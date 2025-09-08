package ez

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/torloejborg/easykube/ekctx"
)

func GetEKContext() *ekctx.EKContext {

	return &ekctx.EKContext{

		Logger:  log.New(os.Stdout, "", log.LstdFlags),
		Printer: &ekctx.Printer{},
	}
}

func TestRenderYaml(t *testing.T) {

	a := &Addon{
		Name: "foo",
		Config: AddonConfig{
			ExtraPorts: []PortConfig{
				{
					HostPort: 9000,
					Protocol: "TCP",
				},
			},
			ExtraMounts: []MountConfig{
				{
					ContainerPath: "pgdata",
					HostPath:      "postgres",
				},
			},
		},
	}

	al := []*Addon{a}

	fmt.Println(al)

}

func TestConfigGeneratedFromAddons(t *testing.T) {

}

func TestCreateCluster(*testing.T) {
}

func TestClusterCreateReport(*testing.T) {

	a := &Addon{
		Name: "alpha",
		Config: AddonConfig{
			ExtraPorts: []PortConfig{
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

	b := &Addon{
		Name: "bravo",
		Config: AddonConfig{
			ExtraPorts: []PortConfig{
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

	c := &Addon{
		Name: "charlie", Config: AddonConfig{
			ExtraMounts: []MountConfig{
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
			ExtraPorts: []PortConfig{
				{
					HostPort: 7743,
					Protocol: "TCP",
					NodePort: 38475,
				},
			},
		},
	}

	var addons []*Addon
	addons = append(addons, a)
	addons = append(addons, b)
	addons = append(addons, c)

}

func TestUpdateConfigmap(t *testing.T) {
}

func TestGetInstalledAddons(t *testing.T) {

}

func TestCreatePersistenceDirectories(t *testing.T) {

}
