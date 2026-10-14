Built: matching field storage for proved overload results, and constraint-backed TNode parameters toward roadmap step 16.
Compiler: dacbdb8cb2ce9033552aa5246a187b666b5cc395, on the already merged compiler/area-next-fixtures f0c6e6fc baseline.
Checks: focused lowering, Node/JavaScript/native/release/sanitizer oracles, dedicated Refused oracle and refreshed counts passed; hidden-06 reveals 8,048 bytes.
Mutants: seven independent mutations failed their intended assertions; every run exited 1, without a build failure.
Uncovered: 01 stays on the checker; 05 needs checked views; the full 13 and 14 result bodies remain unproved and refused.

The paths result.kind and result.value in the full compiler census are overload
admission diagnostics from censusOverload. They are not observed property-read
failures. Field storage cannot prove that a wider implementation result serves a
narrower overload. The full transformAsyncFunctionBody and evaluate bodies still
fail that earlier proof. Hidden-13's implementation body also has the previously
recorded stricter-options checker exclusions.

A proved result already uses the ordinary property read for fields with a single
representation. The new guard checks a union of result objects before reading:
every member's field must use the combined field's storage representation.
Direct overload calls and immutable result aliases are recognized. A mixed
unboxed number/string field cannot be read as a combined Union pointer. Its
refusal names result.value and asks for narrowing to a member. This guard does
not copy results, add checked views, or erase the return proof. The new positive
fixture checks literal kind reads, string/undefined value reads, a result alias,
a direct-call read, and identity. The existing Block and EvaluatorResult fixtures
remain held to Node in both backends.

Hidden-06's rule from 702c3ecb already applies to parameters: signature declares
parameter locals through typeOf, which uses representation. Its storage rule and
focused witnesses were reused without merging the worker's branch. The fallback
lives in constraintStorage, called only after concrete checker mappings and class
substitutions. Scalar constraints retain scalar storage. Structural object
constraints retain Union storage and runtime brands; their structural subtypes
can be arrays or functions, so Object layout cannot be assumed. Unconstrained,
any, unknown and unmapped collection layouts remain unsupported. Unsafe writes
through this constrained storage and generic function values remain NotYet.
The representation test now also checks each actual parameter node through
typeOf, rather than testing only the type-parameter declaration.

The visitNode stop needs three proofs. TIn admits undefined although the
implementation parameter is declared Node: its initial undefined return must
dominate all uses and preserve the overload's missing-value correlations. The
implementation calls a visitor whose admitted input can be narrower than Node:
every invocation must be proved to stay within NonNullable<TIn>. Finally, both
assigned callers pass isExpression, so the returned TOut must satisfy the whole
Expression contract, not just a matching kind. Debug.assertNode's test and the
visit-result array/lift path need a checked view of that returned value, including
reference fields. The current overload mechanism cannot supply that view.
Following the instruction to stop if checked views are needed, visitNode was
left refused. Constraint storage is not a parameter or result contract proof.

| Region | Hidden before | Hidden after | Revealed | First remaining stop |
|---|---:|---:|---:|---|
| 01 large | 13,625 | 13,625 | 0 | declarations.ts:1387:5 body excluded by TS2345 at declarations.ts:1678:110; checker, #k881crd |
| 01 small | 6,899 | 6,899 | 0 | Same excluded body and TS2345 at declarations.ts:1678:110; checker, #k881crd |
| 05 large | 6,078 | 6,078 | 0 | es2018.ts:828:9; visitorPublic.ts:123:5 visitNode parameter node refusal, censusOverload |
| 05 small | 5,748 | 5,748 | 0 | es2015.ts:3193:9; same visitNode refusal, censusOverload |
| 06 | 11,417 | 3,369 | 8,048 | esDecorators.ts:1250:13; visitorPublic.ts:196:5 visitNodes parameter visitor refusal, censusOverload |
| 13 | 7,289 | 7,289 | 0 | es2017.ts:736:5 Block through ConciseBody result refusal at result.kind, censusOverload |
| 14 | 7,102 | 7,102 | 0 | evaluate body entry; utilities.ts:11320:5 EvaluatorResult result refusal at result.value, censusOverload |

The assigned intersections total 58,158 -> 50,110 hidden bytes. All 82 adapted
source hashes match pin 388096e6. The pinned hidden.py union/subtraction algorithm
was applied to the same seven scoped intervals as the previous group. The full
project remains loaded and registered; all independent declarations overlapping
these intervals are attempted. Checker guards, rollback/continuation and the
no-output assertion remain enabled. This is a region measurement, not a claim
that the full compiler can emit. The before census is
[the preceding group's after census](../structural/after.jsonl.gz); the
[raw after census](after.jsonl.gz), [comparison](regions.json) and
[first stops with exact diagnostics](next-stops.json) are preserved here.

Commands actually used, with output redirected directly to logs:

```sh
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/overload-fields-setup.log 2>&1
nproc
# 5

go test ./internal/lower -run '^(TestOverloadStructural|TestHiddenTNode|TestOverloadResults|TestOverloadCallback)' -count=1 -v > /tmp/overload-fields-lower.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode/internal/oracle/testdata/(overload_structural_|hidden_boundary_generic_tnode)' -count=1 -v > /tmp/overload-fields-oracle.log 2>&1
go test ./internal/oracle -run '^TestOverloadStructuralFieldRefusal$' -count=1 -v > /tmp/overload-fields-refusal-oracle.log 2>&1
python3 docs/overload-results/groups/fields/run-mutants.py > /tmp/overload-fields-mutants.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/overload-fields-counts.log 2>&1

python3 stage3/census/latent/make_overlay.py /workspace/adamic /tmp/overload-fields-census-overlay > /tmp/overload-fields-overlay.log 2>&1
# Apply ../c68/census-scope.patch.gz to the generated scratch overlay.
go build -buildvcs=false -overlay=/tmp/overload-fields-census-overlay/overlay.json -o /tmp/overload-fields-census ./stage3/census/latent/tool > /tmp/overload-fields-census-build.log 2>&1
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /tmp/overload-fields-census /tmp/overload-results-adapted/src/compiler /tmp/overload-fields-after.jsonl > /tmp/overload-fields-census.log 2>&1
python3 docs/overload-results/measure-regions.py /workspace/overload-results-sourcepin /tmp/overload-results-adapted/src/compiler docs/overload-results/groups/structural/after.jsonl.gz /tmp/overload-fields-after.jsonl docs/overload-results/groups/fields/regions.json --all --compiler dacbdb8cb2ce9033552aa5246a187b666b5cc395 > /tmp/overload-fields-measure.log 2>&1
python3 docs/overload-results/groups/fields/record-stops.py /tmp/overload-fields-after.jsonl docs/overload-results/groups/fields/regions.json docs/overload-results/groups/fields/next-stops.json > /tmp/overload-fields-next-stops.log 2>&1
```

The positive oracle ran five emitting fixtures and two existing NotYet witnesses;
the five emitted programs matched Node through the JavaScript backend, ASan/UBSan
native, release native and leak checks. The dedicated Refused witness prints
7 under Node and stops before either backend receives IR. The generic negative
fixture runner accepts only NotYet, so this Refused case has its own test.
An initial combined test selector selected no fixtures; it was replaced with the
focused selector above, whose log records all seven fixture runs. Initial mutant
evidence matching looked for the wrong marker on the missing-constraint failure;
the assertion had failed as intended, and the corrected complete run records all
seven caught mutations. No whole package or full gate was run.

| Mutant | Catcher | Observed failure |
|---|---|---|
| Drop matching field storage | TestOverloadStructuralRefuses/mixed_field_storage | Refused replaced by nil error; Node witness still prints 7 |
| Drop constraint storage | TestHiddenTNodeConstraintRepresentation | Supported constraints no longer have storage |
| Use Object layout for object constraints | TestHiddenTNodeConstraintMutationRemainsNotYet | Unsafe field mutation accepted with nil error |
| Admit unmapped array layout | TestHiddenTNodeConstraintRepresentation/array_layout | Unsupported constraint accepted |
| Admit unconstrained/unknown/any layout | TestHiddenTNodeConstraintRepresentation | Unsupported constraints accepted |
| Ignore concrete substitution | TestHiddenTNodeConstraintRepresentation | Concrete Array substitution lost |
| Ignore checker mapper | TestHiddenTNodeConstraintRepresentation | Concrete Number mapping lost |

[mutants.json](mutants.json), their individual logs and run-mutants.py retain the
selectors and observed exits. Production files were never left mutated.

Counts gained exactly three rows; no existing numeric count changed:
hidden_boundary_generic_tnode.a has allocations/frees 2/2;
hidden_boundary_generic_tnode_constraints.a has 8/8;
overload_structural_fields.a has 4/4. All allocated values are freed.

Setup cumulative timing lines: Go ready 0.022s, Node ready 0.027s, submodules
0.069s, markdown ready 0.077s, clang ready 0.195s, Go build ready 40.428s,
deferred test binaries 40.638s, cache warm 40.640s, done 40.669s. nproc is 5,
with cgroup cpu.max 400000 100000. Go 1.27.1, clang 20.1.8, Node 24.19.0.
The complete output is in setup.log.gz. No checker option or runtime header was
changed, no cohere code was copied, and no unlanded worker branch was merged.
