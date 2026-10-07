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
