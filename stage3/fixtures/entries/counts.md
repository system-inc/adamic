# Entries fixture counts

Six baseline fixtures and six single-edit source mutants. Six additional header mutants prove the existing gate predicate. No native heap counters are available on the measured base.

| Fixture | Node stdout bytes | Current main | Ruled mode | Ruled runtime exit |
| --- | ---: | --- | --- | --- |
| 01_scanner_keywords.a | 884 | Refused | proven | 0 |
| 02_hidden_same_type.a | 56 | NotYet | proven | 0 |
| 03_hidden_other_type.a | 60 | NotYet | checked | 70 |
| 04_alias_add_key.a | 54 | Refused | checked | 0 |
| 05_getter.a | 71 | NotYet | refused | compile refusal |
| 06_symbol_key.a | 47 | NotYet | refused | compile refusal |

Scanner coverage: all 84 keyword entries in their upstream insertion order, including the computed constructor string key.

Current outcomes: 2 Refused, 4 NotYet, 0 native executables. All 6 raw Node runs exit 0 with empty stderr. All 6 source mutants exit 0 with empty stderr and differ from their goldens. A-check accepts 6/6 and rejects 6/6 wrong-header inputs. The filtered shared fixture test passes all 6 Node and stage0 subtests.

The separate ruled acceptance run is deliberately red on current main: 0/12 baseline-and-mutant backend contracts match. Expected native exit 70 is an acceptance requirement, not a measured current result.
