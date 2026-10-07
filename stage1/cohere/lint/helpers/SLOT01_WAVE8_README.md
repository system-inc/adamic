# Slot 01 regexp leaf helpers

Each production file owns one Go helper. These are helper APIs, not rule listeners or rule registrations.

- `regexpDecimalEscape(bytes, offset)` returns value and width. It consumes ASCII digits, updates only while the preceding value is below 1 << 20, and continues counting the full digit run after saturation. The retained value may exceed the threshold after its final multiplication.
- `regexpLegacyOctal(bytes, offset)` returns value and width. Leading 0 through 3 permits three digits; leading 4 through 7 permits two. Any other byte returns zero value and zero width.
- `regexpClassAtomCovers(kind, low, high, rune)` uses numeric Go enum values: rune 0, range 1, set 2, dash 3. Rune/dash compare equality, ranges include both bounds, other kinds return false.

The arena contains validated bytes 0 through 255. Offsets are nonnegative integers; octal requires a present byte, decimal permits offsets at or past the end. Class kinds are uint8 and rune values are signed int32. These are adapter preconditions, not arbitrary JavaScript numeric-input validation. Byte offsets preserve Go UTF-8 indexing, including continuation bytes; no conversion to UTF-16 positions is performed.

The slot-owned Go overlay calls the actual private Go helpers, without changing cohere's worktree. Every consuming fixture file contributes complete Go string expressions, including concatenations and local constant references. Decimal and octal examine each byte offset; coverage examines each fixture rune, each supported/unknown kind and range boundaries, deduplicating identical query tuples. Controls cover all 256 leading bytes, digit boundaries, saturation, Unicode and int32 extremes. Comparison runs Go, source Node, sanitized native and emitted JavaScript. Mutants compile and finish with empty stderr before a differing stdout is credited.

No shared rule harness, registration generator, compiler or rule directories are changed. Remaining regex parser, case tables and adapters belong to their existing owners. See SLOT01_WAVE8_REPORT.md and slot01_wave8_readiness.json for measured results and every residual dependency.
