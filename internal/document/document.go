// Package document parses Keep Markdown documents and manages their frontmatter.
package document

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/sidisinsane/keep-go/internal/config"
	"github.com/sidisinsane/keep-go/internal/schema"
)

// Relation is a typed directed edge from one document to another.
type Relation struct {
	Target       string
	Type         string
	AutoInjected bool
}

// Document is a parsed Keep document.
type Document struct {
	Slug        string
	Title       string
	Kind        string
	Status      string
	DateCreated string
	Tags        []string
	Private     bool
	Summary     string
	Relations   []Relation
	Extra       map[string]any
	SourcePath  string
}

// ParseResult carries both the document and any schema violations.
type ParseResult struct {
	Document         *Document
	SchemaViolations []string
}

// Parse reads and parses a single Markdown file.
func Parse(path string, cfg *config.Config) (ParseResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ParseResult{}, err
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return ParseResult{}, fmt.Errorf("%s: no frontmatter block found", path)
	}

	closing := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			closing = i
			break
		}
	}
	if closing == -1 {
		return ParseResult{}, fmt.Errorf("%s: frontmatter block is not closed", path)
	}

	fmText := strings.Join(lines[1:closing], "\n")
	var fm map[string]any
	if err := yaml.Unmarshal([]byte(fmText), &fm); err != nil {
		return ParseResult{}, fmt.Errorf("%s: YAML parse error: %w", path, err)
	}

	// Normalise time.Time values to YYYY-MM-DD strings before schema validation.
	// YAML parses unquoted dates (e.g. 2026-01-01) as time.Time, but the schema
	// expects plain date strings.
	normaliseYAML(fm)

	// Schema validation
	violations := schema.ValidateDocument(fm)

	kind := ""
	if k, ok := fm["kind"].(string); ok {
		kind = k
	}
	if kind != "" && cfg != nil {
		extensions := make(map[string]schema.Extension, len(cfg.Extensions))
		for name, ext := range cfg.Extensions {
			fields := make(map[string]schema.ExtensionField, len(ext.Fields))
			for fn, f := range ext.Fields {
				fields[fn] = schema.ExtensionField{Type: f.Type, Enum: f.Enum}
			}
			extensions[name] = schema.Extension{
				AppliesWhen:        ext.AppliesWhen,
				AdditionalRequired: ext.AdditionalRequired,
				Fields:             fields,
			}
		}
		extViolations := schema.ValidateExtension(fm, kind, extensions)
		violations = append(violations, extViolations...)
	}

	doc := buildDocument(fm, path)
	return ParseResult{Document: doc, SchemaViolations: violations}, nil
}

// WriteReciprocal appends a reciprocal relation to a document's frontmatter.
func WriteReciprocal(path, sourceSlug, reciprocalType string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, nil
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return false, nil
	}

	closing := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			closing = i
			break
		}
	}
	if closing == -1 {
		return false, nil
	}

	fmText := strings.Join(lines[1:closing], "\n") + "\n"

	var node yaml.Node
	if err := yaml.Unmarshal([]byte(fmText), &node); err != nil {
		return false, nil
	}
	if node.Kind != yaml.DocumentNode || len(node.Content) == 0 {
		return false, nil
	}
	root := node.Content[0]
	if root.Kind != yaml.MappingNode {
		return false, nil
	}

	// Find or create relations node
	relIdx := -1
	for i := 0; i < len(root.Content); i += 2 {
		if root.Content[i].Value == "relations" {
			relIdx = i + 1
			break
		}
	}

	if relIdx == -1 {
		// Append relations key and empty sequence
		root.Content = append(root.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: "relations"})
		seq := &yaml.Node{Kind: yaml.SequenceNode}
		root.Content = append(root.Content, seq)
		relIdx = len(root.Content) - 1
	}

	seq := root.Content[relIdx]
	if seq.Kind != yaml.SequenceNode {
		return false, nil
	}

	newRel := &yaml.Node{
		Kind: yaml.MappingNode,
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "target"},
			{Kind: yaml.ScalarNode, Value: sourceSlug},
			{Kind: yaml.ScalarNode, Value: "type"},
			{Kind: yaml.ScalarNode, Value: reciprocalType},
			{Kind: yaml.ScalarNode, Value: "auto_injected"},
			{Kind: yaml.ScalarNode, Value: "true", Tag: "!!bool"},
		},
	}
	seq.Content = append(seq.Content, newRel)

	out, err := yaml.Marshal(&node)
	if err != nil {
		return false, nil
	}

	// Reassemble
	newLines := append([]string{"---"}, strings.Split(string(out), "\n")...)
	newLines = append(newLines, lines[closing:]...)

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(strings.Join(newLines, "\n")), 0o644); err != nil {
		return false, nil
	}
	if err := os.Rename(tmp, path); err != nil {
		return false, nil
	}

	return true, nil
}

func buildDocument(fm map[string]any, path string) *Document {
	known := map[string]bool{
		"slug": true, "title": true, "date_created": true,
		"status": true, "tags": true, "kind": true,
		"private": true, "summary": true, "relations": true,
	}

	doc := &Document{
		Slug:        stringVal(fm, "slug"),
		Title:       stringVal(fm, "title"),
		Kind:        stringVal(fm, "kind"),
		Status:      stringVal(fm, "status"),
		DateCreated: dateVal(fm, "date_created"),
		Summary:     stringVal(fm, "summary"),
		SourcePath:  path,
		Extra:       map[string]any{},
	}

	if p, ok := fm["private"].(bool); ok {
		doc.Private = p
	}
	if t, ok := fm["tags"].([]any); ok {
		for _, v := range t {
			if s, ok := v.(string); ok {
				doc.Tags = append(doc.Tags, s)
			}
		}
	}
	if r, ok := fm["relations"].([]any); ok {
		for _, v := range r {
			if m, ok := v.(map[string]any); ok {
				rel := Relation{
					Target:       stringVal(m, "target"),
					Type:         stringVal(m, "type"),
					AutoInjected: boolVal(m, "auto_injected"),
				}
				doc.Relations = append(doc.Relations, rel)
			}
		}
	}

	for k, v := range fm {
		if !known[k] {
			if t, ok := v.(time.Time); ok {
				doc.Extra[k] = t.Format("2006-01-02")
			} else {
				doc.Extra[k] = v
			}
		}
	}

	return doc
}

func normaliseYAML(m map[string]any) {
	for k, v := range m {
		switch x := v.(type) {
		case time.Time:
			m[k] = x.Format("2006-01-02")
		case map[string]any:
			normaliseYAML(x)
		}
	}
}

func dateVal(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	switch x := v.(type) {
	case string:
		return x
	case time.Time:
		return x.Format("2006-01-02")
	default:
		return fmt.Sprintf("%v", v)
	}
}

func stringVal(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func boolVal(m map[string]any, key string) bool {
	v, ok := m[key]
	if !ok {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}