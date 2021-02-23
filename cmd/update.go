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
	"encoding/json"
	"fmt"
	"os"

	operator "github.com/civo/bizaar-operator/api/v1alpha1"
	utils "github.com/civo/bizaar/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an application",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		appName := args[0]
		if appName == "" {
			fmt.Println("Please provide an app name")
			os.Exit(1)
		}
		utils.DebugPrintf("App name to install: %s\n", appName)

		clientset, err := utils.GetKubeClientSet()
		if err != nil {
			fmt.Printf("Unable to create k8s clientset - %v\n", err)
			os.Exit(1)
		}

		app := &operator.App{}
		path := fmt.Sprintf("/apis/bizaar.civo.com/v1alpha1/namespaces/default/apps/%s", appName)
		err = clientset.RESTClient().
			Get().
			AbsPath(path).
			Do(context.Background()).
			Into(app)
		if err != nil {
			fmt.Printf("Unable to fetch app data - %v\n", err)
			os.Exit(1)
		}

		if !app.ObjectMeta.DeletionTimestamp.IsZero() {
			fmt.Println("This app is being deleted. You can't update it.")
			os.Exit(1)
		}

		if !app.Status.NewUpdateAvailable {
			fmt.Println("There is no new update available for this app. You are already using the latest version.")
			os.Exit(1)
		}

		app.Spec.Action = "update"
		body, err := json.Marshal(app)
		if err != nil {
			fmt.Printf("Unable to marshall app's manifest - %v\n", err)
			os.Exit(1)
		}

		err = clientset.RESTClient().
			Patch(types.MergePatchType).
			AbsPath(path).
			Body(body).
			Do(context.Background()).
			Error()
		if err != nil {
			fmt.Printf("Unable to update app - %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("%s app is now scheduled to be updated\n", appName)
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// updateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// updateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
