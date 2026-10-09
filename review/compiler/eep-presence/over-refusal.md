# Over-refusal comparison

Every `.a` fixture beneath `internal/oracle/testdata` and `stage3` was checked on the current fixture corpus with the compiler from 897d0e7989ecfc2175f25966b8def3efe51b1ec8 and the unchanged eep-presence compiler from 5f08a92cc6363645f7aa0c6cbc4898eaf2c04118. The probe performs `load.Load`, then `lower.Lower`, exactly as `review/compiler/views-v4/sweep/check.go.txt` on compiler/views-v4. Its source is preserved in `followup/check.go.txt`; all 1,916 paths are in `followup/fixtures.json`.

| Comparison | Fixture count |
| --- | ---: |
| Admitted on 897d0e79 and this branch | 1442 |
| Refused on 897d0e79 and this branch | 474 |
| Admitted on 897d0e79, refused on this branch | 0 |
| Refused on 897d0e79, admitted on this branch | 0 |
| Probe errors or timeouts | 0 |

All individual checker/lower verdicts, diagnostics, and timings are recorded in `followup/admission-initial.jsonl`. The binaries read the same current fixture paths and pinned checker, so the comparison isolates the lower guard rather than changes in historical fixture corpora. Each probe had a 60-second process-group limit; two workers ran in parallel. There are no incomplete or infrastructure-invalid rows in the final sweep.

| Fixture admitted before and refused now | Source location | Reason / disposition |
| --- | --- | --- |
| None | Not applicable | No measured over-refusal regression; no guard change required. |

V4's script enumerated all of `internal/oracle`, rather than just its `testdata` directory. The only two additional `.a` files are `internal/oracle/refusals/nested_cycle.a` and `internal/oracle/refusals/nested_rebinding.a`; both are refused by both compilers. Their complete verdicts are in `followup/admission-v4-extra.json`. Thus the requested 1,916-fixture set and V4's 1,918-fixture directory set both have zero newly refused fixtures.

This is a corpus admission result. It does not prove that the static guard precisely tracks reachability for every possible valid source, and it does not replace Node runtime comparisons. The original seven review witnesses remain separate located-refusal fixtures with a caught real revert mutant; their containment is unchanged.
