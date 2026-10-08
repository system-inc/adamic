# 12. reading name

Historical boundary: `reading name`. Raising function: `enumNeverValue` in `internal/lower/expression.go`. The template is present on origin/main `89ac4a8c1de0b02d95965be72f7f8bf1c92433d2`.

**7,705 hidden bytes** across 1 selected region(s):

- #36: `binder.ts` [29754, 37459) = 7,705 bytes

These are assigned head bytes within ranks 11–60, not a whole-corpus total or a promise that removing this stop reveals them all. The smallest recorded boundary containing the head wins. Enclosing blockers remain recorded in TABLE.md.

Largest region head: `binder.ts:29754` (UTF-8 byte offset). Recorded diagnostic: `binder.ts:760:13`.

Recorded attempt owner: `/tmp/hidden-adapted/src/compiler/binder.ts:749:5`. Census stopping mode: `failed statement; state rolled back; continued at next statement or declaration`; boundary span [29754, 37459).

```text
/tmp/hidden-adapted/src/compiler/binder.ts:760:13: stage 0 can't lower reading name yet
```

Replay selector for that head, using the replay tool inherited from `codex/stage3-census-replay`, already present at census pin `388096e6`, with that pin's adapted source:

```sh
source /workspace/adamic-tools/env.sh
# Run from the replay worktree prepared as described in TABLE.md.
go run ./stage3/census/latent/replay \
  -project /tmp/hidden-adapted/src/compiler \
  -where /tmp/hidden-adapted/src/compiler/binder.ts:760:13 -kind NotYet \
  -reason 'reading name' > /tmp/hidden-boundary-12-replay.log 2>&1
```

Observed replay: **exit 0, exact position/kind/reason reproduced**, in 25.428s. Log: `/tmp/hidden-briefs/replay-12.log`. The prebuilt census-overlay worker ran the same flags as the command above, with all 82 adapted file hashes matching RESULT.json. It selected the smallest eligible attempt in the full project.

Node semantic witness, to save as the proposed fixture below after resolving the stop. It illustrates the behavior at issue; it is not claimed to reproduce the census stop by itself.

```typescript
function enclosing(): () => string { const name = 'node'; return () => name; }
console.log(enclosing()());
```

Node 24.19.0, after `node:module.stripTypeScriptTypes`, observed stdout:

```text
node
```

Exit 0; stderr empty. Raw observation: `/tmp/hidden-briefs/node-12.log`.

The short closure witness is already valid JavaScript. The census stop concerns visibility in its recorded attempted unit, so replay must retain ancestor binding context.

Proposed fixture path: `internal/oracle/testdata/hidden_boundary_name_binding.a`. No oracle fixture was added by this documentation unit; the current scratch witness is `/tmp/hidden-briefs/12.a`.

Existing branch: No verified fixing branch found in the 72 searched tips. The raising diagnostic template remains in all searched branches.

Done when the implementation unit:

1. Replays the exact head stop, records its log and keeps unsafe cases refused.
2. Adds the sound `.a` fixture above and any separate negative fixture needed; source Node, native with ASan/UBSan/leak checks, and the JavaScript backend match stdout, stderr and exit status. Run only its focused oracle selector, writing output to a log.
3. Runs this mutant independently and records its intended catcher: Capture name as an empty string; the nested-call oracle must disagree with Node.
4. Refreshes counts with `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts`, writes output to a log, and explains every changed row.
5. Commits its fix and fixtures, pushes its own branch once after local checks, and reports the SHA.
6. Re-measures the same hash-pinned largest region with the hidden census and reports old hidden intersection, new hidden intersection and their byte difference. Report the next boundary too. Do not report the group total as revealed bytes.

Revealed bytes in this documentation unit: **not measured**; no boundary implementation changed.
