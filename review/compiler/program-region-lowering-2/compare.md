Both implementations start at runtime core a3f28d97 and implement all three requested landing changes. This comparison reads the source and committed test logs; it does not rerun the other session's tests.

| Item | compiler/program-region-lowering, 325aded7 | compiler/program-region-lowering-2, implementation efb0be51 |
| --- | --- | --- |
| Allocation-time membership | Two checker-sharing lowering passes; flags set during allocation construction; descriptor replay consistency check | Two checker-sharing lowering passes; flags set during allocation construction; no reflective production IR rewrite |
| Structural shape index | Required/offered property filtering; quoted shape keys; candidate cache shared by shape | Required/offered property filtering; property postings; candidate cache by checker type |
| ASan member storage test | Actual pointer assertion; counted mapper reuses, member mapper allocates fresh | Actual pointer assertion; counted mapper reuses, member mapper allocates fresh |
| Ruling mutants | Extra leaf and dropped cyclic root caught by independent region-count expectations; Node, sanitizers and leaks remain clean | Selection mutations caught by inference assertions; separate actual membership mutations verify Node, sanitizer and leak safety with region deltas |
| Reuse mutant | Caught by storage assertion under ASan/UBSan | Caught by storage assertion under ASan/UBSan |
| Additional mutants | Field-only adoption size, shape-key collision, entry-pair publication | Field-only adoption size, optional-property shape filtering |
| Recorded selection before / after | 7.415796s / 1.785695s; leaves 8.99s / 3.42s | Final independent checkers: 23.147740696s / 11.766952292s; leaves 30.95s / 19.23s. Earlier same-checker warm run: 8.452791265s / 1.348071880s |
| Census | 1,685 records; 1,599 members, 86 counted, zero unresolved | Same; additional logged checker flags and element traversal for K60/K144 |
| Node, sanitizer, leak and counts | Six mirrors and million pass; existing counts unchanged; eleven new rows | Six mirrors and million pass; all 1,045 existing recorded entries unchanged; fifteen new rows |
| Runtime sources | Unchanged | Unchanged |
| Boundaries | Refuses selected canonical closures and member programs materializing Map/Set entry pairs; async refused | Emits canonical closure helper that adopts before cache publication; async refused; no equivalent entry-pair boundary test |
| Lane / vet evidence | Lane passes, including vet | Lane passes with its vet timeout skip; separate targeted vet passes |

The timing logs were collected in separate sessions; the second branch's final census ran with competing validation work. These numbers do not establish which index is faster under equal conditions. Both retain K60 and K144 as selector members while preserving the recorded counted expectations.

I would keep 325aded7 as the conservative landing starting point. Its quoted shape keys avoid delimiter collisions from unusual property names, and its entry-pair publication refusal has a failing mutant. The second branch uses raw NUL/separator joins for shape keys and lacks that boundary test, so its extra canonical closure support is not sufficient reason to prefer it without review. The second branch's optional-property index test, explicit ruling safety variants and detailed K60/K144 diagnostics are useful follow-up evidence.

Evidence: original branch review/compiler-program-region-lowering/{report.md,mutants.log,census-pairwise.log,census-indexed.log}; second branch review/compiler/program-region-lowering/{REPORT.md,program-complete.log,census-split.log,extra-leaf-mutant.log,missing-member-mutant.log,optional-shape-portable.log}. Neither full gate was run here. No implementation files were changed for this comparison.
