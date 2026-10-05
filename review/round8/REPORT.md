# Stream R, round eight (first half): A2's regions

Under review: claude/release-globals-at-exit-uqmaf6 (0222103), `internal/native/region.go`, `runtime/region.c`,
the "Arenas" design in docs/memory.md, and regions.a.

I ran everything on Linux x86-64 with Go 1.27.0, clang 18.1.3 and Node v24.21.0. `review/fxspptb/probe.sh`
runs each probe four ways:

1. Node on the source.
2. Native under ASan, UBSan and LeakSanitizer.
3. Native as `adamic build` makes it.
4. The JavaScript backend.

The second half, re-running `move_throw.a` and the rest of the Perceus-meets-exceptions family on
cloud/integrate-7-flow, waits for that branch. It wasn't pushed when this was written.

## Findings

None so far. Read line by line, the plan is conservative wherever I could find a way out.

## What held, and why

- **The region is handed narrowly.** A fresh function hands its region only to a call that is a direct field
  value of the literal it returns, or the return itself (emit.go's `handRegion` at the return and at each
  literal field). A call inside an array literal, a conditional or `??` gets the heap. So arrays and maps
  inside region objects hold only heap values, and no region object can be reached except through a chain of
  `Property` reads, which `escaping` follows.
- **`escaping` is conservative.** A parameter, or anything derived from it through `Property`, `Narrow`,
  `Unwrap` or `CheckedCast`, escapes when it is stored, declared, assigned, returned, printed, iterated,
  captured, or used as the operand of anything not on borrow.go's pure-consumer list. Passthroughs (`??`, the
  conditional, `Box`, `MaybeOf`) count as escapes.
  - Pure operations that can return an existing reference: `ArrayIndex`, `MapGet` and `Trim` can't return a
    region object, by the point above. A string they return is a counted heap value, which the region's end
    releases correctly.
- **`region_escapes.a`, ten ways out regions.a doesn't try:**
  - a callee that returns `node.left`;
  - one that pushes its parameter into a global array;
  - one that puts it in a global Map;
  - one that hands it to a closure that keeps it;
  - `node.left ?? node` and `flag ? node.right : undefined` kept;
  - two levels of consumer, where the inner one keeps `node.left`;
  - a pure `trim` of a label kept in a global string.

  Each kept value is read after its statement. All four runs agree, with no sanitizer report.
  - Only one statement got a region, the `trim` one, and it is right: the string is counted.
  - Every other shape was correctly kept off the region.

## For when regions meet exceptions

They aren't on one branch yet: region.go is on A2's branch, exceptions.go on integrate-7. When they merge:

- A throw out of a statement that has a region must run `adamic_region_end` on its cleanup path. Otherwise the
  region's blocks, and the heap values its objects hold, leak. A fresh function's own region parameter needs
  no cleanup, since its caller owns the region.
- A region value can't be what's caught. A caught value is an Error made by `new Error`, which isn't an object
  literal, so it's never fresh.

## Not covered

- Mutants of region.go beyond A2's six.
- macOS and arm64.
