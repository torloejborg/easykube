package cmd

import (
	"errors"

	"github.com/torloejborg/easykube/pkg/ez"
)

type SkaffoldOpts struct {
	AddonName     string
	AddonLocation string
}

func skaffoldActual(opts SkaffoldOpts) error {

	ezk := ez.Kube

	ekc, err := ez.Kube.LoadConfig()
	if nil != err {
		return errors.New("cannot proceed without easykube configuration")
	}

	skaf := ez.NewSkaffold(ekc.AddonDir)
	skaf.CreateNewAddon(opts.AddonName, opts.AddonLocation)

	ezk.FmtGreen("addon '%s' created in '%s'", opts.AddonName, opts.AddonLocation)

	return nil
}
