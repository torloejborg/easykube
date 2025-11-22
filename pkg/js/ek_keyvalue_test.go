package jsutils_test

import (
	"testing"

	mock_ez "github.com/torloejborg/easykube/mock"
	"github.com/torloejborg/easykube/pkg/ez"
	jsutils "github.com/torloejborg/easykube/pkg/js"
	"go.uber.org/mock/gomock"
)

func TestEasykube_KeyValue(t *testing.T) {

	ctl := gomock.NewController(t)
	SetUpJsTestEnvironment(t, ctl)

	mockCommand := mock_ez.NewMockICobraCommandHelper(ctl)
	mockCommand.EXPECT().IsDryRun().Return(false).AnyTimes()
	mockCommand.EXPECT().GetStringFlag(gomock.AnyOf("kv")).Return("hello = world, foo=bar ").AnyTimes()

	ez.Kube.ICobraCommandHelper = mockCommand

	script := `
		if(easykube.kv("hello") === "world") {
			console.info("hi!")
		} else {
			helloIsNotWorld
		}

		if(easykube.kv("foo") === "bar") {
			console.info("foo=="+easykube.kv("foo"))
		} else {
			fooIsNotBar
		}

    `

	mock := CreateSyntheticAddon(script, ctl)
	jsu := jsutils.NewJsUtils(mockCommand, mock)
	err := jsu.ExecAddonScript(mock)

	if err != nil {
		t.Fatal(err)
	}
}
