package core

type IEasykubeConfig interface {
	LoadConfig() (*EasykubeConfigData, error)
	MakeConfig() error
	EditConfig() error
	LaunchEditor(config, editor string)
	PathToConfigFile() string
	IsZotConfigInSync(configData *EasykubeConfigData) (bool, error)
	GenerateZotRegistryConfig(*EasykubeConfigData) error
	GenerateZotRegistryCredentials(*EasykubeConfigData) error
	WriteConfig(*EasykubeConfigData) error
	CopyConfigResources() error
	HasConfiguration() bool
}

type MirrorRegistry struct {
	RegistryUrl string `mapstructure:"registry-url"`
	UserKey     string `mapstructure:"username"`
	PasswordKey string `mapstructure:"password"`
}

type EasykubeConfigData struct {
	AddonDir          string `mapstructure:"addon-root"`
	PersistenceDir    string `mapstructure:"persistence-dir"`
	ConfigurationDir  string `mapstructure:"config-dir"`
	ContainerRuntime  string `mapstructure:"container-runtime"`
	ConfigurationFile string
	MirrorRegistries  []MirrorRegistry `mapstructure:"mirror-registries"`
}
