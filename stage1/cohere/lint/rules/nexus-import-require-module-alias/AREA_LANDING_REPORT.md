# Integrated lint area landing

Current compiler landing: c7991b900; area base b84a9d931. See the final
compiler-refresh section for rebuilt results. Earlier sections preserve their
original observations and bases. The branch remains blocked for landing.

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

## Fresh landing-cap check after the next instruction

Fetched every origin branch again on October 7. Both integration tips are
unchanged: main 39638d9e278d38bb5aeae887f46d55a70e47aaad and area/stage1-lint
 d65a8f931c98655936ae04c6899f38f14862b73e. Both owned branches already descend
from them, so no rebase is needed. No implementation, compiler, corpus input or
shared file changed. The helper branch remains at 9697eb4393a53d3862519d260f9a3e05e2fdefaa.

Freshly executed, all output saved:

- go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1 -timeout 30m -v:
  FAIL again in 21.18s with the unchanged Go fix-budget panic.
- python3 stage1/cohere/lint/rules/nexus-import-require-module-alias/validate_landing.py:
  exit 1 again. Nine original-case groups (579 cases) match Go on Node, emitted
  JavaScript and sanitized native; the constructor group fails on the same
  original modified-destructured parameter. The extra-space mutant is freshly
  caught only by byte comparison on all three backends over 32 original cases.
  The Go test process completes in 74.128s; no case was skipped or removed.

The other nine rule mutants, full 368-file corpus and two helper mutants retain
the immediately preceding evidence on exactly the same implementation and inputs;
they were not rerun in this refresh. These are retained observations, not fresh
passes. New raw logs are evidence/refresh-budget.log, refresh-launcher.log and
refresh-fetch.log; area-per-rule-upstream.log records the complete fresh probe.

The rule branch remains blocked and is not declared landing-ready. Per the
landing cap, no further rule/helper is claimed. Resolving these two failures
requires the owners of the shared parser and oracle; this unit stops without
editing their files or weakening either check.

## Compiler landing refresh at c7991b900

Fetched all origin refs and rebased both owned branches onto
area/stage1-lint b84a9d9314b65d3d0261ee017e233287b4f071da, containing current
main c7991b900362796aefd111474e65eb5398e91953. DEDUP_LEDGER.md has identical
content to the already-read prior version; the same ten winners are retained
and the same three losing directories stay dropped. Compiler changes were
accepted without touching or reverting unowned files. Remote tips were checked
again before pushing and have not moved during these validations.

Unlike the preceding unchanged-base refresh, no corpus or mutant evidence was
reused after this compiler change. Fresh commands use
source /workspace/adamic-tools/env.sh and write logs under evidence/c799-*:

- ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-14-typescript-pinned go test
  ./stage1/cohere/lint -run '^TestCompilerAndStage1Agree$' -count=1 -timeout 30m -v:
  PASS, 368 files, 20,750,769 identical bytes on actual Go, source Node,
  emitted JavaScript and ASan/UBSan native (185.57s).
- python3 stage1/cohere/lint/rules/nexus-import-require-module-alias/validate_landing.py:
  exit 1, nine original-rule groups PASS (579 cases), modified-destructured
  constructor group FAILS with expected CloseParenToken, got OpenBraceToken at
  33. The extra-space no-lonely-if mutant freshly compiles and runs cleanly and
  is caught only by Go comparison on all three backends, over all 32 original
  cases. The complete probe process finishes in 82.192s.
- go test ./stage1/cohere/lint -run
  '^TestMutants$/(wrap_close_parenthesis_changed|numeric_duplicate_anchor_advances|unary_plus_accepted_as_static_key|extra_bang_deleted_from_wrong_position|getter_named_new_silently_skipped|trailing_semicolon_retained|namespace_style_silently_accepted|bound_sole_block_exemption_widened|normalized_text_hides_lost_precision)$'
  -count=1 -timeout 30m -v: PASS, all nine selected owned mutations freshly
  compiled and caught only by Go byte comparison on all three backends
  (266.22s). Together with the probe, all ten owned semantic mutants are fresh.
- go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1 -timeout 30m -v:
  FAIL (23.66s), unchanged shared Go oracle panic on the preserved convergence
  witness. Its real fixer reaches ten passes without convergence. This check
  was neither skipped nor weakened; the same checked-in reproducer remains.
- The correctly filtered inherited-static-field compiler fixture and one-byte
  oracle control PASS again (21.812s), as does go vet ./....
- Both retained helper comparisons and their compiling semantic mutants were
  also rebuilt in full and PASS on Node, emitted JavaScript and sanitized
  native. Fixed hex: 68,844 records / 2,149,768 bytes. Unicode: 12,935 records /
  834,966 bytes. All four original consumer suites and targeted leaf calls PASS.

Setup: tools ready 0s each, submodules 0s, build cache warm 149s, total 149s;
nproc 5, CPU quota 4 cores. Removed only two confirmed inactive scratch Go-build
directories from earlier failed runs to make room; no source, input or assertion
was removed. No full repository correctness gate, fresh throughput measurement
or arbitrary-input parameter-property parity is claimed. The broader seventeen
required-input checks were not run by this filtered lint/compiler gate.

The separately rebased helper branch is pushed at
5efb52bac2d03caa5d74d68659b11f9d7650b502. Its eight prerequisites for four
consumers still remove zero complete blocker sets. The rule branch remains
blocked on the shared parser and shared oracle, so no new helper or rule is
claimed. Reproducers above are unchanged. No main or area branch was pushed.
