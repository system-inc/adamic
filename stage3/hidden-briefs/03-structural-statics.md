# 03. a method call through a structural signature in a program with statics; use typeof the declaring class

Historical boundary: `a method call through a structural signature in a program with statics; use typeof the declaring class`. Raising function: `callOrMethod` in `internal/lower/class.go`. The template is present on origin/main `89ac4a8c1de0b02d95965be72f7f8bf1c92433d2`.

**13,111 hidden bytes** across 2 selected region(s):

- #39: `transformers/module/module.ts` [16356, 23548) = 7,192 bytes
- #55: `moduleSpecifiers.ts` [67576, 73495) = 5,919 bytes

These are assigned head bytes within ranks 11–60, not a whole-corpus total or a promise that removing this stop reveals them all. The smallest recorded boundary containing the head wins. Enclosing blockers remain recorded in TABLE.md.

Largest region head: `transformers/module/module.ts:16356` (UTF-8 byte offset). Recorded diagnostic: `transformers/module/module.ts:416:27`.

Recorded attempt owner: `/tmp/hidden-adapted/src/compiler/transformers/module/module.ts:413:5`. Census stopping mode: `failed statement; state rolled back; continued at next statement or declaration`; boundary span [16356, 21300).

```text
/tmp/hidden-adapted/src/compiler/transformers/module/module.ts:416:27: stage 0 can't lower a method call through a structural signature in a program with statics; use typeof the declaring class yet
```

Replay selector for that head, using the replay tool inherited from `codex/stage3-census-replay`, already present at census pin `388096e6`, with that pin's adapted source:

```sh
source /workspace/adamic-tools/env.sh
# Run from the replay worktree prepared as described in TABLE.md.
go run ./stage3/census/latent/replay \
  -project /tmp/hidden-adapted/src/compiler \
  -where /tmp/hidden-adapted/src/compiler/transformers/module/module.ts:416:27 -kind NotYet \
  -reason 'a method call through a structural signature in a program with statics; use typeof the declaring class' > /tmp/hidden-boundary-03-replay.log 2>&1
```

Observed full-project replay: **timed out at 180s**, with no reproduction result. Log: `/tmp/hidden-briefs/replay-03.log`. The current-main reduction below reproduces the same reason and raising function; the exact census-position replay remains unverified.

Verified current-main reduction, to save as the proposed fixture below. Ran `/tmp/hidden-briefs/adamic c /tmp/hidden-briefs/03.a > /tmp/hidden-briefs/reduction-03.log 2>&1`; exit 1 with this boundary reason. This reproduces the reason in its raising function, not the original census source position. Node behavior follows.

```typescript
interface Callable { read(): number; }
class StaticSource { static value = 7; static read(): number { return this.value; } }
function read(source: Callable): number { return source.read(); }
console.log(`${read(StaticSource)}`);
```

Node 24.19.0, after `node:module.stripTypeScriptTypes`, observed stdout:

```text
7
```

Exit 0; stderr empty. Raw observation: `/tmp/hidden-briefs/node-03.log`.

Use a proven representation and preserve Node evaluation order, identity and cleanup.

Proposed fixture path: `internal/oracle/testdata/hidden_boundary_structural_statics.a`. No oracle fixture was added by this documentation unit; the current scratch witness is `/tmp/hidden-briefs/03.a`.

Existing branch: `origin/codex/notyet-statics` removes this guard in `b6aa4f000eb7279da65353b2e39ee359c7171082` (Dispatch structural methods through the actual receiver). The diff routes structural calls through receiver lookup. Existing fixture: `internal/oracle/testdata/structural_statics_static.a` on that branch. Source-level fix identified; its backend tests were not rerun here. `compiler/input-spread` and `compiler/super-closure` also lack the reason, but absence alone is not a verified fix.

Done when the implementation unit:

1. Replays the exact head stop, records its log and keeps unsafe cases refused.
2. Adds the sound `.a` fixture above and any separate negative fixture needed; source Node, native with ASan/UBSan/leak checks, and the JavaScript backend match stdout, stderr and exit status. Run only its focused oracle selector, writing output to a log.
3. Runs this mutant independently and records its intended catcher: Dispatch every structural receiver through the instance method table; the static receiver oracle must disagree with Node.
4. Refreshes counts with `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts`, writes output to a log, and explains every changed row.
5. Commits its fix and fixtures, pushes its own branch once after local checks, and reports the SHA.
6. Re-measures the same hash-pinned largest region with the hidden census and reports old hidden intersection, new hidden intersection and their byte difference. Report the next boundary too. Do not report the group total as revealed bytes.

Revealed bytes in this documentation unit: **not measured**; no boundary implementation changed.
