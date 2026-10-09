Built: fixed ordinary-return admission with result variance and matching storage, plus exact failed-relation diagnostics toward roadmap step 16.
Compiler: c09a96043e7ff87c434c4dcf0f078247b8f9dccd, on ba14c913 and the already merged compiler area baseline.
Checks: focused lowering and four Node/JavaScript/native/release/sanitizer fixtures passed; counts refreshed; seven-region measurement reveals zero additional bytes.
Mutants: six independent mutations failed their intended assertions with exit 1; none was caught by a build failure.
Uncovered: full 13/14 branch-dependent return proofs, visitNodes invocation/element proofs, checked views for 05 (★11), and checker-only 01.

The result relation is implementation-produced value to overload-promised result.
Readonly covariance is forward. It does not reverse merely because the narrow
record could be widened to the implementation's annotation. Writable fields
also require the reverse relation. The diagnostics now name those relations,
the actual types and direction, and the failing field path.

| Region | Overload result | Implementation result | Failed relation | Remaining proof |
|---|---|---|---|---|
| 13 | Block (FunctionBody) | ConciseBody = Block or Expression | censusRelated(ConciseBody, Block) is false; Expression.kind has SyntaxKind, which does not fit SyntaxKind.Block at result.kind | Under MethodDeclaration / AccessorDeclaration / FunctionDeclaration / FunctionExpression inputs, establish that isArrowFunction is false and every reachable assignment to result supplies a complete Block, including statements |
| 14 | EvaluatorResult<string or undefined> | EvaluatorResult<string or number or undefined> | censusRelated(wide result, narrow result) is false; readonly covariance requires string or number or undefined to fit string or undefined at result.value | Under TemplateExpression inputs, prove skipParentheses preserves the admitted domain and the reachable return comes from evaluateTemplateExpression or another string/undefined result, excluding numeric and arbitrary resolver returns |

Exact fresh diagnostics:

```text
transformers/es2017.ts:736:5: Adamic 0.1 refuses overload 1 of transformAsyncFunctionBody result Block cannot be served by implementation result ConciseBody at result.kind; readonly covariance requires SyntaxKind to fit SyntaxKind.Block; prove every return for this overload's admitted arguments, preserving result variance
utilities.ts:11320:5: Adamic 0.1 refuses overload 1 of evaluate result EvaluatorResult<string | undefined> cannot be served by implementation result EvaluatorResult<string | number | undefined> at result.value; readonly covariance requires string | number | undefined to fit string | undefined; prove every return for this overload's admitted arguments, preserving result variance
```

The first field failure in 13 is necessary, not sufficient: a matching Block
kind alone does not establish the required statements and complete Block shape.
The readonly value field fails in 14 because number remains in the implementation
annotation. The other three fields are boolean on both result instantiations;
their writable invariance does not remove the value failure. These annotation
failures do not establish that the original algorithms return wrong results.
They establish that ordinary variance alone cannot prove the narrower promises.

The existing return proof can now use a direct call to a closed ordinary named
function with a fixed nongeneric return contract. The callee remains subject to
ordinary body lowering; callbacks, host declarations, generic instantiations and
overloaded callees receive no shortcut. The source result must fit the overload's
promise through censusRelated. Call effects do not invalidate this proof because
it uses no argument or alias flow narrowing. Matching object-field storage is
also required: covariance does not rebox a Number field into a Union pointer
while preserving a shared object. Different reference payload types still need
a separate layout proof. Whole scalar conversions remain on the existing return
path. No object is copied, no wider annotation is trusted, and the original call
and its effects run once.

The factory fixture exercises a Block returned through a ConciseBody-like wider
annotation with a broad Expression.kind, and a shared Result<string> returned
through a wider result annotation to Result<string | undefined>. It prints:

```text
Block:1:true:1
word:true:1
false
```

That output pins identity, the reference payload, one invocation of each factory,
readonly covariance and a same-type writable flag shared with the original
object. Node, the JavaScript backend, native with ASan/UBSan, release native and
leak checks agree. Narrowing a wider ordinary factory result is refused, even
when the object fields share the same storage. A writable string field cannot
be widened to string | undefined on a shared result: the real alias-write
witness makes Node observe undefined through the original string-typed alias.
A logically covariant number-to-number/string record also stays refused when
its fields would need different storage.

For 06, the overloaded visitor parameter is Visitor<TIn, Node | undefined>;
the implementation parameter is Visitor, defaulting its input to Node. Treating
the supplied callback as the implementation callback requires the reverse input
relation censusRelated(Node, TIn). It fails for a rigid TIn that may be a proper
subtype of Node. The actual modifierVisitor at esDecorators.ts:504:5 takes
ModifierLike and returns VisitResult<Modifier | undefined>; it is not a visitor
accepting every Node.

```text
visitorPublic.ts:196:5: Adamic 0.1 refuses overload 1 of visitNodes parameter visitor cannot be served by implementation parameter visitor; callback contravariance at callback.parameter1 requires Node to fit TIn; make the implementation accept every value admitted by this overload, without mutable widening or bivariance
```

Serving this parameter needs a correlated invocation-domain proof: every visitor
call must consume an element from the admitted NodeArray<TIn>, including through
visitArrayWorker and its start/count handling. The current immediate-callback
mechanism requires a single concrete consumer contract served directly; it does
not specialize this generic forwarding body or establish that correlation.
The next result obligation would be NodeArray<TOut>: every element, including
flattened visitor arrays and reused originals, must pass the supplied predicate.
visitArrayWorker uses Debug.assertNode and Debug.assertEachNode, so a boundary
must prove or check the complete element contract while preserving array identity
and NodeArray metadata. That is not ordinary callback variance. It needs the
interprocedural domain proof and the tested element/view mechanism. visitNodes
therefore remains refused; its diagnostic now exposes the precise first relation.

A reduced correlated loop prints modifier and 1 under Node but remains refused
until that input correlation is proved. The separate visitor-domain liar passes
an arbitrary Node to a Modifier visitor; Node prints undefined and 1. Removing
the callback input refusal accepts that liar, independently demonstrating why
matching closure representations do not prove the invocation domain.

05 is unchanged and waits on checked views (★11), as instructed. 01 is unchanged
on checker TS2345 under stricter options (#k881crd). Hidden-13 also retains its
previously recorded checker-excluded implementation body.

| Region | Hidden before | Hidden after | Additional revealed | First remaining stop |
|---|---:|---:|---:|---|
| 01 large | 13,625 | 13,625 | 0 | declarations.ts:1387:5 excluded body; TS2345 at declarations.ts:1678:110, checker |
| 01 small | 6,899 | 6,899 | 0 | Same body and TS2345 at declarations.ts:1678:110, checker |
| 05 large | 6,078 | 6,078 | 0 | es2018.ts:828:9; visitorPublic.ts:123:5 visitNode parameter node, censusOverload; checked views ★11 |
| 05 small | 5,748 | 5,748 | 0 | es2015.ts:3193:9; same visitNode stop; checked views ★11 |
| 06 | 3,369 | 3,369 | 0 | esDecorators.ts:1250:13; visitorPublic.ts:196:5 callback input Node to TIn, censusOverload |
| 13 | 7,289 | 7,289 | 0 | es2017.ts:736:5; result.kind SyntaxKind to SyntaxKind.Block, censusOverload |
| 14 | 7,102 | 7,102 | 0 | evaluate body entry; utilities.ts:11320:5 result.value wide to narrow readonly value, censusOverload |

Total assigned hidden bytes stay at 50,110. The prior 8,048-byte reveal in 06 is
retained, not counted again. The [comparison](regions.json), [exact next stops](next-stops.json)
and [raw census](after.jsonl.gz) preserve the observations. The before census is
[the preceding group's after census](../fields/after.jsonl.gz). All 82 adapted
hashes match source pin 388096e6. The pinned union/subtraction algorithm and the
same independently overlapping declaration scope were used, with the full
project loaded, checker guards, rollback/continuation and the no-output assertion.
There is no whole-compiler emission or corpus-wide reveal claim.

Commands used, with output redirected directly to logs:

```sh
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/overload-admission-setup.log 2>&1
nproc
# 5

go test ./internal/lower -run '^(TestOverloadStructural|TestOverloadResults|TestOverloadCallback|TestCensusOverload)' -count=1 -v > /tmp/overload-admission-lower.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode/internal/oracle/testdata/overload_structural_' -count=1 -v > /tmp/overload-admission-oracle.log 2>&1
python3 docs/overload-results/groups/admission/run-mutants.py > /tmp/overload-admission-mutants.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/overload-admission-counts.log 2>&1

python3 stage3/census/latent/make_overlay.py /workspace/adamic /tmp/overload-admission-census-overlay > /tmp/overload-admission-overlay.log 2>&1
# Apply ../c68/census-scope.patch.gz to the scratch overlay.
go build -buildvcs=false -overlay=/tmp/overload-admission-census-overlay/overlay.json -o /tmp/overload-admission-census ./stage3/census/latent/tool > /tmp/overload-admission-census-build.log 2>&1
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /tmp/overload-admission-census /tmp/overload-results-adapted/src/compiler /tmp/overload-admission-after.jsonl > /tmp/overload-admission-census.log 2>&1
python3 docs/overload-results/measure-regions.py /workspace/overload-results-sourcepin /tmp/overload-results-adapted/src/compiler docs/overload-results/groups/fields/after.jsonl.gz /tmp/overload-admission-after.jsonl docs/overload-results/groups/admission/regions.json --all --compiler c09a96043e7ff87c434c4dcf0f078247b8f9dccd > /tmp/overload-admission-measure.log 2>&1
python3 docs/overload-results/groups/fields/record-stops.py /tmp/overload-admission-after.jsonl docs/overload-results/groups/admission/regions.json docs/overload-results/groups/admission/next-stops.json > /tmp/overload-admission-next-stops.log 2>&1
```

All final commands passed. The four positive fixtures ran uncached; all negative
witnesses run source Node before asserting the lowering refusal. Counts gained
one row: overload_structural_factory.a has allocations/frees 9/9, retains/releases
8/18, peak 5, regions 0. No existing numeric count changed. No whole package or
full gate was run. An initial drop-proof run failed the intended test but did not
match its evidence marker; the strengthened broad-kind Block witness now catches
it directly, and all six final mutations match their intended assertion.

| Mutant | Catcher | Failure |
|---|---|---|
| Drop fixed-return proof | TestOverloadStructuralResults | Broad-kind Block factory is refused |
| Reverse result covariance | factory_literal_covariance | Same-storage string factory liar accepted with nil error |
| Trust visitor input | visitor_domain_liar | Arbitrary Node invocation accepted with nil error |
| Ignore writable invariance | factory_writable_invariance | Alias-invalidating undefined write accepted with nil error |
| Drop return field storage | factory_field_storage | Unboxed Number viewed as Union field accepted with nil error |
| Drop relation diagnostic | factory_result_covariance | Refusal lacks the required relation and direction |

[mutants.json](mutants.json), run-mutants.py and their individual logs retain the
selectors and observations. The variance and storage mutants have independent
witnesses so neither check masks the other's mutation. Production sources were
never left mutated.

Setup cumulative timing lines: Node ready 0.022s, Go ready 0.023s, submodules
0.055s, markdown ready 0.066s, clang ready 0.188s, Go build ready 36.801s,
deferred test binaries 37.007s, cache warm 37.008s, done 37.042s. nproc is 5,
cgroup cpu.max is 400000 100000. The complete setup output is preserved here.
No checker options, runtime headers or protected emitter files changed; no
cohere code was copied and no unlanded worker branch was merged.
