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

	utils "github.com/kubemart/kubemart-cli/pkg/utils"
	"github.com/spf13/cobra"
)

// systemUpgradeCmd represents the systemUpgrade command
var systemUpgradeCmd = &cobra.Command{
	Use:     "system-upgrade",
	Example: "kubemart system-upgrade",
	Short:   "Upgrade Kubemart operator to latest version",
	RunE: func(cmd *cobra.Command, args []string) error {
		operatorYAML, err := utils.GetLatestManifests()
		if err != nil {
			return fmt.Errorf("unable to download latest manifests - %v", err)
		}

		fmt.Println("Upgrading Kubemart components...")
		manifests := strings.Split(operatorYAML, "---")
		err = utils.ApplyManifests(manifests)
		if err != nil {
			return fmt.Errorf("unable to apply manifest upgrade - %v", err)
		}

		fmt.Println("Upgrade complete successfully")
		return nil
	},
}

func init() {
	KubemartRootCmd.AddCommand(systemUpgradeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// systemUpgradeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// systemUpgradeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
