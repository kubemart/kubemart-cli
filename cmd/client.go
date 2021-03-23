package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	operator "github.com/kubemart/kubemart-operator/api/v1alpha1"
	"github.com/kubemart/kubemart/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

const (
	baseURL = "/apis/kubemart.civo.com/v1alpha1/namespaces/kubemart-system/apps"
)

// GetKubeClientSetIfCRDIsInstalled returns Kubernetes client if the
// App CRD exists in the user's cluster
func GetKubeClientSetIfCRDIsInstalled() (*kubernetes.Clientset, error) {
	clientset := &kubernetes.Clientset{}
	crdExist := utils.IsCRDExist("apps.kubemart.civo.com")
	if !crdExist {
		errMsg := "App CRD is not found in the cluster\n"
		errMsg += "You can install it by running 'kubemart init' command"
		return clientset, fmt.Errorf(errMsg)
	}

	clientset, err := utils.GetKubeClientSet()
	if err != nil {
		return clientset, fmt.Errorf("unable to create k8s clientset - %v", err)
	}

	return clientset, nil
}

// CreateApp will create an App in user's cluster
func CreateApp(appName string, plan int) (bool, error) {
	app := &operator.App{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubemart.civo.com/v1alpha1",
			Kind:       "App",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: "kubemart-system",
		},
		Spec: operator.AppSpec{
			Name:   appName,
			Action: "install",
			Plan:   plan,
		},
	}

	clientset, err := GetKubeClientSetIfCRDIsInstalled()
	if err != nil {
		return false, err
	}

	body, err := json.Marshal(app)
	if err != nil {
		return false, fmt.Errorf("unable to marshall app's manifest - %v", err)
	}

	created := false
	err = clientset.RESTClient().
		Post().
		AbsPath(baseURL).
		Body(body).
		Do(context.Background()).
		WasCreated(&created).
		Error()

	return created, err
}

// GetApp will get an App from user's cluster
func GetApp(appName string) (*operator.App, error) {
	app := &operator.App{}

	clientset, err := GetKubeClientSetIfCRDIsInstalled()
	if err != nil {
		return app, err
	}

	path := fmt.Sprintf("%s/%s", baseURL, appName)
	err = clientset.RESTClient().
		Get().
		AbsPath(path).
		Do(context.Background()).
		Into(app)

	if err != nil {
		return app, fmt.Errorf("unable to fetch app data - %v", err)
	}

	return app, nil
}

// ListApps will get all Apps from user's cluster
func ListApps() (*operator.AppList, error) {
	apps := &operator.AppList{}

	clientset, err := GetKubeClientSetIfCRDIsInstalled()
	if err != nil {
		return apps, err
	}

	res := clientset.RESTClient().
		Get().
		AbsPath(baseURL).
		Do(context.Background())

	if res.Error() != nil {
		return apps, fmt.Errorf("unable to list apps - %v", res.Error())
	}

	err = res.Into(apps)
	if err != nil {
		return apps, fmt.Errorf("unable to parse apps - %v", err)
	}

	return apps, nil
}

// UpdateApp will update an App in user's cluster
func UpdateApp(appName string) error {
	clientset, err := GetKubeClientSetIfCRDIsInstalled()
	if err != nil {
		return err
	}

	app, err := GetApp(appName)
	if err != nil {
		return err
	}

	if !app.ObjectMeta.DeletionTimestamp.IsZero() {
		return fmt.Errorf("this %s app is being deleted - you can't update it", appName)
	}

	if !app.Status.NewUpdateAvailable {
		return fmt.Errorf("there is no new update available for this app - you are already using the latest version")
	}

	app.Spec.Action = "update"
	body, err := json.Marshal(app)
	if err != nil {
		return fmt.Errorf("unable to marshall app's manifest - %v", err)
	}

	path := fmt.Sprintf("%s/%s", baseURL, appName)
	err = clientset.RESTClient().
		Patch(types.MergePatchType).
		AbsPath(path).
		Body(body).
		Do(context.Background()).
		Error()

	return err
}

// DeleteApp will delete an App from user's cluster
func DeleteApp(appName string) error {
	clientset, err := GetKubeClientSetIfCRDIsInstalled()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("%s/%s", baseURL, appName)
	err = clientset.RESTClient().
		Delete().
		AbsPath(path).
		Do(context.Background()).
		Error()

	return err
}
