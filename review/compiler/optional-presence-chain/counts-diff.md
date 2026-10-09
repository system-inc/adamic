# Counts changes

Selective TestCountsAreRecorded updates were merged into the union table. Unselected rows retain the chain values. Additional counters run by the existing harness were measured too.

Existing rows moved: 0.

New rows:

| internal/oracle/testdata/optional_field_write.a | 2 | 2 | 0 | 2 | 1 | 0 |
| internal/oracle/testdata/optional_field_presence.a | 45 | 45 | 41 | 60 | 11 | 0 |
| internal/oracle/testdata/optional_field_construction.a | 85 | 85 | 60 | 131 | 25 | 0 |
| internal/oracle/testdata/optional_field_unknown.a | 22 | 22 | 15 | 44 | 7 | 0 |
| internal/oracle/testdata/optional_field_alias.a | 10 | 10 | 9 | 20 | 6 | 0 |
| internal/oracle/testdata/optional_field_alias_variants.a | 69 | 69 | 32 | 99 | 16 | 0 |
| internal/oracle/testdata/optional_field_checked_copy.a | 6 | 2 | 6 | 9 | 4 | 0 |
| stage3/fixtures/taste/17_binder_flow.a | 51 | 51 | 0 | 51 | 7 | 0 |

The presence and construction fixtures each add one retain/release pair relative to the old source branch The retained chain ownership rules are the inferred cause; that cause was not isolated in a counterfactual build.
