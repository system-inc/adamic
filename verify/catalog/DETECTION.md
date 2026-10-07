# Detection time

How quickly does the randomized test-program generator (`cmd/adamic-fuzz`, `internal/fuzz`) report a
disagreement once a catalog entry's fault is re-applied? Every number below is in
[detection.json](detection.json), with each run's command, load before and after, the first finding's
verdict line and detail, and the shrunk program.

## Method

- Target: origin/main `f8013f0b`. Main moved from `e8ba3d5d` to `f8013f0b` (the fixes-3 landing)
  while the first runs were being set up, so every run was restarted there. Each applying entry got a
  clean worktree at `f8013f0b` with its patch applied. Entries 08 and 12 no longer apply on `f8013f0b`
  (`git apply --check` fails in `internal/lower/object.go` and `internal/lower/refusals.go`): they need
  a refresh. As a supplement they were run on `e8ba3d5d`, where they apply, against that commit's own
  control. Entries 10 and 13 to 16 are skipped by the catalog. check.py reads `/proc` and `nproc` and
  does not run on macOS, so applicability is its own `git apply --check`, run by hand. Main kept
  moving after the runs (`39638d9e` when this was written); applicability was not rechecked there.
- Run: `adamic-fuzz -root <patched worktree> -work <own directory> -seed 1 -count 2000 -parallel 4
  -shrink=false -v` with `GOFLAGS=-trimpath`, as one process, stopped 3 seconds after the first
  finding line, at 2,000 seeds, or at 15 minutes, whichever came first.
- Seeds: the first finding printed. With four workers a lower seed can still be running, so the
  table also gives the lowest finding seed among the seeds judged before the run stopped; every seed
  below that one was judged and agreed.
- Seconds: wall seconds from launching adamic-fuzz to the finding's verdict line. That includes
  building the checkout's adamic command and sanitized runtime: about 26 s cold (entry 01, the first
  i16 run) and 2 to 5 s once Go's build cache was warm.
- Shrink: `adamic-fuzz -root <patched worktree> -seed <reported seed> -count 1` with shrinking on,
  using the generator that found it. Each shrunk program was then tried both ways: it still fails
  the same way on the patched tree, and it agrees on unpatched `f8013f0b`.
- Leak pass: this Mac's ASan has no leak detection, so the leak pass reported every program as
  leaking. Each generator's `internal/fuzz/run.go` was patched in its scratch worktree only
  (uncommitted) to ignore a leak run whose stderr says `detect_leaks is not supported on this
  platform`. Heap-use-after-free, overflows and UBSan still report on macOS. Leaks are not covered
  here.

Generators:

- area: origin/area/developer-tools `ee5c63c5`, the fixed judge. Its `generate.go` is the same
  as `89e586e3`, the tip when this study began.
- i16: origin/devtools/generators-i16 `7ebba450`. Same judge. It adds the shared-slices, ownership
  and overrides features. Its scenes are written after the catalog's shapes (narrowed lends,
  borrowed elements around virtual and super calls, spreads that call methods, capturing
  constructors, slices sharing their owner's bytes, overrides behind static signatures), so its hits
  on those rows come from shapes made for those families.
- latest: origin/area/developer-tools `ac92bf0a`, the tip once these runs were done. It has merged the
  i16 families and adds undefined-numbers (`number | undefined` across every kind of call). It was
  run only on the entries no earlier generator found.
- scenes: origin/devtools/generator-liveness-fields `6615e48d`, the area tip `3f904d53` plus the
  liveness and field-representation scenes (`internal/fuzz/liveness.go`, `field_representation.go`).
  It was run only on 02, 08 and 09, against origin/main `39638d9e` rather than `f8013f0b`, the same
  way as the rest (`-seed 1 -count 2000 -parallel 4 -shrink=false -v`, stopped 3 seconds after the
  first finding), with the same uncommitted leak-pass patch. 02 and 09 apply to `39638d9e`
  unchanged. 08 does not: inside the hunk the patch removes, main now calls
  `l.acceptsUndefined(node)` where the patch has `comparedWithUndefined(node)`, so the patch's copy
  of that one removed line was changed to match, and nothing else
  ([08-narrowed-number-field-39638d9e.patch](08-narrowed-number-field-39638d9e.patch); catalog.json
  still names the original). Go's
  build cache was warm, so its seconds are mostly the programs.

## Table

Target `f8013f0b` unless the row says otherwise. "Not found" means no finding within the cap.

| Entry | Area | area generator | i16 generator | latest generator | scenes generator (`39638d9e`) | Matches recorded shape |
|---|---|---|---|---|---|---|
| 01 shared-slice-append | ownership | not found, 2,000 seeds, 441 s | seed 3 (lowest 1), 26.6 s | not run (found) | not run | yes: ASan heap-buffer-overflow in `adamic_string_put` |
| 02 liveness-throw | liveness | not found, 2,000 seeds, 515 s | not found, 2,000 seeds, 802 s | not found, cap at 1,719 seeds, 900 s | seed 3 (lowest 1), 4.6 s | yes: UBSan member access within null pointer of type `const adamic_object`. Seed 1 is the same fault as stdout differs: a reuse in place wrote into the local the catch reads |
| 03 defined-lent | ownership | not found, 2,000 seeds, 528 s | seed 1 (lowest 1), 5.2 s | not run (found) | not run | yes: ASan heap-use-after-free |
| 04 borrowed-array-move | ownership | not found, 2,000 seeds, 367 s | seed 4 (lowest 4), 7.2 s | not run (found) | not run | yes: ASan heap-use-after-free |
| 05 spread-method-reuse | ownership | not found, 2,000 seeds, 398 s | seed 6 (lowest 5), 4.1 s | not run (found) | not run | no: UBSan member access within null pointer, where the catalog records stdout differs with exit 0. The shrunk program is the same spread-with-a-method shape |
| 06 constructor-capture-region | ownership | not found, 2,000 seeds, 415 s | seed 7 (lowest 7), 5.7 s | not run (found) | not run | yes: ASan heap-use-after-free |
| 07 borrowed-element-reads | ownership | not found, 2,000 seeds, 552 s | seed 10 (lowest 10), 9.1 s | not run (found) | not run | yes: ASan heap-use-after-free, from the overrides scene |
| 08 narrowed-number-field | representation | needs refresh on `f8013f0b`. On `e8ba3d5d`: not found, 2,000 seeds, 568 s | `e8ba3d5d`: not found, 2,000 seeds, 885 s | `e8ba3d5d`: not found, 2,000 seeds, 891 s | patch adapted: seed 3 (lowest 1), 4.2 s | yes: native prints NaN where Node prints undefined. The scene reads only where undefined is taken, so main never stops at an inserted check; the JavaScript backend also differs, at `?? -5` |
| 09 literal-undefined-field | representation | not found, 2,000 seeds, 521 s | not found, 2,000 seeds, 802 s | not found, cap at 1,702 seeds, 900 s | seed 3 (lowest 1), 3.8 s | yes: stdout differs, both exit 0, native prints `0 false` for `undefined true` |
| 11 refuse-definite-assignment | refusal | not found, 2,000 seeds, 351 s | not found, 2,000 seeds, 894 s | not found, 2,000 seeds, 880 s | refusal writer: program 1, 0.056407 s (best of 3); clean 200-program targeted controls; [report](../../internal/refusalprobe/REPORT.md) | yes: the removed refusal compiles |
| 12 refuse-suppression-directives | refusal | needs refresh on `f8013f0b`. On `e8ba3d5d`: not found, 2,000 seeds, 568 s | `e8ba3d5d`: not found, 2,000 seeds, 890 s | not run | refusal writer: program 1, 0.049749 s (best of 3); clean 200-program targeted controls; [report](../../internal/refusalprobe/REPORT.md) | yes: the removed refusal compiles |

On `e8ba3d5d` the patched 08 and 12 runs reported findings, but every one is a seed the same
generator's unpatched `e8ba3d5d` control also reports, all of them the #ht2nwj5 panic (below). No run
found anything beyond its control, so those rows are not found.

The area generator found none of the eleven. The i16 shapes find six of them within ten seeds:
shared-slices reaches 01, the ownership scenes reach 03, 04, 05 and 06, and the overrides scene
reaches 07. Before the scenes generator, no generator saw liveness (02) or representation (08, 09,
even with undefined-numbers). The scenes generator finds all three at the first seeds it judges:
every one of seeds 1 to 8 found 02, seven of them 08, and seven of the seven judged 09. The liveness
scene takes a local's old value with a spread, a consuming call, an array spread or a map, assigns
the local from a call that may throw, and reads it where the throw lands; the field-representation
scene clears narrowed `number | undefined` fields behind the checker's back and widens literals whose
fields are written undefined. No generator sees refusal (11, 12). A generator that writes only programs the checker accepts can see a missing
refusal only if the program it should have refused also runs differently, and none of the three
writes a definite assignment assertion or a suppression directive (no `!:` or `@ts-` in
`internal/fuzz`).

## Controls

The same seed range on unpatched trees:

| Control | Programs | Findings |
|---|---:|---|
| area generator on `f8013f0b` | 2,000 | none (100 inserted checks, 8 checker refusals, 4 not yet, 2 unfit) |
| i16 generator on `f8013f0b` | 2,000 | none (112 inserted checks) |
| latest generator on `f8013f0b` | 1,713 (cap) | none |
| area generator on `e8ba3d5d` | 2,000 | 18, all #ht2nwj5 |
| i16 generator on `e8ba3d5d` | 2,000 | 23, all #ht2nwj5 |
| latest generator on `e8ba3d5d` | 2,000 | 19, all #ht2nwj5 |
| scenes generator on `39638d9e` | 1,000 | none (935 agreed, 64 inserted checks, 1 unfit, 0 not yet, 0 checker refusals) |

The scenes control covers seeds 1 to 1,000, past every seed the three patched runs judged. Its 64
inserted checks and its one unfit program (seed 428) come to the same verdicts with both scenes left
out, so none of them is the scenes' doing, and `adamic c` from `39638d9e` compiles every one of the
1,000 programs with no checker refusal and no not yet.

#ht2nwj5 is the optional-chain-after-call panic: native and the JavaScript backend both exit 70 with
`adamic: panic: TypeError: Cannot read properties of undefined (reading 'next')` where Node exits 0.
None of it reproduces on `f8013f0b`: seed 79 tried there agrees, and both full controls there are
clean over the same seeds.

## Shrunk programs from the scenes generator

Seed 3, the first finding of each run, reduced with `go run ./cmd/adamic-reduce -root <patched tree>`.
Each fails the same way on its patched tree and agrees on unpatched `39638d9e` (detection.json has the
sources). For 02 native's stdout is lost when UBSan kills it, so the derived signature is the first
line that differs, which the reduction kept from the undefined-numbers scene; without those lines the
rest still fails the same way.

- 02 (812 lines to 30): a loop whose try passes the local to `liveConsume`, which spreads it, then
  assigns it from `liveEach`, whose forEach callback throws; the catch is empty, and the next round
  consumes the local again. Native passes the local the first round moved away: UBSan member access
  within null pointer.
- 08 (812 lines to 13): `field1.value !== undefined`, then `field1Alias.value = undefined`, then
  `field1.value` in a template prints `NaN false number NaN` where Node prints
  `undefined true undefined -5`.
- 09 (812 lines to 9): `const field4 = { value: undefined }` widened to
  `FieldReading { readonly value: number | undefined }` prints `0 false number 0` where Node prints
  `undefined true undefined -5`.

## Build flags

Commits: main `f8013f0baac41ddc340d76f83bddde38536a8f07` (08 and 12 supplement:
`e8ba3d5d81de4d3773c723914fccd4c76248b965`); generators `ee5c63c5205cd113366ceb268177924e428ce126`,
`7ebba45064bfe1b6738b817de5719300a9defa0b`, `ac92bf0abf36b0fa442929b4e9a434445cb15a17`.
`sysctl -n hw.ncpu` 16; `go version go1.27.0 darwin/arm64`; `Apple clang version 21.0.0
(clang-2100.3.34.2)`; node `v24.14.1`; `go build -trimpath`, and `GOFLAGS=-trimpath` for the
checkout builds adamic-fuzz makes; native flags are the generator tree's `native.Flags` for a
sanitized build (ASan and UBSan). Load (`vm.loadavg`) before the first run `38.99 71.76 71.02`, after
the last `180.78 174.89 173.45`. Every run's own before and after is in detection.json. The scenes
runs: main `39638d9e278d38bb5aeae887f46d55a70e47aaad`, generator
`6615e48d9d23160b43cfbbb9242c1b0b486a1a89`, load `190.80 129.46 111.63` before the first and
`120.73 120.88 116.36` after the control.

The machine carried other agents' work the whole time. The area control and the area runs of 01 to
05 and 11 had the machine to themselves among this study's runs; every other run overlapped one to
six others, named in its `overlapped_runs`. That is why the i16 runs of 02, 09 and 11 and the latest
runs took 13 to 15 minutes where the first area runs finished 2,000 seeds in six to nine. Seeds are
the stable number; seconds are this machine under this load.
