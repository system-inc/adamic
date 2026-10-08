Built three remaining lookup pairs with 20 source controls over fixed objects and genuine shared record producers.
Commits: follows 8a660bcd; named erased-source admission and descriptor readiness hooks are documented in the lane plan.
Checks: lower/native/JS/oracle targeted gate passed (1.710s / 42.465s / built / 57.805s); vet and candidate reproduction passed.
Mutants: removed view origin loses the named refusal in compiling C/JS; optional-absence, container, skip, wrong-shape, transitive and producer-certificate mutants pass.
Not covered: optional dictionary receiver, finite named-key caches, rich element unions and enumeration; 11 candidate pairs / 58 reads remain, exact reachability unmeasured.

The working date remains October 11, 2026, 23:00 UTC. This group certifies the
MapLike<string>, MapLike<T> and PackageJsonPathFields dynamic-key rows, one
candidate read each. Fixtures are original representative read contracts.
MapLike<T> instantiates T to string in these witnesses; earlier generic
object/array propagation controls supplement them. Unsupported unresolved and
rich union element selections are not claimed implemented.

Scalar/generic lookups have valid, wrong-number and absent-key controls. Nested
package maps add a wrong-container and wrong nested value. Every control runs
on a fixed object and on a genuine Record producer held in a checked holder
field. Valid and absent cases agree with Node in sanitized C, release C and JS,
with the leak check; wrong values pin the entire named exit-70 message.

A structural {} source appears assignable to an index signature because it
declares no keys. That does not certify the runtime storage or its hidden values.
`dictionaryCastNeedsView`, called before apparent upcast admission, installs the
ordinary lazy view for these erased dictionary casts. It preserves reverse
assignability, nominal exclusions and the writable-slot check.

The generic source first exposed invalid C: its checked record read descriptor
contained a Readiness expression for a parameter even after the enclosing
operation's argument had been cleared. The readiness transform now visits only
the dictionary property pointer and applies the same existing facts. It does
not suppress readiness checks or add a new runtime representation.

Ordinary Record-to-{} erasure still refuses because plain fixed-object
operations would see the wrapper's storage fields. Initial producer probes hit
that refusal. Final producer fixtures use checked holder fields, exercising the
actual shared tables without weakening the plain erasure rule.

Removing the new view origin compiles and runs in both backends. Native loses
the named field contract and reaches the existing ordinary storage refusal;
JavaScript exits 0 and prints 42. Both violate the normal control's exact named
exit-70 message. This is a semantic mutation, not a compiler rejection.

Final commands:

    ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/native ./internal/javascript ./internal/oracle -run '^TestCheckedViewDictionary|^TestCheckedViewLazy|^TestOptionalCheckedRead|^TestReadinessMutants$|^TestUninitializedIsNotNullishMutant$|^TestCast|^TestOptionalCastKeepsHiddenFieldCheck$|^TestViewDictionaryDescriptors$|^TestRecordForms$|^TestRecordRefusals$|^TestRecordPrototypeLiteralNames$|^TestRecordsAgainstNode$|^TestRecordMutants$|^TestRecordReadMutants$' -count=1 -v
    go vet ./internal/lower ./internal/native ./internal/javascript ./internal/oracle
    python3 stage3/interface-downcasts/dictionaries/rank-lazy-demand.py --check

Final evidence: logs/group7-gate.log and group7-vet.log. Initial admission,
readiness and producer probes are retained. No full repository gate is claimed;
baseline failures from GROUP3.md remain outside this targeted gate.
