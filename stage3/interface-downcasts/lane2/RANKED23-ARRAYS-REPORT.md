Built producer signature checks for indexed and consumer callable-array reads, with original source contracts on writes.
Commits: 4712e290 owns all production adapters; original declarations, probes, counts and this report follow separately.
Commands: original callbacks, eight mutants, recent array joints and focused callable/array regressions pass in 34.465s; count refresh passes in 5.521s; vet passes.
Mutants: array kind, element kind, parameter signature, actual invocation signature omission, source write arity, physical result representation and two void adapters are caught by stopping or Node oracles in all three modes.
Not covered: optional/rest source-call ABI expansion, valued-result boxing across different physical representations, production consumer/intrinsic totals and remaining ranked pairs.

The complete private FileWatcherWithModifiedTime declaration is retained, including watcher, callbacks and modifiedTime. Preparation adds only a temporary export alias for that original private interface. It preserves all original fields and imports, emits 78 declarations, verifies all four original callback spans, and verifies the original callback's three parameters, two required arguments and void result. No reduced receiver or callable replaces the original declarations.

Indexed and map reads now select a producer signature by immutable generated code identity before allowing the declared calling convention. Initial probes exposed a native mismatch when a string argument reached a number-parameter function. The old function-kind check alone let the call run with a different numeric interpretation. The retained wrong-invoke regression now stops before invocation. Its signature-omission mutant completes without sanitizer errors but disagrees with Node in native output; the exit-70 signature oracle catches the omission in sanitized native, release native and JavaScript. The failed original discovery log is retained as observation, not certification.

Void callback reads preserve JavaScript's ignored extra arguments and discarded results. A producer using fewer parameters is checked against the corresponding prefix of the complete target signature; the existing closure argument-vector convention supplies the values. The remaining target parameters and original declaration are not deleted. A valued implementation behind a void callback is permitted when the caller discards its result. Observing a void call as a value remains the existing named compile refusal, verified against Node's observable return. Nonvoid physical result changes stop until a producer boxing adapter exists; a separate result-signature probe and mutant verify this boundary.

Writes keep the source allocation's original callable contract and check incoming producer metadata before retaining and storing. A replacement succeeds with compatible arguments. A three-parameter replacement into an original two-parameter array remains a named source-arity stop, protecting future calls through the original alias. Its descriptor mutant supplies the missing argument contract and finishes with Node output. Readonly-view admission never supplies source allocation evidence. A later cast also preserves a callable descriptor that an earlier helper already prepared; array interning no longer overwrites that certificate with a generic callable refusal.

Thirteen original fixtures cover direct calls, lazy unread bad storage, bad array and element kinds, bad parameter signatures, actual bad invocation, a later unread incompatible callback, replacement, map, a shorter valid implementation, a valued void implementation, the source-arity stop and the retained earlier frontier. Eight finishing controls and the six finishing stopping mutants have native leak checks. Two adapter mutants reject valid Node controls; their failure is checked against Node and pinned diagnostics. The actual-invocation omission mutant is deliberately not claimed to preserve Node output.

Commands, output retained under original23/evidence:

```sh
source /workspace/adamic-tools/env.sh
node stage3/interface-downcasts/lane2/original23/prepare.cjs /workspace/lane2-original-pin /workspace/lane2-original-declarations23
node stage3/interface-downcasts/lane2/original23/generate.cjs
export ADAMIC_ARRAY23_ORIGINAL_DECLS=/workspace/lane2-original-declarations23
export ADAMIC_TUPLE_ORIGINAL_DECLS=/workspace/lane2-joint-original-declarations
go test ./internal/oracle -run '^(TestCheckedViewRanked23OriginalArrays|TestCheckedViewRanked23ArrayMutants|TestCheckedViewCallableArrayResultStorage|TestCheckedViewCallableArrayVoidAdapterMutants|TestCheckedViewCallableArrayVoidResultObservation|TestCheckedViewRanked23ArrayCounts|TestCheckedViewArrayTupleJoints|TestCheckedViewFileNamesJoint|TestCheckedViewArrayReferenceWrites|TestCheckedViewNativeArrays|TestCheckedViewCallables)$' -count=1 -v -timeout 15m
go vet ./internal/ir ./internal/lower ./internal/native ./internal/javascript ./internal/oracle
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts
go test ./internal/oracle -run '^TestCheckedViewRanked23ArrayCounts$' -count=1 -v -args -update-counts
```

The required global count refresh fails in 40.024s in existing fixtures. The scoped refresh adds thirteen measured rows; removing them reproduces the previous table byte for byte. Automatic approval review rejected deleting the earlier unused frontier without evidence of safe removal. It was retained, added to the measured controls, and passes. No whole package or full gate ran. No cross-lane production file changes were needed in this group; existing lane 5 signature helpers are reused. The latest fetched integration remains 432d4913, already merged.

This adds one lane 2 array-field pair / four original reads. The ledger is now 184 pairs / 2903 reads certified, remaining 150 / 286 of 334 / 3189. Own fields remain 3 / 26 of 30 / 72. The three requested joints are already pushed. The next ranked fields have three reads each.
