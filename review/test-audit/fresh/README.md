# internal/fresh audit, brief v4

Starting origin/main: 8de93800f4a27959e386a5b542e5cc160e410246.
Scope: `go test -list . ./internal/fresh/`, 812 top-level tests.
Logical rows: 801 generated TestFreshWrites units, coverage test and remainder as
one family; nine focused tests. No witness tests were found.

Code under test: Go internal/fresh abstract interpretation and write proof.
Oracles: self, handwritten IR proof assertions, refusal diagnostic/location,
and corpus known-kind/nonzero-site assertions. These tests do not execute an
outside implementation. The regex test comment references other suites, which
were not run here and do not determine this test's expected answer.

The fixed set contains 18 production mutants in four files. M2 inserts an
unknown diagnostic, which is outside the brief's fixed mutation menu. Its
results are retained as supplemental data; no sacred verdict relies on M2.
The other 17 use return early, drop statement, change constant/option, or
change a condition. No test or oracle source was mutated.

All matrix logs are gzip compressed, with uncompressed parsed JSON alongside.
Each JSON records the full command, wall seconds, exact failing top-level tests,
and source output lines. The .diff files contain independent production changes
against the starting commit, with no selector or harness edits. Apply one patch
at a time on that commit. Example:

    git apply review/test-audit/fresh/M6.diff
    go test -count=1 -timeout 90s ./internal/fresh/ -run . > M6.log 2>&1

The switched scratch source used ADAMIC_MUTANT=M1 through M18. It was compiled
once and then reused by go test from its build cache. Separate Go baseline
coverage and standalone probe binaries were also built. Original production
source is restored before this evidence is committed.

Family unique_kills means that only this logical family failed. Exact member
failures are retained in the matrix. A family is not compared to its own
members for subsumption. Repo-wide uniqueness remains unknown.

Warm toolchain: /workspace/adamic-tools/env.sh, nproc=5, Go 1.27.1,
Node 24.19.0, clang 20.1.8. Setup skipped. The first full baseline failed for
15 corpus units because stage3/api lacked @types/node. npm ci there succeeded
in 2.629 seconds. The corrected baseline with coverage passed in 43.056 seconds.
No dependency manifest was changed.

Brief feedback:

* Family naming is narrower than reality. This package has 801 named fixture
  tests plus coverage and remainder, rather than numbered shards and Union.
* Worthy asks about the whole repo, while repo-wide testing is forbidden.
  State the conclusion as package or logical-family uniqueness.
* Exactly one failing test and family grouping need one explicit counting rule.
* The verdict set is incomplete for caught, nonunique tests without one
  no-slower subsumer. cannot-judge is used for that classification gap, not as
  a claim that these tests cannot fail.
* Wall seconds from go test include Go command startup and build work. Tiny
  tests and tiny timing differences make the no-slower condition sensitive
  to machine load. Specify timing repeats and which statistic to compare.
* About three mutants per test should explicitly mean per logical row. The
  812 raw functions would imply thousands. Eighteen were fixed here instead
  of the approximate 30-row target to keep the full matrix near the budget.
* Warm env.sh does not establish all pinned dependencies. The explicit npm ci
  fallback was necessary and worked.
* The per-step 90 second budget should cover compilation plus test execution.
  go test -timeout itself only bounds the test binary. Our driver additionally
  imposed an outer 90 second process-group limit; no run reached it.
* First-catch commands and a mandatory full matrix duplicate execution. Here
  the complete matrix ran first, followed by a single-test reproduction of
  its first failing top-level test. This costs time but gives both forms of
  evidence without selecting mutants from test outcomes.
* A corpus test can pass after lower.Lower declines a fixture, and can pass on
  zero writes. Survivor output probes are necessary to separate changed
  behavior from ineffective mutations.
* My M2 did not obey the fixed menu. Its supplemental status is explicit.
  /usr/bin/time was unavailable, so Python monotonic timing was used after
  one failed timing command. These are execution deviations, not brief bugs.

No other packages, native .a/.ts ports, external runtimes, or memory sanitizers
were tested. No tests were deleted or changed, and no pull request was opened.

The coverage profile measures 97.2% of production statements. Every listed
production function was reached except state.closes (0%); it was included
in the initial static inventory as a candidate, not established reachability.

TestNodeFSFileOperationsAreKnown has only the non-menu M2 catch, so strict v4
judgment is unresolved. The MethodKeepsArgument subsumption hint rests on
two eligible kills. The borrowed-object hint rests on one eligible kill plus
one supplemental M2 kill. Neither hint is a deletion recommendation.

Measured timing breakdown is in timings.json. The initial standalone probe
build was not separately timed. No step in the test driver exceeded 90 seconds.
