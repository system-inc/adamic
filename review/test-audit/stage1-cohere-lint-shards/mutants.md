| ID | Origin shards.go line | Change | Failing rows |
|---|---:|---|---|
| C1 | 78 | `change prefix constant` | TestMergePutsCasesBackInOrder, TestMergeRefusesAMissingOrRepeatedCase, TestMergeKeepsLinesThatOnlyLookLikeCases |
| C2 | 119 | `off-by-one block key` | TestMergeRefusesAMissingOrRepeatedCase, TestMergePutsCasesBackInOrder, TestMergeKeepsLinesThatOnlyLookLikeCases |
| C3 | 116 | `flip duplicate condition` | TestMergeRefusesAMissingOrRepeatedCase |
| C4 | 107 | `drop whole-line condition` | TestMergeKeepsLinesThatOnlyLookLikeCases |
| C5 | 149 | `off-by-one digit bound` |  |
| C6 | 146 | `change digit upper bound` |  |
| C7 | 119 | `off-by-one block slice` | TestMergePutsCasesBackInOrder, TestMergeKeepsLinesThatOnlyLookLikeCases |
| C8 | 123 | `flip case-count condition` | TestMergeRefusesAMissingOrRepeatedCase |
| C9 | 149 | `flip line terminator condition` | TestMergePutsCasesBackInOrder, TestMergeKeepsLinesThatOnlyLookLikeCases, TestMergeRefusesAMissingOrRepeatedCase |
| P1 | 85 | `empty Merge body` | TestMergeRefusesAMissingOrRepeatedCase, TestMergePutsCasesBackInOrder, TestMergeKeepsLinesThatOnlyLookLikeCases |
