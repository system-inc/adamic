# Predicate stack on views integration

Built all six own commits on `432d4913d31daaa49d8ca4eb46f30f21f90da5ee`, including dfced069's views admission.
Linear stack: c4f4b337, b5039caf, 30afab46, edfbb233, 1b4259af, 4b2b2172; no merge commits above the base.
Build and focused lower, Node/backend, and own count checks passed; 120/122 refusal roots admitted.
Ten effective mutants failed their targeted tests; three old single-path mutants survived alternate verification and were replaced by combined/admission mutants.
Limits: two writable emitNode refusals remain; shared counts refresh failed on unrelated fixtures and missing Node type dependencies.

## Stack resolutions

All six own non-merge commits were cherry-picked. Neither the source views merge nor the non-null merge was imported. Foundation resolutions preserve the views base's checks in `internal/lower/expression.go`, `internal/lower/predicates.go`, and `internal/lower/refusals.go`; the last file has no final diff. The own new `internal/lower/predicates_contracts.go` uses the area's flow-proof API and `closedPredicateRefusal`, preserving its independent `provePredicate` summary API. Fixture-location conflicts retain the test-owned `internal/oracle/testdata/predicate_hatches/` directory. `internal/oracle/counts.md` retains existing base rows and places own runtime rows before the predicate-direction section.

The views rule uses the base's `structuralViewCast`, `structuralViewIntersection`, and lazy `view` admission. The base has no `checkedAssertionSource`; its equivalent extension test is local to `internal/lower/predicates_views.go`. No non-null implementation, IR, backend, or runtime change was added. There are no deferred symbol dependencies. The stack is prepared on the requested integration pin; a future area's additional changes have not been replayed or claimed conflict-free.

## Tests and observations

`export GOPROXY='https://proxy.golang.org|direct'`, `bash cloud/setup.sh`, then `source /workspace/adamic-tools/env.sh`.

Setup passed: node 0.047s, go 0.053s, submodules 0.114s, markdown ready 0.132s (step 0.014s), clang 0.319s, done 452.986s. `nproc` is 5, CPU quota is 4. Reference cloning failed because the existing repository is shallow; a local clone also stalled on promisor objects. Exact pinned submodules were initialized with Git object alternates and checked out through Git, without copying source. `absorbgitdirs` encountered a cross-device rename for cohere; its standalone Git directory works and setup subsequently passed. Logs preserve these observations where available.

Final restored-source commands, all redirected to the adjacent compressed logs:

```sh
go test ./internal/lower -run '^(TestPredicateContractRefusals|TestPredicateBodyProof|TestPredicateCallbackContracts|TestPredicateOverloadCallback|TestPredicateUseRegions|TestPredicateUsesBelongToEachCall)$' -count=1 -v
go test ./internal/oracle -run '^(TestPredicateCheckedNarrowing|TestPredicateOpenContractsStayRefused|TestPredicateStructuralViews|TestPredicateWritableViewStaysRefused)$' -count=1 -v
go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/predicate_(callback_contract|helper_return).a$' -count=1 -v
go test ./internal/oracle -run '^(TestPredicateDirectionCountsAreRecorded|TestPredicateStructuralViews)$' -count=1 -v
PREDICATE_CENSUS_ROOT=/tmp/notyet-predicates-adapted PREDICATE_CENSUS_OUTPUT=/tmp/predicate-views-stack-coverage.csv go test ./internal/lower -run '^TestPredicateViewCensus$' -count=1 -v
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -args -update-counts
```

Focused lower: PASS 1.319s. Restored fixtures: PASS 3.775s. Both positive Node/backend fixtures: PASS 0.197s. Own mixed-mode runtime counts and existing predicate-direction section: PASS 3.096s. Valid values agree with Node in JavaScript, release native, and ASan/UBSan native; invalid structural values print 1 and stop with exit 70 at label. Both language boundaries and writable-view refusal are asserted. A temporary own-fixture count probe passed (0.429s), measured the two foundation rows in the base's eight-column format, and was removed. The ten mixed-mode rows were measured and matched by their owning fixture tests. No whole package or full gate was run.

The first lower invocation produced only `FAIL`, without a diagnostic. Its log is retained; no cause is inferred. Both subsequent focused lower runs passed.

The shared counts refresh failed in 61.788s. Failures include missing `@types/node 25.3.3` for FS fixtures and unsupported `process.exit` as a value in process fixtures. It did not write counts.md. Own runtime rows are retained, measured, and correctly placed so the existing predicate-direction table test passes. This report does not claim that the shared counts refresh passed.

The unchanged refusal input audit matched all 122 roots, admitting 120 (80 isExpression callbacks, 19 isStatement callbacks, 21 returns). Remaining transformers/es2015.ts:1349:9 and :1354:9 require broad writable emitNode admission from a readonly refined slot; the existing restriction stays. This audits refusal roots, not full native compilation of TypeScript. Historical source reports are preserved; this directory contains the new stack's own observations.

## Mutants

Each runner restores its sources in finally blocks. All final compiler sources were restored before the final tests.

| Effective mutant | Targeted fixture | Observation |
|---|---|---|
| Disable helper proof in both flow and summary verifiers | predicate_helper_return.a | Lower refuses the wrapper return at line 8 |
| Disable predicate callback admission | predicate_callback_contract.a | Lower refuses callback predicate parameter |
| Disable predicate callback admission for .ts hatch | CheckedNarrowing/callback | Lower refuses callback predicate parameter |
| Skip callback parameter write detection | ContractRefusals/reassigned_callback_parameter | Expected refusal becomes nil |
| Bypass closed hatch .a gate | CheckedNarrowing | Expected .a refusal becomes nil |
| Skip Expression closed narrowing check | CheckedNarrowing/true | Expected narrowing panic is absent |
| Skip Statement closed narrowing check | CheckedNarrowing/false | Expected narrowing panic is absent |
| Skip structural family registration and narrowed-read admission | StructuralViews/.*invalid$ | JavaScript prints 1 and 42, exits 0; ASan also detects invalid reads |
| Bypass structural hatch .a gate | StructuralViews | Expected predicate refusal becomes nil |
| Remove writable restriction from eligibility and structural admission | WritableViewStaysRefused | .ts compiles instead of refusing |

All ten effective mutants returned test exit 1, caught by the expected proof/refusal or runtime assertion rather than a build diagnostic. The initial helper-proof, callback-proof, and callback-hatch single-path mutants returned 0: the base has alternate proof/admission paths. Their survival is recorded, not counted as a killed mutant. The adapted proof runner checks the full helper proving rule and the callback admission boundary. No backend/runtime mutant was needed because no backend/runtime change was authored on this stack.
