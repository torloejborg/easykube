package ez

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/torloejborg/easykube/pkg/constants"
)

type ExternalToolsImpl struct {
}

type IExternalTools interface {
	KustomizeBuild(dir string) string
	ApplyYaml(yamlFile string)
	DeleteYaml(yamlFile string)
	EnsureLocalContext()
	// SwitchContext Change kube context to name
	SwitchContext(name string)
	// RunCommand Runs an OS command
	RunCommand(name string, args ...string) (stdout string, stderr string, err error)
}

func NewExternalTools() IExternalTools {
	return &ExternalToolsImpl{}
}

func (et *ExternalToolsImpl) KustomizeBuild(dir string) string {

	cmd := "kustomize"
	args := []string{
		"build",
		"-enable-helm",
		"--enable-alpha-plugins",
		"--enable-exec", dir}

	outCmd := fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))
	stdout, stderr, err := et.RunCommand(cmd, args...)

	if Kube.IsVerbose() {
		Kube.FmtVerbose(outCmd)
	}

	if err != nil {
		Kube.FmtRed("kustomize failed with %s", stderr)
		panic(err)
	} else {
		// save output to file
		f, err := os.Create(constants.KUSTOMIZE_TARGET_OUTPUT)
		if err != nil {
			panic(err)
		}

		_, err = f.WriteString(stdout)
		if err != nil {
			panic(err)
		}

		defer func(f *os.File) {
			err := f.Close()
			if err != nil {
				panic(err)
			}
		}(f)
	}

	return filepath.Join(dir, constants.KUSTOMIZE_TARGET_OUTPUT)
}

func (et *ExternalToolsImpl) ApplyYaml(yamlFile string) {

	cmd := "kubectl"
	args := []string{"apply", "-f", yamlFile}
	outCmd := fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))

	if Kube.IsDryRun() {
		Kube.FmtDryRun(outCmd)
	} else {
		if Kube.IsVerbose() {
			Kube.FmtVerbose(outCmd)
		}
		_, stderr, err := et.RunCommand(cmd, args...)

		if err != nil {
			Kube.FmtRed("kubectl failed with %s", stderr)
			os.Exit(-1)
		}
	}

}

func (et *ExternalToolsImpl) DeleteYaml(yamlFile string) {

	cmd := "kubectl"
	args := []string{"delete", "-f", yamlFile}

	if Kube.IsDryRun() {
		Kube.FmtDryRun(cmd, strings.Join(args, " "))
	} else {
		_, stderr, err := et.RunCommand(cmd, args...)
		if err != nil {
			Kube.FmtRed("kubectl failed with %s", stderr)
			os.Exit(-1)
		}
	}

}

func (et *ExternalToolsImpl) EnsureLocalContext() {
	k := Kube
	if len(os.Getenv("KUBECONFIG")) == 0 {
		k.FmtGreen("Please configure the KUBECONFIG environment variable to include .kube/easykube configuration file")
		fmt.Println()
		k.FmtGreen("https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/#the-kubeconfig-environment-variable")
		fmt.Println()
		k.FmtYellow("(The cluster is running, but you cannot manage it yet)")

		home, _ := os.UserHomeDir()
		_ = os.Setenv("KUBECONFIG", filepath.Join(home, ".kube", constants.CLUSTER_NAME))
	} else {

		cmd := "kubectl"
		args := []string{"config", "use-context", constants.CLUSTER_CONTEXT}
		cmdStr := fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))
		if k.IsDryRun() && k.IsVerbose() {
			k.FmtDryRun(cmdStr)

		} else {
			if k.IsVerbose() {
				k.FmtVerbose(cmdStr)
			}
			_, stderr, err := et.RunCommand(cmd, args...)

			if err != nil {
				Kube.FmtRed("kubectl failed with %s", stderr)
				os.Exit(-1)
			}
		}
	}
}

func (et *ExternalToolsImpl) SwitchContext(name string) {
	k := Kube
	cmd := "kubectl"
	args := []string{"config", "use-context", name}
	cmdStr := fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))

	if k.IsDryRun() {
		k.FmtDryRun(cmdStr)
	} else {

		if k.IsVerbose() {
			k.FmtVerbose(cmdStr)
		}
		_, stderr, err := et.RunCommand(cmd, args...)

		if err != nil {
			Kube.FmtRed("kubectl failed with %s", stderr)
			os.Exit(-1)
		}
		Kube.FmtYellow("operating in context '%s'", name)
	}
}

func (et *ExternalToolsImpl) RunCommand(name string, args ...string) (stdout string, stderr string, err error) {
	var outBuf, errBuf bytes.Buffer

	cmd := exec.Command(name, args...)
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err = cmd.Run()
	return outBuf.String(), errBuf.String(), err
}
