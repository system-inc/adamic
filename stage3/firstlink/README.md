# First link over tsc

A recorded stop is a measurement lead. It is never a pass or evidence that tsc lowered. Reproduce a lead as an ordinary program before filing it. This harness asks the first-link door to measure the program rooted at `src/tsc/tsc.ts` in the pinned, adapted TypeScript 6.0.3 tree.

## Sealed run

Correctness asks for this run on the door at `ea68172c1592d6ed33e4b19db080d9dbff8b129c`, which records whole-program stops.

```sh
ahra loom job 'source /opt/adamic-block/environment.sh && bash stage3/firstlink/run.sh' --kind proof --size F64 --ceiling 120m --task nm7z8fv --title "first link over tsc in trap mode: row 1" --at adamic@<sha>
```

The runner prepares the pinned tree and builds the door in a fresh directory beneath `LOOM_JOB_SCRATCH`, falling back to `TMPDIR`. It invokes exactly:

```sh
"$work/adamic-firstlink" "$work/tree/src/tsc/tsc.ts" "$work/out"
```

This invocation writes `traps.json`, `traps.md`, `c/main.c` and `js/main.js` without invoking clang or linking. The runner gives the door 60 minutes, including checking and emission. Each phase reports wall time and peak child RSS. `VERDICT PASS` means the door completed, the JSON was read and its SHA256 was printed. The output directory is under the printed `WORK` path.

## Row 1 of #bcbp175

`report.py <out-dir>` reads `traps.json`. It writes `firstlink.json` with the input SHA256, the original fields and the phase summaries. It writes `firstlink.md` with counts by kind and normalized reason and a row for every recorded stop. Each row retains the door's reason and site and shows its meter family. The door makes sites relative to the program's common directory. It keeps an absolute site if the program shares only the filesystem root. The summary counts absolute and empty sites.

The body phases come from `ir.go:153` to `:157` at `ea68172c`. The whole-program phases come from the 33 `firstLinkPass` calls and two `firstLinkProgramStop` calls at that head. A phase outside both sets fails the report. If there are no whole-program recorded stops, the table says so in its header.

Echo and suspect flags are tallied separately and stay in their phase totals. The reader checks the door's `echo` count against the records and lists `rewrites_skipped` in door order. A `Gap` marks a measurement gap.

## Fixtures

`testdata/body.json`, `testdata/body-and-statement.json` and `testdata/setter.json` are byte-for-byte reports from the pinned door on small programs. `testdata/wave2.json` was written by hand to test flags, gaps, skipped rewrites, site fallbacks and additional fields. These fixtures test the report reader.
