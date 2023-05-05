/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by wflicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package workflow

import (
	"context"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	argov1alpha1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
	workv1 "open-cluster-management.io/api/work/v1"
)

var _ = Describe("Argo Workflow controller", func() {

	const (
		wfName      = "wf-1"
		wfName2     = "wf-2"
		wfNamespace = "default"
		clusterName = "cluster1"
	)

	wfKey := types.NamespacedName{Name: wfName, Namespace: wfNamespace}
	wfKey2 := types.NamespacedName{Name: wfName2, Namespace: wfNamespace}
	ctx := context.Background()

	Context("When Workflow without OCM pull label is created", func() {
		It("Should not create ManifestWork", func() {
			By("Creating the Workflow without OCM pull label")
			wf1 := argov1alpha1.Workflow{
				ObjectMeta: metav1.ObjectMeta{
					Name:        wfName,
					Namespace:   wfNamespace,
					Annotations: map[string]string{AnnotationKeyOCMManagedCluster: clusterName},
				},
			}
			Expect(k8sClient.Create(ctx, &wf1)).Should(Succeed())
			wf1 = argov1alpha1.Workflow{}
			Expect(k8sClient.Get(ctx, wfKey, &wf1)).Should(Succeed())

			mwKey := types.NamespacedName{Name: generateManifestWorkName(wf1), Namespace: clusterName}
			mw := workv1.ManifestWork{}
			Consistently(func() bool {
				if err := k8sClient.Get(ctx, mwKey, &mw); err != nil {
					return false
				}
				return true
			}).Should(BeFalse())
		})
	})

	Context("When Workflow with OCM pull label is created/updated/deleted", func() {
		It("Should create/update/delete ManifestWork", func() {
			By("Creating the OCM ManagedCluster")
			managedCluster := clusterv1.ManagedCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: clusterName,
				},
			}
			Expect(k8sClient.Create(ctx, &managedCluster)).Should(Succeed())

			By("Creating the OCM ManagedCluster namespace")
			managedClusterNs := corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: clusterName,
				},
			}
			Expect(k8sClient.Create(ctx, &managedClusterNs)).Should(Succeed())

			By("Creating the Workflow with OCM pull label")
			wf2 := argov1alpha1.Workflow{
				ObjectMeta: metav1.ObjectMeta{
					Name:        wfName2,
					Namespace:   wfNamespace,
					Labels:      map[string]string{LabelKeyEnableOCMMulticluster: strconv.FormatBool(true)},
					Annotations: map[string]string{AnnotationKeyOCMManagedCluster: clusterName},
				},
				Spec: argov1alpha1.WorkflowSpec{
					Entrypoint: "whalesay",
				},
			}
			Expect(k8sClient.Create(ctx, &wf2)).Should(Succeed())
			wf2 = argov1alpha1.Workflow{}
			Expect(k8sClient.Get(ctx, wfKey2, &wf2)).Should(Succeed())

			mwKey := types.NamespacedName{Name: generateManifestWorkName(wf2), Namespace: clusterName}
			mw := workv1.ManifestWork{}
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, mwKey, &mw); err != nil {
					return false
				}
				return true
			}).Should(BeTrue())

			By("Updating the Workflow")
			oldRv := mw.GetResourceVersion()
			Expect(k8sClient.Get(ctx, wfKey2, &wf2)).Should(Succeed())
			wf2.Spec.Entrypoint = "whalesay2"
			Eventually(func() bool {
				if err := k8sClient.Update(ctx, &wf2); err != nil {
					return false
				}
				return true
			}).Should(BeTrue())
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, mwKey, &mw); err != nil {
					return false
				}
				return oldRv != mw.GetResourceVersion()
			}).Should(BeTrue())

			By("Deleting the Workflow")
			Expect(k8sClient.Delete(ctx, &wf2)).Should(Succeed())
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, mwKey, &mw); err != nil {
					return true
				}
				return false
			}).Should(BeTrue())
		})
	})
})
