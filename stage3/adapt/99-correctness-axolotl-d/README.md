# 99d: caught config-file error message

JavaScript-changing at one diagnostic argument in `tryReadFile`:
`e.message` becomes `String(propertyOf(e, "message"))`. No cast, checker
option change, public API edit, or other call-site change is made.

Task axochk, census row D032, TS18046 at stock
`src/compiler/commandLineParser.ts:2301`. This follows Adamic's ruling for
the twin in program.ts. Stock TypeScript is 6.0.3, commit
`050880ce59e30b356b686bd3144efe24f875ebc8`. Main was recorded at
`80c2d3282e44b527555d59690bb413cc883d2bf2`. The delivered branch is exactly
94's head, `8a7cd267b434efa0a60d481529a1256cc3c92a19`, plus this commit.

## Adapter

Stock TypeScript 6.0.3 parses the current file. The adapter requires the exact
94 helper suffix in adaptation 47's hostErrors.ts, a unique named
tryReadFile function, its single catch binding e, and the exact diagnostic
return statement. It accepts only the original or fully adapted import/call
pair and refuses half-adapted files before writing. It checks syntax and
unchanged input before its single write. A second run makes no write.

94 does not export propertyOf through the namespace barrel. This file has
no hostErrors import in the baseline, so one direct named import is added.
If a named import already exists, its declaration is extended instead of
adding a second declaration. Duplicate declarations, type-only imports,
aliases of propertyOf, and unreviewed catch text are refused.

## Node witnesses

`witnesses.cjs` extracts each tree's actual tryReadFile function, transpiles
it with stock 6.0.3, loads that tree's actual hostErrors helper, and supplies
stock createCompilerDiagnostic and Diagnostics. The formatter's
formatStringFromArgs checks each argument with Debug.checkDefined and then
concatenates it. Node 24.19.0 observations follow. Diagnostic cells give the
suffix in `Cannot read file 'config.json': <suffix>.`.

| Thrown value | Before | After |
| --- | --- | --- |
| `new Error("read failed")` | `read failed` | `read failed` |
| `{message: "only message"}` | `only message` | `only message` |
| `"thrown"` | Error: `Debug Failure.` | `undefined` |
| `42` | Error: `Debug Failure.` | `undefined` |
| `{}` | Error: `Debug Failure.` | `undefined` |
| `{message: 42}` | `42` | `42` |
| `null` | TypeError: `Cannot read properties of null (reading 'message')` | same |
| `undefined` | TypeError: `Cannot read properties of undefined (reading 'message')` | same |

A real Error's diagnostic text is identical. Numeric messages already work
in the original formatter; the newly successful witnesses are the three
undefined-message cases. Null and undefined still throw through 94's helper.
The conversion is intentional behavior, not an assertion of message type.

## Harness reachability

tryReadFile invokes its supplied readFile callback inside the try. Its
callers include readConfigFile, readJsonConfigFile,
getParsedCommandLineOfConfigFile and the command-line worker's
parseResponseFile (for @response files). A callback that throws
can reach the catch; the public callback contract does not restrict thrown
values.

The compiler harness's compilerImpl.ts:34 calls readConfigFile with its
host's readFile. fakesHosts.System.readFile:56 catches VFS errors and returns
undefined. ParseConfigHost.readFile:220 and CompilerHost.readFile:321 delegate
to that system. The language-service harness's LanguageServiceAdapterHost
readFile:288 similarly catches VFS failures, and NativeLanguageServiceHost.readFile:428
delegates to sys. The watch TestServerHost.readFile:1043 returns undefined
for missing files. Configuration extension tests pass host.readFile to
readJsonConfigFile at configurationExtension.ts:375. Native sys.readFile
also catches filesystem failures and returns undefined. These ordinary
missing-file paths produce the separate Cannot_read_file_0 diagnostic,
without entering the catch. Custom throwing callbacks or a host failure that
escapes those handlers can enter it. The counter below measures the suite,
not every possible injected host.

## Reproduction

Source the cloud setup environment, install stage3/api packages from the
lockfile, and set ADAMIC_TYPESCRIPT_SOURCE to the pinned pristine checkout.
Use fresh output directories. Run baseline apply from 94 without this
adaptation directory; run final apply from this branch.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/99d-setup.log 2>&1
source /workspace/adamic-tools/env.sh
bash stage3/apply.sh /tmp/99d-after > /tmp/99d-apply.log 2>&1
export NODE_PATH=$PWD/stage3/api/node_modules
node stage3/adapt/99-correctness-axolotl-d/adapt.cjs /tmp/99d-after > /tmp/99d-idempotence.log 2>&1
node stage3/adapt/99-correctness-axolotl-d/witnesses.cjs /tmp/99d-before /tmp/99d-after > /tmp/99d-witnesses.json 2>&1
```

Measurements and exact commands are on `typescript/axolotl-99d-evidence`.

## Measured proof

The primary measurement checkout is recorded main plus cherry-picks
`d42771c36`, `087d20d81` and `8a7cd267`, producing
`8ea8c394ff40e7078675445a7a9afb7a816aaf24` before adding 99d. Baseline and
final apply.sh runs both pass end to end. The delivered adapter is identical
to the measured adapter. Its incremental patch row is one file, two lines
added, one removed: the import and the replaced diagnostic argument. A
recursive source comparison finds no other changed file. The baseline and
final commandLineParser.ts bytes also match the corresponding delivery-base
control trees exactly. Refusal checks cover missing/changed helper, import
only, call only, catch drift and duplicate imports, with no writes on refusal.
An existing named import is extended once and the second run writes nothing.

The unfiltered default oracle ran on the final main-tip tree with a temporary
counter at this catch in a separate scratch copy. npm ci and npm run build
pass. Tests report **106,366 passing, 1 failing, 0 pending**, with only
`api/typescript.d.ts` in baseline_diffs. This is the standing API case, not
an empty-diff pass. No reference baseline is edited. Oracle wall time is
273.296 seconds with eight workers on Node 24.19.0.

The counter appends `[typeof e, typeof e?.message]` per entry to a shared JSONL
file across test processes. It records **zero entries**, so the complete
observed pair-to-count map is `{}`; there are no observed pairs. Both emitted
compiler and test-runner bundles contain the counter. A separate positive
control loads the emitted compiler with only the counter path redirected,
throws an Error into tryReadFile, and records exactly one `object,string`
pair plus the expected diagnostic. The suite counter bytes stay unchanged.
The counter is never part of the delivered adapter or final adapted tree.
The suite does not exercise this catch; the Node witnesses supply its
behavior proof.

The real `adamic types` CLI is built at recorded main. It checks the same 79
sorted .ts/.a roots recursively under src/compiler in each tree, with the
loader's unchanged strict options. Results are whole-program, not filtered
to this file:

| Diagnostics | Before | After | Restoring mutant |
| --- | ---: | ---: | ---: |
| Every code, all roots | 54 | 53 | 54 |
| TS18046, all roots | 4 | 3 | 4 |
| D032 | 1 | 0 | 1 |

The only removed diagnostic is
`commandLineParser.ts:2301:91: error TS18046: 'e' is of type 'unknown'.`
Exact multiset comparison retains complete messages, columns, multiplicity
and elaboration chains, replacing tree roots and mapping positions after
the added import back one line. There are zero additions. The mutant changes
only `String(propertyOf(e, "message"))` back to `e.message`, retaining the
import; its normalized whole-program multiset equals the baseline exactly.
Remaining diagnostics are not claimed repaired.

The measurement driver uses argument arrays for all 79 roots:

```python
roots = sorted(str(p) for p in (tree / "src/compiler").rglob("*")
               if p.suffix in [".ts", ".a"])
subprocess.run([adamic, "types", *roots], stdout=out, stderr=err)
```

Cloud setup passes at recorded main, using
`GOPROXY='https://proxy.golang.org|direct'`, Go 1.27.1, clang 20.1.8 and
Node 24.19.0. Total setup time is 1192.675 seconds, including the cold checker
submodule restoration. Logs, exact measurement drivers, full diagnostic
outputs, oracle phase logs, baseline diff and counter proof are kept only on
the evidence branch, outside this small delivery directory's code commit.
