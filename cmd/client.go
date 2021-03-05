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

func getKubeClientSetIfCRDIsInstalled() (*kubernetes.Clientset, error) {
	clientset := &kubernetes.Clientset{}
	crdExist := utils.IsCRDExist("apps.kubemart.civo.com")
	if !crdExist {
		errMsg := "App CRD is not found in the cluster\n"
		errMsg += "You can install it by running 'kubemart init' command"
		return clientset, fmt.Errorf(errMsg)
	}

	clientset, err := utils.GetKubeClientSet()
	if err != nil {
		return clientset, fmt.Errorf("Unable to create k8s clientset - %v", err)
	}

	return clientset, nil
}

func createApp(appName string, plan int) (bool, error) {
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

	clientset, err := getKubeClientSetIfCRDIsInstalled()
	if err != nil {
		return false, err
	}

	body, err := json.Marshal(app)
	if err != nil {
		return false, fmt.Errorf("Unable to marshall app's manifest - %v", err)
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

func getApp(appName string) (*operator.App, error) {
	app := &operator.App{}

	clientset, err := getKubeClientSetIfCRDIsInstalled()
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
		return app, fmt.Errorf("Unable to fetch app data - %v", err)
	}

	return app, nil
}

func listApps() (*operator.AppList, error) {
	apps := &operator.AppList{}

	clientset, err := getKubeClientSetIfCRDIsInstalled()
	if err != nil {
		return apps, err
	}

	res := clientset.RESTClient().
		Get().
		AbsPath(baseURL).
		Do(context.Background())

	if res.Error() != nil {
		return apps, fmt.Errorf("Unable to list apps - %v", res.Error())
	}

	err = res.Into(apps)
	if err != nil {
		return apps, fmt.Errorf("Unable to parse apps - %v", err)
	}

	return apps, nil
}

func updateApp(appName string) error {
	clientset, err := getKubeClientSetIfCRDIsInstalled()
	if err != nil {
		return err
	}

	app, err := getApp(appName)
	if err != nil {
		return err
	}

	if !app.ObjectMeta.DeletionTimestamp.IsZero() {
		return fmt.Errorf("This %s app is being deleted - you can't update it", appName)
	}

	if !app.Status.NewUpdateAvailable {
		return fmt.Errorf("There is no new update available for this app - you are already using the latest version")
	}

	app.Spec.Action = "update"
	body, err := json.Marshal(app)
	if err != nil {
		return fmt.Errorf("Unable to marshall app's manifest - %v", err)
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

func deleteApp(appName string) error {
	clientset, err := getKubeClientSetIfCRDIsInstalled()
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
