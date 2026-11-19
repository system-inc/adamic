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
