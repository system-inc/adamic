# The nesting bounds, tested directly (#ykd89v9)

Each edit runs on a house box or a sealed one, then is put back. Each test calls its walk past the
expansive-generic pre-pass, so each bound keeps a red witness.

## collectViewedElements' enter forced ok

In `collectViewedElements` (view_arrays.go), keep the call and its deferred leave:

| Old text | New text |
|---|---|
| `		if !ok {` (after `ok, leave := nesting.enter(target)`) | `		if ok = true; !ok {` |

Run `go test ./internal/lower -run '^TestCollectViewedElementsBoundsAGrowingGenericDirectly$' -count=1 -timeout 20s`
under `GOMEMLIMIT=2GiB`. The red is Go's own stack overflow, `fatal error: stack overflow` past the
1,000,000,000-byte limit, as #k031ey0 first met it: the walk recurses into a deeper Grow at every level
(J-4hpk2vx).

## The cycle finder's use enter forced ok

In `use` (cycles.go):

| Old text | New text |
|---|---|
| `	if !ok {` (after `ok, leave := f.nesting.enter(proven)`) | `	if ok = true; !ok {` |

Run `go test ./internal/lower -run '^TestCycleFinderUseBoundsAGrowingGenericDirectly$' -count=1 -timeout 20s`
under `GOMEMLIMIT=2GiB`. The red is the walk still running at Go's 20 s timeout (J-4hpk2vx).
