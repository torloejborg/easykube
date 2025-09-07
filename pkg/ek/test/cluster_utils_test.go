package test

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/torloejborg/easykube/ekctx"
	"github.com/torloejborg/easykube/pkg/ek"
)

func GetEKContext() *ekctx.EKContext {

	return &ekctx.EKContext{

		Logger:  log.New(os.Stdout, "", log.LstdFlags),
		Printer: &ekctx.Printer{},
	}
}

func TestRenderYaml(t *testing.T) {

	a := &ek.Addon{
		Name: "foo",
		Config: ek.AddonConfig{
			ExtraPorts: []ek.PortConfig{
				{
					HostPort: 9000,
					Protocol: "TCP",
				},
			},
			ExtraMounts: []ek.MountConfig{
				{
					ContainerPath: "pgdata",
					HostPath:      "postgres",
				},
			},
		},
	}

	al := []*ek.Addon{a}

	cut := CreateFakeClusterUtils()
	fmt.Println(cut.RenderToYAML(al))
}

func TestConfigGeneratedFromAddons(t *testing.T) {
	ar := CreateFakeAddonReader()
	addonMap := ar.GetAddons()

	var addons []*ek.Addon

	for _, addon := range addonMap {
		addons = append(addons, addon)
	}

	cut := CreateFakeClusterUtils()
	yaml := cut.RenderToYAML(addons)
	fmt.Println(yaml)

}

func TestCreateCluster(*testing.T) {
	cu := CreateFakeClusterUtils()
	cu.CreateKindCluster(map[string]*ek.Addon{})
}

func TestClusterCreateReport(*testing.T) {

	a := &ek.Addon{
		Name: "alpha",
		Config: ek.AddonConfig{
			ExtraPorts: []ek.PortConfig{
				{
					HostPort: 9000,
					Protocol: "TCP",
					NodePort: 234524,
				},
				{
					HostPort: 5432,
					NodePort: 123144,
				},
				{
					HostPort: 7777,
					Protocol: "UDP",
					NodePort: 531144,
				},
			},
		},
	}

	b := &ek.Addon{
		Name: "bravo",
		Config: ek.AddonConfig{
			ExtraPorts: []ek.PortConfig{
				{
					HostPort: 443,
					Protocol: "TCP",
					NodePort: 4342,
				},
				{
					HostPort: 80,
					Protocol: "TCP",
					NodePort: 8080,
				},
			},
		},
	}

	c := &ek.Addon{
		Name: "charlie", Config: ek.AddonConfig{
			ExtraMounts: []ek.MountConfig{
				{
					PersistenceDir: "/some/other/location",
					ContainerPath:  "/mnt/foo-a",
					HostPath:       "docker-a",
				},
				{
					ContainerPath: "/mnt/foo-b",
					HostPath:      "docker-b",
				},
				{
					ContainerPath: "/mnt/foo-c",
					HostPath:      "docker-c",
				},
				{
					ContainerPath: "/mnt/foo-d",
					HostPath:      "/some/abs/dir/docker-d",
				},
			},
			ExtraPorts: []ek.PortConfig{
				{
					HostPort: 7743,
					Protocol: "TCP",
					NodePort: 38475,
				},
			},
		},
	}

	var addons []*ek.Addon
	addons = append(addons, a)
	addons = append(addons, b)
	addons = append(addons, c)

	cu := CreateFakeClusterUtils()
	fmt.Println(cu.ConfigurationReport(addons))

}

func TestUpdateConfigmap(t *testing.T) {
	k8su := CreateFakeK8sUtil()
	k8su.CreateConfigmap("testing", "default")
	k8su.UpdateConfigMap("testing", "default", "myvalue", []byte("hello"))
}

func TestGetInstalledAddons(t *testing.T) {
	k8su := CreateFakeK8sUtil()
	addons, _ := k8su.GetInstalledAddons()
	fmt.Println(addons)
}

func TestCreatePersistenceDirectories(t *testing.T) {
	clu := CreateFakeClusterUtils()
	clu.EnsurePersistenceDirectory()

}
