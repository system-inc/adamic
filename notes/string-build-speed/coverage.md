# String build branch coverage

Base: origin/codex/string-build-speed at 0d3d76c. Read CLAUDE.md, README.md,
docs/0.1.md, docs/memory.md, the three commit messages, the merge-base diff,
and the changed files before writing the programs. The branch changes runtime
metadata and decoding, not the language's accepted call forms.

No differing programs were observed. Six new programs are registered explicitly
in internal/oracle/oracle_test.go; that list, not a testdata directory scan, picks
them up. Each is short source with bounded loops and prints observable UTF-16
units and/or code points. All new program files end in .a.

The table names existing programs in internal/oracle/testdata unless otherwise
stated. New names omit the common string_build_ prefix and .a suffix.
"Existing" means the source uses the case, not measured C branch coverage.

| Changed code case | Existing program using it | Added coverage |
|---|---|---|
| allocate: empty versus nonempty; empty units are known | ascii_scan.a, string_append.a | join (empty array), boundaries (empty slices) |
| from_number: finite, negative zero, NaN, infinities; ASCII fact | integer_format.a, number_edges.a, number_formats.a | join with number values and numeric template interpolation |
| concat: one stack piece, unknown count, normalize adjacent halves | from_codes.a, library_string_existing.a | repeat with low-high copy seams; join with surrogate separators |
| concat: multiple pieces, ASCII versus two/three/four-byte text and lone halves | string_append.a, string_positions.a, utf8_view.a | join, padding, repeat |
| concat: +, template with string/number/boolean, concat method | strings.a, library_string_indices.a | join, including zero method arguments |
| concat: no pieces | none; from_codes.a tests the empty single-piece builder | zero-piece runtime call is not directly expressible |
| put: written < 3; part bytes < 3 (zero, one, two) | string_append.a seams | join separator '', a, é |
| put: left lead != ED; right lead != ED | string_append.a readBetween, utf8_view.a | join and padding with ASCII, é, 世, 🌍 |
| put: left ED but not high; right ED but not low; reversed halves | string_append.a seams | join, repeat, padding with lone high/low and low-high |
| put: high meets low; tail after first three bytes; empty intervening parts | string_append.a seams | join, padding with receiver/fill halves meeting |
| append: initial units plus every part's units, ASCII cleared by Unicode | string_build_caches.a | cached_reads grow with multiple parts |
| append: unique with capacity, allocation growth, alias/refcount > 1 | string_append.a, string_build_caches.a | cached_reads grow, repeated cached reads before append |
| append: itself among parts, constant, shared slice, borrowed parameter | string_append.a, shared_slices.a | already covered |
| append: += and text = text + parts, loop then indexed | string_append.a | cached_reads grow uses both forms and bracket reads |
| append: global/captured assignment uses concat instead of in-place append | string_append.a | already covered |
| builder: unknown units, exact supplied units, ASCII and non-ASCII | from_codes.a, strings_more.a; repeat used throughout | repeat and padding with every encoding width |
| repeat: zero/NaN/fraction/one/many copies; empty source | string_limits.a, defaults.a; ASCII positive counts throughout | repeat adds Unicode, supplementary, lone halves and low-high seams for 0, 1, 2, 3.9, NaN |
| pad: default/undefined/empty fill, no-op target, NaN and negative target | optional_strings.a, string_limits.a, number_edges.a | padding adds no-op targets with Unicode fills |
| pad: at_start true/false; whole fills, no whole fills, remainder 0/nonzero | optional_strings.a for ASCII fills | padding adds two/three/four-byte fills, lone halves, mixed fill, fractional target and pair-splitting remainder |
| share: whole retain, small copy (<64), owner-fraction copy, shared bytes | shared_slices.a | boundaries at lengths 1, 63, 64, 65, 74, 75, 76, 150 for every width |
| share: propagate ASCII in copy/shared; unknown non-ASCII metadata | shared_slices.a | boundaries compares units and nested substring for every width |
| share: owner is original or already-shared; constant versus heap owner | shared_slices.a (nested heap slices), library_string_indices.a (short literal) | cached_reads adds indexed long literal; calls adds a long literal shared slice and a nested shared slice |
| charCodeAt: unknown/built index, ASCII, short walk, long heap cache | ascii_scan.a, string_positions.a, library_string_existing.a | cached_reads covers each width and supplementary threshold at 60/64/68 bytes |
| charCodeAt: literal sentinel versus NULL versus real view | library_string_indices.a (short literal), string_positions.a (heap) | cached_reads adds long supplementary literal |
| charCodeAt: in range, fractions, NaN, negative, end, infinities | ascii_scan.a for ASCII; library_string_indices.a has charAt/codePointAt | cached_reads adds charCodeAt edges before and after cache creation |
| codePointAt: ASCII, short walk, direct view, NaN/fraction/negative/end/infinity | library_string_indices.a, string_positions.a | cached_reads includes codePointAt-first cache creation |
| view decoding: widths 1, 2, 3, 4; high and low slots of supplementary point | string_positions.a, library_string_existing.a | cached_reads covers all with mixed ASCII and supplementary strings |
| codePointAt view: not high, high+low combines, terminal high, high+nonlow | string_build_caches.a for high+low and terminal high | cached_reads adds high+x, high+high, lone low, terminal high in long view |
| index checkpoints: point beginning, low half, exact terminal checkpoint | string_positions.a, shared_slices.a | cached_reads at positions 31, 32, 33, lengths around threshold |
| view lifetime: free on release, invalidate before reused append | string_build_caches.a, string_append.a | cached_reads growth, all new programs leak-checked |
| slice: ASCII/two-byte/three-byte/four-byte/lone surrogate boundaries | shared_slices.a, string_positions.a, utf8_view.a | boundaries exhausts every start/end of a small mixed built string |
| slice: no split, start low, end low, both low, empty/reversed bounds | shared_slices.a, string_positions.a | boundaries full sweep |
| slice/substring: omitted end, negative/clamped/reversed arguments | library_string_indices.a, shared_slices.a, number_edges.a | boundaries on built mixed strings and copied/shared long strings |
| prototype .call for concat, repeat, padStart/padEnd, slice/substring and cached reads | library_string_prototype.a has ASCII repeat/padStart/slice and short reads | calls adds missing concat/padEnd/substring and built supplementary receivers |
| length: before any reads, after canonicalized seams, after append | string_build_caches.a, string_append.a | every new program |

## Limits of Adamic programs

- SIZE_MAX byte-addition overflow and append's SIZE_MAX/2 bound cannot be
  reached through legal strings: the much smaller V8 UTF-16 limit intervenes.
- Existing string_too_long.a, concat_too_long.a and pad_too_long.a cover the
  observable V8 length limit. No additional multigigabyte near-limit allocation
  was attempted on this worker.
- malloc/realloc out-of-memory branches cannot be induced deterministically
  from a small pure program without an external allocation fault injector.
- The index's > UINT32_MAX/2 byte fallback requires a huge allocation and was
  not attempted. Stack pieces have reference count zero and no literal marker;
  pure Adamic cannot set these runtime fields or force an index onto such pieces.
- Unknown versus propagated metadata and the ASCII flag are implementation
  facts with identical output when correct. Programs check semantics, lifetime,
  counts, and the cache mutation below, not a speed guarantee.

## Mutation proof

Temporarily changed only internal/native/runtime/string_index.c's low slot:

```c
index->view[at++] = (uint16_t)(0xdc00 + ((point - 0x10000) & 0x3ff));
```

The mutant used 0xdc01. The uncached oracle for string_build_cached_reads.a
failed by stdout comparison, with exit 0 and empty stderr on both sides.
For the long repeated globe, Node printed `80 55356 127757`; native printed
`80 55356 127758`. Its low-unit read became 57102 instead of 57101.
The C compiled successfully; no sanitizer or -Werror failure killed it.
A Python finally block restored the original. No runtime changes are committed.
