package ez

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/Masterminds/semver/v3"
)

type VersionUtils struct {
	SemVerRegex string
}

func NewVersionUtils() VersionUtils {
	return VersionUtils{
		SemVerRegex: `([~^]|>=|=)?\s*(\d+(?:\.\d+){0,2})`,
	}
}

func (v *VersionUtils) IsCompatible(versionIn, constraintIn string) (bool, error) {
	version, err := v.ExtractVersion(versionIn)

	if err != nil {
		return false, err
	}

	constraint, err := v.ExtractConstraint(constraintIn)
	if err != nil {
		return false, err
	}

	return constraint.Check(version), nil

}

func (v *VersionUtils) ExtractVersion(source string) (*semver.Version, error) {

	re := regexp.MustCompile(v.SemVerRegex)
	match := re.FindStringSubmatch(source)

	if len(match) == 3 {
		op := match[1]
		version := match[2]
		if op == "" {
			v, e := semver.NewVersion(version)
			return v, e
		}
	}

	return nil, errors.New("no version found")
}

func (v *VersionUtils) ExtractConstraint(source string) (*semver.Constraints, error) {

	re := regexp.MustCompile(v.SemVerRegex)
	match := re.FindStringSubmatch(source)

	if len(match) == 3 {
		op := match[1]
		version := match[2]
		if op != "" {
			c, e := semver.NewConstraint(fmt.Sprintf("%s%s", op, version))
			return c, e
		}
	}

	return nil, errors.New("no semver constraint found")
}
