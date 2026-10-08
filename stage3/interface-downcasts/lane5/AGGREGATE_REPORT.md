Built: six original-member aggregate callable fixtures, deferred payload schemas, transitive callback parameter/result reads and callable return-allocation demand.
Commits: follows boxed-union adapter cb6adb25; this checkpoint is the commit containing this report on codex/views-callables.
Commands: Node original-span verification PASS (6 members/326 candidate reads); uncached restored callable oracle PASS (57.928s); aggregate counts PASS (12.705s); touched-package guards and vet PASS, raw logs in logs/aggregate.
Mutants: native/JavaScript arity admission, omitted payload read registration and omitted callable result demand all killed; every source restored.
Uncovered: exact tsc reachability/full-program compilation, overloaded/generic/rest/optional signatures, nominal producer parameters and class-method boxed variance; Debug.assert remains guarded pending predicate integration.

Ranks 7, 22, 31, 36, 51 and 52: NodeFactory.createExpressionStatement (114),
DiagnosticCollection.add (59), NodeFactory.createVoidZero (46),
NodeFactory.createThis (42), TransformationContext.requestEmitHelper (33),
and EmitResolver.hasNodeCheckFlag (32). Declaration spellings and original read
expressions are retained; adjacent helpers and structural data carriers are
reduced. verify-aggregate-witnesses.cjs checks AST spans, source hashes and
original declarations against the independent TypeScript pin
050880ce59e30b356b686bd3144efe24f875ebc8. Its output is
aggregate-original-witnesses.json. This is member-contract fixture certification,
not a claim that the whole original tsc call graph was executed.

Each member has a valid call and pinned exit-70 refusals naming the field,
expected signature and encountered number, wrong arity or wrong result
representation. Release native, ASan/UBSan native and JavaScript are checked;
valid executions match Node, including leak checks. Returned objects do not
receive recursive eager admission: their declared fields are checked at the
read, including the nested expression read inside a shared helper. Callback
parameters with compatible Object representations but narrower field declarations
also check on their reads. Wrong payload and wrong callback-parameter payload
fixtures pin those stops. No arbitrary Object equivalence is treated as proof
of its descendant values.

The returned-demand regression proves a known returned allocation remains in
lazy demand when the callable owner's flow is Unknown. A disjoint ordinary
call is its admitted control. Removing the graph hook loses the required
unsupported-field refusal and kills that test. Removing read registration
allows three wrong returned payloads to print instead of refusing. The two
arity mutants allow all six wrong-arity functions to execute. Logs contain
executable failures, not clang failures. The earlier seven boxing mutants
remain documented in BOXING_REPORT.md.

Conservative inventory: 12/2818 candidate pairs and 441/11063 candidate reads
certified; 2806 pairs and 10622 reads remain. Static inventory: 4/308 pairs and
34/1503 reads certified; 304 pairs and 1469 reads remain. These six members are
absent from the static inventory, so they are not double-counted there. The
11648 reconciled-read inventory remains an upper bound; exact lazy reachability
needs the checker-clean original program. No exact demand claim is made.

All shared hooks are named and listed in docs/checked-views-plan.md under lane 5.
No other lane branch was merged. Oct 12 at 18:00 UTC remains a provisional
planning estimate, not a certified completion date: unresolved optional,
overload, generic, array and predicate boundaries prevent a defensible promise
for all candidates. The compiler-diagnostic ledger and predicate integration
are dependencies of exact measurement and rank 1 respectively; these fixtures
were built without waiting for them.

Validation commands (stdout/stderr go directly to the named log files):
- node stage3/interface-downcasts/lane5/verify-aggregate-witnesses.cjs /workspace/scratch/lane5-ts-pin
- python3 stage3/interface-downcasts/lane5/run-aggregate-mutants.py
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'Test.*View.*Callable|TestCheckedViewCallable|TestPrepareViewCallableRead|TestDefaultTaggedInterface' -count=1 -timeout 10m
- go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -update-counts -count=1 -timeout 5m
- go test ./internal/ir ./internal/lower ./internal/native ./internal/javascript -run 'Test.*(Callable|CallTarget|LazyView)' -count=1 -timeout 5m
- go vet ./internal/ir ./internal/lower ./internal/native ./internal/javascript

The full repository gate was not run. BOXING_REPORT.md records the inherited
stale two-argument graph closure harness excluded from the focused selection;
the integrated calling convention takes self, arguments and argument count.
