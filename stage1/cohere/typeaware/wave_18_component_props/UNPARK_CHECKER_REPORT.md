Branch: codex/typeaware-wave-18; merge aaf95e6407ad56ca7d2aa15d86a36a2c31f59b03 takes origin/area/stage1-lint c4bdc23fa86d55cf7e579989201c11258f4d3a62 without rebasing or discarding checks.
Rules: five React registry integrations now use context.checker, but zero rules are re-green at the new cohere pin because the shared bridge cannot build.
Mutants: new owned verdict mutants are registered; execution blocked before tests. Previous mutant evidence is historical and is not current-pin proof.
Lint: complete all-input package build failed; tests pass=0 fail=0 skip=0, package fail=1; wall 0.120174479s, nproc=5, start/end sampled load 1.86/2.88/1.34.
Stopped on: shared SourceFile API incompatibility; remaining migrations and four native HIR/SSA/capture source analyses are unfinished. No new claims.

## Shared build blocker

After setting GOPROXY=https://proxy.golang.org|direct, bash cloud/setup.sh fails in the Go build, not on module retrieval. Owned jsx_syntax_facts.go and component_property_syntax_facts.go serializers were corrected to explicitly convert RootedFilePath to string. Three errors remain outside those owned files:

* bridge/tsgo/checker/declaration_facts.go:17 passes SourceFile.FileName() to fields.text(string).
* bridge/tsgo/checker/declaration_facts.go:22 calls removed SourceFile.Path().
* bridge/tsgo/checker/facts.go:292 passes SourceFile.FileName() to fields.text(string).

Exact upstream API: SourceFile.FileName() returns tspath.RootedFilePath at cohere/TypeScript/tsc/internal/ast/ast.go:2703; SourceFile.PathKey() returns tspath.PathKey at :2707. Program.IsSourceFileDefaultLibrary requires tspath.PathKey at cohere/TypeScript/tsc/internal/compiler/program.go:1804. Existing serializer callers predate the pin change. No cast is added to shared files, no compiler pin is rolled back, and no check is relaxed. The complete JSON build output is the reproducer.

## Checker migration done, not yet runtime validated

Five owned directories under stage1/cohere/lint/rules/ are prepared: wave18-react-jsx-fragments, wave18-react-jsx-no-undef, wave18-react-no-adjacent-inline-elements, wave18-react-static-property-placement and wave18-react-style-prop-object. Each has a .a entry, .a verdict core and pure fact decoder, typed descriptor with named kinds and ReadsOtherFiles, upstream Go adapter, positive raw witness and verdict mutant. RuleContext.checker.askFile requests the existing raw jsx-syntax-facts or component-property-syntax-facts question. The area's checker owns the program and transcript recording/replay; the rule modules do not open a private program or generate saved answers. Raw source facts are indexed once in prepare; visit acts on the supplied node, matching its byte span and named kind. Missing mappings are explicit refusals. Finding offsets convert UTF-8 to UTF-16 only at report construction. Options adapters decode into upstream Go types.

These five rule cores preserve the prior verdict implementation. Their new parity, upstream-case coverage, sanitizer and mutant checks have NOT executed. Descriptor generation and registry tests pass (0.479s), four top-level tests, plus their subtests; initial descriptor generation required adding the mandatory no-options adapter export. Standalone old harness evidence remains at the earlier pin and does not prove this integration.

The other twelve completed standalone ports have not yet moved into the registry: await-thenable, class-literal-property-style, Next no-title-in-document-head, no-class-assign, no-const-assign, no-constant-binary-expression, no-new-func, no-new-native-nonconstructor, no-new-wrappers, prefer-promise-reject-errors, prefer-regex-literals and prefer-rest-params. They share the same global bridge build blocker. No fabricated checker answer or rule verdict is introduced to work around it. Their existing checks are retained by the merge.

## Analysis dependencies still missing

* react-hooks/static-components and react-hooks/set-state-in-render call high_level_intermediate_representation.ForFunction at cohere/internal/lint/ecmascript/high_level_intermediate_representation/cache.go:63, with MayHoldComponentOrHook at spelling.go:88. Native equivalents of the lowering/SSA/capture path remain required. Call sites: cohere/internal/lint/rules/react/static_components.go:120 and set_state_in_render.go:183.
* react-hooks/set-state-in-effect calls high_level_intermediate_representation.ForFunctionWithoutManualMemoization at cohere/internal/lint/ecmascript/high_level_intermediate_representation/cache.go:178; call site set_state_in_effect.go:267. Native source lowering, memo erasure and SSA translation remain required.
* react/jsx-no-constructed-context-values requires jsxNoConstructedContextValuesStabilityWalk.functionEvaluation at cohere/internal/lint/rules/react/jsx_no_constructed_context_values_stability.go:532, anyEscapes at :603 and jsxNoConstructedContextValuesStaysHome at :649. Its native callback-return/capture/escape evaluation remains required.

Prepared HIR reporters and cores from earlier work are kept, but Go-prepared inputs are not represented as native source analysis.

## Full lint command and inputs

The complete command is go test -json -count=1 -timeout=60m ./stage1/cohere/lint, with installed toolchain environment sourced. Inputs: GOPROXY=https://proxy.golang.org|direct, GOMAXPROCS=4, ADAMIC_LINT_BENCH=1, ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-18-typescript at clean 050880ce59e30b356b686bd3144efe24f875ebc8, ADAMIC_LINT_PROFILE_DIR and ADAMIC_LINT_PROFILE_SNAPSHOTS both /workspace/wave18-unpark-all-inputs-profiles, freshly created. No test filter. Output goes directly to logs. The build fails before any test starts, so there are no named skips and no green rule/mutant claims. The area report's pending TSGoError refusal dependency is retained unchanged; it was not reached in this run. All measured input/load metadata and JSON counts are in validation-unpark-checker/summary.json. The failed earlier witness attempt and setup output are also preserved.

The merge and blocked progress are pushed only to codex/typeaware-wave-18. Integration owns main and area branches. No new claims were taken, and shared harness/registration files were not edited.
