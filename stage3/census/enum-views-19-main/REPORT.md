# Nineteen enum-view sites on main after Lower's binding fix

**Result:** 17 formatter units now complete isolated lowering with no findings.
The getter-owner and emitter-owner units retain their earlier stops. The shared
getEnumMembers unit first stops at debug.ts:412:27, the class-getter guard, not
at the Map at line 399. In the ordinary census, 73 first stops move to
core.ts:1326:26; outcome counts remain 73 Refused, 4 NotYet and 2 accepted.

Frozen main: 01813e07a5b062baa5cabd9b5fb7e4ad6c983382.
The exact frozen SHA and tool hashes are recorded in 01-setup.tar.gz/provenance.json. This is a fresh worktree and a report-only
branch, typescript/enum-views-19-main. Adaptation 40 and production code are
unchanged by this report.

The comparison carries the previous three columns from report 9683a4e4:
main 63e14a0b, adaptation 40 alone, and B2 at 9bdbe75e. As before, the nineteen
sites mean eighteen debug.ts casts and rule 91 in emitter.ts. The new column
will record actual owning-unit outcomes, separately from scan admission.

The frozen main includes 80168a7db, the entry-reach and const-enum-binding fix,
and intervening changes including adaptations 99r and 99z. All ordinary first
stop changes are reported for the complete frozen snapshot; no causal
attribution solely to the enum fix is assumed.

## Completed pieces

01-setup.tar.gz contains the target and baseline pins, tool hashes, reference
records, and build logs. A stale temporary build directory from the prior setup
was recreated before the successful build; the initial failure is retained.
The target's full apply passes. The full debug.ts census and the emitter replay
have completed and are packed in 02-latent-units.tar.gz. The ordinary 79-root
comparison is complete in 04-ordinary.tar.gz. Source-emitter trace confirmations are complete
and are packed in 03-emitter-traces.tar.gz.

## Nineteen-site comparison

The 17 formatter units now finish their isolated lowering attempts with zero
lowering findings. This corrects the Map-at-399 prediction for this method:
they do not stop at the Map. formatEnum's own isolated unit also has no findings.
This is not a whole-program/native acceptance claim: the unit driver registers
callee signatures independently and does not close over all callee bodies and
global initializers as one production program.

R0 below is the earlier main cast scan Refused at invariance.go:673. N0 is the
casts-only namespace-value scan NotYet at module_namespace.go:27. N1 is B2's
unit NotYet, "an enum read through its namespace without a lowered binding".
L means attempted isolated unit completed with no lowering findings.

| Rule | Owner / enum | Earlier main | Casts alone | B2 at 9bdbe75e | Main now |
| --- | --- | --- | --- | --- | --- |
| 72 | formatSyntaxKind / SyntaxKind | R0 | N0 | N1 | L |
| 73 | formatSnippetKind / SnippetKind | R0 | N0 | N1 | L |
| 74 | formatScriptKind / ScriptKind | R0 | N0 | N1 | L |
| 75 | formatNodeFlags / NodeFlags | R0 | N0 | N1 | L |
| 76 | formatNodeCheckFlags / NodeCheckFlags | R0 | N0 | N1 | L |
| 77 | formatModifierFlags / ModifierFlags | R0 | N0 | N1 | L |
| 78 | formatTransformFlags / TransformFlags | R0 | N0 | N1 | L |
| 79 | formatEmitFlags / EmitFlags | R0 | N0 | N1 | L |
| 80 | formatSymbolFlags / SymbolFlags | R0 | N0 | N1 | L |
| 81 | formatTypeFlags / TypeFlags | R0 | N0 | N1 | L |
| 82 | formatSignatureFlags / SignatureFlags | R0 | N0 | N1 | L |
| 83 | formatObjectFlags / ObjectFlags | R0 | N0 | N1 | L |
| 84 | formatFlowFlags / FlowFlags | R0 | N0 | N1 | L |
| 85 | formatRelationComparisonResult / RelationComparisonResult | R0 | N0 | N1 | L |
| 86 | formatCheckMode / CheckMode | R0 | N0 | N1 | L |
| 87 | formatSignatureCheckMode / SignatureCheckMode | R0 | N0 | N1 | L |
| 88 | formatTypeFacts / TypeFacts | R0 | N0 | N1 | L |
| 89 | attachFlowNodeDebugInfoWorker / FlowFlags | R0 | N0 | Refused at debug.ts:496:15 | Refused at debug.ts:496:15 |
| 91 | makeName / GeneratedIdentifierFlags | R0 | N0 | NotYet at emitter.ts:5769:30 | NotYet at emitter.ts:5769:30 |

Rule 89 still stops earlier at debug.ts:496:15 (U1), before its getter cast at
519:56. Rule 91 still stops first at emitter.ts:5769:30 (U2), before its cast at
5790:141. Its subsequent autoGenerate reads remain recovery findings (U3).
The emitter replay exits 1 because the deliberately requested old namespace
signature no longer reproduces; it still returns the complete makeName unit.
Every actual finding is retained, including rollback boundaries.

## The shared helper and the Map prediction

getEnumMembers now reaches the runtime-key read `enumObject[name]` at
**debug.ts:412:27**, then stops at K1, the class-getter syntax guard. It does not
report a Map constructor stop at debug.ts:399 in this measurement. The Map is a
namespace-level initializer and is not a separate function unit in this census;
absence of a constructor finding does not prove whole-program initialization.

The full debug record contains no diagnostic at line 399. After rollback, the
same helper unit records K2, `reading value` at 413:24, and K3, `a comparator
that isn't a function` at core.ts:1049:82 during toSorted's generic body. These
are later recovery findings, not replacements for K1 as its first stop. The
corresponding source expressions are `typeof value` and `array.slice().sort(comparer)`.

## Exact source and lowering emitters

The following locations are on the frozen main. Source snippets and Go guard
excerpts are packed in the raw evidence. Refused/NotYet is preserved, rather
than describing every stop as a type-system refusal.

| ID | Source stop | Lowering file:line | Lowering function | Meaning |
| --- | --- | --- | --- | --- |
| K1 | `debug.ts:412:27` | `internal/lower/reflect_get.go:111` | `checkDynamicKeyedGetterSyntax` | readonly runtime-key read encounters the program-wide class-getter syntax guard |
| K2 | `debug.ts:413:24` | `internal/lower/expression.go:660` | `enumNeverValue` | reading value after the failed initializer was rolled back |
| K3 | `core.ts:1049:82` | `internal/lower/object.go:2065` | `arraySort` | comparator is not held as a closure with one call signature |
| U1 | `debug.ts:496:15` | `internal/lower/unknown.go:46` | `inProperty` | in narrows multiple object members |
| U2 | `emitter.ts:5769:30` | `internal/lower/expression.go:53` | `typeOf` | representation of the EmitNode intersection union is not known |
| U3 | `emitter.ts:5770:48,5771:48,5772:17,5790:84` | `internal/lower/expression.go:660` | `enumNeverValue` | reading autoGenerate after its initializer was rolled back |
| O1 | `core.ts:1326:26` | `internal/lower/cast_proof.go:188` | `castProof` | uncheckable-cast diagnostic construction, propagated through castRefusal |

K1 is called by `viewKeyedRead` at element_access_fields.go:134; its emitting
return is reflect_get.go:111 in checkDynamicKeyedGetterSyntax. The diagnostic
prefix is named by dynamicKeyedRead, and the syntax reason by dynamicKeyedGetterBody.
The program-wide guard can reject a view read because another module class has
a getter whose effects would be hidden inside that read. It is not a Map-key
storage diagnostic. O1 is the newly exposed ordinary core cast observed so far;
the complete per-root comparison is below.

All raw records are packed in numbered tar.gz archives. SHA256SUMS hashes the
archives, and every archive has an inner raw-file SHA256SUMS. Runs are bounded
below 25 minutes and pieces are committed and plainly pushed as completed.

## Source-emitter confirmation

A scratch-only logger in notYet records runtime.Caller's Go file, line and
function while returning the original NotYet unchanged. The existing no-output
replay worker reproduces the getEnumMembers and emitter stops with exit 0.
The helper's six findings and emitter's ten findings are identical parsed records to their corresponding untraced lowering findings. The trace
confirms K1, K2, K3, U2 and U3 at the listed emitters. The logger's Go source is
stored as .go.txt inside the archive; no production file is edited. U1 and O1
are direct Refused constructors at the listed unique source emitters.

## Ordinary 79-root census against the earlier main

All 79 TypeScript compiler roots were measured independently with the ordinary
production compiler. Three non-TypeScript records and the driver's whole-program
record are retained raw but excluded from this per-file comparison.

| Outcome | Earlier main | Main now |
| --- | ---: | ---: |
| NotYet | 4 | 4 |
| Refused | 73 | 73 |
| accepted | 2 | 2 |

**73 changed first stops.** All 73 retain Refused with `a cast the runtime can't
check`, but move from core.ts:810:89 to core.ts:1326:26. The new source is
`values.push((collection as MapLike<T>)[key]!);` in getOwnValues, whose collection
parameter is MapLike<T> | T[]. Diagnostic O1 is constructed at
internal/lower/cast_proof.go:188 in castProof and propagated by castRefusal.
This is the next unproven cast, not a new enum-view failure. The six other roots
retain their exact first stop or accepted outcome. Every changed entry is listed:

| Entry root under src/compiler | Earlier first stop | Main first stop | New emitter |
| --- | --- | --- | --- |
| `_namespaces/ts.moduleSpecifiers.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `_namespaces/ts.performance.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `_namespaces/ts.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `binder.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `builder.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `builderPublic.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `builderState.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `builderStatePublic.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `checker.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `commandLineParser.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `emitter.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `executeCommandLine.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `expressionToTypeNode.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `factory/baseNodeFactory.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `factory/emitHelpers.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `factory/emitNode.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `factory/nodeChildren.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `factory/nodeConverters.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `factory/nodeFactory.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `factory/nodeTests.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `factory/parenthesizerRules.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `factory/utilities.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `factory/utilitiesPublic.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `moduleNameResolver.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `moduleSpecifiers.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `parser.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `path.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `performance.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `performanceCore.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `program.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `programDiagnostics.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `resolutionCache.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `scanner.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `semver.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `sourcemap.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `symbolWalker.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `sys.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `tracing.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `transformer.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `transformers/classFields.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `transformers/classThis.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `transformers/declarations.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `transformers/declarations/diagnostics.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `transformers/destructuring.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `transformers/es2015.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `transformers/es2016.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `transformers/es2017.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `transformers/es2018.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `transformers/es2019.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `transformers/es2020.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `transformers/es2021.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `transformers/esDecorators.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `transformers/esnext.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `transformers/generators.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `transformers/jsx.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `transformers/legacyDecorators.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `transformers/module/esnextAnd2015.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `transformers/module/impliedNodeFormatDependent.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `transformers/module/module.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `transformers/module/system.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `transformers/namedEvaluation.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `transformers/taggedTemplate.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `transformers/ts.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `transformers/typeSerializer.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `transformers/utilities.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `tsbuild.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `tsbuildPublic.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `utilities.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `utilitiesPublic.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `visitorPublic.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `watch.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `watchPublic.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |
| `watchUtilities.ts` | Refused `core.ts:810:89` | Refused `core.ts:1326:26` | O1: castProof, cast_proof.go:188 |

The raw census and ordinary-comparison.json preserve every root's complete kind,
source location, reason, fix and diagnostics. Only paths and timing are normalized.
Intervening main adaptations 99r and 99z remove other earlier core stops; this
comparison reports the whole main snapshot and does not attribute the ordinary
movement exclusively to Lower's enum-binding fix.

## Reproduction, limits and archive audit

The source pin is TypeScript 6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8.
After sourcing /workspace/adamic-tools/env.sh and setting
ADAMIC_TYPESCRIPT_SOURCE, the fresh worktree runs stage3/apply.sh with
--build-product. The ordinary tool is built from cmd/adamic and stage3/census/tool
without a compiler overlay. The latent tool and replay worker use the existing
make_overlay.py output, just as report 9683a4e4 did. LATENT_FULL=1 and
LATENT_ONLY_FILE select only debug.ts for the full latent run; the separate
emitter worker selects makeName. The same nineteen rule IDs are located with
the retained TypeScript parser. All flags, pins, input hashes and raw observations
are packed with the corresponding pieces.

The full debug piece has 105 units and 126 findings. An attempted unit with no
lowering findings is described as completed in isolation, not as native code
emission or closed dependency/global-initializer acceptance. No diagnostic at
line 399 is evidence only about this measurement, not a proof that its global
Map initializer has been lowered. The helper guard at line 412 and its recovery
findings are independently replayed with the original errors unchanged.

Each measurement is wrapped in timeout 1490s with a 10s termination grace. All
pieces completed without timeout and were committed and plainly pushed as they
finished. Raw records are in the four numbered tar.gz files, with archive hashes
in SHA256SUMS and raw member hashes inside each archive. The source/emitter
catalog is in 02 and its runtime origin confirmations in 03. Piece 04 contains
all ordinary comparison rows, source excerpts, and the final scope/cleanup audit.
Only this report directory changes from the frozen main; adaptation 40 and
production compiler source are unchanged. No measurement process remains active.
