# getPathFromPathComponents refusal measurement

Measured on Adamic main **`b1b79cd0f67c9e9b8f22228343d6c9e629aa0076`**, the fetched main tip when this worktree was created (2026-11-12). Upstream is TypeScript **6.0.3**, `050880ce59e30b356b686bd3144efe24f875ebc8`. The report branch is `typescript/path-from-components`. Nothing in `stage3/adapt` was changed.

**Result:** 15 TypeScript call expressions, with **13 instantiations of `T = string` and 2 of `T = Path`**. The function is internal, absent from both requested public API inventories. The scratch overload candidate checks all 79 compiler roots with zero stock diagnostics, emits byte-identical JavaScript for all 79 roots, and has **no latent finding at either overload or the implementation** on this main. Main already serves this primitive-brand result; an overload-result refusal does not occur.

## Inputs and method

The worktree is `/workspace/path-from-components`. `stage3/apply.sh /workspace/scratch/path-components-adapted` produced the adapted tree using `ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3`. The first attempt stopped with child exit 1 because that pinned checkout was missing; after cloning the exact tag and installing locked parser dependencies, apply completed with child exit 0. [apply.log](apply.log) records the successful run. The product key was `f3734246b52c49105a61c857ad540963f5ae7d11a9bf4f0f92ea4ab884a26dc6`. Final adapted `src/compiler/path.ts` SHA-256 is `4cf2621fcd2d659c0e7561d9177d0ff10f51e85b6357d91cfd30601ac5f29938`.

[check.cjs](check.cjs) uses stock TypeScript 6.0.3's checker on all **603 adapted `src/**/*.ts` roots**, including compiler, services, server, harness and testRunner. For every call it resolves the signature and matches its declaration to `src/compiler/path.ts`'s `getPathFromPathComponents`, thereby including aliases. Since the declaration returns exactly `T`, the resolved signature return type gives the inferred `T`; argument types come from `getTypeAtLocation`. Strict checking is enabled. These are checker types, not deductions from variable names. [calls.json](calls.json) preserves the calls and counts.

[reference-searches.json](reference-searches.json) records the whole-checkout name/alias search, including ignored files, excluding `.git` and installed dependencies. There are no direct-name or `vpath.format` hits in `tests`, `src/testRunner`, `src/services`, or `src/server` (grep child exit 1, meaning no matches). The checker walk also finds no alias calls there. The JSDoc example at `path.ts:507` is not a call expression. The harness's `export import format = ts.getPathFromPathComponents` at `vpathUtil.ts:15` is a reference, not a call.

## Every adapted TypeScript call

Positions below are in the **final adapted tree**, not pristine upstream. Columns distinguish multiple calls on one line. `Readonly<PathPathComponents>` is the readonly form of the branded `Path[]` intersection; there are no plain `Path[]` arguments at these calls.

| File:line:column | Inferred T | Checker argument type / classification | Call |
|---|---|---|---|
| `src/compiler/path.ts:1052:12` | `string` | `string[]` | `getPathFromPathComponents(pathComponents)` |
| `src/compiler/path.ts:1082:12` | `string` | `string[]` | `getPathFromPathComponents(pathComponents)` |
| `src/compiler/utilities.ts:9573:27` | `string` | `string[]` | `getPathFromPathComponents(aParts)` |
| `src/compiler/utilities.ts:9573:62` | `string` | `string[]` | `getPathFromPathComponents(bParts)` |
| `src/compiler/moduleNameResolver.ts:2931:43` | `string` | `string[]` | `getPathFromPathComponents(parts)` |
| `src/compiler/program.ts:385:12` | `string` | `string[]` | `getPathFromPathComponents(commonPathComponents)` |
| `src/compiler/resolutionCache.ts:459:14` | `string` | `readonly string[]` | `getPathFromPathComponents(dirComponents, length)` |
| `src/compiler/resolutionCache.ts:460:18` | `Path` | `Readonly<PathPathComponents>` | `getPathFromPathComponents(dirPathComponents, length)` |
| `src/compiler/resolutionCache.ts:462:54` | `string` | `readonly string[]` | `getPathFromPathComponents(dirComponents, packageDirLength)` |
| `src/compiler/resolutionCache.ts:463:58` | `Path` | `Readonly<PathPathComponents>` | `getPathFromPathComponents(dirPathComponents, packageDirLength)` |
| `src/harness/vpathUtil.ts:109:48` | `string` | `string[]` | `format(reduce(components))` |
| `src/harness/vpathUtil.ts:109:83` | `string` | `string[]` | `format(reduce(components))` |
| `src/harness/vfsUtil.ts:1065:36` | `string` | `string[]` | `vpath.format(components)` |
| `src/harness/vfsUtil.ts:1072:33` | `string` | `string[]` | `vpath.format(components.slice(0, step))` |
| `src/harness/vfsUtil.ts:1094:30` | `string` | `string[]` | `vpath.format(components.slice(0, step + 1))` |

| Distinct T | Call count |
|---|---:|
| `string` | 13 |
| `Path` | 2 |
| **Total** | **15** |

The argument breakdown is 11 `string[]`, 2 `readonly string[]`, 2 `Readonly<PathPathComponents>`, and 0 plain `Path[]`. By source area: compiler 10, harness 5, services/server/tests 0. The two `Path` instantiations are both in `resolutionCache.ts`.

## Prebuilt JavaScript copies elsewhere in the checkout

The whole-checkout search additionally finds **20 calls in upstream's prebuilt `lib/_tsc.js` and `lib/typescript.js`**. These call their own bundled, nongeneric JavaScript copies, not the adapted TypeScript declaration. Thus **T is not applicable** to these calls and they are excluded from the 15 generic instantiations above. [bundles.cjs](bundles.cjs) separately runs the stock checker with `allowJs` over these copies; [bundle-calls.json](bundle-calls.json) retains each expression and checker argument type. The checker reports return type `any` for these unannotated copies; these do not supply additional branded-type evidence.

| File:line:column | T | Checker argument type | Call |
|---|---|---|---|
| `lib/_tsc.js:5719:10` | N/A (nongeneric JS) | `any[]` | `getPathFromPathComponents(pathComponents2)` |
| `lib/_tsc.js:5746:10` | N/A (nongeneric JS) | `any[]` | `getPathFromPathComponents(pathComponents2)` |
| `lib/_tsc.js:18392:25` | N/A (nongeneric JS) | `any[]` | `getPathFromPathComponents(aParts)` |
| `lib/_tsc.js:18392:60` | N/A (nongeneric JS) | `any[]` | `getPathFromPathComponents(bParts)` |
| `lib/_tsc.js:41824:32` | N/A (nongeneric JS) | `any[]` | `getPathFromPathComponents(parts)` |
| `lib/_tsc.js:121938:10` | N/A (nongeneric JS) | `never` | `getPathFromPathComponents(commonPathComponents)` |
| `lib/_tsc.js:128061:10` | N/A (nongeneric JS) | `any` | `getPathFromPathComponents(dirComponents, length2)` |
| `lib/_tsc.js:128062:14` | N/A (nongeneric JS) | `any` | `getPathFromPathComponents(dirPathComponents, length2)` |
| `lib/_tsc.js:128064:47` | N/A (nongeneric JS) | `any` | `getPathFromPathComponents(dirComponents, packageDirLength)` |
| `lib/_tsc.js:128065:51` | N/A (nongeneric JS) | `any` | `getPathFromPathComponents(dirPathComponents, packageDirLength)` |
| `lib/typescript.js:9125:10` | N/A (nongeneric JS) | `any[]` | `getPathFromPathComponents(pathComponents2)` |
| `lib/typescript.js:9152:10` | N/A (nongeneric JS) | `any[]` | `getPathFromPathComponents(pathComponents2)` |
| `lib/typescript.js:22367:25` | N/A (nongeneric JS) | `any[]` | `getPathFromPathComponents(aParts)` |
| `lib/typescript.js:22367:60` | N/A (nongeneric JS) | `any[]` | `getPathFromPathComponents(bParts)` |
| `lib/typescript.js:46340:32` | N/A (nongeneric JS) | `any[]` | `getPathFromPathComponents(parts)` |
| `lib/typescript.js:126778:10` | N/A (nongeneric JS) | `never` | `getPathFromPathComponents(commonPathComponents)` |
| `lib/typescript.js:132996:10` | N/A (nongeneric JS) | `any` | `getPathFromPathComponents(dirComponents, length2)` |
| `lib/typescript.js:132997:14` | N/A (nongeneric JS) | `any` | `getPathFromPathComponents(dirPathComponents, length2)` |
| `lib/typescript.js:132999:47` | N/A (nongeneric JS) | `any` | `getPathFromPathComponents(dirComponents, packageDirLength)` |
| `lib/typescript.js:133000:51` | N/A (nongeneric JS) | `any` | `getPathFromPathComponents(dirPathComponents, packageDirLength)` |

## Public API status

`getPathFromPathComponents` is exported from the internal compiler module and reachable through the internal `ts` namespace, but its declaration is marked `@internal` (`path.ts:510`). It is **not in the public TypeScript declaration API**:

- TypeScript 6.0.3 `tests/baselines/reference/api/typescript.d.ts`: no occurrence.
- Adamic main `stage3/lane/sanctioned-api.json`: no occurrence.

Both exact-name searches returned child exit 1 (no match); commands, paths and outputs are in [reference-searches.json](reference-searches.json). Runtime exposure in `lib/typescript.js` does not make it a declared public API.

## Exact refusal family in path.ts

The existing main latent overlay was generated with `stage3/census/latent/make_overlay.py` and built against this worktree, without production compiler edits. Generated physical Go evidence paths were renamed to `.go.txt` and their overlay mappings updated before building. The measurement binary was `/workspace/scratch/path-components-latent`.

The baseline invocation was:

```sh
LATENT_FULL=1 \
LATENT_ONLY_FILE=/workspace/scratch/path-components-adapted/src/compiler/path.ts \
LATENT_ASSERT_NO_OUTPUT=1 \
/workspace/scratch/path-components-latent \
  /workspace/scratch/path-components-adapted/src/compiler \
  /workspace/path-from-components/stage3/census/path-from-components/baseline-census.jsonl
```

This is **full mode with a single selected file**, as in the referenced `clear-census.py`; selection does not shrink the loaded program. The child exited 0, output guards passed, and the header reports zero checker diagnostics (`checker_rejected: false`). The recorder's fixed measurement label still says “measured on a checker-rejected program”; the actual header is authoritative about this run. It attempted 81 declaration units and recorded 102 findings of all kinds, including boundaries and repeated dependency attempts. Those totals are not a count of cast sites.

Deduplicating the exact refusal reason by location yields these **three sites**, all in `refusal_scan`:

| Adapted path.ts line:column | Exact reason | Exact expression |
|---|---|---|
| `513:45` | `a cast the runtime can't check` | `"" as T` |
| `516:12` | `a cast the runtime can't check` | `root + pathComponents.slice(1, length).join(directorySeparator) as T` |
| `1103:28` | `a cast the runtime can't check` | `getDirectoryPath(directory) as P` |

The first two belong to `getPathFromPathComponents`; the third belongs to `forEachAncestorDirectory` (implementation at line 1096). There are no other sites with that exact reason. In particular, `toPath`'s line 766 `getCanonicalFileName(nonCanonicalizedPath) as Path` is **not refused with this reason** on the measured main.

Lowering attempts also repeat the first two sites with more specific reasons: `an unproven relation from "" to T: the source is not assignable to the target` and `an unproven relation from string to T: the source is not assignable to the target`. These are retained separately in the raw evidence, not counted as new sites of the requested exact reason.

See [baseline-census.jsonl](baseline-census.jsonl), [baseline-summary.json](baseline-summary.json), and [baseline-census-status.json](baseline-census-status.json). Exact error text, repair text, attempted unit and phase are retained. [analyze.cjs](analyze.cjs) extracts the expressions from the stock AST at the recorded positions. The production refusal construction is preserved in [cast_proof.go.txt](cast_proof.go.txt), line 195.

## Adaptation pins and owners

[adaptation-grep.txt](adaptation-grep.txt) is the complete `rg -n 'getPathFromPathComponents|path\.ts' stage3/adapt` result. Most matches are historical evidence paths, source hashes, or recorded census results. [path-adaptation-diffs.txt](path-adaptation-diffs.txt) independently compares every adaptation's before/after tree ref for `src/compiler/path.ts`; exactly these three adaptations actually change the file:

- `10-type-imports`: changes `GetCanonicalFileName` and `Path` imports to type imports. It selects owners through the checker rather than a literal `path.ts` filename pin.
- `30-indexed-reads`: explicitly includes `path.ts` in `adapt.cjs:12`. `sites.json:3689–3699` pins owner `getPathFromPathComponents`, expression `pathComponents[0]`, occurrence 1, total 4, action **decline**, class `U-endpoint`. Its invariant is: “The original truthiness test handles an empty or absent root, and the repeated guarded read already narrows; keep both reads.” This is the direct named function pin. The same ledger pins other path owners, and this adaptation adds eight assertions at lines 540, 912 (two), 986 (two), 1016, 1017 and 1076. Neither cast in the target function is rewritten.
- `40-explicit-any`: `class-rules.json` rule **103** (`:65` file entry) owns `PathPathComponents.__pathComponensBrand` at upstream line 457, replacing `any` with `undefined`. The brand audit/proof evidence also pins this file. The separate `Path.__pathBrand` in `types.ts` is `undefined` in the adapted program.

`41-explicit-any-remaining` has path.ts hashes, timer/census and any-proof evidence, but its measured before/after refs do **not** change path.ts. The multimap `owners.json` hits for the target name are copies of a utilities import block, not target-function ownership. Other grep hits in correctness proof results are root lists or historical observations. No adaptation rewrites `getPathFromPathComponents`'s signature or casts in the measured tree.

## Scratch overload experiment

The only source difference in `/workspace/scratch/path-components-candidate` is this function's overloads and removal of its two `as T` assertions:

```ts
export function getPathFromPathComponents(pathComponents: readonly Path[], length?: number): Path;
export function getPathFromPathComponents(pathComponents: readonly string[], length?: number): string;
export function getPathFromPathComponents(pathComponents: readonly string[], length?: number): string {
    if (pathComponents.length === 0) return "";

    const root = pathComponents[0] && ensureTrailingDirectorySeparator(pathComponents[0]);
    return root + pathComponents.slice(1, length).join(directorySeparator);
}
```

Each overload retains `@internal`. The overloads are placed **before the existing `//// Path Formatting` section comment and JSDoc**, leaving those original comments on the implementation. This matters to a byte comparison: putting overloads immediately after the comments makes stock emit discard comments attached to an erased overload. The checked-in scratch source is [candidate-path.ts.txt](candidate-path.ts.txt).

[candidate.cjs](candidate.cjs) runs stock TypeScript 6.0.3 with the adapted `src/compiler/tsconfig.json` options and its exact **79 roots**, using an in-memory replacement of path.ts. Both baseline and candidate have **zero pre-emit diagnostics**. This includes the project configuration's declaration/isolated-declaration checks. Locked upstream dependencies provide Node and source-map-support declarations; no stubs suppress diagnostics.

For JavaScript emission, the same 79 roots are emitted with declaration-only/composite/isolated-declaration/incremental settings disabled and source/declaration maps disabled. Runtime-relevant project options remain. All **79 JavaScript file names and SHA-256 values match**, including path.js. Thus the JavaScript is **byte-identical**, not merely equivalent after removing types or comments. Source maps and declarations are not part of that claim. [candidate-stock.json](candidate-stock.json) records options, zero-diagnostic arrays and both complete JS hash ledgers; [baseline-path.js.txt](baseline-path.js.txt) and [candidate-path.js.txt](candidate-path.js.txt) preserve the exact path.js bytes.

The final scratch source copy was also passed to the same full-mode single-file census, using `LATENT_ONLY_FILE=/workspace/scratch/path-components-candidate/src/compiler/path.ts`. Child exit was 0; output guards passed; the loaded program had zero checker diagnostics. Both overload units (lines 500 and 501) and the implementation (line 514) were attempted. There are **zero findings attributed to their declaration spans or units**: no cast refusal, overload-result refusal, NotYet, or boundary at this function. The remaining same-reason family is only `getDirectoryPath(directory) as P`, shifted to line 1105.

This agrees with main's existing primitive phantom-brand handling: `census_small.go:242` admits an overload result when the primitive phantom relation works in both directions; `phantom_brands.go:51` strips phantom brands while comparing primitive types. Source evidence is [census_small.go.txt](census_small.go.txt) and [phantom_brands.go.txt](phantom_brands.go.txt). The candidate therefore does not require a new primitive-brand optimizer/lowering feature on this main. This experiment is evidence for choosing an owner; no adaptation was shipped. It does not establish that the other generic `P` cast is solved, or that all 73 tsc entry roots now lower successfully.

See [candidate-summary.json](candidate-summary.json), [candidate-census.jsonl](candidate-census.jsonl) and [candidate-census-status.json](candidate-census-status.json).

## Exit codes and time bounds

All final measurement children (call checker, JS bundle checker, stock overload check, baseline census and candidate census) exited **0**. Each [run.py](run.py) invocation records its child's own exit code and elapsed time in the corresponding `*-status.json`; wrapper shells print their returned status too. `run.py` bounds a child at 29 minutes, terminates/reaps its process group, and exits with the child's code (or its signal-derived shell status). Setup/build commands used explicit limits of at most 1750 seconds. No process was allowed to run for 30 minutes. No time bound was reached. API absence searches intentionally returned grep exit 1, not a compiler failure.

To reproduce, use the recorded main and TypeScript pins, install the locked stage3 API and upstream dependencies, apply into the recorded adapted scratch path, generate/build the existing latent overlay with physical Go filenames ending `.go.txt`, and run the scripts/commands above. Scripts record the absolute paths used in this worktree; adjust those paths together when relocating. [provenance.json](provenance.json) records pins and source identity. All committed changes are contained in this report directory.
