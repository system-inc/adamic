Merged ca082675 locally in b0ef560b; withheld push because requested reader guard and conflict validation fail.
Search-shrink merge and counts audit were pushed successfully at f130c846.

Commands: ADAMIC_GATE_UNCACHED=1 timeout 120 go test ./internal/native ./internal/oracle -run '(?i)split|^TestSelfCompare|^TestRetainedTopLevelCoverage$' -count=1 -v -timeout 90s passed: native 0.458s, oracle 0.781s. Split regression 0.44s. Three checker-archive tests skipped. Constant self-equality mutant caught by NaN output.

The existing WASI fixture list gained arguments_length_extended.a, closure_convention_regexp_count.a and closure_convention_nested.a in 4f23173f. Updated retained coverage pins from 35 fixtures and 36 units to 38 fixtures and 39 units, including the exact membership hash. No fixture was removed.

The requested timeout 120 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -v -timeout 90s failed after 20.266s:
- Unapproved internal/lower/exceptions.go:throwsOutReadiness:ArraySort.Comparator.
- Unapproved internal/lower/class_static_guard_test.go:TestClassStaticInitializerCallIsEmitted:Call.Function.
- Stale internal/lower/exceptions.go:throwsOut:ArraySort.Comparator allowlist entry.

Additional timeout 120 go test ./internal/lower -run '^TestUndecidedCycleReadsUseReadyChecks$' -count=1 -v -timeout 90s failed after 1.526s. Incoming agreeEntry expects source Node exit 70, but early value, indirect value, early plain class and early extends return exit 1 with ReferenceError. The merge conflict resolution removed the previous terminal-check runner block in favor of agreeEntry; this needs correction before publishing the merge.

No full gate, WASI execution or new split inline mutant run. The split-header upstream mutant evidence remains on the merged branch. No push of this local merge.
