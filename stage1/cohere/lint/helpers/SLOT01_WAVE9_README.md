# Slot 01 regexp prefix scanners

One production .a file per Go helper:

- regexpBoundedQuantifierWidth(bytes, offset) starts at slice byte one, requires at least one ASCII lower digit, optionally consumes a comma and zero or more upper digits, then requires a closing brace. It deliberately does not check slice byte zero. It parses syntax, not bound magnitudes/order, and permits leading zeros and descending bounds.
- regexpQuantifierWidth(bytes, offset) recognizes one *, + or ?, or invokes the separately named bounded helper for an opening {. It consumes exactly one following lazy ? and returns a byte width, with zero for other prefixes or incomplete braces.
- regexpGroupKind(bytes, offset) returns Go groupPlain 0, groupLookahead 1 for (?= or (?!, groupLookbehind 2 for (?<= or (?<!, and plain for all other prefixes including incomplete and named groups.

The caller supplies a validated dense byte arena (0 through 255) and a nonnegative integer offset within or at its length, representing a Go source[offset:] slice. Offsets and widths count bytes, including UTF-8 continuation bytes. Arbitrary invalid number arrays/offsets are outside the adapter contract, not silently interpreted as full regexp compilation. Missing array entries in prefix tests compare unequal to ASCII punctuation, preserving short-slice decline.

The slot-owned oracle adds a virtual Go export file without changing cohere's worktree. Every consumer test contributes complete Go literal expressions, concatenations and resolved local constants. Every byte suffix, including the empty suffix, is compared against the actual private Go helper. Additional controls cover all 256 initial bytes, ASCII/non-ASCII digits, malformed/lazy braces, huge bounds and group prefixes. Comparison runs Go, source Node, sanitized native and emitted JavaScript; temporary native mutants compile and run with empty stderr before differing output is credited.

No shared rule registration, rule dispatcher, compiler or harness is edited. APIs are standalone helpers, not new rules. See SLOT01_WAVE9_REPORT.md and slot01_wave9_readiness.json for every consumer, residual prerequisite, command and witness.
