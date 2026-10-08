# For-of object roots

Base `b410340dc8f889b5799c3bc519117c63def3aa24`; replay merge
`c68b6ceb0bd43283c6b919f8bc2c823084bd69e2` (includes `9a1f14c5`).
**0 of the 118 object-iteration roots lowered.** The named examples reproduce.

## Observed scope

The raw `roots/raw.csv` on `origin/codex/stage3-notyet-table` has 128 attempt rows
with exact reason `for...of over an object`, deduplicating to 118 diagnostic sites.
Stock TypeScript 6.0.3's checker on the complete adapted tsc entry identifies:

| Type | Sites |
| --- | ---: |
| NodeArray | 103 |
| JSDocArray | 6 |
| MutableNodeArray | 1 |
| Iterable<T> | 4 |
| Iterable<Symbol> | 1 |
| ElaborationIterator | 1 |
| readonly string tuple | 1 |
| SortedReadonlyArray<EmitHelper> | 1 |

[Per-site checker observations](site-types.csv). All 81 source SHA-256 values match
`stage3/meter/runs/20261008T035244Z.latent-full/tsc/source-manifest.json`.
These are source-type observations, not proof of native compilation.

Both examples are NodeArray: binder:432:37 is `NodeArray<Statement>`;
binder:1296:36 is `NodeArray<Expression>`. They are not Map or Set values.
The current compiler already lowers concrete custom iterables through iteration.go.
Its receiver, hidden-return and generic-origin guards were left intact.

## Why the two examples stay stopped

NodeArray extends ReadonlyArray and carries pos, end, hasTrailingComma and
transformFlags. Adamic represents that interface as an object, but the library
array iterator is inherited, not an own closure field. Treating the object pointer
as an array pointer would invent a native layout. Treating the inherited iterator
as an own field would invent a runtime method. Neither is sound.

The reductions keep that interface shape and the source loop, using an arrow to
make the loop lower before the unrelated metadata construction. Node prints `2`
and `true`, respectively. TestForOfObjectNodeArrayStops pins the exact NotYet and
nil IR. Its independent mutant replaces the object stop with an empty statement
list. Both cases fail the stop assertion; the later array/object intersection
construction becomes the next stop. Production code is restored after the run.

The requested release/sanitized native and JavaScript comparisons cannot run for
these reductions: lowering returns no IR. They are explicitly registered as
stopped fixtures, with separate source-Node checks. No green backend comparison
or completed object lowering is claimed.

The design question is whether NodeArray's array metadata construction and writes
are admitted as fixed array storage, or adapted to an object containing an array.
`docs/0.1.md` refuses expandos and prototype mutation by design. A ruling admitting
fixed metadata on arrays would need a representation and ownership design, plus
all metadata accesses and construction held to Node. This unit does not weaken
that refusal or assume that every structural Iterable has an array layout.
The five structural Iterable roots additionally need protocol dispatch that
preserves built-in, class and closure receiver conventions and optional return.
The tuple and SortedReadonlyArray roots need their own storage proofs.

## Commands and results

Every test writes to a log; no whole package or full gate was run.
Environment: Node 24.19.0, Go 1.27.1, clang 20.1.8; nproc 5, CPU quota 4.
`export GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh` succeeded.
Timing lines: node 0.059s, go 0.071s, clang 0.462s, markdown 1.068s,
submodules 15.407s, go build 292.079s, cache warm 292.188s, done 292.221s.
Source `/workspace/adamic-tools/env.sh` in each build/test shell.

```
go run ./stage3/census/latent/replay -project /tmp/for-of-adapted/src/tsc/tsc.ts -where /tmp/for-of-adapted/src/compiler/binder.ts:432:37 -kind NotYet -reason 'for...of over an object'
go run ./stage3/census/latent/replay -project /tmp/for-of-adapted/src/tsc/tsc.ts -where /tmp/for-of-adapted/src/compiler/binder.ts:1296:36 -kind NotYet -reason 'for...of over an object'
```

Both exit 0 and print `reproduced NotYet: for...of over an object`.
Total times 125.841s and 120.009s include cold overlay compilation overlapping setup;
load/register/lower times are 4.879s and 4.341s.
The first unit also stops at binder:423:18 `a BinaryExpression with a value and a value`
and binder:424:9 `reading name`. The second also refuses the unchecked cast at 1304:42.

```
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestForOfObjectNodeArrayStops|TestNativeAgreesWithNode/internal/oracle/testdata/for_of_object_stopped' -count=1 -timeout 10m
```

Exit 0, oracle 0.230s. The omit-object-stop mutant exits 1 with the intended
`want the NodeArray representation stop and no IR` assertion.
Logs: `/tmp/for-of-setup.log`, `/tmp/for-of-apply.log`,
`/tmp/for-of-before-{432,1296}.log`, `/tmp/for-of-stopped.log`,
`/tmp/for-of-stop-mutant.log`, `/tmp/for-of-counts-step2.log`.
