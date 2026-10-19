# The adamic command

What `cmd/adamic` accepts, returns and needs, read from the code on main (141fec77). Every claim
cites the line that makes it true. When the code and this page disagree, the code is right and this
page is a bug.

## Usage

The binary prints this, and only this, for every usage error. `TestDocumentedUsageIsTheCommands`
(`cmd/adamic/cli_doc_test.go`) fails when this block and `usage` (`cmd/adamic/main.go:28-32`) differ,
and when `build` parses a flag the block doesn't name.

```text
usage:
  adamic types <file.a|file.ts>...
  adamic c <file.a|file.ts> [--explain-checks]
  adamic js <file.a|file.ts> [--explain-checks]
  adamic build [--target wasm32-wasi] <file.a|file.ts> -o <out> [--count] [--sanitize] [--tsgo <archive>] [--explain-checks]
```

There is no `--help`, `--version`, `run`, `test` or `fix`. Anything the switch doesn't match,
`--help` included, prints the usage and exits 2 (`cmd/adamic/main.go:102-103`).

## Commands

Every command first loads the program through typescript-go in process. A checker error prints each
diagnostic to stderr and exits 1; any other load error prints `adamic: <error>` and exits 1
(`cmd/adamic/main.go:116-130`). Paths in diagnostics are shortened to be relative to the working
directory (`cmd/adamic/main.go:153-165`). `.a` and `.ts` are both accepted (`docs/0.1.md:766`).

| Command | Does | Writes | Code |
| --- | --- | --- | --- |
| `types <file>...` | checks one or more files and prints every declaration's proven type as `file:line:column: kind name: type` | stdout | `cmd/adamic/main.go:66-74` |
| `c <file>` | checks, lowers and prints the program as C11 | stdout | `cmd/adamic/main.go:75-85` |
| `js <file>` | checks, lowers and prints the program as JavaScript | stdout | `cmd/adamic/main.go:86-96` |
| `build <file> -o <out>` | checks, lowers, emits C and compiles it with the runtime to a binary at `<out>` | `<out>` | `cmd/adamic/main.go:97-100`, `cmd/adamic/tsgo.go:15-97` |

`c` and `js` take exactly one file (`cmd/adamic/main.go:75,86`); `types` takes one or more
(`cmd/adamic/main.go:66`). A program that uses the tsgo checker library can't be printed:
`c` says `adamic: tsgo requires a native build with --tsgo <archive>` and `js` says
`adamic: tsgo is an external native checker library; JavaScript is not supported`, both exit 1
(`cmd/adamic/main.go:80-83,91-94`).

### Where `build`'s arguments go

The switch is positional (`cmd/adamic/main.go:97-100`): the source is the argument right after
`build`, or right after `build --target <t>`, and `-o <out>` comes immediately after the source.
The flags after `-o <out>` may come in any order (`cmd/adamic/tsgo.go:19-45`). So
`build --target wasm32-wasi a.a -o a.wasm` and `build a.a -o a.wasm --target wasm32-wasi` both work
(`cmd/adamic/command_guards_test.go:21-31`), and `build -o out a.a` or `build a.a --count -o out` is a
usage error.

`--explain-checks` is lifted out of the arguments wherever it appears, before the switch
(`cmd/adamic/main.go:43-64`). With `c` or `js` it goes to `renderWithChecks`
(`cmd/adamic/checks.go:14-34`); with `build` it is moved to the end; with `types` or nothing it is a
usage error.

### Flags

| Flag | Applies to | Effect | Code |
| --- | --- | --- | --- |
| `-o <out>` | `build`, required | the binary's path | `cmd/adamic/main.go:97-100` |
| `--target wasm32-wasi` | `build` | a WebAssembly module for WASI instead of a host binary. The only value accepted; any other, a second `--target`, or a missing value is a usage error | `cmd/adamic/tsgo.go:21-27` |
| `--count` | `build` | the binary counts allocations, frees, retains, releases, peak and regions and writes `adamic: counts: ...` to stderr as it exits (`-DADAMIC_COUNT`) | `cmd/adamic/tsgo.go:30-31`, `internal/native/native.go:124-125`, `internal/native/runtime/count.c:18-34` |
| `--sanitize` | `build`, native only | `-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all` in place of the release flags | `cmd/adamic/tsgo.go:32-33`, `internal/native/native.go:139-140` |
| `--tsgo <archive>` | `build`, native only | links a prebuilt typescript-go checker archive for programs that use it; a second `--tsgo` or a missing value is a usage error | `cmd/adamic/tsgo.go:34-40` |
| `--explain-checks` | `c`, `js`, `build` | writes every predicate-overload and non-null check site, proven or checked, and their totals to stderr; stdout is unchanged | `cmd/adamic/checks.go:36-70` |

`--count` and `--sanitize` may be combined (`internal/native/target.go:10-34` refuses neither).
Repeating `--count`, `--sanitize` or `--explain-checks` is harmless.

### What wasm32-wasi changes

- A module whose entry file exports `handleRequest(request: string): string`, as a named function
  declaration or an alias of one, is built as a reactor exporting the request ABI; without one it is
  a command (`cmd/adamic/request.go:52-92`, `internal/native/target.go:55-68`). A `handleRequest` of
  any other shape, or an ambiguous one, exits 1 (`cmd/adamic/request.go:24-25,38-46,68,72`).
- `--sanitize` exits 1 with `adamic: native: sanitizers are not supported for wasm32-wasi`
  (`internal/native/target.go:20-21`), and `--tsgo` exits 1 with
  `adamic: --tsgo is not supported for wasm32-wasi` (`cmd/adamic/tsgo.go:50-53`).

## Exit codes

### The compiler

| Code | Means | Returned at |
| --- | --- | --- |
| 0 | the command did what it says | `cmd/adamic/main.go:74,85,96`, `cmd/adamic/tsgo.go:96` |
| 1 | the program doesn't check, is refused (`Refused`) or not yet compiled (`NotYet`), a target refuses an option, a tool is missing, or clang failed | `cmd/adamic/main.go:124,127,148`, `cmd/adamic/tsgo.go:48,52,73,94`, `cmd/adamic/request.go:25,30,40,46` |
| 2 | usage error | `cmd/adamic/main.go:59-60,102-103`, `cmd/adamic/tsgo.go:24-25,37-38,42-43` |
| 4 | the compiler itself panicked: stderr gets `adamic: compiler bug: <value>` and the stack | `cmd/adamic/main.go:41,106-114` |

Exit 1 covers both "your program" and "your machine": a script can't tell a refusal from a missing
clang without reading stderr. Compiler has ruled the split (#z1ba56x): 1 the program, 2 usage
including an option the target refuses, 3 the machine, 4 a compiler bug; only 4 is in the code yet.

`TestUsageExitStatus` pins 2 and the exact usage text (`cmd/adamic/command_guards_test.go:13-19`);
`TestBuildWASISanitizeTarget{Before,After}Source` pin a target refusal at 1 with its text, on a file
that loads (`cmd/adamic/command_guards_test.go:21-44`); `TestBuildTargetParsing` pins three malformed
`--target`s at 2 (`cmd/adamic/target_test.go:10-21`); `TestCompilerBugExitsFour` and
`TestRunRecoversCompilerBugs` pin 4, its text, and that `run` defers the recovery
(`cmd/adamic/compiler_bug_test.go`).

### A program it built

| Code | Means | Where |
| --- | --- | --- |
| 0, or the program's own exit code | the program finished; `main` returns the stored exit code, 0 when none was set | `internal/native/emit.go:114`, `internal/native/runtime/host_runtime.c:96-98` |
| 70 | a panic: an uncaught throw, a failed inserted check (non-null assertion, checked cast), stack exhaustion, out of memory, an index written out of range, and the runtime's other checked failures. stderr gets `adamic: panic: <message>` after everything the program printed | `internal/native/runtime/adamic.c:311-330`, `internal/native/runtime/exceptions.c:73-90` |
| 70, silently | stdout or stderr couldn't be written (a closed pipe), or a closed standard descriptor couldn't be reopened on `/dev/null` at startup | `internal/native/runtime/adamic.c:64-66,93-103,248-256` |
| a signal | a catchable stop signal flushes output and re-raises; SIGPIPE, SIGXFSZ and SIGUSR1 are ignored | `internal/native/runtime/adamic.c:137-173,263-269` |
| SIGABRT | an internal async-loop failure | `internal/native/runtime/async.c:171`, `internal/native/runtime/async_host_impl.h` |
| the sanitizers' own | under `--sanitize`, ASan and UBSan exit with their defaults; nothing sets `ASAN_OPTIONS` for a user build | `internal/native/native.go:139-140` |

70 is EX_SOFTWARE (`internal/native/runtime/adamic.h:1174`). `process.exit` isn't lowered: any Node
member not registered is `NotYet` (`internal/lower/library_node.go:11-13,62-70`), so a program
calling it doesn't build. A `--count` imbalance never changes the exit code; the counts line is for
the test harness to judge (`internal/leakcheck/leakcheck.go:136-160`).

`adamic js` output panics the same way, `adamic: panic: <message>` and 70, but only where its
`'adamic'` import resolves to `oracle/adamic.mjs` (`internal/javascript/javascript.go:50`,
`oracle/adamic.mjs:10-31`), which only the oracle's loader arranges (`oracle/node.mjs`). Under plain
`node` the import doesn't resolve.

A wasm32-wasi panic leaves through the same `_exit(70)` (`internal/native/runtime/adamic.c:318-329`);
`TestWASIRequestThrows` sees 70 and `adamic: panic: Error: boundary` from the request host
(`cmd/adamic/request_test.go:152-169`).

## Environment

### Read by `adamic build`

| Variable | Effect | Default | Code |
| --- | --- | --- | --- |
| `PATH` | where `clang`, `llvm-ar` and `ar` are found | | `internal/toolchain/toolchain.go:156-174` |
| `ADAMIC_TOOLS` | `<ADAMIC_TOOLS>/bin` then `<ADAMIC_TOOLS>/go/bin` are searched before `PATH`, and put first on the tools' `PATH` | unset | `internal/toolchain/toolchain.go:138-174` |
| `ADAMIC_SANDBOX=1` | no fallback to `PATH`: a tool outside `ADAMIC_TOOLS` is an error naming where it looked, and the cache must be named by `ADAMIC_BUILD_CACHE_DIR` | off | `internal/toolchain/toolchain.go:134-136,165-172`, `internal/buildcache/location/location.go:26-32` |
| `WASI_SYSROOT` | required for `--target wasm32-wasi`, and must be a directory; the SDK's clang is taken from beside it | none | `internal/native/target.go:26-32,37-45` |
| `ADAMIC_BUILD_CACHE_DIR` | where the compiled runtime archive is cached | `os.UserCacheDir()/adamic-build` | `internal/buildcache/location/location.go:24-38` |
| `ADAMIC_BUILD_OUTPUT_DIR` | a separate directory to write cache entries to | unset | `internal/buildcache/location/location.go:20-22` |
| `ADAMIC_CC_JOBSERVER` | a FIFO; every clang call takes a token from it first | unset | `internal/native/jobserver.go:18` |
| `ADAMIC_NATIVE_SPLIT=1` | builds the program as split C units instead of one file | off | `internal/native/native.go:170` |
| `ADAMIC_NATIVE_JOBS` | split builds only: parallel clang jobs, an integer | 1 | `internal/native/units.go:447-451` |
| `ADAMIC_GATE_UNCACHED=1` | split builds only: bypasses the unit cache | off | `internal/native/units.go:477` |
| `TMPDIR` | where build directories are made | system default | `internal/native/native.go:194` |

clang and ar run under a fixed environment, not the caller's: only the variables in
`internal/toolchain/toolchain.go:105-125` pass through, `GOENV=off ZERO_AR_DATE=1` is always set, and
`CC=clang CXX=clang++` are set when unset. So `CPATH`, `SDKROOT` and `DYLD_*` never reach clang, and
setting `CC` doesn't change which compiler a native build runs: that is always `clang` resolved as
above (`internal/native/target.go:44`).

`ADAMIC_BUILD_LOG` and the other `ADAMIC_BUILD_*` store settings belong to `internal/buildcache`,
which `cmd/adamic` doesn't link (`internal/buildcache/location/location.go:1-3`); they matter only to
the test harness.

### Read by a program it built

| Variable | Effect | Code |
| --- | --- | --- |
| `ADAMIC_THREADS` | worker threads, an integer from 1 to 4096; anything else panics (70) | `internal/native/runtime/parallel.c:126-132` |
| `ADAMIC_LEAK_CHECK` | an intentional exit writes `adamic: intentional exit: status N` for the leak harness | `internal/native/runtime/adamic.c:345` |
| `ADAMIC_TSGO_PROFILE`, `ADAMIC_TSGO_TIMING` | tsgo builds only: checker profiling and timing | `internal/native/runtime/tsgo.c:16-17` |
| `TMPDIR`, `TMP`, `TEMP` | `os.tmpdir()` | `internal/native/runtime/node_host.c:211-213` |

## Toolchain

| Tool | Needed for | Found by | Version | Missing |
| --- | --- | --- | --- | --- |
| clang | `build`, native | `toolchain.Path("clang")` (`internal/native/library.go:222`) | not checked: its `--version` report only keys the runtime cache (`internal/native/library.go:226,300-303`). `cloud/setup.sh:94` installs 20.1.8 (`ADAMIC_LLVM_VERSION`) when no clang passes its sanitizer probe | `adamic: native: exec: "clang": executable file not found in $PATH`, exit 1 |
| lld | `build`, shipped native release on Linux | clang's `-fuse-ld=lld` (`internal/native/native.go:159-160`) | | clang's link error, exit 1 |
| `llvm-ar`, else `ar` | the runtime archive | beside the compiler, else `toolchain.Path("ar")` (`internal/native/target.go:47-53`) | | exit 1 |
| wasi-sdk | `--target wasm32-wasi` | `WASI_SYSROOT` (`internal/native/target.go:26-32`); clang from `<sysroot>/../../bin/clang` (`internal/native/target.go:37-45`) | not checked. `cloud/setup.sh --wasi-sdk` installs 27 (`cloud/setup.sh:158-168`) | no `WASI_SYSROOT`: `adamic: native: wasm32-wasi requires WASI_SYSROOT`, exit 1. No SDK clang: falls back to `clang` on `PATH` silently |
| a tsgo checker archive | `--tsgo` | the path given (`internal/native/tsgo.go:145-152`), built with `go build -buildmode=c-archive ./bridge/tsgo/archive` (`bridge/tsgo/README.md`) | | `adamic: native: checker archive: <stat error>`, exit 1 |
| Node | running `adamic js` output, and the oracle | not used by `adamic` itself | `cloud/setup.sh` installs 24; only `internal/native/wasm_test.go:91-93` checks a minimum (24) | |
| Go | building `adamic` | | `go 1.27`, `toolchain go1.27.1` (`go.mod:3,5`) | |

The binary ships no toolchain: it runs whatever clang it finds. The runtime's C files are embedded
and compiled on first use into `<cache>/native-runtime/<sha256>/runtime.a`, keyed by os, arch, the
compiler's path and version report, the archiver, the headers read and the flags
(`internal/native/library.go:169,296-316,411-443`).

### Flags clang gets

Every build: `-std=c11 -Wall -Wextra -Werror -Wcast-function-type-strict -pedantic` and the unused
warnings off (`internal/native/native.go:109`), `-ffp-contract=off` so arithmetic rounds as Node does
(`internal/native/native.go:110-113`), `-fno-optimize-sibling-calls` (`internal/native/native.go:117`).
Native adds `-pthread` (`internal/native/native.go:122`); wasm32-wasi adds
`--target=wasm32-wasi --sysroot=$WASI_SYSROOT -DADAMIC_TARGET_WASI=1 -mno-atomics`
(`internal/native/native.go:119`). Then `--sanitize` gets `-O1 -g` and the sanitizers, and everything
else `-O2` (`internal/native/native.go:139-142`). `adamic build` always asks for a release
(`cmd/adamic/tsgo.go:16`), and a native build with neither `--count` nor `--sanitize` adds
`-flto=thin` (`internal/native/native.go:143-151`). `cmd/adamic/release_test.go:11-56` holds that
policy, and that the two semantic flags never drop.

## Targets

| Target | Selected by | Tested on main by |
| --- | --- | --- |
| native, Linux x86_64 and arm64 | the host; no flag | `internal/oracle` `TestNativeAgreesWithNode_*` (`internal/oracle/families_generated_test.go:8`, body at `internal/oracle/oracle_test.go:696`), Linux-only leak checks (`internal/oracle/oracle_test.go:777`), `internal/native` `TestReleaseFlagsReachLink` (`internal/native/release_build_test.go:48`) |
| native, macOS | the host; no flag | darwin branches in code (`internal/native/native.go:219-231`, `internal/native/library.go:623-627`) and darwin-only tests (`internal/oracle/moves_darwin_test.go`, `internal/native/whole_build_paths_test.go:30`, `internal/native/machine_paths_test.go:135,172`). Loom's Go sources never name darwin, so as far as its code shows these run only where someone runs the suite on a Mac |
| wasm32-wasi | `--target wasm32-wasi` | parsing and refusals ungated (`cmd/adamic/target_test.go`, `cmd/adamic/command_guards_test.go:21-31`, `internal/native/target_test.go:11`); compiled programs only with `ADAMIC_ORACLE_WASI=1` (`internal/oracle/wasi_test.go:45`, `cmd/adamic/request_test.go:76-169`) or `ADAMIC_TEST_WASI=1` (`internal/native/wasm_test.go:64`). Loom's gate sets both (`GateEnvironment` in loom's `planner/plan.go:20`); `cloud/setup.sh` and `cloud/write-env.sh` set neither, so a worker's own `go test` skips them |
| JavaScript | `adamic js` | the oracle's third run of every fixture (`internal/oracle/oracle_test.go:572,727`), the fuzzer (`internal/fuzz/run.go:202`) |
| tsgo-linked native | `--tsgo <archive>` | `bridge/tsgo/products_test.go:352` and the `TestBridge*` tests over it |

There is no Windows, iOS, Android or embedded target: `--target` accepts only wasm32-wasi
(`cmd/adamic/tsgo.go:23`, `internal/native/target.go:17-18`), and nothing in the runtime or the
native Go branches on architecture. `cloud/setup.sh` provisions Linux x86_64 and aarch64 only
(`cloud/setup.sh:47-51`).

## Where other docs disagree

- `docs/wasm.md:5-6` says the branch changes neither `native.go` nor the driver; the driver has
  taken `--target wasm32-wasi` since (`cmd/adamic/main.go:97`).
- `docs/wasm.md:47-48` runs `go test ./internal/native -run '^TestWASI$'`, which matches no test and
  passes having run nothing; the tests are `TestWASIUnit00` to `35` (`internal/native/wasm_test.go:362-537`).
- `docs/wasm.md:38-46` installs WASI SDK 27 by hand; `cloud/setup.sh --wasi-sdk` does it and
  `cloud/write-env.sh:36-37` sets `WASI_SYSROOT`.
- `docs/lint-registration.md:89` shows `go run ./cmd/adamic build stage1/cohere/lint/main.ts` without
  `-o`, which exits 2.
- `cloud/setup.sh:7`'s header names only `--warm-tests`; its usage line names `--wasi-sdk` too
  (`cloud/setup.sh:23`).
- `CLAUDE.md` lists `adamic types | c | build | js <file>`, without `-o` or the flags; this page is
  the full form.
- `README.md` says macOS and Linux today and JavaScript for the web, which matches; the JavaScript
  output still needs the oracle's `'adamic'` module to run.

## Open gaps

- Exit 1 is both a refused program and a missing or failing toolchain (#z1ba56x).
- No `--help` that exits 0 and no `--version` (#a733mfq).
- Nothing checks the host clang's version, and wasm32 falls back to `PATH` clang silently (#h2vm51c).
- `build`'s source and `-o` are positional (#21m2dv3).
- `adamic js` output runs only under the oracle's loader (#qq5dkee).
- Toolchain downloads carry no checksum, and there is no release of `adamic` (#xx66xas).
- The other docs listed above (#gyg0wrm).
