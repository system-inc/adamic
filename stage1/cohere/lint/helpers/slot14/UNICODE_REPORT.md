# Unicode regexp escapes

Claim 3ac34517d was pushed before source creation, after all 610 origin heads,
20 distinct recursive helper claim contents and the HELPERS.md reservation
were inspected. Four consumers ties the highest remaining unclaimed concrete
fan-out. Base: origin/main 39638d9e278d38bb5aeae887f46d55a70e47aaad.

regexp_decode_unicode_escape.a supplies decodeUnicodeEscape and delegates
fixed-width spelling to the owned regexp_decode_fixed_hex.a. Inputs are raw
bytes, byte index, introducer width and Unicode mode. Source bytes are integers
in [0,255]; index/size are nonnegative integers, size <= 2^31-1, and
index + size <= source.length. Actual scanner callers use size 1. Go panics
outside this domain are not covered.

The implementation preserves Unicode-only brace syntax, the first closing
brace, fixed spelling in both modes, exact scanner width, uint32 parsing,
Unicode MaxRune acceptance, Go's acceptance of surrogate rune values, all
decoded default fields, errors.Is(ErrUnsupportedSyntax) and exact error bytes.
The U+2026 in Go's closed-brace failure is retained as E2 80 A6.

Reproduce with the setup environment sourced:
python3 stage1/cohere/lint/helpers/slot14/validate_unicode.py > /tmp/unicode-escape.log 2>&1

12,935 unique input records / 834,966 output bytes are identical with the actual
private Go helper on source Node, emitted JavaScript and ASAN/UBSAN native.
Controls include 0..4095, 0x10f000..0x1100ff, all byte values inside braces,
signed/prefixed/underscored/space-padded payloads, empty/unclosed/nested/multiple
braces, malformed UTF-8, EOF, byte indexes after multibyte prefixes and seeded
uint32/random-byte cases. This is bounded comparison, not an exhaustive proof
over arbitrary lengths or all integers.

All four original consuming Go suites pass assertions. Original private calls:
Next.js 0, no-empty-object-type 1, restricted-exports 0, restricted-imports 0.
Targeted actual consumer paths add 2, 1, 2, 2 calls respectively. Next.js checks
its actual route compiler's matches; the other three invoke their real option
decoder and rule. Captured inputs join the same three-runtime Go comparison.
No shared harness, compiler or Go submodule source is edited; overlays are
oracle-only. Full consumer findings/fixes on an Adamic regexp engine remain
outside this leaf-helper result and are not claimed.

Mutant: omit the closing brace from successful width. It compiles, exits zero
with empty stderr and is caught only by Go byte comparison on all three runtimes.
See evidence/unicode.log, unicode-coverage.json and the named consumer logs.

Dependency entries made available: four, for @next/next/no-html-link-for-pages,
@typescript-eslint/no-empty-object-type, no-restricted-exports and
no-restricted-imports. Complete rule blocker sets removed by this helper alone:
zero. Together with fixed hex: eight entries across the same four rules, zero
complete blocker sets removed.
