Built the runtime zero snapshot and compiler reference-tag retain guard for task #aecx10a, item 143.
Implementation commit: d3d2e838 on branch compiler/fx7-snapshot.
Focused Node differential, call-target and view checks pass; counts and lane receipts are accompanying logs.
Both fixes pass alone; both mutants fail native comparison under release and sanitizer builds.
No whole-package gate or full gate was run; integration owns that gate.

Fixtures erase value through a base interface before reading a mixed union view. The physical field remains maybe-boolean (true, false, undefined) or maybe-number (7, 0, undefined), after spreading a number into the same slot. The additional string arm forces the boxed union reader rather than the unavailable narrowed scalar view conversion. This is the conservative interpretation of p01's checked-view shape.

Node prints `true / true`, `false / false`, `false / undefined` for boolean and `7`, `0`, `undefined` for number. Native release, ASan/UBSan native and JavaScript agree. Lower tests use lowersAndAgreesWithNode. Sanitized native tests also check leaks.

Commands (all output saved here):

- `export GOPROXY='https://proxy.golang.org|direct'; timeout 300 bash cloud/setup.sh` (setup.log). Source `/workspace/adamic-tools/env.sh`; nproc 5, cgroup four CPUs. Setup completed in 294.794s. Initial /opt path failed with no such file; corrected to printed path. A cold compile reached a 120s command limit; retry passed.
- `timeout 120 go test ./internal/lower ./internal/oracle -run '^TestViewSnapshot' -v -count=1 -timeout 90s` (final-fixtures.log).
- `timeout 180 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s` (call-targets.log): pass, 17.454s.
- `timeout 180 go test ./internal/lower -run 'View|view' -v -count=1 -timeout 90s` (lower-views.log): pass, 2.960s.
- `timeout 180 go test ./internal/native -run 'View|view' -v -count=1 -timeout 90s` (native-views.log): pass, 0.406s.
- `timeout 360 python3 review/compiler/fx7-snapshot/run-mutants.py` (mutants.log and individual logs). Restore sources in finally; no mutant remains active.
- `timeout 600 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 9m -args -update-counts` (counts.log).

Mutants and observations:

- Runtime alone with compiler fix disabled: passes.
- Compiler alone with runtime fix disabled: passes.
- Runtime mutant copies raw slot back for maybe-boolean, compiler fix disabled: native exit -1 versus Node 0. Sanitized exit 1: UBSan misaligned adamic_heap access, heap.c:226, address 0x4045000000000002.
- Compiler mutant retains regardless of tag, runtime fix disabled: native exit -1 versus Node 0. Sanitized exit 1: ASan SEGV on a high address in adamic_retain, heap.c:226, stale 0x4045000000000000.
- Initial same-type fixture let the mutant survive; erased-base mixed-union fixture above catches it. The narrowed scalar target exposed an existing unavailable conversion, so the target uses a mixed union. The direct runtime test initially reused one cache for two property names; a separate absent-property cache fixes the test.

New test durations: lower boolean 0.27s, lower number 0.26s; oracle boolean 1.31s, oracle number 0.86s; native payload ownership 0.24s. Every new leaf is below 60s. Native payload test covers raw storage 0 preservation and full-byte zero payloads, besides both present and absent packed representations. Existing native view tests cover selection and reference families; no new heap-representation fixture was added.

Roadmap contribution: item 143's runtime and compiler checked-view snapshot correctness, with independent failure evidence. The brief supplies task #aecx10a and item 143, but no separate numbered roadmap step.

Counts refresh passed (58.552s) after installing stage3/api pinned dependencies with `timeout 120 npm ci --prefix stage3/api`. Two existing fixtures lose one retain of undefined each; the two new rows are recorded. Current origin/main was fast-forwarded before committing.

Lane checks passed: `lane checks 5.3 s: gofmt and tools on 4 Go files, t.Parallel on 3 test packages; vet 3 packages`. The checkout fetched only main by default, so the lane branch remote-tracking refs were populated with explicit refspecs before running the prescribed command. No PR was opened.
