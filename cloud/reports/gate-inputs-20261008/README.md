# Gate input verification — 2026-10-08

Base: `origin/devtools/area-merge-main-74fb` at `a2f2edfce70b8d357a10e6d67b3c7942b4580ab6`. The base already implemented the pinned installers and checked-in npm locks. This change reuses them, makes the gitignore export exactly `1`, documents the opt-in, and explicitly labels cache hits as doing no network work.

All 18 selected tests passed with inputs. No correctness failures were found. All ten explicitly requested inputs have a SKIP-to-PASS proof. Additional inherited formatter and ledger exports were exercised too. CSS numbers, CSS strings and Markdown inline log an absent external comparison instead of skipping; CSS fixtures and ledger output are supporting inputs exercised by their parent harnesses.

The driver ran from the repository root after `source /workspace/adamic-tools/env.sh`. For each baseline it removed all selected gate exports from that sourced environment. With-input runs retain setup exports. Commands use exact anchored names, `-v -count=1`, and a 20-minute timeout; no test source was edited. The driver and verbose logs are retained here.

## Named proofs

### ADAMIC_TYPESCRIPT_SOURCE

`go test ./stage1/typescript/parser -run ^TestWholeCompilerAgrees$ -v -count=1 -timeout=20m`

Without inputs: `--- SKIP: TestWholeCompilerAgrees (0.00s)`

With setup env.sh:

```text
=== RUN   TestWholeCompilerAgrees
--- PASS: TestWholeCompilerAgrees (59.70s)
```

[Without log](stage1-typescript-parser-TestWholeCompilerAgrees-without.log) · [With log](stage1-typescript-parser-TestWholeCompilerAgrees-with.log)

### ADAMIC_CSS_LIBRARY ADAMIC_CSS_FIXTURES

`go test ./stage1/cohere/css -run ^TestThePortParsesAsGoCohereDoes$/^PostCSS$ -v -count=1 -timeout=20m`

Without inputs: `--- SKIP: TestThePortParsesAsGoCohereDoes/PostCSS (0.00s)`

With setup env.sh:

```text
=== RUN   TestThePortParsesAsGoCohereDoes/PostCSS
--- PASS: TestThePortParsesAsGoCohereDoes/PostCSS (2.73s)
```

[Without log](stage1-cohere-css-TestThePortParsesAsGoCohereDoes-PostCSS-without.log) · [With log](stage1-cohere-css-TestThePortParsesAsGoCohereDoes-PostCSS-with.log)

### ADAMIC_CSS_PRINTER_LIBRARY

`go test ./stage1/cohere/css -run ^TestCSSPrinterBoundaryProofs$ -v -count=1 -timeout=20m`

Without inputs: `--- SKIP: TestCSSPrinterBoundaryProofs (19.67s)`

With setup env.sh:

```text
=== RUN   TestCSSPrinterBoundaryProofs
--- PASS: TestCSSPrinterBoundaryProofs (1.85s)
```

[Without log](stage1-cohere-css-TestCSSPrinterBoundaryProofs-without.log) · [With log](stage1-cohere-css-TestCSSPrinterBoundaryProofs-with.log)

### ADAMIC_JSON_PRETTIER

`go test ./stage1/cohere/json -run ^TestUpstreamNumericSeparatorGap$ -v -count=1 -timeout=20m`

Without inputs: `--- SKIP: TestUpstreamNumericSeparatorGap (0.00s)`

With setup env.sh:

```text
=== RUN   TestUpstreamNumericSeparatorGap
--- PASS: TestUpstreamNumericSeparatorGap (13.60s)
```

[Without log](stage1-cohere-json-TestUpstreamNumericSeparatorGap-without.log) · [With log](stage1-cohere-json-TestUpstreamNumericSeparatorGap-with.log)

### ADAMIC_GRAPHQL_LIBRARY

`go test ./stage1/cohere/graphql -run ^TestThePortParsesAsGoCohereDoes$/^as_graphql\-js$ -v -count=1 -timeout=20m`

Without inputs: `--- SKIP: TestThePortParsesAsGoCohereDoes/as_graphql-js (0.00s)`

With setup env.sh:

```text
=== RUN   TestThePortParsesAsGoCohereDoes/as_graphql-js
--- PASS: TestThePortParsesAsGoCohereDoes/as_graphql-js (0.80s)
```

[Without log](stage1-cohere-graphql-TestThePortParsesAsGoCohereDoes-as_graphql-js-without.log) · [With log](stage1-cohere-graphql-TestThePortParsesAsGoCohereDoes-as_graphql-js-with.log)

### ADAMIC_MEDIA_QUERY_LIBRARY

`go test ./stage1/cohere/mediaquery -run ^TestThePortParsesAsGoCohereDoes$/^as_postcss\-media\-query\-parser$ -v -count=1 -timeout=20m`

Without inputs: `--- SKIP: TestThePortParsesAsGoCohereDoes/as_postcss-media-query-parser (0.00s)`

With setup env.sh:

```text
=== RUN   TestThePortParsesAsGoCohereDoes/as_postcss-media-query-parser
--- PASS: TestThePortParsesAsGoCohereDoes/as_postcss-media-query-parser (0.68s)
```

[Without log](stage1-cohere-mediaquery-TestThePortParsesAsGoCohereDoes-as_postcss-media-query-parser-without.log) · [With log](stage1-cohere-mediaquery-TestThePortParsesAsGoCohereDoes-as_postcss-media-query-parser-with.log)

### ADAMIC_SELECTOR_LIBRARY

`go test ./stage1/cohere/selector -run ^TestTheLibraryDoesNotReturnOnUnconsumedNamespaceBars$ -v -count=1 -timeout=20m`

Without inputs: `--- SKIP: TestTheLibraryDoesNotReturnOnUnconsumedNamespaceBars (0.00s)`

With setup env.sh:

```text
=== RUN   TestTheLibraryDoesNotReturnOnUnconsumedNamespaceBars
--- PASS: TestTheLibraryDoesNotReturnOnUnconsumedNamespaceBars (0.00s)
```

[Without log](stage1-cohere-selector-TestTheLibraryDoesNotReturnOnUnconsumedNamespaceBars-without.log) · [With log](stage1-cohere-selector-TestTheLibraryDoesNotReturnOnUnconsumedNamespaceBars-with.log)

### ADAMIC_VALUES_LIBRARY

`go test ./stage1/cohere/values -run ^TestThePortParsesAsGoCohereDoes$/^as_postcss\-values\-parser$ -v -count=1 -timeout=20m`

Without inputs: `--- SKIP: TestThePortParsesAsGoCohereDoes/as_postcss-values-parser (0.00s)`

With setup env.sh:

```text
=== RUN   TestThePortParsesAsGoCohereDoes/as_postcss-values-parser
--- PASS: TestThePortParsesAsGoCohereDoes/as_postcss-values-parser (1.15s)
```

[Without log](stage1-cohere-values-TestThePortParsesAsGoCohereDoes-as_postcss-values-parser-without.log) · [With log](stage1-cohere-values-TestThePortParsesAsGoCohereDoes-as_postcss-values-parser-with.log)

### ADAMIC_GITIGNORE_LARGEST

`go test ./stage1/cohere/gitignore -run ^TestThePortAnswersAsGoCohereAndGitDo$/^catches_R2_the_size_limit_one_byte_lower$ -v -count=1 -timeout=20m`

Without inputs: `--- SKIP: TestThePortAnswersAsGoCohereAndGitDo/catches_R2_the_size_limit_one_byte_lower (0.00s)`

With setup env.sh:

```text
=== RUN   TestThePortAnswersAsGoCohereAndGitDo/catches_R2_the_size_limit_one_byte_lower
--- PASS: TestThePortAnswersAsGoCohereAndGitDo/catches_R2_the_size_limit_one_byte_lower (23.37s)
```

[Without log](stage1-cohere-gitignore-TestThePortAnswersAsGoCohereAndGitDo-catches_R2_the_size_limit_one_byte_lower-without.log) · [With log](stage1-cohere-gitignore-TestThePortAnswersAsGoCohereAndGitDo-catches_R2_the_size_limit_one_byte_lower-with.log)

### ADAMIC_CLANG_TSGO_ARCHIVE

`go test ./internal/native -run ^TestSplitTSGoAgrees$ -v -count=1 -timeout=20m`

Without inputs: `--- SKIP: TestSplitTSGoAgrees (0.00s)`

With setup env.sh:

```text
=== RUN   TestSplitTSGoAgrees
--- PASS: TestSplitTSGoAgrees (46.55s)
```

[Without log](internal-native-TestSplitTSGoAgrees-without.log) · [With log](internal-native-TestSplitTSGoAgrees-with.log)

### ADAMIC_GRAPHQL_PRETTIER

`go test ./stage1/cohere/graphql/printer -run ^TestPrinterUpstreamPreflight$ -v -count=1 -timeout=20m`

Without inputs: `--- SKIP: TestPrinterUpstreamPreflight (0.00s)`

With setup env.sh:

```text
=== RUN   TestPrinterUpstreamPreflight
--- PASS: TestPrinterUpstreamPreflight (29.33s)
```

[Without log](stage1-cohere-graphql-printer-TestPrinterUpstreamPreflight-without.log) · [With log](stage1-cohere-graphql-printer-TestPrinterUpstreamPreflight-with.log)

### ADAMIC_ESTREE_LIBRARY

`go test ./stage1/cohere/estree -run ^TestOriginalLibraries$ -v -count=1 -timeout=20m`

Without inputs: `--- SKIP: TestOriginalLibraries (0.00s)`

With setup env.sh:

```text
=== RUN   TestOriginalLibraries
--- PASS: TestOriginalLibraries (6.12s)
```

[Without log](stage1-cohere-estree-TestOriginalLibraries-without.log) · [With log](stage1-cohere-estree-TestOriginalLibraries-with.log)

### ADAMIC_YAML_LIBRARY

`go test ./stage1/cohere/yaml -run ^TestBundledParserDifference$ -v -count=1 -timeout=20m`

Without inputs: `--- SKIP: TestBundledParserDifference (0.00s)`

With setup env.sh:

```text
=== RUN   TestBundledParserDifference
--- PASS: TestBundledParserDifference (0.15s)
```

[Without log](stage1-cohere-yaml-TestBundledParserDifference-without.log) · [With log](stage1-cohere-yaml-TestBundledParserDifference-with.log)

### ADAMIC_TS_PRETTIER

`go test ./stage1/cohere/tsprinter -run ^TestStatementUpstreamDifferences$ -v -count=1 -timeout=20m`

Without inputs: `--- SKIP: TestStatementUpstreamDifferences (14.24s)`

With setup env.sh:

```text
=== RUN   TestStatementUpstreamDifferences
--- PASS: TestStatementUpstreamDifferences (1.98s)
```

[Without log](stage1-cohere-tsprinter-TestStatementUpstreamDifferences-without.log) · [With log](stage1-cohere-tsprinter-TestStatementUpstreamDifferences-with.log)

### ADAMIC_CYCLE_LEDGER_ROOT ADAMIC_CYCLE_LEDGER_OUTPUT

`go test ./internal/lower -run ^TestOriginalCycleLedger$ -v -count=1 -timeout=20m`

Without inputs: `--- SKIP: TestOriginalCycleLedger (0.00s)`

With setup env.sh:

```text
=== RUN   TestOriginalCycleLedger
--- PASS: TestOriginalCycleLedger (4.07s)
```

[Without log](internal-lower-TestOriginalCycleLedger-without.log) · [With log](internal-lower-TestOriginalCycleLedger-with.log)

### ADAMIC_CSSNUMBERS_LIBRARY

`go test ./stage1/cohere/cssnumbers -run ^TestCSSNumbers$ -v -count=1 -timeout=20m`

Without inputs: `--- PASS: TestCSSNumbers (132.55s)`

With setup env.sh:

```text
=== RUN   TestCSSNumbers
--- PASS: TestCSSNumbers (125.19s)
```

[Without log](stage1-cohere-cssnumbers-TestCSSNumbers-without.log) · [With log](stage1-cohere-cssnumbers-TestCSSNumbers-with.log)

### ADAMIC_CSSSTRINGS_LIBRARY

`go test ./stage1/cohere/cssstrings -run ^TestCSSStrings$ -v -count=1 -timeout=20m`

Without inputs: `--- PASS: TestCSSStrings (108.21s)`

With setup env.sh:

```text
=== RUN   TestCSSStrings
--- PASS: TestCSSStrings (98.22s)
```

[Without log](stage1-cohere-cssstrings-TestCSSStrings-without.log) · [With log](stage1-cohere-cssstrings-TestCSSStrings-with.log)

### ADAMIC_MARKDOWNINLINE_LIBRARY

`go test ./stage1/cohere/markdowninline -run ^TestMarkdownInline$ -v -count=1 -timeout=20m`

Without inputs: `--- PASS: TestMarkdownInline (623.90s)`

With setup env.sh:

```text
=== RUN   TestMarkdownInline
--- PASS: TestMarkdownInline (651.80s)
```

[Without log](stage1-cohere-markdowninline-TestMarkdownInline-without.log) · [With log](stage1-cohere-markdowninline-TestMarkdownInline-with.log)

## Installer and network checks

- `bash -n cloud/setup.sh`: passed.
- `ADAMIC_SETUP_INTEGRATION=1 python3 cloud/test_gate_inputs.py`: 8 tests passed, including real archive rebuild.
- CSS setup: 6 tests passed; cycle ledger setup: 5 passed; archive mode selection: 3 passed.
- Cache mutants: all detected by the intended assertions; see `cache-mutants.log`.
- Final setup and its immediate warm repeat both succeeded. Every locked npm project, source corpus, and archive was reused by the warm repeat. See `setup.log` and `warm-setup.log`.
- An additional full warm run traced all descendants with `strace -f -e trace=connect` and denied curl calls: zero connect syscalls and zero curl invocations. See `network-proof.txt` and `traced-warm-setup.log`. Go module preparation prints its existing “downloading” label on some cache validations; the trace confirms it made no network connections.

## TypeScript unset mutant

After sourcing setup env.sh, only `ADAMIC_TYPESCRIPT_SOURCE` was unset. Ran `go test ./stage1/typescript/parser -run '^TestWholeCompilerAgrees$' -v -count=1 -timeout=10m`:

```text
=== RUN   TestWholeCompilerAgrees
    whole_test.go:24: set ADAMIC_TYPESCRIPT_SOURCE to the pinned v6.0.3 checkout
--- SKIP: TestWholeCompilerAgrees (0.00s)
PASS
ok  	github.com/system-inc/adamic/stage1/typescript/parser	0.093s
```

The skip comes from `compilerManifest`, reached by `TestWholeCompilerAgrees`.
