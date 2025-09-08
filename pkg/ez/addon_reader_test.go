package ez

import (
	"encoding/json"
	"fmt"
	"testing"
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
	fmt.Println(string(data))
}

func TestJsonConfigParsingWithOnlyDepends(t *testing.T) {
	data := []byte(`
	let configuration = {
		"dependsOn" : ["foo","bar"]
	}
	`)

	fmt.Println(string(data))

}

func TestTroublesomeInput(t *testing.T) {
	input := `{"extraPorts": [{"nodePort": 80, "hostPort": 80, "protocol": "TCP"}]}`
	cfg := &AddonConfig{}

	err := json.Unmarshal([]byte(input), &cfg)
	if err != nil {
		panic(err)
	}
}

func TestDiscoverAddons(t *testing.T) {

}
