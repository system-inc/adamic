# Workers benchmark harness

Run the same self-contained Worker modules under the npm **workerd 1.20261007.1**
binary directly. No wrangler, reverse proxy or development middleware is in the
request path. Node 24 drives HTTP over a keep-alive `node:http` agent. No compiler,
compute Worker or Wasm generator is implemented here.

```sh
npm install --prefix /tmp/workers-runtime workerd@1.20261007.1
node bench/workers/run.mjs \
  --workerd /tmp/workers-runtime/node_modules/.bin/workerd \
  --suite bench/workers/fixtures/suite.json \
  --output /tmp/workers-results.json > /tmp/workers-results.md 2>&1
```

The pin was the newest registry version on October 7, 2026 and installed on
linux-x64. Both the binary's date and its adjacent npm package version are verified
and printed. Supply the npm package binary, rather than a copied standalone binary
whose npm patch version cannot be verified. Dependencies stay outside the repo.

The explicit fallback is `--runner node`. It starts a separate Node HTTP adapter
process importing `worker.mjs` and translating HTTP into `Request` and `fetch`.
Its output says **Node HTTP adapter, NOT workerd**. Its CPU and RSS belong to the
adapter process and include its HTTP and Request/Response machinery. These results
are harness evidence, not a Workers performance claim. Switch runners via the CLI;
the exported harness functions also support focused experiments.

## Suite and flags

A suite is JSON. Variant directories resolve relative to the suite file; each has
`worker.mjs` exporting a Workers fetch handler. A directory's recursive `.mjs`,
`.js` and `.wasm` files are listed in a generated Cap'n Proto config, copied beside
it with their relative module names intact, entry module first. Symlinks are
refused. Other extensions are not bundled. Workers must import only bundled
modules and Workers-supported APIs; no nodejs_compat is enabled. Wasm modules use
Cap'n Proto's `wasm` field. The Node adapter cannot directly import `.wasm` like
workerd does.

```json
{
  "variants": [
    { "name": "twin", "directory": "./twin" },
    { "name": "adamic-js", "directory": "./adamic-js" }
  ],
  "coldPath": "/health",
  "checks": [
    { "method": "GET", "url": "/answer?q=hello" },
    { "method": "POST", "url": "/answer", "body": "hello" }
  ],
  "workloads": [
    { "name": "single", "method": "POST", "url": "/answer", "body": "hello" },
    { "name": "mix", "requests": [
      { "weight": 3, "method": "GET", "url": "/answer?q=hello" },
      { "weight": 1, "method": "POST", "url": "/answer", "body": "hello" }
    ] }
  ]
}
```

URLs are local absolute paths including any query, not external origins. The
HTTP Host defaults to `bench.invalid`, so changing ephemeral ports does not change
the Worker-visible URL between variants. Override it with request headers when
needed; the network connection still goes to the local server. Bodies
are UTF-8 strings. Requests may also specify a `headers` object. Omitted method is
GET, omitted weight is 1. Weights are positive integers. Weighted traffic is
reproducible: request index modulo total weight selects successive weighted
buckets; the schedule resets at each phase. Short counts can leave a partial
cycle. Concurrency lanes claim indices in dispatch order.

Every effective flag, including defaults, is printed in the header and JSON:

| Flag | Default | Meaning |
|---|---|---|
| `--runner` | `workerd` | `workerd` or explicit `node` fallback |
| `--workerd` | `workerd` | npm binary path or name on PATH |
| `--suite` | included fixture suite | suite JSON filename |
| `--output` | `workers-results.json` | raw report filename |
| `--requests` | 500 | measured requests per cell |
| `--warmup` | 100 | discarded requests before every cell |
| `--rounds` | 3 | interleaved measurement rounds |
| `--cold-spawns` | 5 | independent cold spawns per variant |
| `--concurrency` | `1,16` | closed-loop client lane counts |
| `--compatibility-date` | `2026-10-01` | fixed workerd compatibility date |
| `--timeout-ms` | 10000 | startup deadline and per-request timeout |
| `--idle-settle-ms` | 100 | pause after readiness, before idle RSS |
| `--memory-poll-ms` | 10 | peak RSS sampling interval |

The generated config uses one service and an HTTP socket bound to 127.0.0.1 on an
available ephemeral port. There is a small close-to-spawn port reservation race;
a collision is reported as startup failure. Generated modules and configs are
removed after process termination, including on failure.

## Instruments and interpretation

Correctness runs **before any timing**, including cold timing. Every variant gets
a separate process and the complete ordered check list, compared with the first
variant's status, byte body and normalized application headers. Date, connection,
keep-alive, transfer-encoding and content-length are ignored as HTTP transport
metadata. A mismatch fails the run, saves the responses and an error in JSON, and
leaves zero timing samples. The check list defines the extent of equivalence; it
is not an exhaustive oracle for requests absent from that list.

Each round iterates workloads, then concurrency, then all variants. Variant order
rotates by one position per round. Each cell uses a fresh server process, probes
only the cheap cold path for readiness, reads idle memory, discards warmup, and
measures exactly the fixed request count. Warmup uses the same workload and
concurrency. No lane sends a request until its previous response has been fully
consumed. Latency is Node's monotonic `performance.now()` from HTTP dispatch to
complete response body, including any keep-alive agent queue time. Percentiles
use nearest rank; mean uses all requests. This is closed-loop latency, with no
correction for coordinated omission. JSON keeps each request's index, latency and
status, all per-cell counters and summaries, PIDs and module lists.

Linux CPU uses **`/proc/<pid>/stat` fields 14 and 15**, user plus system ticks,
with `getconf CLK_TCK` for resolution. Parsing handles spaces and parentheses in
the process name. Counters immediately bracket the measured phase and exclude
warmup. CPU delta divided by measured request count is CPU milliseconds per
request. This is total process CPU across workerd threads, not wall time or client
CPU. After measurement, with the same process idle and the HTTP agent still open,
we sample CPU over a delay matching the measured wall time. JSON records the
actual baseline wall duration and CPU delta; the table normalizes idle drift to
that same duration and request count. Drift is shown, never silently subtracted.
A zero below the counter resolution means unresolvable, not zero cost.

Linux memory uses **`/proc/<pid>/status` VmRSS** after readiness and settling, and
**VmHWM** for peak resident bytes. VmHWM is a lifetime high-water mark, including
startup and warmup; fresh processes prevent previous cells from contaminating it.
Polling also reads it during measurement, and a final read captures brief peaks.
This is process RSS, not isolate heap usage.

macOS CPU uses **`ps -o cputime=,rss= -p PID`**, parsing optional day prefixes.
CPU time has one-second resolution; use sufficiently large request counts.
macOS memory uses idle `ps` RSS and the maximum RSS sampled at the configured
interval, including idle and measured phase endpoints. It can miss brief peaks
and is not equivalent to Linux's kernel VmHWM. The header names this distinction.
macOS instruments are implemented but have not been exercised in the Linux
container. Frequent `ps` subprocesses also impose observer overhead.

Cold start uses separate interleaved rounds of fresh spawns. Timing begins
immediately before spawn, after config/module files are prepared, and ends at the
first complete 200 response on the named cheap path. Module parsing and
instantiation happen inside this interval. Polling is at least 2ms apart; socket
setup, scheduler noise and HTTP handling are included. This is local process
spawn-to-response, not Cloudflare's deployed isolate cold-start metric.

The markdown table reports **best and median of N independently for each metric**
across rounds, not a fictional single fastest row. Cold starts have their own
best and median table. `uptime` with load averages is recorded before and after;
machine, runtime, instruments, pin and effective flags are in the header. Preserve
the full raw JSON when comparing variants. Small differences on a loaded machine,
or CPU deltas near the tick/drift floor, do not establish an optimization.

## Proofs that can fail

```sh
node bench/workers/prove.mjs workerd \
  /tmp/workers-runtime/node_modules/.bin/workerd /tmp/workers-proofs-workerd \
  > /tmp/workers-proofs-workerd.log 2>&1
node bench/workers/prove.mjs node workerd /tmp/workers-proofs-node \
  > /tmp/workers-proofs-node.log 2>&1
```

The proof runner saves every baseline and mutant raw report, source mutant and
`proofs.json` in the destination. It exits nonzero on an undetected mutant or a
failed positive check. It also holds measured counts and rotating order, stable Worker-visible URLs,
UTF-8 request bodies and observed weighted dispatch. Under workerd it verifies
an actual Wasm module against Node WebAssembly API. See [REPORT.md](REPORT.md)
for the commands and observations from this container.

The fixtures are hand-written JavaScript for testing this harness only: fast,
CPU (4,000,000 dependent integer operations per non-health request) and allocate
(touch and retain up to 24 one-MiB buffers). All return the same application
responses. The operation count is known; the time is measured, not assumed.

1. CPU per request must exceed fast by the greater of 0.5ms, five CPU ticks per
   measured count, and three times measured idle drift, at both 1 and 16 lanes.
   A source mutant reads CPU from an unrelated idle child PID; exactly that
   separation check must fail. Its HTTP output and latency still work.
2. A temporary variant changes only `/echo` to return `wrong`. Correctness must
   refuse it with zero measured and zero cold samples. The mutant exists only
   under the proof output directory.
3. The fast fixture's dedicated `/warm-probe` path burns 12,000,000 dependent
   integer operations on its first eight calls. Neither readiness nor checks
   touch that path. A source mutant drops the actual warmup phase while keeping
   the declared flags. In first-round request samples, its first-eight median
   must exceed both the warmed first-eight median and its own later median by
   5ms. JSON also shows actual warmup count 0 instead of 16. This deliberately
   reproducible first-use penalty proves warmup removal is visible without
   relying on unpredictable JIT behavior. It does not estimate production JIT
   warmup duration. The proof's small CPU count is suited to Linux tick resolution;
   macOS needs a larger-count experiment.

## Day-6 compute Worker table

`day6.mjs` runs the TypeScript twin bundled by Wrangler deploy, the generated
Adamic JS Worker, and the explicit primes/summarize split Worker under workerd.
Use a separate local measurement checkout with `codex/workers-compute-a` and
`codex/workers-two-backends` merged onto platforms; retain its exact merge SHA
and any merge resolution. Build before timing, and run nothing else while timing.

Place the three builds under a scratch directory as `variants/twin`,
`variants/adamic-js`, and `variants/adamic-js-wasm`, each with `worker.mjs` as entry.
Bundle the twin with `npx wrangler@4 deploy --dry-run --outdir <scratch>` and a
minimal config naming the original twin, compatibility date 2026-10-01 and no
nodejs_compat. Copy its emitted JS to the twin entry; do not time type stripping.
Build the others with `adamic worker workers/compute/handler.a --out <dir>` and
the same command adding `--wasm primes,summarize`. Save
`adamic-split workers/compute/handler.a` output as scratch `split.txt`; save build
commands, effective flags and versions in scratch `build-metadata.json`, an object
with `commands`, `flags` and `versions` fields. These are included in the header.
The runner preserves the full splitter output even when explicit selection
differs from its suggestions.

```sh
node bench/workers/day6.mjs --repo /path/to/measurement-checkout \
  --directory /tmp/workers-day6 --workerd /path/to/npm/workerd \
  --rounds 5 --cold-spawns 10 --output /tmp/workers-day6/run.json \
  > /tmp/workers-day6/run.log 2>&1
```

`--mode preflight` performs only correctness; `--mode pilot` also measures four
requests on the three largest inputs with a two-request warmup. Default mode
`run` requires at least five rounds and ten cold spawns. Other flags are exactly
those in the example; mode, rounds and cold spawns have defaults. All results,
logs, generated bodies, suite, gzip JSON and markdown stay in scratch and belong
on the task, never in the repository.

Before timing, the unchanged `workers/replay.mjs --url` compares all 600 recorded
requests through each workerd process; anything other than 600 matched refuses
the run. Additional byte comparisons cover every standalone size. Synthetic
inputs use xorshift32 seed 0x6a09e667: stats are deterministic finite decimals;
orders are valid deterministic items; text uses ten ASCII tokens, exactly 1,000,
100,000 or 1,000,000 text bytes before JSON framing. These are decimal KB/MB,
matching the handler's one-million-unit limit. The mix dispatches the entire
600-request corpus in order, including invalid inputs, once per phase, with a
complete-corpus warmup. Concurrency lanes claim indices in dispatch order.

Standalone cells discard 32 warmup requests. Measured counts are 1,000 for health,
primes 10,000, stats 100, both quotes and text 1 KB; 256 for stats 10,000 and text
100 KB; 128 for the larger primes, stats 100,000 and text 1 MB. Counts are fixed
across variants and rounds and named in the output. With 128 samples, nearest-rank
p99 is the second-largest observation; it is a noisy tail estimate.

The general suite accepts optional `requestCount` and `warmupCount` per workload;
these override the command defaults and are validated. Optional `expectedStatus`
on a request refuses an incorrect timed response, including mixed error statuses.
Round load averages are captured before and after each round and printed with
the existing whole-run loads. JS size sums served modules, each compressed with
gzip level 9; Wasm is reported separately. Config, ABI JSON and source maps are
excluded because workerd does not serve them as modules. Cold timing stays
spawn-to-first-200 on /health and includes module parse and Wasm initialization.

```sh
node bench/workers/day6-proof.mjs /path/to/npm/workerd /tmp/day6-harness-proofs \
  > /tmp/day6-harness-proofs.log 2>&1
```

This proves workload counts, awaited callbacks, round loads, deterministic input
sizes and status refusal. Source mutants drop the request-count override,
after-round load and expected-status check; each must fail its intended check.
Run the original `prove.mjs` too to retain CPU, wrong-PID, correctness and warmup
proofs. Nothing in the day-6 unit changes or optimizes the compiler.

A reused keep-alive socket that fails with ECONNRESET or EPIPE before any response
bytes arrive is retried once on a fresh socket. The original latency timer and
timeout cover both attempts. JSON and the per-cell table count measured and
warmup retries separately. Cells with more than 1% retries among measured
requests or among all configured requests, and requests that fail twice, are
marked failed and excluded from timing aggregates. No extra settling pause is
added. Run `node bench/workers/retry-proof.mjs /tmp/workers-retry-proofs` with
output redirected to a scratch log to prove counting, fresh sockets, response
byte refusal, the original timer, the threshold and the corresponding mutants.
