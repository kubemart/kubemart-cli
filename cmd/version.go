package cmd

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/kubemart/kubemart-cli/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/tcnksm/go-latest"
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

			githubTag := &latest.GithubTag{
				Owner:             "kubemart",
				Repository:        "kubemart-cli",
				FixVersionStrFunc: latest.DeleteFrontV(),
			}
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

				isAppCRDInstalled := "not created"
				isJobWatcherCRDInstalled := "not created"
				if utils.IsCRDExist("apps.kubemart.civo.com") {
					isAppCRDInstalled = "created"
				}
				if utils.IsCRDExist("jobwatchers.kubemart.civo.com") {
					isJobWatcherCRDInstalled = "created"
				}
				fmt.Printf("App CRD status: %s\n", isAppCRDInstalled)
				fmt.Printf("JobWatcher CRD status: %s\n", isJobWatcherCRDInstalled)

				namespaceStatus := "not created"
				isNamespaceExist, _ := utils.IsNamespaceExist("kubemart-system")
				if isNamespaceExist {
					namespaceStatus = "created"
				}
				fmt.Printf("Namespace (kubemart-system) status: %s\n", namespaceStatus)

				configMapStatus := "not created"
				cmExists, _ := utils.IsKubemartConfigMapExist()
				if cmExists {
					configMapStatus = "created"
				}
				fmt.Printf("ConfigMap (kubemart-config) status: %s\n", configMapStatus)

				res, err := latest.Check(githubTag, strings.Replace(VersionCli, "v", "", 1))
				if err != nil {
					fmt.Printf("Checking for a newer version failed with %s\n", err)
					os.Exit(1)
				}

				if res.Outdated {
					fmt.Printf("\nFYI, a newer Kubemart CLI version (v%s) is available, please upgrade\n", res.Current)
				}
			case quiet:
				fmt.Printf("v%s\n", VersionCli)
			default:
				fmt.Printf("v%s\n", VersionCli)

				res, err := latest.Check(githubTag, strings.Replace(VersionCli, "v", "", 1))
				if err != nil {
					fmt.Printf("Checking for a newer version failed with %s\n", err)
					os.Exit(1)
				}

				if res.Outdated {
					fmt.Printf("\nFYI, a newer Kubemart CLI version (v%s) is available, please upgrade\n", res.Current)
				}
			}
		},
	}
)

func init() {
	KubemartRootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "display simple output")
	versionCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "display full information")
}
