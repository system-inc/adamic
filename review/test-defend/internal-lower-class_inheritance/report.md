All three assigned rows have a unique production-mutant catch in the current whole-package matrix.
Base: 7b9d4272c28f59530ab13daa5c49067e47933b06; 276 top-level tests, including 37 added since the audit.
No test was changed; all production mutations were restored.

Code under test and oracle

Adamic lowering is the code under test: lowering.instantiate and its recursion refusal; lowering.prepareSuper / constructorSuperFlow.reads; lowering.nominalAncestor and nested nominal argument proof. Lower is the entry, with loading and checker preparation preceding it. All three row oracles are self: handwritten refusal class and diagnostic substrings, not live external agreement. D1 defends repair advice, not the precise limit or resource cost. D2 and D3 defend rejected programs that otherwise lower successfully.

Coverage and semantic differences

The six clean per-test profiles use -coverpkg=./internal/lower and the exact individual test name, with -count=1 and -timeout 90s. Raw profiles and command logs are retained; exclusive-coverage.json lists positive blocks absent from each prior subsumer's positive blocks. These are relative exclusives, not claims of package-exclusive coverage.

Growing classes: 402 exclusive blocks versus HasClassIdentity, including class.go:119-121, the actual generic-class recursion guard. D1 changes its repair constant at base line 120. The identity test compares live native output with Node and never exercises this refusal. The growing test does not pin the depth value, an explicit limit or cost threshold; this defense rests only on the diagnostic contract.

Conditional this: 428 exclusive blocks versus RefusesThisBeforeSuperReturns, including class_super.go:157. The semantic difference is its third program's default constructor parameter closure using this, before the constructor body calls super. The subsumer checks a body closure before super. D2 drops the whole parameter loop at base line 157; the first two conditional programs still refuse, while the third is wrongly accepted.

Generic views: 75 exclusive blocks versus GenericNominalConstraints, including class_inheritance.go:592-599. Both nominally distinct classes have structurally identical methods. The assigned row checks Box<Other> versus Box<Root>; the subsumer checks Other directly against a T extends Root constraint. D3 drops the whole nested argument comparison loop at base line 596. The Box view is wrongly accepted; direct constraints still refuse.

Matrix and validation

Each standalone diff was applied separately and passed go vet ./internal/lower/. Each matrix command is recorded in results.json and used its own ADAMIC_BUILD_CACHE_DIR=/tmp/defend-inherit/cache/Dn, timeout 120, -json -count=1 -timeout 90s ./internal/lower/ -run . . Each completed, without panic or timeout. The matrix covers all current tests rather than only the six coverage rows. Each mutant failed only its assigned row, with 273 other top-level rows passing. Full passed lists and failing lines are in results.json. Standalone diffs have no selector switch. All three apply to the recorded origin/main base and compiled under vet and go test.

Skipped coverage limits: TestOriginalCycleLedger and TestOptionalWideningCensus require externally selected projects and output destinations. One TestMixedUnionContractGraph subcase explicitly awaits compiler support. No uniqueness claim is made against skipped behavior. No repository-wide uniqueness claim is made.

Timing and friction

Warm toolchain setup skipped; nproc 5. npm ci was rerun in stage3/api before the clean baseline. Clean whole-package test binary: 30.535 seconds. Mutant wall times, including go test compilation: D1 38.567 seconds, D2 37.613 seconds, D3 37.248 seconds. Vet timings are not separately measured. Coverage binaries all finished below one second. No run exceeded the budget.

The workspace environment needed to become available before the disk command could run. df reported /tmp as an 8.8 GB filesystem with 8.6 GB free, so the requested 15 GB free there is impossible even when empty. Only the prior /tmp/defend-nodefs scratch/cache directory was removed; /workspace/adamic and tools were untouched. /workspace reports 15 GB free. Neither baseline nor mutation runs failed on disk capacity.

The audit was based on an older revision: 37 tests have been added, none removed, and some inheritance acceptance assertions now compare native output with Node. scope-changes.json lists them. The generic-class row's name promises growing-class refusal and it checks refusal plus repair wording, but not a particular recursion threshold. Its unique diagnostic catch should not be read as a timing or memory proof. The other two names match the refusal behavior demonstrated here. The shared advice about executor twins does not apply to these three refusal rows.
