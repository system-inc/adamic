# Stage 3 adapted-tree pipeline

Requires Git, Node, npm and Python 3. Upstream source stays outside Adamic.
From the repository root, run these in order:

```sh
stage3/apply.sh /tmp/tsc-adapted > /tmp/tsc-apply.log 2>&1
stage3/oracle/run.sh /tmp/tsc-adapted /tmp/tsc-oracle > /tmp/tsc-oracle.log 2>&1
source /workspace/adamic-tools/env.sh # or the env.sh printed by cloud/setup.sh
go build -o /tmp/tsc-census ./stage3/census/tool > /tmp/tsc-census-build.log 2>&1
/tmp/tsc-census /tmp/tsc-adapted/src/compiler /tmp/tsc-census.jsonl > /tmp/tsc-census.log 2>&1
```

Use a new output path for apply and for each oracle run. Apply refuses an existing
output, verifies source.json's tag and commit, and caches upstream under
`~/.cache/adamic-stage3` (`STAGE3_CACHE` overrides that location).
It generates upstream diagnostics, runs `adapt/*/adapt.cjs` in numeric order,
and writes patch-set.md. No changes to apply are needed for a new adaptation.
Rows measure successive edits; the total compares final against pristine plus
upstream generation. Setup has zero source changes.

Each later adaptation must be idempotent and use stock typescript@6.0.3's
compiler API, parse current text, and document changed and declined sites.
When later adapters exist, apply installs api/package-lock.json in its cache
and exposes stock 6.0.3 through NODE_PATH; adapters can require("typescript").
Never weaken Adamic's checker options to improve the census.

The oracle does npm ci, npm run build, and npm test with --light=false,
--workers=4 and --lint=false. It writes phase logs, report.json and baseline.diff;
nonzero exit means failure, including install/build failure or test timeout.
Default tests cover compiler, conformance, project, fourslash, fourslash-server,
transpile and unit tests. Lint, browser integration and ESLint rule tests are
separate tasks. --runners=compiler selects compiler and conformance;
--tests=<regex> narrows a mutant probe. Default test limit is 2400 seconds.
See oracle/REPORT.md for measured runs and throwaway mutants.
