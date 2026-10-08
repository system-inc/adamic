Run from fresh origin/area/stage3 checkout b61e70642652a7f23af113278ce2fc357d7b4646.

Command after sourcing /workspace/adamic-tools/env.sh and selecting Node 24.19.0 first on PATH:

```sh
export TMPDIR=/workspace/stage3-meter-tmp
STAGE3_METER_COMPILER=per-ref STAGE3_METER_RUNS=$PWD/stage3/meter/runs bash stage3/meter/twice-daily.sh > /tmp/stage3-meter-landing-retry.log 2>&1; echo "exit=$?"
```

Meter exit=0. Both source refs are their own compiler refs. Ordinary binaries
embed the exact pinned SHAs and vcs.modified=false, captured separately under
main/ and area/. Latent binaries were built with each source's overlay and
-buildvcs=false; the meter initialized each source's own recorded submodules.
There were no changes outside stage3/ between the two pinned compiler refs.

Setup: GOPROXY='https://proxy.golang.org|direct' bash cloud/setup.sh.
Setup retry exit=0, total 335.951s, nproc=5, Node v24.19.0. Full timing lines
and the initial disk failure and meter failure are retained in setup.log,
setup-first-attempt.log, recovery.md and failed-attempt/. The retry used
workspace scratch storage; a further preserved prior tree was relocated
to /tmp/stage3-meter-preserved-enum to keep build headroom on both disks.

Validation with CENSUS_BINARY and LATENT_CENSUS_BINARY set to this run's
selected binaries, TMPDIR=/workspace/stage3-meter-tmp, output always in logs:

- Main: python3 -m unittest discover -s stage3/meter -p '*test.py', 23 passed.
- Area: python3 -m unittest discover -s stage3/meter -p 'report_test.py', 20 passed.
- Compiler selection: python3 -m unittest discover -s stage3/meter -p 'compiler_test.py', 3 passed.
- Wrong-compiler mutant: METER_SCRIPT_UNDER_TEST=/tmp/stage3-meter-landing-wrong-compiler.sh PYTHONPATH=stage3/meter python3 -m unittest compiler_test.CompilerSelectionTests.test_per_ref_builds_and_runs_each_pinned_compiler, exit=1 with a latent compiler-witness assertion.

The real dependency-error mutant changes whole-program 2 -> 0 and own-file
2 -> 1, affecting only the dependency's own-file attribution. The real latent
mutant plants one NotYet in target: only that reason rises by one, with
Refused unchanged. Both real probes passed using each ref's built binaries.

Integration delivery: branch stage3-meter/20261008T000754Z.hFDR7Q starts at
origin/main f4efdd2369311d1420aa53fdf5c1a55bdda811d4. Its one commit adds
only this run directory. No compiler, adaptation, meter code, or other file
is changed. Raw census and latent JSONL/gzip streams remain scratch data.
