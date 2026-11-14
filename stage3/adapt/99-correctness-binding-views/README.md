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
