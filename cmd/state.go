package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/sidisinsane/keep-go/internal/config"
	"github.com/sidisinsane/keep-go/internal/state"
	"github.com/sidisinsane/keep-go/internal/workspace"
)

func runState(ws string) error {
	cfg, err := config.Load(ws)
	if err != nil {
		return err
	}
	results, err := workspace.Load(cfg)
	if err != nil {
		return err
	}
	current := state.Load(cfg.StatePath)
	updated := state.Update(current, results, cfg)
	if err := state.Write(updated, cfg.StatePath); err != nil {
		return err
	}
	rel, _ := filepath.Rel(cfg.WorkspaceRoot, cfg.StatePath)
	fmt.Printf("state written to %s (%d document(s))\n", rel, len(updated.Documents))
	return nil
}
