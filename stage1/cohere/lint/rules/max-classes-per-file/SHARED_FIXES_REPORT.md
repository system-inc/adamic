Merged both shared fixes into codex/lint-port-max-classes-per-file, without rebasing.
Source merge tip: fc9173e33f512d9501e1da574e8813cc82721c1a.
Included JSX inventory e7c196a9755cbba48809859033158bc9b06f3c7b and parser recovery dbe2fff02270fe9356ee560a699e706a8b444105.
Only this report and owned evidence are added after the merges; shared code is unchanged.

max-classes-per-file is green: all 29 unique upstream source/rule/options cases,
all three firing owned witnesses, and mutant one_excess_class_is_silently_allowed
caught across Node, emitted JavaScript and sanitized native. TestRulesAgree passed
its complete 4,038-case captured inventory. Registry, gofmt and vet pass.

Full command and every input are in evidence/shared-fixes/inputs.json.
Top-level tests: 33 pass, 2 fail, 1 skip. Including subtests: 118 pass, 2 fail, 1 skip.
All 76 registered rule mutants are caught. Wall time 2835.831s;
nproc 5; start load [7.376953125, 2.24609375, 1.3310546875]; end load [3.61279296875, 4.39697265625, 4.654296875];
1-minute load min/median/max [2.10107421875, 4.732421875, 13.591796875].
Setup succeeded in 534.001s. TypeScript compiler, benchmark and both profiling
inputs were provided; no optional-input test skipped.

Failures, classified in the requested categories:

| Test | Location and cause | Smallest verified input | Category |
| --- | --- | --- | --- |
| TestCompilerAndStage1Agree | stage1/cohere/lint/lint_test.go:467 discovers raw recovery input stage1/typescript/parser/testdata/lint_cases/decorated_async_promise_executor.ts:1 as normal corpus; Go refuses Expression expected and comma expected | evidence/shared-fixes/repro-corpus/source.ts.txt: `new Promise(@dec async () => {})` | yours: inherited corpus integration outside the owned rule directory |
| TestProfileSnapshotsAgree | stage1/cohere/lint/profile_test.go:225; parser.ts:1792 does not recognize top-level await before NewKeyword; no-new/rule.ts:15 consequently receives a separate new statement. Go no_new.go:68 skips it as an AwaitExpression | evidence/shared-fixes/repro-await-new/source.ts.txt: `await new A();export{};` | yours: inherited parser gap outside the owned rule directory |
| TestCheckerBridgeRefusalPending (skip) | stage1/cohere/lint/checker_pending_test.go:49 awaits TSGoError from tsgoInspect/C error buffer | the test's pending-refusal-control checker question on its typed witness | bridge errors |

The profile failure's original fixture is stage1/typescript/parser/testdata/lint_cases/wave13_top_level_await_new.ts:1.
Go's matching selector is cohere/internal/lint/rules/core/no_new.go:68; the
shared parser condition is stage1/typescript/parser/parser.ts:1792.
Reproducer stdout/stderr and full test logs are preserved at the paths above.
No shared helper, parser, gate, fixture or check was changed. These two failures
prevent claiming a green full package; they do not involve max-classes-per-file.
No dynamic RegExp, checker-question, React IR or Tailwind blocker exists for this rule.
