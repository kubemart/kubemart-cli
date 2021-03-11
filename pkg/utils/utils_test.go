package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetKubemartPaths1(t *testing.T) {
	homeDir, _ := os.UserHomeDir()
	kp, _ := GetKubemartPaths()

	expectedRootDirectoryPath := fmt.Sprintf("%s/.kubemart", homeDir)
	actualRootDirectoryPath := kp.RootDirectoryPath
	if expectedRootDirectoryPath != actualRootDirectoryPath {
		t.Errorf("Expected %s but actual is %s", expectedRootDirectoryPath, actualRootDirectoryPath)
	}
}

func TestGetKubemartPaths2(t *testing.T) {
	homeDir, _ := os.UserHomeDir()
	kp, _ := GetKubemartPaths()

	expectedAppsDirectoryPath := fmt.Sprintf("%s/.kubemart/apps", homeDir)
	actualAppsDirectoryPath := kp.AppsDirectoryPath
	if expectedAppsDirectoryPath != actualAppsDirectoryPath {
		t.Errorf("Expected %s but actual is %s", expectedAppsDirectoryPath, actualAppsDirectoryPath)
	}
}

func TestGetKubemartPaths3(t *testing.T) {
	homeDir, _ := os.UserHomeDir()
	kp, _ := GetKubemartPaths()

	expectedConfigFilePath := fmt.Sprintf("%s/.kubemart/config.json", homeDir)
	actualConfigFilePath := kp.ConfigFilePath
	if expectedConfigFilePath != actualConfigFilePath {
		t.Errorf("Expected %s but actual is %s", expectedConfigFilePath, actualConfigFilePath)
	}
}

// --------------------------------------------------

func TestGetNodesFromKubeconfigVariableSingle(t *testing.T) {
	// define KUBECONFIG and context
	homeDir, _ := os.UserHomeDir()
	kubeConfigPath := fmt.Sprintf("%s/kind/cluster-1.yaml", homeDir)
	contextName := "kind-cluster-1"

	// set KUBECONFIG and context
	os.Setenv("KUBECONFIG", kubeConfigPath)
	cmd := exec.Command("kubectl", "config", "use-context", contextName)
	_ = cmd.Run()

	cs, err := GetKubeClientSet()
	if err != nil {
		t.Error(err)
	}

	nc := cs.CoreV1().Nodes()
	nodes, err := nc.List(context.Background(), v1.ListOptions{})
	if err != nil {
		t.Error(err)
	}

	var actual string
	for _, node := range nodes.Items {
		actual = node.Name
		break
	}

	expected := "cluster-1-control-plane"
	if expected != actual {
		t.Errorf("Expected %s but actual is %s", expected, actual)
	}

	// cleanup
	os.Unsetenv("KUBECONFIG")
	cmd = exec.Command("kubectl", "config", "use-context", "kind-default")
	_ = cmd.Run()
}

func TestGetNodesFromKubeconfigVariableMultiples(t *testing.T) {
	// Simulate situation when user have multiple contexts/clusters.
	// For testing purpose, use absolute path to kubeconfig.
	// In real live, user's terminal will expand the "~" or "$HOME".
	homeDir, _ := os.UserHomeDir()
	kubeConfigPath1 := fmt.Sprintf("%s/kind/cluster-1.yaml", homeDir)
	fmt.Println("Kubeconfig path:", kubeConfigPath1)

	kubeConfigPath2 := fmt.Sprintf("%s/kind/cluster-2.yaml", homeDir)
	fmt.Println("Kubeconfig path:", kubeConfigPath2)

	kubeConfigPaths := fmt.Sprintf("%s:%s", kubeConfigPath1, kubeConfigPath2)
	os.Setenv("KUBECONFIG", kubeConfigPaths)
	fmt.Println("KUBECONFIG:", os.Getenv("KUBECONFIG"))

	// switch context to cluster-2
	out, err := exec.Command("kubectl", "config", "use-context", "kind-cluster-2").Output()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(string(out))

	cs, err := GetKubeClientSet()
	if err != nil {
		t.Error(err)
	}

	nc := cs.CoreV1().Nodes()
	nodes, err := nc.List(context.Background(), v1.ListOptions{})
	if err != nil {
		t.Error(err)
	}

	var actual string
	for _, node := range nodes.Items {
		actual = node.Name
		break
	}

	expected := "cluster-2-control-plane"
	if expected != actual {
		t.Errorf("Expected %s but actual is %s", expected, actual)
	}

	// cleanup
	os.Unsetenv("KUBECONFIG")
	cmd := exec.Command("kubectl", "config", "use-context", "kind-default")
	_ = cmd.Run()
}

func TestGetNodesFromKubeconfigFlag(t *testing.T) {
	// Simulate situation when user run CLI with "--kubeconfig" flag.
	// For testing purpose, use absolute path to kubeconfig.
	// In real live, user's terminal will expand the "~" or "$HOME".

	// define KUBECONFIG and context
	homeDir, _ := os.UserHomeDir()
	kubeConfigPath := fmt.Sprintf("%s/kind/cluster-2.yaml", homeDir)
	contextName := "kind-cluster-2"

	// set KUBECONFIG and context
	os.Setenv("KUBECONFIG", kubeConfigPath)
	cmd := exec.Command("kubectl", "config", "use-context", contextName)
	_ = cmd.Run()

	cs, err := GetKubeClientSet()
	if err != nil {
		t.Error(err)
	}

	nc := cs.CoreV1().Nodes()
	nodes, err := nc.List(context.Background(), v1.ListOptions{})
	if err != nil {
		t.Error(err)
	}

	var actual string
	for _, node := range nodes.Items {
		actual = node.Name
		break
	}

	expected := "cluster-2-control-plane"
	if expected != actual {
		t.Errorf("Expected %s but actual is %s", expected, actual)
	}

	// cleanup
	os.Unsetenv("KUBECONFIG")
	cmd = exec.Command("kubectl", "config", "use-context", "kind-default")
	_ = cmd.Run()
}

func TestGetNodesFromDefaultKubeconfig(t *testing.T) {
	kc := os.Getenv("KUBECONFIG")
	fmt.Println("KUBECONFIG:", kc) // if this prints "KUBECONFIG:" (w/o filepath), that's expected

	cs, err := GetKubeClientSet()
	if err != nil {
		t.Error(err)
	}

	nc := cs.CoreV1().Nodes()
	nodes, err := nc.List(context.Background(), v1.ListOptions{})
	if err != nil {
		t.Error(err)
	}

	var actual string
	for _, node := range nodes.Items {
		actual = node.Name
		break
	}

	expected := "default-control-plane"
	if expected != actual {
		t.Errorf("Expected %s but actual is %s", expected, actual)
	}
}

// --------------------------------------------------

func TestNotAvailableProgram(t *testing.T) {
	programName := "gitx"
	result := IsCommandAvailable(programName)
	fmt.Printf("%s program found? %t\n", programName, result)
}

func TestAvailableProgram(t *testing.T) {
	programName := "git"
	result := IsCommandAvailable(programName)
	fmt.Printf("%s program found? %t\n", programName, result)
}

// --------------------------------------------------

func TestGitClone(t *testing.T) {
	targetDir := "/tmp/test"
	output, _ := GitClone(targetDir)
	fmt.Printf("Clone output: %s\n", output)
}

func TestGitPull(t *testing.T) {
	targetDir := "/tmp/test"

	hashBefore, _ := GitLatestCommitHash(targetDir)
	fmt.Printf("Before reset commit: %s\n", hashBefore)

	cmdToRun := fmt.Sprintf("git -C %s reset --hard HEAD~1", targetDir)
	cmd := exec.Command("/bin/sh", "-c", cmdToRun)
	if err := cmd.Run(); err != nil {
		t.Error(err)
	}

	hashAfter, _ := GitLatestCommitHash(targetDir)
	fmt.Printf("After reset commit: %s\n", hashAfter)

	output, _ := GitPull(targetDir)
	fmt.Printf("Pull output: %s\n", output)

	hashAfterPull, _ := GitLatestCommitHash(targetDir)
	fmt.Printf("After pull commit: %s\n", hashAfterPull)
}

// --------------------------------------------------

func TestGetContextAndClusterFromKubeconfigFlag(t *testing.T) {
	// Simulate situation when user run CLI with "--kubeconfig" flag.
	// For testing purpose, use absolute path to kubeconfig.
	// In real live, user's terminal will expand the "~" or "$HOME".
	homeDir, _ := os.UserHomeDir()
	kubeconfigPath := fmt.Sprintf("%s/kind/cluster-2.yaml", homeDir)
	fmt.Println("Kubeconfig path:", kubeconfigPath)
	os.Setenv("KUBECONFIG", kubeconfigPath)

	currentContext, err := GetCurrentContext()
	if err != nil {
		t.Error(err)
	}
	fmt.Println("Current context:", currentContext)

	clusterName, err := GetClusterName()
	if err != nil {
		t.Error(err)
	}
	fmt.Println("Cluster name:", clusterName)

	// 	Output:
	// 	Current context: k3d-cluster-2-context
	// Cluster name: k3d-cluster-2-cluster
}

func TestGetContextAndClusterFromKubeconfigEnv(t *testing.T) {
	// Simulate situation when user have multiple contexts/clusters.
	// For testing purpose, use absolute path to kubeconfig.
	// In real live, user's terminal will expand the "~" or "$HOME".
	homeDir, _ := os.UserHomeDir()
	kubeConfigPath1 := fmt.Sprintf("%s/kind/cluster-1.yaml", homeDir)
	fmt.Println("Kubeconfig path:", kubeConfigPath1)

	kubeConfigPath2 := fmt.Sprintf("%s/kind/cluster-2.yaml", homeDir)
	fmt.Println("Kubeconfig path:", kubeConfigPath2)

	kubeConfigPaths := fmt.Sprintf("%s:%s", kubeConfigPath1, kubeConfigPath2)
	os.Setenv("KUBECONFIG", kubeConfigPaths)
	fmt.Println("KUBECONFIG:", os.Getenv("KUBECONFIG"))

	// switch context to cluster-2
	out, err := exec.Command("kubectl", "config", "use-context", "kind-cluster-2").Output()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(string(out))

	currentContext, err := GetCurrentContext()
	if err != nil {
		t.Error(err)
	}
	fmt.Println("Current context:", currentContext)

	clusterName, err := GetClusterName()
	if err != nil {
		t.Error(err)
	}
	fmt.Println("Cluster name:", clusterName)

	// 	Output:
	// 	Current context: k3d-cluster-2-context
	// Cluster name: k3d-cluster-2-cluster
}

// --------------------------------------------------

func TestExtractIPAddressFromURL1(t *testing.T) {
	expectedErr := "IP address is empty"
	inputURL := "https://www.oreilly.com/library/view/regular-expressions-cookbook/9780596802837/ch07s16.html"
	_, err := ExtractIPAddressFromURL(inputURL)
	if err.Error() != expectedErr {
		t.Errorf("Actual error = %v, expected error = %v.", err, expectedErr)
	}
}

func TestExtractIPAddressFromURL2(t *testing.T) {
	expectedErr := "IP address is empty"
	inputURL := "https://www.oreilly.com/"
	_, err := ExtractIPAddressFromURL(inputURL)
	if err.Error() != expectedErr {
		t.Errorf("Actual error = %v, expected error = %v.", err, expectedErr)
	}
}

func TestExtractIPAddressFromURL3(t *testing.T) {
	expectedErr := "IP address is empty"
	inputURL := "https://www.oreilly.com:8080"
	_, err := ExtractIPAddressFromURL(inputURL)
	if err.Error() != expectedErr {
		t.Errorf("Actual error = %v, expected error = %v.", err, expectedErr)
	}
}

func TestExtractIPAddressFromURL4(t *testing.T) {
	expectedErr := "IP address is empty"
	inputURL := "www.oreilly.com"
	_, err := ExtractIPAddressFromURL(inputURL)
	if err.Error() != expectedErr {
		t.Errorf("Actual error = %v, expected error = %v.", err, expectedErr)
	}
}

func TestExtractIPAddressFromURL5(t *testing.T) {
	expectedErr := "IP address is empty"
	inputURL := "www.oreilly.com:8080"
	_, err := ExtractIPAddressFromURL(inputURL)
	if err.Error() != expectedErr {
		t.Errorf("Actual error = %v, expected error = %v.", err, expectedErr)
	}
}

func TestExtractIPAddressFromURL6(t *testing.T) {
	inputURL := "https://192.168.99.109:12345"
	ip, err := ExtractIPAddressFromURL(inputURL)
	if err != nil {
		t.Errorf("Actual error = %v, expected error = %v.", err, nil)
	}
	fmt.Println(ip) // should print "192.168.99.109"
}

func TestExtractIPAddressFromURL7(t *testing.T) {
	inputURL := "https://192.168.99.109"
	ip, err := ExtractIPAddressFromURL(inputURL)
	if err != nil {
		t.Errorf("Actual error = %v, expected error = %v.", err, nil)
	}
	fmt.Println(ip) // should print "192.168.99.109"
}

func TestExtractIPAddressFromURL8(t *testing.T) {
	inputURL := "192.168.99.109"
	ip, err := ExtractIPAddressFromURL(inputURL)
	if err != nil {
		t.Errorf("Actual error = %v, expected error = %v.", err, nil)
	}
	fmt.Println(ip) // should print "192.168.99.109"
}

// --------------------------------------------------

func TestGetAppPlans1(t *testing.T) {
	appName := "mariadb"
	expected := []int{5, 10, 20}
	plans, _ := GetAppPlans(appName)

	if !reflect.DeepEqual(expected, plans) {
		t.Errorf("Expecting %+v but got %+v\n", expected, plans)
	}

	fmt.Printf("Plans - %+v\n", plans)
}

func TestGetAppPlans2(t *testing.T) {
	appName := "minio"
	expected := []int{5, 10, 20}
	plans, _ := GetAppPlans(appName)

	if !reflect.DeepEqual(expected, plans) {
		t.Errorf("Expecting %+v but got %+v\n", expected, plans)
	}

	fmt.Printf("Plans - %+v\n", plans)
}

func TestGetSmallestAppPlan(t *testing.T) {
	sortedPlans := []int{5, 10, 20}
	expected := 5
	result := GetSmallestAppPlan(sortedPlans)

	if expected != result {
		t.Errorf("Expecting %+v but got %+v\n", expected, result)
	}

	fmt.Println("Smallest plan:", result)
}

func TestGetAppDependencies1(t *testing.T) {
	deps, _ := GetAppDependencies("wordpress")
	fmt.Println(deps)
}

func TestGetAppDependencies2(t *testing.T) {
	deps, _ := GetAppDependencies("minio")
	fmt.Println(deps)
}

func TestGetKubeServerVersionHuman(t *testing.T) {
	version, _ := GetKubeServerVersionHuman()
	fmt.Printf("Human version: %s\n", version)
}

func TestGetKubeServerVersionCombined(t *testing.T) {
	combined, _ := GetKubeServerVersionCombined()
	fmt.Printf("Combined version: %d\n", combined)
}
