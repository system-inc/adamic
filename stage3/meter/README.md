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

After each tree's ordinary checker census, the meter runs the existing latent
census through its scratch Go overlay. The overlay is built from the same compiler
checkout and never edits production compiler files. `LATENT_ASSERT_NO_OUTPUT=1`
checks that ordinary loading cannot expose a rejected program and lowering
cannot return usable IR. No backend is invoked.

After the checker table, report.md shows each tree's latent NotYet and Refused
totals and its top ten reasons, labeled **measured on a checker-rejected program**.
The four checker numbers, their calculations and existing JSON fields remain
unchanged. Each tree's JSON adds `latent_lowering`, with the measurement label,
checker_rejected flag, totals, full per_reason counts and ranked top_reasons.
The root's latent_lowering is the area observation, matching existing JSON fields.

Counts deduplicate `(kind, where, reason, text)` across attempts, matching the
latent census's count definition. SkippedDependency, error and panic totals are
retained separately and excluded from the NotYet/Refused ranking. The measurement
skips function bodies with checker diagnostics and can stop at the first error
inside an attempted unit. It measures observed blockers, not exhaustive blockers
or successful compilation. See ../census/latent/README.md for the tool's limits.

Each main/area directory also retains latent.jsonl.gz and latent.log; overlay and
build logs are at the run root. Missing coverage or invalid measurement labels
prevent publishing a paired report. Raw latent JSONL is excluded from Git.

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

To run the real planted-NotYet attribution probe, set LATENT_CENSUS_BINARY to
the built overlay binary in addition to CENSUS_BINARY when invoking unittest.
It uses the latent tool's overlay-only LATENT_MUTANT_FUNCTION hook to plant one
extra NotYet on a checker-rejected program with existing NotYet and Refused findings. Only that reason's
count may increase; every existing reason and the Refused total must stay fixed.

Reason tables group identical reason text across NotYet and Refused and show an
owner column. owners.json is a flat reason-to-owner map. The explicit "seen as" entry wins whenever that phrase occurs inside a reason.
Other reasons use exact matches, then the longest matching prefix. Enum-read
entries cover the enum declarations in the pinned compiler sources; ordinary
function reads and the Error constructor are not assigned as enums.
Unknown reasons get OWNER BLANK. The Unowned section appears immediately after
the checker table and before owned reason tables. It lists only reasons with at least 10 unique sites on either tree, sorted by
the larger per-tree count. A final line summarizes the number of omitted reasons
and their total sites summed across both trees.
Each row compares both trees; JSON retains the complete unowned_reasons list,
and each tree's reason_rows carries its owner and counts by kind.
