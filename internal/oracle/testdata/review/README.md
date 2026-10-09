# Integration review programs

Add a complete `.a` program directly to `agree/` or `refused/`; no test table needs editing.
`TestReviewProgramsAgreeWithNode` lowers through stage 0 and compares the sanitized native,
JavaScript, and release backends with source execution on Node, including stdout, stderr,
and exit status. Successful programs also undergo the oracle's usual leak check.
`TestReviewProgramsRefuse` requires a compiler refusal with a diagnostic and never executes
an accepted program. Each fixture is one subtest named by its filename and must take less
than 30 seconds. `TestReviewProgramsNoLooseFiles` rejects programs outside these directories.

A known miscompile may have a same-stem sidecar, e.g. `agree/example.pending`, containing
exactly one line in this form:

```
awaits compiler/fix-branch: explanation of the miscompile
```

The subtest skips with exactly that message. Remove the sidecar when the fix lands; the
existing gate census tracks `awaits <branch>` skips. A malformed sidecar fails the subtest.

While no fixing branch exists yet, the sidecar may name the task instead: `awaits #<task>: <reason>`
(e.g. `awaits #cxr5x2v: catchable RangeError, step 21`). The Mac-side review census
(cloud/review-census.py on the gate tools) pages integration once that task is Done and the sidecar
is still there.

A ruled difference from Node (an intended departure, such as step 24's loud stack guard where Node
throws a catchable RangeError) is an `<name>.intended` sidecar, one line
`intended: <ruling>: <what differs>`, with the pinned output in `<name>.expected.json`
(`{"stdout": "...", "stderr": "...", "exit": 70}`). Native, the JavaScript backend and the release
build are held to the pin instead of Node. A pin that Node now matches is stale and fails.

`TestReviewProgramsSelfTest` plants a wrong expected stdout for `agree/smoke.a` and requires
every backend comparison to reject it. It also plants an accepted program in a temporary
`refused/` directory and a loose program in a temporary review root, and requires the same
checks the lane uses to reject each.

## Initial validation (2026-10-08)

Based on `origin/main` at `031a1259bc7973934792dc6cb1bd4074fc2204b9`.
With the executor limited to four CPUs, `GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=1 go test
./internal/oracle -run '^TestReviewPrograms|^TestTheOracleCatchesOneByte$' -count=1 -json`
reported these subtest times:

| Subtest | Seconds | Result |
| --- | ---: | --- |
| agree/fxspptb_p06_helper_second_parameter_undefined.a | 0.40 | Native and release stdout disagreement |
| agree/fxspptb_p24_filter_proven_number.a | 0.39 | Native and release stdout disagreement |
| agree/smoke.a | 0.32 | Pass |
| refused/non_null.a | 0.03 | Pass |
| self-test/wrong_output | 0.40 | All three planted comparisons caught |
| self-test/accepted_refused | 0.03 | Accepted program caught |
| self-test/loose_file | 0.00 | Loose program caught |

The first native fixture execution, including building shared native dependencies, took
12.24 seconds, also under the 30-second limit. The existing `TestTheOracleCatchesOneByte`
control passed (0.26 seconds).

Additional temporary probes made the actual lane tests fail: a loose `.a` (0.00 seconds),
an accepted program added to `refused/` (0.04 seconds), and removal of a pending sidecar
from the helper seed (0.41 seconds). With that temporary sidecar present the helper
subtest skipped with exactly its `awaits devtools/review-lane: ...` message. All temporary
probe files were removed.

No verified fixing branch was found among origin's branch names and the fetched compiler
and predicate branches' commit messages/history (`predicates_proof`, `arrayVisit`, `filter`,
second parameters, and narrow refusals). Consequently both supplied miscompile seeds have
no pending sidecar and deliberately leave the agreement lane failing, pending identification
of their actual fixes. Neither the historical predicate branches nor the ahead specification
branch supplied evidence sufficient to assign either fix.
