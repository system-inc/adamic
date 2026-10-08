Built actual-source key enumeration for dictionary views with six source controls over fixed objects and genuine shared records.
Commits: follows 04de3f66; named dictionaryKeysCall and source_keys hooks are listed in the plan.
Checks: focused lower/native/oracle gate passed (2.129s / 50.667s / 82.424s); vet passed; Node, sanitized C, release C and JavaScript controls/leaks passed.
Mutants: eager element check compiles and exits 70 in C/JS where Node keys succeeds; existing skip, wrong-shape and transitive mutants passed the gate.
Not covered: three candidate pairs / 31 reads remain; rich reference alternatives, any, values/entries, finite literals and numeric/symbol key contracts remain uncertified.

Object.keys now observes actual source presence and key ordering without demanding
unread element values. Numeric keys precede string keys in Node order. Empty
sources return empty arrays. Both fixed-shape source objects and independently
certified Record<string,number> producers pass key-only reads through a
MapLike<string> view. An explicit subsequent wrong string lookup still stops
with its exact named message, after all key output. No table copy, second record
representation or alternate initialization bitmap is introduced.

The native adapter validates object heap kind, dispatches actual records to
record_keys, and fixed/class objects to object_keys. Existing public/private and
class-presence semantics remain with that adapter. The frontend recognizes
string dictionaries; ordinary JS Object.keys already has the required behavior.
The previous native path checked an element storage certificate before keys,
eagerly refusing these legal key-only observations. The eager semantic mutant
checks each value as string before returning keys. Both generated backends build
and exit 70; the good key control catches that change against Node exit zero.

recheck-results.json preserves current observations for every remaining candidate
kind and explicitly recorded key-domain/finite-element/enumeration gaps. All
source programs run under Node; malformed dynamic key witnesses are conservative
frontier probes, not successful source contracts. Rich primitive/nullish selected
reads execute correctly; array, map and source-file alternatives still stop at
the read with exit 70 and remain pending under the accepted lane 4/4b handoff.
Any, finite optional literal, numeric and symbol index signatures still refuse.
Values/entries remain demanded enumeration refusals and are this lane's next work.
The first scratch JS execution lacked the Adamic package loader; rerunning through
oracle/node.mjs corrected that harness error. Only final JS observations are in
the results JSON. No rich pair is credited from primitive-only controls.

26 candidate pairs / 197 reads are certified; exact runtime reachability is
unmeasured. This enumeration batch removes no candidate row, since its coverage
is an owned semantic obligation beyond that candidate table. Working date stays
October 11, 2026, 23:00 UTC. Full repository gate was not run.

Command: ADAMIC_GATE_UNCACHED=1 go test ./internal/ir ./internal/flow ./internal/fresh ./internal/lower ./internal/native ./internal/javascript ./internal/oracle -run '^TestCheckedViewDictionary|^TestCheckedViewLazy|^TestViewDictionary|^TestRecordForms$|^TestRecordRefusals$|^TestRecordPrototypeLiteralNames$|^TestRecordsAgainstNode$|^TestRecordMutants$|^TestRecordReadMutants$' -count=1 -v
Then go vet ./internal/lower ./internal/native ./internal/javascript ./internal/oracle.
