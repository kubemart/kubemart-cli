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
	"os"

	utils "github.com/civo/bizaar/pkg/utils"
	"github.com/spf13/cobra"
)

var kubeCfgFile string
var canSkipUpdateApps map[string]bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bizaar",
	Short: "",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		canSkipUpdateApps = make(map[string]bool)
		canSkipUpdateApps["init"] = true
		canSkipUpdateApps["version"] = true

		_, found := canSkipUpdateApps[cmd.Name()]
		if !found {
			utils.UpdateAppsCacheIfStale()
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(replaceKubeconfigEnvIfFlagIsPresent)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&kubeCfgFile, "kubeconfig", "", "kubeconfig file")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func replaceKubeconfigEnvIfFlagIsPresent() {
	if kubeCfgFile != "" {
		os.Setenv("KUBECONFIG", kubeCfgFile)
	}
}
