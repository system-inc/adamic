# Resolution-only proof protocol

Commit parent: 7534c1d7575db02bb7b6512119e0685987b7e333.
TypeScript: 050880ce59e30b356b686bd3144efe24f875ebc8 (v6.0.3).

All proofs run from the exact resolution-only implementation commit. Its SHA,
commands, complete logs, raw streams and results are recorded on
[typescript/binding-views-resolution-evidence](https://github.com/system-inc/adamic/tree/typescript/binding-views-resolution-evidence/stage3/census/binding-views-resolution).

Acceptance requires a complete single-file resolutionCache.ts census with
CommonJSBundle, LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1 and the exact
LATENT_ONLY_FILE. The baseline targets at 774, 1053, 1056 and 1471 must disappear
with zero new and zero changed findings over the entire file. The restoring
mutant must return all four exact refusals with an otherwise empty baseline
delta. No program scan or overload claim is made.

Upstream npm ci and unmodified Hereby dts must produce byte-identical
typescript.d.ts and tsserverlibrary.d.ts against baseline. Stock tsc 6.0.3 must
add zero diagnostics over all 79 compiler roots, using the scout's pinned Node
ambient inputs; Adamic types must report zero. All 604 source files are compared
in both JavaScript comment modes. Normal stage3/apply.sh must succeed; its entire
source must match the measured candidate, including unchanged program.ts.
Adapter drift, duplicate, partial-owner, mixed-line-ending and replay controls
require zero writes on rejection or replay; the mutant writes only its one file.

The whole unfiltered stage3/lane/run.sh must pass with 106,366 passing, exactly
the sanctioned API row, and zero pending, using unchanged workers and settings.
Generated files are refreshed only if their canonical hashed inputs move. One
type-only ledger line cites #ty5myjt for the four resolution sites. No amendment
to 95-checker-axolotl-h is allowed. No source change follows the tested commit.
Nothing is submitted to Loom.

## Diagnostic amendment protocol

Base: 8e3a784aa5a601f57342c07b68f145fe56252c21; same upstream pin.
Evidence: typescript/diag-read-view-evidence, stage3/census/diag-read-view-2.
All four complete main/candidate censuses (program.ts, utilitiesPublic.ts,
utilities.ts, types.ts) use LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1,
LATENT_ONLY_FILE=<absolute adapted file>, CENSUS_MODULE_FORMAT=CommonJSBundle,
and a 6000-second per-file limit. Pair by phase, kind, mapped where and reason;
retain complete added/changed findings. Both program callers at baseline
2796:16 and 2829:16 must disappear. All four files require zero new findings.
program.ts, utilitiesPublic.ts and types.ts require zero changed findings.
utilities.ts permits exactly the refusal_scan/Refused pair at mapped 8808:39
and 8808:44, classified "changed: target type renamed, pre-existing refusal":
"optional property reportsUnnecessary in Diagnostic absent from structural
source DiagnosticRelatedInformation, which can hide fields" becomes "optional
property reportsUnnecessary in DiagnosticReadView absent from structural
source DiagnosticRelatedInformation, which can hide fields". This explicit user
ruling accepts neither another site nor another reason change. A new refusal outside the specified shape stops the amendment.

The restoring mutant returns the exact original diagnostics files and the
resolution half's original file; a full program census must return both target
refusals. Verify zero new stock TypeScript 6.0.3 diagnostics over 79 roots,
zero Adamic types diagnostics, normal apply, every source's transpiled JavaScript
in both comment modes, rejection controls and zero-write replay. Upstream npm ci
and node_modules/.bin/hereby dts run in both adapted checkouts. Both public
declarations must be byte-identical; a diff stops the amendment as an API question.
Run the whole unfiltered stage3/lane with 106,366 passing and only the sanctioned
API row. Refresh generated lane inputs only if their listed proof inputs move.
The exact tested implementation commit and all results are recorded in evidence.
