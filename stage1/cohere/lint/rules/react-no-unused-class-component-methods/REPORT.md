# react/no-unused-class-component-methods

Port of the unchanged Go rule at cohere `7945d102a6c18dd36adf9114a758ce646e8b2359`,
on Adamic base `9156bf5c579a44d687c9955d13e44f9ad8bbb6f8` from
`origin/area/stage1-lint`. The remote descriptor census found no existing port;
`react/require-optimization` provided a positive control for the same path pattern.
Only this rule directory is committed. No shared registration, compiler, runtime,
parser, helper or upstream rule body changes are required.

## Observed parity

[evidence/rule.txt](evidence/rule.txt) records 133 unique captured Go
source/file/options combinations and four owned witnesses, compared byte-for-byte
on source Node, emitted JavaScript and ASan/UBSan native: 63,098 output bytes.
This includes messages, byte ranges, fix and suggestion records, and fixed source.
The rule has no options, fixes or suggestions in Go; its options adapter returns
nil, as the upstream registration has no Decode function.

The witnesses are [class.tsx.txt](testdata/class.tsx.txt),
[object.ts.txt](testdata/object.ts.txt), [nested.ts.txt](testdata/nested.ts.txt),
and [keys.ts.txt](testdata/keys.ts.txt). They cover JSX reads, static-member
exclusion, compound-assignment definitions, a specification in the second
argument slot, renamed destructuring, lifecycle differences, nested classes,
computed keys, numeric keys, private-name exclusion, accessors, and quoted names.
Every witness causes Go findings. The standard owned-witness test also compares
selected and all-rule runs.

[evidence/comparison.log](evidence/comparison.log) records passing
TestOwnedWitnesses and TestRulesAgree. The wider capture contains 5,338 unique
combinations. Its main syntax/corner-case comparison matches 14,260,150 bytes;
typed cases also pass their live-native and JavaScript replay comparisons.

## Oracle behavior retained

TestNoUnusedClassComponentMethodsNestedClass deliberately differs from ESLint:
Go judges both the outer and inner component instead of losing the outer state
on entry to a nested class. This port matches Go, including its recursive
collection of accesses inside nested classes. The nested witness pins it.
TestNoUnusedClassComponentMethodsDefinitionVersusUse treats compound assignment
as a definition and increment as a use. The port preserves that distinction.
Static top-level class members contribute neither definitions nor uses. A bare
createReactClass callee accepts an object in any argument slot. Class expressions
and computed identifier keys remain invisible. No behavior is widened from Go.

The existing certified React projection, component-class/base helpers and
rules/react.isThisExpression are imported. Message quoting uses the existing
Go-generated printable-code-point table without the comment helper's truncation
or whitespace folding. Gather state is local to one component and retains
numeric node identities rather than a cyclic AST.

## Mutant

[mutant.json](mutant.json) makes assignment definitions count as uses instead.
[evidence/rule.txt](evidence/rule.txt) records that it
compiled, ran and disagreed with Go on all three backends. The first mismatch is
the captured constructor assignment to foo in ClassAssignPropertyInMethodTest.
This is an output mismatch, not a compiler refusal or sanitizer termination.
The standard package mutant additionally runs the inherited generated corpus.

## Commands and setup

Source `/workspace/adamic-tools/env.sh` in each shell. Both setup runs used
`GOPROXY='https://proxy.golang.org|direct'`. [setup.log](evidence/setup.log) records
501.421 seconds, 5 processors, cgroup quota 4 CPUs, and 17.6 GB. The second run
[with WASI SDK 27](evidence/wasi-setup.log) took 323.283 seconds including its
wait for the first installer's lock.

- `go run ./cmd/lint-registry` passed.
- `go test ./stage1/cohere/lint -run '^(TestRulesAgree|TestOwnedWitnesses)$' -count=1 -v -timeout=30m` passed on retry.
- The first comparison build succeeded but the harness rejected first-use Go
  module-download lines on stderr. [initial-comparison.log.gz](evidence/initial-comparison.log.gz)
  preserves that failure; warming the downloads resolved it without a source edit.
- `go test -overlay=/workspace/scratch/react-no-unused-class-component-methods/overlay.json ./stage1/cohere/lint -run '^TestReactNoUnusedClassComponentMethodsPort$' -count=1 -v -timeout=30m` passed.

[testdata/selected.go.txt](testdata/selected.go.txt) is the rule-specific test
fragment appended to a scratch copy of harness_test.go through a Go overlay.
It calls t.Parallel, uses the existing capture/comparison/build helpers, filters
this rule's captured cases, and checks its mutant natively as well as on both
JavaScript paths. The repository's harness is unchanged.

The whole package uses a clean TypeScript checkout at
`050880ce59e30b356b686bd3144efe24f875ebc8`, WASI_SYSROOT, ADAMIC_LINT_BENCH=1,
and one fresh directory for ADAMIC_LINT_PROFILE_DIR and
ADAMIC_LINT_PROFILE_SNAPSHOTS. Its command and machine measurements are recorded
in evidence/full-summary.json; the complete output is evidence/full.jsonl.gz.

The first full-package attempt was interrupted by an environment reconnect,
before a package exit status. It is preserved in
[evidence/interrupted-full.json](evidence/interrupted-full.json) and
[evidence/interrupted-full.jsonl.gz](evidence/interrupted-full.jsonl.gz), and is
not counted as a completed gate. The retry uses a detached runner and a new
fresh profile directory.

## Completed full-package result

The completed retry returned exit 0: 152 passing test events, 0 failures, and
1 skip, including subtests. Top-level counts are {"pass": 46, "fail": 0, "skip": 1}.
Wall time was 1,632.112 seconds (Go package time 1,626.521 seconds). nproc was 5;
load before was `3.78 3.84 3.90 2/188 113573`, and load after was
`4.92 5.50 5.29 2/170 248499`.

The sole skip is TestCheckerBridgeRefusalPending. The pinned prelude does not
declare TSGoError, so its existing test awaits codex/tsgo-errors-as-values. All
optional corpus, WASI, benchmark and profile inputs were supplied; none of their
tests skipped. No missing helper or Adamic language gap blocked this rule.

[evidence/full-summary.json](evidence/full-summary.json) records the exact
command, input paths, counts and completed test names.
[evidence/full.jsonl.gz](evidence/full.jsonl.gz) preserves every JSON test event.
The full mutant sweep includes this rule on its witnesses and the inherited
generated corpus; the additional scoped test proves its native mutant too.

The readable scoped log is an excerpt retaining phase timings, parity, mutant
mismatches and the test result. [evidence/rule.log.gz](evidence/rule.log.gz)
preserves its full original bytes, including the intentionally space-bearing
fixed-source records. The initial-download failure log is also compressed
without changing its bytes.
