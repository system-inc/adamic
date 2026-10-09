# Task aecx10a item 144: checked maybe-scalar union views

Built from compiler/fx6-candidates-2 at 3e2166d51eec649396a004d95c91ae331763d8f5. The reported stop is in internal/native/view_unions_read.go:65, not internal/lower. No protected emitter orchestration files were edited.

Native emission now validates union membership and the scalar kind before constructing a packed maybe-number or maybe-boolean from the existing runtime snapshot. Undefined sets the missing presence bit; zero and false remain present payloads. Boxing and ownership use the existing union path. JavaScript needed explicit scalar typeof tests in its narrowed read wrapper; previously it generated an empty scalar test and the original programs compiled but failed during execution. Both backends retain the same contract membership checks and diagnostics.

The original p54_maybe_n and p54_maybe_b witnesses are active agreement rows with their pending sidecars removed. They print undefined followed by 0 or false, matching Node. The existing string | undefined shape remains held to Node. Additional number and boolean state transitions cover 7/true and return to undefined.

Misfit fixtures contain a string behind a number | undefined or boolean | undefined view. Source Node prints wrong and exits 0. Native release, sanitized native, and JavaScript must stop before exposing the field: exit 70, empty stdout, and field read failed: viewed.value matches no member of the declared union; expected the declared union, found string. Explicit oracle tests pin the complete diagnostic.

The skip-conversion-check.diff mutant removes the native conversion check call, leaving the conversion and valid generated C. Both misfit tests then fail: native and sanitized native print undefined and exit 0; JavaScript still stops with exit 70. No sanitizer report. run-mutant.py restores production code in finally. The first runner invocation required the log text got exit 0, but this harness prints native exit=0; the assertion was corrected and the mutant rerun successfully. The mutant was caught by runtime results in both invocations.

Setup: export GOPROXY=https://proxy.golang.org|direct; timeout 240 bash cloud/setup.sh, then source /workspace/adamic-tools/env.sh. Go 0.022s, Node 0.027s, submodules 0.084s, clang 0.209s, build cache 39.119s, total 39.161s. nproc 5, cgroup four-CPU quota.

New lower leaf seconds in focused-restored.log: MaybeNumber 1.31, MaybeBoolean 1.40, MaybeString 0.89, MaybeNumberTransitions 1.20, MaybeBooleanTransitions 1.49. New misfit oracle leaves each 0.60s in focused.log. All new test functions are top-level and parallel.

Commands (all test outputs captured in this directory):
- timeout 180 go test ./internal/lower ./internal/oracle -run '^TestScalarUnionView' -count=1 -v -timeout 90s: PASS lower 0.488s, oracle 0.613s.
- timeout 120 go test ./internal/lower -run '^TestScalarUnionView' -count=1 -v -timeout 90s: PASS 1.557s after restoration, including added transitions.
- timeout 240 python3 review/compiler/fx7-scalar-union-view/run-mutant.py: PASS runner, expected test exit 1.
- timeout 120 go test ./internal/ir -run TestCallTargetReaders -count=1 -timeout 90s: PASS 23.216s.
- timeout 300 go test ./internal/lower -count=1 -timeout 240s: PASS 85.918s, lower.log.
- timeout 360 go test ./internal/oracle -run 'TestCountsAreRecorded|TestScalarUnionView|TestNativeAgreesWithNode/internal/lower/testdata/scalar_union_views|TestReviewProgramsAgreeWithNode/fxspptb_oct9_native_p54_maybe_' -count=1 -v -timeout 5m -args -update-counts: PASS 100.171s, oracle-counts.log; counts.md refreshed with five new fixture rows.
- Integration lane command after commit: git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -.

This implements item 144's checked maybe-scalar conversion. Optional/absent view members and arbitrary narrowed scalar unions outside these representations are not changed. No full oracle package or repository gate was run, no cohere source was copied, and no PR was opened.

Lane checks PASS 1.7s: gofmt/tools on 19 Go files, t.Parallel on 2 test packages, vet 2 packages.
