# Coverage evidence

Inventory pin: `73ac2eb0963e1a4166eaa0fbd160203f11dcdbdf`.
Cohere pin: `715ba94f3608a6500086b1076ce5cb7e51b836db`.
typescript-go pin: `8d550c837c90bd1805b047b7eeccc2baac2d5e7a`.
TypeScript compiler corpus pin: `050880ce59e30b356b686bd3144efe24f875ebc8` (v6.0.3).

The count tables were collected before implementing the ports. `ranking.json`
excludes the existing sixteen rules and sorts remaining inventory entries by
compiler plus repository findings descending, with lexical ties. `selection.json`
records the first ten and their upstream production sources. Counts include zeros
for all 204 registered inventory wave names, not just `NeedsChecker` rules.

The manifests are relative to this repository and the TypeScript checkout. They
freeze the populations counted and checked. Materialize absolute manifests for
the runners, because the bridge resolves query paths from its working directory.
From the repository root, with `ADAMIC_TYPESCRIPT_SOURCE` pointing to the pinned
TypeScript checkout and the setup environment sourced:

```sh
mkdir -p /tmp/adamic-coverage-repro
python3 - <<'PY'
import os
from pathlib import Path
repository = Path.cwd()
compiler = Path(os.environ['ADAMIC_TYPESCRIPT_SOURCE']).resolve()
evidence = repository / 'stage1/cohere/typeaware/validation-coverage'
for name, root in [('repository', repository), ('compiler', compiler)]:
    paths = (evidence / (name + '.manifest')).read_text().splitlines()
    Path('/tmp/adamic-coverage-repro/' + name + '.manifest').write_text(
        ''.join(str(root / path) + '\n' for path in paths))
PY
ADAMIC_COVERAGE_ARTIFACTS=/tmp/adamic-coverage-repro/artifacts \
ADAMIC_COVERAGE_REPOSITORY_MANIFEST=/tmp/adamic-coverage-repro/repository.manifest \
ADAMIC_COVERAGE_COMPILER_MANIFEST=/tmp/adamic-coverage-repro/compiler.manifest \
go test ./stage1/cohere/typeaware -run '^TestCoverageAgreementAndMutants$' \
    -count=1 -timeout=30m -v > /tmp/adamic-coverage-repro/coverage-test.log 2>&1
```

`coverage_test.go` builds the native runner and independent Go oracle, compares
full streams on both corpora and controls under normal and sanitizer builds, then
runs ten source mutants and the released-handle probe/mutant. Child commands'
stdout/stderr are separate numbered files in the artifact directory. Committed
`diagnostic-hashes.json` identifies full comparison streams from this run. Huge
compiler outputs are not committed; the test reproduces them. Control hashes
include original scratch paths in file headings, so a different scratch path
changes the hash while native and Go still must agree with each other.

To reproduce the inventory counter without modifying cohere, create a Go overlay
for an otherwise nonexistent root source file in its module:

```sh
python3 - <<'PY'
import json
from pathlib import Path
repository = Path.cwd()
source = repository / 'stage1/cohere/typeaware/testdata/count_coverage.go'
virtual = repository / 'cohere/adamic_coverage_count.go'
Path('/tmp/adamic-coverage-repro/count-overlay.json').write_text(
    json.dumps({'Replace': {str(virtual): str(source)}}))
PY
go -C cohere build -overlay /tmp/adamic-coverage-repro/count-overlay.json \
    -o /tmp/adamic-coverage-repro/count adamic_coverage_count.go \
    > /tmp/adamic-coverage-repro/count-build.log 2>&1
/tmp/adamic-coverage-repro/count "$ADAMIC_TYPESCRIPT_SOURCE/src/compiler/tsconfig.json" \
    /tmp/adamic-coverage-repro/compiler.manifest \
    > /tmp/adamic-coverage-repro/compiler.counts 2> /tmp/adamic-coverage-repro/compiler.counts.stderr
/tmp/adamic-coverage-repro/count "$PWD/tsconfig.json" \
    /tmp/adamic-coverage-repro/repository.manifest \
    > /tmp/adamic-coverage-repro/repository.counts 2> /tmp/adamic-coverage-repro/repository.counts.stderr
```

Rule options are production defaults. The repository manifest remains the
pre-port population rather than automatically including newly created port files.
The complete report explains exclusions, setup recovery, initial mutant attempts,
regressions and measured limits.
