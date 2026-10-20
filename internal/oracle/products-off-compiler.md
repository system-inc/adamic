# Oracle products off the compiler: audit and decision

Task: #ntdgcnj. Audited origin/main `cb9c0202` on 2026-10-20.

## Decision

Plan only. Move no declarations or products in this unit. The brief permits this
outcome when most products need their builds restructured. Of 45 declarations,
42 directly reach compiler packages, including preparation before Product is
called, or call runtimeproducts, whose public API imports native. Copying those
builds to a package cannot meet the required compiler-free dependency graph.
Keeping compiler files in Inputs.Files preserves soundness but does not remove
that Go import. Removing those files to satisfy a comment-only cache experiment
would lose actual dependencies.

Three declarations can be relocated without a compiler import after extracting
compiler-free helpers: Adamic (a finder), AdamicMapOrderMutant (a Go build), and
MovesDeterminism (executes upstream Adamic). Moving the finder creates no new
producer. The Go build already uses a lean recipe. MovesDeterminism retains
compilerinputs.Lowering(), so relocating it alone does not meet proof (b).
None justifies introducing a package and consumer finders in this plan-only unit.
This is not a claim that future restructuring is impossible.

The 92 products and 1,000 s in the brief are historical measurements, not measured
results on this main. Current source still has exactly 45 declarations, but has
12 runtime feature sets and six variants including Linux parallel_tsan. The 18
runtime declarations describe 72 archive outputs across all enabled variants;
the remaining 27 declarations each address one output, including the Adamic
finder. Thus the static cross-platform inventory describes 99 output slots, not
92 simultaneous builds. Linux omits malloc, and without ADAMIC_ORACLE_WASI=1
also skips the 12 WASI outputs: 86 slots including the upstream-owned finder.
This arithmetic is a source census, not an observed hit/miss count.

## Precedent and current key model

Read the landed css change `3761e98359196b2ccb929e02e734638f86e17374`
(Loom-Original `fdc926a8886cb0cb80db1553554b2deb3433d7c4`). Its compiler-free
products build external Go/Node oracles; its Adamic lowering/native products
stay behind. Shared helpers are ordinary Go files, and consumer packages have
distinct Find declarations ordered after owners. Moving a declaration without
moving its build closure would not reproduce that separation here. The css
owner also retires old name keys with ` oracle-owner-v1`, avoiding a retained
whole-compiler read set in recordUnion, while preserving reproducible build flags
and output protocols. A future migration needs to account for that transition
as well as unique TestProduct_ declaration names.

There is also a newer change relevant to the premise: `62b5038a` introduced lean
recipes, and `4a769ad6` passes runtime compilation its jobserver as a value.
Current runtimeproducts.Build calls LeanProduct, and runtimeproducts is a reviewed
lean caller. Its recipe names runtimeproducts itself, buildcache and buildcache's
dependencies, plus the explicit native build files below, rather than oracle.test's
compiler dependency graph. GoBuild also uses a lean buildcache recipe. Both fall
back to the whole recipe when ADAMIC_BUILD_LEAN=off. Deployment of that switch
and actual Builder cache state were not measured by this source audit.

For ordinary Product calls, the traced recipe names oracle.test's own package,
including test files, and its transitive in-tree linked packages, module pins and
settings. It therefore includes native and lower even when a particular closure
does not call them. Inputs.Name is a logical product name, not the test function
name. NameKey covers name/flags/tool reports; traced read sets add recipe, reads
and upstream product output digests. Untraced local fallback keys declared Files
and is not evidence for traced recipe behavior. This distinction matters for
both the comment-only experiment and the import mutant.

## Key definitions for the inventory

Every row below names its actual build path, build-time compiler reach, key and
upstream product reads. Reach lists compiler entry packages, not merely packages
imported by unrelated assertions. load/lower also reach ir and the TypeScript Go
bridge/checker graph. Runtime rows do not execute lowering just because their
feature sets have witnesses: TestRuntimeSetsAreEachUsed lowers those separately.

- **R**: Inputs.Name `native-runtime`; Files are `internal/native/runtime`,
  `internal/native/jobserver.go`, `internal/native/library.go`,
  `internal/native/native.go`, `internal/native/target.go`. Flags are
  request.Portable (platform, runtime contents, headers, features and options);
  tool reports are request.Compiler and request.Archiver versions. Lean recipe
  as described above. Each group owns indices congruent to its number modulo 3.
- **C**: ordinary Product, Inputs.Name `count-measurements`; flags start with
  `shard N of 16`, followed by every row's key and countDigest; clang version.
  Files is empty; prepare-time values and the ordinary recipe carry those inputs.
  countDigest includes portable emitted C, input-run behavior and read trees,
  runtime source digest and counted flags. countMeasurementInputs invokes
  each.prepare BEFORE opening Product: in-process load/lower/native and the
  additional registered count adapters remain part of the recipe. The build
  closure invokes measures, which build counted native code and execute it.
- **G**: GoBuild input name `go build ./cmd/adamic adamic`; dependency files and
  module pins from go list, reproducible flags, resolved build settings, Go/C
  tool reports. The mutant additionally keys its exact before/after replacement
  of the sort in internal/lower/class.go. The Go build compiles the compiler;
  its harness need not import it. GoBuild already uses a lean recipe on this main.
- **S**: ordinary Product; Files are compilerinputs.Lowering() plus
  `internal/oracle/testdata/moves/accepted/objects.a`; toolchain is Go version,
  GOOS, GOARCH. Lowering() currently starts with the entire `internal` directory,
  module/workspace pins, bridge and TypeScript checker package directories.
  It covers native AND javascript contributions, not just lower.
- **MC**: S with name `moves-counted-control` or
  `moves-counted-shared-mutant`, native counted flags and clang version.
  Reads generated C inside the closure, rewrites the move ABI and prepends its
  counting wrapper, then calls native.Build with Count=true.
- **MN**: S with name `moves-native-VARIANT`, native flags, allocator flag and
  clang version. buildParallelProgram uses native.Build for ordinary allocators;
  slabs/malloc copy runtime sources, alter defines, call RuntimeLibraryForSource
  and BuildWithRuntime. Those native APIs execute during product construction.

## All 45 declarations

| Declaration | Source | Build calls | Compiler reach at build time | Key today | Products read |
|---|---|---|---|---|---|
| `TestProduct_CountMeasurements00` | `counts_test.go` | countMeasurementProducts(t, 0) -> countShard -> countMeasurements | load, lower, native (prepare and measure) | C, shard 0 | Adamic for interface-cast rows; counted runtime archives selected by each emitted source; identity also asks plain sanitized/release/counted archives |
| `TestProduct_CountMeasurements01` | `counts_test.go` | countMeasurementProducts(t, 1) -> countShard -> countMeasurements | load, lower, native (prepare and measure) | C, shard 1 | Adamic for interface-cast rows; counted runtime archives selected by each emitted source; identity also asks plain sanitized/release/counted archives |
| `TestProduct_CountMeasurements02` | `counts_test.go` | countMeasurementProducts(t, 2) -> countShard -> countMeasurements | load, lower, native (prepare and measure) | C, shard 2 | Adamic for interface-cast rows; counted runtime archives selected by each emitted source; identity also asks plain sanitized/release/counted archives |
| `TestProduct_CountMeasurements03` | `counts_test.go` | countMeasurementProducts(t, 3) -> countShard -> countMeasurements | load, lower, native (prepare and measure) | C, shard 3 | Adamic for interface-cast rows; counted runtime archives selected by each emitted source; identity also asks plain sanitized/release/counted archives |
| `TestProduct_CountMeasurements04` | `counts_test.go` | countMeasurementProducts(t, 4) -> countShard -> countMeasurements | load, lower, native (prepare and measure) | C, shard 4 | Adamic for interface-cast rows; counted runtime archives selected by each emitted source; identity also asks plain sanitized/release/counted archives |
| `TestProduct_CountMeasurements05` | `counts_test.go` | countMeasurementProducts(t, 5) -> countShard -> countMeasurements | load, lower, native (prepare and measure) | C, shard 5 | Adamic for interface-cast rows; counted runtime archives selected by each emitted source; identity also asks plain sanitized/release/counted archives |
| `TestProduct_CountMeasurements06` | `counts_test.go` | countMeasurementProducts(t, 6) -> countShard -> countMeasurements | load, lower, native (prepare and measure) | C, shard 6 | Adamic for interface-cast rows; counted runtime archives selected by each emitted source; identity also asks plain sanitized/release/counted archives |
| `TestProduct_CountMeasurements07` | `counts_test.go` | countMeasurementProducts(t, 7) -> countShard -> countMeasurements | load, lower, native (prepare and measure) | C, shard 7 | Adamic for interface-cast rows; counted runtime archives selected by each emitted source; identity also asks plain sanitized/release/counted archives |
| `TestProduct_CountMeasurements08` | `counts_test.go` | countMeasurementProducts(t, 8) -> countShard -> countMeasurements | load, lower, native (prepare and measure) | C, shard 8 | Adamic for interface-cast rows; counted runtime archives selected by each emitted source; identity also asks plain sanitized/release/counted archives |
| `TestProduct_CountMeasurements09` | `counts_test.go` | countMeasurementProducts(t, 9) -> countShard -> countMeasurements | load, lower, native (prepare and measure) | C, shard 9 | Adamic for interface-cast rows; counted runtime archives selected by each emitted source; identity also asks plain sanitized/release/counted archives |
| `TestProduct_CountMeasurements10` | `counts_test.go` | countMeasurementProducts(t, 10) -> countShard -> countMeasurements | load, lower, native (prepare and measure) | C, shard 10 | Adamic for interface-cast rows; counted runtime archives selected by each emitted source; identity also asks plain sanitized/release/counted archives |
| `TestProduct_CountMeasurements11` | `counts_test.go` | countMeasurementProducts(t, 11) -> countShard -> countMeasurements | load, lower, native (prepare and measure) | C, shard 11 | Adamic for interface-cast rows; counted runtime archives selected by each emitted source; identity also asks plain sanitized/release/counted archives |
| `TestProduct_CountMeasurements12` | `counts_test.go` | countMeasurementProducts(t, 12) -> countShard -> countMeasurements | load, lower, native (prepare and measure) | C, shard 12 | Adamic for interface-cast rows; counted runtime archives selected by each emitted source; identity also asks plain sanitized/release/counted archives |
| `TestProduct_CountMeasurements13` | `counts_test.go` | countMeasurementProducts(t, 13) -> countShard -> countMeasurements | load, lower, native (prepare and measure) | C, shard 13 | Adamic for interface-cast rows; counted runtime archives selected by each emitted source; identity also asks plain sanitized/release/counted archives |
| `TestProduct_CountMeasurements14` | `counts_test.go` | countMeasurementProducts(t, 14) -> countShard -> countMeasurements | load, lower, native (prepare and measure) | C, shard 14 | Adamic for interface-cast rows; counted runtime archives selected by each emitted source; identity also asks plain sanitized/release/counted archives |
| `TestProduct_CountMeasurements15` | `counts_test.go` | countMeasurementProducts(t, 15) -> countShard -> countMeasurements | load, lower, native (prepare and measure) | C, shard 15 | Adamic for interface-cast rows; counted runtime archives selected by each emitted source; identity also asks plain sanitized/release/counted archives |
| `TestProduct_AdamicMapOrderMutant` | `determinism_test.go` | adamicMapOrderMutant -> buildcache.GoBuildInMutated | No compiler execution in harness; go build compiles load/lower/native/javascript | G + determinismMutation | None |
| `TestProduct_Adamic` | `interface_cast_counts_test.go` | buildcache.AdamicBuiltElsewhere | None in finder; upstream go build compiles load/lower/native/javascript | G, unmutated, find-only under trace | Adamic, owned by internal/native TestProduct_WASICompiler |
| `TestProduct_MovesNative_malloc` | `moves_darwin_test.go` | moveBinary(t, "malloc") -> buildParallelProgram | native.RuntimeLibraryForSource + native.BuildWithRuntime; source fallback also load/lower/native/javascript | MN, malloc | MovesSource program.c; runtime sources compiled directly with allocator override |
| `TestProduct_MovesCounted_Control` | `moves_test.go` | moveCountedBinary(t, false) -> native.Build | native; load/lower/native/javascript if MovesSource fallback builds | MC, control | MovesSource program.c; counted runtime archive for generated source |
| `TestProduct_MovesCounted_SharingMutant` | `moves_test.go` | moveCountedBinary(t, true) -> native.Build | native; load/lower/native/javascript if MovesSource fallback builds | MC, shared-mutant | MovesSource program.c; counted runtime archive for generated source |
| `TestProduct_MovesSource` | `moves_test.go` | moveSource -> lowered -> native.C + javascript.JavaScript | load, lower, native, javascript | S, moves-source | None |
| `TestProduct_MovesNative_asan` | `moves_test.go` | moveBinary(t, "asan") -> buildParallelProgram | native.Build; source fallback also load/lower/native/javascript | MN, asan | MovesSource program.c; sanitized runtime archive for its emitted features |
| `TestProduct_MovesNative_release` | `moves_test.go` | moveBinary(t, "release") -> buildParallelProgram | native.Build; source fallback also load/lower/native/javascript | MN, release | MovesSource program.c; release runtime archive for its emitted features |
| `TestProduct_MovesNative_asan_slabs` | `moves_test.go` | moveBinary(t, "asan_slabs") -> buildParallelProgram | native.RuntimeLibraryForSource + native.BuildWithRuntime; source fallback also load/lower/native/javascript | MN, asan_slabs | MovesSource program.c; runtime sources compiled directly with allocator override |
| `TestProduct_MovesNative_tsan` | `moves_test.go` | moveBinary(t, "tsan") -> buildParallelProgram | native.Build; source fallback also load/lower/native/javascript | MN, tsan | MovesSource program.c; parallel_tsan runtime archive for its emitted features |
| `TestProduct_MovesDeterminism` | `moves_test.go` | moveDeterminism -> AdamicBuiltElsewhere -> emitWithin twice | load/lower/native executed in upstream Adamic process | S, moves-determinism | Adamic |
| `TestProduct_ParallelTSanRuntime0` | `parallel_test.go` | runtimeproducts.Build(t, "parallel_tsan", 0) -> native.RuntimeRequestFor -> request.Build | native build API; no load/lower/emitter execution | R, parallel_tsan, sets 0,3,6,9 | None (compiles runtime sources directly) |
| `TestProduct_ParallelTSanRuntime1` | `parallel_test.go` | runtimeproducts.Build(t, "parallel_tsan", 1) -> native.RuntimeRequestFor -> request.Build | native build API; no load/lower/emitter execution | R, parallel_tsan, sets 1,4,7,10 | None (compiles runtime sources directly) |
| `TestProduct_ParallelTSanRuntime2` | `parallel_test.go` | runtimeproducts.Build(t, "parallel_tsan", 2) -> native.RuntimeRequestFor -> request.Build | native build API; no load/lower/emitter execution | R, parallel_tsan, sets 2,5,8,11 | None (compiles runtime sources directly) |
| `TestProduct_NativeRuntimesSanitized0` | `runtimes_test.go` | runtimeproducts.Build(t, "sanitized", 0) -> native.RuntimeRequestFor -> request.Build | native build API; no load/lower/emitter execution | R, sanitized, sets 0,3,6,9 | None (compiles runtime sources directly) |
| `TestProduct_NativeRuntimesSanitized1` | `runtimes_test.go` | runtimeproducts.Build(t, "sanitized", 1) -> native.RuntimeRequestFor -> request.Build | native build API; no load/lower/emitter execution | R, sanitized, sets 1,4,7,10 | None (compiles runtime sources directly) |
| `TestProduct_NativeRuntimesSanitized2` | `runtimes_test.go` | runtimeproducts.Build(t, "sanitized", 2) -> native.RuntimeRequestFor -> request.Build | native build API; no load/lower/emitter execution | R, sanitized, sets 2,5,8,11 | None (compiles runtime sources directly) |
| `TestProduct_NativeRuntimesRelease0` | `runtimes_test.go` | runtimeproducts.Build(t, "release", 0) -> native.RuntimeRequestFor -> request.Build | native build API; no load/lower/emitter execution | R, release, sets 0,3,6,9 | None (compiles runtime sources directly) |
| `TestProduct_NativeRuntimesRelease1` | `runtimes_test.go` | runtimeproducts.Build(t, "release", 1) -> native.RuntimeRequestFor -> request.Build | native build API; no load/lower/emitter execution | R, release, sets 1,4,7,10 | None (compiles runtime sources directly) |
| `TestProduct_NativeRuntimesRelease2` | `runtimes_test.go` | runtimeproducts.Build(t, "release", 2) -> native.RuntimeRequestFor -> request.Build | native build API; no load/lower/emitter execution | R, release, sets 2,5,8,11 | None (compiles runtime sources directly) |
| `TestProduct_NativeRuntimesCounted0` | `runtimes_test.go` | runtimeproducts.Build(t, "counted", 0) -> native.RuntimeRequestFor -> request.Build | native build API; no load/lower/emitter execution | R, counted, sets 0,3,6,9 | None (compiles runtime sources directly) |
| `TestProduct_NativeRuntimesCounted1` | `runtimes_test.go` | runtimeproducts.Build(t, "counted", 1) -> native.RuntimeRequestFor -> request.Build | native build API; no load/lower/emitter execution | R, counted, sets 1,4,7,10 | None (compiles runtime sources directly) |
| `TestProduct_NativeRuntimesCounted2` | `runtimes_test.go` | runtimeproducts.Build(t, "counted", 2) -> native.RuntimeRequestFor -> request.Build | native build API; no load/lower/emitter execution | R, counted, sets 2,5,8,11 | None (compiles runtime sources directly) |
| `TestProduct_NativeRuntimesWASI0` | `runtimes_test.go` | runtimeproducts.Build(t, "wasm", 0) -> native.RuntimeRequestFor -> request.Build | native build API; no load/lower/emitter execution | R, wasm, sets 0,3,6,9 | None (compiles runtime sources directly) |
| `TestProduct_NativeRuntimesWASI1` | `runtimes_test.go` | runtimeproducts.Build(t, "wasm", 1) -> native.RuntimeRequestFor -> request.Build | native build API; no load/lower/emitter execution | R, wasm, sets 1,4,7,10 | None (compiles runtime sources directly) |
| `TestProduct_NativeRuntimesWASI2` | `runtimes_test.go` | runtimeproducts.Build(t, "wasm", 2) -> native.RuntimeRequestFor -> request.Build | native build API; no load/lower/emitter execution | R, wasm, sets 2,5,8,11 | None (compiles runtime sources directly) |
| `TestProduct_NativeRuntimesShipped0` | `runtimes_test.go` | runtimeproducts.Build(t, "shipped", 0) -> native.RuntimeRequestFor -> request.Build | native build API; no load/lower/emitter execution | R, shipped, sets 0,3,6,9 | None (compiles runtime sources directly) |
| `TestProduct_NativeRuntimesShipped1` | `runtimes_test.go` | runtimeproducts.Build(t, "shipped", 1) -> native.RuntimeRequestFor -> request.Build | native build API; no load/lower/emitter execution | R, shipped, sets 1,4,7,10 | None (compiles runtime sources directly) |
| `TestProduct_NativeRuntimesShipped2` | `runtimes_test.go` | runtimeproducts.Build(t, "shipped", 2) -> native.RuntimeRequestFor -> request.Build | native build API; no load/lower/emitter execution | R, shipped, sets 2,5,8,11 | None (compiles runtime sources directly) |

## Dependency boundaries required before a future move

Runtime archives (18 declarations) need a compiler-free runtime builder API, or
an upstream product that performs native.RuntimeRequest.Build and supplies a
complete archive/request to a compiler-free consumer. That API and its census
live in internal/native/runtimeproducts, outside this unit's allowed territory.
They already have explicit native build dependencies and lean recipes; a package
move alone offers no additional default lean-key benefit. Do not discard the
native build files merely because a lower edit does not affect archives.

Count measurements (16 declarations) need upstream portable emitted C AND a
complete run specification for each registered adapter, input tree, stack and
identity, followed by a compiler-free counted builder/runner. Moving only the
rows.tsv write would omit prepare-time compiler reads. Output-digest chaining
could cut off compiler changes that leave these inputs equal, but only after
all compiler contributions to native flags/runtime/linking are carried as well.

MovesSource needs its lowering and both emitters upstream. MovesNative and both
MovesCounted variants need the native step upstream too, or a separate build API
that imports no native/lower and keys every file implementing the native step.
Reading MovesSource's output alone is insufficient: changing native.Build,
native flags or runtime/linking can change the executable with identical C.
Keeping Lowering() is conservative and sound, but still rekeys on lower comments.
MovesDeterminism could instead key on Adamic's output and the source fixture,
provided the upstream binary and all emitted-input reads are preserved. Its
current Lowering() would have to be replaced with a proven complete boundary.

For any later owner move, retain an oracle consumer finder with a unique name,
use BuiltElsewhere rather than a second producer, and update every corresponding
product-order edge. Existing external consumers in cmd/adamic and stage3/fixtures
must still find the runtime outputs. No order edge changes are needed for this
plan. No symlinked Go files are introduced; counts.md is untouched.

## Validation scope

No product recipe, key definition, executable code, dependency or check changes.
Therefore
proofs (b), (c) and (d) for moved products are not applicable: there are no moved
products and no new package to mutate. No rebuild count or killed import mutant
is claimed. The required soundness condition is the reason for stopping at this
plan, not a claim that a package-only move passed the experiments.

Command results and setup timing are recorded in the task report. Test output
is kept in separate /tmp/ntdgcnj-* logs. Source inventory is reproducible with
`rg -n '^func TestProduct_' internal/oracle -g '*.go'` (45 rows).

## Observed checks on this branch

`git fetch origin` succeeded, then the task branch was created directly from
origin/main. The short-hash fetch of the original css commit was unavailable;
the full landed commit above, carrying its Loom-Original hash, was read instead.

`export GOPROXY='https://proxy.golang.org|direct'` and
`timeout 840 bash cloud/setup.sh > /tmp/ntdgcnj-setup.log 2>&1` passed.
The generated environment was `/workspace/adamic-tools/env.sh`, sourced for
validation. Setup's cumulative timing lines were:

```text
go ready (0.032s)
clang ready (0.248s)
node ready (0.021s)
markdown dependencies ready (1.119s)
submodules ready (228.646s)
shared cache off (228.647s)
go build ready (590.283s)
test binaries deferred (590.762s)
build cache warm (590.768s)
done on 5 processors (590.799s)
```

`nproc` returned 5; cgroup cpu.max was `400000 100000`. Tools were Go 1.27.1,
clang 20.1.8 and Node 24.19.0. Build warming overlapped validation compilation,
so these timings do not measure an isolated setup or any product speedup.

- `timeout 60 gofmt -l cmd internal`: exit 0, no output,
  `/tmp/ntdgcnj-gofmt.log`.
- `timeout 840 go vet ./internal/oracle/... ./cmd/adamic-gate/...`: exit 0,
  `/tmp/ntdgcnj-vet.log`. An earlier attempt while the nested submodule was still
  cloning could not open cohere/TypeScript/tsc/go.mod; the completed-checkout
  rerun passed.
- `timeout 840 go vet ./...`: exit 0, no output,
  `/tmp/ntdgcnj-vet-all.log`.
- `timeout 120 go list -deps -test ./internal/oracle ./internal/native/runtimeproducts`:
  exit 0, `/tmp/ntdgcnj-deps.log`; includes native and lower.
- `timeout 120 go list -deps -test ./internal/native/runtimeproducts`: exit 0,
  `/tmp/ntdgcnj-runtime-deps.log`; includes native. Therefore extracting just the
  archive declarations, while continuing to call runtimeproducts.Build, does
  not meet the required dependency-graph condition.
- Compared all source TestProduct_ names with inventory table rows: 45 names,
  45 unique rows, no missing or extra names. `git diff --check` passed.

The requested package test command passed:

```sh
timeout 1800 go test -count=1 -timeout 30m ./internal/oracle/... ./cmd/adamic-gate/... > /tmp/ntdgcnj-tests.log 2>&1
```

```text
ok github.com/system-inc/adamic/internal/oracle 609.231s
ok github.com/system-inc/adamic/cmd/adamic-gate 11.335s
```

Exit 0. These are ordinary worker tests, with the normal oracle product/result
cache settings; this is not an uncached integration run. Proof (a) is green.
Proofs (b), (c), (d) remain not applicable for the plan-only outcome, not passes.
