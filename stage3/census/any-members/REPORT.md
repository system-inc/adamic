# Paired any-members census (tmdhz1d): stopped

**Any-member findings and first stops: unknown. This run cannot establish 0.**

The head findings list is unavailable. Head completed 3,865 of 4,056 timed attempts, but its process and driver were found defunct before the census flushed any file records. Only its measurement header survived. No head member location, owner unit, full reason, or first-stop status can be recovered from unit timers.

## Regressions

**Unknown:** previously finding-free units or sites that stop on head, new panics, and units whose first lowering stop changed cannot be assessed. No empty regression list or zero-panic claim is made for head.

Main stage3/census/pair was attempted on these saved streams and refused them with `unit identities differ`. Its traceback is in `evidence/pair-error.log`. `delta.json` uses null for unavailable answers.

## Pins and box

Base: `origin/main` at `b0efce6dce1cd7445ed8a570aecac387a4a7d684`. Head: `optimizer/any-members-3` at `66fff4ecc05c3bfb2258cf4bb2a7b0a2b762c59b`, one commit directly on base.

TypeScript v6.0.3, upstream commit `050880ce59e30b356b686bd3144efe24f875ebc8`.

Box: cloud host `3cbd7fe9a660`; 5 available processors; CPU quota 400000/100000 us (4 cores); cgroup memory limit 17179869184 bytes (16 GiB). Host memory is `18440100 kB`. Go 1.27.1 linux/amd64; Node v24.19.0.

## Unique finding kinds

Meter definition: unique `(kind, where, reason, text)` across attempts. The head header is not a finding-free census.

| Kind | Base, complete | Head |
| --- | ---: | --- |
| NotYet | 620 | unknown |
| Refused | 4708 | unknown |
| panic | 0 | unknown |
| error | 0 | unknown |

## Unit lowering time on this box

One sample per compiler on this box. Base is complete; head is interrupted and partial. These sums are not comparable full-run totals. Timers include unit registration, synchronous dependency lowering and recovery; they exclude checker work, file scans and other shared setup. The active interrupted unit has no completed timer.

| Sum | Base, 4,056 completed attempts | Head, 3,865 completed attempts only |
| --- | ---: | ---: |
| wall_seconds | 5902.087133 | 5175.161776 |
| cpu_seconds | 6887.892272 | 6002.433103 |

## Method and stopping evidence

Read `typescript/kind-predicate-census` (`33dca393`) REPORT.md and compare.py, and the meter launcher and README. Built one adapted tree with b0efce6d stage3/apply.sh, its lockfile-pinned API packages and the upstream pin above. Both census tools were separately built from their exact compiler commits with recorded cohere/checker pins, main make_overlay.py, gofmt, and `go build -buildvcs=false -overlay=... ./stage3/census/latent/tool`. No stage3 files differ between the compiler pins.

The base build and run finished before the head build and run. Both used the same source bytes and exact invocation:

```sh
LATENT_TIMES=times.jsonl LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 latent-census ADAPTED/src/compiler latent.jsonl
```

No full mode, entry mode, backend, replay, subset, or replacement census was run. Base started at 16:43:29 UTC and finished at 18:49:21 UTC (7,552 seconds), exiting 0 with all 79 compiler files and 4,056 timed attempts. Its attempted-unit identities independently match the stock TypeScript top-level catalog.

Head started at 18:51:50 UTC. Its last completed timer was written at 20:45:21.924659 UTC for `src/compiler/utilitiesPublic.ts:1748:1`. The interrupted current unit is not identified by that previous-unit timer.

The tool wait was then reported as aborted and the environment reannounced as starting. On reconnection, both head PID 9441 and driver PID 800 were defunct (state Z, parent PID 1). No head driver exit code or end timestamp was saved. The termination cause is not established. The post-interruption memory observation has `max=2450`, `oom=0`, and `oom_kill=0`; this is not evidence of an OOM kill. The process snapshots, timestamps and counters are retained.

The findings files stayed at their headers while timers advanced. This revision deliberately buffers every file record in `Latent` until all callers finish, to finalize deferred-generic instantiation witnesses (`stage3/census/latent/lower.go.txt:213`, the Latent wrapper). Base emitted all its records only at completion. This was observable progress, not a live timer stall; no mode was switched. After interruption the head was already dead, so a goroutine dump could not be obtained. The logs contain no crash diagnosis.

Stopping here preserves the requested one launch per compiler. A complete paired answer requires a new complete head measurement. No zero result is inferred from the lost buffer.

Both headers agree and report first-error mode with zero checker diagnostics (`checker_rejected=false`). The tool retains its fixed measurement label verbatim in the streams. Ordinary Load/Lower output guards were enabled. Compiler and dependency trees are clean, and all 746 adapted source hashes still match the pre-run manifest.

## Evidence and validation

The compressed streams preserve every saved record, with the shared absolute adapted prefix normalized to relative paths as in the reference method. Full and partial timer streams, logs, source hashes, stock unit catalog, compiler/dependency pins, binary hashes/build metadata and process snapshots are under `evidence/`. `compare.py` uses main stage3/census/pair and refuses to manufacture a comparison when either stream is incomplete. `validate.py` checks the saved evidence and explicit unknown answers; see `validation.txt`.
