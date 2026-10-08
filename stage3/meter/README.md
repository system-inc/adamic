# Twice-daily meter

Run from any directory after sourcing the toolchain env.sh:

```sh
stage3/meter/twice-daily.sh > /tmp/stage3-meter.log 2>&1
```

Every invocation fetches and pins `origin/main` and `origin/area/stage3`.
It snapshots each ref's stage3 directory and runs that ref's own apply.sh and
adaptations into a fresh tree. A single census binary built from the current
checkout measures both trees with ordinary stage 0 options and prelude. This is
the default (`STAGE3_METER_COMPILER=single`); fetching a newer main does not change
the compiler built from the current checkout.

Set `STAGE3_METER_COMPILER=per-ref` to build both the ordinary census and latent
census separately from each pinned ref's own compiler. Detached scratch worktrees
initialize each ref's recorded submodules, and the worker supplies the guarded latent
overlay instrumentation applied to each ref's compiler sources. Build, checkout and submodule logs live under `main/` and `area/`.
Scratch worktrees are retained for inspection, including on failure.
`compiler-mode.json` records the selected mode and both compiler commits before
building. Paired JSON records `compiler_mode` and each tree's `adamic_commit`;
the root `adamic_commit` remains the area's compiler for compatibility. Markdown
names both compilers in per-ref mode. Its first two lines keep the same format.
Per-ref mode compares both source adaptations and compiler versions, so its
changes cannot be attributed to adaptations alone. Unknown modes fail before
creating a run directory.

```sh
STAGE3_METER_COMPILER=per-ref bash stage3/meter/twice-daily.sh > /tmp/stage3-meter-per-ref.log 2>&1
```

The compiler SHA and both adaptation SHAs are recorded. In default mode the
compiler version stays fixed when comparing the two sets of adaptations.

The first two lines of report.md give the four source-file pass counts:
whole program for main and area, then own file for main and area. The table has
both columns for every file in both trees. Whole program retains today's
meaning: the program loaded from a file has no checker diagnostics anywhere.
Own file uses each diagnostic's primary file location, including diagnostics
observed when another root imports it. Repeated observations are deduplicated;
indented elaborations are not treated as locations. Global and external findings
are counted separately and do not belong to a src/compiler file. Input errors
fail both measures. Non-source files are listed but excluded from denominators.

The next line reports the deadline entry, independently of the compiler-file
counts: `tsc entry: main: pass; area: fail`, followed by
`tsc entry diagnostics: main: 0; area: N`. This is the whole-program stage 0
check rooted only at the adapted `src/tsc/tsc.ts`, including diagnostics anywhere
in its resolved imports and global diagnostics. Its observations never enter
the existing per-file numerators, denominators, or own-file attribution.

For each checker-clean entry, a separate guarded measurement binary enumerates
the checker's resolved implementation source files and applies the existing
lowering census to that reach. It excludes declaration files and unrelated
source files. Its top ten NotYet/Refused reasons include owners and are labeled
**measured on a checker-clean entry-root program**. In legacy latent mode a checker failure blocks this lowering census. Full mode
measures eligible nested bodies even on a rejected entry-root program and labels
that stream **measured on a checker-rejected entry-root program**. The entry
checker result and diagnostic count remain unchanged. Errors and panics remain separate totals; this is an observation
ledger and does not establish successful lowering or native output.

The entry measurement extends the selected compiler's existing scratch overlay;
production compiler sources remain untouched. `LATENT_ASSERT_NO_OUTPUT=1` is
required and ordinary Load and Lower remain disabled. The worker supplies the
entry driver in both compiler modes, while compiler code and dependencies still
come from the selected ref. `trees.main.tsc_entry` and `trees.area.tsc_entry`
record checker status, full diagnostics/count, and the optional `lowering_census`.
Each `main/tsc/` and `area/tsc/` retains census and measurement JSONL (compressed
and ignored by Git), logs, and the resolved reach in the measurement header.
Missing entry roots, missing reachable records, inconsistent diagnostics, or a lowering stream on a checker-rejected entry without an explicit full-mode
header fail loudly.

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
census through its scratch Go overlay. The overlay is built from the selected
compiler checkout and never edits production compiler files. `LATENT_ASSERT_NO_OUTPUT=1`
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
retained separately and excluded from the NotYet/Refused ranking. Full measurement splits diagnosed containers into named nested declaration units,
excludes only directly diagnosed bodies, and recovers at statement boundaries
with state rollback. Unsafe child traversal and signature/prologue failures
remain counted byte-span boundaries. It measures observed blockers, not exhaustive blockers
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

`compiler_test.py` also runs the meter against local Git refs with real Go
compiler witnesses. The default must build the worker checkout for both trees;
per-ref mode must produce each pinned compiler's distinct checker and latent
results and record its SHA. Point the test-only `METER_SCRIPT_UNDER_TEST` at a
scratch copy of twice-daily.sh to exercise a mutant that builds the worker
checkout in place of each ref. The per-ref probe must reject that mutant even
when its metadata claims the correct SHAs.

The imported-file probe starts with two checker-clean files. Changing the dependency's number
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

The entry reach probe runs with `ENTRY_CENSUS_BINARY` and `CENSUS_BINARY` set.
It imports an eligible dependency outside src/compiler, leaves an unrelated
checker-failing file unloaded, and requires the dependency's lowering finding.
Planting an imported type error must change the entry to fail and block lowering.
A scratch overlay mutant that leaves Program.Files restricted to the entry root
must fail this test's reach count (one source instead of two). Compiler-selection
probes exercise the entry command wiring in both default and per-ref modes.


Full latent mode is the default for both compiler-file roots and the tsc entry.
`STAGE3_METER_LATENT=first-error` retains the historical body skips and first-error
behavior for comparison. Unknown modes fail before making a run directory.
`compiler-mode.json` records `latent_mode`; report generation checks it against
both streams, so a stale overlay cannot silently claim full coverage. Compiler
selection (`single` by default, or `per-ref`) remains independent of latent mode.
The first two report lines and their checker arithmetic are unchanged.
Each lowering summary retains `boundaries` and the complete
`excluded_nested_functions` ledger (name, location, diagnostics, and bytes).
Excluded parent/child byte spans can overlap; do not add them to claim unique
unexamined bytes. This is partial observational measurement with explicit gaps.
