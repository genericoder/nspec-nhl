# mdgen

Generate Markdown from YAML, preview it in the terminal, and edit it in
the browser with changes synced back to both the `.md` and `.yaml`
files.

## Install

```sh
go install ./cmd/mdgen
# or
make build   # writes bin/mdgen
```

## Usage

### 1. Write a YAML document

```yaml
# yamls/example.yaml
title: "mdgen Example"
blocks:
  - type: heading
    level: 2
    text: "What is this?"
  - type: paragraph
    text: "Plain text, block by block."
  - type: list
    ordered: false
    items: ["one", "two"]
  - type: code
    language: bash
    text: |
      echo hello
  - type: table
    headers: ["A", "B"]
    rows: [["1", "2"]]
  - type: quote
    text: "A quote."
  - type: hr
```

See the full block schema in [`internal/doc/doc.go`](internal/doc/doc.go).

### 2. Generate Markdown

```sh
mdgen generate --dir yamls
# yamls/example.yaml -> yamls/example.md
```

### 3. Preview in the terminal

```sh
mdgen view yamls/example.yaml
```

Renders a styled header (file + block summary) followed by
syntax-highlighted Markdown via [glamour](https://github.com/charmbracelet/glamour).

### 4. Edit in the browser

`view` also starts a local editor server for that file and prints its
URL (Ctrl+C to stop; pass `--no-serve` to skip this and just print the
preview). To serve every document in a directory instead:

```sh
mdgen serve --dir yamls --open
```

Saving in the browser writes both `example.md` and re-derives
`example.yaml` from the edited Markdown. The server only listens on
`127.0.0.1`.

## Commands

| Command | Purpose |
|---|---|
| `mdgen generate --dir yamls [--out DIR]` | Batch-generate `.md` from every `.yaml`/`.yml` in a directory |
| `mdgen view <file.yaml> [--no-serve] [--open] [--port N] [--width N]` | Terminal preview, plus a browser editor for that file |
| `mdgen serve [--dir yamls] [--open] [--port N]` | Browser editor for a whole directory |

## Development

```sh
make build     # bin/mdgen
make test      # go test ./... -race
make coverage  # coverage.out + function-level report
make lint      # gofmt + go vet
make ci        # lint + test
```

## Layout

```
cmd/mdgen/            CLI entrypoint and cobra commands
internal/doc/          YAML document model + validation
internal/mdrender/     doc.Document -> Markdown
internal/mdparse/      Markdown -> doc.Document (goldmark AST walk)
internal/generate/     batch yaml -> md generation
internal/tui/           terminal preview (glamour + lipgloss)
internal/httpserver/    browser editor server + embedded frontend
internal/browseropen/  cross-platform "open the default browser" helper
testdata/yamls/         fixtures used by tests
yamls/                  example documents (mdgen's own default --dir)
```
