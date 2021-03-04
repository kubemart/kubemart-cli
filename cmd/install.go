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
	"github.com/forestgiant/sliceutil"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Plan is used for selected apps i.e. mariadb
var Plan int

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install APP_NAME",
	Short: "Install an application",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		appName := args[0]
		if appName == "" {
			fmt.Println("Please provide an app name")
			os.Exit(1)
		}
		utils.DebugPrintf("App name to install: %s\n", appName)

		appExists := utils.IsAppExist(appName)
		if !appExists {
			fmt.Printf("Unable to find %s app. Please try again.\n", appName)
			os.Exit(1)
		}

		appPlans, err := utils.GetAppPlans(appName)
		if err != nil {
			fmt.Printf("Unable to list app's plans - %v\n", err)
			os.Exit(1)
		}

		if len(appPlans) > 0 {
			if Plan == 0 {
				smallestPlan := utils.GetSmallestAppPlan(appPlans)
				if smallestPlan > 0 {
					Plan = smallestPlan
					fmt.Println("This app require a plan. Next time you could use --plan or -p flag.")
					fmt.Printf("Since the flag is not present, this installation will proceed with the smallest one (%dGB)\n", Plan)
				}
			}

			if Plan > 0 {
				if !sliceutil.Contains(appPlans, Plan) {
					fmt.Printf("The given plan is not supported for this app. Supported values are %v.\n", appPlans)
					os.Exit(1)
				}
			}

			utils.DebugPrintf("Plan to proceed with: %d\n", Plan)
		}

		app := &operator.App{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "bizaar.civo.com/v1alpha1",
				Kind:       "App",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      appName,
				Namespace: "bizaar-system",
			},
			Spec: operator.AppSpec{
				Name:   appName,
				Action: "install",
				Plan:   Plan,
			},
		}

		clientset, err := utils.GetKubeClientSet()
		if err != nil {
			fmt.Printf("Unable to create k8s clientset - %v\n", err)
			os.Exit(1)
		}

		body, err := json.Marshal(app)
		if err != nil {
			fmt.Printf("Unable to marshall app's manifest - %v\n", err)
			os.Exit(1)
		}

		wasCreated := false
		res := clientset.RESTClient().
			Post().
			AbsPath("/apis/bizaar.civo.com/v1alpha1/namespaces/bizaar-system/apps").
			Body(body).
			Do(context.Background())

		res = res.WasCreated(&wasCreated)
		if wasCreated {
			fmt.Println("App created successfully")
			utils.RenderPostInstallMarkdown(appName)
			os.Exit(0)
		} else {
			fmt.Printf("App creation failed - %+v\n", res.Error())
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().IntVarP(&Plan, "plan", "p", 0, "Storage plan for the app (in GB) e.g. '5' for 5GB")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// installCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// installCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
