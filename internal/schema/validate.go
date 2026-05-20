package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// ExtensionField describes a single field in a user-defined extension.
type ExtensionField struct {
	Type string
	Enum []string
}

// Extension describes a user-defined extension declaration.
type Extension struct {
	AppliesWhen        map[string]string
	AdditionalRequired []string
	Fields             map[string]ExtensionField
}

var (
	compiledSchemas     map[string]*jsonschema.Schema
	compiledSchemasOnce sync.Once
)

func initCompiledSchemas() {
	compiledSchemasOnce.Do(func() {
		compiler := jsonschema.NewCompiler()

		schemas := map[string][]byte{
			"config":     ConfigSchema,
			"document":   DocumentSchema,
			"person":     PersonSchema,
			"experiment": ExperimentSchema,
		}

		compiledSchemas = make(map[string]*jsonschema.Schema, len(schemas))
		for name, data := range schemas {
			uri := "resource://" + name + ".json"
			s, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
			if err != nil {
				panic(fmt.Sprintf("unmarshal %s schema: %v", name, err))
			}
			if err := compiler.AddResource(uri, s); err != nil {
				panic(fmt.Sprintf("add %s schema: %v", name, err))
			}
			compiled, err := compiler.Compile(uri)
			if err != nil {
				panic(fmt.Sprintf("compile %s schema: %v", name, err))
			}
			compiledSchemas[name] = compiled
		}
	})
}

// Validate validates a document frontmatter map against the core document
// schema and, if the document's kind matches an extension, against that
// extension schema as well. Returns all violation messages from both passes.
func Validate(fm map[string]any, kind string, extensions map[string]Extension) []string {
	initCompiledSchemas()
	violations := validateAgainst(fm, compiledSchemas["document"])
	violations = append(violations, validateExtension(fm, kind, extensions)...)
	return violations
}

// ValidateDocument validates a document frontmatter map against the core
// document schema and returns all violation messages.
func ValidateDocument(fm map[string]any) []string {
	initCompiledSchemas()
	return validateAgainst(fm, compiledSchemas["document"])
}

// ValidateConfig validates a config map against the config schema and
// returns all violation messages.
func ValidateConfig(fm map[string]any) []string {
	initCompiledSchemas()
	return validateAgainst(fm, compiledSchemas["config"])
}

// ValidateExtension validates a document frontmatter map against the
// extension schema for the given kind (if any). Returns all violation messages.
func ValidateExtension(fm map[string]any, kind string, extensions map[string]Extension) []string {
	return validateExtension(fm, kind, extensions)
}

func validateExtension(fm map[string]any, kind string, extensions map[string]Extension) []string {
	initCompiledSchemas()

	matched := false
	var matchedName string
	for name, ext := range extensions {
		if ext.AppliesWhen["kind"] == kind {
			matched = true
			matchedName = name
			break
		}
	}
	if !matched {
		return nil
	}

	var schema *jsonschema.Schema
	switch matchedName {
	case "genealogy":
		schema = compiledSchemas["person"]
	case "experiment":
		schema = compiledSchemas["experiment"]
	default:
		data := extensionSchemaBytes(matchedName, extensions[matchedName])
		compiler := jsonschema.NewCompiler()
		uri := "resource://extension-" + matchedName + ".json"
		s, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
		if err != nil {
			return []string{fmt.Sprintf("extension schema unmarshal: %v", err)}
		}
		if err := compiler.AddResource(uri, s); err != nil {
			return []string{fmt.Sprintf("extension schema add: %v", err)}
		}
		schema, err = compiler.Compile(uri)
		if err != nil {
			return []string{fmt.Sprintf("extension schema compile: %v", err)}
		}
	}

	return validateAgainst(fm, schema)
}

// ExtensionSchemaFor returns the JSON Schema bytes for a given kind — either
// the embedded static schema for built-ins, or the runtime-generated JSON
// bytes for user-defined extensions. Returns nil if no extension applies.
func ExtensionSchemaFor(kind string, extensions map[string]Extension) []byte {
	for name, ext := range extensions {
		if ext.AppliesWhen["kind"] == kind {
			switch name {
			case "genealogy":
				return PersonSchema
			case "experiment":
				return ExperimentSchema
			default:
				return extensionSchemaBytes(name, ext)
			}
		}
	}
	return nil
}

func validateAgainst(fm map[string]any, schema *jsonschema.Schema) []string {
	if schema == nil {
		return nil
	}
	b, err := json.Marshal(fm)
	if err != nil {
		return []string{fmt.Sprintf("marshal: %v", err)}
	}
	inst, err := jsonschema.UnmarshalJSON(bytes.NewReader(b))
	if err != nil {
		return []string{fmt.Sprintf("unmarshal: %v", err)}
	}

	verr, ok := schema.Validate(inst).(*jsonschema.ValidationError)
	if !ok || verr == nil {
		return nil
	}
	return collectMessages(verr)
}

func collectMessages(err *jsonschema.ValidationError) []string {
	var msgs []string
	if len(err.Causes) == 0 {
		msgs = append(msgs, err.Error())
		return msgs
	}
	for _, c := range err.Causes {
		msgs = append(msgs, collectMessages(c)...)
	}
	return msgs
}

func extensionSchemaBytes(name string, ext Extension) []byte {
	schema := map[string]any{
		"$schema":    "https://json-schema.org/draft/2020-12/schema",
		"type":       "object",
		"required":   ext.AdditionalRequired,
		"properties": map[string]any{},
	}
	props := schema["properties"].(map[string]any)
	for fieldName, field := range ext.Fields {
		prop := map[string]any{"type": "string"}
		if field.Type == "isodate" {
			prop["pattern"] = "^[0-9]{4}-[0-9]{2}-[0-9]{2}$"
		}
		if len(field.Enum) > 0 {
			prop["enum"] = field.Enum
		}
		props[fieldName] = prop
	}
	b, _ := json.Marshal(schema)
	return b
}