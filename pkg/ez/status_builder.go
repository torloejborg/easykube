package ez

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/core"
	"github.com/torloejborg/easykube/pkg/textutils"
)

type StatusBuilderImpl struct {
	ek           *core.Ek
	VersionUtils VersionUtils
}

type binaryCheckStatus struct {
	HasVersionMismatch bool
}

func NewStatusBuilder(ek *core.Ek) core.IStatusBuilder {
	return &StatusBuilderImpl{
		ek:           ek,
		VersionUtils: NewVersionUtils(),
	}
}

func (s *StatusBuilderImpl) DoContainerCheck() error {

	running := func(containerID string) {

		if running, _ := s.ek.ContainerRuntime.IsContainerRunning(containerID); running {
			s.ek.Printer.FmtGreen("✓ %s container", containerID)
		} else {
			s.ek.Printer.FmtRed("⚠ %s container not running", containerID)
		}
	}

	s.ek.Printer.FmtGreen("Container configuration")
	running(constants.RegistryContainer)
	running(constants.KindContainer)

	if connected, _ := s.ek.ContainerRuntime.IsNetworkConnectedToContainer(constants.RegistryContainer, "kind"); connected {
		s.ek.Printer.FmtGreen("✓ %s connected to kind network", constants.RegistryContainer)
	} else {
		s.ek.Printer.FmtRed("⚠ %s not connected to kind network", constants.RegistryContainer)
	}

	return nil
}

func (s *StatusBuilderImpl) checkBinary(name string, vFunc func() string) binaryCheckStatus {
	_, err := exec.LookPath(name)
	if err != nil {
		s.ek.Printer.FmtRed("⚠ " + name)
		return binaryCheckStatus{HasVersionMismatch: false}
	} else {

		version := vFunc()
		if strings.Contains(version, "easykube") {
			s.ek.Printer.FmtYellow("%s %s", name, version)
			return binaryCheckStatus{HasVersionMismatch: true}
		} else {
			s.ek.Printer.FmtGreen("✓ %s %s", name, version)
			return binaryCheckStatus{HasVersionMismatch: false}
		}
	}
}

func (s *StatusBuilderImpl) DoBinaryCheck() error {
	cfg, err := s.ek.Config.LoadConfig()
	if err != nil {
		return err
	}

	s.ek.Printer.FmtGreen("Inspecting binary dependencies")
	versionCheck := make([]binaryCheckStatus, 0)

	runtime := cfg.ContainerRuntime

	if runtime == "docker" {
		if !s.ek.Utils.HasBinary("docker") {
			return errors.New("docker runtime not available")
		}
		versionCheck = append(versionCheck, s.checkBinary("docker", s.GetDockerVersion))
	}

	if runtime == "podman" {
		if !s.ek.Utils.HasBinary("podman") {
			return errors.New("podman runtime not available")
		}
		versionCheck = append(versionCheck, s.checkBinary("podman", s.GetPodmanVersion))
	}

	versionCheck = append(versionCheck, s.checkBinary("kubectl", s.GetKubectlVersion))
	versionCheck = append(versionCheck, s.checkBinary("helm", s.GetHelmVersion))
	versionCheck = append(versionCheck, s.checkBinary("kustomize", s.GetKustomizeVersion))

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
			s.ek.Printer.FmtYellow(textutils.TrimMargin(msg, "|"))
			break
		}
	}

	return nil
}

func (s *StatusBuilderImpl) DoAddonRepositoryCheck() error {

	s.ek.Printer.FmtGreen("Repository configuration")

	addons, aerr := s.ek.AddonReader.GetAddons()
	if aerr != nil {
		s.ek.Printer.FmtRed(aerr.Error())
	}

	cfg, _ := s.ek.Config.LoadConfig()

	na := len(addons)
	if _, err := os.Stat(cfg.AddonDir); err == nil {
		if na == 0 {
			s.ek.Printer.FmtYellow("⚠ %d addons discovered, check if '%s' is an addon repository", na, cfg.AddonDir)
		} else {
			s.ek.Printer.FmtGreen("✓ %d addons discovered at '%s'", na, cfg.AddonDir)
		}

		msg, err := s.ek.AddonReader.CheckAddonCompatibility()
		if err != nil {
			s.ek.Printer.FmtRed("⚠ %s", err.Error())
		}
		if msg != "" {
			s.ek.Printer.FmtGreen("✓ %s", msg)
		}

	} else {
		s.ek.Printer.FmtRed("⚠ addon directory '%s' does not exist, check your config", cfg.AddonDir)
	}

	return nil
}

func (s *StatusBuilderImpl) GetKubectlVersion() string {
	out, _, err := s.ek.ExternalTools.RunCommand("kubectl", []string{"version", "--client"}...)
	return s.GetVersionStr(out, constants.KubectlSemver, err)
}

func (s *StatusBuilderImpl) GetDockerVersion() string {
	out, _, err := s.ek.ExternalTools.RunCommand("docker", []string{"version", "--format", "{{.Server.Version}}"}...)
	return s.GetVersionStr(out, constants.DockerSemver, err)
}

func (s *StatusBuilderImpl) GetHelmVersion() string {
	out, _, err := s.ek.ExternalTools.RunCommand("helm", []string{"version", "--template", "{{.Version}}"}...)
	return s.GetVersionStr(out, constants.HelmSemver, err)
}

func (s *StatusBuilderImpl) GetKustomizeVersion() string {
	out, _, err := s.ek.ExternalTools.RunCommand("kustomize", []string{"version"}...)
	return s.GetVersionStr(out, constants.KustomizeSemver, err)
}

func (s *StatusBuilderImpl) GetPodmanVersion() string {
	out, _, err := s.ek.ExternalTools.RunCommand("podman", []string{"version", "--format", " {{.Version}}"}...)
	return s.GetVersionStr(out, constants.PodmanSemver, err)
}

func (s *StatusBuilderImpl) GetVersionStr(in, wants string, inErr error) string {
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
