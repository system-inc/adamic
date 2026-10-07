# Escape and class syntax helpers

Three separate .a files port the pinned Go syntax helpers, without implementing a matcher.

- decodePropertyEscape takes UTF-8 bytes, the p/P byte offset, its decoded size (one), escape context, identityEscape and relative closing-brace lookup callbacks. Annex B delegates identity; Unicode requires an opening brace and a closing brace. Returned escape kinds, set classification, byte width and errors preserve Go.
- decodeNumericEscape takes UTF-8 bytes, a valid digit offset and context, with decimal-digit, whole decimal-escape and legacy-octal callbacks. It preserves the standalone NUL fast path, outside-class backreferences bounded by group count, Unicode rejection, legacy octal and literal 8/9 fallback.
- readClass takes UTF-8 bytes, a rune-size decoder and an opaque body-slice callback. It starts after the opening bracket, recognizes an optional caret, skips escaped decoded runes, closes at an unescaped bracket and resets all outputs on failure. Byte offsets and widths remain byte offsets; the rule worker converts positions only when constructing findings.

Private escape kind/set numbers are Go enums, not AST kinds. Numeric/property offsets must be valid parser offsets, and marker bytes must be digits or p/P respectively. Callbacks are explicit external dependencies and must behave like the pinned Go implementations. No process-global matcher, option decoder or regex translation table is introduced.

The driver uses hexadecimal body bytes to compare arbitrary invalid UTF-8 without changing Go's bytes into replacement characters. Actual Go rune-size results are supplied per byte offset, including the zero-size EOF case. The private overlay changes only dependency call names inside actual upstream bodies; helpers still execute original Go logic and real dependencies. Traces hold dependency argument positions and call order alongside the returned values.

Run with the setup environment sourced:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch22 -count=1 -v -timeout=20m > /tmp/lint05-batch22-complete.log 2>&1
```

All output is compared against Go on source Node, emitted JavaScript and sanitized native. Native baselines and mutants must exit zero with empty stderr; compiler refusals, panics and sanitizer failures are not semantic-mutant credit. Compressed raw input/expected output and all witnesses are in evidence; REPORT.md gives measured results and limits.
