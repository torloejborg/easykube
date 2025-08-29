package test

import (
	"fmt"
	"testing"

	"github.com/torloj/easykube/pkg/ek"
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

	_ = graph.AddEdge(postgres, storage)
	graph.AddEdge(jenkins, storage)
	graph.AddEdge(jenkins, ingress)

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

	g, err := ek.BuildDependencyGraph(addons["a"], addons)

	if err != nil {
		panic(err)
	}

	l, tserr := g.TopologicalSort()
	if tserr != nil {
		panic(tserr)
	}

	for _, addon := range l {
		fmt.Println(addon.Name)
	}

}
