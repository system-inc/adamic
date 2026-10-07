# Slot 03 third-batch contracts

One helper per `.a` file. Consumer counts come from the frozen parent readiness ledger. Dependency removal is not a completed rule implementation.

`hexValue(character)` receives an integer Go rune code point and returns 0..15 for exactly 0..9, a..f and A..F; everything else returns -1. It matches every point 0..0x10FFFF, signed-int32 extremes and actual fixture runes. The contract is a decoded rune, despite the first claim's byte wording.

`unescapeStringLiteralText(text, decodeEntity)` scans the first semicolon after each ampersand. A candidate shorter than two characters is preserved one ampersand at a time. A successful decoder result replaces the entire reference; a failed result preserves the ampersand and continues scanning. Replacement text is not rescanned. `EntityResult` contains `text` and `ok`. The separately owned decoder is an explicit callback, not another implementation here. Valid UTF-8 text corresponds to JavaScript UTF-16; scanning only ASCII delimiter boundaries preserves the same output. Arbitrary invalid UTF-8 byte strings are outside this adapter contract.

`parameterNodes(parametersPresent, nodes)` returns empty for a nil parameter-list reference and returns the same node-index array for a present list. Node order and identity, including repeated or nil nodes, are preserved. Writes to existing returned entries share the original storage, matching Go slices. This fixed-length AST adapter represents nil slices as empty lists; structural list growth is outside its contract. All consuming rules read or range over the list.

The oracle uses real Go cohere implementations, actual rule runtime source inputs and typescript-go decoded string fields. All XHTML names come from Go's pinned generated table through the oracle, not a browser guess. Decoder observations are supplied to the callback so this unit independently tests the public scanner, its decoder-call order and results. UTF-16 unit output exposes control characters, supplementary characters, replacement runes and Unicode normalization differences.

Go, Node source, emitted JavaScript and ASan/UBSan native must match byte for byte. Every native mutant must compile, exit 0 without stderr, and disagree semantically with Go; compile errors and sanitizer failures do not count. Coverage checks require every consumer, matching source counts and the exact cohere pin. The complete shared rule harness and registration generator are untouched. No complete rule finding/fix/suggestion parity is claimed.
