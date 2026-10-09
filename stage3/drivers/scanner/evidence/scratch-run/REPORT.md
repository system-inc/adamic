Built the one-command scratch scanner runner; neither native scanner built.
Refs: main 3ffb1a835184713998a34874e86326cd21db971f, combined 2648ad33; scratch 365444cd8d0e2bab5db6eb6fef3a93ebef904156.
Both full-tree Node diffs exit 0; first native stop debug.ts:7:1; fifteen fresh Node witnesses.
False-pass conflict-summary mutant exits 1; five targeted tests pass; both Node end mutants caught.
No native execution, runtime measurement or native-output mutant; no compiler/adaptation edits.

## Command and observations

```sh
STAGE3_CACHE=/workspace/scratch/native3-cache stage3/drivers/scanner/scratch-run.sh /workspace/scratch/scanner-scratch-command-proof 2648ad33 > /tmp/scanner-scratch-command-proof.log 2>&1
stage3/drivers/scanner/scratch-run.sh /workspace/scratch/scanner-scratch-conflicts ed6e2975 2648ad33 > /tmp/scanner-scratch-conflicts.log 2>&1
stage3/drivers/scanner/scratch-run.sh /workspace/scratch/scanner-scratch-original-conflicts 28285421 2648ad33 > /tmp/scanner-scratch-original-conflicts.log 2>&1
```

Exit codes: 1 (scanner-blocked), 2 (merge-conflict), 2 (merge-conflict).
All three starts use fetched main `3ffb1a835184713998a34874e86326cd21db971f`. The two conflict
runs stop before setup/compiler/scanner; no conflicting merge is resolved.
The command never pushes. Only the delivery branch is pushed after this unit.

The requested corePublic.ts:9:5 first stop is not reproduced. The combined ref
already closed MapLike in accepted 8822db17; the current first refusal is
`debug.ts:7:1`, namespace, in both split modes. The unchanged scanner builds directly observe this first refusal.

The requested ed6e2975 conflict count is also not reproduced: this ref produces
43 paths. The earlier 28285421 control reproduces all 44. Their sole difference
is `internal/native/runtime/node_process.c`, absent from the ed6e2975 conflict
set. Exact resolved SHAs, paths and failed merge commands are in
[conflict-summary.json](conflict-summary.json) and
[original-conflict-summary.json](original-conflict-summary.json).

The Node reference covers 81 files twice: 509,014 skip-trivia tokens plus
860,418 retained-trivia tokens, with 466 error rows. Both slice/reference diffs
exit 0, and both slice Node runs exit 0. Their raw SHA-256 is
`28e1f7008a99667d506d977f51b69d4e4a1f62f8d950ef54641cec71dbd5b9da` over
108,022,527 bytes. Absolute input paths appear
in pass headers. Normalizing just those paths proves the prior 11a64731 corpus
and this run identical at every remaining byte; see [corpus-check.json](corpus-check.json).

| Phase | Exit | Wall seconds |
| --- | --- | --- |
| setup | 0 | 191.57 |
| compiler-build | 0 | 5.888 |
| apply | 0 | 23.418 |
| reference | 0 | 9.952 |
| slice | 0 | 3.844 |
| scanner-0 | 1 | 7.044 |
| full-tree-check-0 | 0 | 0.215 |
| scanner-1 | 1 | 7.249 |
| full-tree-check-1 | 0 | 0.215 |

Total command wall time: 326.766s. Compiler binary SHA-256: `471bdfc3ffed7b2e63f64c537513d72997e96719154cec01d5ef83ce61e586f4`. nproc: 5; cgroup cpu.max: 400000 100000. Setup timing lines:

```text
setup: node ready (0.023s)
setup: go ready (0.025s)
setup: submodules ready (0.063s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.007s
setup: markdown dependencies ready (0.072s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.153s)
setup: go build ready (191.339s)
setup: test binaries deferred (use --warm-tests) (191.517s)
setup: build cache warm (191.518s)
setup: build-flags commit=365444cd8d0e2bab5db6eb6fef3a93ebef904156 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false load-before=1.57 3.14 2.18 1/210 33351 load-after=8.79 6.02 3.51 1/210 34889
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (191.551s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.ri6UQl
```

## Ordered discovery

Only split 0 is walked after both unchanged builds stop at the same namespace.
Every row's Node witness exits 0 and its native witness build exits 1 with the
same diagnostic. Stops after the first depend on recorded private throwing
placeholders; the final Debug location is in the namespace replacement.
Full messages, provenance and source positions are in [stops.json](stops.json).
Witness source files are under witnesses/. Raw fresh build/Node logs are in
main-records-logs.json.gz. placeholders.json.gz records every before/after
source change, including throwing bodies. Discovery stops at the fifteen bound.

| Stop | File:line:column | Message |
| --- | --- | --- |
| 1 | debug.ts:7:1 | Adamic 0.1 refuses a namespace; use a module: a file of its own, with named exports |
| 2 | types.ts:689:36 | stage 0 can't lower a reference with both null and undefined (nullable reference needs an empty-case tag) yet |
| 3 | diagnosticInformationMap.generated.ts:13:34 | Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) |
| 4 | utilities.ts:67:67 | Adamic 0.1 refuses the comma operator; write each expression as its own statement |
| 5 | scanner.ts:430:21 | Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| 6 | scanner.ts:476:24 | Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| 7 | scanner.ts:476:12 | Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| 8 | scanner.ts:500:67 | Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) |
| 9 | scanner.ts:534:24 | Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| 10 | core.ts:107:20 | stage 0 can't lower new an Identifier yet |
| 11 | utilities.ts:13:17 | stage 0 can't lower a function returning U \| undefined yet |
| 12 | utilities.ts:17:56 | stage 0 can't lower a function value taking string \| number yet |
| 13 | scanner.ts:258:27 | stage 0 can't lower a void call used as a value yet |
| 14 | scanner.ts:364:9 | Adamic 0.1 refuses a value as a condition; compare it explicitly, like name.length > 0 or count !== 0 |
| 15 | debug.ts:9:9 | stage 0 can't lower a function value with an optional parameter yet |

## Checks and mutants

```sh
python3 -B stage3/drivers/scanner/test_scratch_runner.py > /tmp/scanner-scratch-tests-final.log 2>&1
python3 -B stage3/drivers/scanner/scratch-summary.py stage3/drivers/scanner/evidence/scratch-run/conflict-summary.json > /tmp/scanner-scratch-valid-summary.log 2>&1
python3 -B stage3/drivers/scanner/scratch-summary.py stage3/drivers/scanner/evidence/scratch-run/conflict-pass-mutant.json > /tmp/scanner-scratch-mutant-summary.log 2>&1
bash -n stage3/drivers/scanner/scratch-run.sh > /tmp/scanner-scratch-shell-check.log 2>&1
git diff --check > /tmp/scanner-scratch-diff-check.log 2>&1
```

Five tests pass. The valid conflict summary exits 0. Its planted mutant changes
status to pass and supplies otherwise complete successful compiler/scanner
claims. The production validator exits 1 specifically because a merge
conflicted and exited nonzero; it is not caught by unrelated missing pass
fields. The tests also reject a missing split mode, compilation after conflict
and a native-byte mutant diff returning 0. The recorded comparison control
passes in each scanner mode, while each token-end mutant returns diff exit 1.
The full-tree reference's control and end mutant also pass their expected checks.
No native-output mutant was run, because neither scanner binary exists.
No compiler oracle fixture was added; internal/oracle/counts.md is unchanged.
No package-wide test or full gate was run.

## Environment failures and limits

Earlier private dependency setup attempts exposed symlink rejection and shared
submodule-worktree metadata problems. The final command uses independent Git
metadata and pinned checkouts, borrowing local objects only at a matching SHA;
it completes setup without changing the primary dependency checkouts.
Exact failed-attempt stdout/stderr and summaries are retained in
dependency-attempt-logs.json.gz. The symlink error was `error: expected submodule path 'cohere' not to be a symbolic link`; local shared-clone attempts reported `fatal: could not fetch 5752dd0f6f3f996f66a52c2ce34882eef10e850a from promisor remote`, followed by early EOF and invalid index-pack output. A linked worktree changed shared cohere
core.worktree configuration; that configuration was restored to the primary
checkout before switching to independent metadata. No such shared configuration
change occurs in the final runner.
An initial full-command attempt built the compiler but ran out of disk while
checking out TypeScript: `No space left on device`, apply exit 1, status failed.
Its exact logs and summary are retained in disk-full-logs.json.gz. Old,
regenerable Go cache entries were reclaimed and old private token streams
losslessly compressed. Completed conflict/failed source worktrees were retired
to recover space; their reports, logs and scratch branch refs remain. The final
measurement worktree and full raw outputs remain under its output directory.

The native-success branch of the runner delegates comparisons, the one-byte
native mutant and timings to the existing scanner tools. It could not be
exercised on these refs. New diagnostic kinds fail closed when no freshly
matching Node witness or recorded-safe placeholder is available. The witnesses
cover compiler admission, not native scanner correctness after placeholder edits.
