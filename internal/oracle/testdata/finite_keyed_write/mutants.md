# Mutants for finite keyed writes (#13p3y0c)

Each mutant is one exact edit, put back from the commit after its run, and each red counts only beside
its tests green at head. (a) to (f) ran through Loom's door as J-szx64ff at 5faa9634 on Lunatic, and (g)
as J-80fz499 at 0c9141f8 on Lunatic: the tests green unmutated, and every mutant red on the test its Red
names. (a) reds more than its Red lists: every program in TestFiniteKeyedWritesAgreeWithNode.

The tests: `TestFiniteKeyedWritesAgreeWithNode` and `TestFiniteKeyedWriteStops`
(internal/oracle/finite_keyed_write_test.go), `TestComputedFieldsWriteRefusalPin`
(internal/oracle/element_access_fields_test.go) and `TestFiniteElementWriteStoresEachNamedFieldThenPanics`
(internal/lower/element_access_fields_test.go).

## (a) setIndex's fixed-object arm dropped

internal/lower/object.go, setIndex. Old:

```go
	if array.Type() == ir.Object && !checker.IsTupleType(l.checker.GetTypeAtLocation(access.Expression)) {
		return l.finiteElementWrite(target, array, valueNode)
	}
```

New: the three lines gone.

Red: every program in `TestFiniteKeyedWritesAgreeWithNode` stops at its write, "assigning an element of a
value"; `TestComputedFieldsWriteRefusalPin`, fields.a at 12:5 in those words, its stop before #13p3y0c, where it
wants 26;
`TestFiniteElementWriteStoresEachNamedFieldThenPanics` at its lowering; and in `TestFiniteKeyedWriteStops`
every keyed row that stops inside finiteElementWrite (absent_optional_keyed.a, index_signature_union.a,
readonly_constructor.a, receiver_alternatives.a, value_held_differently.a), each now in those words.
checked_view_keyed.a stays green: its stop is the view's, made when the cast lowers, before the write.

## (b) the index-signature loop never fires

internal/lower/element_access_fields.go, finiteElementWrite. Old:

```go
		if len(l.checker.GetIndexInfosOfType(alternative)) != 0 {
```

New:

```go
		if len(l.checker.GetIndexInfosOfType(alternative)) != 0 && false {
```

The loop stays, so the mutant compiles; dropping it whole leaves `alternative` unused, which the compiler
catches, not a test.

Red: `TestFiniteKeyedWriteStops/index_signature_union.a` lowers. It is the one road found to this loop:
a mutable string index signature is a record, written by records.go before setIndex, and a readonly one
held as fixed storage stops at enumerationIndexedWriteRefusal through a parameter or a non-literal const,
and at entriesRecordWrite through a literal const. Only a union with an alternative that has no index
signature (so the union has none) reaches finiteElementWrite with one.

## (c) the readonly check dropped

internal/lower/element_access_fields.go, finiteElementWrite. Old:

```go
		if l.checker.IsReadonlySymbol(field) {
			return nil, l.notYet(target, "a computed write to readonly field "+names[index])
		}
```

New: the three lines gone.

Red: `TestFiniteKeyedWriteStops/readonly_constructor.a` lowers.

## (d) the panic past the last key test dropped

internal/lower/element_access_fields.go, finiteElementWrite. Old:

```go
	chain := []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: l.constant("computed object key outside its proven finite union")}}}
```

New:

```go
	chain := []ir.Statement{}
```

Red: `TestFiniteElementWriteStoresEachNamedFieldThenPanics`, nothing after the last key test. Every
agreement fixture stays green, since the checker proved no key falls outside the union: this is why the
IR test is there.

## (e) the view's computed-write refusal dropped

internal/lower/interface_cast.go, view, the computed-write block. Old:

```go
						if of, _ := l.representation(l.checker.GetTypeOfSymbol(field)); of == ir.Object {
							refused = l.notYet(part, "writing a checked object field without its source-slot type certificate")
							break
						}
```

New:

```go
						if of, _ := l.representation(l.checker.GetTypeOfSymbol(field)); of == ir.Object {
							break
						}
```

Red: `TestFiniteKeyedWriteStops/checked_view_keyed.a`, whose write then lowers while its named twin,
checked_view_named.a, still stops. This is Adamic's condition failing: a keyed arm admitted where o.f = v
stops. Dropping the whole block also turns `TestViewReceiverCallersAreClassified` (internal/lower) red,
since view's calls to checkedViewReceiverField fall from 3 to 2; this narrower edit keeps that count, so
only the twin row catches it.

## (f) the absent-optional judgment dropped

internal/lower/element_access_fields.go, finiteElementWrite. Old:

```go
		if err := l.absentOptionalFieldWrite(target, access.Expression, field, names[index], false); err != nil {
			return nil, err
		}
```

New: the three lines gone.

Red: `TestFiniteKeyedWriteStops/absent_optional_keyed.a` lowers while its named twin,
absent_optional_named.a, stops: the other half of Adamic's condition.

## (g) the keyed placeholder's stop dropped

internal/lower/assignments.go, the assignment statement. Old:

```go
	if l.uninitializedInitializer(binary.Right) && !ast.IsIdentifier(target) && target.Kind != ast.KindPropertyAccessExpression {
```

New:

```go
	if l.uninitializedInitializer(binary.Right) && !ast.IsIdentifier(target) && target.Kind != ast.KindPropertyAccessExpression && target.Kind != ast.KindElementAccessExpression {
```

Red: `TestFiniteKeyedWriteStops/placeholder_keyed.a`. This stop on main, before setIndex, is what keeps
`slots[side] = undefined!` out of finiteElementWrite, which would lower `undefined!` as an ordinary value
where its named twin, `slots.left = undefined!`, is a placeholder reset with setProperty's own checks.

## (h) each named field's slot judgment dropped

internal/lower/element_access_fields.go, finiteElementWrite. Old:

```go
		if err := l.refuseInSlot(valueNode, slot); err != nil {
```

New:

```go
		if err := l.refuseInSlot(valueNode, slot); err != nil && false {
```

Red: the registry twins in `TestFiniteKeyedWriteStops` whose rule the union key's intersected write type
hides, each keyed write then lowering, or stopping in other words than its named twin.
