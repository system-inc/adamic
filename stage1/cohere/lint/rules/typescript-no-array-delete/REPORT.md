# @typescript-eslint/no-array-delete

Typed DeleteExpression listener using the shared checker. Reporting takes the handed node; messages and the three-edit splice suggestion are preserved.

All 44 original upstream cases retain their Go finding counts and match Go, Node, emitted JavaScript and ASan/UBSan native byte for byte. The owned witness fires. The array-guard mutant is caught; all 94 registered mutants are caught. Median program-plus-lint time over cases and witness: Go 59.204ms, native 76.804ms (includes checker startup).

Complete lint gate on ad7bd066: 47 pass, 0 fail, 1 skip, including the scratch-only original-upstream fidelity check. TestCheckerBridgeRefusalPending awaits shared TSGoError support; every optional input is set.

Validation: lint-registry, go vet, empty gofmt -l oracle.go; go test -overlay=/tmp/wave14-area-land-original-overlay.json -json -count=1 -timeout=3h ./stage1/cohere/lint. ADAMIC_TYPESCRIPT_SOURCE, ADAMIC_LINT_BENCH, ADAMIC_LINT_PROFILE_DIR and ADAMIC_LINT_PROFILE_SNAPSHOTS were set.

Logs: /tmp/wave14-area-land-gate.json and .time. Wall 1170.075s; nproc 5; end load [3.46630859375, 4.6982421875, 4.60595703125]. The held branch also passed its full gate: /tmp/wave14-area-held-gate.json.
