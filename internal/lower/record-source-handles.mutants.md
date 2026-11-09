# Record source handles: its mutants (#ny0c3gb)

entriesRecordStorageTypes relates a record literal as the types of the places that hold it (entriesRecordHandles),
and typesMayReach asks, for records, whether `{}` is assignable to the receiver (recordEmptyObject), since any record
can be seen as `{}` and then as whatever `{}` is assignable to. Each mutant is one edit, run on a house box, then put
back, with `go test ./internal/oracle -run '^TestRecordSourceHandles$' -count=1 -v`.

## (a) `satisfies T` counted as a holding place

`satisfies T` holds nothing: the binding holds the operand's own type (rulings.md:14), so
`const table = {} satisfies Record<string, number>` is held as `{}`. In entries_records.go, entriesRecordHandles, the
old text

```go
		if held.Kind == ast.KindAsExpression && !ast.IsConstTypeReference(held.AsAsExpression().Type) {
			handles = append(handles, l.checker.GetTypeFromTypeNode(held.AsAsExpression().Type))
		}
```

becomes

```go
		if held.Kind == ast.KindAsExpression && !ast.IsConstTypeReference(held.AsAsExpression().Type) {
			handles = append(handles, l.checker.GetTypeFromTypeNode(held.AsAsExpression().Type))
		}
		if held.Kind == ast.KindSatisfiesExpression {
			handles = append(handles, l.checker.GetTypeFromTypeNode(held.AsSatisfiesExpression().Type))
		}
```

Red: TestRecordSourceHandles/satisfies. The record is held as Record<string, number> alone, which Options (a required
label) relates to in neither direction, so Options's read loses its weak_field_read and the fixture lowers one helper,
not two. Every other subtest stays as it was: no other fixture holds a record under `satisfies`.

## (b) The `{}` question dropped

In entries_records.go, typesMayReach, the old text

```go
	if records && receiver.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) != 0 {
		if empty := l.recordEmptyObject(); empty == nil || l.checker.IsTypeAssignableTo(empty, receiver) {
			return true
		}
	}
```

becomes

```go
	if false && records && receiver.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) != 0 {
		if empty := l.recordEmptyObject(); empty == nil || l.checker.IsTypeAssignableTo(empty, receiver) {
			return true
		}
	}
```

Red: TestRecordSourceHandles/empty_source, whose record is held as Held, which relates to Read in neither direction, so
Read's read loses its weak_field_read (0 helpers, want 1). And TestRecordSourceHandles/t2_parameter_return, if it
lowers as its .stop reads: Read's read is plain, so native stops at the record's fixed read with the runtime's own
words and the JavaScript backend prints text, where the pin wants the weak read's stop. t1_empty and the other roads
stop at the hiding rule before lowering finishes (rulings.md:71), the second guard, and stay as they were. bare,
satisfies and generic stay green, since their Options declares a required label `{}` lacks, and so does handle.ts,
whose Hints its Record<string, number> handle already reaches.

## (c) A type parameter related as itself

The walk relates declarations, so a type parameter of a declaration around the literal is unknown there (Adamic's
ruling on #ny0c3gb, 16:12:33Z). In entries_records.go, entriesRecordHandles, the old text

```go
	if len(parameters) > 0 {
		unknowns := make([]*checker.Type, len(parameters))
```

becomes

```go
	if false && len(parameters) > 0 {
		unknowns := make([]*checker.Type, len(parameters))
```

Red: TestRecordSourceHandles/generic_holder. The record is held as Record<string, T | undefined> with T rigid, and
Shape, `{ label: string; size?: number }`, relates to it in neither direction (a string isn't a T | undefined), and `{}`
lacks Shape's label, so Shape's read loses its weak_field_read (1 helper, want 2). The other subtests hold no record
inside a generic and stay as they were.
