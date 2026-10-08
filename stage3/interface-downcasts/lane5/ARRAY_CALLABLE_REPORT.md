Built: represented array callable parameters and deferred element schemas; original ranks 13 and 49 certified against Node in both backends.
Commits: follows ebb1f8a6; this report identifies the next commit on codex/views-callables.
Commands: original AST verification PASS (2 pairs/124 candidate reads); 11 member fixtures PASS (5.190s), including sanitized native and leak checks; restored uncached oracle PASS (75.762s), package guards PASS (13.347/1.857/7.258/1.957s), vet PASS and counts PASS (16.944s), logged in logs/aggregate.
Mutants: native/JS wrong-arity admission and omitted array payload schemas killed for both original member fixtures; every source restored.
Uncovered: exact reachability, full compiler execution, overloaded/generic/rest shapes; existing optional-boolean and owner-mutation callback refusals are retained.

Ranks 13 NodeFactory.createCallExpression (89 reads) and 49
NodeFactory.inlineExpressions (35 reads) share the array parameter family.
array-original-witnesses.json verifies declarations and property-read spans
at TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. Adjacent data carriers
and helpers are reduced. Rank 49 retains the original member call expression;
its surrounding || fallback is reduced because these fixture inputs yield a
truthy result. No full-program execution is claimed.

Array representations are compared at the callable read. Expected parameter and
producer element schemas are completed before lazy demand, so an array consumed
inside a callback or a previously lowered helper has read checks too. A callback
that expects numeric elements from a supplied object array passes only the
physical Array shape; its indexed read stops with exit 70 naming the index,
expected number and found object. Scalar/object element representation equality
is not trusted. Undefined array arguments preserve their existing pointer and
argument-count behavior; valid and absent executions match Node.

Every member has a valid read/call, a number in the callable slot, a wrong-arity
producer and a wrong result representation. The element mismatch is pinned in
release native, sanitized native and JavaScript. Deleting both expected and
producer array schemas lets the wrong-element fixtures execute and kills both
oracle tests. Both arity mutants also execute the otherwise-refused producers.
The mutations fail executable checks, not compilation.

Candidate totals: conservative 18/2818 pairs and 805/11063 reads certified;
2800 pairs and 10258 reads remain. Static 4/308 pairs and 34/1503 reads certified;
304 pairs and 1469 reads remain. These two pairs are absent from the static
inventory. Exact production reachability remains unmeasured.

Only lane-owned shape/payload helpers changed; the plan names both hooks.
No callback slot guard, owner-mutation guard or Union exception changed.
The integrator retains responsibility for the owner-change/entry-live-mutation
reconciliation. Rank 14's optional boolean call still needs the existing callback
refusal resolved soundly; it is not counted by this report.

Commands, redirected to logs/aggregate:
- node stage3/interface-downcasts/lane5/verify-aggregate-witnesses.cjs /workspace/scratch/lane5-ts-pin 13,49 array-original-witnesses.json
- python3 stage3/interface-downcasts/lane5/run-array-callable-mutants.py
- go test ./internal/oracle -run '^TestCheckedViewCallableArrays$' -count=1 -timeout 5m
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'Test.*View.*Callable|TestCheckedViewCallable|TestPrepareViewCallableRead|TestDefaultTaggedInterface' -skip '^TestCheckedViewCallableCounts$' -count=1 -timeout 10m
- go test ./internal/ir ./internal/lower ./internal/native ./internal/javascript -run 'Test.*(Callable|CallTarget|LazyView|ArrayContract)' -count=1 -timeout 5m
- go vet ./internal/ir ./internal/lower ./internal/native ./internal/javascript
- go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -update-counts -count=1 -timeout 5m

The full repository gate is not claimed. Counts are checked separately because
this lane's count-recording command writes the table. The unrelated inherited
closure harness and stale optional pins remain documented in prior lane reports.
