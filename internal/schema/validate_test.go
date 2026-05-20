package schema

import (
	"strings"
	"testing"
)

func TestValidateDocumentValid(t *testing.T) {
	fm := map[string]any{
		"slug":         "ragu",
		"title":        "Ragù alla Bolognese",
		"kind":         "recipe",
		"status":       "published",
		"date_created": "2026-01-10",
		"tags":         []any{"italian", "pasta"},
	}
	violations := ValidateDocument(fm)
	if len(violations) > 0 {
		t.Fatalf("expected no violations, got %v", violations)
	}
}

func TestValidateDocumentMissingRequired(t *testing.T) {
	fm := map[string]any{
		"slug":         "ragu",
		"title":        "Ragù",
		"status":       "published",
		"date_created": "2026-01-10",
		"tags":         []any{"italian"},
	}
	violations := ValidateDocument(fm)
	found := false
	for _, v := range violations {
		if strings.Contains(v, "kind") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected violation containing 'kind', got %v", violations)
	}
}

func TestValidateDocumentInvalidStatus(t *testing.T) {
	fm := map[string]any{
		"slug":         "ragu",
		"title":        "Ragù",
		"kind":         "recipe",
		"status":       "banana",
		"date_created": "2026-01-10",
		"tags":         []any{"italian"},
	}
	violations := ValidateDocument(fm)
	if len(violations) == 0 {
		t.Fatal("expected violations for invalid status, got none")
	}
}

func TestValidateDocumentInvalidRelationType(t *testing.T) {
	fm := map[string]any{
		"slug":         "ragu",
		"title":        "Ragù",
		"kind":         "recipe",
		"status":       "published",
		"date_created": "2026-01-10",
		"tags":         []any{"italian"},
		"relations": []any{
			map[string]any{"target": "soffritto", "type": "loves"},
		},
	}
	violations := ValidateDocument(fm)
	if len(violations) == 0 {
		t.Fatal("expected violations for invalid relation type, got none")
	}
}

func TestValidateDocumentExhaustive(t *testing.T) {
	fm := map[string]any{
		"slug":         "ragu",
		"title":        "Ragù",
		"status":       "banana",
		"date_created": "2026-01-10",
		"tags":         []any{"italian"},
		"relations": []any{
			map[string]any{"target": "soffritto", "type": "loves"},
		},
	}
	violations := ValidateDocument(fm)
	if len(violations) < 3 {
		t.Fatalf("expected at least 3 violations, got %d: %v", len(violations), violations)
	}
}

func TestValidateExtensionPerson(t *testing.T) {
	fm := map[string]any{
		"slug":         "marcella",
		"title":        "Marcella Hazan",
		"kind":         "person",
		"status":       "published",
		"date_created": "2026-01-10",
		"tags":         []any{"italian"},
	}
	extensions := map[string]Extension{
		"genealogy": {
			AppliesWhen:        map[string]string{"kind": "person"},
			AdditionalRequired: []string{"birth_date", "family_name"},
			Fields: map[string]ExtensionField{
				"birth_date":  {Type: "isodate"},
				"death_date":  {Type: "isodate"},
				"family_name": {Type: "string"},
			},
		},
	}
	violations := ValidateExtension(fm, "person", extensions)
	found := false
	for _, v := range violations {
		if strings.Contains(v, "birth_date") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected violation containing 'birth_date', got %v", violations)
	}
}

func TestValidateExtensionUserDefined(t *testing.T) {
	fm := map[string]any{
		"slug":         "exp1",
		"title":        "Experiment 1",
		"kind":         "custom",
		"status":       "draft",
		"date_created": "2026-01-10",
		"tags":         []any{"science"},
	}
	extensions := map[string]Extension{
		"custom_ext": {
			AppliesWhen:        map[string]string{"kind": "custom"},
			AdditionalRequired: []string{"field_a"},
			Fields: map[string]ExtensionField{
				"field_a": {Type: "string"},
				"field_b": {Type: "isodate"},
			},
		},
	}
	violations := ValidateExtension(fm, "custom", extensions)
	found := false
	for _, v := range violations {
		if strings.Contains(v, "field_a") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected violation containing 'field_a', got %v", violations)
	}
}

func TestValidateExtensionNoMatch(t *testing.T) {
	fm := map[string]any{
		"slug":         "ragu",
		"title":        "Ragù",
		"kind":         "recipe",
		"status":       "published",
		"date_created": "2026-01-10",
		"tags":         []any{"italian"},
	}
	extensions := map[string]Extension{
		"genealogy": {
			AppliesWhen:        map[string]string{"kind": "person"},
			AdditionalRequired: []string{"birth_date"},
			Fields:             map[string]ExtensionField{},
		},
	}
	violations := ValidateExtension(fm, "recipe", extensions)
	if len(violations) > 0 {
		t.Fatalf("expected no violations when no extension matches, got %v", violations)
	}
}

func TestValidateCombined(t *testing.T) {
	// Document missing kind (core violation) and kind=person missing birth_date (extension violation)
	fm := map[string]any{
		"slug":         "marcella",
		"title":        "Marcella Hazan",
		"kind":         "person",
		"status":       "published",
		"date_created": "2026-01-10",
		"tags":         []any{"italian"},
	}
	extensions := map[string]Extension{
		"genealogy": {
			AppliesWhen:        map[string]string{"kind": "person"},
			AdditionalRequired: []string{"birth_date", "family_name"},
			Fields: map[string]ExtensionField{
				"birth_date":  {Type: "isodate"},
				"family_name": {Type: "string"},
			},
		},
	}
	violations := Validate(fm, "person", extensions)
	foundBirthDate := false
	foundFamilyName := false
	for _, v := range violations {
		if strings.Contains(v, "birth_date") {
			foundBirthDate = true
		}
		if strings.Contains(v, "family_name") {
			foundFamilyName = true
		}
	}
	if !foundBirthDate {
		t.Errorf("expected violation containing 'birth_date', got %v", violations)
	}
	if !foundFamilyName {
		t.Errorf("expected violation containing 'family_name', got %v", violations)
	}
}

func TestValidateNoExtensionMatch(t *testing.T) {
	fm := map[string]any{
		"slug":         "ragu",
		"title":        "Ragù",
		"kind":         "recipe",
		"status":       "published",
		"date_created": "2026-01-10",
		"tags":         []any{"italian"},
	}
	extensions := map[string]Extension{
		"genealogy": {
			AppliesWhen:        map[string]string{"kind": "person"},
			AdditionalRequired: []string{"birth_date"},
			Fields:             map[string]ExtensionField{},
		},
	}
	violations := Validate(fm, "recipe", extensions)
	if len(violations) > 0 {
		t.Fatalf("expected no violations for valid document with no matching extension, got %v", violations)
	}
}

func TestValidateConfigValid(t *testing.T) {
	fm := map[string]any{
		"schema_version": "0.1.0",
		"staleness": map[string]any{
			"draft": map[string]any{"days": 14},
		},
	}
	violations := ValidateConfig(fm)
	if len(violations) > 0 {
		t.Fatalf("expected no violations, got %v", violations)
	}
}

func TestValidateConfigInvalid(t *testing.T) {
	fm := map[string]any{
		"schema_version": 99,
	}
	violations := ValidateConfig(fm)
	if len(violations) == 0 {
		t.Fatal("expected violations for invalid schema_version")
	}
}

func TestExtensionSchemaForBuiltIn(t *testing.T) {
	extensions := map[string]Extension{
		"genealogy": {
			AppliesWhen:        map[string]string{"kind": "person"},
			AdditionalRequired: []string{"birth_date", "family_name"},
			Fields: map[string]ExtensionField{
				"birth_date":  {Type: "isodate"},
				"death_date":  {Type: "isodate"},
				"family_name": {Type: "string"},
			},
		},
	}
	b := ExtensionSchemaFor("person", extensions)
	if b == nil {
		t.Fatal("expected schema bytes for person")
	}
	if len(b) == 0 {
		t.Error("schema bytes empty")
	}
}

func TestExtensionSchemaForUserDefined(t *testing.T) {
	extensions := map[string]Extension{
		"custom": {
			AppliesWhen:        map[string]string{"kind": "custom"},
			AdditionalRequired: []string{"field_a"},
			Fields: map[string]ExtensionField{
				"field_a": {Type: "string"},
				"field_b": {Type: "isodate"},
			},
		},
	}
	b := ExtensionSchemaFor("custom", extensions)
	if b == nil {
		t.Fatal("expected schema bytes for custom")
	}
	if len(b) == 0 {
		t.Error("schema bytes empty")
	}
}

func TestExtensionSchemaForNoMatch(t *testing.T) {
	extensions := map[string]Extension{
		"genealogy": {
			AppliesWhen:        map[string]string{"kind": "person"},
			AdditionalRequired: []string{"birth_date"},
			Fields:             map[string]ExtensionField{},
		},
	}
	b := ExtensionSchemaFor("recipe", extensions)
	if b != nil {
		t.Error("expected nil when no extension matches")
	}
}