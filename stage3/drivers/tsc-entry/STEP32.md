Measured the reference compiler after skipping a conflicting stricter-options merge; no combined-mode result is claimed.
Compiler 88fe8de4; skipped stricter-options-next 8f32e51e; publication base 45487a80.
Both splits exit 1 at builder.ts:1246:69 TS2345; all fifteen saved snapshots remain.
Sixteen Node witnesses pass, sixteen source stdout mutants fail, and eleven evidence mutants are caught.
No working native tsc binary, stricter-mode compatibility proof, runtime execution, or full gate.

## Composition stopped at a conflict

The detached, never-pushed scratch started at
`88fe8de47cba6928539f1b35d4fd84c1ad06f34d` (`codex/project-references-source`).
Loader `16c22268` is its ancestor. The remote `codex/stricter-options-next`
exists at `8f32e51e8fc41b8f1177453213ca5453ce764486`, so the conditional c09
fallback and its siblings were not attempted.

An ordinary merge conflicts in seven paths:

```
internal/load/load.go
internal/load/project_console.go
internal/load/project_loader.go
internal/load/project_loader_test.go
internal/load/project_options.go
internal/load/source_fs.go
internal/oracle/counts.md
```

The merge was aborted and skipped as instructed. No conflict was resolved,
no compiler source was patched, and no scratch commit was pushed. The resulting
binary has `vcs.revision=88fe8de4...` and `vcs.modified=false`. The merge output,
conflict population, submodule pins and binary metadata are retained in
[evidence/step32](evidence/step32/). **Which stops survive the combined mode and
its new first stop are unavailable**, because that compiler was never formed.
The earlier 13/15 admission result in MODE-AND-ADAPTATIONS.md is not transferred
to this incompatible merge or presented as a fresh combined-mode observation.

## Two measured source profiles

The reference branch's unchanged adaptation pipeline reconstructs exactly the
81 reached-source hashes in the older main-efe9f404 experiment. Saved-site
replay uses that tree, preserving original CRLF and UTF-16 replacement offsets.
The new snapshot helper also retains locked Node declarations through a
node_modules symlink in every disposable project.

A second tree uses publication main's current adaptations. Its apply run failed
at adaptation 76 with `Cannot find module 'typescript'`; that last adaptation
was rerun with `CENSUS_TYPESCRIPT` set to the pinned stock 6.0.3 API path and
succeeded. The failure and repair output are saved. Builds made before this
repair are excluded from the final evidence. Full current-source hashes are
in result.json; reference hashes are in comparison.json. Neither profile's
upstream sources are committed.

Both final profiles were built from source with the unchanged scratch compiler:

```sh
ADAMIC_NATIVE_SPLIT=0 /tmp/step32-adamic build TREE/src/tsc/tsc.ts -o /tmp/step32-tsc-0 > split-0.stdout 2> split-0.stderr
ADAMIC_NATIVE_SPLIT=1 /tmp/step32-adamic build TREE/src/tsc/tsc.ts -o /tmp/step32-tsc-1 > split-1.stdout 2> split-1.stderr
```

Both splits exit 1. Complete stdout and stderr match byte for byte within each
profile. The first stop is unchanged:

> src/compiler/builder.ts:1246:69: error TS2345: Argument of type 'Path | undefined' is not assignable to parameter of type 'string'.

The current tree runs on Node 24.19.0 and prints `Version 6.0.3`, exits 0,
and has empty stderr. Native compilation never reaches lowering or clang.

## Fifteen saved sites and fifteen further stops

On the **reference-only** compiler, the historical replay reports **15 remain,
0 changed, 0 disappear**. Stops 9 and 14 remain only in placeholder snapshots;
the other thirteen are present in the pristine diagnostic stream. This is not
an observation of stricter-options-next admission.

A fresh current-tree walk captures the first stop plus fifteen further stops,
with both split builds at every stop. The first fifteen reproduce the exact
historical messages. Coordinates below are mapped back to the unmodified
current adapted tree; raw scratch coordinates and full messages remain in
[walk.json](evidence/step32/walk.json). Every row has a minimal Node witness and
an owner. Ownership is an inference from the observed checking phase: these
are compiler admission/checker stops, not observed runtime or library failures.
Rows 9 and 14 require correcting the exploratory placeholder when investigating
real progress; they do not justify a source adaptation.

| Stop | Original adapted site | Code | Saved result / pristine presence | Owner | Minimal program |
|---|---|---|---|---|---|
| 1 | builder.ts:1246:69 | TS2345 | Remain; pristine | Compiler | [01-indexed-path.a](probes/01-indexed-path.a) |
| 2 | builder.ts:1258:65 | TS2488 | Remain; pristine | Compiler | [02-tuple-parameter.a](probes/02-tuple-parameter.a) |
| 3 | builder.ts:2273:9 | TS2375 | Remain; pristine | Compiler | [03-present-undefined-field.a](probes/03-present-undefined-field.a) |
| 4 | builder.ts:395:118 | TS2345 | Remain; pristine | Compiler | [04-optional-path-argument.a](probes/04-optional-path-argument.a) |
| 5 | builder.ts:991:17 | TS2345 | Remain; pristine | Compiler | [05-optional-path-argument.a](probes/05-optional-path-argument.a) |
| 6 | checker.ts:10022:63 | TS2345 | Remain; pristine | Compiler | [06-optional-symbol-argument.a](probes/06-optional-symbol-argument.a) |
| 7 | checker.ts:10026:97 | TS2345 | Remain; pristine | Compiler | [07-optional-symbol-argument.a](probes/07-optional-symbol-argument.a) |
| 8 | checker.ts:10033:67 | TS18048 | Remain; pristine | Compiler | [08-optional-member-read.a](probes/08-optional-member-read.a) |
| 9 | checker.ts:10034:53 | TS2345 | Remain; placeholder artifact | Compiler | [09-never-argument.a](probes/09-never-argument.a) |
| 10 | checker.ts:10319:83 | TS2345 | Remain; pristine | Compiler | [10-optional-symbol-argument.a](probes/10-optional-symbol-argument.a) |
| 11 | checker.ts:10320:51 | TS2345 | Remain; pristine | Compiler | [11-optional-symbol-array.a](probes/11-optional-symbol-array.a) |
| 12 | checker.ts:10941:33 | TS2379 | Remain; pristine | Compiler | [12-present-undefined-argument.a](probes/12-present-undefined-argument.a) |
| 13 | checker.ts:12034:30 | TS18048 | Remain; pristine | Compiler | [13-optional-file-read.a](probes/13-optional-file-read.a) |
| 14 | checker.ts:12859:13 | TS2322 | Remain; placeholder artifact | Compiler | [14-void-result.a](probes/14-void-result.a) |
| 15 | checker.ts:13986:47 | TS2345 | Remain; pristine | Compiler | [15-optional-declaration-argument.a](probes/15-optional-declaration-argument.a) |
| 16 | checker.ts:14643:66 | TS2345 | Additional stop; pristine | Compiler | [step32-symbol-array.a](probes/step32-symbol-array.a) |

Stop 16 is `resolveAnonymousTypeMembers`'s call
`getIndexInfosOfIndexSymbol(indexSymbol, arrayFrom(members.values()))` at
checker.ts:14643:66. Its argument is inferred as `(Symbol | undefined)[]`, while
the parameter is `Symbol[]`. The new .a witness preserves Map-values-to-array
flow; Node prints `1`, while both `types` and `build` reject TS2345 at 6:8 with
the exact same diagnostic message. It does not demonstrate a runtime fault.

All sixteen witnesses exit 0 on Node. Each native witness build exits 1 at
checking. Thirteen historical witnesses match their complete stop message;
3 and 12 reproduce TS2375/TS2379 with smaller structural object displays.
The additional witness matches its complete message. The scheduled checks of
the skipped mode were not emitted or executed.

## Focused verification and mutants

All test output was written to files and retained as gzip evidence. No package
suite or full gate was run. Commands executed:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/step32-setup.log 2>&1
source /workspace/adamic-tools/env.sh
(cd /tmp/step32-compiler && go build -o /tmp/step32-adamic ./cmd/adamic) > /tmp/step32-compiler-build.log 2>&1
export SCANNER_TYPESCRIPT=/home/agent/.cache/adamic-stage3/api/node_modules/typescript/lib/typescript.js
python3 stage3/drivers/tsc-entry/step32-snapshots.py /tmp/step32-compiler /tmp/step32-adamic /tmp/step32-reference-adapted /tmp/step32-snapshots > /tmp/step32-snapshots.log 2>&1
python3 stage3/drivers/tsc-entry/continue.py /tmp/step32-walk-final/src /tmp/step32-adamic /tmp/step32-progress --limit 16 > /tmp/step32-progress.log 2>&1
python3 stage3/drivers/tsc-entry/step32-collect.py /tmp/step32-annotations.json > /tmp/step32-collect.log 2>&1
python3 stage3/drivers/tsc-entry/step32-verify.py > /tmp/step32-verify.log 2>&1
python3 stage3/drivers/tsc-entry/step32-mutants.py > /tmp/step32-mutants.log 2>&1
```

To reproduce, fetch the two exact pins, add a detached scratch at 88fe8de4,
try the merge without a conflict strategy, save its conflicts and abort it.
Initialize its pinned recursive submodules. Run each checkout's unchanged
stage3/apply.sh into a new external tree; use the explicit CENSUS_TYPESCRIPT
path for main's adaptation 76. Symlink the locked upstream node_modules into
each tree. Copy the final current src into `/tmp/step32-walk-final/src`, with
node_modules beside src. Never use that replaced tree as a pristine oracle.
The collector's observed scratch paths are intentionally explicit. Its annotation
input is retained as evidence/step32/annotations.json.

The verifier reports `15 saved sites verified; 16 walked stops; Node witnesses
and stdout mutants verified; conflicting mode excluded`. One real-source mutant
per witness appends an additional stdout observation. All sixteen still run normally
on Node and produce different stdout; the expected-output oracle kills them.
These are output-oracle sensitivity checks, not proof of inserted runtime guards.
Mutation source files stay in /tmp and are not committed.

Eleven independent evidence mutants fail with retained AssertionErrors:

| Mutant | Catcher |
|---|---|
| Invent combined-mode success | Excluded-composition assertion |
| Change compiler SHA | Exact measured pin assertion |
| Invent first stop | Raw diagnostic comparison |
| Claim split success | Recorded exit assertion |
| Drop saved stop 15 | Required fifteen-site population |
| Corrupt witness Node stdout | Expected source observation |
| Remove valid owner | Ownership population assertion |
| Claim stop 1 disappeared | Classification recomputed from raw diagnostics |
| Change split 1 stderr | Full stream byte comparison |
| Claim dirty binary | Embedded clean provenance assertion |
| Change Node entry output | Version control assertion |

[step32-counts.md](step32-counts.md) refreshes this territory's local fixture
registry. No central compiler-oracle fixture was added, and no allocation counters
are claimed for intentionally refused programs.

Setup timing lines: Go 0.023s, Node 0.025s, markdown skip 0.008s, submodules
0.084s, markdown ready 0.085s, clang 0.248s, validated build-cache skip 0.865s,
test binaries deferred 0.866s, cache warm 0.868s, done 0.917s. `nproc=5`,
`cpu.max=400000 100000` (four CPUs). Environment:
`/workspace/adamic-tools/env.sh`; Node 24.19.0, Go 1.27.1, clang 20.1.8.

No combined-mode compiler, compatible merge proof, native tsc executable,
--tiny native comparison, runtime guard execution, upstream test suite, watch
workload, or full repository gate was produced. Only evidence and driver helpers
under stage3/drivers/tsc-entry are published; compiler, adaptation, runtime and
library files are unchanged.
