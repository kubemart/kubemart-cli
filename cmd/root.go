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

	utils "github.com/kubemart/kubemart-cli/pkg/utils"
	"github.com/spf13/cobra"
)

var kubeCfgFile string
var debug bool
var canSkipUpdateApps map[string]bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "kubemart",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		canSkipUpdateApps = make(map[string]bool)
		canSkipUpdateApps["destroy"] = true
		canSkipUpdateApps["help"] = true
		canSkipUpdateApps["init"] = true
		canSkipUpdateApps["system-upgrade"] = true
		canSkipUpdateApps["version"] = true

		_, found := canSkipUpdateApps[cmd.Name()]
		if !found {
			ok, err := utils.UpdateAppsCacheIfStale()
			if !ok {
				fmt.Printf("%v\n", err)
				os.Exit(1)
			}
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVarP(&kubeCfgFile, "kubeconfig", "k", "", "kubeconfig file")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "print verbose logs when running command")
	rootCmd.SetHelpCommand(&cobra.Command{Use: "no-help", Hidden: true}) // disable "kubemart help <command>"

	// https://github.com/spf13/cobra/issues/340
	rootCmd.SilenceUsage = true

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// This `setLogLevelEnvIfFlagIsTrue` must be the first. Otherwise,
	// we won't see debug statement in other `OnInitialize` functions.
	cobra.OnInitialize(setLogLevelEnvIfFlagIsTrue)
	cobra.OnInitialize(replaceKubeconfigEnvIfFlagIsPresent)
}

// replaceKubeconfigEnvIfFlagIsPresent will set KUBECONFIG env variable
// when user use '--kubeconfig' or '-k' flag
func replaceKubeconfigEnvIfFlagIsPresent() {
	if kubeCfgFile != "" {
		kubeconfigEnvName := "KUBECONFIG"
		os.Setenv(kubeconfigEnvName, kubeCfgFile)
		utils.DebugPrintf("KUBECONFIG: %s\n", os.Getenv(kubeconfigEnvName))
	}
}

// setLogLevelEnvIfFlagIsTrue will set LOGLEVEL env variable
// when user use '--debug' or '-d' flag
func setLogLevelEnvIfFlagIsTrue() {
	if debug {
		os.Setenv("LOGLEVEL", "debug")
	}
}
