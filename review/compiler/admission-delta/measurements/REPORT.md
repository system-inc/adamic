Built explicit diff-only filtering and deterministic shards; every unit retains the complete mandatory diff.
Measurements use implementation in this commit, baseline 2b1be383 and fixes ced32bf9, on the cgroup quota of four CPUs.
Seven bounded real runs exited 0: partial scope, admitted 1, sampled 1, omitted 0, no compiler errors and no disagreements.
Mutants caught: split the diff, repeat the first shard, and label a partial unit pass; restored tests and the call-target guard passed.
Not covered: empty Go package-cache bootstrap, concurrent shard scheduling on one four-CPU box, or a universal runtime bound for larger mandatory deltas.

# Method

Base: 2b1be38362046455e0e5454d8b0e674a8630e95d
Fixes head: ced32bf9b1f410b8be07f078adb78365a20a9f5b
Small synthetic head: 562b2ff017f5394621b684ed27ae56594c448c7f

The fixes comparison intentionally retains the earlier pinned baseline so its revision diff is exactly 21 .a paths. The small snapshot uses the same fixes compiler and other non-.a source as ced32bf9, restoring every .a input except cat.a, every.a and find.a to its base blob. Its diff is exactly those three paths. The synthetic commit was constructed with a private Git index; no compiler files or delivery-tree programs were edited. small.patch, small.commit.txt and run.py reproduce it.

Cold means a newly created, empty ADAMIC_BUILD_CACHE_DIR with zero compiler products. Both compiler products missed. Go package cache, installed SDK, pinned Git objects and npm download cache were retained. Warm reuses both products: both hit. No classification-result or runtime-result caching occurs. Each real command uses --workers 4 --budget 60; all measurements run sequentially. CPU quota was cpu.max = 400000 100000; nproc reports 5 visible CPUs, but all classification uses four workers. Per-command hard limits remain; runner limits every real invocation to 600 s and logs progress.

All results report verdict partial and complete_corpus false because they cover a filter or one shard. Exit zero means the selected unit agrees; it does not mean a full corpus gate passed. All mandatory diff admissions are checked on every unit. Each of the seven runs newly admits array_narrowing/find.a; Node, JavaScript and native all print "found 2 / double 4\n" and exit zero.

# Diff-only cold and warm, seconds

| Phase | 21 paths cold | 21 paths warm | 3 paths cold | 3 paths warm |
| --- | ---: | ---: | ---: | ---: |
| checkout | 1.765 | 4.562 | 4.652 | 4.602 |
| provision | 0.581 | 0.566 | 0.546 | 0.540 |
| base_build | 9.622 | 0.001 | 20.694 | 0.001 |
| head_build | 19.069 | 0.000 | 20.272 | 0.000 |
| manifest_and_verification | 0.093 | 0.091 | 0.039 | 0.027 |
| classification | 0.801 | 0.838 | 0.173 | 0.128 |
| runtime | 0.320 | 0.316 | 0.339 | 0.302 |
| total | 32.255 | 6.378 | 46.720 | 5.605 |
| external wall including cleanup | 32.579 | 6.695 | 47.029 | 5.904 |

# Generated manifest split into three shards, warm seconds

Path SHA-256 selects i/N deterministically. All non-diff generated inputs are assigned exactly once; all 1,031 generated-manifest paths are covered by the shard union. The 21 diff inputs are kept whole in each shard, and overlap with selected generated inputs is classified once. Shards were measured with cached compilers, not with empty compiler-product caches.

| Shard | Generated paths | Unique classified | Checkout | Provision | Verification | Fixed sum | Builds | Classification | Runtime | Tool total | External wall |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1/3 | 352 | 366 | 4.228 | 0.531 | 0.832 | 5.591 | 0.001 | 21.691 | 0.316 | 27.602 | 27.901 |
| 2/3 | 313 | 328 | 4.100 | 0.535 | 0.713 | 5.348 | 0.001 | 13.333 | 0.310 | 18.995 | 19.276 |
| 3/3 | 366 | 380 | 4.525 | 0.517 | 0.861 | 5.903 | 0.000 | 15.986 | 0.310 | 22.204 | 22.531 |

All three measured warm shard units fit 30 s, including cleanup. The slowest unit was shard 1/3. Fixed costs are roughly 5 to 6 s per shard, primarily checkout. Sharding repeats setup and the mandatory diff runtime check, so it trades more total work for smaller independently bounded units. Measurements do not promise these times if all shard units contend for the same four CPUs concurrently.

Compiler products must be prepared outside a strict 30 s unit: both cold diff-only cases exceeded 30 s even though the Go package cache was retained. Warm diff-only units took under 7 s, and three warm generated-corpus shards each stayed under 30 s. No gap compile-observation cache was needed.

# Validation

go test ./cmd/adamic-admission-delta -count=1 -timeout 60s -v passed; TestShardPartitionKeepsDiff took 0.00 s and TestShardCommandScope 0.10 s. The three source mutants each failed its intended test. The restored command tests and go test ./internal/ir -run ^TestCallTargetReaders$ -count=1 -timeout 60s passed. Test logs and mutant patches are under repair/. Lane checks run after the commit before its push.
