package core

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
