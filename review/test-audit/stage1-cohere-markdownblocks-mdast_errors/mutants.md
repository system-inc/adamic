| ID | Origin file:line | Change | Failing grouped rows |
|---|---|---|---|
| M1 | stage1/cohere/markdownblocks/mdastCompile.ts:620 | `const root = this.arena.add('root', true, false); -> const root = this.arena.add('root', false, false);` | TestMdastIdentifierWitnesses, TestNativeMdastConstruction |
| M2 | stage1/cohere/markdownblocks/astPath.ts:82 | `return this.stack.length === 1; -> return this.stack.length === 0;` | TestMarkdownAstPath |
| M3 | stage1/cohere/markdownblocks/preprocess.ts:39 | `code: 65533 -> code: 65534` | TestMarkdownParserPrefixes |
| M4 | stage1/cohere/markdownblocks/quotes.ts:44 | `arena.text('> ') -> arena.text('# ')` | TestMarkdownQuoteLayout family |
| P1 | stage1/cohere/markdownblocks/mdastCompile.ts:619 | `empty entry` | TestMdastIdentifierWitnesses, TestNativeMdastConstruction |
| P2 | stage1/cohere/markdownblocks/pathObserve.ts:72 | `empty entry` | TestMarkdownAstPath |
| P3 | stage1/cohere/markdownblocks/preprocess.ts:7 | `empty entry` | TestMarkdownParserPrefixes |
| P4 | stage1/cohere/markdownblocks/quotes.ts:27 | `empty entry` | TestMarkdownQuoteLayout family |
| W1 | stage1/cohere/markdownblocks/preflight_shards_test.go:260 | `weaken witnessed comparison` | TestWholeDocumentOraclePreflight family |
| W2 | stage1/cohere/markdownblocks/preflight_shards_test.go:451 | `weaken witnessed comparison` | TestWholeDocumentOraclePreflightMutants |
| W3 | stage1/cohere/markdownblocks/quote_layout_shards_test.go:172 | `weaken witnessed comparison` | TestMarkdownQuoteLayout family |
| G2 | stage1/cohere/markdownblocks/quote_layout_shards_test.go:81 | `drop constructed product value` | TestMarkdownQuoteLayoutNative, TestMarkdownQuoteLayout_Setup |
| G3 | stage1/cohere/markdownblocks/sample_test.go:51 | `drop generated input condition` | TestGeneratedLayoutSelection, TestSampleRetainsFixedMarkdownInputs |
| G1 | stage1/cohere/markdownblocks/malformed_events_independent_test.go:374 | `change construction build option` | TestMdastMalformedEvents_Setup |
| P5 | internal/lower/lower.go:20 | `empty Lower entry` | TestOptionalStringInitializationWitness |
| P6 | stage1/cohere/markdownblocks/sample_test.go:35 | `empty construction entry` | TestGeneratedLayoutSelection, TestSampleRetainsFixedMarkdownInputs |
| P7 | stage1/cohere/markdownblocks/sample_test.go:58 | `empty construction entry` | TestGeneratedLayoutSelection, TestSampleRetainsFixedMarkdownInputs |
| P8 | stage1/cohere/markdownblocks/sample_test.go:104 | `empty construction entry` | TestSampleRetainsFixedMarkdownInputs |
| P9 | stage1/cohere/markdownblocks/sample_test.go:16 | `empty construction entry` | TestGeneratedLayoutSelection |
| P10 | stage1/cohere/markdownblocks/malformed_events_independent_test.go:238 | `empty construction entry` |  |
| P11 | stage1/cohere/markdownblocks/quote_layout_shards_test.go:69 | `empty construction entry` |  |
| P12 | stage1/cohere/markdownblocks/quote_layout_shards_test.go:107 | `empty construction entry` |  |
