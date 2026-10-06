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

1. The checker/import graph and its ProjectFiles order, dependent closure and
   post-fix rebuilds. Tsconfig roots are not the complete checker graph.
2. Full CLI execution: type, lint, fix, formatter invocation, cache/replay,
   diagnostics, version/provenance, rules/config listing, profiling, stdin,
   rename validation and rename execution. Main-phase contradictions other than
   verbose/json also depend on run ordering that is not implemented here.
3. Multiple-project child creation, capture, worst exit, signals and Swift
   dispatch. Node succeeds on gaps/child.ts; Adamic reports TS2591 for the missing
   node:child_process binding. This proves that dependency, not that the whole
   command cannot ever be ported.
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
