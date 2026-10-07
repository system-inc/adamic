# Ninth batch

ImportedNameOf reads a named-import specifier's source name. Numeric AST edges preserve node identity without ownership cycles. Negative indexes represent nil; decoded text and exact Name/PropertyName projections come from the common adapter. A present empty PropertyName wins over a nonempty local alias.

Actual pinned Go cohere decides helper results on all four consumers' test-file strings plus alias, string-name and nil-name controls. Source Node, emitted JavaScript and ASan/UBSan native must match. Two compiling native mutants skip aliases and replace a present empty alias with the local name; both are caught. See CONSUMERS.md and readiness.json for dependency removals.

CallExpressionSource and HasAttributeNamed were withdrawn to slot 02's two-second-earlier claim. No duplicate port is delivered. The discarded local ports are not exported or delivered.

## Regexp replacements

isDecimalDigit(source, byteOffset) accepts only ASCII bytes 0 through 9. Offsets at or beyond UTF-8 length answer false; negative offsets panic, as Go indexing does. Offsets supplied by the adapter must be integers. UTF-16 string positions are not byte offsets.

countGroups(source) returns {count, named}. Backslashes skip exactly the next UTF-8 rune, including supplementary characters; a terminal backslash consumes only itself. Brackets toggle class state exactly like Go, without nesting or full syntax validation. Parentheses inside a class do not count. Outside a class, ordinary groups count, and (?< groups count and set named except (?<= and (?<! lookbehinds. Unfinished syntax preserves Go's scan result.

Each regexp helper takes every test-file string from all four consumers plus boundary, malformed, escape, Unicode-digit and group-shape controls. Every Unicode scalar is tested: digit classification at byte zero and EOF, and capture scanning with that scalar escaped before named, lookbehind and noncapturing groups. Go private helpers are called through an additive test-only overlay; no cohere source changes. UTF-8 is validated before serialization, so invalid byte strings cannot quietly change when JSON replaces them.

The owned test package compares Go with source Node, emitted JavaScript and sanitized native. Eight compiling native mutants must produce a semantic mismatch with Go, exit zero and emit no stderr. Corpus hashes and per-consumer literal counts are in evidence. Run with the setup environment sourced:

```
go test ./stage1/cohere/lint/helpers/slot05/batch9 -count=1 -v -timeout=20m > /tmp/lint05-batch9.log 2>&1
```

These helper contracts depend on the common AST adapter; no rule registration, diagnostic spans, fixes or suggestions are implemented. Invalid UTF-8 Go strings, noninteger offsets, corrupted AST arenas and exact Go panic wording are outside this gate. The regexp engine, rewriting and syntax validation remain separate prerequisites.
