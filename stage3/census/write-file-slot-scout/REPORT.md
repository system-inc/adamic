| Road | Target gone? | New and changed findings against main | JavaScript change | Public declaration change |
|---|---|---|---|---|
| (a) Wrap during Node construction; retain the patch helper for its other caller | **No**: production still refuses `sys.ts:1387:5` | Full census: **one new refusal site**, represented twice: +2 Refused records and +2 associated Boundary records; 0 removed, 0 changed; no new normalized signature | Yes: capture/wrapper before the literal, initialize `writeFile` with it, remove the initializer's patch call | None: all 79 emitted declaration files byte-identical |
| (b) Exact writable slot view | **Yes**: production advances to `getNodeSystem`, `sys.ts:1502:57` | Full census: **0 new, 0 removed, 0 changed** after position stripping | Yes: one alias statement and redirected assignment through that same object | None: all 79 emitted declaration files byte-identical |
| (c) Extend the language's method-slot write rule | Not built | Not measured; language question below | Not built | Not built |

Scout only, on **763d5fc83558ae8ceb0202189b18430dff499ab8**. No adaptation, compiler rule, public sanction, or expected baseline is submitted. Road (b) is the smallest measured way past the stated production target, but it does not compile the remaining frontier. A future adaptation must amend **99-correctness-write-file-view**, whose complete directory was read before candidate work. Its `owners.json` pins the whole patch function; adding a separate overlapping adapter would violate that ownership.

Callers and readers

The following references use pristine TypeScript v6.0.3 line numbers; measurement positions below use adapted source.

- `src/compiler/sys.ts:1954–1964`: the exported `sys` initializer calls `getNodeSystem()`, then calls `patchWriteFileEnsuringDirectory(sys)` if a system was returned. **getNodeSystem itself does not call the helper.** It constructs the literal and returns it; the initializer performs the subsequent patch.
- `src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:553–563`: `TestServerHost.createServerHost` constructs a **class instance**, adjusts its parameters, then patches it before returning. This is the only other direct call found by a whole-`src` search. `createWatchedSystem` does not make that call. `getCreateWatchedSystem` selects between the two constructors. Removing the helper or making it a no-op changes server-host behavior.
- `src/compiler/program.ts:477`: `createCompilerHostWorker` forwards through `(path, data, bom) => system.writeFile(path, data, bom)`. `createWriteFileMeasuringIO` at 413 explicitly documents redundant directory creation when this system has been patched. This is a **live slot read on each invocation**, not a capture of the original function.
- `src/compiler/watch.ts:860` likewise forwards to the current system slot. `src/compiler/executeCommandLine.ts:1300` calls `sys.writeFile` for generated configuration. `src/harness/harnessIO.ts:126` forwards to `ts.sys.writeFile`; `src/harness/fakesHosts.ts:327` calls `this.sys.writeFile`. Numerous compiler/watch/build tests call their supplied system slot to edit files.
- The Node server takes `ts.sys` as a `ServerHost` (`src/tsserver/nodeServer.ts:111`, then passes it into the server); the typings installer passes `sys` as its install host (`src/typingsInstaller/nodeTypingsInstaller.ts:93`). Server readers call host slots, e.g. `src/server/scriptInfo.ts:658`, `project.ts:782`, and `session.ts:2628`. No second direct patch-helper call occurs there.
- **Captured identity does matter in helper composition:** `helpers/baseline.ts:460–470` captures and wraps `sys.writeFile`, forwarding with `.call(sys, ...)`; `helpers/virtualFileSystemWithWatch.ts:1383–1390` does the same for written-file tracking. `helpers/solutionBuilder.ts:64–78` and `99–104` capture and later **restore the exact saved function**. These are positive examples of replacement after construction. CompilerHost-specific captures/replacements (`program.ts:529/582`, builder's `maybeBind`, and build caches) are a separate host layer and should not be mistaken for direct reads of global `sys`.

No comparison of `writeFile` function identity was found in a whole-`src` search for equality/inequality involving that member. A stock-parser AST walk also found zero equality/inequality expressions containing a `writeFile` spelling (case-insensitive), and confirmed exactly the two direct calls above; see `reader-ast-check.json`. This is a source-search result, not a claim about external API consumers. Tests do rely on later replacement and restoration; the test server construction specifically relies on patching an already constructed object. The patch helper is exported from the internal source namespace even though it is absent from the bundled public `typescript.d.ts` baseline. `System.writeFile` itself is a public method signature.

Hidden stops and the two measurement modes

The **production** entry census reproduces the question exactly on main: `Refused`, `sys.ts:1387:5`, `a method read as a value (writeFile would lose its object, and this with it)`. Both inspections made by the production census tool agree. After (a), the identical stop remains. After (b), both inspections instead report `NotYet`, `sys.ts:1502:57`:

> node:timers.clearTimeout value needs exactly (timeoutId: unknown) => void, not (timeout: unknown) => void with no slot

The requested **full single-file latent census** sees farther even on unchanged main. Its refusal clone defers eligible interface-method stores, and its resumed lowering reaches `.call` before the deferred-store postcheck. Consequently **the original 1387 refusal is not a finding in the main full stream**. It would be misleading to claim its removal as a full-census delta. The relevant `checkHostMemberWrites` pass-ledger entry says `not_reached_lowering_stop`.

| Input | Next refusal-scan stop in patch helper | Hidden lowering stop in patch helper | First refusal-scan stop in getNodeSystem |
|---|---|---|---|
| Main | None recorded | `1392:49`, Refused: inherited library member `call` read as an own field | `1501:57`, clearTimeout signature/slot NotYet above |
| (a) | None recorded; production still stops at the retained method store | Same `1392:49` | Same `1501:57`; inserted wrapper comes later |
| (b) | None recorded | `1393:49`, same `.call` refusal | `1502:57`, same clearTimeout stop |
| Store-removal diagnostic probe | None recorded | `1392:49`, same `.call` refusal | `1501:57`, unchanged |

The diagnostic probe changes only the store into a typed unused local holding the same arrow. It is **not a road or a proposed adaptation**: it no longer patches the system. Its full stream checks that removing the store exposes no additional refusal-scan stop in that owner; the `.call` failure remains a lowering failure.

The full stream actually records **five** refusal-scan findings in `getNodeSystem`, so it supplies more than one first-stop answer for this unit. Main positions are `1501:57` (clearTimeout), `1548:13` (`watchFile`: PollingInterval viewed as number | undefined), `1549:13` (`watchDirectory`: boolean viewed as boolean | undefined), `1604:192` (namespace object used as a value), and `1614:57` (clearTimeout again). In (a), the first position stays fixed and the other four move +10 lines. In (b), all five move +1 line. Their kind/phase/reason/message signatures are unchanged. The raw streams preserve the remaining file findings, including the separate `getEnvironmentVariable` method read outside these owners.

Full pairing totals (all record kinds included):

| Input | Units | Raw records | Distinct normalized tuples | Added records | Removed / changed |
|---|---:|---:|---:|---:|---:|
| Main | 136 | 180 | 60 | — | — |
| (a) | 136 | 184 | 60 | 4 | 0 / 0 |
| (b) | 136 | 180 | 60 | 0 | 0 / 0 |
| Diagnostic store-removal probe | 136 | 180 | 60 | 0 | 0 / 0 |

Road (a)'s two unmatched tuple groups are:

- `(Refused, lowering, inherited library member call read as an own field, M)`, multiplicity **2**.
- `(Boundary, lowering, failed statement; state rolled back; continued at next statement or declaration, M)`, multiplicity **2**.

For both groups, stripped message **M** is `Adamic 0.1 refuses inherited library member call read as an own field; prototype members are not stored in an object's shape; call the method on its receiver, or wrap that call in an arrow (unbound-method)`. The new source site is `(a) sys.ts:1539:53`, attempted once through the enclosing initializer and once through `getNodeSystem`. Neither group is a new signature: main already reports that same signature at the retained helper's `.call`. No existing tuple is removed or rewritten. Road (b) and the diagnostic probe pair every record with main, including multiplicities.

Exact roads, timing, and identity

**(a)** captures the existing hoisted local `writeFile` in `originalWriteFile`, makes a typed arrow **before** the `nodeSystem` literal, initializes that literal's `writeFile` property to the arrow, and removes only the `sys` initializer's later patch call. The arrow still calls `writeFileEnsuringDirectories`; its write callback uses `originalWriteFile.call(nodeSystem, ...)`, and its directory callbacks read `nodeSystem` at invocation. The declaration reference is captured before initialization but is not evaluated by constructing the arrow. Stock tsc accepts it. Exact source and JavaScript patches are in evidence.

Main temporarily stores the original local function in the literal, returns the object, then reads that property once and overwrites it with a newly allocated wrapper. Road (a) creates the wrapper earlier, before evaluation of literal initializers such as `process.argv.slice(2)` and `_os.EOL`; the object never holds the unwrapped function in that slot. The original local function identity is the same one main would place in and read from this literal. The published slot still holds one fresh wrapper per constructed system, not the original function. `.call` is still looked up on the captured function at invocation, and both receivers are the same Node system object. Directory member reads remain live. The named original function itself does not use `this`.

There is no source-visible reader between this literal's return and the original initializer's patch call. That supports the Node-only construction move, but does **not** justify deleting the exported mutating helper for arbitrary systems: its getter/setter observations, exact capture, repeated patching and class-instance caller would be lost. Hence the measured construction road retains that helper and cannot remove the production target. A more sweeping construction redesign would need to handle the harness class and other permitted helper consumers, and would be a different road. This scout stops that road here.

**(b)** inserts exactly:

```ts
const writeFileSlot: { writeFile: System["writeFile"] } = sys;
writeFileSlot.writeFile = (path, data, writeBom) => /* existing body */;
```

It retains the landed readonly view and `originalWriteFile` capture. The new alias is the **same object**, with no allocation, getter read, function wrapper, bind operation or extra call. The assignment has the same receiver and value as main and therefore the same getter/capture/setter order and later captured-function identity. Only an ordinary local alias statement precedes the store. Existing `.call` behavior remains intact in JavaScript; the full census's native representation refusal is a separate obstacle.

The anticipated writable-field refusal does **not** occur. `internal/lower/invariance.go`'s `widenedProperties` method-to-writable-field check rejects when the target field's type is not assignable back to the source method type (including its enum check). Here the target is exactly that source indexed-access type. This is not the unsafe wider `() => void` view of a `() => never` method described in that rule's comment. The measured positive result should not be generalized to widened writable views.

**(c), language question only:** when may a plain assignment to a public interface method slot be admitted with correct dispatch, aliasing and write compatibility? Main already has a bounded rule: `hostMemberWritable` defers eligible program-interface method stores; `checkHostMemberWrites` requires known plain object-literal reach, an own function field, no spread/accessor/class representation, and compatibility with **every holder type**. Unknown reach returns the original refusal. A public `System` parameter can also receive the harness class instance and objects with accessors, so treating every such assignment as an ordinary known-literal store requires a language/representation decision, not a syntax exemption. Reading a method value and writing its slot must remain separate questions. No compiler rule or spelling workaround was built for (c).

Measurement contract and receipts

- Fresh detached worktree at the exact requested commit under `/workspace/scratch/write-file-slot-scout`; `git submodule update --init --recursive --depth 1` completed. Separate upstream clone at **050880ce59e30b356b686bd3144efe24f875ebc8**, tag v6.0.3. `npm ci` in the private worktree's `stage3/api`, followed by unmodified `stage3/apply.sh`, passed. All candidates were made from that completed private adapted tree. Shared source was not adapted.
- `ADAMIC_TYPESCRIPT_SOURCE` names that upstream clone. `GOTMPDIR`, `TMPDIR`, and `GOCACHE` name `/workspace/scratch/write-file-slot-run` subdirectories during Go builds; measurement temporary/cache paths are likewise scratch paths. No new Go file under stage3 is submitted.
- Full census binary built from unchanged main with `stage3/census/latent/make_overlay.py`, then `go build -buildvcs=false -overlay=... ./stage3/census/latent/tool`. Each run loads all **79 sorted compiler roots**, selects the **whole** exact `src/compiler/sys.ts` via `LATENT_ONLY_FILE`, and sets `LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 CENSUS_MODULE_FORMAT=CommonJSBundle`. No unit/line filter, speculative mode, permissive loader or usable IR output is enabled. All completed headers have zero diagnostics and `checker_rejected:false`; the historical generic measurement label does not change those facts. These are measurements, not native compilation success.
- Finding pairing uses a **multiset** of `(kind, phase, reason, message with position stripped)`. Positions, unit paths, and line shifts are not keys. Boundary records are retained and identified separately. Raw streams and `summary.json` preserve every occurrence; normalized signature identity alone must not hide the duplicate attempts at a new site.
- Stock TypeScript **6.0.3 CLI** checks all 79 roots, with strict, exact optional properties, unchecked indexed access, verbatim modules, ES2024/Bundler, and pinned Node 25.3.3 declarations: **exit 0 on main, (a), (b)**. Actual CLI declaration emits with explicit rootDir, `emitDeclarationOnly` and `stripInternal` also exit 0: **79/79 files byte-identical** for each road. This compares per-module emitted declarations, not a rebuilt Hereby API bundle.
- JavaScript comparison transpiles **604 source files** with ES2024/ESNext in both comments-kept and comments-removed modes: **2,416 comparison pairs / 4,832 emits** across two roads and two modes. Only `src/compiler/sys.ts` differs. The exact emitted changes are attached. No whole-lane oracle was run or claimed for a scout that submits no adaptation.

The source patches, normalized pairing output, completed full streams, production entry stops, source hashes, successful type/declaration receipts, JavaScript comparison inventory, and measurement scripts are under [evidence](evidence). All full census streams include the final relation-report trailer, so a partial run cannot stand in for a completed measurement.
