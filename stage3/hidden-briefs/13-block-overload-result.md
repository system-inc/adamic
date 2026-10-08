# 13. overload 1 of transformAsyncFunctionBody result Block cannot be served by implementation result ConciseBody

Historical boundary: `overload 1 of transformAsyncFunctionBody result Block cannot be served by implementation result ConciseBody`. Raising function: `censusOverload (historical; absent on origin/main)` in `internal/lower/census_small.go`. There is no matching raising function on origin/main `89ac4a8c1de0b02d95965be72f7f8bf1c92433d2`; the function above is located at census pin `388096e6`.

**7,289 hidden bytes** across 1 selected region(s):

- #38: `transformers/es2017.ts` [30237, 37526) = 7,289 bytes

These are assigned head bytes within ranks 11–60, not a whole-corpus total or a promise that removing this stop reveals them all. The smallest recorded boundary containing the head wins. Enclosing blockers remain recorded in TABLE.md.

Largest region head: `transformers/es2017.ts:30237` (UTF-8 byte offset). Recorded diagnostic: `transformers/es2017.ts:736:5`.

Recorded attempt owner: `/tmp/hidden-adapted/src/compiler/transformers/es2017.ts:736:5`. Census stopping mode: `function signature or prologue failed; continued at next declaration`; boundary span [30237, 30435).

```text
/tmp/hidden-adapted/src/compiler/transformers/es2017.ts:736:5: Adamic 0.1 refuses overload 1 of transformAsyncFunctionBody result Block cannot be served by implementation result ConciseBody; make the implementation result covariant with every overload result
```

Replay selector for that head, using the replay tool inherited from `codex/stage3-census-replay`, already present at census pin `388096e6`, with that pin's adapted source:

```sh
source /workspace/adamic-tools/env.sh
# Run from the replay worktree prepared as described in TABLE.md.
go run ./stage3/census/latent/replay \
  -project /tmp/hidden-adapted/src/compiler \
  -where /tmp/hidden-adapted/src/compiler/transformers/es2017.ts:736:5 -kind Refused \
  -reason 'overload 1 of transformAsyncFunctionBody result Block cannot be served by implementation result ConciseBody' > /tmp/hidden-boundary-13-replay.log 2>&1
```

Observed replay: **exit 0, exact position/kind/reason reproduced**, in 28.711s. Log: `/tmp/hidden-briefs/replay-13.log`. The prebuilt census-overlay worker ran the same flags as the command above, with all 82 adapted file hashes matching RESULT.json. It selected the smallest eligible attempt in the full project.

Node semantic witness, to save as the proposed fixture below after resolving the stop. It illustrates the behavior at issue; it is not claimed to reproduce the census stop by itself.

```typescript
type Block = { readonly kind: 'Block'; readonly value: number };
type ConciseBody = Block | { readonly kind: 'Expression'; readonly value: number };
function transformAsyncFunctionBody(body: Block): Block;
function transformAsyncFunctionBody(body: ConciseBody): ConciseBody { return body; }
console.log(`${transformAsyncFunctionBody({ kind: 'Block', value: 7 }).value}`);
```

Node 24.19.0, after `node:module.stripTypeScriptTypes`, observed stdout:

```text
7
```

Exit 0; stderr empty. Raw observation: `/tmp/hidden-briefs/node-13.log`.

Preserve the overload result promise; a Block-only overload cannot accept a general ConciseBody result without proof.

Proposed fixture path: `internal/oracle/testdata/hidden_boundary_block_overload_result.a`. No oracle fixture was added by this documentation unit; the current scratch witness is `/tmp/hidden-briefs/13.a`.

Existing branch: No verified fixing branch found. The historical `censusOverload` function and its reason are absent on origin/main and on 29 searched tips; this is not evidence of a landed fix. `origin/codex/notyet-overloads` still contains the parameter and result refusal templates. See TABLE.md for every absent tip.

Done when the implementation unit:

1. Replays the exact head stop, records its log and keeps unsafe cases refused.
2. Adds the sound `.a` fixture above and any separate negative fixture needed; source Node, native with ASan/UBSan/leak checks, and the JavaScript backend match stdout, stderr and exit status. Run only its focused oracle selector, writing output to a log.
3. Runs this mutant independently and records its intended catcher: Accept an overload implementation returning an Expression for a Block-only promise; the negative fixture must remain refused.
4. Refreshes counts with `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts`, writes output to a log, and explains every changed row.
5. Commits its fix and fixtures, pushes its own branch once after local checks, and reports the SHA.
6. Re-measures the same hash-pinned largest region with the hidden census and reports old hidden intersection, new hidden intersection and their byte difference. Report the next boundary too. Do not report the group total as revealed bytes.

Revealed bytes in this documentation unit: **not measured**; no boundary implementation changed.
