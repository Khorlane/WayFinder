# Maintenance

## Regenerate Go Source Index

Purpose: refresh `docs/dev/go_source_index.md` from current `*.go` files.

From repo root (`C:\Projects\WayFinder`), run:

```powershell
go run ./docs/dev/scripts/generate_go_source_index.go
```

Notes:
- Generator uses Go AST (`go/parser`, `go/ast`, `go/token`), not regex.
- Index includes package-level `Types`, `Functions` (including methods), and `Variables`.
- Variables are package-level only; function-local variables are intentionally excluded.

## ToDo Commit Tags

When completing a `WF-*` item in `docs/dev/ToDo.md`, replace `(commit: pending)` or `(commit: TBD)` with the exact commit hash in the same commit (or immediate follow-up commit).
