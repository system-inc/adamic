Built checked array replacement and push for flat required scalar records in native and JavaScript, preserving original element and incoming allocation contracts.
Commit: follows 9ba1f722976d6d150bb97f1f0fc3e5bc80daff03; the pushed SHA accompanies this report.
Validation: fourteen Node-pinned probes, sanitized/release native, JavaScript, and positive native leak checks; focused package and existing array oracle gate.
Mutants: eight executed witnesses for compatibility, incoming readiness, readonly/mutable field direction and reference ownership; all caught.
Limits: optional/nested/class/union records and other reference families remain refused; corpus counts remain candidate obligations, not exact unlocked cast sites.

Original logical element declarations now travel separately from physical
`element_kind`. Complete required scalar record allocations carry their original
contract ID. Element replacement and push first compare incoming and original
contracts with the shared structural rule, including both directions for mutable
fields, then check the incoming fields' presence, initialization, scalar runtime
kind and finite literals. Only after those checks does native retain the incoming
reference and change the array. No scan occurs when the array field is read.
The incoming object's contract is allocation evidence, never a contextual view's
readonly declaration. Source scalar slot checks remain active through aliases.

The native additions are `adamic_array.element_contract` and
`adamic_object.array_write_contract`, initialized to zero in constructors and
region objects. Zero is an explicit unsupported-production certificate. Dense
slice and the checked sparse slice preserve the element certificate. The two
IDs fit existing padding on this tested x86_64 toolchain: before and after,
object size/slots offset are 48/48, array size/elements offset 64/40. This is a
measured ABI result, not a promise for other architectures. JavaScript uses
WeakMaps. Ownership and destructors are unchanged.

The fixtures cover set, push, an explicitly typed empty array, slice mutation,
original alias mutation and surviving aliases after replacement, evaluation
order, and a malformed later element left untouched by reads. Negative writes
pin missing required source fields, an original literal field, readonly incoming
fields versus mutable originals, mutable field direction, uncertified optional
records, and an incoming field deinitialized after allocation. Node controls
are pinned even where Adamic intentionally stops: the staged-value control
prints `written`, while checked backends exit 70 naming `<array write>.value`
and the expected number versus uninitialized value.

Development runs exposed two bugs and are retained in the logs: contextual
readonly allocation inference rejected a valid alias replacement; and a zero
contextual scalar contract was appended into source metadata and panicked in
an existing lazy NodeArray probe. Original allocation mutability is now retained,
and unsupported contextual certificates are rejected before descriptor append.
A further alias probe pins the original literal slot refusal after the record
enters an array; omitting scalar write backpatching prints `written:1` instead.
The existing reference-write oracle now pins the more specific structural
refusal rather than its former blanket refusal.

| Family | Candidate pairs / reads | Ranked fixture obligations held | Candidate remainder |
| --- | ---: | ---: | ---: |
| Array contracts | 334 / 3,189 | 7 / 884 | 327 / 2,305 |
| Element or consumer reads | 251 / 1,602 | 0 / 0 | 251 / 1,602 |
| Intrinsic contracts | 179 / 794 | 0 / 0 | 179 / 794 |
| Own array fields | 30 / 72 | 3 / 26 | 27 / 46 |
| Tuples | 6 / 9 | 0 / 0 | 6 / 9 |

This foundation adds no ranked production pair credit: complete Type, Symbol
and Diagnostic records include fields outside this subset. Families overlap.
Checker-clean production IR and exact reachability remain unavailable. The
next unheld contracts are VariableDeclarationList.declarations (63 reads),
CallExpression.arguments (62), ObjectLiteralExpression.properties (49),
TupleType.elementFlags (43) and ArrayLiteralExpression.elements (42).

The estimate remains October 10, 12:00 MDT. The original date moved because I
underestimated optional-array descriptor repair, NodeArray ancestry/own storage,
and original-slot write certification. Integration/environment restart added
validation time. No shared-hook handoff is awaited; every new named hook and IR
or runtime addition is listed under Lane 2 in docs/checked-views-plan.md.
Callable files were not changed.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedView(Arrays|NativeArrays|NodeArrayRecords|ArraySearch|RankedArrayContracts|OptionalDeclarations|GenericArrayCast|ArrayReferenceWrites)$' -count=1 -timeout 15m
go test ./internal/ir ./internal/lower ./internal/native ./internal/javascript -run 'Test.*View|TestLazyView|TestSharedArrayContractAdapter' -count=1 -timeout 15m
python3 stage3/interface-downcasts/lane2/run-reference-array-mutants.py
```

The focused package selection passes: lowering 6.585s, native 16.199s,
JavaScript 1.291s; IR compiled, with no tests matching this selection. Full
repository gate was not run. Each mutant must cause an executed oracle mismatch;
build failures, compiler panics, frontend refusals and syntax failures do not
qualify. Original sources are restored in finally. Removing native retaining
produces ASan heap-use-after-free in the surviving-alias probe.

The restored existing array oracle gate passes in 19.428s before adding the
original-slot alias probe. Its final expanded run is included in the logs.

Final restored expanded array gate: PASS, 25.949s. Final package selection:
lowering 11.413s, native 15.600s, JavaScript 2.546s; IR builds, no matching
tests. All eight execution mutants pass their detection script; no source remains
mutated. Raw logs and ABI measurements are in reference-array-logs/.
