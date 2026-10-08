Census control ledger. Five `.a` files form four witness projects; these are
measurement fixtures, outside internal/oracle/testdata. This unit adds no native oracle fixtures; the merged mapper-fix branch brings its
own previously counted oracle fixtures. Counts below deduplicate `(kind, where, reason, text)` per project.

| Control | Files | NotYet | Refused | Total |
|---|---:|---:|---:|---:|
| nested with/with/debugger | 1 | 4 | 4 | 8 |
| async generic signature identity | 1 | 3 | 2 | 5 |
| checker-clean nested signature | 1 | 5 | 0 | 5 |
| imported generic body | 2 | 5 | 2 | 7 |
| Total | 5 | 17 | 8 | 25 |

The imported project has nine raw records and seven distinct sites because its two
NotYet dependency findings are also observed in their own source-file attempt.
