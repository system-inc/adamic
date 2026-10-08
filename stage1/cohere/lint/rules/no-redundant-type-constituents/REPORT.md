# @typescript-eslint/no-redundant-type-constituents

Clean landing base: origin/area/stage1-lint fb6cb5dbf79e247d74dccf16d8e3aefc5b31c039.
Go cohere: 7945d102a6c18dd36adf9114a758ce646e8b2359.

The descriptor subscribes to UnionType and IntersectionType and takes the handed node.
Messages are verbatim. The Go rule has no options. The checker is RuleContext.checker.
All changes stay in this directory; no bridge or shared registration edits, compressed evidence, wave suite, or external oracle fixtures are included.
Existing area modules used unchanged: stage1/cohere/typeaware/facts.ts and flags.ts decode checker types/flags; diagnostic.ts buffers findings for shared reportRange.

Upstream prefix TestNoRedundantTypeConstituents captures 108 unique cases; all pass byte-identically against Go on Node, emitted JavaScript and ASan/UBSan native. No parser-limit case belongs to this rule.
TestOwnedWitnesses passes; the witness reports both any and never constituents.
TestMutants/no-redundant-type-constituents-decision passes: reversing the never return-type exemption compiles and loses the expected never finding; comparison catches it on all three runtimes.

Commands: go run ./cmd/lint-registry; gofmt -w oracle.go; go vet ./stage1/cohere/lint ./stage1/cohere/lint/registry; go test -json -count=1 -timeout=90m ./stage1/cohere/lint. Generator, formatting and vet pass.
Full package: 117 pass, 0 fail, 1 skip (37 top-level passes). Existing TestCheckerBridgeRefusalPending at stage1/cohere/lint/checker_pending_test.go:49 awaits TSGoError in the prelude; it is unchanged.
Inputs: GOMAXPROCS=4, GOFLAGS=-buildvcs=false, ADAMIC_LINT_BENCH=1, ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-lint-landing, fresh ADAMIC_LINT_PROFILE_DIR and ADAMIC_LINT_PROFILE_SNAPSHOTS. TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8 remains clean.
Wall: 1900.559s; nproc 5; one-minute load mean 2.007, peak 4.662.
The 108 ordered upstream runs total Go program/lint 6.171s versus native program/lint/recording 8.082s; native includes transcript recording.
Local logs: /tmp/wave10-clean-validation/lint-all.jsonl, all-inputs-results.json, rule-timing.json; /tmp/wave10-clean-registry.log; /tmp/wave10-clean-vet.log. Logs are not committed.
