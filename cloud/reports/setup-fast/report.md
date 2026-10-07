Built parallel preparation, validated warming stamps, optional test warming, and reproducible timing/mutant runners.
Commits: baseline e011f8f; measured core 8062ddc; final manifest coverage 9d6bdbb.
Commands: all setup trials exit 0; integration 5 tests OK; vet exit 0; filtered uncached oracle ok.
Mutants: 10 omissions killed; HEAD and checksums also failed real setup integration with VCS metadata disabled.
Not covered: full repository gate, ARM64, successful auto download, or final-key all-cold reruns. Snapshot provenance is unobservable.

Warm setup is 0.547 seconds best of three with the final key, versus 13.550 seconds before. The measured parallel/default-test change reduces fully cold setup from 227.005 to 140.486 seconds best of three. These are directory-cold trials on the same box; kernel/CDN/proxy caches were not flushed.

The all-cold comparisons use commit 8062ddc for both scripts. After that experiment, checksum coverage was extended to the workspace and every dependency module in 9d6bdbb. The final warm comparisons use 9d6bdbb for both scripts. cloud/setup.sh itself is byte-identical between those commits; only the key helper and verification/measurement tools changed. Final-key full cold timing was not remeasured, so the cold numbers below belong to 8062ddc.

The baseline script is the exact e011f8f blob, with only its repository location made explicit because the measurement copy lives outside cloud/. Each pair builds the same checkout commit. Order is before/after, after/before, before/after. All six cold runs are below five minutes, so all three pairs were completed. The runner had a scratch cleanup permission failure after the first successful cold baseline; it was fixed and resumed without discarding or overwriting the completed trial.

The step matrix uses the best complete before trial and best complete after trial in each regime, rather than mixing the fastest individual steps. Parallel phase durations overlap and must not be summed. LLVM download and probe are substeps of clang preparation. Command spans come from EPOCHREALTIME bash traces; streamed curl/tar is reported as a pipeline, not separate overlapping process times.

| Step | Cold before | Cold after | Warm before | Warm after |
|---|---:|---:|---:|---:|
| Go install / version / auto resolution | 3.604s | 5.862s | 0.006s | 0.009s |
| Clang preparation | 27.651s | 30.553s | 0.137s | 0.143s |
| LLVM download and extract | 27.541s | 30.441s | not run | not run |
| Sanitizer compile and run | 0.091s | 0.093s | 0.115s | 0.120s |
| Node | 1.523s | 2.207s | 0.008s | 0.007s |
| Recursive submodules | 12.468s | 28.459s | 0.032s | 0.038s |
| Go dependency validation | not run | 106.761s | not run | 0.252s |
| go build ./... | 107.380s | 2.959s | 2.892s | skipped |
| go test -count=1 -run '^$' ./... | 74.356s | deferred | 10.453s | deferred |
| Total | 227.005s | 140.486s | 13.550s | 0.547s |

Build flags for every timing in the matrix are the selected records below; complete per-trial flags, loads and exact commands are in cold-timings.json and warm-timings.json. Cold tools are absent before setup, so their installed versions are recorded afterward. Both sides of every cold pair installed the same Go/clang/Node versions.

- cold before, loop 1: commit=8062ddc5fbf25ca989e2d50fc1b4105e0e1ef9b1, nproc=5, cpu.max=400000 100000, go=go version go1.27.1 linux/amd64, clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261), node=v24.21.0, cached=empty-action-cache, load before=3.68 2.03 1.54 1/171 21231, load after=10.00 5.39 2.93 1/177 26096.
- cold after, loop 1: commit=8062ddc5fbf25ca989e2d50fc1b4105e0e1ef9b1, nproc=5, cpu.max=400000 100000, go=go version go1.27.1 linux/amd64, clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261), node=v24.21.0, cached=empty-action-cache, load before=1.59 3.72 2.60 1/176 26125, load after=6.01 4.83 3.17 1/187 28675.
- warm before, loop 2: commit=9d6bdbb9b56209032121c1ff266eb01066bcb59b, nproc=5, cpu.max=400000 100000, go=go version go1.27.1 linux/amd64, clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261), node=v24.19.0, cached=action-cache, load before=5.05 2.57 3.56 1/238 54790, load after=5.07 2.71 3.59 1/239 55481.
- warm after, loop 3: commit=9d6bdbb9b56209032121c1ff266eb01066bcb59b, nproc=5, cpu.max=400000 100000, go=go version go1.27.1 linux/amd64, clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261), node=v24.19.0, cached=action-cache, load before=4.96 2.76 3.60 1/240 56186, load after=4.96 2.76 3.60 1/241 56394.

| Regime / loop | Before | After | Instrument (exact commands) |
|---|---:|---:|---|
| cold / 1 | 227.005s | 140.486s | before: `bash -x /workspace/setup-trials-hfv9d3_w/before-1/before.sh`; after: `bash -x /workspace/setup-trials-86bm_izh/after-1/repository/cloud/setup.sh` |
| cold / 2 | 242.600s | 143.128s | before: `bash -x /workspace/setup-trials-86bm_izh/before-2/before.sh`; after: `bash -x /workspace/setup-trials-86bm_izh/after-2/repository/cloud/setup.sh` |
| cold / 3 | 250.013s | 144.394s | before: `bash -x /workspace/setup-trials-86bm_izh/before-3/before.sh`; after: `bash -x /workspace/setup-trials-86bm_izh/after-3/repository/cloud/setup.sh` |
| warm / 1 | 14.900s | 0.548s | before: `bash -x /workspace/setup-trials-myl07xgn/before-1/before.sh`; after: `bash -x /workspace/adamic/cloud/setup.sh` |
| warm / 2 | 13.550s | 0.550s | before: `bash -x /workspace/setup-trials-myl07xgn/before-2/before.sh`; after: `bash -x /workspace/adamic/cloud/setup.sh` |
| warm / 3 | 13.769s | 0.547s | before: `bash -x /workspace/setup-trials-myl07xgn/before-3/before.sh`; after: `bash -x /workspace/adamic/cloud/setup.sh` |

Reproduce the comparative trials:

```sh
source /workspace/adamic-tools/env.sh
python3 cloud/measure-setup.py --output /tmp/setup-warm-trials
python3 cloud/measure-setup.py --cold all --output /tmp/setup-cold-trials
```

Every trial log contains timestamped bash -x output, the setup timing lines, build flags before/after and a monotonic wall time. The runner retains logs and JSON before removing only its own named cold scratch directory. Downloaded module directories require restoring directory write permissions before cleanup. Resuming requires the same checkout commit and cold mode.

What is cached: only the completed Go warming work. Go list -deps -export validates/rebuilds real dependency artifacts first. With --warm-tests it also validates test dependencies. The stamp keys HEAD, root/workspace/dependency manifests and checksum files including absence, go version, go env including flags, setup/helper contents, dependency BuildIDs/export state, Go cache state, and warming mode. Go cache action IDs/output IDs/size are keyed; their refreshed access timestamps are not. Go cache data-file size/mtime/ctime and all entry names notice partial removal as well as go clean. A racing cache deletion falls back to a build rather than trusting a key. Successful stamps are atomically published after completion.

Dirty/untracked source files, embeds and changed local replacements are held by Go’s actual dependency validation rather than a HEAD-only shortcut. Existing Go and Node must answer a version check; clang is compiled and executed through the sanitizer probe on every invocation; Git checks out the recorded gitlinks on every invocation. No installation or sanitizer-result cache was added, so they have no new cache-key mutants. The live sanitizer check costs milliseconds and avoids trying to summarize changing system/sanitizer runtime state.

The filtered shallow clone checked out exactly the same gitlinks on both sides: cohere 715ba94f3608a6500086b1076ce5cb7e51b836db, TypeScript 8d550c837c90bd1805b047b7eeccc2baac2d5e7a. Clone duration was longer while overlapping LLVM extraction; this experiment does not isolate a benefit from blob filtering, and none is claimed. Cold elapsed preparation follows the slowest parallel task instead of the sum.

Test warming is optional because the baseline spends many seconds linking/starting the no-test runs even with compilation artifacts already cached, and much more on a cold cache. A worker testing one package does not need that cost for every other package. --warm-tests restores warming every test package; the real integration checks both its first warmup and its subsequent stamp skip. ADAMIC_GATE_UNCACHED=1 bypasses the new stamp and uses go build -a (and go test -a when requested). The integration compares cached/forced-uncached executable bytes, stdout, and env.sh bytes and requires equality. Timing and skip diagnostic lines intentionally differ.

Repository invalidation proof: this checkout has no root go.sum. After a warm skip, a newline-only go.sum was temporarily created, setup printed go build ready, and the file was removed to restore its original absence. The restored input also reran the step. The real integration separately changes an existing go.sum, changes only HEAD, changes an embedded file, introduces an invalid untracked Go file, and removes an actual export artifact. Its GOFLAGS=-buildvcs=false ensures HEAD/checksum checks do not pass incidentally through VCS metadata in compiler actions.

Verification commands and observed outputs:

```sh
ADAMIC_SETUP_INTEGRATION=1 python3 cloud/test_setup.py > /tmp/setup-fast/integration-expanded.log 2>&1
# Ran 5 tests; OK
python3 cloud/setup-mutants.py --integration > /tmp/setup-fast/mutants-complete.log 2>&1
# 10 omitted-key/manifest mutants killed; head/sums also killed by real setup integration
go vet ./... > /tmp/setup-fast/vet.log 2>&1
# exit 0; no diagnostics
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/oracle -run 'TestNativeAgreesWithNode/dedication/dedication.a' > /tmp/setup-fast/oracle.log 2>&1
# ok github.com/system-inc/adamic/internal/oracle
gofmt -l cmd internal > /tmp/setup-fast/gofmt.log
# empty
bash -n cloud/setup.sh
git diff --check
# both exit 0
```

| Mutant | Check that failed |
|---|---|
| Drop HEAD | test_every_component_invalidates; real setup wrongly skipped HEAD-only change |
| Drop checksum/manifest digest | test_every_component_invalidates; real setup wrongly skipped go.sum change |
| Drop Go version | test_every_component_invalidates |
| Drop Go environment | test_every_component_invalidates |
| Drop dependency artifacts | test_every_component_invalidates |
| Drop Go cache state | test_every_component_invalidates |
| Drop warming mode | test_every_component_invalidates |
| Omit root go.sum collection | test_collected_manifests_invalidate |
| Omit go.work.sum collection | test_collected_manifests_invalidate |
| Omit dependency go.sum collection | test_collected_manifests_invalidate |

Each mutant exited 1 at its intended assertion, not a syntax or warning failure. Sources were never mutated in the checkout: each variant lives in a private scratch directory and the test loader selects that module. Individual failure logs are retained in evidence/. The final implementation passed the same checks.

Initial observations and timing caveats: after fetching main to e011f8f, the first requested manual run printed Go ready 0s, clang ready 1s, Node ready 1s, submodules ready 1s, build cache warm 122s, done 122s. Tools and initialized submodules were already present, so it was not a cold tool install. An early timer wrapper failed because /usr/bin/time is absent and its trace overlapped that first compilation; those exploratory numbers are not comparative measurements. The subsequent timestamped baseline printed done 13s (monotonic 13.453s). The reported conclusions use the clean, serial, alternating-order best-of-three trials above, not the exploratory numbers. Exploratory build flags: commit=e011f8f60899586d6373a5ccb07335ad82cfbf3c, nproc=5, cpu.max=400000 100000, go=go version go1.27.1 linux/amd64, clang=clang version 20.1.8, node=v24.19.0, cached=pre-existing tools/submodules and partial action cache, load before/after=unrecorded. Initial exploratory load averages cannot be reconstructed.

Auto-toolchain limitation: no successful auto SDK download occurred in the setup comparisons because fresh setup downloaded Go 1.27.1, which already satisfies go.mod’s go 1.27. A separate permitted Go 1.26.0 bootstrap experiment selected go1.27.0 and failed when proxy.golang.org redirected the SDK zip to storage.googleapis.com: Forbidden. It is a network-policy failure, not a successful timing to compare. The failed probe’s flags, command, load and exit are in evidence/auto-timings.json, with the public artifact’s temporary signed redirect redacted. The permitted direct dl.google.com Go 1.27.1 installation succeeded in all fully cold trials. Successful auto downloading requires that redirect destination to be permitted; no destination-policy bypass was attempted.

Snapshot observation: Go, LLVM, Node, submodules and some Go cache artifacts existed before the first manual run. Stamps and artifacts persist across repeated shells in this container. That is consistent with a prepared environment or prior setup, but neither snapshot creation nor image restore/transfer is visible here. Original environment startup stdout was not accessible. These measurements cover setup execution, not provisioning/image transfer, and do not establish whether a new worker restores the stamp or cache inode timestamps. If restored metadata differs, the conservative key can rerun warming; Go still validates dependency actions.

Limits: no full repository test gate, no ARM64 trial, no successful automatic SDK-download timing, no original bootstrap timing, and no final-key full cold rerun. The final key expansion was validated by real integration and final warm comparisons. SDK download/unpack and broad Go dependency compilation still dominate cold setup; this work does not remove that necessary compilation or claim a zero-cost fresh tool install.
