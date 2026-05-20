// Package schema provides JSON Schema validation and runtime schema generation
// for Keep documents and workspace configuration.
package schema

import _ "embed"

// ConfigSchema holds the embedded JSON Schema for workspace configuration.
//go:embed keep-config.schema.json
var ConfigSchema []byte

// DocumentSchema holds the embedded JSON Schema for Keep documents.
//go:embed keep-document.schema.json
var DocumentSchema []byte

// PersonSchema holds the embedded JSON Schema for person (genealogy) extensions.
//go:embed keep-document-person.schema.json
var PersonSchema []byte

// ExperimentSchema holds the embedded JSON Schema for experiment extensions.
//go:embed keep-document-experiment.schema.json
var ExperimentSchema []byte

//go:generate go run generate.go
