# react/no-typos

This syntax-only port is confined to this directory on `lint-rules/react-no-typos`,
based on `origin/area/stage1-lint` at `9156bf5c579a44d687c9955d13e44f9ad8bbb6f8`.
All remote branches were fetched and searched for the exact registered name before
porting; none held it. Cohere remains unmodified at
`7945d102a6c18dd36adf9114a758ce646e8b2359`.

The rule preserves Go's source-order import state, static configuration casing,
lifecycle casing and static requirements, validator names and qualifiers, recursive
shape/oneOfType validation, declaration search, spans, messages and report order.
It has no options, fixes or suggestions. The oracle adapter returns the unchanged
Go rule and nil options, matching its upstream registration.

Shared helpers supply imports.BindingsOf, React component-class recognition,
parenthesized JSX recognition, Unicode EqualFold and assignment classification.
The import projection is assembled locally and published readonly. Context storage
starts as undefined and is checked through a method after constructor assignment.
These use ordinary supported Adamic and do not change the upstream algorithm.
No shared helper, compiler file or other rule is edited.

## Oracle behavior retained

Go's surprising behavior is preserved: the member arm reports
`target = Example.PropTypes` although it is a read; compound assignments report;
module helpers `checkPropTypes`, `resetWarningCache` and `PropTypes` count as
accepted validator names; a qualifier after an unrelated call can report once
either package is imported; imports below a use do not retroactively bind it.
See upstream `TestNoTyposMemberArmAssignmentShapes`,
`TestNoTyposAcceptsEveryPropTypesModuleKey` and
`TestNoTyposImportBindingsGateThePropTypeArms`.

The source-file walk, declaration search, import flattening and validator-chain
recursion are rule-local operations present in the Go source. Shared helpers are
used rather than privately copied. No missing-helper or rule-local language gap
is claimed. Go's verdicts are preserved.

## Validation

The Go capture overlay records all runs under `TestNoTypos`, deduplicating
source/file/rule/options combinations: **202 captured cases** for this rule.
There are **three owned witnesses** in [testdata](testdata/): class and assignment
surfaces, createReactClass with recursive validators, and JSX-returning functions
with side-effect imports and a nested-function negative control. Each witness
causes a Go finding. No .options.json is needed because the rule has no options.

The owned `validation.go.txt` is injected as an additional lint-package test
through a Go overlay, without changing a shared test file. Both new top-level
tests call `t.Parallel()`. It selects this rule on the captured upstream cases,
owned witnesses and inherited source corpus, then compares against Go on source
Node, emitted JavaScript and sanitized native. A baseline control removes only
this registration in a disposable copy and confirms the base lowers successfully.
The casing-inequality mutant is compiled and run on all three port backends and
must disagree with Go. The package's normal owned-witness, upstream-agreement
and selected-mutant tests are also run.

To reproduce the extra owned checks, map the nonexistent
`stage1/cohere/lint/react_no_typos_validation_test.go` to the absolute path of
`rules/react-no-typos/validation.go.txt` in a Go overlay JSON, then run
`go test -overlay=<overlay> ./stage1/cohere/lint -run 'TestReactNoTypos' -count=1 -v`.
All test output goes to log files.

The selected run passed in **226.319s**, including the package's 5,407 captured
combinations, all owned witnesses, and the selected mutant. The extra owned
comparison verifies **202 cases plus inherited sources and three witnesses** on
all four implementations: **124,622 identical bytes**. Findings, messages,
byte ranges, fixed-source records and suggestion absence agree. The mutant is
caught on source Node, emitted JavaScript and sanitized native; its exact three
caught lines are in [evidence/mutant.log](evidence/mutant.log). Full selected
output is in [evidence/selected.log](evidence/selected.log).

Exact whole-package results are recorded in the evidence below.
The captured tests and witnesses bound these claims; they are not an exhaustive
proof over arbitrary malformed syntax or every possible file.

## Setup and inputs

Initial setup used `GOPROXY='https://proxy.golang.org|direct' bash cloud/setup.sh`:
Node 0.028s, Go 0.037s, clang 0.253s, markdown 0.874s, submodules 2.999s,
build cache 202.845s, total 202.879s, nproc **5**. WASI setup installed SDK 27 and
completed in 43.579s. Subsequent commands sourced
`/workspace/adamic-tools/env.sh`.

The registry generator passed. Whole-package inputs are a clean TypeScript
v6.0.3 checkout at `050880ce59e30b356b686bd3144efe24f875ebc8`, WASI SDK 27,
`ADAMIC_LINT_BENCH=1`, and one fresh directory shared by both profile variables.
The complete command, verified input paths/pin/clean status, pass/fail/skip counts,
wall time, nproc and load readings are recorded in `evidence/full-metrics.json`.
The complete JSON output is preserved as `evidence/full.jsonl.gz`.

## Whole-package result

**PASS**, exit 0: **152 passed, 0 failed, 1 skipped** test results, including
subtests. Top-level counts are 46 passed, 0 failed, 1 skipped. Wall time is
**1,158.610s**; the Go package reports 1,156.792s. nproc is **5**, with CPU quota
4 and GOMAXPROCS=4. Load averages before: **0.99 / 2.37 / 2.81**; after:
**4.05 / 4.40 / 3.77**.

The sole skip is **TestCheckerBridgeRefusalPending**. It is hard-coded when
`internal/load/prelude.d.ts` lacks `TSGoError`, awaiting
`codex/tsgo-errors-as-values`; see `stage1/cohere/lint/checker_pending_test.go:52`.
All requested inputs were provided, so there are no input-based skips. Achieving
zero skips requires that separate base feature, outside this rule directory.
No rule helper or language blocker remains.

[Full metrics](evidence/full-metrics.json) preserve the exact command and inputs;
[full JSON output](evidence/full.jsonl.gz) preserves every test result.
[Registry output](evidence/registry.log),
[initial setup](evidence/react-no-typos-setup.log), and
[WASI setup](evidence/react-no-typos-wasi-setup.log) preserve the other commands
and timing lines. No PR is opened.
