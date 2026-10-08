# Lane diagnostic fixture counts

| Inventory | Count |
|---|---:|
| Existing lane tests | 32 |
| Additional diagnostic tests in test_check.py | 8 |
| Additional artifact tests in test_artifact.py | 7 |
| Additional file-plan tests in test_measure_plan.py | 5 |
| Additional observer tests in test_profile.py | 2 |
| Total lane tests | 54 |
| Real targeted failure fixtures (timeout, baseline difference) | 2 |
| Dropped-error-text mutant | 1 |
| Additional artifact guard mutants | 5 |
| Additional file-plan guard mutants | 3 |
| Additional dropped-IPC-observation mutant | 1 |
| Additional real upstream input mutants | 2 |
| Cold oversized-file measurements | 1 |

The fixtures run upstream TypeScript on scratch inputs. No Adamic compiler
fixture, internal/oracle/counts.md row, or expected lane count was changed.
