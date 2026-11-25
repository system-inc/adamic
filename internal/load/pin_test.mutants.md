# Mutant for the TypeScript pin test (#g7nnryf)

`internal/load/pin_test.go`, `TestTypeScriptIsThePinnedCommit`, where it asks `internal/submodules` for the TypeScript
submodule's path. Old:

```go
	path, err := submodules.Path("../..", "TypeScript")
```

New, a name no `.gitmodules` declares, as a reader would that spelled the path or looked it up wrong:

```go
	path, err := submodules.Path("../..", "TypeScriptX")
```

Run: `go vet ./internal/load/ && go test -count=1 ./internal/load/ ./internal/submodules/ ./internal/compilercensus/ ./cmd/adamic-gate/parallelcensus/`, sealed, after `sandboxed_repositories`, at the mutant's commit 5e26ff1e81 (Loom job J-jkf6d5e).

Red, exit 1, one test, in the helper's refusal: `TestTypeScriptIsThePinnedCommit`, at `pin_test.go:26: ../..: no
submodule is named "TypeScriptX"; found "TypeScript" at cohere/TypeScript (cohere/.gitmodules), "cohere" at cohere
(.gitmodules)`. The rest of `internal/load`, and `internal/submodules`, `internal/compilercensus` and
`cmd/adamic-gate/parallelcensus`, stay green, so the test reads the path through the helper and never falls back to
one of its own.
