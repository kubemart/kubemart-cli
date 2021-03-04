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

	utils "github.com/civo/bizaar/pkg/utils"
	"github.com/spf13/cobra"
)

// systemUpgradeCmd represents the systemUpgrade command
var systemUpgradeCmd = &cobra.Command{
	Use:   "system-upgrade",
	Short: "Upgrade Bizaar CRDs and operator to latest version",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO - change this after we go live
		manifests := strings.Split(operatorYAML, "---")
		err := utils.ApplyOperatorManifest(manifests)
		if err != nil {
			fmt.Printf("Unable to deploy operator - %v\n", err.Error())
			os.Exit(1)
		}

		fmt.Println("System upgrade complete successfully")
	},
}

func init() {
	rootCmd.AddCommand(systemUpgradeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// systemUpgradeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// systemUpgradeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
