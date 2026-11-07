# Nineteen enum-view sites on main after Lower's binding fix

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
comparison and source-emitter trace confirmations are still running.

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
the complete per-root comparison will follow.

All raw records are packed in numbered tar.gz archives. SHA256SUMS hashes the
archives, and every archive has an inner raw-file SHA256SUMS. Runs are bounded
below 25 minutes and pieces are committed and plainly pushed as completed.
