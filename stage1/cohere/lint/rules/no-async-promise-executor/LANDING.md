Status: rebased onto landed area harness and deduplicated; owned witnesses and mutants green, full upstream comparison blocked.
SHAs: area base 7481e0324; harness 41eb6eab2; pre-dedup rules c74d385b6.
Checks: TestOwnedWitnesses PASS 21.08s, 57,889 identical bytes; TestMutants PASS 408.47s; go vet exit 0. Combined package run exits 1 in 512.402s because TestRulesAgree fails.
Mutants: identifier_expression_ignored, parameter_difference_silenced, empty_method_silenced, async_executor_ignored, closing_brace_missing, negative_zero_ignored caught on Node, emitted JavaScript and sanitized native; all successful executions have empty stderr.
Uncovered: complete upstream findings, fresh compiler/stage1 corpus, throughput and full gate; invalid decorated async executor refuses AtToken at 12.

## Ledger applied

Six losing copies are removed entirely: structure/tailwind-no-physical-direction
belongs to wave1-05; @next/next/no-assign-module-variable and
@typescript-eslint/default-param-last belong to wave1-15;
@typescript-eslint/no-unnecessary-type-constraint, prefer-as-const and
prefer-enum-initializers belong to wave1-08. Historical proof is retained in
commit c74d385b6. The record dependency moves to the retained unused-expression
directory and reports complete suggestions through the landed reportRange API.
No shared harness or registry file is edited. Existing named kinds are retained.

Six retained ports: @typescript-eslint/no-unused-expressions,
@typescript-eslint/unified-signatures, class-methods-use-this,
no-async-promise-executor, no-case-declarations, no-compare-neg-zero.
The ledger names no slot-03-specific batch-only assignment.

## Shared comparison boundary

The unchanged TestRulesAgree stops at captured AsyncPromiseExecutor.ts containing
new Promise(@dec async () => {}). Adamic's parser refuses unsupported primary
AtToken at offset 12. This was already an explicit recovery gap in the owned
pre-integration proof; the unified capture does not mark this case as unsupported
recovery. No fixture is removed, Go verdict changed or shared comparator patched.
TestOwnedWitnesses compares selected and all-rule runs against Go, including both
brace insertion edits in no-case-declarations suggestions. All 21 registered
mutants (six owned and fifteen inherited) pass their comparison checks.
Raw outputs are under evidence/landing. The branch is pushed but not claimed
fully oracle-green; full rule landing remains blocked by the shared parser and
recovery comparison boundary. No new work is claimed.

## Area d65a8f931 recheck

Rebased cleanly onto d65a8f931c98655936ae04c6899f38f14862b73e, including its runtime performance changes. The dedup ledger and shared recovery handling are unchanged. Uncached TestOwnedWitnesses passed in 21.26s with 57,889 identical bytes; the six owned mutants passed on Node, emitted JavaScript and sanitized native. go vet passed. TestRulesAgree remains red in 46.46s on unsupported primary AtToken at offset 12 in the invalid decorated async executor. Combined package execution exits 1 in 184.895s. No full oracle-green claim, new reservation, throughput or full gate. Exact command: ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^(TestRulesAgree|TestOwnedWitnesses|TestMutants)$/^(identifier_expression_ignored|parameter_difference_silenced|empty_method_silenced|async_executor_ignored|closing_brace_missing|negative_zero_ignored)$' -count=1 -v -timeout=20m.

## Area b84a9d931 recheck

Rebased cleanly onto b84a9d9314b65d3d0261ee017e233287b4f071da, including main c7991b900 lowering and record-runtime changes. The ledger and shared recovery handling are unchanged. TestOwnedWitnesses PASS 20.24s, 57,889 identical bytes; six owned mutants PASS 116.69s on Node, emitted JavaScript and sanitized native; go vet passed. TestRulesAgree remains red at the decorated async AtToken fixture in 44.50s. Malformed-interface timeout probes on both Node and native also occur in the preceding d65 run; they are existing shared parser recovery limits. No check is skipped, relaxed or deleted to obtain green. Raw b84 logs retain the combined run timing.

The helper branch is re-green on the same area and pushed at b00ae6f63: 21,019 cases PASS 73.208s, six output-only mutants caught on all paths, vet passed, CFG self-edge still refuses adamic/cycle-capable. No new claims, required external-input checks or full gate.

## Area b46914832 recheck

Rebased cleanly onto b46914832d70e00847d82d5d221ab7bb24040c53. The dedup ledger is unchanged. The upstream volume rules now have individual registrations. Uncached TestOwnedWitnesses PASS 22.08s with 102,814 identical bytes across Go, Node, emitted JavaScript and sanitized native; all six owned mutants PASS 116.25s; go vet exits 0. TestRulesAgree remains FAIL 45.04s on new Promise(@dec async () => {}), unsupported primary AtToken at 12. Combined package FAIL 183.392s. No shared file changed, no new claim, no full parity claim. Full gate, required external-input checks and fresh throughput remain unverified. Raw commands and results are retained in b469 logs.

## Area d3a37422c recheck

Clean rebase onto d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898, containing main b6b1538b0cebc4ba6741ac34f1aedb60293c1d06. Ledger unchanged. Uncached TestOwnedWitnesses PASS 21.91s, 102,814 identical bytes on Go, source Node, emitted JavaScript and sanitized native. Six owned mutants PASS 111.38s, caught only by output comparison. go vet exits 0. TestRulesAgree still fails on new Promise(@dec async () => {}), unsupported primary AtToken at 12. Combined package FAIL 174.413s. No shared file altered or check skipped, relaxed or removed. Helper branch 5e64fe3cf is pushed on this base: 21,019 cases PASS 74.928s, all six helper mutants across three paths and vet pass, CFG self-edge still refuses adamic/cycle-capable. No new claims. Fresh full compiler corpus, full gate, 17 required external-input correctness checks and throughput remain unverified.
