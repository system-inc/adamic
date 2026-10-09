Step 30: overload specialization now reads every target through CallTargets or ClosureTargets.
Fix 573bd3e7 on compiler/hidden-boundaries c9beed41; no allowlist change.
Build, vet, target-reader guard, focused overload lowering/oracles and fresh closure-value sanitizer runs pass.
All 32 existing mutant witnesses pass, with the legacy visitor witness explicitly guarding its diagnostic path.
No new fixture, region replay, census change, full package sweep or full gate was performed.

Direct calls use CallTargets for specialization identity and parameter storage. Known closure calls use ClosureTargets, which now returns Value, the original closure operand, even for unknown target sets. Forwarding it preserves identity, captures and evaluation order. Lowering no longer reads Call.Function or CallClosure.Closure. The guard and allowlist remain unchanged. Unknown targets remain stopped. Conservative assumption: virtual target families are finalized after class lowering, so this boundary still requires one closed implementation known now. Existing result proofs, counted checks and .a refusals remain intact.

Final commands, each exit 0 (checks.json and compressed logs):
- go build ./...
- go vet ./internal/ir ./internal/lower
- go test ./internal/ir -run 'TestCallTarget|TestClosureTarget' -count=1 -v
- go test ./internal/lower -run '^TestOverload' -count=1 -v
- go test ./internal/oracle -run '^TestOverload|^TestNativeAgreesWithNode$/internal/oracle/testdata/(overload_results_|overload_callback_|overload_structural_|overload_values/)' -count=1 -timeout=20m -v
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestOverloadValues$|^TestOverloadValueProof$' -count=1 -timeout=20m -v
- python3 docs/overload-results/run-mutants.py
- python3 docs/overload-results/groups/<group>/run-mutants.py for callback, structural, fields, admission, field-hatch, values and visitors.

Fresh values checks cover returned closures, aliases, identity, captures, narrow single-signature promises, modules and evaluation order. Planted wrong results stop with exit 70; .a refusal pairs remain refused. The oracle uses source Node, backend JavaScript, native ASan/UBSan, release and leak checks. Fresh value runs have zero result-cache hits. Reader tests cover subclasses, unknown targets, literals, const bindings, sibling environments and exact original operand preservation. No new .a fixtures means no new counts row; existing focused oracle records remain checked.

Every mutant is named in witnesses.json and its original group's JSON/logs: core 5, callback 4, structural 4, field storage 7, admission 6, field hatch 2, escaping values 2 and visitors 2. The specialization mutation now uses the reader-derived function index.

The legacy admission trust-visitor-input marker was stale. Removing early callback contravariance reaches the later visitor argument path is unproven refusal; the test fails on the changed diagnostic, while later proof still prevents unsafe admission. A parent-overlay run on c9beed41 has the same result. Its harness now names that exact diagnostic-path witness rather than expecting got <nil>. This is not evidence of unchecked execution; the separate erase-TIn-to-Node and admit-unproven-invocation mutants prove actual visitor input checks. The parent negative run deliberately exits 1.

Initial concurrent cold Go builds exhausted the separate 8.8 GB /tmp filesystem, failing to write cohere checking/flow and rules/nexus outputs. Those build failures were excluded as mutant evidence. GOTMPDIR=/workspace/scratch/hidden-go-build provides workspace capacity; the guard, lowering, build and affected core/field groups were rerun successfully. The initial failing build log is retained alongside final logs.

The merge unit's setup used GOPROXY=https://proxy.golang.org|direct and nproc=5 (CPU quota 4): build 39.518s, total 39.742s. Complete timing lines are in the two preceding stack-member reports. No cohere source was copied.
