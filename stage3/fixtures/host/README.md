# tsc Node host fixtures

25 standalone programs cover the 19 requested System members as TypeScript
6.0.3's adapted getNodeSystem implements them. See [ADAPTED.md](ADAPTED.md)
for source pins, per-fixture changes and current loader results. The census declares 44 System members;
33 are reached at 196 sites. Its separate Node-derived ledger has 126 accesses.
These populations overlap and must not be added. In the two measured noEmit
workloads, readFile runs 53 times each, fileExists and directoryExists twice,
getExecutingFilePath twice, and getCurrentDirectory once. Four readFile fixtures
therefore cover the frequently executed decoding and error paths.

Every fixture creates and removes its own scratch directory. Output is stable
on the Linux environment recorded in status.json: paths are compared or made
relative, timestamps are fixed, environment keys are controlled, and symlink
cycles stay inside the scratch directory. The realpath fixture also exercises
the 260-character native/non-native dispatch. The exit fixtures remove their
scratch directories before exiting.

| Fixtures | System member | Real forms exercised |
| --- | --- | --- |
| 01-04 | readFile | UTF-8, UTF-8 BOM, malformed UTF-8, UTF-16LE/BE BOMs, supplementary characters, odd tails, empty and missing files, directory read failure, ignored encoding argument |
| 05 | writeFile | UTF-8 BOM, Unicode, overwrite/truncation, missing parent, finally closing the fd |
| 06 | fileExists | file, directory, absent entry, file/directory symlinks, broken link, ENOTDIR |
| 07 | directoryExists | same stat and symlink distinctions |
| 08 | getDirectories | sorted directory names, file exclusion, symlink following, broken-link exclusion, missing directory, empty path and returned-array independence |
| 09 | realpath | symlink resolution, missing-path identity, relative-path identity on failure, long-path dispatch |
| 10 | getModifiedTime | fixed mtime, symlink following and missing result |
| 11 | setModifiedTime | both atime and mtime set, ignored missing-path failure |
| 12 | deleteFile | unlink a symlink without deleting its target, existing/missing file, swallowed directory-unlink error |
| 13 | createDirectory | creation, existing directory, EEXIST from an existing file, rethrow of ENOENT |
| 14 | getCurrentDirectory | memoized cwd retained across chdir |
| 15 | getExecutingFilePath | sys.js fake bundle location, normal bundle filename, sys.cjs nonmatch |
| 16 | getEnvironmentVariable | absent/empty return empty, string zero stays present, Unicode |
| 17 | write | consecutive writes without added newlines, empty write and Unicode |
| 18-20 | exit | status 0, 1 and 2, inactive CPU profiler, no continuation after exit |
| 21-22 | createHash | SHA256 and absent-crypto djb2 fallback, empty/ASCII/Unicode/NUL input, overflow-prone long fallback input |
| 23 | newLine | the actual OS EOL, diagnostic/version-style concatenation |
| 24 | useCaseSensitiveFileNames | win32/win64 forced false, actual file-name case probe in both directions, swapCase |
| 25 | readDirectory | actual matchFiles and its dependencies, extension/include/exclude filtering, include order, depth, sorted traversal, hidden/package-folder omission, alias and cycle realpaths, missing directory |

The hardest case is 25. Its matcher, glob/path helpers and Debug assertion
functions are copied from upstream rather than replaced by a different walker.
A source audit checks 164 declarations, including 126 function/method spans,
against the actual stage3/apply.sh output using
canonical parsed syntax. Function statements retain their original operators,
control flow, calls and literals. Removing the outer scope, indentation and
CRLF formatting does not change them. Interfaces and helper types are included
where needed. Numeric enum constants used only as named constants are driver
scaffolding with the exact upstream values; these fixtures do not test enum
object reflection. No Node runner change from another branch is required.
The three exit fixtures retain disableCPUProfiler's inactive else branch and
explicitly reject the active-session case in their scaffolding. The executing
filename and case-probe platform are controlled driver inputs.

`status.json` has the requested schema and records Node stdout, stderr and exit,
plus canonical stage 0 diagnostics under the library loader pinned in ADAPTED.md.
All 25 Node observations are unchanged. Current results are 0 Checker,
14 Refused, 11 NotYet and 0 Compiles with library c13622a0 and main's checker
options. Stock tsc accepts all 25. See 25-followup-proof.json for every status
movement and STYLE-OPTIONS.md for checker configuration.
Fixture 11 historically passed native byte comparison, sanitizer checks and
the Linux leak check; the current library integration stops at utimesSync. source-spans.json preserves the precise adapted-source provenance.
Historical validation.md describes the initial pristine-source recording;
ADAPTED.md supersedes its source and stage 0 results.

After sourcing the setup environment, reproduce from the Adamic root:

```sh
python3 stage3/fixtures/host/check.py --compiler-repo <scratch-stage3-library-merge> --mutants --logs /tmp/host-fixture-check > /tmp/host-fixture-check.log 2>&1
NODE_PATH=<stock-typescript-6.0.3-node_modules> node stage3/fixtures/host/source-audit.cjs <stage3-apply-output> > /tmp/host-source-audit.log 2>&1
```

The first command independently invokes
`node --disable-warning=ExperimentalWarning oracle/node.mjs <fixture>` and
`go run ./cmd/adamic build <fixture> -o <scratch-binary>` for every row. It checks
exact bytes and exit codes, including the nonzero success observations for
exit_1 and exit_2. Each fixture has a semantic mutant, listed in check.py and
validation.md, and every final mutant is caught. A changed recorded compiler
diagnostic and a changed upstream hash implementation are also caught.

Left out: watch/timers, active inspector profiling, terminal width/TTY/blocking,
memory usage, general plugin requires, performance clocks, Windows filesystem
execution, permissions under another uid, concurrent mkdir races, and fault
injection of partial writes/close failures. These are outside the requested
19-member fixture list or need a separate controlled workload. Directory
creation still contains tsc's race-handling statements. readFile's ignored
encoding parameter is preserved, not replaced by a new encoding policy.

## For the require-builtins worker

@system_adamic_compiler is implementing literal builtin requires natively on
`codex/require-builtins`. No source adaptation is needed. These seven calls
retain their original conditional or lazy load behavior:

| Source in src/compiler | Builtin | Why retained |
| --- | --- | --- |
| sys.ts:1470 | fs | getNodeSystem body; initialized only on the Node-like branch |
| sys.ts:1471 | path | getNodeSystem body; initialized only on the Node-like branch |
| sys.ts:1472 | os | getNodeSystem body; initialized only on the Node-like branch |
| sys.ts:1476 | crypto | getNodeSystem body on the Node-like branch, plus a catch for reduced Node installations without crypto |
| sys.ts:1650 | inspector | enableCPUProfiler body, after the active-session early return |
| tracing.ts:63 | fs | startTracing body, only while fs is undefined, with a load-error catch |
| performanceCore.ts:35 | perf_hooks | tryGetPerformance body, inside Node-like detection and a catch |

The seven calls include fs twice. `source-map-support` at sys.ts:1597 is an
optional third-party package and remains inside its catch. The plugin require
at sys.ts:1619 has a nonliteral specifier and remains dynamic.

An unconditional static import would move load failures outside these guards
and load inspector/tracing before the first request. Keeping an assignment
conditional after hoisting its import does not preserve the load timing.
Namespace imports can also change the identity and mutability of CommonJS
module objects. The fixtures' explicit static node: imports provide inputs for
the library worker; tsc's original requires retain their load semantics.

The unused adaptation was removed; number 50 belongs to the scanner proof's
temporary adaptations. Historical oracle evidence is retained in this bucket.
The fixture Go test owned by another worker was not edited.

Fixture 14 now applies adaptation 48 A, re-extracted from the actual composed
adapted tree. Its memoize parameter is `callback: (() => T) | undefined` and
its clearing assignment is `callback = undefined;`. Emitted JavaScript and
Node stdout/stderr/exit are unchanged. See
[ADAPTED.md](ADAPTED.md#fixture-14-applied-memoize-a) for current source pins,
audit and mutant evidence. The earlier decline is historical; tsc adaptation
acceptance and native compiler support are separate. The old status.json
compiler observation for fixture 14 has not been remeasured on this source.
