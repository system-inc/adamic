# Record storage nesting (#fzx0j5z)

One edit, run on a house box, then put back.

## The enter checks forced ok

In `sameRecordStorage` (records.go), force both `typeNesting.enter` answers true, keeping the
calls and their deferred leaves:

| Old text | New text |
|---|---|
| `if !fromOK \|\| !toOK {` | `if fromOK, toOK = true, true; !fromOK \|\| !toOK {` |

Run:

```
go test ./internal/lower -run 'TestRecordStorageNesting' -count=1 -v -timeout 60s
```

The red is Go's test timeout: `panic: test timed out after 1m0s`, with
`TestRecordStorageNestingStopsAtTheView` and `TestRecordStorageNestingBoundsTheWalk` still running,
since the walk over Own<number> seen as Deep<number> visits a fresh pair at every level again and
never returns, as on main before the bound. `TestRecordStorageNestingControlLowers` never reaches
the walk (one type seen as itself) and passes before the timeout. The unmutated pins return in
seconds; the job script `fzx-job.sh` runs both.
