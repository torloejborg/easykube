package ez

import (
	"bytes"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/resources"
	"sigs.k8s.io/kind/pkg/cluster"
)

type IClusterUtils interface {
	CreateKindCluster(modules map[string]IAddon) (string, error)
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

func (u *ClusterUtils) CreateKindCluster(modules map[string]IAddon) (string, error) {

	// see if the cluster has been created
	search, _ := Kube.FindContainer(constants.KindContainer)

	addonList := make([]IAddon, 0)
	for _, addon := range modules {
		addonList = append(addonList, addon)
	}

	homedir, _ := Kube.GetUserHomeDir()
	kubeconfigPath := filepath.Join(homedir, ".kube", "easykube")

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

		configDir, _ := Kube.GetEasykubeConfigDir()

		cfg, err := Kube.LoadConfig()
		if err != nil {
			panic(err)
		}

		configFile := u.RenderToYAML(addonList, cfg)

		SaveFile(configFile, filepath.Join(configDir, "easykube-cluster.yaml"))

		optNodeImage := cluster.CreateWithNodeImage(constants.KindImage)

		optKubeConfig := cluster.CreateWithKubeconfigPath(kubeconfigPath)
		optConfig := cluster.CreateWithConfigFile(filepath.Join(configDir, "easykube-cluster.yaml"))

		clusterErr := cp.Create(constants.ClusterName, optKubeConfig, optConfig, optNodeImage)
		if nil != clusterErr {
			panic(err)
		}

		// initial cluster should be running now
		search, _ = Kube.FindContainer(constants.KindContainer)

		if search.IsRunning {

			localhostReg := []string{"mkdir", "-p", "/etc/containerd/certs.d/_default"}
			if err := Kube.Exec(search.ContainerID, localhostReg); err != nil {
				return "", err
			}

			hosts, _ := resources.AppResources.ReadFile("data/cert.d/hosts.toml")
			if err := Kube.ContainerWriteFile(search.ContainerID, "/etc/containerd/certs.d/_default", "hosts.toml", hosts); err != nil {
				return "", err
			}

		}
	}

	if search.Found && !search.IsRunning {

		err := Kube.StartContainer(search.ContainerID)
		if err != nil {
			return "", err
		}

	}

	err := Kube.WaitForKindClusterReady(kubeconfigPath, 5*time.Minute)
	if err != nil {
		return "", err
	}

	return u.ConfigurationReport(addonList), nil
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
