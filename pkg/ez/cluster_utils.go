package ez

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/resources"
	"sigs.k8s.io/kind/pkg/cluster"
)

type IClusterUtils interface {
	CreateKindCluster(modules map[string]IAddon) string
	RenderToYAML(addonList []IAddon, config *EasykubeConfigData) string
	ConfigurationReport(addonList []IAddon) string
	EnsurePersistenceDirectory() error
}

type ClusterUtils struct {
	Debug     bool
	EkConfig  *EasykubeConfigData
	EkContext *CobraCommandHelperImpl
}

func NewClusterUtils() IClusterUtils {
	cfg, err := Kube.LoadConfig()
	if err != nil {
		panic(err)
	}
	return &ClusterUtils{
		EkConfig: cfg,
	}
}

func (u *ClusterUtils) ConfigurationReport(addonList []IAddon) string {

	portTmpl, _ := resources.AppResources.ReadFile("data/createreport.template")
	sb := new(strings.Builder)
	t := new(template.Template)

	_ = template.Must(t.Parse(string(portTmpl))).Execute(sb, addonList)

	return sb.String()
}

func (u *ClusterUtils) CreateKindCluster(modules map[string]IAddon) string {

	// see if the cluster has been created
	search, _ := Kube.FindContainer("kind-control-plane")

	addonList := make([]IAddon, 0)
	for _, addon := range modules {
		addonList = append(addonList, addon)
	}

	if !search.Found {
		var cp *cluster.Provider

		if u.EkConfig.ContainerRuntime == "docker" {
			cp = cluster.NewProvider(cluster.ProviderWithDocker())
		}

		if u.EkConfig.ContainerRuntime == "podman" {
			cp = cluster.NewProvider(cluster.ProviderWithPodman())
		}

		if cp == nil {
			panic("no cluster provider")
		}
		homedir, _ := Kube.GetUserHomeDir()

		configDir, _ := os.UserConfigDir()
		configFile := u.RenderToYAML(addonList, u.EkConfig)

		SaveFile(configFile, filepath.Join(configDir, "easykube", "easykube-cluster.yaml"))

		optNodeImage := cluster.CreateWithNodeImage(constants.KIND_IMAGE)
		optNoGreeting := cluster.CreateWithDisplaySalutation(false)
		//optReady := cluster.CreateWithWaitForReady(10 * time.Second)

		kubeconfigPath := filepath.Join(homedir, ".kube", "easykube")
		optKubeConfig := cluster.CreateWithKubeconfigPath(kubeconfigPath)
		optConfig := cluster.CreateWithConfigFile(filepath.Join(configDir, "easykube", "easykube-cluster.yaml"))

		err := cp.Create(constants.CLUSTER_NAME, optKubeConfig, optConfig, optNodeImage, optNoGreeting)
		if nil != err {
			panic(err)
		}

		// initial cluster should be running now
		search, _ = Kube.FindContainer(constants.KIND_CONTAINER)

		fmt.Println()
		if search.IsRunning {
			_, _ = Kube.FmtSpinner(func() (any, error) {

				localhostReg := []string{"mkdir", "-p", "/etc/containerd/certs.d/localhost:5001"}
				if err := Kube.Exec(search.ContainerID, localhostReg); err != nil {
					return nil, err
				}

				registryReg := []string{"mkdir", "-p", "/etc/containerd/certs.d/registry.localtest.me:5001"}
				if err := Kube.Exec(search.ContainerID, registryReg); err != nil {
					return nil, err
				}

				localhost, _ := resources.AppResources.ReadFile("data/reg-localhost.toml")
				if err := Kube.ContainerWriteFile(search.ContainerID, "/etc/containerd/certs.d/localhost:5001", "hosts.toml", localhost); err != nil {
					return nil, err
				}

				registry, _ := resources.AppResources.ReadFile("data/reg-registry.toml")
				if err := Kube.ContainerWriteFile(search.ContainerID, "/etc/containerd/certs.d/registry.localtest.me:5001", "hosts.toml", registry); err != nil {
					return nil, err
				}

				return nil, nil
			}, "Wiring local registry to control plane")
		}
	}

	if search.Found && !search.IsRunning {

		_, _ = Kube.FmtSpinner(func() (any, error) {
			return nil, Kube.StartContainer(search.ContainerID)
		}, "Starting existing cluster")

	}

	return u.ConfigurationReport(addonList)
}

func (u *ClusterUtils) RenderToYAML(addonList []IAddon, config *EasykubeConfigData) string {
	data, err := resources.AppResources.ReadFile("data/cluster_config.template")
	if err != nil {
		panic(err)
	}

	x := struct {
		Config    *EasykubeConfigData
		AddonList []IAddon
	}{Config: config, AddonList: addonList}

	templ := template.New("config")
	templ, err = templ.Parse(string(data))
	if err != nil {
		panic(err)
	}

	buf := &bytes.Buffer{}

	err = templ.Execute(buf, x)
	if err != nil {
		panic(err)
	}

	return buf.String()
}

func (u *ClusterUtils) EnsurePersistenceDirectory() error {

	addons, err := Kube.GetAddons()
	if err != nil {
		return err
	}

	for _, a := range addons {
		if len(a.GetConfig().ExtraMounts) > 0 {
			mounts := a.GetConfig().ExtraMounts
			for m := range mounts {
				path := filepath.Join(mounts[m].PersistenceDir, mounts[m].HostPath)
				err := Kube.MkdirAll(path, 0777)
				if err != nil {
					panic(err)
				}
				err = Kube.Chmod(path, 0777)
				if err != nil {
					// ignore for now
				}
			}
		}
	}
	return nil
}
