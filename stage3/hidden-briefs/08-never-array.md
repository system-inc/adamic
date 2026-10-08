# 08. an array of never

Historical boundary: `an array of never`. Raising function: `elementType` in `internal/lower/object.go`. The template is present on origin/main `89ac4a8c1de0b02d95965be72f7f8bf1c92433d2`.

**9,790 hidden bytes** across 1 selected region(s):

- #24: `utilities.ts` [356194, 365984) = 9,790 bytes

These are assigned head bytes within ranks 11–60, not a whole-corpus total or a promise that removing this stop reveals them all. The smallest recorded boundary containing the head wins. Enclosing blockers remain recorded in TABLE.md.

Largest region head: `utilities.ts:356194` (UTF-8 byte offset). Recorded diagnostic: `utilities.ts:9049:23`.

Recorded attempt owner: `/tmp/hidden-adapted/src/compiler/utilities.ts:9041:1`. Census stopping mode: `failed statement; state rolled back; continued at next statement or declaration`; boundary span [356194, 365984).

```text
/tmp/hidden-adapted/src/compiler/utilities.ts:9049:23: stage 0 can't lower an array of never yet
```

Replay selector for that head, using the replay tool inherited from `codex/stage3-census-replay`, already present at census pin `388096e6`, with that pin's adapted source:

```sh
source /workspace/adamic-tools/env.sh
# Run from the replay worktree prepared as described in TABLE.md.
go run ./stage3/census/latent/replay \
  -project /tmp/hidden-adapted/src/compiler \
  -where /tmp/hidden-adapted/src/compiler/utilities.ts:9049:23 -kind NotYet \
  -reason 'an array of never' > /tmp/hidden-boundary-08-replay.log 2>&1
```

Observed replay: **exit 0, exact position/kind/reason reproduced**, in 39.521s. Log: `/tmp/hidden-briefs/replay-08.log`. The prebuilt census-overlay worker ran the same flags as the command above, with all 82 adapted file hashes matching RESULT.json. It selected the smallest eligible attempt in the full project.

Verified current-main reduction, to save as the proposed fixture below. Ran `/tmp/hidden-briefs/adamic c /tmp/hidden-briefs/08.a > /tmp/hidden-briefs/reduction-08.log 2>&1`; exit 1 with this boundary reason. This reproduces the reason in its raising function, not the original census source position. Node behavior follows.

```typescript
const items: never[] = [];
console.log(`${items.length}`);
```

Node 24.19.0, after `node:module.stripTypeScriptTypes`, observed stdout:

```text
0
```

Exit 0; stderr empty. Raw observation: `/tmp/hidden-briefs/node-08.log`.

Use a proven representation and preserve Node evaluation order, identity and cleanup.

Proposed fixture path: `internal/oracle/testdata/hidden_boundary_never_array.a`. No oracle fixture was added by this documentation unit; the current scratch witness is `/tmp/hidden-briefs/08.a`.

Existing branch: No verified fixing branch found in the 72 searched tips. The raising diagnostic template remains in all searched branches.

Done when the implementation unit:

1. Replays the exact head stop, records its log and keeps unsafe cases refused.
2. Adds the sound `.a` fixture above and any separate negative fixture needed; source Node, native with ASan/UBSan/leak checks, and the JavaScript backend match stdout, stderr and exit status. Run only its focused oracle selector, writing output to a log.
3. Runs this mutant independently and records its intended catcher: Emit an empty never array with length 1; the length oracle must disagree with Node.
4. Refreshes counts with `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts`, writes output to a log, and explains every changed row.
5. Commits its fix and fixtures, pushes its own branch once after local checks, and reports the SHA.
6. Re-measures the same hash-pinned largest region with the hidden census and reports old hidden intersection, new hidden intersection and their byte difference. Report the next boundary too. Do not report the group total as revealed bytes.

Revealed bytes in this documentation unit: **not measured**; no boundary implementation changed.
