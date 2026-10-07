Built: semantic message-ID token measurements and a short author command guide.\
Commits: implementation and measurement 8483db57037839c37ccc99cc8604cee3a2bdd0e4.\
Commands and outputs: nine semantic samples rebuilt objects; Node, emitted JavaScript and native agreed and were rejected by unchanged Go.\
Mutants: all nine message-ID changes caught by Go; a native-only emitted-C output mutant caught by semantic Node/native parity.\
Not covered: passing behavior-changing fixes, full repository gate or fifty-rule fleet; these samples stop at expected comparison failure.

Each rule gets a separate empty cache for each of three rounds. The unedited author check primes both correct and owned-mutant builds and all observations. The whitespace check runs next, then the semantic token edit. Both use the same cache, toolchain and commit. The edit changes exactly one literal token in the private rule entry, with an anchor-count check; no rule directory is edited.

| rule | entry | literal before | literal after |
|---|---|---|---|
| no-var | rule.a | unexpectedVar | unexpectedVaz |
| no-empty | rule.ts | unexpectedBlock | unexpectedBlocx |
| eqeqeq | rule.ts | unexpected | unexpectee |

The semantic sample builds both the changed correct port and the changed owned-mutant port concurrently, with four jobs per split build. It runs the complete normal corpus: owned witnesses, captured upstream cases and inherited corner cases/options. All three port runtimes must match byte for byte, must differ from Go, and must fail certification. Execution remains under ASan/UBSan. Timing ends at that comparison failure, as an author check on a broken rule does. The later owned-mutant execution phase is not reached; its native build cost is included.

| loop | before warm whitespace, seconds | after warm semantic, seconds | objects recompiled in semantic rounds 1/2/3 | instrument |
|---|---:|---:|---|---|
| no-var | 14.094 | 17.802 | 35/35/35 | `go test ./stage1/cohere/lint -run '^TestRule$' -count=1 -timeout 30m -v -args -rule no-var -rule-semantic-change`; before substitutes `-rule-byte-change` |
| no-empty | 14.734 | 17.928 | 35/35/35 | `go test ./stage1/cohere/lint -run '^TestRule$' -count=1 -timeout 30m -v -args -rule no-empty -rule-semantic-change`; before substitutes `-rule-byte-change` |
| eqeqeq | 15.375 | 19.525 | 44/44/44 | `go test ./stage1/cohere/lint -run '^TestRule$' -count=1 -timeout 30m -v -args -rule eqeqeq -rule-semantic-change`; before substitutes `-rule-byte-change` |

Best of three, wall-clock including `go test` startup/build overhead. Before is a successful whitespace check; after is an expected failure on incorrect diagnostic IDs, not a successful rule certification. Every semantic run has actual clang object compilation, so none is an emitted-C cache hit.

Build-flags line for every sample: commit=8483db57037839c37ccc99cc8604cee3a2bdd0e4; nproc=5; cpu.max=400000 100000; go=go version go1.27.1 linux/amd64; clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261); node=v24.19.0; GOMAXPROCS=4; jobs=4 per split build; two builds concurrent; flags=-std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all; GOCACHE=/home/agent/.cache/go-build. Native object commands also add `-Wno-gnu-line-marker -fdebug-prefix-map=<snapshot>=/adamic-units -x cpp-output -c`. The exact argument lists appear in compiler traces.

| winning sample | round | cache | load before | load after | log |
|---|---:|---|---|---|---|
| no-var warm-whitespace | 3 | cached | `3.57 2.77 1.95 1/164 273020` | `3.14 2.71 1.95 3/164 274550` | [no-var-3-warm-whitespace-split.log](no-var-3-warm-whitespace-split.log) |
| no-var warm-semantic | 3 | cached | `3.14 2.71 1.95 1/164 274550` | `3.74 2.86 2.01 1/164 275833` | [no-var-3-warm-semantic-split.log](no-var-3-warm-semantic-split.log) |
| no-empty warm-whitespace | 2 | cached | `4.26 2.61 1.78 1/166 263639` | `3.61 2.54 1.77 1/164 265170` | [no-empty-2-warm-whitespace-split.log](no-empty-2-warm-whitespace-split.log) |
| no-empty warm-semantic | 2 | cached | `3.61 2.54 1.77 1/164 265170` | `3.30 2.54 1.79 2/165 266462` | [no-empty-2-warm-semantic-split.log](no-empty-2-warm-semantic-split.log) |
| eqeqeq warm-whitespace | 1 | cached | `1.99 1.67 1.39 1/165 254282` | `2.72 1.85 1.45 2/165 255810` | [eqeqeq-1-warm-whitespace-split.log](eqeqeq-1-warm-whitespace-split.log) |
| eqeqeq warm-semantic | 1 | cached | `2.72 1.85 1.45 2/165 255810` | `3.36 2.05 1.53 1/165 257101` | [eqeqeq-1-warm-semantic-split.log](eqeqeq-1-warm-semantic-split.log) |

All individual semantic samples and their largest phase:

| rule | round | wall seconds | objects | port build phase seconds | emitted JavaScript build seconds | comparison seconds |
|---|---:|---:|---:|---:|---:|---:|
| no-var | 1 | 17.830 | 35 | 12.143 | 2.150 | 0.935 |
| no-empty | 1 | 21.803 | 35 | 12.893 | 2.616 | 0.930 |
| eqeqeq | 1 | 19.525 | 44 | 13.826 | 2.090 | 0.941 |
| no-var | 2 | 19.773 | 35 | 13.583 | 2.161 | 0.873 |
| no-empty | 2 | 17.928 | 35 | 12.265 | 2.129 | 0.881 |
| eqeqeq | 2 | 20.150 | 44 | 14.567 | 2.053 | 0.940 |
| no-var | 3 | 17.802 | 35 | 12.291 | 2.025 | 0.813 |
| no-empty | 3 | 20.186 | 35 | 11.977 | 2.031 | 0.825 |
| eqeqeq | 3 | 27.750 | 44 | 21.216 | 2.604 | 1.161 |

Best-of-three semantic checks are under 20 seconds for all three rules, but individual samples are not uniformly below the target. Eqeqeq round 3 took 27.750 seconds wall-clock: 21.216 seconds were in the concurrent port-build phase, versus 2.604 for emitted JavaScript building and 1.161 for comparison. Broad native rebuild work dominates that slower sample; this is a measured failure-check time, not a latency guarantee.

All 34 unit names (32 function groups, state and main) were rebuilt. The pair performs 35 physical object compilations for no-var/no-empty and 44 for eqeqeq because shared objects compile once while differing correct/mutant objects need separate versions. A literal edit can affect shared declarations and program-wide string numbering; this prototype provides no general one-object guarantee. Broad invalidation is observed in the traces. The port-build phase, including keying, loading/lowering, C emission, preprocessing, compilation and linking, is the largest measured phase. No compiler scheduling or native implementation changes were made.

Correctness and reproduction:

```bash
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/lint/semantic-evidence/measure.py
python3 stage1/cohere/lint/semantic-evidence/validate.py
```

[validate.py](validate.py) specifies exact gate commands. [normal-parity.log](normal-parity.log) verifies selected/full registration, split/fresh whole-file output, cached/uncached bytes and compiler selection. [all-rule-uncached.log](all-rule-uncached.log) passes the complete all-rule upstream/inherited comparison on ordinary whole-file compilation. [vet.log](vet.log) passes touched packages. [native-parity-mutant.log](native-parity-mutant.log) adds a `puts` call to actual emitted C only during semantic checks: compilation and execution succeed, then the new parity check fails with `semantic edit Node/native outputs differ`. The nine semantic source mutants compile/run on all three port runtimes and fail only the Go comparison with the explicit expected marker.

[measurements.json](measurements.json) contains every command, time, exit, cache state, full flags and load readings. Every `.clang.json` contains actual compiler invocations; objects are counted from `-c -x cpp-output`, not cache directory sizes. [setup.log](setup.log) records setup timings and build flags. The earlier whitespace timings remain as whitespace measurements, not semantic edit evidence.

Scope: normal author checks remain unchanged in behavior when the benchmark flag is absent. No new cache was introduced or key weakened. Full repository, throughput/profile and fifty-rule-fleet gates were not run. Full all-rule comparison and focused parity were run. Future behavior-preserving port corrections can have different invalidation patterns and timing; these measurements establish observable ID edits and identical failures, not every possible edit.
