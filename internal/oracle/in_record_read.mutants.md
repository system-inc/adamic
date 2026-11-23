# Mutant for #859ksab's in-read stop

**R1, the stop dropped.** `inNarrowedMayBeRecord` (internal/lower/in_narrowed_write.go) answers false, so a read
of a key only an `in` test added lowers on any receiver, a record seen as object among them.

Must turn red: every subtest of `TestInRecordReadStops` (record_number, record_string, record_absent,
record_string_of, unknown_param, keyed_read, generic_param), each lowering where it must stop, its failure
carrying the sanitized native run.

Must stay green: `TestInRecordReadFixedAgrees`, and #yqbg9vz's write pins (the write arm is untouched).
