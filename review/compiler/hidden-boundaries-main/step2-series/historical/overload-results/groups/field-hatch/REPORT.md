Built the ruled step 05 TypeScript overload field checks; unproved Adamic structural results remain refused.  
Compiler 0cfe794bf34086b31382c35e015b703555feefc2, after merging main 54cbc125 through e19e87c7.  
Focused lower, oracle, CLI, counts, local a-check, and the requested whole flow package pass.  
Removing the kind check or value check lets the wrong result through; both mutants fail their pinned oracle assertion.  
No per-overload structural body proof or escaped overload function boundary was added; generic function values remain NotYet.

The narrower resolved call invokes the implementation once and checks exactly
result.kind or result.value before the caller reads it. The kind fixture uses the
numeric SyntaxKind.Block discriminant. The value fixtures cover strings and
undefined, with a planted numeric liar. Checks preserve the original object and
appear once per emitted narrowed call in the existing PredicateChecks report.
Each fixture has two narrowed calls and one wider call; the report counts two,
and successful execution observes three implementation invocations. The CLI
checks that count for c, js and build --explain-checks. The shared wider result
is inspected through DynamicProperty, which boxes the physical field correctly.
Named closure calls keep their implementation argument and result slots until
the wrapper converts them. Numeric arguments must be boxed for the closure ABI.
The sanitizer caught a premature unboxed argument during development; the final
valid string, undefined and wrong-number fixtures hold that correction.

This is the sanctioned TypeScript hatch, not a return proof. Nominal results,
accessors, optional failing fields and unrelated field conversions retain stops.
The value hatch needs readonly fields and identical contracts for its other
fields. An Adamic version of every witness refuses with its result path and
failed readonly covariance. New broader acceptance still needs its own ruling.
No checker option was weakened. No runtime, backend emitter or protected
orchestration file was edited, and no cohere source was copied.

| Region | Hidden before | Hidden after | Additional revealed | First stop | Owner / raising function |
|---|---:|---:|---:|---|---|
| hidden-01-large | 13,625 | 13,625 | 0 | transformers/declarations.ts:1678:110; TS2345 | checker: isSignatureApplicable |
| hidden-01-small | 6,899 | 6,899 | 0 | transformers/declarations.ts:1678:110; TS2345 | checker: isSignatureApplicable |
| hidden-05-large | 6,078 | 6,078 | 0 | visitorPublic.ts:123:5; visitNode parameter node | censusOverload |
| hidden-05-small | 5,748 | 5,748 | 0 | visitorPublic.ts:123:5; visitNode parameter node | censusOverload |
| hidden-06 | 3,369 | 3,369 | 0 | visitorPublic.ts:196:5; visitNodes visitor: Node must fit TIn | censusOverload |
| hidden-13 | 7,289 | 6,965 | 324 | transformers/es2017.ts:764:38; TS18048, outerParameter possibly undefined | checker: checkNonNullTypeWithReporter |
| hidden-14 | 7,102 | 7,102 | 0 | utilities.ts:11445:12; returning evaluate as an indirect overload value | overloadDirectUses |

The comparison verifies all 82 adapted source hashes at pin 388096e6. The before
raw census is ../admission/after.jsonl.gz at compiler c09a9604; that compiler and
9b3882f3 have identical production code. Merging main changed no lower, load or
IR source, so that baseline remains applicable. after.jsonl.gz is the final
measurement-only replay; it emits no executable. The original seven-interval
scope patch is ../c68/census-scope.patch.gz. This is an intersection measurement
on a checker-rejected project, not a whole-compiler emission claim.

13 reveals exactly [30237,30561), the 324-byte overload signature span. The
implementation body is still checker-excluded at es2017.ts:738:5 because its
first diagnostic is es2017.ts:764:38, TS18048 on outerParameter. The checker
raises it through checkNonNullTypeWithReporter and
reportObjectPossiblyNullOrUndefinedError; latentFullSelected /
LatentOwnDiagnosticsIn exclude that body. There is also a later lowering refusal
in independently replayed statements: es2017.ts:739:9 reaches
visitorPublic.ts:393:126, visitParameterList parameter nodesVisitor. That is not
the first stop in the interval.

14 passes the old result admission and reaches overloadDirectUses at
utilities.ts:11445:12: createEvaluator returns evaluate. The associated independent
unit begins at utilities.ts:11322:5 inside the interval. It stays NotYet until
that escaped function can retain checked resolved-overload boundaries. The
outer destructuring refusal at utilities.ts:11316:35 is outside the interval.
01 remains on stricter checker options #k881crd, 05 on checked views ★11, and
06 on the visitor invocation-domain proof. The preceding 8,048-byte reveal in 06
is retained. Total assigned hidden bytes are 49,786, down from 50,110.

The gate regression came from flow's positive *.a glob, which does not consult
the oracle registry. Both TNode value and mutation fixtures were already
registered as NotYet. They now live in internal/oracle/testdata/notyet, outside
that glob, with their negative registration retained. The generic value stop is
pinned exactly to hidden_boundary_generic_tnode_value.a:3:36:

```text
stage 0 can't lower a generic function as a value yet
```

Its source still runs on Node and prints node. The mutation fixture remains
NotYet at 3:79, assigning a field of a union of differently held members. The
owning mechanism begins at 9d1d672c, Specialize generic function values for
concrete contextual signatures, on codex/generic-function-value (inspected tip
5d1b45e1). That is another worker's dependency, so it was inspected without
merging or copying its implementation. No generic-function-value acceptance is
claimed. The requested flow package and registered negative oracle fixtures ran
on codex/overload-results after merging main 54cbc125.

Commands and outputs, after sourcing /workspace/adamic-tools/env.sh:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/overload-hatch-setup.log 2>&1
nproc
# 5
go test ./internal/lower -run '^(TestOverloadStructural|TestOverloadResults|TestOverloadCallback|TestCensusOverload)' -count=1 > /tmp/overload-hatch-final-lower.log 2>&1
# ok, 2.944s
go test ./internal/oracle -run '^TestOverloadFieldHatch$' -count=1 > /tmp/overload-hatch-final-oracle.log 2>&1
# ok
go test ./internal/oracle -run '^TestNativeAgreesWithNode/internal/oracle/testdata/overload_field_hatch/' -count=1 -v > /tmp/overload-hatch-final-positive.log 2>&1
# all three positive fixtures pass
go test ./cmd/adamic -run '^(TestExplainOverloadFieldChecks|TestExplainChecksOutput)$' -count=1 > /tmp/overload-hatch-final-cli.log 2>&1
# ok, 2.304s
python3 docs/overload-results/groups/field-hatch/run-mutants.py > /tmp/overload-hatch-final-mutants.log 2>&1
# drop-kind-check caught; drop-value-check caught; each mutated oracle exits 1
git show 89cbe74a5009047b0fb2fd31c323794d00dbcc12:cloud/fast-gate/run.py > /tmp/overload-hatch-fast-gate.py
python3 docs/overload-results/groups/field-hatch/a-check.py /tmp/overload-hatch-fast-gate.py > /tmp/overload-hatch-final-acheck.log 2>&1
# PASS, all five Adamic refusal witnesses
go test ./internal/flow > /tmp/overload-hatch-flow.log 2>&1
# ok, 118.166s
go test ./internal/oracle -run '^TestHiddenTNodeValueExactNotYet$' -count=1 -v > /tmp/overload-hatch-generic-oracle.log 2>&1
# PASS, exact 3:36 stop and source Node output
go test ./internal/oracle -run '^TestNativeAgreesWithNode/internal/oracle/testdata/notyet/hidden_boundary_generic_tnode_' -count=1 -v > /tmp/overload-hatch-generic-registered.log 2>&1
# both NotYet registrations pass
go test ./internal/lower -run '^TestHiddenTNode' -count=1 > /tmp/overload-hatch-generic-lower.log 2>&1
# ok
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/overload-hatch-final-counts.log 2>&1
# ok, 38.792s
```

Positive oracle runs cover source Node, emitted JavaScript, ASan/UBSan native,
release native and leaks. Wrong-result fixtures pin the deliberate exit 70 rather
than compare their stop to Node, which accepts and prints the planted result.
Both mutants fail at the runtime outcome assertion, not a build or sanitizer
failure. The unchanged old report golden fixtures also pass. Counts gained only
three rows; existing numeric counts did not change. No full gate was run; the
whole flow package was explicitly requested for the reported regression.

Setup cumulative timings: Go 0.023s, Node 0.024s, submodules 0.061s, markdown
0.066s, clang 0.149s, Go build 41.222s, deferred test binaries 41.444s, warm
cache 41.445s, done 41.472s. nproc is 5; cgroup cpu.max is 400000 100000.
Full outputs, raw census, byte comparison, exact next stops, local a-check results
and mutant evidence are preserved alongside this report.
