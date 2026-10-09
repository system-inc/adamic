Reproduced sanitized chunk 1 dying by SIGKILL, signal 9, under the full harness's memory pressure.
Measured compiler/views-v3 14690251 and main beb1be1b; no production change is proposed.
The exact slice and both five-way direct runs passed; the full V3 harness was OOM-killed.
No new check or mutant is applicable; all measurement overlays stayed outside the repository.
The original pool's kernel log was unavailable; this box also has inherited tmpfs pressure.

## Observations

The captured corpus has exactly 3,473 cases and passes the repository and external corpus pins. The five ranges are [0:694], [694:1389], [1389:2083], [2083:2778], [2778:3473]. Chunk 1 has 695 cases, from cohere/docs/data/examples/no-with.json through stage1/cohere/lint/rules/no-empty-pattern/mutant.json. Its protocol input SHA256 is 97d7389561b194530272213fa0e9f93a7170ac549edeca29073341cc8322304a. The isolated input, direct parallel input and full harness's chunk-1 input are byte-identical. The ordered corpus names, sizes and content hashes are retained in evidence/signal-corpus-manifest.txt.gz.

Every run used ASAN_OPTIONS=detect_leaks=0, exactly the failing ASan/UBSan phase's setting. A wait4 recorder captures actual signal and maximum RSS; no per-unit deadline was added. Both isolated runs exited zero with empty stderr and identical stdout hashes. The main binary was built on beb1be1b during the preceding measurement; its build logs and worktree revision are retained. V3's sanitized binary was freshly built at 14690251. Its emitted C is byte-identical to the earlier fixed JSON C. The JSON parser/printer source is unchanged between the two revisions.

| Exact slice alone | Peak RSS KiB | Seconds | Signal | Exit |
| --- | ---: | ---: | ---: | ---: |
| V3 | 3,600,448 | 103.309264 | 0 | 0 |
| Main | 3,599,656 | 101.388425 | 0 | 0 |

Both five-way direct runs completed with empty stderr and identical outputs for every corresponding chunk. No cgroup OOM counter increased during these runs.

| Chunk | V3 peak RSS KiB | Main peak RSS KiB | V3 signal/exit | Main signal/exit |
| --- | ---: | ---: | --- | --- |
| 0 | 566,132 | 564,836 | 0/0 | 0/0 |
| 1 | 3,593,776 | 3,597,972 | 0/0 | 0/0 |
| 2 | 582,760 | 600,024 | 0/0 | 0/0 |
| 3 | 2,019,160 | 2,062,032 | 0/0 | 0/0 |
| 4 | 2,087,684 | 2,092,716 | 0/0 | 0/0 |

Main's chunks are not materially smaller. Only chunk 1 reaches about 3.4 GiB; the five individual peaks sum to 8.44 GiB on V3 and 8.50 GiB on main. These sums are upper bounds assembled from different times, not measurements of simultaneous RSS. The evidence does not support five children each consuming 3.5 GB, or a V3-specific native allocation explosion. The metadata helper records the same facts without allocating. V3 adds array metadata and a sparse pointer to the runtime layout, but the measured difference on identical inputs is small, with main slightly larger in the parallel run.

## Full-harness reproduction

Running only TestPortMatchesGoCohere with a temporary childguard measurement overlay reached the same "native corpus: 5 contiguous chunks, 5 cores, 3473 cases" line, then hit the 17,179,869,184-byte cgroup limit. Sanitized chunk 1 died by signal 9 after 51.519914 seconds, peak RSS 2,947,504 KiB. That peak is censored by the kill; it is not the amount needed to finish the chunk. The Go test parent also died by signal 9, at 237.653277 seconds, peak RSS 3,212,056 KiB. The cgroup counters changed from oom=19, oom_kill=4 to oom=25, oom_kill=6. This records two kernel OOM kills alongside the two SIGKILL exits.

Immediately before the kills, the sampled parent RSS was 2,959,084 KiB and chunk 1 RSS was 2,900,524 KiB, with chunks 3 and 4 still live. The one-second RSS timeline is retained. This box started with roughly 6.8 GB of inherited tmpfs memory; the test's temporary inputs raised shmem to about 7.94 GB. File cache also contributes to memory.current, so merely reaching the limit during a successful direct run is not by itself an OOM observation. The counter increase and wait4 signals in the full run are the distinguishing evidence.

| Full V3 harness child | Peak RSS KiB | Signal | Exit |
| --- | ---: | ---: | ---: |
| Chunk 0 | 564,364 | 0 | 0 |
| Chunk 1 | 2,947,504 | 9 | -1 |
| Chunk 2 | 586,832 | 0 | 0 |
| Chunk 3 | 2,056,128 | 0 | 70 |
| Chunk 4 | 1,937,504 | 0 | 70 |

Chunks 3 and 4 exited 70 after the parent was killed and their output pipe readers disappeared. Their stderr cannot be recovered from the dead Go parent's buffers. The runtime ignores SIGPIPE and reports failed writes, which is consistent with these later exits; the exact messages are not claimed. Their independent direct runs passed with empty stderr. No sanitizer first frame was reported in the retained full-test log, and the observed fatal signal for chunk 1 is SIGKILL, not SIGSEGV or SIGABRT.

## Inference and limits

The reproduced chunk-1 death is an OOM kill. It explains how Go can report exit -1 without an assertion or sanitizer report. The direct native port has essentially the same memory requirement on V3 and main. Retained parent buffers, temporary-file backing, sanitizer workers and other package activity must fit together; CPU count alone is not a memory budget.

The pool's original signal was not recoverable from its quoted Go message alone. This reproduces the same chunk's SIGKILL with OOM evidence on this box, but does not establish the pool's exact background memory or explain why its main run had more headroom. Main's pool corpus and memory timeline would be needed to attribute that difference. The full main harness was not rerun in this follow-up; main was measured on the exact V3 slice and all five identical input chunks.

No crashing case or sanitizer violation was found to reduce, so no compiler fix, new fixture or mutant is claimed. A later harness change should bound simultaneous sanitized children by memory while preserving every case and comparison; the full parent also needs memory headroom. No views check should be removed to address this resource failure.

## Commands and evidence

```sh
source /workspace/adamic-tools/env.sh
GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=1 go test -overlay=/tmp/v3-json-measure/capture-overlay.json ./stage1/cohere/json -run '^TestCaptureSignalCorpus$' -count=1 -v
python3 /tmp/v3-json-measure/followup.py
GOMAXPROCS=4 go test -c -overlay=/tmp/v3-json-measure/v3-overlay.json -o /tmp/v3-json-measure/followup-v3.test ./stage1/cohere/json
python3 /tmp/v3-json-measure/measure.py followup-v3 /workspace/adamic
python3 /tmp/v3-json-measure/followup-memory.py
```

The temporary capture leaf passed in 19.12 seconds; it is not committed as a Go test. The capture addition is preserved as signal-capture.patch, and there is no compilable Go under this unit's review directory. All commands' outputs were redirected to retained logs. The reproduction scripts record the complete native commands, output hashes, signals, peaks and OOM counters. No whole package suite, new language file or full gate was run. Counts.md has zero changed rows. The installed toolchain, five visible cores, four-core quota and 16 GiB limit are unchanged from the preceding reports.

Only evidence is committed to compiler/views-v3-json-memory. No compiler/views-v3 update is made: the other worker's current-main merge is left to its owner. Integration lane checks are run on the committed evidence tip before the own-branch push.
