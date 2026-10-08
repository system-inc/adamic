# Undefined beside a narrowed array, 2026-10-08 UTC

Fixture 25 at 571:47 has string | readonly string[] | undefined narrowed to the array member. Undefined has object-pointer storage but typeof reports undefined, so it does not conflict with an object-valued reference under the existing typeof guard. Skip only concretely undefined members when comparing reference layouts. Null and actual object/array/map conflicts keep their current treatment. No contextual type change or new representation.

The narrowed_array_undefined.a probe agrees with Node in native and JavaScript, uncached 0.453s: present array, present string and absent value, plus the equivalent ordinary object union. Counts alloc/free 6/6, retain/release 11/12, peak 4, regions 0.

TestNarrowedArrayStaleMemberChecked holds its captured-write witness to Node (14), then requires both backends to panic before interpreting a restored string as an array. Its actual IR mutant removes that guard; both backends instead finish with exit 0, stdout 14, no stderr. Both are caught (/tmp/narrowed-array-undefined-oracle.log). A second executed compiler mutant drops the comparison of actual reference kinds; TestNarrowedUnionObjectTagRefusal loses the pinned array/Map refusal, got <nil>, exit 1 (/tmp/union-reference-kind-mutant.log). Both restored.

The positive probe was also run alone with verbose output proving its selection. Full counts regeneration remains blocked by the previously reported regexp_tree project-root attribution gap. No complete fixture 25 pass claimed; rerun the pinned scratch merge after pushing.
