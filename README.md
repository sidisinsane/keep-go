# Keep

Keep is a personal wiki tool for people who think carefully and want their
thinking to compound over time. You write Markdown documents with structured
frontmatter. Keep reads that frontmatter and produces derived artefacts — a
relationship graph, a document index, a staleness report — that make your
knowledge base navigable and maintainable even as it grows.

The problem Keep solves is one most people feel but rarely name: you learn
things, you write things, you save things — and then you lose them. Not because
they are deleted, but because they are scattered, unconnected, and contextless.
Keep is the answer to *I know I wrote about this before, but I cannot find it or
pick up the thread.*

---

## Installation

**Shell script (macOS/Linux):**

```bash
curl -o- https://raw.githubusercontent.com/sidisinsane/keep-go/main/install.sh | bash
```

**PowerShell (Windows):**

```powershell
irm https://raw.githubusercontent.com/sidisinsane/keep-go/main/install.ps1 | iex
```

**Homebrew (macOS/Linux):**

```bash
brew tap sidisinsane/homebrew-tap
brew install keep
```

**Archive download (macOS/Linux):**

```bash
# Download the release asset and its corresponding SHA256 checksum file
curl -OL https://github.com/sidisinsane/keep-go/releases/download/v0.1.0/keep_Darwin_arm64.tar.gz
curl -OL https://github.com/sidisinsane/keep-go/releases/download/v0.1.0/checksums.txt

# Verify the checksum
shasum -a 256 --check --ignore-missing checksums.txt

# Extract the archive
tar -xzf keep_Darwin_arm64.tar.gz

# Move the binary to a directory on your PATH
mv keep /usr/local/bin/
```

**Build from source:**

```bash
# Clone the repository
git clone https://github.com/sidisinsane/keep-go
cd keep-go

# Build the binary
go build -o keep .

# Make it executable and move it to a directory on your PATH
chmod +x keep
mv keep /usr/local/bin/
```

**Verify the installation:**

```bash
keep --help
```

---

## Getting Started

A Keep **workspace** is a folder of Markdown files. Each file is a document.
Each document has a YAML frontmatter block at the top that describes it
structurally — its slug, title, status, relations to other documents, and more.

Keep reads that frontmatter and produces four derived artefacts:

| Command      | Artefact           | Purpose                                                                |
| ------------ | ------------------ | ---------------------------------------------------------------------- |
| `keep graph` | `.keep/graph.json` | Machine-readable graph of all documents and relations                  |
| `keep state` | `.keep/state.json` | Staleness tracking — when was each document last meaningfully changed? |
| `keep lint`  | `.keep/lint.json`  | Validation report — violations, warnings, auto-injections              |
| `keep index` | `index.md`         | Human and LLM-readable catalog of all documents                        |

These artefacts are rebuilt deterministically on each run. You never edit them
directly. The documents themselves are the source of truth.

### Workspace structure

Before running any commands:

```text
my-wiki/
├── keep.yml
├── recipes/
│   ├── ragu.md
│   └── soffritto.md
└── people/
    └── marcella-hazan.md
```

After running `keep state && keep graph && keep index && keep lint`:

```text
my-wiki/
├── keep.yml
├── index.md                  ← generated catalog
├── .keep/
│   ├── graph.json            ← document relationship graph
│   ├── state.json            ← staleness tracking
│   └── lint.json             ← validation report
├── recipes/
│   ├── ragu.md
│   └── soffritto.md
└── people/
    └── marcella-hazan.md
```

### Writing a document

```markdown
---
slug: soffritto
title: Soffritto Base
kind: recipe
status: canon
date_created: "2025-11-03"
tags: [italian, base, vegetables]
summary: Foundational aromatic base of onion, celery, and carrot for Italian sauces.
relations:
  - target: ragu
    type: derived_from
---

Soffritto is the backbone of Italian cooking. Equal parts onion, celery, and
carrot, cooked slowly in olive oil until completely soft and sweet.
```

See [docs/document-schema.md](docs/document-schema.md) for the full frontmatter reference.

---

## Configuration

Every workspace needs a `keep.yml` (or `keep.yaml`, `keep.json`) at its root.

**Minimal:**

```yaml
schema_version: "0.1.0"
```

**Full:**

```yaml
schema_version: "0.1.0"

staleness:
  draft:     { days: 14 }    # flag after 2 weeks of no meaningful change
  review:    { days: 28 }    # flag after 4 weeks
  published: { days: 180 }   # flag after 6 months
  canon:     { days: null }  # never flag canon documents as stale

completeness:
  min_ratio: 0.6             # flag if fewer than 60% of recommended fields are set
  required_fields:
    - kind
    - tags
  required_relations: 1      # flag if document has no authored relations

extensions:
  recipe:
    applies_when: { kind: recipe }
    additional_required:
      - source
    fields:
      source:    { type: string }
      prep_time: { type: string }
      servings:  { type: string }
```

All keys except `schema_version` are optional — omitting a section uses the app
defaults. Built-in extensions (`genealogy` for `kind: person`, `experiment` for
`kind: experiment`) are defined by the app and do not need to be declared.

See [docs/config-schema.md](docs/config-schema.md) for the full configuration reference.

---

## Keep and LLMs

Keep is designed to work alongside an LLM in a persistent, compounding knowledge
base. The intended workflow:

**Ingest** — when you capture a new source (article, chat export, PDF), ask the
LLM to create a Keep document for it. The LLM writes the frontmatter (`slug`,
`title`, `kind`, `tags`, `summary`, `relations`) and the prose body. The
`summary` field is specifically for the LLM — one sentence that describes what
the document is, used to populate `index.md`.

**Orient** — at the start of every session, the LLM reads `index.md` first to
locate relevant documents before drilling into them. This replaces expensive
full-workspace scans.

**Maintain** — run `keep state`, `keep graph`, `keep index`, and `keep lint`
after each session to keep the derived artefacts current. The LLM can run these
commands itself as part of an ingest skill.

**Lint as gate** — `keep lint` exits non-zero on hard violations, making it
suitable as a quality gate before promoting documents to `published` or `canon`.

A minimal ingest skill prompt looks like:

```text
Read index.md to orient yourself.
Create a new Keep document for the attached source.
Write frontmatter with slug, title, kind, status: draft, date_created (today),
tags, summary (one sentence), and any relations you can identify from index.md.
Run: keep state && keep graph && keep index && keep lint
Report the lint result.
```

---

## Documentation

| Document | What it covers |
|----------|---------------|
| [docs/config-schema.md](docs/config-schema.md) | Workspace configuration reference |
| [docs/document-schema.md](docs/document-schema.md) | Document frontmatter reference |
| [docs/document-person-schema.md](docs/document-person-schema.md) | Person extension reference |
| [docs/document-experiment-schema.md](docs/document-experiment-schema.md) | Experiment extension reference |
