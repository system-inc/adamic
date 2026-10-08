Built a minimal composite dependency and referring app with Node's tsc --build; the emitted app prints 7.
Publication base cc503d56; experimental compiler composition remains pinned to the prior c09c22b8 evidence.
Source-root loading clears TS6305; entry options alone stop at debug.ts:200:19 TS2339; including declared Node types reaches sys.ts:1502:51 lowering.
Focused loader tests, reference verification, thirteen evidence mutants and a real unimported dependency mutant pass locally.
Only evidence, reproducers and measurement scripts are published; no compiler changes or working native tsc binary are claimed.

## Minimal two-project observation

Authored sources are the existing
[`app/main.a`](probes/missing-project-output/app/main.a) and
[`dependency/value.a`](probes/missing-project-output/dependency/value.a):

```ts
// dependency/value.a
export const value: number = 7;
// app/main.a
import { value } from "../dependency/value.a";
console.log(value);
```

The runner copies these to `value.ts` and `main.ts` in scratch and changes the
import suffix to `value.js`. Both configs use `composite: true`, `declaration:
true`, `strict: true`, target es2020, module esnext, resolution bundler, `types:
[]`, `skipLibCheck: true`, rootDir `.`, and outDir `../out/dependency` or
`../out/app`. Dependency includes `value.ts`; app includes `main.ts` and has
`"references": [{"path": "../dependency"}]`. A parent package.json declares
`"type": "module"`. Exact configs are retained in `evidence/project-references/configs/`.

Before any output exists, Adamic's auto-project **file-root** loader reports:

> app/main.ts:1:23: error TS6305: Output file '/tmp/tsc-entry-reference-unit/two-project/out/dependency/value.d.ts' has not been built from source file '/tmp/tsc-entry-reference-unit/two-project/dependency/value.ts'.

The loader probe returns `loaded: false`; the native command exits 1. This is
precisely the missing declaration mechanism seen at the real entry's namespace
re-export. Node 24.19.0 runs TypeScript 6.0.3's stock tsc:

```sh
node STOCK_TYPESCRIPT/lib/tsc.js --build app/tsconfig.json --verbose > build.stdout 2> build.stderr
node out/app/main.js > node.stdout 2> node.stderr
```

Build exits 0. Verbose output records **Building dependency**, then **Building
app**, and creates dependency/value.d.ts plus each project's JavaScript and build
info. The emitted app prints `7\n`, exits 0 and has empty stderr. Without changing
source or configs, Adamic's loader now returns `loaded: true` and no errors.
That establishes the declaration-output prerequisite rather than an invalid
program. Loader success is separate from native lowering.

The independent single-program source check also exits 0, without needing
those declaration outputs:

```sh
node STOCK_TYPESCRIPT/lib/tsc.js --ignoreConfig --noEmit --strict --target es2020 --module esnext --moduleResolution bundler app/main.ts dependency/value.ts > single.stdout 2> single.stderr
```

`reference-oracle.cjs` independently parses both configs, takes every configured
source as a root, and creates one stock TypeScript program without project
references. Its fresh minimal-program run accepts exactly two sources.

Reproduce the build/output comparison from the publication checkout:

```sh
source /workspace/adamic-tools/env.sh
export SCANNER_TYPESCRIPT=/home/agent/.cache/adamic-stage3/api/node_modules/typescript/lib/typescript.js
python3 stage3/drivers/tsc-entry/reference-program.py /tmp/tsc-entry-mode-adamic /tmp/tsc-entry-mode-admission/probe NEW_OUTPUT > reference-run.log 2>&1
```

The binaries are the pinned, unchanged experimental composition from the prior
unit, whose provenance is in `evidence/mode-c09c22b8/provenance.json`.

## Where Adamic chooses this behavior

In that compiler, **internal/load/load.go, `loadInput`**, passes
`projectConfig.ProjectReferences()` into `tsoptions.NewParsedCommandLine` and
its `buildProgram` closure calls `compiler.NewProgram`. It does not invoke a
solution builder, recursively emit declarations, or set
`UseSourceOfProjectReference`. It therefore creates an ordinary checker program
whose referenced modules resolve to declaration output. This is the decisive
Adamic choice. **internal/load/project_loader.go, `projectOptionsForRoots`**
selects the nearest tsconfig and clones its options; it does not build the
references. **internal/load/project_options.go, `auditProjectOptions`**, passes
`config.ProjectReferences()` to its ordinary/stricter shadow programs as well.
The actual unchanged call sites are recorded in `loader-trace.json`.

The embedded frontend's
`cohere/TypeScript/tsc/internal/compiler/program.go`,
`ProgramConfig.canUseProjectReferenceSource`, requires
`UseSourceOfProjectReference` to be true; its omitted value here is false.
The TS6305 message is emitted by `Checker.resolveExternalModule` in
`cohere/TypeScript/tsc/internal/checker/checker.go` when the reference redirect
names an output declaration that is absent. Adamic formats the checker diagnostic;
it does not invent TS6305.

A separate API path, `internal/load/project.go`, `parseProject`, explicitly
rejects project references with "project references are not yet supported; build
each project explicitly". That path is not the file-root command used here.
All these names refer to the pinned stricter-options composition, not to a claim
that publication main has the same auto-project implementation.

## Faithful behavior and recommendation

`tsc --build` is a solution build: it builds dependency projects first under
their own options, then checks consumers against emitted declarations. An ordinary
`tsc -p app` does not promise to build references automatically; its TS6305 is
faithful when declarations are missing. An explicit single-program check from
both sets of sources instead uses one checker and one effective option set,
with no reference-output substitution. Both are legitimate operations, and they
are different contracts.

**Recommend source-root loading for Adamic's single native executable build.**
Native lowering needs dependency implementation bodies; declaration-only outputs
cannot supply those bodies. A single checked source program also fits Adamic's
current whole-program lowering and makes a freshly adapted tree buildable
without a prior JavaScript/declaration build. Discover the full reference graph,
include every configured source root, and make the effective option contract
explicit. Retain the original execution entry rather than treating every added
checking root as an executable entry.

This should not silently flatten arbitrary incompatible configs. Different libs,
module settings, strict flags or ambient types can change meaning. The loader
must either prove compatibility for one effective program or refuse and require
separate checker ownership. A future solution/native-object build could preserve
per-project ownership and link compiled implementations, but simply generating
.d.ts files would only solve checking. No production implementation is proposed
or pushed in this unit.

## Disposable source-root experiment

The prior never-pushed compiler worktree remains
`/tmp/tsc-entry-compiler-mode`, merge 29933a3e plus its previously documented
24 reconciliation choices. The experiment changes only its loader, behind
`ADAMIC_REFERENCE_EXPERIMENT=1`. The disposable implementation recursively
parses resolved reference configs, deduplicates all their `FileNames()`, and
adds them to checker roots. It removes project references from the main program
and every option-audit program, widens rootDir to the common source directory,
and permits checker ownership only for the enumerated reference configs.
The original explicit execution roots remain separate. Ordinary diagnostics
are retained; no error message is filtered to force progress.

These loader changes are **not committed anywhere**. The local construction
script is `/tmp/tsc-entry-reference-experiment.py`; before/after source hashes,
binary hashes and environmental switches are in
`evidence/project-references/provenance.json`. That script constructs the first
profile; the final helper adds the separately enabled Node-types union described
below. The old binary is retained as the TS6305 control. The binary before the
optional-types helper is `/tmp/tsc-entry-reference-adamic`; the final experimental
binary is `/tmp/tsc-entry-reference-types-adamic`.

The minimal two-project experiment uses a fresh source copy with no `out/`.
Its loader accepts it, and native lowering then refuses `console.log`. Thus the
reference output requirement is cleared without a generated declaration file.

The real entry uses the same 746 adapted source files as the previous unit,
without any throwing placeholders. The installed lockfile dependencies are
reused through a node_modules symlink; the compiler declaration output remains
absent. Both experiments build `src/tsc/tsc.ts` with `ADAMIC_NATIVE_SPLIT=0` and
`1`; each pair has byte-identical stdout/stderr and exits 1:

| Source program | First stop | Location |
| --- | --- | --- |
| Entry options, all referenced sources | TS2339: Property 'captureStackTrace' does not exist on type 'ErrorConstructor'. | src/compiler/debug.ts:200:19 |
| Same source roots, union of projects' declared `types` | stage 0 can't lower a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration yet | src/compiler/sys.ts:1502:51 |

Both stops are inside `src/compiler`. TS6305 disappears in both profiles.
The strict first result is an **option-composition problem**, not a missing Node
installation: tsc's entry has `types: []`, while its compiler dependency has
`types: ["node"]`. The first profile keeps entry options; Node ambient Error
extensions disappear when compiler sources are brought into that checker.
[`reference-host-error.a`](probes/reference-host-error.a) reproduces the
captureStackTrace diagnostic on the same type profile; Node prints `true`.
Both Adamic and stock tsc report TS2339 at the call's line 2, column 7.

The second profile explicitly enables
`ADAMIC_REFERENCE_EXPERIMENT_TYPES=1`, collecting only the `types` names already
written in the participating configs. For this tree that produces `["node"]`.
It does not add a type stub, change declarations, or weaken strict options.
The independent stock oracle accepts **all 81 sources** with zero diagnostics
under this explicit common profile. With entry-only types, stock has 66 errors;
Adamic has the same 65 diagnostic locations/codes except the stock-only
`console` error at debug.ts:865:16, because Adamic supplies console via its
prelude. This difference is retained explicitly in verification.

Use these switches with the final scratch binary to reproduce either pair:

```sh
ADAMIC_REFERENCE_EXPERIMENT=1 ADAMIC_NATIVE_SPLIT=0 /tmp/tsc-entry-reference-types-adamic build /tmp/tsc-entry-mode-final-adapted/src/tsc/tsc.ts -o /tmp/flat-tsc-0 > flat-0.stdout 2> flat-0.stderr
ADAMIC_REFERENCE_EXPERIMENT=1 ADAMIC_REFERENCE_EXPERIMENT_TYPES=1 ADAMIC_NATIVE_SPLIT=0 /tmp/tsc-entry-reference-types-adamic build /tmp/tsc-entry-mode-final-adapted/src/tsc/tsc.ts -o /tmp/flat-types-tsc-0 > flat-types-0.stdout 2> flat-types-0.stderr
```

Repeat with split 1. The recorded first-profile runs used the equivalent
pre-types binary. The entry still runs on Node and prints `Version 6.0.3`.
This is evidence of the next native refusal, not a native execution comparison.

## Tests, mutants and coverage

Toolchain setup succeeded: Go build 52.337s, cache warm 52.490s, total 52.583s;
`nproc` 5, CPU quota 4. All setup output is retained.

```sh
ADAMIC_REFERENCE_EXPERIMENT=0 go test ./internal/load -run 'TestProduction|TestProject|TestOptionLedger|TestUnsupportedOptionContract' -count=1 -timeout 20m > loader-tests.log 2>&1
python3 stage3/drivers/tsc-entry/verify-references.py > verification.log 2>&1
python3 stage3/drivers/tsc-entry/reference-mutants.py > mutants.log 2>&1
```

Focused loader tests pass in 6.348s. The real-input mutant adds a configured,
**unimported** `dependency/unimported.ts` with `export const broken: number =
"wrong";`. Both stock `tsc --build` and scratch single-program loading reject
it with TS2322 at 1:14; the independent stock source program rejects it too.
This proves that configured reference roots are checked even when import
traversal would never reach them.

Thirteen evidence mutants corrupt the build exit, dependency-before-app order,
Node output, missing-output diagnostic, first source stop, first lowering stop,
referenced root count, a source diagnostic, ambient options, source identity,
publication compiler identity, the rejected unimported dependency, and the
reference-loading call-site trace. `verify-references.py` catches each, and the
mutant runner first requires the original evidence to pass. The four fresh
build/output observations can be repeated with `reference-program.py`.

All 746 adapted source hashes are unchanged. Publication `internal`, `cmd`,
`cohere` and adaptation files are unchanged; experimental Go files stay in
scratch. Raw logs are compressed to preserve their original bytes. No full Go
suite or full gate was run: publication changes are evidence and driver tools,
and focused loader regressions plus the fixtures above cover this experiment.
No native tsc binary, --tiny comparison, declaration-build implementation,
per-project native linker or proof of general config-flattening compatibility
is claimed.
