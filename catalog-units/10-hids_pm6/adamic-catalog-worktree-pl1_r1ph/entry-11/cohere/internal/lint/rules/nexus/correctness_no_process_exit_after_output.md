# `nexus/correctness-no-process-exit-after-output`

| | |
|---|---|
| **Recommendation** | **Registered `off` in ahra, by @system_cohere's ruling of 2026-10-03.** Every finding is true about the code, but ahra blocks stdout and stderr at every process entry (nexus `blockStandardStreams`), and a blocking stream has nothing queued when `process.exit` runs, so the 1,012 findings hold 0 real truncations there. The exactness moved to the entry rule, which requires a process that can exit after a write to block first. On for a project whose entries do not block |
| Findings | **ahra 1,012** (983 in ahra's own files, 29 in the `libraries/structure` submodule), measured 2026-10-03 on a tree being edited live |
| Measured precision | 1,012 of 1,012 true. 1,002 are reached having written on every path into the exit; the other 10 were read one by one and each prints on the path that exits |
| Plugin | cohere-native, `nexus` |
| Auto-fixable | no |
| Needs type information | yes, to resolve `console`, `process.exit` and `process.stdout` / `stderr` by symbol; without a checker the rule registers nothing |

## What it checks

A call of `exit` on `NodeJS.Process` (as `process.exit`, through `import * as NodeProcess from
'node:process'`, or an imported `exit`) reached, on some path through the enclosing function's control-flow
graph, after a write to stdout or stderr in the same function: `console.log`, `info`, `debug`, `warn`,
`error`, `trace`, `table`, `dir`, `dirxml` on the global console, or `write` on the process's own
`stdout` / `stderr`, or a call of a function whose own body makes one of those writes (one level down,
see below). The fix is `process.exitCode = n` and a `return`, which lets the process end once the
output has drained.

## Why

`process.exit()` ends the process without waiting for writes still in flight, and a write to a pipe can be
one. Measured on macOS with Node 24: `console.log` of 4 MB then `process.exit(0)` delivered exactly 65,536
bytes through a pipe, exit status 0, no error; to a file it was complete. Node's documentation for
`process.exit` names this exact misuse. ahra is read through pipes (`ahra ... | grep`, `execSync` captures
between modules), so a long report can arrive cut with nothing saying so.

## What it declines, and why each is a missed finding rather than a false one

- **Writes more than one callee down, or in a callback.** Since #4pyyvfs a call counts as a write when
  its callee's own body writes: `showHelp(); process.exit(0);` reports. The callee is read only where it
  can be and only where its body surely runs before the call returns: in this file or in a module (so the
  findings cache, keyed on the import closure, sees its edits; a dynamic `import()` counts), with a body
  (an overload's resolved declaration has none), not a generator, awaited directly when async, and neither
  declared `never` nor exiting itself (then the caller's exit may never run). A helper that prints through
  another helper, and `rows.forEach((row) => console.log(row)); process.exit(0);`, are still not followed.
- **A write inside a `try`, the exit in its `catch`.** The graph runs the end of every `try` block into
  its `catch` (ESLint's shape), so a write there could be "followed" into the catch on a path where the
  write was the last thing that ran, or the thing that threw. A path entering a `catch` therefore forgets
  writes made in that statement's `try` block. A probe counting what this cut costs found 0 on ahra.
- **Exits reached through an alias** (`const exit = process.exit`) and **writes through a stored
  logger** (`const log = console.log`) are not resolved.
- **The shape is the condition.** Whether bytes are still buffered at the exit depends on how much was
  written and what ran between (an `await` can let a pipe drain). The source cannot say, so a write then
  an exit on one path is the whole condition, as the task's precision line asks.

## Reading one callee: what it added (#4pyyvfs)

Measured 2026-10-04 on ahra with the rule forced on, cohere at main against the change, both cold: 1,017
findings before, 1,029 after, 12 added and none lost. All 12 read in place, each true by the rule's
condition, a helper that writes on a path into the exit:

| site | the helper that writes |
|---|---|
| `libraries/structure/command-line/Structure.ts` 519, 875, 1238; `StructureCohere.ts` 145 | `reportUnresolvableCohereBinary()`, then `exit(1)` |
| `libraries/structure/command-line/Structure.ts` 1627 | `runDoctorScript(...)`, which prints the doctor's output |
| `modules/ahra/AhraCommandLineInterface.ts` 187 | `showHelp()`, then `exit(0)`, the first named hazard |
| `modules/backup/BackupCommandLineInterface.ts` 58, 164 | `reportFreshness` and `reportTreeCoverage` through `import()`, the second |
| `modules/os/lifecycle/AhraOsBootCommandLineInterface.ts` 263 | `printStaleAwokeSweepResult(...)` earlier in boot |
| `modules/phi/PhiCommerceCommandLineInterface.ts` 388, 529 | an awaited `resolveSubscription(...)`, which prints why it found none |
| `modules/porings/PoringsCommandLineInterface.ts` 179 | `cleanup()` writing to stdout in a `SIGINT` handler |

Cost: the scan finds exits first and resolves a call's callee only inside a function that holds one,
since resolving every call in every file took the rule from 2.5s to 4.2s of the run. As built, `--timing`
puts the rule at 157ms of summed worker time against 62ms before, 0.6% of rule time.

## Reconciling with the research count

Research counted **191 of 996** `process.exit` calls following a **stdout** write in the same function,
by **text order** (`probe8.js`). cohere counts **983** in ahra's own files, by control flow, over **stdout
and stderr**. The two numbers measure different conditions:

- **stderr is in.** The task names `console.error` / `warn` and `process.stderr.write` as writes, and
  ahra's dominant shape is `console.error('...'); process.exit(1);` (923 of the 996 exits are
  `exit(1)`). A pipe truncates stderr the same way.
- **With stderr taken out, control flow counts 78, not 191** (a throwaway build restricted to stdout
  writes, 23 of them `exit(0)` against research's 24). Text order counts writes that precede an exit
  without reaching it: a usage block that prints and `return`s before a later exit, a write in a callback
  above the exit, a write on a branch that has already exited. The 113 between the two counts were not
  read one by one; every finding of the rule was.
- Of the 1,009 exit calls in ahra's own files, 26 went unreported on 2026-10-03: exits in event callbacks
  with no write of their own (`process.on('SIGINT', ...)`, `child.on('close', ...)`), exits after a
  callee prints (`showHelp()`, `reportFreshness(...)`, `runLinkCommand(...)`), and `if(!subscription)
  process.exit(1)` after a helper that printed the reason. All read; none is a write in the same function
  that the rule missed. Reading one callee since reaches `showHelp()`, `reportFreshness(...)` and the
  subscription helper (see above); `runLinkCommand(...)`'s exit is still unreported.

## The findings, read

A throwaway build split every finding by whether **every** path into the exit has written (a must
analysis over the same graph, which no infeasible-path combination can make false, since every real path
is a graph path) or only **some** path has.

- **1,002 every-path findings.** True by construction: whatever branch the program took to reach the exit,
  it printed first. The bulk is `console.error(...)` then `process.exit(1)` in argument validation,
  then report blocks ending in `process.exit(0)`.
- **10 some-path findings, read in context**, all true: `FinanceConnectionsCommandLineInterface.ts`
  `runSync`, `runQuickBooksEnrichCommand`, `runQuickBooksLedgerImport`, `runStatement`, `runMonths` (a loop
  over report lines that prints nothing only when there are none);
  `libraries/structure/command-line/Structure.ts` dead-code and duplicates commands (`if(output.trim())
  console.error(output); process.exit(1);`, which prints the analyzer's whole output, the worst case for
  truncation); `StructureLintEngineParity.ts` (loop over report lines); `ClaudeCommandLineInterface.ts`
  usage (prints the table unless `--json`, then exits 1 when every account failed);
  `ReplicateCommandLineInterface.ts` (prints the failed prediction unless it went through a helper).

`sites.tsv` in the build scratch directory lists every finding with its fix and what the fixer must know:
59 exits are the last statement and need only `process.exitCode = n`; 41 sit in functions that return a
value, so a bare `return` will not type-check; 15 sit in `never`-typed helpers
(`exitProcessWithError` in Nexus's `Process.ts`, `fail` helpers in `AhraOsInbox.ts`,
`FinanceTaxCommandLineInterface.ts`, `AhraOsMindsCommandLineInterface.ts`,
`TasksCommandLineInterfaceShared.ts`) whose callers rely on them not returning; 13 are at module top level,
where no `return` is possible; 17 are in callbacks (`main().catch(...)`, `.on('close', ...)`), where a
`return` leaves only the callback.

## Verification

- Fixtures from the real sites, both ways: `runStatement`, `runSync` and `runQuickBooksCardBalances` from
  `FinanceConnectionsCommandLineInterface.ts`, and `godwordRequireInteractiveTerminal` from
  `GodwordEntropy.ts` (through the `NodeProcess` namespace import), each firing as it stands and silent
  with `process.exitCode` set. 17 more firing shapes and 17 silent ones, against a slice of
  `@types/node` 26.2.0 in its real layout (`declare module "node:process" { global { ... } export =
  process }`, the `web-globals` console in `declare global`) plus lib.dom's console.
- Mutation check, each mutant a copy through `go test -overlay -count=1`, each killed but one: the
  written-state test at the exit (the exit before any write, and four more); an exit ending its path (a
  write laid out after an earlier exit); the catch cut (a write inside the try, the exit in its catch);
  the global console check (a local console); the `NodeJS.Process` member check for `exit` (another
  object's exit, a local function named exit) and for `stdout` (another stream's stdout); the
  exit-inside-a-write's-arguments check; alias resolution (members imported by name); keeping an earlier
  chain when adding a later one (a write in an earlier try survives the catch of a later one); the
  catch-binding skip (an exit in the catch binding); narrowing the console methods (every console method
  that writes). **Survived**: removing the unrecorded-exit guard. A probe reporting whenever it fires
  counted 0 on ahra, and the only exit it skips (one in a catch binding's default) cannot make a later
  finding false; the doc comment says why it stays.
