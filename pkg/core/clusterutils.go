package core

type IClusterUtils interface {
	CreateKindCluster(modules map[string]IAddon) (string, error)
	RenderToYAML(addonList []IAddon, config *EasykubeConfigData) string
	ConfigurationReport(addonList []IAddon) string
	EnsurePersistenceDirectory() error
}
