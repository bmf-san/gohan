package processor

import (
	"fmt"

	"github.com/bmf-san/gohan/internal/model"
)

// addNode inserts a node into the graph if it does not already exist.
func addNode(g *model.DependencyGraph, node *model.Node) {
	if _, exists := g.Nodes[node.Path]; !exists {
		g.Nodes[node.Path] = node
	}
}

// addEdge records a directed edge from → to in the graph.
func addEdge(g *model.DependencyGraph, from, to string) {
	g.Edges[from] = appendUnique(g.Edges[from], to)
	// Update dependents on the target node.
	if n, ok := g.Nodes[to]; ok {
		n.Dependents = appendUnique(n.Dependents, from)
	}
	// Update dependencies on the source node.
	if n, ok := g.Nodes[from]; ok {
		n.Dependencies = appendUnique(n.Dependencies, to)
	}
}

// CalculateImpact returns all node paths that are transitively impacted when
// changedPath changes (i.e., changedPath plus all its transitive dependents).
func CalculateImpact(g *model.DependencyGraph, changedPath string) []string {
	visited := make(map[string]bool)
	result := traverseDependents(g, changedPath, visited)
	return result
}

// traverseDependents recursively collects all dependents of path.
func traverseDependents(g *model.DependencyGraph, path string, visited map[string]bool) []string {
	if visited[path] {
		return nil
	}
	visited[path] = true
	result := []string{path}

	node, ok := g.Nodes[path]
	if !ok {
		return result
	}
	for _, dep := range node.Dependents {
		result = append(result, traverseDependents(g, dep, visited)...)
	}
	return result
}

// CalculateDiff compares two dependency graphs and returns a ChangeSet
// describing added, deleted, and modified nodes.
func CalculateDiff(oldGraph, newGraph *model.DependencyGraph) (*model.ChangeSet, error) {
	if oldGraph == nil || newGraph == nil {
		return nil, fmt.Errorf("processor: graph must not be nil")
	}
	cs := &model.ChangeSet{}
	for path := range newGraph.Nodes {
		if _, exists := oldGraph.Nodes[path]; !exists {
			cs.AddedFiles = append(cs.AddedFiles, path)
		} else {
			cs.ModifiedFiles = append(cs.ModifiedFiles, path)
		}
	}
	for path := range oldGraph.Nodes {
		if _, exists := newGraph.Nodes[path]; !exists {
			cs.DeletedFiles = append(cs.DeletedFiles, path)
		}
	}
	return cs, nil
}

// appendUnique appends s to slice only if it is not already present.
func appendUnique(slice []string, s string) []string {
	for _, v := range slice {
		if v == s {
			return slice
		}
	}
	return append(slice, s)
}
