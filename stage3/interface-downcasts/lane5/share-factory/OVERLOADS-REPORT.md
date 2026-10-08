Built: original-declaration overload contracts for createImportClause and createYieldExpression; five pairs now certify 386 reads toward step 09.
Commits: first three pairs 1da34a72; this commit adds the two remaining overload pairs without compiler changes.
Commands and outputs: focused uncached factory/frontier tests passed; owned counts update passed; required global counts update failed outside these fixtures.
Mutants: selecting only the compatible second overload makes each pinned negative execute with stdout 7 in release, sanitized native, and JavaScript; finishing mutants pass leak checks.
Uncovered: cloneNode 91, createModifier 10, onEmitNode 11 remain stopping frontiers; no whole-tsc run or enum-member certificate is claimed.

The original declarations are unchanged. Each original share-a control is byte identical to the new wrong-overload fixture: its producer serves the second signature but fails the first. Positive producers accept every declared parameter domain. Conversion fixtures exercise numeric, boolean, undefined, present object, and absent object arguments. The yield fixture also reaches its third signature. Source Node determines the output for every fixture. Release native, ASan/UBSan native, JavaScript, and independent native leak checks agree on finishing executions. Negatives stop before invocation with the full declared overload-set message and ordinal 1.

The yield third signature is included in the contract. Its producer parameter domains overlap the first two; the mutant selects only signature 2, rather than claiming an independent producer that passes signatures 1 and 2 but fails signature 3.

Validation output is in logs/second-final.log, logs/counts-second.log, and logs/counts-global-second.log. Commands:

```
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableFactory$|^TestCheckedViewCallableFactoryCounts$|^TestCheckedViewCallableFactoryDeclarations$|^TestCheckedViewCallableShareAFrontiers$' -count=1 -v -timeout 5m
go test ./internal/oracle -run '^TestCheckedViewCallableFactoryCounts$' -count=1 -args -update-counts
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 5m -args -update-counts
```

The uncached focused run passed in 31.327s. The owned updater passed in 3.752s and appended six rows without changing existing rows. The global updater failed in 78.863s with existing unrelated fixture failures, including graph_regions_regression_07 free(): invalid pointer, already reproduced at the unchanged base in the first delivery. Its early fixture failure prevents the global interface-downcast update phase. An initial focused pattern also selected the counts test before updating it; only the six unrecorded rows failed, while both new pair tests passed. The final run includes recorded counts and passes.
