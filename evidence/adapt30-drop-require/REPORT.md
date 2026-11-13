Task 93w9rtz, 2026-11-13: verification stopped at an inherited replay guard.

Change commit: `14b54262fc72c73bd515e10fd0d0fa6dade9bf2a`, one commit above fetched main `6286add0122cb4fb0126384bfd6f5799b7a4fb0d`.
The require declaration alone is removed; process is byte-identical. No internal/ file changes.

Actual in-place full replay exits **1** at `declined site changed: core.ts:992:21`.
Untouched main's adaptation 30 checking untouched main's full apply exits **1** with
exactly the same message. The final composed-tree check also exits **1**. This is
an existing addRange decline/composition mismatch, not the require guard. No
unrelated addRange amendment was attempted. The fresh-cache apply replay is
byte-identical, but is not claimed as a successful in-place pipeline replay.

Completed checks: full applies **0**; 30-after check/mutant/restoration children
**0/1/0**, mutant wrapper **0**; 30-after zero-edit replay **0**; stock tsc 6.0.3 and
adamic types **0** on both sets of 79 roots (zero diagnostics); emitted-JS comparison
**0** for 116 emitting .ts files in ESNext and CommonJS, plus one unchanged .d.ts;
source/process audit **0**; API and shard-plan regeneration comparisons **0**.
Whole lane **0**, with 106366 passing, one sanctioned API failure, zero pending;
its oracle child exits **1**, as required by the existing sanction.

Both complete bundle census wrappers and their comparison exit **0**. All 79
outcomes are unchanged: 73 stop at process, core.ts:2599:15; two compile; four
retain other stops. In particular the core.ts root stops at performanceCore.ts:29:15.
All individual CLI exits and first-stop text are retained.

**102/398 unique shard units completed and passed (child exit 0).** The complete
24-unit project-reference batch exits **0**. The initial collector was stopped
(**143**) to use bounded batches; the file batch was then stopped (**143**) after
the replay blocker was reproduced. Remaining shards and the union merge were
not run. No full shard pass is claimed. Completed units stayed below 30 seconds;
all completed producer, whole-lane and census processes stayed below 30 minutes.

`checks.json` contains every check's exit, completed unit records, and earlier
setup/retry failures. `proofs.tar.gz` contains 1262 raw evidence files, proof
scripts, both 79-root reports, emitted hashes, whole-lane records, and completed
shard observations/diffs. MANIFEST.json records each member's SHA-256 and size.
Archive SHA-256: `90b95857c4d01cdd74d3605743f1d44e447969d94b71c3e5fb0ab55cd04b18c6`. Archive integrity check exits **0**.

Extract with `tar -xzf proofs.tar.gz`. The replay command was
`bash stage3/apply.sh --adapt-checkout ISOLATED_FINAL_TREE`; see
`proofs/in-place-full-replay.log` and both `proofs/check-*-composed-tree.log` files.
Proof scripts use stock TypeScript 6.0.3 on NODE_PATH; full apply requires the
pinned TypeScript checkout named by ADAMIC_TYPESCRIPT_SOURCE. The literal stock
tsc commands, including --ignoreConfig and locked Node host typings, are recorded
in `proofs/stock-tsc.json`. There is no Go evidence source in this package.
