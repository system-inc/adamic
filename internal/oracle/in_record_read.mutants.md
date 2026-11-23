# Mutant for #859ksab's dynamic record read stop

**R1, the stop dropped.** `dynamicRecordReadGap` (internal/lower/dynamic_record_read.go) answers nil, so a named
dynamic read lowers whatever record may reach its receiver.

Must turn red: every subtest of `TestInRecordReadStops` (record_number, record_absent, record_string_of,
unknown_param, generic_param), each lowering where it must stop. record_number, record_string_of,
unknown_param and generic_param carry an ASan fault in their failure; record_absent lowers and agrees.

Must stay green: `TestInRecordReadNeighborsAgree` (record_string, fixed_control, fixed_in_record_program) and
`TestInRecordReadEarlierStops` (from_entries, keyed_read), which stop before this rule.
