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

No red pin today, measured (Codex 01a1c13c at 55433524): a `{}` under `satisfies Record<string, number>` is never a
record, since the const holds its own `{}`, a fixed object, and satisfies.ts stops at its declaration ("a {} seen as
Record<string, number> (fixed objects and records have different storage)") before any read. So whether satisfies
holds anything has no program to observe yet, and this entry records the rule until such a record exists.

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
Read's read loses its weak_field_read (0 helpers, want 1).

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
