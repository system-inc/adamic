# Direct primes typed-array comparison

This is the harness behind the Workers typed-array old/new table. `old.a` is
`primes` extracted unchanged from `workers/compute/handler.a` at
area/platforms `e1ae0e25810483fc7b6217a1ae57bb4d4b75566d`. `new.a` changes only
its push-grown number array to `new Uint8Array(limit + 1)`, as in `4caf8433`.
The original compiler measurement checkout was an uncommitted scratch merge
of that platform base with area/runtime
`7c4dae857cc13b334e699880a5af23b2b7a83c15`. Its merge resolutions were reported
in the typed-array unit; neither this harness nor the handler branch merges
runtime. Use a compiler with typed-array lowering and the Workers backend,
such as the compiler's integrated checkout. Both variants use the same binary,
with its default release build flags; this is a storage comparison, not a
comparison between two compiler revisions.

From the repository root, with Node 24, Python 3 and the configured native
clang on PATH:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/typed-primes-setup.log 2>&1
source /workspace/adamic-tools/env.sh # or setup's printed path
# In the integrated compiler checkout:
go build -o /tmp/typed-primes-adamic ./cmd/adamic > /tmp/typed-primes-build.log 2>&1
node bench/workers/handler/typed-primes/run.mjs /tmp/typed-primes-adamic > /tmp/typed-primes.log 2>&1
```

The runner retains build artifacts in a fresh temporary directory and prints
its path on stderr. No measurement files are written to the repository.
It uses `adamic worker` to expose the compiled JavaScript `primes` function;
the generated HTTP entry is unused. Native drivers call the same extracted
functions directly. There is no HTTP, JSON marshalling or Wasm in the timings.

For each backend and limit (1,000,000 and 5,000,000), five rounds alternate
old/new execution order. Each sample runs in a fresh process and calls primes
20 times, consuming counts and checking their known totals. JavaScript does
two untimed warmup calls before each 20-call timed batch. Native does no
warmup. Builds finish before measurements; processes never overlap.

The JSON output has one row per backend/limit:

- `engine`: `js` means Adamic-generated JavaScript running in Node; `native`
  means the compiled native driver, launched by Node through a Python helper.
- `oldBestMs`, `newBestMs`: smallest batch duration divided by 20, in
  milliseconds per call. JavaScript times only its loop with performance.now.
  Native times process launch through exit with Python's monotonic clock,
  including startup, final checksum printing and shutdown, amortized over 20.
- `oldPeakRSSKiB`, `newPeakRSSKiB`: largest whole-process peak across the five
  samples, independent of which sample had the best time. JavaScript uses
  process.resourceUsage().maxRSS; native uses RUSAGE_CHILDREN in a fresh Python
  helper for each native process. Linux reports KiB; the helper converts macOS
  bytes to KiB. Native RSS may include the Python launcher's inherited RSS floor;
  this intentionally preserves the original measurement method.
- `samples`: all five raw samples per variant, including milliseconds per call,
  peak RSS in KiB, and the consumed prime count.

The original measurement used Linux, Node 24.19.0, Go 1.27.1, clang 20.1.8,
five visible processors and a four-CPU quota. Shared-machine timings vary;
retain the compiler's revision and dirty status alongside each run. No recorded
results are checked in.
