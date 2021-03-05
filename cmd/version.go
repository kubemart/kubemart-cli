package cmd

import (
	"fmt"
	"runtime"

	"github.com/kubemart/kubemart/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	quiet   bool
	verbose bool

	// VersionCli is set from outside using ldflags
	VersionCli = "0.0.0"

	// CommitCli is set from outside using ldflags
	CommitCli = "none"

	// DateCli is set from outside using ldflags
	DateCli = "unknown"

	versionCmd = &cobra.Command{
		Use:     "version",
		Example: "kubemart version",
		Short:   "Output the current build information",
		Run: func(cmd *cobra.Command, args []string) {

			// TODO - uncomment this after we go live
			// githubTag := &latest.GithubTag{
			// 	Owner:             "kubemart",
			// 	Repository:        "kubemart-cli",
			// 	FixVersionStrFunc: latest.DeleteFrontV(),
			// }
			switch {
			case verbose:
				fmt.Printf("Client version: v%s\n", VersionCli)
				fmt.Printf("Go version (client): %s\n", runtime.Version())
				fmt.Printf("Build date (client): %s\n", DateCli)
				fmt.Printf("Git commit (client): %s\n", CommitCli)
				fmt.Printf("OS/Arch (client): %s/%s\n", runtime.GOOS, runtime.GOARCH)
				fmt.Println("---")

				serverVersion, _ := utils.GetKubeServerVersionHuman()
				fmt.Printf("Kubernetes version: %s\n", serverVersion)

				operatorVersion, _ := utils.GetInstalledOperatorVersion()
				fmt.Printf("Operator version: %s\n", operatorVersion)

				isAppCRDInstalled := "not installed"
				isJobWatcherCRDInstalled := "not installed"
				if utils.IsCRDExist("apps.kubemart.civo.com") {
					isAppCRDInstalled = "installed"
				}
				if utils.IsCRDExist("jobwatchers.kubemart.civo.com") {
					isJobWatcherCRDInstalled = "installed"
				}
				fmt.Printf("App CRD status: %s\n", isAppCRDInstalled)
				fmt.Printf("JobWatcher CRD status: %s\n", isJobWatcherCRDInstalled)

				namespaceStatus := "not created"
				isNamespaceExist, _ := utils.IsNamespaceExist("kubemart-system")
				if isNamespaceExist {
					namespaceStatus = "created"
				}
				fmt.Printf("Namespace (kubemart-system) status: %s\n", namespaceStatus)

				// TODO - uncomment this after we go live
				// res, err := latest.Check(githubTag, strings.Replace(VersionCli, "v", "", 1))
				// if err != nil {
				// 	utility.Error("Checking for a newer version failed with %s", err)
				// 	os.Exit(1)
				// }

				// if res.Outdated {
				// 	utility.RedConfirm("A newer version (v%s) is available, please upgrade\n", res.Current)
				// }
			case quiet:
				fmt.Printf("v%s\n", VersionCli)
			default:
				fmt.Printf("v%s\n", VersionCli)

				// TODO - uncomment this after we go live
				// res, err := latest.Check(githubTag, strings.Replace(VersionCli, "v", "", 1))
				// if err != nil {
				// 	utility.Error("Checking for a newer version failed with %s", err)
				// 	os.Exit(1)
				// }

				// if res.Outdated {
				// 	utility.RedConfirm("A newer version (v%s) is available, please upgrade\n", res.Current)
				// }
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "display simple output")
	versionCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "display full information")
}
