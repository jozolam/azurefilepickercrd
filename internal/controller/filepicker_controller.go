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
	"encoding/xml"
	"fmt"
	"io"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"math/rand/v2"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"

	azurefilepickerv1 "example.com/azurefilepickercrd/api/v1"
)

// FilePickerReconciler reconciles a FilePicker object
type FilePickerReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	FileGetter FileGetter
}

// +kubebuilder:rbac:groups=azurefilepicker.example.com,resources=filepickers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=azurefilepicker.example.com,resources=filepickers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=azurefilepicker.example.com,resources=filepickers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// the FilePicker object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.2/pkg/reconcile
func (r *FilePickerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	filePicker := &azurefilepickerv1.FilePicker{}
	err := r.Get(ctx, req.NamespacedName, filePicker)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if filePicker.Status.ReconciledAt != 0 {
		// We assume that file was already selected
		return ctrl.Result{}, nil
	}

	if len(filePicker.Spec.Account) == 0 {
		// We cannot retrieve list of files without account
		return ctrl.Result{}, fmt.Errorf("account must be set")
	}

	if len(filePicker.Spec.Container) == 0 {
		// We cannot retrieve list of files without container
		return ctrl.Result{}, fmt.Errorf("container must be set")
	}

	// retrieving sas token we expect it to be named "sas-token" and live inside the same namespace as filePicker resource
	sasSecret := &v1.Secret{}
	err = r.Get(ctx, types.NamespacedName{Namespace: req.Namespace, Name: "sas-token"}, sasSecret)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("secret not found: %w", err)
	}
	// we need to decode it and trim newline (might be issue with how I set it up) also we assume that key is "token"
	sasToken := strings.TrimSuffix((string)(sasSecret.Data["token"]), "\n")

	// we get list of fileNames so we can randomly pick one
	files, err := r.FileGetter.getListOfFiles(filePicker, sasToken)
	if err != nil {
		return ctrl.Result{}, err
	}

	// here we pick random file and storing its name in resource,
	// I was not sure if we want to download it but seemed weird that we would store it in less reliable storage
	// in some volume when we already have it stored in cloud. Also, this does not deal with removal of files from azure.
	filePicker.Spec.FileName = files[rand.IntN(len(files))]
	err = r.Update(ctx, filePicker)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

type EnumerationResults struct {
	XMLName xml.Name `xml:"EnumerationResults"`
	Names   []string `xml:"Blobs>Blob>Name"`
}

type FileGetter interface {
	getListOfFiles(filePicker *azurefilepickerv1.FilePicker, sasToken string) ([]string, error)
}

type AzureFileGetter struct{}

func (*AzureFileGetter) getListOfFiles(filePicker *azurefilepickerv1.FilePicker, sasToken string) ([]string, error) {
	resp, err := http.DefaultClient.Get(
		fmt.Sprintf(
			"https://%s.blob.core.windows.net/%s?restype=container&comp=list&%s",
			filePicker.Spec.Account,
			filePicker.Spec.Container,
			sasToken,
		),
	)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var decodedBody EnumerationResults
	err = xml.Unmarshal(body, &decodedBody)
	if err != nil {
		return nil, err
	}

	return decodedBody.Names, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *FilePickerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&azurefilepickerv1.FilePicker{}).
		Complete(r)
}
