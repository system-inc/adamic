# Escape dispatch, class atoms and regexp test wrapper

Three separate .a helpers preserve the pinned Go contracts with explicit external dependencies.

- decodeEscape takes UTF-8 bytes, the byte position after a backslash, an escape context and EscapeOps. It preserves EOF rejection, decoded-rune dispatch, fixed control escapes, set classifications, inside/outside-class boundary behavior, named-reference rejection and specialized decimal/hex/Unicode/control/property/identity delegation. Widths remain Go byte widths.
- classAtoms takes body bytes, rewrite options, context and ClassOps. It copies context with inClass true, emits ordinary/dash atoms, decodes escapes, expands word/nonword sets only under both Unicode and ignoreCase, marks case-insensitive property sets inexact, preserves raw set bytes through its slice callback and delegates range joining. Decode or joining errors return undefined atoms and exact false.
- regExpTest takes an optional RegExpValue with an optional opaque engine, the subject and a MatchString callback. Missing wrapper or engine returns false without calling the engine. Otherwise it returns true only for a successful match without error. This helper does not compile patterns or implement matching, flags, cancellation or timeout machinery; its supplied engine owns those operations.

Private escape/set/atom kind numbers are upstream enums, not AST kinds. UTF-8 decoding, specialized escapes, word/nonword expansion, range joining and matching remain explicit dependencies. No regex matcher or rule-local Go regex is copied. The native driver uses numeric opaque engine handles and reversible hex projections for subject and atom text bytes; helpers pass those projections unchanged. The callback that matches still executes the actual Go engine in the oracle. Generic object-engine instantiations are outside this driver.

The private Go overlay changes dependency call names only inside the original helper bodies. Real dependencies execute, returning exact data and recording argument/order traces. Nested escape-helper calls are muted when testing classAtoms because decodeEscape is its external dependency. Engine compilation is muted when testing RegExp.Test, so only its MatchString call enters that helper's trace. Rejected patterns exercise the nil-wrapper path. A real nanosecond-budget Go engine timeout and a separately controlled matched-plus-error dependency result exercise error handling.

Operation results are pooled by exact serialized content without removing any input row. The driver replays actual returned data and records the arguments supplied by the port. Unknown dependency keys in a mutant receive a neutral progress-preserving result and still enter the trace, allowing a successful semantic disagreement rather than earning credit for a panic. Production helpers contain no such fallback.

Run with the setup environment sourced:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch23 -count=1 -v -timeout=20m > /tmp/lint05-batch23-complete.log 2>&1
```

Actual Go bytes must match source Node, emitted JavaScript and sanitized native. Baselines and mutants must exit zero with empty stderr; compilation refusals, panics and sanitizer failures are not semantic-mutant credit. Raw compressed corpora, exact expected output and every witness are in evidence. REPORT.md records commands, measured results and limits.
