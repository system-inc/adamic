Built: original tsc members at ranks 6, 26, 50 and 53; optional numeric callable parameters, scalar/object argument storage and complete long-signature refusals.
Commits: boxed adapters cb6adb25, six-member aggregate group 3e1a7f6f; this report identifies the next commit on codex/views-callables.
Commands: original AST verification PASS (4 members/240 candidate reads); Node/release-native/sanitized-native/JavaScript fixtures PASS (12.621s); restored uncached callable oracle PASS (88.849s), package guards PASS (13.757/1.405/7.812/1.699s), vet PASS and measured counts PASS (16.881s); raw logs in logs/aggregate.
Mutants: numeric native/JS arity, missing argument becomes zero, shortened diagnostic buffer; mixed native/JS arity and parameter admission, omitted field boxing. All nine killed and restored.
Uncovered: exact production reachability, full tsc execution, overload/generic/rest signatures, nominal producer parameter contracts, array signatures and class-method boxed variance. Rank 1 remains guarded pending predicate integration.

Original contracts certified in this group:
- NodeFactory.createPropertyAccessExpression, rank 6, 119 candidate reads.
- NodeFactory.createPropertyAssignment, rank 26, 56 candidate reads.
- NodeFactory.createElementAccessExpression, rank 50, 33 candidate reads.
- NodeFactory.createNumericLiteral, rank 53, 32 candidate reads.

mixed-aggregate-original-witnesses.json records exact original declaration and
read spans with source hashes at TypeScript pin
050880ce59e30b356b686bd3144efe24f875ebc8. Adjacent helpers, structural carriers
and the TokenFlags enum representation are reduced. Member declaration spellings
and witnessed call expressions are retained. These certify the member contracts
in fixtures, not the whole original compiler program. String/reference and
number/reference argument unions are exercised in both native and JavaScript.

Optional parameter admission preserves exact recorded parameter count. The
producer's incoming representation and independently recorded members still
have to accept undefined; a required numeric producer fails at the read.
Missing, explicit undefined and supplied numeric flags are observed against
Node. Missing numeric arguments use the existing MaybeNumber representation
and code(self, arguments, argument_count) convention. No alternate calling
convention, callable wrapper or erased signature is introduced. Overloads,
generic parameters and rest parameters remain refused.

Rank 50 exposed a concrete pre-call sanitizer crash: its contextual name field
was declared number | Expression, but record production stored raw number 3.
The subsequent union read retained the number bits as a pointer. The named
viewCallableBoxedRecordField hook makes supported contextual scalar/object union
fields store their boxed declared form. Removing it reproduces the crash and
kills the valid scalar-argument fixture. Object arguments remain valid. The
hook declines unknown, nominal and dictionary members; its admission is not a
blanket Union admission.

The longer rank 53 signature exposed a diagnostic capacity bug in
adamic_object_view: the expected type is formatted twice, while only one copy
was included in the allocation size. The minimal native/runtime/object.c hook
reserves both. The complete field/expected/found message is pinned on release
and sanitized builds. Restoring the short allocation kills that test.

Every member has valid calls, a number instead of a function, a wrong-arity
function and a wrong result representation. The mixed parameters also have a
producer too narrow for the declared union; numeric flags have a producer that
requires the optional parameter. Failures name the member and stop with exit 70
before the callback runs. Node's erased-type behavior is independently pinned;
valid runs match Node and have clean leak checks.

Conservative candidate inventory: 16/2818 pairs and 681/11063 reads certified;
2802 pairs and 10382 reads remain. Static inventory unchanged: 4/308 pairs and
34/1503 reads certified; 304 pairs and 1469 reads remain. None of these four
pairs appears in the static inventory. The reconciled 11648 reads remain an
Unknown-fallback upper bound; exact lazy reachability still requires checker-clean
tsc. The October 12, 18:00 UTC estimate remains provisional; exact scope and
unresolved signature families prevent a defensible guaranteed completion date.

Commands (output goes directly to logs/aggregate):
- node stage3/interface-downcasts/lane5/verify-aggregate-witnesses.cjs /workspace/scratch/lane5-ts-pin 6,26,50,53 mixed-aggregate-original-witnesses.json
- python3 stage3/interface-downcasts/lane5/run-numeric-mutants.py
- python3 stage3/interface-downcasts/lane5/run-mixed-aggregate-mutants.py
- go test ./internal/oracle -run '^TestCheckedViewCallable(ElementAccess|PropertyWitnesses|NumericLiteral)$' -count=1 -timeout 5m
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'Test.*View.*Callable|TestCheckedViewCallable|TestPrepareViewCallableRead|TestDefaultTaggedInterface' -count=1 -timeout 10m
- go test ./internal/ir ./internal/lower ./internal/native ./internal/javascript -run 'Test.*(Callable|CallTarget|LazyView)' -count=1 -timeout 5m
- go vet ./internal/ir ./internal/lower ./internal/native ./internal/javascript
- go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -update-counts -count=1 -timeout 5m

All minimal hooks are listed under lane 5 in docs/checked-views-plan.md. No other
lane branch was merged and no predicate guard removed. The full repository gate
is not claimed; the previously documented inherited closure harness still has
the obsolete two-argument function pointer in its test source.

Additional adjacent regression selection PASS (26.108s):
go test ./internal/oracle -run 'TestMixedUnion|TestOptional|TestCheckedViewObjectPrimitive|TestCheckedViewMixed' -skip '^TestOptionalCheckedWrites$/^nominal-(subclass-)?slot$|^TestOptional(CheckedReads|ClassReads|NullAndLiteralReads)$' -count=1 -timeout 5m
The skip names the inherited nominal-write and stale optional-read pins previously
reproduced on the integration parent. Their failures are not counted as passes.
