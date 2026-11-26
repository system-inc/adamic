# Mutants of the uniform read's shape guard (#8nj5pn4, Emit's head)

The mutant replaces exactly one line, checked to match once, runs `go test -run '^TestNarrowedArmReadStopsOrAgrees$'
./internal/oracle/`, and is put back from a copy of the head file, checked byte for byte. It is red beside the same
pins green at head.

| Mutant | Line replaced | Red |
|---|---|---|
| M1 | `readFieldSlot` (internal/native/emit_objects.go): `return e.uniformFieldGuard(object, name, slot, ...)` made `return slot`, so the uniform read is bare again, as on main | TestNarrowedArmReadStopsOrAgrees/out_of_bounds.ts, ASan's heap-buffer-overflow read past B's allocation where the head stops in the NotYet's words; string_tags.ts, string_tags_handed_on.ts and enum_tags.ts, which print 1 again where Node prints NaN |

The honest twins stay green under M1: each call writes an A, so the bare slot reads A's own field, as the guarded
slot does.
