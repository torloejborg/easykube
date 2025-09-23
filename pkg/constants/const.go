package constants

const KIND_IMAGE = "kindest/node:v1.34.0"
const KIND_CONTAINER = "easykube-control-plane"
const CLUSTER_NAME = "easykube"
const CLUSTER_CONTEXT = "kind-easykube"
const REGISTRY_IMAGE = "registry:2.8.3"
const REGISTRY_CONTAINER = "easykube-registry"
const KIND_NETWORK_NAME = "kind"
const LOCAL_REGISTRY = "localhost:5001"
const JS_LIB = "__jslib"
const ADDON_CM = "installed-addons"
const DEFAULT_NS = "default"

// KUSTOMIZE_TARGET_OUTPUT Name of file kustomize will render to
const KUSTOMIZE_TARGET_OUTPUT = ".out.yaml"

//
// flags for arguments
//

const FLAG_NODEPENDS = "nodepends"
const FLAG_FORCE = "force"
const FLAG_PULL = "pull"
const ARG_SECRETS = "secret"
const ARG_SKAFFOLD_NAME = "name"
const ARG_SKAFFOLD_LOCATION = "location"
const FLAG_KEYVALUE = "kv"
const FLAG_CLUSTER = "cluster"

//
// Binary dependencies
//

const KUBECTL_SEMVER = "~1.34"
const DOCKER_SEMVER = "^28"
const HELM_SEMVER = "~3.19"
const KUSTOMIZE_SEMVER = "~5.6"
