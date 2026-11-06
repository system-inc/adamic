# Enum views: nineteen-site B2 measurement

Target: 9bdbe75ede31d9b18e49ae3a79843eeaa89c63d4, change chg_bzxjsxe1,
from typescript/enum-views-readonly-2. This is a fresh worktree and a report-only
branch, typescript/enum-views-19. Adaptation 40 is unchanged.

The two earlier columns come from the prior retained measurement: main
63e14a0b2b61cf251ad74292e42d5fd09af804b3 had writable-record Refused at
invariance.go:673; adaptation 40 alone had namespace-value NotYet at
module_namespace.go:27. All nineteen rule IDs are retained: eighteen casts in
debug.ts plus rule 91 in emitter.ts, matching SITE-STOPS.md.

B2's four Library changes end at ff09c4494. Their parent is
7b66fad4dbece492846a10f878417ab9ecd1e36b, two main commits later than the
retained main baseline. The ordinary comparison measures the complete target
against that explicitly pinned earlier main, without attributing every change
exclusively to Library.

## Completed pieces

- Setup: the requested commit is checked out, pinned dependencies match, and
  the ordinary compiler, existing full latent census, and replay worker build.
  Raw provenance, tool hashes, build logs and reference records are in
  01-setup.tar.gz.

## Nineteen owning-unit results

The expected result is not observed by this latent measurement. All 19
casts lose the old namespace-value refusal-scan finding, but the 17 formatter
units stop at the enum read itself. None reaches a later stop or completes.
The FlowFlags getter owner and emitter's owner are blocked earlier. This report
records the unchanged existing full latent/unit-replay method; it does not
modify census registration to supply missing enum bindings.

Earlier main and adaptation-only columns describe cast scan stops, as the prior
SITE-STOPS.md did. The B2 column describes actual owning-unit lowering. `R0` is
Refused at invariance.go:673; `N0` is NotYet at module_namespace.go:27. `N1` is
NotYet: "an enum read through its namespace without a lowered binding", emitted
by internal/lower/namespaces.go:94. Every listed unit has status attempted.

| Rule | Owner / enum | Main cast scan | Adaptation-only cast scan | B2 unit first stop |
| --- | --- | --- | --- | --- |
| 72 | formatSyntaxKind / SyntaxKind | R0 | N0 | `debug.ts:424:33`: N1 |
| 73 | formatSnippetKind / SnippetKind | R0 | N0 | `debug.ts:428:33`: N1 |
| 74 | formatScriptKind / ScriptKind | R0 | N0 | `debug.ts:432:33`: N1 |
| 75 | formatNodeFlags / NodeFlags | R0 | N0 | `debug.ts:436:34`: N1 |
| 76 | formatNodeCheckFlags / NodeCheckFlags | R0 | N0 | `debug.ts:440:34`: N1 |
| 77 | formatModifierFlags / ModifierFlags | R0 | N0 | `debug.ts:444:34`: N1 |
| 78 | formatTransformFlags / TransformFlags | R0 | N0 | `debug.ts:448:34`: N1 |
| 79 | formatEmitFlags / EmitFlags | R0 | N0 | `debug.ts:452:34`: N1 |
| 80 | formatSymbolFlags / SymbolFlags | R0 | N0 | `debug.ts:456:34`: N1 |
| 81 | formatTypeFlags / TypeFlags | R0 | N0 | `debug.ts:460:34`: N1 |
| 82 | formatSignatureFlags / SignatureFlags | R0 | N0 | `debug.ts:464:34`: N1 |
| 83 | formatObjectFlags / ObjectFlags | R0 | N0 | `debug.ts:468:34`: N1 |
| 84 | formatFlowFlags / FlowFlags | R0 | N0 | `debug.ts:472:34`: N1 |
| 85 | formatRelationComparisonResult / RelationComparisonResult | R0 | N0 | `debug.ts:476:35`: N1 |
| 86 | formatCheckMode / CheckMode | R0 | N0 | `debug.ts:480:33`: N1 |
| 87 | formatSignatureCheckMode / SignatureCheckMode | R0 | N0 | `debug.ts:484:33`: N1 |
| 88 | formatTypeFacts / TypeFacts | R0 | N0 | `debug.ts:488:34`: N1 |
| 89 | attachFlowNodeDebugInfoWorker / FlowFlags | R0 | N0 | `debug.ts:496:15`: Refused: in that narrows a union of object types (any object with the key passes) |
| 91 | makeName / GeneratedIdentifierFlags | R0 | N0 | `emitter.ts:5769:30`: NotYet: a value of type (EmitNode & { autoGenerate: AutoGenerateInfo; }) | (EmitNode & { autoGenerate: AutoGenerateInfo; }) |

Rule 89's first stop is Refused at debug.ts:496:15: `in that narrows a union of
object types (any object with the key passes)`, before the getter cast at 519:56.
Rule 91's first stop is NotYet at emitter.ts:5769:30 on the EmitNode intersection
union, before the cast at 5790:141. Its unit also records later reads of the
unavailable autoGenerate binding, including 5790:84. All findings, not just the
first, are retained in unit-results.json inside the latent piece.

The emitter replay worker exits 1 because its deliberately requested old
namespace signature does not reproduce; it still returns the complete selected
makeName unit and its actual findings. This is not a timeout or worker crash.
No successful cast lowering is inferred from the absence of a scan finding.
The missing-binding observation may reflect the latent registration context;
it does not by itself establish failure of whole-program production lowering.

The target's complete apply passes. Its 79 compiler source files are byte-for-byte
identical to the prior adaptation-only tree, as source-audit.json records.
The separate ordinary 79-root census is still running. The separate B2 emitter
scan has no finding on the nineteenth cast's line. Its scan-only driver matches
all 39 findings from the full B2 debug scan exactly, lowers no units, and keeps
the no-output guards. It changes only the scratch driver, stored as .go.txt in
03-scan-and-replay.tar.gz. No production or adaptation-40 source is edited.
An independent single-unit formatSyntaxKind replay reproduces N1 with exit 0.
This confirms the missing-binding result without relying on preceding unit
attempts in the full census.

All raw records are packed in the numbered tar.gz files. SHA256SUMS hashes the
archives; each archive also contains a SHA256SUMS for its raw member files.
