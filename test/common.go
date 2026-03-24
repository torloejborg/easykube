package test

import (
	"testing"

	mock_ez "github.com/torloejborg/easykube/mock"
	"go.uber.org/mock/gomock"
)

func CreateOsDetailsMock(t *testing.T) *mock_ez.MockOsDetails {
	ctrl := gomock.NewController(t)
	osd := mock_ez.NewMockOsDetails(ctrl)
	osd.EXPECT().GetEasykubeConfigDir().Return("/home/some-user/.config/easykube", nil).AnyTimes()
	osd.EXPECT().GetUserHomeDir().Return("/home/some-user", nil).AnyTimes()

	return osd
}
