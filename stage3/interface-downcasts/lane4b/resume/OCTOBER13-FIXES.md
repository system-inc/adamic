Built: restored the requested branch, incorporated integration, added three defect regression witnesses and a named lazy tuple union refusal.
Commits: started at 534b8c525de2624b938dd25dd3dc36f49ae51b25; fast-forwarded to integration 213c43771bd2f3a3dd9abd69ea088bf13e4c629e; fix commit contains this report.
Checks: restored fixes and focused counts passed in 1.461s; two measured rows added to counts.md; full integration counts refresh failed in 53.210s.
Mutants: reverted graph adoption size and diagnostic size, each caught by ASan heap-buffer-overflow; removed tuple-member refusal, caught by the named refusal assertion.
Uncovered: no whole repository gate or whole-tsc execution; 41 outside-unit counts fixtures failed, with preexistence unconfirmed and no unrelated repairs attempted.

The newest integration was a descendant of this lane's pushed tip, so there was
no merge conflict and all earlier plan sections remain. It already contains
73c778ae6b3cbfaf08b7f4d20c054d19366da1b7's full object graph adoption size and
f201b579's doubled declared-type diagnostic size. Those production fixes were
retained, not duplicated. The new executable witnesses fail when each is reverted.

Tuple alternatives remain object reference storage, not a certified positional
contract. The lane classifier permits the shared lazy descriptor to retain that
obligation. An unread cast admits; a demanded union read refuses with
`checked view read of field value with unsupported tuple union member contract`.
No tuple member is converted into a supported Unknown descriptor.

Commands, with `/workspace/adamic-tools/env.sh` sourced and output in logs:

```
python3 stage3/interface-downcasts/lane4b/resume/run-fix-mutants.py
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewObjectPrimitiveFix(Counts|es)$' -count=1 -v -timeout 10m -args -update-counts
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts
```

The first two pass. The last fails on 41 fixtures outside this lane and writes
no table. Per the user's instruction those failures are not pursued. The
focused updater adds only the measured new rows in existing registry order:
`fixes-graph.a` 3/3/3/7/3/0/0/0 and `fixes-tuple-unread.a` 1/1/3/6/1/0/0/0.
Columns are allocations/frees/retains/releases/peak/arena/graph regions/merges.
Both controls agree with source Node, sanitized native, release native and
backend JavaScript; successful controls also pass a separate leak check.
The diagnostic witness uses a long declared type name and checks every byte in
both backends. The tuple read is a compile refusal and has no counted run.

Setup: Go ready .075s, clang ready .436s, Node ready .063s, markdown ready .980s,
submodules ready 280.785s, build ready 543.000s, cache warm 543.109s, done 543.137s.
`nproc` is 5 and cpu.max is 400000 100000. Go 1.27.1, clang 20.1.8, Node 24.19.0.
The first test attempt was premature while the checker submodule was still
cloning and could not find its go.mod; it was rerun after setup finished.
No setup failure or credential workaround occurred.

Logs in resume/logs/oct13-*.log preserve setup, focused checks, mutant summary
and the unsuccessful integration counts refresh. Individual sanitizer mutant
logs are under /tmp/lane4b-fix-mutants/. The runner restores every mutation.
