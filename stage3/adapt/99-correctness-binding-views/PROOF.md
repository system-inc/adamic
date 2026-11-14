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
