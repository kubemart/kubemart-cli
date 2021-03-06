// ========================================================================
// Notes:
// * The tests here run in waterfall fashion like an integration test run.
// * For example, running "install" without first running "init" will fail.
// * So, be sure you execute them in the right order and clean-up unused
//   resources so they won't interfere other tests.
// ========================================================================

package cmd

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/kubemart/kubemart-cli/test"
)

func TestDestroyPrompt(t *testing.T) {
	actual, _ := test.RecordStdOutStdErr(func() {
		rootCmd.SetArgs([]string{
			"destroy",
		})
		rootCmd.ExecuteC()
	})

	expected := "Are you sure want to delete ALL apps and completely remove"
	if !strings.Contains(actual, expected) {
		t.Errorf("Expecting output to contain %s but got %s", expected, actual)
	}

}

func TestDestroyBeforeInstall(t *testing.T) {
	actual, _ := test.RecordStdOutStdErr(func() {
		rootCmd.SetArgs([]string{
			"destroy",
			"--yes",
		})
		rootCmd.Execute()
	})

	expected := "All done"
	if !strings.Contains(actual, expected) {
		t.Errorf("Expecting output to contain %s but got %s", expected, actual)
	}
}

func TestInitWithoutEmail(t *testing.T) {
	_, actual := test.RecordStdOutStdErr(func() {
		rootCmd.SetArgs([]string{
			"init",
		})
		rootCmd.Execute()
	})

	expected := "Error: required flag(s) \"email\" not set"
	if !strings.Contains(actual, expected) {
		t.Errorf("Expecting an error when init but got %s", actual)
	}
}

func TestInitWithEmail(t *testing.T) {
	if test.HasNamespaceGone("kubemart-system") {
		actual, _ := test.RecordStdOutStdErr(func() {
			rootCmd.SetArgs([]string{
				"init",
				"--email",
				"test@example.com",
			})
			rootCmd.Execute()
		})

		expected := "You are good to go"
		if !strings.Contains(actual, expected) {
			t.Errorf("Expecting output to contain %s but got %s", expected, actual)
		}
	} else {
		t.Errorf("kubemart-system namespace termination hung")
	}
}

func TestDestroyAfterInstall(t *testing.T) {
	actual, _ := test.RecordStdOutStdErr(func() {
		rootCmd.SetArgs([]string{
			"destroy",
			"--yes",
		})
		rootCmd.Execute()
	})

	expected := "All done"
	if !strings.Contains(actual, expected) {
		t.Errorf("Expecting output to contain %s but got %s", expected, actual)
	}
}

func TestInitWithEmailAndDomain(t *testing.T) {
	if test.HasNamespaceGone("kubemart-system") {
		actual, _ := test.RecordStdOutStdErr(func() {
			rootCmd.SetArgs([]string{
				"init",
				"--email",
				"test@example.com",
				"--domain-name",
				"example.com",
			})
			rootCmd.Execute()
		})

		expected := "You are good to go"
		if !strings.Contains(actual, expected) {
			t.Errorf("Expecting output to contain %s but got %s", expected, actual)
		}
	} else {
		t.Errorf("kubemart-system namespace termination hung")
	}
}

func TestInstall(t *testing.T) {
	actual, _ := test.RecordStdOutStdErr(func() {
		rootCmd.SetArgs([]string{
			"install",
			"rabbitmq",
		})
		rootCmd.Execute()
	})

	expected := "App(s) created successfully: rabbitmq"
	if !strings.Contains(actual, expected) {
		t.Errorf("Expecting output to contain %s but got %s", expected, actual)
	}
}

func TestInstalled(t *testing.T) {
	actual, _ := test.RecordStdOutStdErr(func() {
		rootCmd.SetArgs([]string{
			"installed",
		})
		rootCmd.Execute()
	})

	expected := "rabbitmq"
	if !strings.Contains(actual, expected) {
		t.Errorf("Expecting output to contain %s but got %s", expected, actual)
	}
}

func TestList(t *testing.T) {
	actual, _ := test.RecordStdOutStdErr(func() {
		rootCmd.SetArgs([]string{
			"list",
		})
		rootCmd.Execute()
	})

	expected := "longhorn"
	if !strings.Contains(actual, expected) {
		t.Errorf("Expecting output to contain %s but got %s", expected, actual)
	}
}

func TestSystemUpgrade(t *testing.T) {
	actual, _ := test.RecordStdOutStdErr(func() {
		rootCmd.SetArgs([]string{
			"system-upgrade",
		})
		rootCmd.Execute()
	})

	expected := "Upgrade complete successfully"
	if !strings.Contains(actual, expected) {
		t.Errorf("Expecting output to contain %s but got %s", expected, actual)
	}
}

func TestUninstall(t *testing.T) {
	actual, _ := test.RecordStdOutStdErr(func() {
		rootCmd.SetArgs([]string{
			"uninstall",
			"rabbitmq",
		})
		rootCmd.Execute()
	})

	expected := "App(s) now scheduled for deletion: rabbitmq"
	if !strings.Contains(actual, expected) {
		t.Errorf("Expecting output to contain %s but got %s", expected, actual)
	}
}

func TestUpdate(t *testing.T) {
	appName := "rabbitmq"

	// wait for uninstall (previous test) to complete
	canInstall := false
	current := 0
	maxRetries := 40
	for {
		current++
		if current > maxRetries {
			break
		}

		cs, err := NewClientFromLocalKubeConfig()
		if err != nil {
			t.Error(err)
		}

		_, err = cs.GetApp(appName)
		if err != nil {
			canInstall = true
			break
		}

		fmt.Printf("Waiting for %s app to get completely deleted...\n", appName)
		time.Sleep(3 * time.Second)
	}

	if canInstall {
		// install again because we run uninstall in previous test
		_, _ = test.RecordStdOutStdErr(func() {
			rootCmd.SetArgs([]string{
				"install",
				appName,
			})
			rootCmd.Execute()
		})

		// update the app
		_, actual := test.RecordStdOutStdErr(func() {
			rootCmd.SetArgs([]string{
				"update",
				appName,
			})
			rootCmd.Execute()
		})

		expected := "no new update available for this app"
		if !strings.Contains(actual, expected) {
			t.Errorf("Expecting output to contain %s but got %s", expected, actual)
		}
	} else {
		t.Errorf("Timed out waiting for %s app to get completely deleted", appName)
	}
}

func TestVersion(t *testing.T) {
	out, _ := test.RecordStdOutStdErr(func() {
		rootCmd.SetArgs([]string{
			"version",
			"--quiet",
		})
		rootCmd.Execute()
	})

	actual := strings.Trim(out, "\r\n")
	expected := fmt.Sprintf("v%s", VersionCli)
	if expected != actual {
		t.Errorf("Expecting %s but got %s", expected, actual)
	}
}

func TestVersionVerbose(t *testing.T) {
	actual, _ := test.RecordStdOutStdErr(func() {
		rootCmd.SetArgs([]string{
			"version",
			"--verbose",
		})
		rootCmd.Execute()
	})

	expected := "App CRD status: created"
	if !strings.Contains(actual, expected) {
		t.Errorf("Expecting output to contain %s but got %s", expected, actual)
	}
}

func TestInstallAppWithPlan(t *testing.T) {
	appName := "mariadb"
	plan := "10GB"
	appWithPlan := fmt.Sprintf("%v:%v", appName, plan)

	actual, _ := test.RecordStdOutStdErr(func() {
		rootCmd.SetArgs([]string{
			"install",
			appWithPlan,
		})
		rootCmd.Execute()
	})

	expected := "App(s) created successfully: mariadb"
	if !strings.Contains(actual, expected) {
		t.Errorf("Expecting output to contain %s but got %s", expected, actual)
	}

	cs, err := NewClientFromLocalKubeConfig()
	if err != nil {
		t.Error(err)
	}

	app, _ := cs.GetApp(appName)
	actualPlan := app.Spec.Plan
	expectedPlan := "10Gi"
	if expectedPlan != actualPlan {
		t.Errorf("Expecting %s plan but got %s", expectedPlan, actualPlan)
	}
}
