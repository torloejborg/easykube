package test

import (
	"github.com/torloj/easykube/pkg/ek"
	"log"
	"testing"
)

func TestTopologicalSort(t *testing.T) {
	graph := ek.NewGraph()

	ingress := &ek.Addon{Name: "nginx-controller"}
	jenkins := &ek.Addon{Name: "jenkins-lts"}
	storage := &ek.Addon{Name: "storage-provider"}
	postgres := &ek.Addon{Name: "postgresql"}

	graph.Nodes = append(graph.Nodes,
		ingress,
		jenkins,
		storage,
		postgres)

	graph.AddNode(postgres, storage)
	graph.AddNode(jenkins, storage)
	graph.AddNode(jenkins, ingress)

	sorted, err := graph.TopologicalSort()
	if err != nil {
		panic(err)
	}

	if len(sorted) == 0 {
		t.Error("cycle detected")
	} else {
		t.Log("Sorted order:")
		for _, node := range sorted {
			t.Log(node.Name)
		}
	}
}

func TestExecutionOrder(t *testing.T) {
	adr := ek.NewAddonReader(GetEKContext())
	addons := adr.GetAddons()

	getAddonDependencyList := func(name string) []*ek.Addon {
		depends := make([]*ek.Addon, 0)
		dependsOn := addons[name].Config.DependsOn

		for x := range dependsOn {
			depends = append(depends, addons[dependsOn[x]])
		}

		return depends
	}

	g := ek.NewGraph()
	// add all nodes
	nl := make([]*ek.Addon, 0)
	for _, v := range addons {
		nl = append(nl, v)
	}
	g.Nodes = nl

	for k, v := range addons {
		d := getAddonDependencyList(k)

		for x := range d {
			e := g.AddDependency(v, d[x])
			if e != nil {
				panic(e.Error())
			}
		}
	}

	result, err := g.TopologicalSort()
	if err != nil {
		panic(err)
	}

	for a := range result {
		log.Println(result[a].Name)
	}

}
