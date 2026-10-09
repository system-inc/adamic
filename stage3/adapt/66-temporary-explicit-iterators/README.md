# Temporary explicit iterators

Scratch fallback for the step 32 generator stop. This branch is not a landing
instruction: only @system_adamic decides whether this races native generators
and lands. Remove this directory when native generators reach the entry build.
No PR is opened.

Based on `origin/area/stage3` at
`fa5e90939a9e4c40aaf15cecd5d5b554c10dd2c9`, with the pinned TypeScript 6.0.3 tree.
Only this adaptation directory is changed. No ordering hook is needed:
`stage3/apply.py` already discovers directories in numeric order, placing 66
after temporary Node builtins and before readonly views. The shared patch table,
compiler, loader, native emitter, oracle, and lane remain unchanged.

## Transformation

The stock TypeScript AST finds **13 generator bodies and two yield delegations**:
seven bodies in core.ts, including createSet's entries method, and six in
checker.ts, including the anonymous single-element elaboration function.
`evidence/before-census.json` records the actual adapted bodies and positions.
The post-adaptation AST census requires zero generators and zero delegations;
a remaining generator makes `census.cjs --check` exit 1.

The adapter validates all reviewed body token streams before writing either
file. It removes only the parsed generator stars and replaces the parsed body
spans, retaining every existing name, parameter, type parameter, modifier, and
return annotation. CRLF source is preserved. Module-local helpers introduce no
namespace exports. No explicit any is introduced; the existing flatMapIterator
return annotation remains unchanged, including its original next-input type.
No .a fixtures or authored .ts programs are added to Adamic: CJS templates produce
source only in the external TypeScript checkout. No counts.md refresh is needed.

The explicit objects implement next, return, throw, self-iteration, and disposal.
Iterator acquisition and next-method capture happen on first pull. Mapping,
filtering, array reads, checker queries, and delegate acquisition remain lazy.
Completed, unstarted-cancelled, throwing, and reentrant states are separate.
Cancellation closes active for-of inputs; an existing throw wins over a close
failure. Explicit yield delegation forwards next input, return, and throw,
including delegates that return an additional element during cancellation.
Delegated result identity and getter timing are preserved; a completion value
is read even when the original yield-star expression discards it. A getter that
throws during the input iterator step does not spuriously close the input.

The checker retains live array-length checks for JSX children and the original
first-pull length snapshot for tuple elaboration. Reverse iteration takes its
initial array length only on first next. Set buckets remain live between pulls.
The helpers use undefined completion internally; existing void signatures stay
unchanged, and SetIterator's actual undefined completion type is respected.

**Scoreboard: 498 changed lines in tsc, 420 added and 78 removed, two files.**
This is the incremental row measured by the unmodified apply pipeline.
`evidence/tsc.patch.gz` contains the external compiler source edits.

## Checks and mutants

`verify.cjs` executes extracted original and adapted functions with stock
TypeScript emission and Node 24.19.0. Its 15 focused cases cover every rewritten
body, exact values and side-effect traces, construction laziness, changing arrays
and set buckets between pulls, closing before first pull, repeated completion,
mapper exceptions, throw cancellation, delegated cancellation/recovery,
delegated-result getters, and throwing input-value getters. These probes run
in seconds. The upstream build checks the entire actual adapted source tree.
The probes are a focused behavioral measurement, not native execution.

All three requested real-source mutants were run independently and failed:

| Mutant | Catcher | Exit | Seconds |
|---|---|---:|---:|
| Restore singleIterator's original generator | AST census finds one remaining body | 1 | 1.670 |
| Start singleIterator with its element already consumed | Exact Node element sequence comparison | 1 | 2.048 |
| Read reverse iteration's array length at construction | Empty construction-trace laziness assertion | 1 | 2.124 |

Two additional mutants prove the guards: adding an unsanctioned public declaration
fails the API byte comparison, and changing a reviewed generator body fails the
adapter before writing source. The mutant compiler trees are isolated and removed
after their observations; evidence contains their actual failure logs.

The separately built prerequisite and adapted public declarations are
byte-identical: 590,847 bytes, SHA256
`edd733bf256465ddcbbd753a76cfe14cf4ef0a872ce5da7aa53072e1c9f93a9d`.
The full apply tree also passes the focused probes, declaration comparison, and
idempotence. The source adapter reports zero edits on a second run.

## Full stage3 lane

The unmodified, unfiltered full lane **PASSed**: **106,366 passing, one known
failure, zero pending**. Its sole failure is
`unittests:: Public APIs for typescript.d.ts should be acknowledged when they change`.
Its sole baseline difference is api/typescript.d.ts. The declaration guard
matches all 222 existing sanctions and reports no new public declarations.
Apply and upstream install/build exit 0; upstream tests/oracle exit 1 as required
for that known failure. The final lane wall time is 438.370s; all eight workers
and all runners were used. The oracle phase timings are in
`evidence/lane-oracle-report.json`.

The same fresh full-apply tree passes all focused Node probes, API byte equality,
and idempotence. Independent AST signature comparison confirms all thirteen
headers are unchanged apart from the generator star. Explicit any counts are
unchanged: core.ts 19 to 19 and checker.ts 4 to 4. `source-shape.json` records the
actual full-apply source hashes. Repackaging the runtime text into a CJS template
was independently replayed and reproduces those validated source bytes.
The lane's execution record retains the base HEAD at measurement time; it is not
rewritten to claim a later commit was tested.

## Native race measurement

Adamic was built in an isolated worktree at compiler scratch commit
`4cf6791a4a4ba515ebfa99fe3c12dd01a460948c`. Both native layouts, split 0 and 1,
exit 1 at **src/compiler/core.ts:335:41**, past the original generator refusal
at core.ts:332:1. The exact new refusal is:

```text
Adamic 0.1 refuses a value of type () => { done: true; value: undefined; } | { done: false; value: U; } seen as (value: unknown) => IteratorResult<U, undefined>, which can write false | undefined where false is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)
```

This is an observed callback-result invariance refusal. No native tsc binary was
produced or executed, and no compiler implementation was changed to bypass it.
The full entry logs and the compiler build log are preserved in evidence.

## Reproduction

Run from the repository root with the setup environment sourced. The paths
below are the actual scratch paths used in this run. Keep output directories
fresh. Test output is always redirected to logs.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/adapt66-setup.log 2>&1
source /workspace/adamic-tools/env.sh
export NODE_PATH=/tmp/adapt66-cache/api/node_modules
node stage3/adapt/66-temporary-explicit-iterators/adapt.cjs /tmp/adapt66-after > /tmp/adapt66-adapt.log 2>&1
node stage3/adapt/66-temporary-explicit-iterators/census.cjs /tmp/adapt66-before /tmp/adapt66-census-before.json > /tmp/adapt66-census-before.log 2>&1
node stage3/adapt/66-temporary-explicit-iterators/census.cjs /tmp/adapt66-after /tmp/adapt66-census-after.json --check > /tmp/adapt66-census-after.log 2>&1
node stage3/adapt/66-temporary-explicit-iterators/verify.cjs /tmp/adapt66-before /tmp/adapt66-after /tmp/adapt66-proof.json > /tmp/adapt66-verify.log 2>&1
node stage3/adapt/66-temporary-explicit-iterators/mutants.cjs /tmp/adapt66-before /tmp/adapt66-after /tmp/adapt66-mutants-final > /tmp/adapt66-mutants-final.log 2>&1
node stage3/adapt/66-temporary-explicit-iterators/api.cjs /tmp/adapt66-before /tmp/adapt66-after /tmp/adapt66-api.json > /tmp/adapt66-api.log 2>&1
STAGE3_CACHE=/tmp/adapt66-cache NODE_OPTIONS=--max-old-space-size=1536 stage3/lane/run.sh /tmp/adapt66-lane-validated > /tmp/adapt66-lane-validated.log 2>&1
ADAMIC_NATIVE_SPLIT=0 /tmp/adapt66-native build /tmp/adapt66-after/src/tsc/tsc.ts -o /tmp/adapt66-tsc-native > /tmp/adapt66-native-entry.log 2>&1
ADAMIC_NATIVE_SPLIT=1 /tmp/adapt66-native build /tmp/adapt66-after/src/tsc/tsc.ts -o /tmp/adapt66-tsc-native-split > /tmp/adapt66-native-entry-split.log 2>&1
```

`/tmp/adapt66-before` was produced by the full series before adding 66, and
`/tmp/adapt66-after` was copied from it before applying 66. Both were independently
installed with `npm ci --no-audit --no-fund` and built with `npm run build`, with
logs under /tmp/adapt66-{before-,}{install,build}.log. The final lane obtains a
fresh tree by running the complete series itself, including 66 in its normal order.
The compiler scratch CLI was built with
`go build -o /tmp/adapt66-native ./cmd/adamic` from /tmp/adapt66-compiler.

Setup timing: Go 0.074s, Node 0.110s, clang 0.527s, markdown dependencies 1.205s,
submodules 17.340s, Go build 249.449s, build cache warm 249.542s,
total 249.588s. nproc is 5; cpu.max is 400000 100000, a four-CPU quota.
The printed environment file is /workspace/adamic-tools/env.sh.
No new Go test units or .a fixtures were added. Focused probes and mutants take
less than 60 seconds each; the full existing upstream lane is the requested
measurement. The whole Adamic gate and unrelated package tests were not run.
