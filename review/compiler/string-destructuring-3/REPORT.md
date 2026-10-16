Re-cut string object bindings for static length reads on current main, independently of records.
Commit: single branch HEAD supplied at handoff; base d79302e50c022c997c00f7c68ddb7475f9ee65fc; source 009e3cabfd12b9d48127b43fc70b0d0bb2092ebb.
Commands/results: focused lowering, source Node/native/JavaScript oracle, ASan/UBSan/leaks, flow, guards, counts regeneration, IR readers and vet pass; exact commands and times are in validation.json and the scripts.
Mutants: wrong member, omitted empty-pattern source evaluation, missing diagnostic fix and planted allocation count were all caught by their intended checks.
Not covered: full packages/gate, a new admission or TypeScript census, full scanner, benchmarks, WASI, optional estree/cycle-ledger inputs and the earlier over-budget typeaware archive-building probe.

The requested fetch, checkout from origin/main and squash merge succeeded without conflicts. Only counts.md auto-merged; no compiler conflict resolution was needed. Main's newer compiler, product and test infrastructure remains. The source's two lowering changes are retained: save the source with its actual representation, and give destructured string parameters a string representation. Only plain static length bindings are admitted; unsupported members retain a located NotYet and a fix. Renaming, repeated reads, empty patterns, UTF-16 semantics, effects and the parameter case are held by the fixture.

Counts assumption: current main has superseded the prompt's digest-column description. Commit 1f07ce2d moved measurement identity into products and removed the digest column. I took main's counts.md, retained that format and completed one regeneration with -update-counts on this Linux box using clang 20.1.8. It passed in 59.067s. Exactly one row was added, with no existing row changed or removed:

| Fixture | Allocations | Frees | Retains | Releases | Peak | Regions |
|---|---:|---:|---:|---:|---:|---:|
| internal/oracle/testdata/string_destructuring.a | 34 | 34 | 18 | 46 | 6 | 0 |

counts-audit.json contains the complete before/after audit. TestCountsAreRecordedTable passes after regeneration. No Mac table was used.

internal/native/runtime +/- lines are identical to the source's: both patches are empty, with no differences. runtime-equivalence.json records the source merge base and comparison. No backend or runtime implementation changed.

Guard results and test grain:

- TestEveryMapAProgramHoldsIsSorted passes. No IR map or field was added.
- TestRuntimeSetsAreEachUsed passes. No new feature set was needed.
- TestNativeAgreementRowsNameRegisteredPrograms passes. No native-agreement helper call was added.
- TestEveryTestIsParallelOrSaysWhy passes. The new diagnostic test already calls t.Parallel first; no baseline was changed.
- TestEveryDuplicatedProductDeclarationAgrees passes. No product declaration or recipe was added or changed; main's placeless products remain. The focused typeaware key checks also exercise absence of machine paths and the actual build compiler in product keys.
- No new compiler-consuming package was added, so cloud/fast-gate/compiler-dependencies.json needs no change. Main's registry remains intact.
- The new diagnostic test and fixture run under 60 seconds on the four-CPU cgroup. test-leaf-seconds.json records their exact times and the related checks. The fixture oracle completed in 14.939s, including runtime work. TestCallTargetReaders took 18.40s. No new or touched test needed splitting.
- TestCallTargetReaders ran with -timeout 90s. Vet passed on lower and oracle, the changed Go packages. Go formatting and the production-file whitespace check pass.

Every recorded mutant:

| Mutant | Observed catcher |
|---|---|
| charCodeAt(0) instead of length | Native and JavaScript stdout disagree with source Node, including 97 instead of 3; clean execution and sanitizers |
| Skip source declaration for an empty string pattern | Both backends print calls 2 instead of Node's 3 |
| Remove the unsupported-member fix | TestStringDestructuringMemberDiagnostic rejects the incomplete diagnostic |
| Recorded allocations planted as 35 | Counts family rejects recorded 35 against measured 34 |

All four mutant invocations exited 1 for those intended failures. The compiler overlays are regenerated from this tree, under .go.txt names, without modifying production files. mutants.json records the first three commands/timings; counts-mutant.log records the fourth, including both recorded/product and independent measured comparisons. No build failure, warning or sanitizer failure is counted as a semantic kill.

Consumer coverage is scoped, not exhaustive package testing. consumers.json inventories 46 current lower/oracle consumer packages, including the changed packages, with no missing package in the plan. consumer-selection.json and consumer-results.json record the exact selections and commands; packages without tests are compile-only. Additional IR and flow checks are in validation.json, and TestFixturesObjects in stage3.log passes (12.903s test time). All logs are retained in this directory.

The records attempt's TestInspectRequestRefusals exceeded the floor while building a cold checker C archive, and was not rerun. This branch uses TestTypeAwareKeysCarryNoMachinePath and TestTypeAwareKeysNameTheBuildsCompiler as its focused typeaware consumer checks. This is not a claim that the archive-building probe was repaired or validated. No other package's tests were changed to accommodate either branch. Optional library/cycle-ledger skips are listed in consumer-skips.json; no pass is claimed for unavailable inputs.

Toolchain/setup: Linux, Go 1.27.1, clang 20.1.8, Node 24.19.0, nproc=5 and cpu.max=400000 100000 (four CPUs). GOPROXY was https://proxy.golang.org|direct. Setup used the already established ADAMIC_GOCACHE_OFF=1 workaround and completed successfully: Node 0.018s, Go 0.028s, submodules 0.057s, markdown 0.058s, clang 0.146s, shared cache explicitly off 0.147s, go build 30.539s, build cache warm 30.697s and done 30.722s. setup.log has every timing line. Each build/test shell sourced the printed /workspace/adamic-tools/env.sh. Final tests used GOMAXPROCS=4 and -parallel 4 where applicable. Test output always went to files, and commands had hard limits.

The initial consumer-inventory assertion identified stage3/fixtures missing from the plan; I added and ran TestFixturesObjects successfully. No test was rerun after exceeding 60 seconds. Inherited source-branch evidence and its old census/admission reports remain under review/compiler/string-destructuring; only this directory describes the current re-cut.

The overall first-green-push target of 30 minutes was missed while records reached a floor violation and stopped. No blocked records state was pushed as green. This branch was then processed independently. No PR or Loom submission was made. No cohere code was copied.

Final consumer audit: all 46 selected package commands exited 0; the additional stage3 command also exited 0. Five optional-input tests skipped, listed in consumer-skips.json. No completed consumer leaf reached 60 seconds; the longest was TestCompileProfiles at 26.11s. consumer-timing-audit.json records the final audit.

Whitespace evidence: the all-file staged git diff --check reports space-before-tab and blank context lines inside the six literal mutant .patch files (three inherited and three regenerated). Those are valid unified-diff context, preserved exactly. The check excluding review/**/*.patch passes; production Go and fixture files are clean. The first commit command stopped at this check before creating a commit.
