// Package hashing provides content hashing functions for keep documents.
package hashing

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"github.com/sidisinsane/keep-go/internal/document"
)

// ContentHash computes a hash of the full parsed frontmatter dict.
func ContentHash(doc *document.Document) string {
	data := docToMap(doc)
	return hashMap(data)
}

// MeaningfulHash computes a hash of only the fields that affect the staleness clock.
func MeaningfulHash(doc *document.Document, extensionFields []string) string {
	data := docToMap(doc)
	watched := []string{"status", "kind", "relations"}
	watched = append(watched, extensionFields...)
	sort.Strings(watched)

	subset := make(map[string]any)
	for _, k := range watched {
		if v, ok := data[k]; ok {
			subset[k] = v
		}
	}
	return hashMap(subset)
}

func docToMap(doc *document.Document) map[string]any {
	relations := make([]string, len(doc.Relations))
	for i, r := range doc.Relations {
		relations[i] = fmt.Sprintf("Relation(target='%s', type='%s', auto_injected=%s)", r.Target, r.Type, pythonBool(r.AutoInjected))
	}
	return map[string]any{
		"slug":         doc.Slug,
		"title":        doc.Title,
		"date_created": doc.DateCreated,
		"status":       doc.Status,
		"tags":         doc.Tags,
		"kind":         doc.Kind,
		"private":      doc.Private,
		"summary":      doc.Summary,
		"relations":    relations,
		"extra":        doc.Extra,
		"source_path":  doc.SourcePath,
	}
}

func hashMap(data map[string]any) string {
	b := pythonJSON(data)
	h := sha256.Sum256([]byte(b))
	return fmt.Sprintf("%x", h)[:8]
}

// pythonJSON serialises a value exactly like Python's json.dumps(sort_keys=True, ensure_ascii=False, default=str).
func pythonJSON(v any) string {
	switch x := v.(type) {
	case nil:
		return "null"
	case bool:
		if x {
			return "true"
		}
		return "false"
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case string:
		b, _ := json.Marshal(x)
		return string(b)
	case []string:
		if len(x) == 0 {
			return "[]"
		}
		var parts []string
		for _, item := range x {
			parts = append(parts, pythonJSON(item))
		}
		return "[" + joinParts(parts) + "]"
	case []any:
		if len(x) == 0 {
			return "[]"
		}
		var parts []string
		for _, item := range x {
			parts = append(parts, pythonJSON(item))
		}
		return "[" + joinParts(parts) + "]"
	case map[string]any:
		if len(x) == 0 {
			return "{}"
		}
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var parts []string
		for _, k := range keys {
			parts = append(parts, pythonJSON(k)+": "+pythonJSON(x[k]))
		}
		return "{" + joinParts(parts) + "}"
	default:
		// Fall back to string representation, matching Python's default=str
		return pythonJSON(fmt.Sprintf("%v", x))
	}
}

func joinParts(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += ", " + parts[i]
	}
	return result
}

func pythonBool(b bool) string {
	if b {
		return "True"
	}
	return "False"
}