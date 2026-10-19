# Entries fixture counts

Six baseline fixtures and six single-edit source mutants. Six additional header mutants prove the existing gate predicate. Native heap-event counts are not recorded by this bucket; successful programs are leak-checked by the shared runner.

| Fixture | Node stdout bytes | Current main | Ruled mode | Ruled runtime exit |
| --- | ---: | --- | --- | --- |
| 01_scanner_keywords.a | 884 | Compiles | proven | 0 |
| 02_hidden_same_type.a | 56 | Compiles | proven | 0 |
| 03_hidden_other_type.a | 60 | CheckedStop | checked | 70 |
| 04_alias_add_key.a | 54 | Compiles | checked | 0 |
| 05_getter.a | 71 | Refused | refused | compile refusal |
| 06_symbol_key.a | 47 | Refused | refused | compile refusal |

Scanner coverage: all 84 keyword entries in their upstream insertion order, including the computed constructor string key.

Current outcomes on `e61c738e` over main `80640d7a`: 3 Compiles, 1 CheckedStop,
2 Refused. Four baselines produce native executables. All 6 raw Node runs still
exit 0 with empty stderr and their original stdout. The three successful native
programs match that raw Node behavior. The checked stop prints exactly
`before entries\n`, reports the incompatible hidden string, and exits 70,
matching generated JavaScript on Node. The six source mutants all compile;
two exercise exit-70 checks and four finish successfully. All twelve backend
acceptance contracts are now met. See [LANDING.md](LANDING.md) for the measured
shared-runner gates, individual walls, and independent runner mutants.
