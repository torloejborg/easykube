package ez

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/spf13/afero"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/torloejborg/easykube/pkg/constants"

	"github.com/torloejborg/easykube/pkg/resources"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	applyv1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/utils/ptr"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type K8SUtilsImpl struct {
	Clientset  *kubernetes.Clientset
	RestConfig *rest.Config
	EKContext  *CobraCommandHelperImpl
	Fs         afero.Fs
}

func (k *K8SUtilsImpl) HasKubeConfig() bool {
	return k.Clientset != nil && k.RestConfig != nil
}

// ExternalSecret Defines a struct to capture the structure of an ExternalSecret from the ExternalSecretOperator
type ExternalSecret struct {
	ApiVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
	Spec struct {
		// Define the structure of the spec according to your needs
		RefreshInterval string `yaml:"refreshInterval"`
		SecretStoreRef  struct {
			Name string `yaml:"name"`
			Kind string `yaml:"kind"`
		} `yaml:"secretStoreRef"`
		Data []struct {
			SecretKey string `yaml:"secretKey"`
			RemoteRef struct {
				Key      string `yaml:"key"`
				Property string `yaml:"property"`
			} `yaml:"remoteRef"`
		} `yaml:"data"`
	} `yaml:"spec"`
}

// KubernetesSecret represents the structure of a Kubernetes Secret resource.
type KubernetesSecret struct {
	ApiVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	Metadata   map[string]string `yaml:"metadata"`
	Data       map[string]string `yaml:"data"`
	Type       string            `yaml:"type"`
}

type IK8SUtils interface {
	// ReloadClientSet Reloads the kubernetes client
	//
	// After creating a kind cluster a new kubeconfig is generated, this function make
	// sure we operate on the current one
	ReloadClientSet() error
	// CreateSecret
	//
	// Creates a kubernetes secret, kubernetes likes to base64 encode secrets, but a map of plain
	// strings can be passed, some fuzzy check is done to detect if encoding should happen or not. This is
	// not 100% perfect, and can fail in some instances - A good candidate for rewriting.
	CreateSecret(namespace, secretName string, data map[string]string) error
	//GetSecret Fetches a secret from kubernetes
	//
	// Returns
	GetSecret(name, namespace string) (map[string][]byte, error)
	// CreateConfigmap
	//
	// Creates an empty configmap
	CreateConfigmap(name, namespace string) error

	// DeleteKeyFromConfigmap
	//
	// Removes a key from a configmap and saves it
	DeleteKeyFromConfigmap(name, namespace, key string)
	// ReadConfigmap
	//
	// Fetch a configmap from kubernetes
	ReadConfigmap(name string, namespace string) (map[string]string, error)
	// UpdateConfigMap
	//
	// Updates a key in a configmap
	UpdateConfigMap(name, namespace, key string, data []byte)
	// HasKubeConfig
	//
	// true if K8sUtils are ready to go - the clientset and rest client are initialized.
	HasKubeConfig() bool

	// FindContainerInPod
	//
	// Attempts to locate a container in a given deployment, if there are multiple containers matching
	// containerPartialName, the first is matched. Returns a triplet of (pod.Name, container.Name, error)
	FindContainerInPod(deploymentName, namespace, containerPartialName string) (string, string, error)
	// ExecInPod
	//
	// Runs a 'kubectl exec' using the API, returns stdout and stderr as strings
	ExecInPod(namespace, pod, command string, args []string) (string, string, error)

	// CopyFileToPod
	//
	// Copies a local file into a remote pod
	CopyFileToPod(namespace, pod, container, localPath, remotePath string) error
	// ListPods
	//
	// Gets all pod names in a namespace
	ListPods(namespace string) ([]string, error)
	// GetInstalledAddons
	//
	// Queries the constants.ADDON_CM configmap for installed addons
	GetInstalledAddons() ([]string, error)

	// PatchCoreDNS
	//
	// Teaches CoreDNS to understand the localtest.me loopback domain
	PatchCoreDNS()

	// WaitForDeploymentReadyWatch pauses execution until some deployment enters ready state
	WaitForDeploymentReadyWatch(name, namespace string) error

	// WaitForCRD pauses execution for a duration of time until some custom resources have appeared in a cluster
	//
	// group, version, kind are the coordinates for the CRD, timeout specifies wait time
	WaitForCRD(group, version, kind string, timeout time.Duration) error
	// TransformExternalSecret creates a kubernetes secret from an external secret definition (ExternalSecretOperator)
	//
	// secret is the source which contains keys and references, mockData specifies the replacements for the keys,
	// addonName is the source of the mockData, namespace specifies where to create the secret
	TransformExternalSecret(secret ExternalSecret, mockData map[string]map[string]string, addonName, namespace string) KubernetesSecret
}

func NewK8SUtils() IK8SUtils {

	impl := &K8SUtilsImpl{
		Clientset:  nil,
		RestConfig: nil,
	}

	err := impl.ReloadClientSet()
	if err != nil {
		Kube.FmtRed("Failed to create Kubernetes client: %v", err)
	}

	return impl
}

func (k *K8SUtilsImpl) ReloadClientSet() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		Kube.FmtRed("cannot determine homedir")
		os.Exit(-1)
	}

	kubeconfigPath := filepath.Join(homeDir, ".kube", "easykube")

	if !FileOrDirExists(kubeconfigPath) {
		Kube.FmtYellow("expecting %s to exist, create the cluster and this message will disappear", kubeconfigPath)
		return nil
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	k.RestConfig = config

	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	k.Clientset = clientset
	return nil
}

func (k *K8SUtilsImpl) GetSecret(name, namespace string) (map[string][]byte, error) {

	ctx := context.Background()
	cm, err := k.Clientset.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{
		TypeMeta: metav1.TypeMeta{},
	})

	return cm.Data, err
}

func (k *K8SUtilsImpl) PatchCoreDNS() {

	ctx := context.Background()

	cs, _ := resources.AppResources.ReadFile("data/coredns/coredns-deployment.yaml")
	corefile, _ := resources.AppResources.ReadFile("data/coredns/coredns.config")
	localdb, _ := resources.AppResources.ReadFile("data/coredns/local.db")

	k.UpdateConfigMap("coredns", "kube-system", "local.db", localdb)
	k.UpdateConfigMap("coredns", "kube-system", "Corefile", corefile)

	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, gKV, _ := decode(cs, nil, nil)
	if gKV.Kind == "Deployment" {
		depl := obj.(*appsv1.Deployment)
		_, e := k.Clientset.AppsV1().Deployments("kube-system").Update(ctx, depl, metav1.UpdateOptions{
			TypeMeta: metav1.TypeMeta{},
		})
		if e != nil {
			panic(e)
		}
	}
}

func (k8s *K8SUtilsImpl) CreateConfigmap(name, namespace string) error {
	ctx := context.Background()

	cm := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	exists, _ := k8s.Clientset.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})

	if exists.Name == "" {

		_, err := k8s.Clientset.CoreV1().ConfigMaps(namespace).Create(ctx, &cm, metav1.CreateOptions{})

		if err != nil {
			return err
		}
	}

	return nil
}

func (k *K8SUtilsImpl) UpdateConfigMap(name, namespace, key string, data []byte) {

	ctx := context.Background()

	cm, _ := k.Clientset.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{
		TypeMeta: metav1.TypeMeta{},
	})

	if cm.Data == nil {
		cm.Data = make(map[string]string)
	}

	cm.Data[key] = string(data)
	_, err := k.Clientset.CoreV1().ConfigMaps(namespace).Update(ctx, cm, metav1.UpdateOptions{
		TypeMeta:     metav1.TypeMeta{},
		FieldManager: "easykube",
	})

	if err != nil {
		panic(err)
	}
}

func (k8s *K8SUtilsImpl) GetInstalledAddons() ([]string, error) {
	result := make([]string, 0)
	addons, err := k8s.ReadConfigmap(constants.ADDON_CM, constants.DEFAULT_NS)

	if err != nil {
		return result, err
	}

	for key, _ := range addons {
		result = append(result, key)
	}

	return result, nil
}

func (k8s *K8SUtilsImpl) ReadConfigmap(name string, namespace string) (map[string]string, error) {
	result := make(map[string]string, 0)
	cmap, err := k8s.Clientset.CoreV1().ConfigMaps(namespace).Get(context.Background(), name,
		metav1.GetOptions{
			TypeMeta: metav1.TypeMeta{
				Kind: "ConfigMap",
			},
		})

	if err != nil {
		return nil, err
	}

	for key, val := range cmap.Data {
		result[key] = val
	}

	return result, nil
}

func (k *K8SUtilsImpl) DeleteKeyFromConfigmap(name, namespace, key string) {
	cmap, err := k.Clientset.CoreV1().ConfigMaps(namespace).Get(context.Background(), name,
		metav1.GetOptions{
			TypeMeta: metav1.TypeMeta{
				Kind: "ConfigMap",
			},
		})
	if err != nil {
		panic(err)
	}

	delete(cmap.Data, key)

	_, err = k.Clientset.CoreV1().ConfigMaps(namespace).Update(context.Background(), cmap,
		metav1.UpdateOptions{
			TypeMeta: metav1.TypeMeta{
				Kind: "ConfigMap",
			},
		})

	if err != nil {
		panic(err)
	}
}

func (k *K8SUtilsImpl) WaitForDeploymentReadyWatch(name, namespace string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	watcher, err := k.Clientset.AppsV1().Deployments(namespace).Watch(ctx, metav1.ListOptions{
		FieldSelector:  fmt.Sprintf("metadata.name=%s", name),
		TimeoutSeconds: ptr.To(int64(60 * time.Second)),
		Watch:          true,
	})

	if err != nil {
		return fmt.Errorf("failed to set up watch: %w", err)
	}

	defer watcher.Stop()

	Kube.FmtGreen("Waiting for deployment %q in namespace %q to become ready...", name, namespace)

	for event := range watcher.ResultChan() {
		if event.Type == watch.Error {
			return fmt.Errorf("received error event")
		}

		dep, ok := event.Object.(*appsv1.Deployment)
		if !ok {
			continue
		}

		if dep.Status.ReadyReplicas == *dep.Spec.Replicas {
			Kube.FmtGreen("Deployment is ready!")
			return nil
		}
	}

	return fmt.Errorf("watch closed or timed out before deployment became ready")
}

func (k8s *K8SUtilsImpl) ListPods(namespace string) ([]string, error) {

	pods, err := k8s.Clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	podList := make([]string, 0)

	for i := range pods.Items {
		podList = append(podList, pods.Items[i].Name)
	}
	return podList, nil
}

func (k *K8SUtilsImpl) ExecInPod(namespace, pod, command string, args []string) (string, string, error) {

	// Compose command
	fullCommand := append([]string{command}, args...)

	// Prepare the REST request
	req := k.Clientset.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(pod).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			//Container: containerName, // set if needed
			Command: fullCommand,
			Stdin:   false,
			Stdout:  true,
			Stderr:  true,
			TTY:     false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(k.RestConfig, "POST", req.URL())
	if err != nil {
		return "", "", fmt.Errorf("failed to create executor: %w", err)
	}

	var stdout, stderr bytes.Buffer

	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if err != nil {
		return stdout.String(), stderr.String(), fmt.Errorf("exec failed: %w", err)
	}

	return stdout.String(), stderr.String(), nil
}

func (k *K8SUtilsImpl) WaitForCRD(
	group, version, kind string,
	timeout time.Duration,
) error {
	disco, err := discovery.NewDiscoveryClientForConfig(k.RestConfig)
	if err != nil {
		return err
	}

	gvk := fmt.Sprintf("%s/%s", group, version)

	return wait.PollUntilContextTimeout(context.Background(), 1*time.Second, timeout, true, func(ctx context.Context) (bool, error) {
		apiList, err := disco.ServerResourcesForGroupVersion(gvk)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				fmt.Println("not found")
				return false, nil // keep polling
			}
			panic(err)
			return false, err // real error
		}

		for _, res := range apiList.APIResources {
			if res.Kind == kind {
				return true, nil // CRD is now available
			}
		}
		fmt.Println("polling goes on")
		return false, nil // keep polling
	})
}

func (k *K8SUtilsImpl) CreateSecret(namespace, secretName string, data map[string]string) error {

	var isProbablyBase64 = func(s string) bool {
		decoded, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return false
		}
		// Check if re-encoding gives the original string (trims padding if any)
		reEncoded := base64.StdEncoding.EncodeToString(decoded)
		if s != reEncoded {
			return false
		}

		// Optionally: ensure it's valid UTF-8
		return utf8.Valid(decoded)
	}

	var s = applyv1.Secret(secretName, namespace)
	sdata := make(map[string][]byte)

	for k := range data {
		if isProbablyBase64(data[k]) {
			decoded, _ := base64.StdEncoding.DecodeString(data[k])
			sdata[k] = decoded
		} else {
			// If it's not base64, use as-is (convert to []byte)
			sdata[k] = []byte(data[k])
		}
	}

	s.WithData(sdata)

	_, e := k.Clientset.CoreV1().Secrets(namespace).Apply(context.Background(), s, metav1.ApplyOptions{
		TypeMeta:     metav1.TypeMeta{},
		Force:        false,
		FieldManager: "easykube",
	})

	// todo: better err handling
	if e != nil {
		return e
	} else {
		return nil
	}
}

func (k *K8SUtilsImpl) CopyFileToPod(namespace, pod, container, localPath, remotePath string) error {
	file, err := Kube.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer file.Close()

	// Prepare tar archive in memory
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat local file: %w", err)
	}

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return fmt.Errorf("failed to create tar header: %w", err)
	}
	// Set header.Name to the final filename inside the container
	header.Name = filepath.ToSlash(filepath.Base(remotePath))

	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write tar header: %w", err)
	}
	if _, err := io.Copy(tw, file); err != nil {
		return fmt.Errorf("failed to write file to tar: %w", err)
	}
	if err := tw.Close(); err != nil {
		return fmt.Errorf("failed to close tar writer: %w", err)
	}

	// Prepare the exec command
	cmd := []string{"tar", "xmf", "-", "-C", filepath.ToSlash(filepath.Dir(remotePath))}
	req := k.Clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: container,
			Command:   cmd,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(k.RestConfig, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("failed to create executor: %w", err)
	}

	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  &buf,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Tty:    false,
	})
	if err != nil {
		return fmt.Errorf("failed to stream tar to pod: %w", err)
	}

	return nil
}

func (k *K8SUtilsImpl) FindContainerInPod(deploymentName, namespace, containerPartialName string) (string, string, error) {
	pods, err := k.Clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return "", "", fmt.Errorf("failed to list pods in namespace %q: %w", namespace, err)
	}

	for _, pod := range pods.Items {
		// Match pod owned by ReplicaSet linked to the Deployment
		foundOwner := false
		for _, owner := range pod.OwnerReferences {
			if owner.Kind == "ReplicaSet" && strings.HasPrefix(owner.Name, deploymentName+"-") {
				foundOwner = true
				break
			}
		}
		if !foundOwner {
			continue
		}

		// Match container by partial name
		for _, container := range pod.Spec.Containers {
			if strings.Contains(container.Name, containerPartialName) {
				return pod.Name, container.Name, nil
			}
		}
	}

	return "", "", fmt.Errorf("no pod/container found for deployment=%q, containerPartial=%q", deploymentName, containerPartialName)
}

func (k *K8SUtilsImpl) TransformExternalSecret(secret ExternalSecret, mockData map[string]map[string]string, addonName, namespace string) KubernetesSecret {

	// Initialize the Kubernetes Secret
	k8sSecret := KubernetesSecret{
		ApiVersion: "v1",
		Kind:       "Secret",
		Metadata: map[string]string{
			"name":      secret.Metadata.Name,
			"namespace": namespace,
		},
		Data: make(map[string]string),
		Type: "Opaque",
	}

	// Populate the Secret data from mockData
	for _, dataItem := range secret.Spec.Data {
		// Extract the key and property from remoteRef
		key := dataItem.RemoteRef.Key
		property := dataItem.RemoteRef.Property

		// Look up the value in the nested mockData structure
		if appData, exists := mockData[key]; exists {
			if value, exists := appData[property]; exists {

				k8sSecret.Data[dataItem.SecretKey] = value
			} else {
				Kube.FmtYellow("No supplied value were matched for substitution with property '%s' in external secret '%s' mapped by '%s'",
					property,
					secret.Metadata.Name,
					key)
			}
		} else {

			Kube.FmtYellow("Warning: Key %s not found in map sourced from addon %s for external secret",
				key,
				addonName,
				secret.Metadata.Name)
		}
	}

	return k8sSecret

}
