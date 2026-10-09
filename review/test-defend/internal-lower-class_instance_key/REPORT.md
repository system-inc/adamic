# Defense: definite assignment neighbors and generic empty returns

Base: 7b9d4272c28f59530ab13daa5c49067e47933b06. Fresh discovery finds 276 top-level tests; both assigned names and their subsumers still exist. The audit reported 239 top-level tests at its older base. Every mutation ran the full current package.

## Code under test and oracle, fixed before mutants

TestDefiniteAssignmentSoundNeighbors reaches Lower, public field initialization and constructor writes, local declaration/assignment, and suppression-directive handling. Its current oracle is external-run: lowersAndAgreesWithNode compares source Node with generated JavaScript on Node, including stdout, exit and relevant stderr. The audit described only absence-of-error assertions; that description is stale. Its private-field subsumer also now uses Node agreement. Neither selected row is a cost test or an executor twin.

TestEmptyLiteralGenericReturnUsesSamePath reaches generic function instantiation, concrete type substitution, cast lowering and contextual empty-array representation in elementType/literalArrayElement. Its oracle is self: exactly one Number and one Object element representation across two generic instantiations. Its subsumer checks nested nongeneric empty literals rather than generic function returns. No test, harness, external checker or oracle was mutated.

## Coverage evidence

Each assigned row and each subsumer passed a separate go test -json -count=1 -timeout 90s -run '^Name$' -coverpkg ./internal/lower -coverprofile run. Exact profiles, logs and exclusive block lists are beside this report. SoundNeighbors has 92 covered blocks absent from PrivateRepair, including initializer evaluation and locals. GenericReturn has 511 absent from NestedEmpty, including generic substitution and function instantiation. Generic empty-array helper lines themselves are shared; a live generic mapper is the semantic difference used for D1. Coverage includes only instrumented Go, not executed JavaScript or generated C.

## Mutations and observations

D1 returns early from literalArrayElement when a generic mapper is active. This removes contextual representation for generic empty returns without changing nested nongeneric literal handling. Only TestEmptyLiteralGenericReturnUsesSamePath fails, at empty_literal_test.go:52, with 'stage 0 cannot lower an array of never yet'. All 273 other executing top-level tests pass, including TestNestedEmptyArrayElementKinds. The complete passed list is in matrix.json. Verdict: defended among executed rows.

D2 drops field-initializer expression evaluation; the preexisting zero value remains. SoundNeighbors fails on its public initialized number field: generated stdout is '1\n' while source Node prints '3\n'. Seven other rows catch it. PrivateRepair passes because its private field is assigned in the constructor rather than initialized by a declaration.

D3 drops the emitted ir.Assign from local assignments, preserving preparation statements. SoundNeighbors fails on its declared local followed by assignment: generated stdout is '1\n' instead of '3\n'. Thirteen other rows catch it. PrivateRepair passes. This is a different behavior from its original audit's constructor escape mutation.

D4 changes useOfThis's permitted declaration-kind constant from PropertyDeclaration to Parameter. SoundNeighbors' constructor write is refused, and eleven other rows fail, including PrivateRepair. This third attempt attacks the constructor-write permission shared with the historical subsumer. Verdict after three observed catches: not defended. This finite matrix does not justify deletion.

All standalone diffs are against the base commit. Each passed go vet ./internal/lower and a restored git apply --check. Matrix commands use timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . and a separate ADAMIC_BUILD_CACHE_DIR=/tmp/defend-lower-class-key/cache/Dn. Full commands, passed/failed/skipped lists and per-row diagnostics are in matrix.json. No run cooked, panicked or left an undiscovered top-level result unknown. Production sources were restored after each mutation.

## Friction, limits and owner findings

The /tmp filesystem has only 8.8 GB total, so 15 GB free is impossible. It had 6.3 GB free. The previous unit's /tmp/defend-lower-non-null scratch/cache directory was deleted as requested and disk space checked again. /workspace reported 15 GB free. No disk failure occurred.

The current tests and their oracle differ from the historical audit. We read the old report and plan, preserved those and its other code/mutation notes, then used current bodies and current full-package discovery. The old assertion that SoundNeighbors only checks acceptance is no longer true. Its name promises sound accepted neighbors, and its current assertions compare their behavior. There is no unasserted speed threshold or claim to flag for its owner.

TestOriginalCycleLedger and TestOptionalWideningCensus skip by default; the unsupported MixedUnionContractGraph interface/readonly-array subcase also skips. Their mutation catches are unknown. Full default-package execution is not a proof about opt-in project inventories or other packages. The unique defense is qualified to executed rows.

The menu's return-early operation requires an inserted guard for D1; the mutant is the early return from production code under an actual generic context, not a supplemental harness statement or a test-name selector. D2 leaves the existing error check after dropping evaluation; it compiles and its value defaults to zero. D3 drops assignment emission at the return while retaining computed preparation statements, avoiding unused locals. No empty-answer probe or oracle weakening contributes to a defense.

The full-package matrices took roughly 42 seconds each, so only four mutants were needed: one unique generic defense and three honest neighbor attempts. Native build time is included in matrix wall time, not separately instrumented. No other Go package was tested. No test was deleted, rewritten or weakened; main and pull requests were untouched.

## Timing

Warm setup: 0s, Go 1.27.1, nproc 5. npm ci reported 391ms. Clean baseline: 38.005 test-binary seconds. Coverage binaries: SoundNeighbors 1.269s, PrivateRepair 0.403s, GenericReturn 0.063s, NestedEmpty 0.159s.
Mutant wall times: D1 44.082s, D2 43.14s, D3 42.281s, D4 41.977s. Separate vet durations were not instrumented.

Restored full-package control: PASS, 33.762 test-binary seconds.
