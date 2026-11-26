# First link over tsc

A recorded stop is a measurement lead. It is never a pass or evidence that tsc lowered. Reproduce a lead as an ordinary program before filing it. This harness asks the first-link door to measure the program rooted at `src/tsc/tsc.ts` in the pinned, adapted TypeScript 6.0.3 tree.

## Sealed run

Correctness asks for this run on main's door, `cmd/adamic-firstlink` since `51d5dada497611c4afd7fa382bd4f4aabd3fb770` (CLI's 052a8142), which records whole-program stops.

```sh
ahra loom job 'source /opt/adamic-block/environment.sh && bash stage3/firstlink/run.sh' --kind proof --size F64 --ceiling 120m --task nm7z8fv --title "first link over tsc in trap mode: row 1" --at adamic@<sha>
```

The runner prepares the pinned tree and builds the door in a fresh directory beneath `LOOM_JOB_SCRATCH`, falling back to `TMPDIR`. It invokes exactly:

```sh
"$work/adamic-firstlink" "$work/tree/src/tsc/tsc.ts" "$work/out"
```

This invocation writes `traps.json`, `traps.md`, `c/main.c` and `js/main.js` without invoking clang or linking. The runner gives the door 60 minutes, including checking and emission. Each phase reports wall time and peak child RSS. On a door failure the runner prints the first `panic:`, `fatal error:` or `runtime error` line and the 20 after it, then the log's last 60 lines. Once `traps.json` is read, the runner prints `REPORT-SHA256`, then `firstlink.md`, `firstlink.json` and the first 200,000 bytes of `traps.md`, since a job's files aren't kept. `VERDICT PASS` means the door's report is complete, the JSON was read and its SHA256 was printed. A complete report is the door's exit 0, or its exit 1 before emit with `traps.json` and `traps.md` alone, naming in `judgments_abandoned` the judging pass that gave up partway (#nxb3xh1): that is a reading, not a failure. Any other door failure after `traps.json` still prints the row, then ends `VERDICT FAIL`. The output directory is under the printed `WORK` path.

## Row 1 of #bcbp175

`report.py [--tree-root <tree>] [--door-exit <code>] <out-dir>` reads `traps.json`. The runner passes both. It writes `firstlink.json` with the input SHA256, the original fields and the phase summaries. It writes `firstlink.md` with counts by kind and normalized reason and a row for every recorded stop. Each row retains the door's reason and site and shows its meter family. The door makes sites relative to the program's common directory. It keeps an absolute site if the program shares only the filesystem root. The summary counts absolute and empty sites. A reason can still name a file by its absolute path, so the reader removes the tree root wherever `traps.json` holds it before hashing, and two identical runs give one hash. Such a reason then reads `src/compiler/...`, from the tree, while its site reads `compiler/...`, from the program's common directory. `tree_root_stripped` counts the places. With `--door-exit`, the reader prints a `DOOR` line after the hash, and it exits 3 for a door failure that leaves no complete report, once the row is written.

The body phases come from `ir.go:157` to `:161` at `51d5dada`. The whole-program phases come from the 36 `firstLinkPass` calls and two `firstLinkProgramStop` calls at that head. A phase outside both sets fails the report. If there are no whole-program recorded stops, the table says so in its header.

Echo and suspect flags are tallied separately and stay in their phase totals. The reader checks the door's `echo` count against the records and lists `rewrites_skipped` in door order. A `Gap` marks a measurement gap. `judgments_abandoned`, when the door writes it, is checked as a list of `Pass`, `Site` and `Why` and named in door order.

## Fixtures

`testdata/body.json`, `testdata/body-and-statement.json` and `testdata/setter.json` are byte-for-byte reports from the door at `ea68172c` on small programs. `testdata/wave2.json` was written by hand to test flags, gaps, skipped rewrites, site fallbacks and additional fields. `testdata/judgment-abandoned.json` was written by hand under the key names of Lower's door-2 (24e05b66, not yet on main) to test the report it writes alone. These fixtures test the report reader.
