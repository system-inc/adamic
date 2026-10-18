Added a-check refusal headers to every Adamic program introduced by the closure-target cycle change, preserving all body bytes.
Commits: delivery SHA in final response; one commit on dbdf977f5b8b0979d414ebe6f20e7aadd312e309, containing bc53e520360047509abc8ec6de963ac785cda8ef plus headers and the shifted map expectation.
Commands/results: real pinned Gate.aCheck accepts 3/3 files; 44 focused/reader/consumer test events pass; changed-package vet passes; counts are byte-identical to bc53e520.
Mutants: removed-header and wrong-reason each fail the unchanged pinned predicate; all six implementation mutants also fail their intended assertions, then restore sources.
Not covered: full-tree a-check, full package suites, full repository gate, WASI or renewed admission census. No ahra invocation, PR or Loom submission.

Inventory is exactly git diff --name-only dbdf977f5b bc53e520 -- '*.a', without excluding review or any other directory. There are three files, all listed below; no review .a file is added or modified by this change. Counts by status: checked 0, NotYet 0, refused 3, type error 0. The scan uses the branch's own adamic c through the exact unchanged Gate.aCheck method pinned by stage3/a-check-headers/README.md, fbac28c62493f27a788edc02a18bc8edb68de5da:cloud/fast-gate/run.py. The pinned implementation is saved verbatim, with its SHA256 in predicate-provenance.json. Its real go build compiles this branch's cmd/adamic; each real file is compiled in place before and after header insertion.

| File | Status | Actual stable refusal reason |
|---|---|---|
| internal/oracle/testdata/cycle_closure_targets_refused/direct.a | refused | 'memo', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free |
| internal/oracle/testdata/cycle_closure_targets_refused/joined.a | refused | 'memo', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free |
| internal/oracle/testdata/cycle_closure_targets_refused/map.a | refused | Map<number, () => number>, a map whose values can reach back to a map like it: a cycle reference counting can't free |

The three added first lines are:

internal/oracle/testdata/cycle_closure_targets_refused/direct.a

```text
// a-check: refused 'memo', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free
```

internal/oracle/testdata/cycle_closure_targets_refused/joined.a

```text
// a-check: refused 'memo', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free
```

internal/oracle/testdata/cycle_closure_targets_refused/map.a

```text
// a-check: refused Map<number, () => number>, a map whose values can reach back to a map like it: a cycle reference counting can't free
```

Conservative assumption: use the compiler's stable diagnostic reason clause as the header expectation, excluding source paths, source positions and suggested fixes. These are actual diagnostic substrings, and the unchanged predicate verifies each against the real compiler stderr. The full compiler diagnostics remain in before.json and after.json. No compiler, runtime or refusal-rule implementation differs from bc53e520.

Body preservation is checked both by direct byte equality against git show bc53e520:<path> and by SHA256 before/after excluding exactly the newly inserted first line. headers-added.json records every header, byte count and body digest. There is no newline normalization, reformatting or body edit.

Position-sensitive expectation: internal/oracle/cycle_closure_targets_test.go's map write location changes from <fixture>:14:1 to <fixture>:15:1. The compiler now reports the map declaration at 13:7 and the write at 15:1. Direct/joined diagnostic declaration positions also move down one line, but their tests assert the position-independent What/Fix and need no edit. Searching the fixture paths and directory in *_test.go and testdata found only this active helper/expectation. Historical admission-delta.json captures retain their actual pre-header observations and positions; they are not executable expectations and are not rewritten. The new before/after captures distinguish both versions.

The real pre-header scan fails for all three missing headers (before.json). The real post-header scan accepts all three refused programs (after.json). As the README prescribes, both header-predicate mutants operate on a real file with the captured real post-header diagnostics: removing direct.a's first line fails with expected checked, and replacing it with a deliberately incorrect refusal reason fails the substring requirement. Header bodies are restored in finally blocks. The unchanged predicate accepts the restored files (final.json). These prove the predicate, not a replacement Python imitation. The real compiler scan is separate from diagnostic replay, exactly as in the README.

No source .a files were added for proof. The mutation manifests are JSON, and implementation mutations are .patch evidence, never compilable Go. All six previous implementation mutants were rerun with the new headers and map expectation:

| Mutant | Catcher |
|---|---|
| omit-boundary-inputs | TestClosureTargetsKnownCallerDoesNotHideUnknownCaller |
| unknown-is-empty | TestClosureTargetsBoundOnlyProvenValues |
| omit-captures | TestClosureCycleDirect |
| drop-return-target | TestClosureTargetsFollowParametersReturnsAndMapValues |
| omit-map-captures | TestClosureCycleMap |
| type-matched-captures | TestNamespaceAmbientCwdInitialization and TestClosureCycleAmbientCwd |

Every mutation fails by its intended assertion, without a build-error or timeout kill; all production sources are restored afterward. The map refusal's strict expected reason is retained, with only the necessary line increment.

Counts are unchanged relative to bc53e520, whose generated counts changes are retained in the combined commit. The headered programs are refused and have no counts rows. TestCountsAreRecordedTable passes, and unchanged-compiler-and-counts.json records the byte-identical counts table and cycle implementation hashes. No counts regeneration or hand merge was performed in this pass. No runtime C lines or runtime feature sets changed, and the exact runtime witness guard passes. No new compiler-consuming package was introduced. No cohere code was copied.

Reruns: all three path-reading refusal tests compare Node's stored output and the exact refusal; the broader original focused witnesses and guards pass too. The reader guard, unchanged map guard, parallel-test guard, exact runtime-set witnesses and native-agreement registry guard pass. Focused native/fresh/flow/JavaScript consumers and the flow fixture-coverage guard pass; these nested refused fixtures remain outside flow's top-level glob. All 44 test events pass, maximum 5.65 seconds, on GOMAXPROCS=4 and -parallel=4. Full per-test seconds are in test-seconds.json. No new test functions or tests above budget were introduced.

| Test log | Pass events | Maximum seconds |
|---|---:|---:|
| focused.log | 36 | 5.65 |
| readers.log | 1 | 3.21 |
| consumers.log | 7 | 0.09 |

Header-sensitive test seconds:

| Test | Seconds |
|---|---:|
| TestClosureCycleJoined | 0.23 |
| TestClosureCycleMap | 0.28 |
| TestClosureCycleDirect | 0.19 |

Toolchain setup passes. nproc=5, cgroup cpu.max=400000 100000 (four CPUs), Go 1.27.1, Node 24.19.0, house clang 20.1.8. The environment initially took several minutes to become available. Main advanced to 99c369e0 at fetch; the branch deliberately uses the specifically requested dbdf977f5b base. The prior commit applied cleanly without conflicts. Setup timing lines:

```text
setup: submodules ready (0.087s)
setup: go ready (0.090s)
setup: node ready (0.094s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.021s
setup: markdown dependencies ready (0.254s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.498s)
setup: shared cache ready (1.029s)
setup: go build ready (15.814s)
setup: test binaries deferred (use --warm-tests) (15.959s)
setup: build cache warm (15.961s)
setup: build-flags commit=dbdf977f5b8b0979d414ebe6f20e7aadd312e309 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=0.06 0.02 0.00 1/134 707 load-after=1.05 0.23 0.07 1/139 1347
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (15.988s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.sNsolz
```

All test stdout/stderr was redirected directly to logs. Each long command is bounded. Commands (from repository root):

```sh
timeout 120 git fetch origin
git checkout -b compiler/loom-r-cycle-targets-ac dbdf977f5b
timeout 120 git cherry-pick -n bc53e520360047509abc8ec6de963ac785cda8ef
export GOPROXY='https://proxy.golang.org|direct'
timeout 600 bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
export GOMAXPROCS=4
nproc
timeout 120 git fetch origin fbac28c62493f27a788edc02a18bc8edb68de5da
git show fbac28c62493f27a788edc02a18bc8edb68de5da:cloud/fast-gate/run.py > review/compiler/loom-r-cycle-targets-ac/pinned-gate.py
timeout 300 python3 review/compiler/loom-r-cycle-targets-ac/check-headers.py before
# Insert only measured first-line refusal headers; increment map write expectation.
timeout 300 python3 review/compiler/loom-r-cycle-targets-ac/check-headers.py after
# Each real scan uses timeout 180 go build -o <gate-selected-binary> ./cmd/adamic
# and timeout 90 <that-branch-binary> c <each-listed-path>.
timeout 180 go test ./internal/ir ./internal/lower ./internal/oracle ./cmd/adamic-gate -run '^(TestEveryMapAProgramHoldsIsSorted|TestEveryTestIsParallelOrSaysWhy|TestNativeAgreementRowsNameRegisteredPrograms|TestRuntimeSetsAreEachUsed|TestClosureTargets.*|TestNamespaceAmbient.*|TestClosureCycle.*|TestGettersCensus.*|TestCountsAreRecordedTable)$' -count=1 -parallel=4 -json -timeout 90s
timeout 180 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -parallel=4 -json -timeout 90s
timeout 180 go vet ./internal/ir ./internal/lower ./internal/oracle
timeout 180 go test ./internal/native ./internal/flow ./internal/fresh ./internal/javascript -run '^(TestCallTargetsElementBorrowPlan|TestBorrowChainTargets|TestLoopCallCoverage|TestDebuggerHasNoFlowInstruction|TestRegexOperandsStillJudgeCycleWrites|TestViewMixedUnionSelectsNumber|TestFlowCorpusUnitsCoverEveryProgram)$' -count=1 -parallel=4 -json -timeout 90s
timeout 840 python3 review/compiler/loom-r-cycle-targets-ac/run-final-checks.py
```

The final runner repeats vet after restoring all implementation mutants and explicitly records every exit and wall time:

```text
timeout 600 python3 /workspace/adamic/review/compiler/loom-r-cycle-targets-ac/run-mutants.py
exit=0, seconds=27.613, log=mutants.log
timeout 180 python3 /workspace/adamic/review/compiler/loom-r-cycle-targets-ac/run-admission-mutant.py
exit=0, seconds=10.997, log=admission-mutant.log
timeout 180 go vet ./internal/ir ./internal/lower ./internal/oracle
exit=0, seconds=1.018, log=final-vet.log
```

The runner source records every exact inner mutant command. Formatting and strict docs/internal whitespace checks pass. The combined commit's only changes beyond bc53e520 in docs/internal are the three headers and the one map position expectation. Delivery uses one commit and a normal non-force push to compiler/loom-r-cycle-targets-ac; no existing branch is rewritten.

Delivery commands:

```text
timeout 120 git commit -F review/compiler/loom-r-cycle-targets-ac/commit-message.txt
timeout 120 git push origin HEAD:refs/heads/compiler/loom-r-cycle-targets-ac
timeout 120 git ls-remote origin refs/heads/compiler/loom-r-cycle-targets-ac
```
