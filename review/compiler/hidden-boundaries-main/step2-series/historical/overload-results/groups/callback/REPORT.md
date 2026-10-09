Added immediate callback arguments whose concrete input and result contracts the implementation serves directly.
Compiler commit 0c84ce9c93451b4e5aa7bbeb0d2fb2e38f8e7a9a; baseline measurement commit 8f124d1c257c05f3ae05ed7861d91d3a71927331.
Focused lower tests, six uncached Node oracle fixtures and the counts refresh pass; seven interval measurements remain unchanged.
Missing input relation, missing result relation, changed return representation and missing field-storage check each fail their refusal assertion.
Step 16 gains a conservative callback case; no assigned hidden bytes are revealed and the full structural and visitor contracts remain unfinished.

A callback consumer promising Node does not receive the generic overload's
stronger promise to return the original T. Its implementation can serve the
consumer directly if every admitted input fits, its entire result fits, and
both the call representation and record field storage agree. Only one concrete
contextual signature is allowed; both signatures have fixed required parameters.
Different record types require readonly scalar fields with matching storage.
Optional, rest, generic and unserved callbacks retain explicit stops.

The original closure is passed, so there is no extra callable value to change
identity or capture ownership. The fixture checks both module and nested
callbacks, captured text, and equality of two references to the same callback.
Source Node prints name:module, name:nested, true, true. Both emitted backends,
release native, ASan/UBSan native and the counted build agree. Existing direct
result checks and parameter checks remain green.

Focused validation commands (output files retained beside this report):

```sh
go test ./internal/lower \
  -run 'TestOverloadCallback|TestOverloadResults|TestCensusOverload|TestPredicateOverload' \
  -count=1 > docs/overload-results/groups/callback/focused.log.txt 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle \
  -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/(overload_callback_served|overload_results_(binding|evaluate|parameter|scalar|transform)).a$' \
  -count=1 -v > docs/overload-results/groups/callback/oracle.log.txt 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts \
  > /tmp/overload-callback-counts.log 2>&1
python3 docs/overload-results/groups/callback/run-mutants.py \
  > docs/overload-results/groups/callback/mutants.log.txt 2>&1
```

The four compiler mutants each cause the focused lower test to accept an
unserved callback, making its refusal assertion fail with got <nil>. None fails
through compilation or clang warnings. The field-storage witness specifically
uses number in the input record and number | undefined in the implementation
record; equal outer object pointers cannot establish that layout conversion.
The first input mutant was initially blocked by a redundant semantic test in
the storage helper. That helper now checks every branch's field storage without
using the semantic relation, and the final mutant isolates and kills the input
relation check. An initial identity fixture passed booleans to string-only
console.log and was rejected by the checker; formatting the booleans fixed it.

The census uses the exact seven-interval scope patch from the baseline group,
the pinned subtraction algorithm, all 82 verified source hashes, and the
no-output guards. The prior c68 census supplies the before records. regions.json
identifies the measured compiler commit. Every interval remains fully hidden:
13,625; 6,899; 6,078; 5,748; 11,417; 7,289; 7,102 bytes, totaling 58,158.
The comparison measures a checker-rejected program, not a runnable compiler.

The full visitor callbacks still need a representation change for results that
may be a node or an array, and proof of their input domains. Block's full record
cannot be established by checking only kind, since Expression admits that tag
without proving Block's fields. EvaluatorResult can arrive through a shared
resolver result; checking its value once does not prevent a wider alias from
changing that field later. Those cases remain refused or NotYet. No worker's
unlanded implementation was merged, including hidden-06's separate TNode storage
change. The structural result group has not been delivered.

No protected compiler files were edited. No package-wide tests or full gate ran.
