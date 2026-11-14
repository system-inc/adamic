# Binding views proof protocol

Base: 7534c1d7575db02bb7b6512119e0685987b7e333.
TypeScript: 050880ce59e30b356b686bd3144efe24f875ebc8, v6.0.3.

The exact commands, measured results, raw streams, declaration outputs and
checks are retained on [typescript/binding-views-evidence](https://github.com/system-inc/adamic/tree/typescript/binding-views-evidence/stage3/census/binding-views).

The declaration bisect runs each half separately with upstream Hereby dts.
Only resolution changes the public declaration bytes in the original scout
shape. The neighbouring `/** @internal */ export interface` convention restores
byte identity for both typescript.d.ts and tsserverlibrary.d.ts with both halves
applied. Both checkouts run npm ci and upstream's unmodified Hereby build.

Acceptance requires complete CommonJSBundle single-file resolutionCache.ts full
censuses with LATENT_FULL=1 and LATENT_ASSERT_NO_OUTPUT=1, selecting the exact
LATENT_ONLY_FILE. Four target refusals must disappear with zero new or changed
findings. Program's two targets use the promise-discard-sites wrapper: **scan
only**, no body lowering; zero new or changed findings across the entire file.
The restoring mutant reinstates both complete baseline files and must restore
all six refusals in the same respective census modes.

Stock TypeScript 6.0.3 checks all 79 sorted compiler roots against baseline;
Adamic types checks those same roots. Both JavaScript comment modes are compared
for every source file against baseline. Adapter drift, duplicate, partial-tree,
line-ending and replay controls require zero writes on rejection or replay.
Normal stage3/apply.sh must succeed, and its emitted source must match the
measured candidate. The whole unfiltered stage3/lane/run.sh must report 106,366
passing with only the sanctioned API failure. Generated files are refreshed
only when their actual hashed inputs change; expectations and sanctions are
not weakened. The requested type-only ledger line cites #ty5myjt and Compiler's
accept. No implementation commit is pushed until all required proofs pass.
Nothing is submitted to Loom.
