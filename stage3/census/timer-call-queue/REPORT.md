# Timer receiver and polling queue evidence

Base main: `f869de8c543afc26075c9f49538041b5273dcff5`. TypeScript: `050880ce59e30b356b686bd3144efe24f875ebc8`, exact tag `v6.0.3`. Node: `v24.19.0`. Both amendments implement Adamic and Compiler task #0gj665k (Nov 14 00:59Z); receiver work is #5vc7nbf. No Loom submission.

## Admission gate

Static counts: **0 timer accessor hosts; 0 argument expressions able to write the receiver timer pair**. Runtime counts: **0; 0**, over 5,865 actual calls, with all eight test workers and their parent retained. `gate-runtime-summary.json` and `gate-runtime-counts-3/` preserve every per-process receipt. The GcTimer block was not executed by this oracle; `forwarder-witness.json` exercises all nine exact blocks.

The source audit includes complete adapted src/ (harness and testRunner included) and tests/, plus pinned upstream src/tests. The four direct assignments in `timer-assignments.txt` initialize nodeServer timers or remove optional watch methods in a test before scheduling. None is in an argument capture. `property-definitions.txt` contains no timer-host accessor definition; its actual targets are compiler syntax/debug/cache objects, evaluator symbols and test-runner contexts, or emitted source strings. No direct timer accessor declaration exists.

| Captures | Why no timer write is reachable |
|---|---|
| Dynamic poller | Callback and label select from enum data; delay is a parameter; pollingIntervalQueue only returns one of three local queue consts. |
| Fixed chunk poller | Function reference, enum data and string literal. |
| Two WatchHost blocks | Function references and literals. |
| Solution builder | Function reference, local parameters and literal label. |
| ThrottledOperations | Static method reference and parameters/this. |
| GcTimer | Static method reference, constructor parameter property delay, and this. |
| MultistepOperation | Arrow creation does not execute its body; captures ms and actionType. Receiver getServerHost remains before the window in both shapes. |
| Typing installer | Arrow creation, static numeric delay and projectName/kind data fields from createInstallTypingsRequest. Existing witness also observes its label getter without timer writes. |

The observer checks own and inherited timer descriptors before capture evaluation, then monitors timer assignments and Object.defineProperty during captures and restores descriptors before invoking the saved method. Its accessor and assignment positive controls each record one violation; the clean control records neither (`gate-mutant-counts/`). Real bundle control transforms all nine blocks.

The first two oracle attempts are invalid setup controls: attempt 1 rejected npm stock TypeScript; attempt 2 rejected esbuild’s setTimeout2 alias. Both precede test execution. The corrected whole oracle completed with **106365 passing, 2 failing, 0 pending**. Besides the sanctioned public API acknowledgment, sys::symlinkWatching failed its modification-time assertion. The isolated unobserved rerun passed **16/16** (`watcher-unobserved/`). This is reported as a red full run, never a passing lane.

Amendment 1: `61603dc6761a2f1c8312a636ecb74be4a256ed86`, pushed to `typescript/timer-call-receiver-1`; parent is the base main above.

## Amendment 1 focused proofs

- Stage 0 restoring mutant: red exits 1 with inherited-library-member call refusal at timer-call-mutant-red.ts:22:12. Green exits 0; unchanged oracle/node.mjs resolves the runtime for emitted JS. Source Node and lowered Node both print `250\nlabel\n42\n7\n`. Both fixtures and outputs contain mutant in their names.
- Node forwarder witness: all nine actual main blocks versus their exact proposed forms agree on callbacks, delays, labels, receiver and captured identities and returned handle identity.
- Existing adaptation temporal/label witness: main and amendment 1 are byte-identical.
- Whole-owner contract: all 13 owners, LF/CRLF, no-write replay, byte/mtime preservation, drift/half/duplicate/moved refusals and unrelated edits pass. Exact before owners were reconstructed in scratch because full main is already adapted.
- `adamic types`: exit 0 on 79 compiler roots; stderr retained separately.
- JavaScript: `emission.cjs` asserts exact source equality after only nine receiver substitutions and nine unused-local deletions. Complete stock ES2024/ESNext module diff is javascript.main-1.diff.
- API sanction regeneration is byte-identical; generated shard/plan runtime check passes. No generated input hash moved.
- Full lane: PASS; fresh apply/install/build exit 0, tests/oracle exit 1 solely for the sanctioned public API acknowledgment; 106366 passing, 1 failing, 0 pending. Complete fresh src matches the independent candidate in all 747 files.

### Every amendment 1 census delta

All runs use LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, LATENT_ONLY_FILE and CommonJSBundle. Every accepted stream has a measurement header, an actual selected-file record and a final relation trailer. Compiler files load all 79 compiler roots; each server file is its own entry (including its dependencies), because compiler-only roots do not load server files. The compiler-only server attempts are retained as unloaded setup receipts and are not counted as proofs. Server entries have checker diagnostics, explicitly recorded by checker_rejected=true; their lowering attempts do not constitute compilation.

| File | Main unique | Amendment 1 unique | Gone | New | .call on main |
|---|---:|---:|---:|---:|---|
| server/session.ts | 1148 | 1148 | 0 | 0 | Absent; earlier stops mask the timer blocks |
| compiler/sys.ts | 132 | 130 | 2 | 0 | Fixed chunk refusal visible; dynamic queue block masked |
| compiler/tsbuildPublic.ts | 72 | 72 | 0 | 0 | Absent; earlier stops mask the timer blocks |
| server/typingInstallerAdapter.ts | 8 | 8 | 0 | 0 | Absent; earlier stops mask the timer blocks |
| server/utilities.ts | 16 | 16 | 0 | 0 | Absent; earlier stops mask the timer blocks |
| compiler/watchPublic.ts | 36 | 36 | 0 | 0 | Absent; earlier stops mask the timer blocks |

All selected-file findings, Boundary included, are compared by kind/where/reason/text with unchanged-line mapping. All dependency findings are unchanged: sys.ts retains 5; watchPublic.ts retains 1; other files retain 0. Raw streams include them. Changed findings appear as gone plus new, rather than being hidden.

- Gone: `Refused` `/tmp/adamic-timer-queue-baseline/src/compiler/sys.ts:494:20`: inherited library member call read as an own field.
- Gone: `Boundary` `/tmp/adamic-timer-queue-baseline/src/compiler/sys.ts:494:13`: failed statement; state rolled back; continued at next statement or declaration.

No other new, gone or changed findings.

## Amendment 2

Cut directly on 61603dc6761a2f1c8312a636ecb74be4a256ed86. The NodeArray reasoning was written in the README before the owner pin changed. files admits undefined; pollScheduled remains unknown. No narrowing check was added.

- Stop rule: PASS. Full sys.ts census is 130 -> 115: 15 gone, 0 new. No reader needs narrowing. All 5 dependency findings are unchanged; both checker headers have zero diagnostics.
- Every queue use/reader is counted in USES.md, AST inventories and whole adapted-tree/recursive-checkout search receipts. The helper's two existing readers and its source inventory are unchanged.
- Roads Node witness: the same 236 events and 6 change callbacks, actual amendment 1 versus composition owners, with the original scenario and extracted helpers. Complete traces agree.
- Restoring mutant: restoring the exact amendment 1 array-with-fields owner, keeping receiver calls, yields Refused sys.ts:272:23 again. Census exits 0 as a measurement, records the refusal and exposes no usable IR. A finally block restores composition; its bytes/hash are verified.
- adamic types: exit 0 on all 79 compiler roots.
- Whole-owner contract: PASS for all 13 owners, refusal controls, line endings and no-write byte/mtime replay.
- Existing temporal witness: byte-identical to amendment 1.
- Complete JavaScript diff: only compiler/sys.js changes, at the factory and three .files selections (javascript.1-2.diff).
- API sanction regeneration is byte-identical; generated shard/plan runtime check passes. No generated hash input moved.
- Fresh apply: PASS; all 747 source files match the independent candidate after line-ending normalization, with no added/missing file. Full lane: PASS, 106366 passing, one sanctioned public API acknowledgment, zero pending; apply/install/build exit 0 and tests/oracle exit 1 exactly as required.

### Every amendment 2 census change

- Gone: `Refused` `/tmp/adamic-timer-queue-baseline/src/compiler/sys.ts:272:23`: a cast the runtime can't check.
- Gone: `NotYet` `/tmp/adamic-timer-queue-baseline/src/compiler/sys.ts:246:11`: a value of type PollingIntervalQueue.
- Gone: `Boundary` `/tmp/adamic-timer-queue-baseline/src/compiler/sys.ts:234:1`: function signature or prologue failed; continued at next declaration.
- Gone: `NotYet` `/tmp/adamic-timer-queue-baseline/src/compiler/sys.ts:271:14`: a function returning PollingIntervalQueue.
- Gone: `Boundary` `/tmp/adamic-timer-queue-baseline/src/compiler/sys.ts:271:5`: function signature or prologue failed; continued at next declaration.
- Gone: `NotYet` `/tmp/adamic-timer-queue-baseline/src/compiler/sys.ts:279:61`: a value of type PollingIntervalQueue.
- Gone: `Boundary` `/tmp/adamic-timer-queue-baseline/src/compiler/sys.ts:279:5`: function signature or prologue failed; continued at next declaration.
- Gone: `NotYet` `/tmp/adamic-timer-queue-baseline/src/compiler/sys.ts:291:64`: a value of type PollingIntervalQueue.
- Gone: `Boundary` `/tmp/adamic-timer-queue-baseline/src/compiler/sys.ts:291:5`: function signature or prologue failed; continued at next declaration.
- Gone: `NotYet` `/tmp/adamic-timer-queue-baseline/src/compiler/sys.ts:339:14`: a function returning PollingIntervalQueue.
- Gone: `Boundary` `/tmp/adamic-timer-queue-baseline/src/compiler/sys.ts:339:5`: function signature or prologue failed; continued at next declaration.
- Gone: `Refused` `/tmp/adamic-timer-queue-baseline/src/compiler/sys.ts:351:9`: inherited library member push read as an own field.
- Gone: `Boundary` `/tmp/adamic-timer-queue-baseline/src/compiler/sys.ts:351:9`: failed statement; state rolled back; continued at next statement or declaration.
- Gone: `Boundary` `/tmp/adamic-timer-queue-baseline/src/compiler/sys.ts:361:9`: failed statement; state rolled back; continued at next statement or declaration.
- Gone: `Boundary` `/tmp/adamic-timer-queue-baseline/src/compiler/sys.ts:367:9`: failed statement; state rolled back; continued at next statement or declaration.

New: none. Changed: none beyond those removals. The cast, five queue value/return NotYets, inherited push refusal and associated boundaries disappear.

All header diagnostics are also compared as full-text multisets, mapping only unchanged source lines. Each server entry retains the same 3,222 checker diagnostics on main versus amendment 1, with no substantive change. Compiler entries and both composition sys.ts runs have zero diagnostics. Server entries remain measurements on checker-rejected programs, never successful lowering claims.

Amendment 2 is pushed as `6fefffa3a283c5603f7fd0a46b61db48a677f174` to `typescript/polling-queue-composition-1`, exactly one commit on `61603dc6761a2f1c8312a636ecb74be4a256ed86`. Both full lanes pass independently. No Loom submission was made.

## Additional legacy-control limitation

The old witness.cjs poller red control exits 0 on the frozen original main source as well as after composition. Main setCustomPollingValues no longer changes the enum, so its purported Low/Medium swap does not distinguish late queue selection. This pre-existing control is not counted as red evidence for either amendment. Production witness/criteria were not rewritten to conceal it. Full receipts are in legacy-mutant-results.json and legacy-poller-mutant.*. The request, label and restored-slot controls still exit 1 for the intended reasons.

The new nine-forwarder witness independently catches an in-memory delay mutant (251 versus 250) at Node trace equality, after checking actual source membership. The requested Function.call restoring fixture and array restoring mutant both produce their exact intended stage 0 refusals.

Raw source snapshots, emitted JavaScript, patches and logs retain their original CRLF/trailing whitespace. Generic whitespace checking flags those raw artifacts; the receipt is retained compressed. Authored scripts/docs and both feature diffs pass whitespace checks. Raw evidence was not normalized to make that check green.
