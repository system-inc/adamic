Replayed nine qualifying mutants against the full package with all four candidates skipped. Baseline passed in 36.8 seconds; replay wall time totaled 391.5 seconds. Nothing was stale, broken or panicking. Removing both refusal tests loses three diagnostic catchers; either can preserve that coverage. Tree restored; evidence pushed to `test-defend/deletion-set/internal-lower` under `review/test-defend/deletion-set/internal-lower/`.

```json
{
  "package": "internal/lower",
  "main": "d29d80ceb5d5d42d9b0ffb7b528a57a2272c76f4",
  "skipped": [
    "TestArgumentsLengthRefusalFixtures",
    "TestArgumentsLengthRefusals",
    "TestArrayPredicateCannotInventAnElementContract",
    "TestDefaultTaggedInterfaceNeedsNoFlag"
  ],
  "mutants": [
    {
      "mutant": "M03",
      "file_line": "internal/lower/arguments_length.go:23",
      "branch": "test-audit/internal-lower-arguments_length",
      "candidates_failed": ["TestArgumentsLengthRefusalFixtures", "TestArgumentsLengthRefusals"],
      "still_caught_by": ["TestArgumentsLengthReadKeepsReaderFact", "TestArgumentsLengthReadNeighbors", "TestMixedTupleSpreadIsNotYet"],
      "stale": false
    },
    {
      "mutant": "M01",
      "file_line": "internal/lower/arguments_length.go:44",
      "branch": "test-audit/internal-lower-arguments_length",
      "candidates_failed": ["TestArgumentsLengthRefusalFixtures", "TestArgumentsLengthRefusals"],
      "still_caught_by": ["TestArgumentsLengthReadKeepsReaderFact", "TestArgumentsLengthReadNeighbors", "TestMixedTupleSpreadIsNotYet"],
      "stale": false
    },
    {
      "mutant": "M02",
      "file_line": "internal/lower/arguments_length.go:29",
      "branch": "test-audit/internal-lower-arguments_length",
      "candidates_failed": ["TestArgumentsLengthRefusalFixtures", "TestArgumentsLengthRefusals"],
      "still_caught_by": ["TestArgumentsLengthReadKeepsReaderFact", "TestArgumentsLengthReadNeighbors", "TestMixedTupleSpreadIsNotYet"],
      "stale": false
    },
    {
      "mutant": "M07",
      "file_line": "internal/lower/library_array_predicate.go:51",
      "branch": "test-audit/internal-lower-arguments_length",
      "candidates_failed": ["TestArrayPredicateCannotInventAnElementContract"],
      "still_caught_by": ["TestArrayPredicateCoexistsWithUnknownReflection", "TestArrayPredicatePreservesDeclaredElementContract"],
      "stale": false
    },
    {
      "mutant": "M01",
      "file_line": "internal/lower/interface_cast.go:69",
      "branch": "test-audit/internal-lower-interface_cast",
      "candidates_failed": ["TestDefaultTaggedInterfaceNeedsNoFlag"],
      "still_caught_by": ["TestDefaultTaggedInterfaceAdmission", "TestPredicateBodyProof", "TestPredicateOverloadRuntime", "TestViewObjectContractsAreAvailableToEraser"],
      "stale": false
    },
    {
      "mutant": "D02",
      "file_line": "internal/lower/arguments_length.go:47",
      "branch": "test-defend/internal-lower-arguments_length",
      "candidates_failed": ["TestArgumentsLengthRefusalFixtures", "TestArgumentsLengthRefusals"],
      "still_caught_by": [],
      "stale": false
    },
    {
      "mutant": "D07",
      "file_line": "internal/lower/library_array_predicate.go:51",
      "branch": "test-defend/internal-lower-arguments_length",
      "candidates_failed": ["TestArrayPredicateCannotInventAnElementContract"],
      "still_caught_by": ["TestArrayPredicateCoexistsWithUnknownReflection", "TestArrayPredicatePreservesDeclaredElementContract"],
      "stale": false
    },
    {
      "mutant": "D04",
      "file_line": "internal/lower/arguments_length.go:49",
      "branch": "test-defend/internal-lower-arguments_length",
      "candidates_failed": ["TestArgumentsLengthRefusalFixtures", "TestArgumentsLengthRefusals"],
      "still_caught_by": [],
      "stale": false
    },
    {
      "mutant": "D03",
      "file_line": "internal/lower/arguments_length.go:30",
      "branch": "test-defend/internal-lower-arguments_length",
      "candidates_failed": ["TestArgumentsLengthRefusalFixtures", "TestArgumentsLengthRefusals"],
      "still_caught_by": [],
      "stale": false
    }
  ],
  "keep": [
    {
      "test": "TestArgumentsLengthRefusalFixtures",
      "because": "D02, D03 and D04 lose their last catcher when both refusal candidates are removed. Keep at least one of the pair."
    },
    {
      "test": "TestArgumentsLengthRefusals",
      "because": "D02, D03 and D04 lose their last catcher when both refusal candidates are removed. Keep at least one of the pair."
    }
  ],
  "deletable": [
    "TestArrayPredicateCannotInventAnElementContract",
    "TestDefaultTaggedInterfaceNeedsNoFlag"
  ],
  "notes": [
    "Deletable means every gathered mutant remains caught, not proof against all possible breaks.",
    "Interface defender round2 uses F01-F03 filenames, excluded by the requested M*.diff and D*.diff selection.",
    "No witness-only failures or panics occurred."
  ]
}
```
