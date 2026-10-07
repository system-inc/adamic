# Wave 02 continuation evidence

The original three wave-02 rules were completed and pushed before the
continuation request. The next three were claimed in pushed commit c835aa6d
before any implementation. A later fetch revealed parallel continuation claims;
`claim-collisions.json` records them. The earliest competing commit is wave 23 at
00:43:48 UTC, one second before our commit at 00:43:49. Commit time alone does not
establish push arrival order. Our selection fetch showed no competing claim.
This is explicit merge-coordination evidence, not a claim of exclusive ownership.

Ahra's correction arrived after these three were claimed and their first full
comparison/mutant suite passed. No further rules were claimed. The existing
registration generator and shared test harness were not changed. Native source
modules remain `.a`; standalone wave-owned runners build and test them directly.
The isolated checker dispatcher registration was made before the correction.

Pins and frozen populations are unchanged from validation-wave-02: cohere
715ba94f3608a6500086b1076ce5cb7e51b836db, typescript-go
8d550c837c90bd1805b047b7eeccc2baac2d5e7a, TypeScript v6.0.3
050880ce59e30b356b686bd3144efe24f875ebc8. Portable manifests retain all 77
compiler roots and 287 pre-port repository files. No new wave file enlarges the
population. Corpus zeroes are backed by positive controls and mutants.

From the repository root after setup, sourcing the printed environment, and
setting ADAMIC_TYPESCRIPT_SOURCE to the pinned TypeScript checkout:

```sh
mkdir -p /tmp/wave02-continuation
python3 - <<'PYTHON'
import os
from pathlib import Path
r = Path.cwd()
c = Path(os.environ['ADAMIC_TYPESCRIPT_SOURCE']).resolve()
e = r / 'stage1/cohere/typeaware/validation-wave-02-continuation'
for name, root in [('repository', r), ('compiler', c)]:
    Path('/tmp/wave02-continuation/' + name + '.manifest').write_text(''.join(
        str(root / p) + '\n' for p in (e / (name + '.manifest')).read_text().splitlines()))
PYTHON
ADAMIC_WAVE02_CONTINUATION_ARTIFACTS=/tmp/wave02-continuation/artifacts \
ADAMIC_WAVE02_CONTINUATION_REPOSITORY_MANIFEST=/tmp/wave02-continuation/repository.manifest \
ADAMIC_WAVE02_CONTINUATION_COMPILER_MANIFEST=/tmp/wave02-continuation/compiler.manifest \
go test ./stage1/cohere/typeaware -run '^TestWave02ContinuationAgreementAndMutants$' \
    -count=1 -v -timeout 30m > /tmp/wave02-continuation/wave-test.log 2>&1
python3 bridge/tsgo/profile/volume_bench.py /tmp/wave02-continuation/artifacts/coverage \
    /tmp/wave02-continuation/artifacts/coverage-oracle /tmp/wave02-continuation/timing \
    --rounds 3 \
    --corpus compiler "$ADAMIC_TYPESCRIPT_SOURCE/src/compiler/tsconfig.json" /tmp/wave02-continuation/compiler.manifest \
    --corpus repository "$PWD/tsconfig.json" /tmp/wave02-continuation/repository.manifest \
    > /tmp/wave02-continuation/timing.log 2>&1
```

All complete streams are compressed without changing a byte. Hashes include
absolute path headings, so relocation changes hashes but must preserve Go/native
agreement. The independent Go oracle imports the three unmodified production
Nexus rules through a virtual-main overlay. It imports no bridge code or native
rule judgments. All three production rules intentionally have no fixes or
suggestions; the common serialization still compares both fields exactly.

The pinned cohere CLI predates `.a` support. The clean source gate uses temporary
`.ts` extension/import adapters, the repository's compiler options and native
prelude, and all 276 cohere:typescript rules. The actual Adamic modules remain
`.a`; the native runner compiles them directly. No shared harness is changed.
