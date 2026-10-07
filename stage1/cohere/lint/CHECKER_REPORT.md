# Checker harness work in progress

Branch: lint-checker/harness, based on origin/area/stage1-lint at
65017b318da1995237ff3ea2c80f59b055b39ac3. That base includes main's pin bump.
No library branch was merged. This branch is not landing-ready.

The manifest program line, one run-owned program, RuleContext.checker, typed
selection, no-program coverage, checker-free fix re-linting, programReads
validation, per-rule read guards, native recording code, shared replay code and
the Go oracle's cohere program are written. The program uses the tsconfig's own
root list, without overriding it with the rows selected for linting. One Checker
per row holds the program handle and exact UTF-16-to-UTF-8 selector mapping.

FileQuestion carries a ProgramReads kind and the bridge question text. The
current pilot uses ReadsCompilerOptions/options and type-shape. Answers carry a
value or refusal reason. A recording prefix has a .header file naming the
program path and tsconfig SHA-256; prefix.N names each row, whose header binds
its path and source SHA-256. Each entry frames the exact selector/question key,
Value or Error, and its payload. Replay checks both missing and unasked entries.

The pilot is @typescript-eslint/no-unnecessary-boolean-literal-compare, moved
from the existing P1 rule logic. It asks expression types and strictNullChecks,
and the control exercises its undeclared-program-read refusal. Its listener is
BinaryExpression and takes the node handed to it. Options and nullable repairs
are written but have not received the complete upstream corpus comparison.

## Observed blocker

Native lowering refuses the agreed release return shape:

```
adamic: stage1/cohere/lint/gaps/checker-error-result/release.a:3:10:
stage 0 can't lower a function returning void | TSGoError yet
```

Reproducer, from the repository root:

```
adamic build stage1/cohere/lint/gaps/checker-error-result/release.a \
  -o /tmp/checker-release --tsgo /workspace/checker-harness/checker.a
```

The archive was built successfully from the landed tree with
`go build -buildmode=c-archive -o /workspace/checker-harness/checker.a ./bridge/tsgo/archive`.
Type checking of the complete driver succeeds. The compiler was not changed to
work around the refusal, and the approved result shape was not replaced.

The underlying C tsgo_create, tsgo_inspect and tsgo_release already return
status and error buffers. Adamic's wrappers still panic. In particular,
`tsgoInspect` / `adamic_tsgo_inspect` must return TSGoError with tsgo_inspect's
error buffer text. The harness converts that result to Answer and the refused
wire line, and records an Error answer for replay. Open and release also have
forward-compatible result wrappers.

TestCheckerBridgeRefusalPending explicitly skips with
`awaits codex/tsgo-errors-as-values`. It becomes required when the prelude exposes
TSGoError, builds the native invalid-question control, and compares its recorded
refusal with Node replay. No skip was treated as a pass.

## Validation

Setup used GOPROXY=https://proxy.golang.org|direct. Its timing lines were:
node 0.043s, Go 0.043s, clang 0.327s, markdown dependencies 1.403s,
submodules 5.307s, Go build 358.883s, total 359.392s. nproc was 5, CPU quota 4.

The final filtered lint command used GOMAXPROCS=1, the external TypeScript
6.0.3 checkout at 050880ce59e30b356b686bd3144efe24f875ebc8, and the same fresh
/workspace/checker-harness/profile directory for ADAMIC_LINT_PROFILE_DIR and
ADAMIC_LINT_PROFILE_SNAPSHOTS:

```
go test -count=1 -v -timeout 10m ./stage1/cohere/lint \
  -run '^TestChecker(NoProgramCoverage|ReplayEntryControl|Hashes|BridgeRefusalPending)$'
```

Result: PASS, 10.956s, with the explicitly pending bridge control skipped.
The no-program mutant printed ignored instead of no program and was caught.
A removed live-bridge transcript entry emitted refused and failed Node's
ordinary comparison; restoring it matched Go's finding and applied fix.
An extra entry failed the reverse guard. Inverting the pilot's type verdict
was caught by findings comparison on Node. An undeclared ReadsOtherFiles ask
emitted refused. Empty, ASCII and non-BMP SHA-256 vectors matched Go on Node and
sanitized native code. Those native builds used the existing sanitizer helper
and completed without sanitizer or leak errors.

The replay control's facts came directly from the landed Go bridge's live
Inspect calls. They were not handwritten, but they were not recorded by the
blocked native driver either. This distinction is material.

`go test -count=1 -v -timeout 10m ./stage1/cohere/lint/registry` passed in 2.122s.
`TestRulesAgree` failed in 21.306s at the release-result lowering refusal.
The full lint package and full repository gate were not run to green.

## Unfinished work

Native recording and full-driver sanitizers, emitted-JavaScript replay,
program-header mismatch controls, the complete typed corpus and companion
witness-project integration, shared profile compilation, and single-threaded
pilot instruction/time measurements remain unverified or unfinished because
native lowering blocks the driver. No instruction count or performance claim
is made. The existing TestRulesAgree build helper is checker-enabled; profile
helpers still need their checker-enabled build integration. Directory question
registration and fleet migration remain outside this unit.
