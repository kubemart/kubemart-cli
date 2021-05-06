/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

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
	"time"

	"github.com/kubemart/kubemart-cli/pkg/utils"
	"github.com/spf13/cobra"
)

var proceedWithoutPrompt bool

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:     "destroy",
	Example: "kubemart destroy",
	Short:   "Completely remove Kubemart and all installed applications",
	RunE: func(cmd *cobra.Command, args []string) error {
		var answer string

		if proceedWithoutPrompt {
			answer = "y"
		} else {
			fmt.Println("Are you sure want to delete ALL apps and completely remove Kubemart Kubernetes resources e.g. operator, CRDs & etc from your cluster? y/n")
			fmt.Scanln(&answer)
		}

		if answer != "y" {
			return fmt.Errorf("operation cancelled")
		}

		apps, err := ListApps()
		if err != nil {
			return err
		}

		deletedApps := []string{}
		for _, app := range apps.Items {
			appName := app.ObjectMeta.Name
			fmt.Printf("Deleting %s app...\n", appName)
			err := DeleteApp(appName)
			if err != nil {
				return err
			}

			fmt.Printf("Waiting %s app to get deleted...\n", appName)
			for i := 0; i < 120; i++ {
				_, err := GetApp(appName)
				if err != nil {
					fmt.Printf("%s app has been deleted...\n", appName)
					deletedApps = append(deletedApps, appName)
					break
				}
				time.Sleep(1 * time.Second)
			}
		}

		if len(apps.Items) != len(deletedApps) {
			return fmt.Errorf("some apps didn't get deleted successfully - please rerun this command")
		}

		fmt.Println("All apps have been deleted")
		fmt.Println("Deleting kubemart Kubernetes objects (operator, CRDs & etc)...")
		operatorYAML, err := utils.GetLatestManifests()
		if err != nil {
			return fmt.Errorf("unable to download latest manifests - %v", err.Error())
		}

		manifests := strings.Split(operatorYAML, "---")
		err = utils.DeleteManifests(manifests)
		if err != nil {
			return fmt.Errorf("unable to delete manifest - %v", err.Error())
		}

		fmt.Println("All done")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
	destroyCmd.Flags().BoolVarP(&proceedWithoutPrompt, "yes", "y", false, "skip interactive y/n prompt by answering 'y'")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// destroyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// destroyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
