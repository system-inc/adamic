Built: shipped native release ThinLTO, complete semantic link flags, opt-in release oracle, recursion/FMA mutant proofs; ordinary lanes unchanged.
Commits: implementation b2c8836c9934e343f3300fb409524ab213828941; runtime 36669add9db2bbc06d486a58c37024f880e94a75 merged first as 574ce70a58cde3f27715bc8ceadd437ae6dfbc41.
Commands and outputs: release oracle 420 fixtures PASS; parse 15.67% and service 13.83% less user time; full good outputs identical.
Mutants: runtime dtoa.c-only contraction omission and link-only tail-call and contraction omissions caught; nonshipping LTO policy mutants caught; compiled AST/service byte changes caught; release-oracle one-byte change caught; instruction-accounting +1 caught.
Not covered: complete root/cohere gates, hardware cycles/IPC, host-wide isolation, macOS execution, WASI/TSGo shipping execution, or LTO atop activated SCC stack-check elimination.

# Shipped native release ThinLTO

The default is implemented for native `adamic build` shipping output. Counted,
sanitized, ordinary test/oracle and WASI policies retain their exact previous
flags. The original measurement and instrument-check report is preserved in
[measurement.md](measurement.md), with its original raw evidence. This report
supersedes that report's proposal with the approved implementation and a fresh
comparison after merging the current runtime.

## Implementation and semantic flags

`cmd/adamic/tsgo.go` selects `Options{Release: true}` for builds. `Flags` adds
`-flto=thin` only for native, uncounted, unsanitized release builds.
`RuntimeLibrary` already includes the complete flags, compiler identity and
runtime snapshot in its cache key, so its existing cache separates ThinLTO
bitcode archives from ordinary machine-code archives. Every runtime C unit
uses those flags. `LinkFlags` passes **the entire compilation list** to the
combined generated-C compilation/link invocation, adding `-fuse-ld=lld` on
this Linux box. Darwin uses its platform linker. `BuildTSGo` uses the same
link policy; its external Go archive remains machine code.

Semantic options in the policy are `-std=c11`, `-ffp-contract=off`,
`-fno-optimize-sibling-calls`, and conditional `-DADAMIC_COUNT`,
`-DADAMIC_SLABS`, `-march=...`, `--target=wasm32-wasi`, `--sysroot=...`,
`-DADAMIC_TARGET_WASI=1` and `-mno-atomics`. Warning, optimization and
sanitizer arguments also reach the link. There is no duplicated allowlist
that can omit a future flag. See [native-builds.md](../../../docs/native-builds.md)
and the comment above `native.Flags`.

`TestNonShippingFlagsStayIdentical` compares NUL-separated argument bytes and
order against independent literals from before this change, for ordinary
oracle/test, counted, sanitized and combined count/sanitize policies.
Three actual source-policy mutants independently enable ThinLTO for ordinary,
counted or sanitized lanes; each fails that test. The CLI audit also records
actual compilation and link invocations for all three build modes.

## Node-held semantic proofs

The release builds of `stack_forever.a` and `stack_tail_call.a` both match
Node's stack panic and exit 70. The actual compiler wrapper removes only
`-fno-optimize-sibling-calls` from the generated-program compilation/link;
runtime `clang -c` invocations keep it. The unbounded mutant loops past a
two-second deadline. The self-tail-call mutant exits 0 and prints
`0\n0\nend\n`, disagreeing with Node. The wrapper audit verifies both link
mutations and preserved runtime compilation flags.

The new `.a` fixture [release_fma.a](../../../internal/oracle/testdata/release_fma.a)
loads operands through program arguments to prevent constant folding. It tests
`0.1*10-1`, `(1/3)*3-1` and `0.1*0.1-0.010000000000000002`, with the operands
parsed from decimal strings. Node and protected release ThinLTO print
`0\n0\n0\n`. The mutant drops only `-ffp-contract=off` from the generated-C
compilation/link, retaining the protected runtime archive. This EPYC has FMA;
the check uses `-march=haswell` on **both** the good and mutant release builds.
The mutant prints `5.551115123125783e-17`, `-5.551115123125783e-17`, and
`-8.326672684688674e-19`, so fusion is actually visible. A separate volatile-C
arithmetic probe also catches the omission. No warning or sanitizer kills
these mutants; the Node-held output or termination comparison does.

## Whole release oracle

The environment switch `ADAMIC_ORACLE_RELEASE=1` adds the release lane;
ordinary test runs skip it. Every registered fixture is checked/lowered again
and built directly with `native.Options{Release: true}`, without a CPU override:
these are the **exact generic shipped flags**. Stdout, stderr and status match
source on Node, or the same explicit inserted-check JavaScript oracle used by
the ordinary lane. Inserted-check fixtures must distinguish unchecked Node.
Unsupported entries must still refuse explicitly. Specialized permission,
stream, leak and counter probes retain their original policies.

~~~sh
source /workspace/adamic-tools/env.sh
ADAMIC_ORACLE_RELEASE=1 ADAMIC_GATE_UNCACHED=1 \
  go test -v -count=1 -timeout 30m ./internal/oracle \
  -run '^TestRelease(AgreesWithNode|OracleCatchesOneByte)$' -parallel 4 \
  > release-oracle-final.log 2>&1
~~~

PASS: all **420** fixture subtests, 142.073 seconds including the compiled
one-byte output mutant. Node observation cache: 0 hits, 431 misses. The additional
misses come from checks and the mutant, not additional fixture-table entries.
The new FMA fixture also passes the ordinary sanitized differential lane
(1.387 seconds), and its count row was added without changing any existing row.
This switch can run nightly or once per main push; it is not added to ordinary
runs.

## Box, setup and instrument limitation

The box is Linux x86-64 under KVM, AMD EPYC 9V74, 5 visible CPUs (`nproc=5`),
a four-core cgroup quota (`cpu.max=400000 100000`) and 17.6 GB reported memory.
clang/lld 20.1.8, LLVM revision `87f0227cb60147a26a1eeb4fb06e3b505e9c7261`;
Go 1.27.1; Node 24.19.0; Callgrind 3.24.0; hyperfine 1.19.0.
The ordinary linker is GNU ld 2.44. Production's archive selection on this
PATH falls back to GNU `ar`; lld reads the archived LLVM bitcode successfully.

`bash cloud/setup.sh` initially failed because the historical evidence copy
`evidence/go_parse.go` was accidentally treated as a root-module Go package
and imported the upstream parser shim. It is now preserved as `.go.txt`, with
the same content hash. Setup retry passed: Go 0s, clang 0s, Node 0s, submodules
0s, cache warm 86s, total 86s. Build/test shells source
`/workspace/adamic-tools/env.sh`.

The earlier ten-pair pinned instrument check found native/Go **5.529x user
time**, versus **4.070x Callgrind instructions**, outside the 10% band. The
instruction ranking is therefore **not confirmed**. Hardware `perf stat -r 10`
events were unsupported, rather than denied by `perf_event_paranoid=2`.
The permitted cache/branch simulation is preserved in the original report.
Its hypothetical central cost model moves remainder 2 to 1, scanner 1 to 2,
retains 6 to 5, string equality 5 to 6, substrings 14 to 13 and line table
13 to 14. Releases remain 4th centrally, move to 3rd with higher miss penalties;
slab allocator stays 11th, separate from allocation entry/libc at 10th.
These are **model rankings, not measured cycles**; working PMU counters are
still needed to establish cycle ranks. The report and evidence retain that
limitation explicitly.

## Post-implementation measurements

Fresh production-built baseline and ThinLTO binaries use the same generated
C and current runtime bytes. The parse driver is batch 8's parse-only harness,
with the same 77 TypeScript files; the service is the six-route `a4e0902`
harness. Corpus and harness pins are in the original report.

Both full AST outputs match Node and Go, **44,766,682 bytes**, SHA-256
`8ae015600498b915cc25abab82730299451ae990b50478980d5a3bc465801bfe`.
Both full service outputs match Node, **8,018,702 bytes**, including the checksum
`7394547`. Actual compiled same-length mutants change `SourceFile` to
`XourceFile` (first differing byte 14) and health status 200 to 201 (byte 12);
both are caught by the full-byte comparison despite unchanged checksum.

Hardware events remain unsupported on the fresh ten-run perf probe. Timings use
the specified hyperfine fallback: core 3, five interleaved rounds, alternating
baseline/ThinLTO order, after an explicit untimed warmup of each binary. Each
invocation has `--runs 1 --warmup 0 --shell none --show-output`; output is checked
on every run. Best wall and best user are selected independently. No build, test
or profiler ran concurrently with these timings. Background container services
remain; host-wide isolation/frequency cannot be certified.

| Workload | Variant | Callgrind Ir | Best wall s | Best user s |
| --- | --- | ---: | ---: | ---: |
| parse | baseline | 6,254,103,417 | 0.839264 | 0.798564 |
| parse | thin | 5,658,151,583 | 0.721729 | 0.673444 |

Parse saves 14.00% wall, 15.67% user, and 9.53% instructions.

One-minute load during parse: 2.03 to 2.12. These averages still decay from the earlier gate; its live processes were cleared first. Raw JSON retains all three load averages for every run.
| service | baseline | 93,130,684,617 | 7.572985 | 7.160951 |
| service | thin | 86,530,992,097 | 6.600901 | 6.170402 |

Service saves 12.84% wall, 13.83% user, and 7.09% instructions.

One-minute load during service: 1.23 to 1.80. These averages still decay from the earlier gate; its live processes were cleared first. Raw JSON retains all three load averages for every run.

Both workloads still exceed the approved 10% wall-time threshold. The release
default is therefore supported by this fresh comparison; ordinary policies stay
as they were. These new runtime-base results replace the historical numbers,
rather than mixing binaries from the two bases.

| Workload | Variant | Cold backend wall/user s | Cached-runtime wall/user s | Binary bytes |
| --- | --- | ---: | ---: | ---: |
| parse | baseline | 5.209 / 4.496 | 2.441 / 2.382 | 972,040 |
| parse | thin | 17.559 / 17.444 | 17.708 / 18.030 | 1,339,872 |
| service | baseline | 2.754 / 2.175 | 0.219 / 0.185 | 406,904 |
| service | thin | 2.241 / 2.403 | 0.514 / 1.145 | 163,912 |

These are single build observations, not best-of-five timings. Cold means a
separate empty runtime.a cache for each workload/variant; cached means the next
build with that exact archive present. Inputs are already emitted C, so these
times exclude Adamic checking/lowering/C emission and Go tool compilation.
Tool/filesystem page caches are warm; this is not a host-wide cold-page test.
Builds may use multiple cores, so aggregated user time can exceed wall time.
One-minute build load was 1.00 to 1.07. The parser cached ThinLTO build happens
to be slightly slower than the cold observation: caching saves runtime frontends,
but the much larger link/backend cost and run variation dominate it.

Cached parse backend cost rises from 2.441s to 17.708s wall, and its binary
grows from 972,040 to 1,339,872 bytes (37.84%). Cached service cost rises from
0.219s to 0.514s wall; its binary shrinks from 406,904 to 163,912 bytes.

The exact common baseline flags, in argument order, are:

~~~text
-std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O2
~~~

Every runtime command is `clang [those flags] -c <runtime-file.c> -o <unit.o>`.
The generated-program command is `clang [those flags] -I <runtime-cache-dir>
-o <binary> <temporary-main.c> -Xlinker --whole-archive <runtime-cache-dir>/runtime.a
-Xlinker --no-whole-archive -lm`. ThinLTO adds `-flto=thin` to the same complete
list on every compilation, and `-fuse-ld=lld` at the program link. There is no
benchmark CPU override. The paths for temporary main C change each build; its
bytes are the committed compressed C snapshot. Flag arrays and actual archive
paths for both production builds are in the `*-build.json` evidence.

Reproduction after the original harness preparation, from this repository root:

~~~sh
go build -o /tmp/build-shipped ./cloud/reports/release-lto
/tmp/build-shipped "$scratch/parse.c" "$scratch/baseline/parse"
/tmp/build-shipped -release "$scratch/parse.c" "$scratch/thin/parse"
/tmp/build-shipped "$scratch/service.c" "$scratch/baseline/service"
/tmp/build-shipped -release "$scratch/service.c" "$scratch/thin/service"
python3 cloud/reports/release-lto/instrument.py "$scratch" --lto-only
python3 cloud/reports/release-lto/profile-shipped.py "$scratch" "$original_scratch"
python3 cloud/reports/release-lto/summarize.py "$scratch"
python3 cloud/reports/release-lto/build-times.py "$scratch"
~~~

The timing helper needs `scratch/tools` pointing to the extracted hyperfine
tools. The profile helper uses the original scratch Valgrind install. The raw
four profiles reconcile self costs exactly to instruction summaries; a
summary-plus-one mutant is rejected. Profiles, timings, build observations,
good-output hashes and proof logs are in
[implementation-evidence](implementation-evidence/).

## Validation and limits

All test output was written to logs. The committed evidence includes:

- `gofmt -l cmd internal` and the benchmark Go helper: no output; `go vet ./...`: PASS.
- Focused native release flag/volatile arithmetic proofs: PASS, 2.042s.
- `TestReleaseFixtureKeepsArithmeticUnfused`: PASS, 1.378s, including the three observed fused residuals above.
- `TestReleaseRecursionKeepsFrames`: PASS, 17.024s, both link-only mutants caught and runtime compilation audited.
- `TestBuildSelectsOnlyShippedReleaseLTO`: PASS, 11.430s, actual CLI compiler commands audited.
- Three separately run production policy mutants: ordinary test/oracle, counted and sanitized flags each fail the byte golden; production source restored after every run.
- Complete final opt-in release lane: PASS, 142.073s; full fixture table including the new FMA fixture.
- Ordinary sanitized oracle for the new fixture: PASS, 1.387s.
- `TestCountsAreRecorded -args -update-counts`: PASS, 48.881s; exactly one new row, all old rows byte-identical.
- `ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./...`: interrupted after over 30 minutes because the remaining stage-1 ports were too slow. Complete touched-package results already printed: `cmd/adamic` PASS 41.571s, `internal/native` PASS 648.782s, `internal/oracle` PASS 281.914s. The later-added FMA fixture has the separate focused and final-release runs above. This is **not a full-root-gate pass**. Its owned process tree was terminated and cleared before timing; remaining stage-1 ports were not completed.
- Fresh full AST/service byte checks and their compiled same-length mutants: PASS.
- Fresh four-profile instruction reconciliation: PASS; summary-plus-one mutant caught.
- Additional cohere check for the new `.a` fixture could not run: named-path mode rejects `.a`, and a two-file config inheriting Adamic's options, also tried with explicit `sourceExtensions: [".a"]`, reports that only one of two configured files matched the type graph. Exact errors are retained. No cohere source was changed; stage-0 checked/lowered the actual fixture and both ordinary/release Node oracles passed.

There was no good-output difference. The ordinary gate's broad interruption is
a coverage limit, not evidence that the unfinished ports pass. macOS ThinLTO,
WASI shipping execution, and native TSGo shipping execution were not run.


ThinLTO saves the runtime frontends through `runtime.a` caching, but still
performs ThinLTO backend work at every generated-program link. Developer tools
that split generated C across translation units must use the same compile
policy on every unit and the same full semantic policy at link, invalidate
object/cache keys by flags and toolchain, and expose their split workload for
a fresh benchmark. No emitter split is introduced here.

The requested combined measurement on `codex/stack-check-scc` is pending that
optimization actually landing. The supplied `01a114d8` object is unavailable
in the fetched branch; its current tip is `8abc29aa4b6c0f9fa1db3836c4400fd9fe78ef80`.
That branch's report says **inactive SCC classifier; production stack-check
placement unchanged**, and names the missing frame-headroom proof. The current
runtime also still emits stack checks in every function. Calling this an LTO
measurement atop removed non-recursive checks would be false. Re-run these
same helpers once the active optimization lands.

## Runtime compilation contraction audit

The developer-tools GCC-lane request is covered alongside the link-step tests
by TestRuntimeCompilesEveryUnitUnfused. It builds real cold runtime archives
through RuntimeLibrary for shipped release, ordinary sanitized oracle and
counted oracle policies, records the actual clang invocations, and checks all
48 unique runtime C files in each build. The last effective -ffp-contract=
option must be off; dtoa.c and ieee754.c are explicitly required.

The real mutant wrapper removes only -ffp-contract=off from dtoa.c compilation.
The archive still builds successfully, then the command audit catches the
missing protection. This proof does not depend on incidental FMA generation
in a particular runtime function. The separate release arithmetic fixture
continues to prove the link-only mutant by changed output on an FMA target.

Command: go test -v -count=1 -timeout 30m ./internal/native
-run '^Test(RuntimeCompilesEveryUnitUnfused|ReleaseFlagsReachLink|NonShippingFlagsStayIdentical)$'.
PASS in 15.865s; all three 48-unit audits passed and the dtoa.c mutant was
caught. Output: [runtime-compile-audit.log](implementation-evidence/runtime-compile-audit.log).
Only tests and documentation changed; measured production flags and binaries
remain the same.
