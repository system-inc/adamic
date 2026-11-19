# writeHidesField mutant (#4ysqm2m)

Described, not run. The rows are `TestGenericSlotHiding`'s, in
`internal/oracle/generic_slot_hiding_test.go`, over `internal/oracle/testdata/generic_slot_hiding/`.

## The helper returning false

`writeHidesField` (invariance.go) returns false at once, so the invariant-mutable rule judges each write
by assignability alone, as main did before this change.

Red, on both roads:

- `append_lie.ts`'s row. `held`, a `Target[]` seen as `Source[] | undefined`, passes `widenedArguments`
  again, since a `Source` is assignable to a `Target`. Rule 3300 sees nothing either: `append`'s T is
  `Source` (`infers_source.ts`), so no argument's slot hides a field on a read. The program lowers, `append`'s
  push stores the lying `Source`, and `describe` reads its label, the number 5, as a string: native
  segfaults (main's J-51aezx3, exit 139) where the row wants the refusal at 26:8.
- `append_lie.a`'s row, the same in .a, which wants the refusal at 22:8.
- `append_honest_source.ts`'s row. It lowers and agrees with Node, since its values are honest, where the
  row wants the refusal at 26:8: the type lets a lying Source through, which `append_lie.ts` shows.
- `field_lie.ts`'s row, the writable-field road: `Holder` seen as `Loose` passes `widenedProperties`,
  `loose.t = hidden` stores the lie, and the second read of `holder.t.label` takes 5 as a string, where the
  row wants the refusal at 14:22.

Green under the mutant, as without it: `append_honest.ts` and `readonly_view.ts` (no write can hide a
field: the first's element is `Target` on both sides, the second's view is readonly), and `infers_source.ts`
(tsc's own diagnostic, before lowering).

## Every place invariance.go judges a write by assignability

Each road where the invariant-mutable rule asks whether a value of one type can be written where another
is read, and what this change does there:

| Road | Where | Disposition |
| --- | --- | --- |
| A mutable container's element | `widenedArguments` | Covered: asks `writeHidesField` (`append_lie`) |
| A writable field | `widenedProperties` | Covered (`field_lie`) |
| A method seen as a writable field | `widenedProperties`, the `methodField` arm | Covered: the function written in is judged against the method's type by the same walk, which goes through call signatures' results and parameters |
| A writable record's value | `widened`'s record arm | Covered |
| A literal's setter seen as a property | `literalSetterRefusal` | Covered: a write through the property hands the setter the slot's type, a write into its parameter |
| A function's parameter seen as another's | `widenedParameters` | Not a write: a hand-off to a parameter, which rule 3300's own signature arm judges for hiding at the relation (`optionalWidenedEach` walks each given parameter into the taken one) |
| A census-only never-rest signature's result | `widened`, `literalMethodRefusal` | Not a write: a result read as the view's result, the read direction, which rule 3300 judges |
| A cast's operand | `refuseWidening`'s `AsExpression` arm | Not a write: the cast's relation, judged by the cast rules and rule 3300 |
| A union's or literal's member choice | `widened`'s union arm, `literalSlotMembers`, `methodSlotMembers`, `takenWhole` | Not a judgment: assignability picks which member a value is seen as, and each pick is walked by `widened`, which reaches the roads above |
| An open enum tag literal | `memberLiteralSlotWidened` | Not this hole: a number literal, which no field hides in |
