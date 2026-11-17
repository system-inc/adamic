# Narrowing-copy landing attempt: stopped at helper lowering

Task #ev14ch9. Pinned origin/main at start: `932a6344af63dc7c8728fffac05584c176d0475f`.

The Nov 17 07:58Z and 09:47Z rulings supersede the prior condition-6 stop: boolean overloads and readonly narrowing overloads do not block landing the 1/5 copy. **This attempt stops for a different reason: the copying helper itself does not lower.** No landing directory, ledger line, internal change or landing push was made.

## Failed mandatory gate

The plain helpers in the scratch candidate each build `const result: U[] = []`, push only inside `if (f(value))`, and return that fresh local. The optional twin returns undefined for undefined input. Both are `@internal` and beside filter. Exactly the frozen 16 calls use them. Declarations 1 and 5 alone are blanked, preserving coordinates; remaining signature texts and implementation are unchanged. This is a measurement candidate, not a completed adapter.

The full commandLineParser.ts census actually enters filterNarrowed and records:

```text
core.ts:302:13: stage 0 can't lower checked predicate overload target null yet
```

This is the `f(value)` call, reached from `convertToJson` and its nested `convertArrayLiteralExpressionToJson`. The original call at :2550 retains `AdamicJsonRecoveryValue`, whose type includes null. Before the change these units encounter filter's declaration refusal at :261; after it they reach the new helper's predicate-membership stop. They are **reached stops in already-blocked units**, not counted as newly broken clean units. Independently, their presence in the required helper body means proof 1 fails.

The exact helper bodies were copied to checker-clean reduced controls with concrete calls to both shapes, including undefined input. With `string | undefined`, both instantiate with zero findings. With `string | null | undefined` and the same `v => v !== undefined` predicate, both instantiate and stop with the same null-target reason. The optional helper is instantiated for both defined and undefined calls. This is not inferred from generic deferral.

The unchanged existing `predicateMembership` code in internal/lower/predicates_proof.go distributes union targets, handles undefined and primitive string/number/boolean targets, then reaches its unsupported-target fallback for null. A source snapshot and hash are retained. **No existing rule has been shown to prove the complete helper result**, since this prerequisite fails. No new rule or cast is invented.

This is **not a demonstrated certificate A (#5fws6q8) refusal**: the recorded stop is at the predicate call, not at the loop-built local or returned local. This evidence cannot say whether A would later be needed after the predicate-membership issue is resolved.

## Required proof lines

FAIL for an unrun row means that the required proof is incomplete after the mandatory gate failed, not that an unexecuted test was observed failing.

FAIL 1. New helper body stops at core.ts:302:13, f(value): checked predicate overload target null. Actual caller and checker-clean reduced probes reproduce it. Remaining caller-file censuses stopped after this failure.

FAIL 2. Whole-import-closure comparison incomplete after the stop. PASS limited isolated-owner control: filter declaration 261:1 moves to 263:1, original overload 2 now ordinal 1. Whole closure before first declaration stop is utilitiesPublic.ts:830:1, getParseTreeNode, not filter.

FAIL 3. Not run for a landing adaptation after proof 1 failed. Unchanged baseline apply exits 0; no TestProduct_Stage3AdaptedTree run.

FAIL 4. Not run: new unfrozen copying lane blocked by proof 1.

FAIL 5. Not run: built-compiler before/after Node witnesses blocked by proof 1. Reduced lowering probes are not Node witnesses.

FAIL 6. Not run: alias-restoring mutant blocked by proof 1.

FAIL 7. Not run: public declaration byte comparison blocked by proof 1.

PASS 8. Four sites read by hand on the pinned tip and scratch candidate, including flatMap/addRange and group continuations. No array write through either handle or identity use found in those inspected paths. Read, not run.

## Completed census comparisons

All completed children use `LATENT_FULL=1`, `LATENT_ASSERT_NO_OUTPUT=1`, the exact selected file, and `CENSUS_MODULE_FORMAT=CommonJSBundle`. The complete compiler directory is loaded. Exit 0 means measurement completion, not acceptance.

Comparison keys are unit, phase, kind, position and reason, after mapping source positions through the actual before/after line edits. Raw text, byte spans and complete streams remain archived. Newly reached diagnostics in already-blocked units are separate from new findings.

| File | Before | After | Gone | Changed | Reached | New in previously clean units |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| src/compiler/commandLineParser.ts | 510 | 508 | 10 | 0 | 8 | 0 |
| src/compiler/core.ts | 14 | 14 | 0 | 0 | 0 | 0 |
| src/compiler/factory/utilities.ts | 240 | 239 | 3 | 0 | 2 | 0 |

Every gone, changed, reached and new record is listed in full in [graded-censuses.json](evidence/graded-censuses.json). Core-only has 14 -> 14 findings and defers the new generics without an entry witness; that absence is not a helper proof. The command-line caller and reduced probes supply the decisive entry/failure observations. Factory utilities also completed. Transformer runs were cancelled after the mandatory failure; remaining service/checker file runs were not started. Partial streams, missing completion records, child exit codes and cancellation requests are retained and never treated as passes.

## Scan scope

PASS isolated complete filter owner before/after: :261:1 -> :263:1. Exact next reason:

```text
overload 1 of filter result T[] cannot be served by implementation result readonly T[] | undefined
```

That ordinal 1 is **original overload 2**. This control preserves the full implementation and original source coordinates. It is the independent declaration-census witness; core-only latent defers generic overloads.

The actual core import-closure baseline completes scan-all with 5,054 events and first declaration stop at utilitiesPublic.ts:830:1 (`overload 2 of getParseTreeNode result T | undefined cannot be served by implementation result Node | undefined`). It is a broader scope than the isolated filter owner. Its after-run was cancelled once proof 1 failed, so no whole-closure movement or event-delta claim is made. The cancellation child status is -15; the original wrapper then failed parsing its empty cancelled output (wrapper exit 1). Both statuses are kept. The queued census driver exits 143 after SIGTERM.

## Ownership and preserved obligations

Before any landing edit, all adapt.cjs, owners.json and targets.json files were searched for filter and every caller owner chain. No whole-filter pin was found. Axolotl-j pins getConfigFileSpecs with hashes and an inline after statement; receiver-transformers pins transformTypeScript whole. Checker/service imports are pinned by core-multimap. These would need amendments in their existing adapters; they have not been edited. The full search is archived.

The four hand reviews are repeated on this tip and the scratch candidate: **read, not run**. [hand-review.json](evidence/hand-review.json) states the handles and continuations inspected. No write or identity comparison was found along them; this is not a whole-program alias proof.

Any future landing still must document upstream's hole: filter can return its argument typed narrower while the caller retains a writable handle. The supplied freeze allocation price is 1,366 additional arrays for self compilation and 242,183 for the lane, including checker.ts:16764's 232,451. Should #4yrezf5 is the later erasure task for fresh, dead arguments. The premise remains that no reader compares a narrowing result with its argument by identity, directly or through a slot; the required unfrozen copying lane was not run here and is not claimed as its witness. No ledger acceptance is written.

## Reproduction and provenance

The fresh worktree is pinned at the recorded main tip; recursive pinned submodules and stage3/api npm ci completed. Upstream remains TypeScript v6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8. Go overlays were rebuilt at this tip with scratch temp/cache directories. Unchanged baseline apply exits 0. No compiler source was changed.

Archives contain exact before/after sources, completed and partial census streams, concrete helper controls, scan controls, generator/runner scripts, logs, hashes and proof statuses. To reproduce a helper control, build main's unchanged latent overlay, then run its binary on the extracted helper-reproduction subdirectory with the full/no-output/exact-file/CommonJSBundle environment stated above. The retained prepare-gate.py and run-latent.py use this run's absolute scratch paths and can be relocated.
