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

Run `go test -count=1 -v ./internal/worker -run '^TestWorkerMutants_'`
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
Select `TestWasmWorkerMutants` to run the four
Wasm integration mutants in scratch compiler copies.

## Test products and top-level families

Build test products before running behavior units (this invokes Go tests, not a gate):

```sh
export CC=clang CXX=clang++
go test ./internal/worker -run '^TestProduct_Worker(Compiler|Checks)$' -count=1
go test ./internal/worker -run '^TestProduct_WorkerWasmTransform$' -count=1
go test ./internal/worker -run '^TestProduct_WorkerWasm_[0-9]' -count=1
go test ./internal/worker -run '^TestProduct_Worker(Wasm)?Mutants_' -count=1
go test ./internal/worker -run '^Test(WorkerMutants_|Wasm)' -count=1
```

The slow Worker tests are FNV-64a families of top-level `<name>_000` wrappers.
`families_test.go` mirrors oracle's test-only generator, which cannot be imported.
`TestWorkerFamilyWrappersAreCurrent` checks the generated source;
`TestWorkerFamiliesOwnEveryCaseOnce` checks unique keys and exact-once ownership.
Regenerate wrappers with that first test and `-args -update-worker-family-wrappers`.
The one-case memory and multiple-function families retain their complete workloads.

`TestProduct_Worker*` builds input-keyed CLI/test executables, including each mutant,
and ordinary/counted Wasm Workers once per input hash. Behavior units only find
products and run fresh witnesses: verdicts, response corpora, and memory observations
are never cached. Compiler products use buildcache's Go dependency/overlay keys;
Wasm products also key the fixture, selected names, count mode, build recipe,
WASI runtime/header identity, pinned clang/archiver, and Node version.
Use the same CC/CXX configuration for building and reading products; Go test
otherwise injects its gcc/g++ defaults when those variables are unset, whereas
the repository toolchain defaults to clang/clang++. Both configurations are keyed.
Product order seeds the ordinary runtime before the other compiler processes run,
and counted memory products reuse the ordinary bridge instead of rebuilding it.
Both the immutable Wasm products and mutant binaries travel with the build cache;
a missing product fails with a build-ahead diagnostic rather than rebuilding inside
a behavior unit's clock. All 10,000 memory requests and mutant failure witnesses remain.

Worker tests run without opt-in variables. Build `TestProduct_` inputs ahead, as
ordered in `cloud/fast-gate/product-order.json`. Node 24 runs the generated
module and WASI shim; Go tests require neither workerd nor wrangler. Tests find
setup.sh's WASI SDK beside the tools directory on PATH (or under ADAMIC_TOOLS);
an explicit WASI_SYSROOT still wins. Missing required tools fail by name.
