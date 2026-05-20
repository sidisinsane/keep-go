package hashing

import (
	"testing"

	"github.com/sidisinsane/keep-go/internal/document"
)

func makeDoc() *document.Document {
	return &document.Document{
		Slug:        "ragu",
		Title:       "Ragù alla Bolognese",
		Kind:        "recipe",
		Status:      "published",
		DateCreated: "2026-01-10",
		Tags:        []string{"italian", "pasta", "meat"},
		Private:     false,
		Summary:     "Classic slow-cooked Bolognese built on a soffritto base.",
		Relations: []document.Relation{
			{Target: "soffritto", Type: "derived_from", AutoInjected: false},
		},
		Extra:      map[string]any{},
		SourcePath: "testdata/recipes/ragu.md",
	}
}

func TestContentHashFormat(t *testing.T) {
	doc := makeDoc()
	hash := ContentHash(doc)
	if len(hash) != 8 {
		t.Errorf("len(hash) = %d, want 8", len(hash))
	}
	for _, c := range hash {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			t.Errorf("hash contains non-hex char: %c", c)
		}
	}
}

func TestContentHashStable(t *testing.T) {
	doc := makeDoc()
	h1 := ContentHash(doc)
	h2 := ContentHash(doc)
	if h1 != h2 {
		t.Errorf("%s != %s", h1, h2)
	}
}

func TestContentHashSensitiveToStatus(t *testing.T) {
	doc := makeDoc()
	h1 := ContentHash(doc)
	doc.Status = "draft"
	h2 := ContentHash(doc)
	if h1 == h2 {
		t.Error("content hash should change when status changes")
	}
}

func TestMeaningfulHashIgnoresCosmetic(t *testing.T) {
	doc := makeDoc()
	h1 := MeaningfulHash(doc, nil)
	doc.Summary = "Totally different summary"
	doc.Tags = []string{"completely", "different", "tags"}
	h2 := MeaningfulHash(doc, nil)
	if h1 != h2 {
		t.Error("meaningful hash should not change for cosmetic edits")
	}
}

func TestMeaningfulHashSensitiveToStatus(t *testing.T) {
	doc := makeDoc()
	h1 := MeaningfulHash(doc, nil)
	doc.Status = "draft"
	h2 := MeaningfulHash(doc, nil)
	if h1 == h2 {
		t.Error("meaningful hash should change when status changes")
	}
}

func TestMeaningfulHashIgnoresExtraFields(t *testing.T) {
	// Python meaningful_hash only looks at top-level keys, not nested extra.
	doc := makeDoc()
	doc.Extra = map[string]any{"birth_date": "1924-04-15"}
	h1 := MeaningfulHash(doc, []string{"birth_date"})
	doc.Extra["birth_date"] = "1924-04-16"
	h2 := MeaningfulHash(doc, []string{"birth_date"})
	if h1 != h2 {
		t.Error("meaningful hash should NOT change when extra fields change (matches Python behavior)")
	}
}

func TestContentHashWithComplexExtra(t *testing.T) {
	doc := makeDoc()
	doc.Extra = map[string]any{
		"nested": map[string]any{
			"a": 1,
			"b": 2.5,
		},
		"list": []any{"x", "y", "z"},
		"float": 3.14,
	}
	h := ContentHash(doc)
	if len(h) != 8 {
		t.Errorf("len(hash) = %d, want 8", len(h))
	}
}

func TestContentHashWithNilExtra(t *testing.T) {
	doc := makeDoc()
	doc.Extra = nil
	h := ContentHash(doc)
	if len(h) != 8 {
		t.Errorf("len(hash) = %d, want 8", len(h))
	}
}