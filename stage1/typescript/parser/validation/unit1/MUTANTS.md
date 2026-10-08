# Mutants run for this landing

Each parser mutant is built and run on Node and sanitized native. The external typescript-go parser supplies the independent expected answer. Byte mutants must finish normally; a compiler error, sanitizer crash or refusal does not count as a byte-comparison kill. Counts and termination have their own independent checks.

| Mutant | Check that caught it | Log |
| --- | --- | --- |
| Binary precedence changed | Expression-tree bytes | final-parser.log |
| Optional property access flag removed | Expression-tree bytes | final-parser.log |
| Parenthesized expression made an arrow | Expression-tree bytes | final-parser.log |
| JSX attribute list count | Whole-tree bytes | final-parser.log |
| JSX text payload | Whole-tree bytes | final-parser.log |
| JSX child order | Whole-tree bytes | final-parser.log |
| JSX type argument comma | Whole-tree bytes | final-parser.log |
| JSX text start | Whole-tree bytes | final-parser.log |
| JSX self-closing kind | Whole-tree bytes | final-parser.log |
| JSX namespace kind | Whole-tree bytes | final-parser.log |
| JSX whitespace flag | Whole-tree bytes | final-parser.log |
| JSX expression kind | Whole-tree bytes | final-parser.log |
| JSX raw attribute scanner value | Whole-tree bytes | final-parser.log |
| JSX name scanner payload | Whole-tree bytes | final-parser.log |
| Dashed JSX member accepted | Go-backed refusal check, mutant exits 0 | final-parser.log |
| Private JSX name accepted | Go-backed refusal check, mutant exits 0 | final-parser.log |
| Escaped JSX name accepted | Go-backed refusal check, mutant exits 0 | final-parser.log |
| Expression count forced to zero | Count only; AST remains identical | final-parser.log |
| Whole-tree count excludes its nodes | Count only; AST remains identical | final-parser.log |
| Octal suggestion changed | Diagnostic bytes | final-parser.log |
| Expected-token diagnostic code changed | Diagnostic bytes | final-parser.log |
| Stranded-export diagnostic changed | Diagnostic bytes | final-parser.log |
| Generic parameter child removed | Recovered tree bytes | final-parser.log |
| EOF made an identifier forever | Two-second subprocess bound | final-parser.log |
| Speculative roots preserved | Recovered expression-tree bytes | final-parser.log |
| For-of made for-in | Whole-tree bytes | final-parser.log |
| Type-only import phase removed | Whole-tree bytes | final-parser.log |
| Keyof made readonly | Whole-tree bytes | final-parser.log |
| Scanner diagnostic hook inverted | Diagnostic bytes on testdata/recovery/unterminated-string.ts.txt | final-fastpath.log |
| Reserved return token made an identifier | Diagnostic/tree bytes on testdata/recovery/reserved-binding.ts.txt | final-fastpath.log |
| Lint overlap winner misreported | Findings and repair bytes on Node, emitted JS, sanitized native | final-lint.log |
| Lint count increased by one | Count only on Node/native; ordinary findings identical | final-lint.log |
| Actual extra parse for every compiler file | Original landing instruction budget; AST/count remain identical | final-measure.log |
| Profile summary increased by one | Sum of self costs must equal summary | final-measure.log |

This is 30 parser mutants, two lint mutants and two instruction checks. The four recovery fixtures are also checked positively through the unmodified TestRulesAgree on all four runtimes.
