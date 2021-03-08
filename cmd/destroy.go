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
	"os"
	"strings"

	"github.com/kubemart/kubemart/pkg/utils"
	"github.com/spf13/cobra"
)

var proceedWithoutPrompt bool

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:     "destroy",
	Example: "kubemart destroy",
	Short:   "Completely remove Kubemart and all installed applications",
	Run: func(cmd *cobra.Command, args []string) {
		if !proceedWithoutPrompt {
			var answer string
			fmt.Println("Are you sure want to delete ALL apps and completely remove Kubemart Kubernetes resources e.g. operator, CRDs & etc from your cluster? y/n")
			fmt.Scanln(&answer)
			if answer != "y" {
				fmt.Println("Operation cancelled")
				os.Exit(1)
			}
		}

		apps, err := listApps()
		if err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}

		for _, app := range apps.Items {
			appName := app.ObjectMeta.Name
			fmt.Printf("Deleting %s app...\n", appName)
			err := deleteApp(appName)
			if err != nil {
				fmt.Printf("%v\n", err)
				os.Exit(1)
			}
		}

		fmt.Println("All apps have been deleted")
		fmt.Println("Deleting kubemart Kubernetes objects (operator, CRDs & etc)...")
		operatorYAML, err := utils.GetLatestManifests()
		if err != nil {
			fmt.Printf("Unable to download latest manifests - %v\n", err.Error())
			os.Exit(1)
		}

		manifests := strings.Split(operatorYAML, "---")
		err = utils.DeleteManifests(manifests)
		if err != nil {
			fmt.Printf("Unable to delete manifest - %v\n", err.Error())
			os.Exit(1)
		}

		fmt.Println("All done")
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
