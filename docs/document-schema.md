## <a name="https://raw.githubusercontent.com/sidisinsane/keep-go/v0.1.0/schema/keep-document.schema.json"></a>Keep Document

Referenced by Schema ID: <a href="#"></a>

| Key | Type | Description | Default | Examples | Extra |
|-----|------|-------------|---------|----------|-------|
| `date_created` | `string` | **ISO 8601 creation date in YYYY-MM-DD format.** |  |  |  |
| `kind` | `string` | **Semantic document type, e.g. 'recipe', 'person', 'note', 'experiment'. Drives extension selection.** |  |  |  |
| `private` | `boolean` | If true, the document is excluded from index.md and cannot be the target of relations from non-private documents. | `false` |  |  |
| `relations[].auto_injected` | `boolean` | True when this relation was written by Keep rather than authored by a human or LLM. Never set this manually. | `false` |  |  |
| `relations[].target` | `string` | Slug of the target document. Must resolve to an existing document in the workspace. |  |  |  |
| `relations[].type` | `string` | Semantic type of the relation. 'contradicts' and 'supersedes' are symmetric — Keep auto-injects the reciprocal. |  |  |  |
| `slug` | `string` | **Unique identifier for the document across the workspace. URL-safe, kebab-case.** |  |  |  |
| `status` | `string` | **Lifecycle stage of the document. Documents at 'published' or 'canon' must pass linting with zero hard violations.** |  |  |  |
| `summary` | `string` | One-sentence description of the document. Used to populate index.md. Intended to be written by an LLM during ingest. |  |  |  |
| `tags` | `string[]` | **Freeform tags for discovery and filtering.** |  |  |  |
| `title` | `string` | **Human-readable document title.** |  |  |  |
