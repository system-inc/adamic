# Mutants for the ruled backend divergences the delta reads (#tqm95v6)

The delta reads internal/oracle/testdata/backend_divergences.json, the rows the oracle's backend tests pin, never a second list. Each mutant is one edit, run on a Linux box, then put back.

| Mutant | Edit | Red |
|---|---|---|
| A row removed | internal/oracle/testdata/backend_divergences.json: node_commonjs_require_missing.a's row deleted | Codex cloud task_e_6aeff607759083219325c94ca8aaee84 (Linux x86_64, at ae7f8a35): TestBackendDivergenceCommandAccepted, "ruled backend divergence not counted". The program reads agree false, ruled_backend_divergence false and verdict fail, and the summary reads "ruled backend divergences: 0" |

The refusals that keep the rule narrow are TestBackendDivergenceCommandRefusals' four subtests: a path on no row, the row's Node outcome differing, its native outcome differing, and native and JavaScript disagreeing. Each fails the verdict.
