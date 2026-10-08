V2 partial extraction toward roadmap step 11: complete mixed-union graph construction and independent selectors.
Commit: see the views-v2 commit containing this report.
Validation: go build ./... and go vet ./internal/... passed; IR, JavaScript and native selector controls passed; three lower contract controls failed.
Mutants: not run yet; no new source admission is enabled.
Not covered: source dispatch, object-plus-primitive and untagged object admissions, tuples, full sweep and counts.

The mixed lane net change is extracted from 67d34f3c. Its original commit list is in the commit message. The extraction preserves failed-builder rollback, missing-member refusals, logical kind checks, finite literal checks and reference-adapter requirements. Descriptors do not establish readiness or ownership.

Source admission remains closed because V1 has not installed union selection in its common read dispatch. The runtime slot snapshot declaration is a handoff, not a runtime implementation or a readiness exemption. Its storage normalization must be reconciled with the ruled representation tags before it can be used.

The source lane contract test includes a Node-or-readonly-Node-array case. V3 owns general array admission; this test remains present so that the dependency is visible. The checked-view family hooks for arrays, callables, dictionaries and intersections remain unavailable. Erasure returns no proof. Runtime graph admission is excluded.

No conflict judgment changes behavior in this extraction: shared backend dispatch and representation changes have not been applied. Node helper tests preserve the lane selector decisions; the full differential and per-check mutants still have to run.

Commands used GOFLAGS=-buildvcs=false for the scratch worktree submodule link. go test -count=1 ./internal/ir ./internal/javascript ./internal/native -run 'TestPrimitiveViewMembers|TestViewMixedUnionUnknownAndUnavailable' passed. go test -count=1 ./internal/lower -run TestMixedUnionContract failed three controls: TestMixedUnionContractGraph/Node-or-readonly-Node-array at view_unions_mixed_test.go:45 (array adapter unavailable), TestMixedUnionContractPhantomBrandUsesPrimitiveBase at line 140 (Target representation unavailable), and TestMixedUnionContractPhantomVoidIsUndefined at line 167 (intersection descriptor). The original assertions remain intact. Logs: /tmp/views-v2-build.log, /tmp/views-v2-vet.log, /tmp/views-v2-mixed-components.log, /tmp/views-v2-mixed-lower.log.
