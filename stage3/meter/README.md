# Twice-daily meter

Run from any directory after sourcing the toolchain env.sh:

```sh
stage3/meter/twice-daily.sh > /tmp/stage3-meter.log 2>&1
```

This requires the `stage3-base` and `tsc-census` files to be merged. It runs
apply against a fresh scratch path, builds the census from the current checkout,
and measures every file of the adapted `src/compiler`. Each invocation creates
`runs/<UTC timestamp>.<unique suffix>/`, containing report.json, report.md,
census.jsonl.gz and phase logs. The table is also printed. The adapted tree stays
at the scratch path named in report.json for inspection. `STAGE3_CACHE` controls
apply's upstream cache; `STAGE3_METER_RUNS` can redirect run artifacts.

The JSON has booleans `checker` and `lowering` per file, `lowering_attempted`,
the census outcome, and the two pass totals. Refused and NotYet pass the checker
but fail lowering. A checker failure blocks lowering; it does not establish a
lowering diagnostic. Non-source inputs are listed and fail the extension gate,
but source_files is the denominator for the two totals. Generated sources count.
The whole-program outcome and diagnostic count are retained separately and is not added to file totals.
A missing, duplicate or unknown census row prevents publishing a partial report.

The profile is the census's ordinary unmodified stage 0 options and prelude.
No upstream-config overlay or weaker checker options are used. Every adapter
present in the checkout runs through apply. The in-flight type-import and optional
declaration adapters are not silently fetched or included.

For morning and evening execution, a host can use this cron entry with its own
repository/toolchain paths:

```cron
0 8,20 * * * . /workspace/adamic-tools/env.sh && /workspace/adamic/stage3/meter/twice-daily.sh >> /tmp/stage3-meter-cron.log 2>&1
```

This unit does not install cron or send messages to @system_adamic.

Validate report accounting with:

```sh
python3 -m unittest discover -s stage3/meter -p '*_test.py' > /tmp/stage3-meter-tests.log 2>&1
```
