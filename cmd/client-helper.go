package cmd

import (
	"fmt"
	"strings"

	"github.com/kubemart/kubemart-cli/pkg/utils"
	kubemartclient "github.com/kubemart/kubemart-operator/pkg/client/clientset/versioned"
)

// Namespace where we save App custom resource
var targetNamespace = "kubemart-system"

// KmClientset is used as receiver object in few functions below
// type KmClientset struct {
// 	*kubemartclient.Clientset
// }

func checkIfCrdExists() error {
	crdExist, err := utils.IsCRDExist("apps.kubemart.civo.com")

	if !crdExist || err != nil {
		errMsg := strings.Join([]string{
			"App CRD is not found in the cluster.",
			"You can install it by running 'kubemart init' command.",
		}, " ")
		return fmt.Errorf(errMsg)
	}

	return nil
}

func NewClientFromLocalKubeConfig() (*kubemartclient.Clientset, error) {
	err := checkIfCrdExists()
	if err != nil {
		return &kubemartclient.Clientset{}, err
	}

	restconfig, err := utils.GetRESTConfig()
	if err != nil {
		return &kubemartclient.Clientset{}, err
	}

	clientset, err := kubemartclient.NewForConfig(restconfig)
	if err != nil {
		return &kubemartclient.Clientset{}, fmt.Errorf("unable to create k8s clientset - %v", err)
	}

	return clientset, nil
}

// func NewClientFromLocalKubeConfig() (*KmClientset, error) {
// 	err := checkIfCrdExists()
// 	if err != nil {
// 		return &KmClientset{}, err
// 	}

// 	restconfig, err := utils.GetRESTConfig()
// 	if err != nil {
// 		return &KmClientset{}, err
// 	}

// 	clientset, err := kubemartclient.NewForConfig(restconfig)
// 	if err != nil {
// 		return &KmClientset{}, fmt.Errorf("unable to create k8s clientset - %v", err)
// 	}

// 	return &KmClientset{clientset}, nil
// }

// // CreateApp will create an App in user's cluster
// func (cs *KmClientset) CreateApp(appName string, plan string) (*v1alpha1.App, error) {
// 	return cs.Clientset.KubemartV1alpha1().Apps(targetNamespace).Create(
// 		context.Background(),
// 		&v1alpha1.App{
// 			ObjectMeta: metav1.ObjectMeta{
// 				Name: appName,
// 			},
// 			Spec: v1alpha1.AppSpec{
// 				Name:   appName,
// 				Action: "install",
// 				Plan:   plan,
// 			},
// 		},
// 		v1.CreateOptions{},
// 	)
// }

// // GetApp will get an App from user's cluster
// func (cs *KmClientset) GetApp(appName string) (*v1alpha1.App, error) {
// 	return cs.Clientset.KubemartV1alpha1().Apps(targetNamespace).Get(context.Background(), appName, v1.GetOptions{})
// }

// // ListApps will get all Apps from user's cluster
// func (cs *KmClientset) ListApps() (*v1alpha1.AppList, error) {
// 	return cs.Clientset.KubemartV1alpha1().Apps(targetNamespace).List(context.Background(), v1.ListOptions{})
// }

// // UpdateApp will update an App in user's cluster
// func (cs *KmClientset) UpdateApp(appName string) (*v1alpha1.App, error) {
// 	return cs.Clientset.KubemartV1alpha1().Apps(targetNamespace).Update(
// 		context.Background(),
// 		&v1alpha1.App{
// 			ObjectMeta: metav1.ObjectMeta{
// 				Name: appName,
// 			},
// 			Spec: v1alpha1.AppSpec{
// 				Action: "update",
// 			},
// 		},
// 		v1.UpdateOptions{},
// 	)
// }

// // DeleteApp will delete an App from user's cluster
// func (cs *KmClientset) DeleteApp(appName string) error {
// 	return cs.Clientset.KubemartV1alpha1().Apps(targetNamespace).Delete(context.Background(), appName, v1.DeleteOptions{})
// }
