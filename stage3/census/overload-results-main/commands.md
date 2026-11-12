# Single-main census commands

Measured main is frozen at `219dbb6e7e20ea99d3e5d4b61c0cb14102dc9205`. The report branch starts at the preceding report commit `481552841c9b5e1496fc6237da10501b707178d0`; no main-source commit or scratch composition is merged into that branch. Only this directory is committed. All observations here are fresh.

Fetch main, create `/workspace/orm-main` detached at that pin, and verify with `git merge-base --is-ancestor` that all four landed changes in `pins.json` are ancestors. Fetch and submodule/install children use 1,700-second limits; all other potentially long children use `run.py`'s 1,740-second limit.

```sh
git fetch origin main
git worktree add --detach /workspace/orm-main 219dbb6e7e20ea99d3e5d4b61c0cb14102dc9205
# In /workspace/orm-main:
git -c submodule.cohere.url=https://github.com/system-inc/cohere.git submodule update --init --recursive --depth 1
source /workspace/adamic-tools/env.sh
npm ci --prefix stage3/api
```

The existing pristine `/workspace/op2-upstream` checkout is reused only as the immutable TypeScript input. Its SHA is verified against measured main's `stage3/source.json` (`050880ce59e30b356b686bd3144efe24f875ebc8`, TypeScript 6.0.3). No earlier adapted output or observation stream is reused.

Set `scripts=/workspace/op2-report/stage3/census/overload-results-main`. From measured main, run:

```sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/op2-upstream
python3 "$scripts/run.py" /tmp/orm-apply.log /tmp/orm-apply.time.json bash stage3/apply.sh /workspace/orm-adapted
bash "$scripts/build.sh" main
bash "$scripts/build-scan.sh" main
```

Both executable variants derive from main's own `make_overlay.py` and `entry_overlay.py`. The only additional selector in the second variant conditionally omits repeated refusal scans. The production output guards remain active.

Copy the retained `fixture/*.ts` to `/workspace/orm-fixture`. Run `python3 -B "$scripts/fixture_check.py" main` and `main scan`. The fixture compares an unsplit census with two complementary shards, including a real overloaded callback-value stop and cross-shard generic call witnesses. Both variants pass. The second variant also matches the original binary's unsharded unit/finding records. Build the ordinary guarded census with `latent-overlay/overlay.json` and run main's existing `stage3/census/latent/audit.py`; its no-output, loader, attribution and continuation checks pass.

The normalization/pair-tool tests are run with `python3 -E -S -B -m unittest discover -s stage3/census/pair -p 'test_*.py'`. All 20 pass, including the deliberately red metadata mutant. Running those tests is not a paired census.

## Measurements

Each item is a separate process/session. Keep at most two active children, as in the preceding census's per-side schedule. No persistent orchestration process spans the complete run.

```text
measure.sh main 0
measure.sh main 1
measure-scan.sh main 2
measure-scan.sh main 3
measure-scan.sh main 4
measure-scan.sh main 5
measure-scan.sh main 6
measure-scan.sh main 7
```

Invoke with `bash "$scripts/SCRIPT" main SHARD`. The immutable wrappers set `LATENT_FULL=0`, `LATENT_ASSERT_NO_OUTPUT=1`, a per-unit `LATENT_TIMES` stream, `CENSUS_SHARDS=8`, `CENSUS_SHARD=N` and `GOMEMLIMIT=3GiB`; later shards set `CENSUS_SKIP_SCAN=1`. All read `/workspace/orm-adapted/src/tsc/tsc.ts`. Dependencies and registration stay complete. Every timing/shell wrapper reports its child's own exit code and propagates it. Each child is terminated at 1,740 seconds with at most ten seconds to kill, below 30 minutes.

Merge all eight streams using `merge.py OUTPUT INPUT0 ... INPUT7`. Every attempted unit must occur once. Shards 0 and 1 retain full refusal scans and must agree exactly; later shards contain no scan findings. Repeated scan and deferred/skipped records deduplicate. Generic witnesses union under the original Go order and five-witness cap. Relation/cache reports remain process-scoped and are not added; overflow or global findings stop composition.

## Unpaired analysis and report

```sh
python3 -B "$scripts/analyze.py"
node "$scripts/inventory.cjs" /workspace/orm-build-main/latent.jsonl "$scripts/main-inventory.json"
python3 -B "$scripts/archive.py"
python3 -B "$scripts/report.py"
python3 -B "$scripts/validate.py"
```

These are run through `run.py`, retaining each child's own status and resource measurements. **No pair or paired delta is computed.** `analyze.py` sends matching source/unit positions to measured main's unmodified `stage3/census/pair/normalize.cjs`, capturing its actual child status. Its request and complete identity response are retained.

The selection regexes are anchored at the beginning: `^overload \d+ of (\S+) result(?:\s|$)` and the analogous `parameter` prefix. Every matching observation from every phase is retained. Type-parameter diagnostics are not silently treated as parameter diagnostics. Raw unique identities are `(kind, source path/hash/byte offset, exact reason, full text)`; distinct identities are `(source path/hash/outermost contiguous access offset, exact reason)`, exactly as in the earlier census tool. Per-attempt observations and first-stop-unit counts are also retained to check the handoff without conflating repeated observations with unique sites.

`overload-findings.json` lists every observation and every unique raw/distinct representative, with complete reason/text and observation references. `summary.json` groups both families by function and checks every supplied result count and the supplied parameter expectations. Listed-but-absent and unlisted-but-present result functions are explicitly named. `inventory.cjs` independently verifies every top-level unit and source order. Validation independently recomputes prefix inclusion, all four counts, first stops, timings, exact shard recomposition, fixture results, source hashes and archive hashes.

The previous report is used only for the numeric 45/3,187 comparison and for an input-manifest comparison. Any changed adapted files are listed in `input-difference-from-78933e52.json`; no previous stream is used as an input. Logs and JSONL are gzip compressed without changing bytes (`gzip -dc FILE.gz` reads them). Go evidence uses `.go.txt`. Compiler binaries and scratch source commits are not committed.
