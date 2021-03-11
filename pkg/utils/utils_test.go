package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
// To test GetKubeClientSet() function
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

func TestGetKubeAPIExtensionClientSet(t *testing.T) {
	clientset, _ := GetKubeAPIExtensionClientSet()
	crdClient := clientset.ApiextensionsV1().CustomResourceDefinitions()
	crds, _ := crdClient.List(context.Background(), v1.ListOptions{})
	totalCrds := len(crds.Items)

	if totalCrds != 0 {
		t.Errorf("Expected 0 CRDs but actual is %d", totalCrds)
	}
}

func TestIsKubemartConfigMapExist(t *testing.T) {
	actual, _ := IsKubemartConfigMapExist()
	expected := false
	if expected != actual {
		t.Errorf("Expected %t but actual is %t", expected, actual)
	}
}

func TestCreateKubemartConfigMap(t *testing.T) {
	kcm := KubemartConfigMap{
		EmailAddress: "test@example.com",
		DomainName:   "example.com",
		ClusterName:  "test",
		MasterIP:     "123.456.789.000",
	}
	err := CreateKubemartConfigMap(&kcm)
	if err != nil {
		t.Errorf("Error when creating ConfigMap - %s", err)
	}

	actual, _ := IsKubemartConfigMapExist()
	expected := true
	if expected != actual {
		t.Errorf("Expected %t but actual is %t", expected, actual)
	}
}

func TestIsNamespaceExist1(t *testing.T) {
	actual, _ := IsNamespaceExist("cool-system")
	expected := false
	if expected != actual {
		t.Errorf("Expected %t but actual is %t", expected, actual)
	}
}

func TestIsNamespaceExist2(t *testing.T) {
	actual, _ := IsNamespaceExist("kubemart-system")
	expected := true
	if expected != actual {
		t.Errorf("Expected %t but actual is %t", expected, actual)
	}
}

func TestCreateKubemartNamespace(t *testing.T) {
	// The namespace has been created before running this test.
	// So, recreating it here will cause an error.
	actualError := CreateKubemartNamespace()
	expectedError := "namespaces \"kubemart-system\" already exists"
	if expectedError != actualError.Error() {
		t.Errorf("Expected %s but actual is %s", expectedError, actualError)
	}
}

func TestDeleteNamespace(t *testing.T) {
	actualError := DeleteNamespace("dummy")
	if actualError != nil {
		t.Errorf("Expected nil error but actual is %s", actualError)
	}
}

func TestDebugPrintf(t *testing.T) {
	os.Setenv("LOGLEVEL", "debug")
	bytesNum, _ := DebugPrintf("hello")
	if bytesNum == 0 {
		t.Errorf("Expected non-zero bytes but actual is %d", bytesNum)
	}
}

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

func TestGitClone(t *testing.T) {
	homeDir, _ := os.UserHomeDir()
	targetDir := fmt.Sprintf("%s/.kubemart/apps", homeDir)
	output, _ := GitClone(targetDir)
	fmt.Printf("Clone output: %s\n", output)
}

func TestUpdateConfigFileLastUpdatedTimestamp(t *testing.T) {
	before := GetConfigFileLastUpdatedTimestamp()
	_ = UpdateConfigFileLastUpdatedTimestamp()
	after := GetConfigFileLastUpdatedTimestamp()

	if before == after {
		t.Errorf("Before and after timestamps are same. Before: %d. After: %d.", before, after)
	}
}

func TestGitPull(t *testing.T) {
	homeDir, _ := os.UserHomeDir()
	targetDir := fmt.Sprintf("%s/.kubemart/apps", homeDir)

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

func TestGetConfigFileLastUpdatedTimestamp(t *testing.T) {
	timestampInt := GetConfigFileLastUpdatedTimestamp()
	if timestampInt == 0 {
		t.Errorf("Expected non-zero timestamp but actual is %d", timestampInt)
	}
}

func TestUpdateAppsCacheIfStale(t *testing.T) {
	actual, _ := UpdateAppsCacheIfStale()
	expected := true
	if expected != actual {
		t.Errorf("Expected %t but actual is %t", expected, actual)
	}
}

func TestGitLatestCommitHash(t *testing.T) {
	homeDir, _ := os.UserHomeDir()
	appsDir := fmt.Sprintf("%s/.kubemart/apps", homeDir)
	latestCommitHash, _ := GitLatestCommitHash(appsDir)
	if latestCommitHash == "" {
		t.Errorf("Expected non-empty latest commit hash but actual is %s", latestCommitHash)
	}
}

// Note: this test must be after "git clone" test (see above)
func TestGetPostInstallMarkdown(t *testing.T) {
	postInstallMsg, _ := GetPostInstallMarkdown("rabbitmq")
	postInstallMsgLC := strings.ToLower(postInstallMsg)

	if !strings.Contains(postInstallMsgLC, "rabbitmq") {
		t.Errorf("Actual post-install message doesn't contain 'rabbitmq' - %s", postInstallMsgLC)
	}
}

// Note: this test must be after "git clone" test (see above)
func TestGetAppPlans(t *testing.T) {
	actual, _ := GetAppPlans("mariadb")
	expected := []int{5, 10, 20}
	if !elementsMatch(expected, actual) {
		t.Errorf("Expected %v but actual is %v", expected, actual)
	}
}

// Note: this test must be after "git clone" test (see above)
func TestGetSmallestAppPlan(t *testing.T) {
	sortedPlans := []int{5, 10, 20}
	expected := 5
	result := GetSmallestAppPlan(sortedPlans)

	if expected != result {
		t.Errorf("Expecting %+v but got %+v\n", expected, result)
	}

	fmt.Println("Smallest plan:", result)
}

// GetRESTConfig() depends on GetKubeconfig()
// So, we are testing two functions in one go here
func TestGetRESTConfig(t *testing.T) {
	restConfig, _ := GetRESTConfig()
	actual := restConfig.Host

	o, _ := exec.Command("kubectl", "config", "view", "-o", "jsonpath='{.clusters[0].cluster.server}'").Output()
	expected := string(o)
	expected = strings.TrimSuffix(expected, "'")
	expected = strings.TrimPrefix(expected, "'")

	if expected != actual {
		t.Errorf("Expected %s but actual is %s", expected, actual)
	}
}

// GetCurrentContext() depends on GetConfigAccess()
// So, we are testing two functions in one go here
func TestGetCurrentContext(t *testing.T) {
	actual, _ := GetCurrentContext()
	expected := "kind-default"
	if expected != actual {
		t.Errorf("Expected %s but actual is %s", expected, actual)
	}
}

// GetClusterName() depends on GetConfigAccess()
// So, we are testing two functions in one go here
func TestGetClusterName(t *testing.T) {
	actual, _ := GetClusterName()
	expected := "kind-default"
	if expected != actual {
		t.Errorf("Expected %s but actual is %s", expected, actual)
	}
}

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

func TestExtractPlanIntFromPlanStr1(t *testing.T) {
	actual := ExtractPlanIntFromPlanStr("5Gi")
	expected := 5
	if expected != actual {
		t.Errorf("Expected %d but actual is %d", expected, actual)
	}
}

func TestExtractPlanIntFromPlanStr2(t *testing.T) {
	actual := ExtractPlanIntFromPlanStr("5Gib")
	expected := 5
	if expected != actual {
		t.Errorf("Expected %d but actual is %d", expected, actual)
	}
}

func TestExtractPlanIntFromPlanStr3(t *testing.T) {
	actual := ExtractPlanIntFromPlanStr("5GB")
	expected := 5
	if expected != actual {
		t.Errorf("Expected %d but actual is %d", expected, actual)
	}
}

func TestExtractPlanIntFromPlanStr4(t *testing.T) {
	actual := ExtractPlanIntFromPlanStr("5 Gi")
	expected := 5
	if expected != actual {
		t.Errorf("Expected %d but actual is %d", expected, actual)
	}
}

func TestExtractPlanIntFromPlanStr5(t *testing.T) {
	actual := ExtractPlanIntFromPlanStr("5 Gib")
	expected := 5
	if expected != actual {
		t.Errorf("Expected %d but actual is %d", expected, actual)
	}
}

func TestExtractPlanIntFromPlanStr6(t *testing.T) {
	actual := ExtractPlanIntFromPlanStr("5 GB")
	expected := 5
	if expected != actual {
		t.Errorf("Expected %d but actual is %d", expected, actual)
	}
}

func TestExtractPlanIntFromPlanStr7(t *testing.T) {
	actual := ExtractPlanIntFromPlanStr("five")
	expected := -1
	if expected != actual {
		t.Errorf("Expected %d but actual is %d", expected, actual)
	}
}

func TestExtractPlanIntFromPlanStr8(t *testing.T) {
	actual := ExtractPlanIntFromPlanStr("fiveGi")
	expected := -1
	if expected != actual {
		t.Errorf("Expected %d but actual is %d", expected, actual)
	}
}

func TestExtractPlanIntFromPlanStr9(t *testing.T) {
	actual := ExtractPlanIntFromPlanStr("five Gi")
	expected := -1
	if expected != actual {
		t.Errorf("Expected %d but actual is %d", expected, actual)
	}
}

func TestExtractVersionFromContainerImage(t *testing.T) {
	actual := ExtractVersionFromContainerImage("nginx:1.2.3")
	expected := "1.2.3"
	if expected != actual {
		t.Errorf("Expected %s but actual is %s", expected, actual)
	}
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

func TestGetKubeServerVersionHuman(t *testing.T) {
	version, _ := GetKubeServerVersionHuman()
	fmt.Printf("Human version: %s\n", version)
}

func TestGetKubeServerVersionCombined(t *testing.T) {
	combined, _ := GetKubeServerVersionCombined()
	fmt.Printf("Combined version: %d\n", combined)
}

// --------------------
// Test helpers
// --------------------

type dummyt struct{}

func (t dummyt) Errorf(string, ...interface{}) {}

func elementsMatch(listA, listB interface{}) bool {
	return assert.ElementsMatch(dummyt{}, listA, listB)
}
