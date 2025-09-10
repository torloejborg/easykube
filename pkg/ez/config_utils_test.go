package ez

import (
	"testing"

	"github.com/spf13/afero"
)

type EasykubeConfigMock struct {
	IEasykubeConfig
}

//func (ec *EasykubeConfigMock) MakeConfig() {
//	fmt.Println("MakeConfig")
//}

func init() {

	x := &EasykubeConfigMock{}
	Kube.UseFilesystemLayer(afero.NewMemMapFs())
	Kube.UseEasykubeConfig(x)

}

func TestMakeDefaultConfig(t *testing.T) {
	Kube.IEasykubeConfig.MakeConfig()
}
