# no-useless-assignment

Base d845dccde; source algorithm 335b5ae50. Rule directories only.
142 unique captured upstream cases and a firing witness match live Go,
sanitized native, Node and emitted JavaScript byte for byte, including fixes
and ordered suggestions. The verdict-inversion mutant disagrees on all three
runtimes. Whole-process medians: Go 55.089ms, sanitized native with recording
74.457ms, including startup under the package workload.
ProgramReads is empty, exactly as core/no_useless_assignment.go:114-125.
Only binding-declarations is used, including shorthand value symbols. Local
Go DeclarationsIn filtering is native; no aliases or foreign files are read.
The SourceFile listener acts on its dispatched file. Messages are verbatim.
upstreamTest is Test because the identifier-only registry must cover both
TestNoUselessAssignment and the two TestDeadStore lattice checks; capture
filters results to registered rules. Prior source-wave coverage was 261 full
program runs; the fresh comparison counts unique source/rule/options cases.

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

## Remaining wave-05 facts

All new questions belong to typeaware-08's lint-checker/facts lane. No bridge
question is included here. symbol-description and valid-typeof belong to wave 24.

| Waiting rule | Exact Go call and question needed |
| --- | --- |
| require-await | core/require_await.go:988: Checker_getResolvedSignature(..., CheckModeNormal), Signature.Target(), TypeParameters(), Parameters(), HasRestParameter(), Checker_getTypeOfSymbol and Checker_getReturnTypeOfSignature; declared-call-signature supplies declared and instantiated signature facets. core/require_await.go:1019: Checker_getIndexTypeOfType(..., Checker_numberType); core/require_await.go:1130-1138: GetTypeAtLocation, Checker_getPropertyOfType and GetTypeOfSymbolAtLocation; type-projection supplies numeric index, property, heritage and call-signature projections. The old reference-shape expands deferred reference type arguments via Checker_getTypeArguments; its raw type-ID traversal also awaits facts. |
| @typescript-eslint/restrict-template-expressions | typescript/restrict_template_expressions.go:266: type_checking.GetNumberIndexType (Checker_getIndexTypeOfType(..., Checker_numberType)); type-projection supplies the resulting element type identity. |
| nexus/correctness-no-process-exit-after-output | nexus/correctness_no_process_exit_after_output.go:411-413: GetSymbolAtLocation and GetAliasedSymbol; output-symbol supplies alias-followed identity and full declaration/name-kind/module ancestry. Lines 314-342: GetResolvedSignature, Signature.Declaration(), ast.GetSourceFileOfNode, ast.GetFunctionFlags and GetReturnTypeOfSignature; output-callee supplies signature declaration/body spans and return flags. The native writer-body traversal needs guarded same-program foreign selectors (other-file / Program.Inspect), replacing the legacy foreign read-and-parse path. |

runtime-modules belongs to the separate blocking-stream rule, not these five
landing directories: Compiler.SourceFiles, GetResolvedModuleFromModuleSpecifier
and GetSourceFileForResolvedModule provide its runtime edges and computed loads.
That rule remains outside this facts slice and the held rule-only landing.

