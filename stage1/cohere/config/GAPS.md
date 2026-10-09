Strict CohereSettings and JSONC tsconfig loaders are implemented in Adamic.
Real Adamic, cohere and TypeScript project outputs are compared byte for byte with Go.
Settings defaults, extends, provenance, per-file overrides and source enumeration are exercised.
Three loader mutants join three discovery mutants and twelve formatter mutants.
Directory-link parity is closed; advanced tsconfig diagnostics remain explicit gaps.

## What is built

`discovery.ts` ports `cohere/command/cohere/discovery.go`'s `discoverProjects`, before
TypeScript ownership processing. It imports the existing gitignore matcher and
formatfiles disk adapter. It reads the nearest repository's ignore files, including
ancestors when the requested root is below the repository root, enters each scope,
prunes ignored directories, skips dependency/cache/fixture directories, distinguishes
clones from gitlinks, and refuses markers matched by lint ignorePatterns. Directory
symlinks are not followed. A dangling symlink named tsconfig.json is a marker, as Go's
os.ReadDir sees it; discovery does not validate its contents.

The Go reads directories concurrently and sorts the result. The port reads them
sequentially and sorts by the same segment order, with TypeScript before Swift in one
directory. It does not implement DirectoryListings reuse, discoveryRoot, project
ownership, or expansion of solution references. `main.ts` accepts absolute roots and explicit ignorePatterns, or a configured root.
The configured entry reads strict CohereSettings and inherited ignores through the
Adamic loader before discovery.

`glob.ts` ports `internal/lint/configuration/glob.go`: nested brace alternatives,
whole-segment double stars, slash boundaries, ordinary stars and question marks,
unbalanced braces as literals, and byte matching via utf8At/utf8Length. These are
lint globs, different from the gitignore glob already ported. formatfiles uses this module for ignorePatterns relative to the settings directory,
with Go's restricted directory pruning. The formatter differential driver still takes resolved formatter settings from its
Go oracle. The new settings loader is independent of that formatter driver.

## The Go front door actually found

At cohere submodule `715ba94f3608a6500086b1076ce5cb7e51b836db`:

- `command/cohere/discovery.go` finds project markers and prunes by gitignore.
- `internal/types/program/project_config.go`, `ReadProjectConfig`, delegates tsconfig
  JSONC, extends, defaults, source extensions, include/exclude and file ordering to
  TypeScript's `tsoptions.GetParsedCommandLineOfConfigFile`. A formatfiles walk is
  not this enumeration and cannot replace it.
- `internal/lint/configuration/configuration.go` reads strict JSON with encoding/json,
  validates keys, reads extends chains, merges rules and inherited options,
  concatenates ignores and overrides, and records source/reason provenance.
- `internal/lint/configuration/sets.go` provides embedded cohere rule sets. Zero
  configuration and per-file house defaults are in `house.go`, not an empty config.

Neither cohere nor its TypeScript submodule has a root tsconfig.json. Their command
walks discover projects below the root. Tests explicitly exercise Adamic, cohere and
cohere/TypeScript as separate discovery roots; they do not silently substitute a
made-up root tsconfig. The new loader corpus enumerates the individual configs and prints their references.
Command-level project ownership and expansion of solution references remain outside
this loader; enumerated source records can repeat across overlapping projects.

## Proving program: JSON parsing

`gaps/json.ts`:

```ts
JSON.parse('1');
console.log('parsed');
```

Node prints `parsed` and exits 0. It typechecks through load.Load. Lowering refuses:

```
Adamic 0.1 refuses JSON.parse: its result's type can't be proven from the text;
a checked parse against a declared type is a later design; construct typed values explicitly for now
```

`TestJSONLibraryGap` checks the exact lower.Refused.What, rather than accepting any
compiler failure. This is a deliberate refusal, not NotYet. An earlier probe stored
the result as unknown and was refused earlier for that representation; the discarded
result above isolates the parser itself. No compiler files were changed.

This rules out JSON.parse, not a typed parser. `json.ts` now implements a checked
flat node representation in Adamic and drives both loaders. No host parser or Go
bridge supplies resolved configuration. The proving program remains useful evidence
of the library boundary; it no longer blocks these loaders.

## Measured settings behavior

`TestSettingsBehaviorCensus` overlays a test into Go cohere's command package and calls
its production `configuration.Load` and `program.ReadProjectConfig`. It replaces only
the temporary directory prefix with ROOT in the output. The following are actual
Go observations. The new loader comparisons now require the same Adamic answers:

```
comments: parsing lint config ROOT/comments.json: invalid character '/' looking for beginning of object key string
trailing: parsing lint config ROOT/trailing.json: invalid character '}' looking for beginning of object key string
missing: reading lint config ROOT/missing.json: open ROOT/missing.json: no such file or directory
cycle-a: lint config ROOT/cycle-a.json extends ./cycle-b.json: lint config ROOT/cycle-b.json extends ./cycle-a.json: lint config ROOT/cycle-a.json extends itself: ROOT/cycle-a.json -> ROOT/cycle-b.json -> ROOT/cycle-a.json
child: severity error options ["always"] ignores [base/** child/**] sources [ROOT/child.json ROOT/base.json]
tsconfig-comments: files ROOT/a.ts
missing-tsconfig: no tsconfig at ROOT/missing-tsconfig.json
tsconfig-cycle-a: reading ROOT/tsconfig-cycle-a.json: TS18000:
tsconfig-base-missing: reading ROOT/tsconfig-base-missing.json: TS5083:
```

The last two actual diagnostic strings end with a space after the colon: cohere's
joinDiagnostics sees empty MessageText there. Exact parity would keep that, rather
than inventing a more helpful message. The census test logs the original whitespace.

Upstream observation: the pinned Go cohere uses strict JSON for CohereSettings.json
and JSONC for tsconfig.json. Comments and trailing commas in CohereSettings are
rejected with the exact errors above. The stage 1 contract preserves this asymmetry,
as confirmed by the user, rather than broadening upstream behavior.

## Differential cases and mutants

`config_test.go` overlays `testdata/cohere_side_test.go` into Go's actual command
package, without editing the submodule. It compares full output bytes on Node source,
ASan/UBSan native and the JavaScript backend. Native runs again with LeakSanitizer.
A failing baseline stops before mutants, so an existing disagreement cannot kill a
mutant by accident.

Cases include all three requested roots, 120 seeded generated trees (20261006), a
fixed tree with nested ignores and negations, root-relative and nested lint ignores,
Unicode ordering, clone/gitlink boundaries, directory/file/dangling/self-loop
symlinks, a linked .gitignore refusal, an interior repository root, a missing root,
and 288 pattern/path pairs. All output fields are compared, not only project counts.

| Mutant | First discrepancy, native and Node |
| --- | --- |
| Drop gitignore directory pruning | counts 2 0 4 versus Go 2 3 4 |
| Double star must consume at least one segment | glob 0 versus Go glob 1 |
| Enter nested repositories | gitignore nested-repository error versus Go's root counts |

The third mutant still exits 0 and prints an error value, so its failure is an output
comparison, not a compiler error or sanitizer failure.

## Performance boundary

`ADAMIC_CONFIG_TIMING=1 go test -v -count=1 -timeout 30m ./stage1/cohere/config`
builds native without sanitizers and checks its timed output too. On this machine,
626 discovered markers in the same case set: Go 0.024023 s (26,058 projects/s),
native 0.138377 to 0.155937 s (4,014 to 4,524 projects/s). These are discovery
measurements, not source files checked per second. Native includes process startup
and output capture; Go times production calls and output construction within its
already-started test process. Go also reads directories concurrently. They are not
an isolated same-work throughput comparison. The continuation measures tsconfig source files/s separately below.

## Validation and remaining coverage

Setup printed Go, clang, Node and submodules ready at 0 s, cache warm at 78 s, and
completed at 78 s. nproc: 5; cgroup cpu.max: 400000 100000; reported memory: 17.6 GB.
Tools: Go 1.27.1, clang 20.1.8 with working sanitizers, Node 24.19.0. Environment
file: /workspace/adamic-tools/env.sh.

The first parity run exposed a translation error: readDirectory says `no such
directory`, not fileStatus's `no such file`. The port now translates the former to
Go's `open ...: no such file or directory`. The rerun passed in 16.035 s, including
all three mutants, settings census, gap probe, three-way parity and the leak check.
Logs: /tmp/stage1-config-test-final.log and /tmp/stage1-config-setup.log.

Further validation commands and results are recorded in REPORT.md.

Not covered: command-level ownership; full arbitrary tsconfig diagnostics and
compiler-option validation; advanced package exports resolution; published-version admissibility (the pinned Go
oracle is a dev build); relative loader entry paths; permission/race errors; non-UTF-8 names; macOS or Windows; inherited .gitignore
read failures the disk adapter cannot faithfully represent; arbitrary filesystem
errno (the runtime collapses unknown errors). Unknown listing errors are explicitly
named unrepresented filesystem errors; no exact Go parity is claimed for those.

Formatter integration: 423 trees, 29,366 files offered, 1,970 directories entered,
73 refused walks all agreed in the final run (261.717 s). All twelve formatter
mutants were caught on native and Node. Go harness timing 1.825032 s (16,090
files offered/s); native 19.716329 to 22.206697 s (1,322 to 1,489 files offered/s).
See REPORT.md for commands, exact catches and timing caveats. This formatter measurement does not measure tsconfig source enumeration.

## Loader continuation

`settings.ts` ports Load/LoadFor, schema validation, embedded sets, extends errors,
ancestor conflicts, inherited options, reasons/departures, plugin defaults, version
pin validation and ordered overrides. `house.ts` ports the house variants and
per-file resolution from typed React/Next/Tailwind detection input. AST detection
itself is not implemented here. `sets.ts` embeds upstream source JSON, not resolved
oracle answers. `json.ts` preserves Go's strict syntax errors, Unicode replacement
and raw option spelling; `quote_table.ts` records pinned Go strconv.IsPrint data.
The generator is `testdata/quote_table.go`. See NOTICE.md for source licensing.

`tsconfig.ts` implements JSONC, relative/package extends, inherited include/exclude
and files, configDir substitution, allowJs/resolveJsonModule, output-directory
excludes, extension priority and include-bucket file order. `tsglob.ts` implements
TypeScript's matcher, distinct from lint globs. The pinned Go ignores custom
sourceExtensions for wildcard enumeration; generated cases retain that observation.
Missing files and cycles preserve Go's error text, including trailing spaces after
empty TS diagnostic messages. Malformed JSONC diagnostics and contentMappers are
explicitly declined; complete compiler-option diagnostics and modern package exports
resolution are not claimed by this port. Ordinary file symlinks are followed.

### Directory-link proving program

`gaps/directory_link.ts` now runs under TestDirectoryLinksAgreeWithGoCohere.
The fixture creates a-alias -> z-target with a.ts, a backlink to the root (a cycle),
and a dangling link. Go and all three Adamic execution paths emit a-alias/a.ts once
and omit the duplicate target spelling. The previous explicit realpath gap is closed.

The user authorized the runtime proposal as a separate commit. realPath is now a
typed filesystem primitive with native ownership and JavaScript-backend support.
Canonical directory keys stop duplicate visits and cycles; returned file paths keep
the first alias spelling. The direct Node fs.realpathSync fixture covers a directory
symlink, a dangling link, an empty path, lexical dangling/.. and a file reached through
a directory alias. Native normalizes the input before POSIX realpath, matching the
observed distinction between Node's plain and native realpath variants. Two clean
executable mutants corrupt successful canonical paths and dangling-link errors;
Node stdout comparison catches both, independently of sanitizers and leak checks.

This closes the measured directory-link cases, not every platform or unusual errno.
See REPORT.md for commands, outputs, counts and remaining coverage.

### Loader comparisons and mutants

The loader corpus includes all real configs under the three requested projects,
80 seeded settings chains and 80 seeded tsconfig trees, strict syntax/schema errors,
Unicode quoting, duplicate fields, unknown keys, missing files, cycles, house variants,
registered rule aliases, ignores and overrides. Full canonical settings, source
paths in order, references and errors are compared across Go, Node source, native
ASan/UBSan and JavaScript backend; native LeakSanitizer runs too. A baseline mismatch
stops mutation testing. Mutants must compile and run successfully before their
output differences count.

| Loader mutant | Native and Node catch |
| --- | --- |
| Parse CohereSettings as JSONC | Comments accepted instead of Go's exact syntax error |
| Drop inherited rule options | @typescript-eslint/no-shadow error [] instead of error ["option"] |
| Disable tsconfig exclusion | src/b.ts appears instead of Go's next selected file |

The first inherited-options mutant lost TypeScript narrowing and did not compile;
it was rejected as evidence. The corrected mutant changes only the selected options
value and is caught on both executable sides. Final commands, counts and timings
are recorded in REPORT.md. Full arbitrary configuration parity is not claimed for
uncovered behaviors above; real-project and generated-case parity is measured.

Final result: 143 settings, 130 tsconfigs and 1,232 source-file records agree across
Go, native, Node and backend; final package PASS 124.264 s. Tsconfig-only harness:
Go 63,542 files/s; native 15,800 to 16,722 files/s. Startup/output boundaries differ;
see REPORT.md. All three loader and three discovery executable mutants were caught.

Authorized realpath continuation: directory aliases, cycles and dangling links now
match Go; direct Node canonical-path and error mutants are caught. Final config
package PASS 120.172 s. Scoped compiler packages, seven uncached input fixtures,
full allocation counts, ASan/UBSan, LeakSanitizer and vet pass; no full repository
gate was run. Latest source harness: Go 72,688 files/s, native 15,249 to 16,612.
