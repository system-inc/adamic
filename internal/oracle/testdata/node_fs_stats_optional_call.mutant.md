# Stats optional-call missing-guard witness

The control is `TestInputAgreesWithNode/internal/oracle/testdata/node_fs_stats_optional_call.a`.
It agrees byte for byte with Node on native and JavaScript, including the file,
directory, seven undefined results, and `Stats | Dirent` rows. The receiver count
is 24: seven printed calls and one discarded call on each of three paths.

For the mutant, replace the shared `optionalIntrinsic` conditional's nullish
condition with `ir.BooleanConstant{Value: true}`, preserving its saved receiver.
Run:

```
go test ./internal/oracle -run '^TestInputAgreesWithNode$/internal/oracle/testdata/node_fs_stats_optional_call.a$' -count=1 -v
```

Observed red: the first missing receiver is the `isFile` expression at source
line 9, stdout line 15. Node prints `undefined` and exits 0. Native stops after
14 stdout lines, exits 1, and reports `node_fs_file.c:533:86: runtime error:
member access within null pointer`. JavaScript also stops after 14 stdout
lines, exits 70, and reports `TypeError: Cannot read properties of undefined
(reading 'isFile')`. The control passes after restoring the guard.

`TestNodeFSStatsOptionalCallNullishMutant` permanently removes all eight guards
in the lowered fixture, retains their receiver declarations, and requires both
native and JavaScript to disagree with Node. No mutant is retained in production.
Counts were generated with `TestCountsAreRecorded` and `-update-counts`.
