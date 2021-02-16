package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	allowedMinutes     = 30 // for caching
	marketplaceAccount = "zulh-civo"
	marketplaceBranch  = "b"
)

// AppManifest is used when parsing remote app manifest.yaml
type AppManifest struct {
	Namespace    string   `yaml:"namespace"`
	Dependencies []string `yaml:"dependencies"`
	Plans        []struct {
		Label         string `yaml:"label"`
		Configuration map[string]struct {
			Value string `yaml:"value"`
		} `yaml:"configuration"`
	} `yaml:"plans"`
}

// BizaarConfigFile is the structure of ~/.bizaar/config.json file
type BizaarConfigFile struct {
	AppsLastUpdatedAt int64 `json:"apps_last_updated_at"`
}

// BizaarConfigMap is used when saving ConfigMap
type BizaarConfigMap struct {
	EmailAddress string
	DomainName   string
	ClusterName  string
	MasterIP     string
}

// BizaarPaths contains all important paths for Bizaar operation
type BizaarPaths struct {
	RootDirectoryPath string
	AppsDirectoryPath string
	ConfigFilePath    string
}

// GetBizaarPaths returns BizaarPaths struct
func GetBizaarPaths() (*BizaarPaths, error) {
	bp := &BizaarPaths{}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return bp, err
	}

	bizaarDirPath := fmt.Sprintf("%s/.bizaar", homeDir)
	bp.RootDirectoryPath = bizaarDirPath
	bp.AppsDirectoryPath = fmt.Sprintf("%s/apps", bizaarDirPath)
	bp.ConfigFilePath = fmt.Sprintf("%s/config.json", bizaarDirPath)

	return bp, nil
}

// GetAppNamespace ...
func GetAppNamespace(appName string) (string, error) {
	bp, err := GetBizaarPaths()
	if err != nil {
		return "", err
	}

	appManifestPath := fmt.Sprintf("%s/%s/manifest.yaml", bp.AppsDirectoryPath, appName)
	file, err := ioutil.ReadFile(appManifestPath)
	if err != nil {
		return "", err
	}

	manifest := AppManifest{}
	err = yaml.Unmarshal(file, &manifest)
	if err != nil {
		return "", err
	}

	ns := manifest.Namespace
	return ns, nil
}

// GetAppDependencies will fetch app's dependencies and return in slice type
func GetAppDependencies(appName string) ([]string, error) {
	dependencies := []string{}
	bp, err := GetBizaarPaths()
	if err != nil {
		return dependencies, err
	}

	appManifestPath := fmt.Sprintf("%s/%s/manifest.yaml", bp.AppsDirectoryPath, appName)
	file, err := ioutil.ReadFile(appManifestPath)
	if err != nil {
		return dependencies, err
	}

	manifest := AppManifest{}
	err = yaml.Unmarshal(file, &manifest)
	if err != nil {
		return dependencies, err
	}

	for _, dependency := range manifest.Dependencies {
		dep := strings.ToLower(dependency)
		cleaned, err := SanitizeDependencyName(dep)
		if err != nil {
			return dependencies, fmt.Errorf("Unable to sanitize %s", dep)
		}
		dependencies = append(dependencies, cleaned)
	}

	return dependencies, nil
}

// GetAppPlans returns sorted app plans e.g. [5,10,20]
func GetAppPlans(appName string) ([]int, error) {
	plans := []int{}
	bp, err := GetBizaarPaths()
	if err != nil {
		return plans, err
	}

	appManifestPath := fmt.Sprintf("%s/%s/manifest.yaml", bp.AppsDirectoryPath, appName)
	file, err := ioutil.ReadFile(appManifestPath)
	if err != nil {
		return plans, err
	}

	manifest := AppManifest{}
	err = yaml.Unmarshal(file, &manifest)
	if err != nil {
		return plans, err
	}

	for _, plan := range manifest.Plans {
		conf := plan.Configuration
		keys := reflect.ValueOf(conf).MapKeys()
		strKeys := make([]string, len(keys))
		for i := 0; i < len(keys); i++ {
			strKeys[i] = keys[i].String()
		}

		for _, key := range strKeys {
			p := ExtractPlanIntFromPlanStr(conf[key].Value)
			if p > 0 {
				plans = append(plans, p)
			}
		}
	}

	sort.Ints(plans)
	return plans, nil
}

// GetSmallestAppPlan take plans slice e.g. [20,5,10] and return 5 (int)
func GetSmallestAppPlan(sortedPlans []int) int {
	return sortedPlans[0]
}

// GetKubeClientSet reads default k8s context and return k8s client of it.
// Loading order as follows:
// * If "--kubeconfig" flag was supplied, create k8s client from it
// * If user has "KUBECONFIG" variable defined, create k8s client from it
// * Otherwise, create k8s client from "~/.kube/config" file
func GetKubeClientSet() (*kubernetes.Clientset, error) {
	clientset := &kubernetes.Clientset{}

	rc, err := GetRESTConfig()
	if err != nil {
		return clientset, err
	}

	cs, err := kubernetes.NewForConfig(rc)
	if err != nil {
		return clientset, err
	}

	return cs, nil
}

// IsBizaarConfigMapExist returns true if "bizaar-config" ConfigMap is found
func IsBizaarConfigMapExist() (bool, error) {
	namespace := "bizaar"
	clientset, err := GetKubeClientSet()
	if err != nil {
		return false, err
	}

	cmClient := clientset.CoreV1().ConfigMaps(namespace)
	_, err = cmClient.Get(context.Background(), "bizaar-config", metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// CreateBizaarConfigMap will create "bizaar-config" ConfigMap
func CreateBizaarConfigMap(bcm *BizaarConfigMap) error {
	namespace := "bizaar"
	configMapData := make(map[string]string)
	configMapData["email"] = bcm.EmailAddress
	configMapData["domain"] = bcm.DomainName
	configMapData["cluster_name"] = bcm.ClusterName
	configMapData["master_ip"] = bcm.MasterIP

	configMap := &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bizaar-config",
			Namespace: namespace,
		},
		Data: configMapData,
	}

	clientset, err := GetKubeClientSet()
	if err != nil {
		return err
	}

	cmClient := clientset.CoreV1().ConfigMaps(namespace)
	_, err = cmClient.Create(context.Background(), configMap, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

// IsNamespaceExist returns true if Namespace is found
func IsNamespaceExist(namespace string) (bool, error) {
	clientset, err := GetKubeClientSet()
	if err != nil {
		return false, err
	}

	nsClient := clientset.CoreV1().Namespaces()
	_, err = nsClient.Get(context.Background(), namespace, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// CreateBizaarNamespace will create "bizaar-config" ConfigMap
func CreateBizaarNamespace() error {
	namespace := "bizaar"
	clientset, err := GetKubeClientSet()
	if err != nil {
		return err
	}

	ns := &v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}

	nsClient := clientset.CoreV1().Namespaces()
	_, err = nsClient.Create(context.Background(), ns, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

// DeleteNamespace ...
func DeleteNamespace(namespace string) error {
	clientset, err := GetKubeClientSet()
	if err != nil {
		return err
	}

	nsClient := clientset.CoreV1().Namespaces()
	err = nsClient.Delete(context.Background(), namespace, metav1.DeleteOptions{})
	return err
}

// RenderPostInstallMarkdown will fetch app's post_install.md and render to user's stdout
func RenderPostInstallMarkdown(appName string) {
	bp, err := GetBizaarPaths()
	if err != nil {
		fmt.Printf("Unable to get Bizaar paths - %v\n", err.Error())
		os.Exit(1)
	}

	appManifestPath := fmt.Sprintf("%s/%s/post_install.md", bp.AppsDirectoryPath, appName)
	file, err := ioutil.ReadFile(appManifestPath)
	if err != nil {
		fmt.Printf("Unable to load post-install notes for this app - %v\n", err.Error())
		os.Exit(1)
	}

	out, err := glamour.Render(string(file), "dark")
	if err == nil {
		fmt.Println("App post-install notes:")
		fmt.Println(out)
	}
}

// DebugPrintf is used to print debug messages (useful during development)
func DebugPrintf(format string, a ...interface{}) (n int, err error) {
	logLevel := os.Getenv("LOGLEVEL")
	if logLevel == "debug" {
		withLabel := fmt.Sprintf("[DEBUG] %s", format)
		return fmt.Printf(withLabel, a...)
	}
	return 0, nil
}

// IsCommandAvailable returns true if a program is installed
func IsCommandAvailable(name string) bool {
	cmdToRun := fmt.Sprintf("command -v %s", name)
	cmd := exec.Command("/bin/sh", "-c", cmdToRun)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// GitClone will run `git clone` and save the contents into directory
func GitClone(directory string) (string, error) {
	var stdErrBuf bytes.Buffer
	url := fmt.Sprintf("https://github.com/%s/kubernetes-marketplace.git", marketplaceAccount)
	cmdToRun := fmt.Sprintf("git clone --branch %s --progress %s %s", marketplaceBranch, url, directory)
	cmd := exec.Command("/bin/sh", "-c", cmdToRun)
	cmd.Stderr = &stdErrBuf // because `git clone` use stderr
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	cloneOutput := stdErrBuf.String()
	return cloneOutput, nil
}

// GitPull will run `git pull` in the context of given directory
func GitPull(directory string) (string, error) {
	cmdToRun := fmt.Sprintf("git -C %s pull origin %s", directory, marketplaceBranch)
	out, err := exec.Command("/bin/sh", "-c", cmdToRun).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// GetConfigFileLastUpdatedTimestamp is used to read timestamp field from ~/.bizaar/config.json file
func GetConfigFileLastUpdatedTimestamp() int64 {
	bp, err := GetBizaarPaths()
	if err != nil {
		return 0
	}

	configFile, err := ioutil.ReadFile(bp.ConfigFilePath)
	if err != nil {
		return 0
	}

	config := &BizaarConfigFile{}
	err = json.Unmarshal(configFile, config)
	if err != nil {
		return 0
	}

	return config.AppsLastUpdatedAt
}

// UpdateConfigFileLastUpdatedTimestamp will update timestamp field of ~/.bizaar/config.json file
func UpdateConfigFileLastUpdatedTimestamp() error {
	bp, err := GetBizaarPaths()
	if err != nil {
		return err
	}

	configFilePath := bp.ConfigFilePath
	config := &BizaarConfigFile{
		AppsLastUpdatedAt: time.Now().Unix(),
	}

	file, err := json.MarshalIndent(config, "", " ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(configFilePath, file, 0644)
	if err != nil {
		return err
	}

	return nil
}

// UpdateAppsCacheIfStale will run `git pull` in the context of ~/.bizaar/apps folder
// and update the timestamp field in the ~/.bizaar/config.json file
func UpdateAppsCacheIfStale() {
	lastUpdated := GetConfigFileLastUpdatedTimestamp()
	now := time.Now().Unix()
	diff := now - lastUpdated

	allowedSeconds := int64(60 * allowedMinutes)
	if diff > allowedSeconds {
		bp, err := GetBizaarPaths()
		if err != nil {
			fmt.Printf("Unable to load Bizaar paths - %v\n", err)
			os.Exit(1)
		}

		appsFolder := bp.AppsDirectoryPath
		pullOutput, err := GitPull(appsFolder)
		if err != nil {
			fmt.Printf("Unable to Git pull latest apps - %v\n", err)
			os.Exit(1)
		}
		DebugPrintf("Pull output: %+v\n", pullOutput)

		err = UpdateConfigFileLastUpdatedTimestamp()
		if err != nil {
			fmt.Printf("Unable to config file's timestamp field - %v\n", err)
			os.Exit(1)
		}
		DebugPrintf("Config file's timestamp field updated successfully\n")
	}
}

// GitLatestCommitHash will return the last commit (short version) from Git folder
func GitLatestCommitHash(directory string) (string, error) {
	cmdToRun := fmt.Sprintf("git -C %s rev-parse --short HEAD", directory)
	out, err := exec.Command("/bin/sh", "-c", cmdToRun).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// GetKubeconfig will load kubeconfig from KUBECONFIG environment variable.
// If it's empty, it will load from ~/.kube/config file.
func GetKubeconfig() clientcmd.ClientConfig {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	return kubeConfig
}

// GetRESTConfig returns the kubeconfig's REST config
func GetRESTConfig() (*rest.Config, error) {
	kubeConfig := GetKubeconfig()
	restConfig := &rest.Config{}
	restConfig, err := kubeConfig.ClientConfig()
	if err != nil {
		return restConfig, err
	}
	DebugPrintf("Loading REST config for host %v\n", restConfig.Host)
	return restConfig, nil
}

// GetConfigAccess returns ConfigAccess
func GetConfigAccess() clientcmd.ConfigAccess {
	kubeConfig := GetKubeconfig()
	configAccess := kubeConfig.ConfigAccess()
	return configAccess
}

// GetCurrentContext return current context
func GetCurrentContext() (string, error) {
	configAccess := GetConfigAccess()
	config, err := configAccess.GetStartingConfig()
	if err != nil {
		return "", err
	}

	currentContext := config.CurrentContext
	return currentContext, nil
}

// GetClusterName return the kubeconfig's cluster name
func GetClusterName() (string, error) {
	currentContext, err := GetCurrentContext()
	if err != nil {
		return "", err
	}

	configAccess := GetConfigAccess()
	config, err := configAccess.GetStartingConfig()
	if err != nil {
		return "", err
	}

	clusterName := config.Contexts[currentContext].Cluster
	return clusterName, nil
}

// ExtractIPAddressFromURL takes URL (procotol://IP:port) and returns IP
func ExtractIPAddressFromURL(url string) (string, error) {
	r, err := regexp.Compile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)
	if err != nil {
		return "", err
	}

	ip := r.FindString(url)
	if ip != "" {
		return ip, nil
	}

	return "", fmt.Errorf("IP address is empty")
}

// ExtractPlanIntFromPlanStr takes plan string i.e. "5Gi" and return 5 (int).
// If something goes wrong, it will return -1 (int).
func ExtractPlanIntFromPlanStr(input string) (output int) {
	r, err := regexp.Compile(`[0-9]+`)
	if err != nil {
		return -1
	}

	str := r.FindString(input)
	if str == "" {
		return -1
	}

	output, err = strconv.Atoi(str)
	if err != nil {
		return -1
	}

	return output
}

// SanitizeDependencyName ...
// https://rubular.com/r/5ibwrOnew3vKpf
func SanitizeDependencyName(input string) (string, error) {
	emptyStr := ""
	r, err := regexp.Compile(`^[a-z-0-9]*`)
	if err != nil {
		return emptyStr, err
	}

	cleaned := r.FindString(input)
	if cleaned == emptyStr {
		return emptyStr, fmt.Errorf("Dependency name is empty")
	}

	return cleaned, nil
}

// GetMasterIP returns the master/control-plane IP address
func GetMasterIP() (string, error) {
	kubeConfig := GetKubeconfig()
	restConfig := &rest.Config{}
	restConfig, err := kubeConfig.ClientConfig()
	if err != nil {
		return "", err
	}
	serverURL := restConfig.Host
	return ExtractIPAddressFromURL(serverURL)
}

// --------------------------------------------------------
// Not used
// --------------------------------------------------------

// GetKubeClientSetFromRESTConfig takes k8s REST config and returns k8s clientset
// func GetKubeClientSetFromRESTConfig(restConfig *rest.Config) (*kubernetes.Clientset, error) {
// 	clientset := &kubernetes.Clientset{}
// 	cs, err := kubernetes.NewForConfig(restConfig)
// 	if err != nil {
// 		return clientset, err
// 	}
// 	return cs, nil
// }

// KubeconfigGetter implements clientcmd#KubeconfigGetter
// https://pkg.go.dev/k8s.io/client-go/tools/clientcmd#KubeconfigGetter
// func KubeconfigGetter() (*api.Config, error) {
// 	return clientcmd.NewDefaultClientConfigLoadingRules().Load()
// }
