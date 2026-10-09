Built acceptance-rows unit 1: shared Node agreement helper and three migrations.
Commit: reported with the pushed SHA in the final response; only tests and review evidence changed.
Commands: focused helper and migration tests passed; package vet passed; lane output is in lane-checks.log.
Mutants: skip-stdout, accept-empty, and all five survivor semantics were caught by their named witnesses.
Not covered: remaining acceptance rows, whole-package tests, full gate, and integration's post-push gate.

The helper writes main.a, Loads and Lowers with fatal refusal diagnostics, emits generated.mjs, and runs each through node --disable-warning=ExperimentalWarning <repo>/oracle/node.mjs <file> with a separate 10-second context. It compares stdout bytes and exit codes; expected nonzero exits also compare stderr's first line. Silent source stdout is always refused with an instruction to print what the row computes. The recorder embeds the real test for temporary-directory cleanup and overrides Fatal/Fatalf with a private panic sentinel. It plants changed IR strings without subprocess test harnesses or global mutation seams.

Per-file migration:

- class_static_guard_test.go: moved source loading, lowering, JavaScript emission and source-vs-JavaScript execution/comparison into lowersAndAgreesWithNode. Kept the sanitized native build and execution leg.
- import_cycle_ready_test.go: moved module entry loading/lowering and the two Node executions into agreeEntry, using the expected exit variant. Kept native sanitizer and readiness-elision checks. To reconcile previously silent failure rows with the required empty-answer rule, added an imported probe.a printing the computed 1 + 2 before the cycle runs. Existing cycle outputs follow that one prelude line; panic messages and exit contracts remain checked.
- predicates_overload_test.go: moved Node execution and comparisons and JavaScript emission into the shared helpers. Kept .ts assertion loading for existing .ts rows, native sanitizers, predicate counts and check-elision assertions. Lying-predicate rows retain their explicit checked-failure contract, including complete stderr, because Node trusts those predicates; these are not ordinary acceptance rows. Their source stdout contract is also retained.

Every targeted structural check in the migrated code names what behavior cannot observe. No IR snapshots were added. New sources are inline .a probes, not new oracle corpus fixtures, so counts.md has no new fixture row to record and the counts updater was not run.

Commands and observations:

- export GOPROXY='https://proxy.golang.org|direct'; timeout 600 bash cloud/setup.sh, logged in setup.log. Succeeded. Ready timings: Node 0.025s, Go 0.030s, clang 0.218s, Markdown dependencies 0.914s, submodules 15.890s, go build 241.130s, build cache 241.228s, total 241.254s. nproc=5, cgroup CPU quota=4. Sourced /workspace/adamic-tools/env.sh. The initial formatting attempt before sourcing reported gofmt: command not found; sourcing resolved it.
- timeout 120 go test ./internal/lower -run '^TestAgreement' -count=1 -v -timeout 90s > review/compiler/lower-agree/helper.log 2>&1
- timeout 120 go test ./internal/lower -run '^(TestClassStaticSideEffectMatchesNode|TestUndecidedCycleReadsUseReadyChecks|TestPredicateOverloadRuntime|TestIndirectPredicateOverloadIsPending|TestClassStaticInitializerCallIsEmitted|TestFSOpenStringFlagsLower|TestFSRemoveDefaultRetryDelayLowers|TestFSExistsOperation|TestFSStatThrowsByDefault)$' -count=1 -v -timeout 90s > review/compiler/lower-agree/migrations.log 2>&1
- timeout 60 go vet ./internal/lower > review/compiler/lower-agree/vet.log 2>&1: exit 0, no findings.
- timeout 600 python3 review/compiler/lower-agree/run-mutants.py > review/compiler/lower-agree/mutants-run.log 2>&1: seven expected failing runs, each go test has -timeout 90s and an outer 120-second subprocess timeout; exact commands and wall timings are in mutants.json. All production inputs restored byte for byte by finally blocks.
- Integration lane command after committing: git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -, bounded externally and logged in lane-checks.log. This checkout fetches only main by default; explicit remote refspecs populated origin/devtools/fast-gate and origin/cloud/merge-tree first.
- origin/main advanced only with test262 test changes. Fast-forwarded this branch to c0a7667b before committing, then merged landed test-only main 09769cb5 and re-ran both focused selections and every mutant.

New leaf timings:

| Test | Seconds |
| --- | --- |
| TestAgreementRejectsWrongLoweredOutput | 0.27 |
| TestAgreementInheritedFieldWitness | 0.27 |
| TestAgreementAcceptsExpectedPanic | 0.28 |
| TestAgreementEnumConstantWitness | 0.28 |
| TestAgreementStaticInitializerWitness | 0.29 |
| TestAgreementRejectsEmptyAnswer | 0.14 |
| TestAgreementAcceptsComputedAnswer | 0.18 |
| TestAgreementParseIntRadixWitness | 0.18 |
| TestAgreementParameterPropertyWitness | 0.19 |

Helper file result: ok  	github.com/system-inc/adamic/internal/lower	0.484s

Migrated and collision-renamed leaves:

- TestUndecidedCycleReadsUseReadyChecks: 2.30s
- TestClassStaticInitializerCallIsEmitted: 0.08s
- TestFSOpenStringFlagsLower: 0.83s
- TestFSRemoveDefaultRetryDelayLowers: 0.84s
- TestFSStatThrowsByDefault: 0.77s
- TestFSExistsOperation: 0.86s
- TestIndirectPredicateOverloadIsPending: 0.03s
- TestClassStaticSideEffectMatchesNode: 0.32s
- TestPredicateOverloadRuntime: 6.53s

Focused migration selection result: ok  	github.com/system-inc/adamic/internal/lower	8.843s

Mutant evidence:

| Mutant | Catching test | Observation |
| --- | --- | --- |
| skip-stdout | TestAgreementRejectsWrongLoweredOutput | agree_test.go:168: planted wrong output was not caught by stdout comparison: "" |
| accept-empty | TestAgreementRejectsEmptyAnswer | agree_test.go:176: silent row was not refused by empty-answer probe: "" |
| enum-plus-one | TestAgreementEnumConstantWitness | agree_test.go:192: JavaScript backend stdout = "8\n", source Node = "7\n" |
| drop-base-fields | TestAgreementInheritedFieldWitness | agree_test.go:197: JavaScript backend stdout = "7:11:own,#base@1\n", source Node = "7:11:own\n" |
| drop-parameter-store | TestAgreementParameterPropertyWitness | agree_test.go:202: JavaScript backend stdout = "0\n", source Node = "7\n" |
| drop-static-initializer | TestAgreementStaticInitializerWitness | agree_test.go:207: JavaScript backend stdout = "done\n", source Node = "static side effect\ndone\n" |
| swap-parseInt-radix | TestAgreementParseIntRadixWitness | agree_test.go:212: JavaScript backend stdout = "10,10,10\n", source Node = "10,NaN,2\n" |

Evidence provenance: enum_flags M01 changes enum constants by +1; class_inheritance M02 drops inherited fields; parameter_properties M03 returns before stores; library_language M20 passes the input string as parseInt's radix. These diffs were read from origin/test-audit branches without merging their code. Static initializer omission follows review/compiler/class-static-guard/M15.diff, adapted to current source. The two helper diffs are generated against agree_test.go. All are non-compilable .diff evidence.

Observation versus inference: an initial public-field inheritance row survived the exact base-fields mutant; initial-public-base-mutant.log preserves that result. JavaScript uses property names for public reads, so the final witness observes inherited #private field enumeration instead. The final mutant printed 7:11:own,#base@1 where Node printed 7:11:own. All five final survivor witnesses fail stdout comparison rather than clang or a structural snapshot. No claim is made about the remaining acceptance rows until later units migrate them.

Integration adjustment: landed main introduced fs_method_guards_test.go with another lowersAndAgreesWithNode taking four arguments. Merging it produced a redeclaration and argument-count build failure. The requested two-argument API cannot coexist with that declaration. The minimal conservative resolution renames only that existing filesystem helper and its four calls to fsLoweringAgreesWithNode; its implementation and behavior stay intact. This extra test-only file edit resolves a real integration collision rather than expanding the acceptance-row migration. All four renamed leaves were run. Their first run lacked stage3/api/node_modules/@types/node 25.3.3; timeout 120 npm ci in stage3/api installed the pinned declarations (three packages in 989ms), after which the focused selection passed. initial-missing-node-types.log and initial-lane-collision.log retain the failures. No production source changed in this merge or resolution.
