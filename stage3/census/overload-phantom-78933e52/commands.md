# Composition and census commands

This run reuses the bounded first-error census tools from `stage3/census/overload-proof-3a7ea290` at report commit `353218cbd01286116bd97fc1475b51bc54cdce75`. All measurements, logs and adapted output here are fresh. The report branch remains based on that prior report commit; none of the scratch merge commits are added to its ancestry.

Main was fetched and frozen at `0c7052759671347775615af5b4a8028dd8688af8`. Every merge used `git merge --no-ff --no-edit`, without conflict resolution. The runner recorded each merge child's own status. All four merges returned 0. Exact parents and tree hashes are in `composition.json` and the four `preparation/oph-{base,head}-merge-*.log.gz` / `.time.json` records.

```sh
git fetch origin main typescript/path-components-interface-3 optimizer/overload-proof-widened-4 optimizer/overload-proof-reachable-2 optimizer/overload-phantom-result-2
git worktree add -b scratch/census-phantom-base-78933e52 /workspace/oph-base 0c7052759671347775615af5b4a8028dd8688af8
# In /workspace/oph-base:
git merge --no-ff --no-edit a3f67b011b1a810c282da390f80a3180df1be34b
git merge --no-ff --no-edit 3a7ea290fca9e3f3325ffa8206a861b22f441583
# Base composition: be9e9cc6f77121cdb6970a18915c7c61cdb65b78
git worktree add -b scratch/census-phantom-head-78933e52 /workspace/oph-head HEAD
# In /workspace/oph-head:
git merge --no-ff --no-edit 694834e95c3dd79bc4d350f4560d5a16b175f0ad
git merge --no-ff --no-edit 78933e52b0abd0da2fce91e73b4b71c1b36dc4e4
# Head composition: 65af0b47a56f444ae54cfc2a8761e63b20f73f8e
```

Both head commits are one commit on `3a7ea290`. Their changed paths are disjoint, and neither they nor the composed head diff change adaptation or meter inputs. Checked paths: `stage3/source.json`, `stage3/apply.py`, `stage3/apply.sh`, `stage3/api`, `stage3/adapt`, plus `stage3/census` and `stage3/meter`. The empty adaptation diff is retained. The source trees remain tracked-clean.

Initialize each composition's exact recursive submodules using `git -c submodule.cohere.url=https://github.com/system-inc/cohere.git submodule update --init --recursive --depth 1`. Install `npm ci --prefix stage3/api` at the base. Submodule and install commands use 1,700-second limits and report their own child status. Reuse the installed toolchain from `/workspace/adamic-tools/env.sh` and the pristine TypeScript 6.0.3 checkout `/workspace/op2-upstream`, verified at `050880ce59e30b356b686bd3144efe24f875ebc8`.

Set `scripts=/workspace/op2-report/stage3/census/overload-phantom-78933e52`. From the **base composition**, run:

```sh
source /workspace/adamic-tools/env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/op2-upstream
python3 "$scripts/run.py" /tmp/oph-apply.log /tmp/oph-apply.time.json bash stage3/apply.sh /workspace/oph-adapted
```

This includes adaptation 40 from a3f67b01. `provenance.py` checks that the actual adapted source has `PathPathComponents<T extends Path = Path> extends Array<T>`, retaining its declaration and file hash in `path-components-interface.json`.

Run each of `bash "$scripts/build.sh" base`, `head`, then `bash "$scripts/build-scan.sh" base`, `head`. Each overlay is generated using **its own composed tree's** `make_overlay.py` and `entry_overlay.py`. The second variant adds only the conditional omission of repeated refusal scans. All original production output guards stay active.

Copy the retained `fixture/*.ts` to `/workspace/oph-fixture`. Run `python3 -B "$scripts/fixture_check.py" SIDE` and `SIDE scan` for both sides. Each variant compares an unsharded fixture with two complementary shards, preserving complete unit/finding records, a real overloaded callback-value stop, and cross-shard generic call witnesses. The second variant also compares its unsharded records against the original binary. All four checks pass.

The ordinary guarded census is separately built with each side's `latent-overlay/overlay.json` and checked with that composition's existing `stage3/census/latent/audit.py`. Both audits pass. The base composition's unmodified pairing tests are run using `python3 -E -S -B -m unittest discover -s stage3/census/pair -p 'test_*.py'`; all 20 pass, including the expected failing metadata mutant.

## Bounded measurements

Every item below is a **separate process/session**. No persistent orchestration process spans the complete census. Run at most two shards per side concurrently and launch the next shard as a slot finishes.

```text
measure.sh base 0        measure.sh head 0
measure.sh base 1        measure.sh head 1
measure-scan.sh base 2   measure-scan.sh head 2
measure-scan.sh base 3   measure-scan.sh head 3
measure-scan.sh base 4   measure-scan.sh head 4
measure-scan.sh base 5   measure-scan.sh head 5
measure-scan.sh base 6   measure-scan.sh head 6
measure-scan.sh base 7   measure-scan.sh head 7
```

Invoke these with `bash "$scripts/SCRIPT" SIDE SHARD`. The immutable wrappers set first-error mode (`LATENT_FULL=0`), `LATENT_ASSERT_NO_OUTPUT=1`, per-unit `LATENT_TIMES`, `CENSUS_SHARDS=8`, `CENSUS_SHARD=N`, and `GOMEMLIMIT=3GiB`. Later shards additionally set `CENSUS_SKIP_SCAN=1`; the two full scan ledgers must agree exactly. The source argument is always `/workspace/oph-adapted/src/tsc/tsc.ts`.

Each `run.py` records its actual child's exit code, own elapsed/CPU time and peak RSS, and exits with that status (using shell encoding for signals). Every shell wrapper reports and propagates its own child's status. Children are terminated at 1,740 seconds with at most ten seconds to kill, below 30 minutes. All census children and wrappers complete with 0; no census child times out. The full project, dependencies and registration graph remain present in every shard.

## Merge, pair and report

Merge the eight successful shards separately per side using `merge.py OUTPUT INPUT0 ... INPUT7`. An attempted unit must occur exactly once. Repeated scans and deferred/skipped records deduplicate. Generic call witnesses union under the original Go ordering and five-witness cap. Process-scoped relation/cache counters remain separate; any overflow or global finding stops composition.

Use the actual **base composition's unmodified pairing tool**:

```sh
python3 /workspace/oph-base/stage3/census/pair/pair.py /workspace/oph-build-base/latent.jsonl /workspace/oph-build-head/latent.jsonl /workspace/oph-adapted --typescript /workspace/oph-base/stage3/api/node_modules/typescript --json "$scripts/delta.json" --report "$scripts/pair-report.md"
python3 -B "$scripts/analyze.py"
node "$scripts/inventory.cjs" /workspace/oph-build-base/latent.jsonl "$scripts/base-inventory.json"
node "$scripts/inventory.cjs" /workspace/oph-build-head/latent.jsonl "$scripts/head-inventory.json"
python3 -B "$scripts/archive.py"
python3 -B "$scripts/report.py"
python3 -B "$scripts/validate.py"
```

These commands are run through `run.py` and retain their logs/stats. `analyze.py` requires exactly 11 base getPathComponents result-stop units and five getSupportedExtensions units. Newly admitted follows the user's criterion: a base stop becomes no stop or a non-overload-result first stop. Exact reasons and full diagnostic text stay in the records; invariant-mutable is detected in the full diagnostic text.

The stock-parser inventory proves that every top-level unit and source order is retained. Validation also checks all timings, disjoint shard membership, exact archived recomposition, both fixtures on both sides, all source/archive hashes, composition pins and cohort sizes. JSONL and logs are gzip compressed without changing their bytes (`gzip -dc FILE.gz` reads them). Go evidence is named `.go.txt`; no compiler binaries or `.go` evidence files are committed. Only this report directory is added to the existing report branch in one commit.
