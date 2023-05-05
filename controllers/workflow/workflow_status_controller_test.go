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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	argov1alpha1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	workflowv1alpha1 "github.com/sarabala1979/argo-wf-ocm-multicluster/api/v1alpha1"
)

var _ = Describe("Argo Workflow Status controller", func() {

	const (
		wfName      = "wf-3"
		wfNamespace = "default"
		clusterName = "cluster1"
	)

	wfKey := types.NamespacedName{Name: wfName, Namespace: wfNamespace}
	ctx := context.Background()

	Context("When WorkflowStatusResult is created/update", func() {
		It("Should update Workflow Status", func() {
			By("Creating the Workflow")
			wf := argov1alpha1.Workflow{
				ObjectMeta: metav1.ObjectMeta{
					Name:      wfName,
					Namespace: wfNamespace,
				},
			}
			Expect(k8sClient.Create(ctx, &wf)).Should(Succeed())

			By("Creating the WorkflowStatus")
			wfStatus := workflowv1alpha1.WorkflowStatusResult{
				ObjectMeta: metav1.ObjectMeta{
					Name:      wfName,
					Namespace: wfNamespace,
					Annotations: map[string]string{AnnotationKeyHubWorkflowName: wfName,
						AnnotationKeyHubWorkflowNamespace: wf.Namespace},
				},
				WorkflowStatus: argov1alpha1.WorkflowStatus{Phase: argov1alpha1.WorkflowPending},
			}
			Expect(k8sClient.Create(ctx, &wfStatus)).Should(Succeed())

			Eventually(func() bool {
				if err := k8sClient.Get(ctx, wfKey, &wf); err != nil {
					return false
				}
				return wf.Status.Phase == argov1alpha1.WorkflowPending
			}).Should(BeTrue())

			By("Updating the WorkflowStatus")
			Expect(k8sClient.Get(ctx, wfKey, &wfStatus)).Should(Succeed())
			wfStatus.WorkflowStatus.Phase = argov1alpha1.WorkflowSucceeded
			Eventually(func() bool {
				if err := k8sClient.Update(ctx, &wfStatus); err != nil {
					return false
				}
				return true
			}).Should(BeTrue())
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, wfKey, &wf); err != nil {
					return false
				}
				return wf.Status.Phase == argov1alpha1.WorkflowSucceeded
			}).Should(BeTrue())
		})
	})
})
