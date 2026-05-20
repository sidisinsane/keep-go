package graph

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/sidisinsane/keep-go/internal/config"
	"github.com/sidisinsane/keep-go/internal/document"
)

func makeResults(t *testing.T) []document.ParseResult {
	t.Helper()
	cfg, err := config.Load("../workspace/testdata/workspace")
	if err != nil {
		t.Fatalf("Load config: %v", err)
	}
	results := []document.ParseResult{
		{
			Document: &document.Document{
				Slug:   "ragu",
				Title:  "Ragù alla Bolognese",
				Kind:   "recipe",
				Status: "published",
				Relations: []document.Relation{
					{Target: "soffritto", Type: "derived_from", AutoInjected: false},
				},
			},
		},
		{
			Document: &document.Document{
				Slug:   "soffritto",
				Title:  "Soffritto Base",
				Kind:   "recipe",
				Status: "canon",
				Relations: []document.Relation{
					{Target: "ragu", Type: "superseded_by", AutoInjected: true},
				},
			},
		},
		{
			Document: &document.Document{
				Slug:    "marcella-hazan",
				Title:   "Marcella Hazan",
				Kind:    "person",
				Status:  "published",
				Private: true,
				Relations: []document.Relation{
					{Target: "ragu", Type: "inspired_by", AutoInjected: false},
				},
			},
		},
		{
			Document: &document.Document{
				Slug:   "unfinished-idea",
				Title:  "Unfinished Idea",
				Kind:   "note",
				Status: "draft",
				Relations: []document.Relation{
					{Target: "ghost", Type: "supports", AutoInjected: false},
				},
			},
			SchemaViolations: []string{"some violation"},
		},
	}
	_ = cfg // may be used later if we need schema validation in tests
	return results
}

func TestBuildOnlyValidDocuments(t *testing.T) {
	results := makeResults(t)
	g := Build(results)
	for _, n := range g.Nodes {
		if n.Slug == "unfinished-idea" {
			t.Fatal("schema-invalid document should not be in graph")
		}
	}
}

func TestBuildNodeCount(t *testing.T) {
	results := makeResults(t)
	g := Build(results)
	if len(g.Nodes) != 3 {
		t.Errorf("nodes = %d, want 3", len(g.Nodes))
	}
}

func TestBuildEdgeFields(t *testing.T) {
	results := makeResults(t)
	g := Build(results)
	if len(g.Edges) == 0 {
		t.Fatal("no edges built")
	}
	edge := g.Edges[0]
	if edge.From == "" || edge.To == "" || edge.Type == "" {
		t.Errorf("edge missing required field: %+v", edge)
	}
}

func TestBuildIncludesAutoInjectedEdges(t *testing.T) {
	results := makeResults(t)
	g := Build(results)
	found := false
	for _, e := range g.Edges {
		if e.AutoInjected {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected at least one auto-injected edge")
	}
}

func TestWriteValidJSON(t *testing.T) {
	results := makeResults(t)
	g := Build(results)
	dir := t.TempDir()
	path := filepath.Join(dir, "graph.json")
	if err := Write(g, path); err != nil {
		t.Fatalf("Write: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var decoded Graph
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.GeneratedAt == "" {
		t.Error("GeneratedAt is empty")
	}
	if len(decoded.Nodes) != len(g.Nodes) {
		t.Errorf("nodes = %d, want %d", len(decoded.Nodes), len(g.Nodes))
	}
	if len(decoded.Edges) != len(g.Edges) {
		t.Errorf("edges = %d, want %d", len(decoded.Edges), len(g.Edges))
	}
}

func TestWriteCreatesDir(t *testing.T) {
	results := makeResults(t)
	g := Build(results)
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "graph.json")
	if err := Write(g, path); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file not created: %v", err)
	}
}

func TestWriteOverwrites(t *testing.T) {
	results := makeResults(t)
	g := Build(results)
	dir := t.TempDir()
	path := filepath.Join(dir, "graph.json")
	if err := os.WriteFile(path, []byte("old content"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Write(g, path); err != nil {
		t.Fatalf("Write: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) == "old content" {
		t.Error("file was not overwritten")
	}
}
