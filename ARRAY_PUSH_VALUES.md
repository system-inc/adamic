# Multiple intrinsic push arguments, 2026-10-08 UTC

Fixture 25 at 739:9 calls components.push("**", "*"). Lower arbitrary fixed-count push arguments by binding receiver and all operands in source order before any append. Use ordinary ArrayPush IR for each append in order and return the final length. Zero arguments return the existing length. Preserve the direct one-argument IR so previous ownership counts stay unchanged. Spreads remain outside this addition. No contextual type changes.

The array_push_values.a probe matches Node in both backends (uncached, 0.544s): numbers and argument-time array length, zero arguments, owned strings, object references whose alias is subsequently mutated, final return lengths, and single receiver evaluation. Counts alloc/free 22/22, retain/release 22/40, peak 13, regions 0. Full lower package passed 15.223s; focused oracle passed 0.633s; vet clean.

Executed running mutant pushes the first argument while later arguments are still being evaluated, and repeats receiver evaluation for that early call. Both backends exit 0 with no stderr but produce 10,20,2 rather than Node's 10,20,1, and rrab rather than rab. The oracle catches stdout alone (/tmp/array-push-evaluation-mutant-running.log). Restored. Two earlier malformed mutant attempts caused a Go slice panic and cyclic IR stack overflow; these are not credited as semantic proof (/tmp/array-push-evaluation-mutant.log and /tmp/array-push-evaluation-mutant-valid.log).

Full counts regeneration remains blocked by the previously reported regexp_tree project-root attribution gap. No complete fixture 25 pass claimed; rerun the pinned scratch merge after pushing this stop.
