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
	"fmt"

	"github.com/forestgiant/sliceutil"
	utils "github.com/kubemart/kubemart/pkg/utils"
	"github.com/spf13/cobra"
)

// Plan is used for selected apps i.e. mariadb
var Plan int

// When this is true, do not print post-install message
var hidePostInstall bool

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:     "install APP_NAME",
	Example: "kubemart install rabbitmq",
	Short:   "Install an application",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]
		if appName == "" {
			return fmt.Errorf("please provide an app name")
		}
		utils.DebugPrintf("App name to install: %s\n", appName)

		appExists := utils.IsAppExist(appName)
		if !appExists {
			return fmt.Errorf("unable to find %s app", appName)
		}

		appPlans, err := utils.GetAppPlans(appName)
		if err != nil {
			return fmt.Errorf("unable to list app's plans - %v", err)
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
					return fmt.Errorf("the given plan is not supported for this app - supported values are %v", appPlans)
				}
			}

			utils.DebugPrintf("Plan to proceed with: %d\n", Plan)
		}

		created, err := CreateApp(appName, Plan)
		if !created {
			return fmt.Errorf("app creation failed - %+v", err.Error())
		}

		fmt.Println("App created successfully")
		if !hidePostInstall {
			postInstallMsg, err := utils.GetPostInstallMarkdown(appName)
			if err == nil {
				fmt.Println("App post-install notes:")
				fmt.Println(postInstallMsg)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().IntVarP(&Plan, "plan", "p", 0, "storage plan for the app (in GB) e.g. '5' for 5GB")
	installCmd.Flags().BoolVarP(&hidePostInstall, "quiet", "q", false, "do not show post-install message")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// installCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// installCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
