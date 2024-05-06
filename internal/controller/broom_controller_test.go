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
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	aiv1alpha1 "github.com/m3dev/broom/api/v1alpha1"
)

func TestIsTargeted(t *testing.T) {
	tests := map[string]struct {
		cj       batchv1.CronJob
		target   aiv1alpha1.BroomTarget
		expected bool
	}{
		"namespace is matched": {
			cj: batchv1.CronJob{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "targeted-namespace",
					Name:      "test-cronjob",
				},
			},
			target: aiv1alpha1.BroomTarget{
				Namespace: "targeted-namespace",
			},
			expected: true,
		},
		"namespace is not matched": {
			cj: batchv1.CronJob{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test-namespace",
					Name:      "test-cronjob",
				},
			},
			target: aiv1alpha1.BroomTarget{
				Namespace: "targeted-namespace",
			},
			expected: false,
		},
		"name is matched": {
			cj: batchv1.CronJob{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test-namespace",
					Name:      "targeted-cronjob",
				},
			},
			target: aiv1alpha1.BroomTarget{
				Name: "targeted-cronjob",
			},
			expected: true,
		},
		"name is not matched": {
			cj: batchv1.CronJob{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test-namespace",
					Name:      "test-cronjob",
				},
			},
			target: aiv1alpha1.BroomTarget{
				Name: "targeted-cronjob",
			},
			expected: false,
		},
		"labels are matched": {
			cj: batchv1.CronJob{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test-namespace",
					Name:      "test-cronjob",
					Labels: map[string]string{
						"targeted": "true",
						"env":      "dev",
					},
				},
			},
			target: aiv1alpha1.BroomTarget{
				Labels: map[string]string{
					"targeted": "true",
				},
			},
			expected: true,
		},
		"labels are not matched": {
			cj: batchv1.CronJob{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test-namespace",
					Name:      "test-cronjob",
					Labels: map[string]string{
						"targeted": "true",
						"env":      "prod",
					},
				},
			},
			target: aiv1alpha1.BroomTarget{
				Labels: map[string]string{
					"targeted": "true",
					"env":      "dev",
				},
			},
			expected: false,
		},
		"all fields are matched": {
			cj: batchv1.CronJob{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "targeted-namespace",
					Name:      "targeted-cronjob",
					Labels: map[string]string{
						"targeted": "true",
						"env":      "dev",
					},
				},
			},
			target: aiv1alpha1.BroomTarget{
				Namespace: "targeted-namespace",
				Name:      "targeted-cronjob",
				Labels: map[string]string{
					"targeted": "true",
					"env":      "dev",
				},
			},
			expected: true,
		},
		"not all fields are matched": {
			cj: batchv1.CronJob{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "targeted-namespace",
					Name:      "targeted-cronjob",
					Labels: map[string]string{
						"targeted": "true",
						"env":      "dev",
					},
				},
			},
			target: aiv1alpha1.BroomTarget{
				Namespace: "targeted-namespace",
				Name:      "targeted-cronjob",
				Labels: map[string]string{
					"targeted": "true",
					"env":      "prod",
				},
			},
			expected: false,
		},
		"no target": {
			cj: batchv1.CronJob{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "targeted-namespace",
					Name:      "targeted-cronjob",
					Labels: map[string]string{
						"targeted": "true",
						"env":      "dev",
					},
				},
			},
			target:   aiv1alpha1.BroomTarget{},
			expected: true,
		},
	}
	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := isTargeted(test.cj, test.target)
			assert.Equal(t, test.expected, got)
		})
	}
}

var _ = Describe("Broom Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

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
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &aiv1alpha1.Broom{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance Broom")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
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
		})
	})
})
