Built: automatic WASI handler exports, reactor initialization and persistent module globals; native stays a command.
Commits: driver 282c873e6a444d98ce476fe8b4c71f288c3d6d80; runtime 52959fc merged only in scratch.
Results: merged WASI oracle 295 pass, 5 fail out of 300; handler controls pass, 10,000 requests keep memory at 524,288 bytes.
Mutants: 17 applicable mutants caught; two initial survivors investigated and resolved below.
Not covered: complete repository gate, Workers deployment, arbitrary handler forms, Wasm sanitizers, and five runtime/platform failures.

This follows [report.md](report.md). Its blocked-runtime observation and 8 MiB stack are historical. No runtime C or header was edited or committed on the driver branch.

The command stays adamic build --target wasm32-wasi <file> -o <out.wasm>. The entry module must declare export function handleRequest(request: string): string. Private and dependency-only handlers do not select reactor mode. Validation rejects generics, optional/default/rest parameters, narrow literal input types, extra parameters, non-string results and function values. Export aliases remain refused by existing lowering. Flat IR names cannot identify a source module: collisions with dependency functions are refused rather than selecting an arbitrary function. Rename the conflicting dependency function. No lowering file changed.

The boundary is exactly adamic_string *adamic_request(const unsigned char *bytes, size_t length), exposed as (i32, i32) -> i32. It decodes borrowed UTF-8, calls the handler, releases the decoded request, flushes output and returns the owned response. Uncaught exceptions use runtime panic exit 70. Reuse consumes object and array parameters, never strings, so the ordinary string calling convention applies.

Handlers use -mexec-model=reactor and export memory, malloc, free, adamic_request, adamic_response_bytes, adamic_response_length and adamic_release. Counted builds additionally export the runtime contract's adamic_live and adamic_regions. The host initializes once, allocates/writes input, calls the handler, copies response bytes, releases the response and frees input. C's entry remains main. A constructor runs it at WASI _initialize, after runtime constructors by linking the runtime archive before the program object. Globals remain owned for the instance lifetime. Ordinary WASI commands retain crt1's _start; native emission and linking retain their existing behavior.

The linear stack is now 131,072 bytes, matching runtime's tested contract. With the former 8 MiB setting, the first runtime integration passed 293/299 fixtures; stack_overflow.a reached the engine limit before the linear guard. At 128 KiB it passed 294/299. This observed fix does not prove every recursive function has a linear-memory frame.

The host fixture imports a dependency, creates two dynamic string globals, and exports a handler without a warmup call in source. Responses exercise a persistent counter, returning the decoded input directly, parameter reassignment, and a 63-node statement-region tree. The same source on Node is the response oracle. Inputs include empty strings, Unicode, NUL, BOM, a 1,024-byte string and three malformed UTF-8 encodings. The host checks borrowed bytes unchanged, required exports, two parameters and numeric pointer results, initialization stdout, byte equality, one owned response allocation and the original live count after release.

Committed controls passed in 1.227s: 1,000 warmup calls and 10,000 measured calls, flat 524,288-byte memory, 630,000 region objects reclaimed, two live module globals after every response release. The native command printed only dependency and entry initialization, then released everything under ASan/UBSan. A separate throwing handler printed exactly adamic: panic: Error: boundary plus newline and exited 70. This does not claim zero live allocations for persistent globals or a Worker deployment.

All test output went directly to logs. Toolchain refresh ran bash cloud/setup.sh --wasi-sdk, then source /workspace/adamic-tools/env.sh. Timing lines: Go ready 0s; clang ready 1s; Node ready 1s; WASI SDK ready 1s; submodules ready 1s; build cache warm 75s; done 75s. nproc printed 5. Go 1.27.1, clang 20.1.8, SDK 27 clang 20.1.8-wasi-sdk, Node 24.19.0. The SDK URL, extraction command and optional setup flag are in the earlier report.

Scratch /workspace/scratch/wasm32-driver on codex/wasm32-driver-test merged origin/codex/wasm32-runtime at 52959fc, then the committed local codex/wasm32-driver branch. Final scratch merge e2fc5b776880d1390671408d238b888b3520c85f contains both runtime 52959fc and driver 282c873, with a clean tree before testing. Runtime 52959fc is not an ancestor of the driver branch.

Commands, using the sourced tool environment:

    ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./cmd/adamic ./internal/native > /tmp/wasm-request-native-packages.log 2>&1
    go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -args -update-counts > /tmp/wasm-request-counts.log 2>&1
    ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout 30m ./internal/oracle -run '^TestNativeAgreesWithNode$/cmd/adamic/testdata/wasi/request[.]a$' > /tmp/wasm-request-native-oracle.log 2>&1
    ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout 30m ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/(stack_overflow|stack_tail_call|stack_forever|exceptions_uncaught|regions_throw)[.]a$' >> /tmp/wasm-request-native-oracle.log 2>&1
    go test -count=1 ./cmd/adamic ./internal/native -run 'TestWASIRequestSelection|TestWASIRequestModuleSelection|TestWASIRefusesUnsupportedOptions|TestRequestNativeStaysCommand' > /tmp/wasm-request-checks-final.log 2>&1
    go vet ./... > /tmp/wasm-request-vet.log 2>&1
    gofmt -l cmd internal > /tmp/wasm-request-gofmt.log

Both touched packages passed: cmd/adamic 2.254s, internal/native 74.253s. Counts update passed in 13.848s and added the fixture row. Six native oracle fixtures passed uncached, including the new source fixture, with release, ASan/UBSan, JavaScript and leak comparisons. Final focused checks passed. Vet and gofmt logs were empty. Both host scripts passed node --check; git diff --check passed.

Committed scratch commands:

    ADAMIC_ORACLE_WASI=1 go test -v -count=1 -timeout 10m ./cmd/adamic -run 'TestWASIRequest|TestRequestNativeStaysCommand' > /tmp/wasm-request-committed-host.log 2>&1
    ADAMIC_ORACLE_WASI=1 ADAMIC_GATE_UNCACHED=1 go test -json -count=1 -timeout 30m ./internal/oracle -run '^TestWASIAgreesWithNode$|^TestWASIOracleCatchesMutants$' > /tmp/wasm-request-committed-oracle.json 2>&1

The final oracle passed 295 of 300 fixtures and failed five in 82.144s. The Go test command returned exit 1 because of those five disagreements. TestWASIOracleCatchesMutants passed, catching the emitted stdout and exit-code mutants. Oracle fixtures are built as commands and compared byte for byte on stdout, stderr and exit. The new source adds one command fixture; the separate driver host test exercises its reactor. No failures were skipped or marked expected.

| Fixture | Observed disagreement |
| --- | --- |
| closures_throw.a | Explicit runtime panic for a throwing sort comparator, WASI exit 70 versus Node 0. |
| stack_tail_call.a, stack_forever.a | Frame-less recursion reaches engine RangeError, exit 1 versus expected panic 70, and loses buffered stdout. |
| write_stdout_order.a | /dev/stdout is unavailable in WASI; stdout differs. |
| write_stderr_order.a | /dev/stderr is unavailable in WASI; stderr differs. |

Each applicable mutant ran separately in scratch with ADAMIC_ORACLE_WASI=1 go test -v -count=1 -timeout 5m ./cmd/adamic -run <test>. Native request-mode refusal used ./internal/native. Each returned exit 1 from its intended assertion, with no clang warning or Go build failure. All mutations were restored.

| Mutant | Catching check |
| --- | --- |
| Omit decoded request release | TestWASIRequest: only the owned response may remain. |
| Shorten nonempty UTF-8 input | TestWASIRequest: response differs from source on Node. |
| Shorten response length | TestWASIRequest: response differs from source on Node. |
| Skip module constructor initialization | TestWASIRequest: initialization output must match at _initialize. |
| Release module globals | TestWASIRequest: globals must remain alive. |
| Omit adamic_request export | TestWASIRequest: required export missing. |
| Select command instead of reactor | TestWASIRequest: _initialize export missing. |
| Accept wrong input type | TestWASIRequestSelection: direct selection accepted number/literal input. |
| Accept wrong result type | TestWASIRequestSelection: direct selection accepted number result. |
| Accept generic declaration | TestWASIRequestSelection: unused T with ordinary string input accepted. |
| Accept two parameters | TestWASIRequestSelection: direct selection accepted extra parameters. |
| Accept default parameter | TestWASIRequestSelection: default initializer accepted. |
| Omit uncaught-exception boundary | TestWASIRequestThrows: returned across ABI instead of panic exit 70. |
| Accept request mode on native | TestWASIRefusesUnsupportedOptions: Options{Request:true} accepted. |
| Accept ambiguous function names | TestWASIRequestModuleSelection: two modules' same-named functions accepted. |
| Mutate input after decoding | TestWASIRequest: borrowed request bytes changed. |
| Release response before return | TestWASIRequest: owned response allocation no longer remains. |

Initial survivors: dropping a proposed consumed-parameter retain was unreachable because string parameters are never consumed by reuse; that speculative branch was removed. Dropping generic validation with a T-typed input was masked by string-type validation; replacing the test with unused T and string input killed the same mutant. Initial and final logs are retained. The merged WASI command oracle also runs its existing emitted stdout and exit-code mutants.

Cohere could not check the new .a paths: the current tsconfig include list excludes cmd/adamic/testdata and cohere's outside-program diagnostic treats .a as unsupported. A project-wide explicit-tsconfig attempt, from the cohere checkout without Adamic's lint configuration, returned 1,385 findings and did not cover these paths; it is not a passing gate. Adamic's checker, native oracle and Node handler oracle did check the new source. No cohere, tsconfig or unrelated source changed. An early native oracle selector matched no fixtures; it was corrected, and the six actual passing names are logged.

An early mutant restore matched an existing command flag instead of the mutated reactor flag and broke a temporary throwing-handler control. Restoration was changed to use the exact original bytes, scratch sources were restored from the driver, and all committed controls subsequently passed.

Evidence is archived beside this report. The complete gate and Workers deployment remain untested. Five disagreements require runtime/platform work; none is reported as passing.
