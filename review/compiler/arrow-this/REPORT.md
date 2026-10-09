Built: isolated class constructor capture discovery from unrelated enclosing arrows for task #rh16bxg, after the lowering chain.
Commits: baseline a7edbde53b4ae2dfe93e8e1c721946d89d373155; current-main merge 3a3cea182a656f3acc9ed61a23e96188a81dfad3; implementation 9764a2de83ebfcfb5fbe6529db609c29d1c81118.
Checks: three Node/native/JavaScript oracle leaves pass; required reader guard, lowering, counts and lane results recorded below.
Mutants: constructor-context removal and dropped capture both fail on clang undeclared identifiers; patches and logs are preserved here.
Not covered: full repository gate, all metamorphic variants, generic representation findings, or the unrelated existing empty-object and never-array refusals.

The original report is on origin/devtools/metamorphic at cmd/adamic-metamorphic/REPORT.md. That branch was read, never merged. All five B-family reduced witnesses are preserved as reduced-1.a through reduced-5.a. On after-chain, reduced-1 (the 19-line deep superclass method) and reduced-3 (subclass holder) build. Reduced-2 stops in lowering at its empty-object field, and reduced-4 stops at its never-element array. Reduced-5 still reaches clang with an undeclared this_cell in MetamorphicBlock2.run. Its fixed build exits 0.

The minimal superclass method and direct field initializer also pass on after-chain. The final field fixture adds the essential failing condition: the compiler discovers class construction while lowering an unrelated enclosing arrow. The preserved baseline command-line compiler rejects that final fixture with undeclared adamic_local_6_this_cell, as baseline-field-warm.log shows. The original fixtures class_oct6_deep, class_oct6_release, class_oct6_subclass_holder and borrow_element_super_move pass their focused oracle runs.

Observed cause: inheritanceConstructor runs field initialization with the discoverer's l.closures stack still present. touch then puts the new constructor receiver into the unrelated arrow's environment. That receiver has no declaration in the unrelated arrow's parent. The fix saves, clears and restores the closure stack at constructor lowering entry, as ordinary method body lowering already does. Arrows created inside the constructor still capture its receiver normally. Assumption: a class initializer has its own receiver frame independently of the expression that first instantiates the class. No new refusal is needed.

Three separately parallel top-level tests register .a fixtures with the standard counts harness and reuse callTargetFixtureAgrees. This checks explicit expected source-Node output, backend JavaScript, uncached release native, uncached ASan/UBSan native and leaks. Runtime labels are constructed with join rather than immortal literal labels. Final leaf times: TestArrowThisSuper 2.81s, TestArrowThisField 3.20s, TestArrowThisStored 2.82s. The stored control captures both super and this and invokes the saved arrow later.

Mutants run sequentially with restoration in finally blocks:

- constructor-context removes the three closure-stack isolation statements. TestArrowThisField fails with clang's undeclared adamic_local_6_this_cell.
- dropped-capture removes the environment append in touch while preserving captured-local marking. TestArrowThisSuper fails with clang's undeclared adamic_local_5_this_cell.

Both go test commands exit 1; the runner exits 0 only after checking the intended undeclared-identifier diagnostic. These are hard C errors, not warning-only kills. run-mutants.py, both .patch files and both logs preserve the evidence. No mutated source remains.

Commands (from repository root, each sourced /workspace/adamic-tools/env.sh; all test output redirected to files):

```sh
export GOPROXY='https://proxy.golang.org|direct'
timeout 600 bash cloud/setup.sh
npm ci --prefix stage3/api
go test ./internal/oracle -run '^TestArrowThis' -v -count=1 -timeout 90s
go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/(class_oct6_deep|class_oct6_release|class_oct6_subclass_holder|borrow_element_super_move)' -v -count=1 -timeout 90s
python3 review/compiler/arrow-this/run-mutants.py
go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s
go test ./internal/lower -count=1 -timeout 150s
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 5m -args -update-counts
go vet ./internal/lower ./internal/oracle
git diff --check
```

Setup reports Go ready 0.091s, Node ready 0.120s, clang ready 0.698s, markdown dependencies ready 1.620s, submodules ready 19.824s, Go build ready 474.543s, build cache warm 474.689s, done 474.727s. nproc is 5; cpu.max is 400000 100000 (4 CPUs). The printed environment file is /workspace/adamic-tools/env.sh. Initial cold reproduction/oracle commands reached their outer limits and are superseded by the warm runs. The first lowering run failed for missing pinned @types/node 25.3.3; npm ci installed it and the suite was rerun. The first full counts refresh exceeded its 90-second aggregate limit and was rerun with a five-minute hard test limit. Logs preserve these failures rather than claiming them green; the full counts timeout stack is compressed in counts-timeout.log.gz.

Final results and delivery are appended below.

Final restored results: oracle 3.227s; internal/lower 148.833s; TestCallTargetReaders 6.751s; go vet exit 0 with no diagnostics; git diff --check exit 0. The required internal/lower suite was the explicit package-check exception; no other whole package or full gate was run.

The mandated command after committing was:

```sh
git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -
```

It exited 0 and printed: lane checks 9.2 s: gofmt and tools on 273 Go files, t.Parallel on 22 test packages; vet 22 packages. The dependency branch carries substantial prior lowering-chain changes, so the checker sees more files than this unit's six-line fix and new test file. The checkout's main-only remote fetch configuration required explicit tracking-ref fetches for the dependency and lane-check branches before the prescribed command could read them.

The full TestCountsAreRecorded update passed in 210.896s. The entire counts diff is three added rows, with no existing row changed. In allocations/frees/retains/releases/peak/regions order: super 5/5/8/8/5/0; field 7/7/9/11/6/0; stored 7/7/9/11/7/0.

Delivery branch is compiler/arrow-this, based on the explicitly requested origin/compiler/after-chain dependency and merged with current origin/main ce1c5a2fd91e40b05160e587ea5d88ed5bdfedb2. This unit advances #rh16bxg by eliminating the remaining reproduced receiver-capture leak and guarding both immediate shapes and the stored control against Node. The brief supplied no numbered roadmap step beyond the task identifier. No PR was opened and no main or area branch was pushed.

After committing the counts and report, lane checks passed again: lane checks 3.9 s: gofmt and tools on 273 Go files, t.Parallel on 22 test packages; vet 22 packages.
