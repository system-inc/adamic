# The diagnostics oracle for days 4 and 5

This driver judges the real `tsc --noEmit` CLI. The corpus contains 300 small
programs from TypeScript 6.0.3's `tests/cases/compiler`: 60 clean and 240 with
errors. Each error case retains its upstream `.errors.txt` baseline. Each case
has expectations derived from that baseline and separately captured Node goldens.
`selection.json` records every upstream source path, local `path`, SHA256, baseline path/hash, header
option, feature bucket and diagnostic code. `SELECTION.md` lists all 300 cases.

```sh
stage3/drivers/tsc/run.sh node /absolute/adapted/built/local/tsc.js > /tmp/tsc-node.log 2>&1
stage3/drivers/tsc/run.sh '/absolute/native-tsc' > /tmp/tsc-native.log 2>&1
NATIVE_TSC=/absolute/native-tsc stage3/drivers/tsc/run.sh 'node /absolute/adapted/built/local/tsc.js' > /tmp/tsc-both.log 2>&1
stage3/drivers/tsc/run.sh --tiny /absolute/native-tsc > /tmp/tsc-day4.log 2>&1
python3 stage3/drivers/tsc/audit.py > /tmp/tsc-audit.log 2>&1
python3 stage3/drivers/tsc/mutants.py node /absolute/adapted/built/local/tsc.js > /tmp/tsc-mutants.log 2>&1
```

A command can be one shell-quoted string or separate argv words. It is parsed
without a shell; existing executable/script paths become absolute before changing
working directories. Native binaries use the same cases, options and comparisons.
`NATIVE_TSC` adds a second run after the given command. Without a native path the
native slot does not run. Its Node-forwarder smoke test proves the plumbing only.

The runner prints a results directory, by default a fresh `/tmp/tsc-diagnostics-*`.
Set `TSC_RESULTS` to retain results at a chosen fresh path. Each case writes its
materialized program, tsconfig, full command argv, `actual.stdout`, `actual.stderr`
and `actual.exit`; `report.json` reports all failures. Four independent processes
run by default (`TSC_JOBS` overrides); `TSC_TIMEOUT` is a per-case timeout in
seconds, default 60. Timeouts fail the exit comparison. Capture output to a log,
then inspect the log. Nothing pipes a test process into another process.

## Written selection rule

The rule in `corpus.py` does not run the tested compiler or select cases by whether
its output matches. It pins upstream commit
`050880ce59e30b356b686bd3144efe24f875ebc8` and enumerates `.ts` files directly in
`tests/cases/compiler`, sorted by name. A candidate has at most 80 physical lines
and 8,192 bytes. It must be a single file with supported, single-valued option
headers, no triple-slash reference and no relative import/export. Filename,
virtual filesystem, unsupported harness controls and multi-option variants are
excluded. Variant `.errors.txt` baselines are excluded, even if a default exists.
A nonempty reference summary must locate every diagnostic in that source file.
Missing baselines mean clean, as in upstream's test convention.

Declaration-emission cases, reference summaries with codes below 2000, and
TS18027 (an emit-only private-name collision diagnostic) are excluded. The stock
6.0.3 parser additionally rejects any remaining parse-error candidate. This
keeps expectations in the CLI's noEmit domain: the API suite can collect semantic
errors alongside syntax errors, while the CLI stops before checking after syntax
errors; emission can report errors that noEmit suppresses. These restrictions
are corpus scope choices, not adaptations to compiler behavior.

Features are assigned by the first matching case-insensitive filename pattern
in the ordered `FEATURES` list in `corpus.py`, with `other` as fallback. First
select 60 clean cases, then 240 error cases without replacement. At each step
minimize the tuple: negative count of previously unseen diagnostic codes,
number already selected in that feature in this partition, SHA256 of the UTF-8
filename, filename. Track seen codes across both partitions. This prioritizes
code diversity, then feature balance, with an explicit stable tie-breaker.
The final population has 3,480 eligible candidates, 17 nonempty feature buckets
and 285 distinct diagnostic codes. No random state or filesystem iteration order
enters the choice. The corpus includes classes, generics, unions, functions,
arrays, objects, modules, namespaces, enums, decorators, async, operators,
control flow, literals, types and a few parser-related semantic tests.

To regenerate into an absent `corpus/` directory:

```sh
python3 stage3/drivers/tsc/corpus.py /absolute/upstream-tree /absolute/stock/typescript/lib/typescript.js > /tmp/tsc-selection.log 2>&1
stage3/drivers/tsc/run.sh --record node /absolute/adapted/built/local/tsc.js > /tmp/tsc-record.log 2>&1
```

`--record` compares stdout, stderr and exit against independent expectations
before copying actual bytes to goldens. It refuses to replace different existing
goldens. Reviewed goldens never update during an ordinary run.

## Headers, diagnostics and files

Original bytes are stored under their exact upstream filenames and extensions,
including BOMs and CRLFs. All 300 selected inputs are `.ts`; they are verbatim
upstream TypeScript data, often intentionally ill-typed, not authored Adamic code.
The manifest's `path` locates each input in its case directory; `source` retains
the upstream path and `source_sha256` retains the unchanged byte pin.
New authored tiny programs use `.a`. Their first-line a-check metadata is omitted
when materializing the tiny project, preserving its original diagnostic positions.
No production compiler file or other
fixture bucket is changed, and this unit does not add a `status.json` bucket or
edit `fixtures_test.go`.

The materializer follows upstream `makeUnitsFromTest` in
`src/harness/harnessIO.ts`: remove `// @name: value` metadata lines, normalize
source line endings to LF and discard leading empty content exactly as that
routine does. This matters for line/column positions. Header names are mapped
case-insensitively to canonical compiler options, booleans become booleans,
and `lib` becomes an array. Unsupported headers are rejected, never silently
passed as compiler flags. Emit-only harness settings are consumed because the
CLI is always run with `--noEmit`; declaration cases are excluded.

Each isolated tsconfig explicitly disables ambient type-package discovery with
`types: []`. Defaults mirror the upstream diagnostic harness's
`skipDefaultLibCheck: true` and `noErrorTruncation: true`; `ignoreDeprecations:
"6.0"` permits older targets selected by real tests. Case headers override these
defaults. `--noEmit --pretty false` is appended to the real CLI command.
Standard libraries come from the given compiler's installation, so a native
compiler's library/host failures are also observable.

The upstream baseline's first summary block supplies expected diagnostic text.
Only its CRLFs are converted to LF, matching Node's Linux CLI convention, and one
final newline is retained. The annotated baseline itself stays byte-for-byte.
Actual CLI output is never normalized: every path, line, column, code, message,
indentation, ordering and newline is compared as bytes. stderr is independently
compared, including empty output. Clean CLI exit is 0; noEmit semantic diagnostics
exit 2. Harness success is 0 only when every comparison passes; failure is 1.

## Day 4

`tiny/` has exactly three source files. `clean.a` has a valid typed export;
`assignment.a` produces TS2322; `argument.a` produces TS2345. The generated
project uses strict checking and ES2020. Its golden stdout contains exactly two
diagnostics, golden stderr is empty, and the CLI exit is 2. `--tiny` runs only
this project through the same driver, including the optional native slot.

## Limits

This is a Linux, single-file, bounded corpus plus one three-file project, not the
entire TypeScript suite. JSX `.tsx`, multi-file/module-resolution scenarios,
virtual filesystem behavior, option variants, suggestions, declaration emit,
syntax-error API baselines, watch/build/incremental modes and Windows newline/path
behavior are outside this selection. The small corpus does not prove the adapted
tree passes Adamic's checker or the full upstream suite. See `REPORT.md` for the
observed build failure, runnable bundle, verification, mutants and source hashes.
