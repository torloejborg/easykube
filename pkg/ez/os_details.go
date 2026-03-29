package ez

import (
	"os"
	"path/filepath"

	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/core"
)

type OsDetailsImpl struct {
	Ek *core.Ek
}

func (d OsDetailsImpl) GetEasykubeConfigDir() (string, error) {

	// setting config dir from environment takes precedence
	value, present := os.LookupEnv(constants.EnvEasykubeConfigDir)
	if present {
		return value, nil
	}

	// allow user to override default configuration directory with program argument
	if d.Ek.CommandContext.GetStringFlag(constants.FlagConfigDir) != "" {
		return d.Ek.CommandContext.GetStringFlag(constants.FlagConfigDir), nil
	} else {
		r, err := os.UserConfigDir()
		if err != nil {
			panic(err)
		}

		r = filepath.Join(r, "easykube")

		return r, nil
	}
}

func (d OsDetailsImpl) GetUserHomeDir() (string, error) {
	r, err := os.UserHomeDir()

	if err != nil {
		return "", err
	}
	return r, nil
}
