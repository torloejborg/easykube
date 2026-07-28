package constants

// AddonCm name of the kubernetes configmap to keep track of installed addons
const AddonCm = "installed-addons"

// ClusterContext What the easykube cluster context should be named
const ClusterContext = "kind-easykube"

// ClusterName What to name the kind cluster
const ClusterName = "easykube"

// DefaultNs Namespace where easykube should keep bookkeeping stuff
const DefaultNs = "default"

// JsLib Where to locate the companion javascript wrappers
const JsLib = "__jslib"

// KindNetworkName How to name the docker/podman network connecting kind and the local registry
const KindNetworkName = "kind"

// KustomizeTargetOutput Default result file for rendered kustomize manifests
const KustomizeTargetOutput = ".out.yaml"

// LocalRegistryPort Which port the local registry should listen on
const LocalRegistryPort = "5001"

// LocalRegistry domain name of the local registry
const LocalRegistry = "registry.localtest.me:" + LocalRegistryPort

// EnvEasykubeConfigDir Environment variable overriding the default configuration directory
const EnvEasykubeConfigDir = "EASYKUBE_CONFIG_DIR"

// ZotCredentials Name of the file where credentials for zot is store
const ZotCredentials = "zot-credentials.json"

// ZotConfig Main configuration filename for the zot registry
const ZotConfig = "zot-config.json"
