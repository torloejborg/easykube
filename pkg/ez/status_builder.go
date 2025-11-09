package ez

import (
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/textutils"
	"os"
	"os/exec"
	"strings"
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
	getPodmanVersion() string
	getVersionStr(in, wants string, inErr error) string
}

type binaryCheckStatus struct {
	HasVersionMismatch bool
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

	checkBinary := func(name string, vFunc func() string) binaryCheckStatus {
		_, err := exec.LookPath(name)
		if err != nil {
			Kube.FmtRed("⚠ " + name)
			return binaryCheckStatus{HasVersionMismatch: false}
		} else {

			version := vFunc()
			if strings.Contains(version, "easykube") {
				Kube.FmtYellow("%s %s", name, version)
				return binaryCheckStatus{HasVersionMismatch: true}
			} else {
				Kube.FmtGreen("✓ %s %s", name, version)
				return binaryCheckStatus{HasVersionMismatch: false}
			}
		}
	}

	cfg, err := Kube.IEasykubeConfig.LoadConfig()
	if err != nil {
		return err
	}

	Kube.FmtGreen("Inspecting binary dependencies")
	versionCheck := make([]binaryCheckStatus, 0)

	runtime := cfg.ContainerRuntime

	if runtime == "docker" {
		versionCheck = append(versionCheck, checkBinary("docker", s.getDockerVersion))
	}

	if runtime == "podman" {
		versionCheck = append(versionCheck, checkBinary("podman", s.getPodmanVersion))
	}

	versionCheck = append(versionCheck, checkBinary("kubectl", s.getKubectlVersion))
	versionCheck = append(versionCheck, checkBinary("helm", s.getHelmVersion))
	versionCheck = append(versionCheck, checkBinary("kustomize", s.getKustomizeVersion))

	for i := range versionCheck {
		if versionCheck[i].HasVersionMismatch {

			msg := `
			|Attention
			|  
			|  One or more binary dependencies did not meet their requirements, dont panic. Things still might work 
			|  as expected, however, if strange or unexpected errors occur - this could be a reason.
			|
  			|  Your mileage may vary :)
			`
			Kube.FmtYellow(textutils.TrimMargin(msg, "|"))
			break
		}
	}

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

func (s *StatusBuilderImpl) getPodmanVersion() string {
	out, _, err := Kube.RunCommand("/usr/bin/podman", []string{"version", "--format", " {{.Version}}"}...)
	return s.getVersionStr(out, constants.PODMAN_SEMVER, err)
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

		return fmt.Sprintf("%s (%s)", v.String(), semv.String())
	}
}
