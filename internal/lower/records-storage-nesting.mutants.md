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
`TestRecordStorageNestingBoundsTheWalk` still running, since the walk over Own<number> seen as
Deep<number> visits a fresh pair at every level again and never returns, as on main before the bound.
`TestRecordStorageNestingStopsAtTheView` stays green: since #ykd89v9 the expansive-generic pre-pass
stops its programs before the walk runs. `TestRecordStorageNestingControlLowers` walks two
generics that don't grow, which ends by its visited pairs alone, so it passes before the timeout.
The unmutated pins return in seconds.

## The Promise arm dropped

In `sameRecordStorage` (records.go), make the Promise and PromiseLike arm never apply:

| Old text | New text |
|---|---|
| `if l.isLibraryType(from, "Promise", "PromiseLike") && l.isLibraryType(to, "Promise", "PromiseLike") {` | `if false && l.isLibraryType(from, "Promise", "PromiseLike") && l.isLibraryType(to, "Promise", "PromiseLike") {` |

Run:

```
go test ./internal/lower -run 'TestRecordStorageComparesPromisesByWhatTheySettleTo' -count=1 -v -timeout 60s
```

The red is the promise and like rows answering false with bounded set, as on oracle's pins before the
arm (J-vsqgd66).
