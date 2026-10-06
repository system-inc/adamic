# ESTree gaps

## Observed Go behavior: trailing multibyte whitespace

The pinned Go `lastWhitespaceSize` returns after its first one-byte attempt.
A UTF-8 continuation byte therefore ends trimming even when the character
is NBSP or U+2028. For the input `x` followed by NBSP and `;`, Go's
ExpressionStatement ContentEnd is byte 3. JavaScript `trimEnd` would give
byte 1. The port uses an explicit Go-compatible tail trim for ContentEnd.
The generated agreement corpus includes both NBSP and U+2028 before a
semicolon, so using JavaScript trimEnd fails byte comparison.

This is an observation of this cohere pin, not an inferred language rule.
No runtime helper was changed. A separate original-library comparison and
minimal proving programs will accompany the next checkpoint.

## Parser dependency boundaries

The parser's `GAPS.md` and `WHOLE_REPORT.md` explicitly exclude diagnostic
parity, JSX and AST list ranges beyond the canonical fields they compare.
The converter's source driver currently refuses unrepresented conversion
kinds rather than outputting a partial tree. Generic argument wrappers are
also refused pending a faithful reconstruction of their list positions.
These are port coverage limits, not compiler or runtime failures.
