package lint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sidisinsane/keep-go/internal/config"
	"github.com/sidisinsane/keep-go/internal/document"
	"github.com/sidisinsane/keep-go/internal/state"
)

func makeLintResults() []document.ParseResult {
	return []document.ParseResult{
		{
			Document: &document.Document{
				Slug:    "ragu",
				Title:   "Ragù alla Bolognese",
				Kind:    "recipe",
				Status:  "published",
				SourcePath: "/tmp/ragu.md",
				Relations: []document.Relation{
					{Target: "soffritto", Type: "derived_from", AutoInjected: false},
					{Target: "ghost-doc", Type: "supports", AutoInjected: false},
				},
			},
		},
		{
			Document: &document.Document{
				Slug:    "soffritto",
				Title:   "Soffritto Base",
				Kind:    "recipe",
				Status:  "canon",
				SourcePath: "/tmp/soffritto.md",
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
				SourcePath: "/tmp/marcella-hazan.md",
				Relations: []document.Relation{
					{Target: "ragu", Type: "inspired_by", AutoInjected: false},
				},
				Extra: map[string]any{
					"birth_date":  "1924-04-15",
					"family_name": "Hazan",
				},
			},
		},
		{
			Document: &document.Document{
				Slug:    "unfinished-idea",
				Title:   "Unfinished Idea",
				Kind:    "note",
				Status:  "draft",
				SourcePath: "/tmp/unfinished-idea.md",
				Relations: []document.Relation{},
			},
		},
	}
}

func makeLintConfig() *config.Config {
	return &config.Config{
		Staleness: map[string]config.StalenessThreshold{
			"draft":     {Days: intPtr(7)},
			"published": {Days: intPtr(90)},
			"canon":     {Days: nil},
		},
		Completeness: config.CompletenessConfig{
			MinRatio:          0.7,
			RequiredFields:    []string{"kind", "tags"},
			RequiredRelations: 1,
		},
		RelationSymmetry: map[string]config.RelationSymmetry{
			"contradicts":  {Symmetric: true, ReciprocalType: "contradicts"},
			"supersedes":   {Symmetric: true, ReciprocalType: "superseded_by"},
			"derived_from": {Symmetric: false},
			"extends":      {Symmetric: false},
			"supports":     {Symmetric: false},
			"inspired_by":  {Symmetric: false},
		},
		Extensions: map[string]config.Extension{
			"genealogy": {
				AppliesWhen:        map[string]string{"kind": "person"},
				AdditionalRequired: []string{"birth_date", "family_name"},
				Fields: map[string]config.ExtensionField{
					"birth_date":  {Type: "isodate"},
					"death_date":  {Type: "isodate"},
					"family_name": {Type: "string"},
				},
			},
		},
	}
}

func intPtr(i int) *int { return &i }

func TestSchemaInvalidDocumentInReport(t *testing.T) {
	results := []document.ParseResult{
		{
			Document: &document.Document{
				Slug:   "invalid-doc",
				Title:  "Invalid Doc",
				Kind:   "recipe",
				Status: "published",
			},
			SchemaViolations: []string{"missing required field: kind"},
		},
	}
	cfg := makeLintConfig()
	s := state.Load("")
	report := Run(results, s, cfg)
	dr, ok := report.Documents["invalid-doc"]
	if !ok {
		t.Fatal("schema-invalid document missing from report")
	}
	if !dr.SchemaInvalid() {
		t.Error("expected SchemaInvalid() == true")
	}
	if len(dr.SchemaViolations) == 0 {
		t.Error("expected schema violations")
	}
}

func TestSchemaInvalidDocumentSkipsLinterRules(t *testing.T) {
	results := []document.ParseResult{
		{
			Document: &document.Document{
				Slug:   "invalid-doc",
				Title:  "Invalid Doc",
				Kind:   "recipe",
				Status: "published",
			},
			SchemaViolations: []string{"missing required field: kind"},
		},
	}
	cfg := makeLintConfig()
	s := state.Load("")
	report := Run(results, s, cfg)
	dr := report.Documents["invalid-doc"]
	if len(dr.HardViolations) > 0 {
		t.Errorf("expected no hard violations for schema-invalid doc, got %v", dr.HardViolations)
	}
	if len(dr.Warnings) > 0 {
		t.Errorf("expected no warnings for schema-invalid doc, got %v", dr.Warnings)
	}
}

func TestCleanWorkspace(t *testing.T) {
	// Create a clean workspace with only valid docs and no violations
	results := []document.ParseResult{
		{
			Document: &document.Document{
				Slug:    "ragu",
				Title:   "Ragù",
				Kind:    "recipe",
				Status:  "published",
				SourcePath: "/tmp/ragu.md",
				Relations: []document.Relation{
					{Target: "soffritto", Type: "derived_from", AutoInjected: false},
				},
			},
		},
		{
			Document: &document.Document{
				Slug:    "soffritto",
				Title:   "Soffritto",
				Kind:    "recipe",
				Status:  "canon",
				SourcePath: "/tmp/soffritto.md",
				Relations: []document.Relation{
					{Target: "ragu", Type: "superseded_by", AutoInjected: true},
				},
			},
		},
	}
	cfg := makeLintConfig()
	s := state.Load("")
	report := Run(results, s, cfg)
	if report.Summary.HardViolations != 0 {
		t.Errorf("hard violations = %d, want 0", report.Summary.HardViolations)
	}
	if report.Summary.SchemaInvalid != 0 {
		t.Errorf("schema invalid = %d, want 0", report.Summary.SchemaInvalid)
	}
}

func TestDanglingSlug(t *testing.T) {
	results := makeLintResults()
	cfg := makeLintConfig()
	s := state.Load("")
	report := Run(results, s, cfg)
	dr := report.Documents["ragu"]
	found := false
	for _, v := range dr.HardViolations {
		if strings.Contains(v, "dangling_slug") && strings.Contains(v, "ghost-doc") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected dangling_slug violation, got %v", dr.HardViolations)
	}
}

func TestPrivateTargetInPublicDoc(t *testing.T) {
	results := []document.ParseResult{
		{
			Document: &document.Document{
				Slug:       "public",
				Title:      "Public Doc",
				Kind:       "recipe",
				Status:     "draft",
				SourcePath: "/tmp/public.md",
				Relations: []document.Relation{
					{Target: "private-target", Type: "supports", AutoInjected: false},
				},
			},
		},
		{
			Document: &document.Document{
				Slug:       "private-target",
				Title:      "Private Target",
				Kind:       "recipe",
				Status:     "draft",
				Private:    true,
				SourcePath: "/tmp/private-target.md",
				Relations:  []document.Relation{},
			},
		},
	}
	cfg := makeLintConfig()
	s := state.Load("")
	report := Run(results, s, cfg)
	dr := report.Documents["public"]
	found := false
	for _, v := range dr.HardViolations {
		if strings.Contains(v, "private_target_in_public_doc") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected private_target_in_public_doc violation, got %v", dr.HardViolations)
	}
}

func TestInvalidPromotion(t *testing.T) {
	results := makeLintResults()
	cfg := makeLintConfig()
	s := state.Load("")
	report := Run(results, s, cfg)
	dr := report.Documents["ragu"]
	found := false
	for _, v := range dr.HardViolations {
		if strings.Contains(v, "invalid_promotion") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected invalid_promotion violation, got %v", dr.HardViolations)
	}
}

func TestNoInvalidPromotionOnDraft(t *testing.T) {
	results := []document.ParseResult{
		{
			Document: &document.Document{
				Slug:    "bad-draft",
				Title:   "Bad Draft",
				Kind:    "note",
				Status:  "draft",
				SourcePath: "/tmp/bad-draft.md",
				Relations: []document.Relation{
					{Target: "ghost", Type: "supports", AutoInjected: false},
				},
			},
		},
	}
	cfg := makeLintConfig()
	s := state.Load("")
	report := Run(results, s, cfg)
	dr := report.Documents["bad-draft"]
	for _, v := range dr.HardViolations {
		if strings.Contains(v, "invalid_promotion") {
			t.Errorf("unexpected invalid_promotion on draft: %s", v)
		}
	}
}

func TestCompletenessRelationsWarning(t *testing.T) {
	results := []document.ParseResult{
		{
			Document: &document.Document{
				Slug:    "lonely",
				Title:   "Lonely Doc",
				Kind:    "note",
				Status:  "draft",
				SourcePath: "/tmp/lonely.md",
				Relations: []document.Relation{},
			},
		},
	}
	cfg := makeLintConfig()
	s := state.Load("")
	report := Run(results, s, cfg)
	dr := report.Documents["lonely"]
	found := false
	for _, v := range dr.Warnings {
		if strings.Contains(v, "incomplete") && strings.Contains(v, "0 authored relation") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected incomplete warning, got %v", dr.Warnings)
	}
}

func TestNoCompletenessFieldWarning(t *testing.T) {
	// Field completeness warning is eliminated — handled by schema.
	results := []document.ParseResult{
		{
			Document: &document.Document{
				Slug:    "no-kind",
				Title:   "No Kind",
				Status:  "draft",
				SourcePath: "/tmp/no-kind.md",
				Relations: []document.Relation{
					{Target: "other", Type: "supports", AutoInjected: false},
				},
			},
			SchemaViolations: []string{"missing required field: kind"},
		},
	}
	cfg := makeLintConfig()
	s := state.Load("")
	report := Run(results, s, cfg)
	dr := report.Documents["no-kind"]
	for _, v := range dr.Warnings {
		if strings.Contains(v, "missing recommended field") {
			t.Errorf("unexpected field completeness warning: %s", v)
		}
	}
}

func TestMissingReciprocal(t *testing.T) {
	results := []document.ParseResult{
		{
			Document: &document.Document{
				Slug:       "a",
				Title:      "A",
				Kind:       "recipe",
				Status:     "draft",
				SourcePath: "/tmp/a.md",
				Relations: []document.Relation{
					{Target: "b", Type: "contradicts", AutoInjected: false},
				},
			},
		},
		{
			Document: &document.Document{
				Slug:       "b",
				Title:      "B",
				Kind:       "recipe",
				Status:     "draft",
				SourcePath: "/tmp/b.md",
				Relations:  []document.Relation{},
			},
		},
	}
	cfg := makeLintConfig()
	s := state.Load("")
	report := Run(results, s, cfg)
	dr := report.Documents["a"]
	found := false
	for _, v := range dr.HardViolations {
		if strings.Contains(v, "missing_reciprocal") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected missing_reciprocal violation, got %v", dr.HardViolations)
	}
}

func TestStalenessWarning(t *testing.T) {
	results := []document.ParseResult{
		{
			Document: &document.Document{
				Slug:       "old",
				Title:      "Old Doc",
				Kind:       "recipe",
				Status:     "draft",
				SourcePath: "/tmp/old.md",
				Relations: []document.Relation{
					{Target: "other", Type: "supports", AutoInjected: false},
				},
			},
		},
		{
			Document: &document.Document{
				Slug:       "other",
				Title:      "Other",
				Kind:       "recipe",
				Status:     "draft",
				SourcePath: "/tmp/other.md",
				Relations:  []document.Relation{},
			},
		},
	}
	cfg := makeLintConfig()
	cfg.Staleness["draft"] = config.StalenessThreshold{Days: intPtr(1)}
	s := &state.State{
		SchemaVersion: state.SchemaVersion,
		Documents: map[string]state.DocumentState{
			"old": {
				LastMeaningfulModification: "2020-01-01T00:00:00Z",
				ContentHash:                "abc",
				MeaningfulHash:             "def",
				StatusAtLastCheck:          "draft",
			},
		},
	}
	report := Run(results, s, cfg)
	dr := report.Documents["old"]
	found := false
	for _, v := range dr.Warnings {
		if strings.Contains(v, "stale") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected stale warning, got %v", dr.Warnings)
	}
}

func TestAutoInjectSupersededBy(t *testing.T) {
	// Write real files so WriteReciprocal can modify them.
	dir := t.TempDir()
	aPath := filepath.Join(dir, "a.md")
	bPath := filepath.Join(dir, "b.md")
	contentA := "---\nslug: a\ntitle: A\nkind: recipe\nstatus: draft\ndate_created: 2026-01-01\ntags: []\nrelations:\n  - target: b\n    type: supersedes\n---\n\nbody\n"
	contentB := "---\nslug: b\ntitle: B\nkind: recipe\nstatus: draft\ndate_created: 2026-01-01\ntags: []\n---\n\nbody\n"
	if err := os.WriteFile(aPath, []byte(contentA), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bPath, []byte(contentB), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := makeLintConfig()
	// Parse using the real files
	resultA, err := document.Parse(aPath, cfg)
	if err != nil {
		t.Fatal(err)
	}
	resultB, err := document.Parse(bPath, cfg)
	if err != nil {
		t.Fatal(err)
	}
	results := []document.ParseResult{resultA, resultB}
	s := state.Load("")
	report := Run(results, s, cfg)
	// b should have received an injected superseded_by from a
	dr := report.Documents["b"]
	if len(dr.Injected) == 0 {
		t.Errorf("expected injection for b, got none")
	}
}

func TestNoReinjectExisting(t *testing.T) {
	dir := t.TempDir()
	aPath := filepath.Join(dir, "a.md")
	bPath := filepath.Join(dir, "b.md")
	contentA := "---\nslug: a\ntitle: A\nkind: recipe\nstatus: draft\ndate_created: 2026-01-01\ntags: []\nrelations:\n  - target: b\n    type: supersedes\n---\n\nbody\n"
	contentB := "---\nslug: b\ntitle: B\nkind: recipe\nstatus: draft\ndate_created: 2026-01-01\ntags: []\nrelations:\n  - target: a\n    type: superseded_by\n    auto_injected: true\n---\n\nbody\n"
	if err := os.WriteFile(aPath, []byte(contentA), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bPath, []byte(contentB), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := makeLintConfig()
	resultA, err := document.Parse(aPath, cfg)
	if err != nil {
		t.Fatal(err)
	}
	resultB, err := document.Parse(bPath, cfg)
	if err != nil {
		t.Fatal(err)
	}
	results := []document.ParseResult{resultA, resultB}
	s := state.Load("")
	report := Run(results, s, cfg)
	dr := report.Documents["b"]
	if len(dr.Injected) > 0 {
		t.Errorf("expected no injection for b since reciprocal already exists, got %v", dr.Injected)
	}
}

func TestLintWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "lint.json")
	report := &LintReport{
		GeneratedAt: "2026-01-01T00:00:00Z",
		Summary:     LintSummary{Total: 1, HardViolations: 0, Warnings: 0, Injected: 0},
		Documents: map[string]DocumentReport{
			"a": {Status: "draft", HardViolations: []string{}, Warnings: []string{}, Injected: []string{}},
		},
	}
	if err := Write(report, path); err != nil {
		t.Fatalf("Write: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "generated_at") {
		t.Error("output missing generated_at")
	}
	if !strings.HasSuffix(string(data), "\n") {
		t.Error("output missing trailing newline")
	}
}

func TestExitCodeSchemaInvalid(t *testing.T) {
	results := []document.ParseResult{
		{
			Document: &document.Document{
				Slug:   "invalid-doc",
				Title:  "Invalid Doc",
				Kind:   "recipe",
				Status: "published",
			},
			SchemaViolations: []string{"missing required field: kind"},
		},
	}
	cfg := makeLintConfig()
	s := state.Load("")
	report := Run(results, s, cfg)
	if report.Summary.SchemaInvalid == 0 {
		t.Error("expected schema-invalid documents to cause non-zero schema_invalid count")
	}
	if report.Clean() {
		t.Error("report should not be clean when schema-invalid documents exist")
	}
}