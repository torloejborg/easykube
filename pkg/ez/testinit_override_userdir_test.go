package ez

type EasykubeConfigStub struct {
	IEasykubeConfig
}
type OsDetailsStub struct {
	OsDetails
}

func (o *OsDetailsStub) GetUserConfigDir() (string, error) {
	return "/home/some-user/.config", nil
}

func (o *OsDetailsStub) GetUserHomeDir() (string, error) {
	return "/home/some-user", nil
}
