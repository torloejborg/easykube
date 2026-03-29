package ez

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/core"
)

type ExternalToolsImpl struct {
	ek *core.Ek
}

func NewExternalTools(ek *core.Ek) core.IExternalTools {
	return &ExternalToolsImpl{ek: ek}
}

func (eti ExternalToolsImpl) KustomizeBuild(dir string) string {

	cmd := "kustomize"
	args := []string{
		"build",
		"-enable-helm",
		"--enable-alpha-plugins",
		"--enable-exec", dir}

	outCmd := fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))

	if eti.ek.CommandContext.IsVerbose() {
		eti.ek.Printer.FmtVerbose(outCmd)
	}

	if eti.ek.CommandContext.IsDryRun() {
		eti.ek.Printer.FmtDryRun(outCmd)
	} else {

		stdout, stderr, err := eti.RunCommand(cmd, args...)

		if err != nil {
			eti.ek.Printer.FmtRed("kustomize failed with %s", stderr)
			panic(err)
		} else {
			// save output to file
			f, err := os.Create(filepath.Join(dir, constants.KustomizeTargetOutput))
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
	}

	return filepath.Join(dir, constants.KustomizeTargetOutput)
}

func (eti ExternalToolsImpl) ApplyYaml(yamlFile string) {

	cmd := "kubectl"
	args := []string{"apply", "-f", yamlFile}
	outCmd := fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))

	if eti.ek.CommandContext.IsDryRun() {
		eti.ek.Printer.FmtDryRun(outCmd)
	} else {
		if eti.ek.CommandContext.IsVerbose() {
			eti.ek.Printer.FmtVerbose(outCmd)
		}
		_, stderr, err := eti.RunCommand(cmd, args...)

		if err != nil {
			eti.ek.Printer.FmtRed("kubectl failed with %s", stderr)
			os.Exit(-1)
		}
	}

}

func (eti ExternalToolsImpl) DeleteYaml(yamlFile string) {

	cmd := "kubectl"
	args := []string{"delete", "-f", yamlFile}
	cmdStr := fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))

	if eti.ek.CommandContext.IsVerbose() {
		eti.ek.Printer.FmtVerbose(cmdStr)
	}

	if eti.ek.CommandContext.IsDryRun() {
		eti.ek.Printer.FmtDryRun(cmdStr)
	} else {
		_, stderr, err := eti.RunCommand(cmd, args...)
		if err != nil {
			eti.ek.Printer.FmtRed("kubectl failed with %s", stderr)
			os.Exit(-1)
		}
	}

}

func (eti *ExternalToolsImpl) EnsureLocalContext() {
	if len(os.Getenv("KUBECONFIG")) == 0 {
		eti.ek.Printer.FmtGreen("Please configure the KUBECONFIG environment variable to include .kube/easykube configuration file")
		fmt.Println()
		eti.ek.Printer.FmtGreen("https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/#the-kubeconfig-environment-variable")
		fmt.Println()
		eti.ek.Printer.FmtYellow("(The cluster is running, but you cannot manage it yet)")

		home, _ := os.UserHomeDir()
		_ = os.Setenv("KUBECONFIG", filepath.Join(home, ".kube", constants.ClusterName))
	} else {

		cmd := "kubectl"
		args := []string{"config", "use-context", constants.ClusterContext}
		cmdStr := fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))
		if eti.ek.CommandContext.IsDryRun() {
			eti.ek.Printer.FmtDryRun(cmdStr)
		} else {
			if eti.ek.CommandContext.IsVerbose() {
				eti.ek.Printer.FmtVerbose(cmdStr)
			}
			_, stderr, err := eti.RunCommand(cmd, args...)

			if err != nil {
				eti.ek.Printer.FmtRed("kubectl failed with %s", stderr)
				os.Exit(-1)
			}
		}
	}
}

func (eti *ExternalToolsImpl) SwitchContext(name string) {
	cmd := "kubectl"
	args := []string{"config", "use-context", name}
	cmdStr := fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))

	if eti.ek.CommandContext.IsDryRun() {
		eti.ek.Printer.FmtDryRun(cmdStr)
	} else {

		if eti.ek.CommandContext.IsVerbose() {
			eti.ek.Printer.FmtVerbose(cmdStr)
		}
		_, stderr, err := eti.RunCommand(cmd, args...)

		if err != nil {
			eti.ek.Printer.FmtRed("kubectl failed with %s", stderr)
			os.Exit(-1)
		}
		eti.ek.Printer.FmtYellow("operating in context '%s'", name)
	}
}

func (eti *ExternalToolsImpl) RunCommand(name string, args ...string) (stdout string, stderr string, err error) {
	var outBuf, errBuf bytes.Buffer

	cmd := exec.Command(name, args...)
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err = cmd.Run()

	cmd.Process.Wait()
	return outBuf.String(), errBuf.String(), err
}
