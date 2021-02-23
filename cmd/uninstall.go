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

	utils "github.com/civo/bizaar/pkg/utils"
	"github.com/spf13/cobra"
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
			fmt.Printf("Unable to create k8s clientset - %v\n", err)
			os.Exit(1)
		}

		// TODO - use bizaar-operator Go client
		path := fmt.Sprintf("/apis/bizaar.civo.com/v1alpha1/namespaces/default/apps/%s", appName)
		res := clientset.RESTClient().
			Delete().
			AbsPath(path).
			Do(context.Background())

		if res.Error() != nil {
			fmt.Printf("Unable to delete %s app - %v\n", appName, res.Error())
			os.Exit(1)
		} else {
			fmt.Printf("%s app is now scheduled to be deleted\n", appName)
			os.Exit(0)
		}
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
