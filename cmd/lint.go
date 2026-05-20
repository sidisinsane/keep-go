package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/sidisinsane/keep-go/internal/config"
	"github.com/sidisinsane/keep-go/internal/lint"
	"github.com/sidisinsane/keep-go/internal/state"
	"github.com/sidisinsane/keep-go/internal/workspace"
)

func runLint(ws string) error {
	cfg, err := config.Load(ws)
	if err != nil {
		return err
	}
	results, err := workspace.Load(cfg)
	if err != nil {
		return err
	}
	s := state.Load(cfg.StatePath)
	report := lint.Run(results, s, cfg)
	lintPath := filepath.Join(cfg.KeepDir, "lint.json")
	if err := lint.Write(report, lintPath); err != nil {
		return err
	}

	status := "✓ clean"
	if !report.Clean() {
		status = "✗ violations found"
	}
	fmt.Printf("lint: %s — %d document(s), %d hard violation(s), %d warning(s), %d injected\n",
		status, report.Summary.Total, report.Summary.HardViolations, report.Summary.Warnings, report.Summary.Injected)

	var slugs []string
	for slug := range report.Documents {
		slugs = append(slugs, slug)
	}
	sort.Strings(slugs)

	for _, slug := range slugs {
		dr := report.Documents[slug]
		if len(dr.HardViolations) == 0 && len(dr.Warnings) == 0 && len(dr.Injected) == 0 && len(dr.SchemaViolations) == 0 {
			continue
		}
		tag := dr.Status
		if dr.SchemaInvalid() {
			tag += ", schema-invalid"
		}
		fmt.Printf("\n  %s [%s]\n", slug, tag)
		for _, v := range dr.SchemaViolations {
			fmt.Printf("    ✗ %s\n", v)
		}
		for _, v := range dr.HardViolations {
			fmt.Printf("    ✗ %s\n", v)
		}
		for _, w := range dr.Warnings {
			fmt.Printf("    ⚠ %s\n", w)
		}
		for _, i := range dr.Injected {
			fmt.Printf("    ↩ %s\n", i)
		}
	}

	rel, _ := filepath.Rel(cfg.WorkspaceRoot, lintPath)
	fmt.Printf("\nreport written to %s\n", rel)

	if report.Summary.HardViolations > 0 || report.Summary.SchemaInvalid > 0 {
		os.Exit(1)
	}
	return nil
}
