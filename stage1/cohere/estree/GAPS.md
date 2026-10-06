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
No runtime helper was changed. `gaps/nbsp.ts` and `gaps/lineSeparator.ts` are minimal source inputs.
`TestOriginalLibraries` proves the Go/Prettier ContentEnd difference,
allowing exactly the corresponding field delta and no other change.

## Parser dependency boundaries

The parser's `GAPS.md` and `WHOLE_REPORT.md` explicitly exclude diagnostic
parity, JSX and AST list ranges beyond the canonical fields they compare.
The converter's source driver currently refuses unrepresented conversion
kinds rather than outputting a partial tree. Generic argument wrappers now reconstruct positions from child spans and tokens;
ordinary, nested and trailing-comma cases are held to Go and the original library.
These are port coverage limits, not compiler or runtime failures.

## Observed wrapper behavior: CRLF normalization

`gaps/crlf.ts` contains `x`, CRLF, and a semicolon. The raw original converter
agrees with Go. Prettier's full parse wrapper normalizes CRLF before parsing: its
Program and ExpressionStatement ranges end at UTF-16 position 3, whereas this
Go package's direct ParseTypeScript API preserves the original four-byte source.
The port preserves the Go API boundary. `TestOriginalLibraries` requires exactly
this two-range delta. This does not claim the complete format pipeline disagrees.

## Stage 0 NotYet: postfix increment as a value

`gaps/postfixValue.ts` indexes an array with `index++`, then prints the index.
Node prints `7` and `1`. Loading passes; lowering refuses a PostfixUnaryExpression
at the value use. `TestPostfixValueGap` requires both observations. The converter
splits each index read and increment into statements. No compiler fix was made.
