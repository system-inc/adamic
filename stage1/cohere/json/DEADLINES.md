# JSON corpus deadlines and pin

The harness counts progress from the child's stdout or stderr. Startup allows 30 minutes until its first byte and reports the command as `never started` if no byte arrives. After the first byte the guard allows 300 seconds without output, then reports `stalled without output`. A six-hour overall cap is a last-resort backstop. Cancellation kills the child's process group; output readers retain the original five-second wait bound.

All 2,496 cases were measured individually under ASan with `ADAMIC_JSON_GUARD_CALIBRATE=1`. Calibration includes process startup and ran with five workers. The slowest case under deliberate load was `cohere/internal/lint/rules/tailwind/collapse/testdata/classorder_fixtures.json`, 24.473 seconds. The fixed 300-second stall interval provides 12.3 times that observed bound. An earlier unloaded 2,492-case measurement found the same case at 21.601 seconds, 13.9 times headroom. The interval is pinned, never chosen from runtime timing. Calibration is opt-in and does not change agreement checks.

The sanitized and Linux LeakSanitizer runs each split the cases into `min(runtime.NumCPU(), 8)` contiguous chunks. Each phase runs its chunks concurrently and joins stdout in input order. Every chunk first agrees byte for byte with its corresponding slice of the mandatory single-process release output; the joined output then agrees with that entire release output and the Go oracle. The original release run, Node source, JavaScript backend and all three original port mutants remain mandatory. Leak detection remains enabled for the separate leak phase. macOS retains its original `leaks --atExit` check for every chunk; that platform was not tested here.

# Corpus identity

At cohere `f5d1934a`, the pinned total is 1,108: cohere has 853 identities,
TypeScript 203, generated 34 and provisioned 18. The pin was regenerated through
the test's corpus helpers; non-cohere groups are unchanged. The table and aggregate
digest below record the earlier pin-design unit, rather than the current checkout.

In the original pin-design unit, `testdata/corpus.pin` pinned only 1,094 upstream, generated and provisioned identities, before any gate selection. Each identity is the repository-relative path and SHA256 of the text bytes. The aggregate digest is SHA256 of the sorted `path<TAB>text-hash<LF>` list, with the same digest and a count per pinned group. Every corpus run rejects a missing, extra or changed pinned input by test name, group and path. No gate switch rewrites the pin.

Tracked repository JSON inputs follow Git HEAD instead of a stored identity pin. The walker must include every path returned by `git ls-files --cached -z -- '*.json'`, excluding cohere and `.git` trees. It must contain no duplicate path. The index must match HEAD and the worktree must match the index for those paths. Both comparisons are necessary: comparing worktree directly with HEAD can hide a staged change whose working file was restored. Missing, dirty and unexpected untracked inputs fail with their paths named. Git must be usable and its root must match the walk root; exports without Git metadata fail rather than fall back to walking.

| Group | Count at the pin-design checkout | How checked |
| --- | ---: | --- |
| repository | 1516 | Current Git paths and clean checkout, no stored count or identity pin |
| provisioned | 18 | Identity pin |
| TypeScript | 203 | Identity pin |
| cohere | 839 | Identity pin |
| generated | 34 | Identity pin |
| Pinned total | 1094 | Count and aggregate identity pin |
| Full corpus | 2610 | Repository plus pinned groups, before sampling |

Pin-design aggregate SHA256 (historical): `3d987e9a388d0e354db2ee0dc5b6ce8a780f39a3e30d41e52bedc9973fc1e000`.

The branch starts at integration tip `ddaef2cb`, with main `74fb6490`, cohere `7945d102` and its pinned TypeScript checkout. Step 1 was pushed separately as `27472be539b5bb2f2455138a71edc0db496742a7`: it refreshed the old whole-corpus pin to 2,610 inputs (Adamic 1,534, TypeScript 203, cohere 839, generated 34). Its three requested tests passed in 72.672s, with the full upstream comparison and exact nine discrepancies. The second commit removes 1,516 repository identities from the pin without removing those files from the formatter comparisons.

The 18 provisioned inputs are under `stage3/api/node_modules`: `.package-lock.json`, three package manifests, thirteen localized TypeScript diagnostic JSON files and `typesMap.json`. The API's committed lock selects TypeScript 6.0.3 and @types/node 25.3.3; `npm ci --prefix stage3/api` provisions this group. The batch6 box lacked these files, explaining the former 2,478 versus 2,496 shortage. The existing witness still removes all 18 and proves every missing path is named; the provisioned group reports zero versus 18 even as repository landings add more inputs.

Only paths explicitly named in the provisioned pin are admitted as provisioned. An additional untracked JSON, including one inside node_modules but outside those identities, fails as unexpected. A tracked JSON added by a later landing is included automatically through Git, with no pin update. Gatesample runs only after both Git validation and all pinned identity checks; sampling retains its original controls and changed-path inclusion.

# Validation

The following timing measurements record the earlier deadline and sharding unit before this pin-design change.

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

# Pin design proofs

`TestRepositoryCorpusMutants` runs the real walker and Git checks in private temporary repositories. Named catches include a dropped `tracked.json`, an added `unexpected.json`, an unlisted node_modules input, missing `stage3/api/node_modules/typescript/package.json`, a one-byte change to `cohere/upstream.json`, dirty `tracked.json`, and a staged change whose working file was restored. The last probe demonstrated a false green with the initial combined diff; separate index/worktree checks caught it after correction.

`TestRepositoryLandingWithoutPinEdit` creates a later HEAD only in its private scratch history. `landing-added.json` raises its repository count from two to three, while the pinned identities and serialized pin remain unchanged. No fixture JSON or scratch Git history is committed to the task branch. `TestRepositoryRequiresGit` proves an export without `.git` fails explicitly. The main checkout has usable Git metadata and a clean tracked JSON set.

`TestCorpusPinRejectsMissingProvisionedInputs` retains the full 18-input witness. `TestCorpusPinBeforeSampling` compares the selected size against the current full corpus, instead of a stale 2,496 constant. The formatter agreement, native/leak checks, benchmark opt-in and original three port mutants are unchanged by this design unit.

Correct-tip setup reported Node ready 0.024s, Go ready 0.025s, submodules ready 0.065s, Markdown dependencies ready 0.095s, clang ready 0.171s, Go build ready 34.334s, build-cache warm 34.588s and total 34.618s, on `nproc` 5 with a four-CPU cgroup quota. Detailed setup and test logs remain outside the checkout. Final validation: all requested step-1 tests passed in 72.672s. The unsampled whole JSON package passed over all 2,610 inputs in 252.508s. After the staged-cleanliness correction, all focused pin/repository checks passed in 2.537s and the entire sampled package passed in 39.852s. The full Git/path/identity checks run before selection in both modes; all formatter mutants and controls remain full. Final vet, gofmt and diff checks are clean. The whole-repository gate and macOS were not run.
