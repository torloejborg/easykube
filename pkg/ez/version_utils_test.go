package ez

import (
	"testing"
)

var versionData = []struct {
	version    string
	constraint string
	fail       bool
}{
	{"1.1.4", "~1.1.4", false},
	{"2.2.4", "~1.1.4", true},
	{"1.1.20", "~1.1.5", false},
	{"1.1.20", "~1.1", false},
	{"1.1.20", "~1", false},
	{"1.1.20", "~2", true},
}

func TestVersionUtils_ExtractVersion(t *testing.T) {
	vut := NewVersionUtils()

	for _, tt := range versionData {

		compat, err := vut.IsCompatible(tt.version, tt.constraint)

		if err != nil {
			t.Error(err)
		}

		if compat == tt.fail {
			t.Errorf("expected %s and %s to fail", tt.version, tt.constraint)
		}

	}

}
