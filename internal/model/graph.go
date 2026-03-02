package model

import "time"

// NodeType classifies a node in the dependency graph.
type NodeType int

const (
	NodeTypeArticle NodeType = iota
	NodeTypeTag
	NodeTypeCategory
	NodeTypeArchive
	NodeTypePage
)

// Node is a vertex in the dependency graph.
type Node struct {
	Path         string
	Type         NodeType
	Dependencies []string
	Dependents   []string
	LastModified time.Time
}

// DependencyGraph is a directed graph of content dependencies.
type DependencyGraph struct {
	Nodes map[string]*Node
	Edges map[string][]string
}

// ChangeSet holds the result of diff detection between two builds.
type ChangeSet struct {
	ModifiedFiles []string
	AddedFiles    []string
	DeletedFiles  []string
}
