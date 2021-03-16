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
	"text/tabwriter"

	"github.com/spf13/cobra"
)

// installedCmd represents the installed command
var installedCmd = &cobra.Command{
	Use:     "installed",
	Example: "kubemart installed",
	Short:   "List all installed applications",
	RunE: func(cmd *cobra.Command, args []string) error {
		apps, err := ListApps()
		if err != nil {
			return err
		}

		haveSomething := false
		w := tabwriter.NewWriter(os.Stdout, 15, 0, 1, ' ', tabwriter.TabIndent)
		fmt.Fprintln(w, "NAME\tCURRENT STATUS\tVERSION\tUPDATE AVAILABLE")
		for _, app := range apps.Items {
			if !app.DeletionTimestamp.IsZero() {
				continue // skip deleted apps
			}

			currentStatus := fmt.Sprintf("\t%s", app.Status.LastStatus)
			version := fmt.Sprintf("\t%s", app.Status.InstalledVersion)

			updateAvailable := ""
			if app.Status.InstalledVersion != "" {
				if app.Status.NewUpdateAvailable {
					updateAvailable = fmt.Sprintf("yes (%s)", app.Status.NewUpdateVersion)
				} else {
					updateAvailable = "no"
				}
			}
			newUpdate := fmt.Sprintf("\t%s", updateAvailable)

			fmt.Fprintln(w, app.Name, currentStatus, version, newUpdate)
			haveSomething = true
		}

		if haveSomething {
			w.Flush()
		} else {
			fmt.Println("No resources found")
		}

		return nil
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
