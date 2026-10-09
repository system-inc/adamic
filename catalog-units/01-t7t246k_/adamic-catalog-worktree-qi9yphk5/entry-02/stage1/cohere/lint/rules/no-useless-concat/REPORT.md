Ported no-useless-concat from batch2 onto the unified registry with a node listener and verbatim Go message.
Commits: this rule's commit contains only its descriptor directory; base origin/area/stage1-lint d3a37422.
Commands: lint-registry and vet PASS; TestRulesAgree 256.09s, TestOwnedWitnesses 20.34s, owned TestMutants 20.11s PASS.
Mutant: omitting TemplateExpression from literal operands is caught by Go comparisons on Node, emitted JavaScript and sanitized native.
Uncovered: full repository gate, compiler/stage1 corpus and throughput profiling; inherited parser recovery limitations remain explicitly reported by the shared harness.

Source: origin/codex/stage1-lint-batch2:stage1/cohere/lint/no_useless_concat.ts, verified against pinned Go cohere. Message moved verbatim into messages.a. Descriptor kinds is BinaryExpression and node is true; visit receives the expression without refetching it or testing its kind for relevance. Operand helpers descend only adjacent concatenations and skip parentheses. Findings point at the operator, with no fix or suggestion. Go's gap test rejects CR and LF; the port preserves that exact test, including multiline templates.

Go declares no options type and ignores its options argument (empty schema). The adapter JSON-decodes explicit field-5 rows into the argument instead of returning nil or inventing settings. Default witness needs no options sidecar. No shared files or guards changed.

The TestNoUselessConcat prefix covers all seven actual upstream test functions: Fires, StaysSilent, SkipsParentheses, TemplatesWithSubstitutions, SameLine, Message, OtherOperators. TestRulesAgree captured 2057 unique combinations across the discovered set. The three-runtime comparisons preserve descriptions, positions and fixed-source output byte for byte.

Commands, from the repository root with /workspace/adamic-tools/env.sh sourced:

```
go run ./cmd/lint-registry
gofmt -w stage1/cohere/lint/rules/no-useless-concat/oracle.go
go vet ./stage1/cohere/lint ./stage1/cohere/lint/registry
go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestRulesAgree)$|^TestMutants/concat_substituted_template_ignored$' -count=1 -v -timeout=20m
```

TestMutants is filtered to the owned semantic mutant; inherited mutants were covered in the preceding landing gate, not repeated here. This mutant compiles and runs; the comparison notices the lost first witness finding for a substituted template. All three runtimes caught it. See evidence/gate.log for the differences.

Setup: Go ready 0s; clang ready 1s; Node ready 1s; submodules ready 18s; build cache warm 347s; total 347s; nproc 5. Setup exits zero. Pinned cohere 715ba94f3608a6500086b1076ce5cb7e51b836db and TypeScript 8d550c837c90bd1805b047b7eeccc2baac2d5e7a. No branch was pushed to main or area.
