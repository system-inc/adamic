# ff5eddz metadata spread evidence

The first attempt stopped when JSON.stringify at tracing.ts:104:46 became
visible after removal of the three spreads. The user explicitly accepted
only that next stop, already represented by the same reason at 194:67.
The first-stop report and original comparison failure remain historical.
Current census comparison checks every tracing.ts finding and permits only
that one addition. Current-main checks, the final apply, exact two-state pin
controls, and the complete corrected sharded lane are recorded separately.

The first lane copy incorrectly excluded nested node_modules fixtures.
Its run was stopped with child exit -15 and archived separately. The corrected
copy retains all 70 nested node_modules inventory entries; all 85,763 entries
match the audited tree except the three intentionally instrumented bundles.
The complete final lane restarts every shard from scratch.

See current-census-comparison.json, final-source-verification.json,
counting-copy-verification.json, final-lane-control.json and lane-summary.json.
All raw Go measurement sources remain outside the adaptation directory; no
production compiler or lane implementation was modified.

The final lane passes all 398 shards with 0 tracing starts and a positive
control observing cat,ph,ts,pid,tid. The initial reconciliation helper assumed
a top-level counts field that the version-1 manifest does not contain. It was
corrected to sum each unit expectation and independently compare the unchanged
platform counts; final reconciliation passes. No lane expectation was altered.
