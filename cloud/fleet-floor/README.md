# Fleet floor

Run `python3 cloud/fleet-floor.py` to poll every five minutes, `--once` for one
pass, or `--dry-run` to inspect a pass without messages, dispatches, Git writes,
or state changes. The coordinator needs Python 3.9 or later, Git, and Kirk’s
Mac `ahra` interface. It never calls Claude or `ahra ai list`.

Settings:

| Environment variable | Default |
| --- | --- |
| `ADAMIC_FLEET_FLOOR_STATE` | `~/.adamic-fleet-floor` |
| `ADAMIC_FLEET_FLOOR_CHECKOUT` | `/Users/kirkouimet/Projects/system/adamic` |
| `ADAMIC_FLEET_FLOOR_CREDIT_FLOOR` | `5000` |

Put one prompt file per unit into `<state>/queues/<lane>/`, ordered by filename.
An optional first line `Label: <label>` selects a preferred Idle or Completed
worker in the first fleet. The entire file is sent as the prompt. Prefix fleet
selectors are passed literally to `ahra ai fleet <selector> --json --live`;
results are filtered by fleet prefix and Codex provider, and deduplicated by
session id. Unavailable live status or usage blocks dispatch for the affected
lane or account. Both credit limits use strictly less-than comparisons.

A file `<state>/off` stops dispatch, while polling, notifications, and records
continue. Each pass dispatches at most three units per lane. A successful send
or start consumes one unit toward the floor during that pass; the next pass
counts live status again. Queue warnings re-arm above half the lane floor;
empty warnings re-arm above zero. Notification flags persist across restarts.

Briefs move intact to `<state>/used/<lane>/` before the provider invocation.
They are never deleted or overwritten. A failed or interrupted invocation is
recorded as uncertain and keeps the brief there: inspect the session before
queuing a replacement, since an error may follow a successful provider action.
An inflight journal survives coordinator restarts. A kernel lock prevents two
coordinators using the same state directory from dispatching concurrently.

Rows first persist under `<state>/rows/`. A detached worktree at
`<state>/records` appends the CSV to `records/fleet-floor` and pushes without
force. Pending rows and the attempted commit survive rejected pushes; retries
check remote ancestry to avoid duplicating an already-pushed batch. Remote
updates are preserved when rebuilding a rejected batch. Git failure never
removes the pending rows. `documentation/velocity/fleet-floor.csv` lives on
that records branch, not the implementation branch.

The included launchd plist runs from `adamic-gate-watch` and logs to
`adamic-gate-logs/fleet-floor.log`. Create the log directory and ensure Python
and ahra are on the configured PATH before installing it yourself. It is not
installed by this change.

Run `python3 cloud/fleet-floor-test.py` for stubbed ahra tests and local bare-Git
records recovery tests. These contact no provider and use no network.
