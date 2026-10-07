# Integrated lint area landing

Rebased codex/lint-wave1-14 onto area/stage1-lint
 d65a8f931c98655936ae04c6899f38f14862b73e, containing current main
39638d9e278d38bb5aeae887f46d55a70e47aaad. The remote tips were checked again before
pushing. The branch is not green for landing: two shared blockers are preserved
and reproduced below. No new claim is taken.

## Ledger and implementation

Read DEDUP_LEDGER.md before rebasing. Dropped the losing copies of
@typescript-eslint/no-unnecessary-type-constraint, @typescript-eslint/prefer-as-const
and @typescript-eslint/prefer-enum-initializers; wave1-08 owns their winners.
The remaining ten descriptor directories are retained. The former shared Finding
refusal is removed from these owned copies: fixes and ordered suggestions now
report through the landed reportRange and Suggestion model. The bridge lives in
the retained no-extra-non-null-assertion directory. The retired complete runner
is archived as .a.txt; old parking validators and rates are historical.

All ten rule.json descriptors declare named ast.Kind listeners and node:true.
The driver supplies ParseNode; relevance guards and initial node refetches are
removed. Repair positions also use that handed node. The module-alias adapter
now accepts other rules' option keys in the shared all-rule manifest, while exact
module-alias selections retain strict option decoding. No shared context, finding,
driver, registry generator, oracle, comparison or compiler implementation is edited.
No new Adamic source is .ts. The retained rules contain no Go regex to translate.

## Commands and observed results

Commands use source /workspace/adamic-tools/env.sh. All output is logged under
evidence/area-*.log, including failed checks.

- ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-14-typescript-pinned go test
  ./stage1/cohere/lint -run '^TestCompilerAndStage1Agree$' -count=1 -timeout 30m -v:
  PASS on final source, 368 compiler/stage1 files, 20,750,769 identical bytes across
  actual Go, source Node, emitted JavaScript and ASan/UBSan native (88.20s).
  The pinned TypeScript SHA is 050880ce59e30b356b686bd3144efe24f875ebc8.
- go test ./stage1/cohere/lint -run '^TestRulesAgree$' -count=1 -timeout 30m -v:
  FAIL at the modified-destructured parameter. All original Go assertions pass;
  2,558 source/rule/options combinations are captured. No fixture is excluded here.
- go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1 -timeout 30m -v:
  FAIL in the unchanged Go oracle on the existing no-lonely-if budget witness.
- go test ./stage1/cohere/lint -run '^TestMutants$' -count=1 -timeout 30m -v:
  24 of 25 registered subtests PASS. All nine other owned mutants pass on all
  three backends; no-lonely-if stops in Go before comparison on its budget witness.
- The owned landing_suite.go.txt probe uses the unchanged shared oracle, builds
  and comparison. It adds only a virtual test file in scratch, never modifying a
  live shared harness file. Original cases are grouped independently by owned
  rule; all captured cases remain, and recovery flags use recoveryRows exactly
  as the shared test does. Nine rules PASS (579 original cases); parameter-property
  assignment FAILS on its 92-case group, with the same parser reproducer.
  The default shared test remains red. No finding or repair is inferred from a
  partial group. Recovered numeric inputs check findings, not converged fixes.
- The owned no-lonely-if mutant probe PASSes over all 32 original cases, with
  source Node, emitted JavaScript and sanitized native each compiling/running
  cleanly and differing from actual Go. Its budget witness remains unchanged.
- go vet ./...: PASS. git diff --check: PASS.
- go test ./internal/oracle -run
  'TestNativeAgreesWithNode/internal/oracle/testdata/inherited_static_field_read|TestTheOracleCatchesOneByte'
  -count=1 -timeout 30m -v: PASS after disk cleanup; the first two attempts failed
  before comparison with no space left on device. Removed only 628 old reproducible
  Go cache entries, 6,446,209,870 bytes. The unchanged oracle then passed.

The reproducible owned launcher is python3
stage1/cohere/lint/rules/nexus-import-require-module-alias/validate_landing.py.
It was run end to end: exit 1, nine original-case groups PASS, the constructor
group FAILS, and the no-lonely-if mutant PASSes on all three backends. It saves
output to evidence/area-per-rule-upstream.log and exits nonzero while the constructor
case fails. No failure is converted into a skip.

| Retained rule | Original cases | Four-way bytes | Result |
|---|---:|---:|---|
| nexus/import-require-module-alias | 28 | 5,897 | PASS |
| @typescript-eslint/no-duplicate-enum-values | 57 | 21,418 | PASS |
| @typescript-eslint/no-dynamic-delete | 42 | 8,590 | PASS |
| @typescript-eslint/no-misused-new | 45 | 7,661 | PASS |
| @typescript-eslint/no-extra-non-null-assertion | 19 | 6,116 | PASS |
| @typescript-eslint/no-confusing-non-null-assertion | 28 | 5,842 | PASS |
| @typescript-eslint/no-unnecessary-parameter-property-assignment | 92 | not certified | parser FAIL |
| no-lone-blocks | 77 | 29,583 | PASS |
| no-lonely-if | 32 | 12,358 | PASS; extra budget witness FAIL |
| no-loss-of-precision | 151 | 40,784 | PASS |

## Every retained semantic mutant

Each mutant compiles and exits zero with clean stderr on source Node, emitted
JavaScript and ASan/UBSan native. Only byte comparison with actual Go catches it.
The first nine are proved by the default TestMutants; the last uses all 32 original
no-lonely-if cases through the same shared build/comparison primitives.

| Rule | Mutation |
|---|---|
| module alias | silently accept namespace style |
| duplicate enum values | advance numeric duplicate anchor |
| dynamic delete | accept unary plus as a static key |
| misused new | skip getters named new |
| extra non-null assertion | delete the bang at the wrong position |
| confusing non-null assertion | replace wrap closing parenthesis with ] |
| parameter-property assignment | retain the trailing semicolon |
| lone blocks | suppress the sole nested bound block |
| numeric precision | read normalized text instead of raw source |
| lonely if | add an extra space to the replacement |

## Preserved blockers and reproducers

1. Parser, outside owned directories. Exact checked-in reproducer:
   ../typescript-eslint-no-unnecessary-parameter-property-assignment/gaps/unsupported-destructured-parameter.ts.txt

```typescript
class Foo {
  constructor(public { a }: { a: string }) {
    this.a = a;
  }
}
```

The shared Node parser exits 70: expected CloseParenToken, got OpenBraceToken at
33. This is original upstream ParameterProperty.ts, captured case-381. Both the
default and owned grouped comparison fail on it. No parser file is edited.

2. Shared oracle, outside owned directories. Exact checked-in reproducer:
   ../no-lonely-if/testdata/convergence-budget.ts.txt. Eleven nested else-if levels
   need eleven fix passes. Go's real fixer stops at ten with Converged:false and
   'the pass budget was exhausted before the file converged (10 passes)'. The
   shared oracle panics 'fix failed' rather than serializing that refusal, before
   any port comparison. The unchanged default witness and mutant gates stay red.

These are blockers, not passing checks or permission to start another helper.
The split-UTF-8 parameter-property suggestion remains an explicit refusal and is
not arbitrary-input parity. No full repository gate or fresh throughput benchmark
is claimed. Historical per-rule native/Node/Go rates in PARKED.md predate the
handed-node/shared-driver changes and are not reused as current measurements.

Setup: Go/clang/Node ready 0s each; submodules 0s; build cache warm 89s; total 89s;
nproc 5, cgroup CPU quota 4 cores. Required compiler corpus input was supplied.
The seventeen broader repository correctness checks were not run by this
filtered lint/compiler gate. No failing input requirement was bypassed, relaxed
or deleted to obtain green.
The separately rebased helper branch is pushed at 9697eb4393a53d3862519d260f9a3e05e2fdefaa
with fresh fixed-hex/Unicode primitive comparisons and mutants. Those helpers
supply eight prerequisites for four rules, zero completely unblocked rule sets.
