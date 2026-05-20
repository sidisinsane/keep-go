package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var binaryPath string

func TestMain(m *testing.M) {
	// Build the keep binary once for all subprocess tests.
	dir, err := os.MkdirTemp("", "keep-test-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "mkdir temp: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)
	binaryPath = filepath.Join(dir, "keep")

	// Derive repo root from this test file's location (cmd/ is one level below root).
	repoRoot, err := filepath.Abs("..")
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve repo root: %v\n", err)
		os.Exit(1)
	}

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "build binary: %v\n%s\n", err, out)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func copyWorkspace(t *testing.T) string {
	t.Helper()
	src := "../internal/workspace/testdata/workspace"
	dst := t.TempDir()
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	for _, entry := range entries {
		if err := copyDir(filepath.Join(src, entry.Name()), filepath.Join(dst, entry.Name())); err != nil {
			t.Fatalf("copyDir: %v", err)
		}
	}
	return dst
}

func copyDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		if err := os.MkdirAll(dst, 0o755); err != nil {
			return err
		}
		entries, err := os.ReadDir(src)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if err := copyDir(filepath.Join(src, entry.Name()), filepath.Join(dst, entry.Name())); err != nil {
				return err
			}
		}
		return nil
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, info.Mode())
}

func runKeep(t *testing.T, dir string, args ...string) *exec.Cmd {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = dir
	return cmd
}

func TestGraphCommand(t *testing.T) {
	dir := copyWorkspace(t)
	cmd := runKeep(t, dir, "graph")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("graph failed: %v\noutput: %s", err, out)
	}
	if !strings.Contains(string(out), "graph written to") {
		t.Errorf("stdout missing expected text: %s", out)
	}
	if _, err := os.Stat(filepath.Join(dir, ".keep", "graph.json")); err != nil {
		t.Errorf("graph.json not created: %v", err)
	}
}

func TestIndexCommand(t *testing.T) {
	dir := copyWorkspace(t)
	cmd := runKeep(t, dir, "index")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("index failed: %v\noutput: %s", err, out)
	}
	if !strings.Contains(string(out), "index written to") {
		t.Errorf("stdout missing expected text: %s", out)
	}
	if _, err := os.Stat(filepath.Join(dir, "index.md")); err != nil {
		t.Errorf("index.md not created: %v", err)
	}
}

func TestStateCommand(t *testing.T) {
	dir := copyWorkspace(t)
	cmd := runKeep(t, dir, "state")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("state failed: %v\noutput: %s", err, out)
	}
	if !strings.Contains(string(out), "state written to") {
		t.Errorf("stdout missing expected text: %s", out)
	}
	if _, err := os.Stat(filepath.Join(dir, ".keep", "state.json")); err != nil {
		t.Errorf("state.json not created: %v", err)
	}
}

func TestLintCommandClean(t *testing.T) {
	dir := copyWorkspace(t)
	cmd := runKeep(t, dir, "lint")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("lint failed: %v\noutput: %s", err, out)
	}
	if !strings.Contains(string(out), "lint: ✓ clean") {
		t.Errorf("stdout missing clean indicator: %s", out)
	}
	if !strings.Contains(string(out), "report written to") {
		t.Errorf("stdout missing report path: %s", out)
	}
	if _, err := os.Stat(filepath.Join(dir, ".keep", "lint.json")); err != nil {
		t.Errorf("lint.json not created: %v", err)
	}
}

func TestLintSchemaInvalidDocument(t *testing.T) {
	dir := copyWorkspace(t)
	// Introduce a schema-invalid document (missing kind).
	badDoc := "---\nslug: bad\ntitle: Bad\nstatus: draft\ndate_created: 2026-01-01\ntags: []\n---\n\nbody\n"
	if err := os.WriteFile(filepath.Join(dir, "bad.md"), []byte(badDoc), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := runKeep(t, dir, "lint")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected lint to exit non-zero for schema-invalid document")
	}
	if !strings.Contains(string(out), "schema-invalid") {
		t.Errorf("stdout missing schema-invalid tag: %s", out)
	}
}

func TestLintExitsOneOnHardViolation(t *testing.T) {
	dir := copyWorkspace(t)
	// Introduce a dangling slug.
	badDoc := "---\nslug: bad\ntitle: Bad\nkind: recipe\nstatus: published\ndate_created: 2026-01-01\ntags: []\nrelations:\n  - target: ghost\n    type: supports\n---\n\nbody\n"
	if err := os.WriteFile(filepath.Join(dir, "bad.md"), []byte(badDoc), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := runKeep(t, dir, "lint")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected lint to exit non-zero for hard violation")
	}
	if !strings.Contains(string(out), "dangling_slug") {
		t.Errorf("stdout missing dangling_slug: %s", out)
	}
}

func TestLintExitsZeroOnWarningsOnly(t *testing.T) {
	dir := copyWorkspace(t)
	// A draft with no authored relations gets an "incomplete" warning but no hard violation.
	warnDoc := "---\nslug: warn\ntitle: Warn\nkind: note\nstatus: draft\ndate_created: 2026-01-01\ntags: []\n---\n\nbody\n"
	if err := os.WriteFile(filepath.Join(dir, "warn.md"), []byte(warnDoc), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := runKeep(t, dir, "lint")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("lint exited non-zero for warnings-only: %v\noutput: %s", err, out)
	}
	if !strings.Contains(string(out), "incomplete") {
		t.Errorf("stdout missing incomplete warning: %s", out)
	}
}

func TestNoConfigExitsNonzero(t *testing.T) {
	dir := t.TempDir()
	cmd := runKeep(t, dir, "graph")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected exit non-zero when no config found")
	}
	stderr := string(out)
	if !strings.Contains(stderr, "no keep config file found") {
		t.Errorf("stderr missing 'no keep config file found': %s", stderr)
	}
	if !strings.Contains(stderr, "workspace root") {
		t.Errorf("stderr missing 'workspace root': %s", stderr)
	}
}

func TestVerboseFlagAccepted(t *testing.T) {
	dir := copyWorkspace(t)
	cmd := runKeep(t, dir, "-v", "graph")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("graph with -v failed: %v\noutput: %s", err, out)
	}
	if !strings.Contains(string(out), "graph written to") {
		t.Errorf("stdout missing expected text: %s", out)
	}
}

func TestHelpListsSubcommands(t *testing.T) {
	cmd := exec.Command(binaryPath, "--help")
	out, err := cmd.CombinedOutput()
	if err != nil {
		// --help may exit 0 or non-zero depending on flag impl; just check output.
	}
	help := string(out)
	for _, sub := range []string{"graph", "index", "lint", "state"} {
		if !strings.Contains(help, sub) {
			t.Errorf("--help missing subcommand %q: %s", sub, help)
		}
	}
}

func TestUnknownSubcommand(t *testing.T) {
	dir := copyWorkspace(t)
	cmd := runKeep(t, dir, "banana")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected exit non-zero for unknown subcommand")
	}
	if !strings.Contains(string(out), "unknown subcommand") {
		t.Errorf("stderr missing 'unknown subcommand': %s", out)
	}
}

func TestAllCommandsRunOnFixtureWorkspace(t *testing.T) {
	dir := copyWorkspace(t)
	for _, sub := range []string{"graph", "index", "state", "lint"} {
		cmd := runKeep(t, dir, sub)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%s failed: %v\noutput: %s", sub, err, out)
		}
	}
}