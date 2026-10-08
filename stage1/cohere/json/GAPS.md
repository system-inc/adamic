# JSON formatter: Go cohere is the stage 1 contract

The formatter beside this file is Adamic `.ts`, compiled by stage 0 to native
code and also run as source on Node. It ports `estree/parse_json.go` and JSON's
reachable paths through `print_json.go`, `print_object.go`, `print_array.go`,
`print_expressions.go`, `utility_text.go` and the shared document printer.
`main.ts` formats one file to stdout; `--cases` is the escaped batch test driver.
Trees and documents use child indexes in tables so their ownership is acyclic.

Ahra's decision: stage 1 agrees byte for byte with **Go cohere**. Prettier parity
is cohere's upstream contract. `port_test.go` compares exact successful output
and exact syntax-error messages to Go; `audit_test.go` reports Prettier separately
and checks the entire report against `known-upstream-differences.txt`. Closing,
adding, or changing a known disagreement fails that report test.

## Current pin: f5d1934a

The cohere bump from `7945d102` to `f5d1934a2d7bebe706210cb1cfd01aebff4f8ca7`
closes the numeric-separator oracle gap in commit `1ee5f944`. The six probes in
`TestUpstreamNumericSeparatorGap` now require both Go and Prettier to refuse;
`gaps/numeric-separators.json` remains a witness. The Adamic port still accepts
these malformed literals, so `TestPortMatchesGoCohere` and
`TestProfileSnapshotsAgree` fail at that witness. This bump does not change the port.

The separate Go/Prettier report now contains two empty-input differences.
Although `1ee5f944` makes the outer `native.Formatter.FormatParsed` dispatch accept empty text,
this slice calls `javascript.FormatJSON` directly, which still refuses it.
The two generated empty inputs therefore stay in the checked-in report.
The cohere TypeScript pin is `d92d9bfee114c80be2c375d72edae966176e3a4f`;
Prettier remains 3.9.6. The corpus pin now contains 853 cohere identities;
repository JSON is checked against Git and the other pinned groups are unchanged.

The original slice observations, pins and measurements below are historical.

## Pins and scope

- Adamic main: `fe3b9f236e0672e968bcb7ad53c1882cdde18f29`.
- cohere: `715ba94f3608a6500086b1076ce5cb7e51b836db`.
- cohere/TypeScript: `8d550c837c90bd1805b047b7eeccc2baac2d5e7a`.
- External Prettier: **3.9.6**, matching cohere's embedded standalone version.
  `testdata/library.mjs` checks this version; npm installs into a scratch directory.

`cohere/internal/format/native/json.go` registers `.json` and delegates to
`javascript.FormatJSON`. `JSONParser` selects `json-stringify` for `package.json`,
`package-lock.json`, and `composer.json`; other names use `json`. The former has a
small printer in `print_json.go`, but ordinary JSON uses the shared estree printer
and document/comment machinery. There is no independently dispatched `.jsonc`
formatter here. The parser explicitly lists jsonc/allowEmpty as unsupported.

The port uses `FormatJSON` with `PrettierDefaults`: width 80, two-space indent,
ordinary `json` and the filename-selected stringify mode. External Prettier runs
with `filepath` and its defaults. Its report compares acceptance and successful
bytes; error wording is not compared across those two APIs. Port-to-Go errors
are compared exactly. Go parity runs without a Prettier installation.

`doc.ts` ports concat, indent, group, break propagation, hard/soft lines, fill,
line suffixes and blank-line trimming. `width.ts` uses cohere's exact captured
East Asian width and emoji tables; the latter are compiled into ordered RE2
instruction programs, matched on the same UTF-16 surrogate mapping. The two Go
generators in `testdata` reproduce `widthTables.ts` and `identifierTables.ts`;
cohere formats their output afterward. These are character/pattern tables,
never corpus answers. The port invokes no Go, Node or Prettier subprocess.

## Observations and proving programs

`gaps/numeric-separators.json` is `[1__0]`. Go formats it as `[1__0]\n`; Prettier
refuses with `A numeric separator is only allowed between two digits. (1:3)`.
`TestUpstreamNumericSeparatorGap` also proves `[1_]` and `[0x_1]`, under both
`probe.json` and `package.json`. All six are accepted by Go and refused by Prettier.
Inspection of `estree/parse_json.go` shows underscores stripped without validating
their positions. That explains the observed acceptance; this is upstream parser
behavior, not an observed stage 0 compiler defect.

The generated corpus also proves empty input under both filenames. Go refuses
with `json: empty input`; Prettier succeeds with an empty output. The report
keeps this distinction rather than treating refusal as successful empty output.

The mandatory Go parity test walks every `.json` file in the checkout and both submodules,
excluding only `.git` directories. Its 1,075 inputs comprise 4 Adamic files
(including the new proving fixture), 835 cohere files, 202 TypeScript files, and
34 generated cases. The generated cases include 500 levels of nesting, a
5,000-element array, comments, trailing commas, Unicode and surrogate escapes,
decimals/exponents/precision extremes, hexadecimal/octal/binary forms, numeric
separators, nonfinite literals, incomplete input, and empty input, each under
ordinary and stringify filenames. Non-UTF-8 files fail the audit explicitly. Go produced 922 formatted outputs and
153 syntax refusals; all 1,075 answers agree byte for byte on native, Node and the
JavaScript backend, including the refusal messages.

All **1,040 pre-existing repository files** agreed in acceptance and successful
bytes. The separate Prettier report records exactly **9/1,075** disagreements: the new
proving fixture, six generated separator cases, and two empty cases. All nine
remain in the mandatory port-to-Go corpus, with no exemptions. These are observations at
the pins above, not proof that all other possible inputs agree.

`TestAdditionalJSONBoundaries` adds 72 Go-only probes: invalid exponents and
radix forms, legacy octal and bigint refusals, holes, numeric keys, single quotes,
template escapes and newlines, Unicode identifiers, dangling/leading/trailing
comments and blank lines, invalid escapes, and display-width boundaries for
CJK, combining marks, astral letters, regional flags, ZWJ emoji, keycaps,
variation selectors and modifiers. Both filename modes run natively under
ASan/UBSan and LeakSanitizer, and on Node. `TestSingleFileStdoutDriver` exercises
the ordinary and stringify file-to-stdout entry directly.

## Stage 0 gaps, with proving programs

`TestDocumentedStageZeroGaps` loads and lowers each of these programs and checks
its refusal, then runs its source on Node and checks the stated output:

1. `gaps/multiplePush.ts`: `stage 0 can't lower push with other than one value yet`.
   Node prints `a,b`. Around it: one `push` call per value.
2. `gaps/emptyFallback.ts`: `stage 0 can't lower an array of never yet`.
   Node prints `0`. Around it: comment-bearing inputs initialize all comment lists; indexed reads
   use a checked `panic` fallback instead of an untyped `[]`.
3. `gaps/repeatInTry.ts`: `stage 0 can't lower a try around repeat, whose failure
   is a panic natively but a throw a catch can take on Node (docs/memory.md) yet`.
   Node prints `xxx` with three arguments. Around it: indentation caches string
   prefixes rather than calling dynamic `repeat` inside caught formatting code.

The parser also uses the existing GraphQL slice's table-of-nodes shape and one
self-recursive value method around the documented ownership-cycle and forward
method restrictions. This unit does not change any compiler files.

## Reproduction

From the repository root, after `bash cloud/setup.sh`:

```sh
source /workspace/adamic-tools/env.sh
npm install --prefix /tmp/adamic-json-prettier --save-exact prettier@3.9.6 > /tmp/adamic-json-npm.log 2>&1
go run ./cmd/adamic build stage1/cohere/json/main.ts -o /tmp/adamic-json-port > /tmp/adamic-json-build.log 2>&1
/tmp/adamic-json-port CohereSettings.json > /tmp/adamic-json-formatted.txt 2>/tmp/adamic-json-driver-error.log
ADAMIC_JSON_PRETTIER=/tmp/adamic-json-prettier ADAMIC_JSON_REPORT=/tmp/adamic-json-known.txt go test -v -count=1 -timeout 30m ./stage1/cohere/json > /tmp/adamic-json-final.log 2>&1
```

The Prettier environment variable points to a scratch npm installation; the
helper refuses a version other than 3.9.6. Without it, only external report and
upstream mutation tests skip. The Go corpus, native/Node/JavaScript-backend
parity, leak checks, boundary tests and three port mutants still run. Go overlays
inject helpers and mutants into scratch copies, leaving the submodule untouched.
The test copies the Adamic sources into scratch directories too.

## Three caught mutants

`TestThreePortMutantsAreCaught` changes the **Adamic port**, builds each changed
program with stage 0, then runs it natively and on Node against Go answers:

1. Remove the final hardline: caught on all four controls.
2. Remove the space after a property colon: caught on three object controls.
3. Select ordinary JSON for `package.json`: caught on two stringify controls.

Every mutant compiles and exits successfully on both sides, with empty stderr.
Only its differing answers catch it. The four controls include Unicode, nested
containers, and the accepted malformed numeric separator. The separate original
upstream Go-printer mutations remain as checks of the external comparison.

## Performance and verification

The original port baseline below is retained for comparison. The performance
follow-up, optimized measurements, top twenties and runtime/compiler proposals
are in [PERFORMANCE.md](PERFORMANCE.md). No `internal/` files changed.

On the original frozen source, three back-to-back rounds over all 1,075 inputs:

| Side | Seconds, rounds 1 / 2 / 3 | Texts/s, rounds 1 / 2 / 3 | Median texts/s |
|---|---|---|---|
| Native release | 16.961 / 17.517 / 17.261 | 63.38 / 61.37 / 62.28 | **62.28** |
| Node source | 12.405 / 11.291 / 11.791 | 86.66 / 95.21 / 91.17 | **91.17** |
| Go cohere | 9.788 / 10.024 / 9.837 | 109.83 / 107.24 / 109.28 | **109.28** |

These are observed whole-process wall times, excluding builds. Every side reads
the same approximately 60 MB escaped cases file, formats or refuses each text,
and writes the same approximately 62 MB escaped answer stream. All timed outputs
are checked against Go. Node includes its source/type-stripping startup; native
and Go use already-built binaries. Refused inputs count in texts/s. This measures
the drivers as well as formatting, not only the printer's inner loop. Native is
slower than both Node and Go on this corpus; no performance win is claimed.

Verification, all output redirected to named log files:

```sh
ADAMIC_JSON_PRETTIER=/tmp/adamic-json-prettier ADAMIC_JSON_REPORT=/tmp/adamic-json-known-frozen.txt go test -v -count=1 -timeout 30m ./stage1/cohere/json > /tmp/adamic-json-frozen.log 2>&1
go test -v -count=1 -timeout 30m ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(strings|strings_more|collections|exceptions|bitwise)\.a$' > /tmp/adamic-json-filtered-oracle.log 2>&1
go vet ./... > /tmp/adamic-json-vet-final.log 2>&1
gofmt -l cmd internal stage1/cohere/json > /tmp/adamic-json-gofmt-final.log 2>&1
/tmp/adamic-json-cohere --no-fix --format-only --format-all stage1/cohere/json/*.ts > /tmp/adamic-json-format-check-final.log 2>&1
/tmp/adamic-json-cohere --no-fix --no-format stage1/cohere/json/*.ts > /tmp/adamic-json-lint-final.log 2>&1
```

`/tmp/adamic-json-cohere` is built from the pinned submodule with
`go build -o /tmp/adamic-json-cohere ./command/cohere`, run from `cohere/`.
The JSON slice suite passed in **443.707s**, including whole-corpus native
ASan/UBSan and LeakSanitizer, Node source, JavaScript backend, 72 boundary probes,
the direct file-to-stdout driver, three port mutants, three gap programs, the
separate exact nine-difference report, and the original external-comparison
mutants. No external or corpus tests skipped in that run.

The filtered compiler oracle passed in **15.330s**, on eight matching fixtures:
`strings.a`, `strings_more.a`, `undefined_strings.a`, `optional_strings.a`,
`collections.a`, `maybe_collections.a`, `exceptions.a`, and `bitwise.a`.
Vet and gofmt produced no output. The cohere format check passed; its lint/type
check reported **276 rules, 7 checked, 100% Adamic-ready**. `git diff --check`
passed. The full repository test gate was not rerun for the port: the previous
audit's full gate took roughly eleven minutes, so this change used the complete
touched package plus the filtered compiler oracle, as the unit allows.

Toolchain: Go 1.27.1, clang 20.1.8, Node 24.19.0.
Setup succeeded: Go ready 0s; clang ready 0s; Node ready 0s; submodules ready 0s;
build cache warm 13s; total 13s. `nproc` was 5; CPU quota was 4 CPUs.

## Not covered

This is the default-options JSON slice, not JSONC or the general JavaScript
printer. It does not test configurable widths, tabs or quote options, the native
wrapper's BOM/CR normalization, repository option resolution, or non-UTF-8 input.
The corpus rejects non-UTF-8 files explicitly. Syntax errors follow Go's parser,
not Babel's error messages. Inputs outside this corpus and the 72 extra probes
are not exhaustively proven; notably pathological comment attachment and
arbitrarily large radix keys have not had a generated differential sweep.
