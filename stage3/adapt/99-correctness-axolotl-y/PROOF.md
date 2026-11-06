# 99y proof

Base and ordinary compiler: origin/main
322f80565edaa978d928ca559803c578ed67c105. Upstream source:
050880ce59e30b356b686bd3144efe24f875ebc8, stock TypeScript 6.0.3.
No compiler overlay, refusal suppression, backend change, cast, or runtime
check is introduced. provenance.json identifies the compiler binaries.

MEASURE.md and raw checker results were written before the adapter was created.
Stock TypeScript already infers combine's own T at both append sites. The edit
spells that type argument explicitly at each site. The isolated ordinary binder
census moves from core.ts:966:72 to core.ts:1049:34.

| Check | Result |
| --- | --- |
| Prescribed setup and API npm ci | Exit 0 |
| Full apply before and after, --build-product | Both exit 0 |
| Paired replay, all 82 adapters, continuing after refusals | All 81 non-99y outcomes identical |
| 99y replay outcomes | Before: 1 file, 2 edits; after: 0 files, 0 edits |
| Both completed replay source trees | Equal accepted source, all 746 files |
| Owner contract | Ten no-write refusals; LF/CRLF and unrelated edits preserved |
| Idempotence | Zero replay writes; bytes and mtime preserved |
| Adamic types, all 79 roots together after | Exit 0, zero diagnostics |
| Ordinary compiler builds before and after | Both exit 0 |
| Complete core.ts JavaScript, comments kept and removed | Byte-identical |
| All nine emitted JavaScript files | Byte-identical |
| Both public declarations | Byte-identical |
| Complete source comparison | Only the two pinned core.ts type arguments differ |
| Original-spelling mutant | Exact core.ts:966:72 refusal reproduced |
| Independent-root census, 79 roots | Complete; groups of 20, 20, 20, 19 each committed and pushed |

comparison.json records complete-output hashes and sizes as evidence, never
adaptation gates. Cast count stays 26 and any count stays 14. The mutant's
746 source files equal the baseline byte-for-byte. Both ordinary mutant rows
reproduce Refused: a value of type T[] seen as {}[], which can write {} where
T is read. mutant.patch restores only the two original calls in a copied tree.

The complete combine owner actually has four declarations, three overloads
and its implementation. All four are pinned, including the array-only overload
above the request's excerpt. Exact before/after owner text and SHA-256 gates
reject drift or a half edit before writing. They do not gate on whole-file hashes.

Ordering constraint: the required directory sorts after all correctness
adapters but before 99-wasm-crypto-import. This follows 99w's convention.
It cannot sort last overall without changing the required directory name or
apply.py, outside the permitted commit scope. This deviation was surfaced during
work. sort.json records the exact real Python order, without relabeling it last.

Reproduce using /workspace/adamic-tools/env.sh, the pinned upstream source,
and NODE_PATH=stage3/api/node_modules. apply-results.json and checks.json retain
commands and exits; verify.py, replay.py, and run-groups.py retain the drivers.
Full logs are retained compressed. The before and after full applies each froze
their adaptation list; adding later evidence changes product keys, not the
transformation. The oracle is skipped under the type-only exception, supported
by the complete JavaScript byte comparisons.

An initial compiler build attempted before the nested checkout was ready failed
on its missing go.mod. The unchanged build passed after that checkout completed.
The early failure is retained in build-main-before-setup.log.gz. Group 1 logging
was archived early; its complete 40-line log was recovered from the open file
descriptor before it closed. All group logs and raw row pairs were validated.
