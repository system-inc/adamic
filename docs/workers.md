# Cloudflare Workers

Run `adamic worker entry.a --out output`. The output directory contains
`handler.mjs` (the JavaScript backend's program with entry function exports),
`adamic.mjs` (the Workers runtime), `worker.mjs` (the generated bridge), and
`wrangler.toml` (main, entry base name, and pinned compatibility date).
Run wrangler from that directory. No `nodejs_compat` is needed.

The entry must export a synchronous function `handle(request: HttpRequest): HttpResponse`,
using the types from `'adamic/http'`. The command checks the signature before writing
anything. Module statements run once at startup. Exported generic functions are
refused; export a concrete function instead.

The bridge passes `method` and the whole `url` unchanged, `path` as the URL's
percent-encoded pathname, decoded query fields in search parameter order,
lowercase headers in Request header iteration order, and `body` from
`await request.text()`. It appends response headers in the handler's order and
returns the handler's status and body. Headers and Response retain the platform's
normalization rules, including combining repeated header names.

An `AdamicPanic` is logged with `console.error(message)` and becomes status 500
with body `internal error`; later requests still run. Other errors propagate.
File and argument functions panic with `<name>: not available on Workers`.
The Workers UTF-8 functions use TextEncoder; the independent Node oracle uses Buffer.

`node internal/worker/run.mjs output/worker.mjs requests.jsonl` prints one JSON
response per request, with status, headers as ordered pairs, and body. Requests
provide method, URL, headers as pairs, and an optional body. Tests use the same
bridge against source stripped by Node and the compiled handler.

Run `ADAMIC_WORKER_MUTANTS=1 go test -count=1 -v ./internal/worker -run '^TestWorkerMutants$'`
to prove the named checks catch each mutant in a scratch compiler copy. Ordinary
test runs skip this opt-in test.

Use `adamic worker entry.a --out output --wasm sieve,stats` to run the named
entry functions in Wasm. Private function declarations can be selected too.
The JavaScript handler calls generated ABI crossings; its selected function bodies
are replaced with calls. Without `--wasm`, the four generated files stay identical.

Wasm output additionally contains `handler.wasm`, `abi.json`, `crossing.mjs`,
`wasm-state.mjs`, `wtf8.mjs`, `crossing-runtime.mjs`, and `wasi.mjs`. Wrangler uses
its CompiledWasm rule. Building requires Node 24 for crossing generation and the
WASI SDK (`bash cloud/setup.sh --wasi-sdk`, then source the printed environment).
The [Wasm ABI](wasm-abi.md) defines the supported signatures and ownership.

Selection requires top-level entry declarations with ABI version 1 signatures.
The interim purity check follows direct and bounded callback callees and refuses
console, Adamic file/argument I/O, module-global reads and writes, and unbounded
function-value calls. Every refusal happens before the output directory is touched.

One crossing object lives for the isolate. It initializes its Wasm instance on the
first call, runs module top level once, and frees inputs and returned result handles.
A Wasm panic becomes the same logged 500 as a JavaScript panic; the crossing discards
the terminated instance and initializes a fresh instance for the next call.
Other thrown errors propagate. Strings preserve UTF-16 code units across the
boundary, including lone surrogates.

With `WASI_SYSROOT` configured, `go test ./internal/worker -run TestWasm -count=1`
compares the same generated bridge over Wasm and stripped source, checks refusals,
and verifies 10,000 requests have flat linear memory and zero live values.
Set `ADAMIC_WORKER_MUTANTS=1` and select `TestWasmWorkerMutants` to run the four
Wasm integration mutants in scratch compiler copies.
