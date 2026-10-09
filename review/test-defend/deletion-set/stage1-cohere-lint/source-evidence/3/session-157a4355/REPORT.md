TestProfileCompilationBuildLower was not uniquely defended in the bounded 16-function matrix.
D1 proves it can fail on a real production error; the identical producer and 12 consumers also fail.
All production sources restored; standalone diffs, complete logs and coverage retained here.

Starting origin/main: 157a43552015f41a79331949c2e82b6f8c7caaab.

CODE UNDER TEST: Adamic Lower for the unchanged lint/main.ts checked graph. Mutations are in internal/lower, not Go cohere, port oracle adapters, tests or construction helpers. ORACLE: target checks successful load/lowering/file creation/gob serialization through buildcache.Product. It has no external semantic expected answer. The runtime profile case executes unchanged Go cohere and compares Node, emitted JavaScript, native scanner, profiled and counted answers.

Coverage and difference

Fresh per-test cache roots force lowering to execute. Commands are in coverage-command.txt; target.cover and product.cover use -coverpkg=./internal/load,./internal/lower,./internal/buildcache. Each covers 3447 blocks, neither has exclusive blocks. TestProduct_ProfileCompilationLowered calls exactly compilationLowered(t, compilationInputs(t)), like the target. Only deferred versus immediate wall logging differs. Both log, neither asserts a build-time ceiling. No distinct input history, executor or semantic contract was found.

Baseline and scope

The audit had already observed the whole package exceeding 90 seconds. The current package lists 6760 functions; the audit list had 6739. scope-changes.json records 22 added and one removed, none in this selected profile-compilation graph. No whole-package replay was attempted. A six-function producer/emission clean baseline passed in 62.980 binary seconds. The expanded current 16-function graph passed clean in 24.815 seconds, without skips. Each mutant uses that same expanded graph. Other compiler callers, opt-in ProfileArtifacts and ProfileSnapshotsAgree, and the wider package are outside this bounded uniqueness claim. The new compiler-corpus rows were enumerated but not run; no unique defense is asserted anyway.

Attempts

D1 flips Lower's single-entry validation. It rejects the existing one-entry lint driver. Fourteen functions fail, including target and identical producer; Union and Go-oracle product pass. Binary time 5.121 seconds. This proves the audit's untrue hint was too broad if read as inability to fail. Production errors are caught. Current bounded finding is subsumption by TestProduct_ProfileCompilationLowered.

D2 changes maximumGenericDepth from 32 to zero. All sixteen pass, binary time 80.718 seconds. This was an unsuccessful exploration: cold coverage confirms instantiateFunction and the generic-class guard do not execute on this driver. It is an equivalent candidate for this selected input, not unguarded behavior and not strong evidence about the target. The initial planned generic-cache D3 candidate was likewise stopped when this coverage was noticed, and was replaced before the final D3 attempt. initial-plan.json and driver.log retain the process deviation. Its orphaned command was stopped; it contributes no result or verdict.

D3 disables the reached constant-string interning fast path. The independent direct-Lower probe reports 1779 string constants before, 5490 after, with 1779 distinct values and 758 functions in both. All sixteen functions pass, including runtime agreement of every executor against Go cohere, binary time 81.451 seconds. Thus answers on the profile witness stay right while the IR constant table and emission work grow. Neither target nor producer has a cost assertion. The binary wall remained below 90 seconds; no final matrix cooked. Native rebuild plus compiler work is included in the matrix command times, not separately isolated. The target's late 0.29-second pass in D3 is a shared product cache hit, not a cold performance comparison.

No single outcome here recommends deletion. The row does warm/build a named product and catches build failures, but does not validate the returned product path, its filename or semantic contents. The audit's S03 missing-program.gob finding is prior evidence only; this session did not repeat a prohibited harness mutation. The name BuildLower promises a build stage, which its success/error checks exercise, but an owner should not read that as a correctness or build-budget guarantee.

Validation and replay

D1.diff, D2.diff and D3.diff each apply to the starting origin/main with git apply --check and independently passed go vet ./internal/lower/. The matrices also compiled with actual Go and downstream emit/native tools. No production change is committed. Use each diff separately and the command/cache in its corresponding JSON. All test output is retained in compressed .log.gz files. constant_probe.go is an independent observation program, not a test edit.

Friction and limits

- Full audit REPORT.md contains rows plus construction/witness faults. S03 is a test-harness filename edit, forbidden for this defender, so it was read but not replayed.
- The whole-package requirement conflicts with practical package size and the prior audit's timeout. This report uses all sixteen current profile-compilation functions and labels wider catches unknown.
- The audit's assigned slice omits the identical product wrapper even though it existed at the audit commit. It is now included; it is not a newly added row.
- Initial generic attempts assumed this large driver would use generics. Coverage disproved that assumption. D2 is explicitly weak/inapplicable evidence; the cache candidate was stopped and replaced with a reached fast path. This cost a full 80-second binary run and part of another compile, without adding a defense.
- Deferred versus immediate timing logs can differ because rows share build products and run in parallel. Logs do not establish a threshold or isolated speed measurement. D3 uses constant-table growth, not a noisy wall-time comparison, as its work witness.
- A direct build wrapper is not a Node/native twin, and the audit's untrue label must distinguish sampled output faults from the ability to reject a compiler error. D1 supplies the latter proof.
- The full test list accidentally exceeded tool output limits; complete list.log is preserved, and scope was subsequently processed without dumping it.
- Warm tools worked, so setup was skipped; nproc=5. npm ci in stage3/api completed before both clean baselines. No other Node dependency directory is used by this graph. Fetching main took about one minute. Exact mutant command walls and binary times are in matrix.json. Baselines plus final matrices sum to 255.085 binary seconds, excluding coverage, probes, compilation overhead and restored check.

Restored final matrix: PASS, 5.702 binary seconds. All sixteen rows completed.
