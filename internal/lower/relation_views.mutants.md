# relation_views.go mutants (#0b9xqhh, rulings.md:72's first slice)

Described, not run. Each names the check that must turn red, and why no other check catches it.
The fixtures are in `internal/oracle/testdata/relation_views/`, held by
`TestRelationViewsAgreeWithNode`, `TestRelationViewsPanicAtTheRead` and `TestRelationViewsStop` in
`internal/oracle/relation_views_test.go`. Mutants (a), (c) and (d) each turn a row of `TestRelationViewsStop`
red, and (b) a row of `TestRelationViewsPanicAtTheRead`; (e) to (g) say why no row can.

## (a) The post-pass removed

`checkRelationViewsTaken` returns nil at once (or its call in `lowerWithElementKey` is taken out,
with the scanning reference's beside it, so `TestScanningReferenceCallsEveryPass` stays green).

Red: `never_lowered.ts`'s row. Its relation sits in `admit<Tag>`, a generic nothing calls, so its body
is never instantiated and never lowered, and the entry rule 3300 recorded at `source` (9:28) is never
taken. The post-pass is the only thing that refuses it. With the mutant the program lowers: the row
wants a stop and gets a lowered program. The relation never runs, so nothing reads raw here; the row
holds the pass that stands behind every road the audit below didn't find.

`shorthand.ts` stays red under this mutant only with mutant (e) beside it: `shorthand` refuses its own
road now, in the same words at 14:20.

Why only that row: every other fixture's relation is a value `value` lowers, so it's viewed and taken,
or a shorthand, which refuses before the post-pass runs. The rule's own refusal can't catch it either,
since in .ts it returned nil to defer to the view.

## (b) The entry deleted when it's taken

`relationViewValue` does `delete(l.relationViews, key)` after viewing, instead of marking
`relationViewsTaken`.

Red: `lowered_twice.ts`'s row. `admit<Tag>` holds the relation `const target: Target = source;` and is
lowered once per instantiation, `admit<number>` and then `admit<string>`. The first lowering views
`source` and deletes the entry; the second finds none, so its value crosses with no view and its local
is no view entry. `second` comes only from the second instantiation, and its flow reaches no entry,
so `checkViewWidenedReads` doesn't weigh the read of `seen.label` through `Seen`, and the read takes the
number 5 raw as a string. Since #zpzm9ta the row wants the program to stop at that read, checked where it's
made, with exit 70 on every backend; under the mutant it never stops there.

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
  `TestRelationViewsStop`'s road1_getter.ts and road2_wider_alias.ts, which want stops lowering reaches
  only past the relation.

Why only these: each is a .ts program whose first stop on main was rule 3300's at a relation this slice
records. A pin refused by another rule first, or in .a, has the same outcome with the slice and without.

## (e) shorthand's own refusal removed

`shorthand` no longer refuses when `l.relationViews` holds the shorthand's name (object.go:2035).

No row turns red alone: `shorthand.ts` is still refused, by the post-pass, in the same words at
14:20. The refusal is there for Compiler's gate (below): the post-pass is one boolean per node, so a
shorthand's raw road would pass it whenever another lowering of the same name viewed it. No road
lowers a shorthand's name through `value` today, so no program shows that; with (a) beside it,
`shorthand.ts` lowers and reads its label raw.

## (f) Only the first hidden field asserted

`relationViewValue` asserts registration for `entry.hidden` alone, as the slice did, instead of each
of `entry.fields`; or `recordRelationView` holds only `found` to `relationViewAdmits`.

No row turns red: `two_hidden_depths.ts`'s second field, `Inner`'s `note`, is one viewSchema registers
(it walks every field of every contract it builds, interface_cast.go:332), and a second field a view
can't check (Weak, callable, unstored, reached through a constraint) has no fixture. The fixture is a
witness that a relation hiding two fields at two depths lowers and stops at the deep lie; the
assertion is for the fields the walk reaches that no contract holds (a call signature's result or
parameters, the fields below an unsupported contract).

## (g) The second-slot guard dropped

`recordRelationView` records over an entry it already holds for the node with another source or
target. `refuseInSlot` runs while lowering, and `namedWriteStop` (records_named.go:404) judges one
value against each named property's type in turn, so the entry keeps the last slot's type and the value
is viewed as that type alone, the others' hidden fields never registered. No fixture holds it.

## Every road a keyed node is lowered by (Compiler's gate)

Compiler's question: `relationViewsTaken` is one boolean per node, so a keyed node lowered once
through `value` and once by a road that never reaches `relationViewValue` would cross raw the second
time with the post-pass satisfied. The keyed node is the exact node `recordRelationView` keys,
`ast.SkipParentheses(site.value)`: an `AsExpression` in a slot, a shorthand's name, an object spread's
expression, or a value judged by its own type in a view site or an implied target (an initializer, the
right of `=`, `??=`, `||=` or `&&=`, an argument, a returned value, an arrow's expression body, a
literal's property or element, a parameter's or a class field's default, a conditional's arm, an
operand of `??`, `||` or `&&`). Its type is a structural object type the program declares, never
never, a primitive, a library's type or an instantiable one (`relationViewOpen`).

Found by reading, not by running: every function in internal/lower that takes an `*ast.Node` and gives
an `ir.Expression` (262 of them), each call site that hands it a child rather than the node `value`
dispatched, and every direct `ir.Read` of a local built from a node (`l.local`, `l.locals[...]`).

Verdict: (b). Every lowering of a keyed node views it or refuses. Every shape but one reaches lowering
only through `value`; the shorthand's name never does, and now refuses on its own road.

| Road | Where | Can it receive a keyed node? | What happens |
| --- | --- | --- | --- |
| `value`'s hook | expression.go:1270 | Yes, every shape but the shorthand's name | `relationViewValue` views it on every lowering, or refuses in 3300's words |
| `value`'s inspector trap | expression.go:1240, inspector_doomed.go:287 | Only a node holding a doomed `require` read, whose type is a host's | A trap: the receiver root is lowered through `expression`, then a panic, never the value |
| `value`'s enum never path | expression.go:1243 | No: it takes a node the checker types never, which the hiding walk exempts | `enumNeverCheck`, a helper whose body is a panic, never the value |
| `value`'s never read path | expression.go:1253 | No: the same never type | `neverReadTrap`, a panic after what the read evaluates |
| `expression`'s optional receiver read | expression.go:329 | No: the receiver of `a?.m()` is a property access's object, no view site | Gives back the local holding the receiver its first lowering, through `value`, made, viewed if it ever were keyed |
| `shorthand` | object.go:2034; called at object.go:102, records.go:461, class_accessors.go:186, from.go:86, library_json_stringify.go:269 | Yes, the shorthand's name | Reads the local directly. Refuses in 3300's words when the name is keyed (this change), on every lowering |
| Call arguments | arguments_length.go:127 (`callArguments`, every call, method, super and generator road: expression.go:2220, :2480, class.go:410), generators.go:485 | Yes | `expression`, so `value` |
| Library call arguments | object.go `arrayMethodArguments`, `mapMethod`, set.go `setMethod`, library_object.go `objectCallArguments`, records.go `recordObjectCallArguments`, and the rest of the library roads | Yes where the parameter is a program's type (`xs.push(source)`) | `expression` on each argument; the few that don't (`methodAliasRead`, library_method_values.go:172) read a library function value, never keyed |
| Spread arguments | arguments_length.go:93 | No: the keyed node would be the `SpreadElement`, whose slot is an array, which a view refuses | `expression` on the spread's operand |
| Object literal fields and spreads | object.go:31, :100; records.go:417, :459; class_accessors.go:184 (a spread there is NotYet) | Yes | `expression` |
| Array and tuple elements, Set and Map literals | object.go:313, `tupleLiteral`, library_map_set.go `librarySetLiteral`, object.go `newExpression`'s entries | Yes | `expression`; a spread element is `iterated`'s, whose operand is no view site |
| Initializers | locals.go:89, collections.go:427 (a destructuring's) | Yes | `expression`. Not lowered at all: a detached own method's alias (locals.go:87, a library method, never keyed), a placeholder `null!` (locals.go:74, never), a stored generator's call (locals.go:45, a library `Generator`) |
| Returns | statements.go:174 | Yes | `expression`. The void return of a generic call (statements.go:336) and `return panic(...)` (statements.go:323) are `expressionStatement`'s, and their type is void or never, never keyed |
| Arrow expression bodies | functions.go:401 | Yes | `expression`; a void body (functions.go:397) is a statement, its slot void |
| Parameter defaults, class fields | functions.go:354, class_inheritance.go:452, class_static.go:314 | Yes | `expression` |
| Assignments | class.go:490 (`setProperty`), object.go:2353 (`setIndex`), logical_assignment.go:30, assignments.go's local and destructuring roads, entries_records.go, construction.go, typed_arrays.go, element_access_fields.go | Yes | `expression` on the right |
| Conditional arms, logical operands | expression.go:1992, :2232, :2271, never_branch.go, host_bindings.go | Yes | `expression`; a never arm is `neverValue`'s, never keyed |
| Casts, `satisfies`, `!` | expression.go:1528, :1534, `nonNull` | Yes (the cast or the `!` itself) | Reached only through `value`, which views the node; the operand inside is no view site |
| `refuseInSlot` while lowering | records_named.go:404, records_define.go:135, library_method_values.go:434, element_access_fields.go:189 | It records entries, not lowers them | The value is lowered right after through `expression`. A second slot for the same value refuses (mutant (g)) |
