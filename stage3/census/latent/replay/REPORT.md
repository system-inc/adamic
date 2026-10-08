# Census replay validation

Built an automatic scratch-overlay `go run` entry that replays one actual lowering finding.
The work starts from resolved pin `29e03f8e6cff3171dc521ead1087c4ae2963e7f3` and includes the named base branch through `9d534d3a31814f1a192a528e701f6c2ea7c910bc`.
The two integration tests pass; the parent's-sibling selection mutant fails the positive reproduction assertion.
Three real entry-project findings replay in 6.764s, 2.297s and 2.175s against a 193.807s full census.
No whole package test or full gate was run; backend semantics and all other census signatures are outside this validation.

Assumption: the named branch's full census is the requested starting point. The
specified 29e03f8e pin predates `units.go.txt`; three later commits on that branch
supply the full implementation and the 03:52Z evidence. No production compiler
file was edited. The entry-root comparison is used because that report's 10,551
attempts belong to the tsc entry scope, rather than its compiler-directory scope.

The selector uses the existing census candidates and direct-body checker
eligibility to choose the smallest attempted unit containing the diagnostic node.
Both census and replay call the same attempt loop, register the entire project,
and snapshot the same lowering state. Replay filters attempts before lowering and
omits the file refusal scan. The match requires the exact diagnostic position,
kind and reason. The worker always checks the no-output loader and lowerer guards.

## Observed results

All 81 entry-reach source hashes matched
the external latent-full source manifest after
`bash stage3/apply.sh /tmp/census-replay-adapted` completed. The checker retained
324 diagnostics. No source reduction or binding/declaration deletion was used.

| Finding | Exact reason | Full command wall time | Load/register/lower |
| --- | --- | ---: | ---: |
| binder.ts:331:22, NotYet | a PrefixUnaryExpression on a value | 6.764s | 1.977s |
| parser.ts:433:39, NotYet | new a ParenthesizedExpression | 2.297s | 1.977s |
| scanner.ts:150:5, NotYet | a computed field name | 2.175s | 1.882s |

Each command was `go run ./stage3/census/latent/replay -project
/tmp/census-replay-adapted/src/tsc/tsc.ts -where
/tmp/census-replay-adapted/src/compiler/FILE:LINE:COLUMN -kind NotYet -reason
'EXACT REASON'`, using the corresponding table row. Each exited 0 and printed
`reproduced NotYet`, the exact reason and node position. The command wall time is
measured around the entire subprocess, including the outer Go invocation,
overlay generation and worker compilation. The first invocation rebuilt the
worker; the later ones reused Go's cache. These are single observations, not
statistical benchmarks. timings.json (external run evidence) preserves the exact
argument arrays and printed durations.

The full run used the existing entry overlay and the same sources:

```sh
source /workspace/adamic-tools/env.sh
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/census-replay-overlay > /tmp/census-replay-overlay.log 2>&1
python3 stage3/meter/entry_overlay.py /tmp/census-replay-overlay /tmp/census-replay-entry-overlay > /tmp/census-replay-entry-overlay.log 2>&1
go build -buildvcs=false -overlay=/tmp/census-replay-entry-overlay/overlay.json -o /tmp/census-replay-entry ./stage3/census/latent/tool > /tmp/census-replay-build.log 2>&1
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /tmp/census-replay-entry /tmp/census-replay-adapted/src/tsc/tsc.ts /tmp/census-replay-full.jsonl > /tmp/census-replay-full.log 2>&1
```

A Python parent measured that last subprocess with `time.monotonic()`:
**193.806843s**, exit 0, **10,551 attempted units**. Its build time is excluded.
The parent then asserted that each replay's complete findings array equals the
unfiltered census's findings for its selected unit, with the same order and
measurement labels. All three comparisons passed. The complete census observations
remain at `/tmp/census-replay-full.jsonl`; the committed
measurement log (external run evidence) records these assertions and times.

## Tests and mutant

```sh
python3 stage3/census/latent/replay/replay_test.py > /tmp/census-replay-tests-final.log 2>&1
LATENT_REPLAY_MUTANT_PARENT_SIBLING=1 python3 stage3/census/latent/replay/replay_test.py ReplayTests.test_nested_findings_match_full_census_in_order > /tmp/census-replay-mutant-final.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/census-replay-counts.log 2>&1
go vet ./stage3/census/latent/replay ./stage3/census/latent/replay/worker > /tmp/census-replay-vet.log 2>&1
```

The final two integration tests printed `Ran 2 tests in 4.744s`, `OK`, exit 0.
The nested unit lives under a checker-rejected parent, reads an ancestor capture,
and retains an imported project's binding. Its three refusal reasons and all
Boundary records equal the unfiltered census findings, in order. A non-BMP
character before the target on the same line exercises UTF-16 column conversion.

The second test selects the parent's sibling. That sibling also refuses `var`,
so matching only the reason would be wrong. The replay exits 1 with
`replay signature did not reproduce`; its selected unit is `sibling`, and its
findings exclude the target node. Running this same selection mutant against the
positive test printed `FAILED (failures=1)`, exit 1, at
`self.assertEqual(result.returncode, 0, log)` with `AssertionError: 1 != 0`.
It was caught by the reproduction assertion, without a build failure. This was
the only deliberate implementation mutant run.

The counts check printed `ok github.com/system-inc/adamic/internal/oracle
65.443s`, exit 0. `counts.md` did not change; the synthetic measurement inputs are
created under a temporary directory and are not oracle fixtures. Focused vet,
`gofmt -l` for the two `.go` entries, and `git diff --check` produced no output.
Raw test (external run evidence), mutant (external run evidence)
and counts (external run evidence) output is committed.

## Setup and limits

`export GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh` succeeded.
Its timing lines were Node ready 0.023s, Go ready 0.029s, clang ready 0.217s,
markdown dependencies installed step 0.655s and ready 0.736s, submodules ready
147.171s, Go build ready 453.980s, test binaries deferred 454.369s, build cache
warm 454.371s, done 454.431s. `nproc` was 5, with a four-CPU cgroup quota.
The printed `/workspace/adamic-tools/env.sh` was sourced for builds and tests.
No setup workaround was needed. setup.log.txt (external run evidence) retains
all timing lines and tool versions.

Replay is an observation command. Scan-only findings and a finding observed in a
larger ancestor's context may fail when the required smallest attempted unit is
selected. Such mismatches exit nonzero. This validation does not claim all census
findings reproduce, exhaustive expression coverage, whole-program compilation,
backend equivalence, or a full repository gate. No PR was opened.
