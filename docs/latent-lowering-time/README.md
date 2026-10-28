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

## One shared source identity map, throwaway (push 4, observed)

Performance's ask after push 3: in main's (`85c9d783`) timed overlay only, declare one
`map[*ast.Node]ir.SourceIdentity` before the first-error files loop in `latentRun` and have
`fresh()` set `sourceIdentities` to it ([tools/shared_identity_overlay.py.txt](tools/shared_identity_overlay.py.txt);
the whole change is those two lines). A's profile run was stopped at file 8 for this.

| run | wall | user s | system s | peak RSS |
|---|--:|--:|--:|--:|
| main trial, per-lowering map | 53:49.6 (3,229.6 s) | 4,316.0 | 27.3 | 4.66 GB |
| main, one shared map | 28:41.9 (1,721.9 s) | 1,833.7 | 26.3 | 4.49 GB |

The loop's per-file sum is 1,718.5 s wall, 1,830.5 s user
([shared-throwaway/per-file.txt](shared-throwaway/per-file.txt)): 0.53 of the trial's wall and
0.42 of its user time. `latent.jsonl` is byte-identical to the trial's (`cmp`, 20,129,953 bytes,
sha256 `348511ff...`), and the 79 stderr decision lines are identical
([shared-throwaway/diff-against-trial.txt](shared-throwaway/diff-against-trial.txt)).

Per file, against the trial: `types.ts` 369.3 to 40.9 s, `factory/nodeTests.ts` 130.6 to 43.1,
`utilities.ts` 649.1 to 330.5, `utilitiesPublic.ts` 190.0 to 101.9, `commandLineParser.ts` 119.9 to
61.7. Twelve files keep at least 0.9 of their time, among them `visitorPublic.ts` (248.0 to 241.0)
and `transformers/ts.ts` (75.9 to 73.4): their cost is elsewhere.

This throwaway differs from the head `compiler/latent-shared-identities` (`f2479703`) in one way:
the head also has `statecopy` pass the map through, while here the generated state copier still
copies it at each snapshot. So the head may be faster than this; its own run follows.

## The head, f2479703, as the meter builds it (push 5, observed)

`compiler/latent-shared-identities` at `f2479703` (one commit on main `95a7b63f`), in its own
worktree with submodules (cohere `06af2e39`, TypeScript `d92d9bfe`), built exactly as
`build_census` does: its own `make_overlay.py`, `gofmt -w`, then
`go build -buildvcs=false -overlay=... ./stage3/census/latent/tool`, plus the throwaway timing lines
([tools/timed_overlay.py.txt](tools/timed_overlay.py.txt)). Same adapted tree, same command. A
container restart killed the first attempt at file 73; this is the rerun.

| run | wall | user s | system s | peak RSS |
|---|--:|--:|--:|--:|
| main `85c9d783` trial | 53:49.6 | 4,316.0 | 27.3 | 4.66 GB |
| main `85c9d783` + throwaway shared map | 28:41.9 | 1,833.7 | 26.3 | 4.49 GB |
| head `f2479703` | 32:47.9 (1,967.9 s) | 2,122.1 | 44.9 | 4.44 GB |

Per-file sum 1,956.8 s wall, 2,117.7 s user ([head/per-file.txt](head/per-file.txt),
[head/head.time](head/head.time)).

Output against the `85c9d783` trial: `latent.jsonl` differs (`cmp`: first at line 75; 20,131,373
bytes against 20,129,953), and two decision lines differ
([head/diff-against-main-85c9d783-trial.txt](head/diff-against-main-85c9d783-trial.txt)):
`utilities.ts` refused 781 to 782 and `watch.ts` 43 to 44. Record by record
([head/latent-jsonl-vs-main-85c9d783-trial.txt](head/latent-jsonl-vs-main-85c9d783-trial.txt)), each
of those two files gains exactly one finding and loses none, and no unit differs:
`utilities.ts:10747:65` and `watch.ts:994:28`, both "Adamic 0.1 refuses a read of a void value, as
a returned value ... (adamic/no-void-read)". That rule comes with `95a7b63f`'s
`internal/lower/void_reads.go`, which `85c9d783` lacks, so I expect it from the base and not from
the shared map. The run of `95a7b63f` itself (running now) decides it.

### Against its own base, main `95a7b63f` (push 6, observed)

`95a7b63f` built the same way (its own `make_overlay.py`, `gofmt -w`, `go build -overlay`, plus the
timing lines), same adapted tree, run right after the head:

| run | wall | user s | system s | peak RSS |
|---|--:|--:|--:|--:|
| main `95a7b63f` | 1:15:34 (4,534 s) | 6,126.5 | 78.6 | 4.71 GB |
| head `f2479703` | 32:47.9 (1,967.9 s) | 2,122.1 | 44.9 | 4.44 GB |
| head / main | 0.43 | 0.35 | | 0.94 |

Per-file sums: `95a7b63f` 4,528.6 s wall, 6,122.1 s user
([main-95a7b63f/per-file.txt](main-95a7b63f/per-file.txt)).

**Output: the head's `latent.jsonl` is byte-identical to `95a7b63f`'s** (`cmp`, 20,131,373 bytes,
sha256 `395d5157...` both), and the 79 stderr decision lines are identical
([head/diff-against-main-95a7b63f.txt](head/diff-against-main-95a7b63f.txt)). `95a7b63f` gives the
same two extra void-read refusals against `85c9d783` as the head does, so they come from the base
(`8d39519a`, "lower: a read of a void value is refused", the one lowering commit between
`85c9d783` and `95a7b63f`), not from the shared map. The `sys.ts` difference from the Oct 27 meter
was already absent from every run here; only the meter's `215a0f84` had 182.

Per file, the three compilers side by side: [head/per-file-three-runs.txt](head/per-file-three-runs.txt).
The head against `95a7b63f`: `types.ts` 584.3 to 44.3 s (0.08), `factory/nodeTests.ts` 190.0 to
48.3 (0.25), `utilities.ts` 928.3 to 380.8 (0.41). `visitorPublic.ts` (310.1 to 285.0, 0.92) and
`transformers/ts.ts` (90.1 to 82.1, 0.91) barely move.

Not explained yet: `95a7b63f` is about 1.4 times slower than `85c9d783` on nearly every file
(`utilities.ts` 928.3 against 649.1, `types.ts` 584.3 against 369.3, total user 6,126 against
4,316 s). A per-statement cost in `8d39519a`'s void-read refusal would look like this, and so would
a slower box after the container restart, which came between the two runs; I haven't separated
them. The head is likewise about 14% slower than the throwaway shared map on `85c9d783`. The
head's profile, next, shows the void-read pass's share.

## The head's profile (push 7, observed)

The head `f2479703`'s meter-built overlay with the per-file loop profile
([tools/loop_profile_overlay.py.txt](tools/loop_profile_overlay.py.txt)), all 79 files: 2,137.0 s of
profile duration, 2,337.1 s of samples (main `85c9d783`'s: 3,538.4 and 4,733.9). Its `latent.jsonl`
is byte-identical to the head's timed run and its decision lines are identical. Wall 35:42.8 and
user 2,318.7 s with profiling on ([profile/head/head-profile.time](profile/head/head-profile.time)).
The first attempt was killed by my own mistake (stopping a shell that was waiting on it) and rerun.

Tops: [profile/head/top-cum-40.txt](profile/head/top-cum-40.txt),
[profile/head/top-flat-40.txt](profile/head/top-flat-40.txt), every node
[profile/head/top-cum-all-nodes.txt](profile/head/top-cum-all-nodes.txt).

Cumulative seconds and share, main `85c9d783` against the head (the same stack-depth caveat as above:
callers far from the leaves are lower bounds):

| function | main `85c9d783` | head `f2479703` |
|---|--:|--:|
| `lower.(*lowering).sourceIdentity` | 1,287.9 s, 27.2% | 3.1 s, 0.1% |
| `lower.(*lowering).latentRegister` (overlay) | 1,153.9 s, 24.4% | 48.0 s, 2.1% |
| `runtime.gcBgMarkWorker` (GC) | 1,187.8 s, 25.1% | 209.3 s, 9.0% |
| `lower.(*lowering).provePredicate` | 863.2 s, 18.2% | 1,005.4 s, 43.0% |
| `lower.(*predicateVerifier).verify` | 711.0 s, 15.0% | 843.9 s, 36.1% |
| `lower.(*lowering).checkedPredicateBodyProven` | 638.2 s, 13.5% | 737.1 s, 31.5% |
| `checker.(*Checker).GetSymbolAtLocation` | 486.9 s, 10.3% | 597.1 s, 25.6% |
| `lower.(*lowering).latentRefuse` (the refusal scan, overlay) | 503.0 s, 10.6% | 590.2 s, 25.3% |
| `lower.(*lowering).entriesRecordLiteral` | 322.5 s, 6.8% | 369.6 s, 15.8% |
| `lower.(*lowering).lowerFunction` | 117.2 s, 2.5% | 138.7 s, 5.9% |
| `lower.(*predicateFalseFlow).escapingReference` | 116.1 s, 2.5% | 131.1 s, 5.6% |
| `checker.(*Checker).GetTypeAtLocation` | 77.0 s, 1.6% | 97.5 s, 4.2% |
| `lower.(*lowering).noteInheritance` | 70.6 s, 1.5% | 71.6 s, 3.1% |
| `lower.(*lowering).refuseVoidWidening` | (absent) | 20.3 s, 0.9% |
| `lower.(*lowering).refuseVoidRead` | (absent) | 0.06 s, 0.0% |

What's left on the head is the predicate proof: `provePredicate` through `predicateArguments`
(42.9%), into `predicateVerifier.verify` and `checkedPredicateBodyProven`, plus the refusal scan
and the checker's symbol resolution under them. `entriesProgramHasRecords` (under
`entriesRecordLiteral`, 15.7%) walks the program, which by its name is also a candidate for a
per-lowering memo; I haven't read it.

On the 1.4 times slowdown of `95a7b63f` against `85c9d783`: `8d39519a`'s void passes are under 1%
of the head's samples, and functions whose code did not change between the two mains
(`provePredicate`, `latentRefuse`, `GetSymbolAtLocation`, `entriesRecordLiteral`) each take 1.15 to
1.2 times their seconds on `85c9d783`. That points to a slower box after the second container
restart rather than to new work, inferred, not measured. A same-box check follows.

## One file alone against the 79-file program (push 8, observed)

Does the time per statement grow with the number of files? `factory/nodeTests.ts` named alone
(`latent-census-timed <adapted>/src/compiler/factory/nodeTests.ts out.jsonl`, the same env; `WalkDir`
on a file yields just that file, so `program.Files()` holds one root while the checker still loads
its whole import closure), on each compiler, one after another on today's box
([one-file/](one-file)):

| compiler | alone: loop s (user s) | process wall | in the 79-file program: loop s | 79-file / alone |
|---|--:|--:|--:|--:|
| main `85c9d783` | 27.0 (30.4) | 32.6 s | 130.6 (trial) | 4.8 |
| main `95a7b63f` | 25.5 (29.0) | 29.7 s | 190.0 | 7.4 |
| head `f2479703` | 19.6 (21.4) | 24.5 s | 48.3 | 2.5 |

Yes: on main, the same file's statements cost 4.8 times as much inside the 79-file program as alone,
with the same decision line both ways (`factory/nodeTests.ts 228 227`). The shared identity map
removes part of that growth; the rest is the per-statement walks over all of `program.Files()`
(`latentRegister`, `noteInheritance`, `noteAccessorNames`) and whatever under the predicate proof
scales with the program.

Drift: alone and on the same box, `95a7b63f` (25.5 s) is not slower than `85c9d783` (27.0 s), while
inside the program it measured 190.0 against 130.6 s, the two runs hours and a container restart
apart. That supports a slower box for the second, not new work. A same-box rerun of `85c9d783` over
all 79 files is running now to settle it.

## The predicate proof memo, f27967ea, with the shared identities (push 9, observed)

`compiler/predicate-proof-memo` at `f27967ea` (one commit on main `b7b5ac9e`), merged locally into
`f2479703` as a throwaway (`dab16234`, `git merge --no-ff`, clean, never pushed). Its diff from
`f2479703` under `internal/lower`, `internal/load`, `internal/ir` and `stage3/census` is exactly
`f27967ea`'s own diff (`95a7b63f..b7b5ac9e` changes nothing there); the merge also brings three
conformance test files. Built as the meter builds it plus the timing lines, same adapted tree,
then `f2479703` run again right after it on the same box:

| run | wall | user s | peak RSS | loop sum wall s (user s) |
|---|--:|--:|--:|--:|
| memo merge `dab16234` | 20:15.8 (1,215.8 s) | 1,292.2 | 4.05 GB | 1,211.3 (1,287.9) |
| `f2479703`, rerun right after | 37:58.2 (2,278.2 s) | 2,472.1 | 4.41 GB | 2,272.9 (2,466.6) |
| memo / `f2479703` | 0.53 | 0.52 | 0.92 | |

**Output: identical.** `cmp` finds no difference in `latent.jsonl` (20,131,373 bytes; sha256
`395d5157...` for both, the same as every earlier run of `f2479703` and `95a7b63f`), and the 79
stderr decision lines are identical ([memo/diff-against-f2479703-rerun.txt](memo/diff-against-f2479703-rerun.txt)).

Per file ([memo/per-file-against-f2479703-rerun.txt](memo/per-file-against-f2479703-rerun.txt),
[memo/per-file.txt](memo/per-file.txt), [head-rerun/per-file.txt](head-rerun/per-file.txt)):
`utilities.ts` 457.9 to 255.4 s, `visitorPublic.ts` 305.0 to 150.0, `utilitiesPublic.ts` 124.4 to
49.1, `transformers/ts.ts` 89.6 to 34.0, `factory/nodeTests.ts` 54.6 to 14.5, `binder.ts` 57.6 to
15.5. A few barely move: `moduleNameResolver.ts` (0.98), `types.ts` (0.95), `commandLineParser.ts`
and `builder.ts` (0.89).

The box drifts: this `f2479703` rerun is 1.16 times its earlier run (2,272.9 against 1,956.8 s of loop
wall) with byte-identical output, so only same-box pairs compare. This pair ran back to back.

The merge's profile is running now; then the three-way merge with `compiler/entries-record-memo-2`.

### The memo merge's profile (push 10, observed)

The memo merge's meter-built overlay with the per-file loop profile, all 79 files: 1,359.9 s of
profile duration, 1,468.9 s of samples; `latent.jsonl` byte-identical to its timed run. Tops:
[profile/memo/top-cum-40.txt](profile/memo/top-cum-40.txt),
[profile/memo/top-flat-40.txt](profile/memo/top-flat-40.txt), every node
[profile/memo/top-cum-all-nodes.txt](profile/memo/top-cum-all-nodes.txt).

Cumulative, `f2479703` (push 7, a different hour of the box) against the memo merge:

| function | `f2479703` | memo merge |
|---|--:|--:|
| `lower.(*lowering).provePredicate` | 1,005.4 s, 43.0% | 76.8 s, 5.2% |
| `lower.(*predicateVerifier).verify` | 843.9 s, 36.1% | 65.9 s, 4.5% |
| `lower.(*lowering).checkedPredicateBodyProven` | 737.1 s, 31.5% | 76.9 s, 5.2% |
| `lower.(*lowering).predicateArguments` | 1,002.1 s, 42.9% | 286.0 s, 19.5% |
| `lower.(*lowering).latentRefuse` (refusal scan) | 590.2 s, 25.3% | 50.6 s, 3.4% |
| `lower.(*lowering).entriesRecordLiteral` | 369.6 s, 15.8% | 421.5 s, 28.7% |
| `lower.(*lowering).entriesProgramHasRecords` | 202.4 s, 8.7% | 231.0 s, 15.7% |
| `checker.(*Checker).GetSymbolAtLocation` | 597.1 s, 25.6% | 396.6 s, 27.0% |
| `lower.(*lowering).enumNeverValue` | 288.5 s, 12.3% | 337.9 s, 23.0% |
| `lower.(*lowering).lowerFunction` | 138.7 s, 5.9% | 165.3 s, 11.3% |
| `lower.(*predicateFalseFlow).escapingReference` | 131.1 s, 5.6% | 17.2 s, 1.2% |
| `lower.(*lowering).noteInheritance` | 71.6 s, 3.1% | 82.7 s, 5.6% |
| `lower.(*lowering).latentRegister` | 48.0 s, 2.1% | 54.1 s, 3.7% |
| `checker.(*Checker).GetTypeAtLocation` | 97.5 s, 4.2% | 30.5 s, 2.1% |
| `runtime.gcBgMarkWorker` (GC) | 209.3 s, 9.0% | 119.9 s, 8.2% |

The predicate proof is no longer the cost. What leads now is `entriesRecordLiteral` and its
`entriesProgramHasRecords` walk (the target of `992a6f53`), the checker's symbol resolution, and
`enumNeverValue`, reached through `predicateArguments.func2` (24.9%): the part of
`predicateArguments` outside the memoized fixed point.

## What hasn't run yet, and why

Dropped at Performance's word after push 8: the same-box rerun of `85c9d783` (stopped at file 10),
A's profile, B's timed run and the A/B decision diffs; the predicate passes A and B lack are no
longer suspects. Still to run (running now): a three-way merge of `f2479703`,
`f27967ea` and `992a6f53` (`compiler/entries-record-memo-2`), timed and profiled.

- The timeout is 3,600 s, not 900 s, since one process covers all 79 files and takes 3,230 s on main.
- Per-file invocations (one process per file) are not the meter's step: the meter loads all 79
  files as one program. Per-file seconds come from the in-process lines instead.
- The per-file named profiles ([tools/profiled_overlay.py.txt](tools/profiled_overlay.py.txt)) are
  replaced by the whole-loop profile and won't run.

No question or blocker is open.
