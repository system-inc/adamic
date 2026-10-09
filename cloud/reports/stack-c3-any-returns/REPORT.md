Merged c79c7572 while preserving ordinary generic mapper identity and overload specialization.
Merge f94726a6; stack-owned repair 3d1095b3 cherry-picked as 33c5e37d.
Build, vet, scoped a-check (2 files), stage3, focused Node/native/JavaScript checks and regenerated counts all pass.
All nine branch mutants are caught; the stack repair's optional presence/copy-state mutants remain caught.
No full gate, census replay, new roadmap mechanism or unrelated package sweep was run.

Conflict: internal/lower/generic.go. Keep inferTypes for implementation overload binders and the wider implementation result's boundary checks. Ordinary calls keep the exact checker-owned mapper, resolved result and class specialization. Removing overload inference fails the Node-held sort-and-deduplicate fixture.

The exact stack repair was cherry-picked after the merge, rather than duplicated. It updates both packed-cache presence stamps, both optional-field tests' runtime-API use, and the two stale real fixture records. The cherry-pick had no conflict. The discarded e3380f23 merge never entered history; this branch is a fast-forward of its original tip.

Commands and exits are in any-checks.json and the compressed logs: go build ./...; go vet ./internal/...; go test ./stage3/fixtures -count=1; focused internal/lower Generic, ClassGeneric, Nominal, CensusOverload and Overload tests; internal/load NodeLibrary and NodeNamespace tests; TestCheckerMapperAliasKeepsIdentity; TestNativeAgreesWithNode for any_returns_concrete, node_namespace_timeout, census_overload_contracts, census_append_overload, class_inheritance_generic and generic_instance_key_. The oracle checks source Node, both backends, ASan/UBSan and release. Scoped a-check compares with c79c7572. Counts: GOMEMLIMIT=2GiB go test ./internal/oracle -run TestCountsAreRecorded -parallel=2 -count=1 -args -update-counts. Actual counts exit 0, 294.274s.

Counts were generated, never hand-merged. One existing row changes: hidden_boundary_generic_tnode_constraints.a retains 6 to 4 and releases 14 to 12, with allocations/frees 8/8 and peak 5 unchanged. Observed change follows ordinary-call resolved-result specialization; fewer ownership operations are the inferred explanation. Added against c79: any_returns_concrete.a and node_namespace_timeout.a. No rows removed or reordered. See any-count-changes.json.

Mutants and witnesses:
unsubstituted-return: caught by TestGenericResolvedReturnContracts
trust-any: caught by TestGenericResolvedReturnContracts
opaque-mapper: caught by TestGenericMapperKeepsCheckerIdentity
copier-loses-alias: caught by TestCheckerMapperAliasKeepsIdentity
ignore-namespace: caught by TestNodeNamespaceAnnotationUsesPinnedTimeout
ignore-unresolved-alias: caught by TestNodeNamespaceAnnotationUsesPinnedTimeout
replace-local-namespace: caught by TestNodeNamespaceLocalDeclarationIsNotReplaced
accept-unresolved-type: caught by TestNodeNamespaceUnresolvedTypesStayRejected
drop-overload-inference: caught by the Node-held sort-and-deduplicate oracle

Repaired optional tests: TestOptionalFieldWriteCatchesDroppedSlot, TestOptionalFieldCopyState and TestOptionalFieldPresenceCatchesMutants. Checked-copy mutants lose presence, readiness or representation, or overlap presence and readiness, and are caught by checked reads in sanitized and ordinary builds. See presence-api-tests.log.gz.

Setup used GOPROXY=https://proxy.golang.org|direct. nproc=5 (CPU quota 4). Timings: node 0.024s, go 0.022s, markdown 0.077s, submodules 0.091s, clang 0.200s, build 39.518s, deferred 39.712s, cache 39.713s, done 39.742s. Final actual-tree checks replace earlier repair-overlay previews. Prior setup/experimental runs and an OOM-killed overlapping counts attempt were not counted as passes.
