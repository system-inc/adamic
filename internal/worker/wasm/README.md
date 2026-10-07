# Workers Wasm host

This host consumes the stable `handleRequest(string): string` reactor entry on
`origin/wasm/integrate` (6f7dce3). All Worker runtime files live here and import
neither Node modules nor a runtime compiler. The generator and tests run on Node.

Generate a Worker, then build its entry into the generated directory:

```sh
node internal/worker/wasm/generate.mjs /tmp/adamic-worker
adamic build --target wasm32-wasi path/to/entry.a -o /tmp/adamic-worker/handler.wasm
```

`worker.mjs` creates one host in module scope. `bridge.mjs` reads the body before
entering the synchronous host, preserving the HttpRequest field order and values:
method, whole URL, encoded pathname, ordered decoded query fields, iterated
headers and text body. It parses the returned HttpResponse JSON and appends its
headers in order. A panic gives status 500 and body `internal error`; WASI stderr
already logged the panic. Other execution failures are also logged and give 500.
The template's CompiledWasm rule makes `handler.wasm` a compiled module binding.
The generator does not compile or overwrite a handler artifact.

`host.mjs` exports `createHost(module, { logger })`. It initializes lazily on the
first call, then initializes again only after an execution failure. `call(text)`
is synchronous and rejects reentrancy. It copies UTF-8 into host-owned malloc
storage, calls the handler, reacquires memory after the response accessors,
copies out the bytes, releases the owned response and frees the input. Response
BOMs are preserved and TextDecoder replaces lone surrogate WTF-8 sequences.
Execution failures discard the entire instance without attempting cleanup in
terminated execution. The optional read-only `exports` getter supports counted
build diagnostics; it does not initialize an instance.

The observed import inventory from `WebAssembly.Module.imports` is:

| Artifact | Preview 1 imports |
| --- | --- |
| Driver request fixture | fd_close, fd_fdstat_get, fd_prestat_get, fd_prestat_dir_name, fd_seek, fd_write, proc_exit |
| Large response and panic fixtures | fd_close, fd_prestat_get, fd_prestat_dir_name, fd_seek, fd_write, proc_exit |

The shim has no preopens. Unknown descriptors return EBADF. Logging descriptors
are character devices with FD_WRITE rights; seeking returns ESPIPE. Writes on
1 and 2 decode incrementally and buffer lines for console.log and console.error.
proc_exit flushes partial lines and throws WASIExit with an unsigned exit code.
Unsupported imports fail before instantiation and name the module and import.
Filesystem handlers require explicit additional host behavior and are not covered.

## Verification

```sh
bash cloud/setup.sh --wasi-sdk
source /workspace/adamic-tools/env.sh
export PATH=/workspace/adamic-tools/wasi-sdk/bin:$PATH
ADAMIC_TEST_WASI=1 go test ./internal/worker/wasm -count=1 -v -timeout 15m \
  > /tmp/workers-wasm-test.log 2>&1
```

Use the environment and SDK paths printed by setup on another machine. Without
opt-in, Node 24, Go, WASI_SYSROOT, or a usable clang/sysroot/linker/builtins pair,
the Go test names the missing requirement and skips. Once probes succeed,
compiler or runtime errors fail the test. Every Adamic fixture builds through
`adamic build --target wasm32-wasi`, never the temporary C request adapter.

The reference process extracts the driver's own crossing function from
`cmd/adamic/testdata/wasi/request-host.mjs`, retaining its ownership assertions
and replacing its source comparison with recording. It uses Node WASI only in
a separate process. The full original driver host also executes unchanged
against its stripped-source oracle. The Workers harness rejects any attempt to
load Node WASI, including through transitive imports.

Observed: 1,000 driver responses matched byte for byte. A handler returning
655,360 UTF-8 bytes grew memory and matched the same driver crossing. A counted
reactor without heap globals completed 100,000 calls in one instance with zero
live values, constant 393,216-byte memory after warmup, and 6,300,000 region
objects. Panic stderr matched, the Worker returned 500, and its next response
proved a fresh instance by reporting call number 1. The generated Worker was
also imported and exercised in Node using a loader for its Wasm binding. A
small C Wasm trap probe independently checked recreation after an engine trap.

Mutants are isolated copies of the production modules, automatically exercised
in every opted-in test run:

| Mutant | Observed failure |
| --- | --- |
| Swallow fd 2 | `panic stderr lost` |
| Read through memory captured before the handler | detached ArrayBuffer on the growing response |
| Skip adamic_release | `warmup retained counted values` |
| Keep the terminated instance | `terminated instance reused`, next body fresh:2 rather than fresh:1 |

Additional checks cover one-time module initialization, reentrant call refusal,
exact HttpRequest conversion, Response construction, split UTF-8 writes,
partial-line exit flushing, fdstat fields and named unsupported-import rejection.
No deployed Worker, workerd witness, benchmark, compiler change or Adamic
HttpRequest decoding glue is claimed by this unit.
