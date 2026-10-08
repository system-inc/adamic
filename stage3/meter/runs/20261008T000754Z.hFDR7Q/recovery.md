The first setup attempt exited 1 with these final diagnostics:

```text
# github.com/microsoft/TypeScript/tsc/shim/parser
compile: writing output: write $WORK/b249/_pkg_.a: no space left on device
# github.com/microsoft/TypeScript/tsc/internal/printer
compile: writing output: write $WORK/b187/_pkg_.a: no space left on device
```

The first meter attempt, 20261008T000521Z.InRCRQ, exited 1 at the ordinary
main census build. Its build.log is empty; the meter log prints only its run
path. No detailed meter-build error was captured. These produced files are
preserved under failed-attempt/. /tmp was 96% full (379 MB available).

Three completed prior scratch trees were preserved by moving them from
/tmp/adamic-gate to /workspace/stage3-meter-scratch-archive, with symlinks at
the old paths. This freed 5.9 GB on /tmp. The meter was rerun unchanged with
TMPDIR=/workspace/stage3-meter-tmp, where 12 GB was available. Both refs
remained the requested landing pins. Setup was also retried.

Compiler-selection validation was repeated after recovery: all three tests
passed. A scratch mutant replaces each per-ref build source with the worker
checkout. It successfully builds and runs while recording the expected SHA
metadata; the per-ref test catches its actual latent reason:

```text
AssertionError: {'NotYet: compiled with checkout': 1} != {'NotYet: compiled with main': 1}
```

Mutant exit=1, one assertion failure. No repository implementation was changed.
The initial disk-constrained mutant log was truncated and is not counted as a
valid catch; wrong-compiler-mutant.log is the successful post-recovery proof.

Setup retry exit=0; total 335.951s, nproc=5, Node v24.19.0.
The complete timing lines are in setup.log.
