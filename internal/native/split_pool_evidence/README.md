# Shared clang pool verification

All native build paths share a process-wide semaphore sized to Go's
`runtime.NumCPU()`. Runtime compilation, generated-unit preprocessing and
compilation, version probes, and clang linking acquire it around command execution.
Per-build worker limits retain their existing explicit-jobs/environment precedence.

The baseline overlay restores every changed production file from `c9551640`.
The unused new pool file remains in the baseline package but no baseline command
calls it. The pooled variant uses the working sources. Both use identical input,
flags, toolchain, and CPU quota.

Source the env file printed by `bash cloud/setup.sh --wasi-sdk`, then run:

```bash
python3 internal/native/split_pool_evidence/measure.py
python3 internal/native/split_pool_evidence/mutant.py
```

The measurement script builds/warm-links Go helpers before timing. Loading,
lowering, and C emission precede standalone probe timing. Probe time includes
cold sanitized runtime compilation plus generated C compilation/linking. Each
sample has a fresh runtime cache; `ADAMIC_GATE_UNCACHED=1` bypasses oracle results
and generated-unit objects. Oracle invocations run the complete package with
`-count=1 -timeout=30m`; `-parallel` retains Go's default. No concurrent verification
workloads run during timing. Machine and per-sample load data are recorded.

The eight-build test explicitly requests CPU-sized per-build workers, independent
of the eventual default under `go test`. Its counter observes commands inside the
pool, including preprocessing and linking. The mutant removes token acquisition
and release while keeping the command counter. It must fail the concurrency
assertion, rather than merely failing compilation or program execution.

For oracle measurements, split on uses the branch's automatic default with
`ADAMIC_NATIVE_SPLIT` unset. The explicit `1` override also opts handwritten C
harnesses into the emitter-only splitter. An initial forced-override run and
partial pooled run are retained as `forced-split-*` diagnostics and excluded
from `oracle.jsonl` and the timing table. The split-off sample and standalone
probe samples were valid and retained.

## Results (best of three, seconds)

Largest available box: 5 CPUs visible to Go/affinity; cgroup quota 4 CPUs
(`cpu.max = 400000 100000`). Go 1.27.1, clang 20.1.8, Node v24.19.0.

| Mode | Markdownblocks sanitized build | Complete oracle test step |
| --- | ---: | ---: |
| Split off | 7.523 | 891.302 |
| c9551640, split on | 2.649 | 1354.389 |
| Shared pool, split on, CPU-sized workers | 3.228 | 1342.339 |

The pooled standalone cell uses the final implementation's three-run confirmation.
The earlier pool-only standalone measurement was 2.594 seconds and is retained in
`probe-pooled.jsonl`; the final samples are in `probe-final-pooled.jsonl`.

Best oracle samples' load averages (1/5/15 minutes), before → after:

| Mode | Before | After |
| --- | --- | --- |
| off | 4.28 / 8.18 / 4.64 | 6.52 / 6.79 / 5.93 |
| unpooled | 5.27 / 8.62 / 9.46 | 7.85 / 10.89 / 11.10 |
| pooled | 4.21 / 9.64 / 10.50 | 7.22 / 8.88 / 9.53 |

All nine valid oracle runs passed. Pooling with CPU-sized per-build workers was
50.6% slower than split off on the oracle step. Therefore the final policy defaults
to **one worker under `testing.Testing()`**, retaining CPU-sized workers in ordinary
programs and retaining the process-wide pool in both cases. Explicit `Options.Jobs`
and `ADAMIC_NATIVE_JOBS` still override the default. The reproduction script sets
CPU-sized jobs for the oracle comparison to reproduce the policy before this
fallback; the timing table does not claim to measure the final one-worker policy.

## Final local verification

- Focused native fixtures passed: concurrent builds, structural prototypes, job
  precedence/default, shared frame state and parallel runtime archive state/order.
- The eight-build test observed **5** peak clangs on **5** CPUs; the semaphore-bypass
  mutant observed **35** and failed the bound. The mutant runner passed by rejecting it.
- Existing `TestSplitTSGoAgrees` ran with a built checker archive, without skipping.
  Whole and split checker output matched byte-for-byte with unit caches on and off
  (86 bytes in both cases). All related split/unit fixtures passed.
- `go vet ./internal/native` passed.

See `final-focused.log`, `final-mutant.log`, and `final-parity.log`.

The complete final-policy oracle validation passed in **1425.921 seconds** with
`ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout=30m` and
no native-job override. This is one validation run, not a best-of-three
comparison; it does not establish a speedup over split off. The final ordinary
build probe (after importing `testing`) retained CPU-sized workers: best of three
**3.228 seconds**; samples and loads are in
`probe-final-pooled.jsonl`.
