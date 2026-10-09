Ported no-caller from batch 2 onto the unified harness with a handed-node listener, verbatim message and property-span finding.
Base: d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898; source: 6700c572ee3810524e356af7f92973e38ee5265f:stage1/cohere/lint/no_caller.ts; publish only codex/lint-port-no-caller.
Registry, gofmt and vet PASS; TestRulesAgree and TestOwnedWitnesses PASS in 161.095s, comparing all four backends.
TestMutants/no-caller-property PASS: compiled mutation caught on Node, emitted JavaScript and sanitized native; adapter payload check PASS.
No checker, dynamic RegExp, parser or Tailwind blocker for this rule; full repository gate and other rules' mutants not rerun.

The descriptor subscribes to PropertyAccessExpression and declares node:true. The driver hands the access node to visit; the rule reads only its receiver and property children, using the shared unwrap helper for parentheses. No shared context, dispatch, registry source, oracle, corpus or compiler source is edited. New Adamic modules use .a. Messages are moved verbatim. Upstream has no fixes or suggestions for this rule.

The real upstream test functions are TestNoCallerFires, TestNoCallerStaysSilent, TestNoCallerReadsTheReceiver and TestNoCallerReportsTheProperty. Prefix TestNoCaller captures all four, including tests that inspect ranges directly. The owned witness fires on both forbidden properties, nested parenthesized arguments and optional access, and includes clean computed and other-receiver accesses. Go and all backends agree on selected and all-rule rows.

Upstream NoCaller defines no typed options struct and ignores its any parameter. The adapter therefore decodes field 5 as JSON into any, preserving explicit payloads instead of returning nil. The dedicated Go overlay test verifies empty object, a nonempty object, array, absent payload and malformed JSON. No options-dependent rule behavior is invented. Witness-options support (29c41e102) is not yet an ancestor of this base; no options sidecar is needed for this optionless rule. The first direct adapter-test invocation failed on Go's internal-package boundary; using the same cohere overlay mechanism as the shared oracle passed. Both logs are retained.

Commands, all test output directly to logs:

```sh
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
go run ./cmd/lint-registry
gofmt -w stage1/cohere/lint/rules/no-caller/*.go
gofmt -l stage1/cohere/lint/rules/no-caller/*.go
go vet ./stage1/cohere/lint/...
go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestRulesAgree)$' -count=1 -timeout 30m -v
go test ./stage1/cohere/lint -run '^TestMutants$/^no-caller-property$' -count=1 -timeout 30m -v
cd cohere
go test -overlay=/tmp/no-caller-options-overlay.json adamic_no_caller_options.go adamic_no_caller_options_test.go -count=1 -v
```

TestRulesAgree: 141.21s, 13,055,010 identical bytes. TestOwnedWitnesses: 19.77s, 92,342 identical bytes. The existing shared corpus explicitly records malformed parser-recovery limitations unrelated to no-caller; those remain refusal tests rather than lint-parity claims. The no-caller-property mutation replaces callee with length in the forbidden-name condition, so it compiles and runs but drops real callee findings. Caught on all three port backends: TestMutants 62.623s (owned subtest 29.65s). No compiler failure is counted as killing this mutant.

Setup: Go ready 0s; clang, Node and submodules ready 1s; cache warm 137s; done 137s; nproc=5 (four-CPU quota, 17.6GB). Disk space filled during build-cache warming. Removed obsolete wave-24 worker binaries and only generated Go cache data older than two hours and larger than 10MB (526 entries, 22,084,181,368 bytes), preserving current source, toolchain and observation logs. The requested tests complete successfully. No full-gate certification is claimed.
