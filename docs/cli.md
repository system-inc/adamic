# The adamic command

What `cmd/adamic` accepts, returns and needs, read from the code. Every claim cites what makes it
true: Go by the declaration (`path.go:Name`, a method as `Type.Method`), anything else by line.
`TestDocumentedCitationsResolve` (`cmd/adamic/cli_doc_test.go`) fails when a cited declaration, line or
path is gone, so a change that moves one fixes the page in the same change. When the code and this
page disagree, the code is right and this page is a bug.

## Usage

The binary prints this, and only this, for every usage error. `TestDocumentedUsageIsTheCommands`
(`cmd/adamic/cli_doc_test.go`) fails when this block and `usage` (`cmd/adamic/main.go:usage`) differ,
and when `build` parses a flag the block doesn't name.

```text
usage:
  adamic types <file.a|file.ts>...
  adamic c <file.a|file.ts> [--explain-checks]
  adamic worker <entry.a> --out <directory> [--wasm <name>[,<name>...]]
  adamic js <file.a|file.ts> [--explain-checks]
  adamic build [--target wasm32-wasi] [--reactor] <file.a|file.ts> -o <out> [--count] [--sanitize] [--tsgo <archive>] [--export <name>]... [--abi-json <path>] [--explain-checks]
```

There is no `--help`, `--version`, `run`, `test` or `fix`. Anything the switch doesn't match,
`--help` included, prints the usage and exits 2 (`cmd/adamic/main.go:run`).

## Commands

Every command first loads the program through typescript-go in process. A checker error prints each
diagnostic to stderr and exits 1; any other load error prints `adamic: <error>` and exits 1
(`cmd/adamic/main.go:check`). Paths in diagnostics are shortened to be relative to the working
directory (`cmd/adamic/main.go:relative`). `.a` and `.ts` are both accepted (`docs/0.1.md:776`).

| Command | Does | Writes | Code |
| --- | --- | --- | --- |
| `types <file>...` | checks one or more files and prints every declaration's proven type as `file:line:column: kind name: type` | stdout | `cmd/adamic/main.go:run` |
| `c <file>` | checks, lowers and prints the program as C11 | stdout | `cmd/adamic/main.go:run` |
| `js <file>` | checks, lowers and prints the program as JavaScript | stdout | `cmd/adamic/main.go:run` |
| `build <file> -o <out>` | checks, lowers, emits C and compiles it with the runtime to a binary at `<out>` | `<out>` | `cmd/adamic/main.go:run`, `cmd/adamic/tsgo.go:buildWithNative` |

`c` and `js` take exactly one file (`cmd/adamic/main.go:run`); `types` takes one or more
(`cmd/adamic/main.go:run`). A program that uses the tsgo checker library can't be printed:
`c` says `adamic: tsgo requires a native build with --tsgo <archive>` and `js` says
`adamic: tsgo is an external native checker library; JavaScript is not supported`, both exit 1
(`cmd/adamic/main.go:run`).

### Where `build`'s arguments go

The switch is positional (`cmd/adamic/main.go:run`): the source is the argument right after
`build`, right after `build --target <t>`, or right after `build --target <t> --reactor`, and `-o <out>`
comes immediately after the source. The flags after `-o <out>` may come in any order
(`cmd/adamic/tsgo.go:buildWithNative`), `--reactor` among them. So
`build --target wasm32-wasi a.a -o a.wasm` and `build a.a -o a.wasm --target wasm32-wasi` both work
(`cmd/adamic/command_guards_test.go:TestBuildWASISanitizeTargetBeforeSource,TestBuildWASISanitizeTargetAfterSource`), and `build -o out a.a` or `build a.a --count -o out` is a
usage error.

`--explain-checks` is lifted out of the arguments wherever it appears, before the switch
(`cmd/adamic/main.go:run`). With `c` or `js` it goes to `renderWithChecks`
(`cmd/adamic/checks.go:renderWithChecks`); with `build` it is moved to the end; with `types` or nothing it is a
usage error.

### Flags

| Flag | Applies to | Effect | Code |
| --- | --- | --- | --- |
| `-o <out>` | `build`, required | the binary's path; a relative one is under the working directory | `cmd/adamic/main.go:run`, `cmd/adamic/tsgo.go:buildWithNativeExports` |
| `--target wasm32-wasi` | `build` | a WebAssembly module for WASI instead of a host binary. The only value accepted; any other, a second `--target`, or a missing value is a usage error | `cmd/adamic/tsgo.go:buildWithNative` |
| `--count` | `build` | the binary counts allocations, frees, retains, releases, peak and regions and writes `adamic: counts: ...` to stderr as it exits (`-DADAMIC_COUNT`) | `cmd/adamic/tsgo.go:buildWithNative`, `internal/native/native.go:Flags`, `internal/native/runtime/count.c:18-34` |
| `--sanitize` | `build`, native only | `-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all` in place of the release flags | `cmd/adamic/tsgo.go:buildWithNative`, `internal/native/native.go:Flags` |
| `--tsgo <archive>` | `build`, native only | links a prebuilt typescript-go checker archive for programs that use it; a second `--tsgo` or a missing value is a usage error | `cmd/adamic/tsgo.go:buildWithNative` |
| `--reactor` | `build`, wasm32-wasi only | a reactor module: `_initialize` runs once and the entry module's crossing `export function`s are exported through ABI v1 (`docs/wasm-abi.md`), instead of a command entered at `main`. Without it a module with crossing exports is a reactor anyway | `cmd/adamic/tsgo.go:buildWithNative`, `docs/wasm-abi.md` |
| `--export <name>` | `build`, wasm32-wasi only | exports only the named functions, and implies `--reactor`; repeats, one name each. A name that isn't a crossing export exits 1 | `cmd/adamic/tsgo.go:buildWithNative`, `cmd/adamic/exports.go:exportSignatures,compileExports` |
| `--abi-json <path>` | `build`, wasm32-wasi only | after a successful build, writes the export table (`{"version": 1, "exports": [...]}`) as JSON at `<path>`, which Workers generates its bindings from | `cmd/adamic/tsgo.go:buildWithNative`, `docs/wasm-abi.md` |
| `--explain-checks` | `c`, `js`, `build` | writes every predicate-overload and non-null check site, proven or checked, and their totals to stderr; stdout is unchanged | `cmd/adamic/checks.go:explainPredicateChecks,explainNonNullChecks` |

`--count` and `--sanitize` may be combined (`internal/native/target.go:ValidateOptions` refuses neither).
Repeating `--count`, `--sanitize`, `--reactor` or `--explain-checks` is harmless. `--reactor`, `--export` or
`--abi-json` without `--target wasm32-wasi` exits 1 with `adamic: export ABI requires wasm32-wasi`
(`cmd/adamic/tsgo.go:buildWithNative`).

### What wasm32-wasi changes

- A module whose entry file exports `handleRequest(request: string): string`, as a named function
  declaration or an alias of one, is built as a reactor exporting the request ABI. A plain declaration is
  a crossing export, so ABI v1 carries it (`internal/native/wasm_exports.go:WASIExports`); an alias goes
  through the request path (`cmd/adamic/request.go:requestFunction`). With no such export and no crossing
  exports, the module is a command (`cmd/adamic/tsgo.go:buildWithNative`,
  `internal/native/target.go:WASILinkFlags`). An ambiguous function name exits 1
  (`cmd/adamic/exports.go:compileExports`, `cmd/adamic/request.go:compileWASI`), and so does an alias of the
  wrong shape (`cmd/adamic/request.go:requestFunction`). A plain `handleRequest` of another shape inside the
  crossing set builds as an ABI v1 export, not a refusal.
- `--sanitize` exits 1 with `adamic: native: sanitizers are not supported for wasm32-wasi`
  (`internal/native/target.go:ValidateOptions`), and `--tsgo` exits 1 with
  `adamic: --tsgo is not supported for wasm32-wasi` (`cmd/adamic/tsgo.go:buildWithNative`).
- `--reactor` and `--export` build a reactor exporting the entry module's functions through ABI v1,
  and `--abi-json` writes its table (`docs/wasm-abi.md`, `cmd/adamic/tsgo.go:buildWithNative`).
- The release is `-Oz` rather than `-O2`, with no ThinLTO (`internal/native/native.go:Flags`), and
  links with `-Wl,--strip-debug`, which keeps the name section (`internal/native/target.go:WASILinkFlags`).

## Exit codes

### The compiler

| Code | Means | Returned at |
| --- | --- | --- |
| 0 | the command did what it says | `cmd/adamic/main.go:run`, `cmd/adamic/tsgo.go:buildWithNative` |
| 1 | the program doesn't check, is refused (`Refused`) or not yet compiled (`NotYet`), a target refuses an option, a tool is missing, or clang failed | `cmd/adamic/main.go:check,compileLibrary`, `cmd/adamic/tsgo.go:buildWithNative`, `cmd/adamic/request.go:compileWASI` |
| 2 | usage error | `cmd/adamic/main.go:run`, `cmd/adamic/tsgo.go:buildWithNative` |
| 3 | the machine: the working directory can't be read, so a relative `-o` has nowhere to go. A missing or failing toolchain still exits 1 (#z1ba56x) | `cmd/adamic/tsgo.go:buildWithNativeExports` |
| 4 | the compiler itself panicked on `run`'s goroutine: stderr gets `adamic: compiler bug: <value>` and the stack. A panic on another goroutine, such as `internal/native`'s compiles (`internal/native/library.go:compileRuntime`, `internal/native/units.go:compileUnits`) or typescript-go's own, still ends the process with Go's 2 (#4mwb9et) | `cmd/adamic/main.go:run,reportCompilerBug` |

Exit 1 covers both "your program" and "your machine": a script can't tell a refusal from a missing
clang without reading stderr. Compiler has ruled the split (#z1ba56x): 1 the program, 2 usage
including an option the target refuses, 3 the machine, 4 a compiler bug. In the code so far: 4, and 3 for a
working directory that can't be read.

`TestUsageExitStatus` pins 2 and the exact usage text (`cmd/adamic/command_guards_test.go:TestUsageExitStatus`);
`TestBuildWASISanitizeTarget{Before,After}Source` pin a target refusal at 1 with its text, on a file
that loads (`cmd/adamic/command_guards_test.go:TestBuildWASISanitizeTargetBeforeSource,TestBuildWASISanitizeTargetAfterSource,assertWASISanitizeRefusal`); `TestBuildTargetParsing` pins three malformed
`--target`s at 2 (`cmd/adamic/target_test.go:TestBuildTargetParsing`); `TestCompilerBugExitsFour` and
`TestRunRecoversCompilerBugs` pin 4, its text, and that `run` defers the recovery, and
`TestCompilerPanicExitsFourThroughTheCommand` runs `adamic build` in a driver whose runtime supplier
panics inside `native.Build`, and sees 4 (`cmd/adamic/compiler_bug_test.go`). `TestBuildExitsThreeWhenTheWorkingDirectoryIsGone`
pins 3 and that the message keeps `-o` as typed, on Linux, where a removed working directory can't be read
(`cmd/adamic/output_path_test.go:TestBuildExitsThreeWhenTheWorkingDirectoryIsGone`).

### A program it built

| Code | Means | Where |
| --- | --- | --- |
| 0, or the program's own exit code | the program finished; `main` returns the stored exit code, 0 when none was set | `internal/native/emit.go:cProgram`, `internal/native/runtime/host_runtime.c:102-105` |
| 70 | a panic: an uncaught throw, a failed inserted check (non-null assertion, checked cast), stack exhaustion, out of memory, an index written out of range, and the runtime's other checked failures. stderr gets `adamic: panic: <message>` after everything the program printed | `internal/native/runtime/adamic.c:325-344`, `internal/native/runtime/exceptions.c:73-90` |
| 70, silently | stdout or stderr couldn't be written (a closed pipe), or a closed standard descriptor couldn't be reopened on `/dev/null` at startup | `internal/native/runtime/adamic.c:93-103,250-258` |
| a signal | a catchable stop signal flushes output and re-raises; SIGPIPE, SIGXFSZ and SIGUSR1 are ignored | `internal/native/runtime/adamic.c:137-173,281-290` |
| SIGABRT | an internal async-loop failure | `internal/native/runtime/async.c:171`, `internal/native/runtime/async_host_impl.h` |
| the sanitizers' own | under `--sanitize`, ASan and UBSan exit with their defaults; nothing sets `ASAN_OPTIONS` for a user build | `internal/native/native.go:Flags` |

70 is EX_SOFTWARE (`internal/native/runtime/adamic.h:1438`). `process.exit` and `process.exitCode`
use Node-compatible status validation (`internal/lower/process.go:lowering.processValue`,
`internal/native/runtime/process.c:59-70`). A `--count` imbalance never changes the exit code; the counts line is for
the test harness to judge (`internal/leakcheck/leakcheck.go:Unbalanced`).

`adamic js` output panics the same way, `adamic: panic: <message>` and 70, but only where its
`'adamic'` import resolves to `oracle/adamic.mjs` (`internal/javascript/javascript.go:JavaScriptWith`,
`oracle/adamic.mjs:34-55`), which only the oracle's loader arranges (`oracle/node.mjs`). Under plain
`node` the import doesn't resolve.

A wasm32-wasi panic leaves through the same `_exit(70)` (`internal/native/runtime/adamic.c:334-343`);
`TestWASIRequestThrows` sees 70 and `adamic: panic: Error: boundary` from the request host
(`cmd/adamic/request_test.go:TestWASIRequestThrows`).

## Environment

### Read by `adamic build`

| Variable | Effect | Default | Code |
| --- | --- | --- | --- |
| `PATH` | where `clang`, `llvm-ar` and `ar` are found | | `internal/toolchain/toolchain.go:Path` |
| `ADAMIC_TOOLS` | `<ADAMIC_TOOLS>/bin` then `<ADAMIC_TOOLS>/go/bin` are searched before `PATH`, and put first on the tools' `PATH` | unset | `internal/toolchain/toolchain.go:directories,Path,searchPath` |
| `ADAMIC_SANDBOX=1` | no fallback to `PATH`: a tool outside `ADAMIC_TOOLS` is an error naming where it looked, and the cache must be named by `ADAMIC_BUILD_CACHE_DIR` | off | `internal/toolchain/toolchain.go:Sandboxed,Path`, `internal/buildcache/location/location.go:BuildCache` |
| `WASI_SYSROOT` | required for `--target wasm32-wasi`, and must be a directory; the SDK's clang is taken from beside it, and must be wasi-sdk 27's | none | `internal/native/target.go:ValidateOptions,WASIClang` |
| `ADAMIC_BUILD_CACHE_DIR` | where the compiled runtime archive is cached | `os.UserCacheDir()/adamic-build` | `internal/buildcache/location/location.go:BuildCache` |
| `ADAMIC_BUILD_OUTPUT_DIR` | a separate directory to write cache entries to | unset | `internal/buildcache/location/location.go:OutputDirectoryVariable` |
| `ADAMIC_CC_JOBSERVER` | a FIFO; every clang call takes a token from it first. `internal/ccjobserver` reads it once, as the process starts, and `cmd/adamic` imports that package for its effect (`cmd/adamic/tsgo.go` imports it) | unset | `internal/ccjobserver/ccjobserver.go:Variable,process` |
| `ADAMIC_NATIVE_SPLIT=1` | builds the program as split C units instead of one file | off | `internal/native/native.go:Build` |
| `ADAMIC_NATIVE_JOBS` | split builds only: parallel clang jobs, an integer | 1 | `internal/native/units.go:buildUnitsWithLibrary` |
| `ADAMIC_GATE_UNCACHED=1` | split builds only: bypasses the unit cache | off | `internal/native/units.go:buildUnitsWithLibrary` |
| `ADAMIC_NATIVE_UNIT_CHECK` | split builds only: how often a unit taken from the cache is compiled again to check its bytes; set at all, every hit | one hit in 20 | `internal/native/units.go:unitCheckVariable,unitChecked` |
| `ADAMIC_NATIVE_UNIT_CACHE=off` | split builds only: every unit compiles uncached | on | `internal/native/units.go:unitCacheOffVariable,unitCacheOff` |
| `ADAMIC_BUILD_LOG` | appends a line per clang compile, build and archive step, with its key, outcome and seconds | unset | `internal/native/library.go:buildLog`, `internal/native/units.go:unitBuildLog,noteUnitMismatch` |
| `TMPDIR` | where build directories are made | system default | `internal/native/native.go:buildOneWithRuntime`, `internal/native/units.go:buildUnitsWithLibrary` |

clang and ar run under a fixed environment, not the caller's: only the variables in
`internal/toolchain/toolchain.go:Env,GoSettings,goLocations,carried,gocacheSettings,fixed,compilers` pass through, `GOENV=off ZERO_AR_DATE=1` is always set, and
`CC=clang CXX=clang++` are set when unset. So `CPATH`, `SDKROOT` and `DYLD_*` never reach clang, and
setting `CC` doesn't change which compiler a native build runs: that is always `clang` resolved as
above (`internal/native/target.go:compilerName`).

The other `ADAMIC_BUILD_*` store settings belong to `internal/buildcache`,
which `cmd/adamic` doesn't link (`internal/buildcache/location/location.go`'s package comment); they matter only to
the test harness.

### Read by a program it built

| Variable | Effect | Code |
| --- | --- | --- |
| `ADAMIC_THREADS` | worker threads, an integer from 1 to 4096; anything else panics (70) | `internal/native/runtime/parallel.c:126-132` |
| `ADAMIC_LEAK_CHECK` | an intentional exit writes `adamic: intentional exit: status N` for the leak harness | `internal/native/runtime/adamic.c:368-373` |
| `ADAMIC_TSGO_PROFILE`, `ADAMIC_TSGO_TIMING` | tsgo builds only: checker profiling and timing | `internal/native/runtime/tsgo.c:16-17` |
| `TMPDIR`, `TMP`, `TEMP` | `os.tmpdir()` | `internal/native/runtime/node_process.c:403-410` |

## Toolchain

| Tool | Needed for | Found by | Version | Missing |
| --- | --- | --- | --- | --- |
| clang | `build`, native | `toolchain.Path("clang")` (`internal/native/library.go:runtimeInputs`) | not checked: its `--version` report only keys the runtime cache (`internal/native/library.go:runtimeInputs,runtimeKeyParts`). `cloud/setup.sh:96` installs 20.1.8 (`ADAMIC_LLVM_VERSION`) when no clang passes its sanitizer probe | `adamic: native: exec: "clang": executable file not found in $PATH`, exit 1 |
| lld | `build`, shipped native release on Linux | clang's `-fuse-ld=lld` (`internal/native/native.go:linkFlags`) | | clang's link error, exit 1 |
| `llvm-ar`, else `ar` | the runtime archive | beside the compiler, else `toolchain.Path("ar")` (`internal/native/target.go:archiverName`) | | exit 1 |
| wasi-sdk | `--target wasm32-wasi` | `WASI_SYSROOT` (`internal/native/target.go:ValidateOptions`); clang from `<sysroot>/../../bin/clang` (`internal/native/target.go:WASIClang,compilerName`) | pinned to 27: a clang whose `--version` doesn't begin `clang version 20.1.8-wasi-sdk` is refused by name (`internal/native/target.go:WASISDKVersion,checkWASIClang`). `cloud/setup.sh --wasi-sdk` installs it (`cloud/setup.sh:160-170`) | no `WASI_SYSROOT`: `adamic: native: wasm32-wasi requires WASI_SYSROOT`, exit 1. No SDK clang beside it, or another clang: `adamic: native: wasm32-wasi is pinned to wasi-sdk 27 ...`, exit 1 |
| a tsgo checker archive | `--tsgo` | the path given (`internal/native/tsgo.go:checkerArchivePath`), built with `go build -buildmode=c-archive ./bridge/tsgo/archive` (`bridge/tsgo/README.md`) | | `adamic: native: checker archive: <stat error>`, exit 1 |
| Node | running `adamic js` output, and the oracle | not used by `adamic` itself | `cloud/setup.sh` installs 24; only `internal/native/wasm_test.go:runWASIUnit` checks a minimum (24) | |
| Go | building `adamic` | | `go 1.27`, `toolchain go1.27.1` (`go.mod:3,5`) | |

The binary ships no toolchain: it runs whatever clang it finds. The runtime's C files are embedded
and compiled on first use into `<cache>/native-runtime/<sha256>/runtime.a`, keyed by os, arch, the
compiler's path and version report, the archiver, the headers read and the flags
(`internal/native/library.go:runtimeLibrary,runtimeKeyParts,runtimePlatformParts,cachedRuntime`).

### Flags clang gets

Every build: `-std=c11 -Wall -Wextra -Werror -Wcast-function-type-strict -pedantic` and the unused
warnings off (`internal/native/native.go:Flags`), `-ffp-contract=off` so arithmetic rounds as Node does
(`internal/native/native.go:Flags`), `-fno-optimize-sibling-calls` (`internal/native/native.go:Flags`).
Native adds `-pthread` (`internal/native/native.go:Flags`); wasm32-wasi adds
`--target=wasm32-wasi --sysroot=$WASI_SYSROOT -DADAMIC_TARGET_WASI=1 -mno-atomics`
(`internal/native/native.go:Flags`). Then `--sanitize` gets `-O1 -g` and the sanitizers, wasm32-wasi `-Oz`, and everything
else `-O2` (`internal/native/native.go:Flags`). `adamic build` always asks for a release
(`cmd/adamic/tsgo.go:buildWithNative`), and a native build with neither `--count` nor `--sanitize` adds
`-flto=thin` (`internal/native/native.go:Flags`). `cmd/adamic/release_test.go:auditReleaseBuild` holds that
policy, and that the two semantic flags never drop.

## Targets

| Target | Selected by | Tested on main by |
| --- | --- | --- |
| native, Linux x86_64 and arm64 | the host; no flag | `internal/oracle` `TestNativeAgreesWithNode_*` (`internal/oracle/families_generated_test.go:TestNativeAgreesWithNode_000`, body at `internal/oracle/oracle_test.go:runNativeAgreesShard`), leak checks, whose combined LeakSanitizer run is Linux-only (`internal/oracle/oracle_test.go:nativelyCheckingLeaks`), `internal/native` `TestReleaseFlagsReachLink` (`internal/native/release_build_test.go:TestReleaseFlagsReachLink`) |
| native, macOS | the host; no flag | darwin branches in code (`internal/native/native.go:buildOneWithRuntime`, `internal/native/library.go:RuntimeLinkFlags`) and tests with darwin-only assertions (`internal/oracle/moves_darwin_test.go`, `internal/native/whole_build_paths_test.go:TestAWholeBuildNamesNoTemporaryObject`, `internal/native/machine_paths_test.go:TestCheckerBuildsNameNoPlaceOfTheArchive,TestSplitBuildsNameNoPlaceOfTheirInputs`). Loom's Go sources never name darwin, so as far as its code shows these run only where someone runs the suite on a Mac |
| wasm32-wasi | `--target wasm32-wasi` | parsing and refusals ungated (`cmd/adamic/target_test.go`, `cmd/adamic/command_guards_test.go:TestBuildWASISanitizeTargetBeforeSource,TestBuildWASISanitizeTargetAfterSource`, `internal/native/target_test.go:TestWASITargetFlags`); compiled programs only with `ADAMIC_ORACLE_WASI=1` (`internal/oracle/wasi_test.go:runWASIFixtures`, `cmd/adamic/request_test.go:TestWASIRequest,TestRequestNativeStaysCommand,TestWASIRequestModuleSelection,TestWASIRequestThrows`) or `ADAMIC_TEST_WASI=1` (`internal/native/wasm_test.go:runWASIUnit`). Loom's gate sets both (`GateEnvironment` in loom's planner); `cloud/setup.sh` and `cloud/write-env.sh` set neither, so a worker's own `go test` skips them |
| JavaScript | `adamic js` | the oracle's third run of every fixture (`internal/oracle/oracle_test.go:onJavaScriptBackend,runNativeAgreesShard`), the fuzzer (`internal/fuzz/run.go:Checkout.tryFile`) |
| tsgo-linked native | `--tsgo <archive>` | `bridge/tsgo/products_test.go:buildBridgeProduct` and the `TestBridge*` tests over it |

There is no Windows, iOS, Android or embedded target: `--target` accepts only wasm32-wasi
(`cmd/adamic/tsgo.go:buildWithNative`, `internal/native/target.go:ValidateOptions`), and nothing in the runtime or the
native Go branches on architecture. `cloud/setup.sh` provisions Linux x86_64 and aarch64 only
(`cloud/setup.sh:49-53`).

## Where other docs disagree

- `docs/wasm.md:5-6` says the branch changes neither `native.go` nor the driver; the driver has
  taken `--target wasm32-wasi` since (`cmd/adamic/main.go:run`).
- `docs/wasm.md:47-48` runs `go test ./internal/native -run '^TestWASI$'`, which matches no test and
  passes having run nothing; the tests are `TestWASIUnit00` to `35` (`internal/native/wasm_test.go:TestWASIUnit00,TestWASIUnit35`).
- `docs/wasm.md:38-46` installs WASI SDK 27 by hand; `cloud/setup.sh --wasi-sdk` does it and
  `cloud/write-env.sh:36-37` sets `WASI_SYSROOT`.
- `docs/lint-registration.md:89` shows `go run ./cmd/adamic build stage1/cohere/lint/main.ts` without
  `-o`, which exits 2.
- `cloud/setup.sh:7`'s header names only `--warm-tests`; its usage line names `--wasi-sdk` too
  (`cloud/setup.sh:25`).
- `CLAUDE.md` lists `adamic types | c | build | js <file>`, without `-o` or the flags; this page is
  the full form.
- `README.md` says macOS and Linux today and JavaScript for the web, which matches; the JavaScript
  output still needs the oracle's `'adamic'` module to run.

## Open gaps

- Exit 1 is both a refused program and a missing or failing toolchain (#z1ba56x).
- No `--help` that exits 0 and no `--version` (#a733mfq).
- Nothing checks the host clang's version; wasm32's is pinned to wasi-sdk 27 (#h2vm51c).
- A panic on another goroutine than `run`'s still exits 2 (#4mwb9et).
- `build`'s source and `-o` are positional (#21m2dv3).
- `adamic js` output runs only under the oracle's loader (#qq5dkee).
- Toolchain downloads carry no checksum, and there is no release of `adamic` (#xx66xas).
- The other docs listed above (#gyg0wrm).
