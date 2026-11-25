# Mutant for the ES2024 global audit (#g7nnryf)

`internal/apple/naming/naming_test.go`, `TestES2024GlobalAudit`, where it asks `internal/submodules` for the TypeScript
submodule's path. Old:

```go
	typeScript, err := submodules.Path("../../..", "TypeScript")
```

New, a name no `.gitmodules` declares, as a reader would that spelled the path or looked it up wrong:

```go
	typeScript, err := submodules.Path("../../..", "TypeScriptX")
```

Run: `go vet ./internal/apple/naming/ && go test -count=1 ./internal/apple/naming/ ./internal/submodules/ ./internal/compilercensus/ ./cmd/adamic-gate/parallelcensus/`, sealed, after `sandboxed_repositories`, at the mutant's commit 1b20b0f517 (Loom job J-ggkfnzh).

Red, exit 1, one test, in the helper's refusal: `TestES2024GlobalAudit`, at `naming_test.go:426: ../../..: no
submodule is named "TypeScriptX"; found "TypeScript" at cohere/TypeScript (cohere/.gitmodules), "cohere" at cohere
(.gitmodules)`. The rest of `internal/apple/naming`, and `internal/submodules`, `internal/compilercensus` and
`cmd/adamic-gate/parallelcensus`, stay green, so the test reads the path through the helper and never falls back to
one of its own.
