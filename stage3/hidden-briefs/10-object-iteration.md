# 10. for...of over an object

Historical boundary: `for...of over an object`. Raising function: `forOf` in `internal/lower/object.go`. The template is present on origin/main `89ac4a8c1de0b02d95965be72f7f8bf1c92433d2`.

**8,175 hidden bytes** across 1 selected region(s):

- #29: `checker.ts` [1997085, 2005260) = 8,175 bytes

These are assigned head bytes within ranks 11–60, not a whole-corpus total or a promise that removing this stop reveals them all. The smallest recorded boundary containing the head wins. Enclosing blockers remain recorded in TABLE.md.

Largest region head: `checker.ts:1997085` (UTF-8 byte offset). Recorded diagnostic: `checker.ts:33562:34`.

Recorded attempt owner: `/tmp/hidden-adapted/src/compiler/checker.ts:33527:5`. Census stopping mode: `failed statement; state rolled back; continued at next statement or declaration`; boundary span [1997085, 2005260).

```text
/tmp/hidden-adapted/src/compiler/checker.ts:33562:34: stage 0 can't lower for...of over an object yet
```

Replay selector for that head, using the replay tool inherited from `codex/stage3-census-replay`, already present at census pin `388096e6`, with that pin's adapted source:

```sh
source /workspace/adamic-tools/env.sh
# Run from the replay worktree prepared as described in TABLE.md.
go run ./stage3/census/latent/replay \
  -project /tmp/hidden-adapted/src/compiler \
  -where /tmp/hidden-adapted/src/compiler/checker.ts:33562:34 -kind NotYet \
  -reason 'for...of over an object' > /tmp/hidden-boundary-10-replay.log 2>&1
```

Observed replay: **exit 0, exact position/kind/reason reproduced**, in 19.438s. Log: `/tmp/hidden-briefs/replay-10.log`. The prebuilt census-overlay worker ran the same flags as the command above, with all 82 adapted file hashes matching RESULT.json. It selected the smallest eligible attempt in the full project.

Node semantic witness, to save as the proposed fixture below after resolving the stop. It illustrates the behavior at issue; it is not claimed to reproduce the census stop by itself.

```typescript
const iterable = { *[Symbol.iterator](): Generator<number> { yield 7; yield 8; } };
let total = 0;
for (const value of iterable) { total += value; }
console.log(`${total}`);
```

Node 24.19.0, after `node:module.stripTypeScriptTypes`, observed stdout:

```text
15
```

Exit 0; stderr empty. Raw observation: `/tmp/hidden-briefs/node-10.log`.

The semantic witness uses a custom iterator and generator to expose the contract. Those introduce additional language boundaries; it is not claimed as an isolated Adamic reproduction.

Proposed fixture path: `internal/oracle/testdata/hidden_boundary_object_iteration.a`. No oracle fixture was added by this documentation unit; the current scratch witness is `/tmp/hidden-briefs/10.a`.

Existing branch: No verified fixing branch found in the 72 searched tips. The raising diagnostic template remains in all searched branches.

Done when the implementation unit:

1. Replays the exact head stop, records its log and keeps unsafe cases refused.
2. Adds the sound `.a` fixture above and any separate negative fixture needed; source Node, native with ASan/UBSan/leak checks, and the JavaScript backend match stdout, stderr and exit status. Run only its focused oracle selector, writing output to a log.
3. Runs this mutant independently and records its intended catcher: Skip the final yielded element; the iteration oracle must print a different total. Add an early-break cleanup case under sanitizers.
4. Refreshes counts with `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts`, writes output to a log, and explains every changed row.
5. Commits its fix and fixtures, pushes its own branch once after local checks, and reports the SHA.
6. Re-measures the same hash-pinned largest region with the hidden census and reports old hidden intersection, new hidden intersection and their byte difference. Report the next boundary too. Do not report the group total as revealed bytes.

Revealed bytes in this documentation unit: **not measured**; no boundary implementation changed.
