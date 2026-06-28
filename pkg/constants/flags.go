package constants

// FlagCluster cmdline flag, used for targeting other clusters than kind-easykube
const FlagCluster = "cluster"

// FlagForce used in add command, forces installation of an addon
const FlagForce = "force"

// FlagKeyValue Used for providing key/value pairs of data to javascript context
const FlagKeyValue = "kv"

// FlagNoDepends when installing an addon, only install that one
const FlagNoDepends = "nodepends"

// FlagPull when installing addons, force pull of fresh images
const FlagPull = "pull"

// FlagEdit for configuration, opens the config in an editor
const FlagEdit = "edit"

// FlagConfigDir allows for an alternate configuration dir other that ~/.config/easykube
const FlagConfigDir = "config-dir"

// FlagUseDefaults when creating a new configuration, do not enter interactive mode, use defaults.
const FlagUseDefaults = "use-defaults"

// ArgSkaffoldLocation where to create the new addon
const ArgSkaffoldLocation = "location"

// ArgSkaffoldName what to name a new addon
const ArgSkaffoldName = "name"
