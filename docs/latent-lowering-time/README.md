# Latent lowering time, partial evidence (push 1)

This is a progress push. The measurement is still running; `docs/latent-lowering-time.md` comes
with the final push. Nothing here compares the trees yet.

## What ran

- Box: Linux 6.18.44, 4 CPUs (`nproc` 4), `Intel(R) Xeon(R) Processor @ 2.10GHz`, 15 GB RAM,
  no swap. `/proc/loadavg` before any build: `0.26 0.06 0.02`; at the trial's start:
  `1.28 2.11 1.70` (the tail of the builds); during it: `1.48 1.28 1.35`. Go 1.27.1, Node 22.22.0
  (apply accepted it), GNU time 1.9 installed from Ubuntu's `time` package. clang not needed.
- Trees, each `git worktree add --detach` plus `git submodule update --init --recursive --depth 1`,
  all pinning cohere `06af2e39` and cohere/TypeScript `d92d9bfe`:
  - main `85c9d7837d9b84f5a07c656c67b620d441349559` (origin/main at the start)
  - A `76ea8b17ea0ed467ecb86f355c06f738fd262983` (4ac3f04d's parent)
  - B `22026c7392f55556e721b6f425dd369ceafdf14f` (5e382ce4's parent; contains 4ac3f04d, lacks 5e382ce4)

  B's first submodule update failed on `could not lock config file .../.git/config` because I ran
  the three updates at once; the retry alone succeeded.
- The meter's latent step, read from `stage3/meter/twice-daily.sh` (`build_census`, per-ref mode):
  `make_overlay.py <tree> <dest>/latent-overlay`, `gofmt -w`, then
  `go build -buildvcs=false -overlay=<dest>/latent-overlay/overlay.json -o latent-census ./stage3/census/latent/tool`,
  run as
  `LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 latent-census <adapted>/src/compiler latent.jsonl`.
  It is **one process over the whole directory**: `LatentLoad` takes all 79 paths as one program, and
  `lower.Latent` walks `program.Files()`, buffering every record and printing the per-file stderr lines
  only at the end. `stage3/census/latent` is byte-identical in all three trees, so each tree's own
  overlay is the same script over its own `internal/lower` and `internal/load`. All three overlays and
  binaries built; none refused.
- One adapted tree from main's `stage3/apply.sh` (TypeScript v6.0.3 `050880ce`, `stage3/api` by
  `npm ci`), shared by all three: `ADAMIC_TYPESCRIPT_SOURCE=<clone> bash stage3/apply.sh <adapted>`.
  Exit 0, 79 `.ts` files under `src/compiler`.
- Per-file timing inside the process: a throwaway copy of each overlay
  ([tools/timed_overlay.py.txt](tools/timed_overlay.py.txt)) adds one stderr line per file,
  `LATENT_TIMING <file> <wall s> <process user s>`, around the body of the first-error loop over
  `program.Files()`. It is never committed into the compiler. A second throwaway
  ([tools/profiled_overlay.py.txt](tools/profiled_overlay.py.txt)) adds a CPU profile around named
  files; built for main and A, not yet run.
- One trial, main's timed binary only, to learn the scale before the interleaved runs:
  [trial/main-partial-per-file.txt](trial/main-partial-per-file.txt) holds the first 69 of 79 files
  (file, wall s, user s), copied while the run was at 1,690 s elapsed.

## What it shows so far (main only, one trial, observed)

- The 69 files took 1,661 s of wall and 2,132 s of process user time in the loop.
- Slowest: `factory/nodeTests.ts` 130.6 s, `commandLineParser.ts` 119.9 s,
  `moduleNameResolver.ts` 95.5 s, `factory/utilities.ts` 76.1 s, `transformers/ts.ts` 75.9 s,
  `builder.ts` 73.4 s, `emitter.ts` 66.2 s, `program.ts` 65.0 s.
- Even the smallest `_namespaces/*.ts` files take about 2 s each.

Inferred, not yet measured: the per-statement unit in the first-error loop calls
`noteInheritance`, `noteAccessorNames` and `latentRegister` over all of `program.Files()` for every
top-level statement, which would make even tiny files cost seconds. The profiles will say whether
that or the predicate checks dominate.

## What hasn't run yet, and why

- The trial is still finishing. The whole-process `/usr/bin/time -v` runs (timeout 3,600 s, not
  900 s, since one process covers all 79 files and the trial alone passes 1,690 s), interleaved
  main, A, B and run twice, are next. At this scale that is about six hours of runs.
- Per-file invocations (one process per file) are not the meter's step: the meter loads all 79
  files as one program. Per-file seconds come from the in-process lines instead.
- CPU profiles of main's 3 slowest files and the same on A, and the decision diffs between trees.

No question or blocker is open.
