# Reproduction commands

Pins and absolute paths are recorded in `provenance.json`. The existing report branch was reset to the new base after preserving the tools; all measurements from the superseded pair were discarded. The pristine pinned TypeScript checkout and toolchain caches were reused as inputs, while all compiler builds, adapted-tree output, fixture checks, native checker evidence and measurement streams belong to this new pair. The report worktree is based on 17c6378d. `base` is an untouched worktree at 17c6378d; `head` is an untouched worktree at 3a7ea290. The prior census is 83e07d2f, directory `stage3/census/overload-result-02d48211`; its historical 43 definitions are extracted into `historical-43.json`.

Setup recipe for a fresh checkout (fetch, submodule initialization and npm install use 1,700-second limits; setup, apply and builds use the retained 1,740-second timing wrapper):

```sh
git fetch origin optimizer/overload-proof-widened-4 typescript/overload-result-census main
git worktree add -b typescript/overload-proof-census /workspace/op2-report 17c6378d
git worktree add --detach /workspace/op2-base 17c6378d
git worktree add --detach /workspace/op2-head 3a7ea290
# In each compiler worktree:
git -c submodule.cohere.url=https://github.com/system-inc/cohere.git submodule update --init --recursive --depth 1
# At base:
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
npm ci --prefix stage3/api
# The upstream checkout is pinned by v6.0.3, verified by provenance.py:
git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /workspace/op2-upstream
```

Before measurement, verify the full head diff and the empty adaptation-input diff:

```sh
git diff --stat 17c6378d 3a7ea290
git diff 17c6378d 3a7ea290 -- stage3/source.json stage3/apply.py stage3/apply.sh stage3/api stage3/adapt
```

Set `scripts=/workspace/op2-report/stage3/census/overload-proof-3a7ea290` in a shell. Generate exactly one adapted tree, **from base**:

```sh
cd /workspace/op2-base
source /workspace/adamic-tools/env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/op2-upstream
python3 "$scripts/run.py" /tmp/op2-apply.log /tmp/op2-apply.time.json bash stage3/apply.sh /workspace/op2-adapted
```

Run `bash "$scripts/build.sh" base` and `bash "$scripts/build.sh" head` in separate sessions. The script generates each overlay from that side's own `make_overlay.py`, applies only scratch unit selection, formats it, extends it using that side's own `entry_overlay.py`, and builds the first-error entry tool. `provenance.py` verifies pins, untouched tracked trees, overlay key roots and the identical meter instrumentation; it hashes every shared `src` file.

Copy the two checked-in `fixture/*.ts` files to `/workspace/op2-fixture`. Run `python3 "$scripts/fixture_check.py" base` and then `head`. Each side compares an unsharded fixture with two complementary unit shards, including exact file/unit/finding records and generic call witnesses. The reused fixture includes an overloaded callback value to verify preservation of a real stop; both variants are rerun independently on both new pins. The guarded ordinary census was also built with each side's `latent-overlay/overlay.json` and checked with that side's `stage3/census/latent/audit.py` (see per-side build/audit logs). The base pairing suite was run with `python3 -E -S -B -m unittest discover -s stage3/census/pair -p 'test_*.py'`; its expected failing mutant traceback is part of the passing test log.

## Bounded measurement

Each of the following is a **separate command/session**; no persistent loop or pipeline process spans all shards. Run at most two shards per side at once, then launch the next pair as slots finish:

```text
bash .../measure.sh base 0        bash .../measure.sh head 0
bash .../measure.sh base 1        bash .../measure.sh head 1
bash .../measure-scan.sh base 2        bash .../measure-scan.sh head 2
bash .../measure-scan.sh base 3        bash .../measure-scan.sh head 3
bash .../measure-scan.sh base 4        bash .../measure-scan.sh head 4
bash .../measure-scan.sh base 5        bash .../measure-scan.sh head 5
bash .../measure-scan.sh base 6        bash .../measure-scan.sh head 6
bash .../measure-scan.sh base 7        bash .../measure-scan.sh head 7
```

`measure.sh` clears inherited `LATENT_*` and `CENSUS_SHARD*` variables, then sets exactly `LATENT_FULL=0`, `LATENT_ASSERT_NO_OUTPUT=1`, `LATENT_TIMES`, `CENSUS_SHARDS=8`, `CENSUS_SHARD=N`, `GOMEMLIMIT=3GiB` and the pinned upstream path. Every child loads the full shared `/workspace/op2-adapted/src/tsc/tsc.ts` project. `run.py` records that child's own exit code, elapsed/CPU time and peak RSS; it terminates at 1,740 seconds, with a ten-second kill grace. Shell wrappers print and propagate their own child's code. The overlay selection is after unchanged eligibility and generic deferral, and before any independent attempt. It filters neither dependencies nor the registration graph.

After all eight shards per side finish, invoke `merge.py OUTPUT INPUT0 ... INPUT7` separately for each side. The raw shard JSONL, own-child stats, logs, per-unit timing files and relation reports are retained. Every attempted unit must occur once, repeated scans deduplicate exactly, and deferred witnesses union under the original Go ordering and five-witness cap. Global findings or relation overflows stop composition. There is no synthetic addition of process-scoped relation/cache counters.

## Pair, resolve and analyze

Use the **unmodified base pairing tool**, with outermost-access normalization:

```sh
python3 /workspace/op2-base/stage3/census/pair/pair.py /workspace/op2-build-base/latent.jsonl /workspace/op2-build-head/latent.jsonl /workspace/op2-adapted --typescript /workspace/op2-base/stage3/api/node_modules/typescript --json "$scripts/delta.json" --report "$scripts/pair-report.md"
python3 "$scripts/analyze.py"
node "$scripts/resolve-concatenate.cjs" "$scripts/concatenate-request.json" "$scripts/concatenate.json"
node "$scripts/inventory.cjs" /workspace/op2-build-base/latent.jsonl "$scripts/base-inventory.json"
node "$scripts/inventory.cjs" /workspace/op2-build-head/latent.jsonl "$scripts/head-inventory.json"
```

These commands are run through `run.py`; their logs/stats are retained. The stock checker reports Node-host-type diagnostics, so declaration/index evidence is additionally verified on each side with Adamic's own zero-diagnostic checker. `resolve-native.go.txt` replaces only the virtual main file in a copy of the corresponding entry overlay, then builds the same `./stage3/census/latent/tool` target. It calls `LatentLoad`, expands entry reach, walks concatenate calls, and reads `GetResolvedSignature(...).Declaration()`. It never lowers or emits code. Native call/declaration locations, exact text, indexes and overload counts must agree on both sides and with the five stock resolutions.

`report.py` creates all requested sections from the records, checks both native ledgers agree, checks every measured child completed with exit 0 below 30 minutes, and writes `REPORT.md`. `archive.py` retains logs, timing records and compressed streams and checks shared source hashes again. `validate.py` performs the final evidence checks. All generated Go evidence committed here is named `.go.txt`; scratch overlay `.go` files and binaries are not committed.

## Duplicate-scan optimization

Shards 0 and 1 per side retain complete scans as a repeated control. Before measurements, `build-scan.sh SIDE` used `scan_overlay.py` to clone that side's entry-overlay metadata and replace only the scratch first-error unit file. Its sole additional condition is `if CENSUS_SKIP_SCAN != 1` around the existing `fresh().latentScan(file)` block. The remaining six shards set `CENSUS_SKIP_SCAN=1` and use `entry-census-scan`; registration, eligibility, lowering and records are untouched. The original and optimized executable hashes and overlay maps are both in provenance.

`fixture_check.py SIDE scan` verifies the second variant's unsharded file/unit/finding records against the original binary, then verifies mixed scanning/non-scanning shard composition against that unsharded result. Both sides passed. Final validation additionally requires that the complete scan ledgers from real-project shards 0 and 1 agree exactly and that shards 2–7 contain no scan findings. Thus one complete, independently repeated scan ledger remains in the final stream without paying for all eight repetitions.

Text logs are archived as `.log.gz` without changing their bytes; `gzip -dc PATH.log.gz` reads them. JSONL streams and per-unit timing streams are also gzip compressed.
