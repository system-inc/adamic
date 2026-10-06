# Commands and results

All commands ran from /workspace/adamic, with
source /workspace/adamic-tools/env.sh in each toolchain shell.
All test output was redirected to a file and then read. None was piped to head
or tail. observations.json records the exact argv, exit, stdout and stderr for
each of the final 37 programs (9 successes, 22 differences, 6 refused builds).
Generated JavaScript and executable binaries remain in scratch, not in git.

## Preparation and review

```sh
git fetch origin codex/non-null-check:refs/remotes/origin/codex/non-null-check
git fetch origin main
git switch -c coverage/non-null-check origin/codex/non-null-check
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
nproc
git diff origin/main...origin/codex/non-null-check > /tmp/non-null-check/branch.diff
git log --format=fuller origin/main..origin/codex/non-null-check > /tmp/non-null-check/commits.log
```

CLAUDE.md was read before branch inspection, along with the language/memory
references and the feature's changed files. See coverage.md for setup timings,
coverage audit and exploratory setup/loader failures.

## Final direct builds and all three runs

```sh
python3 notes/non-null-check/observe.py /tmp/non-null-check/evidence > /tmp/non-null-check/evidence-builds.log 2>&1
python3 notes/non-null-check/observe.py /tmp/non-null-check/object-evidence missing_object_reference > /tmp/non-null-check/object-builds.log 2>&1
```

The observer executes these exact forms for every source file in its two lists:

```text
node --disable-warning=ExperimentalWarning oracle/node.mjs <absolute-source.a>
go run ./cmd/adamic build <absolute-source.a> -o /tmp/non-null-check/evidence/<stem>
/tmp/non-null-check/evidence/<stem>
go run ./cmd/adamic js <absolute-source.a>
node --disable-warning=ExperimentalWarning oracle/node.mjs /tmp/non-null-check/evidence/<stem>.mjs
```

A failed build cannot run a native binary; those six diagnostics are recorded.
A successful JavaScript emission is saved to the .mjs path before it runs.
The runner exits 0 after recording the differences; that exit is not a claim
that every program agrees. builds.log contains its classifications.

## Actual oracle on candidate and differing programs

For these first two commands only, the notes missing programs were temporarily
registered with lowers=true and checked=false. They were removed before the
final passing test, counts and full gate. Both tests exit 1 as expected:
20 source-Node disagreements; the combined run also passes nine new fixtures.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/^(internal|notes)$/^(oracle|non-null-check)$/^(testdata|missing_.*[.]a)$/^coverage_non_null_.*[.]a$' -count=1 -timeout 30m -v > /tmp/non-null-check/oracle-candidates.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/^notes$/^non-null-check$/^missing_.*[.]a$' -count=1 -timeout 30m -v > /tmp/non-null-check/oracle-differences.log 2>&1
```

The final nine fixtures pass, exit 0, after restoring the compiler:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^coverage_non_null_.*[.]a$' -count=1 -timeout 30m -v > /tmp/non-null-check/oracle-final.log 2>&1
```

This run compares source Node, sanitized native, release native and the
JavaScript backend, and performs the leak checks. Elapsed: 7.861s.

## Mutation proof

Replace exactly the line at internal/lower/non_null.go:52:

```go
result := ir.Expression(ir.Coalesce{Value: value, Panic: message, Of: value.Type().Present()})
```

with:

```go
result := ir.Expression(ir.Coalesce{Value: ir.Conditional{Condition: ir.BooleanConstant{Value: false}, WhenTrue: value, WhenNot: fit(ir.Undefined{}, value.Type()), Of: value.Type()}, Panic: message, Of: value.Type().Present()})
```

Then run:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^coverage_non_null_maps[.]a$' -count=1 -timeout 30m -v > /tmp/non-null-check/mutant.log 2>&1
```

Exit 1, with runtime disagreements, not a compilation failure. The original
line was restored immediately. mutant.log retains the exact comparison.

## Counts and repository gate

```sh
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/non-null-check/counts.log 2>&1
gofmt -l cmd internal > /tmp/non-null-check/gofmt.log
go vet ./... > /tmp/non-null-check/vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./... > /tmp/non-null-check/full-gate.log 2>&1
git diff --check
```

Counts: exit 0, 80.717s, exactly nine new rows. Formatting: no output. Vet:
exit 0, no output. git diff --check: no output.

## Final object-reference probes

The remaining missing object-reference representation was added as two notes
programs after the first 35 observations. observe.py's optional prefix argument
runs only those two. Their exact build and run argv are included in the merged
37-program observations.json. Both Node sources exit 0; native and generated
JavaScript exit 70 with the expected non-null assertion panic.

The object probes were temporarily registered as ordinary oracle fixtures for
this command, then those registrations were restored away:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/^notes$/^non-null-check$/^missing_object_reference.*[.]a$' -count=1 -timeout 30m -v > /tmp/non-null-check/oracle-object-differences.log 2>&1
```

This exits 1 as expected for the two source-Node differences.

Full uncached repository gate: exit 0. Every package passed. Native: 436.734s;
oracle: 315.859s; Unicode: 1032.058s; JSON: 687.491s. The final gate log is
full-gate.log. No compiler edit remains.

## Commit and push

```sh
git commit -m "Cover non-null assertions and record differences from Node"
git push -u origin coverage/non-null-check
git rev-parse HEAD
git ls-remote --heads origin coverage/non-null-check
```
