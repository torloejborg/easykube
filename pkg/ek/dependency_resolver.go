package ek

import (
	"fmt"
	"slices"
	"strings"
)

type Graph struct {
	Nodes    []*Addon
	adj      map[*Addon][]*Addon
	inDegree map[*Addon]int
}

func NewGraph() *Graph {
	return &Graph{
		Nodes:    make([]*Addon, 0),
		adj:      make(map[*Addon][]*Addon),
		inDegree: make(map[*Addon]int),
	}
}

func (g *Graph) AppendNode(v *Addon) {
	if _, exists := g.inDegree[v]; !exists {
		g.Nodes = append(g.Nodes, v)
		g.inDegree[v] = 0
		g.adj[v] = make([]*Addon, 0)
	}
}

func (g *Graph) AddEdge(u, v *Addon) error {
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
	visited := map[*Addon]bool{}
	recStack := map[*Addon]bool{}
	var cycle []*Addon
	var dfs func(node *Addon) bool
	dfs = func(node *Addon) bool {
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

func formatCycle(cycle []*Addon) string {
	names := make([]string, len(cycle))
	for i, addon := range cycle {
		names[i] = addon.ShortName
	}
	return strings.Join(names, " -> ")
}

func (g *Graph) TopologicalSort() ([]*Addon, error) {
	result := make([]*Addon, 0)
	queue := make([]*Addon, 0)
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

func BuildDependencyGraph(roots []*Addon, allAddons map[string]*Addon) (*Graph, error) {
	g := NewGraph()
	visited := map[string]bool{}
	var build func(*Addon) error
	build = func(addon *Addon) error {
		if visited[addon.ShortName] {
			return nil
		}
		visited[addon.ShortName] = true
		g.AppendNode(addon)
		for _, depName := range addon.Config.DependsOn {
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

func ResolveDependencies(roots []*Addon, allAddons map[string]*Addon) ([]*Addon, error) {
	g, err := BuildDependencyGraph(roots, allAddons)
	if err != nil {
		return nil, err
	}
	return g.TopologicalSort()
}
