Defended TestExactReceiverRejectsAssignments with D1 in a bounded 25-row matrix.
Starting origin/main: 619e7a4cf33741cc04bc78dc4c0c8ba0e31d73fb; all 282 listed top-level test names retained as scope evidence.
The target alone failed; its prior subsumer and 23 other selected rows passed.

CODE UNDER TEST: internal/native exactReceiverClass and its nested-statement traversal. ORACLE: the target's self-written expectation of zero allocation identity after a nested assignment. The subsumer checks self-written generated-C call forms and case counts; parts of its expected answer reuse production exactReceiverMethod.

The four relevant source components were read: both complete test bodies in devirtualize_test.go, devirtualize.go, walkStatement in reuse.go, and the subsumer's devirtualize.a program. The audit report, limitations, mutant table, rows, plan and matrix are preserved as audit-* files.

Per-test coverage used -coverpkg ./internal/native. Exclusive covered Go blocks: zero. The semantic difference is the assignment in ir.Try.Body in the target versus the top-level mutable receiver assignment in devirtualize.a. D1 drops only the recursive walkStatement call inside exactReceiverClass, keeping the shared walker unchanged. The resulting exact identity is 1 instead of 0. This is an erroneous production proof, not an oracle or harness mutation.

D1 location against starting main: internal/native/devirtualize.go:42.
Failing line: devirtualize_test.go:92: assigned receiver taken as exact: 1
Command: ADAMIC_BUILD_CACHE_DIR=/tmp/defend-native/cache/D1 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run "$(cat review/test-defend/internal-native-decode_ascii/matrix.regex)"
Standalone diff: D1.diff. go vet ./internal/native/ passed under D1, and git apply --check passed after restoration. The modified code under test is Go. No runtime C source changed. Each native product request used D1's own ADAMIC_BUILD_CACHE_DIR.

Passed rows:
- TestCEndsInNewline
- TestCallTargetsElementBorrowPlan
- TestClosureConventionDropCount
- TestClosureConventionRuntimeDropCount
- TestClosureConventionRuntimeFeaturesIgnoreLiterals
- TestClosureConventionWrongOrder
- TestDevirtualizeBorrowDocClaim
- TestDevirtualizedCalls
- TestGlobalArgumentLending
- TestLoopArrayHoldC
- TestLoopBorrowPlan
- TestLoopCallCoverage
- TestNbodyBorrowedLoopC
- TestNbodyIndexedElementsBorrow
- TestNoReaderCallingConvention
- TestNodeBufferRuntimeWithoutDeclarations
- TestOptionalMethodThunksMatchNode
- TestOptionalWriteMissingSlotRemainsChecked
- TestParserHasNoUnusedOptionalMethodThunks
- TestRegexProgramsKeepCheckedFieldReads
- TestRuntimeFieldLayoutsAreIncluded
- TestTSGoBuildSeesTheProgramsFeatures
- TestThrowElementBorrowPlan
- TestUniformFieldsMatchNode

Selection and limits:
- The whole clean package timed out at 90.050 seconds, with no earlier test failure. It was not treated as a red baseline. The bounded baseline passed in 6.035 seconds and the restored target/subsumer run passed.
- Matrix rows include every top-level test in current files that directly call C, cProgram, callCode, exactReceiverClass or exactReceiverMethod. This includes overinclusive sibling tests that only inspect lower-level plans. reachability.txt preserves the search. Production callers of exactReceiverClass are exactReceiverMethod and callCode, reached through C emission. Expensive runtime-only decoder/regexp suites were excluded. This is bounded evidence, not an executed whole-package uniqueness claim; results outside the matrix remain unknown.
- The package now lists 282 tests. The audit's 94 requested functions were a slice, so that number is not a package-growth comparison. Selection was made from the current tests, not the historical files alone.
- No target or selected row requires a WASI opt-in. The full baseline's skipped runtime rows are recorded in baseline.log. WASI-only runtime decoder probes do not execute the modified Go emission proof.
- /tmp has only 8.8 GB total capacity, so 15 GB free there cannot be achieved. Earlier-unit /tmp/defend-views scratch was removed. /workspace initially had 15,708,762,112 bytes available. There was no disk failure.
- Shared lines are explicitly allowed as a semantic defense in the brief; zero exclusive blocks did not prevent a production mutant from distinguishing these rows.
- One successful aimed mutant is sufficient under the up-to-three rule. No additional mutants were needed. No test was changed, weakened or deleted, no other package tests ran, and no PR was opened.
- The test name and assertion agree: it checks rejection of exact identity after an assignment, using one nested assignment witness. It does not promise comprehensive coverage of every assignment form.

Warm tool setup skipped: 0 seconds; nproc 5. npm ci completed successfully. Binary times are in timings.json. The whole-package budget timeout and the Go recompilation/coverage overhead were the material time costs. Production sources are restored. Only evidence is committed and pushed.
