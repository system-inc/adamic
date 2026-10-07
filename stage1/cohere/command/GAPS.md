The command front door is partly ported. Full Go CLI parity is not claimed.

Observed and held to Go:

- All 31 main flags and five rename flags, their defaults, help, positional
  stopping, duplicate values, explicit false values, boolean spellings and int64
  parsing. Integers remain decimal text rather than rounded JavaScript numbers.
  Failed setters change stored values, as Go does. Unsigned overflow can precede
  later syntax errors; underscore validation happens after the digit scan.
- Project-root discovery, TypeScript versus Swift markers, explicit-directory
  precedence, lint-config anchoring, and discovery's choice of repository root.
- Named paths, duplicates, UTF-8 sorting, missing-file errors, whole-tree versus
  subdirectory dot, and formatter/population narrowing and descriptions. Narrowing preserves the
  original scope as Go value receivers do.
- Solution references, nearest-project ownership, shared-file ties, projects
  yielding all their files, same-directory cache yields and child lint-config
  arguments, against Go's original command helper tests on their live fixtures.

The fixture with a linked directory observes that namedPathsScope does not follow
it: Go uses Stat for the initial classification, then WalkDir's Lstat sees a link
and retains that directory-link path as one file. Links found within a directory,
including dangling links, are also retained. This differs from tsconfig source
enumeration's canonicalized directory-link traversal, already ported in config.
Named scopes do not apply gitignore; the formatter walk applies its own ignore
layers. Neither file set can substitute for the other.

An escaped JSON key spelling references bypasses Go ownership's literal byte
search when there is only one TypeScript project. This is an upstream observation,
not a corrected behavior. The generated edge fixture preserves it, along with
cycles, missing references and dependency-directory references.

Still open:

1. Compiler graph construction, module resolution and ProjectFiles selection/order.
   graph.ts now ports the reverse-import closure, exact 500-file refusal, whole-tree
   fallback note, dependent counts and rescopeAfterRebuild's retained/lost names.
   Its inputs are already resolved compiler files and edges. The test overlay
   obtains those inputs from real Go compiler graphs; no expected closure enters
   the port. There is no production graph builder yet. TestAdamicGraphConstructionGap
   observes tsconfig roots a.ts,b.ts but ProjectFiles b.ts,a.ts after a imports b.
   Tsconfig roots therefore cannot substitute for the ordered compiler population.
   Go also keeps imported files and bundled declarations outside ProjectFiles.
   Post-fix scope transfer is ported; actual graph rebuilding still needs the engine.
2. Full CLI execution: type, lint, fix, formatter invocation, cache/replay,
   diagnostics, version/provenance, rules/config listing, profiling, stdin,
   rename validation and rename execution. Main-phase contradictions other than
   verbose/json also depend on run ordering that is not implemented here.
3. The multiple-project scheduler, worst-exit aggregation, parent signal forwarding,
   child cancellation, yield-file lifecycle and Swift dispatch remain open.
   process.ts now runs a real child through the new runProcess primitive: arguments,
   cwd, inherited environment with verdict/yield removal and engine/label injection,
   combined stdout/stderr, ordinary exit, self-SIGTERM, missing executable, denied execution and missing/non-directory
   cwd match Go runProject on live fixtures. The primitive is synchronous, Unix-only
   and requires an absolute executable. It does not implement runProjects or install
   parent signal handlers. A parent killed alone can leave its child alive; Go's
   TestATerminatedMixedRunLeavesNoChildBehind is not claimed as passing native.
   A single unlinked capture file avoids pipe deadlocks, rather than matching Go's
   pipe transport; seekability and inherited-writer lifetime are unheld. Native
   capture uses tmpfile; Node uses a temporary directory.
   Temporary-file errors, descriptor exhaustion/inheritance, races, huge strings,
   unusual errno/signals and binary non-UTF-8 Go CLI output are outside measured parity.
   The primitive decodes output as UTF-8; Go's raw captured bytes can differ there.
   Case-insensitive Unicode path keys and Windows normalization are also unheld.
   TestNodeChildProcessModuleGap still proves that arbitrary node:child_process
   imports are unsupported; it no longer establishes a lack of native child execution.
   Node's direct spawnSync is the independent primitive oracle, including invalid
   bytes, NUL output, empty stdin, environment patches and 192 KiB output.
4. Host exit status. Node runs gaps/exit_status.ts and exits 7; ordinary native
   emission treats that ambient process as undefined and panics with exit 70.
   The test's small host adapter binds commandExitCode after cleanup. It also
   transports argv[0] ahead of the user's arguments. No compiler file is edited.
   main.ts requires that adapter protocol; an ordinary adamic build is not a
   standalone replacement for cohere.
5. Filesystem errors beyond measured Linux missing-file/permission cases, races,
   Windows paths, invalid UTF-8 filenames, and logical PWD spelling through symlinked working directories.
   Unsupported errno cases are named rather than represented as exact Go errors.
   Ownership inherits the config slice's unrepresented compiler-option errors
   and package-resolution gaps.

TestExecutionGapNeverReturnsAGreenRun proves the boundary using a real project:
Go --types --no-cache --directory ROOT exits 0; this CLI driver prints
"cohere: stage 1 command execution is not yet ported" on stderr and exits 1.
It does not report an empty successful check in place of the missing engine.

CohereSettings.json remains strict JSON and tsconfig remains JSONC. The upstream
asymmetry and exact strict-JSON errors remain documented in ../config/GAPS.md.

Continuation validation and four new executable mutants are in
[GRAPH_PROCESS_REPORT.md](GRAPH_PROCESS_REPORT.md). These helper boundaries are
not wired into main.ts's missing checker run. A successful Go engine request still
gets the explicit stage 1 execution-gap error, never a green check over no engine.

An upstream process-error observation is preserved: Go os.StartProcess stats cwd
before launching. A missing cwd fails that stat and says "chdir <cwd>: no such
file or directory". An existing file passes stat, then fails the child's chdir and
says "fork/exec <executable>: not a directory". The new fixture caught the initial
adapter's intuitive but incorrect chdir wording. Native and both Node paths now
follow the Go pre-stat split; uncommon pre-stat errors/races remain unheld.
