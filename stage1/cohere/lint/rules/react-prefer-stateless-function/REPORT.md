# react/prefer-stateless-function

Ported from cohere `7945d102a6c18dd36adf9114a758ce646e8b2359`, on area base
`6bf7bcec73a04b85900c5a69a257b3cf34b35a40`. Only this rule directory changes.
The oracle adapter returns the unchanged upstream rule with typed options and
explicit defaults. It accepts the harness's shared all-rule options bag, as the
other adapters do; it does not run the wire decoder against captured struct keys.

The port reuses the certified ES6 and ES5 component predicates, component-base
and identifier predicates, pure-component predicate, optional-parentheses helper
and function-kind helper. The remaining traversal and disqualifiers are rule-local. Reporting uses the
owned `finish` hook to preserve Go's tied finding order: after
`max-classes-per-file`, before `react/require-optimization`.
No fixes or suggestions are produced, matching Go.

## Parity and witnesses

All **91 captured upstream source/rule/options combinations** and **seven owned
witnesses** matched unchanged Go on source Node, emitted JavaScript and
ASan/UBSan native: 50,193 identical output bytes. See [selected.log](evidence/selected.log).
The selected probe uses the existing lint harness through a temporary Go overlay;
its top-level test calls `t.Parallel`. The ordinary package gate discovers this
rule and captures its upstream tests without any shared-list edits.

Witnesses cover props access, a factory call, a redundant constructor,
`ignorePureComponents` on a regular Component, optional and definite-assignment
typed props, spread members, and factory callee names. The last three boundary
checks were also probed directly against Go before correction; the old Node
implementation disagreed and the final three-backend comparison agrees.

The registered mutant removes `props` from the exempt instance names. It compiles
and runs, and Go comparison catches the missing finding on both Node and emitted
JavaScript. Its catch lines are in [selected.log](evidence/selected.log).

## Upstream observations and limits

Go disqualifies every candidate when any JSX `ref` occurs in the file. It compares
external `childContextTypes` receivers directly to a named class, without alias
resolution. A factory call has no component name; the callee's name is not a
substitute. Spread assignments have no property name. Typed `props` remain allowed
with `?` and `!` markers. These are preserved rather than broadened.

This is a syntax-only port. It does not add parser recovery or checker error
handling, and makes no new arbitrary-input proof claim beyond the captured cases,
owned witnesses and inherited corpus checked by the package gate.

## Setup and commands

`GOPROXY='https://proxy.golang.org|direct' bash cloud/setup.sh` succeeded.
Timing lines: Node ready 0.018s; Go ready 0.019s; submodules ready 0.060s;
markdown dependencies ready 0.070s; clang ready 0.148s; Go build ready 34.111s;
test binaries deferred 34.283s; cache warm 34.284s; done 34.311s.
The additional `bash cloud/setup.sh --wasi-sdk` succeeded: SDK ready 11.914s;
done 62.264s. Environment: `/workspace/adamic-tools/env.sh`.
Go 1.27.1, Node 24.19.0, clang 20.1.8, WASI SDK 27; nproc 5, quota 4 CPUs.

Registration passed with `go run ./cmd/lint-registry`.
Selected command (87.356s, pass), including all owned witnesses in selected and all-rule modes:

```sh
go test -overlay=/tmp/s13-react-prefer-stateless-function-overlay.json ./stage1/cohere/lint -run '^TestPreferStatelessFunctionSelected$|^TestOwnedWitnesses$|^TestMutants$/prefer_stateless_function_props_disqualifies$' -count=1 -v -timeout=30m
```

All full runs use a clean original TypeScript checkout at
`050880ce59e30b356b686bd3144efe24f875ebc8`, `ADAMIC_LINT_BENCH=1`, the installed
WASI SDK, and one fresh directory shared by `ADAMIC_LINT_PROFILE_DIR` and
`ADAMIC_LINT_PROFILE_SNAPSHOTS` per run. Test output goes directly to a log.
The initial run passed 152 test entries, failed 0, skipped 1, wall 1087.427s.
The next run after boundary fixes found a tied finding-order mismatch in the
retained two-class typed-props witness (151 pass, 1 fail, 1 skip). The selected
rule's findings were correct, but the all-rule order differed. Moving only this
rule's reports to its existing owned finish hook fixed it; all owned witnesses
then matched Go on all three backends (458,213 identical output bytes). The
final full run follows that correction; its results follow.

```sh
go test ./stage1/cohere/lint -count=1 -json -timeout=30m
```

The inherited corpus guard also requires a clean Git index. Staging this new
rule during the full run caused `TestCompilerAndStage1Agree` to reject the new
`rule.a` and `messages.a` index entries before comparison. Only this directory
was unstaged, with no source changes, and the failed test was rerun:

```sh
ADAMIC_TYPESCRIPT_SOURCE=/tmp/s13-react-prefer-stateless-function-typescript go test ./stage1/cohere/lint -run '^TestCompilerAndStage1Agree$' -count=1 -v -timeout=30m
```

That rerun passed in 243.268s: Go, Node, emitted JavaScript and ASan/UBSan native
matched 31,428,104 bytes. See [corpus.log](evidence/corpus.log). Raw whole-run
counts and resolved results are recorded separately in the gate summary.

Final whole-run results: 151 pass, 1 fail, 1 skip, wall 1116.387s.
The sole failure was the clean-index guard described above and passed on rerun.
Resolved results: 152 pass, 0 fail, 1 skip.
The intrinsic skip is `TestCheckerBridgeRefusalPending`: the shared prelude
lacks `TSGoError`. All optional inputs were set; no input-dependent test skipped.
Load before: 0.60 3.27 3.92; after: 5.83 5.93 4.92.
See [gate-summary.json](evidence/gate-summary.json) for raw and resolved counts.
No language gap or missing shared helper blocked this port.
