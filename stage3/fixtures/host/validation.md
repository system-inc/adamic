# Host validation

Base: current main ef3d907ecdc4c771b016f7d9c52372def057a340.
Source: TypeScript v6.0.3, 050880ce59e30b356b686bd3144efe24f875ebc8.
The base and census inputs were read from origin/codex/stage3-base and
origin/codex/tsc-census in a separate scratch tree. Neither prerequisite was
merged or copied into this branch. Adaptations 10 and 20 were not applied.

Setup command: `bash cloud/setup.sh > /tmp/host-setup.log 2>&1`, then
`source /workspace/adamic-tools/env.sh`.
Go 1.27.1 ready 0s; clang 20.1.8 ready 1s; Node 24.19.0 ready 1s;
submodules ready 1s; build cache warm 385s; done 385s on 5 processors.
`nproc` is 5, CPU quota is 4 (`400000 100000`). No setup failure occurred.

## Commands and observations

All test output was redirected to files, not piped. The adapter audit command
below is historical evidence from commit 2466bd9; the unused adapter and its
audit were subsequently removed because literal builtin requires are being
implemented natively on codex/require-builtins.

```sh
python3 -u stage3/fixtures/host/check.py --record /workspace/scratch/host-fixture-metadata-final.json --mutants --logs /workspace/scratch/host-final-record-logs > /tmp/host-final-record.log 2>&1
NODE_PATH=/workspace/scratch/host-inputs/stage3/api/node_modules node stage3/fixtures/host/source-audit.cjs /workspace/scratch/host-pristine > /tmp/host-source-audit-final-pass.log 2>&1
NODE_PATH=/workspace/scratch/host-inputs/stage3/api/node_modules node stage3/adapt/50-temporary-node-imports/audit.cjs /workspace/scratch/host-pristine > /tmp/host-adapt-audit-final.log 2>&1
go test ./cmd/adamic ./internal/load -count=1 > /tmp/host-package-tests.log 2>&1
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(functions|library_object_order)\.a$' -count=1 -timeout 30m -v > /tmp/host-filtered-oracle.log 2>&1
```

Fixture observation: 25 Node runs have empty stderr and intended stdout/exit.
Exit fixtures produce codes 0, 1 and 2. All 25 stage-0 builds are Checker;
their complete diagnostics, including the go-run exit-status line, are in
status.json. None compiles, so no native host run or silent miscompile is claimed.
The source audit holds 124 upstream function/method spans and catches a changed
SHA256 call. The adapter audit holds seven declined builtin sites, retains the
optional package and plugin calls, and passes source-byte/idempotence checks.

Package check: cmd/adamic has no test files; internal/load passes in 3.744s.
Filtered native oracle: functions.a, generic_functions.a and
library_object_order.a all pass; total 51.077s. Cache: 3 native hits, 6 native
misses, 0 Node hits, 6 Node misses. This was not the full uncached Adamic gate.
No production Go files or native oracle fixtures/counts were changed.

Upstream oracle commands were run from the scratch stage 3 base:

```sh
bash stage3/oracle/run.sh /workspace/scratch/host-pristine /workspace/scratch/host-oracle > /tmp/host-oracle.log 2>&1
bash stage3/oracle/run.sh /workspace/scratch/host-oracle-mutant /workspace/scratch/host-oracle-mutant-result --runners=compiler --tests=unreachableJavascriptChecked > /tmp/host-upstream-mutant.log 2>&1
```

The full run passes 106,367 assertions, zero failures and zero baseline diffs.
Install 45.149s, build 160.752s, tests 801.258s, total 1007.264s. It uses four
workers, light=false and lint=false. The removed adaptation changed no source
bytes; the source tree judged here is therefore unchanged by its removal. The independent
mutant adds ` [host mutant]` to createFileDiagnostic's computed messageText in
a separate source tree, not to build outputs. Install and build both pass;
the targeted run has 5 passing and 1 failing assertion, changing only
unreachableJavascriptChecked.errors.txt. Total 46.750s. Reports and the exact
failing diff are saved in this host bucket as oracle-report.json,
oracle-mutant-report.json and oracle-mutant-baseline.json. No baseline was
accepted.

## Fixture mutants

Each final mutant changes an actual implementation expression or an exit input.
Each is run on Node through oracle/node.mjs and compared to that fixture's
recorded stdout, stderr and exit. The comparison, not a type/build failure,
catches all 25. These are fixture-sensitivity proofs, not mutations of Adamic's
compiler or library implementation.

| Fixture | Mutant | Catch |
| --- | --- | --- |
| 01 | UTF-8 BOM offset 3 becomes 0 | stdout |
| 02 | UTF-16LE BOM arm skipped | stdout |
| 03 | BE byte-pair swap stops moving the next byte | stdout |
| 04 | read error returns empty text instead of undefined | stdout |
| 05 | BOM is omitted | stdout |
| 06 | file stat asks isDirectory | stdout |
| 07 | directory stat asks isFile | stdout |
| 08 | directory sorting becomes reversal | stdout |
| 09 | failed realpath resolves the input instead of returning it | stdout |
| 10 | mtime returns undefined | stdout |
| 11 | timestamp update uses epoch zero | stdout |
| 12 | unlink is skipped | stdout |
| 13 | mkdir guard never enters | stderr/exit |
| 14 | memoize leaves the callback live | stdout |
| 15 | fake filename suffix changes to sys.cjs | stdout |
| 16 | environment lookup always returns empty | stdout |
| 17 | every write adds a question mark | stdout |
| 18 | exit input 0 becomes 1 | exit |
| 19 | exit input 1 becomes 2 | exit |
| 20 | exit input 2 becomes 0 | exit |
| 21 | SHA256 becomes SHA1 | stdout |
| 22 | djb2 seed 5381 becomes 5382 | stdout |
| 23 | LF becomes CRLF | stdout |
| 24 | filesystem case-probe result is inverted | stdout |
| 25 | extension filtering is skipped | stdout |

An initial mutant for 02 changed the first utf16le toString call, which belongs
to the BE arm and is never reached by that fixture. It survived. The final
mutant targets the LE condition and is caught. This is an observed coverage
limit of the initial mutant, not a claimed catch. Early provenance-audit
implementation errors were corrected before the passing run; they were not
counted as source failures or mutant catches.

Other caught mutants:

- Add an invented line to a recorded stage-0 diagnostic: exact diagnostic comparison.
- Change SHA256 to SHA1 in an extracted source function: upstream syntax comparison.
- Change the real inspector require literal to http in a scratch source: exact site census.
- Append a trailing source comment: source SHA256 comparison alone; census and repeat reports still agree.
- Change the real @types/node lock pin to 25.3.2: declaration-version guard.
- Change the upstream diagnostic value: successful build followed by external baseline failure.

These give 25 caught fixture mutants and 6 caught audit/oracle mutants, plus
the explicitly reported initial surviving LE mutant. The checker diagnostic
mutant is an artifact mutation; the source/site/hash/version and upstream
baseline mutants alter real source or lock inputs. The provenance mutant is
passed as actual source text to the audit.

## Limits

No static-require rewrite is needed: codex/require-builtins implements literal
builtin requires natively, retaining these guarded/lazy loads. Native host support, active
profiling, watch/performance work, Windows filesystem runs, permission/fault
injection, cohere on these incomplete inputs, and the full uncached Adamic
gate were not done. See README.md for the exact member and driver coverage.
