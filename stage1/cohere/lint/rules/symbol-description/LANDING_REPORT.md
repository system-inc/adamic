# Narrow wave 24 landing

The original 593a483ee16aeb54d61b82c3b97cc44037ce5b54 tree contained old wave ports and corpus fixtures, although only symbol-description and valid-typeof were registered as additions to the unified harness. The landing tree now restores the area's other files and removes the legacy-only tree. All legacy source, assertions and evidence remain reachable through that pushed commit's Git history. No registry guard, comparator or another owner's rule is changed.

## Original failure

TestCompilerAndStage1Agree, stage1/cohere/lint/lint_test.go:467, reported exactly:

```text
    lint_test.go:467: Node: case 803 line 134154: port "fixed\texport class Foo { f(): Foo | undefined { if(Math.random()) return; return this; } g(): Foo {const self=this;return self;} }\\u000aexport {};\\u000a", Go "/workspace/adamic/stage1/cohere/typeaware/validation-wave-24/control-002.a:1:100"
```

The extra tracked legacy input was stage1/cohere/typeaware/validation-wave-24/control-002.a. Cohere reports thisAssignment for the self binding; the area rule excludes .a at stage1/cohere/lint/rules/typescript-no-this-alias/rule.a:14. Cohere's NoThisAlias reports at cohere/internal/lint/rules/typescript/no_this_alias.go:154; its file predicate in no_explicit_any.go:160 uses sourcename.TreatedAs, which includes .a. The original event log and the isolated source/output remain in evidence/landing. Raw reproducer text uses .a.txt here; it is evidence, not an imported module or a new active corpus source.

## Landing scope

Only two registered directories differ from origin/area/stage1-lint c4bdc23fa86d55cf7e579989201c11258f4d3a62: rules/symbol-description and rules/valid-typeof. Outside them, only bridge/tsgo/checker/facts.go, program.go, symbol_identity.go and symbol_provenance.go differ. These implement the one borrowed symbol-provenance fact both rules consume. The C archive again uses the area's ordinary Program.Inspect entry; all ten legacy question handlers, legacy suites and private Rules/RuleContext adapters are absent from the landing diff. Existing area typeaware files are unchanged.

The normal CLI, not only the harness's compiler path, compiled successfully:

```sh
go build -o /tmp/wave24-narrow-adamic ./cmd/adamic
go build -buildmode=c-archive -o /tmp/wave24-narrow-checker.a ./bridge/tsgo/archive
/tmp/wave24-narrow-adamic build stage1/cohere/lint/main.ts -o /tmp/wave24-narrow-lint-native --tsgo /tmp/wave24-narrow-checker.a
```

Descriptor generation, gofmt and vet pass. The two rules retain their Go options adapters, all upstream cases, witnesses, suggestions and caught mutants. Fresh full-package and area-only results are recorded below when complete. Historical broader-tree reports in evidence/landing/previous-report.md and lint-summary.json are explicitly not the current landing result.

## Untouched area probe

The independent .a probe also fails against the untouched area tree c4bdc23fa86d55cf7e579989201c11258f4d3a62: selected @typescript-eslint/no-this-alias on const self = this; export {}; yields live Go 556 bytes including thisAssignment and Node 54 bytes without it. Both exit zero; their complete bytes differ. The test worktree is /workspace/wave24-area-check and stays clean. No wave descriptor or bridge change is present there. Its ordinary TestCompilerAndStage1Agree corpus is 872 files; the historical wave corpus was 1449 files, including the failing legacy .a fixture. The fresh narrowed corpus is 876 files. The ordinary area test's completed verdict is recorded below separately from this probe.

## Area-only standard test result

From untouched origin/area/stage1-lint c4bdc23fa86d55cf7e579989201c11258f4d3a62, with the same clean TypeScript corpus pin:

```sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-24-corpus GOMAXPROCS=4 go test -json -count=1 -timeout 30m ./stage1/cohere/lint -run '^TestCompilerAndStage1Agree$'
```

PASS: one test passed, zero failed, zero skipped; package 685.578 seconds. The area worktree remains clean. This standard test does not fail on area alone. The isolated .a alias does reproduce against the unchanged area implementation, as separately recorded above, so the extension-guard defect is an area rule issue exposed by the historical wave's additional legacy corpus. Narrowing the branch to the two requested directories removes that out-of-scope legacy corpus without changing the rule or comparison. The original fixture's bytes and outputs remain archived as raw evidence, and all original legacy tests remain in the old pushed commit history. The narrowed branch's full run also passes TestCompilerAndStage1Agree on its 876-file corpus.

## Final narrowed-tree package result

Command: go test -json -count=1 -timeout 60m ./stage1/cohere/lint, after sourcing the installed cloud env, GOMAXPROCS=4, ADAMIC_LINT_BENCH=1, ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-24-corpus at clean 050880ce59e30b356b686bd3144efe24f875ebc8; both profile inputs /workspace/wave24-narrow-profile. Exit zero. Package PASS, 1641.150 seconds; outer wall 1652.495783981 seconds, nproc 5.

Including subtests: 115 passed, zero failed, one skipped. Top level: 34 passed, zero failed, one skipped. All 77 registered mutants caught, including symbol-description-global and valid-typeof-options. TestRulesAgree, TestCompilerAndStage1Agree, TestOwnedWitnesses and all profile/option/cache controls pass on Node, emitted JavaScript and sanitized native. The ordinary CLI build also succeeds; there is no registered wave addition refused by that CLI. The only skip is TestCheckerBridgeRefusalPending, explicitly awaiting tsgoInspect's C error-buffer TSGoError result; no input-controlled test skips. Current logs/counts are narrow-lint.jsonl.gz and narrow-lint-summary.json.

The .a extension-guard issue in the unchanged area rule remains separately reproduced for integration to route; it is not silently repaired or relaxed here. The old larger landing tree failed its corpus check; its blocked legacy additions are excluded from this two-rule landing and retained in existing Git history. No history is rewritten, and no main or area branch is pushed.
