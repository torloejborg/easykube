package test

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/torloejborg/easykube/ekctx"
	"github.com/torloejborg/easykube/pkg/ez"
)

func GetEKContext() *ekctx.EKContext {

	return &ekctx.EKContext{

		Logger:  log.New(os.Stdout, "", log.LstdFlags),
		Printer: &ekctx.Printer{},
	}
}

func TestRenderYaml(t *testing.T) {

	a := &ez.Addon{
		Name: "foo",
		Config: ez.AddonConfig{
			ExtraPorts: []ez.PortConfig{
				{
					HostPort: 9000,
					Protocol: "TCP",
				},
			},
			ExtraMounts: []ez.MountConfig{
				{
					ContainerPath: "pgdata",
					HostPath:      "postgres",
				},
			},
		},
	}

	al := []*ez.Addon{a}

	cut := CreateFakeClusterUtils()
	fmt.Println(cut.RenderToYAML(al))
}

func TestConfigGeneratedFromAddons(t *testing.T) {
	ar := CreateFakeAddonReader()
	addonMap := ar.GetAddons()

	var addons []*ez.Addon

	for _, addon := range addonMap {
		addons = append(addons, addon)
	}

	cut := CreateFakeClusterUtils()
	yaml := cut.RenderToYAML(addons)
	fmt.Println(yaml)

}

func TestCreateCluster(*testing.T) {
	cu := CreateFakeClusterUtils()
	cu.CreateKindCluster(map[string]*ez.Addon{})
}

func TestClusterCreateReport(*testing.T) {

	a := &ez.Addon{
		Name: "alpha",
		Config: ez.AddonConfig{
			ExtraPorts: []ez.PortConfig{
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

	b := &ez.Addon{
		Name: "bravo",
		Config: ez.AddonConfig{
			ExtraPorts: []ez.PortConfig{
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

	c := &ez.Addon{
		Name: "charlie", Config: ez.AddonConfig{
			ExtraMounts: []ez.MountConfig{
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
			ExtraPorts: []ez.PortConfig{
				{
					HostPort: 7743,
					Protocol: "TCP",
					NodePort: 38475,
				},
			},
		},
	}

	var addons []*ez.Addon
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
