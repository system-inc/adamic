Built the 81-file tsc entry graph and measured fifteen sequential native build stops in both split modes.
First evidence pushed in 1f23d5cc and 0a27bcd2 on codex/stage3-tsc-entry-build; compiler base c6761c24.
All 30 builds exit 1 during checking; fifteen minimal witnesses run on Node 24.19.0.
Five mutants caught: dropped graph edge, probe diagnostic, Node byte, split byte, omitted stop row.
No native binary, lowering measurement or tiny diagnostics comparison; stops 9 and 14 are placeholder-induced.

## First stop

Both `ADAMIC_NATIVE_SPLIT=0` and `1 ADAMIC_NATIVE_JOBS=5` run main's unchanged
compiler on `/tmp/tsc-entry-adapted/src/tsc/tsc.ts`. Both exit 1; their complete
stderr streams are identical, 640 lines each. The first returned diagnostic is:

```text
src/compiler/builder.ts:1246:69: error TS2345: Argument of type 'Path | undefined' is not assignable to parameter of type 'string'.
  Type 'undefined' is not assignable to type 'string'.
```

This file is inside `src/compiler`. It is a checker stop, before lowering,
not a NotYet or Refused finding. `probes/01-indexed-path.a` runs successfully
on Node, printing `undefined`, and main refuses its native build with the
same TS2345 message and elaboration. Giving the indexed value a defined
empty-string fallback removes this checker stop: `adamic types` exits 0.
The control native build instead reaches NotYet for an array of Path.
This control proves checker sensitivity, not successful native compilation.

## Module graph

The complete source closure is 81 files, independently confirmed with stock
TypeScript's program loader. Every reached file outside `src/compiler` is:

- `src/tsc/tsc.ts`
- `src/tsc/_namespaces/ts.ts`

`executeCommandLine.ts` is inside `src/compiler` in the pinned 6.0.3 tree.
The graph includes type-only edges. Host loads are external modules, recorded
in `evidence/closure.json`, not additional TypeScript source files.

## Setup and provenance

`bash cloud/setup.sh` succeeded with the requested GOPROXY. Timing lines:
Node ready 0.106s, Go ready 0.273s, clang ready 1.077s, markdown dependencies
ready 1.873s, submodules ready 212.269s, Go build ready 511.191s, cache warm
511.301s, done 511.381s. `nproc` is 5; cgroup quota is 4 CPUs.
The environment is `/workspace/adamic-tools/env.sh`, Node v24.19.0,
Go go1.27.1, clang 20.1.8. Complete setup output is in `evidence/setup.log`.
An early compiler build raced submodule setup and failed on the missing
`cohere/TypeScript/tsc/go.mod`; retry after setup's submodule step succeeded.

`bash stage3/apply.sh /tmp/tsc-entry-adapted` succeeded. The generated table
is at the tree root `patch-set.md`, not `stage3/patch-set.md`; a copy is retained.
It records 21 adaptations, 78 changed files, 5,128 lines added, 5,102 removed.
Source SHA256s and compiler/upstream commits are in `evidence/provenance.json`.
No compiler or adaptation source was edited.

## Node and mutants

The scanner-style stock-TypeScript loader runs the actual tsc entry point.
The initial ESM-only attempt failed because `sys` was undefined. Providing
upstream's CommonJS host globals makes `--version` print `Version 6.0.3`,
exit 0. Both attempts' streams are retained; no source adaptation was applied.

A scratch graph mutant omits traversal of the entry's import. Graph generation
still exits 0 with one file; `check-graph.cjs` exits 1 against TypeScript's
81-file program loader. The unmutated graph passes. This proves the graph
comparison can detect a missing dependency.

## Meter and limits

The newest main meter is `20261008T000754Z.hFDR7Q`, compiler f4efdd23.
All fifteen stop messages have **no exact match** in that run's main
src/compiler latent lowering `per_reason` table. There are **15 unmatched stop
sites and 12 distinct unmatched messages**. They are checker diagnostics,
so there are **zero newly measured lowering reasons**. The ordinary meter
reported zero source files passing lowering; the entry experiment uses the
ordinary compiler gate and no latent overlay. `evidence/meter-comparison.json`
records the matching rule, run and counts.

## Fifteen-stop continuation

The disposable source copy is `/tmp/tsc-entry-progress-src`. It contains fourteen
body replacements, each exactly `{ throw new Error("tsc-entry scratch placeholder"); }`.
It was never staged or committed. The record for each replacement retains the
original body, offsets and replacement; no modified adapted source is committed.
`continue.py` stops after the fifteenth observation, as requested.
Every attempt has complete stdout, stderr and exit evidence for both split modes.
Both splits exit 1 and agree byte for byte at every stop.

The compiler sorts checker diagnostics lexicographically by formatted path and
location. This is the first returned diagnostic order, not source numeric order
or a call traversal order. Removing a body can expose a caller's inference error.
Only stops 9 and 14 have message text absent from the pristine diagnostic stream;
these are marked placeholder-induced rather than original-source blockers.

The following coordinates map back to the original adapted source. Actual scratch
coordinates and exact full messages, including the two large object types, are
in `evidence/stops.json`. Every row is inside `src/compiler`; every meter message
match is **No**.

| Stop | Original file:line:column | Stopping message | Minimal witness |
|---|---|---|---|
| 1 | builder.ts:1246:69 | TS2345: Path or undefined passed as string | [01](probes/01-indexed-path.a) |
| 2 | builder.ts:1258:65 | TS2488: optional [Path, FileInfo] tuple destructured | [02](probes/02-tuple-parameter.a) |
| 3 | builder.ts:2273:9 | TS2375: present-undefined outSignature field | [03](probes/03-present-undefined-field.a) |
| 4 | builder.ts:395:118 | TS2345: Path or undefined passed as Path | [04](probes/04-optional-path-argument.a) |
| 5 | builder.ts:991:17 | TS2345: Path or undefined passed as Path | [05](probes/05-optional-path-argument.a) |
| 6 | checker.ts:10022:63 | TS2345: Symbol or undefined passed as Symbol | [06](probes/06-optional-symbol-argument.a) |
| 7 | checker.ts:10026:97 | TS2345: Symbol or undefined passed as Symbol | [07](probes/07-optional-symbol-argument.a) |
| 8 | checker.ts:10033:67 | TS18048: m is possibly undefined | [08](probes/08-optional-member-read.a) |
| 9 | checker.ts:10034:53 | TS2345: "real" passed as never (placeholder-induced) | [09](probes/09-never-argument.a) |
| 10 | checker.ts:10319:83 | TS2345: Symbol or undefined passed as Symbol | [10](probes/10-optional-symbol-argument.a) |
| 11 | checker.ts:10320:51 | TS2345: array with undefined passed as readonly Symbol array | [11](probes/11-optional-symbol-array.a) |
| 12 | checker.ts:10941:33 | TS2379: present-undefined modifiers argument | [12](probes/12-present-undefined-argument.a) |
| 13 | checker.ts:12034:30 | TS18048: file is possibly undefined | [13](probes/13-optional-file-read.a) |
| 14 | checker.ts:12859:13 | TS2322: void or Type assigned as Type or undefined (placeholder-induced) | [14](probes/14-void-result.a) |
| 15 | checker.ts:13986:47 | TS2345: Declaration or undefined passed as Node | [15](probes/15-optional-declaration-argument.a) |

All fifteen witnesses exit 0 on Node. Each native witness build exits 1 at the
same checker code. Thirteen reproduce the full first-line message text exactly.
The TS2375 and TS2379 witnesses remove unrelated fields and retain respectively
the `outSignature` and `modifiers` present-undefined incompatibility. Their smaller
object type changes the displayed message text; this is recorded explicitly.
Witness 02 uses a branded Path assertion only to supply a valid runtime tuple;
its destructured parameter still triggers the target checker stop before lowering.

No build completed, so the conditional `--tiny` native diagnostics run could not
be performed. The fifteen placeholders are exploratory and do not create a
working or semantically equivalent tsc.

## Final validation and mutants

Complete output went to logs, never through a test-output pipe:

```text
python3 stage3/drivers/tsc-entry/verify.py
15 stops verified; 15 Node witnesses exit 0; both split modes agree
python3 stage3/drivers/tsc-entry/mutants.py
probe-diagnostic: caught, exit 1; 1: probe diagnostic
node-byte: caught, exit 1; 1: Node observation
split-byte: caught, exit 1; 1: split byte comparison
stop-population: caught, exit 1; stop population
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/undefined_references.a$' -count=1 -timeout 30m
ok github.com/system-inc/adamic/internal/oracle 10.547s
```

The diagnostic mutant substitutes the defined-value control's genuine NotYet
output for the first witness's TS2345 stream; exits and Node observation stay
unchanged, so only diagnostic matching catches it. The Node mutant changes its
stdout; the split mutant adds one line only to split 1; the population mutant
omits the fifteenth observation. Each scratch mutation is independent and its
dedicated comparison rejects it. These four plus the dropped graph-edge mutant
are evidence/harness mutants, not compiler mutations or native tsc execution.
The changed indexed-input control passes `adamic types`, while the original
input is caught by TS2345; its later native NotYet is separately retained.

Shell syntax, JavaScript parsing, Python parsing and `git diff --check` pass.
No Go implementation packages were touched. The full repository gate was not
run; setup built all packages and the filtered existing oracle passed.
No performance, C emission, native module initialization, ownership, full
upstream suite, native CLI behavior or source stops after fifteen are covered.
