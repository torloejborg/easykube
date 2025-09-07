package ez

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/torloejborg/easykube/ekctx"
	"github.com/torloejborg/easykube/pkg/constants"
)

type ExternalToolsImpl struct {
	ctx *ekctx.EKContext
}

type IExternalTools interface {
	KustomizeBuild(dir string) string
	ApplyYaml(yamlFile string)
	DeleteYaml(yamlFile string)
	EnsureLocalContext()
	SwitchContext(name string)
	RunCommand(name string, args ...string) (stdout string, stderr string, err error)
}

func NewExternalTools(context *ekctx.EKContext) IExternalTools {
	return &ExternalToolsImpl{
		ctx: context,
	}
}

func (et *ExternalToolsImpl) KustomizeBuild(dir string) string {

	out := et.ctx.Printer

	kustomize := exec.Command(
		"kustomize",
		"build",
		"-enable-helm",
		"--enable-alpha-plugins",
		"--enable-exec", dir)

	var stdout, stderr bytes.Buffer
	kustomize.Stdout = &stdout
	kustomize.Stderr = &stderr

	var err = kustomize.Run()
	if err != nil {
		out.FmtRed("kustomize failed with %s", string(stderr.Bytes()))
		panic(err)
	} else {
		// save output to file
		f, err := os.Create(constants.KUSTOMIZE_TARGET_OUTPUT)
		if err != nil {
			panic(err)
		}

		_, err = f.Write(stdout.Bytes())
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
	kubectl := exec.Command("kubectl", "apply", "-f", yamlFile)
	var stdout, stderr bytes.Buffer
	kubectl.Stdout = &stdout
	kubectl.Stderr = &stderr

	err := kubectl.Run()

	if err != nil {
		et.ctx.Printer.FmtRed("kubectl failed with %s", string(stderr.Bytes()))
		os.Exit(-1)
	}

}

func (et *ExternalToolsImpl) DeleteYaml(yamlFile string) {
	kubectl := exec.Command("kubectl", "delete", "-f", yamlFile)
	var stdout, stderr bytes.Buffer
	kubectl.Stdout = &stdout
	kubectl.Stderr = &stderr

	err := kubectl.Run()
	if err != nil {
		et.ctx.Printer.FmtRed("kubectl failed with %s", string(stderr.Bytes()))
		os.Exit(-1)
	}

}

func (et *ExternalToolsImpl) EnsureLocalContext() {
	kubectl := exec.Command("kubectl", "config", "use-context", constants.CLUSTER_CONTEXT)
	out := et.ctx.Printer
	var stdout, stderr bytes.Buffer
	kubectl.Stdout = &stdout
	kubectl.Stderr = &stderr

	if len(os.Getenv("KUBECONFIG")) == 0 {
		out.FmtGreen("Please configure the KUBECONFIG environment variable to include .kube/easykube configuration file")
		fmt.Println()
		out.FmtGreen("https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/#the-kubeconfig-environment-variable")
		fmt.Println()
		out.FmtYellow("(The cluster is running, but you cannot manage it yet)")
		home, _ := os.UserHomeDir()
		_ = os.Setenv("KUBECONFIG", filepath.Join(home, ".kube", constants.CLUSTER_NAME))
	} else {
		err := kubectl.Run()
		if err != nil {
			et.ctx.Printer.FmtRed("kubectl failed with %s", string(stderr.Bytes()))
			os.Exit(-1)
		}
	}
}

func (et *ExternalToolsImpl) SwitchContext(name string) {
	out := et.ctx.Printer
	kubectl := exec.Command("kubectl", "config", "use-context", name)
	var stdout, stderr bytes.Buffer
	kubectl.Stdout = &stdout
	kubectl.Stderr = &stderr

	err := kubectl.Run()
	if err != nil {
		et.ctx.Printer.FmtRed("kubectl failed with %s", string(stderr.Bytes()))
		os.Exit(-1)
	}
	out.FmtYellow("operating in context '%s'", name)
}

func (et *ExternalToolsImpl) RunCommand(name string, args ...string) (stdout string, stderr string, err error) {
	var outBuf, errBuf bytes.Buffer

	cmd := exec.Command(name, args...)
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err = cmd.Run()
	return outBuf.String(), errBuf.String(), err
}
