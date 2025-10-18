package ez

import (
	"fmt"
	"slices"
	"strings"
)

type Graph struct {
	Nodes    []IAddon
	adj      map[IAddon][]IAddon
	inDegree map[IAddon]int
}

func NewGraph() *Graph {
	return &Graph{
		Nodes:    make([]IAddon, 0),
		adj:      make(map[IAddon][]IAddon),
		inDegree: make(map[IAddon]int),
	}
}

func (g *Graph) AppendNode(v IAddon) {
	if _, exists := g.inDegree[v]; !exists {
		g.Nodes = append(g.Nodes, v)
		g.inDegree[v] = 0
		g.adj[v] = make([]IAddon, 0)
	}
}

func (g *Graph) AddEdge(u, v IAddon) error {
	g.adj[u] = append(g.adj[u], v)
	g.inDegree[v]++
	if err := g.hasCycle(); err != nil {
		g.adj[u] = g.adj[u][:len(g.adj[u])-1]
		g.inDegree[v]--
		return err
	}
	return nil
}

func (g *Graph) hasCycle() error {
	visited := map[IAddon]bool{}
	recStack := map[IAddon]bool{}
	var cycle []IAddon
	var dfs func(node IAddon) bool
	dfs = func(node IAddon) bool {
		if recStack[node] {
			cycle = append(cycle, node)
			return true
		}
		if visited[node] {
			return false
		}
		visited[node] = true
		recStack[node] = true
		for _, neighbor := range g.adj[node] {
			if dfs(neighbor) {
				cycle = append(cycle, node)
				return true
			}
		}
		recStack[node] = false
		return false
	}
	for _, node := range g.Nodes {
		if !visited[node] && dfs(node) {
			slices.Reverse(cycle)
			return fmt.Errorf("cycle detected: %s", formatCycle(cycle))
		}
	}
	return nil
}

func formatCycle(cycle []IAddon) string {
	names := make([]string, len(cycle))
	for i, addon := range cycle {
		names[i] = addon.GetShortName()
	}
	return strings.Join(names, " -> ")
}

func (g *Graph) TopologicalSort() ([]IAddon, error) {
	result := make([]IAddon, 0)
	queue := make([]IAddon, 0)
	for _, node := range g.Nodes {
		if g.inDegree[node] == 0 {
			queue = append(queue, node)
		}
	}
	visited := 0
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)
		visited++
		for _, neighbor := range g.adj[current] {
			g.inDegree[neighbor]--
			if g.inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}
	if visited != len(g.Nodes) {
		return nil, fmt.Errorf("cycle detected in graph")
	}
	return result, nil
}

func BuildDependencyGraph(roots []IAddon, allAddons map[string]IAddon) (*Graph, error) {
	g := NewGraph()
	visited := map[string]bool{}
	var build func(IAddon) error
	build = func(addon IAddon) error {
		if visited[addon.GetShortName()] {
			return nil
		}
		visited[addon.GetShortName()] = true
		g.AppendNode(addon)
		for _, depName := range addon.GetConfig().DependsOn {
			dep, ok := allAddons[depName]
			if !ok {
				return fmt.Errorf("dependency %s not found", depName)
			}
			if err := build(dep); err != nil {
				return err
			}
			if err := g.AddEdge(dep, addon); err != nil {
				return err
			}
		}
		return nil
	}
	for _, root := range roots {
		if err := build(root); err != nil {
			return nil, err
		}
	}
	return g, nil
}

func ResolveDependencies(roots []IAddon, allAddons map[string]IAddon) ([]IAddon, error) {
	g, err := BuildDependencyGraph(roots, allAddons)
	if err != nil {
		return nil, err
	}
	return g.TopologicalSort()
}
