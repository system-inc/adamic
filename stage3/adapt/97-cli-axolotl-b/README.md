# 97b-internal: internal host members say they hold undefined

Axolotl (#axolotl), task #qg1ry2t. This adaptation is byte-identical in JavaScript and changes no public API. It edits internal declarations only: 11 optional members in src/compiler/types.ts, tsbuildPublic.ts and program.ts.

lane: run as part of the Axolotl wave Correctness cuts it into; Loom's run of that wave is the lane's proof.

## The cause

tsc fills its hosts by copying optional members from one host to another:

- **Stored:** mostly `maybeBind(host, host.x)`, which is `fn?.bind(obj)` (core.ts:1522), or a saved `host.x` written back later. Both can be undefined.
- **Declared:** the receiving member is `x?(…): R`. Under `exactOptionalPropertyTypes` that admits a missing member or a defined one, never a present undefined.

So each store is a TS2412 (an assignment) or a TS2375 (an object literal). The inventory (scratchpad a97/inventory.md) found 28 such rows and the 34 declarations they need, 15 of them internal. This head takes 11 of the internal ones. The others go elsewhere:

- `FilePresenceUnknownOnHost.fileWatcher` (I10) is 97c's.
- `CompilerHostLikeForCache.writeFile` (I12) and `ProgramHost.createDirectory` and `writeFile` (I13, I14) go to 97b-public. They are cascades of its P18 and P19, and restoring any of them here brings nothing back, so this head can't pin them.

## The repair

Each declaration is widened to say what tsc stores, the ruled repair. An optional method becomes adaptation 32's indexed method type plus undefined, `x?: { x(…): R; }["x"] | undefined;`, which the tree already uses (program.ts:503, watchPublic.ts:274). It keeps method parameter bivariance; the plain arrow form breaks `WriteFileCallback` (inventory flag F1).

No member becomes required and none goes away. JSDoc and inline `/** @internal */` stay where they are; only the member's own text changes.

| Id | Declaration | Before | After |
|---|---|---|---|
| I1 | types.ts:8190-8195 `CompilerHost.resolveLibrary` | `resolveLibrary?(` … `): ResolvedModuleWithFailedLookupLocations;` | `resolveLibrary?: { resolveLibrary(` … `): ResolvedModuleWithFailedLookupLocations; }["resolveLibrary"] \| undefined;` (its four parameter lines unchanged) |
| I2 | types.ts:8209 `CompilerHost.useSourceOfProjectReferenceRedirect` | `useSourceOfProjectReferenceRedirect?(): boolean;` | `useSourceOfProjectReferenceRedirect?: { useSourceOfProjectReferenceRedirect(): boolean; }["useSourceOfProjectReferenceRedirect"] \| undefined;` |
| I3 | types.ts:8212 `CompilerHost.createDirectory` | `createDirectory?(directory: string): void;` | `createDirectory?: { createDirectory(directory: string): void; }["createDirectory"] \| undefined;` |
| I4 | types.ts:8631 `EmitHost.createHash` | `createHash?(data: string): string;` | `createHash?: { createHash(data: string): string; }["createHash"] \| undefined;` |
| I5 | types.ts:10001 `ModuleSpecifierResolutionHost.directoryExists` | `directoryExists?(path: string): boolean;` | `directoryExists?: { directoryExists(path: string): boolean; }["directoryExists"] \| undefined;` |
| I6 | types.ts:10003 `ModuleSpecifierResolutionHost.realpath` | `realpath?(path: string): string;` | `realpath?: { realpath(path: string): string; }["realpath"] \| undefined;` |
| I7 | types.ts:10007 `ModuleSpecifierResolutionHost.getGlobalTypingsCacheLocation` | `getGlobalTypingsCacheLocation?(): string \| undefined;` | `getGlobalTypingsCacheLocation?: { getGlobalTypingsCacheLocation(): string \| undefined; }["getGlobalTypingsCacheLocation"] \| undefined;` |
| I8 | types.ts:10019 `ModuleSpecifierResolutionHost.trace` | `trace?(s: string): void;` | `trace?: { trace(s: string): void; }["trace"] \| undefined;` |
| I9 | tsbuildPublic.ts:238 `SolutionBuilderHostBase.now` | `now?(): Date;` | `now?: { now(): Date; }["now"] \| undefined;` |
| I11 | program.ts:504 `CompilerHostLikeForCache.createDirectory` | `createDirectory?(directory: string): void;` | `createDirectory?: { createDirectory(directory: string): void; }["createDirectory"] \| undefined;` |
| I15 | tsbuildPublic.ts:205 `getCurrentTime(host: { now?(): Date; })` | `now?(): Date;` | `now?: { now(): Date; }["now"] \| undefined;` |

Why each is internal:

- I1, I2, I3 and I9 carry their own `@internal`.
- I4 to I8 and I11 sit in an `@internal` interface.
- I15 is the parameter type of a function tsbuildPublic.ts doesn't export.

I11 and I15 are cascade owners: they take hosts whose members this head widens, so they must admit the same undefined. I11 takes I3's `CompilerHost`, and I15 takes I9's `SolutionBuilderHostBase`.

## Why it's sound

- **Readers:** under `exactOptionalPropertyTypes`, reading `x?: T` and reading `x?: T | undefined` both give `T | undefined`, so no read changes type. The inventory checked every read of these members (flag F2). Each one is an optional call, a guard (`if (host.x)`, `||` with a fallback, `maybeBind`), or an indexed read like `CompilerHost["x"]`, which already includes undefined.
- **Writers:** the only stores the widening newly admits are the undefined values tsc already stores today. Defined values were admitted before and still are.
- **Assignability:** a widened host passed into a narrower optional shape would be a new TS2379. Inside src/compiler the two cascade owners above are those shapes, and the check finds nothing new.
- **Runtime:** the edits are type-only. Each file's JavaScript is byte-identical.
- **Public API:** every member is stripped from the declarations. `transpileDeclaration` with `stripInternal` gives byte-identical output for all three files before and after.

## Rows

Six of the inventory's 28 rows clear:

| Row | Site | Declarations |
|---|---|---|
| D044 | program.ts:1876:11 TS2375 | I5, I6, I7 |
| D045 | program.ts:2599:9 TS2375 | I4, I6, I7, I8 |
| D124 | tsbuildPublic.ts:302:5 TS2412 | I9 |
| D127 | tsbuildPublic.ts:440:5 TS2412 | I1 |
| D137 | watchPublic.ts:506:5 TS2412 | I2 |
| D145 | watchPublic.ts:662:9 TS2412 | I3 |

The other 22 wait:

- **For 97b-public (21):** D043, D049, D050, D051, D125, D126, D128, D129, D130, D131, D132, D133, D134, D135, D136, D138, D139, D140, D141, D142 and D144. Their messages are unchanged, including the member D043, D051, D134 and D135 name.
- **For 97c (1):** D143 (I10).

## Mutant pins

Each mutant restores one member of the adapted tree to its before text and reruns the check. Each of the 11 brings something back on its own, so no pair is needed. A TS2375 names the first bad member in its literal, and every returning one names the restored member:

| Id | Restored alone, it brings back |
|---|---|
| I1 | D127, tsbuildPublic.ts:440:5 TS2412 |
| I2 | D137, watchPublic.ts:506:5 TS2412 |
| I3 | D145, watchPublic.ts:662:9 TS2412 |
| I4 | D045, program.ts:2599:9 TS2375, naming `createHash` |
| I5 | D044, program.ts:1876:11 TS2375, naming `directoryExists` |
| I6 | D044 and D045, both naming `realpath` |
| I7 | D044 and D045, both naming `getGlobalTypingsCacheLocation` |
| I8 | D045, naming `trace` |
| I9 | D124, tsbuildPublic.ts:302:5 TS2412 |
| I11 | three TS2379s, `CompilerHost` into `CompilerHostLikeForCache` (createDirectory), at executeCommandLine.ts:919:38 and watchPublic.ts:128:38 and 637:46 |
| I15 | four TS2379s into `{ now?(): Date; }` at tsbuildPublic.ts:1099:44, 1128:122, 1832:76 and 1852:63 |

## Proof

- **The shape:** the adapter parses with stock TypeScript 6.0.3 and finds each owner by its AST:
  - the one top-level exported interface of that name in the file;
  - or, for I15, the one local `getCurrentTime` whose single `host` parameter is a one-member type literal.

  Within the owner it requires exactly one member of that name. The member's text must be exactly the inventoried before or after, with a multi-line member spelled in the file's own line breaks. It also checks that each member or owner is still `@internal`, and that `getCurrentTime` is still unexported.
- **Refusals:** each was probed on a scratch copy, and each exits 1 with no write:
  - types.ts with only I4 adapted (a partly adapted file);
  - types.ts adapted and the other two files not (a partly adapted tree);
  - `EmitHost` renamed (a missing owner);
  - I8 in the arrow form `trace?: ((s: string) => void) | undefined;` (a shape not inventoried);
  - `ModuleSpecifierResolutionHost` without its `@internal`;
  - `getCurrentTime` exported;
  - a second `CompilerHostLikeForCache.createDirectory`;
  - program.ts with one bare LF (mixed line breaks).

  A copy of the adapter that appends a statement to program.ts, with its byte-count check removed, is refused by the JavaScript check: "program.ts's JavaScript changes".
- **Planning:** the adapter plans every file before writing any. For each file it re-reads its own result as fully adapted, checks that only the members' bytes moved, and checks the JavaScript. It re-reads each file just before writing, then writes each changed file once.
- **Source:** on a fresh copy of the adapted tree, the diff is exactly the 11 members. That is 12 lines in three files, 12 insertions and 12 deletions: types.ts has 9 (I1 is two of them), tsbuildPublic.ts 2 and program.ts 1. It matches the in-memory widening the inventory measured (scratchpad a97/inv/widen.cjs) for those members.
- **JavaScript:** `ts.transpileModule` before and after, with an ES2020 target, is byte-identical under both CommonJS and ESNext. Under CommonJS the sizes are types.ts 143,226 bytes, tsbuildPublic.ts 87,806 and program.ts 240,346.
- **Idempotence:** a second run edits nothing and leaves the SHA-256s unchanged. An input that is already adapted is accepted with no write.
- **Diagnostics:** measured with the inventory's in-memory checker (stock TS 6.0.3, all of src/compiler, strict, exactOptionalPropertyTypes, noUncheckedIndexedAccess):
  - 141 before and 135 after;
  - the six removed are exactly the six rows above;
  - nothing new appears anywhere, and no remaining diagnostic's text changes. The sorted full-text logs differ in those six lines alone.
- **Golem, `adamic types`** on the adapted tree at main 7b9c0fe7: whole program 144 to 138, tsbuildPublic.ts own-file 10 to 8, watchPublic.ts 10 to 8, program.ts 11 to 9. The six removed are exactly program.ts 1876:11 and 2599:9, tsbuildPublic.ts 302:5 and 440:5, and watchPublic.ts 506:5 and 662:9, with nothing new. A second run edits nothing.

Not yet run: the sharded lane oracle. It runs in Loom as part of the wave, per Correctness' rule on #axochk.
