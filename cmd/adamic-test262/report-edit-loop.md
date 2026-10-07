Built: Node observations survive runtime/lowering edits; native observations invalidate on both; cold scheduling and preparation costs are measured and reduced.
Commits: before b9f652d71b0577f5914f1d3de889b9a1b2a6eaaf, after d314ff43172cfd23c3e62f088efe5563714f4f36; final implementation 3caa76777952f67ceff1f0c2b8c45c420a47a998 additionally supports go run relocation.
Commands and outputs: 54 paired timing samples plus untimed primes; all JSON and ordered progress/table bytes identical; uncached runner, race, vet and filtered oracle PASS.
Mutants: seven independent restored mutations fail their intended assertions; compact detectors and failures are in evidence/edit-mutants.json.
Not covered: full repository gate/corpus, non-Linux systems, semantic library changes, and optimized lowering or clang internals outside cmd/adamic-test262.

## What changed

The previous runner included the runtime key in the program artifact path and the full executable in Node's context. Its Node keys therefore missed when a runtime or lowering edit changed the compiler. Node now uses stable program paths and an identity of the embedded runner harness sources, Go build settings, Node bytes/version, exact program/command/adaptation flags, and execution context. Compiler implementation and tests are excluded from the Node harness manifest; capture, cache and adaptation sources are included. The manifest comes from the built executable, not mutable checkout files.

Compiler and native contexts retain the executed binary identity. Native still keys emitted C, the runtime snapshot key and actual link/run commands. A lowering-source byte fingerprint adds conservative invalidation: Go can elide same-length comment edits from a linked binary. An initial experiment caught a native hit on such a second edit; TestLoweringSourceEdit now holds that case, even with identical executed bytes. The fingerprint adds misses and does not replace binary identity. All three observation caches retain ADAMIC_GATE_UNCACHED=1 bypass.

With multiple jobs, attempted tests are scheduled by decreasing program size, after serial classification and limit selection. Reduction and progress remain in serial order. The normal compiler-worker path no longer builds a backup adamic executable on every invocation; a shared once-only fallback builds it if worker transport fails. The explicit -root and -compiler-subprocess reference paths retain the original build.

`-profile <file>` writes per-test compile, clang, native and Node phase offsets and actual cache-hit flags, plus worker identities and preparation phases, separately from verdict output. A hit means the execution callback did not run. The final implementation also excludes the runner's unused executable location from compiler/native context: go run relocates identical executable bytes. Node/clang locations, binary contents, inputs, actual native command paths and source fingerprints remain keyed. A copied-executable test and its mutant prove this distinction.

## Measurement method

Before is merged 2493821 plus profiling-only commit b9f652d. The paired runs freeze the measured implementation at d314ff43172cfd23c3e62f088efe5563714f4f36. Final implementation 3caa76777952f67ceff1f0c2b8c45c420a47a998 adds the go run location correction; it does not change execution or eligibility in the fixed-executable paired instrument. The supplementary actual go run measurements below use that final implementation.

All compiler/runtime source outside this territory shares the original base e011f8f. Corpus: test262 `5992dc3b60faf62a48fd6be8a40ae9d9a8c84d81`, the Object/function library-report pin used in the first report. Two isolated detached worktrees contain identical cohere source. Each directory is primed once per runner; successive runtime edits replace exactly one comment byte, and lowering edits replace exactly one blank-line byte with a space, in three distinct source locations. Source bytes and length otherwise remain identical and are restored in finally blocks. Every mutation's path, offset, before/after bytes and SHA256 are recorded. The lowering edit preserves Go semantics while moving debug line tables; the source fingerprint independently covers edits Go elides. Edited source hashes must be distinct.

Edit timing includes `go build -o <runner> ./cmd/adamic-test262` after the change, then invoking that runner on the directory. Rebuild and runner durations are recorded separately. Edit rows include this rebuild cost. Private Go build caches per runner are warmed by untimed original-source primes, preventing stopped experiments from giving one side warmed edit artifacts. Each cold run has a fresh result/runtime cache root; Go build caches remain warm. Before and after are paired, three rounds, on this machine. All run output is redirected to files. Time is measured with Python time.monotonic. Tests finish before the final timing suite; the supplementary go run prime overlapped untimed build-cache priming and part of the first Math pair. All load readings are preserved and best-of-three results below select their own samples.

Instrument: `TMPDIR=/workspace/scratch/test262-tmp python3 cmd/adamic-test262/measure_edits.py --before-root /tmp/test262-loop-before --after-root /tmp/test262-loop-after --test262 /tmp/adamic-test262-corpus --output /workspace/scratch/test262-followup-measured > /workspace/scratch/test262-followup-measured.log 2>&1`.

Native build flags on every run: `-std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all`. Sanitizers, optimization, deadlines and output caps are unchanged. Full metadata, exact build/invocation commands, per-worker busy times and best cold per-test phases are in [edit-measurements.json](evidence/edit-measurements.json). Raw JSON, profiles and logs stay at `/workspace/scratch/test262-followup-measured`, not committed. Earlier stopped experiments are excluded from this timing table. Setup reported done in 34s on 5 processors, quota 4 CPUs; used /workspace/adamic-tools/env.sh. Temporary-space exhaustion was resolved by moving build caches and TMPDIR to the workspace; affected checks were rerun successfully.

## Best-of-three timings

Seconds; edit rows include rebuild. Cold rows time the runner with its already-built executable, matching the first report's cold instrument. Every selected row has its exact command and build-flags line below.

| Loop | Before | After | Instrument |
| --- | ---: | ---: | --- |
| built-ins/Math runtime-edit | 25.061 | 18.042 | before build `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-Math-before GOCACHE=/workspace/scratch/test262-followup-measured/go-build-before GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 go build -o /workspace/scratch/test262-followup-measured/runner-before ./cmd/adamic-test262` then run `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-Math-before GOCACHE=/workspace/scratch/test262-followup-measured/go-build-before GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 /workspace/scratch/test262-followup-measured/runner-before -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-followup-measured/Math-runtime-edit-1-before-profile.json built-ins/Math`; after build `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-Math-after GOCACHE=/workspace/scratch/test262-followup-measured/go-build-after GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 go build -o /workspace/scratch/test262-followup-measured/runner-after ./cmd/adamic-test262` then run `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-Math-after GOCACHE=/workspace/scratch/test262-followup-measured/go-build-after GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 /workspace/scratch/test262-followup-measured/runner-after -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-followup-measured/Math-runtime-edit-2-after-profile.json built-ins/Math` |
| built-ins/Math lowering-edit | 24.092 | 17.282 | before build `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-Math-before GOCACHE=/workspace/scratch/test262-followup-measured/go-build-before GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 go build -o /workspace/scratch/test262-followup-measured/runner-before ./cmd/adamic-test262` then run `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-Math-before GOCACHE=/workspace/scratch/test262-followup-measured/go-build-before GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 /workspace/scratch/test262-followup-measured/runner-before -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-followup-measured/Math-lowering-edit-2-before-profile.json built-ins/Math`; after build `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-Math-after GOCACHE=/workspace/scratch/test262-followup-measured/go-build-after GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 go build -o /workspace/scratch/test262-followup-measured/runner-after ./cmd/adamic-test262` then run `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-Math-after GOCACHE=/workspace/scratch/test262-followup-measured/go-build-after GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 /workspace/scratch/test262-followup-measured/runner-after -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-followup-measured/Math-lowering-edit-2-after-profile.json built-ins/Math` |
| built-ins/Math cold | 16.244 | 14.263 | before run `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-Math-cold-2-before GOCACHE=/workspace/scratch/test262-followup-measured/go-build-before GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 /workspace/scratch/test262-followup-measured/runner-before -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-followup-measured/Math-cold-2-before-profile.json built-ins/Math`; after run `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-Math-cold-2-after GOCACHE=/workspace/scratch/test262-followup-measured/go-build-after GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 /workspace/scratch/test262-followup-measured/runner-after -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-followup-measured/Math-cold-2-after-profile.json built-ins/Math` |
| built-ins/String/prototype/padStart runtime-edit | 15.600 | 10.616 | before build `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-padStart-before GOCACHE=/workspace/scratch/test262-followup-measured/go-build-before GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 go build -o /workspace/scratch/test262-followup-measured/runner-before ./cmd/adamic-test262` then run `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-padStart-before GOCACHE=/workspace/scratch/test262-followup-measured/go-build-before GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 /workspace/scratch/test262-followup-measured/runner-before -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-followup-measured/padStart-runtime-edit-0-before-profile.json built-ins/String/prototype/padStart`; after build `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-padStart-after GOCACHE=/workspace/scratch/test262-followup-measured/go-build-after GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 go build -o /workspace/scratch/test262-followup-measured/runner-after ./cmd/adamic-test262` then run `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-padStart-after GOCACHE=/workspace/scratch/test262-followup-measured/go-build-after GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 /workspace/scratch/test262-followup-measured/runner-after -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-followup-measured/padStart-runtime-edit-1-after-profile.json built-ins/String/prototype/padStart` |
| built-ins/String/prototype/padStart lowering-edit | 15.052 | 10.223 | before build `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-padStart-before GOCACHE=/workspace/scratch/test262-followup-measured/go-build-before GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 go build -o /workspace/scratch/test262-followup-measured/runner-before ./cmd/adamic-test262` then run `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-padStart-before GOCACHE=/workspace/scratch/test262-followup-measured/go-build-before GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 /workspace/scratch/test262-followup-measured/runner-before -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-followup-measured/padStart-lowering-edit-2-before-profile.json built-ins/String/prototype/padStart`; after build `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-padStart-after GOCACHE=/workspace/scratch/test262-followup-measured/go-build-after GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 go build -o /workspace/scratch/test262-followup-measured/runner-after ./cmd/adamic-test262` then run `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-padStart-after GOCACHE=/workspace/scratch/test262-followup-measured/go-build-after GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 /workspace/scratch/test262-followup-measured/runner-after -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-followup-measured/padStart-lowering-edit-1-after-profile.json built-ins/String/prototype/padStart` |
| built-ins/String/prototype/padStart cold | 7.119 | 5.359 | before run `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-padStart-cold-2-before GOCACHE=/workspace/scratch/test262-followup-measured/go-build-before GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 /workspace/scratch/test262-followup-measured/runner-before -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-followup-measured/padStart-cold-2-before-profile.json built-ins/String/prototype/padStart`; after run `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-padStart-cold-1-after GOCACHE=/workspace/scratch/test262-followup-measured/go-build-after GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 /workspace/scratch/test262-followup-measured/runner-after -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-followup-measured/padStart-cold-1-after-profile.json built-ins/String/prototype/padStart` |
| built-ins/Array/prototype/sort runtime-edit | 54.913 | 47.832 | before build `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-sort-before GOCACHE=/workspace/scratch/test262-followup-measured/go-build-before GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 go build -o /workspace/scratch/test262-followup-measured/runner-before ./cmd/adamic-test262` then run `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-sort-before GOCACHE=/workspace/scratch/test262-followup-measured/go-build-before GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 /workspace/scratch/test262-followup-measured/runner-before -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-followup-measured/sort-runtime-edit-2-before-profile.json built-ins/Array/prototype/sort`; after build `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-sort-after GOCACHE=/workspace/scratch/test262-followup-measured/go-build-after GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 go build -o /workspace/scratch/test262-followup-measured/runner-after ./cmd/adamic-test262` then run `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-sort-after GOCACHE=/workspace/scratch/test262-followup-measured/go-build-after GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 /workspace/scratch/test262-followup-measured/runner-after -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-followup-measured/sort-runtime-edit-2-after-profile.json built-ins/Array/prototype/sort` |
| built-ins/Array/prototype/sort lowering-edit | 53.416 | 47.824 | before build `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-sort-before GOCACHE=/workspace/scratch/test262-followup-measured/go-build-before GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 go build -o /workspace/scratch/test262-followup-measured/runner-before ./cmd/adamic-test262` then run `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-sort-before GOCACHE=/workspace/scratch/test262-followup-measured/go-build-before GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 /workspace/scratch/test262-followup-measured/runner-before -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-followup-measured/sort-lowering-edit-0-before-profile.json built-ins/Array/prototype/sort`; after build `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-sort-after GOCACHE=/workspace/scratch/test262-followup-measured/go-build-after GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 go build -o /workspace/scratch/test262-followup-measured/runner-after ./cmd/adamic-test262` then run `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-sort-after GOCACHE=/workspace/scratch/test262-followup-measured/go-build-after GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 /workspace/scratch/test262-followup-measured/runner-after -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-followup-measured/sort-lowering-edit-0-after-profile.json built-ins/Array/prototype/sort` |
| built-ins/Array/prototype/sort cold | 45.188 | 42.558 | before run `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-sort-cold-1-before GOCACHE=/workspace/scratch/test262-followup-measured/go-build-before GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 /workspace/scratch/test262-followup-measured/runner-before -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-followup-measured/sort-cold-1-before-profile.json built-ins/Array/prototype/sort`; after run `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-sort-cold-0-after GOCACHE=/workspace/scratch/test262-followup-measured/go-build-after GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 /workspace/scratch/test262-followup-measured/runner-after -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-followup-measured/sort-cold-0-after-profile.json built-ins/Array/prototype/sort` |

**built-ins/Math runtime-edit before**: rebuild 5.734s, runner 19.327s.

Build: `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-Math-before GOCACHE=/workspace/scratch/test262-followup-measured/go-build-before GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 go build -o /workspace/scratch/test262-followup-measured/runner-before ./cmd/adamic-test262`.

Run: `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-Math-before GOCACHE=/workspace/scratch/test262-followup-measured/go-build-before GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 /workspace/scratch/test262-followup-measured/runner-before -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-followup-measured/Math-runtime-edit-1-before-profile.json built-ins/Math`.

Build-flags: commit `b9f652d71b0577f5914f1d3de889b9a1b2a6eaaf`; nproc=5; cpu.max=`400000 100000`; go version go1.27.1 linux/amd64; clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261); node=v24.19.0; GOFLAGS empty; GOMAXPROCS=4; TMPDIR=/workspace/scratch/test262-tmp; cache=primed before edit; native flags as above; load before `4.77 3.84 2.85 1/172 155327`, after `4.43 3.81 2.87 1/173 156895`.

**built-ins/Math runtime-edit after**: rebuild 5.564s, runner 12.478s.

Build: `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-Math-after GOCACHE=/workspace/scratch/test262-followup-measured/go-build-after GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 go build -o /workspace/scratch/test262-followup-measured/runner-after ./cmd/adamic-test262`.

Run: `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-Math-after GOCACHE=/workspace/scratch/test262-followup-measured/go-build-after GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 /workspace/scratch/test262-followup-measured/runner-after -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-followup-measured/Math-runtime-edit-2-after-profile.json built-ins/Math`.

Build-flags: commit `d314ff43172cfd23c3e62f088efe5563714f4f36`; nproc=5; cpu.max=`400000 100000`; go version go1.27.1 linux/amd64; clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261); node=v24.19.0; GOFLAGS empty; GOMAXPROCS=4; TMPDIR=/workspace/scratch/test262-tmp; cache=primed before edit; native flags as above; load before `3.53 3.65 2.85 1/175 159138`, after `3.13 3.54 2.83 1/175 159799`.

**built-ins/Math lowering-edit before**: rebuild 8.986s, runner 15.106s.

Build: `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-Math-before GOCACHE=/workspace/scratch/test262-followup-measured/go-build-before GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 go build -o /workspace/scratch/test262-followup-measured/runner-before ./cmd/adamic-test262`.

Run: `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-Math-before GOCACHE=/workspace/scratch/test262-followup-measured/go-build-before GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 /workspace/scratch/test262-followup-measured/runner-before -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-followup-measured/Math-lowering-edit-2-before-profile.json built-ins/Math`.

Build-flags: commit `b9f652d71b0577f5914f1d3de889b9a1b2a6eaaf`; nproc=5; cpu.max=`400000 100000`; go version go1.27.1 linux/amd64; clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261); node=v24.19.0; GOFLAGS empty; GOMAXPROCS=4; TMPDIR=/workspace/scratch/test262-tmp; cache=primed before edit; native flags as above; load before `3.32 3.51 2.88 1/176 164063`, after `4.12 3.67 2.95 1/178 165584`.

**built-ins/Math lowering-edit after**: rebuild 9.412s, runner 7.870s.

Build: `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-Math-after GOCACHE=/workspace/scratch/test262-followup-measured/go-build-after GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 go build -o /workspace/scratch/test262-followup-measured/runner-after ./cmd/adamic-test262`.

Run: `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-Math-after GOCACHE=/workspace/scratch/test262-followup-measured/go-build-after GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 /workspace/scratch/test262-followup-measured/runner-after -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-followup-measured/Math-lowering-edit-2-after-profile.json built-ins/Math`.

Build-flags: commit `d314ff43172cfd23c3e62f088efe5563714f4f36`; nproc=5; cpu.max=`400000 100000`; go version go1.27.1 linux/amd64; clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261); node=v24.19.0; GOFLAGS empty; GOMAXPROCS=4; TMPDIR=/workspace/scratch/test262-tmp; cache=primed before edit; native flags as above; load before `4.12 3.67 2.95 1/178 165584`, after `3.81 3.62 2.95 1/178 166193`.

**built-ins/Math cold before**: rebuild 0.000s, runner 16.244s.

Run: `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-Math-cold-2-before GOCACHE=/workspace/scratch/test262-followup-measured/go-build-before GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 /workspace/scratch/test262-followup-measured/runner-before -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-followup-measured/Math-cold-2-before-profile.json built-ins/Math`.

Build-flags: commit `b9f652d71b0577f5914f1d3de889b9a1b2a6eaaf`; nproc=5; cpu.max=`400000 100000`; go version go1.27.1 linux/amd64; clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261); node=v24.19.0; GOFLAGS empty; GOMAXPROCS=4; TMPDIR=/workspace/scratch/test262-tmp; cache=cold; native flags as above; load before `3.79 3.65 3.01 1/180 172313`, after `3.86 3.67 3.03 1/179 173791`.

**built-ins/Math cold after**: rebuild 0.000s, runner 14.263s.

Run: `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-Math-cold-2-after GOCACHE=/workspace/scratch/test262-followup-measured/go-build-after GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 /workspace/scratch/test262-followup-measured/runner-after -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-followup-measured/Math-cold-2-after-profile.json built-ins/Math`.

Build-flags: commit `d314ff43172cfd23c3e62f088efe5563714f4f36`; nproc=5; cpu.max=`400000 100000`; go version go1.27.1 linux/amd64; clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261); node=v24.19.0; GOFLAGS empty; GOMAXPROCS=4; TMPDIR=/workspace/scratch/test262-tmp; cache=cold; native flags as above; load before `3.86 3.67 3.03 1/179 173863`, after `3.91 3.69 3.04 1/179 175260`.

**built-ins/String/prototype/padStart runtime-edit before**: rebuild 5.285s, runner 10.315s.

Build: `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-padStart-before GOCACHE=/workspace/scratch/test262-followup-measured/go-build-before GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 go build -o /workspace/scratch/test262-followup-measured/runner-before ./cmd/adamic-test262`.

Run: `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-padStart-before GOCACHE=/workspace/scratch/test262-followup-measured/go-build-before GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 /workspace/scratch/test262-followup-measured/runner-before -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-followup-measured/padStart-runtime-edit-0-before-profile.json built-ins/String/prototype/padStart`.

Build-flags: commit `b9f652d71b0577f5914f1d3de889b9a1b2a6eaaf`; nproc=5; cpu.max=`400000 100000`; go version go1.27.1 linux/amd64; clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261); node=v24.19.0; GOFLAGS empty; GOMAXPROCS=4; TMPDIR=/workspace/scratch/test262-tmp; cache=primed before edit; native flags as above; load before `3.67 3.64 3.03 1/179 175531`, after `3.06 3.50 3.00 1/182 175830`.

**built-ins/String/prototype/padStart runtime-edit after**: rebuild 5.290s, runner 5.326s.

Build: `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-padStart-after GOCACHE=/workspace/scratch/test262-followup-measured/go-build-after GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 go build -o /workspace/scratch/test262-followup-measured/runner-after ./cmd/adamic-test262`.

Run: `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-padStart-after GOCACHE=/workspace/scratch/test262-followup-measured/go-build-after GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 /workspace/scratch/test262-followup-measured/runner-after -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-followup-measured/padStart-runtime-edit-1-after-profile.json built-ins/String/prototype/padStart`.

Build-flags: commit `d314ff43172cfd23c3e62f088efe5563714f4f36`; nproc=5; cpu.max=`400000 100000`; go version go1.27.1 linux/amd64; clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261); node=v24.19.0; GOFLAGS empty; GOMAXPROCS=4; TMPDIR=/workspace/scratch/test262-tmp; cache=primed before edit; native flags as above; load before `2.73 3.38 2.97 3/179 176509`, after `2.68 3.35 2.97 1/179 176703`.

**built-ins/String/prototype/padStart lowering-edit before**: rebuild 8.787s, runner 6.265s.

Build: `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-padStart-before GOCACHE=/workspace/scratch/test262-followup-measured/go-build-before GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 go build -o /workspace/scratch/test262-followup-measured/runner-before ./cmd/adamic-test262`.

Run: `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-padStart-before GOCACHE=/workspace/scratch/test262-followup-measured/go-build-before GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 /workspace/scratch/test262-followup-measured/runner-before -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-followup-measured/padStart-lowering-edit-2-before-profile.json built-ins/String/prototype/padStart`.

Build-flags: commit `b9f652d71b0577f5914f1d3de889b9a1b2a6eaaf`; nproc=5; cpu.max=`400000 100000`; go version go1.27.1 linux/amd64; clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261); node=v24.19.0; GOFLAGS empty; GOMAXPROCS=4; TMPDIR=/workspace/scratch/test262-tmp; cache=primed before edit; native flags as above; load before `1.75 2.88 2.83 2/176 177974`, after `1.79 2.84 2.82 1/175 178220`.

**built-ins/String/prototype/padStart lowering-edit after**: rebuild 9.093s, runner 1.130s.

Build: `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-padStart-after GOCACHE=/workspace/scratch/test262-followup-measured/go-build-after GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 go build -o /workspace/scratch/test262-followup-measured/runner-after ./cmd/adamic-test262`.

Run: `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-padStart-after GOCACHE=/workspace/scratch/test262-followup-measured/go-build-after GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 /workspace/scratch/test262-followup-measured/runner-after -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-followup-measured/padStart-lowering-edit-1-after-profile.json built-ins/String/prototype/padStart`.

Build-flags: commit `d314ff43172cfd23c3e62f088efe5563714f4f36`; nproc=5; cpu.max=`400000 100000`; go version go1.27.1 linux/amd64; clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261); node=v24.19.0; GOFLAGS empty; GOMAXPROCS=4; TMPDIR=/workspace/scratch/test262-tmp; cache=primed before edit; native flags as above; load before `1.63 2.90 2.83 1/176 177833`, after `1.75 2.88 2.83 1/176 177974`.

**built-ins/String/prototype/padStart cold before**: rebuild 0.000s, runner 7.119s.

Run: `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-padStart-cold-2-before GOCACHE=/workspace/scratch/test262-followup-measured/go-build-before GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 /workspace/scratch/test262-followup-measured/runner-before -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-followup-measured/padStart-cold-2-before-profile.json built-ins/String/prototype/padStart`.

Build-flags: commit `b9f652d71b0577f5914f1d3de889b9a1b2a6eaaf`; nproc=5; cpu.max=`400000 100000`; go version go1.27.1 linux/amd64; clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261); node=v24.19.0; GOFLAGS empty; GOMAXPROCS=4; TMPDIR=/workspace/scratch/test262-tmp; cache=cold; native flags as above; load before `1.61 2.67 2.76 1/172 179394`, after `1.56 2.64 2.75 1/172 179598`.

**built-ins/String/prototype/padStart cold after**: rebuild 0.000s, runner 5.359s.

Run: `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-padStart-cold-1-after GOCACHE=/workspace/scratch/test262-followup-measured/go-build-after GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 /workspace/scratch/test262-followup-measured/runner-after -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-followup-measured/padStart-cold-1-after-profile.json built-ins/String/prototype/padStart`.

Build-flags: commit `d314ff43172cfd23c3e62f088efe5563714f4f36`; nproc=5; cpu.max=`400000 100000`; go version go1.27.1 linux/amd64; clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261); node=v24.19.0; GOFLAGS empty; GOMAXPROCS=4; TMPDIR=/workspace/scratch/test262-tmp; cache=cold; native flags as above; load before `1.66 2.69 2.77 1/171 179193`, after `1.61 2.67 2.76 1/171 179322`.

**built-ins/Array/prototype/sort runtime-edit before**: rebuild 5.463s, runner 49.450s.

Build: `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-sort-before GOCACHE=/workspace/scratch/test262-followup-measured/go-build-before GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 go build -o /workspace/scratch/test262-followup-measured/runner-before ./cmd/adamic-test262`.

Run: `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-sort-before GOCACHE=/workspace/scratch/test262-followup-measured/go-build-before GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 /workspace/scratch/test262-followup-measured/runner-before -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-followup-measured/sort-runtime-edit-2-before-profile.json built-ins/Array/prototype/sort`.

Build-flags: commit `b9f652d71b0577f5914f1d3de889b9a1b2a6eaaf`; nproc=5; cpu.max=`400000 100000`; go version go1.27.1 linux/amd64; clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261); node=v24.19.0; GOFLAGS empty; GOMAXPROCS=4; TMPDIR=/workspace/scratch/test262-tmp; cache=primed before edit; native flags as above; load before `1.50 1.93 2.40 1/160 181439`, after `1.35 1.82 2.33 1/162 181776`.

**built-ins/Array/prototype/sort runtime-edit after**: rebuild 5.379s, runner 42.453s.

Build: `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-sort-after GOCACHE=/workspace/scratch/test262-followup-measured/go-build-after GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 go build -o /workspace/scratch/test262-followup-measured/runner-after ./cmd/adamic-test262`.

Run: `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-sort-after GOCACHE=/workspace/scratch/test262-followup-measured/go-build-after GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 /workspace/scratch/test262-followup-measured/runner-after -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-followup-measured/sort-runtime-edit-2-after-profile.json built-ins/Array/prototype/sort`.

Build-flags: commit `d314ff43172cfd23c3e62f088efe5563714f4f36`; nproc=5; cpu.max=`400000 100000`; go version go1.27.1 linux/amd64; clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261); node=v24.19.0; GOFLAGS empty; GOMAXPROCS=4; TMPDIR=/workspace/scratch/test262-tmp; cache=primed before edit; native flags as above; load before `1.35 1.82 2.33 2/162 181776`, after `1.20 1.71 2.27 1/162 181982`.

**built-ins/Array/prototype/sort lowering-edit before**: rebuild 8.782s, runner 44.634s.

Build: `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-sort-before GOCACHE=/workspace/scratch/test262-followup-measured/go-build-before GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 go build -o /workspace/scratch/test262-followup-measured/runner-before ./cmd/adamic-test262`.

Run: `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-sort-before GOCACHE=/workspace/scratch/test262-followup-measured/go-build-before GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 /workspace/scratch/test262-followup-measured/runner-before -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-followup-measured/sort-lowering-edit-0-before-profile.json built-ins/Array/prototype/sort`.

Build-flags: commit `b9f652d71b0577f5914f1d3de889b9a1b2a6eaaf`; nproc=5; cpu.max=`400000 100000`; go version go1.27.1 linux/amd64; clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261); node=v24.19.0; GOFLAGS empty; GOMAXPROCS=4; TMPDIR=/workspace/scratch/test262-tmp; cache=primed before edit; native flags as above; load before `1.20 1.71 2.27 1/162 181982`, after `1.23 1.65 2.21 1/161 182277`.

**built-ins/Array/prototype/sort lowering-edit after**: rebuild 9.269s, runner 38.555s.

Build: `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-sort-after GOCACHE=/workspace/scratch/test262-followup-measured/go-build-after GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 go build -o /workspace/scratch/test262-followup-measured/runner-after ./cmd/adamic-test262`.

Run: `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-edit-sort-after GOCACHE=/workspace/scratch/test262-followup-measured/go-build-after GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 /workspace/scratch/test262-followup-measured/runner-after -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-followup-measured/sort-lowering-edit-0-after-profile.json built-ins/Array/prototype/sort`.

Build-flags: commit `d314ff43172cfd23c3e62f088efe5563714f4f36`; nproc=5; cpu.max=`400000 100000`; go version go1.27.1 linux/amd64; clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261); node=v24.19.0; GOFLAGS empty; GOMAXPROCS=4; TMPDIR=/workspace/scratch/test262-tmp; cache=primed before edit; native flags as above; load before `1.23 1.65 2.21 1/161 182277`, after `1.15 1.57 2.16 1/161 182433`.

**built-ins/Array/prototype/sort cold before**: rebuild 0.000s, runner 45.188s.

Run: `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-sort-cold-1-before GOCACHE=/workspace/scratch/test262-followup-measured/go-build-before GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 /workspace/scratch/test262-followup-measured/runner-before -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-followup-measured/sort-cold-1-before-profile.json built-ins/Array/prototype/sort`.

Build-flags: commit `b9f652d71b0577f5914f1d3de889b9a1b2a6eaaf`; nproc=5; cpu.max=`400000 100000`; go version go1.27.1 linux/amd64; clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261); node=v24.19.0; GOFLAGS empty; GOMAXPROCS=4; TMPDIR=/workspace/scratch/test262-tmp; cache=cold; native flags as above; load before `1.40 1.60 2.02 1/162 183971`, after `1.29 1.54 1.98 1/161 184214`.

**built-ins/Array/prototype/sort cold after**: rebuild 0.000s, runner 42.558s.

Run: `XDG_CACHE_HOME=/workspace/scratch/test262-followup-measured/cache-sort-cold-0-after GOCACHE=/workspace/scratch/test262-followup-measured/go-build-after GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 /workspace/scratch/test262-followup-measured/runner-after -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-followup-measured/sort-cold-0-after-profile.json built-ins/Array/prototype/sort`.

Build-flags: commit `d314ff43172cfd23c3e62f088efe5563714f4f36`; nproc=5; cpu.max=`400000 100000`; go version go1.27.1 linux/amd64; clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261); node=v24.19.0; GOFLAGS empty; GOMAXPROCS=4; TMPDIR=/workspace/scratch/test262-tmp; cache=cold; native flags as above; load before `1.36 1.61 2.05 1/162 183738`, after `1.40 1.60 2.02 1/162 183899`.

## Edit hit/miss proof

For each edit kind and every round, the updated runner executes every native observation callback and executes no Node observation callback. Skipped/refused tests never reach native or Node; counts refer to eligible observations.

| Directory | Tests / attempted | Native misses per edit | Native hits | Node hits per edit | Node misses |
| --- | ---: | ---: | ---: | ---: | ---: |
| built-ins/Math | 327 / 166 | 118 | 0 | 118 | 0 |
| built-ins/String/prototype/padStart | 13 / 7 | 3 | 0 | 3 | 0 |
| built-ins/Array/prototype/sort | 54 / 43 | 6 | 0 | 6 | 0 |

All edited JSON and entire progress/table logs match the original-source prime byte-for-byte. Their SHA256 digests are recorded; the script compares actual bytes, not only verdict totals. The final reports also match the first report's original serial snapshots.

## Cold phases and worker overlap

The following uses the selected best cold sample for each runner. Phase totals are accumulated worker wall time, including waiting for child processes; they are not CPU seconds. `native-observation` wraps clang/native/cache work and must not be added again to clang plus native. Per-test durations for every attempted test in these samples are in the measurement JSON.

| Directory / runner | Compile | Clang | Native run | Node run | Peak overlapping workers |
| --- | ---: | ---: | ---: | ---: | ---: |
| built-ins/Math / before | 7.201 | 19.379 | 0.413 | 8.387 | 4 |
| built-ins/Math / after | 7.479 | 19.754 | 0.405 | 8.444 | 4 |
| built-ins/String/prototype/padStart / before | 0.355 | 0.452 | 0.010 | 0.178 | 4 |
| built-ins/String/prototype/padStart / after | 0.340 | 0.540 | 0.011 | 0.181 | 4 |
| built-ins/Array/prototype/sort / before | 21.390 | 23.835 | 0.024 | 0.451 | 4 |
| built-ins/Array/prototype/sort / after | 21.329 | 23.544 | 0.023 | 0.477 | 4 |

Each phase total carries the same build-flags line as its selected cold row above. All four workers overlap. Their busy intervals explain the sort tail:

| Sort runner / worker | Tests | Busy wall seconds | First start | Last finish |
| --- | ---: | ---: | ---: | ---: |
| before / 0 | 15 | 5.648 | 6.689 | 12.337 |
| before / 1 | 6 | 0.858 | 6.689 | 7.547 |
| before / 2 | 14 | 38.384 | 6.689 | 45.073 |
| before / 3 | 8 | 0.840 | 6.689 | 7.530 |
| after / 0 | 1 | 37.402 | 5.062 | 42.465 |
| after / 1 | 20 | 1.492 | 5.063 | 6.555 |
| after / 2 | 1 | 5.014 | 5.063 | 10.077 |
| after / 3 | 21 | 1.500 | 5.063 | 6.563 |

Offsets are seconds from the runner's common monotonic origin. These intervals carry their corresponding selected sort cold row's build-flags line.

before: longest test `built-ins/Array/prototype/sort/stability-2048-elements.js` spans 37.852s, starting at 7.221s: compile 18.377s, clang 19.393s, native 0.005s, Node 0.072s. It carries the corresponding selected sort cold build-flags line above.

after: longest test `built-ins/Array/prototype/sort/stability-2048-elements.js` spans 37.402s, starting at 5.062s: compile 18.105s, clang 19.218s, native 0.006s, Node 0.069s. It carries the corresponding selected sort cold build-flags line above.

The measured single-test compile-plus-clang floor is 37.323s. Extra jobs cannot subdivide that one load/lower/C request or clang translation unit. The runner schedules it earlier and avoids building the unused compiler; it preserves native.Flags(Sanitize:true), including -O1 and ASan/UBSan. Changing optimization/sanitizers or rewriting emitted C for a timing win would change the execution being checked. Speeding the dominant lowering/emission work requires changes outside this territory; clang itself is external. Runtime archive preparation is also shared internal/native behavior: selected after-cold preparation costs are listed below. Compiler identities remain conservative.

built-ins/Math: go-build 0.000s, runtime-library 4.221s, cache-context 0.841s. Same build-flags line as its selected cold row.

built-ins/String/prototype/padStart: go-build 0.000s, runtime-library 4.082s, cache-context 0.786s. Same build-flags line as its selected cold row.

built-ins/Array/prototype/sort: go-build 0.000s, runtime-library 4.074s, cache-context 0.923s. Same build-flags line as its selected cold row.

## Actual go run warm path

Latest code, actual library-driver command, no -work. One cold prime followed by three warm runs. This supplementary check is not an interleaved before/after comparison. All eligible compiler successes, native observations and Node observations hit on each warm run; refusals stay fresh. JSON and progress/table bytes equal the prime. Full [metadata](evidence/go-run-measurements.json); raw files in /workspace/scratch/test262-driver.

| Loop | Before | After | Instrument |
| --- | ---: | ---: | --- |
| sort (54), warm best of 3 | n/a | 1.767s | `TMPDIR=/workspace/scratch/test262-tmp XDG_CACHE_HOME=/workspace/scratch/test262-driver/cache GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 GOCACHE=/home/agent/.cache/go-build go run ./cmd/adamic-test262 -jobs 4 -adapt -json -test262 /tmp/adamic-test262-corpus -profile /workspace/scratch/test262-driver/3-profile.json built-ins/Array/prototype/sort` |

Build-flags: commit `3caa76777952f67ceff1f0c2b8c45c420a47a998`; nproc=5; cpu.max=`400000 100000`; go version go1.27.1 linux/amd64; clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261); node=v24.19.0; GOFLAGS empty; GOMAXPROCS=4; TMPDIR=/workspace/scratch/test262-tmp; cache=warm; native flags as above; load before `4.83 3.74 2.78 2/190 153194`, after `4.60 3.71 2.77 2/181 153285`.

## Checks and mutants

Exact validation commands and outputs are in [edit-validation.json](evidence/edit-validation.json). Uncached runner tests, the targeted race run, go vet, gofmt and git diff --check pass. The filtered uncached oracle covers seven Math/string/padding fixtures, with zero observation-cache hits. Full gate and corpus were not run. Go source and runtime edits occurred only in detached experiment worktrees, were restored, and are not part of the pushed diff.

| Mutant | Detector | Observed assertion |
| --- | --- | --- |
| node-native-context | TestEditCacheSeparation | edit_test.go:38: lowering Node hit=false want true |
| node-runtime-path | TestEditCacheSeparation | edit_test.go:38: runtime Node hit=false want true |
| native-context | TestEditCacheSeparation | edit_test.go:41: lowering native result hit after changed identity |
| native-runtime | TestEditCacheSeparation | edit_test.go:41: runtime native result hit after changed identity |
| node-harness | TestNodeHarnessIdentity | edit_test.go:57: changed runner execution reused Node harness |
| lowering-source | TestLoweringSourceEdit | edit_test.go:130: one-byte lowering edit did not change native identity |
| runner-location | TestRunnerLocationIdentity | edit_test.go:180: relocating identical runner bytes invalidates observations: 8efb7a49915a349ee1e56514220b0ff830b6208b49c6278ebd5827f6bd68042a b9a974135f20fa302c8a1b13ffcbdce9f236ceebd8b05223e9b671e33d000432 |

All mutants exited 1 on the intended test assertion, were restored, and were followed by passing checks. Raw failure logs remain under /tmp/test262-followup/mutant-*.log. No new raw logs are committed.
