# Mutants for a captured Union local held in a cell (#n7sx9gz)

Each is one edit, run on a house box, then put back.

| Mutant | Edit | Red |
|---|---|---|
| The guard restored | locals.go, `localOf`: `declared.Captured && !boxedSlot(declared.Type) && l.unlowerable == nil` made main's `declared.Captured && slotless(declared.Type) && !(declared.Placeholder != "" && declared.Type == ir.Union) && l.unlowerable == nil`, so only a placeholder Union is admitted | every pin, both dialects, stops with main's words (54397e8f's proof) |
| The cell held as a plain reference | locals.go, `localOf`: directly above the guard, a captured Union that isn't a placeholder is retyped Object (the block below), so the cell and every read after the capture take it for an object | every pin stops at its first `=== "stopping"`, a BinaryExpression with a value and a string (54397e8f's proof) |
| The overwritten value kept | native/emit_locals.go, `store`: a plain cell's old reference outside an async frame is no longer released, `e.line("adamic_release(%s);", old)` inside `} else {` made `e.line("(void)%s;", old)`, so an overwritten member leaks (Memory's mutant: the counted build reds it, where ASan alone may not) | by_value, released and alternating leak, on the sanitized leak check and the counted build alike (54397e8f's proof). class, object, graph, cycle and nested hold graph cells, whose store drops the old value through `dropIn` instead, so this edit doesn't reach them |
| The overwritten value kept in a graph cell | native/emit_locals.go, `store`: a graph cell's old reference is no longer dropped, `e.dropIn(e.cellReference(local), old)` made `e.line("(void)%s;", old)` | measured on the head's proof |

The plain reference, inserted directly above the line `		if declared := l.result.Locals[local]; declared.Captured && !boxedSlot(declared.Type) && l.unlowerable == nil {`:

```go
		if declared := l.result.Locals[local]; declared.Captured && declared.Type == ir.Union && declared.Placeholder == "" {
			l.result.Locals[local].Type = ir.Object
		}
```

The overwritten value kept, the three lines in `store` that are unique in the file:

```go
		} else {
			e.line("adamic_release(%s);", old)
		}
```

made

```go
		} else {
			e.line("(void)%s;", old)
		}
```
