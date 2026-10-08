# @typescript-eslint/no-array-delete

Typed `DeleteExpression` listener using the shared checker. Reporting takes the handed node; no root node refetch. Messages and the three-edit splice suggestion are preserved.

All 44 original upstream cases retain their Go finding counts and match Go, Node, emitted JavaScript and ASan/UBSan native. The owned witness fires. `wave14-no_array_delete` reverses the array guard, compiles, runs and is caught by findings comparison. All 94 registered mutants are caught.

The complete lint package passes: 42 pass, 0 fail, 1 skip, including the extra scratch-only original-upstream fidelity test. `TestCheckerBridgeRefusalPending` awaits the shared TSGoError feature; no input was omitted.

Validation: lint-registry, go vet, empty gofmt -l output for oracle.go, and go test -json -count=1 -timeout=60m ./stage1/cohere/lint with the scratch original-upstream overlay. ADAMIC_TYPESCRIPT_SOURCE, ADAMIC_LINT_BENCH, ADAMIC_LINT_PROFILE_DIR and ADAMIC_LINT_PROFILE_SNAPSHOTS were all set.

Logs: /tmp/wave14-clean-final-gate.json and /tmp/wave14-clean-final-gate.time. Wall time: 1200.061s; nproc: 5; end load: [4.2890625, 4.505859375, 3.888671875].
