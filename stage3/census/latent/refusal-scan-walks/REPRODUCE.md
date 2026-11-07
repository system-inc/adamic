# Reproduce the scan profiles

Run from the compiler checkout at the report's base, with the normal cloud
setup and stage3 API dependencies installed. All Go probe sources here are
`.go.txt`; only scratch files become Go compilation inputs. No `internal/`
file is written. `prepare.py` excludes ordinary lower tests in its scratch
manifest because the census overlay removes the output-producing helpers
those tests reference. It adds only `TestTask2bz5m7vRefusalScanProfile`, whose
first statement is `t.Parallel()` and which runs no Go subprocess.

```sh
source /workspace/adamic-tools/env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source
stage3/apply.sh --build-product /workspace/2bz-adapted > /workspace/scratch/2bz-apply.log 2>&1
python3 stage3/census/latent/make_overlay.py "$PWD" /workspace/scratch/2bz-plain > /workspace/scratch/2bz-overlay.log 2>&1
python3 stage3/census/latent/refusal-scan-walks/prepare.py "$PWD" /workspace/scratch/2bz-plain > /workspace/scratch/2bz-prepare-plain.log 2>&1
go test -c -buildvcs=false -overlay=/workspace/scratch/2bz-plain/overlay.json -o /workspace/scratch/2bz-plain.test ./internal/lower > /workspace/scratch/2bz-build-plain.log 2>&1
python3 stage3/census/latent/make_overlay.py "$PWD" /workspace/scratch/2bz-counts > /workspace/scratch/2bz-overlay-counts.log 2>&1
python3 stage3/census/latent/refusal-scan-walks/prepare.py "$PWD" /workspace/scratch/2bz-counts counts > /workspace/scratch/2bz-prepare-counts.log 2>&1
go test -c -buildvcs=false -overlay=/workspace/scratch/2bz-counts/overlay.json -o /workspace/scratch/2bz-counts.test ./internal/lower > /workspace/scratch/2bz-build-counts.log 2>&1
python3 stage3/census/latent/refusal-scan-walks/run.py "$PWD" /workspace/scratch /workspace/2bz-adapted stage3/census/latent/refusal-scan-walks/evidence > /workspace/scratch/2bz-profiles.log 2>&1
python3 stage3/census/latent/refusal-scan-walks/summarize.py stage3/census/latent/refusal-scan-walks/evidence > /workspace/scratch/2bz-summary.log 2>&1
```

Regenerate an overlay before preparing it again in counts mode. The two modes
use the same generated census code. Plain mode adds only the test harness.
Counts mode also rewrites lowering's `ForEachChild` call sites through a
counter and wraps each registry check. `FileWalks` means calls whose receiver
is a SourceFile node, including walks that short circuit. `ChildWalks` counts
calls at all nodes; `Children` counts callbacks entered. These are lowering
walk counts, not additional walks inside checker implementations. Site names
identify scratch source function and line before instrumentation.

The outer shared-visitor/preflight bucket has walk counters only; its call and
time columns in rules.csv are blank. Per-rule elapsed times include counter
overhead and any nested checks. They
are inclusive, so do not sum them as disjoint CPU buckets. Plain profiles start
`runtime/pprof` immediately before the scan, after the complete entry program
has been loaded and checked. Evidence serialization follows the scan; scan wall and
process CPU timers stop before that serialization. The pprof scope also
includes serialization and release. Each run uses fresh lowering state and
the same entry-closure enumeration as the replay worker. It attempts no units
and imports no backend. Plain and counted finding hashes must agree.
