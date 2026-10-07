# Twice-daily meter

Run from any directory after sourcing the toolchain env.sh:

```sh
stage3/meter/twice-daily.sh > /tmp/stage3-meter.log 2>&1
```

Every invocation fetches and pins `origin/main` and `origin/area/stage3`.
It snapshots each ref's stage3 directory and runs that ref's own apply.sh and
adaptations into a fresh tree. A single census binary built from the current
checkout measures both trees with ordinary stage 0 options and prelude.
The compiler SHA and both adaptation SHAs are recorded, so the comparison does
not conflate source adaptation changes with different checker versions.

The first two lines of report.md give the four source-file pass counts:
whole program for main and area, then own file for main and area. The table has
both columns for every file in both trees. Whole program retains today's
meaning: the program loaded from a file has no checker diagnostics anywhere.
Own file uses each diagnostic's primary file location, including diagnostics
observed when another root imports it. Repeated observations are deduplicated;
indented elaborations are not treated as locations. Global and external findings
are counted separately and do not belong to a src/compiler file. Input errors
fail both measures. Non-source files are listed but excluded from denominators.

Each run creates `runs/<UTC timestamp>.<unique suffix>/` with report.json and
report.md, fetch/build logs, and main/area subdirectories containing census.jsonl.gz,
apply/census logs and individual reports. The adapted trees and snapshot sources
stay at the scratch paths for inspection. `STAGE3_CACHE` controls apply's upstream
cache; `STAGE3_METER_RUNS` can redirect artifacts. Generated patch-set.md files
stay in the snapshots and do not modify your checkout.

JSON preserves the existing single-tree fields as the **area** observation,
including `checker`, `lowering`, `lowering_attempted`, outcome and totals.
It adds `checker_whole_program`, `checker_own_file` and `own_file_diagnostics`
per file, both checker totals, unlocated/external diagnostic counts, tree_ref
and tree_commit. `trees.main` and `trees.area` contain the complete observations.
The existing whole_program field remains the separate aggregate all-roots run.
Refused and NotYet pass the checker and fail lowering. A checker failure blocks
lowering. Missing, duplicate or unknown census rows prevent a partial report.

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

To exercise the actual checker with a planted type error in an imported file:

```sh
go build -o /tmp/stage3-census ./stage3/census/tool > /tmp/stage3-meter-build.log 2>&1
CENSUS_BINARY=/tmp/stage3-census python3 -m unittest discover -s stage3/meter -p '*_test.py' > /tmp/stage3-meter-tests.log 2>&1
```

The probe starts with two checker-clean files. Changing the dependency's number
initializer to a string makes both loaded programs fail, while only the
dependency's own-file result changes. Without CENSUS_BINARY this probe is skipped.
