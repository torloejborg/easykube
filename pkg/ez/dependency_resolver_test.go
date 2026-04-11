package ez_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/torloejborg/easykube/pkg/core"
)

func TestTopologicalSort(t *testing.T) {
	graph := core.NewGraph[core.IAddon]()

	ingress := &core.Addon{Name: "nginx-controller"}
	jenkins := &core.Addon{Name: "jenkins-lts"}
	storage := &core.Addon{Name: "storage-provider"}
	postgres := &core.Addon{Name: "postgresql"}

	graph.Nodes = append(graph.Nodes,
		ingress,
		jenkins,
		storage,
		postgres)

	_ = graph.AddEdge(postgres, storage)
	_ = graph.AddEdge(jenkins, storage)
	_ = graph.AddEdge(jenkins, ingress)

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
			t.Log(node.GetName())
		}
	}
}

func TestDiamondGraph(t *testing.T) {
	graph := core.NewGraph[core.IAddon]()

	a := &core.Addon{Name: "a"}
	b := &core.Addon{Name: "b"}
	c := &core.Addon{Name: "c"}
	d := &core.Addon{Name: "d"}

	graph.Nodes = append(graph.Nodes, a, b, c, d)

	_ = graph.AddEdge(a, b)
	_ = graph.AddEdge(a, c)
	_ = graph.AddEdge(b, d)
	_ = graph.AddEdge(c, d)

	sorted, err := graph.TopologicalSort()
	slices.Reverse(sorted)

	if err != nil {
		panic(err)
	}

	for _, n := range sorted {
		fmt.Println(n.GetName())
	}

}
