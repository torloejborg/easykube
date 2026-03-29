package core

type IStatusBuilder interface {
	DoContainerCheck() error
	DoBinaryCheck() error
	DoAddonRepositoryCheck() error
	GetDockerVersion() string
	GetHelmVersion() string
	GetKubectlVersion() string
	GetKustomizeVersion() string
	GetPodmanVersion() string
	GetVersionStr(in, wants string, inErr error) string
}
