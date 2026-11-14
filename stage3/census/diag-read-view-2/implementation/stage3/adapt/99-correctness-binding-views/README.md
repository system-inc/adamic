# Resolution binding view

TypeScript pin: 050880ce59e30b356b686bd3144efe24f875ebc8 (v6.0.3).
Adamic base: 7534c1d7575db02bb7b6512119e0685987b7e333.

This is the resolutionCache.ts half of scout #ty5myjt, using the neighbouring
`/** @internal */ export interface` convention for ResolutionReadView. The view
extends ResolutionWithResolvedFileName and redeclares resolvedFileName as
`readonly resolvedFileName: string | undefined`. R constraints and the default
of GetResolutionWithResolvedFileName use this view. The original interface,
wrapper T types, runtime bodies, calls, ordering and values are unchanged.

The first ownership grep found no earlier whole-owner pin for these edited
resolution declarations. owners.json pins the complete ResolutionCache
interface, GetResolutionWithResolvedFileName alias and createResolutionCache
function, plus the inserted view. A stricter full-file text/SHA-256 pin protects
all surrounding source bytes. Unknown text, duplicate declarations, partial
adaptation, parse errors, mixed line endings and source changes during planning
fail before writes. LF and CRLF are preserved. Both commented and comment-free
JavaScript are asserted identical before writing. Replay writes zero files.

`node adapt.cjs TREE` edits only resolutionCache.ts. `node mutant.cjs TREE`
restores only that file's exact original; it is a measurement tool outside normal
apply. This directory has no program.ts pin, dependency or edit. The diagnostics
overload candidate is excluded after the full program census found a lowering
refusal; no 95-checker-axolotl-h amendment is included.

The ledger line covers only the four resolution sites under #ty5myjt.
[PROOF.md](PROOF.md) defines acceptance. Commands, raw full censuses, mutant,
declaration builds, type checks, JavaScript comparisons and the whole lane are
retained on the [resolution evidence branch](https://github.com/system-inc/adamic/tree/typescript/binding-views-resolution-evidence/stage3/census/binding-views-resolution).
The exact implementation commit is created before these proofs run and pushed
only after acceptance. Nothing is submitted to Loom.

## Diagnostic read view amendment

Base: 8e3a784aa5a601f57342c07b68f145fe56252c21. The resolution half above,
its original pin entry, engine and restoring behavior are retained unchanged.
The diagnostic amendment adds an `/** @internal */ export` DiagnosticReadView
with readonly file, start and length. The public sortAndDeduplicateDiagnostics
signature stays unchanged between an internal read-view overload and its
read-view implementation; its body stays unchanged.

The enclosing createProgram implementation is owned whole by
95-checker-axolotl-h. Its diagnostic-owners.json therefore owns the entire
createProgram declaration group before and after the constraint amendment,
and the complete namespace import. Its diagnostic-owner.cjs
recognizes precisely that amended owner on 95h replay and normalizes only the
constraint for the existing original-owner proof. 95h's historical intermediate
and all six folds are unchanged. The binding-views adapter invokes these
amendment pins together with its own pins as one diagnostic group, before any
writes. No independent late edit bypasses the 95h owner contract.

All eight diagnostic reader helpers use DiagnosticReadView parameters:
compareDiagnostics, compareDiagnosticsSkipRelatedInformation,
compareRelatedInformation, compareMessageText, diagnosticsEqualityComparer,
getDiagnosticFilePath, getDiagnosticCode and getDiagnosticMessage. Bodies are
unchanged. The v1 full census demonstrated readonly-to-writable refusals through
the inner helper chain. The v2 user ruling authorizes completing all six remaining
substitutions. Results and complete raw censuses live on
typescript/diag-read-view-evidence, under stage3/census/diag-read-view-2.
Diagnostic owners and complete edited import declarations are pinned whole, as text
and SHA-256. diagnostic-engine.cjs plans every file before writing, rejects
owner drift, duplicates, partial diagnostics adaptations and mixed line endings,
checks both JavaScript comment modes, and preserves LF/CRLF. Resolution planning
still uses the unchanged original engine. Replay writes zero files.

The existing ledger entry remains the resolution half's historical entry; this
amendment does not add a drafted upstream bug. Nothing is submitted to Loom.

The utilities.ts namespace import is already pinned whole by
99-correctness-core-multimap. Its diagnostic-owners.json owns the read-view
import amendment at both the binding and final MultiMap boundaries, with exact
whole import text/SHA-256 pairs. Binding-views invokes those pins; MultiMap
retains the type import in its own before/after pins. Replay preserves the
current known boundary, and restoration removes only the diagnostic type import,
retaining the baseline MultiMap runtime imports.

The user accepts exactly the refusal_scan/Refused pair at mapped utilities.ts
8808:39 and 8808:44 as "changed: target type renamed, pre-existing refusal".
At both recursive arguments, the old reason is "optional property
reportsUnnecessary in Diagnostic absent from structural source
DiagnosticRelatedInformation, which can hide fields"; the new reason replaces
Diagnostic with DiagnosticReadView and changes nothing else. Both texts remain
in evidence. No other new or changed finding is accepted.
