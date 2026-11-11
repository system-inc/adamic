# Mutants for a captured Union local held in a cell (#n7sx9gz)

Each is one edit, run on a house box, then put back.

| Mutant | Edit | Red |
|---|---|---|
| The guard restored | locals.go, `localOf`: `declared.Captured && !boxedSlot(declared.Type) && l.unlowerable == nil` made main's `declared.Captured && slotless(declared.Type) && !(declared.Placeholder != "" && declared.Type == ir.Union) && l.unlowerable == nil`, so only a placeholder Union is admitted | measured on the head's proof |
| The cell held as a plain reference | locals.go, `localOf`: directly above the guard, a captured Union that isn't a placeholder is retyped Object (the block below), so the cell and every read after the capture take it for an object | measured on the head's proof |
| The overwritten value kept | native/emit_locals.go, `store`: a cell's old reference outside an async frame is no longer released, `e.line("adamic_release(%s);", old)` inside `} else {` made `e.line("(void)%s;", old)`, so an overwritten member leaks | measured on the head's proof |

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
