# Enum views: nineteen-site B2 measurement

**Result:** the old cast scan stops disappear at all nineteen sites, but the
expected owning-unit lowering is not observed. Seventeen formatter units stop
at the enum live-binding lookup; the other two units stop earlier. The ordinary
79-root census has zero changed first stops against the pinned earlier main.

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
- Full apply and nineteen-unit observations: 02-latent-units.tar.gz.
- Independent replay and scan-only control: 03-scan-and-replay.tar.gz.
- Complete ordinary comparison and final audit: 04-ordinary.tar.gz.

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
| 91 | makeName / GeneratedIdentifierFlags | R0 | N0 | `emitter.ts:5769:30`: NotYet: EmitNode intersection union |

Rule 89's first stop is Refused at debug.ts:496:15: `in that narrows a union of
object types (any object with the key passes)`, before the getter cast at 519:56.
Rule 91's first stop is NotYet at emitter.ts:5769:30 on the EmitNode intersection
union, before the cast at 5790:141. Its unit also records later reads of the
unavailable autoGenerate binding, including 5790:84. The first union reason is
`a value of type (EmitNode & { autoGenerate: AutoGenerateInfo; }) | (EmitNode & { autoGenerate: AutoGenerateInfo; })`. All findings, not just the
first, are retained in unit-results.json inside the latent piece.

The emitter replay worker exits 1 because its deliberately requested old
namespace signature does not reproduce; it still returns the complete selected
makeName unit and its actual findings. This is not a timeout or worker crash.
No successful cast lowering is inferred from the absence of a scan finding.
The missing-binding observation may reflect the latent registration context;
it does not by itself establish failure of whole-program production lowering.

The target's complete apply passes. Its 79 compiler source files are byte-for-byte
identical to the prior adaptation-only tree, as source-audit.json records.
The ordinary 79-root comparison is complete; see below. The separate B2 emitter
scan has no finding on the nineteenth cast's line. Its scan-only driver matches
all 21 findings from the full B2 debug scan exactly, lowers no units, and keeps
the no-output guards. It changes only the scratch driver, stored as .go.txt in
03-scan-and-replay.tar.gz. No production or adaptation-40 source is edited.
An independent single-unit formatSyntaxKind replay reproduces N1 with exit 0.
This confirms the missing-binding result without relying on preceding unit
attempts in the full census.

All raw records are packed in the numbered tar.gz files. SHA256SUMS hashes the
archives; each archive also contains a SHA256SUMS for its raw member files.

## Ordinary 79-root compiler census

The ordinary production compiler driver ran independently on all 79 .ts roots.
Three non-TypeScript file records and a whole-program record are preserved raw
and excluded from this per-file comparison. Only tree prefixes and timings are
normalized; kind, location, reason, fix and any diagnostics are compared.

| Kind | Main | B2 |
| --- | ---: | ---: |
| NotYet | 4 | 4 |
| Refused | 73 | 73 |
| accepted | 2 | 2 |

Changed first stops: **0**.

Every root retains its exact first stop or accepted outcome. There is no
ordinary per-file first-stop movement to attribute to the B2 stack.

The raw census, full normalized per-root comparison and source hash audit are
in 04-ordinary.tar.gz. The comparison baseline is the earlier pinned main, not
a claim about an unmeasured moving origin/main tip.

## Reproduction and validation

The ordinary tool is built without a compiler overlay from the pinned target.
The latent tool and worker use the existing make_overlay.py output; only the
supplementary scan-only tool adds an early emit after the existing refusal scan.
The latter's .go.txt source and overlay map are packed in piece 03. No production
Go, adaptation-40 rule, owner pin, or source fixture is changed.

Commands were run from the fresh worktree after sourcing
/workspace/adamic-tools/env.sh, with stdout/stderr redirected to the retained
logs. ADAMIC_TYPESCRIPT_SOURCE points to the pinned TypeScript source. Essential
commands, using the paths recorded in provenance.json, are:

```sh
bash stage3/apply.sh /workspace/scratch/ev19/applied --build-product
go build -o /workspace/scratch/ev19/bin/ ./cmd/adamic ./stage3/census/tool
python3 stage3/census/latent/make_overlay.py /workspace/enum-views-19 /workspace/scratch/ev19/overlay
go build -buildvcs=false -overlay=/workspace/scratch/ev19/overlay/overlay.json -o /workspace/scratch/ev19/bin/latent ./stage3/census/latent/tool
LATENT_FULL=1 LATENT_ONLY_FILE=/workspace/scratch/ev19/applied/src/compiler/debug.ts LATENT_ASSERT_NO_OUTPUT=1 /workspace/scratch/ev19/bin/latent /workspace/scratch/ev19/applied/src/compiler /workspace/scratch/ev19/debug-latent.jsonl
/workspace/scratch/ev19/bin/replay -project /workspace/scratch/ev19/applied/src/compiler -where /workspace/scratch/ev19/applied/src/compiler/debug.ts:424:33 -kind NotYet -reason 'an enum read through its namespace without a lowered binding'
/workspace/scratch/ev19/bin/tool /workspace/scratch/ev19/applied/src/compiler /workspace/scratch/ev19/ordinary.jsonl
```

Each measurement command is wrapped in timeout 1490s with a 10s termination
grace. All pieces completed without timeout and were committed and plainly
pushed as completed. The full debug census has 105 units and 166 findings;
its scan-only control reproduces all 21 refusal-scan findings exactly.
The old nineteen namespace cast-scan findings are absent, but no owning unit
completes beyond its cast in these observations. No native lowering or execution
success is claimed. A census-local missing binding is not, by itself, a proof
of production whole-program failure.

Archive and member hashes were checked. The final audit records zero active
measurement processes and confirms that only this report directory differs from
the requested target. Adaptation 40 remains unchanged. A scan-control count in
an earlier report commit was corrected from 39 to the observed 21; its raw
control record was correct throughout.
