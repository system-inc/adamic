# Cohere regular expression inputs

`inventory.json` is source input to `TestRegExpCohereInventory` in
`internal/native/regexp_cohere_patterns_test.go`. It refreshes the binder-based
inventory from `library/l2-land` 23538b4f for cohere commit
7945d102a6c18dd36adf9114a758ce646e8b2359, using TypeScript 6.0.3.
The inventory includes the stage1 cohere ports. Source hashes and original
literal tokens distinguish rejected test source from executable expressions.

`internal/native/testdata/regexp_cohere_patterns.json.gz` is input to
`TestRegExpCoherePatterns` and `TestRegExpCohereCorpusPin`. It combines the
inventory with constant Go regexp declarations, JavaScript/TypeScript scripts,
and expressions embedded in Go lint fixtures. Leading Go `(?ims)` modifiers
are represented as JavaScript flags. Subjects retain source and kind metadata:

- `go-trace`: actual matcher arguments from named cohere lint tests, captured
  through a scratch Go source overlay without changing cohere.
- `scanner-replay`: matcher arguments from cohere's original scanner replayed
  over embedded lint source. This does not establish original rule execution.
- `selector-trace`: actual splitting arguments from the stage1 selector port.
- candidate fixture strings: full source, lines, identifiers and literals;
  these are conservative input candidates, not per-call traces.
- `prior-generated` and `prior-fixture`: earlier corpus subjects for patterns
  still present, with every Node answer recomputed.

Pattern and input UTF-16 units preserve lone surrogates. The fixture contains
no recorded native or Node matcher answers. Tests recompute them with Node
24.19.0, compare captures, named groups, spans and lastIndex, and execute native
bounded, unbounded and forced-VM modes. Patterns without subjects receive
compile and bytecode-serialization checks only. Unresolved dynamic constructors
are outside this constant-pattern corpus.

Run the focused checks from the repository root with the configured toolchain:

```
go test ./internal/native -run '^TestRegExpCohere' -count=1 -timeout 30m
go test ./internal/oracle -run '^TestRegExpCohereInterval' -count=1
```

Keep extraction traces and test output outside the checkout. Source refreshes
must repeat the binder inventory, Go AST constant extraction and fixture trace
collection; a changed hash is not sufficient evidence of unchanged patterns.
The old generated `docs/regexp-cohere-patterns.md` is intentionally not copied.
