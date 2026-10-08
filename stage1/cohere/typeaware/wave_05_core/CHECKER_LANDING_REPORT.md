# Wave 05 shared checker landing

Merged origin/area/stage1-lint c4bdc23fa86d55cf7e579989201c11258f4d3a62 into the
owned wave branch, with merge commit 7b908dc74. No rebase or protected-branch push.
The adjacent-overload-signatures branch 2f55a1da6c4cb9cfa38d81f4ac678318078411cb
is already an ancestor of the fetched area; its descriptor remains unchanged.

Nine owned descriptors now use RuleContext.checker. Legacy algorithms receive
callbacks into that checker; foreign-file selectors also travel through its
record/replay stream. There is one program per run. The original independent
historical drivers and their checks remain in the tree.

Generic declaration-origin and same-program foreign-file questions live in
bridge/tsgo/checker/type_declaration_origins.go and other_file.go, with native
question wrappers inside owned rule directories. The pinned bridge API requires
RootedFilePath.AsString, SourceFile.PathKey and the resolved-module object.
Owned files were updated to those APIs. No shared registration generator or
existing harness file was changed.

The additional owned wave05_programs_test.go check materializes every original
typed fixture, including in-memory fixtures, declarations, modules, options and
compiler configuration. It uses unchanged cohere rules and assertions. The
shared source-only capture does not preserve those programs. There are 800
original typed runs: type parameters 225; templates 100; process output 56;
race timeout 21; blocking streams 52; dead stores 261; await 15; Symbol 34;
typeof 36. The final replay retained all 800: 797 pass, three parser failures.

The template port preserves absent versus false flags, bare and structured
allowlists, captured Go options, explicit empty allowlists, intrinsic names and
array number-index recursion. valid-typeof preserves requireStringLiterals and
its exact suggestion behavior. Full findings, fixes and ordered suggestions
use the shared serialization model.

## Exact blockers retained

- Type-parameter rule: TestNoUnnecessaryTypeParametersStaysSilent/valid84,
  cohere/internal/lint/rules/typescript/no_unnecessary_type_parameters_test.go:316.
  The nested function type containing a type predicate fails the native parser
  at byte 281 with expected CloseParenToken, got Identifier. Reproducer:
  ../../lint/rules/typescript-no-unnecessary-type-parameters/gaps/nested-function-predicate.ts.txt.
- Blocking-stream rule: correctnessRequireBlockingStandardStreamsLintEngineParity,
  cohere/internal/lint/rules/nexus/correctness_require_blocking_standard_streams_test.go:219,
  specifically await import(output) at line 226. Both original and fixed real-site
  cases fail the native parser. Reproducer:
  ../../lint/rules/nexus-correctness-require-blocking-standard-streams/gaps/dynamic-import.ts.txt.
- Shared JSX test: jsxSources in stage1/cohere/lint/jsx_integration_test.go:53
  hard-codes its rule/count map. Registering no-useless-assignment adds 24 real
  JSX cases and fails TestJsxLintReleaseAndThroughput and TestJsxLintTrees.
  Upstream provenance: cohere/internal/lint/rules/core/no_useless_assignment_test.go:19,
  TestNoUselessAssignmentFires at line 199 and StaysSilent at line 222.
  Those cases remain in the owned complete-program comparison; the shared map
  was not changed, nor were cases filtered away.
- Existing TestCheckerBridgeRefusalPending at checker_pending_test.go:43 skips
  because internal/load/prelude.d.ts lacks TSGoError. Its stated dependency is
  codex/tsgo-errors-as-values. No input-dependent test is intentionally skipped.

React reservations remain parked, not implemented or certified. Current exact
missing helpers are reference.WritesToBinding in cohere/internal/lint/ecmascript/reference/write.go:100
(called by globals.go:265), high_level_intermediate_representation.ForFunction
in cohere/internal/lint/rules/react/immutability.go:170, and
ForFunctionWithoutManualMemoization in react/no_deriving_state_in_effects.go:142,
with capture/effect analysis at line 325. JSX itself is no longer their blocker.

## Validation

Final validation used the pinned corpus and all opt-in lint inputs:

```sh
source /workspace/adamic-tools/env.sh
export GOMAXPROCS=4 GOFLAGS=-buildvcs=false
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-05-typescript
export ADAMIC_LINT_BENCH=1
export ADAMIC_LINT_PROFILE_DIR=/tmp/wave05-final-profiles
export ADAMIC_LINT_PROFILE_SNAPSHOTS=/tmp/wave05-final-profiles
go test ./stage1/cohere/lint -json -count=1 -timeout=90m
go test ./stage1/cohere/lint -run '^TestWave05SourceOnlyCorporaAgree$' -json -count=1 -timeout=30m
go test ./bridge/tsgo ./bridge/tsgo/checker -v -count=1 -timeout=20m
go run ./cmd/lint-registry
go vet ./...
gofmt -l bridge/tsgo/checker stage1/cohere/lint/wave05_programs_test.go
```

The complete lint gate exited 1: **28 top-level pass, 7 fail, 1 skip**;
including nested cases, 913 pass, 10 fail, 1 skip (package exit excluded).
No missing-input skips. The sole skip is TestCheckerBridgeRefusalPending,
whose exact unresolved TSGoError dependency is named above. Failed top-level
checks are TestJsxLintReleaseAndThroughput, TestJsxLintTrees, TestRulesAgree,
TestNodeTableIsLinkOnly, TestShardsAgree, TestProfileSnapshotsAgree and
TestWave05OriginalProgramsAgree. Four aggregate comparisons stop on the
same type-parameter parser refusal; the JSX checks stop on the stale shared map.
Nothing was removed or weakened. The branch is **not fully green**.

The original-program check retained every case and continues after a failure:
797 of 800 complete programs are byte-identical on Go, Node, emitted JavaScript
and ASan/UBSan/LSan native. The type-parameter rule has 224 pass / 1 parser
failure; blocking streams has 50 pass / 2 parser failures. Seven rules have
all original typed programs green: templates 100, dead stores 261, process
output 56, race timeout 21, await 15, Symbol 34 and typeof 36.

The separate source-only check, added after the full gate had begun, passed
all 276 cases in 207.333s: adjacent-overload-signatures 107, require-await 88,
symbol-description 30, valid-typeof 51. It uses the same production sources
and the same four-runtime comparison. All owned witnesses passed; exact fixes
and ordered suggestions are included in wire equality. The supplemental test
and its counts are separate from the full gate above.

TestMutants passed all 84 registry mutants, including all nine owned verdict
mutants and adjacent_overload_separation_ignored. The declaration-origin mutant
is caught by TestTypeDeclarationOrigins' cwd assertion; a foreign-file mutant
replacing the delegated question with options is caught by its direct-versus-
delegated assertion. Production source is unchanged by that Go overlay.
Both bridge packages pass in 111.096s, including 100 C-ABI queries, released and
zero-handle rejection, stale-handle mutant and 162-position independent Go
comparison under ASan/UBSan/LSan. No bridge skips. Full go vet passes;
gofmt output and git diff --check are empty. lint-registry enumerates the nine
new descriptors beside the area's existing rules, including adjacent.

The corpus comparison passes on 934 compiler/stage1 files and 31,798,081
identical bytes. The release throughput manifest has 77 files and 28,312
findings: best-of-five native 11.764095s, Go 1.523049s, Node 5.839256s.
This source-only release benchmark is not typed whole-corpus certification:
typed selection without a program is explicitly reported by the shared model.
The complete-program timing observations below include program construction
and native recording, so they are not pure rule-execution timings.

| Rule | Go median ms | Sanitized native + recording median ms | Node replay median ms | Emitted JS replay median ms |
| --- | ---: | ---: | ---: | ---: |
| @typescript-eslint-no-unnecessary-type-parameters | 20.227 | 35.992 | 315.489 | 72.763 |
| @typescript-eslint-restrict-template-expressions | 19.849 | 36.634 | 311.859 | 71.911 |
| nexus-correctness-no-process-exit-after-output | 58.890 | 80.648 | 321.286 | 81.051 |
| nexus-correctness-no-uncleared-race-timeout | 20.257 | 37.383 | 308.866 | 75.257 |
| nexus-correctness-require-blocking-standard-streams | 21.731 | 39.966 | 324.544 | 76.150 |
| no-useless-assignment | 20.655 | 37.517 | 317.828 | 75.439 |
| require-await | 22.632 | 45.085 | 329.840 | 77.630 |
| symbol-description | 19.041 | 34.531 | 306.369 | 71.503 |
| valid-typeof | 20.071 | 36.271 | 323.662 | 73.372 |

Full-gate wall time: 2410.856s (40m10.856s); nproc 5, cgroup quota four cores.
One-minute load sampled every ten seconds: min 0.05, median 1.71, max 5.70,
final 3.35. The native cache relocation paused four owned gate processes
for 11.047s; this pause is included in wall time. Setup final timing is recorded
in setup-final.log (38.053s total). Full commands, outcomes and compressed JSON
logs live in evidence/checker-landing/. Historical preflight logs retain the
previously fixed options/foreign-file failures and are not final verdicts.

Three React reservations remain parked for the exact analysis symbols named
above. No new rules were claimed. No main or area branch was pushed.
