# Color and length helpers

One `.a` module per claimed Go helper:

- `named_color.a`: `isNamedColor(value, lower)` looks up the exact Go-lowercased value in the pinned 169-name table. The supplied callback must implement Go strings.ToLower. Whitespace is preserved. Go's `İndigo` becomes `indigo`; JavaScript's Unicode lowercase operation gives a combining-dot spelling instead.
- `color.a`: `isColor(value, lower)` accepts a hash prefix, a case-insensitive color-function prefix, or a named color. It does not validate CSS. Its patterns are static JS RegExp literals, with ordinary `i` rather than Unicode `iu` for the function test: Go's fixed-byte prefix comparison cannot accept the non-ASCII lookalikes that Unicode regex folding would introduce.
- `length.a`: `isLength(value, numberWithSuffix, hasMathFunction)` delegates numeric-unit and math checks, and recognizes the spacing-function prefix with a static JS RegExp literal. Unit names are exact and case-sensitive, including uppercase Q.

Go lowercasing, numberWithSuffix and hasMathFunction are separately owned runtime dependencies, supplied as callbacks. The test adapter obtains lowercase and numeric-prefix results from actual Go; it checks the supplied suffix list against the real numeric tail. This batch adds no numeric parser, Unicode casing table or handwritten function matcher.

With the setup environment sourced, run from the repository root:

```
python3 stage1/cohere/lint/helpers/slot04_wave17/testdata/generate.py > /tmp/slot04-wave17-generate.log 2>&1
python3 stage1/cohere/lint/helpers/slot04_wave17/testdata/capture.py > /tmp/slot04-wave17-capture.log 2>&1
go test -count=1 -v ./stage1/cohere/lint/helpers/slot04_wave17 > /tmp/slot04-wave17-tests.log 2>&1
```

The Go oracle and live-call capture use temporary overlays. No shared harness or pinned cohere source is modified. The oracle replays every actual Go named-color entry at runtime, so changes to that table cannot silently escape the fixed generated controls. Complete rule findings, fixes, suggestions and arbitrary callback implementations remain outside this helper comparison.
