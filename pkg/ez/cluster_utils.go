package ez

import (
	"bytes"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/core"
	"github.com/torloejborg/easykube/pkg/resources"
	"sigs.k8s.io/kind/pkg/cluster"
)

type ClusterUtils struct {
	EkConfig *core.EasykubeConfigData
	ek       *core.Ek
}

func NewClusterUtils(ek *core.Ek) core.IClusterUtils {
	cfg, err := ek.Config.LoadConfig()
	if err != nil {
		panic(err)
	}
	return &ClusterUtils{
		EkConfig: cfg,
		ek:       ek,
	}
}

func (u *ClusterUtils) ConfigurationReport(addonList []core.IAddon) string {

	portTmpl, _ := resources.AppResources.ReadFile("data/createreport.template")
	sb := new(strings.Builder)
	t := new(template.Template)

	_ = template.Must(t.Parse(string(portTmpl))).Execute(sb, addonList)

	return sb.String()
}

func (u *ClusterUtils) CreateKindCluster(modules map[string]core.IAddon) (string, error) {

	// see if the cluster has been created
	search, _ := u.ek.ContainerRuntime.FindContainer(constants.KindContainer)

	addonList := make([]core.IAddon, 0)
	for _, addon := range modules {
		addonList = append(addonList, addon)
	}

	homedir, _ := u.ek.OsDetails.GetUserHomeDir()
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

		configDir, _ := u.ek.OsDetails.GetEasykubeConfigDir()

		cfg, err := u.ek.Config.LoadConfig()
		if err != nil {
			panic(err)
		}

		configFile := u.RenderToYAML(addonList, cfg)

		u.ek.Utils.SaveFile(configFile, filepath.Join(configDir, "easykube-cluster.yaml"))

		optNodeImage := cluster.CreateWithNodeImage(constants.KindImage)

		optKubeConfig := cluster.CreateWithKubeconfigPath(kubeconfigPath)
		optConfig := cluster.CreateWithConfigFile(filepath.Join(configDir, "easykube-cluster.yaml"))

		clusterErr := cp.Create(constants.ClusterName, optKubeConfig, optConfig, optNodeImage)
		if nil != clusterErr {
			panic(err)
		}

		// initial cluster should be running now
		search, _ = u.ek.ContainerRuntime.FindContainer(constants.KindContainer)

		if search.IsRunning {

			localhostReg := []string{"mkdir", "-p", "/etc/containerd/certs.d/_default"}
			if err := u.ek.ContainerRuntime.Exec(search.ContainerID, localhostReg); err != nil {
				return "", err
			}

			hosts, _ := resources.AppResources.ReadFile("data/cert.d/hosts.toml")
			if err := u.ek.ContainerRuntime.ContainerWriteFile(search.ContainerID, "/etc/containerd/certs.d/_default", "hosts.toml", hosts); err != nil {
				return "", err
			}

		}
	}

	if search.Found && !search.IsRunning {

		err := u.ek.ContainerRuntime.StartContainer(search.ContainerID)
		if err != nil {
			return "", err
		}

	}

	err := u.ek.Kubernetes.WaitForKindClusterReady(kubeconfigPath, 5*time.Minute)
	if err != nil {
		return "", err
	}

	return u.ConfigurationReport(addonList), nil
}

func (u *ClusterUtils) RenderToYAML(addonList []core.IAddon, config *core.EasykubeConfigData) string {
	data, err := resources.AppResources.ReadFile("data/cluster_config.template")
	if err != nil {
		panic(err)
	}

	x := struct {
		Config    *core.EasykubeConfigData
		AddonList []core.IAddon
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

	addons, err := u.ek.AddonReader.GetAddons()
	if err != nil {
		return err
	}

	for _, a := range addons {
		if len(a.GetConfig().ExtraMounts) > 0 {
			mounts := a.GetConfig().ExtraMounts
			for m := range mounts {
				path := filepath.Join(mounts[m].PersistenceDir, mounts[m].HostPath)
				err := u.ek.Fs.MkdirAll(path, 0777)
				if err != nil {
					panic(err)
				}
				err = u.ek.Fs.Chmod(path, 0777)
				if err != nil {
					// ignore for now
				}
			}
		}
	}
	return nil
}
