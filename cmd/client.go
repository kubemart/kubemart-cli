package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kubemart/kubemart-cli/pkg/utils"
	operator "github.com/kubemart/kubemart-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	baseURL = "/apis/kubemart.civo.com/v1alpha1/namespaces/kubemart-system/apps"
)

// Clientset is used as receiver object in few functions below
type Clientset struct {
	*kubernetes.Clientset
}

func checkIfCrdExists() error {
	crdExist, err := utils.IsCRDExist("apps.kubemart.civo.com")
	if err != nil {
		return err
	}

	if !crdExist {
		errMsg := "App CRD is not found in the cluster\n"
		errMsg += "You can install it by running 'kubemart init' command"
		return fmt.Errorf(errMsg)
	}
	return nil
}

func NewClientFromLocalKubeConfig() (*Clientset, error) {
	err := checkIfCrdExists()
	if err != nil {
		return &Clientset{}, err
	}

	cs, err := utils.GetKubeClientSet()
	if err != nil {
		return &Clientset{}, fmt.Errorf("unable to create k8s clientset - %v", err)
	}

	return &Clientset{cs}, nil
}

// NewClientFromKubeConfigString is called by Civo CLI
func NewClientFromKubeConfigString(kubeconfig string) (*Clientset, error) {
	kcBytes := []byte(kubeconfig)
	rc, err := clientcmd.RESTConfigFromKubeConfig(kcBytes)
	if err != nil {
		return &Clientset{}, err
	}

	cs, err := kubernetes.NewForConfig(rc)
	if err != nil {
		return &Clientset{}, err
	}

	return &Clientset{cs}, nil
}

// CreateApp will create an App in user's cluster
func (cs *Clientset) CreateApp(appName string, plan string) (bool, error) {
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

	body, err := json.Marshal(app)
	if err != nil {
		return false, fmt.Errorf("unable to marshall app's manifest - %v", err)
	}

	created := false
	err = cs.RESTClient().
		Post().
		AbsPath(baseURL).
		Body(body).
		Do(context.Background()).
		WasCreated(&created).
		Error()

	return created, err
}

// GetApp will get an App from user's cluster
func (cs *Clientset) GetApp(appName string) (*operator.App, error) {
	app := &operator.App{}

	path := fmt.Sprintf("%s/%s", baseURL, appName)
	err := cs.RESTClient().
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
func (cs *Clientset) ListApps() (*operator.AppList, error) {
	apps := &operator.AppList{}

	res := cs.RESTClient().
		Get().
		AbsPath(baseURL).
		Do(context.Background())

	if res.Error() != nil {
		return apps, fmt.Errorf("unable to list apps - %v", res.Error())
	}

	err := res.Into(apps)
	if err != nil {
		return apps, fmt.Errorf("unable to parse apps - %v", err)
	}

	return apps, nil
}

// UpdateApp will update an App in user's cluster
func (cs *Clientset) UpdateApp(appName string) error {
	app, err := cs.GetApp(appName)
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
	err = cs.RESTClient().
		Patch(types.MergePatchType).
		AbsPath(path).
		Body(body).
		Do(context.Background()).
		Error()

	return err
}

// DeleteApp will delete an App from user's cluster
func (cs *Clientset) DeleteApp(appName string) error {
	path := fmt.Sprintf("%s/%s", baseURL, appName)
	err := cs.RESTClient().
		Delete().
		AbsPath(path).
		Do(context.Background()).
		Error()

	return err
}
