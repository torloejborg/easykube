package jsutils_test

import (
	"fmt"
	"testing"

	mock "github.com/torloejborg/easykube/mock"
	jsutils "github.com/torloejborg/easykube/pkg/js"
	"go.uber.org/mock/gomock"
)

func TestAddonConfigurationGet(t *testing.T) {

	ctl := gomock.NewController(t)
	ek := SetUpJsTestEnvironment(t, ctl)

	mockCommand := mock.NewMockICobraCommandHelper(ctl)
	mockCommand.EXPECT().IsDryRun().Return(false).AnyTimes()
	mockCommand.EXPECT().GetStringFlag(gomock.AnyOf("kv")).Return("hello = world, foo=bar ").AnyTimes()

	ek.CommandContext = mockCommand

	script := `
		let configuration = {
			description:"testing",
			dependsOn:["foo","bar"]
		}

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

	mockAddon := CreateSyntheticAddon(script, ctl)
	jsu := jsutils.NewJsUtils(ek, mockAddon, false)
	cfg, err := jsu.ExtractConfigurationObject(mockAddon)

	fmt.Println(cfg)

	if err != nil {
		t.Fatal(err)
	}
}
