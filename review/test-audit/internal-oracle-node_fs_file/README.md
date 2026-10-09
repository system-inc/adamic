u065: internal/oracle node filesystem and non-null audit

Starting origin/main: cb04a5ac599ede196c934d353e1aa9dec9748867. Branch test-audit/internal-oracle-node_fs_file. All 15 requested top-level tests were present at discovery; none moved or vanished. Regex callback and offset call the same librarySmallRuntimeMutant checker with different fixture/rule inputs, so they form TestNotYetLibraryRegex family (two members). Total 14 grouped rows. No subprocess helpers were found.

CODE UNDER TEST: Adamic non-null lowering/refusal and native filesystem operations. Go coverage inventory from the clean scoped run is reached-functions.txt (532 named lower/native functions); functions-coverage.txt retains all percentages. C code mutation is in flags, reached through write_file/write_buffer -> open_file -> flags -> bytes/invalid_value. Additional filesystem entries include close, mkdir/make_directory, date_new/date_time, write_data and descriptor/argument/path/error helpers. runtime-function-inventory.txt is a conservative source inventory rather than dynamic C coverage. ORACLE: source Node for filesystem bytes/effects and successful source behavior; handwritten panic/refusal expectations and count snapshots for intentional checked deviations. No external-authority value independently checked.

Results: {'sacred': 2, 'witness': 7, 'subsumed': 3, 'cannot-judge': 1, 'setup-check': 1}. Two sacred rows are bounded worthiness claims, not package/repository uniqueness. Four production mutants; four weakened-check changes; one setup construction break; two empty-entry probes. No production survivor. Witness and setup failures never appear in production kills. Mixed witness rows contain useful production assertions, but this brief requires production precondition failures to be excluded when judging their guarded check. results.json records these excluded failures explicitly.

TestCheckedNonNullCounts is cannot-judge. M01 makes Lower refuse its fixture before counting, which proves a precondition can fail but does not prove the count oracle catches incorrect allocation/reference counts. No frozen mutation changes the count producer. Its empty native.C probe fails at linking; this proves it cannot accept an empty backend answer, not that its snapshot is worthy.

Commands

Tools warm: source /workspace/adamic-tools/env.sh; nproc=5; Go 1.27.1 and Node 24.19.0. Setup skipped (0s). npm ci ran in stage3/api before discovery/baseline; npm duration was not separately instrumented.

Discovery: go test -list . ./internal/oracle/ > list.log 2>&1. Whole baseline: ADAMIC_BUILD_CACHE_DIR=/tmp/u065/cache/baseline timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run . > baseline.log 2>&1. It cooked at 90.473s on the binary line, with no assertion-failing test event before the timeout. The bounded clean run over all 15 names passed; coverage used -coverpkg=./internal/lower,./internal/native. Its command took 40.113s.

Each cost measurement: timeout 120 go test -count=1 -timeout 90s ./internal/oracle/ -run '^(ROW_MEMBERS)$' > time-ROW-N.log 2>&1, three runs. The regex family runs both members together. Binary ok-line medians range from 0.026s to 2.331s. These are ordinary gate-cache timings, not cold uncached execution benchmarks.

Every matrix command: ADAMIC_GATE_UNCACHED=1 ADAMIC_MUTANT=ID ADAMIC_BUILD_CACHE_DIR=/tmp/u065/cache/ID timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(ALL_15_DISCOVERED_NAMES)$' > ID.log 2>&1. Exact expanded regex is in run-status.json/results.json. The whole baseline exceeded budget, so no per-mutant whole-package attempt was made. The chosen names reach Lower/native.C and the mutated functions, established by clean coverage and source call sites. All outside-slice kills remain unknown. Every bounded matrix completed within the binary budget without a panic.

Oracle result caches must be bypassed: an environment-selected C runtime change leaves compiled source hashes constant between selectors. Per-mutant build caches alone do not distinguish cached oracle observations. ADAMIC_GATE_UNCACHED=1 was applied to production, witness, setup and probe runs.

Witness proofs

W01 makes disagreement return empty. It fails object-entry, both regexp members, string-bounds, parser, and the checked TypeScript no-check witness. W02 changes the filesystem witness's stdout equality to self-equality; TestNodeFSFileMutants fails. W03 changes the literal-initializer output comparison to self-equality; its built-in drop-check witness fails. W04 disables the checked TypeScript return-representation comparison; its narrowed subcase fails. The TypeScript row therefore has both its output and IR comparison components independently demonstrated. These are the explicit harness exceptions permitted for witness judgment; no Node source or expected output was changed.

S01 changes the fixture registration for non_null.ts from lowers=true to false. The migration metadata remains unchanged, and the setup row fails original runnable fixture lost its registration.

Probes

P01 returns nil,nil at Lower entry for the four ordinary refusal/assertion rows. P02 returns an empty string at native.C entry for filesystem agreement and counted native products. Every applicable row fails its own probe, so none is vacuous. Witness/setup rows have vacuous=null because no applicable production entry probe was judged. Each probe ran its row alone, preventing a nil-answer panic from hiding other rows.

Standalone validation

Every saved unified diff applies to the starting origin/main and passes its own tool check. Go edits pass go vet of lower, native, or oracle as appropriate. M04 compiles node_fs_file.c using native.Flags' warning, C11, sanitizer, contraction and frame flags; the exact clang command is in standalone-checks.json. All 11 checks passed, totaling 6.523s. All production and harness switches were restored; the restored clean slice passed in 10.780s command wall time. No other package's tests or PR was run/opened. Source switch copies and scripts are evidence only. Log gzip files are lossless.

Survivors

None among M01-M04. No equivalent-candidate claim. No witnessed behavior was classified as unguarded without a before/after output difference.

Brief ambiguities, limitations and costs

1. File locations are pinned to 8de93800f4 while the start rule requires current origin/main. Actual base is cb04a5ac599ede196c934d353e1aa9dec9748867; every source line uses that base.
2. Fifteen top-level names become 14 rows after the regex family rule. Other rows have different assertions; shared Lower or disagreement preparation alone was not used to merge them.
3. Three mutants per grouped row would exceed both the 20-change cap and the native rebuild advice of at most four. Four production changes were frozen across four functions, rather than designing one per test. This leaves native count-oracle quality uncovered and cannot-judge.
4. Checked TypeScript, literal initializers, and parser rows mix positive production checks with built-in witness checks. A single verdict cannot represent both roles. The brief's witness rule governed; raw production failures are retained but excluded from kills/uniqueness/subsumption for those rows.
5. Handwritten checked-panic outputs intentionally differ from source Node. Calling every Node-running row a pure external oracle would misstate its expected answer. Mixed oracle kinds and strength limits are recorded, including the literal-initializer row's Node check only forbidding exit 70.
6. NodeFSFileMutants silently continues past cases whose fixtures are in fsFileRefusedFixtures. Only six built-in cases ran, rather than its full 31-case source table; 25 were excluded before mutation. The agreement row runs five admitted fixtures. Refused read/open/stat/etc runtime behavior was not measured here.
7. The migration test checks construction, not compiler output. Unregistering a fixture is an allowed setup break; changing its pinned hashes/expected migration report would mutate its oracle and was avoided.
8. A first logging command used a relative evidence directory from stage3/api and failed before npm/tests ran. The named empty child directory was removed and the command rerun with an absolute path. Existing unrelated parent contents were left alone.
9. The Go coverage build took substantial time even with warm tools. It inventories Go compiler/emitter functions; C reachability is a source call-chain inventory, not instrumented runtime coverage.
10. Still skipped in the cooked whole baseline: TestWASIShardPlantedFixture, TestWASIAgreesWithNode, TestWASIOracleCatchesMutants, TestWASIRunnerCatchesMutants, TestWASIEmission, TestStage3FixtureHook, TestCheckedViewUntaggedArrayPending/good, TestEntriesAcceptance, TestCheckedViewUntaggedArrayPending/empty, TestCheckedViewUntaggedArrayPending/mixed, TestCheckedViewUntaggedArrayPending/wrong, TestCheckedViewUntaggedArrayPending/nested. None of the 15 scoped rows skipped. WASI tools were unavailable and outside this unit; no SDK/corpus installation or replay of those rows was attempted. Unfinished whole-package rows remain unknown.
11. Native product rebuild durations were not isolated from fixture execution. Per-selector binary/command times below include rebuild and run costs. The C runtime switch compiles once per relevant runtime configuration and selects at execution; separate compiler build caches and uncached oracle runs prevent stale observations.
12. Snapshot counts were not weakened as an oracle and no count-producer mutation was added after observing the frozen plan. The inability to judge that row is reported rather than turning its compile refusal into count evidence.

Timing and remaining scope

Setup 0s, npm duration unmeasured, whole baseline binary 90.473s, clean coverage slice command 40.113s, standalone checks 6.523s, restored slice command 10.780s. Elapsed through evidence assembly 23.5min. Initial timing values and every actual selector run are preserved. No package-wide uniqueness, repo-wide replay, independent external-authority check, native allocation/count producer mutation, per-product build timing, or refused filesystem fixture coverage was completed.

| ID | Binary seconds | Command wall seconds |
|---|---:|---:|
| M01 | 26.238 | 46.959 |
| M02 | 7.643 | 10.040 |
| M03 | 8.198 | 10.567 |
| M04 | 8.476 | 10.941 |
| W01 | 8.881 | 11.170 |
| W02 | 8.478 | 10.896 |
| W03 | 8.791 | 11.143 |
| W04 | 9.894 | 12.631 |
| S01 | 8.817 | 11.370 |

| ID | Origin file:line | Change | Raw failing rows |
|---|---|---|---|
| M01 | internal/lower/non_null.go:106 | change source-extension constant | TestCheckedNonNullTypeScript, TestCheckedNonNullAdamicRefusal, TestCheckedNonNullCounts, TestImpossibleNonNullFixturesAreRefused, TestPossibleNonNullAdamicAssertionsAreRefused, TestNonNullLiteralUnionInitializers, TestParserNonNullSourceExtension, TestNonNullAdamicRefusal |
| M02 | internal/lower/non_null.go:114 | change increment constant | TestCheckedNonNullTypeScript, TestImpossibleNonNullFixturesAreRefused, TestParserNonNullSourceExtension |
| M03 | internal/lower/non_null.go:96 | change diagnostic constant | TestCheckedNonNullTypeScript, TestImpossibleNonNullFixturesAreRefused, TestNonNullLiteralUnionInitializers |
| M04 | internal/native/runtime/node_fs_file.c:171 | change open option | TestNodeFSFileAgreesWithNode |
| W01 | internal/oracle/oracle_test.go:716 | witness: comparison always agrees | TestCheckedNonNullTypeScript, TestParserNonNullSourceExtension, TestNotYetLibraryObjectEntriesConstMutant, TestNotYetLibraryRegex family, TestNotYetLibraryStringBoundsMutants |
| W02 | internal/oracle/node_fs_file_test.go:223 | witness: stdout comparison always agrees | TestNodeFSFileMutants |
| W03 | internal/oracle/non_null_migration_test.go:139 | witness: stdout comparison always agrees | TestNonNullLiteralUnionInitializers |
| W04 | internal/oracle/non_null_checked_test.go:122 | witness: disable return representation comparison | TestCheckedNonNullTypeScript |
| S01 | internal/oracle/oracle_test.go:50 | setup: unregister runnable migrated fixture | TestMigratedNonNullFixturesRemainCovered |
