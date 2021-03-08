package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8syaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

// manifestOperation for Server Side Apply
type manifestOperation string

const (
	applySSA           manifestOperation = "apply"
	deleteSSA          manifestOperation = "delete"
	allowedMinutes                       = 30 // for caching
	marketplaceAccount                   = "zulh-civo"
	marketplaceBranch                    = "b"
)

// AppManifest is used when parsing app manifest.yaml from ~/.kubemart/apps folder
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

// KubemartConfigFile is the structure of ~/.kubemart/config.json file
type KubemartConfigFile struct {
	AppsLastUpdatedAt int64 `json:"apps_last_updated_at"`
}

// KubemartConfigMap is used when saving ConfigMap
type KubemartConfigMap struct {
	EmailAddress string
	DomainName   string
	ClusterName  string
	MasterIP     string
}

// KubemartPaths contains all important paths for kubemart operation
type KubemartPaths struct {
	RootDirectoryPath string
	AppsDirectoryPath string
	ConfigFilePath    string
}

// LatestGitHubReleaseResponse ...
type LatestGitHubReleaseResponse struct {
	TagName string `json:"tag_name"`
}

// GetKubemartPaths returns KubemartPaths struct
func GetKubemartPaths() (*KubemartPaths, error) {
	bp := &KubemartPaths{}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return bp, err
	}

	kubemartDirPath := filepath.Join(homeDir, ".kubemart")
	bp.RootDirectoryPath = kubemartDirPath
	bp.AppsDirectoryPath = filepath.Join(kubemartDirPath, "apps")
	bp.ConfigFilePath = filepath.Join(kubemartDirPath, "config.json")

	return bp, nil
}

// GetAppNamespace ...
func GetAppNamespace(appName string) (string, error) {
	bp, err := GetKubemartPaths()
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
	bp, err := GetKubemartPaths()
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
	bp, err := GetKubemartPaths()
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

// GetKubeAPIExtensionClientSet is similar to GetKubeClientSet.
// It does not have core APIs. It has APIs under "apiextensions.k8s.io/v1beta1"
// e.g. for "CustomResourceDefinition" kind.
func GetKubeAPIExtensionClientSet() (*apiextensionsclientset.Clientset, error) {
	clientset := &apiextensionsclientset.Clientset{}

	rc, err := GetRESTConfig()
	if err != nil {
		return clientset, err
	}

	cs, err := apiextensionsclientset.NewForConfig(rc)
	if err != nil {
		return clientset, err
	}

	return cs, nil
}

// IsKubemartConfigMapExist returns true if "kubemart-config" ConfigMap is found
func IsKubemartConfigMapExist() (bool, error) {
	namespace := "kubemart-system"
	clientset, err := GetKubeClientSet()
	if err != nil {
		return false, err
	}

	cmClient := clientset.CoreV1().ConfigMaps(namespace)
	_, err = cmClient.Get(context.Background(), "kubemart-config", metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// CreateKubemartConfigMap will create "kubemart-config" ConfigMap
func CreateKubemartConfigMap(bcm *KubemartConfigMap) error {
	namespace := "kubemart-system"
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
			Name:      "kubemart-config",
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
	ns, err := nsClient.Get(context.Background(), namespace, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	if !ns.DeletionTimestamp.IsZero() {
		DebugPrintf("%s namespace is being terminated\n", namespace)
		return false, nil
	}

	return true, nil
}

// CreateKubemartNamespace will create "kubemart-system" Namespace
func CreateKubemartNamespace() error {
	namespace := "kubemart-system"
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
	bp, err := GetKubemartPaths()
	if err != nil {
		fmt.Printf("Unable to get kubemart paths - %v\n", err.Error())
		os.Exit(1)
	}

	appManifestPath := filepath.Join(bp.AppsDirectoryPath, appName, "post_install.md")
	DebugPrintf("App post install file - %s\n", appManifestPath)
	file, err := ioutil.ReadFile(appManifestPath)
	if err != nil {
		fmt.Printf("Unable to load post-install notes for this app - %v\n", err.Error())
		os.Exit(1)
	}

	if runtime.GOOS == "windows" {
		fmt.Println(string(file))
	} else {
		out, err := glamour.Render(string(file), "dark")
		if err == nil {
			fmt.Println("App post-install notes:")
			fmt.Println(out)
		}
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
	_, err := exec.LookPath(name)
	if err == nil {
		return true
	}
	return false
}

// GitClone will run `git clone` and save the contents into directory
func GitClone(directory string) (string, error) {
	var stdErrBuf bytes.Buffer
	url := fmt.Sprintf("https://github.com/%s/kubernetes-marketplace.git", marketplaceAccount)

	args := []string{
		"clone",
		"--branch",
		marketplaceBranch,
		"--progress",
		url,
		path.Clean(directory),
	}
	DebugPrintf("git command args - %v\n", args)

	cmd := exec.Command("git", args...)
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
	args := []string{
		"-C",
		path.Clean(directory),
		"pull",
		"origin",
		marketplaceBranch,
	}
	DebugPrintf("git command args - %v\n", args)

	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// GetConfigFileLastUpdatedTimestamp is used to read timestamp field from ~/.kubemart/config.json file
func GetConfigFileLastUpdatedTimestamp() int64 {
	bp, err := GetKubemartPaths()
	if err != nil {
		return 0
	}

	configFile, err := ioutil.ReadFile(bp.ConfigFilePath)
	if err != nil {
		return 0
	}

	config := &KubemartConfigFile{}
	err = json.Unmarshal(configFile, config)
	if err != nil {
		return 0
	}

	return config.AppsLastUpdatedAt
}

// UpdateConfigFileLastUpdatedTimestamp will update timestamp field of ~/.kubemart/config.json file
func UpdateConfigFileLastUpdatedTimestamp() error {
	bp, err := GetKubemartPaths()
	if err != nil {
		return err
	}

	configFilePath := bp.ConfigFilePath
	config := &KubemartConfigFile{
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

// UpdateAppsCacheIfStale will run `git pull` in the context of ~/.kubemart/apps folder
// and update the timestamp field in the ~/.kubemart/config.json file
func UpdateAppsCacheIfStale() {
	lastUpdated := GetConfigFileLastUpdatedTimestamp()
	now := time.Now().Unix()
	diff := now - lastUpdated

	allowedSeconds := int64(60 * allowedMinutes)
	if diff > allowedSeconds {
		bp, err := GetKubemartPaths()
		if err != nil {
			fmt.Printf("Unable to load kubemart paths - %v\n", err)
			os.Exit(1)
		}

		appsFolder := bp.AppsDirectoryPath
		pullOutput, err := GitPull(appsFolder)
		if err != nil {
			fmt.Printf("Unable to Git pull latest apps - %v\n", err)
			fmt.Printf("You may need to run 'kubemart init' command to resolve this problem\n")
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
	args := []string{
		"-C",
		path.Clean(directory),
		"rev-parse",
		"--short",
		"HEAD",
	}
	DebugPrintf("git command args - %v\n", args)

	out, err := exec.Command("git", args...).Output()
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

// ExtractVersionFromContainerImage ...
func ExtractVersionFromContainerImage(image string) string {
	splitted := strings.Split(image, ":")
	if len(splitted) == 2 {
		return splitted[1]
	}

	return ""
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

// SanitizeVersionSegment ...
// Sometimes, k8s server minor version contains special characters
// e.g. 1.21+ (for v1.21.0-beta.0). This function will exclude all special
// characters from a version segment. For instance, if the input
// is "21+" (using minor segment as example here), this function will
// return "21" so we can parse it to integer without any issues.
func SanitizeVersionSegment(input string) string {
	re := regexp.MustCompile("[0-9]+")
	cleanedArr := re.FindAllString(input, -1)
	return cleanedArr[0]
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

// IsCRDExist ...
func IsCRDExist(crdName string) bool {
	clientset, err := GetKubeAPIExtensionClientSet()
	if err != nil {
		return false
	}

	crdClient := clientset.ApiextensionsV1().CustomResourceDefinitions()
	_, err = crdClient.Get(context.Background(), crdName, metav1.GetOptions{})
	if err != nil {
		return false
	}

	return true
}

// ApplyManifests ...
func ApplyManifests(manifests []string) error {
	for _, manifest := range manifests {
		operatorYAMLBytes := []byte(manifest)

		operation := applySSA
		err := ExecuteSSA(operatorYAMLBytes, &operation, "kubectl")
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteManifests ...
func DeleteManifests(manifests []string) error {
	for _, manifest := range manifests {
		operatorYAMLBytes := []byte(manifest)

		operation := deleteSSA
		err := ExecuteSSA(operatorYAMLBytes, &operation, "kubectl")
		if err != nil {
			return err
		}
	}

	return nil
}

// ExecuteSSA ...
// Inspired from: https://bit.ly/3b6tB6y
func ExecuteSSA(yamlData []byte, action *manifestOperation, owner string) error {
	DebugPrintf("==========\n")

	// Get REST config
	restConfig, err := GetRESTConfig()
	if err != nil {
		return err
	}

	// Prepare a decoder to read YAML manifest into `unstructured.Unstructured`
	decUnstructured := k8syaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	// Prepare a RESTMapper to find GVR
	dc, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	// Prepare the dynamic client
	dyn, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	// Decode YAML manifest into unstructured.Unstructured
	obj := &unstructured.Unstructured{}
	_, gvk, err := decUnstructured.Decode(yamlData, nil, obj)
	if err != nil {
		return err
	}

	DebugPrintf("GVK: %+v\n", gvk)
	serverVersion, err := GetKubeServerVersionCombined()
	if err != nil {
		return err
	}

	if serverVersion >= 119 {
		// To allow kubectl CLI user to manually run "kubectl apply -f kubemart-operator.yaml",
		// we need to add "kubectl.kubernetes.io/last-applied-configuration" annotation to this object.
		// Otherwise, they will get warnings when running that "kubectl apply" command.
		// * This applies to users using Kubernetes 1.19.x version and onwards.
		// * The warnings (not errors) still appear for users using Kubernetes 1.17.x and 1.18.x version.
		metadataInterface := obj.Object["metadata"]
		metadata, _ := metadataInterface.(map[string]interface{})
		annotationsInterface := metadata["annotations"]
		annotations, _ := annotationsInterface.(map[string]interface{})
		newAnnotations := make(map[string]interface{})

		// Take all current annotations and add them to new annotations
		for k, v := range annotations {
			newAnnotations[k] = v
		}
		newAnnotations["kubectl.kubernetes.io/last-applied-configuration"] = yamlData
		metadata["annotations"] = newAnnotations
		obj.Object["metadata"] = metadata
	}

	// Find GVR
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return err
	}

	// Obtain REST interface for the GVR
	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// namespaced resources should specify the namespace
		dr = dyn.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	} else {
		// for cluster-wide resources
		dr = dyn.Resource(mapping.Resource)
	}

	// Marshal object into JSON
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	if *action == applySSA {
		// Create or Update the object with SSA
		//     * types.ApplyPatchType indicates it's SSA operation
		//     * FieldManager specifies the field owner ID
		// A note from https://kubernetes.io/docs/reference/using-api/server-side-apply:
		// "It is strongly recommended for controllers to always "force" conflicts,
		// ...since they might not be able to resolve or act on these conflicts."
		force := true
		DebugPrintf("Applying manifest for %s using SSA...\n", obj.GetName())
		_, err = dr.Patch(context.Background(), obj.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{
			FieldManager: owner,
			Force:        &force,
		})
	}

	if *action == deleteSSA {
		DebugPrintf("Deleting manifest for %s using SSA...\n", obj.GetName())
		gp := int64(0)
		dpb := metav1.DeletePropagationBackground
		err = dr.Delete(context.Background(), obj.GetName(), metav1.DeleteOptions{
			GracePeriodSeconds: &gp,
			PropagationPolicy:  &dpb,
		})

		// if the target resource is not found, just move on
		if errors.IsNotFound(err) {
			err = nil
		}
	}

	DebugPrintf("==========\n")
	return err
}

// GetKubeServerVersion ...
func GetKubeServerVersion() (*version.Info, error) {
	v := &version.Info{}

	rc, err := GetRESTConfig()
	if err != nil {
		return v, err
	}

	dc, err := discovery.NewDiscoveryClientForConfig(rc)
	if err != nil {
		return v, err
	}

	return dc.ServerVersion()
}

// GetKubeServerVersionHuman ...
func GetKubeServerVersionHuman() (string, error) {
	version, err := GetKubeServerVersion()
	if err != nil {
		return "", err
	}

	return version.GitVersion, nil
}

// GetKubeServerVersionCombined ...
func GetKubeServerVersionCombined() (int, error) {
	version, err := GetKubeServerVersion()
	if err != nil {
		return 0, err
	}

	major := SanitizeVersionSegment(version.Major)
	minor := SanitizeVersionSegment(version.Minor)
	// DebugPrintf("k8s major version: %s\n", major)
	// DebugPrintf("k8s minor version: %s\n", minor)

	combined := fmt.Sprintf("%s%s", major, minor)
	vInt, err := strconv.Atoi(combined)
	if err != nil {
		return 0, err
	}

	return vInt, nil
}

// GetInstalledOperatorVersion ...
func GetInstalledOperatorVersion() (string, error) {
	cs, err := GetKubeClientSet()
	if err != nil {
		return "", err
	}

	deployClient := cs.AppsV1().Deployments("kubemart-system")
	deployment, err := deployClient.Get(context.Background(), "kubemart-operator-controller-manager", metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	containers := deployment.Spec.Template.Spec.Containers
	for _, container := range containers {
		if container.Name == "manager" {
			imageVersion := ExtractVersionFromContainerImage(container.Image)
			return imageVersion, nil
		}
	}

	return "", fmt.Errorf("Manager container not found")
}

// IsAppExist ...
func IsAppExist(appName string) bool {
	bp, err := GetKubemartPaths()
	if err != nil {
		return false
	}

	appFolderPath := filepath.Join(bp.AppsDirectoryPath, appName)
	if _, err := os.Stat(appFolderPath); !os.IsNotExist(err) {
		return true
	}
	return false
}

// GetLatestOperatorReleaseVersion ...
func GetLatestOperatorReleaseVersion() (string, error) {
	response, err := http.Get("https://api.github.com/repos/kubemart/kubemart-operator/releases/latest")
	if err != nil {
		return "", err
	}

	resp, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	var latestRelease LatestGitHubReleaseResponse
	json.Unmarshal(resp, &latestRelease)

	return latestRelease.TagName, nil
}

// GetLatestManifests ...
func GetLatestManifests() (string, error) {
	var manifests string

	latestVersion, err := GetLatestOperatorReleaseVersion()
	if err != nil {
		return manifests, err
	}

	url := fmt.Sprintf("https://github.com/kubemart/kubemart-operator/releases/download/%s/kubemart-operator.yaml", latestVersion)
	response, err := http.Get(url)
	if err != nil {
		return manifests, err
	}
	defer response.Body.Close()

	buf := new(bytes.Buffer)
	n, err := io.Copy(buf, response.Body)
	if err != nil {
		return manifests, err
	}

	if n == 0 {
		return manifests, fmt.Errorf("Manifests are empty")
	}

	manifests = buf.String()
	return manifests, nil
}
