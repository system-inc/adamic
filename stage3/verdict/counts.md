# Verdict fixture counts

| Fixture inventory | Count | Observed A passes | B mutants caught |
|---|---:|---:|---:|
| Remaining-exclusion reason groups | 6 | 240/240 configurations | 6/6 |
| Pinned diagnostic fixtures, one per diagnostic-bearing group | 5 | 5/5 | 5/5 |
| Authored case-host diagnostic fixture (`fixtures/case-host/input.a`) | 1 | 1/1 | 1/1 |
| Focused harness tests (`test_*.py`) | 45 | 45/45 | See REMAINING.md, FINAL.md and DIFFERENTIAL.md |
| Final-command orchestration fixture outcomes (A, B, C) | 3 | A reaches timing | B and C refuse performance |
| Differential baseline coverage campaign | 13,693 configurations | 13,693/13,693 | Source and map mutants below |
| Differential one-function witness | 9 configurations | 9/9 | stdout-only function change caught |
| Differential missing-function and stale-source maps | 2 mutants | Not oracle passes | Both reject with exit 2 |
| Fresh compiler-module fixture | 1 | Two identical UTF-8 outputs and hit sets | Reused-module mutant rejected |

The pinned fixtures are references, not duplicated TypeScript sources:
`evidence/remaining/fixtures.json` records their inputs, baselines and hashes.
The authored fixture supplies a diagnostic for the case-host group beyond its option-deprecation baseline. Group totals include all newly admitted configurations.

No fixture in Adamic's compiler/oracle corpus was added by this unit.
