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

package status_sync

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	argov1alpha1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	workflowv1alpha1 "github.com/sarabala1979/argo-wf-ocm-multicluster/api/v1alpha1"
	workflowcontroller "github.com/sarabala1979/argo-wf-ocm-multicluster/controllers/workflow"
)

const (
	wfName      = "wf-5"
	wfNamespace = "default"
)

var _ = Describe("OCM Workflow Status Sync Addon controller", func() {
	wfKey := types.NamespacedName{Name: wfName, Namespace: wfNamespace}
	ctx := context.Background()

	Context("When Workflow is created/update", func() {
		It("Should craete/update WorkflowStatusResult", func() {
			By("Creating the Workflow")
			wf := argov1alpha1.Workflow{
				ObjectMeta: metav1.ObjectMeta{
					Name:      wfName,
					Namespace: wfNamespace,
					Annotations: map[string]string{
						workflowcontroller.AnnotationKeyHubWorkflowName:      wfName,
						workflowcontroller.AnnotationKeyHubWorkflowNamespace: wfNamespace},
				},
			}
			Expect(k8sClient.Create(ctx, &wf)).Should(Succeed())

			wfStatusResultKey := types.NamespacedName{Name: generateHubWorkflowStatusResultName(wf), Namespace: wfNamespace}
			wfStatusResult := workflowv1alpha1.WorkflowStatusResult{}
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, wfStatusResultKey, &wfStatusResult); err != nil {
					return false
				}
				return true
			}).Should(BeTrue())

			By("Updating the Workflow status")
			Expect(k8sClient.Get(ctx, wfKey, &wf)).Should(Succeed())
			wf.Status = argov1alpha1.WorkflowStatus{Phase: argov1alpha1.WorkflowPending}
			Expect(k8sClient.Update(ctx, &wf)).Should(Succeed())
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, wfKey, &wf); err != nil {
					return false
				}
				return wf.Status.Phase == argov1alpha1.WorkflowPending
			}).Should(BeTrue())

			Eventually(func() bool {
				if err := k8sClient.Get(ctx, wfStatusResultKey, &wfStatusResult); err != nil {
					return false
				}
				return wfStatusResult.WorkflowStatus.Phase == argov1alpha1.WorkflowPending
			}).Should(BeTrue())
		})
	})
})
