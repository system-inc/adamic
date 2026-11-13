# SourceFileLike position candidate: stopped

Main: `751879544ea6e627e67c91d6ab3c4ba595be6356`.
Upstream: `050880ce59e30b356b686bd3144efe24f875ebc8`, exact tag `v6.0.3`.
Measurements used a fresh detached scratch worktree and recursively initialized submodules. The shared checkout source was not edited.

## Result

No adaptation was built. Check 4 did not pass: the **measurement harness failed on the baseline**, before candidate declaration emission. This is not evidence that the candidate breaks declarations.

The stripInternal probe flattened project references into one program using `src/typescript/tsconfig.json`, whose inherited `types` is empty. It failed to preserve `src/compiler/tsconfig.json`'s `types: ["node"]`. Stock TypeScript reported 64 diagnostics, beginning with missing `Error.captureStackTrace`, `node:perf_hooks`, and `node:fs`. The probe exited 1. Its exact script and complete diagnostics are retained. No compiler options or sources were changed to bypass this result.

Per the stop-on-failure instruction, the upstream declaration bundler, lane API guard, and `adamic types` were not subsequently run. Declaration equivalence remains unestablished. No sanctioned-api.json change, adapter directory, restoring mutant, shard regeneration, lane run, adaptation commit, or Loom submission was made.

## Passing observations before the stop

- Ownership: `71-writable-views` modifies only the `lineMap` member. No whole `SourceFileLike` interface pin was found. This would be a new `99-correctness-sourcefile-like-position` directory if built.
- JavaScript: stock TypeScript 6.0.3 emitted identical JavaScript for 604 source files (1,208 emits, comments retained and removed). Identical source bytes reuse their deterministic identical emit. Only the intended types.ts member spelling differs. The JSDoc and @internal remain untouched.
- Full single-file census: scanner findings 138 -> 137; the sole removed finding is scanner.ts:469:12. No new or changed scanner findings. types.ts findings 0 -> 0. All four census commands exited 0. The headers report no checker diagnostics; the historical measurement-label string still says checker-rejected.
- Guarded full-census replay of the actual SourceFile-to-SourceFileLike argument at utilities.ts:7959:37 attempted `positionsAreOnSameLine` on both trees and found no refusal on either. Both signature-match commands exited 1 because the requested callable-contract refusal was absent; these are negative controls, not new candidate findings.
- Production sys.ts first stop: main scanner.ts:469:12; candidate scanner.ts:2934:41. Both exited 1 as expected for refusals. The paired full census confirms 2934:41 already existed on main.

`finding-diff.json` lists all headline and attempt-context changes. Raw census streams, unit timings, build/apply logs, setup pins, the exact trial patch, and scripts are retained here. `commands.txt` records commands, exits, and result paths. An environment interruption killed the first apply without an exit code; a fresh output apply completed successfully and was the only baseline used for measurements.
