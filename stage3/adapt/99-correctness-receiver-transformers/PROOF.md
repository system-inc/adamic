# Proof

JavaScript-changing, value-identical; #trxxkge: 52 exact-parameter forwarding arrows at 17 original factory scopes, with 189,211 forwarding calls and 1,601,433 arrow creations per complete lane run.

Original measured proof base: `8ada94542fbdcb5c1d1689e27564e471eae241b8`. TypeScript v6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`. The unmodified accepted generator at `24c61e74d26caa79c0f313112902e325d090afcd` regenerated all 19 source files; its complete rewrite receipt is identical to the accepted one, and --verify reproduced every generated file byte-for-byte. Source replacements are never hand-edited. Pins use LF-normalized text and sha256; adapters preserve homogeneous input endings and reject mixed endings.

Every edited enclosing owner and changed namespace import is pinned whole before/after. Parser additionally pins all 34 changed nearest function owners, complete overload groups included, under its namespace pin. Prior owner greps found no existing whole-owner amendment for these edits. All 52 context references resolve to their own factory parameter.

## Complete single-file censuses

Each base/change run selects one whole file, LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, CommonJSBundle, unchanged checker options, zero checker diagnostics, and a complete relation trailer. Pairing is a multiset of (kind, phase, reason, message with source positions stripped). All survivors pair; new and changed findings outside targets are both zero. Source mapping resolves identical-key printer var multiplicities at line 1215. The seven hidden transformer sites are not claimed to have lowered bodies; full AST inventories independently prove their target bindings gone. The overlay's historical checker-rejected measurement label is fixed text; the actual headers have checker_rejected=false.

| File | Base | Built | Paired | Target records removed | New/changed outside |
| --- | ---: | ---: | ---: | ---: | ---: |
| transformers/es2021.ts | 27 | 25 | 25 | 2 | 0 |
| transformers/classFields.ts | 639 | 639 | 639 | 0 | 0 |
| transformers/es2015.ts | 1118 | 1118 | 1118 | 0 | 0 |
| transformers/es2016.ts | 57 | 55 | 55 | 2 | 0 |
| transformers/es2017.ts | 392 | 390 | 390 | 2 | 0 |
| transformers/es2018.ts | 497 | 497 | 497 | 0 | 0 |
| transformers/es2020.ts | 152 | 150 | 150 | 2 | 0 |
| transformers/esDecorators.ts | 572 | 572 | 572 | 0 | 0 |
| transformers/esnext.ts | 256 | 254 | 254 | 2 | 0 |
| transformers/generators.ts | 528 | 528 | 528 | 0 | 0 |
| transformers/jsx.ts | 211 | 209 | 209 | 2 | 0 |
| transformers/legacyDecorators.ts | 300 | 298 | 298 | 2 | 0 |
| transformers/module/esnextAnd2015.ts | 122 | 120 | 120 | 2 | 0 |
| transformers/module/module.ts | 405 | 405 | 405 | 0 | 0 |
| transformers/module/system.ts | 419 | 417 | 417 | 2 | 0 |
| transformers/ts.ts | 674 | 674 | 674 | 0 | 0 |
| transformers/typeSerializer.ts | 97 | 95 | 95 | 2 | 0 |

Full pairing details are in [stage3/census/receiver-split-evidence/transformers/proofs/](https://github.com/system-inc/adamic/tree/42f0c552bb93c523b6c5eb8b4681719cf0bc3197/stage3/census/receiver-split-evidence/transformers/proofs) at `42f0c552bb93c523b6c5eb8b4681719cf0bc3197`census.json. Raw complete findings and logs stay in scratch with size/sha256 in the evidence branch manifests.

## Restoration and drift controls

The real mutant restores every one of the 79 compiler files byte-for-byte to the baseline, so its diagnostic oracle is the exact baseline full census above. Real-adapter controls verify creation, zero-write replay, restoration, drift/duplicate/partial-owner/mixed-ending refusals, and multi-file half-adaptation refusal, with zero writes on refusal. Guard and real-file receipts are under [stage3/census/receiver-split-evidence/transformers/proofs/](https://github.com/system-inc/adamic/tree/42f0c552bb93c523b6c5eb8b4681719cf0bc3197/stage3/census/receiver-split-evidence/transformers/proofs) at `42f0c552bb93c523b6c5eb8b4681719cf0bc3197`.

## Runtime observations

The complete counting lane has setter writes 0, typed-assignment writes 0, defineProperty writes 0, direct member identity reads 0, direct member length reads 0, and new-arrow length reads 0 (callees included). No late receiver registration or descriptor failure occurred. Implementation this reads are 0 by AST, including nested arrows and excluding independent nested ordinary functions; parse failures are 0. Both printer default implementations are separately checked. Positive controls run the actual instrumenter and prove direct and alias identity/length observations, callee length reads, writes, defineProperty, operation counts, and one deliberate AST this read. Printer controls prove once-only getter order and receiver, bare callback this, defaults and sentinel identity. Controls are outside lane snapshots.

The zero identity/length claim is direct-member/new-arrow scoped. Preserved printer sentinel identity comparisons and unrelated Function.bind length reads on shared primitives are reported separately in [stage3/census/receiver-split-evidence/transformers/proofs/](https://github.com/system-inc/adamic/tree/42f0c552bb93c523b6c5eb8b4681719cf0bc3197/stage3/census/receiver-split-evidence/transformers/proofs) at `42f0c552bb93c523b6c5eb8b4681719cf0bc3197`counts.json; no blanket zero is claimed. Counts are operations/initializer executions, not a timing forecast or physical-allocation claim.

## Factory scopes and creations

All 17 original statements and all 52 arrows are directly in their top-level transformer/serializer factory body, with context introduced by that factory parameter. None sits in a per-node visitor or a nested call scope, so no scope lift is eligible. factory remains a property binding. The original compiler creates no forwarding arrows at these bindings. No scope rewrite was made: the before/after decision columns describe the same generated artifact, using its observed site count in both columns. They are not claims of two separate runs. The accepted split likewise totaled 1,601,433 creations.

| File | Scope | Original tsc | Before scope decision | After scope decision | Calls/lane |
| --- | --- | ---: | ---: | ---: | ---: |
| transformers/ts.ts | factory invocation | 0 | 245995 | 245995 | 16794 |
| transformers/classFields.ts | factory invocation | 0 | 295194 | 295194 | 12423 |
| transformers/typeSerializer.ts | factory invocation | 0 | 402 | 402 | 38 |
| transformers/legacyDecorators.ts | factory invocation | 0 | 2054 | 2054 | 1627 |
| transformers/esDecorators.ts | factory invocation | 0 | 186612 | 186612 | 6469 |
| transformers/es2017.ts | factory invocation | 0 | 145124 | 145124 | 7645 |
| transformers/es2018.ts | factory invocation | 0 | 146512 | 146512 | 55851 |
| transformers/es2020.ts | factory invocation | 0 | 36871 | 36871 | 432 |
| transformers/es2021.ts | factory invocation | 0 | 37105 | 37105 | 36 |
| transformers/esnext.ts | factory invocation | 0 | 189796 | 189796 | 3387 |
| transformers/jsx.ts | factory invocation | 0 | 590 | 590 | 148 |
| transformers/es2016.ts | factory invocation | 0 | 36212 | 36212 | 228 |
| transformers/es2015.ts | factory invocation | 0 | 18770 | 18770 | 36312 |
| transformers/generators.ts | factory invocation | 0 | 18770 | 18770 | 4142 |
| transformers/module/module.ts | factory invocation | 0 | 193584 | 193584 | 25281 |
| transformers/module/system.ts | factory invocation | 0 | 1854 | 1854 | 2493 |
| transformers/module/esnextAnd2015.ts | factory invocation | 0 | 45988 | 45988 | 15905 |

## Required checks

- Stock tsc 6.0.3 API audit: 0 diagnostics on 79 explicit roots; CLI: exit 0 with strictFunctionTypes, noEmit, incremental/composite false.
- adamic types: exit 0 on those same 79 roots.
- Fresh stage3/apply.sh with only this new directory installed: exit 0, all 79 compiler files identical to the independent built tree.
- Stock emitted JavaScript differs only at the selected rewrite sites; complete masked files are identical.
- hereby dts: typescript.d.ts and tsserverlibrary.d.ts byte-identical to baseline, with sizes/sha256 in [stage3/census/receiver-split-evidence/transformers/proofs/](https://github.com/system-inc/adamic/tree/42f0c552bb93c523b6c5eb8b4681719cf0bc3197/stage3/census/receiver-split-evidence/transformers/proofs) at `42f0c552bb93c523b6c5eb8b4681719cf0bc3197`declarations.json.
- Ordinary and counting whole lanes: all runners, eight workers, no test filter, unchanged 40,000 ms hook timeout; 106,366 passing, one sanctioned public API acknowledgement row, zero pending. Sole baseline diff: api/typescript.d.ts. Locked install/build both exit 0.
- Counting copies preserve every tests/cases and tests/baselines/reference input byte-for-byte.

No Go source is added. No Loom submission.

Re-cut base: `7bc402ece2dda307e451112c8a735d516f9b6720`. Runtime files and pins are byte-identical to the accepted first head. Evidence lives at `stage3/census/receiver-split-evidence/transformers` on `typescript/receiver-split-evidence` at `42f0c552bb93c523b6c5eb8b4681719cf0bc3197`. This directory contains only the five house files; moved JSON is not a runtime dependency. New-stack verification receipts are published forward on the same evidence branch.
