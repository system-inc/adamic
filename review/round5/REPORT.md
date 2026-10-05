# Stream R, round five: A2's reuse fixes (claude/release-globals-at-exit-uqmaf6, 4057ffb)

I ran everything on Linux x86-64 with Go 1.27.0, clang 18.1.3 and Node v24.21.0. `review/fxspptb/probe.sh`
runs each probe four ways:

1. Node on the source.
2. Native under ASan, UBSan and LeakSanitizer.
3. Native as `adamic build` makes it (-O2).
4. The JavaScript backend.

The probes here need A2's branch to compile (Weak and reuse).

## The four integrate-5 findings, re-run on 4057ffb: all closed

| Probe | integrate-5 | 4057ffb |
|---|---|---|
| `review/integrate5/forward.a` | heap-use-after-free; -O2 printed `2 start1 2 start1` | `1 start1 2 start1` on all four runs |
| `review/integrate5/global_sibling.a` | UBSan null dereference; -O2 segfault | `2 tree1` on all four |
| `review/integrate5/weak_during_spread.a` | `100 touched100` | `1 touched100` on all four |
| `review/integrate5/weak_after_reuse.a` | `new2 true` | `old1 false` on all four |

B's aliasing-range finding is closed too: `go test ./internal/flow` passes on 4057ffb (33 s), including
doubly_linked.a.

## The family hunted further, including the new array reuse (ad5e5fb): nothing found

- **Weak on the array paths:** `uniquelyHeld` (count 1 and no Weak) guards all three takeovers: the object
  spread, the in-place map and the leading array spread.
- **A map in place from `Weak<Box>[]` to `Box[]`** (`map_weak.a`): it does reuse the array of handles, and is
  correct. Handles are counted heap values, so releasing one and storing a target in its slot is sound.
  Agrees on all four runs, no leak. The other direction, `Box[]` seen as `Weak<Box>[]`, is NotYet.
- **A map in place while a Weak to the array is read by the callback** (`weak_array_map.a`): not reachable
  today. Reading through `Weak<number[]>` is NotYet. The probe is kept for when it lowers.
- **The rest, in `family.a`,** agrees on all four runs with no leaks:
  - one variable handed to two consumed calls in one statement (each gets a count);
  - a field of the source read in the same call;
  - a reassigned consumed parameter in a loop;
  - a popped element spread;
  - an in-place map whose callback returns its element;
  - a leading spread whose source is read twice (not taken over);
  - a discarded splice of an element a for-of still holds.
- **Labeled break and continue,** the liveness edges I wanted to test, are refused in 0.1. That path is
  closed.

## Gate on 4057ffb

- gofmt clean, vet clean.
- flow, load, lower and native pass.
- The oracle passes in a real checkout (exit 0, 201 s). Run in my worktree, its five input fixtures fail
  with "fork/exec ... permission denied". That's the worktree under a 0700 directory, which the drop to
  nobody can't reach, not the branch.

## B2's exceptions design

It isn't in docs/memory.md on any branch yet: I searched every remote branch for it. I'll read it when it
appears.

## Not covered

- Nothing ran on macOS or arm64.
- No mutants of 4057ffb beyond A2's own, which its commit lists.
