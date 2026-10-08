# Wave 05 checker landing audit

Merged origin/area/stage1-lint 9b7547976bea1b00f8b04508b9102d03511f65b3 into codex/lint-wave1-05; no rebase. The wave1-05 claim file assigns syntax-only rules, not type-aware rules. No unassigned typed rule is claimed or ported. Existing syntax ports retain their descriptors and use the shared harness.

## Still blocked: boundaries/dependencies

At pinned cohere 7945d102a6c18dd36adf9114a758ce646e8b2359, internal/lint/rules/boundaries/dependencies.go:553 calls ctx.Program.GetCurrentDirectory(), :590 calls ctx.Program.ResolveModule(ctx.SourceFile, specifier), and :591/:594 read resolved.IsExternalLibraryImport and resolved.ResolvedFileName. bridge/tsgo/checker/facts.go exposes neither directory nor module-resolution questions. Its options question at :240 only returns StrictNullChecks; it cannot replace these Program reads. RuleContext.checker alone therefore does not unblock this rule. No private checker, resolver or shared-helper copy was added. The incomplete implementation remains outside the registry in parked/wave05/boundaries-dependencies.

Reproducer: Go TestDependenciesProjectModuleBoundary (cohere/internal/lint/rules/boundaries/dependencies_test.go:184) resolves a project import before evaluating the policy. From cohere: go test ./internal/lint/rules/boundaries -run '^TestDependenciesProjectModuleBoundary$' -count=1 -v. The complete port requires a shared recorded/replayed module-resolution answer. The old rule-owned decision-engine controls do not prove full findings parity.

Full-package gate uses pinned TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8, ADAMIC_LINT_BENCH=1 and fresh profile directories. Logs and exact counts are recorded after the gate.

## Owned parity after the area merge

The new Go cohere sourcename.TreatedAs gate checks Adamic .a sources as TypeScript. The owned structure/tailwind-no-physical-direction guard excluded .a; the first diagnostic run caught its missing finding. Added .a to that owned guard. A fresh TestRulesAgree passes all 4,235 captured source/rule/options combinations on Go, Node, emitted JavaScript and sanitized native. No new type-aware assignment was invented.

| Owned rule | Upstream combinations matched |
| --- | ---: |
| @eslint-community/eslint-comments/require-description | 109 |
| structure/tailwind-no-physical-direction | 56 |
| @typescript-eslint/no-non-null-asserted-optional-chain | 24 |
| @typescript-eslint/no-non-null-assertion | 36 |
| @typescript-eslint/prefer-function-type | 39 |
| @typescript-eslint/prefer-namespace-keyword | 22 |
| @typescript-eslint/triple-slash-reference | 70 |

## Remaining landing blockers

TestJsxLintReleaseAndThroughput and TestJsxLintTrees fail at the fixed-map assertion in jsx_integration_test.go:61. Their capture contains one additional JSX case for each owned comment-description and physical-direction rule. The five-rule expected map excludes those two registered rules. Reproduce with all normal inputs set and go test ./stage1/cohere/lint -run '^TestJsxLintTrees$' -count=1 -v. Upstream case capture and shared checks were neither narrowed nor edited. This branch is not full-package green.

On the helper merge branch, TestWave05LeadingInteger, TestWave05PackageRoot and TestWave05CompareBreakpoints fail before comparison: each rule-owned Python validator line 17 asserts an old macOS theme fixture path that the new cohere tests no longer contain. Reproduce with go test ./stage1/cohere/lint/helpers -run '^TestWave05' -count=1 -v. Foundation checks pass, but the three helpers are not re-certified on this pin. No obsolete anchor guard was deleted or relaxed.

TestCheckerBridgeRefusalPending skips for the inherited missing TSGoError bridge surface. The required compiler, benchmark and profile inputs are provided. No input skip is requested.

## Complete gate results

w05-rules-final: {'pass': 118, 'skip': 1, 'fail': 2}; 2531.367s wall; nproc 5; load first [2.505859375, 4.30859375, 5.21435546875], last [2.00732421875, 3.138671875, 4.044921875], peak [7.21826171875, 5.84423828125, 5.220703125]. See checker-landing-evidence/w05-rules-final.log.gz and w05-rules-final-time.json.

w05-helpers: {'pass': 113, 'skip': 1}; 3303.086s wall; nproc 5; load first [8.75634765625, 10.44970703125, 5.7060546875], last [2.466796875, 3.5712890625, 4.26318359375], peak [9.74853515625, 10.6005859375, 6.3369140625]. See checker-landing-evidence/w05-helpers.log.gz and w05-helpers-time.json.

Owned witnesses, all 82 registered mutants, TestRulesAgree and the 915-file compiler/stage1 comparison pass on the rule branch. All seven owned mutants compile and run on Node, emitted JavaScript and sanitized native; summary.json maps each to mutants.log line numbers. The only rule-branch failures are the two unmodified JSX count guards. The helper branch full lint package passes, but its separate helper package has the three named owned-validator failures. Neither branch is presented as wholly landing-ready.

Setup failed only during cache warming with No space left on device. The first broad compiler clone also exhausted disk. Removed only regenerable Go build cache, fetched the exact compiler commit shallowly into /tmp, and reran with every required input. The diagnostic rule run was interrupted after finding the .a regression; the final gate above is a fresh complete run. No shared check, refusal guard, input requirement, registration code or compiler implementation was edited.

The claim file contains no type-aware assignment; boundaries/dependencies remains blocked on the exact Program questions above. No new rule or helper was claimed. No push to main or area. Full repository correctness checks outside lint were not run.
