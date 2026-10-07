Built wasm32-wasi defaults -Oz and --strip-debug, preserving the name section and native flags.
Base: f53f0e6f71b880b88216106103703140ac8e5174; claim: 43e2197; branch: codex/wasm-release-flags.
Checks: full uncached WASI oracle 650 pass/0 fail/0 skip; runtime 37/0/0; command 20/0/0.
Mutants: omitted strip-debug, strip-all and reverted Oz all caught by their intended assertions.
Limits: no deployed Worker, performance timing or Binaryen execution; native evidence normalizes scratch paths.

The new branch starts from the requested wasm-exports commit; that branch was
not modified or pushed. The first push added only claim.md. W3's size table was
read from origin/codex/wasm-size:bench/wasm/README.md. No benchmark sources were
changed. internal/native/native.go adds a three-line WASI return in Flags;
internal/native/target.go adds --strip-debug to WASILinkFlags. Native -O2 and
sanitizer flags retain their existing paths. The existing target flags test's
WASI optimization expectation was updated; no explicit flags in the runtime
TestWASI were changed. Counted builds use Oz and keep their exports.

Setup: bash cloud/setup.sh --wasi-sdk > /tmp/wasm-release-setup.log 2>&1 exited 0.
Go, clang, Node, WASI SDK and submodules were ready at 0s; build cache warm and
done at 46s. nproc=5, cgroup cpu.max=400000 100000. Environment:
/workspace/adamic-tools/env.sh, including WASI_SYSROOT. Go 1.27.1, native clang
20.1.8, WASI SDK 27 clang 20.1.8, Node v24.19.0.

## Native equivalence

Before changing flags, adamic-before emitted C for
internal/oracle/testdata/collections.a. The identical C was built before and
after with native.Build and native.Options{}. Each run used a fresh isolated
XDG_CACHE_HOME, forcing all 48 runtime translation units to compile. A clang
wrapper logged every argv and copied each emitted object. The sample was also
compiled to an object using native.Flags with the runtime header include path.
The same native compiler and wrapper paths were used for both runs.

Both runs recorded 51 invocations: --version, 48 runtime object compilations,
the fixture compile/link, and the separate fixture object compilation. Raw argv
logs are /tmp/wasm-release/native-before.argv.jsonl and native-after.argv.jsonl.
Their printed JSON preserves every argument and argument boundary. The only
path normalization replaces randomized program/runtime temporary directories
and the before/after cache roots. The normalized argv files compare identically:

```sh
diff -u /tmp/wasm-release/native-before.argv.normalized.jsonl /tmp/wasm-release/native-after.argv.normalized.jsonl > /tmp/wasm-release/native-argv.diff
# exit 0; empty diff
```

All 49 native objects are byte-identical, including sample.o and every runtime
object. The SHA-256 map for both sides is /tmp/wasm-release/native-objects.json.
Generated fixture C, all argv logs and object copies are retained in
/tmp/wasm-release. This is a comparison of the sample and complete runtime,
not a claim to have compared every possible native program.

## WASI checks

Test output was written directly to logs, never piped. Counts below count
terminal Go test/subtest events with a Test name, excluding package events.
All required opt-in runs were enabled and none skipped.

| Run | Pass | Fail | Skip | Seconds | Exit |
| --- | ---: | ---: | ---: | ---: | ---: |
| Full WASI oracle | 650 | 0 | 0 | 217.751 | 0 |
| Runtime TestWASI | 37 | 0 | 0 | 35.517 | 0 |
| Command WASI/request checks | 20 | 0 | 0 | 18.708 | 0 |

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_ORACLE_WASI=1 ADAMIC_GATE_UNCACHED=1 go test -json ./internal/oracle -run '^TestWASI' -count=1 -timeout 30m > /tmp/wasm-release/oracle.jsonl 2>&1
PATH=/workspace/adamic-tools/wasi-sdk/bin:$PATH ADAMIC_TEST_WASI=1 ADAMIC_GATE_UNCACHED=1 go test -json ./internal/native -run '^TestWASI$' -count=1 -timeout 15m > /tmp/wasm-release/runtime.jsonl 2>&1
ADAMIC_ORACLE_WASI=1 ADAMIC_GATE_UNCACHED=1 go test -json ./cmd/adamic -run 'WASI|Request' -count=1 -timeout 15m > /tmp/wasm-release/command.jsonl 2>&1
```

The full oracle has 323 agreement fixture passes and 323 emission fixture passes,
plus their two parent tests and the two existing oracle/runner mutant tests.
Runtime TestWASI compiled all 48 runtime units with its explicitly supplied O2
flags, passed 35/35 fixture comparisons and its 100,000-request loop: live 0,
flat memory 393,216 bytes and 6,300,000 region objects. Its flags remain explicit.
The compiler-generated export test passes 37 boundary cases and 100,000 mixed
calls: live 0 after every call, flat memory 1,507,328 bytes, regions 406,476.
The command request test passes 10,000 calls, module initialization once,
524,288-byte flat memory, live baseline 2 for retained module state and 630,000
region objects. Uncaught request throws and native request command behavior pass.

The release section test builds hello with and without --count. The ordinary
module is 14,236 bytes and counted module 29,512 bytes; both list exactly
name, producers, target_features as custom sections.

Additional native runtime cache/count tests passed (6 tests, 2.322s). The WASI
target flags and unsupported-options tests passed (2 tests, 0.006s).
The ordinary touched-package gate passed, exit 0: internal/native 127.902s,
cmd/adamic 2.016s. Command:
go test ./internal/native ./cmd/adamic -count=1 -timeout 15m
Log: /tmp/wasm-release/touched-packages.log. Final gofmt -l cmd internal and
go vet ./... produced empty logs; vet exited 0. git diff --check was clean.

## Size observations

These use the unmodified W3 fixture paths: hello is
internal/load/testdata/0.1/compile/01_hello.ts; dedication is
dedication/dedication.a; collections is internal/oracle/testdata/collections.a;
request is internal/native/wasm/request.a. Each was built with the before/after
compiler, --target wasm32-wasi, no --count or postprocessing; request is selected
automatically as a reactor. The current W1 request includes its ordinary typed
export as well as its legacy entry, so these are new observations, not W3's
older request artifact. Raw hello/dedication differ by five bytes: the retained module-name entries
are hello.wasm and dedication.wasm, respectively.

| Fixture | Before raw | Before gzip9 | Before Brotli11 | After raw | After gzip9 | After Brotli11 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| hello | 237,404 | 83,599 | 69,967 | 14,236 | 6,032 | 5,231 |
| dedication | 237,409 | 83,603 | 70,009 | 14,241 | 6,029 | 5,195 |
| collections | 313,447 | 111,015 | 93,208 | 76,235 | 28,832 | 24,986 |
| request | 272,884 | 98,847 | 83,198 | 43,738 | 18,791 | 16,110 |

Gzip uses gzip -9 -n -c. The brotli CLI was unavailable; Node's built-in Brotli
encoder supplied quality 11 (BROTLI_PARAM_QUALITY=11). Every gzip and Brotli
artifact was decoded and checked byte-for-byte against its raw module. Node
also validated all eight modules with WebAssembly.Module. Full size/section
rows are /tmp/wasm-release/sizes.json; the measurement script is sizes.mjs.

Before: each module lists .debug_loc, .debug_abbrev, .debug_info, .debug_str,
.debug_line, .debug_ranges, name, producers, target_features (the request's
section order differs). After: each lists only name, producers, target_features.
No name section was stripped. The retained name section is why these results
are larger than W3's llvm-strip rows, which stripped more custom metadata.
No execution-speed conclusion follows from these size measurements.

## Mutants

Each production-source mutant was restored in a finally block. All three
compiled and failed the intended assertion; no compiler warning killed them.

| Mutant | Check and observation | Exit | Log under /tmp/wasm-release |
| --- | --- | ---: | --- |
| Omit --strip-debug | TestWASIReleaseSections finds all six .debug_ sections in ordinary and counted hello, sizes 236,962 and 252,238 | 1 | mutant-no-strip.log |
| Use --strip-all | Same test reports name missing for ordinary and counted hello, sizes 12,766 and 27,753 | 1 | mutant-strip-all.log |
| Revert Oz to O2 | TestWASITargetFlags reports missing -Oz in actual WASI flags | 1 | mutant-optimization.log |

Restored TestWASIReleaseSections passed in 0.446s (sections-restored.log).
The ABI build section documents the defaults and optional wasm-opt -Oz -g
postprocessing; Binaryen is not in the SDK and was not run here.

This unit did not rerun the full repository gate or deploy to Workers. Its
scope is the full requested WASI oracle and runtime/request/export checks,
ordinary touched packages, native compile/object comparison, size observations
and section/optimization mutants. No runtime C, emission or lowering changed.
