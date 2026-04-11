package ez_test

import (
	"testing"

	"github.com/torloejborg/easykube/test"
)

func TestScanAddons(t *testing.T) {

	ek := initAddonReaderTest(t)

	test.CopyTestAddonToMemFs("../../test_addons", "diamond", "/home/some-user/addons", ek.Fs)

	all, err := ek.AddonReader.GetAddons()
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range expectedAddonsForDiscoverTest {
		t.Run(tt.name, func(t *testing.T) {
			if all[tt.name] == nil {
				t.Errorf("expected addon %v to be in the list", tt.name)
			}
		})
	}
}
