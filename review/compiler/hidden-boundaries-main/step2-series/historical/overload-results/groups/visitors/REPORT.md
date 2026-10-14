Built coupled visitor input proofs and counted implementation-invocation checks under step 05, toward roadmap step 16.
Compiler 5a718e9356a53f6909ae6abc227bf50da1ce2952; prior region 14 delivery 3f65645870580e574b909f22b222151fd0dacdd1; main 54cbc125 already merged.
Focused lower, Node/JavaScript/native sanitizer and release oracles, CLI checked sites, local a-check and records pass.
Erasing TIn metadata and admitting an unproven invocation both fail the planted exit-70 oracle; region 14's two mutants were caught in its earlier delivery.
No additional region 06 bytes: its independent visitNodes result remains unproved; 01/13 checker stops and 05 checked views stay unchanged.

The overload couples the original array's element type to the visitor input.
Proof follows unchanged array, element and callback aliases and known module
helpers, including overloaded helper signatures. Every invocation must take a
present original-array element. The explicit presence comparison must name that
same element and the real undefined symbol. Writes, callback storage and escape,
recursive or opaque helpers and untracked effects do not prove this input domain.
Default-created arrays retain their implementation element type, rather than
inheriting the caller's TIn. Consumers with captures remain an explicit stop.

A known concrete call specializes the consumer and its helper bodies with the
retained TIn input metadata. It passes the original callback unchanged. The
private invocation entry takes that closure first and evaluates its argument
once; it does not expose a wrapper callback. A proven invocation calls directly
without a check. An unproven TypeScript invocation validates every required flat
readonly scalar field of the concrete input view before invoking that closure.
The guard uses the existing checked-view and boxed dynamic-field paths; a tag
alone is insufficient. Union inputs test their concrete members. Complex views,
mutable/optional fields, accessors and unadapted closure union slots retain stops.
The failure exits 70 naming the visitor, implementation call site and both types.
Each emitted guard is counted in --explain-checks. Adamic sources refuse the
unproven original-array or fabricated-input path instead.

The original, helper and overloaded-helper valid pairs match Node and emit zero
checks. They print true, modifier, true, preserving visitor identity and original
array identity. Their fabricated-node TypeScript partners print undefined on
Node but stop before invoking the visitor in both compiler backends, under ASan
and UBSan and release builds; their Adamic partners refuse. A default-created
fabricated array forwarded through the overloaded helper has the same stop and
refusal. The successful unproven-source fixture instantiates Named and Numbered
inputs separately: Node and both backends print modifier and 7, and exactly two
checked implementation invocation sites are recorded. Successful native runs
also pass the leak check. Node observations are saved in node.json.

Independent result tests refuse an unproved readonly U[] result and reject a
callback's covariant readonly-array result when field storage would change.
The actual callback result is validated separately from the resolved overload's
wider callback promise. Identical outer pointers cannot authorize reboxing
fields or elements. Numeric-index element failures are diagnosed as result[]
with their own covariance direction. Node observations for these reductions are
in node-result-probes.json. No NodeArray<TOut> result proof is inferred from the
visitor input proof.

Both visitor mutants compile and run: erasing concrete TIn to Node, and clearing
the unproven-invocation finding, let a planted fabricated node through with
JavaScript exit 0. Each is caught by TestOverloadVisitors' expected exit 70,
not a compile error or a checked-count mismatch. See mutants.json and the two
individual logs. The local a-check records three proven .a fixtures and five
expected refusal pairs. counts.md has seven new rows; existing numeric counts
are unchanged. No backend emitter, runtime, protected orchestration file,
checker option or copied cohere source was changed.

The measurement-only replay verifies all 82 adapted hashes at source pin
388096e6, loads/registers the full compiler project and visits the seven assigned
intervals. It cannot emit IR. It compares against ../values/after.jsonl.gz.
An initial invocation mistakenly selected the adapted checkout root and reached
an unrelated Node.Text template panic in library discovery; rerunning with the
pinned src/compiler directory succeeded without changing source or checker
options. This is a boundary-intersection measurement on a checker-rejected
project, not whole-compiler emission.

Region 06's parameter refusal is passed, but the independent result relation now
stops its first boundary at esDecorators.ts:1250:13, reaching visitNodes at
visitorPublic.ts:194:1. The exact first failing relation is nullable-result
correlation: undefined must fit NodeArray<TOut> | (TInArray & undefined). Proving
that alone would still require the NodeArray<TOut> element result contract; a
Node result or original input proof does not establish arbitrary TOut. That
result proof/check is not built here. Additional revealed bytes in this group
are zero, with 3,369 still hidden in 06. Region 14's additional 1,195-byte reveal,
06's preceding 8,048 bytes and 13's 324 signature bytes remain. Total assigned
hidden bytes remain 48,591. Exact diagnostics and in-region versus dependency
locations are in next-stops.json.

| Region | Hidden after | Additional revealed in 06 group | Cumulative revealed | Next stop | Owner / raising function |
|---|---:|---:|---:|---|---|
| 01 large | 13,625 | 0 | 0 | declarations.ts:1678:110 TS2345 | checker: isSignatureApplicable |
| 01 small | 6,899 | 0 | 0 | declarations.ts:1678:110 TS2345 | checker: isSignatureApplicable |
| 05 large | 6,078 | 0 | 0 | visitorPublic.ts:123:5 visitNode parameter node, from es2018.ts:828:9 | lowering: censusOverload; checked views |
| 05 small | 5,748 | 0 | 0 | visitorPublic.ts:123:5 visitNode parameter node, from es2015.ts:3193:9 | lowering: censusOverload; checked views |
| 06 | 3,369 | 0 | 8,048 | visitorPublic.ts:194:1 visitNodes result covariance, from esDecorators.ts:1250:13 | lowering: censusOverload |
| 13 | 6,965 | 0 | 324 | es2017.ts:764:38 TS18048 | checker: checkNonNullTypeWithReporter |
| 14 | 5,907 | 0 | 1,195 | utilities.ts:5041:1 skipParentheses result._expressionBrand, from utilities.ts:11338:9 | lowering: censusOverload |


Commands (each saved directly to the accompanying log):

```sh
source /workspace/adamic-tools/env.sh
go test ./internal/lower -run 'TestOverload(Visitor|Value|Structural|Callback|Result)' -count=1 > /tmp/overload-visitors-final-lower.log 2>&1
# ok 2.947s
go test ./internal/oracle -run '^(TestOverloadVisitors|TestOverloadVisitorInstantiations|TestOverloadValues|TestOverloadValueProof|TestOverloadFieldHatch|TestPredicateDirectionCountsAreRecorded)$' -count=1 > /tmp/overload-visitors-final-oracle.log 2>&1
# ok 4.074s
go test ./cmd/adamic -run '^TestExplainOverload(Visitor|Value)Checks$' -count=1 > /tmp/overload-visitors-final-records.log 2>&1
# ok 0.642s
python3 docs/overload-results/groups/visitors/run-mutants.py > /tmp/overload-visitors-mutants.log 2>&1
# both caught; each mutant oracle exits 1
python3 docs/overload-results/groups/visitors/a-check.py /tmp/overload-hatch-fast-gate.py > /tmp/overload-visitors-acheck.log 2>&1
# PASS, three proven and five expected refusals
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -args -update-counts > /tmp/overload-visitors-counts.log 2>&1
# ok 43.696s, seven new rows
```

Tooling a-check pin is 89cbe74a. No whole package or full gate was run. Setup and
machine observations for this same environment are recorded in the preceding
values report: nproc 5, cgroup CPU quota 4; setup cumulative timings Go 0.021s,
Node 0.022s, submodules 0.069s, markdown 0.072s, clang 0.174s, build 34.857s,
deferred tests 35.087s, cache 35.089s, done 35.118s.
