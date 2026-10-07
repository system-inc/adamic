Built parseCSSDeclaration in one .a helper file with complete declaration fields and Go Unicode trimming.
Claim c461d147 was pushed before source; the preceding CSS-string helper was tested and pushed at 5f016e4f.
validate.py passed 8,836 cases and 619,181 identical output bytes on Go, Node, emitted JavaScript and ASan/UBSan native; original consumer tests passed.
The important-search-before-colon mutant compiled, exited normally with empty stderr and was caught by output comparison on all three backends.
Not covered: whole CSS parser or linter integration, invalid byte views/equal-or-negative-invalid colon indexes, arbitrary unbounded configurations and the full repository gate.

The consumers are better-tailwindcss/enforce-canonical-classes,
better-tailwindcss/enforce-consistent-class-order,
better-tailwindcss/enforce-consistent-variant-order,
better-tailwindcss/enforce-shorthand-classes,
better-tailwindcss/no-conflicting-classes and better-tailwindcss/no-unknown-classes.
This removes one helper prerequisite each, six total, and zero final blockers.
Combined with this slot's parseCSSString, twelve distinct prerequisites across
the same six rules are supplied. Neither is six fully ported lint rules.

The helper accepts raw bytes and a byte colon offset. Nil is a tagged absent
result, separate from a fresh declaration with an empty value. Kind, property,
value, ValuePresent and Important are all compared. The original private Go
parseCSSDeclaration and actual Declaration constructor decide the answers.
The owned oracle overlay adds only an exported seam and standalone test entry.
No shared helper, compiler, harness or cohere source is changed.

Important search starts after the colon, matches exact lowercase bytes, selects
the first occurrence, and drops all following text, rather than trimming around
a suffix. strings.TrimSpace's Unicode White_Space set is used for both fields.
The byte implementation preserves malformed UTF-8, trims U+0085 and excludes
U+FEFF, U+180E and U+200B, where JavaScript trim or a historical Unicode table
would give a different answer. Empty values remain present. Missing colon (-1)
and an offset beyond input length return absent as Go does.

Caller contract: the view contains bytes, colon is an integer. An ordinary
separator returned by IndexByte is -1 or in [0, length). Go panics for colon ==
length or colon < -1; the owned helper explicitly refuses those caller errors,
with a descriptive Adamic panic rather than Go's runtime slice-panic wording.
Those panic texts are not claimed byte-identical and were not used as mutants.
The representation adapts Go byte strings into immutable byte arrays, not lossy
Unicode decoding. Consumers can translate the tagged result into their AST model.
No CSS parser integration is implied.

    source /workspace/adamic-tools/env.sh
    python3 stage1/cohere/lint/helpers/wave15/css_declaration/validate.py > /tmp/wave15-helper-css-declaration-validation.log 2>&1

The collector uses Go AST parsing and strconv.Unquote on every string literal
from all matching test files for each consumer: 131, 196, 48, 159, 130 and 139
distinct literals. It generates raw/multiple-colon/important-in-property and
important-in-value declarations, plus absent separators. This is fixture-derived
helper coverage, not measured production call counts. All six original rule
test families are independently run with their own assertions.

Extra controls exercise all 256 byte values in property and value boundaries,
all Go White_Space scalars and near misses, NUL, empty strings/values, repeated
important markers, wrong-case markers and trailing text. Baseline outputs require
normal exit and empty stderr. The mutant moves the search from colon+1 to zero;
important in a property then sets the wrong flag/value. It rebuilds emitted
JavaScript and sanitized native and is killed solely by comparison, never a
compiler refusal or sanitizer. Logs, generated cases, source-derived counts,
output hashes and the all-origin ownership check are in evidence/.

The prior CSS-string report records the successful setup (ready lines 0s,
warm/done 74s, nproc 5) and filtered uncached input oracle (six probes, PASS
6.521s). No new compiler or shared driver source changed after those checks.
Only owned sources and the claim were added. The initial VariantKind was yielded
to slot 04's earlier reservation; its source was removed in forward history and
its historical evidence is not counted as a delivered helper.
