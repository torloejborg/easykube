package test

import (
	"fmt"
	"slices"
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
	slices.Reverse(sorted)
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

func TestDiamondGraph(t *testing.T) {
	graph := ek.NewGraph()

	a := &ek.Addon{Name: "a"}
	b := &ek.Addon{Name: "b"}
	c := &ek.Addon{Name: "c"}
	d := &ek.Addon{Name: "d"}

	graph.Nodes = append(graph.Nodes, a, b, c, d)

	_ = graph.AddEdge(a, b)
	graph.AddEdge(a, c)
	graph.AddEdge(b, d)
	graph.AddEdge(c, d)

	sorted, err := graph.TopologicalSort()
	slices.Reverse(sorted)

	if err != nil {
		panic(err)
	}

	for _, n := range sorted {
		fmt.Println(n.Name)
	}

}
