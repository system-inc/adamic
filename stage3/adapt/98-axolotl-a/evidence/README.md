# Evidence for 98-axolotl-a

Box: this session's Linux container, Node v24.22.2, Go 1.27.1, stock TypeScript 6.0.3 from stage3/api (its digest
matches api/node_modules.sha256). Upstream source: a depth-1 clone of v6.0.3 at
050880ce59e30b356b686bd3144efe24f875ebc8.

- apply-before.log, apply-after.log: `stage3/apply.sh` on main's tip (a23ba096) and on the head; both exit 0.
  patch-set-before.md, patch-set-after.md: their tables.
- census-*.log, census-*.jsonl.gz: `go build ./stage3/census/tool` run over each tree's src/compiler, as the meter runs it.
- count-*.txt: own-file counts through stage3/meter/report.py's report() (count.py). compare-*.txt: every file whose
  count moved (compare.py). Before 6 and 6, after 0 and 0, mutant classFields.ts 3, mutant2 classFields.ts 1.
- transpile/: both files transpiled before and after, comments kept and removed (transpile.cjs); the four .diff files
  are empty. tree-diff.txt: the two trees differ in these two files only.
- unit-run1.log, unit-run2.log, idem-*: first and second runs; the second makes 0 assertions and leaves the same bytes.
- pin90.txt: es2017.ts in the before tree matches 90's afterSha256, which is what 90 sees during apply.
- mutant.diff, mutant2.diff: the two probes, each on a scratch copy of the applied tree.
- setup.log, setup-retry.log: `cloud/setup.sh --wasi-sdk` failed both times downloading LLVM (connection reset);
  nothing here needs clang.
- The sharded oracle was not run: under Correctness's rule it is Loom's run of the wave.
