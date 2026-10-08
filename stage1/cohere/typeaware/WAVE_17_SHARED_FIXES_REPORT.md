Merged both shared fixes into codex/typeaware-wave-17, without rebasing.
Source merge tip f22028f934458fb83ae2ef9877580bcef36e1d7c includes JSX inventory
e7c196a9755cbba48809859033158bc9b06f3c7b and parser recovery
dbe2fff02270fe9356ee560a699e706a8b444105. No production source changed after
those merges: this commit adds only owned validation evidence and this report.
No rule or checker has been claimed in this unit. The area remains the one
production checker; native records its live facts and Node/JavaScript replay them.

Full lint package: 31 pass, 4 fail,
1 skip (top-level). Including subtests: {'pass': 130, 'skip': 1, 'fail': 4}.
Wall 3427.765s; nproc 5; load start [7.376953125, 2.24609375, 1.3310546875],
end [1.82861328125, 1.76611328125, 3.05859375]; one-minute min/median/max [1.0322265625, 4.248046875, 13.591796875].
Setup succeeded in 587.091s, including the shared toolchain-lock wait.
Command: go test -json -count=1 -timeout=90m ./stage1/cohere/lint.
Every input, exact environment and pinned source are in
validation-wave-17-shared-fixes/inputs.json. TypeScript compiler, benchmark and
both profiling inputs were provided. Registry, gofmt, vet and diff checks pass.

Owned agreement against Go (findings, fixes and suggestions byte-identical on
Node, emitted JavaScript and sanitized native):

| Rule | Unique upstream source/rule/options cases | Result |
| --- | ---: | --- |
| @next/next/no-async-client-component | 43 | green |
| @next/next/no-duplicate-head | 27 | green |
| @next/next/no-script-component-in-head | 23 | green |
| nexus/correctness-no-collection-misuse | 6 | green |
| nexus/correctness-no-discarded-pure-result | 22 | green |
| no-global-assign | 53 | green, including separate no-checker guard |
| no-throw-literal | 48 | green, including separate no-checker guard |
| no-useless-backreference | 408 | green |
| react-hooks/unsupported-syntax | 61 | green |
| react-hooks/use-memo | 79 | green |
| prefer-arrow-callback | 83 | 81 agree; two shared fix-validation failures below |
| nexus/correctness-no-discarded-outcome | 31 | full parity blocked: all 31 have companion files |
| no-implicit-globals | 148 | full parity blocked: original JavaScript/script configuration is lost |
| no-implied-eval | 197 | full parity blocked: 170 have companion files |

Ten rules are fully green on 770 unique captured combinations. The shared
TestRulesAgree stopped at its first JavaScript project error, so independent
validation checked all 853 combinations of the eleven dependency-independent
rules with faithful default upstream projects. Of these, 851 agreed and the two
repair failures did not. Supplemental scripts, per-case outputs, durations and
capture audits are in owned-replay/ and owned-replay-six/ under the evidence dir.
The capture audits compare original finding IDs/counts as well as the full Go
wire-output comparison. Their only mismatches are the two intentional no-checker
guards; guard-audit-explanations.json and each no-program-guard/result.json retain
separate four-engine checks in the correct untyped context.
Malformed upstream inputs use the existing recovery protocol and retain their
original Go diagnostic refusals; none was discarded to make a comparison pass.
Supplemental process totals: Go 30.782s; sanitized native including live-program
creation and fact recording 52.794s (1.72x Go). These are observations
under concurrent gate load, not an isolated performance benchmark.

All 14 owned firing witnesses pass TestOwnedWitnesses. All 14
owned_judgment_suppressed mutations are caught on the three runtimes; all 90
registered rule mutants pass TestMutants. The compilation/counting control
reports allocations 328, frees 328 and regions 0. No sanitizer failure appeared.

Remaining failures and blockers (requested classification):

| Test / rule | Cause and exact location | Small verified input / evidence path | Category |
| --- | --- | --- | --- |
| TestJsxLintReleaseAndThroughput | jsx_integration_test.go:83; nexus-consistency-no-single-line-jsdoc/comment_anchors.a:67 does not anchor inside empty JSX expressions; Go comments.collectListInteriors, cohere/internal/lint/ecmascript/comments/comments.go:291-292, calls collectAt(node.Pos()+1) | repro-jsdoc/source.tsx.txt: `<>{/**x*/}</>`; Go reports useSimpleComment, Node reports none | yours: inherited branch parity outside owned rule directories |
| TestRulesAgree / no-implicit-globals | lint_test.go:376-382 rebuilds strict-only projects without allowJs; upstream runImplicitGlobalsAsAScript at cohere/internal/lint/rules/core/no_implicit_globals_test.go:34 configures allowJs and moduleDetection:auto | repro-project/source.js.txt: `x=1;`; strict-only Go exits 2, faithful config exits 0 | shared project replay |
| TestCompilerAndStage1Agree | lint_test.go:467 discovers raw recovery input stage1/typescript/parser/testdata/lint_cases/decorated_async_promise_executor.ts:1 as normal corpus and Go refuses it | repro-corpus/source.ts.txt: `new Promise(@dec async () => {})` | yours: inherited corpus integration outside owned rule directories |
| TestProfileSnapshotsAgree | profile_test.go:225; stage1/typescript/parser/parser.ts:1792 misses top-level await before NewKeyword, passing a separate NewExpression statement to no-new/rule.ts:15; Go cohere/internal/lint/rules/core/no_new.go:68 receives an AwaitExpression instead | repro-await-new/smaller/source.ts.txt: `await new A;export{}` | yours: inherited parser gap outside owned rule directories |
| Supplemental TestPreferArrowCallbackFires cases 470 and 495 | cohere/internal/lint/rules/core/prefer_arrow_callback_test.go:73,95; findings and edits agree, but shared Linter.fixed at lint.ts:198 reparses without rejecting syntax diagnostics. Go cohere/internal/edit/engine.go:182-194 refuses the whole malformed pass | repro-repair/0/source.ts.txt: `f(x||function(){}.bind(this))`; repro-repair/1/source.ts.txt: `f(function(){}.bind(this).bind(x))`; all three runtimes differ from Go on applied source/rejections | yours: shared fix pipeline outside owned rule directories |
| no-implied-eval project replay | shared_test.go:254-257 record drops companions; upstream runImpliedEvalWithGlobals at cohere/internal/lint/rules/core/no_implied_eval_test.go:79-85 provides globals.d.ts | repro-implied-eval/source.ts.txt: `setTimeout('');`; shared single-file project reports 0, upstream ambient declarations report 1 | shared project replay |
| nexus/correctness-no-discarded-outcome project replay | same record loses 4-5 companions; cohere/internal/lint/rules/nexus/correctness_no_discarded_outcome_test.go:83 adds them | previous retained validation-wave-17-unpark/project-replay-gap/README.md, original-capture.json and faithful/ exact files; faithful 1, subject-only 0 | shared project replay |
| TestCheckerBridgeRefusalPending (sole skip) | checker_pending_test.go:49 awaits tsgoInspect returning TSGoError from the C error buffer | pending-refusal-control checker question on the typed pilot witness | bridge errors |
| Parked react/boolean-prop-naming (not a registered package failure) | cohere/internal/lint/rules/react/boolean_prop_naming.go:182 regexp.Compile(settings.Rule) needs a runtime-compiled option pattern; static literals cannot substitute | prior validation-wave-17-unpark/regexp-options.a.txt and regexp-options.log | dynamic RegExp |

Relative repro paths above are inside validation-wave-17-shared-fixes/.
All lint paths are relative to stage1/cohere/lint/ unless fully qualified.
The two hook rules do not require React IR; no owned rule waits on #2kb2kje.
"Yours" above identifies inherited branch parity under the supplied categories;
these locations are shared files or another worker's rule, not an owned judgment.
No shared code, fixture inventory, check or refusal was changed or bypassed.
No new claims were made, and no full repository gate or fresh standalone legacy
suite is claimed here. Full-package green remains blocked by the named failures.
