package constants

const KIND_IMAGE = "docker.io/kindest/node:v1.35.0"
const KIND_CONTAINER = "easykube-control-plane"
const CLUSTER_NAME = "easykube"
const CLUSTER_CONTEXT = "kind-easykube"
const REGISTRY_IMAGE = "docker.io/library/registry:2.8.3"
const REGISTRY_CONTAINER = "easykube-registry"
const KIND_NETWORK_NAME = "kind"
const LOCAL_REGISTRY = "registry.localtest.me:5001"
const JS_LIB = "__jslib"
const ADDON_CM = "installed-addons"
const DEFAULT_NS = "default"
const EASYKUBE_SECRET_NAME = "easykube-secrets"

// KUSTOMIZE_TARGET_OUTPUT Name of file kustomize will render to
const KUSTOMIZE_TARGET_OUTPUT = ".out.yaml"

//
// flags for arguments
//

const FLAG_NODEPENDS = "nodepends"
const FLAG_FORCE = "force"
const FLAG_PULL = "pull"
const ARG_SECRETS = "secret"
const ARG_CONFIG_FILE = "config"
const ARG_SKAFFOLD_NAME = "name"
const ARG_SKAFFOLD_LOCATION = "location"
const FLAG_KEYVALUE = "kv"
const FLAG_CLUSTER = "cluster"
const INGRESS_HTTP_PORT = "http-port"
const INGRESS_HTTPS_PORT = "https-port"

//
// Binary dependencies
//

const KUBECTL_SEMVER = "~1.35"
const DOCKER_SEMVER = "^29"
const HELM_SEMVER = "~3.19"
const KUSTOMIZE_SEMVER = "~5.8"
const PODMAN_SEMVER = "~5.8"
