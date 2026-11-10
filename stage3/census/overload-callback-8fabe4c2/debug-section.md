## debug.ts, full mode

Both streams load the same checker-clean tsc entry graph (81 source files), then select only debug.ts for full-mode observation. Both end with a relation trailer and child exit 0. There are 81 debug units on each side. At locations in debug.ts itself, raw events are 20 -> 12 and unique findings are 17 -> 11; the remaining two unique findings are in core.ts. The full event streams, including two unchanged dependency findings, contain 22 -> 14 findings; unique raw findings and distinct reads are both 19 -> 13, including Boundary markers. Refused findings are 3 -> 1, NotYet findings remain 5, and Boundary markers are 11 -> 7. Two callback refusal locations and four continuation Boundary markers clear. There are no moved reasons, new findings, new sites, panics or regressions in this full-mode pair. Four attempted units become finding-free.

The predicted direction holds: assertNode and assertOptionalNode clear their callback refusal, while assertEachNode keeps it. The observed callback locations are 273:82 and 299:78, each emitted in two attempting-unit contexts. This full census therefore observes two distinct callback locations clearing, rather than three. The source has three narrowing callback overload signatures across these two functions (one assertNode and two assertOptionalNode), but the overload-group census returns at the first failed overload, so the baseline does not emit a separate overload-2 diagnostic for assertOptionalNode. The source evidence is retained in parent-census-small.go.txt and head-census-small.go.txt. Generic signature declarations remain deferred and are not counted as independently lowered bodies.

Every assert-family finding site follows. Boundary rows are the meter's continuation markers, retained to make the raw pairing fully accountable. Locations are relative to src/compiler/debug.ts; reason strings are quoted in full.

| Function | Site | Parent | Head |
| --- | --- | --- | --- |
| assertNever | 252:5 | "Boundary: function signature or prologue failed; continued at next declaration" | "Boundary: function signature or prologue failed; continued at next declaration" |
| assertNever | 252:33 | "NotYet: a value of type never" | "NotYet: a value of type never" |
| assertEachNode | 257:86 | "Refused: overload 1 of assertEachNode parameter test cannot be served by implementation parameter test" | "Refused: overload 1 of assertEachNode parameter test cannot be served by implementation parameter test" |
| assertEachNode | 261:5 | "Boundary: function signature or prologue failed; continued at next declaration" | "Boundary: function signature or prologue failed; continued at next declaration" |
| assertEachNode | 262:5 | "Boundary: function signature or prologue failed; continued at next declaration" | "Boundary: function signature or prologue failed; continued at next declaration" |
| assertNode | 273:82 | "Refused: overload 1 of assertNode parameter test cannot be served by implementation parameter test" | None |
| assertNode | 274:5 | "Boundary: function signature or prologue failed; continued at next declaration" | None |
| assertNode | 275:5 | "Boundary: function signature or prologue failed; continued at next declaration" | None |
| assertOptionalNode | 299:78 | "Refused: overload 1 of assertOptionalNode parameter test cannot be served by implementation parameter test" | None |
| assertOptionalNode | 301:5 | "Boundary: function signature or prologue failed; continued at next declaration" | None |
| assertOptionalNode | 302:5 | "Boundary: function signature or prologue failed; continued at next declaration" | None |

Every assert-family declaration, including finding-free siblings and deferred signatures:

| Function | Declaration | Eligibility, parent -> head | First non-Boundary stop, parent | First non-Boundary stop, head |
| --- | --- | --- | --- | --- |
| assert | 192:5 | attempted -> attempted | None | None |
| assertEqual | 202:5 | deferred_generic_instantiation -> deferred_generic_instantiation | None | None |
| assertLessThan | 209:5 | attempted -> attempted | None | None |
| assertLessThanOrEqual | 215:5 | attempted -> attempted | None | None |
| assertGreaterThanOrEqual | 221:5 | attempted -> attempted | None | None |
| assertIsDefined | 227:5 | deferred_generic_instantiation -> deferred_generic_instantiation | None | None |
| assertEachIsDefined | 239:5 | deferred_generic_instantiation -> deferred_generic_instantiation | None | None |
| assertEachIsDefined | 240:5 | deferred_generic_instantiation -> deferred_generic_instantiation | None | None |
| assertEachIsDefined | 241:5 | deferred_generic_instantiation -> deferred_generic_instantiation | None | None |
| assertNever | 252:5 | attempted -> attempted | "NotYet: a value of type never" | "NotYet: a value of type never" |
| assertEachNode | 257:5 | deferred_generic_instantiation -> deferred_generic_instantiation | None | None |
| assertEachNode | 258:5 | deferred_generic_instantiation -> deferred_generic_instantiation | None | None |
| assertEachNode | 259:5 | deferred_generic_instantiation -> deferred_generic_instantiation | None | None |
| assertEachNode | 260:5 | deferred_generic_instantiation -> deferred_generic_instantiation | None | None |
| assertEachNode | 261:5 | attempted -> attempted | "Refused: overload 1 of assertEachNode parameter test cannot be served by implementation parameter test" | "Refused: overload 1 of assertEachNode parameter test cannot be served by implementation parameter test" |
| assertEachNode | 262:5 | attempted -> attempted | "Refused: overload 1 of assertEachNode parameter test cannot be served by implementation parameter test" | "Refused: overload 1 of assertEachNode parameter test cannot be served by implementation parameter test" |
| assertNode | 273:5 | deferred_generic_instantiation -> deferred_generic_instantiation | None | None |
| assertNode | 274:5 | attempted -> attempted | "Refused: overload 1 of assertNode parameter test cannot be served by implementation parameter test" | None |
| assertNode | 275:5 | attempted -> attempted | "Refused: overload 1 of assertNode parameter test cannot be served by implementation parameter test" | None |
| assertNotNode | 286:5 | deferred_generic_instantiation -> deferred_generic_instantiation | None | None |
| assertNotNode | 287:5 | attempted -> attempted | None | None |
| assertNotNode | 288:5 | attempted -> attempted | None | None |
| assertOptionalNode | 299:5 | deferred_generic_instantiation -> deferred_generic_instantiation | None | None |
| assertOptionalNode | 300:5 | deferred_generic_instantiation -> deferred_generic_instantiation | None | None |
| assertOptionalNode | 301:5 | attempted -> attempted | "Refused: overload 1 of assertOptionalNode parameter test cannot be served by implementation parameter test" | None |
| assertOptionalNode | 302:5 | attempted -> attempted | "Refused: overload 1 of assertOptionalNode parameter test cannot be served by implementation parameter test" | None |
| assertOptionalToken | 313:5 | deferred_generic_instantiation -> deferred_generic_instantiation | None | None |
| assertOptionalToken | 314:5 | deferred_generic_instantiation -> deferred_generic_instantiation | None | None |
| assertOptionalToken | 315:5 | attempted -> attempted | None | None |
| assertOptionalToken | 316:5 | attempted -> attempted | None | None |
| assertMissingNode | 327:5 | attempted -> attempted | None | None |
| assertMissingNode | 328:5 | attempted -> attempted | None | None |

Other debug.ts sites that move: **none**. The unchanged non-assert debug findings are listed here for completeness.

| Function | Site | Parent | Head |
| --- | --- | --- | --- |
| formatSymbol | 350:9 | "Boundary: failed statement; state rolled back; continued at next statement or declaration" | "Boundary: failed statement; state rolled back; continued at next statement or declaration" |
| formatSymbol | 350:135 | "NotYet: a template interpolating an object, an array, a map, a function or undefined" | "NotYet: a template interpolating an object, an array, a map, a function or undefined" |
| getEnumMembers | 400:13 | "Boundary: failed statement; state rolled back; continued at next statement or declaration" | "Boundary: failed statement; state rolled back; continued at next statement or declaration" |
| getEnumMembers | 400:27 | "NotYet: a readonly view read by a run-time key in a program with a class getter that does more than return its fields and literals joined by operators (the getter runs inside the read, unseen by the passes that follow calls)" | "NotYet: a readonly view read by a run-time key in a program with a class getter that does more than return its fields and literals joined by operators (the getter runs inside the read, unseen by the passes that follow calls)" |
| getEnumMembers | 401:13 | "Boundary: failed statement; state rolled back; continued at next statement or declaration" | "Boundary: failed statement; state rolled back; continued at next statement or declaration" |
| getEnumMembers | 401:24 | "NotYet: reading value" | "NotYet: reading value" |

Full-mode dependency findings: the comparator finding at core.ts:1049:82 and its Boundary at core.ts:1049:5 are unchanged. A debug-only observation can report a dependency stop; these are not additional changed debug.ts sites.

The exact access-chain pair output is in [debug-paired.json](debug-paired.json) and [debug-paired.md](debug-paired.md).
