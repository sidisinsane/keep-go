package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadValidYML(t *testing.T) {
	cfg, err := Load("testdata")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Completeness.MinRatio != 0.7 {
		t.Errorf("min_ratio = %f, want 0.7", cfg.Completeness.MinRatio)
	}
}

func TestLoadValidYAML(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "keep.yaml"), []byte("schema_version: \"0.1.0\"\nstaleness:\n  draft: { days: 3 }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if draft := cfg.Staleness["draft"].Days; draft == nil || *draft != 3 {
		t.Errorf("draft days = %v, want 3", draft)
	}
}

func TestLoadValidJSON(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "keep.json"), []byte(`{"schema_version":"0.1.0","staleness":{"draft":{"days":5}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if draft := cfg.Staleness["draft"].Days; draft == nil || *draft != 5 {
		t.Errorf("draft days = %v, want 5", draft)
	}
}

func TestLoadPrecedence(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "keep.yml"), []byte("schema_version: \"0.1.0\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "keep.yaml"), []byte("schema_version: \"0.1.0\"\nstaleness:\n  draft: { days: 99 }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	// keep.yml wins; it has no staleness override so defaults apply.
	if draft := cfg.Staleness["draft"].Days; draft == nil || *draft != 14 {
		t.Errorf("draft days = %v, want 14 (default)", draft)
	}
}

func TestLoadMissingFile(t *testing.T) {
	dir := t.TempDir()
	if os.Getenv("_TEST_EXIT") == "1" {
		_, _ = Load(dir)
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestLoadMissingFile")
	cmd.Dir = "."
	cmd.Env = append(os.Environ(), "_TEST_EXIT=1")
	out, _ := cmd.CombinedOutput()
	if cmd.ProcessState == nil || cmd.ProcessState.ExitCode() != 1 {
		t.Errorf("exit code = %d, want 1", cmd.ProcessState.ExitCode())
	}
	stderr := string(out)
	if !strings.Contains(stderr, "no keep config file found") {
		t.Errorf("stderr missing 'no keep config file found': %q", stderr)
	}
	if !strings.Contains(stderr, "workspace root") {
		t.Errorf("stderr missing 'workspace root': %q", stderr)
	}
}

func TestLoadSchemaViolation(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "keep.yml"), []byte("schema_version: 99\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if os.Getenv("_TEST_EXIT") == "1" {
		_, _ = Load(dir)
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestLoadSchemaViolation")
	cmd.Dir = "."
	cmd.Env = append(os.Environ(), "_TEST_EXIT=1")
	out, _ := cmd.CombinedOutput()
	if cmd.ProcessState == nil || cmd.ProcessState.ExitCode() != 1 {
		t.Errorf("exit code = %d, want 1", cmd.ProcessState.ExitCode())
	}
	stderr := string(out)
	if !strings.Contains(stderr, "schema_version") {
		t.Logf("stderr: %q", stderr)
	}
}

func TestDeepMergeListsReplace(t *testing.T) {
	base := map[string]any{
		"completeness": map[string]any{
			"required_fields": []any{"kind", "tags"},
		},
	}
	override := map[string]any{
		"completeness": map[string]any{
			"required_fields": []any{"kind"},
		},
	}
	merged := deepMerge(base, override)
	fields := merged["completeness"].(map[string]any)["required_fields"].([]any)
	if len(fields) != 1 || fields[0] != "kind" {
		t.Errorf("required_fields = %v, want [kind]", fields)
	}
}

func TestDeepMergeMapsMerge(t *testing.T) {
	base := map[string]any{
		"staleness": map[string]any{
			"draft": map[string]any{"days": 14},
			"review": map[string]any{"days": 28},
		},
	}
	override := map[string]any{
		"staleness": map[string]any{
			"draft": map[string]any{"days": 7},
		},
	}
	merged := deepMerge(base, override)
	st := merged["staleness"].(map[string]any)
	if st["draft"].(map[string]any)["days"] != 7 {
		t.Errorf("draft days = %v, want 7", st["draft"])
	}
	if st["review"].(map[string]any)["days"] != 28 {
		t.Errorf("review days = %v, want 28", st["review"])
	}
}

func TestRelationSymmetryFixed(t *testing.T) {
	cfg, err := Load("testdata")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if _, ok := cfg.RelationSymmetry["contradicts"]; !ok {
		t.Error("relation_symmetry.contradicts missing")
	}
	if _, ok := cfg.RelationSymmetry["supersedes"]; !ok {
		t.Error("relation_symmetry.supersedes missing")
	}
}