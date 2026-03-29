package core

import (
	"fmt"
	"slices"
	"strings"
)

type EzNode interface {
	GetName() string
	GetDependencies() []string
}

type Graph[T EzNode] struct {
	Nodes    []T
	adj      map[string][]T // Key: node ID (string), Value: list of dependent nodes
	inDegree map[string]int // Key: node ID (string), Value: in-degree count
	idToNode map[string]T   // Key: node ID (string), Value: node
}

func NewGraph[T EzNode]() *Graph[T] {
	return &Graph[T]{
		Nodes:    make([]T, 0),
		adj:      make(map[string][]T),
		inDegree: make(map[string]int),
		idToNode: make(map[string]T),
	}
}

func (g *Graph[T]) AppendNode(v T) {
	if _, exists := g.idToNode[v.GetName()]; !exists {
		g.Nodes = append(g.Nodes, v)
		g.inDegree[v.GetName()] = 0
		g.adj[v.GetName()] = make([]T, 0)
		g.idToNode[v.GetName()] = v
	}
}

func (g *Graph[T]) AddEdge(u, v T) error {
	uID := u.GetName()
	vID := v.GetName()
	g.adj[uID] = append(g.adj[uID], v)
	g.inDegree[vID]++
	if err := g.hasCycle(); err != nil {
		g.adj[uID] = g.adj[uID][:len(g.adj[uID])-1]
		g.inDegree[vID]--
		return err
	}
	return nil
}

func (g *Graph[T]) DependsOn(u, v T) error {
	uID := u.GetName()
	vID := v.GetName()
	g.adj[uID] = append(g.adj[uID], v)
	g.inDegree[vID]++
	if err := g.hasCycle(); err != nil {
		g.adj[uID] = g.adj[uID][:len(g.adj[uID])-1]
		g.inDegree[vID]--
		return err
	}
	return nil
}

func (g *Graph[T]) hasCycle() error {
	visited := map[string]bool{}
	recStack := map[string]bool{}
	var cycle []T
	var dfs func(node T) bool
	dfs = func(node T) bool {
		if recStack[node.GetName()] {
			cycle = append(cycle, node)
			return true
		}
		if visited[node.GetName()] {
			return false
		}
		visited[node.GetName()] = true
		recStack[node.GetName()] = true
		for _, neighbor := range g.adj[node.GetName()] {
			if dfs(neighbor) {
				cycle = append(cycle, node)
				return true
			}
		}
		recStack[node.GetName()] = false
		return false
	}
	for _, node := range g.Nodes {
		if !visited[node.GetName()] && dfs(node) {
			slices.Reverse(cycle)
			return fmt.Errorf("cycle detected: %s", g.formatCycle(cycle))
		}
	}
	return nil
}

func (g *Graph[T]) formatCycle(cycle []T) string {
	names := make([]string, len(cycle))
	for i, addon := range cycle {
		names[i] = addon.GetName()
	}
	return strings.Join(names, " -> ")
}

func (g *Graph[T]) TopologicalSort() ([]T, error) {
	result := make([]T, 0)
	queue := make([]T, 0)
	for _, node := range g.Nodes {
		if g.inDegree[node.GetName()] == 0 {
			queue = append(queue, node)
		}
	}
	visited := 0
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)
		visited++
		for _, neighbor := range g.adj[current.GetName()] {
			g.inDegree[neighbor.GetName()]--
			if g.inDegree[neighbor.GetName()] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}
	if visited != len(g.Nodes) {
		return nil, fmt.Errorf("cycle detected in graph")
	}
	return result, nil
}

func BuildDependencyGraph[T EzNode](roots []T, allNodes map[string]T) (*Graph[T], error) {
	g := NewGraph[T]() // Explicitly specify T
	visited := map[string]bool{}
	var build func(node T) error
	build = func(node T) error {
		id := node.GetName()
		if visited[id] {
			return nil
		}
		visited[id] = true
		g.AppendNode(node)
		for _, depID := range node.GetDependencies() {
			dep, ok := allNodes[depID]
			if !ok {
				return fmt.Errorf("dependency %s not found", depID)
			}
			if err := build(dep); err != nil {
				return err
			}
			if err := g.AddEdge(dep, node); err != nil {
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

func ResolveDependencies[T EzNode](roots []T, allNodes map[string]T) ([]T, error) {
	g, err := BuildDependencyGraph(roots, allNodes)
	if err != nil {
		return nil, err
	}
	return g.TopologicalSort()
}
