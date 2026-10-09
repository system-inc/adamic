# Adamic's WASI runtime

The runtime builds for `wasm32-wasi` (WASI Preview 1), using clang and a WASI
sysroot. Every runtime difference uses `ADAMIC_TARGET_WASI`, set with
`-DADAMIC_TARGET_WASI=1`. The compiler's target and driver are a separate unit;
this branch changes neither `native.go` nor the driver. Executable programs
still enter through `main`.

`wasm32-unknown-unknown` alone is insufficient: the runtime needs allocation,
libm, byte/string functions, output and process exit. WASI libc supplies those,
and translates the existing POSIX-shaped file operations to WASI imports.
A libc-free build would need replacements for those services, not just another
clang target flag.

## Portability inventory

| Area | WASI behavior |
| --- | --- |
| `adamic.c` signals | SIGPIPE handling and SIGTERM/SIGINT/SIGHUP flushing are compiled out. WASI has no process signals; there is no public runtime signal API to reach. Ordinary output, flush, explicit panics and exit 70 remain. |
| Atomics | The only runtime atomics here are `atomic_signal_fence` in the output buffer. They compile unchanged; these are compiler fences, not shared-memory synchronization. |
| Threads, pool, `_Thread_local`, `parallel.c` | None exists in the base checkout (`ef3d907`). No threading API was stubbed and no sequential `parallelMap` fallback was implemented. The compiler unit must emit a sequential loop, like the Node witness, when that operation arrives. |
| `stack.c` | No `getrlimit`. The constructor uses wasm-ld's `__stack_low` plus a 16 KiB panic margin. Generated `ADAMIC_CHECK_STACK` checks linear-memory frames. The tests reserve a 128 KiB stack and independently protect its bottom 32 bytes with a canary. |
| Engine call stack | Functions without linear-memory frames can exhaust the engine's separate call stack first. Two oracle fixtures demonstrate this limitation below. A balanced call-depth guard in compiler output, or a tested host strategy preserving buffered output and panic semantics, is still needed. |
| `input.c` | `open`, `read`, `write`, `close` and `fstat` go through WASI libc. Empty read/write paths are normalized to ENOENT; `fstat` identifies a directory before `fd_read`, whose EBADF otherwise differs from Node's EISDIR. UTF-8 decoding is unchanged. |
| `directory.c` | `opendir`, `readdir`, `lstat` and `stat` go through WASI libc. Directory listing and file status work inside preopened directories. No host filesystem is granted implicitly. |
| `sort.c` | Ordinary TimSort is unchanged. WASI SDK's `setjmp`/`longjmp` requires experimental wasm exception handling. That path is compiled out; a comparator that throws panics with `wasm32: throwing sort comparators are not supported`, exit 70. Ordinary emitted throw/catch cleanup paths and region cleanup still work. |
| `weak.c` | Pointer hashing widens a wasm32 pointer to `uint64_t` before shifting by 33. Weak lifetime, reuse and ownership behavior are unchanged. |
| Sanitizer-only paths | `heap.c` detects AddressSanitizer and includes its poisoning interface only in native sanitized builds. The normal WASI build does not enter these paths. Native ASan/UBSan flags are not supported by this WASI toolchain and are not used for wasm. Counted wasm builds supply the independent lifetime probes. |
| `tsgo.c` | The external native Go-checker bridge remains disabled unless `ADAMIC_TSGO` is explicitly selected. A native Go archive cannot be linked into this WASI module. |

All 48 runtime translation units compile with `-std=c11 -Wall -Wextra -Werror
-pedantic`, without the generated-C unused-variable/function exemptions. The
five changed runtime files produce byte-identical native objects to the base
with native clang `-O2`; the native package gate and full native oracle pass.

## Running the opt-in test

Run `bash cloud/setup.sh`, then source the environment file it prints. Install
[WASI SDK 27](https://github.com/WebAssembly/wasi-sdk/releases/tag/wasi-sdk-27)
or supply an equivalent clang, linker, compiler builtins and WASI sysroot.
Node 24 supplies the command runtime and the independent source oracle.

```sh
source /workspace/adamic-tools/env.sh
export PATH=/workspace/adamic-tools/wasi-sdk-27.0-x86_64-linux/bin:$PATH
export WASI_SYSROOT=/workspace/adamic-tools/wasi-sdk-27.0-x86_64-linux/share/wasi-sysroot
ADAMIC_TEST_WASI=1 ADAMIC_WASI_ARTIFACT=/tmp/adamic-request.wasm \
  go test ./internal/native -run '^TestWASI$' -count=1 -v -timeout 15m \
  > /tmp/wasm-test.log 2>&1
```

Paths are examples; use your installation. Without opt-in, Node 24 or a usable
WASI toolchain, the test skips with the missing requirement. Once the minimal
toolchain probe succeeds, runtime compilation failures are failures, never
skips. The test invokes `adamic c <fixture>` from the existing compiler, compiles
all runtime C itself, and links each fixture with:

```sh
clang --target=wasm32-wasi --sysroot="$WASI_SYSROOT" \
  -DADAMIC_TARGET_WASI=1 -std=c11 -Wall -Wextra -Werror -pedantic \
  -O2 -ffp-contract=off -fno-optimize-sibling-calls \
  -Wl,-z,stack-size=131072 -Wl,--export=__stack_low \
  -I runtime-directory main.c runtime-objects/*.o -lm -o main.wasm
node --disable-warning=ExperimentalWarning internal/native/wasm/run.mjs main.wasm
```

Generated C has the same unused-code warning exemptions as native builds.
`run.mjs` preopens only its working directory and checks the reserved-stack
canary after `main` returns or exits through WASI. Oracle command objects are
uncounted; the reactor's objects are rebuilt counted for lifetime measurements.
`ADAMIC_WASI_RUNTIME` selects a scratch runtime directory for isolated mutants.

The test lists 33 fixtures: 30 agree exactly in stdout, stderr and exit status.
They cover all ten 0.1 examples, strings, numbers, bitwise operations, functions,
regions and region exceptions, Weak and reuse, normalization, maps, JSON,
RegExp collections, Object keys, array flattening, file reads and a file/directory
round trip. `read_files.a` runs in its own fixture directory, including malformed
UTF-8, BOMs, NULs, missing files, a directory and an empty path.

The other three are explicitly classified and checked, not counted as equivalent:

| Fixture | Observation |
| --- | --- |
| `closures_throw.a` | Output agrees until the first throwing sort comparator; WASI then gives the explicit unsupported panic and exits 70. Node catches the comparator's error and continues. |
| `stack_forever.a` | Node prints `start`, then the Adamic stack-overflow panic and exit 70. WASI's engine raises RangeError, loses buffered stdout, and exits 1. |
| `stack_tail_call.a` | Node prints `0`, then the same panic and exit 70. WASI has the same engine-stack mismatch as above. |

The opt-in test passes only if the equivalent fixtures agree and these three
known limitations produce their specifically checked observations. Unknown
differences fail. `stack_overflow.a`, which allocates linear stack frames,
agrees with Node and preserves the canary. This is not full wasm oracle coverage.
Permission probes, paths outside preopens, symlink escape behavior, special files,
empty directory/status paths, very large allocations and every library fixture
are not covered here.

## Request ABI and host

The exported application entry is:

```c
adamic_string *adamic_request(const unsigned char *bytes, size_t length);
```

On wasm32 this is `(i32, i32) -> i32`: a borrowed pointer and byte length of UTF-8
request data, returning one owned Adamic string pointer. The function decodes
the request using `adamic_decode_utf8`, calls the handler, releases the decoded
request and returns the response. Invalid UTF-8 uses the existing WHATWG
replacement rules. The host copies the result before releasing it. Adamic
response bytes are WTF-8; the host's TextDecoder replaces lone surrogates.

The supporting exports are `memory`, `malloc`, `free`, `adamic_release`,
`adamic_response_bytes(const adamic_string *) -> const char *` and
`adamic_response_length(const adamic_string *) -> size_t`. All pointer/length
results are wasm32 i32 values. `adamic_live()` and `adamic_regions()` are counted
prototype probes, not required application ABI. Input buffers are host-owned;
responses must be released exactly once. Reacquire memory views after a wasm
call, since allocation can grow memory.

`internal/native/wasm/request.a` is the source witness. Its handler builds and
counts a 63-object tree inside a declaration's statement region and returns
`hello <request>: 63`. The tree dies when that statement ends; the response is
an ordinary owned heap string and survives the region. No current-region global
and no garbage collector is added.

`request-abi.c` is a temporary adapter around today's executable C output,
selecting the emitted handler by name. Only this test adapter renames its
included fixture `main` and links with `-mexec-model=reactor`; ordinary programs
keep `main`. The compiler unit can replace the adapter by emitting the export
and module initialization directly. Node calls `_initialize` exactly once before
serving requests. This fixture has no module initialization to replay beyond
runtime constructors; a general module will need its top-level initialization.

`request-host.mjs` implements a Worker-shaped async `fetch(Request)` returning
`Response`. The synchronous handler runs after the body is read, with no
filesystem preopens. The test verifies a fetch call, warms up with varied ASCII,
Unicode, 1 KiB and NUL-containing requests, then checks 100,000 responses. Between
every request it requires zero live counted values and no linear-memory growth;
it also requires exactly 6,300,000 objects ended through regions. Linear memory
is reusable capacity and cannot shrink; flat capacity alone would not prove
correct release.

The host also imports the very same `.a` source with Node's type stripping and
measures its plain TypeScript handler. Rates include response assertions; wasm
also includes UTF-8 copies, ownership calls and memory/lifetime probes. They are
synchronous handler rates, not HTTP throughput or a Cloudflare performance claim.
The observed numbers and mutants are recorded below.

## Compiler and Worker integration

The compiler target needs the WASI sysroot, `ADAMIC_TARGET_WASI=1`, wasm-compatible
archive tools if it builds an archive, libm, the existing C11 flags, and a linker
stack larger than the 16 KiB panic margin. Preserve sibling-call prevention and
floating-point contraction settings. Initialize constructors before invoking
exports. Keep the existing ownership and statement-region conventions; a response
or other escaping value must never point into an ended request region. Compile
`parallelMap` sequentially when it is introduced, and refuse or clearly panic
for unsupported sort exceptions. The separate engine-stack limit needs a guard
that also works when a function needs no linear-memory frame.

The stripped request artifact imports exactly `fd_close`, `fd_prestat_get`,
`fd_prestat_dir_name`, `fd_seek`, `fd_write` and `proc_exit` from
`wasi_snapshot_preview1`. Its other runtime work uses linear memory and wasm
instructions. A command or a handler using filesystem input will import more.

A real Worker would add an ES-module wasm binding/bundling configuration,
instantiation once per isolate, and a tested WASI Preview 1 import adapter for
allocation startup, output and `proc_exit`. Workers do not supply Node's
`node:wasi`; this host is a local stand-in, not a deployment. Give filesystem
imports explicit behavior with no preopens, route output to host logging, and
translate a panic into a request failure without trying to reuse an instance
whose WASI execution has terminated. Do not share an instance across simultaneous
wasm execution or native threads; runtime counters, caches and pending exceptions
are single-threaded. Fetch bodies may await before the synchronous wasm call.
Measure deployed module size, cold startup, isolate reuse and request throughput
on the actual Worker. No Cloudflare account or deployment was used here.

## Measurements and mutants

Observed on the shared cloud machine with Node 24.19.0 and WASI SDK 27. The
prototype is 272,389 bytes, or 48,940 bytes after `llvm-strip`. Stripping preserves
all exports and the full 100,000-request assertions. The counted prototype's
linear memory remains exactly 393,216 bytes (six 64 KiB pages), with zero live
counted values between requests and 6,300,000 objects ended through regions.

One complete run measured 202,734 wasm requests/s against 904,284 plain Node
requests/s, with 1.86 ms for read/compile and 2.03 ms for instantiate/initialize.
A later run while native tests were also running measured 143,098 against
771,085, with 7.45 and 5.19 ms startup stages. The stripped module measured
138,543 against 429,829 under concurrent load. These are observations on a
shared CPU, not a controlled performance comparison. The wasm boundary and
counted probes are slower for this handler; no speedup is claimed.

Every mutant below compiled and was caught by executing the named check:

| Mutant | Check that caught it |
| --- | --- |
| Console newline replaced by `!` | `01_hello.ts`: exact stdout comparison. |
| Panic prefix replaced | `panic.a`: exact stderr comparison. |
| Panic `_exit(70)` changed to `_exit(0)` | `panic.a`: exit status comparison. |
| Empty-path normalization disabled | `read_files.a`: empty-read message; `wasm/io.a`: empty-write message. |
| Directory-read detection disabled | `read_files.a`: EISDIR message becomes generic failure. |
| A live Weak target always reads absent | `weak_parent.a`: exact stdout comparison. |
| Throwing sort comparator continues sorting | `closures_throw.a`: required explicit unsupported panic is missing, and sort output changes. |
| Linear stack limit set to zero | `stack_overflow.a`: reserved-stack canary is overwritten. Initially stdout/stderr/exit alone missed this mutant; that observation led to the independent canary. |
| Region block `free` omitted | Request loop: linear memory grows after warmup, despite zero logical live counts. |
| `adamic_release` never lets go | Request loop: live allocation assertion fails on dynamically allocated request/response data. |
| Region live count decremented without recording region objects | Request loop: memory and live checks pass, but region count is 0 instead of 6,300,000. |

An initial sort mutant also changed the out-of-memory branch and was stopped by
clang; it was corrected to alter only the comparator's throw path, and only the
runtime failure is counted above. An initial subtest filter selected no fixtures;
those green runs were discarded and all listed witnesses rerun with full paths.

Native verification used the native clang PATH, separately from WASI SDK:

```sh
go test ./internal/native/... -count=1 -timeout 15m > /tmp/wasm-native-final.log 2>&1
go test ./internal/oracle -count=1 -timeout 30m > /tmp/wasm-oracle-full-final.log 2>&1
go vet ./... > /tmp/wasm-vet-all.log 2>&1
gofmt -l cmd internal > /tmp/wasm-format.log
```

The native package run passed in 87.711 s and the complete oracle package in
69.672 s; vet and format checks produced no output. Both new `.a` fixtures were also built with `--sanitize` and compared
to Node: exit 0, identical stdout, empty stderr. The complete repository test
gate was not run; native and oracle packages were the test scope.

Setup printed `go ready`, `clang ready`, `node ready` and `submodules ready` at
0 s, `build cache warm` at 106 s, and `done in 106s on 5 processors` (`nproc` = 5,
cgroup quota four CPUs, 17.6 GB reported memory). Go was 1.27.1 and native clang
20.1.8. The setup environment was `/workspace/adamic-tools/env.sh`.

## Threads

Plain `wasm32-wasi` uses unshared memory and `-mno-atomics`. W7's guarded pool
startup selects one executor regardless of `ADAMIC_THREADS`; no worker is started.
`parallelMap` uses the existing single-thread path, visiting ascending indices,
preserving result positions, and stopping at the first thrown exception. Input,
closure and result graph publication still follows the native ownership ABI.
The request handler remains synchronous within one module instance.

WASI SDK 27 clang lowers the runtime's `_Thread_local` storage and C atomics to
ordinary memory accesses without shared memory. A disassembled probe shows
`i32.load`, `i32.add` and `i32.store` for TLS increment and atomic fetch-add,
with no wasm atomic instruction. No replacement counter implementation is added.
WASI libc supplies single-thread mutex/once operations. Its `pthread_create`
stub returned 6 (ENXIO) and never called the worker in the executed probe; that
stub is a failure path, not a pool implementation.

The WASI panic path skips the native wait for a competing panic thread. Worker
stack initialization leaves the constructor's linear-stack limit intact.
Native preprocessing retains the existing pool, TLS and atomic implementation;
all 50 release runtime objects were compared byte for byte at `-O2`.

The first merged-tree checks exposed two additional integration blockers:
`-pthread` conflicts with `-mno-atomics` in the driver, and `share.c` hashes a
32-bit pointer with a shift by 33. The W7 evidence distinguishes these compile
failures from runtime agreement; a green WASI oracle is required before landing.
See `cloud/reports/wasm-threads/` for the hooks, commands, counts and limitations.

Actual wasm threads would require a wasi-threads-capable host, shared linear
memory, atomics-enabled compilation and linking, worker instantiation and TLS
initialization, per-thread stacks, and tested scheduling/ownership and shutdown.
The native pool cannot acquire those facilities from ordinary WASI Preview 1.
That future target needs its own Node comparisons and race/lifetime proofs.
W7 makes no speed measurement.
