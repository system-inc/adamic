# 05. a value of type ((node: Node) => boolean) | undefined

Historical boundary: `a value of type ((node: Node) => boolean) | undefined`. Raising function: `typeOf` in `internal/lower/expression.go`. The template is present on origin/main `89ac4a8c1de0b02d95965be72f7f8bf1c92433d2`.

**11,826 hidden bytes** across 2 selected region(s):

- #49: `transformers/es2018.ts` [35690, 41768) = 6,078 bytes
- #59: `transformers/es2015.ts` [144178, 149926) = 5,748 bytes

These are assigned head bytes within ranks 11–60, not a whole-corpus total or a promise that removing this stop reveals them all. The smallest recorded boundary containing the head wins. Enclosing blockers remain recorded in TABLE.md.

Largest region head: `transformers/es2018.ts:35690` (UTF-8 byte offset). Recorded diagnostic: `visitorPublic.ts:151:5`.

Recorded attempt owner: `/tmp/hidden-adapted/src/compiler/transformers/es2018.ts:827:5`. Census stopping mode: `failed statement; state rolled back; continued at next statement or declaration`; boundary span [35690, 35769).

```text
/tmp/hidden-adapted/src/compiler/visitorPublic.ts:151:5: stage 0 can't lower a value of type ((node: Node) => boolean) | undefined yet
```

Replay selector for that head, using the replay tool inherited from `codex/stage3-census-replay`, already present at census pin `388096e6`, with that pin's adapted source:

```sh
source /workspace/adamic-tools/env.sh
# Run from the replay worktree prepared as described in TABLE.md.
go run ./stage3/census/latent/replay \
  -project /tmp/hidden-adapted/src/compiler \
  -where /tmp/hidden-adapted/src/compiler/visitorPublic.ts:151:5 -kind NotYet \
  -reason 'a value of type ((node: Node) => boolean) | undefined' > /tmp/hidden-boundary-05-replay.log 2>&1
```

Observed replay: **exit 1, signature mismatch**, in 23.161s. All 82 adapted file hashes matched. The selected dependency declaration differs from the recorded caller attempt above; this selector is not a reproduction pass. Log: `/tmp/hidden-briefs/replay-05.log`.

```text
replay signature did not reproduce: NotYet a value of type ((node: Node) => boolean) | undefined at /tmp/hidden-adapted/src/compiler/visitorPublic.ts:151:5 (selected /tmp/hidden-adapted/src/compiler/visitorPublic.ts:148:1)
```

The required exact head reproduction remains blocked by the selector mismatch. The tool exposes only a diagnostic-position selector, not a separate recorded caller-attempt selector. Do not weaken the match or claim the semantic witness below reproduces this stop. Simple independent reductions also did not emit this reason; they are not substitutes for the recorded caller context.

Node semantic witness, to save as the proposed fixture below after resolving the stop. It illustrates the behavior at issue; it is not claimed to reproduce the census stop by itself.

```typescript
interface Node { readonly value: number; }
function test(node: Node, predicate: ((node: Node) => boolean) | undefined): boolean { return predicate?.(node) ?? false; }
console.log(`${test({ value: 7 }, node => node.value === 7)}|${test({ value: 7 }, undefined)}`);
```

Node 24.19.0, after `node:module.stripTypeScriptTypes`, observed stdout:

```text
true|false
```

Exit 0; stderr empty. Raw observation: `/tmp/hidden-briefs/node-05.log`.

Optional callback representation may require generic specialization before it is concrete. Test both presence and absence.

Proposed fixture path: `internal/oracle/testdata/hidden_boundary_optional_node_callback.a`. No oracle fixture was added by this documentation unit; the current scratch witness is `/tmp/hidden-briefs/05.a`.

Existing branch: No verified fixing branch found in the 72 searched tips. The raising diagnostic template remains in all searched branches.

Done when the implementation unit:

1. Replays the exact head stop, records its log and keeps unsafe cases refused.
2. Adds the sound `.a` fixture above and any separate negative fixture needed; source Node, native with ASan/UBSan/leak checks, and the JavaScript backend match stdout, stderr and exit status. Run only its focused oracle selector, writing output to a log.
3. Runs this mutant independently and records its intended catcher: Drop the callback when present; the true case must disagree with Node while the absent case still prints false.
4. Refreshes counts with `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts`, writes output to a log, and explains every changed row.
5. Commits its fix and fixtures, pushes its own branch once after local checks, and reports the SHA.
6. Re-measures the same hash-pinned largest region with the hidden census and reports old hidden intersection, new hidden intersection and their byte difference. Report the next boundary too. Do not report the group total as revealed bytes.

Revealed bytes in this documentation unit: **not measured**; no boundary implementation changed.
