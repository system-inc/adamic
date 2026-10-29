# Latent lowering time on tsc's 79 files

Asked: why the stage 3 meter went from 59m 18s (Oct 26) to 72m 4s (Oct 27), by timing and profiling the
meter's latent step, the only place lowering runs on tsc's checker-rejected program. The full record,
push by push, is [latent-lowering-time/README.md](latent-lowering-time/README.md), and the shared-memo build
is [latent-lowering-time/shared-memos/README.md](latent-lowering-time/shared-memos/README.md).

## Answer

The cost was not the predicate passes suspected at first (the false-direction walk 4ac3f04d and its escape
scan were 2.5% of main's samples). It was **`sourceIdentity`** (from 1a24528863, Oct 26), which caches per
`lowering`, inside a step that makes a **fresh `lowering` for every top-level statement**. Every statement
re-walked every module: 35% of samples, plus most of the 25% that was GC. Behind it came per-lowering
memos that likewise reset each statement: the predicate proof fixed point (43% once identities were
shared) and the assignment walk under `predicateArguments` (25% after that).

Fixes, each with `latent.jsonl` and the 79 decision lines byte-identical to the tree before it:

| tree | what it adds | latent step wall | against |
|---|---|--:|---|
| main `85c9d783` | | 53:50 | |
| main `95a7b63f` | | 1:15:34 | (another hour of the box) |
| `f2479703` compiler/latent-shared-identities | one `sourceIdentities` map per program | 32:48 | 0.43 of `95a7b63f`, back to back |
| `dab16234` + `f27967ea` predicate-proof-memo | the proof's fixed point once per lowering | 20:16 | 0.53 of `f2479703`, back to back |
| `a1ce733b` + `992a6f53` entries-record-memo-2 | entries answers once per lowering | 21:55, 20:11 | no measurable gain: in the meter its memo misses every time |
| `4a9c4665` + `761ee21c` assigned-symbol-memo-3 | one assignment walk per lowering | 15:32 | 0.77 of `a1ce733b`, back to back |
| shared memos (overlay only, `shared-memos/`) | proofs and assigned symbols shared per (program, checker) | 14:33 | 0.94 of `4a9c4665`, back to back |

Landed on main (all four in `51981922`, with Correctness's statementrewrite fix `55c1c38f` cherry-picked since
the meter's overlay refuses `51981922` without it): **13:00 against `95a7b63f`'s 1:02:23, back to back (0.21)**,
while main's new void-container refusals raise the findings from 3,940 to 13,402. Reverting #3hxh5jm
(`07b24606`) changes 17 files' findings and the time by 0.5%
([latent-lowering-time/landed/README.md](latent-lowering-time/landed/README.md)).

`95a7b63f` against `85c9d783` (1.4 times) is the box, not the code: alone on one box they take 25.5 s and
27.0 s on `factory/nodeTests.ts`, and the same binary ran 12% slower an hour later with nothing changed. Only
back-to-back pairs compare.

The time per statement grows with the program's file count: `factory/nodeTests.ts` takes 27.0 s alone and
130.6 s inside the 79-file program on `85c9d783`, with the same decisions.

What leads now, on the four heads: the entries walk (`entriesRecordLiteralWalk` 39.6% of samples), whose
memo is per lowering and so misses on every fresh lowering in the meter; `enumNeverValue` (29.6%); and the
per-statement walks over all files (`noteInheritance`, `noteAccessorNames`, `latentRegister`, 19.4%).

## Box and method

Linux 6.18.44, 4 CPUs (`Intel(R) Xeon(R) Processor @ 2.10GHz`), 15 GB, no swap; `/proc/loadavg` before any
build `0.26 0.06 0.02`, and each run's start and end load is recorded beside it. Go 1.27.1, Node 22.22.0,
GNU time 1.9. Each tree is its own `git worktree add --detach` with submodules (cohere `06af2e39`,
TypeScript `d92d9bfe` in all), built as `stage3/meter/twice-daily.sh`'s `build_census` builds the latent
binary (`make_overlay.py`, `gofmt -w`, `go build -buildvcs=false -overlay=... ./stage3/census/latent/tool`),
with a throwaway per-file timing line ([tools/timed_overlay.py.txt](latent-lowering-time/tools/timed_overlay.py.txt)),
over one adapted tree from main's `stage3/apply.sh` (TypeScript v6.0.3 `050880ce`). Each run is the meter's
command, `LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 latent-census <adapted>/src/compiler latent.jsonl`, under
`/usr/bin/time -v`. It is one process over all 79 files, so per-file seconds come from the timing lines.
Profiles are one CPU profile per file iteration, merged with `go tool pprof -proto`.

## What this didn't do, and why

- **Per-file processes, interleaved and twice, for main, A and B.** The meter runs one process over all 79
  files, so per-file processes would not be the meter's step. After main's profile named the cost,
  Performance dropped A's profile, B's run and the A/B decision diffs (the passes A and B lack were no longer
  suspects). Repeats were replaced by back-to-back pairs, since the box drifts by up to 12% over hours.
- **Timeouts.** One process takes up to 75 minutes, so the timeout is 7,200 s, not 900 s.
- **Interruptions.** Two container restarts killed runs, which were rerun; I stopped one run myself by
  mistake and reran it. One build ran the disk out; I removed four worktrees and the Go build cache and
  rebuilt. No reported number comes from an interrupted run.
- **Profile caveat.** Deep AST recursion overflows the profiler's stack depth, so cumulative shares for
  callers far from the leaves (`latentRun`, `provePredicate`) are lower bounds.
- **Local merges.** `dab16234`, `a1ce733b`, `4a9c4665` and the shared-memo tree exist only on this box. The
  four-head tree cherry-picks `761ee21c` rather than merging it, so its base's `process.platform` change stays out.
