/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"slices"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	aiv1alpha1 "github.com/m3dev/broom/api/v1alpha1"
	"github.com/m3dev/broom/internal/random"
	"github.com/m3dev/broom/internal/slack"
)

// BroomReconciler reconciles a Broom object
type BroomReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	ResolvedJobs map[types.UID]struct{}
}

//+kubebuilder:rbac:groups=ai.m3.com,resources=brooms,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ai.m3.com,resources=brooms/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ai.m3.com,resources=brooms/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch
//+kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Broom object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *BroomReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	var broom aiv1alpha1.Broom
	if err := r.Get(ctx, req.NamespacedName, &broom); err != nil {
		log.Error(err, "unable to fetch Broom")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	br, err := r.listCandidateBatchResources(ctx, broom.Spec.Target.Namespace)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("unable to fetch candidate resources: %w", err)
	}
	cronJobOwnedReferences := r.traceOOMKilledOwnerReference(br)
	for uid, info := range cronJobOwnedReferences {
		cronJob, ok := br.cronJobs[uid]
		if !ok {
			return ctrl.Result{}, fmt.Errorf("unable to fetch CronJob: UID=%s", uid)
		}
		if !isTargeted(cronJob, broom.Spec.Target) {
			continue
		}
		oldSpec := cronJob.Spec.DeepCopy()
		newSpec, err := r.getNewCronJobSpec(&cronJob, broom.Spec.Adjustment, info.OOMContainerNames)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("unable to get updated CronJob spec: %w", err)
		}

		err = r.updateCronJob(ctx, &cronJob, newSpec)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("unable to update CronJob: %w", err)
		}

		var retriedJobName string
		if broom.Spec.RetryPolicy == aiv1alpha1.RetryAllowPolicy {
			retriedJobName, err = r.retryJob(ctx, &cronJob, info)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("unable to retry Job: %w", err)
			}
		}

		if err := r.notifyResult(ctx, broom.Spec.Webhook, &cronJob, *oldSpec, retriedJobName); err != nil {
			return ctrl.Result{}, fmt.Errorf("unable to notify Slack: %w", err)
		}
	}
	return ctrl.Result{}, nil
}

type batchResources struct {
	pods     map[types.UID]corev1.Pod
	jobs     map[types.UID]batchv1.Job
	cronJobs map[types.UID]batchv1.CronJob
}

// listCandidateResources lists all pods, jobs, cronjobs in the specified namespace
func (r *BroomReconciler) listCandidateBatchResources(ctx context.Context, namespace string) (*batchResources, error) {
	b := &batchResources{
		pods:     make(map[types.UID]corev1.Pod, 0),
		jobs:     make(map[types.UID]batchv1.Job, 0),
		cronJobs: make(map[types.UID]batchv1.CronJob, 0),
	}
	opts := client.ListOptions{
		Namespace: namespace,
	}

	var pods corev1.PodList
	if err := r.List(ctx, &pods, &opts); err != nil {
		return nil, fmt.Errorf("unable to fetch Pods: %w", err)
	}
	for _, pod := range pods.Items {
		b.pods[pod.UID] = pod
	}

	var jobs batchv1.JobList
	if err := r.List(ctx, &jobs, &opts); err != nil {
		return nil, fmt.Errorf("unable to fetch Jobs: %w", err)
	}
	for _, job := range jobs.Items {
		b.jobs[job.UID] = job
	}

	var cronJobs batchv1.CronJobList
	if err := r.List(ctx, &cronJobs, &opts); err != nil {
		return nil, fmt.Errorf("unable to fetch CronJobs: %w", err)
	}
	for _, cronJob := range cronJobs.Items {
		b.cronJobs[cronJob.UID] = cronJob
	}

	return b, nil
}

type cronJobOOMInfo struct {
	LastFailedJob     *batchv1.Job
	OOMContainerNames []string
}

// traceOOMKilledOwnerReference returns a reference from CronJob UID to OOMKilled Pod and referenced Job information
func (r *BroomReconciler) traceOOMKilledOwnerReference(br *batchResources) map[types.UID]cronJobOOMInfo {
	jobOwnedOOMContainers := make(map[metav1.OwnerReference][]string, 0)
	for _, p := range br.pods {
		for _, cs := range p.Status.ContainerStatuses {
			if cs.State.Terminated == nil || cs.State.Terminated.Reason != "OOMKilled" {
				continue
			}
			for _, ref := range p.OwnerReferences {
				if ref.Kind == "Job" {
					jobOwnedOOMContainers[ref] = append(jobOwnedOOMContainers[ref], cs.Name)
				}
			}
		}
	}

	if r.ResolvedJobs == nil {
		r.ResolvedJobs = make(map[types.UID]struct{}, 0)
	}

	cronJobOwnedReferences := make(map[types.UID]cronJobOOMInfo, 0)
	for ref, containerNames := range jobOwnedOOMContainers {
		job, ok := br.jobs[ref.UID]
		if !ok {
			continue // possibly deleted while reconciling
		}
		for _, ownerRef := range job.OwnerReferences {
			if _, ok := r.ResolvedJobs[ref.UID]; !ok && ownerRef.Kind == "CronJob" {
				if oomInfo, ok := cronJobOwnedReferences[ownerRef.UID]; !ok {
					cronJobOwnedReferences[ownerRef.UID] = cronJobOOMInfo{
						LastFailedJob:     &job,
						OOMContainerNames: containerNames,
					}
				} else {
					if job.CreationTimestamp.After(oomInfo.LastFailedJob.CreationTimestamp.Time) {
						oomInfo.LastFailedJob = &job
					}
				}
				r.ResolvedJobs[ref.UID] = struct{}{}
			}
		}
	}

	return cronJobOwnedReferences
}

// getNewCronJobSpec returns a spec with increased memory limit for CronJob
func (r *BroomReconciler) getNewCronJobSpec(cj *batchv1.CronJob, adj aiv1alpha1.BroomAdjustment, containers []string) (*batchv1.CronJobSpec, error) {
	newSpec := cj.Spec.DeepCopy()
	for i, c := range cj.Spec.JobTemplate.Spec.Template.Spec.Containers {
		if !slices.Contains(containers, c.Name) { // Ignore non-OOM Pod containers
			continue
		}
		if m := c.Resources.Limits.Memory(); m != nil {
			if err := adj.IncreaseMemory(m); err != nil {
				return &batchv1.CronJobSpec{}, fmt.Errorf("unable to increase memory: %w", err)
			}
			newSpec.JobTemplate.Spec.Template.Spec.Containers[i].Resources.Limits[corev1.ResourceMemory] = *m
		}
	}
	return newSpec, nil
}

// updateCronJob updates the CronJob in the Kubernetes cluster with given spec
func (r *BroomReconciler) updateCronJob(ctx context.Context, cj *batchv1.CronJob, spec *batchv1.CronJobSpec) error {
	log := log.FromContext(ctx)
	cj.Spec = *spec
	if err := r.Update(ctx, cj); err != nil {
		return fmt.Errorf("unable to update CronJob: %w", err)
	}
	log.Info("Updated CronJob", "name", cj.Name)
	return nil
}

// retryJob creates Job for the failed Job with updated CronJob jobTemplate spec
func (r *BroomReconciler) retryJob(ctx context.Context, cj *batchv1.CronJob, info cronJobOOMInfo) (string, error) {
	log := log.FromContext(ctx)
	randomString := random.GetRandomString(5)
	retriedJobName := fmt.Sprintf("%s-retry-%s", info.LastFailedJob.Name, randomString)
	annotations := map[string]string{
		"rendezvous.m3.com/retried-by-broom": "true",
	}
	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   cj.Namespace,
			Name:        retriedJobName,
			Labels:      info.LastFailedJob.Labels,
			Annotations: annotations,
		},
		Spec: cj.Spec.JobTemplate.Spec,
	}
	if err := r.Create(ctx, &job); err != nil {
		return "", fmt.Errorf("unable to create retried Job: %w", err)
	}
	log.Info("Retried Job", "name", job.Name)
	return job.Name, nil
}

// isTargeted returns whether the Cronjob matches all the fields specified for the target or not
func isTargeted(cj batchv1.CronJob, target aiv1alpha1.BroomTarget) bool {
	if target.Namespace != "" && cj.Namespace != target.Namespace {
		return false
	}
	if target.Name != "" && cj.Name != target.Name {
		return false
	}
	for k, v := range target.Labels {
		val, ok := cj.Labels[k]
		if !ok || val != v {
			return false
		}
	}
	return true
}

// notifyResult notifies the result of changes with webhook information retrieved from Secret
func (r *BroomReconciler) notifyResult(ctx context.Context, w aiv1alpha1.BroomWebhook, cj *batchv1.CronJob, oldSpec batchv1.CronJobSpec, rj string) error {
	secret := &corev1.Secret{}
	if err := r.Get(ctx, client.ObjectKey{Namespace: w.Secret.Namespace, Name: w.Secret.Name}, secret); err != nil {
		return fmt.Errorf("unable to get Secret for webhook URL: %w", err)
	}
	webhookURL := string(secret.Data[w.Secret.Key])
	webhookChannel := w.Channel

	res := slack.UpdateResult{
		CronJobNamespace: cj.Namespace,
		CronJobName:      cj.Name,
		ContainerUpdates: []slack.ContainerUpdate{},
		RetriedJobName:   rj,
	}

	for _, oc := range oldSpec.JobTemplate.Spec.Template.Spec.Containers {
		for _, nc := range cj.Spec.JobTemplate.Spec.Template.Spec.Containers {
			if oc.Name == nc.Name && !(oc.Resources.Limits.Memory().Equal(*nc.Resources.Limits.Memory())) {
				res.ContainerUpdates = append(res.ContainerUpdates, slack.ContainerUpdate{
					Name:         oc.Name,
					BeforeMemory: oc.Resources.Limits.Memory().String(),
					AfterMemory:  nc.Resources.Limits.Memory().String(),
				})
			}
		}
	}

	if err := slack.SendMessage(res, webhookURL, webhookChannel); err != nil {
		return fmt.Errorf("unable to send message to Slack: %w", err)
	}
	return nil
}

// findObjectsForPod finds Brooms to create a reconcile request
func (r *BroomReconciler) findObjectsForPod(ctx context.Context, pod client.Object) []reconcile.Request {
	attachedBrooms := &aiv1alpha1.BroomList{}
	if err := r.List(ctx, attachedBrooms); err != nil {
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(attachedBrooms.Items))
	for i, item := range attachedBrooms.Items {
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		}
	}
	return requests
}

// SetupWithManager sets up the controller with the Manager.
func (r *BroomReconciler) SetupWithManager(mgr ctrl.Manager) error {
	p := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			new := e.ObjectNew.(*corev1.Pod)
			return new.Status.Phase == "Failed"
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&aiv1alpha1.Broom{}).
		Watches(
			&corev1.Pod{},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForPod),
			builder.WithPredicates(p),
		).
		Complete(r)
}
