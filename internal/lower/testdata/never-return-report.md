# Never calls in value positions

Built shared IR continuation markers fitted to the destination, emitted in C and JavaScript.
Delivery branch: codex/never-return-c; base 784b577a7488ddd0ce4cb2b82a96fa6535395896.
Focused Node/sanitizer oracle, IR contract, C audit and counts refresh passed; commands below.
Drop-fit failed clang and the IR contract; skipped JavaScript call effects failed Node comparison.
No whole packages or full gate were run; execution coverage does not span every IR representation.

The explicit unit base, origin/compiler/area-next-drop, takes precedence over the
shared boilerplate's origin/main instruction. No other worker branch was merged.

Observed cause: functions returning never have a void call ABI. The value lowerer
wrapped a void call in Effects with an Undefined result. That result defaults to
Object. fit did not adapt Effects, so returnStatement declared an object pointer
and returned it from a string function. The mutant reproduces that exact clang
-Wincompatible-pointer-types diagnostic under -Werror.

Calls proven never now keep their effects followed by Unreachable. Contextual
slots and fit choose the marker's ABI without boxing, unwrapping, retaining or
converting a nonexistent call result. Existing never conditional/logical/coalesce
arms use the same marker. Native emits __builtin_unreachable followed by an ABI
placeholder; JavaScript emits an unreachable throwing expression. Native's
existing pending-exception check still runs after the call and before the marker,
so a thrown Error takes the ordinary cleanup path. The active catch fixtures
observe that path and finish leak-clean.

The exact supplied program prints ok. The second fixture executes never calls
returned as number, boolean, object, string | undefined, optional scalar and mixed
union; it also exercises closure results, arguments, both conditional selections,
locals, object fields and array elements. The existing host_never_branches and
host_scanner_never fixtures were rerun. Source Node, backend Node, native release,
ASan, UBSan and LeakSanitizer agree. The IR contract checks 14 destination ABIs,
including arrays, maps, closures, weak and typed-array representations.

The native emitter audit found no separate source-level never fallback. Void ABI
branches remain in emit_functions.go (closure dispatch), emit_statements.go
(discarded calls), and emit_objects.go (method thunks). Their placeholder words
are ABI/discard results; the shared lowerer prevents them being used as source
never values. TestOracleCWarningFree builds all registered lowering fixtures,
including stack_overflow, with the standard native warning-as-error flags.
The counts refresh also builds all counted input fixtures. Existing counts rows
are unchanged; new rows are 0/0/0/0/0/0 for never_string and
14/14/41/35/2/0 for never_values (allocations/frees/retains/releases/peak/regions).

Setup: nproc 5, cpu.max 400000 100000, Go 1.27.1, clang 20.1.8, Node 24.19.0.
First setup failed while its go-list/build overlapped the newly written IR:
internal/lower/expression.go reported undefined: ir.Unreachable at lines 208,
211, 729 and 730, and never_branch.go at line 20. Retrying after the writes
completed passed. Retry timing lines: go ready 0.018s; node ready 0.017s;
submodules ready 0.048s; markdown ready 0.049s; clang ready 0.103s;
go build ready 28.621s; test binaries deferred 28.739s;
build cache warm 28.740s; done 28.765s.
Environment: source /workspace/adamic-tools/env.sh.
The first C audit lacked stage3/api/node_modules/@types/node 25.3.3; installing
the repository lock with npm ci --prefix stage3/api fixed all ten load failures.

Commands (all test output written directly to logs):

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/never-return-setup.log 2>&1
bash cloud/setup.sh > /tmp/never-return-setup-retry.log 2>&1
source /workspace/adamic-tools/env.sh
npm ci --prefix stage3/api > /tmp/never-return-node-types.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle ./internal/lower -run 'TestNeverFitsWithoutConversion|TestNativeAgreesWithNode/internal/oracle/testdata/(never_|host_never_branches|host_scanner_never)' -count=1 -timeout 10m > /tmp/never-return-focus-final.log 2>&1
# oracle 1.753s; lower 0.016s, both ok
go test ./internal/oracle -run TestOracleCWarningFree -count=1 -timeout 15m > /tmp/never-return-c-audit-final.log 2>&1
# oracle 28.339s, ok
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 15m -args -update-counts > /tmp/never-return-counts.log 2>&1
# oracle 57.665s, ok
python3 internal/lower/testdata/run-never-mutants.py > /tmp/never-return-mutants-final.log 2>&1
# drop-fit, drop-fit-ir and drop-call-js caught; exit 0
```

The runner restores every edited source in finally. drop-fit returns the default
Object marker instead of the target marker: the exact witness fails with the
requested incompatible pointer diagnostic. The same mutant separately fails
TestNeverFitsWithoutConversion, proving this check independently of clang.
drop-call-js skips Effects.Body only for an unreachable result: native still
compiles and runs normally; backend stdout loses the labels, so the Node oracle
reports stdout differs. Every final mutant test exits 1.
An exploratory lowering drop-call mutant returned Undefined; it failed clang on
the number return before reaching its intended output check. It was rejected as
a behavior proof and replaced by the isolated JavaScript mutation above.

Limits: no full gate, no whole-package test invocation, no performance claim,
no WASI/split-build audit. The full existing oracle was compiled, not executed;
only the named never fixtures were run through all three execution paths.

After restoring the final mutants:

```sh
go test ./internal/oracle ./internal/lower -run 'TestNeverFitsWithoutConversion|TestCountsAreRecorded|TestNativeAgreesWithNode/internal/oracle/testdata/(never_|host_never_branches|host_scanner_never)' -count=1 -timeout 10m > /tmp/never-return-restored.log 2>&1
# oracle 34.185s; lower 0.032s, both ok
git diff --check
# exit 0, no output
gofmt -l internal/ir/taste.go internal/javascript/javascript.go internal/lower/expression.go internal/lower/never_branch.go internal/lower/never_test.go internal/native/emit_expressions.go internal/oracle/never_test.go
# exit 0, no output
```
