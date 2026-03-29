package core

type IOsDetails interface {
	GetEasykubeConfigDir() (string, error)
	GetUserHomeDir() (string, error)
}
