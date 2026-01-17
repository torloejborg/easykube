/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"log"

	"github.com/torloejborg/easykube/cmd"
	"github.com/torloejborg/easykube/pkg/ez"
)

func main() {
	ez.Kube = &ez.EasykubeSingleton{}
	err := ez.InitializeKubeSingleton()
	if err != nil {
		log.Fatal(err)
	}

	_ = ez.Kube.IEasykubeConfig.PatchConfig()

	cmd.Execute()
}
