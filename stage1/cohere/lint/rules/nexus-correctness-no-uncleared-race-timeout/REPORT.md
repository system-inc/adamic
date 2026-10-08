# nexus/correctness-no-uncleared-race-timeout

Base d845dccde; source algorithm 335b5ae50. Rule directories only.
21 unique captured upstream cases and a firing witness match live Go,
sanitized native, Node and emitted JavaScript byte for byte, including fixes
and ordered suggestions. The lost-handle verdict mutant disagrees on all three
runtimes. Whole-process medians: Go 53.155ms, sanitized native with recording
75.113ms, including startup under the package workload.
ProgramReads is exactly ReadsCompilerOptions | ReadsDefaultLibrary (Go line 80).
The CallExpression listener receives its node; helpers are created on a typed
visit, never on a syntax-only factory call. Messages are verbatim.
Shared binding-declarations and symbol-provenance replace the old ancestry
question. Provenance is guarded by askFile(ReadsDefaultLibrary), with no added
ReadsOtherFiles permission. No private program or foreign parser is used.
These facts supply GetSymbolAtLocation, ast.GetSourceFileOfNode,
IsSourceFileDefaultLibrary, ast.IsExternalModule and ast.IsGlobalScopeAugmentation
at nexus/correctness_no_uncleared_race_timeout.go:132-174 and :234-269;
shorthand values use GetShorthandAssignmentValueSymbol at :335-339.
The prefix TestCorrectnessNoUnclearedRaceTimeout captures all five named tests,
including the real generator and DOM cases. No bridge file is changed.

Validation: lint-registry, gofmt and vet passed. TestRulesAgree,
TestOwnedWitnesses, all 82 registered mutants, profile artifacts/compilation/
snapshots, and the 890-file compiler/repository comparison passed.
The counted build reported 270 allocations and 270 frees.
Whole package: 33 top-level passes, 2 failures, 1 skip (282/2/1 with subtests).
Wall 2280.177s; nproc 5, quota 4 cores; one-minute load min/median/max
0.163/1.417/5.255.

The failures are the base's fixed JSX inventories: jsx_integration_test.go:93
and :138 expect a map without the 24 legitimate no-useless-assignment JSX cases.
Reproduce with TestJsxLintReleaseAndThroughput or TestJsxLintTrees on this tree.
No shared check was changed. TestCheckerBridgeRefusalPending is the only skip:
awaits codex/tsgo-errors-as-values, so tsgoInspect returns TSGoError from C.
These are not counted as green. The earlier package attempt stopped at an eager
checker factory bug; lazy creation fixed it and TestDotARename passed.

Command: source /workspace/adamic-tools/env.sh; GOFLAGS=-buildvcs=false,
GOMAXPROCS=4, ADAMIC_LINT_BENCH=1, ADAMIC_TYPESCRIPT_SOURCE set to clean
050880ce at /workspace/wave-05-typescript, both profile variables set to the
fresh /workspace/facts-wave05-profiles-certified directory. Then:
`go test -overlay=/workspace/facts-wave05-certification-overlay.json -json -count=1 -timeout 3h ./stage1/cohere/lint`.
The external overlay adds one owned-upstream check and replaces no existing
test; it is not shipped here. Ordinary TestRulesAgree also passed the capture.
The syntax corpus comparison does not claim typed execution over the compiler.

Setup failed on Git's rejection of the cohere symlink. Timings: Go 0.022s,
Node 0.027s, markdown 0.098s, clang 0.257s. The workaround uses the verified
7945d102 checkout and installed toolchain; buildvcs=false affects stamping only.
Six GB of old workspace build cache was removed; source and logs were preserved.
Full logs: /workspace/facts-wave05-package-certified.jsonl and
/workspace/facts-wave05-package-certified-metrics.json. Summary:
/workspace/facts-wave05-results.json. Small excerpts are in checks.log.
