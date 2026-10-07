# Tenth batch

decodeControlEscape preserves Go byte offsets and all decoded-escape fields. It decodes ASCII letters modulo 32; digits and underscore work only in non-Unicode character classes. Unicode invalid escapes return the exact Go error and zero fields. Annex B invalid escapes return backslash with width zero, leaving c unconsumed.

The oracle calls real private Go for every fixture string from all four consumers, every byte position through EOF plus one, four Unicode/class contexts and multiple sizes on short inputs. Source Node, emitted JavaScript and sanitized native must match. Four compiling semantic mutants prove modulo decoding, class-only widening, Unicode refusal and zero-width fallback.

No rule listener or dispatcher is added. Common valid-string adapters supply integer byte offsets. Invalid UTF-8 Go strings and exact panic prose are not covered. The earlier bounded-width and group-kind claims are withdrawn before any duplicate code.

## Named-prefix readers

namedBackreference and namedGroupOpener return {text, width, found}. Width is a UTF-8 byte count. The first greater-than delimiter closes the returned prefix, including when the name is empty or contains an escape, NUL, newline or supplementary character. No name validation is added. Group openers decline both lookbehind forms before searching for a delimiter. Anchors are exact and case-sensitive.

The named readers are tested on every consumer's fixture string and every suffix at a Go rune boundary, plus explicit malformed, empty-name, delimiter, escape and Unicode controls. Each also tests every Unicode scalar as a name followed by a delimiter and trailing text. The private Go functions provide all expected fields and returned bytes; source Node, emitted JavaScript and sanitized native must match. Four mutants per reader test empty-name acceptance, UTF-8 width, first-delimiter behavior, and either the backreference anchor or negative-lookbehind exclusion.

A nonempty control-escape error represents Go's wrapped ErrUnsupportedSyntax and must propagate as an invalid pattern, not a clean lint result. Prefix found=false is a normal decline. Invalid UTF-8 Go strings are explicitly rejected by corpus generation before JSON serialization changes them. Exact Go panic wording and arbitrary noninteger numeric arguments are not compared.

Run with the setup environment sourced:

```
go test ./stage1/cohere/lint/helpers/slot05/batch10 -count=1 -v -timeout=20m > /tmp/lint05-batch10.log 2>&1
```

The Go overlay adds test-only exports without modifying cohere. Every helper has its own .a file. Shared registration, rule dispatch, diagnostics and compiler files are untouched. See CONSUMERS.md, readiness.json and evidence for the frozen cohort's residual dependencies, per-consumer literal counts and corpus hashes.
