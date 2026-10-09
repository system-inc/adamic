Both requested rows are defended. Keep them.
D1 and D2 each have exactly one failing row and ten passing rows in completed full-package runs.
Evidence starts from origin/main b8bcadb2c493173855f19d7e5c508b34f5eeb5b6; tests and production source are unchanged.

Read rows.json for verdict objects and every passing row; matrix.json for commands, failures, complete scope and timings; code-and-oracle.md for the oracle and semantic differences; coverage-differences.json for individual coverage comparisons; and *.functions.log for all reached IR functions. All eleven current top-level tests ran, including TestMixedOptionalFixedLayoutPreservesMaybeNumber and TestUnknownClosurePreservesThrowingEffect, added since the audit. None skipped. No panic, timeout, narrowing or unknown matrix cell occurred.

| ID | Starting-main file:line | Change | Sole failure |
|---|---|---|---|
| D1 | internal/ir/call_targets.go:102 | Invert ClosureArgumentLayout(call).Count in the wrapper return | TestClosureArgumentsCountTargets |
| D2 | internal/ir/call_targets.go:15 | return targets -> return targets[1:] | TestCallTargetsIncludeEveryDescendant |

D1: arguments_length_test.go:13: type 10: true, want false.
D2: call_targets_test.go:36: missing Base override 3 in [9 12 15]; line 40: want four implementations, got [9 12 15].

D1 is a condition flip on a wrapper block that TestArgumentLayouts never reaches. D2 changes an implicit slice lower bound from zero to one on shared code. The effect test's synthetic virtual sets begin with a nonthrowing implementation, so omitting it does not change any asserted may-throw result. The integration descendant test checks exact membership and cardinality. This is a semantic defense even though CallTargets query blocks are shared. There are no survivors and no requested row remains undefended.

Brief issues and actual costs:
1. /tmp has only 8.8 GB total capacity, so the required 15 GB free threshold cannot be achieved. Earlier scratch directories were removed; /tmp became nearly empty. /workspace had 17 GB free. No disk-full failure occurred.
2. The audit report is REPORT.md, not report.md. The first case-sensitive report read failed and was corrected before selecting mutants; PRE-MUTATION.md and rows.json were also read.
3. Current main adds two tests. Both were included, rather than replaying the audit's nine-row census.
4. Exclusive coverage alone does not prove the descendant row's value. Its target-membership oracle detects a nonthrowing omission that the shared-code effect oracle cannot detect. The full matrix proves this observed distinction.
5. TestCallTargetReaders runs Go package/type discovery on other compiler packages as its own construction check. Its export compilation contributes substantially to the 29-45 second binary times even though the query assertions are fast. No tests from another package were run and no separate native product was built.
6. The argument-count row is about argument cardinality, not performance. The cost-row rule does not require an artificial slowdown mutant here. Neither row is an executor twin.

There is no unsupported name/assertion promise for an undefended row because both rows are defended. The descendant row asserts set membership, cardinality, direct-call behavior and closure-binding proofs; it does not assert executed dispatch output. ClosureArgumentsCountTargets checks the public wrapper's boolean answer for known reader, known nonreader and unknown type evidence.

Warm setup was skipped; npm reported one second; nproc=5. Baseline binary: 28.968 seconds. D1 binary: 44.738 seconds; D2 binary: 38.264 seconds. Matrix wall time including vet and compilation: 117.252 seconds. Every standalone diff applies against starting main, passes go vet ./internal/ir/, and compiled in the full package test run. No other package tests, main push, PR, test edit or production change was made.
