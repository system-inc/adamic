# A presence test of an optional member: mutants (#p2n17c4 shape 1)

Each is one edit, applied alone, confirmed with `git diff`, run against the test it names, then put back with
`git checkout -- <file>`. Lower's tests are internal/lower/optional_member_presence_test.go (with the two rows moved
out of method_presence_test.go and optional_method_call_test.go), the oracle's internal/oracle/optional_member_presence_test.go.
M1 to M8 are the design's (shape1-design.md, section 5), M9 and M10 Lower's Object.prototype must, M11 and M12 the
prototype stops the check rests on, M13 the host's read it leaves to its own rule. M1 to M12 ran red as written
here at 42f21e23 (M8 green, as written); M13 is what the code says until a run quotes it.

## M1. The class instance admitted

File: internal/lower/optional_member_presence.go, checkOptionalMemberPresence.
Old: `			case literal.Class != 0:`
New: `			case false && literal.Class != 0:`
Run: `go test ./internal/lower -run '^TestOptionalMemberPresenceStops$' -count=1 -v`
Red: `a_class_instance_method.a` and `a_class_instance_field.a` lower, though natively the class's m isn't an own slot
and the test says none where Node says has.

## M2. The spread admitted

File: the same.
Old: `			case literal.Spread != nil:`
New: `			case false && literal.Spread != nil:`
Run: as M1.
Red: `a_spread_method.a` and `a_spread_field.a` lower. Only these rows catch it, since the copy's slot may answer as
Node does.

## M3. Unknown reach admitted

File: the same.
Old: `		if reach.Unknown {`
New: `		if false && reach.Unknown {`
Run: as M1.
Red: `an_arrow's_parameter_*.a`, `an_exported_entry_function's_parameter_*.ts` and `a_Map's_value_*.ts` lower.

## M4. A read lowering didn't keep admitted

File: the same.
Old:

```go
		if len(found[key]) != 1 {
			return l.notYet(read.node, tested+" that lowering didn't leave as one field read")
		}
```

New:

```go
		if len(found[key]) != 1 {
			continue
		}
```

Run: as M1.
Red: `a_class's_getter_field.a` and `a_literal's_getter_field.a`: finishAccessors made the read a dispatcher's call, so
nothing judges it and the program lowers. The method rows stay green, stopped earlier by eepPropertyPresence.

## M5. Every consumer admitted

File: the same, optionalMemberPresenceTest.
Old: `	if node.Kind != ast.KindPropertyAccessExpression || !presenceConsumer(node) {`
New: `	if node.Kind != ast.KindPropertyAccessExpression {`
Run: `go test ./internal/lower -run '^(TestOptionalMemberPresenceStops|TestMethodReadsThatKeepTheValueAreRefused|TestOptionalMethodCallGapsStop)$' -count=1 -v`
Red: `an_alias_method.a` (not refused), `a_field's_literal_method_read_as_a_value_field.a` (the scan is bypassed for an
alias too), `a_write_to_an_optional_method_method.a` (the write's read is admitted too),
TestMethodReadsThatKeepTheValueAreRefused/an_optional_method's_alias and TestOptionalMethodCallGapsStop/an_alias.

## M6. The scan bypassed for every read

File: internal/lower/iteration_origin.go, erasedLiteralMethod.
Old: `	if l.optionalMemberPresenceTest(node) {`
New: `	if true {`
Run: as M1.
Red: `a_field's_literal_method_read_as_a_value_field.a` no longer says "a literal method through a view that erases its
receiver".

## M7. The test folded to a constant

File: internal/lower/control.go, condition. Insert as its first statement:

```go
	if l.optionalMemberPresenceTest(ast.SkipParentheses(node)) {
		return ir.BooleanConstant{Value: true}, nil
	}
```

Run: `go test ./internal/lower -run '^TestOptionalMemberPresenceLowers$' -count=1 -v`, then
`ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestOptionalMemberPresenceAgreesWithNode$' -count=1 -v`
Red: every lower row whose test is a condition or the test of ?: finds no read of its member off its slot
(`a_condition_*`, `the_test_of_?:_*`, `a_literal's_method_shorthand_*`, `through_?._*`,
`a_slot_written_after_the_literal_is_built_*`, `a_presence_test_of_an_optional_method_on_{}_*`). The oracle's six programs
print has where Node prints none: a literal without m, and written's slot after undefined is written in.

## M8. Record storage admitted (unpinned)

File: internal/lower/optional_member_presence.go, checkOptionalMemberPresence.
Old: `			case literal.Record:`
New: `			case false && literal.Record:`
Run: as M1.
Red: none expected. No program reaches an H receiver with a record literal without a cast, so this stop has no pin
yet. The run records that it stays green.

## M9. Object's declared members not asked

File: the same, objectPrototypeName.
Old: `	return l.checker.GetPropertyOfType(l.checker.GetDeclaredTypeOfSymbol(object), name) != nil`
New: `	return false`
Run: `go test ./internal/lower -run '^TestOptionalMemberPresenceOfAnObjectPrototypeNameStops$' -count=1 -v`
Red: the two toString rows lower, where Node finds Object.prototype.toString on `{}` and says has.

## M10. The names the library doesn't declare not taken

File: the same, objectPrototypeName.
Old: `	if strings.HasPrefix(name, "__") {`
New: `	if false && strings.HasPrefix(name, "__") {`
Run: as M9.
Red: the two __lookupGetter__ rows lower, where Node finds Object.prototype.__lookupGetter__ on `{}`.

## M11. A __proto__ key not refused

File: internal/lower/object.go, objectLiteral.
Old: `			if property.Kind == ast.KindPropertyAssignment && fieldName == "__proto__" {`
New: `			if false && property.Kind == ast.KindPropertyAssignment && fieldName == "__proto__" {`
Run: as M1.
Red: `a___proto___key_*.a` aren't refused: the literal sets its prototype on Node, and the presence check, which reads
own slots, would admit it.

## M12. Object.setPrototypeOf and Object.create not refused

File: internal/lower/library_object.go, objectCallArguments.
Old: `	case "getPrototypeOf", "setPrototypeOf", "create":`
New: `	case "getPrototypeOf":`
Run: as M1.
Red: `Object.setPrototypeOf_*.a` and `Object.create_*.a` aren't refused as those calls.

## M13. stdout's handle's setBlocking taken from the host's rule

File: internal/lower/optional_member_presence.go, optionalMemberPresenceTest.
Old: `	if node.Name().Text() == "setBlocking" && l.stdoutHandleOrigin(node.AsPropertyAccessExpression().Expression, 0) {`
New: `	if false && node.Name().Text() == "setBlocking" && l.stdoutHandleOrigin(node.AsPropertyAccessExpression().Expression, 0) {`
Run: `go test ./internal/lower -run '^TestNodeProcessHandleSetBlockingCallAndTestLower$' -count=1 -v`
Red: its `if` and `not` rows stop as a presence test whose receiver may hold a value the program didn't build
(untracked source ir.ProcessCall), where the host's rule lowers them.

## M14. The record made where object.go's property reaches it, not where the refusal pass lets it through

File: internal/lower/refusals.go, methodValueStop, and internal/lower/iteration_origin.go, erasedLiteralMethod.
Old (refusals.go): `		l.recordOptionalPresenceRead(node)`
New (refusals.go): `		_ = node`
Old (iteration_origin.go): `		// refusal pass recorded it already (methodValueStop), for every road that lowers it.`
New (iteration_origin.go): `		l.recordOptionalPresenceRead(node)`
Run: `go test ./internal/lower -run '^(TestOptionalMemberPresenceRoadsStopOverAClassInstance|TestOptionalMemberPresenceLowers|TestOptionalMemberPresenceStops)$' -count=1 -v`
Red: any row of TestOptionalMemberPresenceRoadsStopOverAClassInstance that a road other than object.go's property
lowers is admitted unjudged. If every row still stops, each of those roads stops before lowering it today, and the
record at the exemption is a defense, as #4gpbn7w's fix was: the rows hold the stop, and this mutant has no witness
yet. TestOptionalMemberPresenceLowers and TestOptionalMemberPresenceStops stay green either way (Lower's pin: the
record taken out of erasedLiteralMethod moves nothing that passes through it).

Measured on Home at 8c082703 (J-6e61jk0, mutant applied as above, exit 0): every row stays green. Each of
the five roads reaches object.go's property and stops at the class arm, so M14 has no witness today and the
record at the exemption is a defense. TestOptionalMemberPresenceRoadsStopOverAClassInstance pins each road at
that stop, so a road that starts lowering elsewhere turns its row red.

## M15. The note finishAccessors takes of a dispatched presence read dropped

File: internal/lower/class_accessors.go, finishAccessors' ir.Property case.
Old: `					if expression.Absent && expression.Of == ir.Closure && expression.ViewWhere != "" {`
New: `					if false && expression.Absent && expression.Of == ir.Closure && expression.ViewWhere != "" {`
Run: `go test ./internal/lower -run '^TestOptionalMemberPresenceThroughAGetterDispatcher$' -count=1 -v` and
`ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestAccessorDispatchAbsentAgreesWithNode$' -count=1 -v`.
Red: the lowering rows, and internal/oracle's accessor_dispatch_absent presence.a and presence.ts, stop as a
presence test lowering didn't leave as one field read.

## M16. A getter-owning class passes whether or not the dispatcher calls its getter

File: internal/lower/optional_member_presence.go, judgePresenceReceiver.
Old: `		case literal.Class != 0 && (owners == nil || !owners.owners[site][read.name] || l.classChainHasMethod(literal.Class, read.name)):`
New: `		case literal.Class != 0 && owners == nil:`
Run: `go test ./internal/lower -run '^TestOptionalMemberPresenceThroughAGetterDispatcher$' -count=1 -v`
Red: `a class's getter and a class's method` lowers, though C's method lives on its prototype, where native's
fallback reads an own slot.
