# Slot 01 word-class and boundary helpers

Each production .a file owns one Go helper. These are standalone helpers, not rule listeners.

- regexpWordClassAtoms(options) returns fresh mutable records and a fresh list, in Go order: ASCII 0-9, A-Z, underscore, a-z. Only ignoreCase AND unicode appends U+017F LONG S and U+212A KELVIN SIGN. Multiline/dotAll have no effect.
- regexpNonWordClassAtoms(options) returns fresh records for the complement gaps between word atoms through U+10FFFF inclusive. A singleton gap uses the rune kind with high zero; wider gaps use ranges. It reuses the separately named word-atom helper.
- regexpWordBoundary(negated, options, wordCharacters) calls the independently owned wordCharacters dependency exactly once with all four original flags. It then writes the exact Go positive or negative boundary lookaround expression, reusing the returned word string unchanged.

RegexpRewriteOptions carries the four boolean fields. RegexpClassAtom preserves numeric Go kind (rune 0, range 1), low/high and text, including zero high for rune atoms and empty text. Outputs may be mutated like Go slices, but each later call is independent.

All sixteen combinations of the four boolean flags are checked against actual Go, and both negation choices for boundaries. All four consumer fixture files supply literal corpus provenance; these helpers have only flag/negation inputs, so corpus strings determine repetitions rather than additional helper arguments. Each atom-list query records its first result, mutates its first atom and verifies a new result. An observational Go overlay records actual boundary callback count and forwarded flags without changing returned behavior or cohere's worktree.

Source Node, sanitized native and emitted JavaScript match Go. Native mutants must compile, exit zero and emit no stderr before wrong stdout is credited. Tests prove flag conjunction, complement endpoint, polarity, fresh results, callback forwarding and callback count can fail.

The dependency callback must supply the wordCharacters contract; production does not guess Unicode class spelling. No shared rule harness, registration generator, rule directory or compiler changes. See SLOT01_WAVE10_REPORT.md and slot01_wave10_readiness.json for commands, every consumer and residual dependency.
