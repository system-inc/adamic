# 01. A method in object destructuring

Roadmap step 30. Historical rank 14: `Refused: a method in object destructuring`, 34,480 credited hidden bytes across 12 contributing boundaries. Delivery base `4885cec50290686df487b62aac47c85d871ed40c`; corrected ranking `ec0b16c04f3bbdfbaf3932ab01307f90b08156fd`, `stage3/census/hidden-ranking/RESULT.json`. Adapted sources are hash-pinned by `stage3/census/hidden/RESULT.json` at the ranking pin, matching all 82 files prepared from census pin `388096e6`.

## Current boundary and replay

Raising function on this base: `destructureFrom`, `internal/lower/collections.go:281`; its method-symbol guard at line 328 refuses extraction because it could lose `this`. Both the historical sys.ts head and the largest contiguous credited region reproduce the exact refusal on this base. No replacement first stop was observed.

Largest contiguous credited region: `utilities.ts [455532, 462634)`, 7,102 UTF-8 bytes. Its enclosing recorded boundary is `[455254, 463614)`, attempted unit `utilities.ts:11316:1`, diagnostic `utilities.ts:11316:35`. The sys.ts boundary `[23159, 36459)` credits 10,405 fragmented bytes; it is a larger credited boundary, not the largest contiguous region.

From an assigned-base worktree with the pin's adapted sources and every manifest hash verified:

```sh
source /workspace/adamic-tools/env.sh
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/hidden-wave-2/overlay > /tmp/hidden-wave-2/overlay.log 2>&1
go build -buildvcs=false -overlay=/tmp/hidden-wave-2/overlay/overlay.json -o /tmp/hidden-wave-2/replay ./stage3/census/latent/replay/worker > /tmp/hidden-wave-2/build.log 2>&1
GOMAXPROCS=1 /tmp/hidden-wave-2/replay -project /tmp/hidden-adapted/src/compiler -where /tmp/hidden-adapted/src/compiler/utilities.ts:11316:35 -kind Refused -reason 'a method in object destructuring' > /tmp/hidden-wave-2/replay-14-largest.log 2>&1
```

Observed: exit 0, exact position/kind/reason reproduced; worker elapsed 2.200962132s. The second selector `sys.ts:597:5` with the same kind/reason also reproduced, exit 0 (`/tmp/hidden-wave-2/replay-14.log`). The replay is marked “measured on a checker-rejected program”; it is a latent observation, not whole-project acceptance. Replay keeps the full project and ancestor context; a short witness alone does not establish census reproduction.

## Assigned credited regions

Adjacent attributed spans have been merged below; all intervals are half-open UTF-8 byte offsets. These 47 regions total 34,480 credited bytes. Attribution is a historical estimate, not measured revealed bytes. Enclosing blockers remain in the pinned ranking.

| File | Interval | Bytes |
| --- | --- | ---: |
| moduleSpecifiers.ts | [26372, 26845) | 473 |
| moduleSpecifiers.ts | [27195, 32030) | 4,835 |
| sys.ts | [23159, 24305) | 1,146 |
| sys.ts | [24589, 26011) | 1,422 |
| sys.ts | [26032, 26082) | 50 |
| sys.ts | [26180, 26236) | 56 |
| sys.ts | [26317, 26399) | 82 |
| sys.ts | [26435, 26481) | 46 |
| sys.ts | [26508, 26929) | 421 |
| sys.ts | [26964, 27227) | 263 |
| sys.ts | [27352, 29147) | 1,795 |
| sys.ts | [29274, 29895) | 621 |
| sys.ts | [30023, 30579) | 556 |
| sys.ts | [30627, 30675) | 48 |
| sys.ts | [30804, 31606) | 802 |
| sys.ts | [31727, 32205) | 478 |
| sys.ts | [32253, 32400) | 147 |
| sys.ts | [32706, 32828) | 122 |
| sys.ts | [32965, 33009) | 44 |
| sys.ts | [33185, 35340) | 2,155 |
| sys.ts | [35853, 35930) | 77 |
| sys.ts | [35941, 35948) | 7 |
| sys.ts | [36385, 36449) | 64 |
| sys.ts | [36456, 36459) | 3 |
| sys.ts | [41507, 42126) | 619 |
| sys.ts | [42285, 44048) | 1,763 |
| sys.ts | [44105, 44221) | 116 |
| sys.ts | [44285, 44407) | 122 |
| sys.ts | [47300, 47730) | 430 |
| sys.ts | [48127, 49778) | 1,651 |
| sys.ts | [52272, 53289) | 1,017 |
| sys.ts | [53612, 53646) | 34 |
| sys.ts | [53716, 55473) | 1,757 |
| sys.ts | [55603, 57847) | 2,244 |
| sys.ts | [59629, 59636) | 7 |
| sys.ts | [59868, 60420) | 552 |
| sys.ts | [60427, 60430) | 3 |
| transformers/es2016.ts | [534, 618) | 84 |
| transformers/es2020.ts | [1028, 1112) | 84 |
| transformers/es2021.ts | [725, 809) | 84 |
| transformers/legacyDecorators.ts | [2053, 2181) | 128 |
| transformers/module/system.ts | [3661, 3813) | 152 |
| transformers/typeSerializer.ts | [4973, 5057) | 84 |
| utilities.ts | [455254, 455450) | 196 |
| utilities.ts | [455532, 462634) | 7,102 |
| utilities.ts | [462889, 463402) | 513 |
| utilities.ts | [463589, 463614) | 25 |

## Node witnesses

Proposed positive fixture: `internal/oracle/testdata/hidden_wave2_method_destructuring.a`.

```typescript
const host = { increment(value: number): number { return value + 1; } };
const { increment } = host;
console.log(`${increment(4)}`);
```

Node 24.19.0, using `node:module.stripTypeScriptTypes`, observed stdout `5\n`, stderr empty, exit 0. The assigned-base compiler rejects this reduction at line 2, column 9 with the same Refused reason, exit 1 (`/tmp/hidden-wave-2/reduction.log`). This is evidence that the receiver-independent case remains blocked, not a proof that the census site is independently lowerable.

Receiver-sensitive semantic witness, proposed separate fixture `internal/oracle/testdata/hidden_wave2_method_destructuring_receiver.a`:

```typescript
const host = {
    value: 7,
    read(this: { value: number } | undefined): number {
        return this === undefined ? -1 : this.value;
    },
};
const { read } = host;
console.log(`${read()} ${host.read()}`);
```

Node observed stdout `-1 7\n`, stderr empty, exit 0. Destructuring extracts the function; it does not bind the original receiver. Automatically wrapping every method to call on its holder is unsound. The explicit-this witness is a Node observation, not a claim of current compiler acceptance. Preserve the refusal for cases whose unbound receiver behavior cannot be represented soundly.

Scratch witnesses are `/tmp/hidden-wave-2/method.a` and `receiver.a`. Observation command (Node warnings suppressed for this API):

```sh
NODE_NO_WARNINGS=1 node --input-type=module -e "import {readFileSync} from 'node:fs'; import {stripTypeScriptTypes} from 'node:module'; for (const path of ['/tmp/hidden-wave-2/method.a','/tmp/hidden-wave-2/receiver.a']) { await import('data:text/javascript,'+encodeURIComponent(stripTypeScriptTypes(readFileSync(path,'utf8')))); }" > /tmp/hidden-wave-2/node.stdout 2> /tmp/hidden-wave-2/node.stderr
```

Observed combined stdout `5\n-1 7\n`, stderr 0 bytes, exit 0. No oracle fixture was added in this documentation unit.

## Ownership and soundness

The existing `codex/notyet-destructuring` unit (`903ba7b229bad116798714b05a0ef93164da7316`, `stage3/notyet-destructuring/REPORT.md`) owns computed keys and binding patterns, and shares `destructureFrom`. Its completed lessons do not include receiver-proof method extraction. This brief conservatively assigns that remaining stop; coordinate that function after its changes land, without merging an unlanded worker branch. Views, generic returns and never-array representation remain their owners' work.

Prove receiver independence or preserve unbound-call semantics, evaluation order and extracted callable identity. Do not infer receiver independence merely from a missing explicit TypeScript `this` parameter. Keep unsupported receiver-sensitive cases refused. Do not convert the whole remaining project into accepted code to bypass another owned boundary.

## Done when

1. Replay the exact largest-region head on the implementation base with the hash-pinned source, log the result and identify any new first stop. Keep unsafe cases refused.
2. Add the positive `.a` fixture and a receiver-sensitive semantic or separate pinned negative fixture. Run only their focused oracle selector, recording commands and output to a log. Source Node, JavaScript backend, and native with ASan/UBSan/leak checks must match stdout, stderr and exit status for accepted cases.
3. Independently run a mutant that binds an extracted method to its original holder. The receiver oracle must catch `7 7` instead of Node's `-1 7`. If that case remains unsupported, pin its refusal and run a mutant that drops the receiver proof: the negative test must catch incorrect acceptance. Record the intended catcher and actual failure; a compiler that still refuses before the mutant is exercised is not a caught semantic mutant.
4. Refresh counts with `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts`, sending output to a log and explaining every changed row.
5. Commit implementation and fixtures on the implementation's own branch, then push once after focused checks and mutants pass. Report the SHA; no PR.
6. Re-measure the same hash-pinned `utilities.ts [455532, 462634)` with the hidden census. Report its old hidden intersection, new hidden intersection and byte difference, plus the next boundary. Recompute the actual old intersection from the pinned hidden manifest rather than equating credited attribution with hidden coverage. Do not report 34,480 as revealed bytes because a reason disappears.

Revealed bytes in this documentation unit: **not measured**. No boundary implementation changed.
