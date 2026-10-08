# JSON corpus deadlines and pin

The harness counts progress from the child's stdout or stderr. Startup allows 30 minutes until its first byte and reports the command as `never started` if no byte arrives. After the first byte the guard allows 300 seconds without output, then reports `stalled without output`. A six-hour overall cap is a last-resort backstop. Cancellation kills the child's process group; output readers retain the original five-second wait bound.

All 2,496 cases were measured individually under ASan with `ADAMIC_JSON_GUARD_CALIBRATE=1`. Calibration includes process startup and ran with five workers. The slowest case under deliberate load was `cohere/internal/lint/rules/tailwind/collapse/testdata/classorder_fixtures.json`, 24.473 seconds. The fixed 300-second stall interval provides 12.3 times that observed bound. An earlier unloaded 2,492-case measurement found the same case at 21.601 seconds, 13.9 times headroom. The interval is pinned, never chosen from runtime timing. Calibration is opt-in and does not change agreement checks.

The sanitized and Linux LeakSanitizer runs each split the cases into `min(runtime.NumCPU(), 8)` contiguous chunks. Each phase runs its chunks concurrently and joins stdout in input order. Every chunk first agrees byte for byte with its corresponding slice of the mandatory single-process release output; the joined output then agrees with that entire release output and the Go oracle. The original release run, Node source, JavaScript backend and all three original port mutants remain mandatory. Leak detection remains enabled for the separate leak phase. macOS retains its original `leaks --atExit` check for every chunk; that platform was not tested here.

# Corpus identity

`testdata/corpus.pin` pins 2,496 identities, before any gate selection. Each identity is the repository-relative path and SHA256 of the text bytes. The aggregate digest is SHA256 of the sorted `path<TAB>text-hash<LF>` list, with the same digest and a count per group. Every corpus run logs its groups and rejects a missing, extra or changed input by test name, group and path. No gate switch rewrites the pin. Updating a corpus requires reviewing and updating the identities and aggregate/group digests together.

| Group | Pinned count |
| --- | ---: |
| Adamic | 1420 |
| TypeScript | 203 |
| cohere | 839 |
| generated | 34 |
| Total | 2496 |

Aggregate SHA256: `d783351f69ade8b7fc7ce6bd9ae3605d09f6f2f0ebfa90f8e0e1322f3aa6ebb5`.

The pin is anchored to main `855d114e` and cohere `7945d102`. This task branch was updated to that main: its prior ancestor lacked four tracked meter JSON reports. The source tree at 855d114e contains 1,402 tracked Adamic JSON inputs. Provisioning `stage3/api/node_modules` adds exactly 18: `.package-lock.json`, three package manifests (`@types/node`, `typescript`, `undici-types`), thirteen localized TypeScript diagnostic JSON files and `typesMap.json`. They account exactly for 2,478 versus 2,496. The API's committed lock selects TypeScript 6.0.3 and @types/node 25.3.3; `npm ci --prefix stage3/api` provisions this group. A box without those dependencies must fail rather than pass on the shorter corpus.

The fetched Threadripper whole-gate log for 855d114e reports 1,420/203/839 plus 34 generated, matching this pin. Removing those 18 identities reproduces 2,478, Adamic 1,402, and a named failure listing every missing path. The short box's own inventory was not available; its missing files are inferred from this exact tracked/provisioned accounting and reproducer. Its owner can confirm by comparing the named pin diagnostics.

# Validation

Box: AMD EPYC 7763, `nproc` 5, four-CPU cgroup quota, 16 GiB memory; Go 1.27.1, Node 24.19.0, clang 20.1.8. Setup environment is `/workspace/adamic-tools/env.sh`, with pinned Prettier 3.9.6. Test output stayed in local files.

Commands use `go test -v -count=1 -timeout 30m ./stage1/cohere/json`, with `ADAMIC_JSON_PRETTIER` pointing to the pinned install. Both sample variables are unset for these measurements. The after runs also enable `ADAMIC_JSON_GUARD_CALIBRATE=1`, so their wall times include additional isolated-case calibration absent in the baseline.

| Whole JSON package | Wall time | Inputs |
| --- | ---: | ---: |
| Prior unsharded baseline on this box | 455.054s | 2492 |
| Sharded, unloaded, including calibration | 242.280s | 2492 |
| Sharded and pinned, one concurrent `yes` CPU load, including calibration | 278.041s | 2496 |

The loaded run's port comparison took 271.37s and upstream report 84.31s, with the exact nine known discrepancies. The load process was killed and reaped after the run. These are package measurements, not a claim about Threadripper timing; the four additional meter inputs and calibration prevent treating the final pair as an identical-work benchmark.

Kill/check proofs:

- A silent child fails `sh: never started`; a child which prints then hangs fails `sh: stalled without output`. Test policies use short intervals; production bounds remain as above.
- Continuing stdout or stderr output survives beyond the startup interval. A continuing-output fixture is still killed by the overall backstop.
- A planted one-byte change in chunk 1 fails with `beta.json byte 3`, through the same join/comparison path used by the corpus.
- All three original port mutants remain caught: missing final newline, missing colon space, wrong filename parser. The three upstream printer mutants also remain caught.
- Compiling scratch guard mutants which lose the first-output observation or fail to age the stall interval are caught behaviorally. An earlier stall mutant failed to build because it made a variable unused; that result is excluded and the corrected mutant was rerun.
- Pin tests catch missing/extra identities and changed text, preserve sorted-order determinism, reproduce the full 18-input shortage, and prove pin validation precedes sampling. The fake SHA and changed dependency path select 79/2462 physical files, retain that changed path and all 34 generated cases, and still validate the full 2496 pin.

Package vet, gofmt and diff checks pass. Linux is covered; macOS and a Threadripper rerun are not. No compiler source, deadline producer or cloud setup changes are part of this unit.
