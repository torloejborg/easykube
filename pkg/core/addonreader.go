package core

type IAddonReader interface {
	GetAddons() (map[string]IAddon, error)
	ExtractConfiguration(unconfigured IAddon) (*AddonConfig, error)
	ExtractJSON(input string) (string, bool)
	CheckAddonCompatibility() (string, error)
}
