# Lane diagnostic fixture counts

| Inventory | Count |
|---|---:|
| Existing lane tests | 32 |
| Additional diagnostic tests in test_check.py | 8 |
| Additional artifact tests in test_artifact.py | 7 |
| Additional file-plan tests in test_measure_plan.py | 5 |
| Additional observer tests in test_profile.py | 2 |
| Additional case-shard and shared-artifact tests in test_shards.py | 14 |
| Additional case observer tests | 1 |
| Additional upstream worker environment tests | 1 |
| Total lane tests | 70 |
| Real targeted failure fixtures (timeout, baseline difference) | 2 |
| Dropped-error-text mutant | 1 |
| Additional artifact guard mutants | 5 |
| Additional file-plan guard mutants | 3 |
| Additional dropped-IPC-observation mutant | 1 |
| Additional case-shard guard mutants | 10 |
| Additional dropped-case-observation mutant | 1 |
| Additional dropped-worker-environment mutant | 1 |
| Additional real upstream input mutants | 3 |
| Cold oversized-file measurements | 1 |
| Additional 10-second-plan metadata/declaration guard mutants | 5 |
| Additional real single-execution-deletion union mutant | 1 |
| Additional real-work 25-second budget mutant | 1 |
| Rejected noEmit state-pair split witness | 1 |

The fixtures run upstream TypeScript on scratch inputs. No Adamic compiler
fixture, internal/oracle/counts.md row, or expected lane count was changed.

The manifest has 377 units: 348 file shards, 12 exact groups from four expensive
suites, 16 watcher cases and one Public APIs byte-comparison unit. Historical
151-shard evidence is in CASE_SHARDS_REPORT.md. Current budget, union and mutant
evidence is in SHARDS_10S_REPORT.md.
