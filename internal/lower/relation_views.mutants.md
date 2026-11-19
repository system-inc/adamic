# relation_views.go mutants (#0b9xqhh, rulings.md:72's first slice)

Described, not run. Each names the check that must turn red, and why no other check catches it.
The fixtures are in `internal/oracle/testdata/relation_views/`, held by
`TestRelationViewsAgreeWithNode`, `TestRelationViewsPanicAtTheRead` and `TestRelationViewsStop` in
`internal/oracle/relation_views_test.go`. Each mutant below turns a row of `TestRelationViewsStop` red.

## (a) The post-pass removed

`checkRelationViewsTaken` returns nil at once (or its call in `lowerWithElementKey` is taken out,
with the scanning reference's beside it, so `TestScanningReferenceCallsEveryPass` stays green).

Red: `shorthand.ts`'s row. Its relation is `{ target }`, a shorthand property: rule 3300 records the
entry under the shorthand's name, but `objectLiteral` lowers it through `shorthand`, which reads the
local directly and never calls `value`, so no lowering views it. The post-pass is the only thing that
refuses it, in rule 3300's words at 14:20. With the mutant the program lowers with no view, `Target`'s
`label` is never registered checked, and `box.target.label` reads the number 5 as a string: the row
wants a stop and gets a lowered program, and the backends disagree with Node's 5.

Why only that row: every other fixture's relation is a value `value` lowers, so it's viewed and taken,
and the post-pass has nothing to refuse. The rule's own refusal can't catch it either, since in .ts it
returned nil to defer to the view.

## (b) The entry deleted when it's taken

`relationViewValue` does `delete(l.relationViews, key)` after viewing, instead of marking
`relationViewsTaken`.

Red: `lowered_twice.ts`'s row. `admit<Tag>` holds the relation `const target: Target = source;` and is
lowered once per instantiation, `admit<number>` and then `admit<string>`. The first lowering views
`source` and deletes the entry; the second finds none, so its value crosses with no view and its local
is no view entry. `second` comes only from the second instantiation, and its flow reaches no entry,
so `checkViewWidenedReads` doesn't weigh the read of `seen.label` through `Seen`: the program lowers,
and the read takes the number 5 raw as a string. The row wants the NotYet naming that read.

Why only that row: the post-pass sees the deleted entry as gone, not untaken, so it stays silent. The
reads through `Target` stay checked, since view registers `Target:label` program-wide on the first
lowering, which is why the fixture reads the second instantiation's value through another type. Every
other fixture lowers its relation once.

## (c) The .ts condition dropped

`recordRelationView` no longer asks whether the node's file ends in `.ts`.

Red: `adamic_twin.a`'s row, which wants rule 3300's refusal at 12:22 and gets a lowered program,
since the twin's relations are recorded and viewed as argument.ts's are. In the lower package,
`TestOptionalWideningRefused`'s .a fixtures whose relation is one plain object value entering a plain
object slot (testdata/optional_widening/initializer.a, pinned word for word) go red too.

Why only these: every .ts fixture already takes the view, and the .a rows are the only ones whose
answer depends on the file's extension alone. Rows whose source is `object`, a class instance or a
type parameter stay refused under this mutant, since their own condition refuses them.

## (d) The slice undone: rule 3300 refusing in .ts again

`refuseOptionalWidening` stops calling `recordRelationView` (or it returns false at once), so a .ts
relation that fails only on an absent optional field is refused by the hiding rule as it was before
#0b9xqhh. Each pin this slice moved from that refusal to a panic at the read goes red: its .stop file
holds `panics`, so its harness wants the program to lower and stop at the read with exit 70
(`panicsAtTheRead`), and gets rule 3300's refusal from Lower instead ("Lower: ... want it to lower and
panic at the read").

- `TestLogicalAssignmentViewsArePinned` (internal/oracle/logical_assignment_views_test.go):
  `v07_hiding_nullish.ts` and `v09_hiding_assign_control.ts`, refused at 4:9 and 4:7.
- `TestCastSlotHiding` (internal/oracle/cast_slot_hiding_test.go): the .ts twins s1, s2, s3, s4, s5,
  s6, s7, s8, s9, s11, s12, s13 and s14, each refused where its cast or value enters the Positioned
  slot. `s15_angle_assertion.ts` goes red too: it wants the TypeAssertionExpression NotYet and gets the
  refusal, which runs first. s10 and every .a twin keep the refusal, and stay green.
- `TestConstraintRelationSites` (internal/oracle/constraint_relation_test.go): `h2_hidden_direct.ts`,
  refused at 5:18. The other .ts rows are refused by constraintRelationRefusal (rule 3360) at the
  binding, in the same words, which this mutant doesn't touch.
- `TestRelationViewsAgreeWithNode` and `TestRelationViewsPanicAtTheRead` (every row), and
  `TestRelationViewsStop`'s road1_getter.ts, road2_wider_alias.ts, road3_other_type.ts and
  lowered_twice.ts, which want stops lowering reaches only past the relation.

Why only these: each is a .ts program whose first stop on main was rule 3300's at a relation this slice
records. A pin refused by another rule first, or in .a, has the same outcome with the slice and without.
