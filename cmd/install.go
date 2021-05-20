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
	"strconv"
	"strings"

	"github.com/forestgiant/sliceutil"
	utils "github.com/kubemart/kubemart-cli/pkg/utils"
	"github.com/spf13/cobra"
)

// When this is true, do not print post-install message
var HidePostInstall bool

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:     "install APP_NAME[:PLAN]",
	Example: "kubemart install rabbitmq\nkubemart install wordpress:10GB,jenkins:10GB",
	Short:   "Install application(s)",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appsAndPlans, err := PreRunInstall(args)
		if err != nil {
			return err
		}

		cs, err := NewClientFromLocalKubeConfig()
		if err != nil {
			return err
		}

		err = cs.RunInstall(appsAndPlans)
		if err != nil {
			return err
		}

		return nil
	},
}

func PreRunInstall(args []string) (map[string]string, error) {
	appsAndPlans := make(map[string]string)

	appsCombined := args[0]
	apps := strings.Split(appsCombined, ",")
	for _, app := range apps {
		splitted := strings.Split(app, ":")
		appName := splitted[0]
		appPlan := ""

		if len(splitted) > 1 {
			appPlan = splitted[1]
		}

		appsAndPlans[appName] = appPlan
	}

	for name, plan := range appsAndPlans {
		appExists := utils.IsAppExist(name)
		if !appExists {
			return appsAndPlans, fmt.Errorf("unable to find %s app", name)
		}

		appPlans, err := utils.GetAppPlans(name)
		if err != nil {
			return appsAndPlans, fmt.Errorf("unable to list app's plans - %v", err)
		}

		// TODO - revise this
		if plan == "" {
			plan = "0"
		}

		// TODO - revise this
		appPlan, err := strconv.Atoi(plan)
		if err != nil {
			return appsAndPlans, fmt.Errorf("unable to convert %s app plan to int - %v", name, err)
		}

		if len(appPlans) > 0 {
			if appPlan == 0 {
				smallestPlan := utils.GetSmallestAppPlan(appPlans)
				if smallestPlan > 0 {
					appPlan = smallestPlan
					fmt.Printf("This %s app require a plan. Next time you could use 'APP_NAME[:PLAN]'.\n", name)
					fmt.Printf("Since the plan is not present, this %s installation will proceed with the smallest one (%dGB).\n", name, appPlan)
				}
			}

			if appPlan > 0 {
				if !sliceutil.Contains(appPlans, appPlan) {
					return appsAndPlans, fmt.Errorf("the given plan is not supported for this app - supported values are %v", appPlans)
				}
			}
		}

		utils.DebugPrintf("Plan to proceed with for %s app: %d\n", name, appPlan)
	}

	return appsAndPlans, nil
}

func (cs *Clientset) RunInstall(appsAndPlans map[string]string) error {
	for name, plan := range appsAndPlans {
		// TODO - revise this
		if plan == "" {
			plan = "0"
		}

		// TODO - revise this
		appPlan, err := strconv.Atoi(plan)
		if err != nil {
			return fmt.Errorf("unable to convert %s app plan to int - %v", name, err)
		}

		created, err := cs.CreateApp(name, appPlan)
		if !created {
			return fmt.Errorf("%s app creation failed - %+v", name, err.Error())
		}

		fmt.Printf("App %s created successfully\n", name)
		if !HidePostInstall {
			postInstallMsg, err := utils.GetPostInstallMarkdown(name)
			if err == nil {
				fmt.Printf("App %s post-install notes:\n", name)
				fmt.Println(postInstallMsg)
			}
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().BoolVarP(&HidePostInstall, "quiet", "q", false, "do not show post-install message")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// installCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// installCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
