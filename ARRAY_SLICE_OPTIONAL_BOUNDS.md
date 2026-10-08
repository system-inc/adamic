# Optional slice bounds, 2026-10-08 UTC

Fixture 25 at 541:34 calls from.slice(start,end) with optional numeric bounds. Bind receiver and both bounds before observing a default end length: an argument may change the source array. Undefined start means zero; undefined end means current length. Lower through the existing ArraySlice representation, preserving direct IR for entirely numeric/omitted bounds. No contextual type changes.

The generic array_slice_optional_bounds.a probe matches Node in both backends, uncached 0.487s. Cover number, string and object instantiations, missing/undefined bounds, finite negative indices, NaN and infinity, and operand effects in rse order with end growing the receiver. Counts alloc/free 29/29, retain/release 17/41, peak 13, regions 0.

Executed mutant defaults undefined end to zero instead of current length. Both backends finish with exit 0 and empty stderr but return empty slices where Node returns the elements. The oracle catches stdout (/tmp/array-slice-undefined-end-mutant.log), exit 1. Restored before final gates.

The initial side-effect probe used functions returning exactly undefined, which the base branch does not yet lower. Their annotations now use number | undefined without changing Node behavior, keeping the probe on this unit's representation. The full counts update remains blocked by the previously reported regexp_tree project-root attribution gap. No complete fixture 25 pass claimed; rerun its pinned scratch merge after pushing.
