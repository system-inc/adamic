Merged lint area c4bdc23fa86d55cf7e579989201c11258f4d3a62 into codex/typeaware-wave-11; preserved checks from both histories.
All 21 implemented legacy rules pass their existing native production-default comparison suites, sanitizers, semantic mutants and released-handle checks.
Full lint package: 32 PASS, 2 FAIL, 1 SKIP, all optional inputs supplied; both storage-failed benchmark checks pass on retry.
Registry mutants caught: 75; three listener declaration mutants caught; legacy Node/emitted-JavaScript certification remains incomplete.
Full-gate wall 2110.182s; nproc 5; one-minute load min 0.78, max 6.64, final 1.59; four HIR/evaluation rules remain parked.

## What changed

The merge contains the area's live checker Program. No private checker stand-in or
hand-authored checker-answer transcript was found: the legacy entries ask the
actual merged bridge/tsgo/checker.Program. They still call its native ABI directly,
rather than using unified RuleContext descriptors. That remaining work is not
credited as complete by these native-only suites.

The compiler pin now returns RootedFilePath and accepts rooted config/file lists.
Owned bridge fact writers use FileName().AsString() and PathKey(); owned independent
oracle loaders use the actual new FS/compiler-host and rooted-file APIs. Assertions
and source populations are retained. Shared facts.go has only the required
FileName().AsString() compatibility change beyond the merge.

SourceKeyword shifted later numeric SyntaxKind values. All three legacy numeric
tables and all 21 manifests now match the actual compiler. Go also moved
NoFuncAssign from SourceFile to FunctionExpression and FunctionDeclaration listeners;
its declaration matches that current registration. The listener probe imports its
actual nominal classes so the merged compiler can lower their methods. The wrong-kind
and foreign-binding mutants were adapted to the changed types and still compile,
run normally and lose only the independent byte comparison. A failed build is not
counted as a mutant catch.

## Coverage and limits

All eight owned suites have a passing rerun in summary.json: first, second, third,
fourth, fifth, seventh, eighth and listener declarations. They retain the original
production-default controls, captured parse-clean upstream sources, repository corpus,
TypeScript src/compiler corpus, sanitizer checks and invalid/released-handle checks.
Native versus Go process/load/query timings are in the full owned logs. Options across
all upstream rows and all three execution backends are not certified by these legacy
suites. The original filters excluding parse diagnostics were not changed.

The full lint-package run supplies ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-11-typescript,
ADAMIC_LINT_BENCH=1, ADAMIC_LINT_PROFILE_DIR and ADAMIC_LINT_PROFILE_SNAPSHOTS.
The pinned TypeScript corpus is 050880ce59e30b356b686bd3144efe24f875ebc8.
The compiler-and-stage1 comparison passes on 930 files. Both full-run failures are
checker-archive writes exhausting scratch storage, in TestJsxLintReleaseAndThroughput
and TestThroughput. Their serial retries PASS. Original full-run counts remain intact.

The sole skip is stage1/cohere/lint/checker_pending_test.go:49,
TestCheckerBridgeRefusalPending: the library prelude/tsgoInspect must expose TSGoError
from the C error buffer. codex/tsgo-errors-as-values is the named dependency.
It cannot be supplied as a missing test input; its check remains unchanged.

## Exact remaining blockers

Legacy driver replay/migration: stage1/cohere/typeaware/rules.ts:104 invokes
native tsgoInspect directly; oracle/adamic.mjs:137 returns the documented refusal
without a native recording. Running wave_11_suite.a on unpark-wave-control.a
(the existing for-in-array judgment) exits 70 with TypeError: this.text.indexOf
is not a function. The required upstream operation is GetNumberIndexType at
cohere/internal/lint/rules/typescript/no_for_in_array.go:163. The actual bridge can
answer this question natively. The missing part is a RuleContext/replay adapter for
the legacy drivers, not a missing Go fact and not a passing Node comparison.
The emitted-JavaScript CLI refuses the same direct native call at rules.ts:104
as an unlinked typescript-go library call. stdout, stderr, exits and minimal source
are retained here. No fabricated answers were used to make these backends agree.

Native HIR/evaluation prerequisites remain absent, even though the area's SSA module
is now present. The parked calls are ForFunctionWithoutManualMemoization at
cohere/internal/lint/ecmascript/high_level_intermediate_representation/cache.go:178,
used by set_state_in_effect.go:267; ForFunction at cache.go:63, used by
set_state_in_render.go:183 and static_components.go:120; Lower at lower.go:120,
Construct at ssa.go:87, and the manual-memo/capture/post-dominator pipeline.
For jsx/no-constructed-context-values the missing native function-return and escape
evaluation is functionEvaluation at jsx_no_constructed_context_values_stability.go:532
and anyEscapes at :603. Fresh upstream controls for all four PASS, but that does not
prove native ports; no native validators or mutants are claimed for these four.

## Reproduction and evidence

All test output was redirected directly to logs, then compressed here.
unpark-lint-gate.py reproduces the full all-input run from a supplied branch worktree.
Commands include go test -json -count=1 -timeout=90m ./stage1/cohere/lint,
the eight named owned suites, go test ./bridge/tsgo/checker -count=1,
go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware ./stage1/cohere/lint/..., and
gofmt on the owned changed Go files. Vet and formatting pass; git diff --check passes.
Setup PASS 168.770s, cache warm 168.744s, nproc 5. Cohere pin remains
7945d102a6c18dd36adf9114a758ce646e8b2359 and its TypeScript submodule is
d92d9bfee114c80be2c375d72edae966176e3a4f. No new rules were claimed, no PR opened,
and no main or area branch was pushed.
