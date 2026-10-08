# Step 26 (#35e525f): import clauses and attributes ported; both expression regressions closed

Default, namespace, named and type clauses, plus `with` attributes, use the existing indexed parser and document engine.
The source-context merge closes both reported object-method regressions, with their shortest retained inputs checked independently.
Whole-file Go identity over the same 152,660 pinned files rises 114 → 116 → 117 → 124 → 140; attributes retain 140.
61 new import fixtures match Go and Prettier; seven new import mutants, four retained import mutants and a context mutant require normal completion.
The full package is green with zero skips; nine prior byte differences remain and four invalid Babel imports raise acceptance conflicts 17 → 21.

## Pins, input contract and reproducibility

Base: origin/area/stage1-lint `9156bf5c579a44d687c9955d13e44f9ad8bbb6f8` (main `3ffb1a83` included).
Go cohere: `7945d102a6c18dd36adf9114a758ce646e8b2359`; its TypeScript submodule: `d92d9bfee114c80be2c375d72edae966176e3a4f`.
npm Prettier 3.9.6 is independent. TypeScript 6.0.3 source pin: `050880ce59e30b356b686bd3144efe24f875ebc8`.

The quiet hundred lives on Kirk's Mac. The corpus for this brief is **all 23 public pins supplied by Kirk**: 127,288 `.ts` and 25,372 `.tsx` files.
Each repository was shallow-fetched at its SHA into scratch, without installing or running its code. Scratch materialization was made sparse to TS/TSX; tracked paths and HEAD pins were checked, and every measured source has a SHA-256.
The npm oracle installation and the Adamic/Go test runners are separate from corpus code.
TypeScript compiler tests, malformed Babel fixtures and declarations remain in the denominator. Fifteen non-UTF-8 files and two isolated Node panics are recorded, not skipped.
No tsconfig, ignore rules or dependency installation narrows this source census.

All comparisons use Go's **house options** (width 120, tab width 4, spaces, single quotes, semicolons, trailingComma all, bracketSpacing true, bracketSameLine false, arrowParens always, LF).
Repository configuration resolution is surveyed but not composed into the port. This is not a claim of per-repository configured formatting parity.
`formatFile` now accepts a filename so TSX uses the existing parser's TSX mode; the old statement CLI still defaults to `statements.ts`.

`scout/surface.py` regenerates physical source counts and lexical function/member-read sites.
`scout/measure.py <checkout-manifest> <report> --before` runs the unchanged baseline gates; omit `--before` for the port.
The checkout manifest schema is repo, pin, checkout, files. `scout/manifest.json` retains all pins and counts; `before.json.gz`/`after.json.gz` retain every file, source hash, outcome and accepted output.
`testdata/scout_corpus_side_test.go` is overlaid into Go native and produces `scout/go.jsonl.gz`: every file's original Go AST kinds, formatting refusal or output hash.
`scout/summary.json` retains every mismatch and Go-refused input accepted by Node. Hash equality was computed over the actual formatted UTF-8 bytes, without newline normalization.
Batch paths are unique to their report. Every process failure was replayed on its own file in the final post-change census. Earlier overlapping scratch runs were discarded.

## Whole-file evidence

Go accepts 150,363 files and refuses 2,297. The original scout accepted 99 on Node: 73 matched Go bytes, nine differed, and 17 conflicted with Go acceptance. Side-effect imports raised identity to 114 (140 accepted). The method-shorthand merge leaves those whole-file counts unchanged.

| Cumulative import capability | Files accepted by Node | Whole files identical to Go | Output differences | Go refuses, Node accepts |
| --- | ---: | ---: | ---: | ---: |
| Side effects + merged expression context | 140 | 114 | 9 | 17 |
| Default | 142 | 116 | 9 | 17 |
| Namespace | 143 | 117 | 9 | 17 |
| Named | 151 | 124 | 9 | 18 |
| Type clauses and specifiers | 170 | 140 | 9 | 21 |
| Import attributes | 170 | 140 | 9 | 21 |

The default/namespace/named rows are explicit capability ablations of the final printer on source Node (`scout/stages.mjs`), not claims of separate implementation commits. Each runs all 152,660 files. The type row is the actual pre-attributes census; the attributes row is the actual final census. `scout/compare-stages.py` checks identical path sets and source SHA-256 for every row before comparing output hashes to Go. `scout/import-stages.json` retains every reason count, difference, conflict and change between rows; `side-effect.json.gz` and final `after.json.gz` retain the complete outcomes. All 26 newly Go-identical whole files are present, without exemptions.

Attributes add no accepted whole file in this corpus: their containing files meet another existing refusal. The 61-fixture gate independently covers their output, including exact public and TypeScript test excerpts. A green fixture is never counted as a green complete source file.

The raw-marker and source-trivia gates still account for 94,437 and 52,602 files. Fifteen non-UTF-8 files and both Node panics stay in the denominator. The final total of 170 accepted files is **140 identical + 9 differing + 21 acceptance conflicts**, not 170 identical.

First refusal is a disjoint execution classification; AST presence below is overlapping and unmasks constructs behind earlier gates. Neither is a percentage of implemented TypeScript.

### Refusals by exact reason

| Reason | Original scout before | Current after clauses/attributes |
| --- | ---: | ---: |
| `comment-attachment` | 94437 | 94437 |
| `source-trivia` | 52602 | 52602 |
| `variable-modifiers` | 1780 | 1833 |
| `ExportDeclaration` | 1457 | 1467 |
| `TypeAliasDeclaration` | 839 | 845 |
| `function-types` | 265 | 270 |
| `EnumDeclaration` | 238 | 240 |
| `ClassDeclaration` | 210 | 234 |
| `InterfaceDeclaration` | 185 | 188 |
| `accepted` | 99 | 170 |
| `ModuleDeclaration` | 103 | 103 |
| `variable-types` | 67 | 74 |
| `ExportAssignment` | 32 | 37 |
| `ImportEqualsDeclaration` | 21 | 21 |
| `TypeAssertionExpression` | 19 | 19 |
| `assignment-pattern` | 16 | 16 |
| `input-not-utf8` | 15 | 15 |
| `type-arguments` | 15 | 15 |
| `AsExpression` | 14 | 14 |
| `parameter-pattern` | 12 | 14 |
| `variable-pattern` | 9 | 9 |
| `ClassExpression` | 5 | 5 |
| `ExpressionWithTypeArguments` | 5 | 5 |
| `FunctionExpression` | 0 | 5 |
| `ForStatement` | 3 | 3 |
| `IfStatement` | 3 | 3 |
| `tag-type-arguments` | 3 | 3 |
| `ForOfStatement` | 2 | 2 |
| `MissingDeclaration` | 2 | 2 |
| `ImportDeclaration` | 195 | 1 |
| `ImportKeyword` | 1 | 1 |
| `JsxElement` | 0 | 1 |
| `JsxSelfClosingElement` | 1 | 1 |
| `NamespaceExportDeclaration` | 1 | 1 |
| `SatisfiesExpression` | 1 | 1 |
| `TryStatement` | 1 | 1 |
| `node-exit-70:adamic: panic: JSX expected name at 10 in packages/babel-parser/test/fixtures/typescript/type-arguments-bit-shift-left-like/jsx-opening-element/input.tsx` | 1 | 1 |
| `node-exit-70:adamic: panic: missing variable declaration` | 1 | 1 |

### Original unported AST constructs ranked by files affected

| Construct | Files | Sites |
| --- | ---: | ---: |
| ImportDeclaration | 109305 | 557705 |
| TypeReference | 95267 | 1331478 |
| IfStatement | 41259 | 315114 |
| ClassDeclaration | 31134 | 57410 |
| AsExpression | 27548 | 169742 |
| TypeAliasDeclaration | 25635 | 82271 |
| JsxSelfClosingElement | 19842 | 94356 |
| JsxElement | 19037 | 182890 |
| InterfaceDeclaration | 18654 | 57505 |
| ForOfStatement | 11785 | 33278 |
| TryStatement | 9866 | 20015 |
| ExportAssignment | 7883 | 8384 |
| ExportDeclaration | 6830 | 24764 |
| ModuleDeclaration | 4746 | 10942 |
| JsxFragment | 4471 | 7500 |
| ForStatement | 4240 | 9435 |
| SwitchStatement | 3664 | 7029 |
| EnumDeclaration | 2495 | 4522 |
| WhileStatement | 2068 | 4197 |
| SatisfiesExpression | 1969 | 3773 |
| ClassExpression | 924 | 1970 |
| TypeAssertionExpression | 824 | 2783 |
| ForInStatement | 719 | 1354 |
| DoStatement | 508 | 873 |
| LabeledStatement | 309 | 1156 |

This table records the initial import ranking that selected this step. Import clauses and attributes are now ported; source comments/trivia still mask them. `TypeReference` itself occurs in 95,267 files; type printing is also absent.
All original AST-kind presence/site counts are retained in summary.json, not just this declaration/statement/JSX shortlist.

## Entire Go formatter surface

Physical lines include comments, blanks, generated files and platform variants. Testdata is excluded; tests are counted separately. This is an inventory, not semantic line credit.
Go formatter total: 290 production Go files / 66,167 lines; 95 test files / 21,623 lines.
Every file, function declaration line and option member-read candidate is in `scout/surface.json`.

| Go package | Production files | Lines | Port and boundary |
| --- | ---: | ---: | --- |
| arena | 5 | 169 | stage1/cohere/tsprinter/doc.ts; indexed document representation, not Go slab API |
| comparison | 1 | 133 | No dedicated port; Go comparison machinery |
| css | 23 | 4188 | stage1/cohere/css; CSS/SCSS parser and printer; default options |
| css/mediaquery | 4 | 592 | stage1/cohere/mediaquery; media query parser |
| css/postcss | 9 | 2328 | stage1/cohere/css; CSS/SCSS parser and printer; default options |
| css/selector | 5 | 1286 | stage1/cohere/selector; selector parser |
| css/values | 4 | 1202 | stage1/cohere/values; value parser |
| differential | 1 | 413 | No dedicated port; Go differential machinery |
| doc | 7 | 1921 | stage1/cohere/tsprinter/doc.ts; document algebra; settings: width, tabs, indentation |
| doc/tools/generate_string_width | 1 | 247 | stage1/cohere/tsprinter/doc.ts; document algebra; settings: width, tabs, indentation |
| estree | 9 | 4240 | stage1/cohere/estree; non-JSON conversion; tsprinter uses parser directly |
| formatfiles | 1 | 544 | stage1/cohere/formatfiles; enumeration slice |
| formatoptions | 2 | 540 | No dedicated port; no dedicated options resolver port |
| graphql | 9 | 3238 | stage1/cohere/graphql; parser plus graphql/printer |
| javascript | 73 | 14122 | stage1/cohere/tsprinter; partial expressions and statements; no complete TS/TSX contract |
| markdown | 16 | 2959 | stage1/cohere/markdownblocks; partial construction/layout; embedding off |
| markdown/mdast | 4 | 1448 | stage1/cohere/markdownblocks; construction from Go resolved events, not source parser |
| markdown/micromark | 52 | 12914 | stage1/cohere/markdownblocks; event primitives; grammar unfinished |
| markdown/micromark/tools/generate_characters | 1 | 166 | stage1/cohere/markdownblocks; event primitives; grammar unfinished |
| markdown/micromark/tools/generate_entities | 1 | 137 | stage1/cohere/markdownblocks; event primitives; grammar unfinished |
| markdown/tools/generate_classes | 1 | 253 | stage1/cohere/markdownblocks; partial construction/layout; embedding off |
| markdown/tools/generate_languages | 1 | 144 | stage1/cohere/markdownblocks; partial construction/layout; embedding off |
| native | 9 | 496 | No dedicated port; no composed multi-language router port |
| oracletest | 1 | 221 | No dedicated port; Go oracle helpers |
| prettier | 7 | 638 | No dedicated port; Go oracle and bundled fork, not production Adamic |
| printing | 7 | 1630 | No dedicated port; no generic shared print/comment/embedding core port |
| yaml/compose | 15 | 4014 | stage1/cohere/yaml; parser/printer, default options |
| yaml | 8 | 1629 | stage1/cohere/yaml; parser/printer, default options |
| yaml/cst | 5 | 2287 | stage1/cohere/yaml; parser/printer, default options |
| yaml/unist | 8 | 2068 | stage1/cohere/yaml; parser/printer, default options |


Port physical TypeScript source counts (excludes testdata/gaps/scout; not semantic credit):

| Slice | Files | Lines |
| --- | ---: | ---: |
| stage1/cohere/tsprinter | 14 | 5601 |
| stage1/cohere/estree | 34 | 9027 |
| stage1/cohere/markdownblocks | 49 | 10322 |
| stage1/cohere/markdowninline | 3 | 327 |
| stage1/cohere/graphql | 9 | 3313 |
| stage1/cohere/yaml | 43 | 8196 |
| stage1/cohere/css | 16 | 3234 |
| stage1/cohere/json | 7 | 4240 |
| stage1/cohere/formatfiles | 4 | 831 |

### Every language route and embedding boundary

`cohere/internal/format/native/typescript.go:27`: `.ts`, `.tsx`, `.a` => javascript/ESTree, with parsed-tree and doc entries. Port: tsprinter (partial); estree conversion is separate and not its printer input.
`native/javascript.go:17`: `.js`, `.mjs`, `.cjs`, `.jsx` => Babel compatibility adapter plus the same JavaScript printer. No composed Adamic JS/Babel driver.
`native/json.go:13`: `.json` => ordinary JSON or json-stringify by filename (`package.json`, package-lock etc.). Port: stage1/cohere/json, separate from tsprinter's TypeScript driver and estree's non-JSON slice.
`native/css.go:22`: `.css` => CSS; `.scss` at :19 is embedding-only. Port: css plus selector/mediaquery/values/cssnumbers/cssstrings; no Less parser route.
`native/graphql.go:25`: `.graphql`, `.gql` => GraphQL; port: graphql/printer with all 45 visitor kinds and standalone normalization.
`native/yaml.go:23`: `.yaml`, `.yml` => YAML, raw doc entries for embedding; port: yaml default-option complete slice.
`native/markdown.go:36`: `.md` => Markdown plus embedding callback. Port: markdownblocks/markdowninline; no `stage1/cohere/markdown` directory and no complete native Markdown source-parser/formatter composition.

`native/native.go:81` maps 12 embedding parser names. json5/jsonc/less have no native route despite being named there. `javascript/embed.go:16` embeds CSS/SCSS and GraphQL, while HTML/Angular and JS-template Markdown remain unsupported/raw on failure; whitespace-only embed handling still runs.
Markdown fences and YAML/TOML front matter go through `markdown/embed.go`; TOML has no native printer. A named language or an inference-table entry does not establish a native printer.
The port has no multi-language router corresponding to native and no generic comment/embedding core corresponding to printing. `markdownblocks/GAPS.md:3` explicitly excludes embedding and retains Go-supplied tokenizer events.

### Every JavaScript/TypeScript printer file

These are partial family mappings, never whole-file coverage credits. Literal, object, array and document machinery is also reused by the separate JSON port.

| Go printer | Lines | Port status |
| --- | ---: | --- |
| `internal/format/javascript/embed.go:1` | 522 | Unported composition |
| `internal/format/javascript/print.go:1` | 416 | Partial: tsprinter expressions/syntax/parentheses |
| `internal/format/javascript/print_array.go:1` | 246 | Partial: tsprinter expressions/syntax/parentheses |
| `internal/format/javascript/print_arrow_function.go:1` | 288 | Partial: tsprinter expressions/syntax/parentheses |
| `internal/format/javascript/print_assignment.go:1` | 441 | Partial: tsprinter expressions/syntax/parentheses |
| `internal/format/javascript/print_binary_cast_expression.go:1` | 32 | Unported |
| `internal/format/javascript/print_binaryish.go:1` | 287 | Partial: tsprinter expressions/syntax/parentheses |
| `internal/format/javascript/print_block.go:1` | 110 | Partial: tsprinter expressions/syntax/parentheses |
| `internal/format/javascript/print_call_arguments.go:1` | 413 | Partial: tsprinter expressions/syntax/parentheses |
| `internal/format/javascript/print_call_expression.go:1` | 179 | Partial: tsprinter expressions/syntax/parentheses |
| `internal/format/javascript/print_class.go:1` | 335 | Unported |
| `internal/format/javascript/print_class_body.go:1` | 270 | Unported |
| `internal/format/javascript/print_decorators.go:1` | 76 | Unported |
| `internal/format/javascript/print_expression_statement.go:1` | 120 | Partial: tsprinter expressions/syntax/parentheses |
| `internal/format/javascript/print_expressions.go:1` | 321 | Partial: tsprinter expressions/syntax/parentheses |
| `internal/format/javascript/print_function.go:1` | 270 | Partial: tsprinter expressions/syntax/parentheses |
| `internal/format/javascript/print_function_parameters.go:1` | 315 | Partial: tsprinter expressions/syntax/parentheses |
| `internal/format/javascript/print_function_type.go:1` | 48 | Unported |
| `internal/format/javascript/print_json.go:1` | 113 | Separate stage1/cohere/json slice |
| `internal/format/javascript/print_jsx.go:1` | 431 | Unported |
| `internal/format/javascript/print_jsx_children.go:1` | 507 | Unported |
| `internal/format/javascript/print_literal.go:1` | 73 | Partial: tsprinter expressions/syntax/parentheses |
| `internal/format/javascript/print_mapped_type.go:1` | 120 | Unported |
| `internal/format/javascript/print_member.go:1` | 78 | Partial: tsprinter expressions/syntax/parentheses |
| `internal/format/javascript/print_member_chain.go:1` | 473 | Partial: tsprinter expressions/syntax/parentheses |
| `internal/format/javascript/print_miscellaneous.go:1` | 158 | Partial: tsprinter expressions/syntax/parentheses |
| `internal/format/javascript/print_module.go:1` | 378 | Side-effect/default/namespace/named/type imports and with attributes; exports and comment/trivia composition remain unported |
| `internal/format/javascript/print_module_declaration.go:1` | 27 | Unported |
| `internal/format/javascript/print_object.go:1` | 184 | Partial: tsprinter expressions/syntax/parentheses |
| `internal/format/javascript/print_statements.go:1` | 278 | Partial: tsprinter expressions/syntax/parentheses |
| `internal/format/javascript/print_template_literal.go:1` | 371 | Partial: tsprinter expressions/syntax/parentheses |
| `internal/format/javascript/print_ternary.go:1` | 393 | Partial: tsprinter expressions/syntax/parentheses |
| `internal/format/javascript/print_type_alias.go:1` | 26 | Unported |
| `internal/format/javascript/print_type_annotation.go:1` | 112 | Unported |
| `internal/format/javascript/print_type_parameters.go:1` | 195 | Unported |
| `internal/format/javascript/print_typescript_types.go:1` | 260 | Unported |
| `internal/format/javascript/print_union_type.go:1` | 127 | Unported |
| `internal/format/javascript/printer.go:1` | 163 | Partial: tsprinter expressions/syntax/parentheses |

Other-language printer entry/functions and their exact source lines are enumerated in surface.json, including CSS print.go, GraphQL print.go, YAML printer.go, Markdown print.go/children/table/list/heading/paragraph/sentence/word and the shared doc interpreter. Auxiliary serializers, comments, parser adapters and generated tables are included in the package/file inventory rather than silently omitted.

### Every externally resolved formatter option

`formatoptions/options.go:15` has ten fields; `formatoptions/resolve.go:95` gives Prettier defaults. Counts below are lexical production member-read candidates, with all file:line locations in surface.json.

| Option | Prettier default / house | Sites | Files | tsprinter contract |
| --- | --- | ---: | ---: | --- |
| TabWidth | 2 / 4 | 20 | 13 | Exposed in SettingsOptions |
| UseTabs | false / false | 15 | 11 | Exposed in SettingsOptions |
| Semi | true / true | 12 | 8 | House value fixed; non-default variation unported |
| SingleQuote | false / true | 10 | 9 | House value fixed; non-default variation unported |
| PrintWidth | 80 / 120 | 11 | 10 | Exposed in SettingsOptions |
| TrailingComma | all / all | 7 | 6 | House value fixed; non-default variation unported |
| BracketSpacing | true / true | 8 | 8 | House value fixed; non-default variation unported |
| BracketSameLine | false / false | 3 | 3 | House value fixed; non-default variation unported |
| ArrowParens | always / always | 3 | 3 | House value fixed; non-default variation unported |
| EndOfLine | lf / lf | 7 | 4 | House value fixed; non-default variation unported |

Internal option declarations are retained verbatim with file:line in surface.json, including printing.Options, doc.Options, parser/schema/compose/group options. Parser/filepath/parentParser are derived adapter state rather than config fields. `printing/printer.go:103` supports embeddedLanguageFormatting auto/off; standalone Go routes set auto while the Markdown port's contract is off. Markdown `format.go:59` fixes proseWrap preserve; test-only helpers cover always/never. These do not appear in the ten-field resolver.
Go's config walk, house-tier enforcement, stale Prettier-config refusal, ignore/ignorePatterns and caching live in formatoptions/resolve.go and formatfiles/enumerate.go. There is no corresponding composed port option resolver.

## Original side-effect piece and design (bb8eb916)

`cohere/internal/format/javascript/print_module.go:19` prints keyword, phase/kind, specifiers, source, attributes and optional semicolon; :151 supplies the side-effect source's space. `print_literal.go:13` and utility_text.go select/canonicalize quotes.
At `bb8eb916`, `stage1/cohere/tsprinter/syntax.ts:481` admitted an ImportDeclaration only when it has exactly one StringLiteral child, then applied the existing literal refusal check.
At that commit, `stage1/cohere/tsprinter/expressions.ts:631` emitted `import `, the existing non-directive string printer, and `;` through the existing indexed document arena. SourceFile already ends directive mode when it sees a non-expression statement.
At that first increment, clauses, phases, attributes and comments were explicit refusals. The current clause/attribute design below replaces the first three boundaries.

The source contains no new side-effect import execution: the formatted user import is data. This does not add side-effect module loading to Adamic.
This uses existing valid TypeScript switch cases, checked numeric child indexes, readonly settings and existing literal/doc objects. It adds no syntax, reference ownership cycle, GC dependency, parser adaptation or language-gap workaround.
The optional filename parameter preserves `.tsx` mode without editing the parser.
The follow-up now ports clauses/specifiers, phase/type flags, attributes and width-dependent groups using the same numeric node links and doc algebra. Blank-line/comment decisions remain for the next slice.
Every newly encountered lowering/ownership question must become an open @system_adamic question rather than a source rewrite around a language gap.

## Hard boundaries and shortest inputs

`import 'x'` now matches Go and npm byte-for-byte: `import 'x';\n`.
`import {x} from 'x'`, `import type {T} from 'x'`, and `import 'x' with {type:'json'}` now match Go and npm. Legacy `assert` remains `ImportDeclaration`: the pinned Go parser rejects it (`cohere/TypeScript/tsc/internal/parser/parser.go:2573`), rather than formatting it.
`import 'http://x'` remains `comment-attachment` because files.ts tests raw markers even inside literals.
`// x` is still refused. A separate investigation proved comment-only files are independent, but that piece was discarded because the complete AST census ranked imports first.
`// x\t\u00a0` demonstrates a real upstream difference: Go retains the trailing tab/NBSP, npm removes them. Go `estree/postprocess.go:200` returns after testing the last UTF-8 continuation byte, before considering the multibyte whitespace. No exception was added to accepted imports.
`export default abstract;` is shortened from Babel's class/abstract-false-positive fixture. The shared parser collects export/default modifiers (`statements.ts:1210`) then falls through to an expression and loses them (:1295). Go retains the export. A broader leading-comment/body wrapper exposed this, so body composition was stopped.
Neither reported regression remains after merging `06c6e7cbf46e6d53d575c292e2a11733e36547c4`. Their shortest retained inputs are `({resolve(){return {foo:100};}})` (103_circularReferenceInReturnType2.ts:1078) and `({[z](){return z;}})` (193_capturedParametersInInitializers1.ts:857). Both now print Go's exact bytes on Node, native and emitted JavaScript. The responsible entry point is `expressions.ts:1882`; the fix opens the source list at :1894 and closes it at :1896. The old entry point called expression() with no source context, letting speculative arrow recovery consume the method body. `TestScoutSourceContext` removes that context as a mutant; it must compile and exit normally before the differing refusal counts as a kill. The full broad expression and focused tsc gates pass after the merge; no shared parser edit was needed.
`scout/parser-blockers.json` retains live Go/npm bytes and the Node expression-driver outcomes for these shortest blockers and the NBSP input.
The two isolated public Node panics and all existing malformed-input acceptance disagreements are retained in summary.json. They are not converted to NotYet or hidden behind a selector.

## Ruling and stop boundary

No language ruling is needed for these import clauses or attributes. They introduce no Adamic syntax, runtime/compiler change or ownership workaround.
@system_adamic resolved the method-origin question under the method-value ruling **#bw3xg7c**: a method passed through a readonly callback interface is the same method value, never a copy, so its prototype origin is preserved. A method that does not read `this` is a plain value. A method that reads `this` keeps its receiver only when bound. An unbound `this`-reading method is refused in `.a`; in `.ts` it follows Node, including TypeError. This is the supplied language ruling, not a decision by this scout. The existing `gaps/classInterfaceMethod.ts.txt:1` remains the shortest implementation proof (Node prints 17; stage 0 currently records NotYet). Compiler implementation is a separate lane; no compiler/runtime change or wrapper is introduced here. **No open language questions remain for this import piece.**
Parser-lane handoff (implementation, not a new language decision): preserve export/default modifiers on the fallback at `stage1/typescript/parser/statements.ts:1295`. The method regressions are closed by the authorized printer-only merge. Import-local validation remains a parser-lane handoff at `statements.ts:322`, :332, :342 and :348: propertyName() accepts reserved words and string literals in the local binding slot, while Go distinguishes imported names from local identifiers (`cohere/TypeScript/tsc/internal/parser/parser.go:2482`, :2547). Those shared files must be owned and changed by that lane. This scout stops before those changes.

## Fixture list and validation

All 69 inputs, exact Go/npm bytes, before/after Node outcomes and source provenance are in `scout/fixtures.json`; the list below identifies each fixture. Thirty are complete pinned public files; the TypeScript/Go entries are exact labelled excerpts, not fabricated production sources. The remaining shortest/boundary cases intentionally expose missing families.

| Fixture | Origin | Result |
| --- | --- | --- |
| supported-control | minimal reduction of formatter family; SCOUT.md maps source sites | byte-identical accepted |
| line-comment | minimal reduction of formatter family; SCOUT.md maps source sites | comment-attachment |
| line-comment-only | minimal reduction of formatter family; SCOUT.md maps source sites | comment-attachment |
| block-comment | minimal reduction of formatter family; SCOUT.md maps source sites | comment-attachment |
| comment-marker-string | minimal reduction of formatter family; SCOUT.md maps source sites | comment-attachment |
| blank-line | minimal reduction of formatter family; SCOUT.md maps source sites | source-trivia |
| crlf | minimal reduction of formatter family; SCOUT.md maps source sites | source-trivia |
| typed-variable | minimal reduction of formatter family; SCOUT.md maps source sites | variable-types |
| variable-pattern | minimal reduction of formatter family; SCOUT.md maps source sites | variable-pattern |
| variable-modifier | minimal reduction of formatter family; SCOUT.md maps source sites | variable-modifiers |
| if | minimal reduction of formatter family; SCOUT.md maps source sites | IfStatement |
| import | minimal reduction of formatter family; SCOUT.md maps source sites | byte-identical accepted |
| type-alias | minimal reduction of formatter family; SCOUT.md maps source sites | TypeAliasDeclaration |
| interface | minimal reduction of formatter family; SCOUT.md maps source sites | InterfaceDeclaration |
| class | minimal reduction of formatter family; SCOUT.md maps source sites | ClassDeclaration |
| jsx | minimal reduction of formatter family; SCOUT.md maps source sites | JsxSelfClosingElement |
| typed-function | minimal reduction of formatter family; SCOUT.md maps source sites | function-types |
| anonymous-function | minimal reduction of formatter family; SCOUT.md maps source sites | FunctionExpression |
| call-type-arguments | minimal reduction of formatter family; SCOUT.md maps source sites | type-arguments |
| as | minimal reduction of formatter family; SCOUT.md maps source sites | AsExpression |
| satisfies | minimal reduction of formatter family; SCOUT.md maps source sites | SatisfiesExpression |
| assignment-pattern | minimal reduction of formatter family; SCOUT.md maps source sites | assignment-pattern |
| parameter-pattern | minimal reduction of formatter family; SCOUT.md maps source sites | parameter-pattern |
| jest | minimal reduction of formatter family; SCOUT.md maps source sites | jest-template-table |
| import-boundary-0 | Go print_module.go:19 and print_literal.go:13 | byte-identical accepted |
| import-boundary-1 | Go print_module.go:19 and print_literal.go:13 | byte-identical accepted |
| import-boundary-2 | Go print_module.go:19 and print_literal.go:13 | byte-identical accepted |
| import-boundary-3 | Go print_module.go:19 and print_literal.go:13 | byte-identical accepted |
| import-boundary-4 | Go print_module.go:19 and print_literal.go:13 | byte-identical accepted |
| import-boundary-5 | Go print_module.go:19 and print_literal.go:13 | byte-identical accepted |
| import-boundary-6 | Go print_module.go:19 and print_literal.go:13 | byte-identical accepted |
| import-boundary-7 | Go print_module.go:19 and print_literal.go:13 | byte-identical accepted |
| import-boundary-8 | Go print_module.go:19 and print_literal.go:13 | byte-identical accepted |
| import-boundary-9 | Go print_module.go:19 and print_literal.go:13 | byte-identical accepted |
| import-boundary-10 | Go print_module.go:19 and print_literal.go:13 | byte-identical accepted |
| import-boundary-11 | Go print_module.go:19 and print_literal.go:13 | comment-attachment |
| import-boundary-12 | Go print_module.go:19 and print_literal.go:13 | source-trivia |
| real-source-0 | actualbudget/actual@9732a4463aac2909ac2aad1627085e0eb7a207b5:packages/desktop-client/src/globals.ts | byte-identical accepted |
| real-source-1 | angular/angular@c0dc8c4bbeea70879aef54e9fcc7888359dfd1a5:packages/zone.js/zone.ts | byte-identical accepted |
| real-source-2 | babel/babel@f67453d563918a1ea4a1dcb1d7af01d4a528d99b:packages/babel-parser/test/fixtures/typescript/import/import-side-effects/input.ts | byte-identical accepted |
| real-source-3 | backstage/backstage@532931240eec03b318efaadb39e24d9ee6d7076c:packages/cli-module-new/templates/frontend-plugin-module/src/setupTests.ts | byte-identical accepted |
| real-source-4 | backstage/backstage@532931240eec03b318efaadb39e24d9ee6d7076c:packages/cli-module-new/templates/frontend-plugin/src/setupTests.ts | byte-identical accepted |
| real-source-5 | backstage/backstage@532931240eec03b318efaadb39e24d9ee6d7076c:packages/cli-module-new/templates/legacy-frontend-plugin/src/setupTests.ts | byte-identical accepted |
| real-source-6 | backstage/backstage@532931240eec03b318efaadb39e24d9ee6d7076c:packages/cli-module-new/templates/plugin-web-library/src/setupTests.ts | byte-identical accepted |
| real-source-7 | backstage/backstage@532931240eec03b318efaadb39e24d9ee6d7076c:packages/cli-module-new/templates/scaffolder-field-extension-module/src/setupTests.ts | byte-identical accepted |
| real-source-8 | backstage/backstage@532931240eec03b318efaadb39e24d9ee6d7076c:packages/cli-module-new/templates/web-library/src/setupTests.ts | byte-identical accepted |
| real-source-9 | backstage/backstage@532931240eec03b318efaadb39e24d9ee6d7076c:packages/create-app/templates/default-app/packages/app/src/setupTests.ts | byte-identical accepted |
| real-source-10 | backstage/backstage@532931240eec03b318efaadb39e24d9ee6d7076c:packages/create-app/templates/legacy-app/packages/app/src/setupTests.ts | byte-identical accepted |
| real-source-11 | date-fns/date-fns@717ce0a807ea4c6b540d015b5c408723175b2838:pkgs/tz/test/engines/src/javascriptcore.ts | byte-identical accepted |
| real-source-12 | n8n-io/n8n@e77e30c7d92f4fbc34337f7bcc80ce3f52e5d073:packages/@n8n/utils/src/__tests__/setup.ts | byte-identical accepted |
| real-source-13 | n8n-io/n8n@e77e30c7d92f4fbc34337f7bcc80ce3f52e5d073:packages/cli/src/modules/dynamic-credentials.ee/context-establishment-hooks/index.ts | byte-identical accepted |
| real-source-14 | n8n-io/n8n@e77e30c7d92f4fbc34337f7bcc80ce3f52e5d073:packages/cli/src/modules/dynamic-credentials.ee/credential-resolvers/index.ts | byte-identical accepted |
| real-source-15 | n8n-io/n8n@e77e30c7d92f4fbc34337f7bcc80ce3f52e5d073:packages/cli/src/modules/instance-registry/checks/index.ts | byte-identical accepted |
| real-source-16 | n8n-io/n8n@e77e30c7d92f4fbc34337f7bcc80ce3f52e5d073:packages/frontend/@n8n/chat/src/__tests__/setup.ts | byte-identical accepted |
| real-source-17 | n8n-io/n8n@e77e30c7d92f4fbc34337f7bcc80ce3f52e5d073:packages/frontend/@n8n/composables/src/__tests__/setup.ts | byte-identical accepted |
| real-source-18 | n8n-io/n8n@e77e30c7d92f4fbc34337f7bcc80ce3f52e5d073:packages/frontend/@n8n/frontend-module-sdk/src/__tests__/setup.ts | byte-identical accepted |
| real-source-19 | n8n-io/n8n@e77e30c7d92f4fbc34337f7bcc80ce3f52e5d073:packages/frontend/@n8n/frontend-utils/src/__tests__/setup.ts | byte-identical accepted |
| real-source-20 | n8n-io/n8n@e77e30c7d92f4fbc34337f7bcc80ce3f52e5d073:packages/frontend/@n8n/i18n/src/__tests__/setup.ts | byte-identical accepted |
| real-source-21 | n8n-io/n8n@e77e30c7d92f4fbc34337f7bcc80ce3f52e5d073:packages/frontend/@n8n/rest-api-client/src/__tests__/setup.ts | byte-identical accepted |
| real-source-22 | n8n-io/n8n@e77e30c7d92f4fbc34337f7bcc80ce3f52e5d073:packages/frontend/@n8n/stores/src/__tests__/setup.ts | byte-identical accepted |
| real-source-23 | n8n-io/n8n@e77e30c7d92f4fbc34337f7bcc80ce3f52e5d073:packages/frontend/editor-ui/src/app/plugins/index.ts | byte-identical accepted |
| real-source-24 | nestjs/nest@35142c3eca8edaaf6abc5984d915da2fbd458aa2:tools/gulp/gulpfile.ts | byte-identical accepted |
| real-source-25 | prisma/prisma@c882b03377e70c090d7bbe05cf85bf04fe9e33b2:examples/prisma-8-demo/test/setup-temporal.ts | byte-identical accepted |
| real-source-26 | prisma/prisma@c882b03377e70c090d7bbe05cf85bf04fe9e33b2:examples/react-router-demo/test/setup-temporal.ts | byte-identical accepted |
| real-source-27 | prisma/prisma@c882b03377e70c090d7bbe05cf85bf04fe9e33b2:packages/3-targets/3-targets/postgres/test/setup-temporal.ts | byte-identical accepted |
| real-source-28 | prisma/prisma@c882b03377e70c090d7bbe05cf85bf04fe9e33b2:packages/3-targets/6-adapters/postgres-codec-testkit/test/setup-temporal.ts | byte-identical accepted |
| real-source-29 | prisma/prisma@c882b03377e70c090d7bbe05cf85bf04fe9e33b2:packages/3-targets/6-adapters/postgres/test/setup-temporal.ts | byte-identical accepted |
| typescript-corpus-imports | microsoft/TypeScript@050880ce59e30b356b686bd3144efe24f875ebc8:tests/cases/compiler/sideEffectImports1.ts:5 | byte-identical accepted |
| go-cohere-own-module-test | cohere@7945d102:internal/format/javascript/format_test.go:71 exact source prefix | ExportDeclaration |

Four retained mutants change the import keyword, separating space, semicolon or attribute document emission. Every mutant must compile, exit zero with empty stderr on Node and sanitized native, then fail exact bytes/refusal output; build failures do not kill mutants.
Tests compare live Go, fresh pinned npm output, source Node, sanitized native, emitted JavaScript and successful-run LeakSanitizer. New top-level Go tests call t.Parallel. All externally gated inputs are configured, with no scout skips.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TS_PRETTIER=/tmp/formatter-prettier ADAMIC_TYPESCRIPT_SOURCE=/tmp/formatter-typescript go test -v -count=1 -run '^TestScoutSideEffectImports$' ./stage1/cohere/tsprinter
ADAMIC_TS_PRETTIER=/tmp/formatter-prettier ADAMIC_TYPESCRIPT_SOURCE=/tmp/formatter-typescript go test -v -count=1 -parallel=2 -timeout=0 ./stage1/cohere/tsprinter
go vet ./stage1/cohere/tsprinter
gofmt -l stage1/cohere/tsprinter
```

The earlier side-effect-only gate had 69 fixtures / 40 accepted and a red package. Current fixtures retain all 69 sources; 44 now print exact bytes and the remaining reasons are refreshed. The new 61-fixture import gate passes Go/npm/Node/sanitized native/backend and leaks; every fixture is accepted and matches both printers. Its seven mutants and the four retained import mutants finish normally and die on byte mismatches. The full current package passes with zero skips, including both original regression gates, all 12 scoped successful-run mutants and all existing mutants. Its live broad oracles cover 171,146 expression fragments and 58,398 statement/program fragments. Native release build, vet, gofmt and git diff --check are clean. The final gate uses `-parallel=5`; all externally gated inputs are configured. Logs preserve substantive output and omit indentation on otherwise blank lines; they are retained in `scout/evidence/followup-package.log`, with focused merge/context/body-mutant and native build/vet/gofmt logs alongside it.
No shared lint harness, parser, bridge, runtime, compiler or submodule source file is changed. The Go oracle drivers are overlays from this package.

## Import design and Go evidence

`cohere/internal/format/javascript/print_module.go:19` composes keyword, phase/kind, specifiers, source, attributes and semicolon. :166 separates default/namespace docs from grouped named docs, preserves source order, and allows a named group to break only with multiple named items or a standalone binding. :151 emits `from`; :338 preserves aliases, including explicit `x as x`. The adapter distinguishes clause `type` and per-specifier `type` at `cohere/internal/format/estree/convert.go:2045`; defer is an existing parser phase and follows the same composition. The port counterparts are `expressions.ts:594` (specifier), :608 (clause) and :666 (declaration), guarded by `syntax.ts:473`. The parser's existing ImportClause children and semantic phase are consumed directly (`stage1/typescript/parser/statements.ts:409`, :488); no type-expression printer is required for a type-only import's identifier names.

Attributes follow Go `print_module.go:317`, `print_object.go:36` and `print_assignment.go:139`: ordinary object spacing, source-driven breaks, trailing commas only when broken, and values staying with keys. A sole string-valued `type` attribute has its lines removed even if the source brace is multiline or its value exceeds width (:289). `expressions.ts:639` implements this with the existing document arena and property-key/string printers. `syntax.ts:494` refuses other attribute shapes and legacy assertions. A single long named specifier stays flat; multiple named specifiers or a default-plus-named group may wrap. Quoted imported names remain quoted. Attribute string keys reuse the established safe unquoting rule. Nothing sorts imports or passes raw source through.

All new state is local arrays of numeric document links. The document arena owns docs and the parser owns indexed nodes; no ownership cycle, callback/interface dispatch, new syntax or collector is introduced. Native lowering, sanitizer and leak checks validate this composition. The supplied #bw3xg7c method-value ruling resolves the prior method-origin interface question; this implementation does not use that representation.

## Four new malformed-import acceptance conflicts

These are complete pinned Babel negative tests, not accepted parity fixtures. Go emits `Identifier expected.`; Node currently prints them. They stay in `import-stages.json` and final `summary.json`, with no exception or suppressing selector. The parser lane must own the shared repair; this scout stops before changing it.

| Shortest retained input | Shared port site | Go parser distinction |
| --- | --- | --- |
| `import { type as if } from "mod";` | statements.ts:332 | local reserved word, parser.go:2547 |
| `import { type if } from "./mod.js";` | statements.ts:342 | local reserved word, parser.go:2547 |
| `import { type as as if } from "mod";` | statements.ts:322 | local reserved word, parser.go:2547 |
| `import { type foo as "bar" } from "mod";` | statements.ts:348 | string local binding, parser.go:2474 |

## New fixture list

Every source, exact Go/npm byte pair and provenance is retained in `scout/import-fixtures.json`. Public/TypeScript entries are exact lines at their pins; the Go entry is its exact two-import prefix. Layout boundaries are reductions, labelled separately.

| Fixture | Source provenance |
| --- | --- |
| default-0 | shortest reduction / Go print_module.go:19,166,317 |
| default-1 | shortest reduction / Go print_module.go:19,166,317 |
| default-2 | shortest reduction / Go print_module.go:19,166,317 |
| default-3 | shortest reduction / Go print_module.go:19,166,317 |
| namespace-0 | shortest reduction / Go print_module.go:19,166,317 |
| namespace-1 | shortest reduction / Go print_module.go:19,166,317 |
| namespace-2 | shortest reduction / Go print_module.go:19,166,317 |
| named-0 | shortest reduction / Go print_module.go:19,166,317 |
| named-1 | shortest reduction / Go print_module.go:19,166,317 |
| named-2 | shortest reduction / Go print_module.go:19,166,317 |
| named-3 | shortest reduction / Go print_module.go:19,166,317 |
| named-4 | shortest reduction / Go print_module.go:19,166,317 |
| named-5 | shortest reduction / Go print_module.go:19,166,317 |
| named-6 | shortest reduction / Go print_module.go:19,166,317 |
| named-7 | shortest reduction / Go print_module.go:19,166,317 |
| type-0 | shortest reduction / Go print_module.go:19,166,317 |
| type-1 | shortest reduction / Go print_module.go:19,166,317 |
| type-2 | shortest reduction / Go print_module.go:19,166,317 |
| type-3 | shortest reduction / Go print_module.go:19,166,317 |
| type-4 | shortest reduction / Go print_module.go:19,166,317 |
| type-5 | shortest reduction / Go print_module.go:19,166,317 |
| attributes-0 | shortest reduction / Go print_module.go:19,166,317 |
| attributes-1 | shortest reduction / Go print_module.go:19,166,317 |
| go-cohere-own-module-import-prefix | cohere@7945d102a6c18dd36adf9114a758ce646e8b2359:internal/format/javascript/format_test.go:71 exact first two import lines |
| attributes-3 | shortest reduction / Go print_module.go:19,166,317 |
| attributes-4 | shortest reduction / Go print_module.go:19,166,317 |
| attributes-5 | shortest reduction / Go print_module.go:19,166,317 |
| attributes-6 | shortest reduction / Go print_module.go:19,166,317 |
| attributes-7 | shortest reduction / Go print_module.go:19,166,317 |
| defer-0 | shortest reduction / Go print_module.go:19,166,317 |
| default-width-30 | shortest reduction / Go print_module.go:19,166,317 |
| namespace-width-31 | shortest reduction / Go print_module.go:19,166,317 |
| named-width-32 | shortest reduction / Go print_module.go:19,166,317 |
| named-width-33 | shortest reduction / Go print_module.go:19,166,317 |
| named-width-34 | shortest reduction / Go print_module.go:19,166,317 |
| type-width-35 | shortest reduction / Go print_module.go:19,166,317 |
| attributes-width-36 | shortest reduction / Go print_module.go:19,166,317 |
| attributes-width-37 | shortest reduction / Go print_module.go:19,166,317 |
| public-default-0 | actualbudget/actual@9732a4463aac2909ac2aad1627085e0eb7a207b5:bin/__tests__/validate-publish-imports.test.ts:1 |
| public-default-1 | actualbudget/actual@9732a4463aac2909ac2aad1627085e0eb7a207b5:packages/api/index.browser.ts:5 |
| public-default-2 | actualbudget/actual@9732a4463aac2909ac2aad1627085e0eb7a207b5:packages/ci-actions/bin/check-migrations.ts:19 |
| public-namespace-0 | actualbudget/actual@9732a4463aac2909ac2aad1627085e0eb7a207b5:packages/api/browser-worker.ts:8 |
| public-namespace-1 | actualbudget/actual@9732a4463aac2909ac2aad1627085e0eb7a207b5:packages/api/e2e/consumer/main.ts:5 |
| public-namespace-2 | actualbudget/actual@9732a4463aac2909ac2aad1627085e0eb7a207b5:packages/api/methods.test.ts:1 |
| public-named-0 | actualbudget/actual@9732a4463aac2909ac2aad1627085e0eb7a207b5:bin/__tests__/validate-publish-imports.test.ts:5 |
| public-named-1 | actualbudget/actual@9732a4463aac2909ac2aad1627085e0eb7a207b5:packages/api/browser-worker.ts:9 |
| public-named-2 | actualbudget/actual@9732a4463aac2909ac2aad1627085e0eb7a207b5:packages/api/e2e/browser.test.ts:1 |
| public-type-0 | actualbudget/actual@9732a4463aac2909ac2aad1627085e0eb7a207b5:packages/api/app/query.ts:5 |
| public-type-1 | actualbudget/actual@9732a4463aac2909ac2aad1627085e0eb7a207b5:packages/api/browser-worker.ts:11 |
| public-type-2 | actualbudget/actual@9732a4463aac2909ac2aad1627085e0eb7a207b5:packages/api/index.browser.ts:3 |
| public-attributes-0 | angular/angular@c0dc8c4bbeea70879aef54e9fcc7888359dfd1a5:adev/src/app/routing/navigation-entries/index.ts:13 |
| public-attributes-1 | babel/babel@f67453d563918a1ea4a1dcb1d7af01d4a528d99b:babel.config.ts:14 |
| public-attributes-2 | babel/babel@f67453d563918a1ea4a1dcb1d7af01d4a528d99b:packages/babel-helper-compilation-targets/src/filter-items.ts:3 |
| typescript-importAttributes-0 | TypeScript@050880ce59e30b356b686bd3144efe24f875ebc8:tests/cases/conformance/importAttributes/importAttributes1.ts:10 |
| typescript-importAttributes-1 | TypeScript@050880ce59e30b356b686bd3144efe24f875ebc8:tests/cases/conformance/importAttributes/importAttributes3.ts:13 |
| typescript-importAssertion-0 | TypeScript@050880ce59e30b356b686bd3144efe24f875ebc8:tests/cases/conformance/importAssertion/importAssertion1.ts:10 |
| typescript-importAssertion-1 | TypeScript@050880ce59e30b356b686bd3144efe24f875ebc8:tests/cases/conformance/importAssertion/importAssertion3.ts:13 |
| typescript-externalModules-0 | TypeScript@050880ce59e30b356b686bd3144efe24f875ebc8:tests/cases/conformance/externalModules/typeOnly/ambient.ts:7 |
| typescript-externalModules-1 | TypeScript@050880ce59e30b356b686bd3144efe24f875ebc8:tests/cases/conformance/externalModules/typeOnly/chained.ts:13 |
| typescript-importDefer-0 | TypeScript@050880ce59e30b356b686bd3144efe24f875ebc8:tests/cases/conformance/importDefer/importBindingDefer.ts:8 |
| typescript-importDefer-1 | TypeScript@050880ce59e30b356b686bd3144efe24f875ebc8:tests/cases/conformance/importDefer/importDefaultBindingDefer.ts:9 |

The existing body-brace mutant selector was narrowed to the original block-body line because a new import group shared its shorter search string; its semantic mutation and corpus remain the same. All existing mutation targets are unique. New mutants cover default binding, namespace alias, named alias, clause type, specifier type, attribute value and single-type flattening. Each must compile and finish normally on Node and native before a mismatch kills it. Retained side-effect mutants separately cover keyword, separator, semicolon and dropping attributes. The source-context mutant separately proves both shortest regression checks can fail. New top-level Go tests call t.Parallel.

## Next scout

Take the ranked type-printing family next: TypeReference affects 95,267 files / 1,331,478 sites. Preserve import comment/trivia and export boundaries as explicit follow-up work. First hand the four invalid import-local cases and `export default abstract;` loss to the parser lane; these are implementation handoffs, not language rulings. Carry all nine output differences, 21 acceptance conflicts, two Node panics and 15 encoding exclusions forward without exceptions. Carry the supplied #bw3xg7c method-value ruling above into future callback composition. No open language questions remain for this piece.
