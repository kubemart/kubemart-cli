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
	"strings"

	"github.com/forestgiant/sliceutil"
	utils "github.com/kubemart/kubemart-cli/pkg/utils"
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:     "install APP_NAME[:PLAN]",
	Example: "kubemart install rabbitmq\nkubemart install wordpress:10GB,linkerd:\"Linkerd with Dashboard\"",
	Short:   "Install application(s)",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		processedAppsAndPlanLabels, err := PreRunInstall(cmd, args)
		if err != nil {
			return err
		}

		cs, err := NewClientFromLocalKubeConfig()
		if err != nil {
			return err
		}

		err = cs.RunInstall(processedAppsAndPlanLabels)
		if err != nil {
			return err
		}

		return nil
	},
}

func PreRunInstall(cmd *cobra.Command, args []string) (map[string]string, error) {
	appsAndPlanLabels := make(map[string]string)

	appsCombined := args[0]
	apps := strings.Split(appsCombined, ",")
	for _, app := range apps {
		splitted := strings.Split(app, ":")
		appName := splitted[0]
		appPlanLabel := ""

		if len(splitted) > 1 {
			appPlanLabel = splitted[1]
		}

		appsAndPlanLabels[appName] = appPlanLabel
	}

	processedAppsAndPlanLabels := make(map[string]string)
	for appName, planLabel := range appsAndPlanLabels {
		appExists := utils.IsAppExist(appName)
		if !appExists {
			return appsAndPlanLabels, fmt.Errorf("unable to find %s app", appName)
		}

		appPlanLabels, err := utils.GetAppPlans(appName)
		if err != nil {
			return appsAndPlanLabels, fmt.Errorf("unable to list app's plans - %v", err)
		}

		if len(appPlanLabels) > 0 {
			firstPlan := utils.GetSmallestAppPlan(appPlanLabels)
			if planLabel == "" {
				planLabel = firstPlan
				fmt.Printf("This %s app require a plan. Next time you could use '%s APP_NAME[:PLAN]' format.\n", appName, cmd.CommandPath())
				fmt.Printf("Since the plan is not present, this %s installation will proceed with the smallest one (%s).\n", appName, planLabel)
			}

			if !sliceutil.Contains(appPlanLabels, planLabel) {
				return appsAndPlanLabels, fmt.Errorf("the given plan is not supported for %s app - supported values are %v", appName, strings.Join(appPlanLabels, ", "))
			}
		}

		processedAppsAndPlanLabels[appName] = planLabel
		utils.DebugPrintf("Plan to proceed with for %s app: %s\n", appName, planLabel)
	}

	return processedAppsAndPlanLabels, nil
}

func (cs *Clientset) RunInstall(processedAppsAndPlanLabels map[string]string) error {
	createdApps := []string{}

	for appName, appPlan := range processedAppsAndPlanLabels {
		if appPlan != "" {
			plan, err := utils.GetAppPlanValueByLabel(appName, appPlan)
			if err != nil {
				return err
			}
			appPlan = plan
		}

		created, err := cs.CreateApp(appName, appPlan)
		if !created {
			return fmt.Errorf("%s app creation failed - %+v", appName, err.Error())
		} else {
			createdApps = append(createdApps, appName)
		}
	}

	if len(createdApps) > 0 {
		fmt.Printf("App(s) created successfully: %s\n", strings.Join(createdApps, ", "))
	}

	return nil
}

func init() {
	rootCmd.AddCommand(installCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// installCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// installCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
