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

	"github.com/kubemart/kubemart-cli/pkg/utils"
	"github.com/spf13/cobra"
)

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:     "show",
	Example: "kubemart show APP_NAME",
	Short:   "Show the application's post-install message",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cs, err := NewClientFromLocalKubeConfig()
		if err != nil {
			return err
		}

		err = cs.RunShow(args)
		if err != nil {
			return err
		}

		return nil
	},
}

func (cs *Clientset) RunShow(args []string) error {
	appName := args[0]
	if appName == "" {
		return fmt.Errorf("app name is empty")
	}

	// check if app exists in cluster
	_, err := cs.GetApp(appName)
	if err != nil {
		return fmt.Errorf("%s app is not installed in this cluster", appName)
	}

	appPostInstall, err := utils.GetPostInstallMarkdown(appName)
	if err != nil {
		return err
	}

	fmt.Println(appPostInstall)
	return nil
}

func init() {
	rootCmd.AddCommand(showCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// showCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// showCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
