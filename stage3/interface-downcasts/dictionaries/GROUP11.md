Built checked Object.values and Object.entries snapshots over existing fixed-object and dictionary record storage.
Commits: follows 7a3ea2a8; all minimal shared hooks are named in checked-views-plan.md.
Checks: 48 runtime controls, six demanded tuple refusals, two prior blocker rechecks, focused package gate and vet; logs accompany this report.
Mutants: skipping selection, accepting the wrong shape and dropping nested checks compile and are caught in C/JS; removing derived origins removes the pinned tuple refusal.
Not covered: three candidate pairs / 31 reads, rich reference alternatives, any, reached array-entry tuple consumption, and numeric/template/symbol key contracts.

The named dictionaryValuesCall lowers actual source keys followed by checked
lookup for each selected element. It reuses dictionaryReadContract, existing
record/fixed-object storage and ordinary array builders. Entries use the existing
tuple-object representation. Each receiver is evaluated once, and nested values
remain checked at subsequent reads. Empty sources produce empty snapshots.

Forty-two runtime controls cover values of strings, numbers, objects and arrays,
and entries of strings, numbers and objects, over fixed objects and independently
produced records. They run against Node in sanitized C, release C and JavaScript.
Every wrong selected value has its exact exit-70 message pinned; nested object
and array reads stop at their own field/element. Successful controls pass leaks.
Six further runtime controls prove array-entry creation remains lazy, even with
wrong unread array leaves. Six consumers of those tuples are conservatively
refused at the reached element read, and are not certified as supported.

The private-result ArrayPush.DictionaryProduction hook avoids requiring an
alias-write source certificate while building checked snapshots. It applies only
to the generated helper's unpublished result; normal ownership is retained and
no later write certificate is granted. The initial gate caught this eager guard
on valid nested arrays. Another initial failure exposed JavaScript treating a
valid entry tuple array as an object. The final implementation preserves the
existing tuple frontier and uses DictionaryEntryOrigins in the common lazy
reachability graph to refuse reached unsupported tuples in the shared frontend.
It does not add another tuple representation or eagerly reject creation.

Three semantic runtime mutants compile and execute in both backends: skipping
the element probe (with safe native representation conversion), accepting an
array where Entry is required, and dropping the nested name read check. Each
produces Node's unchecked successful output instead of the pinned refusal, so
the corresponding control catches it. The derived-origin IR mutant removes only
the new origin metadata; the fresh unsupported tuple read is then admitted, so
its pinned frontend refusal catches that mutation. No compiler error is counted
as a killed runtime mutant. The existing eager keys mutant remains in the gate.

The two values/entries witnesses in recheck-results.json now pass; their prior
refusals have been replaced with observations held by EnumerationRecheck. Rich
compiler-options reference alternatives remain dependencies on the accepted
lane 4b handoff. Primitive/nullish selection remains lane 4's handoff; a pair is
not credited from primitive-only observations. Any still needs an explicit
sound erased-value contract. No other worker's unlanded branch was merged.

Candidate accounting is unchanged: 26 pairs / 197 reads certified, three pairs /
31 reads remaining, out of 29 / 228. Enumeration is an additional owned semantic
obligation rather than another candidate row. Exact runtime reachability remains
unmeasured. Working date remains October 11, 2026, 23:00 UTC, dependent on these
contracts. Full repository gate was not run.

Commands:
ADAMIC_GATE_UNCACHED=1 go test ./internal/ir ./internal/lower ./internal/native ./internal/javascript ./internal/oracle -run '^TestCheckedView.*Dictionary|^TestDictionary|^TestViewDictionary|^TestLazyView|^TestCheckedViewArrayReferenceWrites$|^TestRecord|^TestPartialRecord' -count=1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewDictionaryValuesEntries$|^TestCheckedViewDictionaryArrayEntryCreationIsLazy$|^TestCheckedViewDictionaryEnumerationMutants$' -count=1 -v
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/oracle -run '^TestDictionaryEntryDerivedOriginMutant$|^TestCheckedViewDictionaryEnumerationRecheck$' -count=1 -v
go vet ./internal/ir ./internal/lower ./internal/native ./internal/javascript ./internal/oracle

The first broad filter also selected TestCountsAreRecorded, whose first failure
was internal/oracle/testdata/node_fs_file_rm.a:10:1, rmSync option literals
containing evaluated expressions. That unrelated refusal is preserved in
GROUP11-broad-gate.log. The final focused filter uses explicit test prefixes;
no unrelated compiler feature was changed to green the unit.

Final focused gate: exit 0; lower 2.275s, native 46.229s, oracle 103.034s.
IR/JavaScript had no matching tests in this final selection; IR's broader first
run passed 0.025s. Final controls 22.511s; blocker rechecks 0.933s; vet exit 0.
Formatting and git diff --check are clean.
