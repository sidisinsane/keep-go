package index

import (
	"strings"
	"testing"

	"github.com/sidisinsane/keep-go/internal/document"
)

func makeIndexResults() []document.ParseResult {
	return []document.ParseResult{
		{
			Document: &document.Document{
				Slug:    "ragu",
				Title:   "Ragù alla Bolognese",
				Kind:    "recipe",
				Status:  "published",
				Summary: "Classic slow-cooked Bolognese built on a soffritto base.",
			},
		},
		{
			Document: &document.Document{
				Slug:    "soffritto",
				Title:   "Soffritto Base",
				Kind:    "recipe",
				Status:  "canon",
				Summary: "Foundational aromatic base.",
			},
		},
		{
			Document: &document.Document{
				Slug:    "marcella-hazan",
				Title:   "Marcella Hazan",
				Kind:    "person",
				Status:  "published",
				Summary: "Italian cooking teacher.",
				Private: true,
			},
		},
		{
			Document: &document.Document{
				Slug:    "unfinished-idea",
				Title:   "Unfinished Idea",
				Kind:    "note",
				Status:  "draft",
				Summary: "A half-formed thought.",
			},
			SchemaViolations: []string{"missing field"},
		},
	}
}

func TestBuildOnlyValidDocuments(t *testing.T) {
	results := makeIndexResults()
	rows := Build(results)
	for _, r := range rows {
		if r.Slug == "unfinished-idea" {
			t.Fatal("schema-invalid document should not be in index")
		}
	}
}

func TestBuildExcludesPrivate(t *testing.T) {
	results := makeIndexResults()
	rows := Build(results)
	for _, r := range rows {
		if r.Slug == "marcella-hazan" {
			t.Fatal("private document should not be in index")
		}
	}
}

func TestBuildSortedBySlug(t *testing.T) {
	results := makeIndexResults()
	rows := Build(results)
	for i := 1; i < len(rows); i++ {
		if rows[i].Slug < rows[i-1].Slug {
			t.Fatalf("not sorted: %s before %s", rows[i-1].Slug, rows[i].Slug)
		}
	}
}

func TestRenderSeparatorWidths(t *testing.T) {
	rows := []IndexRow{
		{Slug: "ragu", Title: "Ragù", Kind: "recipe", Status: "published", Summary: "Classic"},
	}
	rendered := Render(rows)
	wantSep := "| ---- | ----- | ---- | ------ | ------- |"
	if !strings.Contains(rendered, wantSep) {
		t.Errorf("separator not found in:\n%s", rendered)
	}
}

func TestRenderTrailingNewline(t *testing.T) {
	rows := []IndexRow{
		{Slug: "ragu", Title: "Ragù", Kind: "recipe", Status: "published", Summary: "Classic"},
	}
	rendered := Render(rows)
	if !strings.HasSuffix(rendered, "\n") {
		t.Error("rendered output missing trailing newline")
	}
}
