Built: nexus/consistency-no-multiline-arrow-function as an independently registered .a module, using the handed node and rule-local conversion helpers.
Commits: new branch codex/lint-port-nexus-consistency-no-multiline-arrow-function starts from area bb2ece564 and is rebased onto shared harness fix 4f18a05c9; source batch3 is fa9781c2c.
Commands: lint-registry, gofmt, vet, TestRulesAgree, TestOwnedWitnesses, the owned TestMutants subtest and the real-input TestCompilerAndStage1Agree all pass; complete logs are under evidence/.
Mutant: removing the single-line exemption compiles and finishes cleanly, then independently disagrees with Go on source Node, emitted JavaScript and ASan/UBSan native.
Not covered: shared explicitly unsupported parser recovery, other required correctness packages, other registry mutants on this new branch, new throughput and the full gate. No known blocker prevents this rule.

## Rule and provenance

The assigned ledger source is origin/codex/stage1-lint-batch3:stage1/cohere/lint/rules/no_multiline_arrow_function.ts at fa9781c2c0911c52312002ba97ffd4b560d178ae. Adapted its listener and arrow conversion without copying the old batch context or dispatcher. The descriptor listens only to CallExpression and ArrowFunction and has node: true. The handed root node supplies its children and end; descendants and ancestors are read only where the semantic judgment needs them. There is no per-file rule-selection check or global tree walk in the listener.

The existing RuleContext provides parent, name, scanner, unwrap and line services. Only the first/last-child helpers and Go strings.TrimSpace equivalent needed by this rule live in helpers.a. There is no shared helper addition or shared source edit. Go regexes are absent from this upstream rule.

All three messages are copied verbatim from cohere/policy/messages/consistency-no-multiline-arrow-function.json. Hook-name substitution preserves the upstream message. Findings, spans, fix eligibility, signature text, function-expression wrapping, type parameters, async modifiers and inherited-binding exemptions are compared against the unchanged upstream Go rule.

The upstream rule accepts options any but reads none and declares no concrete options type. Its owned adapter decodes manifest field 5 as an actual JSON value into that upstream any parameter. It does not return nil for an object-bearing row. testdata/witness.options.json supplies {} and is consumed by the integrated witness-options harness. The shared nil-options guard is unchanged.

## Upstream test coverage

The upstreamTest prefix is TestConsistencyNoMultilineArrowFunction, covering every actual test name in cohere/internal/lint/rules/nexus/consistency_no_multiline_arrow_function_test.go:

- TestConsistencyNoMultilineArrowFunctionFires
- TestConsistencyNoMultilineArrowFunctionStaysSilent
- TestConsistencyNoMultilineArrowFunctionReportsEachArrowOnce
- TestConsistencyNoMultilineArrowFunctionFixes
- TestConsistencyNoMultilineArrowFunctionDeclinesAFixThatRebinds
- TestConsistencyNoMultilineArrowFunctionFixesWhatOnlyLooksInherited

All six groups are captured, including bare and typed parameters, TSX disambiguating commas, arrow types inside signatures, import.meta, async generic arrows, expression-leading wrapping, this/arguments/new.target/super refusals and property-name/nested-function exemptions. The firing witness also includes a single-line clean control; the mutant incorrectly fixes that control and the Go-output comparison catches it.

## Commands and measured output

With source /workspace/adamic-tools/env.sh:

```
gofmt -w stage1/cohere/lint/rules/nexus-consistency-no-multiline-arrow-function/oracle.go
go run ./cmd/lint-registry > /tmp/arrow-fixed-registry.log 2>&1
go vet ./stage1/cohere/lint ./stage1/cohere/lint/registry > /tmp/arrow-fixed-vet.log 2>&1
go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestRulesAgree|TestMutants)$/(multiline-arrow-single-line-exemption-removed)$' -count=1 -v -timeout=20m > /tmp/arrow-fixed-tests.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 go test ./stage1/cohere/lint -run '^TestCompilerAndStage1Agree$' -count=1 -v -timeout=20m > /tmp/arrow-compiler-stage1.log 2>&1
```

TestRulesAgree passes in 58.21s: 2,057 unique Go source/rule/options combinations and 13,071,239 identical bytes. TestOwnedWitnesses passes in 23.98s: 116,856 identical bytes across selected and all-rule witness rows. TestMutants passes in 25.33s with the owned semantic mutant caught in all three modes after clean compilation and execution. Combined selected suite passes in 107.538s.

The actual TypeScript checkout is pinned to 050880ce59e30b356b686bd3144efe24f875ebc8. TestCompilerAndStage1Agree passes without skipping in 120.469s: 398 actual compiler/stage1 files, 21,285,967 identical Go bytes on source Node, emitted JavaScript and sanitized native. Vet passes with empty output; registry generation includes this public rule. No selected correctness check skips. Initial bb2ece564 comparison logs are retained separately as historical evidence.

The prior setup environment is reused (logged setup 128s, nproc 5). Publication writes only this worker branch; no PR, main push or area push is made. The landing branch's delay parser reproducer remains separate and this new port does not include either the parked delay rule or unrelated wave13 ports.
