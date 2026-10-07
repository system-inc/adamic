# Generic families, math detection and utility identity

Each helper has its own `.a` file:

- `isGenericName(value)` matches the thirteen exact Go generic-family keywords: serif, sans-serif, monospace, cursive, fantasy, system-ui, ui-serif, ui-sans-serif, ui-monospace, ui-rounded, math, emoji and fangsong.
- `hasMathFunction(value)` finds any of the nineteen literal substrings calc(, min(, max(, clamp(, mod(, rem(, sin(, cos(, tan(, asin(, acos(, atan(, atan2(, pow(, sqrt(, hypot(, log(, exp( and round(. It scans anywhere, including quoted or malformed values. Case matters and whitespace before the opening parenthesis is rejected. No parsing or balancing is added.
- `loadedUtilities(system)` returns `system.utility` unchanged. An initialized non-null system is required; -1 denotes Go's nil evaluator, and nonnegative handles identify existing evaluator records in the caller's arena. The helper neither constructs nor compiles utilities. Callers preserve stable handle identity and must check nil before dereferencing.

Text inputs are valid UTF-8-derived Unicode strings. Invalid UTF-8 and unpaired UTF-16 are outside the adapter contract. No trimming, case folding or Unicode normalization.

The owned Go overlay exports the actual private predicates and calls the actual Utilities getter. All four consumers supply 118 captured runtime sources. Full sources, derived tokens, every accepted keyword/function and uppercase, whitespace, NUL, Unicode, prefix and quotation controls produce 556 strings. Nil and two distinct evaluator pointers are compared for returned identity, mutation aliasing, unchanged storage and repeated reads. This produces 1,118 output lines compared against actual Go on source Node, emitted JavaScript and ASan/UBSan native.

Seven compiling semantic mutants must run with exit 0 and no stderr before the comparator catches wrong output. The complete owned helper gate also reruns every previous mutant. Regenerate with `python3 stage1/cohere/lint/helpers/slot03/batch11/testdata/regenerate.py`; test with `go test ./stage1/cohere/lint/helpers/slot03 -count=1 -v -timeout=20m`, writing stdout/stderr directly to a log. Source `/workspace/adamic-tools/env.sh` first.

Capture's upstream whole-rule gate still fails on unavailable external Tailwind installations/corpora. These are helper parity checks, not complete findings/fixes/suggestions integration. No rule or shared harness changes.
