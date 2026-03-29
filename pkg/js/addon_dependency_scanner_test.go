package jsutils_test

import (
	"testing"

	mock_ez "github.com/torloejborg/easykube/mock"
	"github.com/torloejborg/easykube/pkg/ez"
	jsutils "github.com/torloejborg/easykube/pkg/js"
	"go.uber.org/mock/gomock"
)

func TestAddonNoopScan(t *testing.T) {

	ctl := gomock.NewController(t)
	SetUpJsTestEnvironment(t, ctl)

	mockCommand := mock_ez.NewMockICobraCommandHelper(ctl)
	ez.Kube.ICobraCommandHelper = mockCommand

	script := `

		let configuration = {
			dependsOn : ["a"]
		}

		if(easykube.kv("hello") === "world") {
			console.info("hi!")
		} else {
			console.info("bye")
		}

		if(easykube.kv("foo") === "bar") {
			console.info("foo=="+easykube.kv("foo"))
		} else {
			console.info("bye")
		}

    `

	mock := CreateSyntheticAddon(script, ctl)
	jsu := jsutils.NewJsUtils(mockCommand, mock, true)
	err := jsu.ExecAddonScript(mock)

	if err != nil {
		t.Fatal(err)
	}
}
