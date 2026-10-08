# 02. a function returning any

Historical boundary: `a function returning any`. Raising function: `signature` in `internal/lower/functions.go`. The template is present on origin/main `89ac4a8c1de0b02d95965be72f7f8bf1c92433d2`.

**15,966 hidden bytes** across 1 selected region(s):

- #13: `utilities.ts` [48518, 64484) = 15,966 bytes

These are assigned head bytes within ranks 11–60, not a whole-corpus total or a promise that removing this stop reveals them all. The smallest recorded boundary containing the head wins. Enclosing blockers remain recorded in TABLE.md.

Largest region head: `utilities.ts:48518` (UTF-8 byte offset). Recorded diagnostic: `core.ts:1891:17`.

Recorded attempt owner: `/tmp/hidden-adapted/src/compiler/utilities.ts:1384:1`. Census stopping mode: `failed statement; state rolled back; continued at next statement or declaration`; boundary span [48518, 64484).

```text
/tmp/hidden-adapted/src/compiler/core.ts:1891:17: stage 0 can't lower a function returning any yet
```

Replay selector for that head, using the replay tool inherited from `codex/stage3-census-replay`, already present at census pin `388096e6`, with that pin's adapted source:

```sh
source /workspace/adamic-tools/env.sh
# Run from the replay worktree prepared as described in TABLE.md.
go run ./stage3/census/latent/replay \
  -project /tmp/hidden-adapted/src/compiler \
  -where /tmp/hidden-adapted/src/compiler/core.ts:1891:17 -kind NotYet \
  -reason 'a function returning any' > /tmp/hidden-boundary-02-replay.log 2>&1
```

Observed replay: **exit 1, signature mismatch**, in 35.753s. All 82 adapted file hashes matched. The selected dependency declaration differs from the recorded caller attempt above; this selector is not a reproduction pass. Log: `/tmp/hidden-briefs/replay-02.log`.

```text
replay signature did not reproduce: NotYet a function returning any at /tmp/hidden-adapted/src/compiler/core.ts:1891:17 (selected /tmp/hidden-adapted/src/compiler/core.ts:1891:1)
```

The current-main reduction below reproduces the same reason and raising function. The exact census-position replay remains unverified.

Verified current-main reduction, to save as the proposed fixture below. Ran `/tmp/hidden-briefs/adamic c /tmp/hidden-briefs/02.a > /tmp/hidden-briefs/reduction-02.log 2>&1`; exit 1 with this boundary reason. This reproduces the reason in its raising function, not the original census source position. Node behavior follows.

```typescript
function identity(value: any): any { return value; }
console.log(`${identity(7)}`);
```

Node 24.19.0, after `node:module.stripTypeScriptTypes`, observed stdout:

```text
7
```

Exit 0; stderr empty. Raw observation: `/tmp/hidden-briefs/node-02.log`.

any is forbidden by doctrine. Reveal the source only through a proven type or an adaptation; do not add an unchecked runtime any representation.

Proposed fixture path: `internal/oracle/testdata/hidden_boundary_any_return.a`. No oracle fixture was added by this documentation unit; the current scratch witness is `/tmp/hidden-briefs/02.a`.

Existing branch: No verified fixing branch found in the 72 searched tips. The raising diagnostic template remains in all searched branches.

Done when the implementation unit:

1. Replays the exact head stop, records its log and keeps unsafe cases refused.
2. Adds the sound `.a` fixture above and any separate negative fixture needed; source Node, native with ASan/UBSan/leak checks, and the JavaScript backend match stdout, stderr and exit status. Run only its focused oracle selector, writing output to a log.
3. Runs this mutant independently and records its intended catcher: Replace the proven number result with 0; the sound, any-free oracle fixture must disagree with Node. Keep an any-bearing negative fixture refused.
4. Refreshes counts with `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts`, writes output to a log, and explains every changed row.
5. Commits its fix and fixtures, pushes its own branch once after local checks, and reports the SHA.
6. Re-measures the same hash-pinned largest region with the hidden census and reports old hidden intersection, new hidden intersection and their byte difference. Report the next boundary too. Do not report the group total as revealed bytes.

Revealed bytes in this documentation unit: **not measured**; no boundary implementation changed.
