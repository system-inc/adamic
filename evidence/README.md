# Census evidence

- `base/` comes from `44361d277f655f96613a60a7d3c2e8516ce61d2c`.
- `revised-head/` comes from `bfe663fba078d96362cabd80801f6a8ec225a116`.
- Each directory contains the surviving directory census in `census.jsonl.gz`, the entry header in `latent.jsonl.gz`, and the entry unit timings in `latent-times.jsonl.gz`. The recovered singleton runs are in `recovered/`. Their first record is the per-root result.
- `adapted-src.sha256` records the source contents of the single shared adapted tree. `upstream.txt` records the upstream commit and Node version. The compiler directories record the Go version, submodule pins, and census binary digests.
- `interruption.txt` records the process states after the environment interruption. Neither interrupted invocation recorded an exit code.
- Run `python3 evidence/compare-interrupted.py REPORT.md` from the report checkout to regenerate the report and `comparison.json`.

## Rerun

Create worktrees named `base` and `revised-head` under a fresh absolute scratch path at the two commits above. Initialize each with `git submodule update --init --recursive --depth 1`. Clone the upstream input with `git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git "$scratch/upstream"`, as in the meter.

Create `evidence/base`, `evidence/revised-head`, and `tmp` under scratch. Copy `run.sh` into `$scratch/evidence` and put the recorded Go toolchain and Node on `PATH`. Run `bash "$scratch/evidence/run.sh" "$scratch" prepare` once. For each label, run the same script with `"$scratch" build "$label"` and then `"$scratch" measure "$label"`.

For a recovered root, run `CENSUS_MODULE_FORMAT=CommonJSBundle "$scratch/$label-bin/census" "$scratch/adapted/$root" "$output"` in the pinned worktree. The program emits the singleton case first and repeats it as its whole-program case.
