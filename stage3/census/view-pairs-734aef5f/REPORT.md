# TypeScript view-pairs census

Task vqzty8t. Measurement only. Counts describe observed latent unit lowerings, not accepted whole programs.

Box `755daee0c6ba`, CPU quota `400000 100000`, memory limit `17179869184` bytes. Source/compiler base `734aef5f3a28fdc6b21d0decb05d6338e78a027b`. Probe `86157bea8c6eafd207a3b518e1ec817fed52c072`. TypeScript 6.0.3 pin `050880ce59e30b356b686bd3144efe24f875ebc8`.

## Base and probe

The base is 734aef5f, the main parent before Memory's first throwaway probe commit 8bedbc18. The named tip 86157bea fixes that probe's program-name conversion. Its immediate parent already contains the probe, so it is not used as a production base. bcbe5d12 has different adaptation inputs; this run builds the keyed adapted product at 734aef5f rather than reusing bcbe's adapted bytes. apply-comparison.json records both input keys and the changed adapted source files. Prepared Go/cohere and pinned parser dependencies were reused.

Memory's original probe is preserved as memory-probe-original.go.txt and copied into scratch. The scratch context version calls logViewPairs() at every unit end beside the boundary observer, using only &cycleFinder{l:l, where:map[*checker.Type]*ast.Node{}}. findCycles never runs. Original pair traversal, per-unit checker-pointer seen set, first encountered path and depth bound 6 are preserved. The original program field names the entry; file and unit fields supply actual owner attribution. No probe is added to production lower source.

Scalar enum/boolean unions are classified as scalar fields using the adapted source declarations, rather than treated as objects merely because TypeFlagsUnion is set. The raw object_pair flag records broad probe eligibility; the final split applies scalar-types.json to that candidate set. Annotations add full nested paths, nil-source target types and reference fields, and deriveRelated results for intersection records. Object classification accepts object/union shapes and intersections, with intersection targets counted as objects even when the original probe labels them target-not-object. Source intersections also qualify when the target is object/union shaped. The original skipped flags remain in raw rows. Nil-source records remain a separate bucket even when the target is an intersection. Intersections do not add a second recursive traversal, so their fields skipped by the original probe remain unwalked.

## Counts

**Unrelated object pair observations: 1; distinct textual (source, target, nested path) pairs: 1. Nil-source views with a known runtime reference field: 0.** The object-capable reference companion is 0. Nil-source views with no proven reference field but unknown fields or missing target: 0.

Selected-unit pair observations: **37432**. Raw lines across all passes, including superseded first-error and retry observations: **50380**. Selected units held 2016 viewSources entries; pointer deduplication suppressed 210912 repeat visits and the depth bound suppressed 18164 visits.

| Split | Selected unit observations | All raw lines |
| --- | ---: | ---: |
| object_related_top | 1324 | 1618 |
| object_related_nested | 23189 | 31380 |
| object_unrelated_top | 1 | 1 |
| object_unrelated_nested | 0 | 0 |
| scalar_top | 3 | 3 |
| scalar_nested | 12915 | 17378 |
| nil_source | 0 | 0 |
| nil_target_only | 0 | 0 |

Intersection targets are a subset of the object counts, not an added bucket: 15, split as `{'object_related_top': 0, 'object_related_nested': 15, 'object_unrelated_top': 0, 'object_unrelated_nested': 0, 'scalar_top': 0, 'scalar_nested': 0, 'nil_source': 0, 'nil_target_only': 0}`. Scalar/non-object rows include source-proven primitive, enum and scalar-union fields that the probe's broad union flag test admits; scalar-types.json preserves that source-based classification. Intersection targets retain object status even when scalar-like; top-level non-object rows are shown separately from nested scalar fields.

Reference field means a cycleFinder.fields field whose lowering representation is known and ir.Type.IsReference() is true. This runtime definition includes strings and Weak. The companion object-capable count excludes strings and Weak. Unknown representations remain explicitly unknown, not silently declared scalar. The nil-source count is a count of views; unlike non-nil pairs, the original probe does not deduplicate nil sources.

## Coverage

Observed **7261 unit results**: `{'lowered': 4385, 'stopped': 2876}`. Deferred standalone units in emitted inventories: **290**. Unobserved stock units by status: `{'deferred_generic_instantiation': 290}`. Full-mode files attempted: **20/81**; emitted complete full-mode file inventories: 19. Stock inventory has 5049 units. Probe-incomplete selected units: 0; pair lines without a completed probe unit row: 0.

One result per (unit file, unit position) is selected: full mode supersedes first-error mode, and later full pieces supersede earlier ones. Repeated dependency pairs in different units remain repeated observations. TypeToString text can be truncated or reused for distinct checker identities; the distinct-pair list groups displayed source/target strings plus the full first-encountered nested path, not globally shared checker identities.

First-error stopped units expose the attempted lowering state before the stop, even when their IR body would be discarded. Full mode rolls failed statements back and retains successful statements. The probe sees the retained state at the end of that unit. Neither mode counts rolled-back state, skipped/deferred bodies, units not reached before a time limit, paths hidden behind a statement's first failure, pair paths suppressed by the probe's per-unit seen set, or recursion past depth 6. Target fields absent from the source type are skipped by the original probe, and non-object records do not recurse. A completed file traversal still has those gaps. Missing buffered latent inventories are not evidence that a file had no units.

The latent hook does not compute GraphTypes or establish graph membership. GraphTypes was non-empty in 0 observed results. No acyclicity is inferred. Probe errors: 0.

## Every distinct unrelated object pair

All counts and owner positions are also in [unrelated-object-pairs.json](unrelated-object-pairs.json). Paths are the nested field path at its first encounter in each unit. Files below are owner-unit files, not independent roots.

| Source type | Target type | Nested path | Count | Files and units |
| --- | --- | --- | ---: | --- |
| ArrayLiteralExpression \| CallExpression \| NewExpression | ArrayLiteralExpression | (top level) | 1 | checker.ts:28352:5 x1 |

## Every nil-source view with a runtime reference field

Each observation is retained, including repeats in one unit, in [nil-source-reference-views.json](nil-source-reference-views.json). Unknown-only cases are in [nil-source-reference-unknown.json](nil-source-reference-unknown.json).

| Unit | Target type | Runtime reference fields | Object-capable fields |
| --- | --- | --- | --- |
| none observed | | | |

## Per-file coverage

| Owner file | Lowered | Stopped | Selected pairs | Full mode |
| --- | ---: | ---: | ---: | --- |
| _namespaces/ts.moduleSpecifiers.ts | 1 | 0 | 0 | not yet attempted |
| _namespaces/ts.performance.ts | 1 | 0 | 0 | not yet attempted |
| _namespaces/ts.ts | 0 | 0 | 0 | not yet attempted |
| binder.ts | 88 | 92 | 2635 | complete traversal |
| builder.ts | 73 | 51 | 0 | complete traversal |
| builderPublic.ts | 10 | 6 | 0 | complete traversal |
| builderState.ts | 21 | 12 | 0 | complete traversal |
| builderStatePublic.ts | 3 | 0 | 0 | complete traversal |
| checker.ts | 653 | 716 | 9232 | bounded/nonzero exit |
| commandLineParser.ts | 104 | 69 | 325 | complete traversal |
| core.ts | 54 | 3 | 0 | not yet attempted |
| corePublic.ts | 9 | 0 | 0 | not yet attempted |
| debug.ts | 5 | 0 | 0 | not yet attempted |
| diagnosticInformationMap.generated.ts | 3 | 0 | 0 | complete traversal |
| emitter.ts | 276 | 147 | 1834 | complete traversal |
| executeCommandLine.ts | 42 | 26 | 29 | complete traversal |
| expressionToTypeNode.ts | 0 | 0 | 0 | not yet attempted |
| factory/baseNodeFactory.ts | 2 | 1 | 0 | not yet attempted |
| factory/emitHelpers.ts | 40 | 5 | 79 | not yet attempted |
| factory/emitNode.ts | 15 | 3 | 0 | not yet attempted |
| factory/nodeChildren.ts | 1 | 5 | 77 | not yet attempted |
| factory/nodeConverters.ts | 2 | 1 | 0 | not yet attempted |
| factory/nodeFactory.ts | 81 | 421 | 1243 | complete traversal |
| factory/nodeTests.ts | 228 | 0 | 0 | not yet attempted |
| factory/parenthesizerRules.ts | 1 | 2 | 76 | not yet attempted |
| factory/utilities.ts | 61 | 39 | 1428 | not yet attempted |
| factory/utilitiesPublic.ts | 3 | 0 | 0 | not yet attempted |
| hostErrors.ts | 0 | 3 | 0 | not yet attempted |
| moduleNameResolver.ts | 105 | 63 | 18 | complete traversal |
| moduleSpecifiers.ts | 25 | 23 | 7 | not yet attempted |
| parser.ts | 224 | 321 | 3281 | complete traversal |
| path.ts | 63 | 14 | 0 | not yet attempted |
| performance.ts | 23 | 1 | 0 | not yet attempted |
| performanceCore.ts | 10 | 2 | 0 | not yet attempted |
| program.ts | 123 | 112 | 660 | complete traversal |
| programDiagnostics.ts | 0 | 0 | 0 | not yet attempted |
| resolutionCache.ts | 37 | 38 | 7 | complete traversal |
| scanner.ts | 141 | 9 | 0 | complete traversal |
| semver.ts | 21 | 8 | 0 | not yet attempted |
| sourcemap.ts | 24 | 6 | 0 | not yet attempted |
| symbolWalker.ts | 1 | 1 | 56 | not yet attempted |
| sys.ts | 82 | 55 | 2 | complete traversal |
| tracing.ts | 10 | 0 | 0 | not yet attempted |
| transformer.ts | 0 | 0 | 0 | not yet attempted |
| transformers/classFields.ts | 20 | 3 | 267 | not yet attempted |
| transformers/classThis.ts | 2 | 3 | 56 | not yet attempted |
| transformers/declarations.ts | 0 | 0 | 0 | not yet attempted |
| transformers/declarations/diagnostics.ts | 0 | 0 | 0 | not yet attempted |
| transformers/destructuring.ts | 9 | 13 | 194 | not yet attempted |
| transformers/es2015.ts | 53 | 129 | 1936 | complete traversal |
| transformers/es2016.ts | 0 | 0 | 0 | not yet attempted |
| transformers/es2017.ts | 0 | 0 | 0 | not yet attempted |
| transformers/es2018.ts | 0 | 0 | 0 | not yet attempted |
| transformers/es2019.ts | 0 | 0 | 0 | not yet attempted |
| transformers/es2020.ts | 0 | 0 | 0 | not yet attempted |
| transformers/es2021.ts | 0 | 0 | 0 | not yet attempted |
| transformers/esDecorators.ts | 8 | 1 | 0 | not yet attempted |
| transformers/esnext.ts | 0 | 0 | 0 | not yet attempted |
| transformers/generators.ts | 0 | 0 | 0 | not yet attempted |
| transformers/jsx.ts | 0 | 0 | 0 | not yet attempted |
| transformers/legacyDecorators.ts | 1 | 1 | 0 | not yet attempted |
| transformers/module/esnextAnd2015.ts | 0 | 0 | 0 | not yet attempted |
| transformers/module/impliedNodeFormatDependent.ts | 0 | 0 | 0 | not yet attempted |
| transformers/module/module.ts | 0 | 0 | 0 | not yet attempted |
| transformers/module/system.ts | 0 | 0 | 0 | not yet attempted |
| transformers/namedEvaluation.ts | 3 | 16 | 56 | not yet attempted |
| transformers/taggedTemplate.ts | 3 | 2 | 121 | not yet attempted |
| transformers/ts.ts | 4 | 1 | 0 | not yet attempted |
| transformers/typeSerializer.ts | 5 | 1 | 79 | not yet attempted |
| transformers/utilities.ts | 17 | 27 | 1119 | not yet attempted |
| tsbuild.ts | 0 | 0 | 0 | not yet attempted |
| tsbuildPublic.ts | 0 | 0 | 0 | not yet attempted |
| types.ts | 875 | 0 | 0 | complete traversal |
| utilities.ts | 555 | 312 | 7314 | complete traversal |
| utilitiesPublic.ts | 155 | 87 | 5301 | not yet attempted |
| visitorPublic.ts | 4 | 24 | 0 | not yet attempted |
| watch.ts | 0 | 0 | 0 | not yet attempted |
| watchPublic.ts | 0 | 0 | 0 | not yet attempted |
| watchUtilities.ts | 0 | 0 | 0 | not yet attempted |
| tsc/_namespaces/ts.ts | 0 | 0 | 0 | not yet attempted |
| tsc/tsc.ts | 5 | 1 | 0 | not yet attempted |

## Pass exits

Each child is capped at 1490 seconds with five seconds for forced cleanup, below 25 minutes. Priority and files larger than 300000 bytes run alone. At most two smaller files run concurrently. Completed pieces are committed and pushed as they finish. Each process group is reaped and cleaned by supervise.py.

| Pass | Child exit | Seconds | Reason |
| --- | ---: | ---: | --- |
| first | -15 | 1491.012 | configured elapsed bound |
| full-000 | 0 | 621.917 | completed |
| full-001 | 0 | 82.12 | completed |
| full-002 | 0 | 20.013 | completed |
| full-003 | 0 | 43.079 | completed |
| full-004 | 0 | 18.018 | completed |
| full-005 | 0 | 47.044 | completed |
| full-006 | 0 | 77.234 | completed |
| full-007 | 0 | 84.082 | completed |
| full-008 | -15 | 1490.747 | configured elapsed bound |
| full-009 | 0 | 24.033 | completed |
| full-010 | 0 | 638.232 | completed |
| full-011 | 0 | 481.444 | completed |
| full-012 | 0 | 79.666 | completed |
| full-013 | 0 | 235.276 | completed |
| full-014 | 0 | 466.631 | completed |
| full-015 | 0 | 225.359 | completed |
| full-016 | 0 | 276.508 | completed |
| full-017 | 0 | 251.477 | completed |
| full-018 | 0 | 145.189 | completed |
| full-019 | 0 | 184.259 | completed |

## Evidence and replay

Raw streams, pass environments, timings, logs and actual child exits are in raw/. selected-pairs.jsonl.gz and selected-units.jsonl.gz preserve the selected population. overlay/ archives scratch Go inputs as .go.txt only. source-manifest.json hashes the adapted TypeScript bytes. provenance.json records pins, box and binary identity. SHA256SUMS covers evidence. report.py can recount finished archived/scratch streams without lowering. run-census.py resumes finished pieces without rerunning them.

```sh
source /workspace/adamic-tools/env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source
bash stage3/apply.sh /workspace/view-pairs-adapted
python3 stage3/census/latent/make_overlay.py /workspace/view-pairs /workspace/scratch/vqzty8t/overlay
python3 stage3/census/view-pairs-734aef5f/instrument.py /workspace/view-pairs /workspace/scratch/vqzty8t/overlay
python3 stage3/census/view-pairs-734aef5f/prepare-probe.py
go build -overlay /workspace/scratch/vqzty8t/overlay/overlay.json -o /workspace/scratch/vqzty8t/census ./stage3/census/latent/tool
python3 stage3/census/view-pairs-734aef5f/run-census.py
```
