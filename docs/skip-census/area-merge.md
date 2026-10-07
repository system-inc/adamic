Merged: origin/area/developer-tools at 2adf65c2 into devtools/skip-census, without rebasing.
Merge commit: adadcc21b204d5241af318295c744aed4f02e92b.
Checks: all eight package test functions pass; package go vet and diff checks pass.
Mutants: removing opt-in-lane support and reverting formatted skip matching both fail their intended tests.
Not covered: running the full gate or the dedicated verification lanes.

The six newly landed sites bring the census to 57 declarations. Existing
classifications remain intact. There are 34 required-input, 17 measurement,
3 not-applicable and 3 opt-in-lane sites.

| Site | Condition | Class and rationale |
| --- | --- | --- |
| TestRuntimeCacheKeepsCoverageApart | CoverageRequested() | required-input: ADAMIC_C_COVERAGE=1 forces every runtime build to be covered, disabling the covered/ordinary separation proof. Leave the override unset in the normal correctness gate. CoverageRequested in native.go reads this variable. |
| TestRecordBenchmark | os.Getenv("ADAMIC_RECORD_BENCH") != "1" | measurement: five-round timing comparison, no timing pass threshold. |
| TestGCCAgreesWithNode | os.Getenv("ADAMIC_GCC_LANE") != "1" | opt-in-lane: separately selected compiler verification. Missing compilers fail after opt-in. |
| TestNativeReleaseFlagsAgreeWithNode | os.Getenv("ADAMIC_RELEASE_LANE") != "1" | opt-in-lane: separately selected shipping-build verification. Missing tools fail after opt-in. |
| TestNativeExistingReleaseAgreesWithNode | os.Getenv("ADAMIC_RELEASE_MEASURE") != "1" | measurement: existing-release timing control. |
| releaseFixturesAgreeWithNode | expected.exitCode != 0 | opt-in-lane: both release runs explicitly restrict their scope to finishing fixtures. Ordinary oracle verification covers non-finishing behavior. |

The new opt-in-lane class distinguishes verification on separately selected
shards from timing observations. Accepting such a skip in an ordinary log does
not prove that the dedicated shard ran; retain the shard's ordinary completion
and pass verdict. This change adds no cache or performance claim.

Formatted numeric skip reasons now distinguish the release helper from the
opt-in entry guard in the same enclosing test. Unsupported or ambiguous reasons
still fail closed. TestReleaseLaneSkippedFixture exercises a rendered Node exit
70, and rejects a nonnumeric replacement. The existing historical regression
still reports skips=33 required-input=17 unknown=0.

Validation ran with output redirected to files:

```sh
go test -v -count=1 ./internal/skipcensus/... > /tmp/skip-census-area-merge-tests.log 2>&1
go vet ./internal/skipcensus/... > /tmp/skip-census-area-vet.log 2>&1
git diff --check
```

The package test log and both isolated implementation-mutant logs are checked
in beside this report. Removing opt-in-lane from CheckLog's accepted classes
fails TestLogClassesAndMutants; reverting rendered Skipf reason matching to
literal matching fails TestReleaseLaneSkippedFixture. Both mutants compile
and exit 1 because of their intended assertions.
