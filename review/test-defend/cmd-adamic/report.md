# Target parsing defense

Starting main: b8bcadb2c493173855f19d7e5c508b34f5eeb5b6. Branch: test-defend/cmd-adamic. Audit evidence read from test-audit/cmd-adamic, including audit.json, limitations.txt, environment.json, scope.txt, reached-functions.txt, mutants.json and the report generator. Its starting main was 1f34d0d300301faebc94d397adee1a090923d2c7. Full target_test.go and all other current CLI test files were read before mutation, followed by run/build source.

## Code under test and oracle

Code under test: CLI run dispatch and build target-option parsing in cmd/adamic/main.go and cmd/adamic/tsgo.go. The requested row reaches run and build, returning before loading or compiling a program. Only build's existing error exit constant was mutated.

Oracle: self, the test's handwritten usage-error exit code 2. It checks unknown target, missing target operand and duplicate target options. It does not assert stderr text; this proof establishes classification rather than diagnostic content. Its named subsumer passes a valid target and checks the distinct --tsgo unsupported-option rejection, code 1 plus diagnostic substring.

## Coverage lead

Separate isolated commands with -coverpkg=./cmd/adamic produced target.cover and subsumer.cover:

```
ADAMIC_ORACLE_WASI=1 WASI_SYSROOT=/tmp/defend-cmd-adamic/sdk/share/wasi-sysroot ADAMIC_BUILD_CACHE_DIR=/tmp/defend-cmd-adamic/cache/coverage-target timeout 120 go test -json -count=1 -timeout 90s -coverpkg=./cmd/adamic -coverprofile=/tmp/defend-cmd-adamic/target.cover ./cmd/adamic/ -run '^TestBuildTargetParsing$' > target-coverage.log 2>&1
ADAMIC_ORACLE_WASI=1 WASI_SYSROOT=/tmp/defend-cmd-adamic/sdk/share/wasi-sysroot ADAMIC_BUILD_CACHE_DIR=/tmp/defend-cmd-adamic/cache/coverage-subsumer timeout 120 go test -json -count=1 -timeout 90s -coverpkg=./cmd/adamic -coverprofile=/tmp/defend-cmd-adamic/subsumer.cover ./cmd/adamic/ -run '^TestWASIRejectsTSGoArchive$' > subsumer-coverage.log 2>&1
```

Exclusive blocks versus that subsumer: main.go:97 (dispatch of target options after the source), and tsgo.go:20..21 (usage diagnostic and code 2 in the invalid-target guard). D1 changes the latter return constant. The semantic difference is malformed option syntax versus a well-formed target with an unsupported archive option.

## Observed defense

D1 was planned in plan.json before its run. At origin/main cmd/adamic/tsgo.go:21 change return 2 to return 1 in build's guard for unknown, missing or repeated target values. This is the change-constant menu operation, not an empty-answer probe.

Baseline ran all 13 current top-level tests, with ADAMIC_ORACLE_WASI=1, a scratch WASI SDK and a fresh baseline build cache. It passed in 31.853 binary seconds. The audit had ten rows; current main adds TestUsageExitStatus and two WASI sanitizer-option placement rows. Current main also moved the sanitizer assertions out of TestBuildTargetParsing; this row now checks malformed target syntax.

D1's whole-package command and all outcomes are in matrix.json and raw D1.log. It completed in 21.583 test-binary seconds and 26.178 wall seconds, including compilation. Only TestBuildTargetParsing failed:

target_test.go:18: [build --target unknown missing.a -o out.wasm]: got 1, want usage error (2)

No skips, panics or unknown rows. The 12 passing top-level rows are:
- TestExplainChecksDriver
- TestWASIRequestSelection
- TestWASIRequest
- TestRequestNativeStaysCommand
- TestWASIRequestModuleSelection
- TestWASIRequestThrows
- TestWASIRejectsTSGoArchive
- TestUsageExitStatus
- TestBuildWASISanitizeTargetAfterSource
- TestBuildWASISanitizeTargetBeforeSource
- TestExplainChecksOutput
- TestNonNullExplainChecks

TestExplainChecksDriver is an idle subprocess helper in its ordinary top-level invocation; its parent rows also passed. D1.diff is standalone against starting main, passed go vet ./cmd/adamic, and passed git apply --check after source restoration. Production source and tests are unchanged. The restored target and named subsumer pass, as recorded in restored.log. One unique catch suffices to defend the requested row, so no further mutant was needed.

## Costs and brief issues

Warm Go/Node tools worked; setup was skipped, nproc=5. npm ci in stage3/api succeeded in 455 ms. The optional WASI SDK was absent, so SDK 27 was downloaded into this unit's scratch directory to enable both integration rows. The download took about 22 seconds, plus extraction. Nothing was installed in the repository.

/tmp has total capacity 8.8 GB, making the requested 15 GB free threshold impossible. Only previous-unit /tmp/defend-fuzz was removed. The repository and tools were untouched; /workspace had 18 GB free initially and 17 GB after running. No disk baseline failure occurred.

The full-refspec audit fetch also began downloading unrelated historical nested submodule commits. That recursive work exceeded 90 seconds and was stopped after the requested audit ref had arrived. Current pinned cohere was already present; no historical submodule checkout was used. This cost time but did not bound any test matrix result.

A first coverage-summary write used the read-only sandbox and failed before changing anything; the authorized write retry succeeded. No requested action remains blocked. Test output is preserved as raw logs, explicitly force-added because repository ignore rules normally exclude logs.

The target row's name matches its assertions about target parsing classification. It does not check exact usage text or successful target compilation; the defense does not infer those claims. Package uniqueness is observed for D1 on current main only, with no repo-wide uniqueness claim. No tests were deleted, rewritten or weakened and no pull request was opened.
