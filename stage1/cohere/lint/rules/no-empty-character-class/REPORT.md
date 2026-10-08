# no-empty-character-class

Ported from origin/codex/stage1-lint-batch2:stage1/cohere/lint/no_empty_character_class.ts, the batch2 source named by DEDUP_LEDGER.md. Base: origin/area/stage1-lint d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898. This branch adds only this descriptor directory. All new Adamic modules are .a.

The descriptor subscribes to RegularExpressionLiteral and uses node: true. The listener acts on its supplied node, without relevance kind comparisons or refetching that node. The message is moved verbatim. The local escape scanner ports regexsyntax.SkipPatternEscape, a regex syntax tokenizer, not a hand-built regex matcher. No Go regexp is present in the upstream rule, no pattern is executed and no dynamic RegExp is introduced. Nested v classes buffer spans until the outer class closes; UTF-16 spans derive from the literal end, preserving leading trivia and astral source characters. The wire encoder converts positions to Go bytes. There are no fixes or suggestions upstream.

The upstream rule has no options type and does not inspect options. The adapter decodes supplied field-5 JSON to a non-nil Go value instead of silently dropping it or weakening the options guard. No options-bearing witness is necessary for this optionless rule. Witness option sidecars require the separate shared witness-options integration and are not added here.

upstreamTest TestNoEmptyCharacterClass captures all nine real top-level Go tests: Fires, FiresInsideNestedClasses, StaysSilent, StaysSilentInsideNestedClasses, ReportsTheClassAndNotTheLiteral, ReportsPastLeadingTrivia, ReportsEachClassSeparately, ScansEachLiteralSeparately and SkipsAnUnterminatedClass. A direct run of that complete prefix passes. The unified oracle discovers the same prefix without shared corpus edits. Witnesses cover empty, nested and multiple classes, astral offsets, leading trivia, negated/escaped classes and the intentionally ignored constructor form.

Validation, all raw output in evidence:

- go run ./cmd/lint-registry: PASS.
- gofmt -l stage1/cohere/lint/rules/no-empty-character-class/oracle.go: empty output.
- go vet ./stage1/cohere/lint: PASS, empty output.
- In cohere: go test ./internal/lint/rules/core -run '^TestNoEmptyCharacterClass' -count=1 -v -timeout=10m: PASS 0.008s.
- ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^(TestRulesAgree|TestOwnedWitnesses|TestMutants)$/^empty_character_class_silenced$' -count=1 -v -timeout=20m: PASS 89.022s. TestRulesAgree PASS 46.78s; TestOwnedWitnesses PASS 22.85s, 95,460 identical bytes on Go, source Node, emitted JavaScript and ASan/UBSan native. The slash filter runs full non-subtest parity/witness checks and the one new mutant. The same area baseline's complete 45-mutant suite passed on the preceding landing branch, so inherited mutants were not repeated here.
- empty_character_class_silenced changes empty-span width from one to two interior offsets. It compiles, exits normally with empty stderr on all three paths, and only the Go output comparison catches it. TestMutants PASS 19.38s.

The first build refused an inferred empty array of never at the reportRange call. Using the context's typed default fixes this owned-code issue; the initial refusal log is retained. No shared parser, registry, context or comparator changed. No known blocker remains for this port. Full repository gate, 17 external-input correctness checks, full compiler source corpus and throughput are not run.
