Setup now installs the independent Markdown width oracle with exact npm pins and checked-in SHA-512 integrity, and exports ADAMIC_MARKDOWNWIDTH_DEPS in the generated env.sh. The directory is outside /root, including the ADAMIC_TOOLS fallback. Source the env.sh printed by setup before running tests.

The Node background preparation installs these dependencies after Node becomes available, overlapping Go, LLVM and submodules. A fresh box does not need npm: the installer downloads npm 11.9.0, verifies its separately pinned SHA-512 archive integrity, then runs npm ci with an empty private npm cache and scripts disabled. Installation is staged and published only after success; npm errors preserve the previous working installation. No global npm installation or pre-existing npm cache is used.

The installer captures immutable manifest bytes once and uses those same bytes for both the key and installation. A real concurrent-edit test changes the checkout during the download; the installed lock must still equal its keyed snapshot. Global/dry-run/install-strategy settings are explicitly controlled by npm arguments. The npm warming key contains package-lock.json, package.json, npm-bootstrap.json, installer source and Node version. Warm hits also hash installed file bytes, paths and modes, including directory permissions, and reject unexpected symlinks. ADAMIC_GATE_UNCACHED=1 bypasses the stamp and reinstalls; real integration requires cached and uncached installation bytes to match. The Go warming key additionally collects all three manifests and the installer, so a changed lockfile reruns that step as requested.

Dependency installation cost, best of three alternating-order runs on the same box and commit: **cold 0.636259s; warm 0.035869s**. Cold means no installed dependencies, npm bootstrap or npm cache for this step, with Node already ready. Kernel/proxy/CDN caches were not flushed. Wall times include Python process startup, bootstrap download/verification/extraction, npm ci and publication; warm includes process startup and actual byte validation. Before this change the step was absent and the independent oracle failed, so its former cost is not a successful zero-second installation.

| Loop | Cold installation | Warm validation | Instrument (exact command) |
|---|---:|---:|---|
| 1 | 0.636259s | 0.038007s | cold: `python3 /workspace/adamic/cloud/setup-markdown-width.py /workspace/adamic/cloud/markdown-width /tmp/adamic-gate/markdown-timings-qctgkia2/cold-1 /workspace/adamic-tools/bin/node`; warm: `python3 /workspace/adamic/cloud/setup-markdown-width.py /workspace/adamic/cloud/markdown-width /tmp/adamic-gate/markdown-timings-qctgkia2/cold-1 /workspace/adamic-tools/bin/node` |
| 2 | 0.795597s | 0.035869s | cold: `python3 /workspace/adamic/cloud/setup-markdown-width.py /workspace/adamic/cloud/markdown-width /tmp/adamic-gate/markdown-timings-qctgkia2/cold-2 /workspace/adamic-tools/bin/node`; warm: `python3 /workspace/adamic/cloud/setup-markdown-width.py /workspace/adamic/cloud/markdown-width /tmp/adamic-gate/markdown-timings-qctgkia2/cold-1 /workspace/adamic-tools/bin/node` |
| 3 | 0.811621s | 0.045875s | cold: `python3 /workspace/adamic/cloud/setup-markdown-width.py /workspace/adamic/cloud/markdown-width /tmp/adamic-gate/markdown-timings-qctgkia2/cold-3 /workspace/adamic-tools/bin/node`; warm: `python3 /workspace/adamic/cloud/setup-markdown-width.py /workspace/adamic/cloud/markdown-width /tmp/adamic-gate/markdown-timings-qctgkia2/cold-3 /workspace/adamic-tools/bin/node` |

Build flags for every dependency timing above: commit=b13f759028c2703c5b4d54acff8627e6cf7d0be7, nproc=5, cpu.max=400000 100000, go=go version go1.27.1 linux/amd64, clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261), node=v24.19.0. The cached status and before/after loads for each record are:

- cold / 1: cached=false, load before=2.70 4.69 3.21 1/287 78129, load after=2.70 4.69 3.21 1/287 78156.
- warm / 1: cached=true, load before=2.70 4.69 3.21 1/287 78165, load after=2.70 4.69 3.21 1/287 78176.
- warm / 2: cached=true, load before=2.70 4.69 3.21 1/287 78185, load after=2.70 4.69 3.21 1/287 78196.
- cold / 2: cached=false, load before=2.70 4.69 3.21 1/287 78205, load after=2.70 4.69 3.21 1/287 78232.
- cold / 3: cached=false, load before=2.70 4.69 3.21 1/287 78241, load after=2.70 4.69 3.21 1/287 78268.
- warm / 3: cached=true, load before=2.70 4.69 3.21 1/287 78276, load after=2.70 4.69 3.21 1/287 78287.

The full already-fast setup was also measured before/after on this checkout at the same commit, alternating before/after, after/before, before/after. Before is the exact 4e71837 setup.sh blob with only its measurement-copy repository path made explicit; both scripts use this checkout's helper and compiler sources. Both use warm Go action caches. These trials overlapped the separate fresh-checkout proof and setup integration, so their loads are not isolated and the small difference is not a speed claim. Full per-run flags, loads and commands are in warm-setup-timings.json and timestamped traces in evidence/.

| Loop | Before | After | Instrument (exact command) |
|---|---:|---:|---|
| 1 | 0.638270s | 0.593173s | before: `bash -x /workspace/setup-trials-e8wcw_93/before-1/before.sh`; after: `bash -x /workspace/adamic/cloud/setup.sh` |
| 2 | 0.600886s | 0.616308s | before: `bash -x /workspace/setup-trials-e8wcw_93/before-2/before.sh`; after: `bash -x /workspace/adamic/cloud/setup.sh` |
| 3 | 0.609547s | 0.613351s | before: `bash -x /workspace/setup-trials-e8wcw_93/before-3/before.sh`; after: `bash -x /workspace/adamic/cloud/setup.sh` |

Reproduce measurement and installer checks:

```sh
source /workspace/adamic-tools/env.sh
python3 cloud/measure-markdown-setup.py /tmp/markdown-step-final
python3 cloud/measure-setup.py --before 4e71837 --output /tmp/markdown-warm-final
ADAMIC_SETUP_INTEGRATION=1 python3 cloud/test_markdown_setup.py > /tmp/markdown-install-tests.log 2>&1
python3 cloud/markdown-setup-mutants.py > /tmp/markdown-install-mutants.log 2>&1
ADAMIC_SETUP_INTEGRATION=1 python3 cloud/test_setup.py > /tmp/markdown-existing-integration-final.log 2>&1
go vet ./... > /tmp/markdown-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/oracle -run 'TestNativeAgreesWithNode/dedication/dedication.a' > /tmp/markdown-oracle.log 2>&1
```

Observed installer checks: six tests passed, including real lock/manifest/bootstrap changes, installed-byte corruption, forced-uncached byte equality and npm EINTEGRITY rejection without replacing the previous installation. Existing setup integration also passed all five tests. Shell and Python syntax checks passed; git diff --check passed. go vet ./... exited 0 with no diagnostics; the filtered uncached dedication oracle passed. An earlier concurrent integration observed an extra conservative test-binary rebuild and failed its expected skip assertion; the serial rerun passed. The initial failure log is retained, and no incorrect output was observed. An early permission test changed a file to the mode it already had under this box's umask; the fixture now explicitly sets its starting mode before changing it.

Fourteen independent mutants were run and caught at their intended assertions, not syntax errors:

| Mutant | Check that caught it |
|---|---|
| Drop lock digest | test_each_component_invalidates; real integration wrongly skipped a changed lock |
| Drop package manifest digest | test_each_component_invalidates |
| Drop npm bootstrap digest | test_each_component_invalidates |
| Drop installer source digest | test_each_component_invalidates |
| Drop Node version | test_each_component_invalidates |
| Ignore installed bytes | test_tree_bytes_paths_and_modes |
| Ignore installed file modes | test_tree_bytes_paths_and_modes |
| Accept installed symlinks | test_unexpected_symlink_is_rejected |
| Bypass npm bootstrap integrity | test_integrity_rejects_corruption |
| Omit package.json from Go warming stamp | test_collected_manifests_invalidate |
| Omit package-lock.json from Go warming stamp | test_collected_manifests_invalidate |
| Omit npm-bootstrap.json from Go warming stamp | test_collected_manifests_invalidate |
| Omit installer from Go warming stamp | test_collected_manifests_invalidate |
| Reread the lock after computing its key | Real concurrent-edit integration: installed lock bytes differ from the keyed snapshot |

The actual repository proof ran bash cloud/setup.sh warm, appended a newline to cloud/markdown-width/package-lock.json, reran setup, restored the exact original bytes, and reran setup. The warm invocation skipped both npm installation and Go warming. The changed and restored lock each caused npm installation and go build ready. Original and restored lock SHA-256 is 6c4de5ee8e3aa54ab20c2a25de3c1e36387d7216380bfd000bd089b0071598db; the changed hash is 0dad3ff26fac67e3d06f08858f8cd068223c36e6e45a8792727cc20a2b423852. Complete timing/build-flags lines are in evidence/markdown-lock-*.log; lock-proof.json records the hashes, commands and loads.

The full test reads ADAMIC_MARKDOWNWIDTH_DEPS, defaults to /tmp/adamic-markdown-width, and imports three index.js files directly from its node_modules. There is no README beside that test; its REPORT.txt documents the original manual npm install, and width_library.mjs shows the exact imports. The initial installed-dependency run passed the exhaustive Unicode width test and all three native mutants. With the variable unset, the same test exited 1 with ERR_MODULE_NOT_FOUND for /tmp/adamic-markdown-width/node_modules/emoji-regex/index.js. The fallback directory did not exist. Both complete outputs are retained; these are correctness runs, not performance claims.

Fresh provisioning proof is recorded separately below. This is a new checkout and empty directory state on this same container, not a newly allocated VM, and does not establish kernel/network cache coldness or snapshot behavior.

At 87a9dbe a new checkout was created at /workspace/markdown-fresh-zubr_3s9/repository with PATH=/usr/local/bin:/usr/bin:/bin and absent tools, GOPATH and GOCACHE directories. The submodules were cloned remotely, and setup downloaded its own Go, LLVM, Node and npm bootstrap/packages. proof.json records their initial absence. bash cloud/setup.sh exited 0; the test then sourced exactly /workspace/markdown-fresh-zubr_3s9/tools/env.sh and passed uncached, including all three native width mutants. This is the full directory-cold provisioning proof.

After the snapshot guard was added at b13f759, that checkout fetched the final revision from the local branch, retained the first successful npm installation, and began with no installed Markdown dependencies. Final setup reinstalled them. Every installed file byte excluding the bookkeeping stamp is identical between the first and final installation; final-proof.json records this assertion, and installed-versions.json confirms emoji-regex 10.6.0, get-east-asian-width 1.6.0 and narrow-emojis 0.0.3. The exhaustive uncached width test was then rerun with final setup's env.sh.

Final result at b13f759: PASS, 1,216,287 widths from 928 repository files and generated scalar/sequence cases. The East Asian width, narrow emoji and ASCII DEL native mutants all reported caught by observable width output. The complete final output is evidence/fresh-width-final.log. The unset-variable rerun at the same revision still exited 1 with ERR_MODULE_NOT_FOUND; evidence/markdown-width-unset-final.log retains that failure.

Commands for both positive runs:

```sh
bash cloud/setup.sh > setup.log 2>&1
source /workspace/markdown-fresh-zubr_3s9/tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/markdownblocks -run '^TestMarkdownUnicodeWidths$' -count=1 -timeout 30m -v > width.log 2>&1
```

The fresh setup logs retain all cumulative timing lines and full build flags. Their Node version is v24.21.0; the comparative step trials use v24.19.0 within every pair. Fresh SDK installation, module caches and test binaries were initially absent, while the final-revision rerun reused the downloaded Go/LLVM/Node and Go action cache. The final rerun began with absent npm dependencies. Cold provisioning timings are evidence from one correctness proof, not a new best-of-three performance comparison.

Limits: the full repository gate was not run; coverage is the requested exhaustive width test, real setup integration, installer proofs, fourteen cache/integrity mutants, vet, the filtered uncached dedication oracle and shell/Python syntax checks. ARM64 and an actual new VM were not tested. Tools and caches already existed in the primary checkout; original image/snapshot creation is not observable.

Built pinned, integrity-checked npm preparation with a validated millisecond warm skip.
Commits: merged origin/main using a merge; implementation 87a9dbe and snapshot guard b13f759, evidence committed separately on devtools/setup-fast.
Commands: installer and existing setup integrations, vet, filtered uncached oracle and final width test passed; unset-variable run failed as required.
Mutants: fourteen killed, including real lock invalidation; changed and restored repository lock reran both warming steps.
Not covered: full repository gate, ARM64 or new VM provisioning; directory-cold provisioning proof is separate.
