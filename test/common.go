package test

import (
	"testing"

	mock "github.com/torloejborg/easykube/mock"
	"go.uber.org/mock/gomock"
)

func CreateOsDetailsMock(t *testing.T) *mock.MockIOsDetails {
	ctrl := gomock.NewController(t)
	osd := mock.NewMockIOsDetails(ctrl)
	osd.EXPECT().GetEasykubeConfigDir().Return("/home/some-user/.config/easykube", nil).AnyTimes()
	osd.EXPECT().GetUserHomeDir().Return("/home/some-user", nil).AnyTimes()

	return osd
}
