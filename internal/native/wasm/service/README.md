Built a six-route, stateless Adamic service and a seeded 100,000-request Node/WASI oracle.
Commits: claim 7c2d29a, implementation 098e052, framing fix 03e1a87, main merge f0f36df.
WASI/native request comparisons, the native package, filtered WASI oracle, vet and format checks pass.
All three mutants compiled and failed: changed byte, retained Parser (11 live), Unicode escape (+1).
Not covered: deployment, concurrent calls, exhaustive JSON conformance, or the complete repository gate.

The handler exports `handleRequest(request: string): string`. Routes are `GET /health`,
`GET /catalog`, `POST /orders`, `POST /summary`, `POST /listing`, and `POST /quote`.
Orders validate a customer and 1..128 items against a price map. Quantities are integers
1..100; totals use integer cents and tax is rounded at 8.25%. Summary counts words in
request-local maps and returns a sorted Unicode summary. Listing sorts UTF-16 strings
stably. Quote validates a bounded numeric amount and rounds cents and tax. Responses
carry numeric status codes: 200, 201, 400, 404, 413, or 422.

`JSON.parse` is refused by the compiler because its result type cannot be proven from text. `service.a` contains a bounded recursive
descent parser for objects, arrays, strings, numbers, booleans and null, with errors as
values. It recognizes JSON whitespace, number grammar, all string escapes, and four-hex
UTF-16 escapes (including surrogate pairs and lone surrogates). Duplicate fields use the
last value. Requests over 65,536 UTF-16 units, nesting over 24, and nonfinite numbers are
rejected by service policy. Serialization uses compiled `JSON.stringify`. Every request
owns its parser, syntax values, maps, arrays and OrderLine class instances; there is no
module-level application state. The compiler currently selects no statement regions
for this program, so the region advance assertion is conditional on emitted C using them.

`host.mjs` seeds an LCG with `0x6a09e667` and generates 100,000 request lines. It includes
valid orders, invalid fields/products/quantities, six routes, unknown routes, malformed
JSON, escaped and literal Unicode, lone surrogates, control characters, 128-item orders,
large summaries and oversized bodies. It streams them, holding one request at a time: each
generated request draws exactly one number from the LCG, in index order, so every pass
reseeds and generates the same sequence again, and a range skips the indices below it by
drawing their numbers without building their text. Holding all 100,000 at once peaked at
about 2 GiB per host. With module `-` it steps through all 100,000 and writes only the
requests in its range, `[start, end)`, and Node's responses to them, as JSONL files for the
native check, reporting the range and the lines written. Node imports the same `.a` file
through type stripping. Independent expected-response pins hold the health response,
tax/rounding and Unicode escapes; Node's JSON.parse and default string sort additionally
judge malformed inputs and successful listings, on every request either mode generates.

Each release or counted run initializes one WASI reactor, with no filesystem preopens,
once, and writes no files. It warms that instance with the first 1,000 generated requests,
which hold all 20 request shapes 50 times each, then measures all 100,000 calls on the same
instance. The host allocates borrowed UTF-8 input, copies the owned response through the
ABI, and releases both. Every measured response must equal Node exactly, and every measured
memory capacity must equal the post-warmup capacity. The counted run also requires zero
live counted values after every warmup and measured request. This is reusable memory
capacity, not a claim that memory shrinks. A plain pass then answers the same requests in
Node again, each held to the length and 32-bit FNV-1a hash of its measured pass response,
and the two passes' summed response lengths must agree. The wasm rate times each handle
call and its comparison with Node's response; the Node rate times each `handleRequest` call
and its add to the checksum. Generating requests, computing the expected responses, the
JSON.parse checks, the live and memory probes, and hashing fall outside both.

Run:

```sh
bash cloud/setup.sh --wasi-sdk > /tmp/wasm-requests-setup.log 2>&1
source /workspace/adamic-tools/env.sh
ADAMIC_TEST_WASI=1 go test ./internal/native -run '^TestWASIService' -count=1 -v -timeout 45m > /tmp/wasm-service-test.log 2>&1
```

The oracle runs as separate tests, so the gate can hold each to its 60 s per-test budget, and
no test builds anything: the tree builds two products ahead, outside any test's budget.
`TestProduct_WASIServiceModules` builds the release and counted modules, one counted module per
source mutant, and whether the emitted C uses statement regions; `TestProduct_WASIServiceNative`
builds `command.a`, a native command importing the same service, once with `adamic build --sanitize`.
`TestWASIService` (release) and `TestWASIServiceCounted` serve the requests through the compiler's
existing reactor ABI, and `TestWASIServiceNativeAgrees_0` to `_7` each have the host generate
the requests and Node's responses only (module `-`), for its own range alone. Each native part's
host steps through all 100,000 and writes one eighth, 12,500 lines, whose report must name that
range and line count, and the command's stdout must match Node's response file line for line,
with ASan, UBSan and LeakSanitizer enabled. All 100,000 in one sanitized run took 57.5 s on an idle
Linux box (wasm legs 23 to 24 s, mutants 23 s), and a quarter about 14 s, past the 30 s aim on a
loaded gate box, so every request is checked under ASan in eight instances rather than one; the
counted wasm leg still holds live 0 over all 100,000 in a single instance.
`TestWASIServicePartsTile` requires one test per part and refuses parts with a gap, an overlap, a
short end or an empty range. Each mutant is built from an isolated temporary `.a` source while the
Node oracle keeps the original source. A mutant must compile successfully and fail its named runtime
assertion (`TestWASIServiceMutantResponseByte`, `TestWASIServiceMutantRetainedObject` and
`TestWASIServiceMutantUnicodeEscape`). Every leg refuses a host report of fewer than 100,000 requests,
and `TestWASIServiceMutantShrunkCount`, a host generating 1,000, proves that refusal fires.

Machine: shared Linux 6.18.44 container, AMD EPYC 9V74, Node 24.19.0, Go 1.27.1, clang 20.1.8,
WASI SDK 27. `nproc` = 5; cgroup CPU quota = 400000/100000 (four CPUs); memory limit
16 GiB. Setup printed Go/clang/Node ready at 0s, WASI SDK/submodules ready at 3s,
build cache warm at 109s, done in 109s. Environment: `/workspace/adamic-tools/env.sh`,
which sets WASI_SYSROOT to `/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot`.

Final observations after merging origin/main f8013f0baac41ddc340d76f83bddde38536a8f07:

| Build | Module bytes | Wasm requests/s | Node requests/s | One-minute load, start/end |
| --- | ---: | ---: | ---: | --- |
| Release | 347,557 | 7,863 | 32,032 | 2.74 / 2.85 |
| Counted | 348,518 | 7,088 | 47,326 | 2.71 / 3.72 |

Both served 100,000 measured requests in one instance after 100,000 warmup requests.
Memory minimum and maximum were both 1,638,400 bytes. The counted run had zero live
values after every generated request, and zero region objects (no regions emitted).
The native package and WASI oracle were running concurrently during these service
measurements. Rates include exact-response assertions; Wasm includes UTF-8 boundary
copies, allocation/release, memory and lifetime probes. These are synchronous service
observations on a shared machine, not HTTP throughput or a speedup. They predate the
streaming host, whose rates time only the handle calls and comparisons described above, so
they don't compare directly with its reports.

Every requested mutant compiled and failed at execution, against the original Node source:

| Mutant | Assertion that caught it |
| --- | --- |
| Change the health response's final service-name byte from `s` to `t` | Exact response pin: `orders` differs from `ordert`. |
| Append each request's Parser to a module-level array | First generated warmup request: 11 live counted values instead of 0. |
| Decode a JSON Unicode escape with `String.fromCharCode(code + 1)` | Unicode listing pin: `世界`, `🌍`, and `\ud800` differ from the corrupted values. |

The earlier native comparison exposed a generator framing error: a malformed JSON
string contained a raw newline, so the JSONL command read it as two requests. The first
difference was at response 32. The generator now uses a raw tab for that malformed
control-character case and asserts that every generated request occupies one line.
The corrected sanitized native command matches all 100,000 Node responses.

Verification after the main merge:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TEST_WASI=1 go test ./internal/native -run '^TestWASIService' -count=1 -v -timeout 45m > /tmp/wasm-service-land-service.log 2>&1
ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run '^TestWASI' -count=1 -timeout 30m > /tmp/wasm-service-land-oracle.log 2>&1
go test ./internal/native -count=1 -timeout 30m > /tmp/wasm-service-land-native.log 2>&1
go vet ./... > /tmp/wasm-service-land-vet.log 2>&1
gofmt -l cmd internal > /tmp/wasm-service-land-format.log
git diff --check
```

The service integration passed in 191.177s, including all 100,000 native sanitized
responses and all three compiled mutants. The native package passed in 85.399s; the
filtered WASI oracle passed in 151.679s.
Vet, formatting and diff checks produced no diagnostics. The service integration log
contains the full release/counted observations, native comparison and mutant failures.
The full repository test gate was not run. WASI sanitizer support, deployment, concurrent
execution, exhaustive JSON conformance and statement-region behavior are outside this
witness. The emitted-C region detector will require advancing counters if this compiler
starts choosing regions for the service.

Base: origin/wasm/integrate at 6f7dce3. This checkout initially fetched only main; fetching
`refs/heads/wasm/integrate:refs/remotes/origin/wasm/integrate` supplied the required base.
The first push contained only the claim. Current main was merged (never rebased) into
codex/wasm-requests as f0f36df, and all final checks above use that merge. Only the worker
branch is pushed; no pull request is opened.
