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
large summaries and oversized bodies. It writes the request file and expected response
file for the native check. Node imports the same `.a` file through type stripping.
Independent expected-response pins hold the health response, tax/rounding and Unicode
escapes; Node's JSON.parse and default string sort additionally judge malformed inputs
and successful listings.

Each release or counted run initializes one WASI reactor, with no filesystem preopens,
once. It warms that instance with all 100,000 generated requests, then measures another
100,000 calls on the same instance. The host allocates borrowed UTF-8 input, copies the
owned response through the ABI, and releases both. Every measured response must equal
Node exactly, and every measured memory capacity must equal the post-warmup capacity.
The counted run also requires zero live counted values after every warmup and measured
request. This is reusable memory capacity, not a claim that memory shrinks.

Run:

```sh
bash cloud/setup.sh --wasi-sdk > /tmp/wasm-requests-setup.log 2>&1
source /workspace/adamic-tools/env.sh
ADAMIC_TEST_WASI=1 go test ./internal/native -run '^TestWASIService$' -count=1 -v -timeout 45m > /tmp/wasm-service-test.log 2>&1
```

The test invokes `adamic build --target wasm32-wasi` twice (release and `--count`),
serves requests through the compiler's existing reactor ABI, and builds a temporary
native `.a` command importing the same service with `adamic build --sanitize`. The
native command reads the identical JSONL request file and its full stdout must match
Node's response file, with ASan, UBSan and LeakSanitizer enabled. Each mutant is built
from an isolated temporary `.a` source while the Node oracle keeps the original source.
A mutant must compile successfully and fail its named runtime assertion.

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
observations on a shared machine, not HTTP throughput or a speedup.

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
ADAMIC_TEST_WASI=1 go test ./internal/native -run '^TestWASIService$' -count=1 -v -timeout 45m > /tmp/wasm-service-land-service.log 2>&1
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
