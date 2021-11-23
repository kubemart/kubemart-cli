package cmd

import (
	"fmt"
	"strings"

	"github.com/kubemart/kubemart-cli/pkg/utils"
	kubemartclient "github.com/kubemart/kubemart-operator/pkg/client/clientset/versioned"
)

// Namespace where we save App custom resource
var targetNamespace = "kubemart-system"

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
		return &kubemartclient.Clientset{}, fmt.Errorf("unable to create kubemart clientset - %v", err)
	}

	return clientset, nil
}
