## <a name="https://raw.githubusercontent.com/sidisinsane/keep-go/v0.1.0/schema/keep-config.schema.json"></a>Keep Workspace Configuration

Referenced by Schema ID: <a href="#"></a>

| Key | Type | Description | Default | Examples | Extra |
|-----|------|-------------|---------|----------|-------|
| `completeness.min_ratio` | `number` | Minimum ratio of populated recommended fields required before a document is considered complete. Value between 0 and 1. | `0.6` |  |  |
| `completeness.required_fields` | `string[]` | List of frontmatter field names that must be present for a document to be considered complete. Missing fields produce an incomplete warning. | `[kind tags]` |  |  |
| `completeness.required_relations` | `integer` | Minimum number of authored (non-auto-injected) relations a document must have to be considered complete. | `1` |  |  |
| `extensions` | `object` | User-defined extensions that add required fields to documents of a specific kind. Each key is the extension name. Built-in extensions (genealogy, experiment) are defined by the app and do not need to be declared here. |  |  |  |
| `schema_version` | `string` | **The semantic version of the Keep configuration schema, e.g. '0.1.0'. Must be a valid semver string.** |  |  |  |
| `staleness.canon` | `object` | [Canon Staleness](#e1574da0c09386d3a765c91b73df4240deb9d28498937b5336307717f030700e) |  |  |  |
| `staleness.draft` | `object` | [Draft Staleness](#b428a399a0b99f4c8ba593df52963b5e23077044ce7341d55180985667e49fe5) |  |  |  |
| `staleness.published` | `object` | [Published Staleness](#8fde7276fb84e43fa1245c1bd22252ce68c3988c24468e8b57de879f8ffaa887) |  |  |  |
| `staleness.review` | `object` | [Review Staleness](#b5dfede73c51f3d053d08f3a75713e2c42efe575ad308b248e8272ec24ece06c) |  |  |  |

## <a name="b428a399a0b99f4c8ba593df52963b5e23077044ce7341d55180985667e49fe5"></a>Draft Staleness

Referenced by Schema ID: <a href="#https://raw.githubusercontent.com/sidisinsane/keep-go/v0.1.0/schema/keep-config.schema.json"><https://raw.githubusercontent.com/sidisinsane/keep-go/v0.1.0/schema/keep-config.schema.json></a>

| Key | Type | Description | Default | Examples | Extra |
|-----|------|-------------|---------|----------|-------|
| `days` | `unknown` | **Number of days before a draft document is considered stale. Accepts an integer or null. Set to null to disable.** | `14` |  |  |

## <a name="b5dfede73c51f3d053d08f3a75713e2c42efe575ad308b248e8272ec24ece06c"></a>Review Staleness

Referenced by Schema ID: <a href="#https://raw.githubusercontent.com/sidisinsane/keep-go/v0.1.0/schema/keep-config.schema.json"><https://raw.githubusercontent.com/sidisinsane/keep-go/v0.1.0/schema/keep-config.schema.json></a>

| Key | Type | Description | Default | Examples | Extra |
|-----|------|-------------|---------|----------|-------|
| `days` | `unknown` | **Number of days before a review document is considered stale. Accepts an integer or null. Set to null to disable.** | `28` |  |  |

## <a name="8fde7276fb84e43fa1245c1bd22252ce68c3988c24468e8b57de879f8ffaa887"></a>Published Staleness

Referenced by Schema ID: <a href="#https://raw.githubusercontent.com/sidisinsane/keep-go/v0.1.0/schema/keep-config.schema.json"><https://raw.githubusercontent.com/sidisinsane/keep-go/v0.1.0/schema/keep-config.schema.json></a>

| Key | Type | Description | Default | Examples | Extra |
|-----|------|-------------|---------|----------|-------|
| `days` | `unknown` | **Number of days before a published document is considered stale. Accepts an integer or null. Set to null to disable.** | `180` |  |  |

## <a name="e1574da0c09386d3a765c91b73df4240deb9d28498937b5336307717f030700e"></a>Canon Staleness

Referenced by Schema ID: <a href="#https://raw.githubusercontent.com/sidisinsane/keep-go/v0.1.0/schema/keep-config.schema.json"><https://raw.githubusercontent.com/sidisinsane/keep-go/v0.1.0/schema/keep-config.schema.json></a>

| Key | Type | Description | Default | Examples | Extra |
|-----|------|-------------|---------|----------|-------|
| `days` | `unknown` | **Number of days before a canon document is considered stale. Accepts an integer or null. Set to null to disable.** | `null` |  |  |
