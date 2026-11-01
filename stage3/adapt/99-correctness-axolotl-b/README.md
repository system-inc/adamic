# 99b: two disproved required-read invariants

Byte-identical in JavaScript. Both requested reads remain unasserted because real command lines reach them with undefined. This adapter validates the declined sites and writes nothing. Removing either diagnostic with a required-value assertion would introduce an Adamic failure on existing command-line behavior.

lane: run as part of Correctness's wave

Base: main c74051ffc5ee94f4b7f732b1532cb219032b183d. Source: stock TypeScript 6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8. References read in full: 74dda102's stage3/adapt/99-correctness-axolotl-a/adapt.cjs and README.md, stage3/apply.sh, and ad5d2151's stage3/census/axolotl/commands.sh. Measurements include cherry-picked 99a. The published commit adds only this directory on top of the recorded main tip.

Coordinates in the Site column are the requested adapted-tree diagnostic coordinates. Other line numbers refer to the pinned stock source. Paths below are relative to src/compiler/.

| Site | Read and function | Invariant review and witness | JavaScript |
| --- | --- | --- | --- |
| commandLineParser.ts:2077:79 TS2345 (D031) | `args[i]` in `parseOptionValue`, argument to `parseInt` | **Unasserted.** `parseStrings`:1978 bounds the flag's index by `i < args.length`, then increments i at 1980 before both calls at 1988 and 1993. Thus the value index can equal args.length. The missing-argument test at 2070-2072 pushes a diagnostic without returning. Undefined passes `args[i] !== "null"` at 2074 and reaches the number branch at 2077. That branch increments i again at 2078 and returns it at 2115, ending the outer loop. Node witness: `ts.parseCommandLine(['--maxNodeModuleJsDepth'])` calls parseInt with undefined, returns NaN for the option, and records TS6044. The same caller contract applies to watch options and response files, whose tokens reenter parseStrings. No presence invariant holds. | byte-identical, unchanged |
| commandLineParser.ts:2957:13 TS2322 (D036) | `formatSingleValue(value, map)` in `formatValueOrArray` | **Unasserted.** The two callers of formatValueOrArray are emitOption's comment and active branches at 2935 and 2938, both using `(options[setting] ?? defaultValue) as PresetValue` at 2933. The cast does not exclude undefined at runtime. Fixed emitOption calls at 2845-2895 pass strings, enum numbers, booleans, or arrays, with options.lib as another array source. The remaining-options loop at 2898-2902 instead supplies `options[allSetOptions[0]]` as defaultValue. Object.keys at 2838 retains own keys holding undefined. `--init --outFile null` stores own outFile=undefined at 2111; outFile has no fixed emitOption call, so both operands of `??` are undefined. The scalar branch at stock 2952 calls formatSingleValue with undefined; outFile's string type has no map, and JSON.stringify at 2960 returns undefined. The array callback at 2949 also calls formatSingleValue, but cannot prove the scalar branch. Node witness emits `"outFile": undefined,` with zero parse errors. executeCommandLine.ts:577-584 permits this command and passes the same options to writeConfigFile; 1300 calls generateTSConfig. No presence invariant holds. | byte-identical, unchanged |

## Adapter contract

Stock TypeScript 6.0.3 parses commandLineParser.ts. Each declined read is pinned by its unique named function, AST kind, exact text, occurrence ordinal, total count, and exact parent text. The numeric read is occurrence 4 of 8 args[i] expressions in parseOptionValue; the scalar call is the sole exact formatSingleValue(value, map) in formatValueOrArray. Stores and asserted declined reads are rejected. Both partially and fully asserted inputs are refused. Validation of all sites and the unchanged input completes before returning. There are no edits to plan and no writes on any run. No lane manifest is edited.

## Node witnesses

Run with stock TypeScript 6.0.3 available on NODE_PATH:

```sh
NODE_PATH=/workspace/adamic/stage3/api/node_modules node stage3/adapt/99-correctness-axolotl-b/witness.cjs > /workspace/scratch/witness.json 2>&1
```

The script temporarily wraps parseInt and JSON.stringify to observe the actual absent argument and result, restores both functions in finally blocks, and asserts the resulting NaN, TS6044, zero init parse errors, and emitted undefined. It uses TypeScript's stock built library, without source edits. No generated configuration is written by the witness.

## Proof

Both apply.sh runs pass end to end. Both diagnostics reproduce at the exact requested coordinates on the baseline with 99a and remain afterward. Their removal proof cannot truthfully pass for these sites.

| Measurement | Before | After |
| --- | ---: | ---: |
| Sorted compiler roots | 79 | 79 |
| Whole program, every checker code | 129 | 129 |
| Requested diagnostics | 2 | 2 |
| Added anywhere | | 0 |
| Removed anywhere | | 0 |
| commandLineParser.ts transpileModule JavaScript bytes | 161985 | 161985 |

Complete diagnostic multisets are equal, including columns, messages, multiplicity and elaboration chains. Source bytes are equal as well. Stock ts.transpileModule uses target ES2024, module ESNext and retained comments. The baseline's 99a-after snapshot equals the final tree's 99b-before snapshot. The first and second adapter runs both return `{"files":0,"assertions":0,"javascriptChanging":false}`. All five refusal fixtures pass without writes. No assertion is added at either site.

Material commands from the repository root, with complete output in scratch files:

```sh
export GOPROXY='https://proxy.golang.org|direct'
timeout --kill-after=60s 1800 bash cloud/setup.sh > /workspace/scratch/setup.log 2>&1
git cherry-pick 74dda102 > /workspace/scratch/cherry-pick-99a.log 2>&1
source /workspace/adamic-tools/env.sh
git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /workspace/scratch/upstream > /workspace/scratch/upstream.log 2>&1
(cd stage3/api && npm ci --ignore-scripts --no-audit --no-fund > /workspace/scratch/packages.log 2>&1)
# Started before creating 99b, with 99a present.
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/upstream bash stage3/apply.sh /workspace/scratch/before > /workspace/scratch/apply-before.log 2>&1
# Includes 99a and 99b.
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/upstream bash stage3/apply.sh /workspace/scratch/after > /workspace/scratch/apply-after.log 2>&1
go build -o /workspace/scratch/adamic ./cmd/adamic > /workspace/scratch/build-adamic.log 2>&1
NODE_PATH="$PWD/stage3/api/node_modules" node stage3/adapt/99-correctness-axolotl-b/witness.cjs > /workspace/scratch/witness.json 2>&1
python3 /workspace/scratch/check99b.py > /workspace/scratch/check99b.log 2>&1
NODE_PATH="$PWD/stage3/api/node_modules" node /workspace/scratch/verify99b.cjs > /workspace/scratch/verify99b.log 2>&1
```

An early checker build before setup finished failed because cohere/TypeScript/tsc/go.mod had not yet been checked out. The listed build was rerun with the submodule available. The Node witnesses were also run directly through `/workspace/adamic-tools/typescript/lib/tsc.js`, in isolated scratch directories: the missing numeric argument exited 1, and `--init --outFile null` exited 0 and wrote the undefined property into tsconfig.json. CLI logs and generated configuration are under /workspace/scratch/cli-number and /workspace/scratch/cli-init.

The whole-program checker commands use sorted roots and argument arrays, as in 99a:

```python
import pathlib, subprocess
for label in ['before', 'after']:
    tree = pathlib.Path('/workspace/scratch', label)
    roots = sorted(str(p) for p in (tree / 'src/compiler').rglob('*')
                   if p.suffix in ['.ts', '.a'])
    with open(f'/workspace/scratch/types-{label}.stdout', 'w') as out, \
         open(f'/workspace/scratch/types-{label}.stderr', 'w') as err:
        result = subprocess.run(['/workspace/scratch/adamic', 'types', *roots],
                                stdout=out, stderr=err)
    assert result.returncode == 1
```

The comparison retains complete diagnostic strings and elaboration chains, normalizing only the before/after tree root paths, and compares exact multisets across every checker code. measurement99b.json and diagnostics-before.json/diagnostics-after.json retain those results. verify99b.json retains source and transpileModule byte equality, idempotence, refusal checks, and equality of the baseline's 99a-after snapshot with the final tree's 99b-before snapshot. The refusal fixtures introduce either assertion alone, both assertions, occurrence-count drift, and an assignment target; each must throw without changing input bytes.
