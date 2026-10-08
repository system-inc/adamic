# Public any scout

[REPORT.md](REPORT.md) records each of adaptation 41's 19 remaining source
contracts, the separate concrete-timer checker anomaly, public declaration
owners, all internal caller/reference locations, runtime observations and three
costed options. Nothing is decided or adapted. [contracts.json](contracts.json)
retains full declaration text, source hashes, caller uses and all observed shapes.
The questions for @system_adamic are at the report's end.

Reproduction, with every run redirected to a log:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/public-any-setup.log 2>&1
source /workspace/adamic-tools/env.sh
git worktree add --detach /tmp/public-any-adaptation 1e920a31b601422c75a26ead083e9c1423c34b66
bash /tmp/public-any-adaptation/stage3/apply.sh /tmp/public-any-tree > /tmp/public-any-apply.log 2>&1
NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules node stage3/scouts/step09/public-any/inspect.cjs /tmp/public-any-tree stage3/scouts/step09/public-any/evidence/contracts.json > /tmp/public-any-inspect.log 2>&1
(cd /tmp/public-any-tree && npm ci --ignore-scripts --no-audit --no-fund > /tmp/public-any-install.log 2>&1 && npx hereby tsc --no-typecheck > /tmp/public-any-build.log 2>&1 && npx hereby services --no-typecheck > /tmp/public-any-api-build.log 2>&1)
NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules node stage3/scouts/step09/public-any/instrument.cjs /tmp/public-any-tree/built/local/_tsc.js stage3/scouts/step09/public-any/evidence/instrumentation.json > /tmp/public-any-instrument.log 2>&1
PUBLIC_ANY_OUTPUT=/tmp/public-any-captures TSC_RESULTS=/tmp/public-any-acceptance TSC_JOBS=4 stage3/drivers/tsc/run.sh node --require "$PWD/stage3/scouts/step09/public-any/observe.cjs" /tmp/public-any-tree/built/local/tsc.js > /tmp/public-any-acceptance.log 2>&1
python3 -B stage3/scouts/step09/public-any/aggregate.py /tmp/public-any-captures /tmp/public-any-acceptance > /tmp/public-any-aggregate.log 2>&1
python3 -B stage3/scouts/step09/public-any/run-fixtures.py /tmp/public-any-tree/built/local/typescript.js > /tmp/public-any-fixtures.log 2>&1
node stage3/scouts/step09/public-any/counterexamples.cjs /tmp/public-any-tree/built/local/typescript.js stage3/scouts/step09/public-any/evidence/counterexamples.json > /tmp/public-any-counterexamples.log 2>&1
python3 -B stage3/scouts/step09/public-any/report.py > /tmp/public-any-report.log 2>&1
python3 -B stage3/scouts/step09/public-any/audit.py stage3/scouts/step09/public-any/contracts.json /tmp/public-any-tree > /tmp/public-any-audit.log 2>&1
```

Adjust NODE_PATH to the stock 6.0.3 API cache used by apply. All outputs are new
scratch paths. Only the scratch emitted CLI is instrumented; source is unchanged.
Build the services/API bundle separately: rebuilding the CLI would remove its
instrumentation. The no-typecheck build is for measuring Node behavior, not an
Adamic admission or an upstream full build claim.

Setup succeeded in 9.316 seconds: Node ready 0.022s, Go ready 0.023s, clang ready
0.144s, Go build ready 9.171s, test binaries deferred 9.288s, cache warm 9.289s.
nproc is 5; cgroup CPU quota is 4. The exact setup log is retained. Acceptance
passes 301/301, with no golden rewrite. Every capture is committed compressed.
The exact counters and domains remain separate from inferred design costs.

The audit compares every residue location with the independent source document,
checks all 301 captures and their hashes, and checks three fixture/mutant records.
A dropped-row mutant and a missing-project-capture mutant both fail. Full logs,
Node outputs, native compile diagnostics and source-mutant captures are retained.
No whole-package tests, full gate, or real native scanner/tsc proof was run.
These fixtures live only in this scout; internal/oracle counts are unchanged.
