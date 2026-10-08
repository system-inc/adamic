# 15. reading oldSourceFiles

Historical boundary: `reading oldSourceFiles`. Raising function: `enumNeverValue` in `internal/lower/expression.go`. The template is present on origin/main `89ac4a8c1de0b02d95965be72f7f8bf1c92433d2`.

**6,260 hidden bytes** across 1 selected region(s):

- #48: `program.ts` [117206, 123466) = 6,260 bytes

These are assigned head bytes within ranks 11–60, not a whole-corpus total or a promise that removing this stop reveals them all. The smallest recorded boundary containing the head wins. Enclosing blockers remain recorded in TABLE.md.

Largest region head: `program.ts:117206` (UTF-8 byte offset). Recorded diagnostic: `program.ts:2390:37`.

Recorded attempt owner: `/tmp/hidden-adapted/src/compiler/program.ts:2345:5`. Census stopping mode: `failed statement; state rolled back; continued at next statement or declaration`; boundary span [117206, 123466).

```text
/tmp/hidden-adapted/src/compiler/program.ts:2390:37: stage 0 can't lower reading oldSourceFiles yet
```

Replay selector for that head, using the replay tool inherited from `codex/stage3-census-replay`, already present at census pin `388096e6`, with that pin's adapted source:

```sh
source /workspace/adamic-tools/env.sh
# Run from the replay worktree prepared as described in TABLE.md.
go run ./stage3/census/latent/replay \
  -project /tmp/hidden-adapted/src/compiler \
  -where /tmp/hidden-adapted/src/compiler/program.ts:2390:37 -kind NotYet \
  -reason 'reading oldSourceFiles' > /tmp/hidden-boundary-15-replay.log 2>&1
```

Observed replay: **exit 0, exact position/kind/reason reproduced**, in 29.401s. Log: `/tmp/hidden-briefs/replay-15.log`. The prebuilt census-overlay worker ran the same flags as the command above, with all 82 adapted file hashes matching RESULT.json. It selected the smallest eligible attempt in the full project.

Node semantic witness, to save as the proposed fixture below after resolving the stop. It illustrates the behavior at issue; it is not claimed to reproduce the census stop by itself.

```typescript
function enclosing(): () => number { const oldSourceFiles = [7]; return () => oldSourceFiles[0] ?? 0; }
console.log(`${enclosing()()}`);
```

Node 24.19.0, after `node:module.stripTypeScriptTypes`, observed stdout:

```text
7
```

Exit 0; stderr empty. Raw observation: `/tmp/hidden-briefs/node-15.log`.

The short closure witness is already valid JavaScript. The census stop concerns visibility in its recorded attempted unit, so replay must retain ancestor binding context.

Proposed fixture path: `internal/oracle/testdata/hidden_boundary_old_source_files.a`. No oracle fixture was added by this documentation unit; the current scratch witness is `/tmp/hidden-briefs/15.a`.

Existing branch: No verified fixing branch found in the 72 searched tips. The raising diagnostic template remains in all searched branches.

Done when the implementation unit:

1. Replays the exact head stop, records its log and keeps unsafe cases refused.
2. Adds the sound `.a` fixture above and any separate negative fixture needed; source Node, native with ASan/UBSan/leak checks, and the JavaScript backend match stdout, stderr and exit status. Run only its focused oracle selector, writing output to a log.
3. Runs this mutant independently and records its intended catcher: Capture an empty oldSourceFiles array rather than the enclosing binding; the nested-call oracle must disagree with Node.
4. Refreshes counts with `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts`, writes output to a log, and explains every changed row.
5. Commits its fix and fixtures, pushes its own branch once after local checks, and reports the SHA.
6. Re-measures the same hash-pinned largest region with the hidden census and reports old hidden intersection, new hidden intersection and their byte difference. Report the next boundary too. Do not report the group total as revealed bytes.

Revealed bytes in this documentation unit: **not measured**; no boundary implementation changed.
