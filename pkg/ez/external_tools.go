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

	stdout, stderr, err := et.RunCommand(
		"kustomize",
		"build",
		"-enable-helm",
		"--enable-alpha-plugins",
		"--enable-exec", dir)

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

	_, stderr, err := et.RunCommand("kubectl", "apply", "-f", yamlFile)

	if err != nil {
		Kube.FmtRed("kubectl failed with %s", stderr)
		os.Exit(-1)
	}

}

func (et *ExternalToolsImpl) DeleteYaml(yamlFile string) {
	_, stderr, err := et.RunCommand("kubectl", "delete", "-f", yamlFile)
	if err != nil {
		Kube.FmtRed("kubectl failed with %s", stderr)
		os.Exit(-1)
	}

}

func (et *ExternalToolsImpl) EnsureLocalContext() {

	if len(os.Getenv("KUBECONFIG")) == 0 {
		Kube.FmtGreen("Please configure the KUBECONFIG environment variable to include .kube/easykube configuration file")
		fmt.Println()
		Kube.FmtGreen("https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/#the-kubeconfig-environment-variable")
		fmt.Println()
		Kube.FmtYellow("(The cluster is running, but you cannot manage it yet)")
		home, _ := os.UserHomeDir()
		_ = os.Setenv("KUBECONFIG", filepath.Join(home, ".kube", constants.CLUSTER_NAME))
	} else {
		_, stderr, err := et.RunCommand("kubectl", "config", "use-context", constants.CLUSTER_CONTEXT)

		if err != nil {
			Kube.FmtRed("kubectl failed with %s", stderr)
			os.Exit(-1)
		}
	}
}

func (et *ExternalToolsImpl) SwitchContext(name string) {

	_, stderr, err := et.RunCommand("kubectl", "config", "use-context", name)

	if err != nil {
		Kube.FmtRed("kubectl failed with %s", stderr)
		os.Exit(-1)
	}
	Kube.FmtYellow("operating in context '%s'", name)
}

func (et *ExternalToolsImpl) RunCommand(name string, args ...string) (stdout string, stderr string, err error) {
	var outBuf, errBuf bytes.Buffer

	if Kube.IsVerbose() {
		Kube.FmtVerbose("%s %s", name, strings.Join(args, " "))
	}

	cmd := exec.Command(name, args...)
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err = cmd.Run()
	return outBuf.String(), errBuf.String(), err
}
