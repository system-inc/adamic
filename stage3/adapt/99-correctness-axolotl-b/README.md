# 99b: explicit numeric argument conversion

JavaScript-changing in text only, following adaptation 97a on cli/axolotl-97a. D031 changes `parseInt(args[i])` to `parseInt(String(args[i]))` in parseOptionValue. Behavior and public APIs are unchanged. D036 remains unasserted and validated; its fork repair is proposed below, pending a ruling for a separate 99g head.

lane: run as part of Correctness's wave

Base: main c74051ffc5ee94f4b7f732b1532cb219032b183d. Source: stock TypeScript 6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8. This follows the initial decline commit 3fd68da2624a895405d4d9466f8680f75725c6fd. The measurement tree includes 99a (74dda102), as in the original proof. No lane manifest changes.

## Sites and behavior

| Site in adapted commandLineParser.ts | Treatment | Evidence |
| --- | --- | --- |
| 2077:79 TS2345, D031 | Explicit String conversion; no assertion | parseStrings bounds the flag index, then increments i before calling parseOptionValue. A missing argument reaches parseInt with undefined after recording TS6044. parseInt performs ToString internally. For string or undefined inputs, String produces exactly that same string; parseInt then parses it unchanged. The read is still evaluated once, at the same point. |
| 2957:13 TS2322, D036 | Unasserted, unchanged | `--init --outFile null` creates an own outFile property with value undefined. The remaining-options loop passes undefined as both options[setting] and defaultValue. The scalar formatSingleValue call reaches JSON.stringify(undefined), returning undefined. The stock compiler writes `"outFile": undefined,`. |

Stock source witnesses: parseStrings at 1978-1993 increments the index before both parser calls; parseOptionValue at 2070-2078 reports the missing argument and continues into numeric conversion. The `"null"` test at 2074 instead routes to 2111-2112, storing undefined and advancing the index without numeric conversion. generateTSConfig gathers own keys at 2838, passes remaining option values as defaults at 2901, and computes options[setting] ?? defaultValue at 2933. Its two formatValueOrArray callers are at 2935 and 2938. The scalar branch at 2952 calls formatSingleValue; JSON.stringify is at 2960. executeCommandLine.ts:577-584 allows the zero-error init command, and 1300 writes the generated text.

For the allowed string | undefined domain, there is no additional user conversion or observable read. A present string is unchanged; undefined becomes "undefined" and parses as NaN in both forms. As with stock tsc, this assumes the standard built-in String and parseInt implementations.

## Adapter contract

Stock TypeScript 6.0.3 parses commandLineParser.ts. Both sites are pinned by unique named function, AST kind, exact expression text, occurrence ordinal, total count and exact parent text. The numeric read is occurrence 4 of 8 args[i] expressions in parseOptionValue. Its single parseInt call and entire assignment statement must match the original or converted spelling exactly. The scalar call is the sole exact formatSingleValue(value, map) in formatValueOrArray, with its original return statement. Stores, asserted reads, double conversions and other shapes are refused.

Both sites validate before the edit is planned. The planned result is parsed and validated again as fully adapted, its exact length change is checked, and the input is reread before one write. An already converted input is accepted without a write. D036's decline validation remains mandatory before and after conversion. No assertion is introduced.

## Witnesses and measurement

The witness loads the stock TypeScript library twice, replacing exactly its one inventoried numeric call for the second copy. It verifies that the supplied before and after source files differ by exactly that call, then compares complete parse results including options and diagnostics. The real option definitions and validators run in both copies.

| Numeric argument | Result before and after | Diagnostic codes |
| --- | --- | --- |
| Missing | NaN | 6044 |
| `12` | 12 | none |
| `12.9` | 12 | none |
| `0x10` | 16 | none |
| `abc` | NaN | none |
| `null` | undefined, numeric branch bypassed | none |

The direct two parseInt forms also both produce NaN for "null"; the actual parser does not send that token to either conversion. The unchanged D036 witness still generates the invalid undefined property with zero parse errors.

| Measurement | Before | After |
| --- | ---: | ---: |
| Sorted compiler roots | 79 | 79 |
| Whole program, every checker code | 129 | 128 |
| Requested diagnostics | 2 | 1 |
| commandLineParser.ts transpileModule JavaScript bytes | 161985 | 161993 |

Exactly D031 at 2077:79 disappears. No diagnostic is added anywhere. The comparison uses complete diagnostic multisets with columns, codes, messages, multiplicity and elaboration chains, normalizing only tree root paths. D036 remains at 2957:13. Stock ts.transpileModule (ES2024, ESNext, comments retained) differs only by `parseInt(args[i])` becoming `parseInt(String(args[i]))`, eight added bytes. A second run returns zero files and zero edits and preserves bytes and mtime. Seven refusal fixtures pass without writes: numeric assertion, declined assertion on original and converted inputs, count drift, assignment target, double conversion and changed enclosing assignment context.

The original before and after apply.sh trees passed end to end with 99a and the decline-only 99b. For this repair, /workspace/scratch/after is the unchanged 129-diagnostic baseline from that proof. A complete copy at /workspace/scratch/99b-string-after receives only this adapter's one edit. Both whole-program invocations were rerun using sorted .ts/.a roots recursively under src/compiler and `/workspace/scratch/adamic types`, exactly as in 99a's command record. No checker build, checker option or source overlay changed. The prior decline-only measurements remain in git history.

Material commands, with logs in files:

```sh
source /workspace/adamic-tools/env.sh
cp -a /workspace/scratch/after /workspace/scratch/99b-string-after
NODE_PATH="$PWD/stage3/api/node_modules" node stage3/adapt/99-correctness-axolotl-b/adapt.cjs /workspace/scratch/99b-string-after > /workspace/scratch/99b-string-adapt.log 2>&1
NODE_PATH="$PWD/stage3/api/node_modules" node stage3/adapt/99-correctness-axolotl-b/witness.cjs /workspace/scratch/after/src/compiler/commandLineParser.ts /workspace/scratch/99b-string-after/src/compiler/commandLineParser.ts > /workspace/scratch/99b-string-witness.json 2>&1
NODE_PATH="$PWD/stage3/api/node_modules" node /workspace/scratch/verify99b-string.cjs > /workspace/scratch/99b-string-verify.log 2>&1
```

99b-string-types-before.stdout/stderr and 99b-string-types-after.stdout/stderr contain both complete checker runs. 99b-string-measurement.json records the counts and the full removed diagnostic. 99b-string-verify.json records the exact JavaScript comparison, no-write second run and refusal checks.

## D036 research and proposed 99g repair, not implemented

Pinned typescript-go: system-inc/TypeScript commit d92d9bfee114c80be2c375d72edae966176e3a4f, recorded by the cohere submodule. Its existing CLI, built from tsc/cmd/tsc, runs `--init --outFile null` with exit 0 and writes `"outFile": null,`. The generated config has zero syntax diagnostics with stock TypeScript's JSONC parser. The Go parser stores nil for a null token (tsc/internal/tsoptions/commandlineparser.go:330-332), and generateTSConfig retains it: init.go:132-135 distinguishes missing map keys from present nil; formatSingleValue serializes nil with json.MarshalIndent at 68-72. This is valid config, differing from 6.0.3's undefined token.

Upstream TypeScript main was inspected at [6161772c1370bdd4a82f808e6823bd755eb1870d](https://github.com/microsoft/TypeScript/commit/6161772c1370bdd4a82f808e6823bd755eb1870d), dated 2026-10-30. Main now contains the native Go compiler. Its [active init.go](https://github.com/microsoft/TypeScript/blob/6161772c1370bdd4a82f808e6823bd755eb1870d/tsc/internal/execute/tsc/init.go) is byte-identical to the pinned Go version above. The retained [TypeScript fixture](https://github.com/microsoft/TypeScript/blob/6161772c1370bdd4a82f808e6823bd755eb1870d/tsc/testdata/fixtures/compiler/commandLineParser.ts) has a generateTSConfig section, including emitOption, identical to 6.0.3. Thus the retained TypeScript implementation has no fix for this bug at the inspected main commit; the active Go implementation writes null.

Proposed fork policy: omit an explicitly cleared option. After removing its key from allSetOptions, return before formatting when options has that own property and its value is undefined. Placing the guard after splice ensures the remaining-options loop advances. The own-property test preserves normal default emission for absent options. Defined values and explicit API null values retain existing output. No new placeholder string, cast or required-value assertion is needed.

Proposed source diff inside generateTSConfig's emitOption, against stock 6.0.3:

```diff
         if (existingOptionIndex >= 0) {
             allSetOptions.splice(existingOptionIndex, 1);
         }

+        if (hasProperty(options, setting) && options[setting] === undefined) {
+            return;
+        }
+
         let comment: boolean;
```

For `--init --outFile null`, outFile would be omitted. This also skips own undefined values for preset options, while preserving presets for options absent from the input object. It changes output only for own options whose value is undefined. This diff has not been applied or built. If ruled in, 99g would implement it with an entry in stage3/upstream/LEDGER.md documenting the stock defect, the pinned Go behavior, upstream snapshot and changed output. No ledger entry claims a fix yet.

Research logs: /workspace/scratch/build-pinned-tsgo.log, pinned-tsgo-init/stdout.log and stderr.log, its generated tsconfig.json, upstream-main-fetch.log, and 99g-research.json. The upstream read used a depth-1 main fetch; an initial unrestricted-depth fetch was stopped to avoid downloading unnecessary history. No 99g source has been changed.
