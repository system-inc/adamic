Project discovery, lint globs and formatter lint ignores are ported; this unit is incomplete.
128 roots, 626 project markers and 288 glob queries agreed with Go cohere.
Three discovery mutants and twelve formatter mutants were caught on native and Node; sanitizer and leak checks passed.
Settings resolution and tsconfig source enumeration are not implemented.
JSON.parse is deliberately refused; CohereSettings is strict JSON in Go, unlike tsconfig.

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
ownership, or expansion of solution references. `main.ts` takes absolute roots and
ignorePatterns as input. It does not read those patterns out of settings or pretend
that they have been resolved by Adamic.

`glob.ts` ports `internal/lint/configuration/glob.go`: nested brace alternatives,
whole-segment double stars, slash boundaries, ordinary stars and question marks,
unbalanced braces as literals, and byte matching via utf8At/utf8Length. These are
lint globs, different from the gitignore glob already ported. formatfiles uses this module for ignorePatterns relative to the settings directory,
with Go's restricted directory pruning. Settings resolution remains an oracle
input, not an Adamic settings reader.

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
made-up root tsconfig. TypeScript source enumeration still needs to run over the
individual configs, including solution ownership and references.

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

This blocks a direct library-based settings port. It is not proof that a checked
parser written in Adamic is impossible. Such a parser, its JSON/JSONC diagnostics,
settings validation/merge and TypeScript configuration semantics are not built here.
No settings are decoded by a Go bridge and presented as an Adamic implementation.

## Measured settings behavior

`TestSettingsBehaviorCensus` overlays a test into Go cohere's command package and calls
its production `configuration.Load` and `program.ReadProjectConfig`. It replaces only
the temporary directory prefix with ROOT in the output. The following are actual
Go observations, not implemented Adamic answers:

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

The requested JSONC acceptance for CohereSettings and exact Go parity disagree on
the first two rows. No behavior was broadened. The census records this discrepancy;
there is no Adamic settings loader or canonically printed resolved settings yet.

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
an isolated same-work throughput comparison. Requested source files/s against Go
is not measured because the source enumeration port is not built.

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

Not covered: settings loading/defaults/embedded sets or canonical resolved settings;
tsconfig include/exclude, extensions, files ordering, extends and references in
Adamic; permission/race errors; non-UTF-8 names; macOS or Windows; inherited .gitignore
read failures the disk adapter cannot faithfully represent; arbitrary filesystem
errno (the runtime collapses unknown errors). Unknown listing errors are explicitly
named unrepresented filesystem errors; no exact Go parity is claimed for those.

Formatter integration: 423 trees, 29,366 files offered, 1,970 directories entered,
73 refused walks all agreed in the final run (261.717 s). All twelve formatter
mutants were caught on native and Node. Go harness timing 1.825032 s (16,090
files offered/s); native 19.716329 to 22.206697 s (1,322 to 1,489 files offered/s).
See REPORT.md for commands, exact catches and timing caveats. This still does not
measure or implement tsconfig source enumeration.
