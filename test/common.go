package test

import (
	"testing"

	mock_ez "github.com/torloejborg/easykube/mock"
	"go.uber.org/mock/gomock"
)

func CreateOsDetailsMock(t *testing.T) *mock_ez.MockOsDetails {
	ctrl := gomock.NewController(t)
	osd := mock_ez.NewMockOsDetails(ctrl)
	osd.EXPECT().GetUserConfigDir().Return("/home/some-user/.config", nil).AnyTimes()
	osd.EXPECT().GetUserHomeDir().Return("/home/some-user", nil).AnyTimes()

	return osd
}
