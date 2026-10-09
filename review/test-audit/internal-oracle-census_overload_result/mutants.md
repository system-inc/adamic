| ID | Origin file:line | Change | Failed rows |
|---|---|---|---|
| M01 | internal/lower/census_overload_proof.go:27 | !l.censusHasUndefined(produced) -> l.censusHasUndefined(produced) | TestCensusAppendResultProof, TestCensusOverloadResultStop, TestNativeAgreesWithNode family (append and lie inputs only) |
| M02 | internal/lower/census_overload_proof.go:119 | !reachable(flow) -> reachable(flow) | TestCensusOverloadResultStop, TestNativeAgreesWithNode family (append and lie inputs only) |
| M03 | internal/lower/census_overload_proof.go:144 | len(flows) > 256 -> len(flows) > 0 | TestCensusAppendResultProof |
| M04 | internal/lower/census_overload_proof.go:170 | !trusted -> trusted | TestCensusAppendResultProof |
| M05 | internal/lower/census_overload_proof.go:58 | l.typeMapper = newTypeMapper(sources, targets)  -> | TestCensusAppendResultProof |
| M06 | internal/lower/census_overload_proof.go:102 | !unchanged -> unchanged | TestCensusAppendResultProof |
| M07 | internal/lower/census_small.go:295 | ordinal := 0 -> 			ordinal := 1 | TestCensusOverloadResultStop |
| M08 | internal/lower/census_small.go:312 | "overload %d of %s result: expected %s, got undefined" -> "overload %d of %s result: expected type %s, got undefined" | TestCensusOverloadResultStop |
| M09 | internal/lower/census_overload_proof.go:250 | admitted[index].Flags()&checker.TypeFlagsUndefined == 0 -> admitted[index].Flags()&checker.TypeFlagsUndefined != 0 |  |
