# baobab: stopped at the SourceFile augmentation conflict

No census ran and no adaptation was written. This is preflight evidence only, not the requested 79-file census report. The two deliverable branches were not created.

The exact requested optional member is already an optional internal method of SourceFileLike in pinned TypeScript 6.0.3, src/compiler/types.ts:4291. SourceFileObject implements the exact same three-parameter signature in src/services/services.ts:1130, including allowEdits?: true. That explicit signature stop condition does not apply.

There is another declaration that prevents the requested checks from passing. src/services/types.ts:161 augments the compiler SourceFile interface, and line 171 declares:

```typescript
getPositionOfLineAndCharacter(line: number, character: number): number;
```

It is required. The requested optional member in the compiler interface merges with that required overload. TypeScript reports TS2386 at src/services/types.ts:171:9: `Overload signatures must all be optional or required.`

merge-probe.cjs loads the actual pinned services/services.ts program twice. Its CompilerHost overlays only compiler/types.ts in memory on the second pass. It copies the optional member character for character from SourceFileLike, adding no internal marker. The probe asks for semantic diagnostics of the actual services/types.ts augmentation and checks the number of TS2386 diagnostics. The original has zero. The requested adaptation has exactly one. Both assertions pass; the probe exits 0. The source checkout is unmodified.

This prevents Part 2's zero-error services and implementer check. Fixing it would require a further decision about the required services augmentation and its public declaration. The current request authorizes one added optional member and requires no other public declaration changes, so no broader change was attempted.

Setup and baseline preparation were started before discovering the conflict, then deliberately cancelled. Their logs do not establish green setup, green apply, a census, or successful lowering. The wrapper child exits and cancellation record are retained. No declaration emit, Node witness, restoring mutant, API sanction, lane regeneration, or ledger change was performed. docs/rulings.md was not modified.

Reproduce from a checkout with the pinned upstream source and the API dependencies installed:

```sh
source /workspace/adamic-tools/env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/path/to/TypeScript-050880ce59e30b356b686bd3144efe24f875ebc8
NODE_PATH="$PWD/stage3/api/node_modules" node stage3/census/sourcefile-like-c504dd8f/preflight/merge-probe.cjs "$ADAMIC_TYPESCRIPT_SOURCE" > /tmp/baobab-merge-probe.log 2>&1
result=$?
printf 'merge probe child exit: %s\n' "$result"
exit "$result"
```

status.json records the source hashes and exact base commits. The raw successful probe output is merge-probe.log. No JavaScript output or census site count is claimed.
