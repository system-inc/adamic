Native-composition follow-up: see [NATIVE_REPORT.md](NATIVE_REPORT.md). The
Node-only observations below describe the original printer commit.

The CSS and SCSS printer runs on Node and matches Go cohere on 48,152 cases across two option sets.
Both pinned Prettier oracles have 4,952 byte-identical formats per option set.
Three running printer mutants are caught in both option sets.
Native composition remains blocked by the recorded cycle-proof gap assigned to another worker.
Throughput and final verification are recorded below.

# Implementation

`print.ts` implements stylesheet, declaration, rule, at-rule, selector, media and
value printing. It composes the existing selector, value, media query, quote and
number slices. `print_doc.ts` implements the CSS-used document operations,
including width-dependent groups, fills, indentation, trailing line suffixes,
line removal and Unicode width. `print_width_data.ts` retains cohere's pinned
emoji and East Asian width data. `print_utilities.ts` models parent paths,
siblings, raw source positions and contextual spacing predicates.

The port follows cohere's Go printer and its explicit CSS/SCSS scope. The line
driver accepts escaped batch cases and has a checksum mode for throughput.
Full formatting starts with the existing composed parser. The canonical Go
oracle calls actual `formatWithParser`; the original oracles call complete
Prettier formatting with either npm Prettier 3.9.6 or all seven plugins in
cohere's pinned bundle. Original-library paths are external environment variables,
not vendored installations. MIT attribution and width-data notices are included.

# Branch and setup

Branch: `codex/stage1-css-printer`, based on current origin/main
`d799ede40af1d947efe8c4a897efbb90c2ef2264`.
Main lacked some expected slice dependencies. Existing accepted selector, CSS
parser, quote and number commits were reused, then the accepted native-regex
branch was merged in `8d21b63`. Dependency commits on this branch are `2a9f2e4`,
`ab996e2`, `3a7a3e9`, `371fac6`, `c98ba88`; the merge carries regex `59d82c5`.
No hand edits were made to the four protected compiler files.

The older regex dependency's five runtime shapes and generated named-capture
shape lacked main's new method-table field. Integration adds explicit NULL
initializers in `internal/native/runtime/regexp.c` and `internal/regexp/native.go`.
The filtered native regex oracle catches their omission through clang's
`-Werror,-Wmissing-field-initializers`, then passes after the fix. Compatibility commit: `2a3416f`. No regex
algorithm or cycle proof was changed. Three counts rows are retained from the
accepted regex dependency; the complete counts gate was not rerun.

`bash cloud/setup.sh` was rerun successfully after an initial setup/build raced
the dependency checkout and saw a mixed IR tree with missing RegExp fields.
Both logs are retained. The successful timing lines were:

```text
setup: go ready (1s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (39s)
setup: done in 39s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

`nproc` is 5. Environment: Go 1.27.1, clang 20.1.8, Node 24.19.0.
Commands source `/workspace/adamic-tools/env.sh`.

# Corpus and oracle results

The unchanged parser-unit corpus contains 12,038 texts and 24,076 CSS/SCSS
pairs: cohere CSS test constants and concatenations, repository/submodule CSS,
SCSS and Less, pinned fork fixtures, 6,000 seeded generated cases, comment and
custom-property combinations, and malformed Unicode-boundary truncations.
The external fixture checkout is `/tmp/adamic-css-prettier`, commit
`cb4b33fba24a8428d00e54be85fc886288a374ea`. It contributes 157 CSS, 90 SCSS and
43 Less files; one additional CSS file comes from the repository/submodules.
See the earlier `REPORT.md` for corpus collection and checkout commands.

Default options are width 80, two spaces, double quotes, trailing commas all.
The second set is width 24, tab width 4, tabs, single quotes, trailing commas none.
For each set, all 24,076 port formats and refusals match Go exactly: 4,966
formatted and 19,110 refused. Both npm and bundled Prettier each have 4,952
byte-identical formats, 19,080 shared refusals and 44 recorded discrepancy
occurrences (BOM 12, CR 2, nonbreaking space 2, YAML 28).

Those discrepancies concern the full-file wrapper or YAML delegation, proved
in `gaps/printer_boundaries.ts` and `gaps/printer_library.mjs`, and documented
in `GAPS.md`. Exact input, mode, variant and answers are allowlisted in
`testdata/printer_library_gaps.json`. Unlisted acceptance or format differences
fail. Formatted strings are decoded from JSON before byte comparison so U+2028
transport escaping does not become a false format discrepancy. Go refusal
messages and positions are compared exactly; Prettier public code frames and
error positions are not held to Go's different internal error API here.

# Mutants

All three mutants run successfully with clean stderr and are rejected by the
Go-byte oracle on Node in both option sets; compiler failure is not counted.

| Mutant | What caught it |
|---|---|
| Declarations omit semicolons | First difference line 2, byte 69 |
| Rule bodies omit indentation | First difference line 6, byte 7 |
| Groups ignore remaining width | First difference line 26, byte 12 |

The explicit native-gap test fails when either the small regex/media proof or
complete printer starts lowering, requiring native, JavaScript backend,
sanitizer, leak and throughput checks then. Node prints `2` and `Ok` for the
small proof; both builds currently refuse unknown `ir.RegExpNew` in the cycle
finder. Native printer mutants cannot run while composition is refused.

# Verification commands

Every test writes to a log file directly, never a pipe. Commands run:

```sh
ADAMIC_CSS_LIBRARY=/tmp/adamic-css-library ADAMIC_CSS_PRINTER_LIBRARY=/tmp/adamic-css-printer-library go test -v -count=1 -timeout=30m ./stage1/cohere/css
ADAMIC_CSS_FIXTURES=/tmp/adamic-css-prettier ADAMIC_CSS_PRINTER_LIBRARY=/tmp/adamic-css-printer-library go test -v -count=1 -timeout=30m ./stage1/cohere/css -run '^TestCSSPrinterAgreesWithGo$'
go test -count=1 -timeout=30m ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/regexp'
go test -count=1 ./internal/regexp
ADAMIC_CSSSTRINGS_LIBRARY=/tmp/adamic-css-printer-library ADAMIC_CSSNUMBERS_LIBRARY=/tmp/adamic-css-printer-library go test -count=1 -timeout=30m ./stage1/cohere/cssstrings ./stage1/cohere/cssnumbers
go vet ./...
git diff --check
```

Final outputs: full CSS PASS in 154.929s; complete printer corpus PASS in
99.517s; filtered regex oracle PASS in 1.488s; regex package PASS in 4.750s;
quote package PASS in 12.719s; number package PASS in 61.553s. `go vet ./...`
and `git diff --check` produce no output and exit successfully.

The full CSS regression uses 23,496 pairs without external fork fixtures; the
separate printer run includes all 24,076 pairs. Existing raw-parser regression
includes native, Node, JavaScript backend, sanitizer/leak checks and its original
mutants. The complete repository gate was not run. Printer boundary proofs and
Adamic load/lower type checks pass; native lowering refuses at the recorded gap.
Failed pre-fix regression and regex runs are retained with the successful logs.

# Throughput

The opt-in benchmark uses 4,952 cases accepted identically by Go and Prettier,
10 repetitions, three rounds, 49,520 formats and 5,102,980 output UTF-16 units
per round. Each runtime checksum must match. Go measures in-process parse and
print. Node timings include process startup, case decoding, parse and print;
Go's command startup and case decoding are excluded. This is a stylesheets/s
comparison, not document rendering alone. The first benchmark overlapped tests
and is discarded; the final benchmark runs without competing test processes.

```sh
ADAMIC_CSS_FIXTURES=/tmp/adamic-css-prettier ADAMIC_CSS_PRINTER_LIBRARY=/tmp/adamic-css-printer-library ADAMIC_CSS_PRINTER_BENCH=1 go test -v -count=1 -timeout=30m ./stage1/cohere/css -run '^TestCSSPrinterThroughput$'
```

Final benchmark PASS in 226.708s. Every round has the exact output checksum.

| Runtime | Round 1 / s | Round 2 / s | Round 3 / s | Median / s |
|---|---:|---:|---:|---:|
| Adamic source on Node | 1708 | 1580 | 1596 | 1596 |
| Prettier fork on Node | 1777 | 1632 | 1703 | 1703 |
| Go | 4432 | 5347 | 5568 | 5347 |

Printer implementation commit: `b902761`; regex compatibility: `2a3416f`;
dependency merge: `8d21b63`. This report and the throughput log are committed
separately. Verification logs have trailing whitespace trimmed for review.

# Limits

Native composition, native throughput, printer sanitizer/leak evidence and the
JavaScript backend printer remain blocked. This unit follows Go's CSS/SCSS
entry points; testing Less files through them does not implement Less semantics.
Nonempty YAML delegation, HTML style attributes, CSS-in-JS placeholder
orchestration, private application files, incremental edits and arbitrary
invalid-UTF-8 output are not covered. Only document operations used by Go's CSS
printer are implemented. The public Prettier error code-frame API is not ported.
