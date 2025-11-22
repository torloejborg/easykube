/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"github.com/torloejborg/easykube/cmd"
	"github.com/torloejborg/easykube/pkg/ez"
)

func main() {
	ez.Kube = &ez.EasykubeSingleton{}
	ez.InitializeKubeSingleton()

	_ = ez.Kube.IEasykubeConfig.PatchConfig()

	cmd.Execute()
}
