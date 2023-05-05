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
	clusterv1beta1 "open-cluster-management.io/api/cluster/v1beta1"
)

var _ = Describe("Argo Workflow Placement controller", func() {

	const (
		wfName        = "wf-4"
		wfNamespace   = "default"
		clusterName   = "cluster1"
		placementName = "placement-1"
	)

	wfKey := types.NamespacedName{Name: wfName, Namespace: wfNamespace}
	pdKey := types.NamespacedName{Name: placementName, Namespace: wfNamespace}
	ctx := context.Background()

	Context("When Workflow is created with OCM Placement annotation", func() {
		It("Should evaluate the OCM Placement and update the Workflow annotation with the result", func() {
			By("Creating the OCM PlacementDecision")
			pd := clusterv1beta1.PlacementDecision{
				ObjectMeta: metav1.ObjectMeta{
					Name:      placementName,
					Namespace: wfNamespace,
					Labels:    map[string]string{clusterv1beta1.PlacementLabel: placementName},
				},
			}
			Expect(k8sClient.Create(ctx, &pd)).Should(Succeed())
			Expect(k8sClient.Get(ctx, pdKey, &pd)).Should(Succeed())
			pd.Status = clusterv1beta1.PlacementDecisionStatus{
				Decisions: []clusterv1beta1.ClusterDecision{{ClusterName: clusterName}}}
			Expect(k8sClient.Status().Update(ctx, &pd)).Should(Succeed())
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, pdKey, &pd); err != nil {
					return false
				}
				return len(pd.Status.Decisions) > 0
			}).Should(BeTrue())

			By("Creating the Workflow")
			wf := argov1alpha1.Workflow{
				ObjectMeta: metav1.ObjectMeta{
					Name:        wfName,
					Namespace:   wfNamespace,
					Labels:      map[string]string{LabelKeyEnableOCMMulticluster: strconv.FormatBool(true)},
					Annotations: map[string]string{AnnotationKeyOCMPlacement: placementName},
				},
			}
			Expect(k8sClient.Create(ctx, &wf)).Should(Succeed())

			Eventually(func() bool {
				if err := k8sClient.Get(ctx, wfKey, &wf); err != nil {
					return false
				}
				return wf.Status.Phase == argov1alpha1.WorkflowPending &&
					wf.Annotations[AnnotationKeyOCMManagedCluster] == clusterName
			}).Should(BeTrue())
		})
	})
})
