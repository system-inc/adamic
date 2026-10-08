Migrated 14 existing wave-17 rules to the unified registry and the area's one checker; no new claims.
Merged area c4bdc23fa into wave-17 at 62235a4de; max-classes branch is already contained in area.
Full lint: 30 top-level pass, 4 fail, 1 named skip; corpus repair separately passes on 939 files.
All 89 registry mutants caught, including all 14 added judgment-suppression mutants.
Full original-project upstream parity remains blocked by shared replay/census gaps; configurable boolean-prop-naming remains parked.

## Branches and implementation

`codex/typeaware-wave-17` merges origin/area/stage1-lint, rather than rebasing.
`codex/lint-port-max-classes-per-file` at 1b2737c32e88b0023cdeb7a25b6c058a1ca85e1a
is already an ancestor of the fetched area. No main or area push is made.
The merged area includes the RuleContext checker, typed replay, named kind-indexed
registry, options guard and allocator checks. Every inherited check is retained.

The following existing judgments now have owned descriptor directories beneath
`stage1/cohere/lint/rules/`, typed and node descriptors, upstream Go adapters,
firing witnesses, and comparison-only mutants:

- @next/next/no-async-client-component
- @next/next/no-duplicate-head
- @next/next/no-script-component-in-head
- nexus/correctness-no-collection-misuse
- nexus/correctness-no-discarded-outcome
- nexus/correctness-no-discarded-pure-result
- no-global-assign
- no-implicit-globals
- no-implied-eval
- no-throw-literal
- no-useless-backreference
- prefer-arrow-callback
- react-hooks/unsupported-syntax
- react-hooks/use-memo

Descriptors select ast.Kind names and receive the selected loaded ParseNode.
Nine existing judgments gained a one-node visitNode entry point; their legacy
validation runners retain their original loops. SourceFile rules perform their
whole-file judgment only when handed the source node. Messages remain verbatim
in the existing judgment modules. `upstreamTest: "Test"` captures all production
rule cases through the harness's exact rule-name filter, including custom-named
helpers/test groups. It deliberately does not narrow away failing captures.

`rules/no-global-assign/context.a` adapts the existing judgments to
`context.checker` and preserves edits/suggestions, converting byte positions to
UTF-16 only while constructing findings. It forwards questions to the area's
checker; it neither fabricates answers nor has a production transcript of its
own. The shared CheckerFacts structural view avoids nominal class identity
across the harness's copied module tree. The pre-existing raw bridge route is
retained only by the standalone legacy validation drivers. Factories defer
checker-dependent construction until selected, so unselected typed rules do
not require a program. Configured witnesses exercise global exceptions,
lexical bindings and arrow-callback options. Go adapters decode upstream option
structs/defaults rather than dropping options.

The merged compiler pin changed rooted path APIs. Owned bridge fact helpers and
six retained Go oracle builders now use RootedFilePath and the filesystem loader
API. No compiler implementation, shared harness or registry generator was edited.
The raw archived JSX probe was renamed from jsx-boundary.a to
jsx-boundary.tsx.txt without changing its bytes: executable corpus discovery
previously tried to compile this old failure fixture as an Adamic module.

## Checks and inputs

Evidence is in [validation-wave-17-unpark](validation-wave-17-unpark/).
Full command and counts are in lint-result.json; all output is in lint.jsonl.
The lint run used Go1.27.1, Node24.19.0, clang20.1.8, GOMAXPROCS=4,
GOFLAGS=-buildvcs=false, ADAMIC_LINT_BENCH=1, both profile input variables,
and ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-17-typescript, a clean pinned
TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8. Repository TypeScript is
d92d9bfee114c80be2c375d72edae966176e3a4f; cohere is
7945d102a6c18dd36adf9114a758ce646e8b2359. Profiles were provided, not skipped.

`go test -json -count=1 -timeout=90m ./stage1/cohere/lint` took
1993.882s wall (package 1991.405s): 30 top-level pass, 4 fail, 1 skip;
including subtests 123 pass, 4 fail, 1 skip. nproc=5, CPU quota=4;
one-minute load min/median/max 0.550/2.043/5.766. All 14 owned witnesses,
including configured variants, match Go byte-for-byte across Go, sanitized
native, Node and emitted JavaScript. TestMutants caught all 89 registered
mutants, including max-classes and all 14 new owned-judgment-suppressed mutants.
No mutant is claimed caught by a compiler failure.

After preserving the original failing run, the raw JSX archive rename repairs
TestCompilerAndStage1Agree. Its filtered uncached rerun passes in 546.183s:
939 files and 31,786,829 identical diagnostic bytes across the four execution
modes. This source-only corpus comparison does not claim full typed coverage;
typed selections without a program explicitly emit the harness's coverage rows.

Bridge tests pass in 0.283s. `go vet ./stage1/cohere/lint
./stage1/cohere/typeaware ./bridge/tsgo/checker`, gofmt and diff-check are clean.
Registry discovery validates 89 descriptors. The copied-module/import controls
pass after the structural view correction and still catch their mutation.
Setup initially failed on the new rooted path APIs; after repair it passed in
34.779s, including Go build 34.674s (timing lines in setup-retry.log).

307 completed typed captured-replay timing samples before the JavaScript
refusal: median Go program plus lint 51.625ms, native plus recording 69.580ms;
median per-case native/Go ratio 1.351. These are startup-heavy replay timings,
not a full original-project typed corpus throughput claim (timings.json).

The retained eight-suite run after the rooted-path repair took 879.361s wall:
six suites passed, two failed because newly introduced relative type imports
could not resolve in their scratch mutant copies. Their failures are preserved
in legacy-after.jsonl. The owned types now route through the existing rules
module export. Rerunning both affected suites passes in 415.139s
wall (package 413.253s, GOMAXPROCS=2), including every rule/bridge mutant,
configured controls, released handles, frozen compiler/repository corpus and
sanitizers. The eight suites' latest results are all pass, zero fail/skip, in
retained-final-status.json; this is a union of recorded runs, not a claim that
the earlier failed run was green.

Final TestOwnedWitnesses, TestWitnessScriptKind and TestNestedOutsideModuleCopy
pass in 49.018s wall (package45.729s). The firing witnesses,
option witnesses and copied-module/import mutants therefore pass with the final
source. The earlier registry TestMutants run caught all89; the final type-only
import repair reruns the affected retained mutants and compilation witnesses.
Registry validation and vet are clean again after that repair.


## Exact remaining blockers

1. `TestRulesAgree` loses each upstream project's compiler configuration and
companion sources: shared_test.go:254's record retains only Rule/File/Source and
options; lint_test.go:376-382 builds only strict=true and one file. The first
refusal is no-implicit-globals on JavaScript. Cohere's
`runImplicitGlobalsAsAScript`,
`cohere/internal/lint/rules/core/no_implicit_globals_test.go:34`, supplies
allowJs=true, target/lib ES2022 and module detection. Saved raw subject,
original capture, both configs/manifests and stdout/stderr are in
project-replay-gap/javascript/. The shared config refuses (exit2); the faithful
config returns the original one finding (exit0).

The second independent reproducer proves false agreement: cohere's
`correctnessNoDiscardedOutcomeRun`,
`cohere/internal/lint/rules/nexus/correctness_no_discarded_outcome_test.go:82`,
supplies five companion files. The original case reports one finding; the
one-file shared replay reports zero with exit0. Original capture, source, all five companion files and an independent unchanged-Go
0-versus-1 run are in project-replay-gap/. This must be repaired before claiming every upstream
case agrees in its original program. No guard or options adapter is relaxed.

2. TestJsxLintReleaseAndThroughput and TestJsxLintTrees stop at the fixed
`jsxSources` census, stage1/cohere/lint/jsx_integration_test.go:62. New captures
add async-client-component4, duplicate-head27, script-in-head23,
unsupported-syntax48 and use-memo64. The census still expects the five former
rules. Their witnesses pass; shared corpus census update is left to integration.
The current Go hook implementations use `IsInsideComponentOrHook`,
cohere/internal/lint/ecmascript/react/compiled.go:72, with AST/binding facts;
the former high-level IR parking reason does not apply to these two rules.

3. TestCheckerBridgeRefusalPending is the only skip, explicitly inherited from
stage1/cohere/lint/checker_pending_test.go:49: tsgoInspect still needs TSGoError
from the C error buffer (codex/tsgo-errors-as-values). Its gate is preserved.
No missing corpus/throughput/profile inputs are skipped.

4. react/boolean-prop-naming remains partial and is not registered as complete.
Its `regexp.Compile(settings.Rule)` at
cohere/internal/lint/rules/react/boolean_prop_naming.go:182 requires the
options pattern to become new RegExp(pattern, 'u'). The exact probe is
regexp-options.a.txt with compiler refusal in regexp-options.log at3:24:
nonconstant RegExp patterns are not lowerable. No handwritten matcher replaces
it. Existing partial/default-pattern checks and mutants are retained.

No new claims were made. Full lint remains red for three shared-harness failures
and the named pending skip after the separately repaired corpus check.
