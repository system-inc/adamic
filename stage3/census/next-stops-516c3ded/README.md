# Measurement recipe

Measured head: 516c3dedf370293d4e4f8f78ab4628d391aed5e4.
Parent for sys.ts: 751879544ea6e627e67c91d6ab3c4ba595be6356.
Baseline and script source: typescript/next-stops-f2546f1f at
2cda42f68690691c4ed080293684e1ff911adcfa.

`measure.py` records bounded commands. `run.py` starts a process group,
waits at most 1780 seconds, terminates or kills as necessary, waits for the
actual exit, and records it. `group.py` runs at most ten roots and three
children concurrently; `run-one.py` verifies the ordinary tool's two
observations are identical except seconds. Every group is itself run by
`run.py`. `checkpoint.py` packages logs and pushes a non-forced checkpoint.

Run CommonJSBundle completely, analyze, resolve emitting source, render,
validate and push the complete bundle before starting EcmaScript. Reuse the
same adapted tree and ordinary binary without apply or build between formats.
`analyze.py`, `resolve-evidence.py`, `write-report.py`, `validate.py`, and
`finalize.py` generate and verify the evidence. Reports include both the
requested baseline and the older 464e5daa table per root.

For sys.ts, use the repository's existing `latent/make_overlay.py` and
`latent/tool` with `LATENT_FULL=1 LATENT_ONLY_FILE=<absolute adapted sys.ts>`
and CommonJSBundle. This overlay is measurement only, disables production
output, and does not change compiler or adaptation files. Limit each full
run to 1500 seconds. If it times out, retain the actual exit and use the
scan-only wrapper at e960ce28d8465ee6f2cc809ef8ddf2b922965a6e from
`typescript/promise-discard-sites` for the scan half; state that lowering
comparison is incomplete if full output is unavailable.

Only this directory is committed. Generated binaries and adapted trees stay
in scratch; artifacts exceeding 20 MB are omitted and identified by sha256.
Go source snapshots under this directory use `.go.txt` names.

The environment reconnected during the original CommonJSBundle 11–20 group,
removing its outer runner. Its children were terminated, its evidence was
archived without a fabricated group exit, and the full group was repeated.
`launch.py` starts only the bounded outer runner in a detached session to
preserve its timeout/exit recording across tool reconnects.
