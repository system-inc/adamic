# 99d evidence

Delivery commit: cc2bfae014073a16ba9c54712f67b372e2c76429.
This evidence-only commit adds files beneath this directory. The delivery
branch contains only README.md, adapt.cjs and witnesses.cjs in its new unit.

Primary measurements use main 80c2d3282e44b527555d59690bb413cc883d2bf2
plus all three commits comprising helper 94. Cherry-pick receipts are logged.
The delivered branch stays on 94's original head plus its single 99d commit.
metadata.json records all heads, CLI binary hash and every diagnostic-code
count. sha256.json hashes the collected proof artifacts, excluding itself
and this index.

## Results and artifacts

- 99d-main-diagnostics.json: 54 before, 53 after, 54 for the mutant; D032
  alone is removed, zero diagnostics are added, mutant restores the exact
  baseline multiset. Complete CLI stdout/stderr for all three trees is gzipped.
- 99d-main-witnesses.json: all eight requested Node observations.
- 99d-main-contracts.json and 99d-existing-import.json: no-write refusals,
  idempotence and extension of one existing import declaration.
- oracle-report.json: full default all-runner suite, eight workers,
  106366 passing, 1 failing, 0 pending. Only api/typescript.d.ts differs.
  Install, build, tests and full baseline diff are preserved as gzip files.
- 99d-main-catch.jsonl: the empty raw suite counter. The complete observed
  typeof pair map is {}, total zero. 99d-main-counter-proof.json checks
  emission and a positive control written to the separate sanity JSONL file.
- 99d-source-contexts.json: commandLineParser before/after hashes agree
  between the delivery-base and main-tip contexts.
- 99d-setup.log.gz: successful standard setup at recorded main, GOPROXY
  configured, including the cold submodule restoration and build warming.

An additional delivery-base control oracle completed before the main-tip
measurement checkout was corrected. Its report and counter proof are retained
as delivery-control-* for transparency. It also reports 106366 passing, one
API failure, zero catch entries. Its baseline diff bytes equal the primary
run's diff. It is not substituted for the main-tip measurements.

## Commands and drivers

All command output went to files. Paths below describe the actual scratch
layout; use equivalent fresh directories on another machine.

```sh
export GOPROXY='https://proxy.golang.org|direct'
# Original main checkout at the recorded SHA:
(cd /workspace/adamic && bash cloud/setup.sh > /workspace/scratch/99d-setup.log 2>&1)
source /workspace/adamic-tools/env.sh
# In 99d-measure-repo, cloned from recorded main:
git cherry-pick d42771c36 087d20d81 8a7cd267 > /workspace/scratch/99d-cherry-pick.log 2>&1
# stage3/api's lockfile install was shared with the delivery checkout.
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/99d-upstream bash stage3/apply.sh /workspace/scratch/99d-main-before > /workspace/scratch/99d-main-apply-before.log 2>&1
# Copy the delivered 99d directory into stage3/adapt, then:
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/99d-upstream bash stage3/apply.sh /workspace/scratch/99d-main-after > /workspace/scratch/99d-main-apply-after.log 2>&1
# CLI source is recorded main, with identical loader/checker options:
(cd /workspace/adamic && go build -o /workspace/scratch/99d-adamic ./cmd/adamic > /workspace/scratch/99d-main-build-adamic.log 2>&1)
python3 /workspace/scratch/99d-main-measure.py > /workspace/scratch/99d-main-measure.log 2>&1
# Run from the main-plus-94 measurement checkout:
python3 /workspace/scratch/99d-main-oracle-start.py > /workspace/scratch/99d-main-oracle-driver.log 2>&1
node /workspace/scratch/99d-counter-proof.cjs > /workspace/scratch/99d-main-counter-proof.json 2> /workspace/scratch/99d-main-counter-proof.stderr
export NODE_PATH=/workspace/scratch/adamic-99d/stage3/api/node_modules
node /workspace/scratch/adamic-99d/stage3/adapt/99-correctness-axolotl-d/witnesses.cjs /workspace/scratch/99d-main-before /workspace/scratch/99d-main-after > /workspace/scratch/99d-main-witnesses.json 2> /workspace/scratch/99d-main-witnesses.stderr
node /workspace/scratch/99d-contracts.cjs /workspace/scratch/99d-main-before /workspace/scratch/99d-main-after > /workspace/scratch/99d-main-contracts.json 2> /workspace/scratch/99d-main-contracts.stderr
```

99d-main-measure.py copies only src into a mutant tree, restores e.message,
checks all 79 sorted compiler roots through adamic types, and compares full
normalized diagnostic chains. Every run exits 1 because unrelated diagnostics
remain. Only the added import's line shift and tree root paths are normalized.

99d-main-oracle-start.py copies the final tree, inserts the temporary counter
by exact catch text, initializes its JSONL file, and calls the unchanged
oracle runner. Its child exit 1 is the recorded standing API failure.
The code delivery and final source measurement tree are never instrumented.
99d-counter-proof.cjs runs after the suite; its emitted-code probe redirects
the counter filename before execution and checks the suite bytes stay equal.
The counter statements and measurement drivers are scratch proof tooling,
not additions to the TypeScript adaptation.
