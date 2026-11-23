# Weak union record read mutants

weakUnionRecordRead (`entries_records.go`) holds a boxed union field read through a weak type to the read's
union, member by member, and stops with exit 70 on a mismatch (#hwjx7gw, family 46, rulings.md:407 applied).
Each planted change ran in a sealed Loom job, J-zyce1qy, against a head where every case of
`TestWeakFieldUnionMembership` and every `TestRecordDeclaredKeyStopsArePinned` pin is green, and each was
restored after.

| Mutant | Changed | Red test and assertion |
| --- | --- | --- |
| The mismatch panic dropped (`_, _ = matches, message`) | `entries_records.go:748` | `TestWeakFieldUnionMembership/mismatch` at `weak_field_union_membership_test.go:101`; `members` and `decoder` stay green |
| The boolean member's test dropped | `entries_records.go:728` | `TestWeakFieldUnionMembership/decoder` at `weak_field_union_membership_test.go:23` (NotYet at `decoder.ts:11:16`); `members` and `mismatch` stay green |
| The whole change undone (main's `entries_records.go`) | `entries_records.go` | `TestRecordDeclaredKeyStopsArePinned/r13p_direct_write.ts` at `record_declared_keys_test.go:155` (its stop back at 6:17), the only pin that moves, and all three `TestWeakFieldUnionMembership` cases at `weak_field_union_membership_test.go:23` |
