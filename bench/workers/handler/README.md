# Four-column handler benchmark

This bench runs `handleRequest(request: string): string` synchronously, without
HTTP. It builds and runs the same `.a` handler as native, Wasm through the actual
Workers host, Adamic JavaScript, and Node evaluating the source through
`oracle/node.mjs`'s type-stripping hooks. Each batch reads a line corpus and
answers it K times in a fresh process, consuming every response in a checksum.
No compiler changes or profiling are part of this unit.

## Required checkout and run

The default service and the Workers host belong to sibling branches, not this
harness. Merge them in a local measurement checkout first:

```sh
git checkout -b local-handler-measure origin/area/platforms
git merge --no-edit origin/codex/wasm-requests
git merge --no-edit origin/codex/workers-wasm-host
bash cloud/setup.sh --wasi-sdk > /tmp/handler-setup.log 2>&1
source /workspace/adamic-tools/env.sh
node /path/to/bench/workers/handler/run.mjs \
  --repo /path/to/measurement-checkout \
  --iterations 1000 --rounds 3 \
  --output /tmp/handler-results.json > /tmp/handler-results.md 2>&1
```

Use setup's printed environment path on another machine. On macOS, Node 24,
Go, Xcode clang, a WASI SDK clang/sysroot, and Binaryen 126 `wasm-opt` are required. Set
`WASI_SYSROOT` to the SDK's `share/wasi-sysroot`. Native clang stays on PATH;
The harness creates an isolated SDK wrapper beside the real sysroot for Wasm. Put `wasm-opt` on PATH or supply `--wasm-opt /path/to/wasm-opt`. The SDK must include `llvm-strip` and `llvm-ar`.

The delivery branch contains only `bench/workers/handler/`. The local measurement
checkout used service `f4ec96cf0a43c0b03b8f99dfcfd3bacb28e80569` and Workers host
`c7196b920db29c40e30210744a20a6c33a29d27f`, merged onto platforms `c323650`.
The exact merged compiler commit, dependency refs, source/corpus and binary hashes,
commands, compiler versions and dirty status are recorded in every raw report.
The service is imported by its original path,
`internal/native/wasm/service/service.a`; it is never copied or edited.

Default workloads are a balanced six-line service corpus covering `/health`,
`/catalog`, `/orders`, `/summary`, `/listing`, and `/quote`, and the sieve control
with limits 1000, 10000 and 100000. The service corpus includes Unicode, an escaped
lone surrogate, ordering, word counts and cent/tax rounding. It is not the
sibling's seeded 100,000-request corpus or its error/large-body distribution.
The sieve uses a boolean array and `!== true` tests; its independent source pins
are 168, 1229 and 9592 primes.

For an arbitrary handler and corpus:

```sh
node bench/workers/handler/run.mjs --repo /path/to/merged-checkout \
  --entry /path/to/entry.a --corpus /path/to/requests.txt --name custom \
  --iterations 1000 --rounds 5 \
  --output /tmp/custom-handler.json > /tmp/custom-handler.md 2>&1
```

Each physical line is one request string. A final line terminator is removed,
CRLF is accepted, empty interior lines are real requests, and nothing is trimmed
or parsed as JSON by the harness. Literal newlines inside one request require a
handler-specific escaped encoding. Custom handlers may import other `.a` modules;
the generated driver imports the original entry by relative path. They must
export the named handler and obey the Workers host's supported import inventory.
A handler needing filesystem imports inside Wasm will be refused by that host.

## Flags and artifacts

| Flag | Default | Meaning |
|---|---|---|
| `--repo` | current directory | merged compiler/host checkout |
| `--entry` | default service and sieve | custom `.a` handler; pair with corpus |
| `--corpus` | included corpora | custom request-lines file; pair with entry |
| `--name` | `handler` | custom workload name |
| `--iterations` | 1000 | K complete corpus iterations per batch |
| `--rounds` | 3 | interleaved rounds |
| `--output` | `handler-results.json` | complete raw report |
| `--artifacts` | new scratch directory | builds, drivers, stdout/stderr and rusage files |
| `--adamic` | build from repo | optional existing compiler binary path |
| `--wasm-opt` | `wasm-opt` on PATH | release optimizer executable (126 used here) |
| `--node` | current Node executable | executable for all three Node-hosted variants |
| `--timeout-ms` | 300000 | per benchmark-child timeout |
| `--build-timeout-ms` | 600000 | compiler/build timeout |

All effective flags appear in both outputs. Fixed build choices are also named:
**clang -O2** for native, **clang -Oz** for Wasm entry and runtime,
**llvm-strip** followed by **wasm-opt -Oz**, no sanitizers, no allocation counting,
and Node's experimental-warning suppression. This matches the release configuration
measured in `codex/wasm-size` at `aada8365085b8fea9f6cfd79230badbaa0dfce20`.
No compiler edits are needed: `clang-wrapper.mjs` replaces the compiler's `-O2`,
then strips and optimizes successful links. Its version output includes the
configuration, preventing reuse of runtime archives built with other flags.
Every wrapped command is retained in `release-commands.jsonl`. Whole-archive linking
and reactor initialization are retained. The real SDK is unchanged.
Artifacts and compiler binaries are hashed; an external compiler does not prove
it came from the recorded checkout, so retain its provenance separately.
Build commands and their stdout/stderr paths are recorded in `builds.json` and the
raw report. Build time is excluded from benchmark samples. All artifacts remain
available after the run for inspection.

## Artifact sizes and startup phases

A separate table reports raw and Brotli bytes, Wasm compile and instantiate plus
module initialization, JS parse plus module load, and full K=0 process startup,
with best and median of N. Raw artifacts are the native executable (including its
driver/runtime), the Wasm reactor, emitted unminified JS driver, and the Node source
graph (driver, checksum, handler and relative imports). These differ in bundled
contents; external host/oracle runtime files are excluded. Node source sizes sum
the individual original files. All variants use Node zlib Brotli quality 11,
generic mode; modules are compressed individually, round-tripped and retained as
`.br` files. The header names the Brotli version, and JSON retains a size manifest.
Compression is outside all timers. No service source is copied into this unit.

`startup-probe.mjs` runs in a fresh Node process for every phase observation. Wasm
bytes and the production WASI shim are loaded before timing. It times asynchronous
`WebAssembly.compile(bytes)`, then `new WebAssembly.Instance` plus `_initialize`
using that shim, just as the Workers host initializes its reactor. Import validation
precedes the instantiate timer; no request runs. These probes do not replace the
actual Workers crossing in the measured variant. JS probes time the unchanged
oracle bootstrap, source parsing/type stripping where applicable, module loading
and the K=0 corpus read together. Those phases cannot be separated reliably by
this public interface. Native AOT compilation is build work; separate native loader
and module initialization timers are unavailable, so these cells say `n/a` and the
full K=0 process startup is shown. Node process launch is excluded from phase
probes and included in full process startup. Phase probe samples and invocations
are retained separately from request samples; variants rotate during both passes.

## What a sample means

Native and Adamic-JS use the same generated `.a` driver, which reads the corpus
with `readTextFile`, calls the original handler in nested K/corpus loops and prints
one final digest. Node-source runs that exact driver with the repository's oracle
hooks, never Adamic's lowering. Adamic-JS also uses this bootstrap to resolve its
`adamic` runtime import; emitted `.mjs` is executed without type stripping. Wasm builds the original entry as a reactor with
`adamic build --target wasm32-wasi`; its Node driver calls
`internal/worker/wasm/host.mjs`'s `createHost(module).call(request)`. That includes
Workers' UTF-8 encode/copy, malloc, request ABI, response copy/decode and owned
response/input cleanup. It does not substitute Node WASI or a faster crossing.

`checksum.a` supplies a length-framed, order-sensitive pair of 32-bit rolling
hashes over every UTF-16 response unit, plus total response count and unit count.
Only the final digest is printed, avoiding per-response terminal I/O. All four
consume every returned response, so dead-code elimination cannot discard the
handler work. Hashes are checksums, not a collision-free byte transcript; agreement
covers the supplied corpus, with the ordinary possibility of hash collisions.
Hashing and iteration overhead are included, with compiled hashing in native and
Adamic-JS, and source hashing in Node-source and the Wasm host. This is a
handler-plus-driver comparison, not a claim to have removed all driver overhead.

Before timing, all four run K=1 and must match Node-source's digest. A separate
Node-source run supplies the expected K digest. Every measured batch is checked
against it, and every startup batch must have the empty K=0 digest. Invalid counts,
child errors, timeouts and checksum mismatches refuse the run. A refused run saves
its checks, failed child and any already-collected raw samples, but prints no
performance table. The checksum mutation proof fails preflight with zero timing
samples.

Each measurement uses a small C helper's **`wait4` direct-child rusage**:
user plus system CPU, minor/major faults, context switches and maximum RSS.
Linux `ru_maxrss` is KiB and is converted to bytes; macOS reports bytes. Peak RSS
belongs to the whole process, not the isolate or heap. The helper times immediate
pre-fork to reap with `CLOCK_MONOTONIC`; helper launch before that is excluded.
A signal alarm terminates the benchmark child's process group at timeout. The
benchmark parent's CPU and the helper's CPU are not charged to the child. The
helper is compiled with `-O2 -std=c11 -Wall -Wextra -Werror -pedantic`, recorded
in the header; it avoids a Python launcher's inherited RSS floor. Child maximum
RSS can still include the small pre-exec image, as OS rusage defines it.

Every variant/round has two fresh child processes: K measured, and **K=0 startup**.
For each pair, wall/request is `(wall(K) - wall(0)) / (K * corpusLines)`; CPU/request
uses the same subtraction on user+system CPU. Raw wall/CPU per request and startup
wall/CPU are shown alongside the corrected numbers. RSS is reported for both
processes, never subtracted: a difference of two lifetime peaks is not live memory.
The JSON keeps both complete child samples, stdout, stderr, checksum and all
counters. No negative difference is clamped; small K can be dominated by startup
noise. Large enough K and repeated rounds are needed for performance conclusions.

K=0 loads the corpus and modules but calls no handler. The production Workers host
is lazy: K=0 compiles the Wasm module and creates the host, while first-call reactor
instantiation is inside K>0 and amortized over its requests. That is the actual
host behavior, not a replacement initializer. No warmup is performed; JIT and
first-use costs are included. Source type stripping is part of Node-source startup.
This comparison measures that explicit source execution path, not a pretranspiled
TypeScript Worker bundle.

Variants rotate one position each round. Even rounds run startup before measured;
odd rounds reverse that order. Processes never overlap. `uptime` load averages
are recorded before/after timing, after builds. Best and median of N are computed
independently for each metric across paired samples. Comparisons on a shared
machine are observations; no compiler or platform speedup is inferred here.

## Failure proofs

```sh
node bench/workers/handler/prove.mjs /path/to/merged-checkout /tmp/handler-proofs \
  > /tmp/handler-proofs.log 2>&1
```

The proof builds all four variants, checks corpus counts/order, pins the sieve
against Node, and runs both mutations:

1. A generated `.a` driver changes exactly its first handler response by replacing
   its final unit with `!`, before the unchanged checksum. Count and total length
   stay equal, so only the response hashes can catch it. The mutant is compiled with `adamic js`.
   All other variants keep their original source. The checksum comparison must
   refuse it at preflight, with no timing samples.
2. A source mutant removes startup subtraction from both wall and CPU reductions.
   A real K=1/K=0 run supplies three paired rounds. The original and mutated
   reducers consume the **same captured samples**, so process noise cannot hide
   the removed operation. The per-pair subtraction check must fail; the proof
   reports corrected and mutant wall/CPU per request for every workload/variant.
   Node-source's service wall difference must exceed 1 ms/request to hold visible
   startup impact. Negative corrected samples remain visible.

`proofs.json`, baseline/small-K markdown and JSON, actual mutant sources, compiled
mutant output and build diagnostics are retained in the proof directory. The
proof exits nonzero if either mutant escapes or the positive comparison fails.
Linux execution is verified in this unit; macOS wait4 unit conversion is
implemented but has not been exercised here. No diagnosis/profile pass is run.

Keep run results and evidence on the task, outside the repository. Run with
`--output` pointing to a scratch path, such as `/tmp/handler-results.json`, and
redirect the markdown output there too. Node-source's parse time includes Node's
type stripping, which a bundled TypeScript Worker does not pay.
