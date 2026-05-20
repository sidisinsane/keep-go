package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sidisinsane/keep-go/internal/config"
)

func TestLoadWorkspace(t *testing.T) {
	cfg, err := config.Load("testdata/workspace")
	if err != nil {
		t.Fatalf("Load config: %v", err)
	}
	results, err := Load(cfg)
	if err != nil {
		t.Fatalf("Load workspace: %v", err)
	}
	// Workspace has 4 valid docs (ragu, soffritto, marcella-hazan, unfinished-idea)
	// plus bad.md which has no frontmatter and is skipped with a warning.
	var validCount int
	for _, r := range results {
		if len(r.SchemaViolations) == 0 {
			validCount++
		}
	}
	if validCount != 4 {
		t.Errorf("valid documents = %d, want 4", validCount)
	}
}

func TestLoadOnlyValidDocuments(t *testing.T) {
	cfg, err := config.Load("testdata/workspace")
	if err != nil {
		t.Fatalf("Load config: %v", err)
	}
	results, err := Load(cfg)
	if err != nil {
		t.Fatalf("Load workspace: %v", err)
	}
	slugs := map[string]struct{}{}
	for _, r := range results {
		if len(r.SchemaViolations) == 0 {
			slugs[r.Document.Slug] = struct{}{}
		}
	}
	expected := []string{"ragu", "soffritto", "marcella-hazan", "unfinished-idea"}
	for _, slug := range expected {
		if _, ok := slugs[slug]; !ok {
			t.Errorf("expected slug %s not found", slug)
		}
	}
}

func TestLoadExcludesKeepDir(t *testing.T) {
	// Create a temp workspace with a .keep dir containing markdown files
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "keep.yml"), []byte("schema_version: \"0.1.0\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "valid.md"), []byte("---\nslug: a\ntitle: A\nkind: recipe\nstatus: draft\ndate_created: 2026-01-01\ntags: []\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	keepDir := filepath.Join(dir, ".keep")
	if err := os.MkdirAll(keepDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(keepDir, "inside.md"), []byte("---\nslug: b\ntitle: B\nkind: recipe\nstatus: draft\ndate_created: 2026-01-01\ntags: []\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	results, err := Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Document.Slug != "a" {
		t.Fatalf("expected 1 result with slug 'a', got %d: %+v", len(results), results)
	}
}

func TestLoadExcludesIndexPath(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "keep.yml"), []byte("schema_version: \"0.1.0\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.md"), []byte("---\nslug: idx\ntitle: Index\nkind: recipe\nstatus: draft\ndate_created: 2026-01-01\ntags: []\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	results, err := Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) > 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestLoadSortedByPath(t *testing.T) {
	cfg, err := config.Load("testdata/workspace")
	if err != nil {
		t.Fatalf("Load config: %v", err)
	}
	results, err := Load(cfg)
	if err != nil {
		t.Fatalf("Load workspace: %v", err)
	}
	for i := 1; i < len(results); i++ {
		if results[i].Document.SourcePath < results[i-1].Document.SourcePath {
			t.Error("results not sorted by path")
		}
	}
}

func TestLoadWithSchemaInvalidDocuments(t *testing.T) {
	// Use the invalid testdata directory
	cfg, err := config.Load("testdata/invalid")
	if err != nil {
		t.Fatalf("Load config: %v", err)
	}
	// Write a keep.yml in the invalid dir
	if err := os.WriteFile(filepath.Join("testdata/invalid", "keep.yml"), []byte("schema_version: \"0.1.0\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	results, err := Load(cfg)
	if err != nil {
		t.Fatalf("Load workspace: %v", err)
	}
	// invalid dir has 4 invalid files + no keep.yml file (we just wrote it)
	// Let's check that schema-invalid docs are included in results
	foundInvalid := false
	for _, r := range results {
		if len(r.SchemaViolations) > 0 {
			foundInvalid = true
			break
		}
	}
	if !foundInvalid {
		t.Error("no schema-invalid documents found in results")
	}
}
