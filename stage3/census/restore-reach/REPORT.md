a source-level trace, not 949a2198's reach analysis; it can over- or under-count where lowering's graph differs, and it names each place it can't follow.

Paths in the table are relative to `src/compiler/`. L1–L4 and each outside boundary are quoted below. “No” means the trace does not establish exactly one literal on both sides; it does not mean that the two runtime objects must differ.

| Restore site and quoted line | Written object: allocation endpoints | Object originally read for saved value: allocation endpoints | One literal, same one |
|---|---|---|---|
| `program.ts:4957` `host.compilerHost.fileExists = originalFileExists;` | L1 `program.ts:474`; L2 `watch.ts:755`; outside **P1–P8** (open); guard caveat below | L1 `program.ts:474`; L2 `watch.ts:755`; outside **P1–P8** (open); guard caveat below | **no** — 2 syntactic literal sources + open outside |
| `program.ts:4958` `host.compilerHost.directoryExists = originalDirectoryExists;` | L1 `program.ts:474`; L2 `watch.ts:755`; outside **P1–P8** (open); guard caveat below | L1 `program.ts:474`; L2 `watch.ts:755`; outside **P1–P8** (open); guard caveat below | **no** — 2 syntactic literal sources + open outside |
| `program.ts:4959` `host.compilerHost.getDirectories = originalGetDirectories;` | L1 `program.ts:474`; L2 `watch.ts:755`; outside **P1–P8** (open); guard caveat below | L1 `program.ts:474`; L2 `watch.ts:755`; outside **P1–P8** (open); guard caveat below | **no** — 2 syntactic literal sources + open outside |
| `tsbuildPublic.ts:782` `host.readFile = cache.originalReadFile;` | L3 `watch.ts:844`; outside **S1–S2** (open) | L3 `watch.ts:844`; outside **S1–S2** (open) | **no** — 1 known literal + open outside |
| `tsbuildPublic.ts:783` `host.fileExists = cache.originalFileExists;` | L3 `watch.ts:844`; outside **S1–S2** (open) | L3 `watch.ts:844`; outside **S1–S2** (open) | **no** — 1 known literal + open outside |
| `tsbuildPublic.ts:784` `host.directoryExists = cache.originalDirectoryExists;` | L3 `watch.ts:844`; outside **S1–S2** (open) | L3 `watch.ts:844`; outside **S1–S2** (open) | **no** — 1 known literal + open outside |
| `tsbuildPublic.ts:785` `host.createDirectory = cache.originalCreateDirectory;` | L3 `watch.ts:844`; outside **S1–S2** (open) | L3 `watch.ts:844`; outside **S1–S2** (open) | **no** — 1 known literal + open outside |
| `tsbuildPublic.ts:786` `host.writeFile = cache.originalWriteFile;` | L3 `watch.ts:844`; outside **S1–S2** (open) | L3 `watch.ts:844`; outside **S1–S2** (open) | **no** — 1 known literal + open outside |
| `tsbuildPublic.ts:787` `compilerHost.getSourceFile = cache.originalGetSourceFile;` | L2 `watch.ts:755` (1 literal) | L2 `watch.ts:755` (1 literal) | **yes** |
| `tsbuildPublic.ts:788` `state.readFileWithCache = cache.originalReadFileWithCache;` | L4 `tsbuildPublic.ts:495` (1 literal) | L4 `tsbuildPublic.ts:495` (1 literal) | **yes** |
| `watchPublic.ts:658` `compilerHost.readFile = originalReadFile;` | L2 `watch.ts:755` (1 literal) | L2 `watch.ts:755` (1 literal) | **yes** |
| `watchPublic.ts:659` `compilerHost.fileExists = originalFileExists;` | L2 `watch.ts:755` (1 literal) | L2 `watch.ts:755` (1 literal) | **yes** |
| `watchPublic.ts:660` `compilerHost.directoryExists = originalDirectoryExists;` | L2 `watch.ts:755` (1 literal) | L2 `watch.ts:755` (1 literal) | **yes** |
| `watchPublic.ts:661` `compilerHost.createDirectory = originalCreateDirectory;` | L2 `watch.ts:755` (1 literal) | L2 `watch.ts:755` (1 literal) | **yes** |
| `watchPublic.ts:662` `compilerHost.writeFile = originalWriteFile!;` | L2 `watch.ts:755` (1 literal) | L2 `watch.ts:755` (1 literal) | **yes** |

**Result: 15 rows; 7 source-level yes, 8 no.** The 14 host-member restores plus the auxiliary `state.readFileWithCache` restore are the same 15 candidates recorded earlier. No row is a head reach-analysis result. Open boundaries have no finite allocation count.

## Limits, starting with the public API boundary

1. **A host parameter of an exported API is open to external callers.** Internal calls do not close that boundary. An external caller may supply any compatible object, including a literal, class instance or proxy whose allocation is outside these 79 roots. This trace stops at that parameter/field and records it as **outside**, rather than assuming that the CLI is the only caller. Source `export` boundaries are intentionally included even for `@internal` exports; they are not a measurement of which exports a particular lowered entry exposes.

   In the head, `internal/lower/view_flow.go:204–213` marks parameters of **entry-module exported functions** unknown regardless of internal call edges. `hostMemberReach` (`internal/lower/host_member_write.go:201`) begins `if reaches.Unknown { ... return stop }`. Its comment names “a closure's parameter, an opaque call's result, a host value.” It fails closed, returning the deferred original stop unchanged; it does not treat an external host as a known literal or as a zero-site proof. It also returns that stop for a nonliteral allocation or class instance. Whether this exact whole tsc input reaches that check is still unmeasured. The earlier two-literal control remains evidence that *multiple known literals alone* do not fail this head’s current layout-based restore rule.

2. **This is a hand-checked source trace, not a points-to proof or the head’s graph.** It follows the explicit assignments, destructuring, returns, fields, casts, call arguments and known callback forwarding in the frozen adapted input. It joins possible calls rather than proving path feasibility or liveness. Multiple executions of the same literal count as one allocation *site*. “Yes” is the source-level singleton/site-equality result, not a claim about native layout, method compatibility, temporal patch values, or the head’s acceptance.

   **Known program guard caveat:** `program.ts:4900`, `if (!host.useSourceOfProjectReferenceRedirect) return { onProgramCreateComplete: noop, fileExists };`, bypasses the restore closure when redirect support is false. L1’s unmodified default/CLI hosts do not define `useSourceOfProjectReferenceRedirect`, so those plain paths do not execute the three restores. L1 is retained in the table as a *syntactic receiver source*, not a witnessed active restore. `program.ts:1725` computes the guard from the host method; `watchPublic.ts:505` forwards that method from a possibly external watch host into L2. I did not solve guard/callback feasibility globally. Pruning the plain L1 path does not close P1–P8 or turn any program row into a source-level singleton proof.

3. **External callback implementations are unknown.** U1–U3 below stop target enumeration where a public callback is supplied. The known default builder and the CLI forwarding wrapper are traced. These callbacks consume a compiler host and return a program; they do not supply the locally allocated receiver at the watch or solution `compilerHost` restores. If an unknown callback reenters a public API with a different host, that host stops at the corresponding open API boundary. I have not analyzed callback code outside the 79 roots, reflective mutation, monkey-patched factories, or arbitrary computed writes hidden in external code. These are places this source account can under-count; infeasible call joins can over-count. No unexplored callback is silently assigned a literal.

## Allocation terminals

The trace stops at these four receiver allocations. It does not follow the methods’ underlying `System`/Node objects after a fresh wrapper literal is reached. No in-root class-instance allocation was found on the traced receiver paths; external boundaries can include class instances, and their sites cannot be enumerated here.

- **L1 — `src/compiler/program.ts:474`**: `const compilerHost: CompilerHost = {` createCompilerHostWorker allocates and returns compilerHost; System-backed methods do not change the receiver allocation.
- **L2 — `src/compiler/watch.ts:755`**: `const compilerHost: CompilerHost = {` createCompilerHostFromProgramHost allocates a new compilerHost; copied or bound methods do not make the input host its receiver.
- **L3 — `src/compiler/watch.ts:844`**: `return {` createProgramHost returns this literal. Watch and solution-host factories extend this same object.
- **L4 — `src/compiler/tsbuildPublic.ts:495`**: `const state: SolutionBuilderState<T> = {` createSolutionBuilderState allocates state and returns it at line 551.

L1’s callers are `program.ts:390`, `watchPublic.ts:123`, and `executeCommandLine.ts:915`. L2’s internal callers are `watchPublic.ts:492` and `tsbuildPublic.ts:434`. L3’s internal callers are `watch.ts:872` and `tsbuildPublic.ts:295`. L4’s only caller is `tsbuildPublic.ts:2269`. Exported factory parameters can affect fields/functions, but these factories return their own literal, not the input host/System. A cast does not allocate. `copyProperties` (`core.ts:1510–1515`) writes into its first argument, so extending L3 with a watch-host object does not change its allocation identity.

## Outside terminals: quotes and reasons

These are boundaries, not additional counted literal sites. P1/P2 are the two overload routes into the same API; the remaining P entries identify upstream exported APIs. They are not eight disjoint sets of runtime objects.

| Boundary | Exact line | Why tracing stops |
|---|---|---|
| **P1** `src/compiler/program.ts:1517` | `export function createProgram(_rootNamesOrOptions: readonly string[] \| CreateProgramOptions, _options?: CompilerOptions, _host?: CompilerHost, _oldProgram?: Program, _configFileParsingDiagnostics?: readonly Diagnostic[]): Program {` | Public positional createProgram: external _host; createCreateProgramOptions forwards it to options.host. |
| **P2** `src/compiler/program.ts:1501` | `export function createProgram(createProgramOptions: CreateProgramOptions): Program;` | Public options-object createProgram: external createProgramOptions.host (types.ts:7769). |
| **P3** `src/compiler/builder.ts:1607` | `export function getBuilderCreationParameters(newProgramOrRootNames: Program \| readonly string[] \| undefined, hostOrOptions: BuilderProgramHost \| CompilerOptions \| undefined, oldProgramOrHost?: BuilderProgram \| CompilerHost, configFileParsingDiagnosticsOrOldProgram?: readonly Diagnostic[] \| BuilderProgram, configFileParsingDiagnostics?: readonly Diagnostic[], projectReferences?: readonly ProjectReference[]): BuilderCreationParameters {` | Exported @internal getBuilderCreationParameters: external oldProgramOrHost in the root-name-array branch can supply the host. |
| **P4** `src/compiler/builderPublic.ts:179` | `oldProgramOrHost?: CompilerHost \| SemanticDiagnosticsBuilderProgram,` | Public createSemanticDiagnosticsBuilderProgram implementation third argument; root-name-array overload host at line 171. |
| **P5** `src/compiler/builderPublic.ts:218` | `oldProgramOrHost?: CompilerHost \| EmitAndSemanticDiagnosticsBuilderProgram,` | Public createEmitAndSemanticDiagnosticsBuilderProgram implementation third argument; root-name-array overload host at line 210. |
| **P6** `src/compiler/builderPublic.ts:256` | `oldProgramOrHost?: CompilerHost \| BuilderProgram,` | Public createAbstractBuilder implementation third argument; root-name-array overload host at line 248. |
| **P7** `src/compiler/watchPublic.ts:136` | `host?: CompilerHost;` | Public createIncrementalProgram input.host, destructured at 140-147; if absent, default is L1. |
| **P8** `src/compiler/watch.ts:986` | `export function performIncrementalCompilation(input: IncrementalCompilationOptions): ExitStatus {` | Exported @internal performIncrementalCompilation input.host; declaration watch.ts:978, then forwarded through createIncrementalProgram(input). |
| **S1** `src/compiler/tsbuildPublic.ts:339` | `export function createSolutionBuilder<T extends BuilderProgram>(host: SolutionBuilderHost<T>, rootNames: readonly string[], defaultOptions: BuildOptions): SolutionBuilder<T> {` | Public createSolutionBuilder host parameter; no allocation is imposed on caller-supplied object. |
| **S2** `src/compiler/tsbuildPublic.ts:343` | `export function createSolutionBuilderWithWatch<T extends BuilderProgram>(host: SolutionBuilderWithWatchHost<T>, rootNames: readonly string[], defaultOptions: BuildOptions, baseWatchOptions?: WatchOptions): SolutionBuilder<T> {` | Public createSolutionBuilderWithWatch host parameter; no allocation is imposed on caller-supplied object. |

The options-object host declaration for P2 is `types.ts:7769`, `host?: CompilerHost | undefined;`. For P7, `watchPublic.ts:140` exports `createIncrementalProgram` and destructures the options whose `host` member is declared at line 136. For P8, `watch.ts:978` declares `host?: CompilerHost;`. These fields retain the supplied object; passing through an options object does not make that options literal the host allocation.

## Hand-checked paths and every caller on those paths

### Program restores: `program.ts:4957–4959` → L1, L2, or P1–P8

Written object: `host.compilerHost`. Saved reads: `program.ts:4895–4897`, respectively `const originalFileExists = host.compilerHost.fileExists;`, `const originalDirectoryExists = host.compilerHost.directoryExists;`, and `const originalGetDirectories = host.compilerHost.getDirectories;`.

The helper `updateHostForUseSourceOfProjectReferenceRedirect` has exactly one in-root caller: `program.ts:1727`. The argument is a wrapper literal with `compilerHost: host` at 1728. Follow that field: **the wrapper is not the receiver allocation**. The same captured wrapper/field is used by its returned `onProgramCreateComplete` closure. `createProgram`’s `const host = createProgramOptionsHost || createCompilerHost(options);` at 1579 gives:

- Default host → `createCompilerHost` at `program.ts:389`, return of `createCompilerHostWorker` at 390 → **L1**.
- Supplied host → destructured `host: createProgramOptionsHost` at `program.ts:1519`. The options branch uses the caller’s options field (**P2**); the positional branch at 1518 calls `createCreateProgramOptions`, which returns the same `host` field at 1484 (**P1**). Neither options carrier is the host.

The checker-resolved **internal callers of the actual `program.ts` function** are exactly `builder.ts:1620` and `executeCommandLine.ts:929`. The similarly named `tsbuildPublic.ts:1171` call resolves to the nested build-step function, not to this function.

- CLI `executeCommandLine.ts:929` passes `programOptions`, whose `host` field is at 926; its local `host` was created at 915 → **L1**.
- `builder.ts:1620` passes `{ host: oldProgramOrHost as CompilerHost, ... }` (field at 1623). `getBuilderCreationParameters` itself is exported (**P3**). Its three internal callers are `builderPublic.ts:186`, 225 and 261. Each forwards `oldProgramOrHost`, through the semantic, emit-and-semantic, or abstract public builder implementation (**P4**, **P5**, **P6**). Only the root-name-array branch creates a new program; the undefined-root and existing-Program branches do not create an additional receiver.
- The only in-root function-value uses of the emit-and-semantic builder are its defaults at `watch.ts:861` and `watchPublic.ts:149`; the semantic and abstract builders have no other in-root nondeclaration references. Follow those defaults to the callback call sites described next. All three also remain public entry boundaries.

The callback/field routes feeding the default builder’s third (`host`) argument are:

- `watchPublic.ts:151`, `createIncrementalProgram`: `host = host || createIncrementalCompilerHost(options)` at 148. Supplied field → **P7**; default → `watchPublic.ts:123`, worker → **L1**. Its only in-root caller is `watch.ts:989`, `createIncrementalProgram(input)`. That function is exported (**P8**) and at 988 uses `input.host` or writes the incremental-host factory result into it. Its only in-root caller is `executeCommandLine.ts:951`, importing it as `ts_performIncrementalCompilation`; CLI `host` at 949 also comes from the incremental factory → **L1**. The same-named private CLI function is a separate symbol; its callers do not introduce another host source.
- `watchPublic.ts:642` and 696: `createProgram` was destructured from input host at 455. Both calls pass the local `compilerHost` made at 492 → **L2**. The undefined-root call at 642 reuses an old program in the default builder branch. The root-name call at 696 can reach the program restores with L2.
- `tsbuildPublic.ts:1048`: `host.createProgram(..., compilerHost, ...)`, with third argument at 1051. `state.compilerHost` comes from its factory at 434 → **L2**.
- `executeCommandLine.ts:981–989`, `updateCreateProgram`, saves `compileUsingBuilder = host.createProgram` at 982 and installs an arrow at 983. Its `return compileUsingBuilder(rootNames, options, host, ...)` at 988 forwards the third argument unchanged. The wrapper is installed for the CLI solution and watch hosts; it does not manufacture another host. Its original callback is the default above on those CLI paths. User-supplied alternatives stop at U1–U3.

### Solution `host` restores: `tsbuildPublic.ts:782–786` → L3 or S1–S2

Written receiver is `state.host`, destructured at 780. The saved values have **that same source receiver**, not `cache`, `state`, or `state.compilerHost`:

`enableCache(state)` destructures `{ compilerHost, host }` at 745; calls `changeCompilerHostLikeToUseCache(host, ...)` at 758–762; that helper saves `host.readFile`, `host.fileExists`, `host.directoryExists`, `host.createDirectory`, `host.writeFile` at `program.ts:527–531`; its returned record at 625 carries those bindings; the destructuring at `tsbuildPublic.ts:750–758` and `state.cache = { ... }` at 766 preserve them. `disableCache` reads the corresponding cache fields and writes to the same state’s `host`.

The checker finds exactly four in-root cache-helper calls: `executeCommandLine.ts:919`, `watchPublic.ts:127`, `tsbuildPublic.ts:758`, and `watchPublic.ts:636`. The first two patch L1 during CLI/incremental-host setup; the last two supply the saved values restored by these solution/watch rows. I retain the call context when following the returned original fields. Merging all four helper calls would incorrectly add L1 and external solution hosts to the watch rows.

`state.host` is initialized by shorthand at `tsbuildPublic.ts:496`. Its `host` is `hostOrHostWithWatch as SolutionBuilderHost<T>` at 429. `createSolutionBuilderState`’s only caller is `createSolutionBuilderWorker` at 2269. The worker’s only callers are the public APIs at 340 and 344, forwarding their parameter unchanged → **S1**, **S2**.

Their only in-root callers are `executeCommandLine.ts:895` and 878. The former’s `buildHost` comes from `createSolutionBuilderHost` at 885; the latter’s comes from `createSolutionBuilderWithWatchHost` at 855. Those factories use `createSolutionBuilderHostBase` at `tsbuildPublic.ts:312` and 324; the base at 295 calls `createProgramHost` → **L3**. Both return the same base object (314/327). The watch version adds watch operations with `copyProperties` at 326. There is no second solution-host literal here.

### Compiler-host restores: watch’s five rows and `tsbuildPublic.ts:787` → L2 only

For watch, the five `compilerHost.* = original*` lines at `watchPublic.ts:658–662` share the `const compilerHost = createCompilerHostFromProgramHost(...) as CompilerHost & ResolutionCacheHost` at 492 → **L2**. The cache helper call at 636 receives exactly that local. The helper returns the saved originals, destructured at 629–636. Saved-source and written-object receiver therefore both end at **L2**, including `writeFile`’s non-null assertion, which does not change object identity.

The public `createWatchProgram(host)` input at `watchPublic.ts:419` is open, but is only an **input** to this wrapper factory. Its two in-root callers are `executeCommandLine.ts:1031` and 1051. They obtain their watch hosts from the config and file-list factories at 1020/1042. Those are also returned by the public `watchPublic.ts:368` overload implementation, delegating at 370/382 to `watch.ts:922`/954. The two internal factory bodies call private `watch.ts:createWatchCompilerHost` at 933/964, which gets **L3** at 872, extends it, and returns it at 891. Neither the open input host nor this L3 is the freshly created compilerHost’s identity. Stop at **L2** for the restore.

For the solution builder, `state.compilerHost` is initialized at `tsbuildPublic.ts:519` from the call at 434 → **L2**. `enableCache` reads `const originalGetSourceFile = compilerHost.getSourceFile;` at 748 and carries it into `state.cache.originalGetSourceFile` at 773; line 787 writes it to the same field’s object. `setGetSourceFileAsHashVersioned` may replace the member, but never replaces the compilerHost object. Context matters: the original method can already be a wrapper; the source *object* is still L2.

### Auxiliary state-property restore: `tsbuildPublic.ts:788` → L4 only

`const originalReadFileWithCache = state.readFileWithCache;` at 747 reads from **state itself**. The saved binding becomes `state.cache.originalReadFileWithCache` at 772; line 788 writes back to `state.readFileWithCache`. The value may be a closure; its function body’s use of `host.readFile` does not change which object that saved property read came from. State is the literal **L4**, returned at 551 to the worker at 2269. The returned public builder methods capture that state; they do not accept a replacement state from their caller.

The state chain was checked recursively across the private call graph: 61 functions have a `SolutionBuilderState` parameter; the index contains **228 resolved calls** into those parameters. Of those arguments, 227 are the lexical identifier `state`. The remaining call is the timer alias at `tsbuildPublic.ts:2074`: `callback(t1, t2, t3)`, where 2069 binds `callback = buildNextInvalidatedProject` and 2072 binds `t2 = state`. Its scheduling arrow captures `t2`; the external timer cannot substitute its own state argument. Recursive cycles therefore add no state allocation source.

In particular, `enableCache` has two callers, `setupInitialBuild` at 816 and `invalidateProject` at 2050. `disableCache` has three: `enableCache` at 742, `buildWorker` at 1973 and `buildNextInvalidatedProjectWorker` at 2115. Those paths lead through builder methods or captured timer/watch closures back to the worker’s state. All selected function-value references were checked; the timer alias is the only non-call use of these state-parameter functions. The source has no assignment to `state.host` or `state.compilerHost`, and no computed `state[...]` write or `copyProperties(state, ...)` on this path. The only cache-field writes are at 766/793; the two `state.readFileWithCache` assignments are at 763/788. Cache enable/disable cycles retain state and host identities.

## Explicit unknown stops

These are unknown *callback targets*, not extra counted receiver allocations. They limit the interprocedural trace; the default and CLI forwarding targets above were followed.

| Stop | Quoted line | Why |
|---|---|---|
| **U1** `src/compiler/watchPublic.ts:151` | `return createProgram(rootNames, options, host, oldProgram, configFileParsingDiagnostics, projectReferences);` | Unknown external CreateProgram callback target when supplied in IncrementalProgramOptions; known default is builderPublic.ts:215. |
| **U2** `src/compiler/watchPublic.ts:696` | `builderProgram = createProgram(rootFileNames, compilerOptions, compilerHost, builderProgram, configFileParsingDiagnostics, projectReferences);` | Unknown external host.createProgram callback target, captured at line 455; another call at line 642. Known default and CLI wrapper are followed in report. |
| **U3** `src/compiler/tsbuildPublic.ts:1048` | `program = host.createProgram(` | Unknown external host.createProgram callback target; known default and CLI wrapper are followed in report. |

## Evidence and validation

The input is the already completed adapted tree, unchanged from the stopped run. `source_trace.py` verifies all 79 source hashes against `provenance.json` before extracting any quotes. All locations above refer to that adapted tree, not pristine TypeScript. `source-index.cjs` uses stock TypeScript 6.0.3’s checker for symbol/alias resolution, overload signatures, function references, field declarations and call arguments; symbol keys include exact declaration positions, so the private build-step `createProgram` and imported public function are not conflated. It indexes evidence only and does not compute flow sets. The trace and the conclusions are hand-checked, encoded explicitly in `source_trace.py`/`source-trace.json`.

`source-index.json` retains the call/reference evidence, including the 61 state functions. `source-trace.json` retains quoted path steps, all endpoints, the 228 state-call edges and one row for each restore. All 228 direct `state`/`t2` argument passes resolve into the indexed private state functions; none is an unexamined direct pass of the raw state to an external callback. The state argument set, function privacy, non-call references, literal first lines and counts are checked when rendering; these checks validate the recorded evidence, not correctness of arbitrary source-level alias reasoning. The prior direct graph controls remain passing; no compiler or adaptation was changed or rerun for this source trace.

Reproduce the evidence in the prepared private worktree (with the adapted input path substituted):

```sh
node stage3/census/restore-reach/source-index.cjs "$ADAPTED" > stage3/census/restore-reach/source-index.json
python3 stage3/census/restore-reach/source_trace.py "$ADAPTED"
python3 stage3/census/restore-reach/source_report.py
```

## Earlier stopped measurement (preserved below)

The following is the previous report verbatim. Its “Unmeasured” table records the **head reach-analysis attempt**, not the new source trace above. `STOPPED.md` preserves those original bytes separately. The whole-program direct reach instrument is still not built.

---

| Restore site and quoted line | Written object allocations | Saved value source object allocations | One literal, same one |
|---|---|---|---|
| `src/compiler/program.ts:4957` — `host.compilerHost.fileExists = originalFileExists;` | Unmeasured | Unmeasured | Unmeasured |
| `src/compiler/program.ts:4958` — `host.compilerHost.directoryExists = originalDirectoryExists;` | Unmeasured | Unmeasured | Unmeasured |
| `src/compiler/program.ts:4959` — `host.compilerHost.getDirectories = originalGetDirectories;` | Unmeasured | Unmeasured | Unmeasured |
| `src/compiler/tsbuildPublic.ts:782` — `host.readFile = cache.originalReadFile;` | Unmeasured | Unmeasured | Unmeasured |
| `src/compiler/tsbuildPublic.ts:783` — `host.fileExists = cache.originalFileExists;` | Unmeasured | Unmeasured | Unmeasured |
| `src/compiler/tsbuildPublic.ts:784` — `host.directoryExists = cache.originalDirectoryExists;` | Unmeasured | Unmeasured | Unmeasured |
| `src/compiler/tsbuildPublic.ts:785` — `host.createDirectory = cache.originalCreateDirectory;` | Unmeasured | Unmeasured | Unmeasured |
| `src/compiler/tsbuildPublic.ts:786` — `host.writeFile = cache.originalWriteFile;` | Unmeasured | Unmeasured | Unmeasured |
| `src/compiler/tsbuildPublic.ts:787` — `compilerHost.getSourceFile = cache.originalGetSourceFile;` | Unmeasured | Unmeasured | Unmeasured |
| `src/compiler/tsbuildPublic.ts:788` — `state.readFileWithCache = cache.originalReadFileWithCache;` | Unmeasured | Unmeasured | Unmeasured |
| `src/compiler/watchPublic.ts:658` — `compilerHost.readFile = originalReadFile;` | Unmeasured | Unmeasured | Unmeasured |
| `src/compiler/watchPublic.ts:659` — `compilerHost.fileExists = originalFileExists;` | Unmeasured | Unmeasured | Unmeasured |
| `src/compiler/watchPublic.ts:660` — `compilerHost.directoryExists = originalDirectoryExists;` | Unmeasured | Unmeasured | Unmeasured |
| `src/compiler/watchPublic.ts:661` — `compilerHost.createDirectory = originalCreateDirectory;` | Unmeasured | Unmeasured | Unmeasured |
| `src/compiler/watchPublic.ts:662` — `compilerHost.writeFile = originalWriteFile!;` | Unmeasured | Unmeasured | Unmeasured |

# Stopped measurement: #fj0vqtf, Lower’s ground

**This is not the requested allocation census. No tsc restore has an allocation verdict.**
“Unmeasured” means the instrument never queried that receiver on a complete whole-program graph;
it does **not** mean the head returned `Unknown`, and it does not imply “no” or zero allocations.
The table retains all discovered restore candidates, including callable-property siblings.

## Controls

The scratch-only `probe_test.go.txt` fully lowers two small TypeScript programs using this exact
head, assigns allocation sites, builds `newAllocationFlowGraph`, and calls its unchanged
`ReachingAllocations` directly. Both the save and restore name the same module-level `host` local.
The programs are embedded in the probe; they differ only in `const host: Host = a` versus
`const host: Host = flag ? a : b`, with `flag = Date.now() > 0`, two arrow-bearing object literals,
`const original = host.read`, and `host.read = original`.

| Control | Written receiver | Saved-read receiver | Requested one-and-same predicate |
|---|---|---|---|
| One literal | `Sites:[-1], Unknown:false` | Same local, same query | **yes** |
| Choice of two literals | `Sites:[-2 -1], Unknown:false` | Same local, same query | **no, 2** |

`-1` is the `a` literal (control source line 2, `const a: Host = {read: (): string => 'a'};`);
`-2` is `b` (line 3, `const b: Host = {read: (): string => 'b'};`). These numbers are graph-local,
not source identities. `controls.log` records the passing run. These controls test a cardinality
predicate over the head’s actual reach results, not a new compiler rule.

Both restore programs **lower successfully on 949a2198**. This matters: this head’s
`hostMemberValueRestore` (`internal/lower/host_member_value.go:254`) asks the graph for both sets,
checks unknown reach and known closures, then applies `hostMemberLayout` over the combined sites.
It does not require a singleton or equality of the two sets. `hostMemberReach`
(`internal/lower/host_member_write.go:201`) likewise permits multiple plain literals. The proposed
one-and-same ruling is a separate measurement target, not this head’s existing acceptance rule.

## Method and stopping reason

The requested complete whole-program instrument was **not built**. The preserved probe only
attempts ordinary loading/lowering for the whole roots and makes direct reach queries for the
controls. It does not bypass earlier walls or reach tsc’s restore bodies. Thus the central task
remains undone; the report stops here rather than presenting a source-flow approximation as the
head’s own answer.

The reach API consumes `ir.Expression` and an `ir.Program`, not checked AST nodes:
`shape_flow.go:22` delegates to the shared allocation graph, whose sources, calls, returns,
closure parameters, liveness, stores and projections are collected from IR (`view_flow.go:114`).
The normal pipeline can stop before complete bodies or receiver expressions exist. Querying a
partial graph, dropping unsupported statements, or inventing AST-to-IR edges would not establish
what this head would judge on a complete program. I did not implement or validate a faithful
scratch construction of that complete IR. This is an implementation limit of this attempt,
not a claim that the requested scratch instrument is impossible, and not a request for permission.

The final whole-root attempt is retained in `whole-program.log`. No per-site allocation query was
made on that input. The final adapted program loads, but `Lower` returns
`lower: stage 0 compiles a program from one entry file, got 79` and no IR. This is only the public
entry guard, not an exploration of the later walls or a direct reach census. It is evidence of where
this attempt stopped, not a substitute for the direct
whole-program census the task explicitly requested.

## Site discovery

`host-sites.txt` was not found in the checkout; the `ahra` CLI was unavailable. I used the allowed
AST derivation route with the pinned stock TypeScript API in `stage3/api`. `inventory.cjs` traverses
every assignment in the compiler roots, follows direct variable initializers and aliases to member
reads, and selects interface/type-literal method targets. It additionally keeps callable targets
whose RHS is named `original…` or `cache.original…`, to expose restores of values returned by a
helper or stored in a cache. This latter selection is a naming heuristic, not a proof of saved-value
provenance or an exhaustive interprocedural derivation. No reach inference occurs in this script.

I manually checked the cache provenance: `changeCompilerHostLikeToUseCache` in `program.ts` saves
`host.readFile`, `host.fileExists`, `host.directoryExists`, `host.createDirectory`, and `host.writeFile`,
returns them as `original…` fields, and `watchPublic.ts` destructures that result before restoring.
`tsbuildPublic.ts` stores the helper’s returned originals in `state.cache`, along with
`compilerHost.getSourceFile`, and restores them in `disableCache`. `program.ts`’s three direct restores
read their originals from `host.compilerHost`. These are syntactic provenance observations only;
they do not identify allocation sites of those host objects.

`sites.json` contains the exact roots, quoted assignment lines and raw inventory metadata.
Its `saved` field can be an intermediate cache-field read, not the original host-method read;
it must not be interpreted as the object whose allocation reach the task asks to measure.
Rows marked `declaredMethod:false` are callable-property siblings in the **adapted** tree, retained
for visibility; existing adaptations have converted some optional methods to properties. The
`state.readFileWithCache` row is an auxiliary state-property restore, not a host-interface method.
This discovery has not been proven exhaustive for arbitrarily renamed values returned through
other functions or fields. There is no claim that this is a complete census.

## Setup, provenance and limits

Base: `949a2198c5a869d0a1598949add6708cb7949880`, fetched from
`compiler/host-member-value-3`; its complete change was reviewed (the `.a` and `.ts` fixture twins
were also checked byte-identical). A fresh detached worktree was created at
`/workspace/scratch/restore-reach`, separate from `/workspace/adamic`.
`git submodule update --init --recursive --depth 1` completed, and `npm ci` ran in `stage3/api`.
`ADAMIC_TYPESCRIPT_SOURCE` named a separate checkout of TypeScript v6.0.3 at
`050880ce59e30b356b686bd3144efe24f875ebc8`. `stage3/apply.sh` generated the adapted input;
no adaptation was added or edited. TMPDIR, GOTMPDIR and GOCACHE were under
`/workspace/scratch/restore-reach-runtime`. Go was 1.27.1.

The only instrument installed outside this report directory was a temporary test in the private
worktree’s `internal/lower`, copied from `probe_test.go.txt` and removed after the run. No compiler
implementation was changed. All preserved Go files under stage3 end in `.go.txt`.
`reproduce.sh` installs that test temporarily and requires an already prepared private worktree
and completed adapted tree. The report branch adds only `stage3/census/restore-reach` artifacts.

A useful continuation must construct and validate the whole-program IR used by this exact graph,
retain both receiver expressions and literal-to-source locations, and query every site directly.
Until then, allocation locations, literal quotations, counts and tsc yes/no decisions remain absent.


Inventory: 79 compiler roots; 15 restore candidates, of which 6 have an interface/type-literal method declaration in the stock API.
