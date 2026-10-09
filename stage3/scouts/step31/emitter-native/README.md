Native emitter compilation is blocked in both split modes. The first stop is
`moduleNameResolver.ts:2772:21`, TS2322: an explicit possibly-undefined
`originalPath` cannot inhabit the optional field of `Resolved` under
`exactOptionalPropertyTypes`. No emitter binary, native comparison, C compilation,
or ownership/leak result is claimed. A run with recorded resolver answers is
component evidence, not Step 31's proof.

This scout delivers a closed emitted-text replay for one real compiler project,
a bounded walk of fifteen native build stops, minimal Node witnesses, proposed
owners, and checks that reject source-input and evidence mutants. It changes no
compiler implementation or adaptation.

Pins used on selection:

| Input | Commit |
|---|---|
| Main and delivery branch base | `031a1259bc7973934792dc6cb1bd4074fc2204b9` |
| Newest area-next candidate, `origin/cloud/land-area-next-auto-547551cb` | `79dd1abae85972b8123fb813d2d98e49ef612730` |
| Requested Node observer/comparator scout | `2de9fc1b01d3da3beb35c554e70d98cd6d9fe536` |
| TypeScript 6.0.3 | `050880ce59e30b356b686bd3144efe24f875ebc8` |
| Cohere, shared unchanged by main and candidate | `7945d102a6c18dd36adf9114a758ce646e8b2359` |

The candidate was not an ancestor of main. Candidate selection, binary hashes,
and source hashes are retained under [evidence](evidence/summary.json).
Both compiler builds use `-buildvcs=false` because the detached candidate shares
the pinned cohere submodule through a symlink. Production compiler files have
no scratch edits. Neither scratch worktree is pushed. Only this new directory
is committed on `codex/step31-emitter-native`.

The replay and its boundary

`record.mjs` uses stock TypeScript's AST to find the real `emitFiles` function
body and insert a forwarding recorder in memory. The rest of each compiler
module is independently transpiled by stock TypeScript. The scout's real Node
observer runs `Program.emit` with the real checker; the recorder retains the
resolver method, argument node kind/positions, and returned answer. It never
substitutes a resolver answer during capture.

The project contains an exported typed constant and a function returning that
constant, emitted as CommonJS. Its **17 resolver calls** include the source-file
answer which turns the return expression into `exports.value`. The complete
[transcript](evidence/resolver.json) includes the input and host observations.

`make-driver.cjs` reads the actual `EmitResolver` and `EmitHost` interfaces through
stock TypeScript's checker, then generates a closed `.a` replay. It creates and
binds a fresh source file, constructs the real transformers, and calls the real
`emitFiles`. Recorded resolver answers are consumed in order, checked against
node kinds/positions and primitive arguments, and counted at completion.
Unrecorded resolver methods throw. The request pathname is read with Adamic's
file API and its bytes must equal the recorded request. The emitter's write
callback supplies text, BOM, output paths, and source-file identities; these
observations form the scout protocol records.

This is a fixed-request replay, not a general resolver serializer or compiler
CLI. It supports the recorded primitive/source-file answers; other answer
shapes are rejected by the generator. It does not compile a native checker.
The materialized root is retained as [replay.a.txt](evidence/replay.a.txt);
upstream source is fetched and gathered in scratch, not committed here.

The real stock emitter cache and the gathered-source replay on Node match
byte for byte, with empty stderr and exit 0:

- One project, three protocol records, including one emitted JS file.
- Golden SHA256 `e3e51e9be884f3637464149535c91c63ff088d6dc167780bebe6f246eb1319aa`.
- [Comparator result](evidence/node-comparison.json): success, no differences.
- Gathered closure: 3,381 code declarations in 62 declaration files, 3,472 copied
  spans, 79 evaluation modules. The upstream tool audits every copied byte and
  ordered import list. [Byte audit](evidence/slice-verify.log) passes.

The ordered native stops

The final measurement attempts the untouched replay with split 0 and 1, then
walks fifteen first stops in each mode: **32 failed driver build attempts**.
Every build exits 1 at the checker; both modes have identical diagnostics at
all fifteen positions. The order below is the compiler's actual diagnostic
order, not a sorted source inventory.

To expose the next first stop, `stub.cjs` replaces the smallest enclosing
function body with a throwing discovery placeholder, retaining an explicit or
checker-inferred signature. Fourteen such edits exist only in the walk clone.
These are semantic deletions for discovery, not validated adaptations. Removing
one body may hide other errors in that body; this walk is not an exhaustive
inventory of the closure. No modified walk tree supplies Node truth.

`map-stops.cjs` reverses previous edit offsets and maps through the audited
slice spans to the adapted upstream coordinates. It rejects a diagnostic in a
synthetic placeholder. [stops.json](evidence/stops.json) retains full messages,
function names, span hashes, witnesses, and owners; [discovery-edits.json](evidence/discovery-edits.json)
retains each removed body's hash, extent, and replacement without copying whole
upstream functions into the repository.

Owners below are routing recommendations to **@system_adamic**, not new
assignments: O = stage3 optional-declaration adaptation; I = stage3 checked
indexed reads/writes; H = Node host/library lane. The checker errors are
observations of stricter contracts, not assertions that the checker is wrong.

| Order | Adapted upstream site | Function / operation | First error | Owner | Witness |
|---:|---|---|---|---|---|
| 1 | moduleNameResolver.ts:2772:21 | loadModuleFromTargetExportOrImport, originalPath | TS2322 | O | 01-search-result |
| 2 | moduleNameResolver.ts:286:9 | createResolvedModuleWithFailedLookupLocations | TS2322 | O | 02-resolved-module |
| 3 | moduleNameResolver.ts:132:13 | withPackageId, peerDependencies | TS2375 | O | 03-package-id |
| 4 | moduleNameResolver.ts:474:36 | getPackageJsonTypesVersionsPaths, indexed paths | TS2322 | I | 04-version-paths |
| 5 | sourcemap.ts:216:58 | appendSourceMap, indexed source content | TS2345 | I | 05-source-content |
| 6 | sys.ts:1281:43 | watchPresentFileSystemEntry, unknown caught error | TS18046 | H | 06-watcher-error |
| 7 | sys.ts:1598:69 | tryEnableSourceMapsForHost, optional dependency | TS2307 | H | 07-source-map-support |
| 8 | sys.ts:1617:13 | require callback, unknown error result | TS2322 | H | 08-module-import-error |
| 9 | transformer.ts:367:9 | enableSubstitution, indexed compound write | TS2532 | I | 09-substitution |
| 10 | transformer.ts:395:9 | enableEmitNotification, indexed compound write | TS2532 | I | 10-emit-notification |
| 11 | transformer.ts:563:9 | endLexicalEnvironment, indexed stack restore | TS2322 | I | 11-lexical-stack |
| 12 | transformer.ts:614:9 | endBlockScope, indexed stack restore | TS2322 | I | 12-block-stack |
| 13 | transformers/classFields.ts:1835:21 | visitInNewClassLexicalEnvironment, first private member | TS2532 | I | 13-private-member |
| 14 | transformers/classFields.ts:2251:45 | transformConstructorBodyWorker, optional super index | TS2538 | I | 14-super-index |
| 15 | transformers/classFields.ts:2376:56 | transformConstructorBody, indexed original node | TS2345 | I | 15-original-node |

Each [minimal witness](witnesses/) reduces the named source operation, has its
required first-line a-check header, and runs independently on Node. All fifteen
match their recorded outputs. Each reproduces its named checker code on main
and the candidate, in both split modes. The source-map-support witness exercises
the disabled lazy path: its type import exposes the missing dependency, while
Node executes no optional module load. Enabled source-map-support is untested.
[Counts](counts.md) and both complete witness observation reports are retained.
No internal/oracle fixture registration or native allocation row was added.

Mutants actually run

| Mutant | Catcher |
|---|---|
| Each of fifteen witness inputs changes independently | Original Node stdout, successful mutant execution with empty stderr |
| Drop the first resolver call | Resolver order guard |
| Change a recorded argument span | Resolver argument guard |
| Append an unconsumed resolver answer | Transcript completeness guard |
| Change request source bytes | Exact request-byte guard |
| Change the replay's input initializer | Stock emitted stdout, only stdout differs |
| Drop callback source-file identities | Stock emitted stdout, only stdout differs |
| Flip callback BOM | Stock emitted stdout, only stdout differs |
| Flip actual emitSkipped | Stock emitted stdout, only stdout differs |
| Process stdout changes one byte | Scout comparator stdout check alone |
| Process stderr gains one byte | Scout comparator stderr check alone |
| Process exit changes to 1 | Scout comparator exit check alone |
| Falsify first stop | Captured-stderr recount |
| Claim native execution after failed builds | Native-status audit |
| Falsify adapted-source hash | Source-byte audit |
| Falsify Node golden hash | Stock golden-byte audit |
| Falsify witness Node output | Witness output recount |

The three process producers are explicitly test doubles. They do not occupy a
reported native slot. The fifteen witness mutants were repeated against main's
observation set. All **31 distinct mutants** are caught; no clang-warning kill
is counted. [Replay mutant results](evidence/replay-mutant-results.json) retain
which streams differ, and the audit mutant logs retain their named assertions.

Reproduction and commands run

Use the setup environment and stock API, a never-pushed candidate worktree,
the scout checkout at the specified commit, and a fresh apply.sh output. Locked
`@types/node` 25.3.3 must be installed in the candidate's `stage3/api`; native
builds run with that worktree as their working directory.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/step31-setup.log 2>&1
source /workspace/adamic-tools/env.sh
npm ci --prefix /workspace/step31-candidate/stage3/api --ignore-scripts --no-audit --no-fund
bash /workspace/step31-candidate/stage3/apply.sh /workspace/step31-adapted > /tmp/step31-apply.log 2>&1
export STEP31_TYPESCRIPT=/home/agent/.cache/adamic-stage3/api/node_modules/typescript/lib/typescript.js
bash stage3/scouts/step31/emitter-native/run.sh /workspace/step31-candidate \
    /workspace/step31-scout/stage3/scouts/step31 /workspace/step31-adapted \
    /workspace/step31-final-run > /tmp/step31-final-run.log 2>&1
python3 stage3/scouts/step31/emitter-native/test-witnesses.py /tmp/step31-main-adamic \
    /workspace/adamic /workspace/step31-main-witness-proof > /tmp/step31-main-witness.log 2>&1
python3 stage3/scouts/step31/emitter-native/audit.py /workspace/step31-adapted > /tmp/step31-audit.log 2>&1
```

`run.sh` builds the actual candidate compiler, captures resolver answers, gathers
and audits the real closure, compares Node output through the upstream scout
comparator, attempts both native modes, walks stops when blocked, then runs the
focused witnesses and replay mutants. Operations within this measurement run
sequentially. The `--mutant` audit modes are `stop`, `native-status`,
`source-hash`, `node-output`, and `witness-output`; each exits 1. The ordinary
audit exits 0. Test output is retained in log files, never piped.

Setup succeeded: Node ready 0.020s; Go 0.021s; submodules 0.058s; markdown
0.067s; clang 0.161s; Go build 48.022s; tests deferred 48.152s; cache warm
48.153s; done 48.179s. `nproc=5`, `cpu.max=400000 100000`, 17.6 GB,
Go 1.27.1, clang 20.1.8, Node 24.19.0. The environment file is
`/workspace/adamic-tools/env.sh`. An initial native invocation from the delivery
checkout lacked its pinned Node declarations; rerunning from the installed
candidate worktree resolved that infrastructure error before stop measurement.

Not covered: native execution, C split emission, runtime ownership/leaks, a
native resolver or checker, linking/Step 32, the full emitter project corpus,
full upstream suite, whole packages, full gate, or stops beyond fifteen. The
independent Node comparison is component evidence for this fixed request only.
