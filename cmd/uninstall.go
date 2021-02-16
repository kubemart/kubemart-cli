/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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
package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	operator "github.com/civo/bizaar-operator/api/v1alpha1"
	utils "github.com/civo/bizaar/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"
)

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall an application",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		appName := args[0]
		if appName == "" {
			fmt.Println("Please provide an app name")
			os.Exit(1)
		}
		utils.DebugPrintf("App name to uninstall: %s\n", appName)

		clientset, err := utils.GetKubeClientSet()
		if err != nil {
			fmt.Printf("Unable to create k8s clientset - %v", err)
			os.Exit(1)
		}

		app := &operator.App{}
		path := fmt.Sprintf("/apis/app.bizaar.civo.com/v1alpha1/namespaces/default/apps/%s", appName)
		clientset.RESTClient().
			Get().
			AbsPath(path).
			Do(context.Background()).
			Into(app)

		if app.Status.LastStatus != "installation_finished" {
			fmt.Printf("Unable to delete %s app. This can be due to few factors such as:\n", appName)
			fmt.Printf("* There is no such %s app in the marketplace\n", appName)
			fmt.Printf("* You don't have %s app installed in this cluster\n", appName)
			fmt.Printf("* The %s app is still in create or update phase\n", appName)
			os.Exit(1)
		}

		namespace, err := utils.GetAppNamespace(appName)
		if err != nil {
			fmt.Printf("Unable to determine app namespace from app's manifest.yaml file - %v\n", err)
			os.Exit(1)
		}

		if namespace == "" {
			fmt.Println("Unable to uninstall this app because it does not have namespace declared in its manifest.yaml file")
			os.Exit(1)
		}

		namespaceExists, err := utils.IsNamespaceExist(namespace)
		if err != nil {
			fmt.Printf("Unable to check namespace - %v\n", err.Error())
			os.Exit(1)
		}

		if namespaceExists {
			fmt.Printf("Deleting %s namespace for %s app...\n", namespace, appName)
			err = utils.DeleteNamespace(namespace)
			if err != nil {
				fmt.Printf("Unable to delete namespace - %v\n", err.Error())
				os.Exit(1)
			}
		}

		fmt.Printf("Waiting for %s namespace deletion to finish...\n", namespace)
		var sleepDuration time.Duration = 5 // seconds
		var maxTries int = 60
		var tries int = 0
		var handleFinalizer bool = false

		for {
			if tries > maxTries {
				handleFinalizer = true
				break
			}

			nsExists, err := utils.IsNamespaceExist(namespace)
			if err != nil {
				fmt.Printf("Unable to check namespace - %v\n", err.Error())
				os.Exit(1)
			}

			if !nsExists {
				break
			}

			time.Sleep(sleepDuration * time.Second)
			tries++
		}

		if handleFinalizer {
			fmt.Println("Clearing namespace finalizer...")
			app.ObjectMeta.Finalizers = []string{}
			updateFinalizerRes := clientset.RESTClient().
				Patch(types.JSONPatchType).
				AbsPath(path).
				Body(app).
				Do(context.Background())
			if updateFinalizerRes.Error() != nil {
				fmt.Printf("Unable to clear %s namespace finalizers - %v\n", namespace, updateFinalizerRes.Error())
				os.Exit(1)
			}
		}

		fmt.Printf("Deleting %s app...\n", appName)
		deleteRes := clientset.RESTClient().
			Delete().
			AbsPath(path).
			Do(context.Background())

		if deleteRes.Error() != nil {
			fmt.Printf("Unable to delete %s app - %v\n", appName, deleteRes.Error())
			os.Exit(1)
		}

		fmt.Printf("%s app (was running in %s namespace) successfully deleted\n", appName, namespace)
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// uninstallCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// uninstallCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
