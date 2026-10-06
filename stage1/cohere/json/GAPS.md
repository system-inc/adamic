# JSON formatter: upstream audit, port blocked

This directory contains a reproducible upstream audit, not an Adamic formatter.
There are no Adamic `.ts` implementation files, native driver, or four-way parity
claim. At these pins, an implementation cannot agree with both Go cohere and
Prettier on every input: the two required oracles disagree. Choosing either
behavior would require an explicit exception to the unit's parity requirement.

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

The audit invokes `FormatJSON` with `PrettierDefaults`, and external Prettier with
`filepath` and its defaults. It compares acceptance and exact bytes of successful
outputs. Error wording is not compared because the APIs differ. It does not test
the native wrapper's BOM and CR/CRLF normalization, repository option resolution,
or CLI behavior.

## Observations and proving programs

`gaps/numeric-separators.json` is `[1__0]`. Go formats it as `[1__0]\n`; Prettier
refuses with `A numeric separator is only allowed between two digits. (1:3)`.
`TestUpstreamNumericSeparatorGap` also proves `[1_]` and `[0x_1]`, under both
`probe.json` and `package.json`. All six are accepted by Go and refused by Prettier.
Inspection of `estree/parse_json.go` shows underscores stripped without validating
their positions. That explains the observed acceptance; this is upstream parser
behavior, not an observed stage 0 compiler defect.

The generated corpus also proves empty input under both filenames. Go refuses
with `json: empty input`; Prettier succeeds with an empty output. The strict audit
keeps this distinction rather than treating refusal as successful empty output.

The full audit walked every `.json` file in the checkout and both submodules,
excluding only `.git` directories. Its 1,075 inputs comprise 4 Adamic files
(including the new proving fixture), 835 cohere files, 202 TypeScript files, and
34 generated cases. The generated cases include 500 levels of nesting, a
5,000-element array, comments, trailing commas, Unicode and surrogate escapes,
decimals/exponents/precision extremes, hexadecimal/octal/binary forms, numeric
separators, nonfinite literals, incomplete input, and empty input, each under
ordinary and stringify filenames. Non-UTF-8 files fail the audit explicitly.

All **1,040 pre-existing repository files** agreed in acceptance and successful
bytes. The strict audit failed on **9/1,075** inputs: the new proving fixture,
six generated separator cases, and two empty cases. These are observations at
the pins above, not proof that all other possible inputs agree.

## Reproduction

From the repository root, after `bash cloud/setup.sh`:

```sh
source /workspace/adamic-tools/env.sh
npm install --prefix /tmp/adamic-json-prettier --save-exact prettier@3.9.6 > /tmp/adamic-json-npm.log 2>&1
ADAMIC_JSON_PRETTIER=/tmp/adamic-json-prettier go test -v -count=1 -timeout 30m ./stage1/cohere/json > /tmp/adamic-json-checks.log 2>&1
ADAMIC_JSON_PRETTIER=/tmp/adamic-json-prettier ADAMIC_JSON_CORPUS=1 ADAMIC_JSON_REPORT=/tmp/adamic-json-corpus-differences.txt go test -v -count=1 -timeout 30m ./stage1/cohere/json > /tmp/adamic-json-audit.log 2>&1
```

The first test command passed in 6.838 seconds. The opt-in corpus command failed
as expected, reporting every disagreement without exempting cases. Without the
Prettier environment variable the external tests skip; without the corpus
variable the expensive, currently red corpus test skips. Go's overlay injects
the test helper and mutants into scratch copies; it does not edit the submodule.

## Three caught mutants

`TestExternalComparisonCatchesThreePrinterMutants` first checks four successful
controls against actual Prettier, then changes actual Go source through overlays:

1. Remove `hardline` from the stringify root: both `package.json` controls lose
   their final newline and fail byte comparison.
2. Change property `": "` to `":"`: both stringify controls lose spaces after
   colons and fail byte comparison.
3. Select `json` instead of `json-stringify`: both stringify controls become
   compact layouts and fail byte comparison.

Every mutated Go formatter compiled, exited successfully, and produced output;
each mutant was caught on two inputs. These demonstrate the external comparison,
**not** Adamic compiler or port checks. No port mutants were run.

## Timing and remaining work

One corpus round, excluding process startup/build: Go cohere formatted 1,075 texts in 9.738 seconds (**110.39 texts/second**);
Node running external Prettier took 20.617 seconds (**52.14 texts/second**).
Go timing covers its format loop; Node timing also includes case-file parsing
and answer-file serialization/writing. These include refused inputs and generated
cases, and are single-round observations. They are not native Adamic versus Node-port performance numbers.

Setup succeeded: Go ready 0s; clang ready 0s; Node ready 0s; submodules ready 0s;
build cache warm 13s; total 13s. `nproc` was 5; CPU quota was 4 CPUs.

Remaining: resolve or explicitly scope upstream parity disagreements, then build
the Adamic parser/printer and stdout driver, compile with stage 0, compare native
and Node outputs against both oracles, exercise native memory/sanitizer checks,
run three port mutants, and measure native/Node-port throughput. This audit
establishes none of those results and changes no compiler files.
