# Verdict fixture counts

| Fixture inventory | Count | Observed A passes | B mutants caught |
|---|---:|---:|---:|
| Remaining-exclusion reason groups | 6 | 240/240 configurations | 6/6 |
| Pinned diagnostic fixtures, one per diagnostic-bearing group | 5 | 5/5 | 5/5 |
| Authored case-host diagnostic fixture (`fixtures/case-host/input.a`) | 1 | 1/1 | 1/1 |
| Focused harness tests (`test_*.py`) | 23 | 23/23 | See REMAINING.md |

The pinned fixtures are references, not duplicated TypeScript sources:
`evidence/remaining/fixtures.json` records their inputs, baselines and hashes.
The authored fixture supplies a diagnostic for the case-host group beyond its option-deprecation baseline. Group totals include all newly admitted configurations.
No fixture in Adamic's compiler/oracle corpus was added by this unit.
