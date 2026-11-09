# Record source handles: its mutant (#ny0c3gb)

entriesRecordStorageTypes relates a record literal as the types of the places that hold it (entriesRecordHandles),
and `satisfies T` is not one: the binding holds the operand's own type (rulings.md:14), so
`const table = {} satisfies Record<string, number>` is held as `{}`. One edit, run on a house box, then put back, with
`go test ./internal/oracle -run '^TestRecordSourceHandles$' -count=1 -v`.

The mutant counts `satisfies T`'s type as a holding place. In entries_records.go, entriesRecordHandles, the old text

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
not two. The other three subtests stay green: bare and generic hold no satisfies, and handle.ts already holds its record
as Record<string, number>.
