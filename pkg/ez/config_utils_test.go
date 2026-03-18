package ez_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/torloejborg/easykube/pkg/ez"
	"github.com/torloejborg/easykube/pkg/textutils"
	"github.com/torloejborg/easykube/test"
)

func initConfigTests(t *testing.T) {

	osd := test.CreateOsDetailsMock(t)
	osd.EXPECT().GetEasykubeConfigDir().Return("/home/some-user/.config", nil).AnyTimes()
	osd.EXPECT().GetUserHomeDir().Return("/home/some-user", nil).AnyTimes()

	ez.Kube.UseOsDetails(osd)
	ez.Kube.UseFilesystemLayer(afero.NewMemMapFs())
	config := ez.NewEasykubeConfig()
	ez.Kube.UseEasykubeConfig(config)
	_ = ez.Kube.MakeConfig()

	ez.Kube.UseAddonReader(ez.CreateAddonReaderImpl(config))
	ez.Kube.UseClusterUtils(ez.CreateClusterUtilsImpl())

}

func TestMakeDefaultConfig(t *testing.T) {
	initConfigTests(t)
	cfgdir, _ := ez.Kube.GetEasykubeConfigDir()
	homeDir, _ := ez.Kube.GetUserHomeDir()

	exists := ez.FileOrDirExists(filepath.Join(cfgdir, "config.yaml"))
	if !exists {
		t.Errorf("expected easykube config file to exist")
	}

	data, err := ez.Kube.LoadConfig()
	if err != nil {
		panic(err)
	}

	if data.AddonDir != filepath.Join(homeDir, "addons") {
		t.Errorf("expected addons dir to be ./addons")
	}

}

func TestLoadDefaultConfigWithPrivateRegistries(t *testing.T) {
	initConfigTests(t)

	// use config with private registries enabled
	cfg := textutils.TrimMargin(`
	|easykube:
	|  addon-root: /home/tor/code/research/easykube-addons
	|  config-dir: /home/tor/.config/easykube
	|  persistence-dir: /home/tor/.config/easykube/persistence
	|  container-runtime: docker
	|  mirror-registries:
	|   - repository-url: https://foo.com
	|     username: userkey1
	|     password: passkey1
	|   - repository-url: https://bar.com
	|     username: userkey2
	|     password: passkey2
	`, "|")

	f, _ := ez.Kube.Fs.OpenFile(ez.Kube.IEasykubeConfig.PathToConfigFile(), os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	_, _ = f.WriteString(cfg)

	exists := ez.FileOrDirExists(ez.Kube.IEasykubeConfig.PathToConfigFile())
	if !exists {
		t.Errorf("expected easykube config file to exist")
	}

	data, err := ez.Kube.LoadConfig()
	if err != nil {
		panic(err)
	}

	reg1 := data.MirrorRegistries[0]
	reg2 := data.MirrorRegistries[1]

	if reg1.RepositoryURL != "https://foo.com" {
		t.Errorf("expected https://foo.com got %s", reg1.RepositoryURL)
	}

	if reg2.RepositoryURL != "https://bar.com" {
		t.Errorf("expected https://bar.com got %s", reg2.RepositoryURL)
	}

}

var filesExist = []struct {
	file   string
	exists bool
}{
	{"config.yaml", true},
	{"localtest.me.crt", true},
	{"localtest.me.key", true},
	{"zot-config.json", true},
	{"persistence", false},
	{"easykube-cluster.yaml", false},
}

func TestVerifyConfigurationFilesCopiedToConfigDir(t *testing.T) {
	initConfigTests(t)
	cfgdir, _ := ez.Kube.GetEasykubeConfigDir()

	for _, tt := range filesExist {
		t.Run(tt.file, func(t *testing.T) {
			found := ez.FileOrDirExists(filepath.Join(cfgdir, tt.file))
			if found != tt.exists {
				t.Errorf("expected file %v file to exist in %v, was %v", tt.file, cfgdir, found)
			}
		})
	}
}

func TestZotConfigGeneration(t *testing.T) {
	initConfigTests(t)

	cfg := &ez.EasykubeConfigData{
		AddonDir:          "",
		PersistenceDir:    "",
		ConfigurationDir:  "",
		ContainerRuntime:  "",
		ConfigurationFile: "",
		MirrorRegistries: []ez.MirrroRegistry{
			{
				RepositoryURL: "https://foo.com",
				UserKey:       "none",
				PasswordKey:   "nil",
			},
			{
				RepositoryURL: "https://bar.com",
			},
		},
	}

	ez.Kube.SyncWithZot(cfg)

}
