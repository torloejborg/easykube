package test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/torloj/easykube/pkg/ek"
)

func TestJsonConfigParsing(t *testing.T) {
	data := []byte(`
let configuration = {
	"dependsOn" : ["foo","bar"],
	"extraPorts" : [
		{
			"nodePort" : 8080,
			"hostPort" : 8080,
			"protocol" : "TCP"
		}
	],
	"extraMounts" : [
		{
			"hostPath" : "/var/run/docker.sock",
			"containerPath" : "docker.sock"	
		}
	]
	}
	`)

	cfg := &ek.AddonConfig{}
	reader := ek.NewAddonReader(GetEKContext())

	extracted, _ := reader.ExtractJSON(string(data))

	err := json.Unmarshal([]byte(extracted), &cfg)
	if err != nil {
		panic(err)
	}
}

func TestJsonConfigParsingWithOnlyDepends(t *testing.T) {
	data := []byte(`
	let configuration = {
		"dependsOn" : ["foo","bar"]
	}
	`)

	cfg := &ek.AddonConfig{}
	extracted, _ := ek.NewAddonReader(GetEKContext()).ExtractJSON(string(data))

	err := json.Unmarshal([]byte(extracted), &cfg)
	if err != nil {
		panic(err)
	}
}

func TestTroublesomeInput(t *testing.T) {
	input := `{"extraPorts": [{"nodePort": 80, "hostPort": 80, "protocol": "TCP"}]}`
	cfg := &ek.AddonConfig{}

	err := json.Unmarshal([]byte(input), &cfg)
	if err != nil {
		panic(err)
	}
}

func TestDiscoverAddons(t *testing.T) {
	addons := ek.NewAddonReader(GetEKContext()).GetAddons()

	for i := range addons {
		fmt.Println(addons[i].Config)
	}
}
