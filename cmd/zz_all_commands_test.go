// ========================================================================
// Notes:
// * The tests here run in waterfall fashion like an integration test run.
// * For example, running "install" without first running "init" will fail.
// * So, be sure you execute them in the right order and clean-up unused
//   resources so they won't interfere other tests.
// ========================================================================

package cmd

import (
	"strings"
	"testing"

	"github.com/kubemart/kubemart/test"
)

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
		t.Errorf("Expecting %s but got %s", expected, actual)
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
	if test.CanProceedWithInit() {
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
			t.Errorf("Expecting %s but got %s", expected, actual)
		}
	} else {
		t.Errorf("kubemart-system namespace termination stucked")
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
		t.Errorf("Expecting %s but got %s", expected, actual)
	}
}

func TestInitWithEmailAndDomain(t *testing.T) {
	if test.CanProceedWithInit() {
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
			t.Errorf("Expecting %s but got %s", expected, actual)
		}
	} else {
		t.Errorf("kubemart-system namespace termination stucked")
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

	expected := "App created successfully"
	if !strings.Contains(actual, expected) {
		t.Errorf("Expecting %s but got %s", expected, actual)
	}
}
