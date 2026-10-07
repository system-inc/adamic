# Async feasibility prototype

Proposal only, task p286ycm. Kirk and Ahra have not approved these decisions. See [the design](../../docs/concurrency-async.md).

`prototype.c` is a hand-written three-await state machine with counted frames, Promises and dynamic text. It runs an eager prefix, then awaits a fulfilled Promise, a monotonic timer and a real file read performed on a helper pthread. Only the loop settles Promises or changes reference counts. Success, throw after the file await, and cooperative cancellation after the timer all drain queued reactions and release frames.

`oracle.a` is a Node-only reference program, not an Adamic fixture. `check.py` copies it verbatim to a scratch `.cjs` file. No compiler feature or production runtime is wired in. The repository's existing tsconfig does not include this prototype directory.

```sh
source /workspace/adamic-tools/env.sh
python3 prototypes/async/check.py > /tmp/p286ycm-check.log 2>&1
```

Requires Linux/POSIX, clang with ASan/UBSan/LeakSanitizer, Python 3 and Node 24. The harness writes every child stdout/stderr and build diagnostic directly to files in the scratch directory it prints. It compares native stdout and exit with Node, requires Node stderr empty, checks native stderr contains only exact counts, and runs six mutant cases. A timeout or compilation failure fails the harness; it never counts as a mutant kill. The controls use `ASAN_OPTIONS=detect_leaks=1:halt_on_error=1`.

The prototype's counts cover Object-header allocations only, excluding jobs and platform allocations; LeakSanitizer covers the rest. Each allocation starts with one count, so releases = allocations + retains, rather than releases = retains. This is separate from Adamic ADAMIC_COUNT.

This loop has one timer slot and one file request slot. The file buffer is 128 bytes and trims CR/LF for this ASCII fixture; it is not a UTF-8 file API. The Node timer starts the file read causally, avoiding a race between independent external completions. Cancellation is a source-visible flag checked after the timer, not OS request cancellation or an extension of JavaScript Promise semantics. General loops, finally, network I/O, Promise resolution assimilation and production scheduling remain in the design's landing plan.
