# godocgen

Renders `./pkg/...` into Markdown for the docs site.

ADR-0016.

## What it does

1. Walks `./pkg/` top-level subdirectories.
2. Uses `go/ast` + `go/doc` to extract every exported declaration
   (types, methods, functions, consts, vars).
3. Renders Markdown with `go/printer` for stable, formatted signatures.
4. Writes one `pkg-<name>.md` per package + a top-level `index.md`.
6. For non-English locales, writes an index that links to the English version.

## CLI

```bash
go run ./cmd/godocgen --locale=en --version=vX.Y.Z
go run ./cmd/godocgen --locale=en,zh,ja,ko,fr,de,ru,es,it --version=vX.Y.Z
```

Flags:
- `--pkg`: input dir (default `./pkg`)
- `--out`: output dir (default `./website/content/docs/api`)
- `--locale`: comma-separated codes (default `en`)
- `--version`: label written into `index.md` (default `dev`)

## Library use

```go
opts := godocgen.Options{
  PkgDir:  "/abs/pkg",
  OutDir:  "/abs/website/content/docs/api",
  Locale:  "en",
  Version: "v0.3.0",
}
if err := godocgen.Generate(opts); err != nil { ... }
```

## Tests

```bash
go test ./internal/godocgen/...
```

Covers happy path, non-English index, empty pkg, and `printDecl` round-trip.

## CI

`.github/workflows/governance.yml` runs `go run ./cmd/godocgen --locale=...
--version=${GITHUB_SHA::7}` after `go test ./...`, then uploads
`website/content/docs/api/**` as a GitHub Actions artifact.

The output dir is in `website/.gitignore` so it never reaches `main`.
