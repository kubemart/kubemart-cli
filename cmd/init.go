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
	"strings"

	utils "github.com/kubemart/kubemart/pkg/utils"
	"github.com/spf13/cobra"
)

// Email is used for KUBEMART:EMAIL_ADDRESS
var Email string

// DomainName is used for KUBEMART:DOMAIN_NAME
var DomainName string

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:     "init",
	Example: "kubemart init --email your@email.com",
	Short:   "Setup local environment and install Kubemart operator",
	RunE: func(cmd *cobra.Command, args []string) error {
		masterIP, err := utils.GetMasterIP()
		if err != nil {
			return fmt.Errorf("unable to determine master node IP address - %v", err)
		}

		if DomainName == "" {
			DomainName = fmt.Sprintf("%s.xip.io", masterIP)
		}

		clusterName, err := utils.GetClusterName()
		if err != nil {
			return fmt.Errorf("unable to determine cluster name - %v", err)
		}

		bcm := &utils.KubemartConfigMap{
			EmailAddress: Email,
			DomainName:   DomainName,
			ClusterName:  clusterName,
			MasterIP:     masterIP,
		}
		utils.DebugPrintf("Kubemart ConfigMap: %+v\n", bcm)

		gitProgram := "git"
		if !utils.IsCommandAvailable(gitProgram) {
			return fmt.Errorf("%s program not found - please install it first and retry", gitProgram)
		}

		kubemartPaths, err := utils.GetKubemartPaths()
		if err != nil {
			return fmt.Errorf("unable to load Kubemart paths - %v", err.Error())
		}

		kubemartDirPath := kubemartPaths.RootDirectoryPath
		appsDirPath := kubemartPaths.AppsDirectoryPath
		configFilePath := kubemartPaths.ConfigFilePath
		if _, err := os.Stat(kubemartDirPath); os.IsNotExist(err) {
			// When kubemartDir is not exist, create it (with apps folder and config.json file)
			fmt.Println("Fetching apps...")

			// Create apps folder
			err = os.MkdirAll(appsDirPath, 0755)
			if err != nil {
				return fmt.Errorf("unable to create ~/.kubemart/apps directory ($ mkdir -p ~/.kubemart/apps)")
			}

			// Create config.json file
			_, err := os.Create(configFilePath)
			if err != nil {
				return fmt.Errorf("unable to create ~/.kubemart/config.json file")
			}

			// Clone
			cloneOutput, err := utils.GitClone(appsDirPath)
			if err != nil {
				return fmt.Errorf("unable to clone marketplace - %v", err)
			}
			utils.DebugPrintf("Clone output: %s\n", cloneOutput)

			// Update timestamp
			err = utils.UpdateConfigFileLastUpdatedTimestamp()
			if err != nil {
				return fmt.Errorf("unable to config file's timestamp field - %v", err)
			}
		}

		namespaceExists, err := utils.IsNamespaceExist("kubemart-system")
		if err != nil {
			return fmt.Errorf("unable to check namespace - %v", err.Error())
		}

		if !namespaceExists {
			fmt.Println("Creating Namespace (for operator)...")
			err = utils.CreateKubemartNamespace()
			if err != nil {
				return fmt.Errorf("unable to create namespace - %v", err.Error())
			}
		}

		configMapExists, err := utils.IsKubemartConfigMapExist()
		if err != nil {
			return fmt.Errorf("unable to check ConfigMap - %v", err.Error())
		}

		if !configMapExists {
			fmt.Println("Creating ConfigMap (for operator)...")
			err = utils.CreateKubemartConfigMap(bcm)
			if err != nil {
				return fmt.Errorf("unable to create ConfigMap - %v", err.Error())
			}
		}

		fmt.Println("Applying manifests...")
		operatorYAML, err := utils.GetLatestManifests()
		if err != nil {
			return fmt.Errorf("unable to download latest manifests - %v", err.Error())
		}

		manifests := strings.Split(operatorYAML, "---")
		err = utils.ApplyManifests(manifests)
		if err != nil {
			return fmt.Errorf("unable to apply manifest - %v", err.Error())
		}

		fmt.Println("You are good to go")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVarP(&Email, "email", "e", "", "email address (required)")
	initCmd.Flags().StringVarP(&DomainName, "domain-name", "n", "", "domain name (will default to master_ip.xip.io if not supplied)")
	initCmd.MarkFlagRequired("email")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
