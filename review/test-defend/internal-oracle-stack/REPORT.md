Defended within the bounded stack/recursion/new-test matrix: D1 fails only the long-input row.
The former subsumer, recursion neighbors and all five added top-level tests pass.
Keep the row; complete package and repo-wide uniqueness remain unproven.

# Stack-limit defense

Starting main: 7b890befc8fc43fa57435c0959510d13349dd31c. Branch: test-defend/internal-oracle-stack. Audit base: 6c60da091afddc9c2fe88b3a1067845b6dc79cb3. Read CLAUDE.md, full audit REPORT.md, rows.json and report.json, runtime/reached-function notes, both stack test bodies, native stack.c and ADAMIC_CHECK_STACK, and the fixture source. Audit copies are retained. Required repository documentation was read earlier in this warm session; its content has not changed relative to the audited branch.

CODE UNDER TEST is native runtime stack-limit initialization, stack checks and overflow/panic path. ORACLE is unchanged original TypeScript run under Node. The long row compares complete stdout, stderr and exit against sanitized and release builds. Its former subsumer uses Node at 1 MiB and a self-written panic contract at 512/256 KiB. Neither fixture, Node runner, comparison, harness nor test was mutated. The pre-mutation declaration and fixed aimed menu are saved in CODE-AND-ORACLE.md and plan.json.

## Space, setup and scope

First action: df -h /tmp /workspace. /tmp had 8.2G free and /workspace 14G free. Removed only prior-unit /tmp/defend-css and /tmp/adamic-gate. After cleanup /tmp had 8.6G free. Its entire 8.8G capacity is below the requested 15G; /workspace remained 14G free. No tools or repository removed. TMPDIR recreated with mode1777. No disk-related baseline failure occurred. Exact figures: DISK.md and disk-final.log.

Warm env.sh worked, so setup was skipped. npm ci in stage3/api ran before baseline. nproc=5; versions and cache configuration are in environment.json. Current go test -list has 198 top-level functions versus audit's 193. The five added names are TestFractionalPowersReachRuntime, TestReviewProgramsAgreeWithNode, TestReviewProgramsRefuse, TestReviewProgramsNoLooseFiles, TestReviewProgramsSelfTest. No names vanished. All five were included in passing clean and D1 replay. Exact complete scope: scope.json.

## Baseline correction and narrowing

I initially enabled optional ADAMIC_NATIVE_SPLIT=1. TestEntriesRuntimeReadiness and two RegExp mutant families construct C snippets that this optional splitter rejects as unsupported declarations. This was a setup choice I introduced, not mutation evidence. No mutant was run on that baseline. Split mode was removed and the default native build mode rerun without source changes. The corrected whole package hit its 90-second limit at 90.090 binary seconds with no observed assertion failure. It was narrowed rather than extending the budget.

Clean isolated coverage runs passed in 0.564s and 0.474s binary elapsed. Clean reached stack matrix passed; clean three recursion neighbors passed; all added top-level tests passed in 43.637s binary time. Later whole-package tests not reached are unknown. Optional WASI and other opt-in jobs were not newly enabled for this host-native stack defense.

## Coverage and semantic distinction

Required per-test profiles used -run anchored names, -coverpkg=./internal/native and -coverprofile with ADAMIC_GATE_UNCACHED=1. Both rows cover exactly the same positive Go native blocks: zero exclusive blocks either way. C runtime execution is outside Go's profiler; no dynamic runtime-exclusive-line claim is made. Function coverage and profiles are preserved. The lead is semantic on shared code, expressly allowed by the brief.

The long row passes three 120000-byte strings as arguments, arguments with an entirely empty environment, or environment variables. The subsumer uses small stacks at 1024/512 KiB with no long strings, and 256 KiB with only one 100000-byte argument. 128 KiB lies above 100000 bytes and below 360000 bytes. The sole executed mutant changes stack.c:58 from reserved = arguments + margin to reserved = ARGUMENTS_FLOOR. It under-reserves stack for long initial inputs without altering the shared panic implementation or source program.

## Result and replay

D1 is an aimed constant change. It fails all three long-row subcases. In arguments_alone, stack_test.go:75 reports plain: exit codes differ: Node exits 70 with start and the expected RangeError panic; the release native binary exits -1 with empty output. Sanitized native exits 1 with AddressSanitizer: stack-overflow. Full lines and source locations are in D1-stack-rows.log. These are observed child failures, not a Go panic aborting later matrix rows.

TestSmallStacksStillPanic passes all three size subcases. TestNativeAgreesWithNode passes selected stack_overflow.a, library_fnexpr_recurse.a, route_targets_recursive.a fixtures. All five added top-level tests pass too. Complete passed subcases are recorded in passed-rows.json; the top-level matrix is eight rows. The corpus row was only sampled at these three explicitly listed fixtures; its other members are unknown. Stack initialization is shared by native products more broadly, so this matrix does not prove package-wide uniqueness. The defender verdict is bounded as allowed by the big-package replay rule, not a claim that all native products were replayed.

D2 and D3 were fixed in the initial plan but not planted because the first attempt established the distinct catch. No survivor among executed production mutants in this bounded matrix. No empty-answer probe, supplemental insertion or comparison weakening was used. The stack-limit row is defended, so the rule requiring a cost attempt before a not-defended verdict does not apply.

The standalone D1.diff applies cleanly to starting main. It was compiled separately using both the actual release and sanitizer clang flags from native.Flags. Existing builder flags suppress unused variables; leaving the now-unused reservation temporaries is accepted by the specified toolchain, not a compile-error kill. Actual sanitized and release products were also built and executed by the tests. ADAMIC_GATE_UNCACHED=1 disables observation reuse and D1 has its own ADAMIC_BUILD_CACHE_DIR=/tmp/defend-stack/cache/D1. Runtime artifacts additionally key embedded source content. Source was restored; a final clean two-row run passed. No test or production change is committed.

## Brief ambiguities, costs and owner findings

* A 15G free-space threshold cannot be reached on the 8.8G /tmp mount. Cleanup still prevented space pressure.
* Optional split mode was an unnecessary setup choice and caused the first baseline's harness-snippet build failures. It was corrected before any mutant, and both logs are preserved rather than calling those failures mutation kills.
* The audit predates five added tests. The matrix includes them, and lists its sampled corpus fixtures rather than implying the whole corpus row ran.
* Go coverage cannot measure the C runtime. Zero exclusive compiler blocks does not rule out different argument/environment occupancy. The semantic mutation demonstrated that distinction.
* The whole package exceeds 90 seconds. Runtime stack initialization is broadly shared, so bounded uniqueness must not be presented as full package or repo uniqueness. Other package rows and unsampled corpus fixtures remain unknown.
* The supplied audit evidence command was truncated. The fetched full report supplied its command and bounds.
* Both stack rows use original Node, but are not executor twins: their inputs and assertions differ. Below 1 MiB, the subsumer's expected panic is self-written.
* The long-row name matches its graceful-failure assertion under large arguments/environment. It does not measure maximum usable recursion depth or reserved byte count, and promises no throughput threshold. Assertions check complete output and exit, not merely that a process failed.
* Native recompilation and execution are combined in test command timings. No unsupported separate compiler-only timing is claimed.

No deletion, rewrite, weakened assertion, main push or pull request was made. Linux default native behavior was exercised; Darwin and WASI alternate stack initialization branches were not covered. Whole-package replay and repo-wide replay were not completed.

Costs: total bounded runner command wall 91.403s, including clean neighbors, standalone compiler validations, D1 rebuilding and replay. D1 stack rows wall 21.054s; recursion neighbors 2.881s; added tests 60.855s. Release/sanitizer standalone C compile 0.025/0.029s. Corrected whole baseline 90.090s; initial split baseline 90.137s. Coverage Go compilation is outside reported binary elapsed. Final restored binary 0.963s. Warm setup skipped.
