# What WebAssembly can't do, and what Adamic does instead

Adamic compiles to `wasm32-wasi` (WASI Preview 1) with the same C and the same runtime as native, and every
module answers to the same oracle: Node running the source, stdout, stderr and exit code byte for byte.
WebAssembly is a smaller machine than a native process. This page lists each place it can't do what native
does, what Adamic does there, and which check holds that answer. The rule underneath every row: when Wasm
can't do something, Adamic either does the same thing another way that the oracle still accepts, or stops
loudly. It never does something different quietly.

How the target is built and tested is in [wasm.md](wasm.md); the boundary contract is in
[wasm-abi.md](wasm-abi.md).

## Threads

Plain `wasm32-wasi` has no shared memory and no way to start a thread. Modules are built with `-mno-atomics`,
so C11 atomics lower to ordinary loads and stores, and `_Thread_local` becomes an ordinary global.

- `parallelMap` runs its sequential path: the pool never starts, and items are visited in order, which is
  the order Node's sequential witness prints. The parallel runtime's thread-only code sits behind
  `#ifdef ADAMIC_TARGET_WASI` hooks, with every native object byte-identical before and after.
- If the pool were started anyway, wasi-libc's `pthread_create` stub returns `ENXIO` and the program panics
  with exit 70. A mutant that forces this is caught, as is one that visits items in reverse.
- One instance never runs two calls at once. Hosts must not call into an instance concurrently or reenter it.
- What real Wasm threads would need later: the `wasi-threads` proposal, shared linear memory, a module built
  with atomics and bulk memory, and a host that supports both. Cloudflare Workers doesn't, today.

Held by: the `WASI` oracle over the concurrency fixtures, on `codex/wasm-threads` (in area/runtime with the
concurrency stack).

## Files

WASI has no ambient file system. A module sees only the directories its host preopens, and nothing is
preopened implicitly.

- The test runner (`oracle/wasi.mjs`) preopens `/` so file fixtures behave as on Node. A deployed module
  should be given the narrowest directory it needs, or none.
- Paths behave as Node's do where WASI libc would differ: an empty path is an error, never the current
  directory, for reads and writes here and for `readDirectory` and `fileStatus` with `codex/wasi-empty-path`
  (merging through area/runtime). Reading a directory reports `EISDIR` rather than WASI's `EBADF`.
- Writing to `/dev/stdout` or `/dev/stderr` uses the output descriptor the host supplied, since Preview 1
  has no `/dev` namespace to reopen. Ordinary files still go through `open`.
- Hosts without a file system (Workers, browsers) use the shim in `internal/native/wasm/shim.mjs`: every
  `fd_` and `path_` call other than stdout and stderr returns `EBADF`, and `fd_prestat_get` tells libc there
  are no preopens. A program that reads files fails there with Adamic's ordinary error values, never with
  data that isn't there.
- wasmtime, the second engine, refuses things Node's host allows: opening `/dev/stdout`, `/dev/stderr` or
  `/dev/stdin` by path, malformed `UTF-8` arguments, and non-`UTF-8` file names. Each is asserted by name to
  its exact behavior, so a new difference still fails.

Held by: the `WASI` oracle including every input fixture, `empty-path.a`, and the wasmtime witness.

## The engine's own stack

A Wasm engine keeps a call stack the program can't see, separate from the linear-memory stack Adamic's
overflow check measures. A function that keeps everything in Wasm locals never moves the linear stack, so
deep recursion would hit the engine's limit first. The engine traps, buffered output is lost, and the exit
code is 1 instead of Adamic's panic.

- Under `WASI`, `ADAMIC_CHECK_STACK` keeps a small volatile frame in every emitted function, so the linear
  stack always moves and Adamic's own check fires first: buffered output is flushed, the panic message
  matches Node's `RangeError`, and the exit code is 70.
- The linear stack is 128 KiB with a 16 KiB panic margin. Recursion depth differs from Node; where it ends,
  and how, does not.
- Not covered: recursion entirely inside runtime C, and hosts configured with a much smaller engine stack.

Held by: `stack_forever.a`, `stack_tail_call.a` and `stack_overflow.a`, and a mutant that swaps in the
native check.

## Exceptions

WASI SDK's `setjmp` and `longjmp` need Wasm exception handling, which isn't enabled.

- Ordinary throw and catch need neither: emitted code propagates a pending error explicitly.
- A comparator that throws inside `sort` propagates the error out through each sort frame before the next
  comparison, so it is caught exactly where Node catches it.
- An uncaught throw or panic ends the instance through `proc_exit(70)`. A host must discard that instance
  and start a new one; it can't be reused. Under the shim this arrives as a thrown `AdamicExit`.

Held by: `closures_throw.a` and the shim's exit mutant.

## Signals and processes

There are no signals in WASI. The native runtime's `SIGPIPE` handling and its flush on `SIGTERM`, `SIGINT`
and `SIGHUP` are compiled out. Output is still flushed on normal exit, on `proc_exit` and before a panic.
There is no subprocess, no environment unless the host passes one, and no clock or randomness that Adamic's
runtime depends on; the shim traps those calls by name if anything reaches them.

## Memory

Linear memory grows in 64 KiB pages and never shrinks. Flat memory therefore means capacity stops growing
after warmup, not that it returns. Over 100,000 requests in one instance, capacity holds steady and the live
value count returns to the module's baseline after every request.

Held by: the request witness (`request.a`), the JSON service (`internal/native/wasm/service/`) and the
exports test, each with mutants that retain a value or skip a release.

## Checks that can't run inside Wasm

AddressSanitizer, UndefinedBehaviorSanitizer and LeakSanitizer don't exist for this target. The same C runs
under all three natively in every oracle run, and Wasm builds use counted builds (`--count`) to check
lifetimes independently. The in-process TypeScript checker bridge (`tsgo.c`) can't link into a Wasm module
and stays native-only.

## Numbers that differ by host

The Wasm module computes Math with the same C on every host, so it gives the same bits everywhere. Node on
macOS arm64 fuses multiply-adds in some Math functions and differs from Node on Linux in the last bit
(#myatdyv). On a Mac, `navigation.a` disagrees for that reason alone. Linux is the gate of record.

## Speed

On the 100,000-request JSON service, Wasm ran about 30% slower than the same program compiled natively.
Most of the gap to Node is shared with native, in string slicing, allocation and destruction, which the
runtime is fixing for both targets. The Wasm-only part is the boundary: copying and decoding each request.
Measured, not claimed: see `cloud/reports/wasm-requests-profile/report.md`.
