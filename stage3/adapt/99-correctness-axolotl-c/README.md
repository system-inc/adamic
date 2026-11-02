# 99c: present undefined in serialized configuration

lane: run as part of Correctness's wave

Base: main `80c2d3282e44b527555d59690bb413cc883d2bf2`.
Source: TypeScript 6.0.3, commit `050880ce59e30b356b686bd3144efe24f875ebc8`.
Adaptation 20 sanctions adding undefined to existing optional data properties
when the implementation stores a present undefined. Object keys, spreads,
assignments, readonly modifiers and question marks stay in place.

## Member and producer review

Coordinates are pinned upstream coordinates. The common producer is
`src/compiler/commandLineParser.ts`, `convertToTSConfig`. The compilerOptions
literal clears ten keys before JSON serialization; the reference mapper spreads
each reference, replaces path, and clears originalPath. An undefined key is still
an own property. Removing the key would change enumeration and spread behavior.

| Owner.member | Visibility in pinned API | Explicit undefined producer | 99c action |
| --- | --- | --- | --- |
| CompilerOptions.showConfig | @internal | commandLineParser.ts:2665 | boolean to boolean \| undefined |
| CompilerOptions.configFile | @internal | commandLineParser.ts:2666 | TsConfigSourceFile to TsConfigSourceFile \| undefined |
| CompilerOptions.configFilePath | @internal | commandLineParser.ts:2667; :3087 when configFileName is undefined | Already string \| undefined after 20; unchanged |
| CompilerOptions.help | @internal | commandLineParser.ts:2668 | boolean to boolean \| undefined |
| CompilerOptions.init | @internal | commandLineParser.ts:2669 | boolean to boolean \| undefined |
| CompilerOptions.listFiles | @internal | commandLineParser.ts:2670 | boolean to boolean \| undefined |
| CompilerOptions.listEmittedFiles | @internal | commandLineParser.ts:2671 | boolean to boolean \| undefined |
| CompilerOptions.project | Public, API :7095 | commandLineParser.ts:2672 | string to string \| undefined |
| CompilerOptions.build | @internal | commandLineParser.ts:2673 | boolean to boolean \| undefined |
| CompilerOptions.version | @internal | commandLineParser.ts:2674 | boolean to boolean \| undefined |
| ProjectReference.originalPath | Public, API :6987 | commandLineParser.ts:2677 | string to string \| undefined |
| TSConfig.exclude | @internal | commandLineParser.ts:2681, conditional spread | readonly string[] to readonly string[] \| undefined |
| TSConfig.include | @internal | commandLineParser.ts:2680, conditional spread | readonly string[] to readonly string[] \| undefined |

The generic command line producer `parseOptionValue` also stores undefined at
:2111 for a `null` argument to any of the named command line options. Its
TSConfig-only branch does the same at :2050, but these flags and project are not
TSConfig-only options. `validateJsonOptionValue` can return undefined for invalid
values. Those paths write through OptionsBase's string index. Config JSON
conversion, extension merges and option cloning can copy those values.
`setConfigFileInOptions` (:3032) only defines a non-enumerable configFile when
its argument is truthy; it does not store undefined. The harness preProgram literal (`src/harness/compilerImpl.ts:267`) also
copies a potentially undefined `compilerOptions.configFile` into a present key.
The parsed reference
producer (:3209) stores the checked string ref.path. Other originalPath owners
in module resolution are separate contracts and remain unchanged.

The first ten-member probe removed both requested rows but exposed a TS2375
at the containing config literal on `TSConfig.exclude`. Widening exclude alone
then exposed `TSConfig.include` at the same literal. These are the only sibling
contracts added. The conditional spread checks configFileSpecs, not either array.
`validatedExcludeSpecs` is explicitly `readonly string[] | undefined`;
`filterSameAsDefaultInclude` (:2718) returns undefined for absent, empty or default
include specifications. Thus both keys can be present with undefined. Neither
property appears in the pinned API, so no additional API sanction is needed.

## Readers

The following list enumerates symbol-resolved reads across upstream src,
including services, server, harness and test runners. Duplicate reads on a line
are shown once. Paths are relative to src. Dynamic indexed option consumers
additionally serialize/filter values in commandLineParser (`serializeOptionBaseObject`,
`generateTSConfig`, `convertCompilerOptionsForTelemetry`,
`convertToOptionsWithAbsolutePaths`, `handleOptionConfigDirTemplateSubstitution`),
and compare/copy option maps in the compiler and harness. Those consumers already
handle undefined through CompilerOptions' string index. JSON serialization omits
undefined values, preserving the purpose of these clearing literals.

- `ProjectReference.originalPath`: `compiler/commandLineParser.ts:2677`
- `CompilerOptions.build`: No direct named value read; generic option-map consumers only.
- `CompilerOptions.configFilePath`: `compiler/utilities.ts:711,6682`; `compiler/commandLineParser.ts:2984,2985`; `compiler/moduleNameResolver.ts:485,486,2897`; `compiler/moduleSpecifiers.ts:638,639`; `compiler/emitter.ts:482,584,649,651,689,727,742`; `compiler/program.ts:1142,1574,1791,4117,4134,4266,4284,4285,4410`; `compiler/watchPublic.ts:112`; `compiler/tsbuildPublic.ts:1765,1831,1850,1881`; `compiler/executeCommandLine.ts:1021,1152`; `services/documentRegistry.ts:287`; `server/project.ts:1119`; `testRunner/compilerRunner.ts:245,246`
- `CompilerOptions.configFile`: `compiler/commandLineParser.ts:2650,2652,2653,2679,2680,2681`; `compiler/moduleNameResolver.ts:2887,2974`; `compiler/watchUtilities.ts:414,606`; `compiler/program.ts:1267,1297,1324,1325,2578,2579,3275,3276,4021,4043,4624,4671,4673,4700,4703,4726,4742,4745`; `compiler/programDiagnostics.ts:359,364,368,375,388,422`; `compiler/watch.ts:417,428,497,507`; `compiler/watchPublic.ts:927,932,993`; `compiler/tsbuildPublic.ts:1239,1243,1769`; `services/utilities.ts:2246`; `services/refactors/moveToFile.ts:282`; `services/codefixes/importFixes.ts:1222`; `server/project.ts:1119,1251,1274,3155`; `server/editorServices.ts:2929,3207,3458,4507`; `harness/harnessIO.ts:366`; `harness/compilerImpl.ts:267`; `testRunner/compilerRunner.ts:246`; destructuring reads at `services/getEditsForFileRename.ts:105`, `services/codefixes/fixEnableJsxFlag.ts:19,32`, `services/codefixes/fixModuleAndTargetOptions.ts:28`
- `CompilerOptions.help`: `compiler/executeCommandLine.ts:592`
- `CompilerOptions.init`: `compiler/executeCommandLine.ts:582`
- `CompilerOptions.listEmittedFiles`: `compiler/emitter.ts:767`
- `CompilerOptions.listFiles`: `compiler/watch.ts:343`
- `CompilerOptions.project`: `compiler/executeCommandLine.ts:602,608,612,619`; `services/stringCompletions.ts:759`; `harness/compilerImpl.ts:245,246,256`; `testRunner/projectsRunner.ts:148,150`; `testRunner/unittests/config/showConfig.ts:28,29`
- `CompilerOptions.version`: `compiler/executeCommandLine.ts:587`
- `CompilerOptions.showConfig`: `compiler/executeCommandLine.ts:636,655,704`

The two TSConfig siblings have no direct named reads after construction. Their
whole config is read by JSON.stringify in `compiler/executeCommandLine.ts:666,706`
and the show-config harness at `testRunner/unittests/config/showConfig.ts:34-37`.
JSON serialization intentionally omits their undefined values.

Indexed accesses and spreads of the reviewed typed contracts occur at the
following locations (includes write targets and computed keys limited to other
options, so this is a conservative reader/copy inventory):

- `compiler/utilities.ts:9370,9402`
- `compiler/commandLineParser.ts:2677,2901,2933,2977,2979,3613,3785,4269`
- `compiler/moduleNameResolver.ts:1886`
- `compiler/checker.ts:8637`
- `compiler/program.ts:1525,5099`
- `compiler/builder.ts:1447,1449`
- `compiler/tsbuildPublic.ts:334`
- `services/transpile.ts:124,125,135,252,256`
- `server/project.ts:2368,2755,2756`
- `server/editorServices.ts:463,465,479,492`
- `harness/harnessIO.ts:292`
- `harness/compilerImpl.ts:267`
- `harness/fourslashImpl.ts:105`
- `testRunner/projectsRunner.ts:478,486`
- `testRunner/unittests/customTransforms.ts:23`
- `testRunner/unittests/helpers/tsc.ts:431`
- `testRunner/unittests/tsbuild/commandLine.ts:63,84,182,228,271`
- `testRunner/unittests/helpers/declarationEmit.ts:12,45`
- `testRunner/unittests/helpers/noEmit.ts:400,431,461`
- `testRunner/unittests/tsc/incremental.ts:446,469`
- `testRunner/unittests/tscWatch/sourceOfProjectReferenceRedirect.ts:116`
- `testRunner/unittests/tsserver/projectReferences.ts:448,1347,1550`

## Adapter contract

Stock TypeScript 6.0.3 parses types.ts and commandLineParser.ts. Each of twelve
members is pinned by its
unique interface owner and exact complete member text. Each member must be
pristine or have exactly the sanctioned undefined union. A half-adapted tree,
missing member, duplicate owner or text drift throws before any write. Parse
checks and unchanged-input validation for both files precede either write. A second
run writes nothing. No source coordinate is a rewrite address.

## Proof

Both full apply runs pass. The baseline final snapshot equals the final tree's
99c-before snapshot byte for byte. The incremental source diff is twelve lines
replaced in two files: ten in types.ts and two in commandLineParser.ts.

The current main CLI is built with `go build ./cmd/adamic`. Both checks pass the
same 79 sorted .ts/.a roots to `adamic types`, using the unchanged checker and
prelude. Complete diagnostic strings, multiplicity, columns and related chains
are compared after replacing only the tree root path.

| Whole-program measurement | Before | Final |
| --- | ---: | ---: |
| All diagnostics | 53 | 51 |
| TS2375 | 9 | 8 |
| TS2322 | 7 | 6 |
| D033 and D034 | 2 | 0 |
| Added or changed diagnostic chains | | 0 |

Every other checker-code count and every other diagnostic chain is unchanged.
The ten-member intermediate probe was 53 to 52: both target rows disappeared,
but exclude appeared. The exclude-only probe remained at 52, replacing exclude
with include. The final twelve-member result removes exactly the two requested
rows with no additions anywhere.

All thirteen restoration mutants are killed by the real whole-program checker.
Each of the twelve edited members, restored independently, adds exactly one
construction error and raises the total from 51 to 52. The two TSConfig mutants
restore the exposed outer-literal TS2375. Restoring the already-widened
configFilePath adds D033 and its independent assignment TS2412 at adapted
commandLineParser.ts:3092, raising the total to 53. No existing diagnostic is
removed by any mutant. Each mutation is restored before the next one.

Stock `npm run build` passes before and after, including its checked
`build-tests` task. No compiler flag is weakened. Stock TypeScript 6.0.3
transpileModule, ES2024 target, ESNext modules and comments retained, proves:

| Edited file | Identical JavaScript bytes | SHA-256 |
| --- | ---: | --- |
| src/compiler/commandLineParser.ts | 161985 | dc05a488396d003e9ec4d5ff78ef11af0e7e5f99d6264d5860872975b6ec7c5d |
| src/compiler/types.ts | 139659 | 4839bedc439cad06969bad2a3d805a7c0b480f06df73fcff62d2cd0957009fc3 |

All ten built JavaScript artifacts, 28,602,533 bytes including run.js and
both compiler bundles, are also byte-identical. Every other source file is
byte-identical. The second adapter run reports zero files and zero declarations,
with both saved source files unchanged. A partly adapted tree, last-file member
drift and last-file duplicate owner each throw before either source is written.
A pristine two-file fixture also adapts once and is stable on its second run.

The real built convertToTSConfig implementations preserve all thirteen reviewed
own-undefined keys. An actual parsed empty config supplies the include/exclude
witness; JSON serialization is identical before and after.

The generated public declarations equal stage3/lane/generate-api.cjs's full
sanctioned expected text byte for byte. 99c adds exactly two replacement lines
(two added and two removed) beyond main's output. The two public properties keep
their question marks; all ten edited internal properties are absent from the API.
Only tests/baselines/reference/api/typescript.d.ts is accepted in the external
checkout, and only after byte equality with the generator. The raw ten-member
oracle has 106,366 passing, one API snapshot failure, zero pending, and exactly
that one baseline difference. Its accepted-snapshot rerun has 106,367 passing,
zero failing, zero pending and no baseline differences.

The final twelve-member full oracle also passes: **106,367 passing, zero failing,
zero pending and zero baseline differences**. Its install and checked build both
exit zero. The byte, declaration and no-write proofs pass again after this oracle.

Evidence is on [typescript/axolotl-99c-evidence](https://github.com/system-inc/adamic/tree/typescript/axolotl-99c-evidence/stage3/census/axolotl-99c),
under `stage3/census/axolotl-99c/`. This adaptation directory contains only the adapter,
this review and the public sanctions. That evidence branch retains full apply,
setup, build, oracle, diagnostic and mutant logs, the comparison scripts, and the
intermediate exclude/include probes. Setup exits zero in 918.367 seconds.

Material commands, from Adamic's root, with complete output redirected to files:

```sh
export GOPROXY='https://proxy.golang.org|direct'
timeout --kill-after=60s 1800 bash cloud/setup.sh > /workspace/scratch/setup.log 2>&1
source /workspace/adamic-tools/env.sh
# Baseline apply ran at the recorded main base before creating this directory.
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/upstream bash stage3/apply.sh /workspace/scratch/before > /workspace/scratch/apply-before.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/upstream bash stage3/apply.sh /workspace/scratch/after-final > /workspace/scratch/apply-after-final.log 2>&1
go build -o /workspace/scratch/adamic ./cmd/adamic > /workspace/scratch/build-adamic.log 2>&1
NODE_PATH=/workspace/adamic/stage3/api/node_modules node stage3/lane/generate-api.cjs /workspace/scratch/upstream > /workspace/scratch/api-expected.json 2> /workspace/scratch/api-expected.stderr
# After the generated declarations pass exact reconstruction, accept only that
# API reference in the external tree and run the unchanged full oracle.
bash stage3/oracle/run.sh /workspace/scratch/after-final /workspace/scratch/oracle-final > /workspace/scratch/oracle-final.log 2>&1
```

The CLI checks use argument arrays to pass every sorted compiler root:

```python
import pathlib, subprocess
for label, tree in [("before", "/workspace/scratch/before"),
                    ("after", "/workspace/scratch/after-final")]:
    roots = sorted(str(p) for p in pathlib.Path(tree, "src/compiler").rglob("*")
                   if p.suffix in [".ts", ".a"])
    with open(f"/workspace/scratch/types-{label}.stdout", "w") as out, \
         open(f"/workspace/scratch/types-{label}.stderr", "w") as err:
        result = subprocess.run(["/workspace/scratch/adamic", "types", *roots],
                                stdout=out, stderr=err)
    assert result.returncode == 1  # Remaining whole-program diagnostics.
```

No lane manifest or shared apply, checker, API generator or oracle is changed.
