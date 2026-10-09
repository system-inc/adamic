Built: addressed WASI function frames; both formerly frame-less recursion fixtures now match Node's loud panic.
Commits: emitted-C fix 8bea81a0d454abe4628acf953d49d7d3f10857b5; requested main/runtime merges 5a32755 and f5fb307.
Results: final uncached WASI oracle 299 pass, 3 fail out of 302; native packages and six native fixture comparisons pass.
Mutants: omitted frame, inverted stack check and oversized stack each compiled and failed the intended oracle check.
Not covered: complete repository gate, forthcoming codex/wasi-agreement changes, Workers deployment, custom smaller engine stacks and runtime-internal recursion.

Requested baseline: origin/main e011f8f was merged into codex/wasm32-driver, followed by runtime 52959fc. Those are the requested baseline merges, not runtime fixes authored by this unit. The only production-code change after them is internal/native/emit_functions.go. No internal/native/runtime/, internal/native/wasm/, internal/native/wasm_test.go or docs/wasm.md change was authored or committed. Initial local edits to the runtime-owned test and docs were restored on the scope change.

| Fixture | Side needed for the reproduced defect | Current outcome |
| --- | --- | --- |
| closures_throw.a | Runtime only: sort.c order(), sorted_or_stopped(), adamic_timsort(). | Still exit 70 at the first throwing sort comparator; Node catches it and exits 0. |
| stack_forever.a | Ours, emitted C: missing linear-memory frame. Existing runtime guard in 52959fc works once it has a frame to measure. | Exact Node stdout, stderr and exit 70. |
| stack_tail_call.a | Ours, emitted C: same missing frame. Existing runtime guard is sufficient for this reproducer. | Exact Node stdout, stderr and exit 70. |
| write_stdout_order.a | Runtime only: input.c adamic_write_text_file() reopens a pipe through /dev/stdout. | Still returns Error instead of writing second. |
| write_stderr_order.a | Runtime only: input.c adamic_write_text_file() reopens a pipe through /dev/stderr. | Still loses second on stderr because the returned Error is discarded. |

No driver-flag defect was reproduced. Stack flags already reserve 131,072 bytes and disable sibling-call optimization. A second runtime implementation is not required to fix these two specific stack reproducers. Runtime's forthcoming independent depth guard may complement or supersede this frame strategy. Integration of that guard, and the runtime-owned TestWASI/documentation updates, belongs to 01a1143a on codex/wasi-agreement. The generated ADAMIC_CHECK_STACK() invocation is deliberately retained so a future runtime entry guard is not bypassed. No new runtime signature or target macro was invented.

Before the fix, optimized stack_forever C reads __stack_pointer and compares it with adamic_stack_limit but never subtracts a frame before calling itself. V8's separate engine stack fills while the linear pointer never moves. The engine trap bypasses adamic_panic and output flushing: stdout is empty, exit is 1 and stderr is a raw Wasm RangeError trace.

The emitted WASI branch now declares an addressed volatile 32-byte local array, stores its first and last bytes and checks its address against the existing runtime limit. Its address use prevents scalar replacement and the volatile stores keep the memory access observable under -O2. Native retains its existing runtime macro. The optimized Wasm now subtracts 32 from __stack_pointer on entry and restores it on return. This applies to every emitted function body, including closures and region variants. The runtime's adamic_stack_overflow() remains unchanged and flushes stdout before the exact panic text and exit 70.

Named target bound: wasm32-wasi uses a 128 KiB linear stack, the existing 16 KiB panic margin and a minimum 32-byte emitted function frame. Its recursion depth differs from Node; its termination in these probes does not. Node 24.19.0 with the default engine stack is the verified host. Smaller engine stack configurations and recursion entirely inside runtime/library C remain outside this generated-function guarantee. The forthcoming runtime depth guard has not been tested here.

Exact stack observations:
- stack_forever.a: Node and fixed WASI stdout are "start\n"; stderr is "adamic: panic: RangeError: Maximum call stack size exceeded\n"; exit 70.
- stack_tail_call.a: Node and fixed WASI stdout are "0\n"; the same stderr and exit 70.
- Both pre-fix WASI artifacts had empty stdout, raw engine RangeError stderr and exit 1. Complete unmodified traces, generated C and optimized assembly are archived under the fixture directories.

Runtime handoffs contain the exact original source, complete Node/WASI stdout, stderr and exit code, and the function at fault:
- [closures_throw.a](closures_throw/handoff.md)
- [write_stdout_order.a](write_stdout_order/handoff.md)
- [write_stderr_order.a](write_stderr_order/handoff.md)

The snapshots are byte-identical to internal/oracle/testdata sources; their hashes are in results.json. Raw streams are also stored beside each program. observations.json includes both pre-fix and fixed observations for all five fixtures.

For the stream failures, instrumentation retained the oracle's piped stdout/stderr and root preopening. path_open fd 3 with dev/stdout or dev/stderr returned WASI errno 44, ENOENT, in both cases. The first console line had already been flushed correctly. The failure is reopening a pipe via a filesystem path, rather than emitted arguments or output ordering in the driver. An initial trace used regular-file redirection and path_open succeeded; it was discarded and repeated with pipes. The retained trace asserts identical stdout/stderr to the actual oracle observation. Runtime owns fd handling, output order, UTF-8 encoding and borrowed descriptor lifetime.

All tests wrote directly to logs; none was piped. Setup:

    bash cloud/setup.sh --wasi-sdk > /tmp/wasm-diagnosis-setup.log 2>&1
    source /workspace/adamic-tools/env.sh

Timing: Go ready 0s; clang ready 1s; Node ready 1s; WASI SDK ready 1s; submodules ready 1s; build cache warm 80s; done 80s. nproc is 5. Go 1.27.1, native clang 20.1.8, SDK 27 clang 20.1.8-wasi-sdk, Node 24.19.0.

Baseline reproduction:

    ADAMIC_ORACLE_WASI=1 ADAMIC_GATE_UNCACHED=1 go test -json -count=1 -timeout 30m ./internal/oracle -run '^TestWASIAgreesWithNode$/internal/oracle/testdata/(closures_throw|stack_forever|stack_tail_call|write_stdout_order|write_stderr_order)[.]a$' > /tmp/wasm-five-before.json 2>&1

All five failed, as integration reported. Node's source observations were read first. Frozen pre-fix and post-fix compiler binaries separately built each exact source with build --target wasm32-wasi; Node source and WASI runs used subprocess pipes, captured as raw bytes.

Final commands:

    ADAMIC_ORACLE_WASI=1 ADAMIC_GATE_UNCACHED=1 go test -json -count=1 -timeout 30m ./internal/oracle -run '^TestWASIAgreesWithNode$|^TestWASIOracleCatchesMutants$' > /tmp/wasm-diagnosis-final.json 2>&1
    ADAMIC_ORACLE_WASI=1 ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./cmd/adamic ./internal/native > /tmp/wasm-diagnosis-packages.log 2>&1
    ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout 30m ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/(closures_throw|stack_overflow|stack_forever|stack_tail_call|write_stdout_order|write_stderr_order)[.]a$' > /tmp/wasm-diagnosis-native-oracle.log 2>&1
    ADAMIC_ORACLE_WASI=1 go test -v -count=1 -timeout 10m ./cmd/adamic -run 'TestWASIRequest|TestRequestNativeStaysCommand' > /tmp/wasm-diagnosis-handler-final.log 2>&1
    go vet ./... > /tmp/wasm-diagnosis-vet.log 2>&1
    gofmt -l cmd internal > /tmp/wasm-diagnosis-format.log

Main added two fixtures since the prior 300-case report. The final 302-case WASI result is 299 pass and 3 genuine runtime failures, with no skips or expected-failure changes. The command returns exit 1 because of those three. TestWASIOracleCatchesMutants passes: its existing emitted stdout and exit-code mutations are caught. Exact final elapsed time is recorded in results.json.

Touched packages passed: cmd/adamic 2.963s, internal/native 88.119s. The six selected native fixtures all passed uncached in 1.454s, with Node, JavaScript, release, ASan/UBSan and leak comparisons. Final handler controls passed in 1.508s: 10,000 requests, flat 524,288-byte memory, two persistent globals and 630,000 region objects, plus uncaught handler panic exit 70. Vet and formatting logs are empty. Native -O2 objects for both stack reproducers are byte-identical before and after when compiled at the same source pathname; native-object-comparison.json records the result.

Runtime-owned test, deliberately unchanged:

    PATH=/workspace/adamic-tools/wasi-sdk/bin:$PATH ADAMIC_TEST_WASI=1 go test -v -count=1 -timeout 15m ./internal/native -run '^TestWASI$' > /tmp/wasm-diagnosis-runtime-test.log 2>&1

It returned exit 1 in 24.518s only because its two stack exception assertions still require WASI exit 1 and empty stdout. Their diagnostics show Node and WASI now have identical stdout, stderr and exit 70. Its other 30 equivalent fixture checks and explicit sort limitation passed; the reserved-stack canary remains intact. Its 100,000-request control passed with zero live values, flat 393,216-byte memory and 6,300,000 region objects. The runtime worker must replace those two obsolete expected-mismatch assertions with ordinary exact comparisons and update its docs. No test failure was hidden.

Mutants each ran independently from the final source and were restored using the exact original bytes:

    ADAMIC_ORACLE_WASI=1 ADAMIC_GATE_UNCACHED=1 go test -json -count=1 -timeout 10m ./internal/oracle -run '^TestWASIAgreesWithNode$/internal/oracle/testdata/(stack_forever|stack_tail_call)[.]a$'

| Mutant | Catching check |
| --- | --- |
| Omit the addressed WASI frame, leaving the ordinary entry macro | Both stack fixtures: engine exit 1 instead of panic 70 and lost stdout. |
| Invert the WASI frame-address comparison | stack_tail_call.a: shallow count(1000) panics before printing 0, stdout differs. stack_forever.a alone cannot catch this mutant. |
| Restore the old 8 MiB linker stack | Both stack fixtures: V8's engine limit arrives before the linear guard, exit 1 instead of 70. |

All three compiled and returned test exit 1 from those comparisons, not from warnings or build failures. Raw compressed JSON logs preserve every result.

Scope note: runtime fixes for the throw path, fd_write ordering and a runtime depth check belong to the named runtime worker. This branch contains only the requested baseline runtime merge, this isolated emitted-C fix and diagnostic evidence. No runtime-owned edits from the temporary local test/doc changes remain.
