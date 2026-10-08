# 04. a value of type readonly T[] | undefined

Historical boundary: `a value of type readonly T[] | undefined`. Raising function: `typeOf` in `internal/lower/expression.go`. The template is present on origin/main `89ac4a8c1de0b02d95965be72f7f8bf1c92433d2`.

**13,061 hidden bytes** across 2 selected region(s):

- #40: `checker.ts` [620304, 627479) = 7,175 bytes
- #56: `checker.ts` [1623136, 1629022) = 5,886 bytes

These are assigned head bytes within ranks 11–60, not a whole-corpus total or a promise that removing this stop reveals them all. The smallest recorded boundary containing the head wins. Enclosing blockers remain recorded in TABLE.md.

Largest region head: `checker.ts:620304` (UTF-8 byte offset). Recorded diagnostic: `core.ts:22:27`.

Recorded attempt owner: `/tmp/hidden-adapted/src/compiler/checker.ts:10648:13`. Census stopping mode: `failed statement; state rolled back; continued at next statement or declaration`; boundary span [620304, 627479).

```text
/tmp/hidden-adapted/src/compiler/core.ts:22:27: stage 0 can't lower a value of type readonly T[] | undefined yet
```

Replay selector for that head, using the replay tool inherited from `codex/stage3-census-replay`, already present at census pin `388096e6`, with that pin's adapted source:

```sh
source /workspace/adamic-tools/env.sh
# Run from the replay worktree prepared as described in TABLE.md.
go run ./stage3/census/latent/replay \
  -project /tmp/hidden-adapted/src/compiler \
  -where /tmp/hidden-adapted/src/compiler/core.ts:22:27 -kind NotYet \
  -reason 'a value of type readonly T[] | undefined' > /tmp/hidden-boundary-04-replay.log 2>&1
```

Observed replay: **exit 1, signature mismatch**, in 21.605s. All 82 adapted file hashes matched. The selected dependency declaration differs from the recorded caller attempt above; this selector is not a reproduction pass. Log: `/tmp/hidden-briefs/replay-04.log`.

```text
replay signature did not reproduce: NotYet a value of type readonly T[] | undefined at /tmp/hidden-adapted/src/compiler/core.ts:22:27 (selected /tmp/hidden-adapted/src/compiler/core.ts:22:1)
```

The required exact head reproduction remains blocked by the selector mismatch. The tool exposes only a diagnostic-position selector, not a separate recorded caller-attempt selector. Do not weaken the match or claim the semantic witness below reproduces this stop. Simple independent reductions also did not emit this reason; they are not substitutes for the recorded caller context.

Node semantic witness, to save as the proposed fixture below after resolving the stop. It illustrates the behavior at issue; it is not claimed to reproduce the census stop by itself.

```typescript
function first<T>(items: readonly T[] | undefined): T | undefined { return items?.[0]; }
console.log(`${first([7]) ?? 0}|${first<number>(undefined) ?? 0}`);
```

Node 24.19.0, after `node:module.stripTypeScriptTypes`, observed stdout:

```text
7|0
```

Exit 0; stderr empty. Raw observation: `/tmp/hidden-briefs/node-04.log`.

This is an uninstantiated generic signature. Ordinary specialized calls can already succeed; the full census context is essential.

Proposed fixture path: `internal/oracle/testdata/hidden_boundary_generic_optional_array.a`. No oracle fixture was added by this documentation unit; the current scratch witness is `/tmp/hidden-briefs/04.a`.

Existing branch: No verified fixing branch found in the 72 searched tips. The raising diagnostic template remains in all searched branches.

Done when the implementation unit:

1. Replays the exact head stop, records its log and keeps unsafe cases refused.
2. Adds the sound `.a` fixture above and any separate negative fixture needed; source Node, native with ASan/UBSan/leak checks, and the JavaScript backend match stdout, stderr and exit status. Run only its focused oracle selector, writing output to a log.
3. Runs this mutant independently and records its intended catcher: Return undefined for every present array input; the present and absent input oracle must disagree with Node.
4. Refreshes counts with `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts`, writes output to a log, and explains every changed row.
5. Commits its fix and fixtures, pushes its own branch once after local checks, and reports the SHA.
6. Re-measures the same hash-pinned largest region with the hidden census and reports old hidden intersection, new hidden intersection and their byte difference. Report the next boundary too. Do not report the group total as revealed bytes.

Revealed bytes in this documentation unit: **not measured**; no boundary implementation changed.
