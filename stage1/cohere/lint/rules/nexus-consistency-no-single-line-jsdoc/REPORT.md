# Unified harness port

Source: `origin/codex/stage1-lint-batch3:stage1/cohere/lint/rules/no_single_line_jsdoc.ts`, as named by DEDUP_LEDGER.md. Base: origin/area/stage1-lint at d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898.

The descriptor listens only to SourceFile and receives the dispatched node. Its private comment-range and anchor helpers preserve the batch's parser-reachable trivia scan, stepping over literal spans provided by RuleContext. No shared harness, driver, registry implementation, or RuleContext files changed.

The message is copied verbatim from cohere/policy/messages/consistency-no-single-line-jsdoc.json. Findings and automatic edits use reportRange; the rule has no suggestions. Repairs retain the description and are withheld for multiline comments, descriptions containing //, or code after the comment on the same line. The trim helper follows Go's Unicode White_Space rather than JavaScript trim, preserving Go's NEL/BOM behavior.

Upstream has no options type and ignores its options argument. The oracle adapter explicitly decodes field 5 as JSON instead of discarding it. `upstreamTest: TestConsistencyNoSingleLineJsDoc` captures all five real upstream test functions: Fires, StaysSilent, FixesWithOnlyWhitespaceAfter, Fixes, and WithholdsUnsafeFixes (19 asserted cases before source/options deduplication).

Owned witnesses cover a firing simple comment, a leading BOM boundary, string/regex/template exclusions, shebang, tags, unsafe fixes, trailing prose, UTF-16 positions, NEL and BOM. The mutant corrupts only automatic replacement text: `// ${description}` becomes `// MUTANT ${description}`. A successful run with different serialized fixes is the required catch.

Validation:

```sh
source /workspace/adamic-tools/env.sh
export TMPDIR=/tmp/adamic-gate
go run ./cmd/lint-registry
gofmt -w stage1/cohere/lint/rules/nexus-consistency-no-single-line-jsdoc/oracle.go
go vet ./stage1/cohere/lint/...
go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestRulesAgree|TestMutants)$' -count=1 -v -timeout=60m
go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestRulesAgree)$' -count=1 -v -timeout=10m
go test ./stage1/cohere/lint -run '^TestMutants$/^JSDoc_replacement_text_corrupted$' -count=1 -v -timeout=10m
```

Registry generation and vet exit 0; gofmt -l is empty. The complete registry mutant sweep caught all 41 mutants on Node, emitted JavaScript and sanitized native (1466.46 seconds). That run exposed a baseline owned-witness failure: the private batch helper missed a comment after a shebang. The fix was narrowed to that path after a separate Go comparison proved that Go's ASCII anchor guard misses comments immediately after a leading BOM. Both reproducer inputs are retained as witnesses; no Go behavior or shared code was changed.

After the helper correction, the complete TestRulesAgree passed on 2034 unique captured upstream combinations plus generated cases (13,060,121 identical output bytes, 85.85 seconds). TestOwnedWitnesses passed for every descriptor, including all four new witnesses both selected and in all-rules mode (99,482 identical bytes, 38.38 seconds). The affected JSDoc mutant was rechecked on this final source and caught on all three port executions (34.47 seconds), specifically `// MUTANT Trailing prose.` versus Go's `// Trailing prose.`. The other 40 rules and their mutations were unchanged; the full sweep is retained with its initial witness failure clearly marked.

An additional comparison ran all four rule-owned witnesses with absent options, {}, and {"Unused":true}: 12 rows, 13,046 identical bytes, exit 0 and empty stderr on both Go and sanitized native. A single process timing on those small rows measured Go 26.670 ms against sanitized native 68.437 ms (2.57x); it includes startup and is not a release-performance benchmark.

Logs are in validation/. Native executions use the unified harness's sanitizers; no check was skipped or weakened. Its existing malformed-input recovery refusals remain explicit in the upstream log. No JSDoc input hit the checker, dynamic-RegExp, parser, or Tailwind blockers. The full repository gate, large compiler corpus, and release-performance benchmark were not run.

Toolchain setup earlier in this unit reported ready 0s, warm 116s, total 116s; nproc is 5. Step 1 was skipped because all twelve checker-backed wave 27 ports are blocked on unified checker linkage; their names and reproducer were pushed in PARKED_UNIFIED.md on codex/typeaware-wave-27. This branch starts from the unified area tip and contains only this descriptor directory.
