No unique defense was found in the bounded three-attempt matrix.
The current family catches real production breaks, so the audit's untrue finding is obsolete.
All production sources were restored; no test was edited, deleted or weakened.

Starting origin/main: 92d196011e78208b45b3768dcf2c8c88e7ec132d. The fetched audit is preserved in audit-README.md, audit-rows.json, audit-report-data.json, audit-scope.json, audit-production-plan.json and audit-matrix.json. Earlier defense evidence at 59f8da69 is preserved under prior-59f8da69/. Its history is retained without force-pushing.

CODE UNDER TEST: the TypeScript lint port, including Linter.run, ancestry, walk, iterative fixed(), and FileCommentCache.get(), compiled through Adamic. ORACLE: current Go cohere byte output for every selected corpus case, followed by equality between plain and junk-row native output. Owned witnesses independently compare Node, emitted JavaScript and native output with Go cohere. These are different input families, not executor twins over one identical case set.

All eight requested members exist: TestNodeTableIsLinkOnly_000 through _007. They moved from node_table_split_test.go to node_table_oracle_floor_test.go. Commit ca87bd7a added a Go comparison and nonempty-oracle assertion after the audit. Current go test -list records 6760 top-level Test names in list.log. The family label itself is not a top-level function, so its members were selected by regex.

The whole clean package timed out after 90.299 binary seconds, with no individual failed-test event before the timeout. The eight-member clean baseline passed in 3.481 seconds. Final restored-source control passed all eight in 3.491 seconds. Warm tools worked and setup was skipped. nproc=5. npm ci in stage3/api succeeded before baseline; its separate duration was not recorded.

Coverage: family.cover comes from the eight members with -coverpkg=./internal/lower,./internal/native. other.cover comes from TestNestedOutsideModuleCopy with the same instrumentation. There were 16 covered family blocks and 2574 comparison blocks, with just one family-exclusive block: internal/native/native.go:106, the release -O2 flag return. This is a build-mode difference, not an exclusive lint behavior. Cached products hid most compiler work. Go coverage does not measure executed TypeScript port lines. The attempted ProfileSnapshotsAgree comparison skipped because ADAMIC_LINT_PROFILE_SNAPSHOTS was unset. These profiles are not coverage of the whole rest of the package. The semantic lead is that only the target family passes --junk-rows; perturbation construction was left unchanged because it is a test-only hook.

Three honest production attempts, each with an independent ADAMIC_BUILD_CACHE_DIR:

- D1, lint.ts:83: drop parent-table initialization. Aim: the distinction between table setup and root-linked ancestry with detached rows. All eight target members failed, as did all sixteen owned-witness members. Native port exit status 70 proved a real break, but not uniqueness. D1-final.log is the completed matrix.
- D2, lint.ts:202: drop next.run() in the fix pass. Aim: captured multi-pass fix histories versus minimal owned witnesses. All eight target members failed with byte differences, including nested conditional simplification and BOM removal; TestOwnedWitnesses_004 and _011 also failed. D2-final.log is the completed matrix, including newly added product tests.
- D3, helpers/comments/for_file.ts:18: drop this.ready = true. Aim: the family's 60-second shard work guard. Each cache request now recomputes comments. The source probe changed from counts=2,2 same=true ready=true to counts=2,2 same=false ready=false. Every target member passed, with logged work below 60 seconds; completed narrowed binary time was 3.733 seconds. This attempt does not defend the family. Owned-witness outcomes for D3 remain unknown because their product preparation repeatedly exceeded 90 seconds. D3-family.log is the completed target matrix. Do not call D3 a whole-package survivor.

The initial third candidate, an index-zero boundary change, was never planted or run. It is retained as unexecuted-boundary.diff and is not an attempt, mutant result or compile claim. It was replaced with the cost-preserving D3 because the current shards assert a work budget.

All D1-D3 standalone diffs independently passed git apply --check --cached against the clean starting source. Each was actually compiled by the port's lowering, native.C and native.Build path, as shown by the native product logs and completed family runs. No Go vet is presented as TypeScript compilation proof. No oracle, harness or test source was mutated.

Current finding: not defended by these three attempts, with production kills shared by TestOwnedWitnesses family. The old untrue verdict is refuted. This is a small bounded subsumption hint, not a deletion recommendation or proof that no unique break exists. Raw member pass/fail/skip results, bounded matrix rows and unknown comparisons are in observed-rows.json and matrix.json. Other package rows, including large compiler-corpus and RulesAgree families, were not replayed per mutant. Package-wide and repository-wide uniqueness are unknown.

Brief costs and ambiguities:

1. The 15 GB free requirement cannot be met on an 8.8 GB /tmp filesystem. Initial /tmp was full; deleting explicitly named prior-unit /tmp/u003, /tmp/u048 and /tmp/u048-profile-bin recovered 3.0 GB. Shared temporary directories were preserved because other processes were active. /workspace initially had 8.8 GB free. No workspace or tool content was deleted.
2. The audit's file and oracle description refer to older test bodies. Reading current tests was essential: ca87bd7a changed what can fail.
3. A literal family-name -run pattern selects no tests. Family members must be enumerated or selected by regex.
4. Requested Go coverage cannot instrument the actual TypeScript code under test. Warm product hits can make two strong runtime tests appear to cover no compiler lines. The small comparison profile does not replace a whole-package coverage map.
5. The 6760-test package exceeded the 90-second cap. Cold C emission and separate owned-witness products repeatedly consumed nearly the entire cap. Cooked runs are unknown, never kills. Subsequent runs narrowed to exposed build phases and the requested family. One D2 replay repeated the mixed bounded selection before its remaining product was separated; that cost an additional cooked run.
6. The special instruction says a formerly untrue row caught by another row becomes subsumed, while the final defense enum omits subsumed. rows.json uses not defended, names the catcher and this report states the bounded subsumption finding.
7. A skipped profile row provided no comparison evidence. Its log and profile are retained to make this limitation explicit.
8. The requested push branch already existed. Prior evidence and commit history were preserved, rather than overwritten or force-pushed.

Name/assertions finding: the current family checks its promised observable link-only invariance and compares with Go. It does not assert that the junk-row option actually appended the intended detached rows. Disabling perturbation might therefore pass both comparisons. This is an inspection finding, not a new demonstrated survivor. The nonempty Go-output check is not itself a diagnostic-count assertion; the subsequent byte comparison is the substantive oracle.

Elapsed session was about 30 minutes. Summed recorded binary durations are 865.207 seconds, including cooked runs and overlapping controls; this is not elapsed wall time. Successful cold build timings, including C emission near 78-86 seconds and the separate owned lowering/emission near 80-84 seconds, are in timings.json and overlap their enclosing runs. No whole-repository test run, exhaustive per-row coverage, profile snapshot setup, or completed D3 owned-witness matrix was performed.
