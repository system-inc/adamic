# 99w compiler census

Compiler: unmodified main a28effb2c. Source: the completed full apply of main
plus 99w, exactly the tree checked by types, builds and comparison.json.
No overlay, suppression, backend change or partial-root fallback.

All 79 TypeScript compiler roots were measured independently.

| Roots | Kind | First stop | Exact reason |
| ---: | --- | --- | --- |
| 73 | Refused | `core.ts:810:89` | a cast the runtime can't check |
| 2 | accepted | `-` | - |
| 1 | NotYet | `debug.ts:243:59` | a generic instantiation whose type argument for T, constrained to Node, can't be read |
| 1 | NotYet | `semver.ts:71:9` | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |
| 1 | NotYet | `commandLineParser.ts:293:35` | reading an enum before its runtime initialization; move the call after the enum declaration |
| 1 | NotYet | `diagnosticInformationMap.generated.ts:13:45` | reading an enum before its runtime initialization; move the call after the enum declaration |

The 73 roots still stop at core.ts:810:89. The bridge is now `as unknown as`,
but the exact ordinary compiler reason remains `a cast the runtime can't check`.
No root movement or accepted per-overload erasure is claimed. Optimizer can use
the unchanged overloads quoted in README.md to establish that proof separately.

mutant.cjs restores only `as any as` in an isolated copy. mutant.patch records
the one-line control. mutant-comparison.json verifies every one of its 746 source
files equals baseline. mutant-binder.jsonl returns to core.ts:810:89, Refused,
with the same exact reason `a cast the runtime can't check` as baseline-binder.jsonl.

census-results.json retains every independent TypeScript root and its timing.
census-after.jsonl.gz and census-after.log.gz retain all raw rows, including
non-TypeScript inputs and the ordinary whole-program single-entry limitation.
Those extra rows are excluded from the 79-root summary. The census exits 0.
