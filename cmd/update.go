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

	utils "github.com/kubemart/kubemart-cli/pkg/utils"
	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:     "update APP_NAME",
	Example: "kubemart update rabbitmq",
	Short:   "Update an application",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]
		if appName == "" {
			return fmt.Errorf("please provide an app name")
		}
		utils.DebugPrintf("App name to install: %s\n", appName)

		err := UpdateApp(appName)
		if err != nil {
			return fmt.Errorf("unable to update app - %v", err)
		}

		fmt.Printf("%s app is now scheduled to be updated\n", appName)
		return nil
	},
}

func init() {
	KubemartRootCmd.AddCommand(updateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// updateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// updateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
