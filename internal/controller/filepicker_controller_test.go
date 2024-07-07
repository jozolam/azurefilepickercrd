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
	v1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	azurefilepickerv1 "example.com/azurefilepickercrd/api/v1"
)

var _ = Describe("FilePicker Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		filepicker := &azurefilepickerv1.FilePicker{}
		sasSecret := &v1.Secret{}
		token := "just for testing check for presence of secret"
		BeforeEach(func() {
			By("creating the custom resource for the Kind FilePicker")
			err := k8sClient.Get(ctx, typeNamespacedName, filepicker)
			if err != nil && errors.IsNotFound(err) {
				resource := &azurefilepickerv1.FilePicker{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: azurefilepickerv1.FilePickerSpec{
						Account:   "filepickereon",
						Container: "images",
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}

			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: typeNamespacedName.Namespace, Name: "sas-token"}, sasSecret)
			if err != nil && errors.IsNotFound(err) {
				resource := &v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "sas-token",
						Namespace: typeNamespacedName.Namespace,
					},
					Data: map[string][]byte{
						"token": ([]byte)(token),
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &azurefilepickerv1.FilePicker{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance FilePicker")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &FilePickerReconciler{
				Client:     k8sClient,
				Scheme:     k8sClient.Scheme(),
				FileGetter: &SimpleFileGetter{},
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			resource := &azurefilepickerv1.FilePicker{}
			err = k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			Expect(resource.Spec.FileName).To(Equal("test"))
		})
	})
})

type SimpleFileGetter struct{}

func (*SimpleFileGetter) getListOfFiles(filePicker *azurefilepickerv1.FilePicker, sasToken string) ([]string, error) {
	return []string{"test"}, nil
}
