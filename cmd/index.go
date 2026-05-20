package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/sidisinsane/keep-go/internal/config"
	"github.com/sidisinsane/keep-go/internal/index"
	"github.com/sidisinsane/keep-go/internal/workspace"
)

func runIndex(ws string) error {
	cfg, err := config.Load(ws)
	if err != nil {
		return err
	}
	results, err := workspace.Load(cfg)
	if err != nil {
		return err
	}
	rows := index.Build(results)
	if err := index.Write(rows, cfg.IndexPath); err != nil {
		return err
	}
	rel, _ := filepath.Rel(cfg.WorkspaceRoot, cfg.IndexPath)
	fmt.Printf("index written to %s (%d document(s))\n", rel, len(rows))
	return nil
}
