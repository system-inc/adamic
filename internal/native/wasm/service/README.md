Built a six-route, stateless Adamic service and a seeded 100,000-request Node/WASI oracle.
Claim commit: 7c2d29a; implementation and final branch SHA are reported with the delivery.
WASI release and counted requests passed; final package and oracle results are recorded below.
Mutants change a response byte, retain a per-request Parser, and corrupt a Unicode escape.
Not covered: deployment, concurrent calls, exhaustive JSON conformance, or the complete repository gate.

The handler exports `handleRequest(request: string): string`. Routes are `GET /health`,
`GET /catalog`, `POST /orders`, `POST /summary`, `POST /listing`, and `POST /quote`.
Orders validate a customer and 1..128 items against a price map. Quantities are integers
1..100; totals use integer cents and tax is rounded at 8.25%. Summary counts words in
request-local maps and returns a sorted Unicode summary. Listing sorts UTF-16 strings
stably. Quote validates a bounded numeric amount and rounds cents and tax. Responses
carry numeric status codes: 200, 201, 400, 404, 413, or 422.

`JSON.parse` is not compiled by this compiler. `service.a` contains a bounded recursive
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

Machine: shared Linux container, AMD EPYC 9V74, Node 24.19.0, Go 1.27.1, clang 20.1.8,
WASI SDK 27. `nproc` = 5; cgroup CPU quota = 400000/100000 (four CPUs); memory limit
16 GiB. Setup printed Go/clang/Node ready at 0s, WASI SDK/submodules ready at 3s,
build cache warm at 109s, done in 109s. Environment: `/workspace/adamic-tools/env.sh`,
which sets WASI_SYSROOT to `/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot`.

Initial counted observation: 348,337-byte module, memory 1,638,400 bytes throughout,
live = 0, regions = 0, Wasm 7,878 requests/s, Node 47,696 requests/s; one-minute load
6.25 at start and 3.93 at end. Rates include exact-response assertions; Wasm includes
UTF-8 boundary copies, allocation/release, memory and lifetime probes. These are
synchronous service observations on a shared machine, not HTTP throughput or a speedup.
