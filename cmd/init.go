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

// Email is used for BIZAAR:EMAIL_ADDRESS
var Email string

// DomainName is used for BIZAAR:DOMAIN_NAME
var DomainName string

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Setup local environment and install Bizaar operator",
	Run: func(cmd *cobra.Command, args []string) {
		masterIP, err := utils.GetMasterIP()
		if err != nil {
			fmt.Printf("Unable to determine master node IP address - %v\n", err)
			os.Exit(1)
		}

		if DomainName == "" {
			DomainName = fmt.Sprintf("%s.xip.io", masterIP)
		}

		clusterName, err := utils.GetClusterName()
		if err != nil {
			fmt.Printf("Unable to determine cluster name - %v\n", err)
			os.Exit(1)
		}

		bcm := &utils.BizaarConfigMap{
			EmailAddress: Email,
			DomainName:   DomainName,
			ClusterName:  clusterName,
			MasterIP:     masterIP,
		}
		utils.DebugPrintf("Bizaar ConfigMap: %+v\n", bcm)

		gitProgram := "git"
		if !utils.IsCommandAvailable(gitProgram) {
			fmt.Println(gitProgram, "program not found. Please install it first and retry.")
			os.Exit(1)
		}

		bizaarPaths, err := utils.GetBizaarPaths()
		if err != nil {
			fmt.Printf("Unable to load Bizaar paths - %v\n", err.Error())
			os.Exit(1)
		}

		bizaarDirPath := bizaarPaths.RootDirectoryPath
		appsDirPath := bizaarPaths.AppsDirectoryPath
		configFilePath := bizaarPaths.ConfigFilePath
		if _, err := os.Stat(bizaarDirPath); os.IsNotExist(err) {
			// When bizaarDir is not exist, create it (with apps folder and config.json file)

			// Create apps folder
			err = os.MkdirAll(appsDirPath, 0755)
			if err != nil {
				fmt.Println("Unable to create ~/.bizaar/apps directory ($ mkdir -p ~/.bizaar/apps)")
				os.Exit(1)
			}

			// Create config.json file
			_, err := os.Create(configFilePath)
			if err != nil {
				fmt.Println("Unable to create ~/.bizaar/config.json file")
				os.Exit(1)
			}

			// Clone
			cloneOutput, err := utils.GitClone(appsDirPath)
			if err != nil {
				fmt.Printf("Unable to clone marketplace - %v\n", err)
				os.Exit(1)
			}
			utils.DebugPrintf("Clone output: %s\n", cloneOutput)

			// Update timestamp
			err = utils.UpdateConfigFileLastUpdatedTimestamp()
			if err != nil {
				fmt.Printf("Unable to config file's timestamp field - %v\n", err)
				os.Exit(1)
			}
		}

		namespaceExists, err := utils.IsBizaarNamespaceExist()
		if err != nil {
			fmt.Printf("Unable to check namespace - %v\n", err.Error())
			os.Exit(1)
		}

		if !namespaceExists {
			err = utils.CreateBizaarNamespace()
			if err != nil {
				fmt.Printf("Unable to create namespace - %v\n", err.Error())
				os.Exit(1)
			}
		}

		configMapExists, err := utils.IsBizaarConfigMapExist()
		if err != nil {
			fmt.Printf("Unable to check ConfigMap - %v\n", err.Error())
			os.Exit(1)
		}

		if !configMapExists {
			err = utils.CreateBizaarConfigMap(bcm)
			if err != nil {
				fmt.Printf("Unable to create ConfigMap - %v\n", err.Error())
				os.Exit(1)
			}
		}

		fmt.Println("You are good to go")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVarP(&Email, "email", "e", "", "Email address (required)")
	initCmd.Flags().StringVarP(&DomainName, "domain-name", "d", "", "Domain name (will default to master_ip.xip.io if not supplied)")
	initCmd.MarkFlagRequired("email")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
