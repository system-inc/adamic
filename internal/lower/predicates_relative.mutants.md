# predicates_relative.go's wrong versions

Each is planted alone in the head's tree and must turn the named row of `TestPredicateRelativeTarget`
(internal/oracle/predicate_relative_test.go) red. The lie rows' programs are in predicate_relative_lies_test.go. Run on Loom as one red job; the n rows read nothing after the
call, so only the call's own check stops their lie.

- **optional-as-required**: a field held optional in a source member is judged by its declared type alone, the
  `!optional` clause dropped and the checker's missing `undefined` taken off its type. Red: `n1_optional_unread.ts`
  prints `named` where the call's check stops it (Language's condition (a)). Dropping the clause alone changes
  nothing (J-74w6cqx): the checker types an optional field with its missing `undefined`, which `castAssignable`
  refuses too, so condition (a) is held twice.
- **skip-refined**: a field whose type is assignable in no source member is skipped (the `castAssignable` clause
  dropped). Red: `n3_refined_unread.ts` prints `dst`.
- **absent-skipped**: a field a source member doesn't declare counts as held. Red: `n4_absent_unread.ts` prints
  `missing`.
- **bound-raised**: `untaggedRecord` answers false for every record (predicates_proof.go), so family 9 A's shape
  takes the plain overload-target words. Red: `r6_family9a.ts`, whose depth-bound words are pinned.
- **collection-as-membership**: a refined field held as a Map or a Set gets a checked read as a membership. Red:
  `s5_custom_option.ts`, which keeps main's stop by Adamic's road (a).
- **remainder-unviewed**: the false side's remainder registers no view. Red: `r5_lying_false.ts`, whose lying
  false then reads a field its value lacks unchecked instead of stopping at that first read (ruling 12 A).

Not planted: **index-as-declared**. A source member with an index signature is held as a record, never
`ir.Object`, so relativePredicateFields leaves its target on main's road before it asks which fields the member
declares (p03u keeps main's stop). Counting an index signature's answer as declared changes nothing a program can
reach today; condition (b) is held by the representation test and by `GetPropertyOfType`, which answers declared
properties alone.

- **undefined-raw** (Lower's must on #4cvdkk3, the arm this head had and dropped): a refined field held as undefined
  alone, read raw and tested undefined. Red: `v_cast_view.ts` and `v_param_view.ts`, which reach the call through a
  checked view and keep main's stop, where the raw read let JavaScript print `tail` and native stop or panic on a
  compiler bug (J-wb50y6y). Not planted as a job: restoring the arm is the whole of the dropped code, and the two
  rows' pending words are what it would turn.
