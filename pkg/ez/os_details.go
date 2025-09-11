package ez

import "os"

type OsDetails interface {
	GetUserConfigDir() (string, error)
	GetUserHomeDir() (string, error)
}

type OsDetailsImpl struct{}

func (d *OsDetailsImpl) GetUserConfigDir() (string, error) {
	r, err := os.UserConfigDir()

	if err != nil {
		return "", err
	}
	return r, nil
}

func (d *OsDetailsImpl) GetUserHomeDir() (string, error) {
	r, err := os.UserHomeDir()

	if err != nil {
		return "", err
	}
	return r, nil
}
