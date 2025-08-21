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

func (g *Graph) SetNodeList(addons []*Addon) {
	g.Nodes = addons
}

func (g *Graph) AppendNode(v *Addon) {
	g.Nodes = append(g.Nodes, v)
}

func (g *Graph) AddNode(u, v *Addon) {
	if _, ok := g.adj[u]; !ok {
		g.adj[u] = make([]*Addon, 0)
	}
	g.adj[u] = append(g.adj[u], v)
	g.inDegree[v]++
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
		// Pop the first element
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

	// If visited nodes != total nodes, there is a cycle
	if visited != len(g.Nodes) {
		cycle := g.findCycle()
		return nil, fmt.Errorf("cycle detected in Graph involving addons: %v", formatAddonNames(cycle))
	}

	return result, nil
}

func (g *Graph) AddDependency(A, B *Addon) error {
	// Check if adding this dependency creates a cycle
	// Temporarily add the edge
	g.adj[A] = append(g.adj[A], B)
	g.inDegree[B]++

	// Check for cycles
	if err := g.checkCycle(); err != nil {
		// If a cycle is detected, revert the change
		g.adj[A] = g.adj[A][:len(g.adj[A])-1]
		g.inDegree[B]--
		return err
	}

	// If no cycle, we are good to go
	return nil
}

// checkCycle checks for cycles in the Graph using DFS.
func (g *Graph) checkCycle() error {
	visited := map[*Addon]bool{}
	recStack := map[*Addon]bool{}
	var cycle []*Addon

	// Helper DFS function to detect cycles and build the cycle path
	var dfs func(node *Addon) bool
	dfs = func(node *Addon) bool {
		if recStack[node] {
			// Cycle detected, add to the cycle path
			cycle = append(cycle, node)
			return true
		}
		if visited[node] {
			return false // already visited
		}

		visited[node] = true
		recStack[node] = true

		for _, neighbor := range g.adj[node] {
			if dfs(neighbor) {
				// If cycle is found, add current node to the cycle path
				cycle = append(cycle, node)
				return true
			}
		}

		recStack[node] = false
		return false
	}

	// Check all nodes for cycles
	for _, node := range g.Nodes {
		if !visited[node] && dfs(node) {
			// Reverse cycle for better readability (start from the node where the cycle was found)
			slices.Reverse(cycle)
			return fmt.Errorf("cycle detected in the Graph involving addons: %v", formatAddonNames(cycle))
		}
	}

	return nil
}

// findCycle tries to find a cycle in the Graph and return the involved nodes.
// This is a DFS-based approach to help identify one cycle.
func (g *Graph) findCycle() []*Addon {
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
			break
		}
	}

	// Reverse cycle for better readability
	slices.Reverse(cycle)
	return cycle
}

func formatAddonNames(addons []*Addon) string {
	names := make([]string, 0)
	for _, a := range addons {
		names = append(names, a.ShortName)
	}

	return strings.Join(names, ", ")

}
