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
	"io/ioutil"
	"os"
	"strings"

	"github.com/kubemart/kubemart-cli/pkg/utils"
	"github.com/spf13/cobra"
)

// to store all folder names that we don't want e.g. "bin"
var excludeList map[string]bool

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "list",
	Example: "kubemart list",
	Short:   "List all the applications that can be installed",
	Long:    `This command will display the list of all the applications that can be installed onto the Kubernetes cluster`,
	RunE: func(cmd *cobra.Command, args []string) error {
		excludeList = make(map[string]bool)
		excludeList["bin"] = true
		dir, _ := utils.GetKubemartPaths()
		path := dir.AppsDirectoryPath
		files, err := ioutil.ReadDir(path)
		if err != nil {
			return fmt.Errorf("unable parse get list of files - %v", err)
		}
		apps := []string{}
		for _, file := range files {
			fileName := file.Name()
			filePath := fmt.Sprintf("%s/%s", path, fileName)
			fileInfo, err := os.Stat(filePath)
			if err != nil {
				return fmt.Errorf("unable to locate file - %v", err)
			}
			if fileInfo.IsDir() && isValid(fileName) {
				apps = append(apps, fileName)
			}
		}

		fmt.Println(strings.Join(apps, "\n"))
		return nil
	},
}

// isValid returns true if the folder is not a hidden folder
// and not part of excluded items e.g. 'bin' folder
func isValid(folderName string) bool {
	isHidden := isHidden(folderName)
	if isHidden {
		return false
	}

	isNotAllowed, ok := excludeList[folderName]
	if ok && isNotAllowed {
		return false
	}

	return true
}

// isHidden returns 'true' if the folder starts with '.' (dot)
func isHidden(folderName string) bool {
	firstCharacter := folderName[0:1]
	return firstCharacter == "."
}

func init() {
	KubemartRootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
