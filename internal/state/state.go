// Package state manages the workspace state store, tracking per-document hashes and modification times.
package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/sidisinsane/keep-go/internal/config"
	"github.com/sidisinsane/keep-go/internal/document"
	"github.com/sidisinsane/keep-go/internal/hashing"
)

// SchemaVersion is the current state file format version.
const SchemaVersion = 1

// DocumentState holds per-document state.
type DocumentState struct {
	LastMeaningfulModification string `json:"last_meaningful_modification"`
	ContentHash                string `json:"content_hash"`
	MeaningfulHash             string `json:"meaningful_hash"`
	StatusAtLastCheck          string `json:"status_at_last_check"`
}

// State is the full workspace state store.
type State struct {
	SchemaVersion int                      `json:"schema_version"`
	GeneratedAt   string                   `json:"generated_at"`
	Documents     map[string]DocumentState `json:"documents"`
}

// Load reads state from disk, returning an empty state on any error.
func Load(path string) *State {
	data, err := os.ReadFile(path)
	if err != nil {
		return &State{SchemaVersion: SchemaVersion, Documents: map[string]DocumentState{}}
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return &State{SchemaVersion: SchemaVersion, Documents: map[string]DocumentState{}}
	}
	if s.Documents == nil {
		s.Documents = map[string]DocumentState{}
	}
	return &s
}

// Update recomputes state from the current document set.
func Update(current *State, results []document.ParseResult, cfg *config.Config) *State {
	updated := &State{
		SchemaVersion: SchemaVersion,
		GeneratedAt:   now(),
		Documents:     map[string]DocumentState{},
	}

	for _, r := range results {
		if len(r.SchemaViolations) > 0 {
			continue
		}
		d := r.Document
		extFields := fieldsForKind(d.Kind, cfg.Extensions)
		cHash := hashing.ContentHash(d)
		mHash := hashing.MeaningfulHash(d, extFields)

		existing, ok := current.Documents[d.Slug]
		if !ok {
			info, err := os.Stat(d.SourcePath)
			mtime := now()
			if err == nil {
				mtime = info.ModTime().UTC().Format(time.RFC3339Nano)
			}
			updated.Documents[d.Slug] = DocumentState{
				LastMeaningfulModification: mtime,
				ContentHash:                cHash,
				MeaningfulHash:             mHash,
				StatusAtLastCheck:          d.Status,
			}
		} else if existing.ContentHash == cHash {
			updated.Documents[d.Slug] = existing
		} else if existing.MeaningfulHash == mHash {
			updated.Documents[d.Slug] = DocumentState{
				LastMeaningfulModification: existing.LastMeaningfulModification,
				ContentHash:                cHash,
				MeaningfulHash:             mHash,
				StatusAtLastCheck:          d.Status,
			}
		} else {
			updated.Documents[d.Slug] = DocumentState{
				LastMeaningfulModification: now(),
				ContentHash:                cHash,
				MeaningfulHash:             mHash,
				StatusAtLastCheck:          d.Status,
			}
		}
	}

	return updated
}

// Write persists state to disk.
func Write(s *State, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}

func now() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

func fieldsForKind(kind string, extensions map[string]config.Extension) []string {
	var result []string
	for _, ext := range extensions {
		if ext.AppliesWhen["kind"] == kind {
			for name := range ext.Fields {
				result = append(result, name)
			}
		}
	}
	return result
}
