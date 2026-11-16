#!/usr/bin/env python3
"""Render the source trace and retain the earlier stopped report below it."""
import json
from pathlib import Path
here=Path(__file__).resolve().parent
x=json.loads((here/'source-trace.json').read_text())
def q(s): return '`'+s.replace('|','\\|')+'`'
def site(d): return f"{d['file']}:{d['line']}"
cells={
 'P':'L1 `program.ts:474`; L2 `watch.ts:755`; outside **P1–P8** (open); guard caveat below',
 'S':'L3 `watch.ts:844`; outside **S1–S2** (open)',
 'C':'L2 `watch.ts:755` (1 literal)',
 'D':'L4 `tsbuildPublic.ts:495` (1 literal)',
}
lines=[x['label'],'','Paths in the table are relative to `src/compiler/`. L1–L4 and each outside boundary are quoted below. “No” means the trace does not establish exactly one literal on both sides; it does not mean that the two runtime objects must differ.','',
'| Restore site and quoted line | Written object: allocation endpoints | Object originally read for saved value: allocation endpoints | One literal, same one |',
'|---|---|---|---|']
for r in x['rows']:
 s=r['site'];g=r['group']; verdict='**yes**' if r['one_and_same'] else '**no** — '+('2 syntactic literal sources + open outside' if g=='P' else '1 known literal + open outside')
 lines.append(f"| {q(s['file'].removeprefix('src/compiler/')+':'+str(s['line']))} {q(s['text'])} | {cells[g]} | {cells[g]} | {verdict} |")
lines += ['', '**Result: 15 rows; 7 source-level yes, 8 no.** The 14 host-member restores plus the auxiliary `state.readFileWithCache` restore are the same 15 candidates recorded earlier. No row is a head reach-analysis result. Open boundaries have no finite allocation count.', '',
'''## Limits, starting with the public API boundary

1. **A host parameter of an exported API is open to external callers.** Internal calls do not close that boundary. An external caller may supply any compatible object, including a literal, class instance or proxy whose allocation is outside these 79 roots. This trace stops at that parameter/field and records it as **outside**, rather than assuming that the CLI is the only caller. Source `export` boundaries are intentionally included even for `@internal` exports; they are not a measurement of which exports a particular lowered entry exposes.

   In the head, `internal/lower/view_flow.go:204–213` marks parameters of **entry-module exported functions** unknown regardless of internal call edges. `hostMemberReach` (`internal/lower/host_member_write.go:201`) begins `if reaches.Unknown { ... return stop }`. Its comment names “a closure's parameter, an opaque call's result, a host value.” It fails closed, returning the deferred original stop unchanged; it does not treat an external host as a known literal or as a zero-site proof. It also returns that stop for a nonliteral allocation or class instance. Whether this exact whole tsc input reaches that check is still unmeasured. The earlier two-literal control remains evidence that *multiple known literals alone* do not fail this head’s current layout-based restore rule.

2. **This is a hand-checked source trace, not a points-to proof or the head’s graph.** It follows the explicit assignments, destructuring, returns, fields, casts, call arguments and known callback forwarding in the frozen adapted input. It joins possible calls rather than proving path feasibility or liveness. Multiple executions of the same literal count as one allocation *site*. “Yes” is the source-level singleton/site-equality result, not a claim about native layout, method compatibility, temporal patch values, or the head’s acceptance.

   **Known program guard caveat:** `program.ts:4900`, `if (!host.useSourceOfProjectReferenceRedirect) return { onProgramCreateComplete: noop, fileExists };`, bypasses the restore closure when redirect support is false. L1’s unmodified default/CLI hosts do not define `useSourceOfProjectReferenceRedirect`, so those plain paths do not execute the three restores. L1 is retained in the table as a *syntactic receiver source*, not a witnessed active restore. `program.ts:1725` computes the guard from the host method; `watchPublic.ts:505` forwards that method from a possibly external watch host into L2. I did not solve guard/callback feasibility globally. Pruning the plain L1 path does not close P1–P8 or turn any program row into a source-level singleton proof.

3. **External callback implementations are unknown.** U1–U3 below stop target enumeration where a public callback is supplied. The known default builder and the CLI forwarding wrapper are traced. These callbacks consume a compiler host and return a program; they do not supply the locally allocated receiver at the watch or solution `compilerHost` restores. If an unknown callback reenters a public API with a different host, that host stops at the corresponding open API boundary. I have not analyzed callback code outside the 79 roots, reflective mutation, monkey-patched factories, or arbitrary computed writes hidden in external code. These are places this source account can under-count; infeasible call joins can over-count. No unexplored callback is silently assigned a literal.

## Allocation terminals

The trace stops at these four receiver allocations. It does not follow the methods’ underlying `System`/Node objects after a fresh wrapper literal is reached. No in-root class-instance allocation was found on the traced receiver paths; external boundaries can include class instances, and their sites cannot be enumerated here.
''']
for key,d in x['allocations'].items():
 lines += [f"- **{key} — {q(site(d))}**: {q(d['quote'])} {d['why']}"]
lines += ['', '''L1’s callers are `program.ts:390`, `watchPublic.ts:123`, and `executeCommandLine.ts:915`. L2’s internal callers are `watchPublic.ts:492` and `tsbuildPublic.ts:434`. L3’s internal callers are `watch.ts:872` and `tsbuildPublic.ts:295`. L4’s only caller is `tsbuildPublic.ts:2269`. Exported factory parameters can affect fields/functions, but these factories return their own literal, not the input host/System. A cast does not allocate. `copyProperties` (`core.ts:1510–1515`) writes into its first argument, so extending L3 with a watch-host object does not change its allocation identity.

## Outside terminals: quotes and reasons

These are boundaries, not additional counted literal sites. P1/P2 are the two overload routes into the same API; the remaining P entries identify upstream exported APIs. They are not eight disjoint sets of runtime objects.

| Boundary | Exact line | Why tracing stops |
|---|---|---|''']
for key,d in x['outside_boundaries'].items():
 lines.append(f"| **{key}** {q(site(d))} | {q(d['quote'])} | {d['why']} |")
lines += ['', '''The options-object host declaration for P2 is `types.ts:7769`, `host?: CompilerHost | undefined;`. For P7, `watchPublic.ts:140` exports `createIncrementalProgram` and destructures the options whose `host` member is declared at line 136. For P8, `watch.ts:978` declares `host?: CompilerHost;`. These fields retain the supplied object; passing through an options object does not make that options literal the host allocation.

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
|---|---|---|''']
for key,d in x['unknown_callback_targets'].items():
 lines.append(f"| **{key}** {q(site(d))} | {q(d['quote'])} | {d['why']} |")
lines += ['', '''## Evidence and validation

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
''', (here/'STOPPED.md').read_text()]
(here/'REPORT.md').write_text('\n'.join(lines))
