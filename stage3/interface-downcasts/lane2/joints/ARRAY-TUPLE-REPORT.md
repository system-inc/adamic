Built live array-backed tuple reads, checked tuple lengths, and original primitive-union source writes.
Commits: e3cb66a1 own adapters, f4563b82 shared hooks, 78c73bab lane 4 union reads, c451c9e8 lane 4c runtime and frontier.
Commands: focused original tuple and array checks pass in 16.094s; final negatives, mutants, boundary and count verification pass in 9.892s; vet passes.
Mutants: first-position kind, second-position membership, receiver arity, length arity and source-write membership all fail their pinned stopping oracles in all three modes and finish with Node output.
Not covered: reference-element and readonly source arrays, tuple writes, optional/rest target casts and consumers requiring raw object storage; fileNames and ranked callbacks remain next.

The original lane 4c cast frontier now returns string in sanitized native, release native and JavaScript, matching Node. The complete original tuple is [fileId: IncrementalBuildInfoFileId, signature: [] | EmitSignature]. Its four original positional source spans are verified by tuples/prepare.cjs against the pinned compiler source; declaration preparation retains all 78 declaration files. The new four-reads fixture exercises both positions and both repeated signature reads without reducing the declaration. Each runtime read checks live source-array arity, storage, position and readiness. The cast retains the original allocation, and the alias control reads second after updating the original array. Unread wrong elements and unread wrong arity remain lazy.

The five stopping negatives pin complete messages in checked_views_array_tuple_joints_test.go. They stop at a wrong numeric position, excluded numeric signature, changed tuple arity, changed length or a boolean write into an original string | number array. Each mutant changes only the corresponding logical descriptor or arity. All five finish with Node output, empty stderr and no native leaks. Seven successful owned fixtures, the handed-off frontier and all finishing mutants have native leak checks. JavaScript follows the same exact messages. Scalar source storage is also exercised by the number[] control. Structural JSON serialization retains its existing named compile refusal. An unrelated primitive array still serializes in a program containing the tuple view.

Commands, output retained under tuple/evidence:

```sh
source /workspace/adamic-tools/env.sh
node stage3/interface-downcasts/tuples/prepare.cjs /workspace/lane2-original-pin /workspace/lane2-joint-original-declarations
export ADAMIC_TUPLE_ORIGINAL_DECLS=/workspace/lane2-joint-original-declarations
go test ./internal/oracle -run '^(TestCheckedViewArrayTupleJoints|TestCheckedViewArrayTupleConsumerBoundary|TestCheckedViewTupleLane4bCastFrontier|TestCheckedViewTupleOriginalSignaturePositions|TestCheckedViewTupleOriginalReferencedMap|TestCheckedViewArrayReferenceWrites|TestCheckedViewNativeArrays)$' -count=1 -v
go vet ./internal/ir ./internal/lower ./internal/native ./internal/javascript ./internal/oracle
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts
go test ./internal/oracle -run '^TestCheckedViewArrayTupleJointCounts$' -count=1 -v -args -update-counts
go test ./internal/oracle -run '^(TestCheckedViewArrayTupleJoints|TestCheckedViewArrayTupleConsumerBoundary|TestCheckedViewArrayTupleJointCounts)$' -count=1 -v
```

The required global count refresh fails in 52.188s in pre-existing fixtures and leaves counts.md unchanged. The scoped refresh passes in 4.410s and adds twelve measured rows. Removing those rows reproduces the prior table byte for byte. No whole package or full gate ran. Initial probe failures were fixture typing mistakes in numeric logging and structural JSON typing; corrected controls and final evidence pass.

Cross-lane commits are isolated above. f4563b82 adapts shared lowering, reference conversion, length metadata and receiver hooks. 78c73bab changes only the two lane 4 object/primitive union emitters. c451c9e8 changes only lane 4c tuple runtime shape checks and its frontier oracle. Native object-layout consumers are conservatively closed when the array alias can reach them; the shared allocation graph supplies reachability. This is a fixed-position storage certificate, not general tuple-object storage conversion.

This certifies the requested tuple frontier: one tuple element pair, four original reads. It receives no duplicate lane 2 array-field credit. Lane 2 remains 182 pairs / 2895 reads, remaining 152 / 294 of 334 / 3189. Own fields remain 3 / 26 of 30 / 72. fileNames is next, followed by the preserved ranked callback group.
