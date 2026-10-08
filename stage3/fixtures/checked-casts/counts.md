# Checked-cast fixture counts

Measured on main 45487a809f89885a3fc651cd590e7dabf31362dc.

| Fixture | Allocations | Frees | Retains | Releases | Peak | Regions |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 02_name_subunion.a | 2 | 2 | 3 | 3 | 2 | 0 |

The failing twin exits 70 at its tag check. Counts after normal cleanup are not
claimed for it. The other 18 programs are refused before native emission, so
there are no allocation counts for them.

These are external stage3 fixtures exercised by observe.cjs, rather than new
entries in internal/oracle's compiled fixture registry. Its counts table is
unchanged. The command used here is:

```sh
adamic build stage3/fixtures/checked-casts/02_name_subunion.a -o /tmp/name-counted --count
/tmp/name-counted
```
