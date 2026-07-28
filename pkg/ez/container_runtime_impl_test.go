package ez_test

import (
	"testing"

	"github.com/torloejborg/easykube/pkg/core"
	"github.com/torloejborg/easykube/pkg/ez"
	"github.com/torloejborg/easykube/test"
)

func initContainerRuntimeTest(t *testing.T) *core.Ek {
	osd := test.CreateOsDetailsMock(t)
	osd.EXPECT().GetEasykubeConfigDir().Return("/home/some-user/.config", nil).AnyTimes()
	osd.EXPECT().GetUserHomeDir().Return("/home/some-user", nil).AnyTimes()

	ek := &core.Ek{
		OsDetails: osd,
	}

	ek.Utils = ez.NewUtils(ek)
	ek.Config = ez.NewEasykubeConfig(ek)
	_ = ek.Config.MakeConfig()

	return ek
}
