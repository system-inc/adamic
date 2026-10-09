# Optional calls with the lowering-a stack

Merged origin/cloud/land-stack-lowering-a-82484e42 at
82484e42aaa915853370a50d62949bb188b6bbb1 into compiler/optional-calls-main,
whose previous tip was 12dce3bf4ed8a7f3a0c04b29e0b1b56e26dbdc39.
The stack includes main, optional-chain-after-call 5028fa69 and spread-shorthand
e82c9646. No rebase or history rewrite.

The only textual conflict was internal/oracle/counts.md. Its values were not hand
merged: one Linux `go test ./internal/oracle -run TestCountsAreRecorded -count=1
-parallel=4 -args -update-counts` generated the table, with GOMEMLIMIT=2GiB.
It passed in 120.487 s. The final table has 981 rows, preserving every row from
both parents. Compared with the previous branch tip, 58 rows are added and 16
change. lowering-stack-count-changes.json records every row and before/after value.
Eight additions are the batch's six optional-after-call and two spread/shorthand
fixtures; fifty are inherited checked-view fixture registrations from the stack's
main. Ten existing oracle rows and two step-18 chain rows change because required
reads now construct catchable TypeError objects and use the incoming exception
paths. Four step-18 retain counts decrease by one because nullish tagged slots
return directly instead of retaining a null pointer; their other counters stay
unchanged. No existing row is removed.

Automatic compiler merges preserve both branches' lowering and emission paths.
A runtime interaction needed repair: the stack introduces semantic null/undefined
slot tags, which optional native selection initially treated as non-callable.
The absent slot evaluated arguments and stopped with exit 70. Selection now
recognizes both tags and skips the call, preserving the receiver and argument
rules. The new nullish-slot-tags-rejected overlay removes this exemption; the
optional-method Node fixture catches it. No fixture or expected output was changed.

Uncached step-18 controls agree with Node in native ASan/UBSan, release/leak
checks and JavaScript, including stored field, optional method, getter and saved
callable cases. The represented scalar-result boundary remains NotYet as before.
All fourteen storage overlays are caught, including the eight ruled behaviors
and the new tag interaction. Five optional-after-call overlays are caught;
the spread and JSON shorthand IR mutants are caught in both backends. Evidence
selectors and log paths are in lowering-stack-mutants.json.

Commands write test output directly to /tmp/storage-stack-*.jsonl or .log:

- `ADAMIC_GATE_UNCACHED=1 go test -p 1 ./internal/oracle -run 'TestStep18Storage|TestStep18DetachedMethod|TestNativeAgreesWithNode/docs/step-18' -count=1 -json`: pass, 4.688 s.
- `ADAMIC_GATE_UNCACHED=1 go test -p 1 ./internal/oracle ./internal/flow -run '^(TestOptionalAfterCall|TestCallSpread|TestJSONStringifyShorthand)' -count=1 -json`: pass, oracle 25.906 s and flow 0.042 s. All six optional-after-call fixtures and both spread/shorthand fixtures run in both backends with native sanitizers.
- `go test -p 1 ./internal/lower ./internal/flow ./internal/ir -run 'TestStep18|TestCallTargetReaders|TestOptionalAfterCall' -count=1 -json`: pass, lower 0.194 s, flow 0.632 s and IR 11.854 s.
- `go build -p 1 ./cmd/adamic`: pass.
- `go vet ./internal/lower ./internal/native ./internal/javascript ./internal/oracle ./internal/flow ./internal/ir`: pass, empty output.
- `STORAGE_MUTANTS=/workspace/scratch/storage-stack-mutants python3 docs/step-18/run-storage-mutants.py`: all fourteen caught.
- The incoming optional-after-call overlay runner uses a scratch evidence directory: all five caught, without modifying its committed evidence.
- The ten moved existing oracle rows pass an uncached explicit Node fixture filter in 1.383 s.

The instance has a four-CPU quota and nproc=5. An incoming JavaScript test helper received gofmt only (a trailing blank line); no new top-level Go test function is
added or edited for this merge. Focused top-level times are recorded in
lowering-stack-test-durations.json; the longest batch leaf is TestCallSpreadArray
at 25.89 s. The existing complete counts sweep is unchanged. Full packages,
the full gate, WASI and the entire inherited checked-view acceptance suite are
not claimed run. No stage-3 status record or Node observation is regenerated.

After committing the merge, run the requested repository-root lane command:
`git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -`.
Its exact output accompanies the delivery report before the single push.
