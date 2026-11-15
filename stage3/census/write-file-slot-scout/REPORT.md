| Road | Target gone? | New / changed findings against main | JavaScript change | Public declaration change |
| --- | --- | --- | --- | --- |
| (a) Wrap during Node construction; retain the harness helper | **No.** The helper still writes `sys.writeFile`. | **+1 / 0**, no removals; 54 → 55 unique sites | Two locals and the wrapper move into construction; literal selects it; initializer patch call removed. Wrapper `.name` changes. | **0 bytes** against main |
| (b) Exact writable slot view | **Yes in the production tsc entry**: advances to `1502:57`. The focused helper reproduction exposes `.call`. | **0 / 0**, no removals; 54 → 54 unique sites | One object alias; assignment selects that alias. | **0 bytes** against main |
| (c) Language rule for method-slot writes | **Not built or measured**, as requested | — | None submitted | No public property conversion proposed |

# Write-file-slot scout

Base: **fe5a1a54f045e9bce4e5aa32ca74885a317fbebd**. No adaptation, compiler change, public API sanction, or expected-count change is submitted. Both built roads are private source experiments, not amendments ready to land.

**Production and full mode:** the unmodified production tsc-entry check reproduces the requested `sys.ts:1387:5` refusal on main and (a). With (b), it advances to **NotYet at `sys.ts:1502:57`**, the clearTimeout slot-signature stop. The full census sees farther already on main: its 54 unique Refused/NotYet sites include the helper's `1392:49` inherited-library-`call` lowering refusal, but no 1387 finding. Thus (b) removes the **production** target without earning a spurious **full-census** removal. Neither road compiles the remaining frontier.

## Callers and readers

The entire `99-correctness-write-file-view` directory was read before making experiments: adapter, README, owner pins, contract, witness, mutant and proof. Its whole-owner pin is authoritative. Road (b), or any subsequent change inside this helper, would be an amendment there. The experiments do not alter that directory.

Upstream v6.0.3 has exactly two calls to `patchWriteFileEnsuringDirectory` across `src`:

| Caller | Timing and requirement |
| --- | --- |
| `compiler/sys.ts:1961` upstream; `:1982` on adapted main | The sys initializer calls it **after** `getNodeSystem()` returns and before publishing the initialized exported `sys`. `getNodeSystem` itself does not call it. |
| `testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:561`, `TestServerHost.createServerHost` | After `new TestServerHost(...)` and before returning the host. Its class method `writeFile` at upstream `:1155` throws if the parent directory is missing. This caller still needs the patch behavior. |

There is also an import in that test helper, the declaration/export itself, and the explanatory comment in `compiler/program.ts:424`; no other named reference or call was found. `createWatchedSystem` does **not** call the patch helper. Removing the helper just because the Node initializer no longer calls it would break the harness route. Road (a) retains that route, so its original assignment remains in the compiler-root source.

The syntax audit parsed every `.ts` file under upstream `src`. It retains **all 479** `writeFile` property accesses, conservatively including unrelated CompilerHost and test-host implementations: **455 calls, 15 value reads, 9 writes, zero direct identity comparisons, zero destructuring reads, zero string-literal element accesses**. The complete locations and contexts are in [readers.json.gz](evidence/readers.json.gz); the compact audit is [readers-summary.json](evidence/readers-summary.json). This is a source audit, not a claim about arbitrary external consumers or dynamic computed keys.

The important System routes and saved-value readers are:

| Reader / route, upstream locations | What it observes |
| --- | --- |
| `compiler/program.ts:477`, `createCompilerHostWorker` | Its IO-measuring writer calls `system.writeFile(...)` at invocation time. It does not capture the original slot value. |
| `compiler/watch.ts:860`, `createProgramHost` | Arrow forwards to the current `system.writeFile`. |
| `compiler/executeCommandLine.ts:1300` | Direct `sys.writeFile` call for generated config output. |
| `harness/harnessIO.ts:126`; `harness/fakesHosts.ts:327` | Live forwarding to `ts.sys.writeFile` / `this.sys.writeFile`. |
| `server/scriptInfo.ts:658`, `server/project.ts:782`, `server/session.ts:2628` | Host calls; `tsserver/nodeServer.ts` carries `sys` into `ServerHost`. No saved raw Node writer or construction-interval observation was found. |
| `testRunner/unittests/helpers/baseline.ts:460–469` | Captures the current writer, replaces it, and calls the capture with `.call(sys, ...)` while changing build-info versions. |
| `helpers/solutionBuilder.ts:64,78,99,104` | Saves and **restores the exact captured function** after build wrappers. These readers rely on later replacement/restoration, not on observing the Node constructor's raw writer. |
| `helpers/virtualFileSystemWithWatch.ts:1383–1386` | Saves and replaces the current writer to count written files, invoking its saved function with the host receiver. |
| `compiler/program.ts:529,582,599,628`; `watchPublic.ts:635,663`; `tsbuildPublic.ts:756,772,787` | CompilerHost writer caching, replacement and restoration downstream of System forwarding. Preserve these later operations. |
| `compiler/builder.ts:1727,1815,1914`; `watch.ts:764`; `watchUtilities.ts:133,221`; `tsbuildPublic.ts:1117`; `unittests/customTransforms.ts:24` | Binding, presence tests or callback selection for host writers. Full contexts are retained in the inventory. |

The other direct calls are enumerated in the inventory, including unit-test edits and baseline writers. Searches of the saved `originalWriteFile` aliases found calls/restoration, not identity tests. No `writeFile.name`, `.length` or `.prototype` reader was found. No reader was found that obtains `nodeSystem.writeFile` between object construction and the old initializer patch. That interval does not publish `nodeSystem`: construction completes, returns to the initializer, and the helper runs before exported `sys` receives its result.

**Identity and timing of (a):** capture the hoisted local `writeFile` after watch-function setup and before the `nodeSystem` literal. Allocate the three-argument arrow there, with deferred closures over `nodeSystem`; put it in the literal's `writeFile` property. Remove the initializer's later patch call. The object now starts with the wrapper rather than first storing the raw function and subsequently replacing it. The raw local function is never exposed in that slot. The wrapper's directory-method reads remain live; its original writer and receiver remain fixed, including after `setSys` or a later slot replacement. Its dynamic own `.call` lookup remains on every invocation. There is no extra hot-path call; the one initialization-time helper call disappears. The captured raw function is unchanged within an execution, while a newly allocated final wrapper remains distinct from it.

This spelling changes the final arrow's inferred `.name` from `""` to `"patchedWriteFile"`; its `.length` remains 3. It is **not** globally metadata-identical JavaScript. The [real-body execution witness](evidence/witness.json) records that difference separately from matching write events, BOM handling, directory checks/creation, receiver identity, two changing own `.call` overrides, and a later replacement of the public slot. The filesystem/watch/ensure helper dependencies are focused stubs; this is not a full lane run.

**Identity and timing of (b):** `writeFileSlot` is the same object, not a copy, binding or live-reader wrapper. The readonly original capture remains unchanged. The extra alias does no property read, call or allocation. The original getter is still read once and the replacement is written to the same object at the same point, with the same setter receiver. Final arrow name, argument count, captured identity and call events match main.

## Hidden stops and the two measurement paths

The unmodified main census overlay and driver were used. Main's production `methodValueRefusal` recognizes a plain assignment to a program interface method and defers its old refusal through `deferHostMemberWrite`. Production's `lowerWithElementKey` uses `hostMemberPending` to restore that saved refusal on an earlier failure. The census's separate per-unit path does not apply that production return wrapper. See [host_member_write.go](../../../internal/lower/host_member_write.go), [refusals.go](../../../internal/lower/refusals.go), [lower.go](../../../internal/lower/lower.go), and [census units](../latent/units.go.txt). This explains why a saved production slot-write stop must not be equated with a `refusal_scan` row in this stream.

The literal claim that this revision's census reports only one scan stop per unit also does not describe the receipts: `getNodeSystem` has **five** `refusal_scan` findings. The continuing scan exposes them already; no language rule was disabled to find them.

| Unit / first scan stop | Main | (a) construction | (b) writable view |
| --- | --- | --- | --- |
| `patchWriteFileEnsuringDirectory`, `refusal_scan` | **No finding** | **No finding** | **No finding**; in particular, no writable-method-field refusal |
| Helper's next **lowering**, not scan, refusal | `1392:49`, inherited library member `call` read as an own field | Same `1392:49` | `1393:49`, same refusal; position shifts by one line |
| `getNodeSystem`, first `refusal_scan` stop | `1501:57` | `1501:57` | `1502:57` |
| Additional construction-wrapper lowering refusal | — | **`1539:53`, inherited library member `call` read as an own field** | — |

The first getNodeSystem scan stop is **NotYet**: `node:timers.clearTimeout value needs exactly (timeoutId: unknown) => void, not (timeout: unknown) => void with no slot`. The next scan refusal is the `watchFile` relation: a function taking `PollingInterval` viewed as one taking `number | undefined`, at main `1548:13`, (a) `1558:13`, (b) `1549:13`. Then come the `watchDirectory` boolean/optional-boolean relation at `1549:13 / 1559:13 / 1550:13`, a namespace-object-as-value stop at `1604:192 / 1614:192 / 1605:192`, and the second clearTimeout stop at `1614:57 / 1624:57 / 1615:57`. All are paired unchanged by meaning. Other lowering findings, including byteOrderMarkIndicator, comparator-less sorts and the self-capturing object initializer, remain in the raw streams.

The actual compiler entry, checked with unmodified `adamic js <tree>/src/tsc/tsc.ts --module-format commonjs-bundle`, gives:

| Input | Production first stop | Exit / stdout |
| --- | --- | --- |
| Main | **Refused, `sys.ts:1387:5`**, method read as a value (`writeFile` would lose its object and `this`) | 1 / 0 bytes |
| (a) | Same **1387:5** refusal | 1 / 0 bytes |
| (b) | **NotYet, `sys.ts:1502:57`**, clearTimeout slot-signature stop quoted above | 1 / 0 bytes |

[entry-production.json](evidence/entry-production.json) retains the fresh exact-main commands and outputs. An additional `sys.ts`-as-entry spot check instead stops at `utilitiesPublic.ts:1284:52` on all three trees; that different entry is retained in [production.json](evidence/production.json) and is not substituted for the requested compiler frontier.

To discriminate the slot write itself, [slot-repro.json](evidence/slot-repro.json) uses the **exact extracted helper body** and the three **actual method signatures** from System, with only the filesystem helper replaced by a focused stub. Unmodified production main reports the original method-read-as-value refusal at the assignment. With (b), that refusal disappears and production reports the inherited `.call` refusal instead. Both exit 1 with zero output. This is evidence for the narrow road, not successful native compilation.

## Findings and declaration price

Every full run completed with `LATENT_FULL=1`, `LATENT_ASSERT_NO_OUTPUT=1`, `LATENT_ONLY_FILE` naming the whole private `sys.ts`, and `CENSUS_MODULE_FORMAT=CommonJSBundle`. Headers have **zero diagnostics**, `checker_rejected:false`, `latent_mode:full`; trailers are present. The tool's historical `measurement` label still says “measured on a checker-rejected program”; the actual header is authoritative. No usable IR/output guard was removed.

Pairing key: **(kind, phase, reason, message with position stripped)**, also canonicalizing the private tree prefix. Site findings are deduplicated before pairing, and multiplicity is retained so a second identical diagnostic at a new site cannot disappear into a set. [pairs.json.gz](evidence/pairs.json.gz) retains the complete contexts.

| Measurement | Main | (a) | (b) |
| --- | ---: | ---: | ---: |
| Unique Refused/NotYet sites | 54 | 55 | 54 |
| Unique site/phase findings | 54 | 55 | 54 |
| Raw Refused/NotYet observations, including repeated attempts | 87 | 89 | 87 |
| New unique findings vs main | — | 1 | 0 |
| Gone unique findings vs main | — | 0 | 0 |
| Changed findings after requested pairing | — | 0 | 0 |
| Stock tsc 6.0.3, explicit sorted 79 roots, inherited upstream compiler options | exit 0 | exit 0 | exit 0 |

The sole unpaired addition is (a)'s `1539:53`: **Refused / lowering / inherited library member call read as an own field**, with message “Adamic 0.1 refuses inherited library member call read as an own field; prototype members are not stored in an object's shape; call the method on its receiver, or wrap that call in an arrow (unbound-method)”. It is the new wrapper's `originalWriteFile.call(...)`, not a new language reason. There are no unpaired removals to classify as replacements. Counting **every record kind and every attempt**, main/(a)/(b) have **180 / 184 / 180** records and **60 / 60 / 60** distinct normalized tuples. (a) adds two observations of that one Refused site and two associated Boundary observations; neither tuple signature is new, and no existing tuple changes or disappears. (b) pairs every record, including multiplicity and Boundaries. [all-record-pairs.json](evidence/all-record-pairs.json) retains the complete multiset delta; the site-count table above counts compiler Refused/NotYet findings rather than continuation markers.

Road (b) is **not refused as a widened writable method field**. Main's `widenedProperties` rule checks whether the writable field could store something the original method cannot hold. Here source and target are the exact same `System["writeFile"]` type. Its same-type exception is documented/tested in [method_field_widening_test.go](../../../internal/lower/method_field_widening_test.go). This is one **new local view relation**, with no widening; it changes **zero existing public declaration relations**. It still fails later on `.call`.

Neither road converts System.writeFile's public method declaration into a function property. The complete System declaration has the same normalized SHA-256 on all sides (`d81501788e981b02ec97aab31f072333e4c47a39f403417dacbf89d30e68a49c`). Each separate stock-6.0.3 `stripInternal` declaration emit completed without emit diagnostics; upstream's unmodified declaration bundler then exited 0. All three bundled public `typescript.d.ts` files are byte-identical, SHA-256 **`f03adcf915df559d616fcafd75181d3423eaa0f272237482cedbd025525c9516`**. See [declarations.json](evidence/declarations.json) and the empty road-to-main public diffs in [diffs](diffs/).

The comparisons against upstream `tests/baselines/reference/api/typescript.d.ts` are also retained in `diffs/*-vs-upstream-api.diff.gz`. They contain main's existing API differences, with **no additional difference from either road**. Byte identity here is against adapted main, not an assertion that main equals the untouched upstream baseline.

Thus neither road spends Nov 1's named method-to-property sanction or its Nov 14 extensions; neither invokes the internal-only exception from sys.ts:594 (#trxxkge) for a public member. **Public members converted: 0; changed existing public relations: 0.** A conversion of the public System member would be a different, unmeasured road requiring its own declaration and relation accounting; no such conversion is made here.

The JavaScript comparison uses stock 6.0.3, ES2024/ESNext, with comments both kept and removed. Only `src/compiler/sys.ts` changes among source files. Both exact source patches and both emitted-JavaScript diffs are retained in [diffs](diffs/). (a) has only the construction insertion, selected literal property, and removed initializer patch block; (b) has only the extra alias and redirected assignment. StripInternal sys declarations also match exactly.

## Unbuilt language road

**Question:** may a method slot of a public program interface be assigned through an arbitrary System parameter while preserving its method declaration, and what receiver/allocation and stored-value proof is required? Main already has a narrow rule: `hostMemberWritable` defers a plain `=` write to a required program interface/type-literal method; `checkHostMemberWrites` subsequently requires known plain-object-literal receivers and compatible stored values/holders. Unknown allocation reach fails closed. A TestServerHost class instance is not a plain object literal. This rule is not a general authorization to replace any host or class method.

A broader receiver/allocation rule, interprocedural proof, or different boundary for the exported helper would need to answer that language question. No language implementation, widened exemption, method-to-property declaration change, or proposed adaptation is built/submitted for (c).

## Reproduction and limits

Private detached worktree: `/workspace/scratch/write-file-slot-scout`, from the exact base. `git submodule update --init --recursive --depth 1` completed: cohere `5360753e19e4417675a90e92a1823bea66b13918`, its TypeScript `d92d9bfee114c80be2c375d72edae966176e3a4f`. Upstream clone tag v6.0.3 was confirmed at **050880ce59e30b356b686bd3144efe24f875ebc8**. `npm ci` ran in the private stage3/api, then unmodified `stage3/apply.sh` completed successfully. The shared checkout's files were not adapted or edited.

`ADAMIC_TYPESCRIPT_SOURCE` names that clone. `TMPDIR`, `GOTMPDIR`, `GOCACHE`, and the stage3 build cache all name `/workspace/scratch/write-file-slot-measure` subdirectories. The census overlay/binary and ordinary production binary were built from the private unchanged base. No new Go file is authored under stage3. Upstream's additional npm install supplies its locked bundler dependencies (including its locked TypeScript 5.9.3); declaration emit and stock checks explicitly use the API's **6.0.3**.

[Scripts](scripts/) retain experiment construction, exact census flags, checks, diffs, reader audit, witnesses and pairing. Paths in these receipts are the actual scratch paths; rerunning elsewhere requires substituting those paths. The stock-check config explicitly resolves the inherited `node` type library from the pinned API installation. Both candidate sys.ts files preserve main's CRLF endings. [checks.json](evidence/checks.json) records exact source hashes, root counts, commands and stock exits; raw census streams, root lists, logs and declaration receipts are retained in [evidence](evidence/).

This is a scout, not a landing proof: no whole lane, amendment contract/replay, or successful native compilation of TypeScript is claimed. Road (a) does not remove the generic helper's write because of its real harness caller. Road (b) gets past that specific slot refusal in the production tsc entry and focused helper reproduction without spending a public declaration ruling, but still stops on inherited `.call`. That is the concrete next road to measure before proposing an amendment.


## Earlier branch evidence preserved

At publication the requested remote branch already contained scout commit `90a2ab30292aa5cf4dc6a3fb9ed5857a7d04b161`, based on `763d5fc8`. Its complete report/evidence snapshot is preserved byte for byte under [previous-763d5fc8](previous-763d5fc8/REPORT.md), and its commits remain in the branch ancestry. The report above and its receipts are fresh measurements from **fe5a1a54**, including the upstream bundled declaration comparison and wrapper-name witness. Publication merges the report histories and uses a normal fast-forward push, never force.
