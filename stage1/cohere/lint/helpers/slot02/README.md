# Slot 02 helpers

Three public helpers, each in its own `.a` file. These remove dependencies in the frozen readiness ledger; they do not port rules or replace the existing linter.

| File | API | Go contract |
|---|---|---|
| `../jsx_attribute_name.a` | `attributeName(nodes, property)` | `jsx.AttributeName`: only a JsxAttribute with an Identifier name answers. The result keeps `name` separate from `named`, including present-empty. Nil, spread, namespaced and missing names decline. |
| `../property_name.a` | `propertyName(nodes, key, accept)` | `property.Name`: flags Named=1, Quoted=2, Templated=4, Numeric=8, Private=16, Computed=32; Static=63 and Textual=3. Computed identifiers and private identifiers decline even when their flags are enabled. Parentheses are skipped only inside computed keys. |
| `../tailwind_reader_for.a` | `classLiteralReaderFor(settings, settingsKey, compiledReader, cached)` | `tailwind.ClassLiteralReaderFor`: namespace the settings key, reuse the per-file bound reader, share compiled fields and allocate a fresh values map on each fill. |

The AST APIs accept the read-only projection in `../slot02_ast.a`. Edges are numeric arena indices, with -1 for nil; this keeps the arena acyclic as a reference-counted value. `kind` is the Go parser kind name. Relevant kinds are Identifier, PrivateIdentifier, StringLiteral, NoSubstitutionTemplateLiteral, NumericLiteral, ComputedPropertyName, ParenthesizedExpression and JsxAttribute; other kinds decline. `text` must be Go AST Text(), including decoded identifiers, the leading hash of a private identifier and the parser's normalized numeric text. `expression` and `name` preserve the corresponding Go fields and node identity. The production AST adapter remains rule-worker territory. Inputs must be valid finite projections; missing indices and cycles panic explicitly. Arbitrary malformed graphs and exact Go internal panic prose are not part of the parity claim.

The helpers never unwrap an assertion, a JSX expression or a member access in place of parentheses. `propertyName` reads a key node, not a whole access expression: `AccessedName` and `NameTagged` are other helpers and remain unported here. The accept mask is the caller's Go Kinds value; choosing it is a rule decision. No default widens it.

`testdata/slot02_inputs.json.gz` holds actual asserted consumer inputs, including dynamically assembled fixtures. The capture workflow runs the consumer tests with Go cohere's docs capture. An overlay records the localization rule's custom multi-file harness, without editing cohere. File names are canonicalized to preserve the source extension, which selects the real TS/TSX/JS/JSX parser mode; these helpers do not read file paths. The compressed corpus is deterministic and `slot02_coverage.json` maps every AST-helper consumer to captured evidence. Tests refuse a consumer with no actual captured input. The oracle calls the real Go helpers on every parsed node, then the same projected inputs run on Node source and ASan/UBSan native. Property names are checked with all 64 combinations of the six flags. Nil and factory-built missing/empty JSX names supplement the corpus.

The reader helper takes prerequisite functions rather than reimplementing other workers' settings keys, compiled regexp readers or FileCache. Its cache adapter receives the cache key, compiled key, settings, compile function and bind function; it must invoke compile then bind only on a miss and implement Go `rule.Cached` behavior. Nil caches, zero caches and heterogeneous type collisions compute fresh readers. This explicit factory protocol avoids a cycle-capable escaping closure in Adamic. Values use arena node indices; the caller owns the Go classValues projection. The oracle exercises the real private Go prerequisites and reader for eight settings, including defaults, empty settings, duplicate names, invalid/duplicate patterns, embedded NUL and key collisions. It compares cache identity, value isolation, shared fields and fill keys on Node and sanitized native. These are helper-level observations covering the shared contract of all twelve consumer rules; live Tailwind rule findings and external design-system loading are not covered.

From the repository root, after sourcing the setup environment:

```
python3 stage1/cohere/lint/helpers/testdata/slot02_capture.py > /tmp/lint-helpers-02-capture.log 2>&1
go test ./stage1/cohere/lint/helpers -run '^TestSlot02' -count=1 -v -timeout=20m > /tmp/lint-helpers-02-tests.log 2>&1
```

Mutants use temporary `.a` copies. Every credited mutant compiles and executes successfully, then differs from Go semantically; a compilation or sanitizer failure is not credited. Full rule finding/fix parity, linter integration, all possible AST programs and the full repository test gate are outside this slice. Exact results and consumer lists are in REPORT.md.
