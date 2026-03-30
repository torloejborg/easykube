package cmd

import (
	"errors"

	"github.com/torloejborg/easykube/pkg/core"
	"github.com/torloejborg/easykube/pkg/ez"
)

type SkaffoldOpts struct {
	AddonName     string
	AddonLocation string
}

func skaffoldActual(opts SkaffoldOpts, ek *core.Ek) error {

	ekc, err := ek.Config.LoadConfig()
	if nil != err {
		return errors.New("cannot proceed without easykube configuration")
	}

	skaf := ez.NewSkaffold(ek, ekc.AddonDir)
	skaf.CreateNewAddon(opts.AddonName, opts.AddonLocation)

	ek.Printer.FmtGreen("addon '%s' created in '%s'", opts.AddonName, opts.AddonLocation)

	return nil
}
