# Escape widths, class ends and React base names

skipPatternEscape(bytes, index, flags, decodeWidth) returns {value, ok}, where value is the consumed byte count. classEnd(bytes, start, flags, decodeWidth) returns the byte offset just after the closing bracket. A failed class scan returns the original start. isReactComponentBaseName(bytes) accepts exactly the original UTF-8 bytes of Component or PureComponent.

Bytes must be original Go string bytes, each integer 0..255. Index/start must be nonnegative integers. Negative offsets are outside the upstream caller contract and are not given a new refusal policy here. flags is RegexFlags from batch25/flags.a. decodeWidth(index) must supply Go utf8.DecodeRuneInString's width for the original suffix: valid rune width 1..4, invalid UTF-8 width 1, empty suffix width 0. The scanner does not implement UTF-8 decoding or a regex matching engine. Supply the existing decoding adapter in integration; the oracle supplies actual Go widths at every byte.

ClassEnd reuses this batch's SkipPatternEscape. Escape scanning reuses the completed batch24 hex helpers and batch25 UV helper. It preserves upstream best-effort recovery for malformed brace and hex escapes. Only v enables nested brackets; u and v both enable brace escapes. No Go regex is introduced by these helpers, and no finding span is constructed.

See REPORT.md, readiness.json and evidence/mutants.json for the Go/source-Node/emitted-JavaScript/native comparisons, all thirteen compiling mutants and coverage limits. Helpers are separate .a files; test drivers and oracle-only Go files remain owned by this directory. No shared registration or harness edits.
