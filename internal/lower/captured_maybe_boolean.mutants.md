# Mutants for a captured boolean | undefined local held in a cell (#fds729j)

Each is one edit, run on a house box, then put back.

| Mutant | Edit | Red |
|---|---|---|
| The cell read as a plain boolean | native/emit_locals.go:151, `read`: a cell's MaybeBoolean is read as present, true only when the byte is 1, so the packed undefined (2) reads as false | written, read_after, optional_parameter, narrowed and reporter, both dialects, natively, sanitized and on the counted build, each printing false where Node prints undefined, their JavaScript runs green (5661e965, J-g0g67fj). uninitialized, initialized and hoisted the same way: an uninitialized let's first read prints false (fresh false, fresh false), the let initialized to undefined prints false, and hoisted's cell, made before its declaration, prints false before and after true, all three green on JavaScript (0b1bb8dd, J-4hkj0ar: 48 red, 16 JavaScript green). The JavaScript backend's cells hold JavaScript's own values, so it has no representation to get wrong |

The plain-boolean read replaces the line in `read` that is unique in the file:

```go
		value := unslotted(read.Of, slot+"."+member(read.Of))
```

with

```go
		value := unslotted(read.Of, slot+"."+member(read.Of)); if read.Of == ir.MaybeBoolean { value = "((adamic_maybe_boolean){true, " + slot + ".maybe_boolean == 1})" }
```

## Not yet a cell

Two shapes stop at other walls, pinned by the oracle's TestCapturedMaybeBooleanStops in the words
measured at 5661e965 (J-g0g67fj):

- async: flow/async_normalize.go refuses every boolean | undefined local of an async function,
  captured or not, before its frame's cells are laid out ("async slots holding boolean or undefined
  are not yet proven"). The frame cell async_ordinary.go fills from a parameter is the next step.
- early: an arrow written above the let it captures is lowered before the let's local exists, so
  its read stops as "reading seen", as getters_census_lazy_object.a's does with an object.
