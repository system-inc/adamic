Built instantiated NodeArray storage, lazy object element reads, and source-certified scalar own-field mutation in native and JavaScript.
Commits: this checkpoint follows published 5ca21a57 and integrated ba59427c; the new pushed SHA is in the handoff.
Validation: ten original Node controls, sanitized native, release native, emitted JavaScript, and the integrated checked-view gate.
Mutants: native property aliasing, native and JavaScript own-field/element kind, and omitted source-slot certificate are caught during execution.
Limits: candidate counts only; production exact reachability is unmeasured; native-array family estimate remains October 10, 12:00 MDT.

## Behavior and named hooks

The next ranked array field is Block.statements, NodeArray<Statement>, 110 candidate
reads. NodeArray extends ReadonlyArray<T> and ReadonlyTextRange. It retains array
storage and its own pos/end/hasTrailingComma contracts rather than becoming an
object or dictionary. Instantiated ancestry and array/record source intersections
share viewArrayBase; viewArrayElementType supplies the same selected element
contract to indexes and supported consumers.

Object.assign of a fresh literal and one exact scalar record produces ArrayRecord.
The native emitter copies that record's slots, readiness, physical kinds and original
logical slot certificates into the existing owned adamic_array.properties. It labels
the storage array for diagnostics. JavaScript copies own fields and their physical
and logical certificates onto the actual array. Later mutations of the source record
are independent. No new native header fields, destruction paths or element metadata
were needed in this group: the existing array properties ownership and element_kind
are reused. ArrayProperties selects own-field storage without visiting elements.

viewArrayOwnWriteReceiver registers original slot certification for assignment and
compound assignment. A pos: number view may update an original number slot; it may
not update an original pos: 3 literal slot to 4. The exact pinned write refusal is:

    adamic: panic: field write failed: property 'pos' on array at node-array-write-literal.a:11 has no compatible declared slot

Other pinned exit-70 diagnostics name statements (expected NodeArray<Statement>,
found object), statements.pos (expected number, found missing or string), and
statement.text (expected string, found number). The lazy probe reads length, pos and
only the first text while a later text contains 9; it prints 2:3 and first. Scalar
own-field writes remain visible through the original array alias.

Every new shared hook, including the earlier optional array hook, is listed under
Lane 2 in docs/checked-views-plan.md. ArraySearch now participates in the named
arrayRead lazy-demand dispatch; TestLazyViewArrayDemand pins that adapter obligation.
Callables were not edited. No other lane branch was merged.

## Candidate remainder

This ledger uses lazy admission's read-demand candidate inventory. Fixtures use
small representative interfaces, not unchanged complete TypeScript interfaces.
The generic scalar own-field mechanism is exercised with NodeArray<Statement>;
its three representative obligations are associated explicitly with the ranked
NodeArray<Node> scalar fields: end 11, pos 10, hasTrailingComma 5. That association
is a candidate mechanism test, not a production allocation-flow certificate.

| Family | Candidate pairs / reads | Cumulative ranked representative obligations held | Candidate remainder |
| --- | ---: | ---: | ---: |
| Array contracts | 334 / 3,189 | 2 / 474 | 332 / 2,715 |
| Element or consumer reads | 251 / 1,602 | 0 / 0 | 251 / 1,602 |
| Intrinsic contracts | 179 / 794 | 0 / 0 | 179 / 794 |
| Own array fields | 30 / 72 | 3 / 26 | 27 / 46 |
| Tuples | 6 / 9 | 0 / 0 | 6 / 9 |

Families overlap and are not summed. Other prior array mechanisms have executable
fixtures but are not silently credited to this new ranked ledger. The old 2,936-site
transitive-family counts are not an unlocked count. Exact runtime reachability and
production compilation remain unmeasured/blocked by checker diagnostics.

## Evidence and limits

run-node-array-mutants.py restores every production file in finally, and rejects
build failures, frontend NotYet, compiler panics and JavaScript syntax failures as
mutation proof. The property-copy mutant prints 12:12 instead of 3:12. The JavaScript
kind mutant prints wrong with exit 0. Omitting source-slot certification permits the
write and instead fails at the later original literal read, with the wrong refusal
site and message. The native kind mutant likewise executes the unsafe numeric read.
Raw logs and original Node controls accompany this report.

Only a fresh array literal plus one exact scalar record is admitted as a fixed own
shape producer. Optional source properties, reference own properties, intrinsic/index
replacement, class/tuple ancestry and overridden intrinsics remain excluded. General
array shape mutation and reference element writes remain fail-closed. This group does
not certify every mutator, narrowing source scalar element contract, nested join or
all tuple positions. The existing unsupported-consumer scan is still conservative;
this checkpoint does not claim every array consumer refusal is fully read-local.

The full repository gate was not run. The interrupted first gate was not counted as
an overall pass; the resumed uncached gate passes lower 8.323s, native 11.514s,
JavaScript 1.490s, IR 0.019s (no matching tests), oracle 63.363s:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/native ./internal/javascript ./internal/ir ./internal/oracle -run 'Test.*View|TestLazyView|TestSharedArrayContractAdapter' -count=1 -timeout 15m
python3 stage3/interface-downcasts/lane2/run-node-array-mutants.py
```

The managed workspace restarted during the first gate; tracked and untracked edits
survived. The cloud-environment runtime skill was read on resumption; current network
policy was enforced and no credential values were inspected. Setup timing and nproc
are recorded in OPTIONAL_ARRAYS_REPORT.md. Final restored fixture, vet and additional
regression results are appended below.

Final restored observations: the focused NodeArray, ordinary array family mutants,
and bounds mutant oracle passes in 5.857s; touched-package vet exits 0. The ten
Node controls now pin original stdout even for unsafe casts. Native wrong-kind
mutation prints numeric pointer bits instead of refusing (exit 0); both backends'
wrong-value checks are therefore execution-held.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestCheckedViewNodeArrayRecords|TestArrayFamilyMutants|TestArrayWithBoundsMutant' -count=1 -timeout 15m
go vet ./internal/lower ./internal/native ./internal/javascript ./internal/ir
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/oracle -run TestPhantom -count=1 -timeout 10m
```

The additional phantom command has lower-package failures in
TestPhantomArrayCastsAreErased, TestPhantomArrayRequiredCastsAreErased and
TestPhantomArrayProofs/cycle. These exact three failures are already reproduced on
c01ae313 in stage3/interface-downcasts/integration/REPORT.md, lines 32-34. They
were not suppressed; the phantom source oracle passes in 4.108s. The additional
RegExp census assertion fails and is retained in the logs; its baseline comparison
is recorded separately below. No counts.md change was made.

The RegExp count failure is independently reproduced unchanged on detached
5ca21a57 in /workspace/scratch/lane2-nodearray-baseline. That checkout references
the original cohere submodule through a symlink, rather than copying its code.
Both checkouts observe 631/631/621/628/17/0/0/0; the stored row remains stale.

The tenth probe, node-array-wrong-element, separately pins:

    adamic: panic: element read failed: statements[0] expected Statement, found number

Node prints undefined. Native and JavaScript element-kind omission mutants are
caught during execution (native ASan refusal mismatch; JavaScript reaches the
wrong descendant-field refusal). The first run used an incorrect field-read
prefix in the expected diagnostic; it is retained as a failure and corrected to
the observed element-read prefix. Candidate credits are unchanged by this probe.

Final restored ten-probe oracle passes in 2.320s:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewNodeArrayRecords$' -count=1 -timeout 10m
```

Six execution mutants were run in two batches: the four own-storage guards and
the two selected-element guards. All caught results and restored gate logs are
under node-array-logs/. The script defaults to all six and accepts names to rerun
an individual guard. No production runtime guard remains modified.
