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

// KubemartRootCmd represents the base command when called without any subcommands
var KubemartRootCmd = &cobra.Command{
	Use:   "kubemart",
	Short: "Manage apps in your Kubernetes cluster",
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
// This is called by main.main(). It only needs to happen once to the KubemartRootCmd.
func Execute() {
	if err := KubemartRootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	KubemartRootCmd.PersistentFlags().StringVarP(&kubeCfgFile, "kubeconfig", "k", "", "kubeconfig file")
	KubemartRootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "print verbose logs when running command")
	KubemartRootCmd.SetHelpCommand(&cobra.Command{Use: "no-help", Hidden: true}) // disable "kubemart help <command>"

	// https://github.com/spf13/cobra/issues/340
	KubemartRootCmd.SilenceUsage = true

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// KubemartRootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	cobra.OnInitialize(replaceKubeconfigEnvIfFlagIsPresent)
	cobra.OnInitialize(setLogLevelEnvIfFlagIsTrue)
}

// replaceKubeconfigEnvIfFlagIsPresent will set KUBECONFIG env variable
// when user use '--kubeconfig' or '-k' flag
func replaceKubeconfigEnvIfFlagIsPresent() {
	if kubeCfgFile != "" {
		os.Setenv("KUBECONFIG", kubeCfgFile)
	}
}

// setLogLevelEnvIfFlagIsTrue will set LOGLEVEL env variable
// when user use '--debug' or '-d' flag
func setLogLevelEnvIfFlagIsTrue() {
	if debug {
		os.Setenv("LOGLEVEL", "debug")
	}
}
