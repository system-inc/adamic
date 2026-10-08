Built: eight original-declaration callable pair contracts, 498 candidate reads, toward roadmap step 09.
Commits: 1da34a72 first three overload pairs; c1f80ec5 import/yield pairs; d48d2ad0160de5f79cf9c3548e2adf922f89e15c generic and higher-order code.
Commands and outputs: final filtered uncached tests pass in all five packages; 26 own fixture rows and 90 inherited rows pass; required global counts update fails outside this unit.
Mutants: twelve own IR changes and eleven compiler guard changes are caught; the final regression also catches all 60 inherited arity-mutant executions.
Uncovered: capturing generic producers, generic methods, deeper or generic callback descriptors, finite enum domains, whole-tsc execution and exact runtime reachability.

Delivery branch: codex/views-callables-factory, based on a62778110b958f54d52622da01bf7758bce629dc. No lane branch, main, or integration branch was merged. The protected compiler/oracle files named in the unit were not edited. New Adamic source uses .a. Existing checker shims and type-mapper access are referenced through the cohere dependency; no cohere code was copied.

| Pair | Candidate reads | Contract exercised |
| --- | ---: | --- |
| NodeFactory.createIdentifier | 206 | Both original signatures, omitted and explicitly undefined parameters |
| NodeFactory.createStringLiteral | 104 | Both original signatures and boolean packing |
| NodeFactory.createUniqueName | 58 | Both original signatures, optional number/string/object domains |
| NodeFactory.createImportClause | 9 | Boolean and phase-modifier signatures, present and absent object arguments |
| NodeFactory.createYieldExpression | 9 | All three declared signatures, present and absent asterisk/expression arguments |
| NodeFactory.cloneNode | 91 | T instantiated as Node, a subtype with label, undefined, and Node or undefined |
| NodeFactory.createModifier | 10 | Separate literal instantiations, with a kind: T carrier field |
| TransformationContext.onEmitNode | 11 | Outer signature, complete nested callback signature, callback argument check |
| Total | 498 | Eight original member declaration fixtures |

These are source-member contract certificates under the producer scopes below, not execution of all inventoried reads in tsc. The read counts and original witness provenance come from the pinned share-a needs-code.json entries. certificates.json retains their declarations, source pin, hashes and read expressions. The declaration test verifies every fixture contains its original declaration and original read. For the first three pairs the original good control is byte identical. For import, yield, clone, modifier and onEmitNode the original reduced producer is retained byte identical as the stopping negative; positive controls change the producer to fulfill the original declaration.

Calls use GetResolvedSignature on the call node. Arguments are fitted to that signature's parameter representations. An overload resolution must name one of the declared set's declarations. The viewed member read checks the entire overload set against immutable producer code metadata, including each parameter domain and result. Overloads form a conjunction. Extra evaluated arguments may be ignored by a shorter producer; omitted producer parameters must admit undefined. Existing single-signature rules stay strict. The yield third signature is retained and reached; the omission mutant selects only its second signature and does not claim an independently distinguishable third producer domain.

Generic reads require an immutable named module producer with the same quantified declaration after alpha-renaming: constraints, parameters and result all match. This deliberately uses exact declarations rather than method bivariance. A canonical template closure preserves function identity. Each reached call selects a separate instantiated body through the template code pointer. The existing instantiation cache uses checker type identity, not only native representation. Each returned type is rechecked against that instantiation before body lowering; existing instantiated-write checks also run. The clone fixture has four distinct bodies although all four types have the same object representation. Modifier has two distinct literal bodies, and its conversion fixture retains kind: T in the returned carrier. This is a static instantiation result proof; it does not substitute the constraint for T. Ordinary generic class methods cannot borrow the module producer certificate. Generic arrows, captured or nested generic producer environments, unconstrained generic values and generic overload sets retain their refusals.

Higher-order parameter descriptors retain the callback's parameter and result signatures in both the producer metadata and the declared contract. Nested variance is checked at the member read. The callback actually passed is checked against its descriptor at the invocation boundary. The check returns the original callable value and preserves existing receiver binding. One nested callback level is supported. Recursive, optional or generic callbacks and callback-valued result descriptors remain unsupported. Existing conservative method checks for union ABI adaptation are preserved.

All 26 own .a fixtures are held to source Node. Finishing positives agree with native release, native ASan/UBSan, and the JavaScript backend; they also pass the independent native leak check. Runtime negatives pin empty stdout, exit 70, and the complete field/signature/found message. The malformed callback boundary is an IR witness derived from the Node-valid boundary fixture and pins call argument 3. The temporary generic result-cast controls are rejected by the existing no-unchecked-cast rule before instantiation; they are not presented as evidence for the new result check. The new result-check test instead proves that replacing a Detail instantiation by the broader Node constraint must fail.

Own IR mutants, with source Node determining finishing output and independent leak checks on every finishing mutant:

| Change | Witness and what catches it |
| --- | --- |
| Keep only overload 1, separately for identifier, string literal and unique name | Each wrong-overload control prints 7 in all three execution modes; its pinned signature-2 exit 70 catches the omission |
| Remove resolved undefined packing, separately for those three pairs | Both native modes finish with different output: identifier 9/110/0, string literal 3/110/0, unique name 11/100/3; Node requires 9/110/110, 3/110/110, 11/100/103 |
| Keep only overload 2, separately for import and yield | Each original one-arm producer prints 7 in all three modes; pinned signature-1 exit 70 catches it |
| Admit the ordinary producer at the generic read and instantiation dispatch, separately for clone and modifier | Each non-generic original producer prints 7 in all three modes; the pinned unknown-signature exit 70 catches it |
| Erase onEmitNode expected outer arity | Its zero-argument producer prints done in all three modes; pinned arity-0 exit 70 catches it |
| Erase the callback argument descriptor | The boundary witness prints wrong/done in all three modes; pinned call-argument arity-0 exit 70 catches it |

Those are twelve distinct IR changes and 33 finishing backend executions. The inherited share-a regression additionally exercises 30 distinct arity changes in release native and JavaScript, producing 60 inherited mutant observations.

Isolated compiler guard mutants are reproducible through run-mutants.py, run-generic-mutants.py and run-nested-mutants.py. Every runner restores exact source bytes, requires a failing test and checks its specific failure observation:

| Mutant | What caught it |
| --- | --- |
| Native overload result variance | Wrong-result fixture finishes with read instead of exit 70; sanitized native also finishes cleanly |
| JavaScript overload result variance | Wrong-result fixture finishes with read instead of exit 70 |
| Overload-bearing union refusal | TestPrepareViewCallableRead reports unsupported read admitted |
| Resolved overload declaration membership | TestResolvedCallableDeclarationMembership reports undeclared resolution admitted: nil |
| Generic method producer identity | Ordinary class method prints 7 instead of read rejection; release and sanitized native finish cleanly |
| Generic instantiation cache type identity | Clone's four instantiated bodies collapse to one; the fixture's distinct-instantiation assertion fails |
| Generic declared result identity | A wider Node-or-undefined result is admitted for T; the declaration identity assertion fails |
| Generic instantiated result check | Node is accepted where Detail was instantiated; the result proof assertion fails |
| Generic instantiated effects | ClosureTargets follows template function 0 rather than instantiated bodies 2 and 3; its effect assertion fails |
| Native nested callback comparison | Wrong-nested fixture prints done instead of exit 70; sanitized native also finishes cleanly |
| JavaScript nested callback comparison | Wrong-nested fixture prints done instead of exit 70 |

The first four source guard logs are retained from the first delivery; their checks remain in the final compiler. The final generic and nested guard observations are in logs/generic-guards.log and logs/nested-guards.log, with each individual failing-test log adjacent. No compiler mutation is left applied.

Exact final commands, with output redirected directly to files:

```
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/ir ./internal/lower ./internal/native ./internal/javascript ./internal/oracle -run '^Test(ResolvedCallableDeclarationMembership|PrepareViewCallableRead|ViewCallableShapeContract|GenericCallableDeclarationIdentity|GenericCallableResultInstantiation|GenericClosureTargetsUseInstantiatedBodies|ClosureTargetsBoundOnlyProvenValues|GenericUnionFixtureHasSeparateInstances|CensusGenericRefusalInsideClosure|ViewCallableAggregateReturnedDemand|ViewCallableShapeNative|ViewCallableShapeNode|ViewCallableProducerCertificateNative|ViewCallableProducerCertificateNode|ViewCallableBoxingUnknownProducer|CheckedViewCallableFactory.*|CheckedViewCallableShareA.*)$' -count=1 -v -timeout 10m > /tmp/views-callables-factory-complete-restored.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableFactoryCounts$' -count=1 -args -update-counts > /tmp/views-callables-factory-counts-all.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 5m -args -update-counts > /tmp/views-callables-factory-counts-global-all.log 2>&1
go vet ./internal/ir ./internal/lower ./internal/native ./internal/javascript ./internal/oracle > /tmp/views-callables-factory-vet-final.log 2>&1
python3 stage3/interface-downcasts/lane5/share-factory/run-generic-mutants.py > /tmp/views-callables-factory-generic-guards-complete.log 2>&1
python3 stage3/interface-downcasts/lane5/share-factory/run-nested-mutants.py > /tmp/views-callables-factory-nested-guards.log 2>&1
git diff --check
```

The restored focused test results are IR 0.106s, lower 3.031s, native 9.026s, JavaScript 1.400s and oracle 142.722s, all PASS. The own counts updater passes in 6.388s. Twenty-six rows have been appended across the three delivery groups; existing rows are unchanged. The final regression verifies those rows and the 90 inherited share rows. Vet and whitespace checks pass. No whole-package test or full gate was run.

The required global counts updater exits 1 in 79.289s with failures outside these fixtures and does not write a partial table. Four representative failures, census_overload_contracts, census_small_boolean, maybe_number_slots and graph_regions_regression_07, were reproduced on the unchanged a6277811 base during the first delivery; logs/counts-baseline.log records that evidence. Other global failures remain untriaged. The global table is not certified. Installing the existing pinned stage3 API dependencies removed the initial package setup failure.

The first combined final attempt passed the own eight-pair tests but exposed two test changes: the existing C signature initializer lacked the new metadata field, and a broad replacement accidentally changed inherited receiver-name pins. The initializer was completed, original receiver-name assertions were restored, and the same combined command passes. logs/complete-initial-regression.log and logs/complete-restored.log retain both observations. The initial descriptor-erasure higher-order mutant remained rejected as an unknown descriptor; it is not counted. Direct source comparison omissions provide the two counted nested guard mutants instead.

Setup was completed before building: GOPROXY=https://proxy.golang.org|direct, bash cloud/setup.sh, then source /workspace/adamic-tools/env.sh. Reported cumulative readiness lines: Node 0.167s, Go 0.168s, clang 0.875s, markdown 1.491s, submodules 217.572s, go build 615.586s, cache 615.765s, done 615.813s. nproc is 5; cpu.max is 400000 100000. logs/setup.log retains the output.

All eight needs-code entries now have original-declaration positive, stopping negative and mutant evidence within the supported scopes. Capturing generic producer environments and deeper callback contracts still need compiler work. Adjacent enum carriers remain reduced to number; this certifies no finite enum membership. Original Node and token carriers are reduced, so this certifies their member signatures rather than complete upstream object schemas. Full tsc compilation/execution, all actual factory producer bodies, and exact read reachability remain unmeasured. Integration remains owned by the views integration branch.
