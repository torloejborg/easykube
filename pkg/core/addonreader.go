package core

type IAddonReader interface {
	GetAddons() (map[string]IAddon, error)
	ExtractConfiguration(unconfigured IAddon) (*AddonConfig, error)
	CheckAddonCompatibility() (string, error)
}
