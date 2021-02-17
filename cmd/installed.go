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
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	operator "github.com/civo/bizaar-operator/api/v1alpha1"
	utils "github.com/civo/bizaar/pkg/utils"
	"github.com/spf13/cobra"
)

// installedCmd represents the installed command
var installedCmd = &cobra.Command{
	Use:   "installed",
	Short: "List all installed applications",
	Run: func(cmd *cobra.Command, args []string) {
		clientset, err := utils.GetKubeClientSet()
		if err != nil {
			fmt.Printf("Unable to create k8s clientset - %v", err)
			os.Exit(1)
		}

		// TODO - use bizaar-operator Go client
		statusCode := 0
		res := clientset.RESTClient().
			Get().
			AbsPath("/apis/bizaar.civo.com/v1alpha1").
			Namespace("default").
			Resource("apps").
			Do(context.Background())

		res = res.StatusCode(&statusCode)
		apps := &operator.AppList{}

		if statusCode == 200 {
			err = res.Into(apps)
			if err != nil {
				fmt.Printf("Unable parse app list - %v", err)
				os.Exit(1)
			}

			if len(apps.Items) == 0 {
				fmt.Println("No resources found")
				os.Exit(0)
			}

			w := tabwriter.NewWriter(os.Stdout, 15, 0, 1, ' ', tabwriter.TabIndent)
			fmt.Fprintln(w, "NAME\tVERSION\tCURRENT STATUS")
			for _, app := range apps.Items {
				version := fmt.Sprintf("\t%s", app.Status.InstalledVersion)
				currentStatus := fmt.Sprintf("\t%s", app.Status.LastStatus)
				fmt.Fprintln(w, app.Name, version, currentStatus)
			}
			w.Flush()
			os.Exit(0)
		} else {
			fmt.Printf("Unable to list apps - %v\n", res.Error())
			os.Exit(1)
		}

	},
}

func init() {
	rootCmd.AddCommand(installedCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// installedCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// installedCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
