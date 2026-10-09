New required-string stop row:

| internal/oracle/testdata/optional_field_checked_view_required_string.a | 2 | 0 | 5 | 4 | 2 | 0 |

Existing row moved because its fixture now creates a spread copy and writes undefined through it:

| internal/oracle/testdata/optional_field_checked_view_string_undefined.a | 7 | 7 | 10 | 15 | 5 | 0 |
| internal/oracle/testdata/optional_field_checked_view_string_undefined.a | 10 | 10 | 20 | 28 | 5 | 0 |

All other measured optional rows were unchanged; all unmeasured rows preserved.
