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
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/kubemart/kubemart-cli/test"
)

func TestDestroyPrompt(t *testing.T) {
	actual, _ := test.RecordStdOutStdErr(func() {
		KubemartRootCmd.SetArgs([]string{
			"destroy",
		})
		KubemartRootCmd.ExecuteC()
	})

	expected := "Are you sure want to delete ALL apps and completely remove"
	if !strings.Contains(actual, expected) {
		t.Errorf("Expecting output to contain %s but got %s", expected, actual)
	}

}

func TestDestroyBeforeInstall(t *testing.T) {
	actual, _ := test.RecordStdOutStdErr(func() {
		KubemartRootCmd.SetArgs([]string{
			"destroy",
			"--yes",
		})
		KubemartRootCmd.Execute()
	})

	expected := "All done"
	if !strings.Contains(actual, expected) {
		t.Errorf("Expecting output to contain %s but got %s", expected, actual)
	}
}

func TestInitWithoutEmail(t *testing.T) {
	_, actual := test.RecordStdOutStdErr(func() {
		KubemartRootCmd.SetArgs([]string{
			"init",
		})
		KubemartRootCmd.Execute()
	})

	expected := "Error: required flag(s) \"email\" not set"
	if !strings.Contains(actual, expected) {
		t.Errorf("Expecting an error when init but got %s", actual)
	}
}

func TestInitWithEmail(t *testing.T) {
	if test.HasNamespaceGone("kubemart-system") {
		actual, _ := test.RecordStdOutStdErr(func() {
			KubemartRootCmd.SetArgs([]string{
				"init",
				"--email",
				"test@example.com",
			})
			KubemartRootCmd.Execute()
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
		KubemartRootCmd.SetArgs([]string{
			"destroy",
			"--yes",
		})
		KubemartRootCmd.Execute()
	})

	expected := "All done"
	if !strings.Contains(actual, expected) {
		t.Errorf("Expecting output to contain %s but got %s", expected, actual)
	}
}

func TestInitWithEmailAndDomain(t *testing.T) {
	if test.HasNamespaceGone("kubemart-system") {
		actual, _ := test.RecordStdOutStdErr(func() {
			KubemartRootCmd.SetArgs([]string{
				"init",
				"--email",
				"test@example.com",
				"--domain-name",
				"example.com",
			})
			KubemartRootCmd.Execute()
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
		KubemartRootCmd.SetArgs([]string{
			"install",
			"rabbitmq",
		})
		KubemartRootCmd.Execute()
	})

	expected1 := "App created successfully"
	if !strings.Contains(actual, expected1) {
		t.Errorf("Expecting output to contain %s but got %s", expected1, actual)
	}

	expected2 := "App post-install notes:"
	if !strings.Contains(actual, expected2) {
		t.Errorf("Expecting output to contain %s but got %s", expected2, actual)
	}
}

func TestInstalled(t *testing.T) {
	actual, _ := test.RecordStdOutStdErr(func() {
		KubemartRootCmd.SetArgs([]string{
			"installed",
		})
		KubemartRootCmd.Execute()
	})

	expected := "rabbitmq"
	if !strings.Contains(actual, expected) {
		t.Errorf("Expecting output to contain %s but got %s", expected, actual)
	}
}

func TestList(t *testing.T) {
	actual, _ := test.RecordStdOutStdErr(func() {
		KubemartRootCmd.SetArgs([]string{
			"list",
		})
		KubemartRootCmd.Execute()
	})

	expected := "longhorn"
	if !strings.Contains(actual, expected) {
		t.Errorf("Expecting output to contain %s but got %s", expected, actual)
	}
}

func TestSystemUpgrade(t *testing.T) {
	actual, _ := test.RecordStdOutStdErr(func() {
		KubemartRootCmd.SetArgs([]string{
			"system-upgrade",
		})
		KubemartRootCmd.Execute()
	})

	expected := "Upgrade complete successfully"
	if !strings.Contains(actual, expected) {
		t.Errorf("Expecting output to contain %s but got %s", expected, actual)
	}
}

func TestUninstall(t *testing.T) {
	actual, _ := test.RecordStdOutStdErr(func() {
		KubemartRootCmd.SetArgs([]string{
			"uninstall",
			"rabbitmq",
		})
		KubemartRootCmd.Execute()
	})

	expected := "rabbitmq app is now scheduled to be deleted"
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

		_, err := GetApp(appName)
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
			KubemartRootCmd.SetArgs([]string{
				"install",
				appName,
			})
			KubemartRootCmd.Execute()
		})

		// update the app
		_, actual := test.RecordStdOutStdErr(func() {
			KubemartRootCmd.SetArgs([]string{
				"update",
				appName,
			})
			KubemartRootCmd.Execute()
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
		KubemartRootCmd.SetArgs([]string{
			"version",
			"--quiet",
		})
		KubemartRootCmd.Execute()
	})

	actual := strings.Trim(out, "\r\n")
	expected := fmt.Sprintf("v%s", VersionCli)
	if expected != actual {
		t.Errorf("Expecting %s but got %s", expected, actual)
	}
}

func TestVersionVerbose(t *testing.T) {
	actual, _ := test.RecordStdOutStdErr(func() {
		KubemartRootCmd.SetArgs([]string{
			"version",
			"--verbose",
		})
		KubemartRootCmd.Execute()
	})

	expected := "App CRD status: created"
	if !strings.Contains(actual, expected) {
		t.Errorf("Expecting output to contain %s but got %s", expected, actual)
	}
}

func TestInstallAppWithPlan(t *testing.T) {
	appName := "mariadb"
	plan := 10

	actual, _ := test.RecordStdOutStdErr(func() {
		KubemartRootCmd.SetArgs([]string{
			"install",
			appName,
			"--plan",
			strconv.Itoa(plan),
			"--quiet",
		})
		KubemartRootCmd.Execute()
	})

	expected := "App created successfully"
	if !strings.Contains(actual, expected) {
		t.Errorf("Expecting output to contain %s but got %s", expected, actual)
	}

	app, _ := GetApp(appName)
	actualPlan := app.Spec.Plan
	expectedPlan := plan
	if expectedPlan != actualPlan {
		t.Errorf("Expecting %d Gi plan but got %d Gi", expectedPlan, actualPlan)
	}
}
