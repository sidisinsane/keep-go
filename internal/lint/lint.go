// Package lint runs validation over workspace documents and produces a structured
// report of schema violations, hard violations, warnings, and auto-injected
// reciprocals.
package lint

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/sidisinsane/keep-go/internal/config"
	"github.com/sidisinsane/keep-go/internal/document"
	"github.com/sidisinsane/keep-go/internal/state"
)

// DocumentReport holds lint results for a single document.
type DocumentReport struct {
	Status           string   `json:"status"`
	SchemaViolations []string `json:"schema_violations,omitempty"`
	HardViolations   []string `json:"hard_violations"`
	Warnings         []string `json:"warnings"`
	Injected         []string `json:"injected"`
}

// SchemaInvalid returns true if the document has schema violations.
func (r *DocumentReport) SchemaInvalid() bool { return len(r.SchemaViolations) > 0 }

// Clean returns true if the document has no hard violations.
func (r *DocumentReport) Clean() bool { return len(r.HardViolations) == 0 }

// LintSummary holds aggregate counts.
type LintSummary struct {
	Total          int `json:"total"`
	SchemaInvalid  int `json:"schema_invalid"`
	HardViolations int `json:"hard_violations"`
	Warnings       int `json:"warnings"`
	Injected       int `json:"injected"`
}

// LintReport is the full lint output.
type LintReport struct {
	GeneratedAt string                    `json:"generated_at"`
	Summary     LintSummary               `json:"summary"`
	Documents   map[string]DocumentReport `json:"documents"`
}

// Clean returns true if there are no hard violations overall.
func (r *LintReport) Clean() bool { return r.Summary.HardViolations == 0 && r.Summary.SchemaInvalid == 0 }

// Run executes the full lint pass.
func Run(results []document.ParseResult, s *state.State, cfg *config.Config) *LintReport {
	report := &LintReport{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339Nano),
		Documents:   map[string]DocumentReport{},
	}

	// Build slug index from valid documents
	slugIndex := map[string]*document.Document{}
	privateSlugs := map[string]bool{}
	validDocs := []*document.Document{}
	for _, r := range results {
		if len(r.SchemaViolations) == 0 {
			slugIndex[r.Document.Slug] = r.Document
			if r.Document.Private {
				privateSlugs[r.Document.Slug] = true
			}
			validDocs = append(validDocs, r.Document)
		}
	}

	// Auto-inject reciprocals
	injections := map[string][]string{}
	for _, d := range validDocs {
		for _, rel := range d.Relations {
			sym, ok := cfg.RelationSymmetry[rel.Type]
			if !ok || !sym.Symmetric {
				continue
			}
			target := slugIndex[rel.Target]
			if target == nil {
				continue
			}
			reciprocalType := sym.ReciprocalType
			if reciprocalType == "" {
				reciprocalType = rel.Type
			}
			alreadyPresent := false
			for _, tr := range target.Relations {
				if tr.Target == d.Slug && tr.Type == reciprocalType {
					alreadyPresent = true
					break
				}
			}
			if alreadyPresent {
				continue
			}
			injected, _ := document.WriteReciprocal(target.SourcePath, d.Slug, reciprocalType)
			if injected {
				injections[target.Slug] = append(injections[target.Slug], fmt.Sprintf("injected '%s' from '%s'", reciprocalType, d.Slug))
			}
		}
	}

	// Re-parse injected documents
	injectedDocs := map[string]*document.Document{}
	for slug := range injections {
		if targetDoc, ok := slugIndex[slug]; ok {
			result, err := document.Parse(targetDoc.SourcePath, cfg)
			if err == nil {
				injectedDocs[slug] = result.Document
			}
		}
	}
	for slug, d := range injectedDocs {
		slugIndex[slug] = d
	}

	// Process all documents
	var slugs []string
	docMap := map[string]*document.Document{}
	for _, r := range results {
		slugs = append(slugs, r.Document.Slug)
		docMap[r.Document.Slug] = r.Document
	}
	sort.Strings(slugs)

	for _, slug := range slugs {
		r := findResult(results, slug)
		dr := DocumentReport{
			Status:         r.Document.Status,
			HardViolations: []string{},
			Warnings:       []string{},
			Injected:       []string{},
		}

		if len(r.SchemaViolations) > 0 {
			dr.SchemaViolations = r.SchemaViolations
			report.Documents[slug] = dr
			continue
		}

		d := r.Document
		if id, ok := injectedDocs[slug]; ok {
			d = id
		}

		// Hard violations
		for _, rel := range d.Relations {
			target := slugIndex[rel.Target]
			if target == nil {
				dr.HardViolations = append(dr.HardViolations, fmt.Sprintf("dangling_slug: target '%s' does not exist", rel.Target))
				continue
			}
			if !d.Private && target.Private {
				dr.HardViolations = append(dr.HardViolations, fmt.Sprintf("private_target_in_public_doc: '%s' is private", rel.Target))
			}
			if rel.Type == "contradicts" || rel.Type == "supersedes" {
				reciprocalType := "contradicts"
				if rel.Type == "supersedes" {
					reciprocalType = "superseded_by"
				}
				hasReciprocal := false
				for _, tr := range target.Relations {
					if tr.Target == d.Slug && tr.Type == reciprocalType {
						hasReciprocal = true
						break
					}
				}
				if !hasReciprocal {
					dr.HardViolations = append(dr.HardViolations, fmt.Sprintf("missing_reciprocal: '%s' has no '%s' back to '%s'", rel.Target, reciprocalType, d.Slug))
				}
			}
		}

		// Warnings: completeness (relations only)
		authored := 0
		for _, rel := range d.Relations {
			if !rel.AutoInjected {
				authored++
			}
		}
		if authored < cfg.Completeness.RequiredRelations {
			dr.Warnings = append(dr.Warnings, fmt.Sprintf("incomplete: has %d authored relation(s), minimum is %d", authored, cfg.Completeness.RequiredRelations))
		}

		// Warnings: staleness
		if ds, ok := s.Documents[slug]; ok {
			if th, ok := cfg.Staleness[d.Status]; ok && th.Days != nil {
				last, err := time.Parse(time.RFC3339Nano, ds.LastMeaningfulModification)
				if err == nil {
					age := int(time.Since(last).Hours() / 24)
					if age > *th.Days {
						dr.Warnings = append(dr.Warnings, fmt.Sprintf("stale: %d day(s) since last meaningful modification (threshold: %d)", age, *th.Days))
					}
				}
			}
		}

		// Invalid promotion
		if (d.Status == "published" || d.Status == "canon") && len(dr.HardViolations) > 0 {
			dr.HardViolations = append(dr.HardViolations, fmt.Sprintf("invalid_promotion: '%s' requires zero hard violations but %d found", d.Status, len(dr.HardViolations)))
		}

		if inj, ok := injections[slug]; ok {
			dr.Injected = inj
		}
		report.Documents[slug] = dr
	}

	// Summary
	for _, dr := range report.Documents {
		report.Summary.Total++
		if dr.SchemaInvalid() {
			report.Summary.SchemaInvalid++
		}
		report.Summary.HardViolations += len(dr.HardViolations)
		report.Summary.Warnings += len(dr.Warnings)
		report.Summary.Injected += len(dr.Injected)
	}

	return report
}

// Write persists the lint report to disk.
func Write(r *LintReport, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}

func findResult(results []document.ParseResult, slug string) document.ParseResult {
	for _, r := range results {
		if r.Document.Slug == slug {
			return r
		}
	}
	return document.ParseResult{}
}