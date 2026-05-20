// Package graph builds the workspace document graph and serialises it to JSON.
package graph

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/sidisinsane/keep-go/internal/document"
)

// Node represents a document in the graph.
type Node struct {
	Slug    string `json:"slug"`
	Title   string `json:"title"`
	Kind    string `json:"kind"`
	Status  string `json:"status"`
	Private bool   `json:"private"`
}

// Edge represents a relation between documents.
type Edge struct {
	From         string `json:"from"`
	To           string `json:"to"`
	Type         string `json:"type"`
	AutoInjected bool   `json:"auto_injected"`
}

// Graph is the full workspace graph.
type Graph struct {
	GeneratedAt string `json:"generated_at"`
	Nodes       []Node `json:"nodes"`
	Edges       []Edge `json:"edges"`
}

// Build creates a graph from parsed documents.
func Build(results []document.ParseResult) *Graph {
	g := &Graph{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339Nano),
	}
	for _, r := range results {
		if len(r.SchemaViolations) > 0 {
			continue
		}
		d := r.Document
		g.Nodes = append(g.Nodes, Node{
			Slug:    d.Slug,
			Title:   d.Title,
			Kind:    d.Kind,
			Status:  d.Status,
			Private: d.Private,
		})
		for _, rel := range d.Relations {
			g.Edges = append(g.Edges, Edge{
				From:         d.Slug,
				To:           rel.Target,
				Type:         rel.Type,
				AutoInjected: rel.AutoInjected,
			})
		}
	}
	return g
}

// Write serialises a graph to disk.
func Write(g *Graph, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}
