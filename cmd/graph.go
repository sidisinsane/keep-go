package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/sidisinsane/keep-go/internal/config"
	"github.com/sidisinsane/keep-go/internal/graph"
	"github.com/sidisinsane/keep-go/internal/workspace"
)

func runGraph(ws string) error {
	cfg, err := config.Load(ws)
	if err != nil {
		return err
	}
	results, err := workspace.Load(cfg)
	if err != nil {
		return err
	}
	g := graph.Build(results)
	if err := graph.Write(g, cfg.GraphPath); err != nil {
		return err
	}
	rel, _ := filepath.Rel(cfg.WorkspaceRoot, cfg.GraphPath)
	fmt.Printf("graph written to %s (%d node(s), %d edge(s))\n", rel, len(g.Nodes), len(g.Edges))
	return nil
}
