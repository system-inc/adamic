Built: resolved overload boundary entries, checked results and nullable parameters, toward roadmap step 16.
Base: 4885cec50290686df487b62aac47c85d871ed40c; delivery SHA accompanies the pushed report.
Checks: five reduced Node fixtures, three pinned liar stops, affected overload regressions and refreshed counts pass.
Mutants: unchecked result, erased specialization, missing parameter check, missing freshness proof and missing required-field presence are caught.
Limits: this implements specialized boundary wrappers around a shared body; full tsc regions remain blocked.

The proposal is in ../0.1.md's overload section. An implementation's wider annotation never becomes a proof of the narrower result. A direct checker-resolved overload gets an entry keyed by implementation, overload ordinal, concrete result and parameter contracts, and argument representations. It evaluates the original call once, keeps the wider result until proven or checked, then returns the promised representation. Reusing an entry still converts each argument. Existing nullable-result and predicate-overload lowering remains in use.

The boundary can check primitives and literals, an already proven tagged-union member, and fresh scalar records with compatible field storage. Freshness comes only from direct literals or closed scalar factories; shared writable and readonly records remain refused. Deferred generic tagged-node parameters are judged against the concrete resolved signature. A nullable parameter excluded by the implementation stops before that implementation runs. Pinned failures exit 70 and name the overload and both types.

Unreifiable result contracts, indirect function values, callbacks, spread/rest calls requiring the new boundary, differing arity and void boundaries remain refused or NotYet. Callback variance, arbitrary structural casts and mutable parameter widening are not accepted. These are conservative limits, not additional trust in overload declarations. The new mechanism does not clone or rewrite an implementation body per overload.

The transform, binding and nullable visitNode witnesses reduce the reported tsc seams to independently executable programs. The evaluate witness has mutable EvaluatorResult fields and a fresh generic factory, including a write through the narrowed result. It explicitly instantiates the factory with its implementation's wide scalar union. An implicit narrower mutable factory result still meets the existing invariance refusal before this boundary; this unit does not resolve that independent factory-return problem. Full compiler Block/Expression contracts and callable visitor-domain specialization remain unsupported. Hidden-13 and hidden-14 must replay against the delivery SHA; their region bytes are not claimed here.

The focused commands were run with output redirected directly to log files:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/overload-results-area-setup.log 2>&1
source /workspace/adamic-tools/env.sh
npm ci --prefix stage3/api --ignore-scripts > /tmp/overload-results-api-install.log 2>&1
go test ./internal/lower -run 'TestOverloadResults|TestCensusOverload|TestPredicateOverload' -count=1 > /tmp/overload-results-fixtures.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/(overload_results|census_overload_contracts)' -count=1 -v > /tmp/overload-results-oracle-positive.log 2>&1
go test ./internal/oracle -run '^(TestCensusAppendResultProof|TestCensusOverloadResultStop|TestPredicateOverloadRuntime)$' -count=1 -v > /tmp/overload-results-oracle-regression.log 2>&1
python3 docs/overload-results/run-mutants.py > /tmp/overload-results-mutants.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/overload-results-counts-final.log 2>&1
```

TestPredicateOverloadRuntime belongs to internal/lower and executes in the first selector; the oracle regression selector executes the two census tests. The oracle fixtures compare source Node, JavaScript, sanitized native, release native and leak checks. The liar tests independently pin Node's continuation and the boundary's stop in JavaScript, release native and ASan/UBSan native. No whole-package tests or full gate ran. The first counts attempt exposed a rest-overload selection regression and missing pinned Node declarations; both were fixed, and the final refresh passed.

Setup reports Go 0.024s, Node 0.025s, submodules 0.069s, markdown 0.082s, clang 0.166s, Go build 47.079s, cache 47.260s, done 47.284s. nproc is 5; cgroup cpu.max is 400000 100000. The toolchain is Go 1.27.1, Node 24.19.0 and clang 20.1.8. The first branch checkout needed an explicit fetch because the remote fetch refspec only covered main; the reported setup ran on the correct area base.

| Mutant | Catcher | Observed failure |
|---|---|---|
| Treat an optional source field as required | Required-property refusal witness | A fresh optional record is accepted where the overload promises own-field presence |
| Trust the implementation result unchecked | TestOverloadResultsLiarStops | String liar prints number and fresh-record liar prints 42, finishing instead of exit 70 |
| Erase concrete specialization from entry keys | Binding fixture held to Node | Second call reuses the BindingElement contract for OmittedExpression; both backends exit 70 where Node prints true |
| Drop the nullable parameter check | Parameter liar pinned stop | Implementation runs and prints undefined, finishing instead of stopping before entry |
| Drop result freshness requirement | Shared writable and readonly refusal witnesses | Shared mutable views are accepted with nil error |

The runner uses scratch Go overlays and never mutates production files. Its exact selectors and exit codes are in evidence/mutants.json. All five mutations fail the intended assertions, without Go compilation failures.

Counts add the five new positive fixtures. Allocations equal frees for each; all finish without leaks. The generated table also moves the existing logical_and_reference_maybe row without changing its numbers and removes a stale taste/17_binder_flow row: that fixture is already registered as lowers=false on this area base. No existing numeric count changed.

The census loads and registers the whole pinned adapted project while independently attempting every declaration in transformers/declarations.ts, transformers/es2018.ts and transformers/esDecorators.ts. All 82 adapted source hashes are verified against the stock manifest at 388096e6a83a4e9d287fb827f793c599ba1bf0ad. The pinned hidden.py union/subtraction calculation is intersected with each assigned UTF-8 half-open range. This is measurement-only output from a checker-rejected program, not usable IR and not a whole-corpus reveal claim. The scope overlays and compressed records are preserved under evidence/.

The delivery carries only this unit's changes on the requested area tip. Hidden-01's reference 333354b3 and hidden-06's 702c3ecb were inspected, never merged. A separate scratch-only hidden-06 overlay adds its constraint-storage expression change on both measurement sides so its already-revealed bytes can be distinguished from this unit's contribution. No cohere implementation was copied. Native and JavaScript emission use existing IR operations; protected emitter files are unchanged.

| Region | Assigned range | Hidden before | Hidden after | Newly revealed |
|---|---|---:|---:|---:|
| hidden-01 | declarations.ts [68180, 81805) | 13,625 | 13,625 | 0 |
| hidden-05 | es2018.ts [35690, 41768) | 6,078 | 6,078 | 0 |
| hidden-06, area base | esDecorators.ts [61324, 72741) | 11,417 | 11,417 | 0 |
| hidden-06, with 702c3ecb constraint storage | same range | 3,369 | 3,369 | 0 |

Hidden-06's dependency-aware replay reproduces the visitNodes boundary at esDecorators.ts:1250:13, [61665, 61755), refusing visitorPublic.ts:196:5's callback parameter. Its worker's prior 8,048-byte reveal remains intact and is not attributed to this unit. Hidden-01 retains callable visitor-domain stops, including visitNodes. Hidden-05 retains the enclosing transformFunctionBody Block-versus-ConciseBody refusal, alongside other visitor-contract stops. Reduced literal-tag records are reifiable here; the full compiler contracts are not. This is a partial step-16 mechanism, not completion of those hidden boundaries.

The replay commands are:

```sh
# The scoped measurement overlays are generated by stage3/census/latent/make_overlay.py.
# before.json substitutes the area-base census_small.go and expression.go.
# hidden06-{before,after}.json additionally applies hidden06-dependency.patch in scratch.
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /tmp/overload-results-before-census /tmp/overload-results-adapted/src/compiler /tmp/overload-results-before.jsonl > /tmp/overload-results-before-census.log 2>&1
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /tmp/overload-results-after-census /tmp/overload-results-adapted/src/compiler /tmp/overload-results-after.jsonl > /tmp/overload-results-after-census.log 2>&1
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /tmp/overload-results-hidden06-before-census /tmp/overload-results-adapted/src/compiler /tmp/overload-results-hidden06-before.jsonl > /tmp/overload-results-hidden06-before-census.log 2>&1
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /tmp/overload-results-hidden06-after-census /tmp/overload-results-adapted/src/compiler /tmp/overload-results-hidden06-after.jsonl > /tmp/overload-results-hidden06-after-census.log 2>&1
python3 docs/overload-results/measure-regions.py /workspace/overload-results-sourcepin /tmp/overload-results-adapted/src/compiler /tmp/overload-results-before.jsonl /tmp/overload-results-after.jsonl docs/overload-results/evidence/regions.json > /tmp/overload-results-regions.log 2>&1
```

The hidden-06 dependency comparison uses the same measurement script on its two corresponding jsonl records. Absolute scratch paths in overlay JSON record the actual run; regenerate them using make_overlay.py, the named baseline files and the decompressed evidence/census-scope.patch.gz and evidence/hidden06-dependency.patch.gz to repeat on another checkout.

The environment restart interrupted the final census; both after replays were resumed and completed with exit 0. The preserved final records each contain the header and all three file records.
