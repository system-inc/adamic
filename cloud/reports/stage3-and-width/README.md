# API npm preparation and Markdown oracle preflight

Built conditional integrity-checked API npm preparation and an early missing-module harness check.
Implementation starts from ac92bf0 after merging origin/area/developer-tools; the report commit is the branch tip.
Named Node library tests fail without declarations and pass after setup; full Markdown width test and go vet pass.
All 11 cache mutants and 3 harness guard omissions are caught; uncached installation preserves installed bytes.
No fresh VM, full gate, ARM run, or 30-minute hang reproduction; scratch integration merges remain uncommitted.

## API installation

`prepareNode` now runs the API installer after Node and Markdown npm preparation, in parallel with Go, clang and submodule preparation. If stage3/api/package-lock.json is absent it prints a skip. Otherwise pinned npm 11.9.0 is downloaded with its SHA512 verified and runs `npm ci --prefix stage3/api --ignore-scripts` with the package lock's integrity checks and an empty private download cache. No package versions or stage3 implementation files were changed on this branch.

The stamp includes lockfile, package manifest, npm bootstrap metadata, both helper implementations and Node version. Warm validation also hashes actual installed filenames, bytes, modes, directories and internal symlink targets. A damaged installation is repaired. Manifests changed during npm ci prevent publication of the stamp. The outer Go warming key also collects the API manifests and helper. `ADAMIC_GATE_UNCACHED=1` bypasses the install stamp; integration checks compare the installed byte tree before and after. See api-integration/ for install, warm, changed-lock, uncached, corruption repair and EINTEGRITY failures.

origin/main c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06 already contains @types/node 25.3.3. The area/developer-tools base has a lockfile but only TypeScript 6.0.3. The named tests were absent from both main and the fetched stage3 branch. To exercise the real named tests, the detached scratch worktree at /workspace/stage3-setup-proof used this uncommitted merge:

```
git merge --no-commit --no-ff origin/area/stage3 origin/library/merge-host-node-types
```

Proof refs: stage3 e71f3490d0d1803e5206ef8fec614a1a48b1e6ae (contains 638b79fdd5ec35a5bcdac58cb8593c9aa1744cd1); library 1a0b0f77efda168ceb43d65f0c0086beeeaae9a2. Neither merge is committed to devtools/setup-fast. After deleting that scratch seat's own node_modules, both commands failed with the actual missing @types/node 25.3.3 diagnostic. After copying the current setup helpers and running `bash cloud/setup.sh`, both passed:

```
go test ./internal/load -run '^TestNodeLibraryUsesPinnedDeclarations$' -count=1 -timeout=30m -v
go test ./internal/lower -run '^TestNodeLibrary' -count=1 -timeout=30m -v
```

See stage3-{load,lower}-{before,final}.log. Package times after installation: 0.515s and 2.657s. The fresh API install in actual setup took 0.932s; setup's Go build cache also needed warming. This was a fresh npm directory on the existing box, not a fresh VM. The toolchain and Markdown dependencies were already installed.

The real setup lock mutation proof appends one newline to package-lock.json, runs setup, then restores the exact original bytes and runs setup again. The initial API check skips, both changes reinstall, and Go warming reruns. See stage3-lock-*.log/json. The initial Go warming also ran due to conservative cache validation; it is not claimed as a Go warm skip.

## Measurements

Same box, same HEAD ac92bf0abf36b0fa442929b4e9a434445cb15a17, alternating order, three loops. Modified sources are the before/after variants on that HEAD. Best of three API cold cost is 0.976061s and warm cost 0.088331s, including Python process startup and complete installed-tree validation. There was no API step before this change, and missing packages made fresh seats fail; zero previous cost would not represent successful setup. These are incremental step costs, not total setup speedups.

| Loop | API cold after | API warm after | Instrument |
| --- | ---: | ---: | --- |
| 1 | 0.976061s | 0.088331s | python3 cloud/measure-stage3-setup.py /tmp/stage3-step-timings |
| 2 | 0.985658s | 0.099525s | same driver, warm then cold |
| 3 | 1.051090s | 0.100927s | same driver, cold then warm |

Exact spawned helper commands, cold directories and all build flags/load observations are in api-timings/timings.json and the individual logs. Build flags: nproc=5; cgroup cpu.max=`400000 100000`; Go 1.27.1 linux/amd64; clang 20.1.8 LLVM 87f0227cb60147a26a1eeb4fb06e3b505e9c7261; Node v24.19.0. Cold has no API stamp, modules, or npm download cache; warm uses the validated stamp. Load before and after is recorded per observation, not held constant.

| Loop | Missing Markdown modules before | After | Instrument |
| --- | ---: | ---: | --- |
| 1 | 6.766458s | 0.004324s | /tmp/markdown-{before,after}.test -test.run=^TestMarkdownUnicodeWidths$ -test.timeout=30m -test.v |
| 2 | 8.981314s | 0.003983s | same commands, after then before |
| 3 | 12.608837s | 0.013749s | same commands, before then after |

Best before 6.766458s, after 0.003983s. Binary compilation is excluded: Go test binaries were compiled from the original and changed harness. `ADAMIC_MARKDOWNWIDTH_DEPS=/tmp/markdown-empty-deps` points at an empty directory. Tests execute fresh observations against a warm Go action cache. Full flags, commands and load before/after are in markdown-empty-timings.json; same tools and SHA as above. Load rises during other compilation, so these are observed box measurements, not CPU-isolated results.

## Cause and harness change

The reported 30-minute hang did not reproduce. An initial original test failed in 4.403s. An instrumented original harness reported corpus generation complete at 0.129s, Go oracle compilation complete at 0.858s, the Go width oracle complete at 11.207s, then Node's missing import at 11.273s. The main observed delay was the Go oracle processing the entire corpus *before* attempting the Node imports. This phase trace is diagnostic evidence, not another best-of-three benchmark; load was not captured for it.

The Node driver imports emoji-regex before reading the corpus, does not read stdin, and exits in 0.031901s even with an empty stdin pipe whose writer remains open. The harness's exec helper supplies no stdin and uses bounded execution and WaitDelay. The evidence does not support an stdin/pipe hang on this box; it cannot rule out a different historical timeout or machine condition.

Only harness lines in width_test.go changed: resolve the existing dependency location first and stat all three imported index.js paths before corpus generation or any oracle work. A missing package fails (never skips) and names the module, path, and setup/env remedy. The imports and all actual correctness comparisons remain unchanged. With modules present:

```
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/markdownblocks -run '^TestMarkdownUnicodeWidths$' -count=1 -timeout=30m -v
```

PASS: 1,264,328 width observations, 998 repository files, plus the existing three native correctness mutants; package time 162.310s. This passing observation is not a performance comparison. Each of the three modules was removed independently in a fixture. The correct harness fails with its early named diagnostic; omitting that module's guard via a Go overlay loses the required diagnostic and reaches Node ERR_MODULE_NOT_FOUND. All three omissions were caught. Compile `/tmp/markdown-after.test` with `go test -c -o /tmp/markdown-after.test ./stage1/cohere/markdownblocks` before running `python3 cloud/verify-width-preflight.py`.

## Cache mutants and validation

`ADAMIC_TOOLS=/workspace/adamic-tools ADAMIC_SETUP_INTEGRATION=1 python3 cloud/test_stage3_setup.py`: four tests pass, including real fresh-cache npm, integrity rejection, lock mutation, missing lock, tree repair and uncached equality. `python3 cloud/stage3-setup-mutants.py`: dropping lock, manifest, npm bootstrap metadata, helper bytes, Node version, installed bytes, modes or link targets each fails its intended assertion. Dropping the lock also fails the actual install/skip integration assertion. Omitting each of the three newly collected outer Go-key inputs fails its manifest invalidation assertion. Eleven distinct cache mutants, plus three harness guard omissions, all caught. Individual failures are retained under cache-mutants/ and width-mutants/.

`ADAMIC_TOOLS=/workspace/adamic-tools ADAMIC_SETUP_INTEGRATION=1 python3 cloud/test_setup.py`: all five tests pass when run serially. The first concurrent attempt conservatively rebuilt Go rather than meeting the integration test's expected warm skip; its failure log is retained as setup-two-integration.log. No incorrect cache hit was observed. The serial rerun is setup-two-integration-final.log. `go vet ./...`, `bash -n cloud/setup.sh`, Python compilation and `git diff --check` pass. Test output was written directly to logs.

Uncovered: full repository gate and ARM; fresh VM provisioning and snapshot behavior; reported 30-minute timeout. No setup snapshot boundary is observable from this process. Scratch trees model absent npm dependencies, with installed toolchains and a warm Go cache. Only cloud files and the minimal cohere test harness are committed.

Retained log whitespace-only lines are normalized to satisfy git diff --check; diagnostics and observations are unchanged.
