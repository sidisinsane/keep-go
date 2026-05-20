// Package config reads, validates and merges workspace configuration for Keep.
package config

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/sidisinsane/keep-go/internal/schema"
)

//go:embed keep.defaults.json
var defaultsJSON []byte

var configFileCandidates = []string{
	"keep.yml",
	"keep.yaml",
	"keep.json",
}

// StalenessThreshold holds the days threshold for a status; nil means never stale.
type StalenessThreshold struct {
	Days *int
}

// CompletenessConfig holds completeness scoring thresholds.
type CompletenessConfig struct {
	MinRatio          float64
	RequiredFields    []string
	RequiredRelations int
}

// RelationSymmetry describes symmetry and reciprocal type for a relation.
type RelationSymmetry struct {
	Symmetric     bool
	ReciprocalType string
}

// ExtensionField describes a single field in a user-defined extension.
type ExtensionField struct {
	Type string
	Enum []string
}

// Extension describes an extension declaration.
type Extension struct {
	AppliesWhen        map[string]string
	AdditionalRequired []string
	Fields             map[string]ExtensionField
}

// Config is the fully merged runtime configuration for a workspace.
type Config struct {
	SchemaVersion    int
	Staleness        map[string]StalenessThreshold
	Completeness     CompletenessConfig
	RelationSymmetry map[string]RelationSymmetry
	Extensions       map[string]Extension
	WorkspaceRoot    string
	KeepDir          string
	GraphPath        string
	StatePath        string
	IndexPath        string
}

// Load reads, validates and merges workspace configuration.
func Load(workspaceRoot string) (*Config, error) {
	// 1. Read app defaults.
	var defaults map[string]any
	if err := json.Unmarshal(defaultsJSON, &defaults); err != nil {
		return nil, fmt.Errorf("unmarshal defaults: %w", err)
	}

	// 2. Probe config files.
	var workspaceRaw map[string]any
	var found string
	for _, name := range configFileCandidates {
		p := filepath.Join(workspaceRoot, name)
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		found = p
		if filepath.Ext(name) == ".json" {
			if err := json.Unmarshal(data, &workspaceRaw); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		} else {
			if err := yaml.Unmarshal(data, &workspaceRaw); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		}
		break
	}

	if found == "" {
		fmt.Fprintf(os.Stderr, "error: no keep config file found in %s (keep.yml, keep.yaml, or keep.json)\nMake sure you are running 'keep' from your workspace root.\n", workspaceRoot)
		os.Exit(1)
	}

	// 3. Validate workspace config against schema.
	normaliseYAML(workspaceRaw)
	violations := schema.ValidateConfig(workspaceRaw)
	if len(violations) > 0 {
		for _, v := range violations {
			fmt.Fprintf(os.Stderr, "error: %s\n", v)
		}
		os.Exit(1)
	}

	// 4. Deep-merge.
	merged := deepMerge(defaults, workspaceRaw)

	// 5. Populate Config.
	cfg, err := buildConfig(merged)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	cfg.WorkspaceRoot = workspaceRoot
	cfg.KeepDir = filepath.Join(workspaceRoot, ".keep")
	cfg.GraphPath = filepath.Join(cfg.KeepDir, "graph.json")
	cfg.StatePath = filepath.Join(cfg.KeepDir, "state.json")
	cfg.IndexPath = filepath.Join(workspaceRoot, "index.md")

	return cfg, nil
}

func buildConfig(raw map[string]any) (*Config, error) {
	cfg := &Config{
		SchemaVersion:    intValue(raw, "schema_version"),
		Staleness:        map[string]StalenessThreshold{},
		RelationSymmetry: map[string]RelationSymmetry{},
		Extensions:       map[string]Extension{},
	}

	if s, ok := raw["staleness"].(map[string]any); ok {
		for k, v := range s {
			if m, ok := v.(map[string]any); ok {
				cfg.Staleness[k] = StalenessThreshold{Days: intPtrValue(m, "days")}
			}
		}
	}

	if c, ok := raw["completeness"].(map[string]any); ok {
		cfg.Completeness = CompletenessConfig{
			MinRatio:          floatValue(c, "min_ratio"),
			RequiredFields:    stringSliceValue(c, "required_fields"),
			RequiredRelations: intValue(c, "required_relations"),
		}
	}

	if rs, ok := raw["relation_symmetry"].(map[string]any); ok {
		for k, v := range rs {
			if m, ok := v.(map[string]any); ok {
				cfg.RelationSymmetry[k] = RelationSymmetry{
					Symmetric:      boolValue(m, "symmetric"),
					ReciprocalType: stringValue(m, "reciprocal_type"),
				}
			}
		}
	}

	if exts, ok := raw["extensions"].(map[string]any); ok {
		for name, v := range exts {
			if m, ok := v.(map[string]any); ok {
				cfg.Extensions[name] = buildExtension(m)
			}
		}
	}

	return cfg, nil
}

func buildExtension(raw map[string]any) Extension {
	ext := Extension{
		AppliesWhen:        map[string]string{},
		AdditionalRequired: stringSliceValue(raw, "additional_required"),
		Fields:             map[string]ExtensionField{},
	}
	if aw, ok := raw["applies_when"].(map[string]any); ok {
		for k, v := range aw {
			ext.AppliesWhen[k] = fmt.Sprintf("%v", v)
		}
	}
	if f, ok := raw["fields"].(map[string]any); ok {
		for k, v := range f {
			if fm, ok := v.(map[string]any); ok {
				ext.Fields[k] = ExtensionField{
				Type: stringValue(fm, "type"),
				Enum: stringSliceValue(fm, "enum"),
			}
			}
		}
	}
	return ext
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

func deepMerge(base, override map[string]any) map[string]any {
	result := make(map[string]any, len(base))
	for k, v := range base {
		result[k] = deepCopy(v)
	}
	for k, v := range override {
		if bm, ok := result[k].(map[string]any); ok {
			if vm, ok := v.(map[string]any); ok {
				result[k] = deepMerge(bm, vm)
				continue
			}
		}
		result[k] = deepCopy(v)
	}
	return result
}

func deepCopy(v any) any {
	switch x := v.(type) {
	case map[string]any:
		m := make(map[string]any, len(x))
		for k, v2 := range x {
			m[k] = deepCopy(v2)
		}
		return m
	case []any:
		s := make([]any, len(x))
		for i, v2 := range x {
			s[i] = deepCopy(v2)
		}
		return s
	default:
		return v
	}
}

func intValue(m map[string]any, key string) int {
	v, ok := m[key]
	if !ok {
		return 0
	}
	switch x := v.(type) {
	case int:
		return x
	case int64:
		return int(x)
	case float64:
		return int(x)
	default:
		return 0
	}
}

func intPtrValue(m map[string]any, key string) *int {
	v, ok := m[key]
	if !ok {
		return nil
	}
	if v == nil {
		return nil
	}
	switch x := v.(type) {
	case int:
		return &x
	case int64:
		i := int(x)
		return &i
	case float64:
		i := int(x)
		return &i
	default:
		return nil
	}
}

func floatValue(m map[string]any, key string) float64 {
	v, ok := m[key]
	if !ok {
		return 0
	}
	switch x := v.(type) {
	case float64:
		return x
	case float32:
		return float64(x)
	case int:
		return float64(x)
	case int64:
		return float64(x)
	default:
		return 0
	}
}

func stringValue(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func boolValue(m map[string]any, key string) bool {
	v, ok := m[key]
	if !ok {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

func stringSliceValue(m map[string]any, key string) []string {
	v, ok := m[key]
	if !ok {
		return nil
	}
	if s, ok := v.([]any); ok {
		out := make([]string, len(s))
		for i, item := range s {
			if str, ok := item.(string); ok {
				out[i] = str
			}
		}
		return out
	}
	return nil
}