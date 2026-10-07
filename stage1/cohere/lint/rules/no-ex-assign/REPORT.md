# no-ex-assign unified harness port

Source: origin/codex/stage1-lint-batch2 at 6700c572ee3810524e356af7f92973e38ee5265f,
stage1/cohere/lint/no_ex_assign.ts (DEDUP_LEDGER's batch2:no_ex_assign.ts).
The message is copied verbatim. The adapter calls unchanged core.NoExAssign.
Upstream defines no options struct and reads no options; the adapter validates
field 5 as JSON into an explicit empty struct instead of returning nil.

The descriptor subscribes to CatchClause, node: true, and consumes the handed
node. Its recursive write search stays inside the catch body, as unchanged Go
requires, including nested functions and catches. Plain identifiers only;
destructured catch bindings and property writes are excluded. All assignment
operators and prefix/postfix updates count. The full TestNoExAssign prefix
captures Fires, StaysSilent, SeesUpdateExpressions and
IgnoresReadingUnaryOperators. No helper or shared registration edit was needed.

Step 1: existing owned type-aware ports cannot be integrated into the unified
harness without the checker/CFG driver. Their parked note and minimal probes
remain on codex/typeaware-wave-01. This branch starts from area/stage1-lint and
contains none of those implementations. There were no unblocked owned
unified descriptors to land separately.

Validation results are recorded below after all requested checks complete.

## Validation

Base: origin/area/stage1-lint d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898.
All commands sourced /workspace/adamic-tools/env.sh.

- go run ./cmd/lint-registry: PASS, descriptor discovered.
- gofmt -w and gofmt -l on oracle.go: PASS, no remaining format output.
- go vet ./stage1/cohere/lint/... and go vet ./...: PASS, empty logs.
- go test ./stage1/cohere/lint -run
  '^(TestOwnedWitnesses|TestMutants|TestRulesAgree)$' -count=1 -timeout 30m -v:
  PASS, 897.278 seconds.
- TestRulesAgree: PASS in 98.19 seconds, 2036 captured source/rule/options
  combinations, 13060392 identical output bytes on Go, Node, emitted
  JavaScript and sanitized native. Existing explicit parser-recovery
  refusals were exercised without relaxing their checks.
- TestMutants: PASS in 776.93 seconds, the complete discovered mutant list.
  miss-postfix-writes was caught on Node, emitted JavaScript and native
  by omitted unexpectedExceptionAssignment at range 22..27 of
  try {} catch(error) { error++; const a = -error; const b = typeof error; }.
  The mutant compiled and ran; the output comparison caught it.
- TestOwnedWitnesses: PASS in 22.14 seconds, all owned witnesses caused Go
  findings; selected and all-rule output matched across all four paths,
  98281 identical bytes.

Raw logs are under evidence/. No shared harness files were edited.
No known checker, dynamic RegExp, parser or Tailwind blocker hit this rule.
The full repository test gate and external compiler/parser checks were not
run; this report certifies the requested unified lint gates only.
