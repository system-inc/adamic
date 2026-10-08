Built: download every manifest's complete module graph with private modfiles, preserving committed sources and checksum verification.
Implementation: bb7feea2b7d875be3759491aab64b42243db763f, on devtools/setup-warm-modules from origin/area/developer-tools 358dbccd95a909a849d1d9db370ae7122a87fc5b.
Proof: origin/main ce0750f28ef3943057f1f852b3ae5d93e6c5d644 passes the requested offline lint tests and go build ./... with no download message.
Mutants: removing regexp2 v1 and v2 independently fails offline naming each; 17 installer mutants caught, including workspace inheritance and omitted all; real unused-transitive fixture catches omitted all.
Limits: this is a reused box with initially empty module caches, not a fresh VM; full gate not run. Old no-argument download already lists both regexp2 versions here, so r2's exact cache omission was not reproduced.

The installer recursively enumerates every repository/submodule go.mod, including tool and testdata modules: 22 on this main checkout. It uses go mod download -json all rather than a package dependency list or the no-argument default. Every download and verify runs with GOWORK=off. Shim manifests require workspace-local modules at v0.0.0; private copies of go.mod/go.sum retain each explicit replacement and add missing local replacements derived from the same checkout. Neither real manifests nor workspace files are edited. Existing proxy fallback, Go sum checks and go mod verify remain. The existing stamp still covers all manifest/workspace bytes and absence, Go version, Go environment, helper implementation and every downloaded artifact's metadata; all downloaded graphs contribute protected paths.

The actual empty-cache setup used:
```sh
source /workspace/css-gate-tools/env.sh
ADAMIC_TOOLS=/workspace/css-gate-tools GOMODCACHE=/workspace/warm-modules-cache GOCACHE=/home/agent/.cache/go-build GOFLAGS=-trimpath ADAMIC_SETUP_REPOSITORY=/workspace/main bash /workspace/adamic/cloud/setup.sh
```
It downloaded and verified 22 graphs in 31.889s; complete setup took 57.419s. Final serial warm setup took 8.981s; the module step skipped in 0.393s. The earlier overlapping warm run is excluded because the standalone proof temporarily changed the module manifest.

Offline proof (normal committed workspace; module lookup and sumdb networking disabled):
```sh
source /workspace/css-gate-tools/env.sh
export GOMODCACHE=/workspace/warm-modules-cache GOCACHE=/home/agent/.cache/go-build
export GOPROXY=off GONOPROXY=none GOPRIVATE= GOSUMDB=off GONOSUMDB='*' GOWORK=auto GOTOOLCHAIN=local
export GOFLAGS='-mod=readonly -trimpath' ADAMIC_GATE_UNCACHED=1
cd /workspace/main
go test -count=1 -run 'TestEmittedJavaScriptMismatch|TestRulesAgree' ./stage1/cohere/lint > /tmp/adamic-gate/warm-modules-proof/offline-workspace-lint.log 2>&1
go build ./... > /tmp/adamic-gate/warm-modules-proof/offline-build.log 2>&1
```
Both selected tests pass, package time 84.298s. The build succeeds with zero stdout/stderr bytes. Neither positive log contains go: downloading. The harness also rejects any oracle-build stderr, so a hidden download cannot satisfy its ordinary success.

Go rejects -mod=mod in workspace mode. The initial GOWORK=off/-mod=mod attempt rewrote main's proof go.mod, created go.sum, and ended with only FAIL [build failed] lines. Those edits were preserved in /tmp/adamic-gate/warm-modules-proof/standalone-go.mod and standalone-go.sum, then restored; no compiler change was made. -mod=readonly with the existing workspace is the valid proof mode. GONOPROXY=none prevents direct-fetch exemptions, GOSUMDB=off prevents checksum-network requests, and GOTOOLCHAIN=local prevents toolchain fetching. The offline removal mutants prove missing modules fail loudly.

Removing both versions failed with:
```text
github.com/dlclark/regexp2@v1.11.5: module lookup disabled by GOPROXY=off
```
Restoring v1 while keeping v2 absent failed with:
```text
github.com/dlclark/regexp2/v2@v2.5.2: module lookup disabled by GOPROXY=off
```
Both tests fail in each stable removal run. Source and download entries were moved outside the unit-owned cache to recoverable backups, then restored and verified. Go's readonly directory permissions required temporarily adding owner write during the moves; original permissions were restored. A first negative command started before removal succeeded and passed; it is not counted as a mutant kill. The subsequent stable removals above are the evidence.

The real local-proxy fixture has no imports but requires a module that itself requires another unused module. Both are downloaded. Dropping all fails the nested-source assertion. It verifies a warm skip, uncached rerun and unchanged real go.mod with no generated go.sum. Normal setup tests: 51 passed, six opt-in integration checks skipped; actual module suite with ADAMIC_SETUP_INTEGRATION=1: seven passed. mutate_setup_modules.py catches all 17 independent mutants.

Cached/uncached preparation protects 156 module archives and 24,313 artifact files. All protected files' names, bytes and modes produce the same SHA256 before and after ADAMIC_GATE_UNCACHED=1:
```text
ba635a799046eae263227df978e5ee1a0b2a60d0cd6771acda5113d5fe51ee51 24313 artifact files; 156 module archives
```

Timings use identical committed copies of all 22 main manifests/workspaces in a scratch tree, avoiding benchmark changes to the checkout being tested. Module download reads manifests, not package source. The old helper is exactly 358dbccd's version; the new helper is the implementation commit above. Each variant gets an empty GOMODCACHE, then repeats warm. Order is before/after, after/before, before/after. Downloads retain checksum verification and the proxy fallback. Only each timing driver's temporary cache is disposed after its measurements.

| Loop | Cold before | Cold after | Warm before | Warm after |
| --- | ---: | ---: | ---: | ---: |
| 1 | 18.956s | 23.818s | 0.164s | 0.365s |
| 2 | 16.293s | 26.279s | 0.164s | 0.315s |
| 3 | 18.265s | 27.179s | 0.215s | 0.315s |

Best cold 16.293s -> 23.818s (+7.525s); best warm 0.164s -> 0.315s (+0.150s).

Exact measurement commands, cache modes, commit SHAs, nproc, cpu.max, Go/clang/Node versions and before/after loads are in timings.json. The instrument is:
```sh
source /workspace/css-gate-tools/env.sh
PYTHONDONTWRITEBYTECODE=1 python3 /tmp/adamic-gate/warm-modules-proof/measure-modules.py
```
Each inner command is python3 <helper> /tmp/adamic-gate/warm-modules-proof/manifest-snapshot <fresh-tools-directory>; the full per-run paths are recorded in timings.json.

All timing rows: installer bb7feea2b7d875be3759491aab64b42243db763f; main ce0750f28ef3943057f1f852b3ae5d93e6c5d644; nproc=5; cpu.max=400000 100000; Go 1.27.1 linux/amd64; clang 20.1.8 (LLVM 87f0227cb60147a26a1eeb4fb06e3b505e9c7261); Node v24.19.0; cached=yes (fresh prefixes cold, matching stamps warm). Load pairs appear in every timings.json record.

Main and all initialized submodules are clean after final setup. No module-host Forbidden failure occurred in the successful setup/measurement runs, so the requested retries were not needed. Final full-setup timing output is below:
```text
setup: go ready (0.079s)
setup: submodules ready (0.126s)
setup: node v24.19.0 skipped (validated version, checksum and installed bytes); step-duration=0.160s
setup: clang ready (/workspace/css-gate-tools/llvm/bin/clang) (0.276s)
setup: node ready (0.275s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.010s
setup: markdown dependencies ready (0.334s)
setup: stage3 API dependencies skipped (validated API lock and installed bytes); step-duration=0.105s
setup: stage3 API dependencies checked (0.487s)
setup: module dependencies skipped (validated manifests and downloaded module cache); step-duration=0.393s
setup: module dependencies ready (0.569s)
setup: go build ready (8.497s)
setup: test binaries deferred (use --warm-tests) (8.831s)
setup: build cache warm (8.835s)
setup: workspace sums restored to the commit (8.920s)
setup: build-flags commit=ce0750f28ef3943057f1f852b3ae5d93e6c5d644 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gate-inputs=false gate-archive=false load-before=0.19 0.81 1.02 1/1171 299993 load-after=0.98 0.97 1.07 1/1172 300826
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (8.981s)
setup: source /workspace/css-gate-tools/env.sh
setup: logs /tmp/adamic-gate/setup.5fgzdF
setup: node v24.19.0
```
