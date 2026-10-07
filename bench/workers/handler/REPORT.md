Built a repeatable native/Wasm/Adamic-JS/Node-source handler benchmark, without HTTP or profiling.
Delivery contains only bench/workers/handler; dependency merges remain in a local measurement checkout.
K=1000, N=3: four checksums agree; 24 paired request samples and 18 startup phase samples retained.
Both source mutants were caught: one equal-length response change, and removed wall/CPU startup subtraction.
Linux validated; macOS is implemented but untested, and separate native loader/init phases are unavailable.

## Measurement provenance

Delivery was rebased onto platforms `be2e9c1615c5cc74a27f07b1f57cde1443b7c65f`.
The performance observations below were taken against clean local compiler merge
`de6e1de0b4eb3aea066df66606a46de0bd49a5ce`, based on platforms `c3236500`, with
service `f4ec96cf0a43c0b03b8f99dfcfd3bacb28e80569` and Workers host
`c7196b920db29c40e30210744a20a6c33a29d27f`. The service remains at its original path.
Only the release configuration from size reference
`aada8365085b8fea9f6cfd79230badbaa0dfce20` is used; that branch was not merged.

Linux 6.18.44, AMD EPYC 9V74, nproc 5, cgroup quota 4 CPU equivalents;
Node 24.19.0; Go 1.27.1; clang 20.1.8; WASI SDK 27 clang 20.1.8-wasi-sdk;
Binaryen 126. Native uses clang -O2. Wasm entry and runtime use clang -Oz,
llvm-strip, then wasm-opt -Oz. Brotli is Node zlib quality 11, generic mode.
All effective flags, versions, sources/binaries/corpora hashes and raw child
rusage appear in [run.json](evidence/run.json). Wrapped release commands are in
[release-commands.jsonl](evidence/release-commands.jsonl).

Setup was run as `bash cloud/setup.sh` and then `bash cloud/setup.sh --wasi-sdk`,
with file logs and `source /workspace/adamic-tools/env.sh`.
Normal timing lines: Go 0s, clang 0s, Node 0s, submodules 15s,
build-cache warm 457s, done 457s. WASI setup: Go/clang/Node 0s,
SDK 8s, submodules 8s, build-cache warm 380s, done 380s; nproc 5.
The two setup/build jobs overlapped; benchmark processes did not overlap in the
retained standalone run or standalone proofs. An earlier run overlapped a proof
build and was discarded.

## Commands and observations

```sh
export PATH=/tmp/handler-tools/binaryen-version_126/bin:$PATH
node bench/workers/handler/run.mjs --repo /tmp/adamic-handler-measure \
  --iterations 1000 --rounds 3 --output /tmp/handler-release.json \
  --artifacts /tmp/handler-release-build > /tmp/handler-release.md 2> /tmp/handler-release.stderr
node bench/workers/handler/prove.mjs /tmp/adamic-handler-measure \
  /tmp/handler-release-proofs-alone > /tmp/handler-release-proofs-alone.log 2>&1
```

Both exited 0. The retained [run.md](evidence/run.md) and JSON are a subsequent
standalone repeat through exported `runPrepared`, reusing those exact built
artifacts and recording output `/tmp/handler-release-final.json`. No build,
compression or other benchmark was concurrent with that repeat.
Load before was 0.79/1.14/2.61 and after 1.06/1.17/2.56.

Median startup-subtracted timings, ms/request:

| workload | native wall / CPU | wasm wall / CPU | adamic-js wall / CPU | node-source wall / CPU |
|---|---:|---:|---:|---:|
| six-route service | 0.020284 / 0.020280 | 0.030743 / 0.043316 | 0.014072 / 0.019712 | 0.005249 / 0.011352 |
| sieve | 0.829349 / 0.826971 | 1.379223 / 1.395161 | 0.601258 / 0.612685 | 0.593166 / 0.601920 |

There are 108,000 measured requests total. Startup subtraction uses independent
fresh processes, so its noise is appreciable even at K=1000 for the cheap service.
These observations do not establish a speedup claim. Full best/median raw and
subtracted wall/CPU, K=0 and peak RSS tables are retained.

Service artifact sizes and median startup:

| variant | raw / Brotli bytes | compile ms | instantiate + init ms | parse + module load ms | K=0 process ms |
|---|---:|---:|---:|---:|---:|
| native | 398944 / 131297 | n/a | n/a | n/a | 1.278249 |
| wasm | 86846 / 32293 | 0.944966 | 0.323488 | n/a | 74.417964 |
| adamic-js | 31734 / 5649 | n/a | n/a | 3.762056 | 43.259088 |
| node-source | 10455 / 3026 | n/a | n/a | 55.366256 | 91.099988 |

The size definitions and phase boundaries are in the README and table header.
The observed Wasm artifact differs from the sibling's 33,488-byte size result;
these are different recorded compiler checkouts. No cause is inferred here.
The production Workers host initializes lazily: K=0 compiles and creates the
host, but first-call instance/init remains in measured K>0. A separate phase
probe exposes compile and instance/init directly through the production WASI shim.

## Failure proofs and limits

[proofs.json](evidence/proofs.json) reports the actual source mutants and results:

- Equal-length first response mutation, compiled with `adamic js`, refused at
  `service/adamic-js preflight`, with zero timed samples. Count and units remained
  `6:388`; expected digest `478017052:139489648`, actual `3086986072:1823433884`.
  [Changed-response evidence](evidence/changed-response.json) retains the failed child.
- Removed wall and CPU startup subtraction from the reducer source; checks of the
  corrected per-pair equations fail under that mutant. Both reducers consume the
  same K=1/K=0 samples, so their difference is precisely startup/request.
  Median service wall ms/request (correct / mutant): native 0.077444 / 0.329140;
  Wasm -0.122957 / 19.076659; Adamic-JS 0.509960 / 8.155858;
  Node-source 1.592774 / 19.875166. Raw CPU and all pairs are retained in
  [small-K.json](evidence/small-K.json) and [small-K.md](evidence/small-K.md).
- Positive checks cover all four columns, fixed counts, interleaving, source sieve
  pins 168/1229/9592, and wait4 exec without arguments and timeout process killing.

JavaScript syntax checks ran on every harness .mjs file. The uncached filtered
compiler/oracle gate was:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./cmd/adamic ./internal/oracle \
  -run 'TestNativeAgreesWithNode/internal/oracle/testdata/modules/main.a$|TestRequest' \
  -count=1 -timeout 30m > /tmp/handler-filtered-oracle.log 2>&1
```

It passed: cmd/adamic 6.145s, internal/oracle 9.499s. This was a filtered gate,
not the full compiler suite. The corpus is a small balanced six-route sample,
not the sibling's seeded 100,000-request distribution. Checksums can collide;
they are not a collision-free response transcript. Hashing/driver/JIT costs are
included. No workerd/HTTP measurements or profiling were added in this unit.

## Landing compatibility check

After rebasing the delivery branch onto current platforms `f573c146`, the local
measurement checkout merged that base at `2bf5040314daa5fc058f5bf4fc0db3fe70ff4981`.
`node bench/workers/handler/prove.mjs /tmp/adamic-handler-measure
/tmp/handler-landing-proofs > /tmp/handler-landing-proofs.log 2>&1` exited 0
again: four columns agree, both source mutants are caught, and helper/pins/counts
pass. [Landing proofs](evidence/landing-proofs.json) and the complete
[landing baseline](evidence/landing-baseline.json) preserve the new checkout
provenance and all its raw samples. These checks do not replace the named
performance observation above.

The same uncached filtered gate passed after the update: cmd/adamic 5.756s,
internal/oracle 9.906s, captured in `/tmp/handler-landing-gate.log`.

A final platforms update to `be2e9c1615c5cc74a27f07b1f57cde1443b7c65f`
added only tests and notes. Delivery was rebased again. Local measurement merge
`38b277629b24a392e1cbafdf8280bf8806f47f34` passed the same uncached filtered gate:
cmd/adamic 0.852s, internal/oracle 0.788s, log `/tmp/handler-final-gate.log`.
The production compiler and host files did not change in this final update.
