# Baseline oracle

`run.sh <tree> <new-results-directory>` installs the tree's locked dependencies
with npm ci, builds using npm run build, then runs npm test with full baseline
verification. Commands come from v6.0.3 package.json, Herebyfile.mjs and
CONTRIBUTING.md. The compiler, services and test harness are built from the
supplied source. It never accepts or updates reference baselines.

Default runners in src/testRunner/runner.ts are compiler regression, conformance,
project, native fourslash, server fourslash, transpile and unit tests.
Generated fourslash is commented out upstream. Lint is explicitly disabled;
browser integration and ESLint rule tests are independent tasks.

Upstream harnessIO.ts writes only mismatches to tests/baselines/local, with
.delete markers for missing outputs. The oracle compares those to reference,
saves unified diffs, and fails if tests fail, any baseline differs, or no tests
pass. Unexecuted reference baselines are not falsely counted as missing outputs.
Pass/fail/pending counts come from upstream's final reporter totals, not test
file counts. A timeout is a failure (exit 124 in report.json); it never passes a
partial run. Subset options are recorded in every report. Regex-filtered probes use one
worker because upstream's parallel host does not apply Mocha's grep filter.

The 40 minute test limit excludes dependency installation and build. Four workers
match this machine's CPU quota, although nproc reports five. Reported wall time
includes all phases. Always use a fresh results directory and do not run two
oracles concurrently on the same tree, since upstream cleans local baselines.
