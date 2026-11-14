# Scout: sys.ts:514 and utilitiesPublic.ts:1280

Status: setup checkpoint; no candidate measurements yet. Compiler and adaptations unchanged.

Measured-head target: `4f460a611b61a5b11c12ce61fde8ad5d0f11a912` (fresh detached worktree, then report branch). Prior evidence: `716d205b`, not counted as a new measurement.

## Recommended roads (provisional)

- Wall A: withhold selection until all three scratch candidates have stock 6.0.3 diagnostics, full sys.ts census and emitted JavaScript comparisons. New measured numbers: none.
- Wall B: investigate the language/compiler proof transport boundary; do not rewrite getFirstJSDocTag merely to bypass it. New measured numbers: none. At this head, predicateArgumentProof explicitly handles bodies and producer calls but not a parameter declaration.

## Setup

The shared checkout was not edited. The requested commit was initially absent (`fatal: invalid reference: 4f460a611b61a5b11c12ce61fde8ad5d0f11a912`); fetching the two named branches succeeded. Scratch worktree: `/workspace/scratch/scout-514-1280`. `npm ci` in its stage3/api succeeded (3 packages). Recursive submodule initialization and upstream TypeScript clone are in progress; no credential error observed.

## What this checkpoint cannot see

No adapted tree, baseline census, candidate diagnostics or emit comparisons have run yet. No claim that 514:17 clears, that diagnostics stay unchanged, or that emitted behavior is identical is supported at this checkpoint.
