/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"os"

	"github.com/torloejborg/easykube/cmd"
	"github.com/torloejborg/easykube/pkg/ez"
)

func main() {
	ez.Kube = &ez.EasykubeSingleton{}
	err := ez.InitializeKubeSingleton()
	if err != nil {
		ez.Kube.FmtRed(err.Error())
		os.Exit(1)
	}

	_ = ez.Kube.IEasykubeConfig.PatchConfig()

	cmd.Execute()
}
