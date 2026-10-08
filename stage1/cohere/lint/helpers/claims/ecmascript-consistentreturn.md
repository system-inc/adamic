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
control_flow_graph.Build (judgment.go:366), whose complete graph package is
unported and reserved by lint-helpers/ecmascript-control-flow-graph. Stop on that
helper and its callers unless the prerequisite lands; build independent helpers
first. No substitute reachability algorithm or Go-answer callback in production.
Property.Name is already on area. Typed extension hooks remain explicit inputs.
