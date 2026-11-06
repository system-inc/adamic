# Twice-daily meter

The daily command, after the toolchain is installed, is:

```sh
source /workspace/adamic-tools/env.sh && STAGE3_METER_COMPILER=per-ref STAGE3_METER_LATENT=first-error timeout --kill-after=60s 7200 bash /workspace/adamic/stage3/meter/twice-daily.sh > /tmp/stage3-meter-daily.log 2>&1
```

Every invocation fetches only `origin/main`, then resolves one immutable commit.
`STAGE3_METER_MAIN_REF` can select an explicit commit; its default is the fetched
`origin/main`. Only main is measured. The historical `area` reporter label reuses that
measurement: the complete `main/` directory (raw census, latent and timing
streams, entry streams, and preparation/measurement logs) is copied to `area/`
before reporting and gzip. Both reporter arguments name `main-adapted`, preserving
the absolute paths in the copied records. `compiler-mode.json.measurement_sources`
and the run README explicitly record this reuse.
`STAGE3_METER_AREA_REF`, if supplied, must resolve to the same SHA.
Unresolvable or unequal pins fail before creating an output directory.
Cold runs clone the pinned upstream TypeScript tag when
`ADAMIC_TYPESCRIPT_SOURCE` is unset and install each snapshot's API lockfile with
`npm ci`. Apply validates the upstream SHA and package digest. All preparation
runs unattended and its logs remain beside the measurement.

The default compiler mode remains `single`, using the launcher checkout's
compiler. The daily command selects `per-ref` so the compiler comes from the
measured SHA. The main snapshot uses its own apply script and adaptations.

Set `STAGE3_METER_COMPILER=per-ref` to build both the ordinary census and latent
census once from main's pinned compiler. Detached scratch worktrees
initialize each ref's recorded submodules, and the worker supplies the guarded latent
overlay instrumentation applied to each ref's compiler sources. Build, checkout and submodule logs live under `main/` and `area/`.
Scratch worktrees are retained for inspection, including on failure.
`compiler-mode.json` records the selected mode and both compiler commits before
building. Paired JSON records `compiler_mode` and each tree's `adamic_commit`;
the root `adamic_commit` remains the area's compiler for compatibility. Markdown
names both compilers in per-ref mode. Its first two lines keep the same format.
Across daily runs, per-ref mode changes both adaptations and compiler versions,
so a score change cannot be attributed to adaptations alone. Unknown modes fail before
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

Each run records `box.json` before any build: hostname, available processor count,
CPU quota and period in microseconds (null when unlimited or unavailable), total
memory, and the starting load average. The ending load average is added after
measurement, before reporting. `report.json` preserves this as `box` and Markdown
prints the box below the run header.

Each latent pass writes its own `latent-times.jsonl`, compressed after reporting:
`main/tsc/latent-times.jsonl` for the entry pass and `main/latent-times.jsonl`
for the compiler-directory pass. Each completed unit adds one record directly
to the file, with no user-space buffer. Read that pass's times file for progress
while it runs. A slow unit can leave the file unchanged until the unit finishes.
The entry timing sidecar is ignored by `report.py`, preserving today's report;
the compiler timing sidecar still supplies its existing timing tables.

Each timing record contains `{file, where, wall_seconds, cpu_seconds}` for an
attempted unit, including recovered panics. Deferred and checker-skipped units
have no time record. Wall time uses Go's monotonic clock; CPU is the process's
user plus system time from `getrusage(RUSAGE_SELF)`. The driver is sequential,
so each unit includes synchronous dependency work and registration within its
attempt. Shared file setup and writing the timing record are excluded. Both
first-error and full mode record times. Only main runs; area copies its times
as well as its findings.

`latent.jsonl` keeps its initial metadata line during a pass and writes its file
records only when the pass returns. This buffering preserves final record bytes:
a concrete generic call in a later file can set an earlier deferred unit's
`instantiated` field. Preserving traversal order alone cannot make immediate
file records identical. See `stage3/census/latent/audit_deferred.py` for the
cross-file witness fixture. In speculative mode, boundary inventory and depth
also depend on later files. Lack of growth in `latent.jsonl` does not identify
a stuck unit; inspect the matching timing sidecar.

Each tree's `latent_time` JSON block contains wall/CPU totals, `files` summed and
sorted by descending wall time, and `slowest_units` (at most ten). Source locations
start at `src/`, like panic sites. Markdown shows totals and the five slowest
files, explicitly labeled as this box's times, not comparable across boxes.
Older runs use `latent_time: null` and say "no unit times recorded".

`python3 -E -S -B stage3/meter/check-unit-time-mutant.py` drops the time block in
a scratch reporter and requires
`ReportTests.test_meter_unit_times_block_sorts_files_and_shortens_units` to fail
with `KeyError: 'latent_time'`.

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
0 8 * * * . /workspace/adamic-tools/env.sh && STAGE3_METER_COMPILER=per-ref STAGE3_METER_LATENT=first-error timeout --kill-after=60s 7200 bash /workspace/adamic/stage3/meter/twice-daily.sh >> /tmp/stage3-meter-cron.log 2>&1
```

This unit does not install cron or send messages to @system_adamic.

The single-measurement regression counts four census invocations, checks copied
streams byte for byte after decompression, and checks the shared adapted tree and
measurement provenance. Restore a real second measurement and require that check
to go red with:

```sh
python3 stage3/meter/check-measure-once-mutant.py /tmp/measure-once-mutant-test.log > /tmp/measure-once-mutant.log 2>&1
```

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
per-ref mode must produce main's compiler witness for both labels and record
the same pinned SHA. A missing local dependency transport retries the recorded
origins without changing either gitlink pin. Point the test-only `METER_SCRIPT_UNDER_TEST` at a
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

Run meter and census test cases as independent gate units after preparing the
four probe executables, four Go test binaries and the shared fixture oracle hook once. Preparation logs all builds; fetches validate the action
key and every binary's SHA-256 and refuse missing or corrupt products.

```sh
python3 stage3/meter/build-probes.py --prepare > /tmp/probe-build.log 2>&1
ADAMIC_TEST_SHARD=0/8 python3 stage3/meter/shards.py meter > /tmp/meter-0.log 2>&1
ADAMIC_TEST_SHARD=0/2 python3 stage3/meter/shards.py census > /tmp/census-0.log 2>&1
ADAMIC_TEST_SHARD=0/8 python3 stage3/meter/shards.py meter --list > /tmp/meter-cases.json
python3 stage3/meter/audit-shards.py /tmp/new-probe-audit > /tmp/probe-audit.log 2>&1
```

Repeat meter with indices 0 through 7 and census with indices 0 and 1.
Assignment is SHA-256 of each complete unittest case ID, first eight bytes
big-endian, modulo the selected shard count.
The audit checks the complete disjoint union against the case inventory and
plants an additional failure in exactly one case's shard for each group. Every
original assertion still runs. In particular it retains meter's Refused-count
failure and census's nested-unit panic on the measured main revision. Meter's
real optional probes are enabled by the runner, rather than silently skipped.
Census replay's ordinary unittest invocation also requires prepared products.
The local content store can be copied to another worker at the same repository
path with the same keyed inputs; this helper does not configure remote storage.
See [the full timing report](../fixtures/SHARD-REPORT.md) for measured failures
and limits. The audit is an orchestrator, not one gate test unit.

Each run's README compares main's whole-program, own-file and ordinary lowering
counts and tsc entry diagnostic count with the latest earlier paired report
committed in the measured main revision. Both measured SHAs are printed. Uncommitted scratch runs cannot
become the previous baseline. The Oct 19 report fixture retains the relevant
fields from commit 11a32835 and is tested beside the committed Oct 8 iBLy18 run.

A per-root stop is a root-order artifact only for the two exact runtime-readiness
reasons, when the entry census attempted units in the stop's file without any
readiness finding there, and that root throws on Node. The meter runs
`node_readiness.py` itself: TypeScript 6.0.3 emits ES modules with
`verbatimModuleSyntax: true`, preserving value imports used only as types and
const enums, then the installed setup-pinned Node v24 runs
each qualifying root with a 5-second limit (emit has a separate 60-second limit).
JSON retains exit, first error line, version and timeout; build failures,
timeouts and clean runs keep counting against the root file. All three measured
roots (`debug.ts`, `diagnosticInformationMap.generated.ts`, `types.ts`) throw
with the loader's module graph and qualify when the entry condition also holds.
`LOADER_OPTIONS` beside the runner mirrors `compilerOptions()` in
`internal/load/load.go`: strict, unchecked-index and exact-optional checks;
non-erasable syntax; verbatim imports; TypeScript extensions; ESNext modules,
forced module detection, bundler resolution, ES2024 target/lib and empty types.
Only emit settings are overridden: noEmit is disabled, noCheck avoids another
checker pass, TS extensions are rewritten, and output/source-map paths are set.
Each probe retains the effective settings in `compiler-options.json`.
`root_order_artifacts` lists the excluded readiness stops; `readiness_node_checks`
also retains clean and inconclusive probes. Every artifact's root, stop, reason
and Node result appears beside lowering counts in single, paired and daily
README tables, without truncation. This ruling changes the file's meter score,
not its raw outcome or latent finding-site totals. An artifact root counts as
lowering whole only when every implementation source in its program has
attempted entry-census units and no finding of any kind in any program file.
The Node runner asks tsc `--listFilesOnly` separately for each root using the
same options as emit, excludes declarations, and records `program_files` and
`program_files_complete`. An unavailable reach never proves whole lowering.
Otherwise `lowering_stop` records the first entry finding in the program
(kind, reason, location), labeled `entry census`. Compiler source lists do not
establish ES-module evaluation order, so the report explicitly uses source
order: normalized path, numeric line and column, then entry-stream order for
ties. The table shows that next stop behind the artifact. Unattempted program
files are named in `entry_unattempted_sources` and prevent a pass.
No successful IR after the per-root first stop is claimed. Probe logs remain in each run's `node-readiness/`.

For a partial historical ordinary census, `python3 -E -S -B
stage3/meter/report.py --root-order-only <adapted-tree> <run/main>` writes and
prints the artifact table and Node checks without claiming complete coverage.
Paths in both streams must name that adapted tree. A complete ordinary census
uses the normal single-tree report command. The two direction mutants run with
`python3 -E -S -B stage3/meter/check-root-order-mutants.py`; each names its red
`RootOrderTests` regression. A third mutant disables verbatim imports and must
fail `NodeReadinessTests.test_value_import_used_only_as_type_keeps_its_evaluation_edge`.
The Node runner has its own real-tool tests in
`node_readiness_test.py`. Go wrappers in `meter_test.go` run all meter Python
suites; new tests keep the existing `*_test.py` naming and the stage3 Python
suite census stays unchanged.

The `artifact-clean-own-file-implies-whole` mutant must go red on
`RootOrderTests.test_artifact_with_clean_own_file_and_refused_dependency_does_not_lower_whole`.
A refusal in a dependency prevents whole-program lowering even when the root
file itself is clean. A separate test pins missing attempted dependency units.
