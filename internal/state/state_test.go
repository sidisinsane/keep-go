package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sidisinsane/keep-go/internal/config"
	"github.com/sidisinsane/keep-go/internal/document"
)

func makeDocResult(slug, status string, relations []document.Relation) document.ParseResult {
	return document.ParseResult{
		Document: &document.Document{
			Slug:        slug,
			Title:       slug,
			Kind:        "recipe",
			Status:      status,
			DateCreated: "2026-01-01",
			Tags:        []string{"tag"},
			SourcePath:  "/tmp/" + slug + ".md",
			Relations:   relations,
			Extra:       map[string]any{},
		},
	}
}

func TestLoadMissing(t *testing.T) {
	s := Load("/nonexistent/state.json")
	if s.SchemaVersion != SchemaVersion {
		t.Errorf("schema_version = %d, want %d", s.SchemaVersion, SchemaVersion)
	}
	if s.Documents == nil {
		t.Error("Documents is nil")
	}
	if len(s.Documents) != 0 {
		t.Errorf("documents = %d, want 0", len(s.Documents))
	}
}

func TestLoadCorrupt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	if err := os.WriteFile(path, []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	s := Load(path)
	if s.SchemaVersion != SchemaVersion {
		t.Errorf("schema_version = %d, want %d", s.SchemaVersion, SchemaVersion)
	}
	if len(s.Documents) != 0 {
		t.Errorf("documents = %d, want 0", len(s.Documents))
	}
}

func TestUpdateSeedsFromMtime(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "doc.md")
	if err := os.WriteFile(path, []byte("---\nslug: a\ntitle: A\nkind: recipe\nstatus: draft\ndate_created: 2026-01-01\ntags: []\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "keep.yml"), []byte("schema_version: \"0.1.0\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	result := document.ParseResult{
		Document: &document.Document{
			Slug:        "a",
			Title:       "A",
			Kind:        "recipe",
			Status:      "draft",
			DateCreated: "2026-01-01",
			Tags:        []string{},
			SourcePath:  path,
			Extra:       map[string]any{},
		},
	}
	current := Load("")
	updated := Update(current, []document.ParseResult{result}, cfg)
	if len(updated.Documents) != 1 {
		t.Fatalf("documents = %d, want 1", len(updated.Documents))
	}
	ds := updated.Documents["a"]
	info, _ := os.Stat(path)
	want := info.ModTime().UTC().Format(time.RFC3339Nano)
	if ds.LastMeaningfulModification != want {
		t.Errorf("LastMeaningfulModification = %s, want %s", ds.LastMeaningfulModification, want)
	}
}

func TestUpdateUnchanged(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "keep.yml"), []byte("schema_version: \"0.1.0\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	result := makeDocResult("a", "draft", nil)
	result.Document.SourcePath = filepath.Join(dir, "a.md")
	if err := os.WriteFile(result.Document.SourcePath, []byte("---\nslug: a\ntitle: A\nkind: recipe\nstatus: draft\ndate_created: 2026-01-01\ntags: []\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// First run to get real hashes
	current := Load("")
	first := Update(current, []document.ParseResult{result}, cfg)
	// Second run with identical document
	second := Update(first, []document.ParseResult{result}, cfg)
	if len(second.Documents) != 1 {
		t.Fatalf("documents = %d, want 1", len(second.Documents))
	}
	ds1 := first.Documents["a"]
	ds2 := second.Documents["a"]
	if ds2.ContentHash != ds1.ContentHash {
		t.Error("content hash changed for unchanged document")
	}
	if ds2.LastMeaningfulModification != ds1.LastMeaningfulModification {
		t.Error("last meaningful modification changed for unchanged document")
	}
}

func TestUpdateCosmeticChange(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "keep.yml"), []byte("schema_version: \"0.1.0\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "a.md")
	content1 := "---\nslug: a\ntitle: A\nkind: recipe\nstatus: draft\ndate_created: 2026-01-01\ntags: []\nsummary: First\n---\n"
	if err := os.WriteFile(path, []byte(content1), 0o644); err != nil {
		t.Fatal(err)
	}
	result1 := document.ParseResult{
		Document: &document.Document{
			Slug:        "a",
			Title:       "A",
			Kind:        "recipe",
			Status:      "draft",
			DateCreated: "2026-01-01",
			Tags:        []string{},
			Summary:     "First",
			SourcePath:  path,
			Extra:       map[string]any{},
		},
	}
	current := Load("")
	first := Update(current, []document.ParseResult{result1}, cfg)

	// Cosmetic change: summary only
	result2 := document.ParseResult{
		Document: &document.Document{
			Slug:        "a",
			Title:       "A",
			Kind:        "recipe",
			Status:      "draft",
			DateCreated: "2026-01-01",
			Tags:        []string{},
			Summary:     "Second",
			SourcePath:  path,
			Extra:       map[string]any{},
		},
	}
	second := Update(first, []document.ParseResult{result2}, cfg)
	ds := second.Documents["a"]
	if ds.LastMeaningfulModification != first.Documents["a"].LastMeaningfulModification {
		t.Error("cosmetic change should not reset meaningful modification")
	}
	if ds.ContentHash == first.Documents["a"].ContentHash {
		t.Error("content hash should change for cosmetic edit")
	}
}

func TestUpdateMeaningfulChange(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "keep.yml"), []byte("schema_version: \"0.1.0\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "a.md")
	result1 := document.ParseResult{
		Document: &document.Document{
			Slug:        "a",
			Title:       "A",
			Kind:        "recipe",
			Status:      "draft",
			DateCreated: "2026-01-01",
			Tags:        []string{},
			SourcePath:  path,
			Extra:       map[string]any{},
		},
	}
	current := Load("")
	first := Update(current, []document.ParseResult{result1}, cfg)

	// Meaningful change: status
	result2 := document.ParseResult{
		Document: &document.Document{
			Slug:        "a",
			Title:       "A",
			Kind:        "recipe",
			Status:      "published",
			DateCreated: "2026-01-01",
			Tags:        []string{},
			SourcePath:  path,
			Extra:       map[string]any{},
		},
	}
	second := Update(first, []document.ParseResult{result2}, cfg)
	ds := second.Documents["a"]
	if ds.LastMeaningfulModification == first.Documents["a"].LastMeaningfulModification {
		t.Error("meaningful change should reset meaningful modification")
	}
}

func TestUpdateOnlyProcessesValidDocuments(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "keep.yml"), []byte("schema_version: \"0.1.0\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	results := []document.ParseResult{
		makeDocResult("valid", "draft", nil),
		{Document: &document.Document{Slug: "invalid", Title: "Invalid", Kind: "recipe", Status: "draft", DateCreated: "2026-01-01", Tags: []string{}, SourcePath: "/tmp/invalid.md", Extra: map[string]any{}}, SchemaViolations: []string{"missing field"}},
	}
	current := Load("")
	updated := Update(current, results, cfg)
	if _, ok := updated.Documents["valid"]; !ok {
		t.Error("valid document missing from state")
	}
	if _, ok := updated.Documents["invalid"]; ok {
		t.Error("invalid document should not be in state")
	}
}

func TestUpdateDropsDeletedDocument(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "keep.yml"), []byte("schema_version: \"0.1.0\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	results := []document.ParseResult{
		makeDocResult("a", "draft", nil),
	}
	current := &State{
		SchemaVersion: SchemaVersion,
		Documents: map[string]DocumentState{
			"a": {LastMeaningfulModification: "2026-01-01T00:00:00Z", ContentHash: "abc", MeaningfulHash: "def", StatusAtLastCheck: "draft"},
			"b": {LastMeaningfulModification: "2026-01-01T00:00:00Z", ContentHash: "ghi", MeaningfulHash: "jkl", StatusAtLastCheck: "draft"},
		},
	}
	updated := Update(current, results, cfg)
	if _, ok := updated.Documents["b"]; ok {
		t.Error("deleted document should be dropped")
	}
}

func TestWriteRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	s := &State{
		SchemaVersion: SchemaVersion,
		GeneratedAt:   "2026-01-01T00:00:00Z",
		Documents: map[string]DocumentState{
			"a": {
				LastMeaningfulModification: "2026-01-01T00:00:00Z",
				ContentHash:                "abc12345",
				MeaningfulHash:             "def67890",
				StatusAtLastCheck:          "draft",
			},
		},
	}
	if err := Write(s, path); err != nil {
		t.Fatalf("Write: %v", err)
	}
	loaded := Load(path)
	if loaded.SchemaVersion != s.SchemaVersion {
		t.Errorf("schema_version = %d, want %d", loaded.SchemaVersion, s.SchemaVersion)
	}
	if len(loaded.Documents) != 1 {
		t.Fatalf("documents = %d, want 1", len(loaded.Documents))
	}
	ds := loaded.Documents["a"]
	if ds.ContentHash != "abc12345" {
		t.Errorf("content_hash = %s, want abc12345", ds.ContentHash)
	}
}