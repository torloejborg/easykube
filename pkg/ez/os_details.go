package ez

import (
	"os"
	"path/filepath"

	"github.com/torloejborg/easykube/pkg/constants"
)

type OsDetails interface {
	GetEasykubeConfigDir() (string, error)
	GetUserHomeDir() (string, error)
}

type OsDetailsImpl struct {
}

func (d *OsDetailsImpl) GetEasykubeConfigDir() (string, error) {

	// setting config dir from environment takes precedence
	value, present := os.LookupEnv(constants.EnvEasykubeConfigDir)
	if present {
		return value, nil
	}

	// allow user to override default configuration directory with program argument
	if Kube.GetStringFlag(constants.FlagConfigDir) != "" {
		return Kube.GetStringFlag(constants.FlagConfigDir), nil
	} else {
		r, err := os.UserConfigDir()
		if err != nil {
			panic(err)
		}

		r = filepath.Join(r, "easykube")

		return r, nil
	}
}

func (d *OsDetailsImpl) GetUserHomeDir() (string, error) {
	r, err := os.UserHomeDir()

	if err != nil {
		return "", err
	}
	return r, nil
}
