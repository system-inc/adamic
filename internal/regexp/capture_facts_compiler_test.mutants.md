# Mutant for the compiler capture table test (#g7nnryf)

`internal/regexp/capture_facts_compiler_test.go`, `TestCaptureFactsCompilerTable`, where it asks `internal/submodules`
for the TypeScript submodule's path before walking `tsc/testdata/fixtures/compiler` under it. Old:

```go
	typeScript, err := submodules.Path("../..", "TypeScript")
```

New, a name no `.gitmodules` declares, as a reader would that spelled the path or looked it up wrong:

```go
	typeScript, err := submodules.Path("../..", "TypeScriptX")
```

Run: `go vet ./internal/regexp/ && go test -count=1 ./internal/regexp/ ./internal/submodules/ ./internal/compilercensus/ ./cmd/adamic-gate/parallelcensus/`, sealed, after `sandboxed_repositories`, at the mutant's commit a3e680bdc1 (Loom job J-5xsrnpd).

Red, exit 1, one test, in the helper's refusal: `TestCaptureFactsCompilerTable`, at
`capture_facts_compiler_test.go:34: ../..: no submodule is named "TypeScriptX"; found "TypeScript" at cohere/TypeScript
(cohere/.gitmodules), "cohere" at cohere (.gitmodules)`. The rest of `internal/regexp`, and `internal/submodules`,
`internal/compilercensus` and `cmd/adamic-gate/parallelcensus`, stay green, so the test walks the path the helper
reads and never falls back to one of its own.
