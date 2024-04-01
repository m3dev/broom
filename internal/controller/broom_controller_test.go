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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	aiv1alpha1 "github.com/m3dev/broom/api/v1alpha1"
)

var _ = Describe("Broom Controller", func() {
	Context("When reconciling a resource", func() {
		const (
			resourceName     = "broom-sample"
			cronJobName      = "oom-sample"
			cronJobNamespace = "broom"
		)

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		broom := &aiv1alpha1.Broom{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind Broom")
			err := k8sClient.Get(ctx, typeNamespacedName, broom)
			if err != nil && errors.IsNotFound(err) {
				resource := &aiv1alpha1.Broom{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					// TODO(user): Specify other spec details if needed.
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}

			By("creating the Namespace")
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: cronJobNamespace,
				},
			}
			Expect(k8sClient.Create(ctx, ns)).To(Succeed())

			By("creating the CronJob")
			cronJob := &batchv1.CronJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cronJobName,
					Namespace: cronJobNamespace,
					Labels: map[string]string{
						"m3.com/use-broom": "true",
					},
				},
				Spec: batchv1.CronJobSpec{
					Schedule: "*/2 * * * *",
					JobTemplate: batchv1.JobTemplateSpec{
						Spec: batchv1.JobSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:    "oom-container",
											Image:   "ubuntu:latest",
											Command: []string{"/bin/bash", "-c"},
											Args: []string{
												"echo PID=$$; for i in {0..9}; do eval a$i'=$(head --bytes 5000000 /dev/zero | cat -v)'; echo $((i++));	done",
											},
											Resources: corev1.ResourceRequirements{
												Limits: corev1.ResourceList{
													corev1.ResourceMemory: resource.MustParse("100Mi"),
												},
												Requests: corev1.ResourceList{
													corev1.ResourceMemory: resource.MustParse("50Mi"),
												},
											},
										},
									},
									RestartPolicy: corev1.RestartPolicyNever,
								},
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, cronJob)).To(Succeed())
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &aiv1alpha1.Broom{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance Broom")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())

			By("Cleanup the CronJob")
			cronJob := &batchv1.CronJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cronJobName,
					Namespace: cronJobNamespace,
				},
			}
			Expect(k8sClient.Delete(ctx, cronJob)).To(Succeed())

			By("Cleanup the Namespace")
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: cronJobNamespace,
				},
			}
			Expect(k8sClient.Delete(ctx, ns)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &BroomReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.

			By("Checking if the CronJob is targeted correctly")
			cronJob := &batchv1.CronJob{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: cronJobName, Namespace: cronJobNamespace}, cronJob)).To(Succeed())

			target := aiv1alpha1.BroomTarget{
				Namespace: cronJobNamespace,
				Name:      cronJobName,
				Labels: map[string]string{
					"m3.com/use-broom": "true",
				},
			}
			res := isTargeted(*cronJob, target)
			Expect(res).To(BeTrue(), "Expected CronJob to be targeted but it is not")

			emptyTarget := aiv1alpha1.BroomTarget{}
			res = isTargeted(*cronJob, emptyTarget)
			Expect(res).To(BeTrue(), "Expected CronJob to be targeted but it is not")

			wrongNamespaceTarget := aiv1alpha1.BroomTarget{
				Namespace: "default",
			}
			res = isTargeted(*cronJob, wrongNamespaceTarget)
			Expect(res).To(BeFalse(), "Expected CronJob not to be targeted but it is")

			wrongLabelTarget := aiv1alpha1.BroomTarget{
				Labels: map[string]string{
					"m3.com/foo": "bar",
				},
			}
			res = isTargeted(*cronJob, wrongLabelTarget)
			Expect(res).To(BeFalse(), "Expected CronJob not to be targeted but it is")
		})
	})
})
