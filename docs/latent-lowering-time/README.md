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
- One trial, main's timed binary only, the meter's exact command plus the timing lines, under
  `/usr/bin/time -v` with `timeout 3600`: [trial/main.time](trial/main.time),
  [trial/main-per-file.txt](trial/main-per-file.txt) (file, wall s, process user s, all 79 files),
  and its per-file decision lines [trial/main.stderr-decisions.txt](trial/main.stderr-decisions.txt).

## What it shows so far (main only, one trial, observed)

- Whole process: exit 0, wall 53:49.6 (3,229.6 s), user 4,316.0 s, system 27.3 s, peak RSS
  4.66 GB. The loop over the 79 files is 3,226.2 s of that wall, so loading and the checker are
  about 3 s.
- Slowest ten, wall s (user s): `utilities.ts` 649.1 (894.7), `types.ts` 369.3 (584.9),
  `visitorPublic.ts` 248.0 (273.9), `utilitiesPublic.ts` 190.0 (270.2), `factory/nodeTests.ts`
  130.6 (184.0), `commandLineParser.ts` 119.9 (149.5), `moduleNameResolver.ts` 95.5 (131.0),
  `factory/utilities.ts` 76.1 (101.9), `transformers/ts.ts` 75.9 (83.8), `builder.ts` 73.4 (89.3).
  Those ten are 2,028 s, 63% of the loop.
- Even the smallest `_namespaces/*.ts` files take about 2 s each.
- Decisions: the trial's 79 lines match the Oct 27 meter's main (`215a0f84`, which has both
  4ac3f04d and 5e382ce4) on 78 files; `sys.ts` differs, refused 168 here against 182 there
  ([trial/main-vs-meter-20261027.diff](trial/main-vs-meter-20261027.diff)). Between the Oct 26 and
  Oct 27 meter runs, refused counts fell on 33 files; `visitorPublic.ts` fell from 349 to 40.
  The Oct 26 meter's main, `fa8909ad`, is not in this clone, so I can't say which of the two
  commits it had.

Inferred, not yet measured: the per-statement unit in the first-error loop calls
`noteInheritance`, `noteAccessorNames` and `latentRegister` over all of `program.Files()` for every
top-level statement, which would make even tiny files cost seconds. The profiles will say whether
that or the predicate checks dominate.

## What hasn't run yet, and why

Order, as Performance asked after push 1: a CPU profile of main's whole first-error loop over all
79 files (running now, [tools/whole_profile_overlay.py.txt](tools/whole_profile_overlay.py.txt)),
then the same on A, then one timed run of B, then the decision diffs across the three trees, then
the one-file against 79-file comparison, and only then a second round.

- The timeout is 3,600 s, not 900 s, since one process covers all 79 files and takes 3,230 s on main.
- Per-file invocations (one process per file) are not the meter's step: the meter loads all 79
  files as one program. Per-file seconds come from the in-process lines instead.
- The per-file named profiles ([tools/profiled_overlay.py.txt](tools/profiled_overlay.py.txt)) are
  replaced by the whole-loop profile and won't run.

No question or blocker is open.
