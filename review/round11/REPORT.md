# Stream R, round eleven: P's shared slices, appends in place, and indexOf in place

Under review: claude/port-gitignore-adamic-1bstfd at 68f078b (6b0826d and on): string_share.c,
string_append.c, and the string header's new `owner` and `capacity`.

I ran everything on Linux x86-64 with Go 1.27.0, clang 18.1.3 and Node v24.21.0. `review/fxspptb/probe.sh` runs
each probe four ways:

1. Node on the source.
2. Native under ASan, UBSan and LeakSanitizer.
3. Native as `adamic build` makes it.
4. The JavaScript backend.

## Findings

None on the branch as it stands.

## What held

- **The design's one invariant is that count 1 means nobody else can see the bytes.** Every holder counts:
  - a shared slice retains its owner;
  - a slice of a slice points at the root owner, and retains that;
  - an array, a map key, an object field and another local each take a count.

  An append writes in place only when the count is 1 and there's room. A slice has capacity 0 and never
  grows in place, and neither does a constant. Every string-making path initializes `owner` and
  `capacity`: the heap allocators, and every stack piece, which the C zero-fills.
- **`strings.a`**, all four runs agree, no sanitizer report. Its cases:
  - a for...of over a string whose body appends to that same variable (the loop's string isn't freed under
    it);
  - a shared slice kept after its parent's variable moves on;
  - a slice of a slice of a slice over two-, three- and four-byte characters;
  - `+=` on a string an array, a map key, an object and another local also hold;
  - a lone high surrogate and a lone low one meeting across two appends;
  - a slice taken just before its parent appends.
- **`slice_sweep.a`**: 40 shared slices of a long mixed string, each cut at a different offset, plus an inner
  slice of each. Every slice's own length, every unit by `charCodeAt`, `codePointAt`, `indexOf` from a
  position and `lastIndexOf` agree on all four runs, so each slice's caches count from its own first byte.

## Not covered: the integration

P's branch merged main at 1d72913. Merging today's main (4ff4657) into it conflicts in ten files: emit.go,
lower.go, object.go, javascript.go, the oracle list, and others. So I couldn't test P's appends against main's
lent reads, Perceus and exceptions together.

On reading, lent reads take no count only as the direct operand of a pure consumer, and an append is no such
consumer. So I expect no collision, but I haven't observed that. These probes are the ones to run once P
merges main.
