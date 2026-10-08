# Positioned lastIndexOf, 2026-10-08 UTC

Fixture 25 at 470:73 searches backward with an explicit position. Lower string lastIndexOf using the existing UTF-16 slice and backward search: clamp the position, keep the prefix ending at position plus the needle length, then search it. A match must start no later than position. Capture receiver, search and position once in source order. Missing, undefined and NaN positions mean the full length. Support present string search and number/undefined positions; unsupported ToPrimitive conversions remain NotYet. No contextual type changes and no new runtime helper.

Uncached fixture string_last_index_position.a agrees with Node in native under sanitizers and JavaScript (0.457s). Cover fractions, both infinities, NaN, undefined, omitted positions, empty/overlong needles, surrogate halves and operand side effects. Node stdout: 6,6,-1,8,8,-1,8,7,11,0,-1,0,1,0,8,6,rsp on successive lines. Counts: alloc/free 32/32, retain/release 9/43, peak 4, regions 0.

Executed mutant omits the needle length from the prefix bound. Both backends finish with exit 0 and empty stderr but give -1 rather than 6 for the bounded multi-unit matches. The oracle catches stdout disagreement, exit 1 (/tmp/string-last-index-length-mutant.log). Restored before final verification.

The full counts updater remains blocked by the previously reported regexp_tree project-root attribution failure. The new row was measured with the ordinary counted helper. No complete host fixture pass claimed here; the pinned scratch merge is rerun after pushing this change.
