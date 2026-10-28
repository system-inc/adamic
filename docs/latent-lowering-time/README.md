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

## Main's profile (push 3, observed)

The first profile run died in a container restart at file 73 of 79: `pprof` writes only on stop,
so it left an empty profile. Its 73 timing lines are kept
([killed/main-profile-run-73-files.txt](killed/main-profile-run-73-files.txt)) and agree with the
trial: 2,083.4 s against 2,098.7 s on the same 73 files, per-file ratio 0.90 to 1.14.

The rerun writes one CPU profile per iteration of the first-error loop
([tools/loop_profile_overlay.py.txt](tools/loop_profile_overlay.py.txt)), merged with
`go tool pprof -proto` into the whole loop: 79 files, 3,538.4 s of profile duration, 4,733.9 s of
samples. With profiling on, the run passed 3,600 s at file 76, so I killed only the `timeout`
wrapper (the census process kept running, reparented to init, and finished all 79 files). Its
`time -v` ([profile/main/time-v-void.txt](profile/main/time-v-void.txt)) is therefore the
wrapper's, not the run's; use the trial's for the process. Its 79 decision lines are identical to
the trial's. Per-file lines: [profile/main/per-file.txt](profile/main/per-file.txt) (sum 3,540.3 s
wall, 4,705.4 s user; profiling costs about 10% on this run).

Tops: [profile/main/top-cum-40.txt](profile/main/top-cum-40.txt),
[profile/main/top-flat-40.txt](profile/main/top-flat-40.txt), and every node,
[profile/main/top-cum-all-nodes.txt](profile/main/top-cum-all-nodes.txt).

Cumulative shares of the 4,733.9 s, from the every-node listing:

| function | cum s | share |
|---|--:|--:|
| `lower.(*lowering).sourceIdentity.func1` (its walk) | 1,665.4 | 35.2% |
| `lower.(*lowering).noteLocal`, which calls `sourceIdentity` | 1,260.7 | 26.6% |
| `runtime.gcBgMarkWorker` (GC) | 1,187.8 | 25.1% |
| `lower.(*lowering).latentRegister` (overlay) | 1,153.9 | 24.4% |
| `lower.(*lowering).provePredicate` | 863.2 | 18.2% |
| `lower.(*predicateVerifier).verify` | 711.0 | 15.0% |
| `lower.(*lowering).checkedPredicateBodyProven` | 638.2 | 13.5% |
| `lower.(*lowering).latentRefuse` (the refusal scan, overlay) | 503.0 | 10.6% |
| `lower.(*lowering).checkedPredicateDeclaration` | 220.6 | 4.7% |
| `lower.(*lowering).checkedPredicateArgumentProven` | 165.6 | 3.5% |
| `lower.(*lowering).lowerFunction` | 117.2 | 2.5% |
| `lower.(*predicateFalseFlow).escapingReference` | 116.1 | 2.5% |
| `lower.(*predicateFalseFlow).escapes.func1` | 116.0 | 2.5% |
| `lower.(*predicateFalseFlow).checkedDeclaration` | 116.2 | 2.5% |
| `lower.(*lowering).noteInheritance` | 70.6 | 1.5% |
| `lower.(*lowering).noteAccessorNames` | 70.3 | 1.5% |
| `checker.(*Checker).GetTypeAtLocation` | 77.0 | 1.6% |
| every other `predicateFalseFlow` method | under 0.3 s each | 0% |

Flat, `maps.ctrlGroup.matchH2` (map probing) is 11.2% and `ForEachChild` 4.0%; `ast.ForEachChild`
is 71.8% cumulative.

Caveat: `latentRun` shows only 38.4% cumulative while `ForEachChild` shows 71.8%. The deep
recursive AST walks overflow the profiler's stack depth, so many samples lose their outer
frames. The shares above are lower bounds for callers far from the leaves (`latentRun`,
`provePredicate`); `sourceIdentity.func1` is a leaf-side closure and is not affected that way.

What the code says (read, not measured): `sourceIdentity` indexes a whole source file once per
`lowering` and caches it in `l.sourceIdentities`. The latent first-error loop builds a `fresh()`
`lowering` for every top-level statement, so the cache starts empty each time and every
statement that declares a local re-walks its whole file: cost grows with statements times file
size, which fits `utilities.ts` (the largest file) at 649 s. `latentRegister`, `noteInheritance`
and `noteAccessorNames` run once per statement over all of `program.Files()`. All of these exist
in A as well; of the suspects, only `predicateFalseFlow.escapes` is absent from A. A's profile
decides how much is new.

## What hasn't run yet, and why

Order, as Performance asked after push 1: a CPU profile of main's whole first-error loop over all
79 files (done, above), then the same on A (running now, timeout raised to 7,200 s), then one timed run of B, then the decision diffs across the three trees, then
the one-file against 79-file comparison, and only then a second round.

- The timeout is 3,600 s, not 900 s, since one process covers all 79 files and takes 3,230 s on main.
- Per-file invocations (one process per file) are not the meter's step: the meter loads all 79
  files as one program. Per-file seconds come from the in-process lines instead.
- The per-file named profiles ([tools/profiled_overlay.py.txt](tools/profiled_overlay.py.txt)) are
  replaced by the whole-loop profile and won't run.

No question or blocker is open.
