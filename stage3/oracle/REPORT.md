# Stage 3 baseline evidence

Source: TypeScript v6.0.3, commit
050880ce59e30b356b686bd3144efe24f875ebc8. Adamic branch starts at main
`ef3d907` and fast-forwards the requested census commit `429c117`.
Only new files under stage3 are changed by this unit; census files are untouched.

Toolchain setup: `bash cloud/setup.sh > /tmp/stage3-setup.log 2>&1`, then
`source /workspace/adamic-tools/env.sh`. Go ready 0s, clang ready 0s,
Node ready 0s, submodules ready 1s, build cache warm 105s, done 105s.
Go 1.27.1, clang 20.1.8, Node 24.19.0. `nproc`: 5; CPU quota: 4;
memory limit: 16 GiB. No setup failure or workaround was needed.

Upstream commands were read from package.json, Herebyfile.mjs and CONTRIBUTING.md.
`npm ci --no-audit --no-fund` installs the unchanged upstream lockfile.
`npm run build` runs `hereby local` and `hereby tests`, including upstream
esbuild bundles and type checking, with generated diagnostics and libraries.
`npm test -- --workers=4 --lint=false --no-colors` runs upstream
`runtests-parallel --light=false`: compiler regression, conformance, project,
native fourslash, server fourslash, transpile and unit tests (327 unit suites,
19,676 runner test files discovered). Counts below are test assertions, not files.

## Observations

Pristine and setup-only adapted src/compiler are byte-identical: 81 files,
including upstream's two generated diagnostic outputs. Repeating 00-setup also
leaves all bytes unchanged. `evidence/setup-equivalence.json` records this check.
Apply's final patch-set.md reports 00-setup and total as zero files and zero lines.
Two temporary 90-probe directories were discovered without changing apply.sh:
the value probe measured 1 file, 1 line added and 1 removed; the comment probe
measured 1 file and 1 line added. Neither temporary adapter is committed. A final drop-in comment probe also
verified require("typescript") resolves the cached locked 6.0.3 API, producing
exactly the same 81 compiler files as the comment control. Setup-only apply
skips the API installation; its final output remains byte-identical to pristine.

## Full baseline runs

```sh
stage3/oracle/run.sh /tmp/stage3-pristine /tmp/stage3-run-pristine-serial > /tmp/stage3-pristine-serial-oracle.log 2>&1
stage3/oracle/run.sh /tmp/stage3-adapted /tmp/stage3-run-adapted-serial > /tmp/stage3-adapted-serial-oracle.log 2>&1
```

| Tree | Wall seconds | nproc | Passed | Failed | Baseline diffs | Oracle |
|---|---:|---:|---:|---:|---:|---|
| Pristine | 436.375 | 5 | 106,367 | 0 | 0 | pass |
| Apply, only 00-setup | 354.767 | 5 | 106,367 | 0 | 0 | pass |

Pristine phases: install 6.544s, build 7.436s, tests 422.337s.
Adapted phases: install 2.496s, build 2.457s, tests 349.774s.
These retries reuse upstream's existing incremental build outputs; the build
commands rerun and bundle current sources. Times are not a performance comparison.
The complete default upstream suite fits under 40 minutes; no default runner
was removed. See evidence/pristine-report.json and evidence/adapted-report.json for exact
phase commands.

## Census before and after

Built and ran the merged census tool with stock Adamic options, unchanged:

```sh
go build -o /tmp/stage3-census ./stage3/census/tool > /tmp/stage3-census-build.log 2>&1
/tmp/stage3-census /tmp/stage3-pristine/src/compiler /tmp/stage3-census-before.jsonl > /tmp/stage3-census-before.log 2>&1
/tmp/stage3-census /tmp/stage3-adapted/src/compiler /tmp/stage3-census-after.jsonl > /tmp/stage3-census-after.log 2>&1
```

Both inputs include upstream diagnostic generation, matching the trees the oracle
builds. All 82 observations agree exactly after removing elapsed seconds and
normalizing the checkout path: 78 individual TypeScript roots and their whole
program are checker-rejected, and 3 JSON inputs return the tool's non-source error.
Whole-program diagnostics: **6,552 before, 6,552 after**. Every reason is unchanged:
TS1484 3,719; TS2412 617; TS18048 380; TS2532 225; TS1294 180; TS2345 733;
the remaining codes and every observation fingerprint are recorded in
evidence/census-comparison.json. No corpus root reaches lowering; this is not a
claim that the native compiler can compile tsc.

Summed gate times: 450.384s before, 189.818s after. The first run overlapped the
failed concurrent suites and memory pressure; these times are observations,
not evidence of a speed change. Raw JSONL and logs remain at the /tmp paths above.
The census branch's original 6,569 count was for the unprepared 77-source tree;
it is not the prepared-to-prepared comparison measured here.

## Mutants and controls

Both temporary adapters used stock npm typescript@6.0.3's compiler API, parsing
the current utilities.ts and obtaining checker types or symbols rather than
using regex or line numbers. The value adapter changes createFileDiagnostic's
string-valued messageText from `text` to `text + " [stage3 mutant]"`.
The comment adapter adds one comment before the same function and changes no
computed value. Both trees were produced by apply.sh, not patched build outputs.

The two decisive commands, with output redirected to separate log files, were:

```sh
stage3/oracle/run.sh /tmp/stage3-mutant /tmp/stage3-run-value-final --runners=compiler --tests=unreachableJavascriptChecked
stage3/oracle/run.sh /tmp/stage3-comment /tmp/stage3-run-comment-final --runners=compiler --tests=unreachableJavascriptChecked
```

Initial targeted runs took 25.092s and 19.915s with the same outcomes. The table
records final verification after making diff comparison byte-exact.

Regex probes select one worker because upstream's parallel host ignores --tests;
the one-worker path runs Mocha with its documented grep argument. Both probes
install locked dependencies and complete the upstream build before testing.

| Tree | Wall seconds | nproc | Passed | Failed | Baseline diffs | Oracle |
|---|---:|---:|---:|---:|---:|---|
| Computed diagnostic value | 22.144 | 5 | 5 | 1 | 1 | fail |
| Comment only | 16.714 | 5 | 6 | 0 | 0 | pass |

Caught by upstream's baseline comparison:
`unreachableJavascriptChecked.errors.txt` changes TS7027's message at two sites
from `Unreachable code detected.` to `Unreachable code detected. [stage3 mutant]`.
See evidence/value-baseline.json (exact unified_diff plus SHA256) and the reports for exact commands,
phase exit codes, times and counts. Compilation succeeds for both; this failure
is a behavioral baseline failure, not a compiler build rejection.

An earlier objectLitStructuralTypeMismatch probe passed all 6 assertions on the
value mutant: it uses a different diagnostic constructor. This is an observed
coverage limit, not evidence of semantic equivalence; its report is preserved.
The unreachable-code probe was chosen after inspecting the diagnostic path.

Two guard probes also fail: a scratch copy of source.json with an all-zero commit
is rejected by apply's pin check (exit 1, actual commit printed); a nonexistent
test regex makes upstream exit 0 with 0 tests, but the oracle returns failure.
The empty-run report is saved in evidence/empty-report.json.

## Resource failure and recovery

The first two full suites were mistakenly launched concurrently with four workers
each. They exhausted the 16 GiB limit: memory.events reported two OOM kills.
Pristine failed after 162.157s wall, adapted after 159.098s, before test totals.
These are infrastructure failures, not passing baselines; reports are preserved.
Orphan workers were terminated. The oracle now terminates remaining process-group
workers after an early exit or timeout. Full-suite retries run serially.
An initial parallel regex probe was stopped once inspection showed that upstream
ignores the regex in that mode; it is not counted as a targeted test result.

## Coverage limits

Lint is explicitly disabled. Upstream browser integration, ESLint-rule tests,
and generated fourslash (commented out in upstream's default runner) are separate
from the test suite exercised here. No baseline is accepted or overwritten.
No native tsc build or full Adamic integration gate was run: this unit establishes
the external upstream oracle and source preparation pipeline.
Test logs remain under /tmp in accord with CLAUDE.md; checked-in JSON reports and
the failing baseline diff preserve the reviewable measurements.
