# ESTree syntax defense

Keep both tests pending full replay. Each caught an aimed production break missed by TestSyntaxGrammar and all six other replayed checks. Strict package-wide uniqueness remains cannot-judge. No test was edited.

Starting origin/main: 92d196011e78208b45b3768dcf2c8c88e7ec132d. Audit base: 3a1c8b5792fb66503290d2d494ce96302bc4386d. All 245 top-level names still exist, with no added names. Read and preserved audit report, rows, inventory and menu. Pre-mutation code and oracle declaration: code-and-oracle.md.

## Differences and coverage

Go coverage used -coverpkg=./internal/lower,./internal/native,./internal/javascript with isolated -run and -coverprofile. Each target has 5014 blocks and zero exclusive blocks versus Grammar. These profiles cover compiler execution, not the TypeScript port. Supplemental NODE_V8_COVERAGE measures runtime port execution. Complete compressed records, original transformed sources and exclusive ranges are preserved.

Refusals uniquely reaches invalid reference resolution mode in referenceError, sourceReferences.ts, transformed offsets 2368..2594. D1 drops the invalid-mode refusal return at original line 79. The port admits the malformed directive and emits 487 bytes while independent Go cohere refuses it. Grammar exercises accepted or ignored trailing references instead.

Cooked uniquely reaches lone-surrogate handling in written, protocol.ts. D2 changes <= 0xdfff to < 0xdfff at original line 13. U+DFFF now emits a surrogate escape instead of Go cohere's three replacement escapes. Grammar does not observe that boundary. D1 and D2 are respectively drop statement and off-by-one menu changes. See plan.json and standalone diffs.

## Matrix and validation

Whole clean package: timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run . with library/benchmark opt-ins enabled. It timed out during TestAcceptanceGrammar at 90.017 binary seconds, 92.131 command wall, without observed assertion failure. This is over budget, not a red baseline. Isolated clean coverage runs passed: Refusals 27.750s, Cooked 25.169s, Grammar 28.546s (command wall total 91.667s).

Bounded matrix: TestSyntaxRefusals, TestCookedSurrogates, TestSyntaxGrammar, TestGeneratedAgreement, TestThroughput, TestLossyInputRefusal, TestCookedSurrogateMutant. These directly exercise the reference/serialization pipeline; the last is a witness control. Other package tests can reach it indirectly and remain unknown. Each matrix row ran alone under timeout 120 with -json -count=1 -timeout 90s, with cache /tmp/defend-estree/cache/D1 or D2. All 14 runs completed without timeout or panic. D1 failed only Refusals; D2 failed only Cooked. Exact commands, failing lines, passed rows and timings: results.json and rows.json. No survivors in this bounded matrix.

Both standalone diffs apply cleanly to starting main. Both compile with the actual native port builder during matrix execution: runtime comparison failures occur after native build. Source was restored after each mutant. Native rebuild and runtime costs are combined in recorded command wall time because the test harness rebuilds for each row. No separate rebuild-only timing is claimed.

## Ambiguities, costs and uncovered work

The brief allows narrowing big packages, but defines defended as no other package row catching the mutant. We conservatively report cannot-judge for package-wide uniqueness, with positive bounded unique kills. This is keep evidence, not confirmation of subsumption or three unsuccessful attempts.

Go profiles do not measure TypeScript execution, requiring supplemental V8 coverage. Exact production file URLs were selected; V8 offsets are UTF-16, with ASCII critical mapped fragments. The schema has no bounded field, so supplemental fields disclose the scope.

Fresh main differs from the audit base; no top-level name was added or vanished. The broad baseline exhausted 90 seconds before later tests. Isolated runs avoid hiding later rows but repeat native rebuilds.

Refusals checks generic parser-error labels, status count, empty stdout and timely nonzero exit, not exact diagnostic cause. Another parser error could satisfy it. Its name does not promise that stronger check. Cooked compares full bytes against independent Go cohere, including its replacement convention, not independently specified ECMAScript UTF-16 behavior. Neither target promises unasserted performance; neither is an executor twin or cost row.

Warm toolchain setup skipped, nproc=5. npm ci ran in stage3/api; pinned independent oracle dependencies installed under /tmp/defend-estree/library. Dependency logs are saved. Library and benchmark opt-ins enabled. External optional corpus was not supplied. No bounded matrix row skipped. The other 238 top-level functions, external corpus and repo-wide uniqueness were not covered.

Matrix command wall seconds including 14 native rebuilds: 418.006

Restored combined targets and former subsumer passed: [('TestSyntaxGrammar', 24.84), ('TestSyntaxRefusals', 27.72), ('TestCookedSurrogates', 32.36), (None, 84.928)]
