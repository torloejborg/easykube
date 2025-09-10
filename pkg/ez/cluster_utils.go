package ez

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/torloejborg/easykube/ekctx"
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/resources"
	"sigs.k8s.io/kind/pkg/cluster"
)

type IClusterUtils interface {
	CreateKindCluster(modules map[string]*Addon) string
	RenderToYAML(addonList []*Addon) string
	ConfigurationReport(addonList []*Addon) string
	EnsurePersistenceDirectory()
}

type ClusterUtils struct {
	Debug     bool
	EkConfig  *EasykubeConfigData
	EkContext *ekctx.EKContext
}

func NewClusterUtils(ctx *ekctx.EKContext) IClusterUtils {
	cfg, err := Kube.LoadConfig()
	if err != nil {
		panic(err)
	}
	return &ClusterUtils{
		EkConfig:  cfg,
		EkContext: ctx,
	}
}

func (u *ClusterUtils) ConfigurationReport(addonList []*Addon) string {

	portTmpl, _ := resources.AppResources.ReadFile("data/createreport.template")
	sb := new(strings.Builder)
	t := new(template.Template)

	template.Must(t.Parse(string(portTmpl))).Execute(sb, addonList)

	return sb.String()
}

func (u *ClusterUtils) CreateKindCluster(modules map[string]*Addon) string {

	// kind already exists, but not started
	search := Kube.FindContainer("kind-control-plane")

	addonList := make([]*Addon, 0)
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

		currentDir, e := os.Getwd()
		if e != nil {
			Kube.FmtRed("cannot get current working directory")
		}
		defer os.Chdir(currentDir)

		e = os.Chdir(homedir)
		if e != nil {
			Kube.FmtRed("cannot change directory to %s", homedir)
			panic(e)
		}

		configDir, _ := os.UserConfigDir()
		configFile := u.RenderToYAML(addonList)

		SaveFile(configFile, filepath.Join(configDir, "easykube", "easykube-cluster.yaml"))

		optNodeImage := cluster.CreateWithNodeImage(constants.KIND_IMAGE)
		optNoGreeting := cluster.CreateWithDisplaySalutation(false)
		optReady := cluster.CreateWithWaitForReady(20 * time.Second)

		kubeconfigPath := filepath.Join(homedir, ".kube", "easykube")

		optKubeConfig := cluster.CreateWithKubeconfigPath(kubeconfigPath)

		optConfig := cluster.CreateWithConfigFile(filepath.Join(configDir, "easykube", "easykube-cluster.yaml"))

		Kube.FmtGreen("Waiting for cluster ready")

		err := cp.Create(constants.CLUSTER_NAME, optKubeConfig, optConfig, optNodeImage, optNoGreeting, optReady)
		if nil != err {
			panic(err)
		}

		// initial cluster should be running now
		search = Kube.FindContainer(constants.KIND_CONTAINER)

		if search.IsRunning {
			Kube.FmtGreen("Configuring containerd")
			c1 := []string{"mkdir", "-p", "/etc/containerd/certs.d/localhost:5001"}
			Kube.Exec(search.ContainerID, c1)

			Kube.FmtGreen("Adding registry host")
			toml, err := resources.AppResources.ReadFile("data/hosts.toml")
			Kube.ContainerWriteFile(search.ContainerID, "/etc/containerd/certs.d/localhost:5001", "hosts.toml", toml)
			if err != nil {
				panic(err)
			}
		}
	}

	if search.Found && !search.IsRunning {
		Kube.FmtGreen("Starting existing cluster")
		Kube.StartContainer(search.ContainerID)
	}

	return u.ConfigurationReport(addonList)
}

func (u *ClusterUtils) RenderToYAML(addonList []*Addon) string {
	data, err := resources.AppResources.ReadFile("data/cluster_config.template")
	if err != nil {
		panic(err)
	}

	templ := template.New("config")
	templ, err = templ.Parse(string(data))
	if err != nil {
		panic(err)
	}

	buf := &bytes.Buffer{}

	err = templ.Execute(buf, addonList)
	if err != nil {
		panic(err)
	}

	return buf.String()
}

func (u *ClusterUtils) EnsurePersistenceDirectory() {

	addons := Kube.GetAddons()

	for _, a := range addons {
		if len(a.Config.ExtraMounts) > 0 {
			mounts := a.Config.ExtraMounts
			for m := range mounts {
				path := filepath.Join(mounts[m].PersistenceDir, mounts[m].HostPath)
				err := os.MkdirAll(path, 0777)
				if err != nil {
					panic(err)
				}
				err = os.Chmod(path, 0777)
				if err != nil {
					// ignore for now
				}
			}
		}
	}
}
