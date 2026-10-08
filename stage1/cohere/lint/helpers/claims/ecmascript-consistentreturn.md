# ecmascript/consistentreturn

Branch: lint-helpers/ecmascript-consistentreturn. Base: area/stage1-lint 334509ee.
Triage origin/lint-helpers/triage d7ab0bc4, rank 26, thirteen missing helpers.
Reserve HasValue, IsGenerator, IsScope, Judge, Name, ReportRange, Verb,
canRunOffEnd, capitaliseFirst, isExemptFromEndJudgment, judgeReturns,
judgeScope, staticName. Rules: consistent-return and
@typescript-eslint/consistent-return, conditional on control_flow_graph and
property. Alone zero; frozen cumulative readiness 135 with prior packages.

All origin lint-helpers/* and codex/lint-helpers* claims checked after fetch.
Earlier eligible packages are reserved or delivered; imports/jsx/property/
structure/module/react excluded as tonight landings. Runtime regex packages
regexp, tailwind class literals and dotnotation skipped. Directives is already
ported by suppression outside the helper catalog. Decorators has earlier claim
b9fa10f4 (2026-10-08T12:48:53Z); our unpublished local claim was withdrawn in a
follow-up deletion commit, without overwriting the owner's branch.

No retained partial consistentreturn port listed in triage. canRunOffEnd calls
control_flow_graph.Build (judgment.go:356), whose complete graph package is
unported and reserved by lint-helpers/ecmascript-control-flow-graph. Stop on that
helper and its callers unless the prerequisite lands; build independent helpers
first. No substitute reachability algorithm or Go-answer callback in production.
Property.Name is already on area. Typed extension hooks remain explicit inputs.

## Delivered independent subunit / dependency stop

Nine helpers: HasValue, IsGenerator, IsScope, Name, ReportRange, Verb,
capitaliseFirst, isExemptFromEndJudgment, staticName. Symbol/file map and exact
scope: ../consistentreturn/README.md. Real pinned-Go captures: 131 distinct
upstream cases (75 core / 56 extension), 3,136 consumer calls and 3,918 total
observations with controls. All nine semantic mutants compile and are caught
on source Node, emitted JavaScript and ASan/UBSan native.

Stop on canRunOffEnd (Go judgment.go:356, control_flow_graph.Build), judgeReturns,
judgeScope and Judge. The prerequisite package is unported and reserved; do not
release this reservation or claim complete rule readiness. No proof-rule
registration was created: core and extension consistent-return remain blocked.
Newly fully unblocked rules: zero. No compiler change, substitute reachability
or recorded Go-answer dependency is supplied. Existing property.Name is reused;
there is no partial consistentreturn source branch to credit.

Next.js filename-preserving witness need remains in ecmascript-nextjs.md on
lint-helpers/ecmascript-nextjs, notes commit 9da5e0d5; no-head-import-in-document
remains reserved by its existing owner.
