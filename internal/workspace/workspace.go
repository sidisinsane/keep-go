// Package workspace discovers and loads Markdown documents from a keep workspace.
package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sidisinsane/keep-go/internal/config"
	"github.com/sidisinsane/keep-go/internal/document"
)

// Load walks the workspace and parses all Markdown documents.
func Load(cfg *config.Config) ([]document.ParseResult, error) {
	var results []document.ParseResult
	var files []string

	err := filepath.Walk(cfg.WorkspaceRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}
		rel, err := filepath.Rel(cfg.WorkspaceRoot, path)
		if err != nil {
			return err
		}
		if strings.Contains(rel, ".keep") {
			return nil
		}
		if path == cfg.IndexPath {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(files)

	for _, path := range files {
		result, err := document.Parse(path, cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", path, err)
			continue
		}
		results = append(results, result)
	}

	return results, nil
}
