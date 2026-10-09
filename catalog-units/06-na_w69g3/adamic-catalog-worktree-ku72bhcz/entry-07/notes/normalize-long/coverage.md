# Normalize long coverage

Base: origin/codex/normalize-long, b1fec32. Its one commit is
`Normalize long strings without a whole output code point vector`.
Read CLAUDE.md, README.md, docs/0.1.md, docs/memory.md, the three-dot diff,
commit message, and all six changed files before writing fixtures.

Every new successful fixture calls NFC, NFD, NFKC and NFKD and repeats normalization
on its result. `show` prints UTF-16 length and every code point as hexadecimal
`point:consecutive-count` runs. This encoding is lossless, including lone surrogates,
NUL and supplementary points. It avoids hashes and sampling.

## Conditions and existing coverage

Existing programs below refer to internal/oracle/testdata. New filenames start
with normalize_coverage_. Covered means an input exercises the condition; it does
not claim measured C branch coverage or a separate mutant for every condition.

| Condition or value | Existing use | Added use |
| --- | --- | --- |
| NFC default and explicit; NFD, NFKC, NFKD; runtime form | normalize.a, normalize_long_marks.a, normalize_form.a | quick, all five successful fixtures |
| Undefined or omitted optional form | optional_strings.a, normalize.a | quick |
| Invalid form, including wrong case | normalize_form.a | Already covered |
| ASCII shortcut, empty string | normalize.a covers ASCII, no empty normalization | quick, long (million units), embedded NUL |
| Canonical mappings, compatibility mappings applied or skipped, recursive decomposition, mapping miss | normalize.a | quick, repeat, cache |
| Hangul LV and LVT decomposition; L+V and LV+T composition | normalize.a, normalize_long_marks.a | stream at batch boundaries, quick, long |
| Class lookup ASCII fast return, table hits and misses | normalize.a | cache visits 0x300..0x3ff and 0xa0..0x2ff three times |
| Mapping lookup low-point fast return, cache hit/miss and collisions | normalize.a has ordinary lookups | cache, repeat |
| Composition pair hits, negative hits, colliding keys, repeated revisits | normalize.a has ordinary pairs | cache scans 58 bases with four marks, three times |
| Quick check individually normalized composed starters, repeated cached point, different point, self join and neighbour join | normalize.a, normalize_long_marks.a have short cases | quick with long composed Latin, Hangul, Greek, Cyrillic, CJK, emoji, surrogate strings; repeat with V+L motifs |
| Quick check decomposed mapping absent, compatibility ignored, mapping present, Hangul rejected | normalize.a | quick, repeat, long |
| Ordered nonzero classes retained for D forms; descending classes rejected; composed marks conservatively rejected | normalize.a, normalize_long_marks.a | quick long sorted runs, stream many classes |
| Stable insertion sorting up to 64 marks; counting sorting above 64; ordered run; leading marks; multiple runs; starter terminates run | normalize_long_marks.a at 63/64/65 and 127/128/129 | stream with distinct equal-class acute/grave marks, 64/65/257/1025 motifs, CGJ separators; long with 12,288 marks |
| Composition adjacent, blocked equal class, allowed intervening lower class, failed pair, leading marks | normalize.a, normalize_long_marks.a | stream, cache |
| Segment below 64; at least 64 at a starter; kept final starter; final mark not kept; final flush | normalize_long_marks.a | stream with distinct prefixes of 62/63/64/65/127/128 and Hangul/Bengali tails |
| Repeated prefix below 64 bytes and at 64/65; copies below eight and at eight/nine; search up to eight points, motif of nine falls back | normalize_long_marks.a only has 16 copies | repeat with exact byte cutoffs and 1/2/7/8/9/16/17 copies; eight/nine-point motifs |
| Repeated prefix starts with mark (unsafe); cross-copy starter composition unsafe; safe marked/starter motif; partial final copy | normalize_long_marks.a has safe prefixes with tails | repeat with mark+A, V+L, FDFA, ffi, marked Latin, emoji, trailing scrambled marks |
| Last repeated copy stays decomposed for trailing reorder | normalize_long_marks.a | repeat, stream |
| Stream scalar decomposition cached vs recomputed; all starters vs containing marks; separate vs cross-copy composition | normalize_long_marks.a has some repetitions | repeat following nonperiodic prefix, including jamo, LV/LVT, FDFA, ffi, emoji |
| Stream repeated-scalar count zero, one and many; preceding batch changes final starter and rejects bulk path | No targeted program | repeat with L followed by repeated V, or repeated LV, plus two and nineteen ligatures |
| Encoded block copies zero blocks, one block, doubling, final partial doubling | normalize_long_marks.a uses 16 copies | repeat counts 2/7/8/9/16/17/19; long FDFA |
| UTF-8 1/2/3/4-byte encoding, UTF-16 BMP vs supplementary length, WTF-8 surrogates | normalize.a | quick, repeat, stream, long |
| Two passes: sizing with NULL output and writing; exact UTF-16 result units | All nontrivial existing normalization | Every new successful fixture observes length and all points |
| Limit guard in emit_blocks | No normalization limit fixture | limit: 29,826,161 FDFA -> 536,870,898 units rejected |
| Heap concatenation, append, short copied slice, long shared slice, split supplementary boundary | normalize.a has small built strings and surrogate slices | quick, stream, long |
| Mixed scripts and very long input | normalize.a has scripts separately; no million-unit normalization | stream mixed Greek/Cyrillic/Indic/Hangul/Arabic/music/emoji/surrogates; long |

## Cases without an oracle program

- malloc/realloc failure: deterministic allocation failure injection is unavailable in
  an Adamic program. Resource exhaustion would depend on the machine and affect Node
  differently. Buffer growth is exercised without forcing failure.
- Thread-local cache independence: Adamic programs have no exposed thread API for
  calling normalization concurrently. Cache collisions and repeated keys are covered
  within one thread.
- emit_point's over-limit branch (as distinct from emit_blocks): reaching it needs
  a non-bulk output above 536,870,888 units. The guard is inspected but not executed;
  an ordinary oracle fixture of that size with exact code-point output is impractical.
  The bulk guard is executed by limit.a.
- Benchmark-only argv validation, clocks, RSS limits and Python measurement loops
  are host test infrastructure, not callable normalization features in `.a` programs.
  The randomized host test is exercised by the repository gate.

## Practical limits

The first long fixture trial used 100,000 copies of each of three descending mark
classes. Its Node process exceeded the oracle's process deadline (exit -1, empty
stderr after the million-unit cases); native completed. That is an incomplete
reference observation, not a completed normalization-output mismatch. The obsolete
standalone Node trial was terminated. The retained fixture uses 4,096 copies of each
class, still well above the counting-sort threshold. Million-unit strings remain.
No completed semantic disagreement has been observed.
