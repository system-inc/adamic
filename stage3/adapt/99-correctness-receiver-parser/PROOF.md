# Proof

JavaScript-changing, value-identical; #trxxkge: 46 parser callee rewrites call through factory, measured at 18,022,708 added per-use property reads per complete lane run.

Original measured proof base: `8ada94542fbdcb5c1d1689e27564e471eae241b8`. TypeScript v6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`. The unmodified accepted generator at `24c61e74d26caa79c0f313112902e325d090afcd` regenerated all 19 source files; its complete rewrite receipt is identical to the accepted one, and --verify reproduced every generated file byte-for-byte. Source replacements are never hand-edited. Pins use LF-normalized text and sha256; adapters preserve homogeneous input endings and reject mixed endings.

Every edited enclosing owner and changed namespace import is pinned whole before/after. Parser additionally pins all 34 changed nearest function owners, complete overload groups included, under its namespace pin. Prior owner greps found no existing whole-owner amendment for these edits. All 52 context references resolve to their own factory parameter.

## Complete single-file censuses

Each base/change run selects one whole file, LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, CommonJSBundle, unchanged checker options, zero checker diagnostics, and a complete relation trailer. Pairing is a multiset of (kind, phase, reason, message with source positions stripped). All survivors pair; new and changed findings outside targets are both zero. Source mapping resolves identical-key printer var multiplicities at line 1215. The seven hidden transformer sites are not claimed to have lowered bodies; full AST inventories independently prove their target bindings gone. The overlay's historical checker-rejected measurement label is fixed text; the actual headers have checker_rejected=false.

| File | Base | Built | Paired | Target records removed | New/changed outside |
| --- | ---: | ---: | ---: | ---: | ---: |
| parser.ts | 3183 | 3181 | 3181 | 2 | 0 |

Full pairing details are in [stage3/census/receiver-split-evidence/parser/proofs/](https://github.com/system-inc/adamic/tree/42f0c552bb93c523b6c5eb8b4681719cf0bc3197/stage3/census/receiver-split-evidence/parser/proofs) at `42f0c552bb93c523b6c5eb8b4681719cf0bc3197`census.json. Raw complete findings and logs stay in scratch with size/sha256 in the evidence branch manifests.

## Restoration and drift controls

The real mutant restores every one of the 79 compiler files byte-for-byte to the baseline, so its diagnostic oracle is the exact baseline full census above. Real-adapter controls verify creation, zero-write replay, restoration, drift/duplicate/partial-owner/mixed-ending refusals, and multi-file half-adaptation refusal, with zero writes on refusal. Guard and real-file receipts are under [stage3/census/receiver-split-evidence/parser/proofs/](https://github.com/system-inc/adamic/tree/42f0c552bb93c523b6c5eb8b4681719cf0bc3197/stage3/census/receiver-split-evidence/parser/proofs) at `42f0c552bb93c523b6c5eb8b4681719cf0bc3197`.

## Runtime observations

The complete counting lane has setter writes 0, typed-assignment writes 0, defineProperty writes 0, direct member identity reads 0, direct member length reads 0, and new-arrow length reads 0 (callees included). No late receiver registration or descriptor failure occurred. Implementation this reads are 0 by AST, including nested arrows and excluding independent nested ordinary functions; parse failures are 0. Both printer default implementations are separately checked. Positive controls run the actual instrumenter and prove direct and alias identity/length observations, callee length reads, writes, defineProperty, operation counts, and one deliberate AST this read. Printer controls prove once-only getter order and receiver, bare callback this, defaults and sentinel identity. Controls are outside lane snapshots.

The zero identity/length claim is direct-member/new-arrow scoped. Preserved printer sentinel identity comparisons and unrelated Function.bind length reads on shared primitives are reported separately in [stage3/census/receiver-split-evidence/parser/proofs/](https://github.com/system-inc/adamic/tree/42f0c552bb93c523b6c5eb8b4681719cf0bc3197/stage3/census/receiver-split-evidence/parser/proofs) at `42f0c552bb93c523b6c5eb8b4681719cf0bc3197`counts.json; no blanket zero is claimed. Counts are operations/initializer executions, not a timing forecast or physical-allocation claim.

## Required checks

- Stock tsc 6.0.3 API audit: 0 diagnostics on 79 explicit roots; CLI: exit 0 with strictFunctionTypes, noEmit, incremental/composite false.
- adamic types: exit 0 on those same 79 roots.
- Fresh stage3/apply.sh with only this new directory installed: exit 0, all 79 compiler files identical to the independent built tree.
- Stock emitted JavaScript differs only at the selected rewrite sites; complete masked files are identical.
- hereby dts: typescript.d.ts and tsserverlibrary.d.ts byte-identical to baseline, with sizes/sha256 in [stage3/census/receiver-split-evidence/parser/proofs/](https://github.com/system-inc/adamic/tree/42f0c552bb93c523b6c5eb8b4681719cf0bc3197/stage3/census/receiver-split-evidence/parser/proofs) at `42f0c552bb93c523b6c5eb8b4681719cf0bc3197`declarations.json.
- Ordinary and counting whole lanes: all runners, eight workers, no test filter, unchanged 40,000 ms hook timeout; 106,366 passing, one sanctioned public API acknowledgement row, zero pending. Sole baseline diff: api/typescript.d.ts. Locked install/build both exit 0.
- Counting copies preserve every tests/cases and tests/baselines/reference input byte-for-byte.

No Go source is added. No Loom submission.

Re-cut base: `c2f3c21f72b04e4c3acf7731913f0fc8ab89c2ca`. Runtime files and pins are byte-identical to the accepted first head. Evidence lives at `stage3/census/receiver-split-evidence/parser` on `typescript/receiver-split-evidence` at `42f0c552bb93c523b6c5eb8b4681719cf0bc3197`. This directory contains only the five house files; moved JSON is not a runtime dependency. New-stack verification receipts are published forward on the same evidence branch.
