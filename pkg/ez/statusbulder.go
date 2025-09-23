package ez

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/torloejborg/easykube/pkg/constants"
)

type StatusBuilderImpl struct {
	VersionUtils VersionUtils
}

type IStatusBuilder interface {
	DoContainerCheck() error
	DoBinaryCheck() error
	DoAddonRepositoryCheck() error
	getDockerVersion() string
	getHelmVersion() string
	getKubectlVersion() string
	getKustomizeVersion() string
	getVersionStr(in, wants string, inErr error) string
}

func NewStatusBuilder() IStatusBuilder {
	return &StatusBuilderImpl{
		VersionUtils: NewVersionUtils(),
	}
}

func (s *StatusBuilderImpl) DoContainerCheck() error {
	if !Kube.IsContainerRuntimeAvailable() {
		Kube.FmtRed("Container runtime not available, is docker running??")
		os.Exit(-1)
	}

	running := func(containerID string) {
		if Kube.IsContainerRunning(containerID) {
			Kube.FmtGreen("✓ %s container", containerID)
		} else {
			Kube.FmtRed("⚠ %s container not running", containerID)
		}
	}

	Kube.FmtGreen("Container configuration")
	running(constants.REGISTRY_CONTAINER)
	running(constants.KIND_CONTAINER)

	if Kube.IsNetworkConnectedToContainer(constants.REGISTRY_CONTAINER, "kind") {
		Kube.FmtGreen("✓ %s connected to kind network", constants.REGISTRY_CONTAINER)
	} else {
		Kube.FmtRed("⚠ %s not connected to kind network", constants.REGISTRY_CONTAINER)
	}

	return nil
}

func (s *StatusBuilderImpl) DoBinaryCheck() error {

	hasBinary := func(name string, vFunc func() string) {
		_, err := exec.LookPath(name)
		if err != nil {
			Kube.FmtRed("⚠ " + name)
		} else {

			version := vFunc()
			if strings.Contains(version, "easykube") {
				Kube.FmtYellow("%s %s", name, version)
			} else {
				Kube.FmtGreen("✓ %s %s", name, version)
			}
		}
	}

	Kube.FmtGreen("Inspecting binary dependencies")
	hasBinary("kubectl", s.getKubectlVersion)
	hasBinary("docker", s.getDockerVersion)
	hasBinary("helm", s.getHelmVersion)
	hasBinary("kustomize", s.getKustomizeVersion)

	return nil
}

func (s *StatusBuilderImpl) DoAddonRepositoryCheck() error {

	Kube.FmtGreen("Repository configuration")

	addons, aerr := Kube.GetAddons()
	if aerr != nil {
		Kube.FmtRed(aerr.Error())
	}

	cfg, _ := Kube.LoadConfig()

	na := len(addons)
	if _, err := os.Stat(cfg.AddonDir); err == nil {
		if na == 0 {
			Kube.FmtYellow("⚠ %d addons discovered, check if '%s' is an addon repository", na, cfg.AddonDir)
		} else {
			Kube.FmtGreen("✓ %d addons discovered at '%s'", na, cfg.AddonDir)
		}

		msg, err := Kube.CheckAddonCompatibility()
		if err != nil {
			Kube.FmtRed("⚠ %s", err.Error())
		}
		if msg != "" {
			Kube.FmtGreen("✓ %s", msg)
		}

	} else {
		Kube.FmtRed("⚠ addon directory '%s' does not exist, check your config", cfg.AddonDir)
	}

	return nil
}

func (s *StatusBuilderImpl) getKubectlVersion() string {
	out, _, err := Kube.RunCommand("kubectl", []string{"version", "--client"}...)
	return s.getVersionStr(out, constants.KUBECTL_SEMVER, err)
}

func (s *StatusBuilderImpl) getDockerVersion() string {
	out, _, err := Kube.RunCommand("docker", []string{"version", "--format", "{{.Server.Version}}"}...)
	return s.getVersionStr(out, constants.DOCKER_SEMVER, err)
}

func (s *StatusBuilderImpl) getHelmVersion() string {
	out, _, err := Kube.RunCommand("helm", []string{"version", "--template", "{{.Version}}"}...)
	return s.getVersionStr(out, constants.HELM_SEMVER, err)
}

func (s *StatusBuilderImpl) getKustomizeVersion() string {
	out, _, err := Kube.RunCommand("kustomize", []string{"version"}...)
	return s.getVersionStr(out, constants.KUSTOMIZE_SEMVER, err)
}

func (s *StatusBuilderImpl) getVersionStr(in, wants string, inErr error) string {
	if inErr != nil {
		return "?"
	}

	v, err := s.VersionUtils.ExtractVersion(in)
	semv, _ := semver.NewConstraint(wants)

	if err != nil {
		return "?"
	} else {

		if !semv.Check(v) {
			return fmt.Sprintf("(easykube want %s, actual is %s)", semv.String(), v.String())
		}

		return v.String()
	}
}
