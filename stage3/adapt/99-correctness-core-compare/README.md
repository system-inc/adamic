# compareComparableValues compares like with like

Compiler amendment, task b5w3ycg: the unreachable mixed arm fails loudly. One amendment commit follows 3cf7ef3e29ab4ec52078d6f85e3545cbd0c6d45c. Both full applies and all affected proofs were rerun; the accepted lane count is retained unchanged. AMENDMENT.json and amendment-evidence.tar.gz are the current proof.

JavaScript-changing, count-authorized established-shape adaptation for task b5w3ycg, wave row L1, site F18 at adapted core.ts:2003:9. Upstream is TypeScript 6.0.3, 050880ce59e30b356b686bd3144efe24f875ebc8. Base is origin/main f0271c6a7b2e92900eb7046d8246cfb89f52cfb6.

Adamic, task b5w3ycg, Nov 7 03:20Z: established-shape adaptations, Compiler accepts.

The two overloads and the first three conditional arms are unchanged. The final comparison splits on both operands' typeof: numbers use their own `<`, and strings use their own `<`. Source line count is preserved. The third arm calls `Debug.fail("compareComparableValues: a string compared with a number")` for a mixed string/number pair. Debug is already imported, and Debug.fail returns never. It is reached on 0 of the 144,697,378 counted calls. The overloads forbid that pair, so a well-typed tsc caller cannot reach it. The complete lane count found 0 mixed calls. Deliberate mixed Node witnesses use the actual Debug.fail implementation and show the original coerced LessThan answer against the new thrown Debug Failure; this off-count class has count 0 and is recorded in the ledger.

Equal operands, undefined on either side, ordinary numbers, strings with different case, and NaN have equal before/after witnesses. NaN remains GreaterThan, including NaN compared with itself. Every observed operand-type pair is witnessed. No new cast, import, lib change, compiler overlay for ordinary checks, or production lowering change is introduced.

## Count before construction

All 398 unchanged lane units pass their normal identity, baseline, API and count guards: 106,366 passing tests, 1 sanctioned failure, 0 skipped. The separate unchanged `stage3/lane/run.sh` control passes, with the identical test-identity multiset, counts and raw baseline diff. Counted calls total **144,697,378**, with **0 mixed calls and 0 result differences**. No adapter was constructed until the count, independent reconciliation and whole-lane control completed.

| typeof pair | Calls |
| --- | ---: |
| number / number | 79,569,686 |
| string / string | 64,252,980 |
| undefined / undefined | 868,781 |
| string / undefined | 5,879 |
| undefined / string | 52 |
| number / undefined | 0 |
| undefined / number | 0 |
| Mixed string and number | 0 |

The retained original measurement computed the original and initially accepted typeof candidate on the same operands, recorded both results, and returned the original result to tsc. This amendment changes only its mixed arm, which those type-pair records reach 0 times. The lane count was not rerun; COUNT-REUSE.json binds the unchanged count and proves Debug.fail is reached on 0 of the 144,697,378 calls. The scratch observer is absent from adapt.cjs. The frozen source/built manifest and recorder hash bind the measurement. A byte records four base-3 tags: typeof a, typeof b, original Comparison, adapted Comparison. BINARY-SCHEMA.json describes the lossless encoding of these required fields. Every call is written synchronously, including calls before an intentional process termination. Mixed pairs or differences also retain full operand values in a sidecar and stop the count.

The compact recorder matches the original JSON observer on all 676 excluded seed calls. Earlier JSON attempts hit unchanged timing guards and are excluded and archived. An additional real-call comparison verifies all 16,586,589 ordered JSON records from the first passing JSON piece against the original observer. The accepted census uses the single frozen compact recorder throughout. Piece 1 used ordinary files; piece 2 retains 19 passing ordinary-file units and the isolated memory-backed retry of reallyLargeFile.ts, which passed in 18.82 seconds. Later pieces use two concurrent units with synchronous memory-backed files copied into the archive after each unit. The runner's 90-second unit bound and upstream 40-second test bound are unchanged. Seeds and failed attempts are excluded from the canonical 398-unit union.

RECONCILED.json reopens every archive, verifies every raw member hash, reconstructs the type/result matrix and validates every report against the unchanged lane manifest. COUNT.json retains piece provenance and the isolated retry. CONTROL.json records the whole-lane match.

## Proof

| Check | Result |
| --- | --- |
| Main full apply | PASS |
| Main plus compare full apply | PASS |
| Paired replay of every adapter | PASS; all other outcomes identical, compare applies once before and performs zero edits after |
| Owner contract | PASS; exact full overload owner pins, no-write drift refusals, LF/CRLF and foreign edits preserved, replay bytes and mtime preserved |
| Adamic types | PASS; all 79 roots, 0 diagnostics before and after |
| Source scope | All 746 source files checked; only the pinned owner differs |
| Emitted JavaScript | All 9 files checked; only _tsc.js, typescript.js differ, only at the owner |
| Public declarations | typescript.d.ts and tsserverlibrary.d.ts byte-identical |
| Full core.ts census | F18 removed; 59 to 58 distinct sites; no new site anywhere; all 3 overload probes attempted |
| Original-owner mutant | Restores the entire baseline site set, including F18 |

JAVASCRIPT-DIFF.md shows every JavaScript hunk. The full census uses the same LATENT_FULL, LATENT_ONLY_FILE, LATENT_ASSERT_NO_OUTPUT and timing procedure as core-sites-4186581e. Raw findings, timings, manifests, both type logs, all replay outcomes, exact diffs and tool provenance are archived. Every retained Go source is named .go.txt. Evidence-only file updates can change the product cache key without changing the frozen transformation.

The initial ownership scan found no earlier owner pin. Matches in axolotl-j are checker.ts call-site edits. The adapter pins the entire three-declaration owner, refuses missing, duplicate, nested, drifted or half-adapted owners before a write, and edits nothing outside that owner. Replay performs zero writes.

Every new launcher records the child's actual return code and propagates unexpected failures. The exit-code canary proves child 7 remains wrapper 7. Expected nonzero validation subjects, including diff status 1 and matching old-adapter refusals, retain their own codes. Initial construction and measurement helpers are historical; amendment-apply.py and amendment-proof.py are the current rerun entry points.
