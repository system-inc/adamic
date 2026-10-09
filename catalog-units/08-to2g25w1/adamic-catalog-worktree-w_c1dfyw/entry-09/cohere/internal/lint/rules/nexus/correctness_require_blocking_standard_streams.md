# `nexus/correctness-require-blocking-standard-streams`

| | |
|---|---|
| **Recommendation** | **Yes, at `error`.** 1 finding on ahra today, 0 false. It replaces `nexus/correctness-no-process-exit-after-output`, which is registered `off` |
| Findings | **ahra 1** at 11:16 on 2026-10-03, after step 1 landed (`modules/figma/FigmaMcpLauncher.ts:20`). **16** on the tree as it stood before step 1: the 15 unblocked entries of `ExitEntries.tsv` and the Figma launcher |
| Measured precision | 16 of 16 true: each file runs as a process, and on some path its load-time code writes and then exits with nothing blocked before |
| Plugin | cohere-native, `nexus` |
| Auto-fixable | no |
| Needs type information | yes, for `process.exit`, the console and Nexus's `blockStandardStreams` by symbol; reads other files (`ReadsOtherFiles`), so the findings cache never replays it |
| Cost on ahra | 403ms summed across workers, 51ms of it the program index; it listens on the 282 files that spell `exit` |

## What it checks

A file that runs as a process, and can reach `process.exit()` after a write to stdout or stderr, must call
Nexus's `blockStandardStreams()` first. Node writes to a pipe asynchronously on POSIX, and `process.exit()`
does not wait for the queue (measured: 65,536 of 1 MiB arrived, exit status 0). Blocking the two streams at
the entry makes every exit of the process deliver what was written before it, which is why the per-exit rule
is off: the fix belongs at the entry, once.

One finding per entry, at its first qualifying exit, naming how many there are.

## The design, clause by clause

**What is an entry.** A file that starts with `#!` (declared executable; an entry even if other files import
it too), or a file nothing in the program imports (any `import`, `export ... from` not written `type`,
`import x = require()`, or a dynamic `import()` / `require()` of a string literal). A file without `#!` that
something imports is declined whole: its top-level code may run inside another entry's process after that one
blocked, and its own AST cannot say otherwise. No name pattern: `main` is not special. A top-level
`process.on` or `process.argv` read proves nothing on its own and is not used.

**What the process runs.** The module's top level, every function of the same file it calls (by an identifier
resolving to a function declaration or a `const` bound to a function), every immediately invoked function, and
every function handed to a call as an argument (`main().catch(function onFatal(...))`,
`socketServer.on('error', ...)`), transitively. An exported function no load-time code reaches is not counted.

**Can reach `process.exit` after a write.** The per-site rule's own condition, reused from its file (same
package): an exit reached, on some path through its function's control-flow graph, after a write in that
function, with its `try`/`catch` forgetting and its guards. Cross-file reach is not followed. It is not needed
on ahra: every one of the 15 entries has an exit after a write in its own file, reached at load. Following
imports would only add sites inside entries already reported (`Process.ts:17` through
`GraphQlUpdateSchemas.ts`; `StructureUtilities.ts:201` is imported but never called).

**Blocks first.** Decided with a program index built once per run (no checker): which files reach Nexus's
`source/system/StandardStreams.ts` through imports, transitively, together with any file that loads a module
by a computed `import()` / `require()`.

- A file that cannot reach it cannot block. Every exit the process runs counts.
- A file that can is walked in order, and anything that may block ends a path: a call resolving by symbol to
  Nexus's `blockStandardStreams` (its declaration's name and file), or to any function whose body may call
  one (`runCommandLineInterface`, a `main` whose first statement blocks, a helper in another file); an
  unresolvable callee, a `declare`d function, a dynamic `import()`, `require()`; a blocking function handed to
  any call; an `await` (the rest runs on a later turn). Callbacks and generators are taken to start blocked.
  A file it imports that may block while loading makes the whole entry blocked.
- Every one of those can only cost a finding. A call into a file that cannot reach the streams never blocks.

So a usage guard that prints and exits above `runCommandLineInterface(...)`, or a `main` that writes and exits
before its own `blockStandardStreams()`, is reported; the same guard below it is not.

## What it declines

- **A file without `#!` that something imports.** Its load may run after another entry blocked.
- **Exits reached only through another file**, a method, an object's property, or a dispatch table
  (`commands[name]()`). Missed findings.
- **Writes and exits in different functions** (`showHelp(); process.exit(0);`), as the per-site rule.
- **Anything after a possible block in a file that can block.** `StructureLintEngineParity.ts` as it stood had
  three exits; its computed `await import(...)` may load anything, so only the first, before it, counts. The
  entry is still reported.
- **Two things it cannot see, either of which could make a finding false**: `blockStandardStreams` stored and
  called through a variable, property or registry (on ahra both Nexus functions are only ever called directly,
  checked by grep), and a `node --import` preload that blocks before the entry (ahra runs entries under `tsx`).

## Reconciling with the research count

`ExitEntries.tsv` names 15 entries that block nothing and reach an exit site: Facets, the six step-1 scripts
and workers in `modules/`, and eight Structure tools. Measured on a copy of the tree with each of those files
at its version before step 1 (`git show c0269f0e~1`, `c2f95018~1`, Structure `fed59eb3~1`, everything else
from HEAD), the rule reports all 15 and one more, `FigmaMcpLauncher.ts`, which the table lists as a script
with no invoker. The 15 since received the call; on the live tree the count is 1.

Per entry, the exits counted match the table's in-file `Exposed` sites, except `StructureLintEngineParity.ts`
(1 of 3, above) and `LintTranslations.ts:258` (`process.exit(0)` with no write on its path, not a site).

The table's other unblocked rows are correctly silent: `PensieveDailies.ts` / `PensieveWeeklies.ts` (imported
by the pensieve module, and their `isMain` branch is dead), `InstallAhraShims.ts` and `icons/generate.ts` (no
exit after a write), `ClaudeUsageApi.ts` (`#!`, no load-time work), the Next server, `instrumentation.ts` and
the test runner (not files that run their own process's top level, and nothing qualifying at load).

## The findings

| Site | Reading |
|---|---|
| `modules/figma/FigmaMcpLauncher.ts:20` (and `:43`) | **True.** Nothing imports it, its top level forks the socket server and spawns the MCP, and both `on('error', ...)` callbacks print to stderr and exit. Launched with stderr piped (an MCP server is), the message can be cut. No invoker in the repo or `~/.claude.json`: block at the top, or delete the launcher |

## Verification

- Fixtures from the real entries, both ways: Facets before and after c0269f0e (`#!`, `main().catch`, 3 exits in
  the model, reported at the first with "the first of 3"), the forked `DataConversionShardWorker.ts`,
  `PhiSocialUpload.ts`'s `main().catch(function onFatal(...))`, `StructureLintEngineParity.ts`, and
  `NewMigration.ts`, each firing as it stood and silent with the call added; a `runCommandLineInterface` CLI,
  silent as it stands and firing with a usage guard above the hand-off. 13 more firing shapes and 19 silent
  ones, among them a library other files import, a script without `#!` that is imported, a `#!` file that is
  imported (fires), an exported function nothing runs, and every way of blocking the walk honors.
- Mutation check, 22 mutants each a copy through `go test -overlay -count=1`, 21 killed: the imported-by test,
  the `#!` test, functions handed on never walked, functions called never walked, a block not ending the path,
  the write-state test, imports blocking at load ignored, `await` not blocking, callbacks walked in a file that
  blocks, a blocking function handed on, Nexus's declaration matched by name only, every call into a reaching
  file blocking, the unplaced-block guard, never ordered, a block in the arguments of the call that starts
  `main`, unresolvable callees, `import()` not blocking, the load scan entering function bodies, computed
  `import()` not a seed, the closure not propagated, and `declare`d functions. **Survived**: removing the
  unplaced-exit guard, for the reason the per-site rule's twin survives: the only exit it skips sits in a catch
  binding's default, which cannot make a later finding false.
- `gofmt -l` and `go vet` clean through the overlay. The whole suite is green apart from the two expected under
  an overlay: `TestRulesFlagListsEveryRegisteredRule` (480 against 481) and
  `TestEveryRegisteredRuleIsReachableFromTheLiveConfig` (this rule is not in the live config yet).
