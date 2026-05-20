package document

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sidisinsane/keep-go/internal/config"
)

func loadTestConfig(t *testing.T) *config.Config {
	t.Helper()
	cfg, err := config.Load("testdata")
	if err != nil {
		t.Fatalf("Load config: %v", err)
	}
	return cfg
}

func TestParseValidRagu(t *testing.T) {
	cfg := loadTestConfig(t)
	result, err := Parse(filepath.Join("testdata", "valid-ragu.md"), cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(result.SchemaViolations) > 0 {
		t.Fatalf("unexpected schema violations: %v", result.SchemaViolations)
	}
	d := result.Document
	if d.Slug != "ragu" {
		t.Errorf("slug = %q, want ragu", d.Slug)
	}
	if d.Title != "Ragù alla Bolognese" {
		t.Errorf("title = %q, want Ragù alla Bolognese", d.Title)
	}
	if d.Kind != "recipe" {
		t.Errorf("kind = %q, want recipe", d.Kind)
	}
	if d.Status != "published" {
		t.Errorf("status = %q, want published", d.Status)
	}
	if d.DateCreated != "2026-01-10" {
		t.Errorf("date_created = %q, want 2026-01-10", d.DateCreated)
	}
	if len(d.Tags) != 3 || d.Tags[0] != "italian" {
		t.Errorf("tags = %v, want [italian pasta meat]", d.Tags)
	}
	if d.Private {
		t.Error("private = true, want false")
	}
	if d.Summary != "Classic slow-cooked Bolognese built on a soffritto base." {
		t.Errorf("summary = %q", d.Summary)
	}
	if len(d.Relations) != 1 {
		t.Fatalf("relations = %d, want 1", len(d.Relations))
	}
	rel := d.Relations[0]
	if rel.Target != "soffritto" || rel.Type != "derived_from" || rel.AutoInjected {
		t.Errorf("relation = %+v, want soffritto/derived_from/not-auto", rel)
	}
}

func TestParseValidSoffritto(t *testing.T) {
	cfg := loadTestConfig(t)
	result, err := Parse(filepath.Join("testdata", "valid-soffritto.md"), cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(result.SchemaViolations) > 0 {
		t.Fatalf("unexpected schema violations: %v", result.SchemaViolations)
	}
	d := result.Document
	if d.Slug != "soffritto" {
		t.Errorf("slug = %q, want soffritto", d.Slug)
	}
	if len(d.Relations) != 1 {
		t.Fatalf("relations = %d, want 1", len(d.Relations))
	}
	rel := d.Relations[0]
	if rel.Target != "ragu" || rel.Type != "superseded_by" || !rel.AutoInjected {
		t.Errorf("relation = %+v, want ragu/superseded_by/auto", rel)
	}
}

func TestParseValidMarcellaHazan(t *testing.T) {
	cfg := loadTestConfig(t)
	result, err := Parse(filepath.Join("testdata", "valid-marcella-hazan.md"), cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(result.SchemaViolations) > 0 {
		t.Fatalf("unexpected schema violations: %v", result.SchemaViolations)
	}
	d := result.Document
	if d.Slug != "marcella-hazan" {
		t.Errorf("slug = %q, want marcella-hazan", d.Slug)
	}
	if d.Kind != "person" {
		t.Errorf("kind = %q, want person", d.Kind)
	}
	if d.Extra["birth_date"] != "1924-04-15" {
		t.Errorf("extra.birth_date = %v, want 1924-04-15", d.Extra["birth_date"])
	}
	if d.Extra["family_name"] != "Hazan" {
		t.Errorf("extra.family_name = %v, want Hazan", d.Extra["family_name"])
	}
}

func TestParseValidUnfinishedIdea(t *testing.T) {
	cfg := loadTestConfig(t)
	result, err := Parse(filepath.Join("testdata", "valid-unfinished-idea.md"), cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(result.SchemaViolations) > 0 {
		t.Fatalf("unexpected schema violations: %v", result.SchemaViolations)
	}
	d := result.Document
	if d.Slug != "unfinished-idea" {
		t.Errorf("slug = %q, want unfinished-idea", d.Slug)
	}
	if d.Kind != "note" {
		t.Errorf("kind = %q, want note", d.Kind)
	}
	if d.Status != "draft" {
		t.Errorf("status = %q, want draft", d.Status)
	}
}

func TestParseAutoInjectedRelation(t *testing.T) {
	cfg := loadTestConfig(t)
	result, err := Parse(filepath.Join("testdata", "valid-soffritto.md"), cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	d := result.Document
	if len(d.Relations) != 1 {
		t.Fatalf("relations = %d, want 1", len(d.Relations))
	}
	if !d.Relations[0].AutoInjected {
		t.Error("auto_injected = false, want true")
	}
}

func TestParseExtraFields(t *testing.T) {
	cfg := loadTestConfig(t)
	result, err := Parse(filepath.Join("testdata", "valid-marcella-hazan.md"), cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	d := result.Document
	if d.Extra == nil {
		t.Fatal("extra is nil")
	}
	if _, ok := d.Extra["birth_date"]; !ok {
		t.Error("birth_date not in extra")
	}
	if _, ok := d.Extra["family_name"]; !ok {
		t.Error("family_name not in extra")
	}
}

func TestParseNoFrontmatter(t *testing.T) {
	cfg := loadTestConfig(t)
	_, err := Parse(filepath.Join("testdata", "invalid-no-frontmatter.md"), cfg)
	if err == nil {
		t.Fatal("expected error for missing frontmatter")
	}
	if !strings.Contains(err.Error(), "no frontmatter") {
		t.Errorf("error message = %q, want 'no frontmatter'", err.Error())
	}
}

func TestParseUnclosedFrontmatter(t *testing.T) {
	cfg := loadTestConfig(t)
	_, err := Parse(filepath.Join("testdata", "invalid-unclosed-frontmatter.md"), cfg)
	if err == nil {
		t.Fatal("expected error for unclosed frontmatter")
	}
	if !strings.Contains(err.Error(), "not closed") {
		t.Errorf("error message = %q, want 'not closed'", err.Error())
	}
}

func TestParseMalformedYAML(t *testing.T) {
	cfg := loadTestConfig(t)
	_, err := Parse(filepath.Join("testdata", "invalid-malformed-yaml.md"), cfg)
	if err == nil {
		t.Fatal("expected error for malformed YAML")
	}
	if !strings.Contains(err.Error(), "YAML parse error") {
		t.Errorf("error message = %q, want 'YAML parse error'", err.Error())
	}
}

func TestParseSchemaViolation(t *testing.T) {
	cfg := loadTestConfig(t)
	result, err := Parse(filepath.Join("testdata", "invalid-missing-required-field.md"), cfg)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if len(result.SchemaViolations) == 0 {
		t.Fatal("expected schema violations, got none")
	}
}

func TestParseExtensionViolation(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "keep.yml"), []byte("schema_version: \"0.1.0\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("Load config: %v", err)
	}
	content := "---\nslug: test-person\ntitle: Test\nkind: person\nstatus: draft\ndate_created: 2026-01-01\ntags: []\n---\n\nbody\n"
	path := filepath.Join(dir, "test.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := Parse(path, cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	found := false
	for _, v := range result.SchemaViolations {
		if strings.Contains(v, "birth_date") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected extension violation containing 'birth_date', got %v", result.SchemaViolations)
	}
}

func TestParseExhaustive(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "keep.yml"), []byte("schema_version: \"0.1.0\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("Load config: %v", err)
	}
	content := "---\nslug: bad\ntitle: Bad Doc\nstatus: banana\ndate_created: 2026-01-01\ntags: []\n---\n\nbody\n"
	path := filepath.Join(dir, "bad.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := Parse(path, cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(result.SchemaViolations) < 2 {
		t.Fatalf("expected at least 2 violations, got %d: %v", len(result.SchemaViolations), result.SchemaViolations)
	}
}

func TestWriteReciprocal(t *testing.T) {
	dir := t.TempDir()
	content := "---\nslug: target\ntitle: Target\nkind: recipe\nstatus: draft\ndate_created: 2026-01-01\ntags: []\n---\n\nbody\n"
	path := filepath.Join(dir, "target.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	ok, err := WriteReciprocal(path, "source", "contradicts")
	if err != nil {
		t.Fatalf("WriteReciprocal: %v", err)
	}
	if !ok {
		t.Fatal("WriteReciprocal returned false")
	}
	result, err := Parse(path, nil)
	if err != nil {
		t.Fatalf("Parse after write: %v", err)
	}
	if len(result.Document.Relations) != 1 {
		t.Fatalf("relations = %d, want 1", len(result.Document.Relations))
	}
	rel := result.Document.Relations[0]
	if rel.Target != "source" || rel.Type != "contradicts" || !rel.AutoInjected {
		t.Errorf("relation = %+v", rel)
	}
}

func TestWriteReciprocalNoFrontmatter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "no-fm.md")
	if err := os.WriteFile(path, []byte("no frontmatter\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ok, err := WriteReciprocal(path, "source", "contradicts")
	if err != nil {
		t.Fatalf("WriteReciprocal: %v", err)
	}
	if ok {
		t.Fatal("WriteReciprocal returned true for no frontmatter")
	}
}

func TestWriteReciprocalCreatesRelations(t *testing.T) {
	dir := t.TempDir()
	content := "---\nslug: target\ntitle: Target\nkind: recipe\nstatus: draft\ndate_created: 2026-01-01\ntags: []\n---\n\nbody\n"
	path := filepath.Join(dir, "target.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	ok, err := WriteReciprocal(path, "source", "contradicts")
	if err != nil {
		t.Fatalf("WriteReciprocal: %v", err)
	}
	if !ok {
		t.Fatal("WriteReciprocal returned false")
	}
	// Verify it appends, not overwrites
	ok2, err := WriteReciprocal(path, "other", "contradicts")
	if err != nil {
		t.Fatalf("WriteReciprocal second: %v", err)
	}
	if !ok2 {
		t.Fatal("WriteReciprocal returned false on second call")
	}
	result, err := Parse(path, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(result.Document.Relations) != 2 {
		t.Fatalf("relations = %d, want 2", len(result.Document.Relations))
	}
}

func TestWriteReciprocalExistingRelation(t *testing.T) {
	dir := t.TempDir()
	content := "---\nslug: target\ntitle: Target\nkind: recipe\nstatus: draft\ndate_created: 2026-01-01\ntags: []\nrelations:\n  - target: source\n    type: contradicts\n---\n\nbody\n"
	path := filepath.Join(dir, "target.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	ok, err := WriteReciprocal(path, "source", "contradicts")
	if err != nil {
		t.Fatalf("WriteReciprocal: %v", err)
	}
	if !ok {
		t.Fatal("WriteReciprocal returned false")
	}
	result, err := Parse(path, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(result.Document.Relations) != 2 {
		t.Fatalf("relations = %d, want 2", len(result.Document.Relations))
	}
}

func TestWriteReciprocalAppendsToExisting(t *testing.T) {
	dir := t.TempDir()
	content := "---\nslug: target\ntitle: Target\nkind: recipe\nstatus: draft\ndate_created: 2026-01-01\ntags: []\nrelations:\n  - target: existing\n    type: supports\n---\n\nbody\n"
	path := filepath.Join(dir, "target.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	ok, err := WriteReciprocal(path, "source", "contradicts")
	if err != nil {
		t.Fatalf("WriteReciprocal: %v", err)
	}
	if !ok {
		t.Fatal("WriteReciprocal returned false")
	}
	result, err := Parse(path, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(result.Document.Relations) != 2 {
		t.Fatalf("relations = %d, want 2", len(result.Document.Relations))
	}
}
