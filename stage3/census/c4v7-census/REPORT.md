# c4v-7 census — setup checkpoint

sys.ts:304:13: unmeasured. scanner.ts:2934:41: unmeasured. New refusal-scan findings: unmeasured.

Measurement and reporting only; no adaptation or compiler changes.
Head: `1d48ecb5a5321717daf901579e49a23291995589`. Main: `5c8338c4232c37afb304ce6865458b67655c590a`.
`git merge-base` returned exactly the main commit. Detached worktrees were freshly created under `/workspace/scratch/c4v7-census`.
Both stage3/api `npm ci` installs passed. Recursive submodule setup passed at both commits. Upstream TypeScript HEAD and peeled v6.0.3 both match `050880ce59e30b356b686bd3144efe24f875ebc8`. Both scan-only builds and all four promise controls passed. Both apply runs completed successfully. All 747 adapted src files are byte-identical between commits, including exactly 79 compiler TypeScript files; the compiler source also matches the prior census manifest. Line mapping is identity. Part 2 paired batches are running.

The first npm installs used the default npm cache; both were repeated successfully with an explicit scratch cache and preserved logs.

Part 2 precedes Part 1. Scan-only sees refusal-scan findings only, and cannot establish lowering-phase stops.
The requested promise-discard-sites README is absent at branch tip `2dc0593b937f8fbaf6c32f8107a899f78c223966`; its REPORT.md and the latent README document the wrapper and build procedure.
The baseline next-stops branch `d4f3ac9023013b7eb3d9dac4e8f61e1d64ecb76c` contains a validated CommonJSBundle table; its report says EcmaScript has not started.

The inherited raw header says `latent_mode: first-error`; the scratch driver calls `LatentScanOnly`, and all controls and file records require zero lowering units.
