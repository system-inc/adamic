# Next stops after the core.ts wave

Task b5w3ycg. Report only, not for landing.

Main/compiler/apply SHA: `af42d3dc1261a3db536064c3dfbb5eaa7e8b6405`. TypeScript 6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`.

Both requested runs use the same completed full apply. No ordinary compiler overlay, adaptation subset or partial-root fallback is used.

## Grouped first stops

Roots are relative to `src/compiler/`. Every one of the 79 sorted roots is retained.

| Count | Location | Kind | Exact reason | Root names |
| ---: | --- | --- | --- | --- |
| 73 | `core.ts:2116:26` | Refused | a method read as a value (compare would lose its object, and this with it) | `_namespaces/ts.moduleSpecifiers.ts`, `_namespaces/ts.performance.ts`, `_namespaces/ts.ts`, `binder.ts`, `builder.ts`, `builderPublic.ts`, `builderState.ts`, `builderStatePublic.ts`, `checker.ts`, `commandLineParser.ts`, `emitter.ts`, `executeCommandLine.ts`, `expressionToTypeNode.ts`, `factory/baseNodeFactory.ts`, `factory/emitHelpers.ts`, `factory/emitNode.ts`, `factory/nodeChildren.ts`, `factory/nodeConverters.ts`, `factory/nodeFactory.ts`, `factory/nodeTests.ts`, `factory/parenthesizerRules.ts`, `factory/utilities.ts`, `factory/utilitiesPublic.ts`, `moduleNameResolver.ts`, `moduleSpecifiers.ts`, `parser.ts`, `path.ts`, `performance.ts`, `performanceCore.ts`, `program.ts`, `programDiagnostics.ts`, `resolutionCache.ts`, `scanner.ts`, `semver.ts`, `sourcemap.ts`, `symbolWalker.ts`, `sys.ts`, `tracing.ts`, `transformer.ts`, `transformers/classFields.ts`, `transformers/classThis.ts`, `transformers/declarations/diagnostics.ts`, `transformers/declarations.ts`, `transformers/destructuring.ts`, `transformers/es2015.ts`, `transformers/es2016.ts`, `transformers/es2017.ts`, `transformers/es2018.ts`, `transformers/es2019.ts`, `transformers/es2020.ts`, `transformers/es2021.ts`, `transformers/esDecorators.ts`, `transformers/esnext.ts`, `transformers/generators.ts`, `transformers/jsx.ts`, `transformers/legacyDecorators.ts`, `transformers/module/esnextAnd2015.ts`, `transformers/module/impliedNodeFormatDependent.ts`, `transformers/module/module.ts`, `transformers/module/system.ts`, `transformers/namedEvaluation.ts`, `transformers/taggedTemplate.ts`, `transformers/ts.ts`, `transformers/typeSerializer.ts`, `transformers/utilities.ts`, `tsbuild.ts`, `tsbuildPublic.ts`, `utilities.ts`, `utilitiesPublic.ts`, `visitorPublic.ts`, `watch.ts`, `watchPublic.ts`, `watchUtilities.ts` |
| 2 | `-` | accepted | - | `corePublic.ts`, `hostErrors.ts` |
| 1 | `commandLineParser.ts:293:35` | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | `diagnosticInformationMap.generated.ts` (artifact classification incomplete) |
| 1 | `debug.ts:243:59` | NotYet | a generic instantiation whose type argument for T, constrained to Node, can't be read | `core.ts` |
| 1 | `diagnosticInformationMap.generated.ts:13:45` | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | `types.ts` (artifact classification incomplete) |
| 1 | `semver.ts:71:9` | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | `debug.ts` (artifact classification incomplete) |

## Whole-root lowering

2/79 roots lower whole in the ordinary independent-root census: `corePublic.ts`, `hostErrors.ts`.
This credits ordinary Lower acceptance of the entire resolved root program. Native emission and a multi-root aggregate are not measured.

## Comparison

The 37f8314b report had 73 roots at core.ts:1049:34, Refused: a value of type never seen as T, a type parameter whose constraint any can be written, so it can write what never can't hold. The 4186581e state had 73 roots at core.ts:1326:26, getOwnValues, Refused: a cast the runtime can't check.

- 73 root(s) at `core.ts:2116:26` (Refused): 73 changed from the 37f8314b table; 0 unchanged.
- 2 root(s) at `-` (accepted): 0 changed from the 37f8314b table; 2 unchanged.
- 1 root(s) at `commandLineParser.ts:293:35` (NotYet): 0 changed from the 37f8314b table; 1 unchanged.
- 1 root(s) at `debug.ts:243:59` (NotYet): 0 changed from the 37f8314b table; 1 unchanged.
- 1 root(s) at `diagnosticInformationMap.generated.ts:13:45` (NotYet): 0 changed from the 37f8314b table; 1 unchanged.
- 1 root(s) at `semver.ts:71:9` (NotYet): 0 changed from the 37f8314b table; 1 unchanged.

`comparison.json` retains the exact previous and current tuple for each root. `previous-REPORT.md` and `previous-grouped-first-stops.json` retain the comparison baseline.

## Every new first-stop group: source and decision

### core.ts:2116:26

Exact reason: a method read as a value (compare would lose its object, and this with it).

Exact adapted source, `core.ts-2116-26.txt`. The whole enclosing function follows.

```ts
    function createIntlCollatorStringComparer(locale: string | undefined): Comparer<string> {
        // Intl.Collator.prototype.compare is bound to the collator. See NOTE in
        // http://www.ecma-international.org/ecma-402/2.0/#sec-Intl.Collator.prototype.compare
        const comparer = new Intl.Collator(locale, { usage: "sort", sensitivity: "variant", numeric: true }).compare;
        return (a, b) => compareWithCallback(a, b, comparer);
    }
```

Emitter: `internal/lower/refusals.go:282`, `(*lowering).methodValueRefusal`. The deciding function starts at line 260. Its guard at line 261 selects a property access read outside a direct call, without an allowed method-read proof. The method-symbol check at line 269 selects this refusal, emitted at line 282. The enclosing source comments that Intl.Collator.prototype.compare is bound; this report records the current compiler decision without changing it.

## Root-order artifacts

STOPPING REASON: The bounded entry census did not complete; the meter requires its attempted-source and finding ledger before classifying root-order artifacts.

No artifact label or whole-root lowering credit is inferred from an incomplete entry probe. The ordinary 79-root census is complete. `root-order.json` retains the pending roots and the child's actual exit code.
Entry child exit -15 after 1800.756 seconds: configured elapsed-time bound. Last completed timing: `/workspace/after-wave-adapted/src/compiler/utilities.ts:10805:1`. The interrupted active unit was not recorded. Partial stream records: 1. See ENTRY-STOP.json. `entry-probe.tar.gz` retains the exact stopped checkpoint bundle, with its own SHA-256 file.

- `debug.ts`: artifact classification incomplete; Node child exit 1, TypeError: Cannot read properties of undefined (reading 'assert').
- `diagnosticInformationMap.generated.ts`: artifact classification incomplete; Node child exit 1, TypeError: Cannot read properties of undefined (reading 'AbstractKeyword').
- `types.ts`: artifact classification incomplete; Node child exit 1, TypeError: Cannot read properties of undefined (reading 'Error').

## Method, limits and evidence

The ordinary census was built from the pinned main compiler. Four root groups contain 20, 20, 20 and 19 roots. Each root invocation emits the independent observation and an identical single-root aggregate. The runner verifies all non-timing fields agree and counts the first observation only. Non-TypeScript rows and a 79-root aggregate are excluded.

The separate entry probe uses the meter guarded entry overlay with LATENT_FULL=1 and LATENT_ASSERT_NO_OUTPUT=1. Physical scratch Go files use .go.txt; virtual overlay paths preserve the original compiler paths. The probe has an 1800-second bound. `supervise.py` records actual child exit codes, enforces 30 minutes without output, kills the entire child process group and reaps the direct child.

Completed apply and each completed group were pushed to the append-only checkpoint branch. The final branch contains one report commit on main. `CHECKPOINT.json` records the last checkpoint. No force push or main push was used.

Raw ordinary/entry streams, progress, timing and preparation logs are packed in `raw-streams.tar.gz`; verify with `sha256sum -c raw-streams.tar.gz.sha256`. Summaries, provenance, source/product manifests, emitter evidence and validation remain directly readable. The core.ts report is in `../core-sites-af42d3dc/REPORT.md`.

## Reproduction

```sh
source /workspace/adamic-tools/env.sh
bash cloud/setup.sh > setup.log 2>&1
npm ci --prefix stage3/api > npm-ci.log 2>&1
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source
python3 stage3/census/next-stops-af42d3dc/supervise.py apply bash stage3/apply.sh --build-product /workspace/after-wave-adapted > apply-runner.log 2>&1
python3 stage3/census/next-stops-af42d3dc/prepare-tools.py > prepare-tools.log 2>&1
python3 stage3/census/core-sites-af42d3dc/run-census.py > core-runner.log 2>&1
python3 stage3/census/next-stops-af42d3dc/run-groups.py > groups.log 2>&1
python3 stage3/census/next-stops-af42d3dc/run-entry.py > entry-runner.log 2>&1
```
