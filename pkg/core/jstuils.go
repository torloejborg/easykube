package core

type IJsUtils interface {
	ExecAddonScript(a IAddon) error
	ExtractConfigurationObject(a IAddon) (*AddonConfig, error)
}
