Built the oracle's exact ruled-divergence list and terminal JavaScript stack catch guard for task #z1vjxxd, step 21.
Branch: compiler/per-backend-stops, based on main 031a1259; this report accompanies its implementation commit.
Focused uncached oracle tests and changed-package vet pass; new fixture counts are refreshed and checked.
The unlisted-output and missing-stack-guard mutants fail semantically; both guard-condition removal mutants fail too.
No whole package test or repository gate was run; native lifetime rules and other checked-error semantics are unchanged.

The only divergence entry is 4ddd17f_weak_single_narrowed. It records the ruling,
reason and exact stdout, stderr and exit for each backend. Native prints before,
then exits 70 with its weak-after-free diagnostic. JavaScript and original Node
print before and after: nn, exiting 0. No reference counting is added to JavaScript.
Unknown fixture names use the existing byte-exact comparison. Mutating any listed
outcome's exit, stdout or stderr remains an error, independently for both backends.
The list and comparison helper live in internal/oracle/backend_divergences_test.go.
The only edit to oracle_test.go is the checked-fixture comparison call: that central
call is the necessary hook; the list and its tests are kept outside the protected file.

Every emitted IR catch now checks instanceof RangeError and the exact V8 message
Maximum call stack size exceeded before declaring the catch binding, marking catch
entry or running its body. It calls the existing terminal panic function, using
RangeError: Maximum call stack size exceeded from internal/native/runtime/stack.c.
Stack overflow is absent from the divergence list. The pinned try_stack source
catches on original Node, printing caught the overflow and after. Both Adamic
backends instead emit the exact native diagnostic and exit 70 with empty stdout.
Catch instrumentation is independently shown not to run before that stop.

The review sources are pinned unchanged as .a oracle fixtures. Source commit,
original and classified paths, and SHA-256 are in per-backend-stops-fixtures.json.
Only those source bytes are taken, without merging the review branch. A third
fixture holds ordinary thrown Error catches, including an Error with the overflow
message, and ordinary finally execution, to source Node, release native, ASan/UBSan,
backend Node and LeakSanitizer. Terminal native fixtures do not undergo a clean-exit
leak assertion. A generated JavaScript control separately proves a V8 RangeError
with another message remains catchable.

A first control used String.fromCodePoint(-1) inside try. Native's existing invalid
code-point check is terminal while V8 catches it. That control was removed rather
than registering another divergence or changing an unrelated language boundary.
A first direct-source control used relative paths and failed the cache harness's
absolute-path requirement; those paths were corrected. Neither failure is counted
as a mutant kill or hidden by refreshed expectations.

Final verification commands, each with complete output sent to a file:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'Test(RuledBackendOutcomes|UnlistedBackendAgreement|TerminalStackStop|NonOverflowRangeErrorCatch)$|TestNativeAgreesWithNode/internal/oracle/testdata/(4ddd17f_weak_single_narrowed|d96d304_try_stack|backend_stop_catches|exceptions|panic_in_try|stack_overflow)' -count=1 -v -timeout 10m > /tmp/stops-restored.log 2>&1
python3 internal/oracle/testdata/run-backend-stop-mutants.py > /tmp/stops-mutants.log 2>&1
go vet ./internal/javascript ./internal/oracle > /tmp/stops-vet.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 10m -args -update-counts > /tmp/stops-counts-final-update.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 10m > /tmp/stops-counts-check.log 2>&1
git diff --check > /tmp/stops-diff-check.log 2>&1
```

The restored focused run passes four direct controls and ten fixture cases, including
release and sanitized native comparison. Package output: ok, 3.961s. Vet and diff
checks exit 0 with empty output. Counts update passes in 42.384s and verification
passes in 38.507s. The entire counts diff adds only three rows:
weak_single_narrowed 4/2/4/6/4/0, try_stack 0/0/0/0/0/0 and backend_stop_catches
2/2/6/4/1/0 (allocations/frees/retains/releases/peak/regions). Existing rows are
unchanged. The mutant runner restores production source after
each mutation and rejects a build failure as a kill. Each go test exits 1:

| Mutant | Independent catcher |
|---|---|
| unlisted-output: append ! only to JavaScript's console output | TestUnlistedBackendAgreement: unlisted backend divergence: stdout differs. Native still prints the dedication correctly. |
| drop-terminal-guard: remove the emitted catch test | TestTerminalStackStop: exit codes differ. JavaScript exits 0 with caught the overflow and after; the pinned stop requires 70. |
| drop-range-error-kind: remove instanceof RangeError | Ordinary Error with the same text in backend_stop_catches.a incorrectly stops at 70; the source/backend comparison fails. |
| drop-overflow-message: remove the exact message restriction | TestNonOverflowRangeErrorCatch: a non-overflow V8 RangeError incorrectly stops at 70. |

Complete mutant logs: /tmp/backend-stop-mutants/unlisted-output.log,
/tmp/backend-stop-mutants/drop-terminal-guard.log,
/tmp/backend-stop-mutants/drop-range-error-kind.log and
/tmp/backend-stop-mutants/drop-overflow-message.log. No clang or Go build failure
is credited. The listed-outcome exit/stdout/stderr corruptions are also rejected
by TestRuledBackendOutcomes.

Toolchain setup succeeded after export GOPROXY='https://proxy.golang.org|direct'.
Timing lines: node ready 0.024s; go ready 0.032s; markdown dependencies skipped
step-duration 0.007s; submodules ready 0.080s; markdown dependencies ready 0.081s;
clang ready 0.196s; go build ready 45.086s; test binaries deferred 45.239s;
build cache warm 45.240s; done 45.270s. nproc=5, CPU quota=4. Go 1.27.1,
clang 20.1.8, Node 24.19.0. Environment: /workspace/adamic-tools/env.sh.
