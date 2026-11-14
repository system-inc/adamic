# Mutants for a captured boolean | undefined local held in a cell (#fds729j)

Each is one edit, run on a house box, then put back.

| Mutant | Edit | Red |
|---|---|---|
| The cell read as a plain boolean | native/emit_locals.go:151, `read`: a cell's MaybeBoolean is read as present, true only when the byte is 1, so the packed undefined (2) reads as false | to be measured. Expected: every pin, both dialects, natively, sanitized and on the counted build, on each line that prints a captured undefined. JavaScript stays green, since its cells hold JavaScript's own values and the JavaScript backend has no representation to get wrong |

The plain-boolean read replaces the line in `read` that is unique in the file:

```go
		value := unslotted(read.Of, slot+"."+member(read.Of))
```

with

```go
		value := unslotted(read.Of, slot+"."+member(read.Of)); if read.Of == ir.MaybeBoolean { value = "((adamic_maybe_boolean){true, " + slot + ".maybe_boolean == 1})" }
```
