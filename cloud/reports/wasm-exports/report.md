Built ABI v1 typed per-function WASI exports, signature JSON, reference host and differential fixtures.
Commits: claim 6b87570; implementation 5926f5e; origin/main f8013f0 merged as 4658552.
Checks: 37 boundary cases and 100,000 mixed calls pass; zero live values, flat 1,507,328-byte memory.
Mutants: string length, f64 padding, leaked result, record order and unsupported selection all caught.
Limits: prefix flag order needs an unowned main.go hook; duplicate lowered function names are refused.

The branch is codex/wasm-exports, based on origin/wasm/integrate at
6f7dce3dc1eace606fe081c8f4ab12034ae11bb4. The first pushed commit added only
cloud/reports/wasm-exports/claim.md. No main or area branch was pushed. No PR was opened.

The build selects crossing entry-module function exports automatically, or narrows
with repeatable --export. Explicit unsupported selections fail with a field path.
The adapters retain the normal compiler ownership/reuse conventions and statement
regions. Host-created record layouts participate in the uniform-field-offset proof
through unreachable layout witnesses. Top level runs once through the existing
adamic_module_start constructor. handleRequest also retains its legacy request API.
No existing runtime C or protected compiler emitter/lowerer files were edited.
The optional claimed wasm_exports.c translation unit was not needed: support is
emitted with the application C.

The signature table and ABI are documented in docs/wasm-abi.md. The reference
host preserves UTF-16 code units through WTF-8; a lone U+D800 is ed a0 80 and
returns to JS as U+D800, rather than U+FFFD. Ordinary UTF-8 remains unchanged.
The source fixture covers all crossing kinds, empty and large buffers, astral
characters, lone high/low surrogates, NULs, -0, NaN, infinities, depth-2 records,
reversed construction order, empty records, multiple buffer arguments, a consumed
record parameter, a region-heavy tree and persistent module state.

Setup: bash cloud/setup.sh --wasi-sdk, with output in /tmp/wasm-exports-setup.log.
Go ready 0s; clang ready 1s; Node ready 1s; WASI SDK ready 5s; submodules ready 5s;
build cache warm 180s; done 180s. nproc = 5, cgroup quota 4 CPUs, memory 17.6 GB.
Go 1.27.1, Node 24.19.0, native clang 20.1.8, WASI SDK clang 20.1.8.
Every build/test shell sourced /workspace/adamic-tools/env.sh and WASI shells set
WASI_SYSROOT=/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot.

Commands run, with complete output sent to logs:

```sh
ADAMIC_ORACLE_WASI=1 ADAMIC_GATE_UNCACHED=1 go test ./cmd/adamic ./internal/native -count=1 -timeout 15m > /tmp/wasm-exports-packages.log 2>&1
ADAMIC_ORACLE_WASI=1 go test ./cmd/adamic -run '^TestWASIExports$' -count=1 -v -timeout 5m > /tmp/wasm-exports-final-boundary.log 2>&1
ADAMIC_ORACLE_WASI=1 ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestWASI' -count=1 -v -timeout 15m > /tmp/wasm-exports-wasi-oracle.log 2>&1
ADAMIC_ORACLE_WASI=1 ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./... > /tmp/wasm-exports-final-gate.log 2>&1
gofmt -l cmd internal > /tmp/wasm-exports-final-format.log
go vet ./... > /tmp/wasm-exports-final-vet.log 2>&1
```

The pre-merge touched packages passed: cmd/adamic 6.707s, internal/native 112.889s.
The post-merge boundary test passed in 74.125s under concurrent gate load: 37 cases,
100,000 mixed calls, live 0 after every call, linear memory 1,507,328 bytes after
warmup, regions 406,476. It also checks automatic omission, explicit unsupported
refusal, and one-export narrowing in signature JSON. Output is in
/tmp/wasm-exports-final-boundary.log. Post-merge format and vet exited 0 with no output.

The separately filtered WASI oracle started before the main merge and passed in
525.202s, including command/source comparison (361.71s), emission and its existing
mutants. It is pre-merge evidence; the complete gate was started after the merge.
The full gate was stopped after more than ten minutes under the authorized
worker-gate exception, using SIGTERM on its identified process group. It exited
143. Its log is partial, not a full pass. The filtered worker gate was then
rerun from the merged state, with no competing broad gate:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./cmd/adamic ./internal/native -count=1 -timeout 15m > /tmp/wasm-exports-postmerge-packages.log 2>&1
ADAMIC_ORACLE_WASI=1 ADAMIC_GATE_UNCACHED=1 go test ./cmd/adamic ./internal/native ./internal/oracle -run 'TestWASI|TestRequestNativeStaysCommand' -count=1 -v -timeout 15m > /tmp/wasm-exports-postmerge-worker-gate.log 2>&1
```

Full post-merge touched packages passed: cmd/adamic 2.008s and internal/native
105.241s. The focused worker gate's command package passed in 6.727s; its native
WASI target checks passed in 0.009s. The older native prototype TestWASI skipped
because it requires the separate ADAMIC_TEST_WASI opt-in; the command's reactor
oracle and the required internal/oracle WASI tests use ADAMIC_ORACLE_WASI.
The post-merge required WASI oracle passed in 183.288s, including command/source
agreement, emission, and its byte/exit/runner mutants. The complete focused worker
gate exited 0. This is the re-green result for landing after merging origin/main.

Every requested mutant was applied independently and restored:

| Mutant | Observed failure | Log |
| --- | --- | --- |
| Wrapper string result length reduced by one | deepStrictEqual stringValue: empty string versus ! | /tmp/wasm-exports-mutant-length.log |
| Compiler f64 padding removed | numbers result decode: Truncated ABI result | /tmp/wasm-exports-mutant-alignment.log |
| Wrapper never releases result handles | memory at call 0 grows 43,122,688 to 43,515,904 | /tmp/wasm-exports-mutant-release.log |
| Record encoding field order reversed | recordValue score is +0 instead of -0 | /tmp/wasm-exports-mutant-order.log |
| Named unsupported signature accepted by selection | TestWASIExportSelection: nil error instead of bad.value refusal | /tmp/wasm-exports-mutant-refusal.log |

All five executions exited 1 through their intended assertions/runtime checks;
none was stopped by clang warnings. The first leaked-result mutant survived the
original short-call sweep: large leaked warmup results left reusable allocator
capacity and hid subsequent small leaks. The oracle was strengthened to include
a large number-array result every 256 calls. The rerun then failed immediately
on memory growth. The strengthened normal sweep passes. An initial test expected
11 exports when the source had 10; that test expectation was corrected before
using its result. The final source now has 14 crossing exports.

The unit's exact prefix command is not supported by the pre-existing main.go
parser. Its run function dispatches before tsgo.go parses options; supporting
--reactor before entry.a requires a main.go hook outside the listed ownership.
Approval for that narrow hook was requested, with no response received. The concrete
main.go switch case to insert before its existing wasm build case is:

```go
case len(arguments) >= 7 && arguments[0] == "build" && arguments[1] == "--target" && arguments[3] == "--reactor" && arguments[5] == "-o":
    return build(arguments[4], arguments[6], append([]string{"--target", arguments[2], "--reactor"}, arguments[7:]...))
```

This edit has not been applied or tested because main.go is unowned. The usable
invocation is:

```sh
adamic build --target wasm32-wasi entry.a -o out.wasm --reactor --abi-json out.json
```

Ambiguous function declaration names across modules remain explicitly refused
because the IR lacks source identity. Deliberately retained dynamic module globals
are instance state, not per-call temporaries; the lifetime oracle uses globals
with scalar state only. General malicious host buffers, allocation failure limits,
concurrent execution, Cloudflare deployment and all possible WASI import combinations
were not covered. Hosts must discard an instance after uncaught throw/panic
(proc_exit 70) or an engine trap. Existing command tests cover uncaught request
throws; the reference host implements this discard rule.
