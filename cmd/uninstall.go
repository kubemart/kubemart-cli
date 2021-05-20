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

	"github.com/spf13/cobra"
)

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:     "uninstall APP_NAME",
	Example: "kubemart uninstall rabbitmq\nkubemart uninstall wordpress,jenkins",
	Short:   "Uninstall application(s)",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cs, err := NewClientFromLocalKubeConfig()
		if err != nil {
			return err
		}

		err = cs.RunUninstall(args)
		if err != nil {
			return err
		}

		return nil
	},
}

func (cs *Clientset) RunUninstall(args []string) error {
	apps := strings.Split(args[0], ",")

	for _, app := range apps {
		err := cs.DeleteApp(app)
		if err != nil {
			return err
		}

		if err != nil {
			return fmt.Errorf("unable to delete %s app - %v", app, err)
		}

		fmt.Printf("%s app is now scheduled to be deleted\n", app)
	}

	return nil
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
