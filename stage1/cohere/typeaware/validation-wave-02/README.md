# Wave 02 evidence and reproduction

Base: `0d540f413625f016f20fea39761c7b184f335de6`.
Cohere: `715ba94f3608a6500086b1076ce5cb7e51b836db`.
typescript-go: `8d550c837c90bd1805b047b7eeccc2baac2d5e7a`.
TypeScript: `050880ce59e30b356b686bd3144efe24f875ebc8` (v6.0.3).

The two portable manifests are unchanged copies of validation-coverage's frozen
77 compiler files and 287 repository files. New source files do not enlarge the
repository population. Full output hashes include absolute file headings, so
reproduction at a different path changes hashes but must retain byte agreement.

From the repository root, after `bash cloud/setup.sh` and sourcing its environment,
with `ADAMIC_TYPESCRIPT_SOURCE` pointing to the pinned TypeScript checkout:

```sh
mkdir -p /tmp/wave02
python3 - <<'PYTHON'
import os
from pathlib import Path
r = Path.cwd()
c = Path(os.environ['ADAMIC_TYPESCRIPT_SOURCE']).resolve()
e = r / 'stage1/cohere/typeaware/validation-wave-02'
for name, root in [('repository', r), ('compiler', c)]:
    Path('/tmp/wave02/' + name + '.manifest').write_text(''.join(
        str(root / p) + '\n' for p in (e / (name + '.manifest')).read_text().splitlines()))
PYTHON
ADAMIC_WAVE02_ARTIFACTS=/tmp/wave02/artifacts \
ADAMIC_WAVE02_REPOSITORY_MANIFEST=/tmp/wave02/repository.manifest \
ADAMIC_WAVE02_COMPILER_MANIFEST=/tmp/wave02/compiler.manifest \
go test ./stage1/cohere/typeaware -run '^TestWave02AgreementAndMutants$' \
    -count=1 -v -timeout 30m > /tmp/wave02/wave-test.log 2>&1
python3 stage1/cohere/typeaware/validation-wave-02/checker_mutants.py /tmp/wave02 \
    > /tmp/wave02/checker-mutants.log 2>&1
python3 bridge/tsgo/profile/volume_bench.py /tmp/wave02/artifacts/coverage \
    /tmp/wave02/artifacts/coverage-oracle /tmp/wave02/timing --rounds 3 \
    --corpus compiler "$ADAMIC_TYPESCRIPT_SOURCE/src/compiler/tsconfig.json" /tmp/wave02/compiler.manifest \
    --corpus repository "$PWD/tsconfig.json" /tmp/wave02/repository.manifest \
    > /tmp/wave02/timing.log 2>&1
```

The wave test compiles actual `.a` sources, checks full findings/fixes/suggestions,
runs ASan/UBSan/LSan builds, runs one compiling source mutant per rule, and checks
released handles with a compiling registry mutant. Its Go oracle imports the
unmodified production cohere rules through an overlay; no lint logic is copied
into that oracle. Timing uses quiet, alternating count-only runs and excludes
builds; every count must agree.

The pinned cohere CLI predates `.a` support. `formatter.go.txt` is a tooling adapter
for its embedded Prettier oracle, using a virtual `.ts` filename while writing
the actual `.a` files. For the recorded lint gate, temporary copies changed only
extensions and import paths; the repository compiler options, native prelude and
cohere:typescript rules were retained. All six new modules passed type, lint and
format checking. No adapter or generated fixture changes the repository's source
extension promise.
