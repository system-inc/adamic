Built the #7c6b4pq self-compare fix-forward: static inline helpers stay defined in every split translation unit through the shared header.
Commits: continued 387c2826; fix 9679fbe4; main 60397548 merged in 9ebca615 after the earlier main merge 510c9f33.
Commands: uncached self-compare tests passed in 15.503s; uncached split tests passed in 6.210s; new regression took 0.23s.
Mutants: restoring ordinary static-function splitting failed the regression with -Wundefined-inline; existing constant self-equality mutant was caught by Node's NaN output.
Not covered: full gate and whole packages; TSGo split tests skipped because ADAMIC_CLANG_TSGO_ARCHIVE is unset.

Read-only inspection confirmed self_compare.go emits static inline definitions into declarations, while splitC removes static and emits only an inline prototype to the header. Before fixing, the regression reproduced the product failure: functions_0001.c used adamic_unit_adamic_same_number without a definition in that translation unit. The error was -Werror,-Wundefined-inline, with the same inline _Bool prototype and used-here note.

The existing sweep alone fits in one sixteen-function group and passed before the fix. The new test reads that fixture and appends sixteen self-comparison functions to a temporary .a source so callers span groups. Node strips the TypeScript syntax and executes exactly that temporary source; its output and successful exit must match native. This conservative extension exercises the product's group boundary without adding a registered fixture. No new counts row is required. The native build uses the standard -Werror flags.

splitC now writes the whole static inline definition to the shared header, using the existing token renaming consistently, and emits no separate body unit. self_compare.go is unchanged.

The explicit requested compile-regression mutant disables that header path and restores the old splitter behavior. run-mutant.py saves the source, records a .patch, checks exit 1 and the exact diagnostic, and restores the source in finally. This is intentionally a compile-error mutant, as requested by this unit. The final uncached split rerun proves restoration.

Exact validation commands, each test output redirected directly to its named log:

```sh
source /workspace/adamic-tools/env.sh
timeout 60 go test ./internal/native -run '^TestSplitSelfCompareAgreesWithNode$' -count=1 -v -timeout 90s
timeout 75 python3 review/compiler/self-compare-main/run-mutant.py
ADAMIC_GATE_UNCACHED=1 timeout 60 go test ./internal/oracle -run '^TestSelfCompare' -count=1 -v -timeout 90s
ADAMIC_GATE_UNCACHED=1 timeout 60 go test ./internal/native -run '(?i)split|^TestUnitsPreserveSharedState$' -count=1 -v -timeout 90s
```

Focused fixed run: 0.258s package, 0.25s leaf. Before fix: exit 1, 0.238s package. Splitter mutant: exit 1, 0.187s package. Mutant runner: exit 0. Self-compare oracle leaves: constant mutant 11.99s, array-fill 12.32s, closure 15.49s, sweep 15.49s. All are under sixty seconds.

Setup: export GOPROXY='https://proxy.golang.org|direct' then timeout 240 bash cloud/setup.sh. First attempt hit the 240-second limit during dependency cache warming; the two initial focused invocations also hit their 60-second wall limits before any test output. Retried setup completed: Node ready 0.019s, Go ready 0.022s, submodules ready 0.061s, markdown dependencies ready 0.071s, clang ready 0.150s, Go build ready 33.362s, build cache warm 33.475s, done 33.503s. Sourced /workspace/adamic-tools/env.sh. nproc=5, cpu.max=400000 100000 (four CPUs).

No unlanded worker branch was merged. Only current origin/main was merged into the explicitly requested continuation branch. No numbered roadmap step was supplied in the unit heading; this delivers its named #7c6b4pq self-compare fix-forward and preserves step 79's test grain.

Integration lane-check output is recorded in lane-checks.log after committing. All commands the new test invokes (node and the native clang toolchain) are declared in origin/devtools/fast-gate:cloud/fast-gate/tools.txt.

Final re-green after merging main 60397548:

```sh
ADAMIC_GATE_UNCACHED=1 timeout 60 go test ./internal/native ./internal/oracle -run '(?i)split|^TestUnitsPreserveSharedState$|^TestSelfCompare' -count=1 -v -timeout 90s
```

Exit 0: native 0.996s, oracle 0.876s. New regression leaf 0.29s. The same two archive-dependent TSGo tests skipped.

Required lane command, after committing and from the repository root:

```sh
git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -
```

Final output: lane checks 1.5 s: gofmt and tools on 5 Go files, t.Parallel on 2 test packages; vet 2 packages. Exit 0. The five Go files include the prior self-compare branch changes; this fix changes units.go and adds units_self_compare_test.go.
