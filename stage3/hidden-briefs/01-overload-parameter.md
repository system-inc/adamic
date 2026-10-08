# 01. overload 1 of visitBindingElement parameter elem cannot be served by implementation parameter elem

Historical boundary: `overload 1 of visitBindingElement parameter elem cannot be served by implementation parameter elem`. Raising function: `censusOverload (historical; absent on origin/main)` in `internal/lower/census_small.go`. There is no matching raising function on origin/main `89ac4a8c1de0b02d95965be72f7f8bf1c92433d2`; the function above is located at census pin `388096e6`.

**20,524 hidden bytes** across 2 selected region(s):

- #16: `transformers/declarations.ts` [68180, 81805) = 13,625 bytes
- #43: `transformers/declarations.ts` [82928, 89827) = 6,899 bytes

These are assigned head bytes within ranks 11–60, not a whole-corpus total or a promise that removing this stop reveals them all. The smallest recorded boundary containing the head wins. Enclosing blockers remain recorded in TABLE.md.

Largest region head: `transformers/declarations.ts:68180` (UTF-8 byte offset). Recorded diagnostic: `transformers/declarations.ts:619:54`.

Recorded attempt owner: `/tmp/hidden-adapted/src/compiler/transformers/declarations.ts:259:1`. Census stopping mode: `function signature or prologue failed; continued at next declaration`; boundary span [7221, 95509).

```text
/tmp/hidden-adapted/src/compiler/transformers/declarations.ts:619:54: Adamic 0.1 refuses overload 1 of visitBindingElement parameter elem cannot be served by implementation parameter elem; make the implementation accept every value admitted by this overload, without mutable widening or bivariance
```

Replay selector for that head, using the replay tool inherited from `codex/stage3-census-replay`, already present at census pin `388096e6`, with that pin's adapted source:

```sh
source /workspace/adamic-tools/env.sh
# Run from the replay worktree prepared as described in TABLE.md.
go run ./stage3/census/latent/replay \
  -project /tmp/hidden-adapted/src/compiler \
  -where /tmp/hidden-adapted/src/compiler/transformers/declarations.ts:619:54 -kind Refused \
  -reason 'overload 1 of visitBindingElement parameter elem cannot be served by implementation parameter elem' > /tmp/hidden-boundary-01-replay.log 2>&1
```

Observed replay: **exit 0, exact position/kind/reason reproduced**, in 28.751s. Log: `/tmp/hidden-briefs/replay-01.log`. The prebuilt census-overlay worker ran the same flags as the command above, with all 82 adapted file hashes matching RESULT.json. It selected the smallest eligible attempt in the full project.

Node semantic witness, to save as the proposed fixture below after resolving the stop. It illustrates the behavior at issue; it is not claimed to reproduce the census stop by itself.

```typescript
function visitBindingElement(elem: { readonly value: number }): number;
function visitBindingElement(elem: { readonly value: number | string }): number { return Number(elem.value); }
console.log(`${visitBindingElement({ value: 7 })}`);
```

Node 24.19.0, after `node:module.stripTypeScriptTypes`, observed stdout:

```text
7
```

Exit 0; stderr empty. Raw observation: `/tmp/hidden-briefs/node-01.log`.

This is a soundness refusal, not permission to trust an incompatible implementation. Preserve a negative test while finding a sound implementation or validated call specialization.

Proposed fixture path: `internal/oracle/testdata/hidden_boundary_overload_parameter.a`. No oracle fixture was added by this documentation unit; the current scratch witness is `/tmp/hidden-briefs/01.a`.

Existing branch: No verified fixing branch found. The historical `censusOverload` function and its reason are absent on origin/main and on 29 searched tips; this is not evidence of a landed fix. `origin/codex/notyet-overloads` still contains the parameter and result refusal templates. See TABLE.md for every absent tip.

Done when the implementation unit:

1. Replays the exact head stop, records its log and keeps unsafe cases refused.
2. Adds the sound `.a` fixture above and any separate negative fixture needed; source Node, native with ASan/UBSan/leak checks, and the JavaScript backend match stdout, stderr and exit status. Run only its focused oracle selector, writing output to a log.
3. Runs this mutant independently and records its intended catcher: Narrow the implementation parameter to a type not admitted by one overload; the negative overload fixture must remain refused.
4. Refreshes counts with `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts`, writes output to a log, and explains every changed row.
5. Commits its fix and fixtures, pushes its own branch once after local checks, and reports the SHA.
6. Re-measures the same hash-pinned largest region with the hidden census and reports old hidden intersection, new hidden intersection and their byte difference. Report the next boundary too. Do not report the group total as revealed bytes.

Revealed bytes in this documentation unit: **not measured**; no boundary implementation changed.
