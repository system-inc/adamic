# Escaped adapter length and name

Source function reflection is recorded separately from callable contracts and
physical argument slots. Native and JavaScript select that metadata by the actual
underlying code identity, after adapter normalization. Implementation forwarders
follow their original source function. Optional parameters count toward length;
the first default or rest parameter stops it. Named, inferred, computed constant
and anonymous names retain their Node spelling.

Reading does not invoke the function or check its argument/result relation. The
misfit producers in the length/name witnesses remain uncalled. The two .a
witnesses still refuse the unproven escape with its original path and fix.
Runtime headers and the cleared cache were unchanged.

ASan exposed an ownership bug when a declaration's adapter was treated as a
borrowed field read. The field-chain planner now excludes adapter-producing .ts
reads. A self-restoring source mutant removes that exclusion; ASan catches the
use-after-free in TestV4EscapeAdapterName. Its patch, raw output and decoded ASan
trace are saved beside this report.

## Commands and observations

Every Go test used -count=1 -timeout 90s and an outer timeout 90. Test output went
to the named log files in this directory.

- go test ./internal/ir -run 'Test(CallTargetReaders|CallTargetsIncludeEveryDescendant|ClosureTargets)': PASS, 15.563 seconds, surface-call-target-readers.log.
- go test ./internal/oracle -run '^TestV4EscapeAdapter(Length|Name|Surface.*)$' -v: PASS, 1.910 seconds, surface-tests.log. Native counted ASan/UBSan/LSan, ordinary native and JavaScript match source Node.
- go test ./internal/oracle -run '^TestV4EscapeAdapter(Argument|Result|Identity|NoStacking)$' -v: PASS, 1.488 seconds, surface-adapter-regression.log, including those existing mutants and unchanged bounded-allocation rows.
- go test ./internal/lower: PASS, 29.277 seconds, surface-lower-final.log. The initial run failed because stage3/api lacked @types/node 25.3.3; npm ci --ignore-scripts --prefix stage3/api installed the pinned dependencies in 0.884 seconds without changing manifests.
- go test ./internal/native -run 'Test.*(ViewCallable|BorrowChain|ClosureConvention)': PASS, 7.695 seconds, surface-native-focused.log. The earlier whole-package attempt reached its outer 90-second limit without recording a result; it is not counted as a pass.
- go test ./internal/javascript: PASS, 0.670 seconds, surface-javascript.log.
- python3 review/compiler/views-v4/run-surface-borrow-mutant.py: ASan caught heap-use-after-free; source restored byte for byte, surface-borrow-mutant-result.log.

The new oracle leaf seconds were Length 1.59, Name 1.43, Defaults 1.00, Returned
1.00, LengthRefusal 0.32, NameRefusal 0.14, ComputedName 1.06 and AnonymousName
0.89. All have top-level parallel tests. No persistent oracle fixture was added,
so there are no new or changed counts.md rows.

The surface omission mutants bypass only reflection's underlying normalization.
Both finish cleanly with exit zero and wrong Node output: length is 2 instead of
3, name is empty instead of producer. Both native sanitizer and JavaScript runs
catch each mutant. Argument and result checks remain present.

This unit does not implement call, apply, bind, new or actual later-call-site
blame. Optional reflection and opaque library function observations retain their
existing unsupported status. Runtime pool and region-adoption pendings remain.
