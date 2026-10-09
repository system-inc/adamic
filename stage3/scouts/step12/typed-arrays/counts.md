# Step 12 witness counts

Measured with main plus 04f18a2a, using `adamic build --count`.
These scouts are outside the shared oracle fixture registry; its counts table is unchanged.
Blocked witnesses have no native counts. Every finished counted run matches Node stdout.

| Witness | Allocations | Frees | Retains | Releases | Peak | Regions |
|---|---:|---:|---:|---:|---:|---:|
| 04_length | 7 | 7 | 1 | 9 | 3 | 0 |
| 05_read | 3 | 3 | 1 | 5 | 3 | 0 |
| 06_write | 5 | 5 | 2 | 8 | 3 | 0 |
| 10_subarray_control | 10 | 10 | 5 | 16 | 4 | 0 |
| 11_constructor_boolean_control | 33 | 33 | 0 | 35 | 3 | 0 |
| 12_dense_or_control | 5 | 5 | 2 | 8 | 3 | 0 |
