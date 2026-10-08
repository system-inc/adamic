@typescript-eslint/no-explicit-any on the unified harness, from lint-batch/wave2-02 at 95968dd93ad0876245f181af63931c3134e14b3d. Only this registered directory is added; the small file-extension helper is in rule.a.

209 captured upstream source/file/options cases agree byte for byte with independent Go cohere on source Node, emitted JavaScript and ASan/UBSan native. The witness fires; typescript-no-explicit-any-message compiles, runs and is caught by comparison. Messages are verbatim and the adapter decodes FixToUnknown and IgnoreRestArgs.

The clean lint package exits 0: 147 PASS, 0 FAIL, 1 SKIP including subtests; 41/0/1 top-level. All 94 registered mutants are caught. TestRulesAgree, TestCompilerAndStage1Agree, TestProfileSnapshotsAgree and TestOwnedWitnesses pass. Wall 1470.143s, nproc 5.

The sole skip is TestCheckerBridgeRefusalPending at stage1/cohere/lint/checker_pending_test.go:49: internal/load/prelude.d.ts lacks TSGoError from the C error buffer. Its pending-refusal-control question at :51 is the existing reproducer. No external-input check skips; no shared code or skip guard was changed.

Commands: go run ./cmd/lint-registry; gofmt -l stage1/cohere/lint/rules/typescript-no-explicit-any/oracle.go; go vet ./stage1/cohere/lint; go test -json -count=1 -timeout=90m ./stage1/cohere/lint. Registry, formatting and vet pass.

Inputs: ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-11-typescript (050880ce59e30b356b686bd3144efe24f875ebc8); ADAMIC_LINT_BENCH=1; both ADAMIC_LINT_PROFILE_DIR and ADAMIC_LINT_PROFILE_SNAPSHOTS=/tmp/unpark-batch2-clean-profiles, initially empty. The command sources /workspace/adamic-tools/env.sh and uses GOCACHE=/workspace/batch2-go-cache, XDG_CACHE_HOME=/tmp/unpark-runtime-cache, GOFLAGS=-buildvcs=false and GOMAXPROCS=4.

Logs: /tmp/unpark-batch2-clean-package.jsonl and /tmp/unpark-batch2-clean-package-result.json; /tmp/batch2-clean-{registry,gofmt,vet}.log. The held branches merged batch 2 without rebasing and pass separately: no-explicit-any 147/0/1; wave-11 146/0/1. Their logs use batch2-held-any-retry and batch2-held-wave in the same /tmp/unpark- naming.

No legacy wave port is included: those ports lack complete unified source-Node/emitted-JavaScript replay certification. No duplicate or partially certified rule, wave suite, evidence archive or shared edit is added.
