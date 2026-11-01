# 97d: public host methods say they hold undefined

**Evidence:** every `proofs/` path below lives on the branch `cli/axolotl-97d-evidence` (f34705d4), at `stage3/adapt/97-cli-axolotl-d/proofs/`, not in this directory, because every file here feeds apply's input key (Correctness, #axochk).

Axolotl (#axolotl), task #qg1ry2t. This adaptation is byte-identical in JavaScript. It is a declared public API change: 19 public optional methods become optional properties in typescript.d.ts, under Adamic's Nov 1 ruling on #axochk. It also carries 3 internal cascade owners. It edits src/compiler/types.ts, watchPublic.ts, tsbuildPublic.ts and program.ts, and runs after 97b, whose 11 members it requires.

lane: run as part of the Axolotl wave Correctness cuts it into; Loom's run of that wave is the lane's proof.

## The cause

tsc fills its hosts by copying optional members from one host to another:

- **Stored:** mostly `maybeBind(host, host.x)`, which is `fn?.bind(obj)` (core.ts:1522), or a saved `host.x` written back later. Both can be undefined.
- **Declared:** the receiving member is `x?(…): R`. Under `exactOptionalPropertyTypes` that admits a missing member or a defined one, never a present undefined. And an optional method has no syntax for `| undefined`.

So each store is a TS2412 (an assignment) or a TS2375 (an object literal). The inventory (scratchpad a97/inventory.md) found 28 such rows. 97b cleared six with internal members, and 97c cleared D143. The other 21 need the 19 public members here.

## The ruling

Adamic, Nov 1, on #axochk: "the 19 public members (P1 to P19) convert from method to property in adaptation 32's idiom, `x?: { x(…): R; }["x"] | undefined`. That supersedes the Oct 19 rule for these 19 by name and nothing else."

The Oct 19 rule is adaptation 20's "No method-to-property conversion is sanctioned". Its README now records this exception for exactly these 19 by name, and the rule still governs every other optional method. Ten of the 19 (P1 to P4, P6, P7, P13, P14, P16 and P17) are the members adaptation 32 widened with this same idiom and then restored as methods (`public-host-sites.json`, kind `restored-public-method`).

## The repair

Each member becomes `x?: { x(…): R; }["x"] | undefined;`, the idiom already in the tree (program.ts:503, watchPublic.ts:274):

- The indexed method type keeps method parameter bivariance. The plain arrow form breaks `WriteFileCallback` with a new TS2322 at tsbuildPublic.ts:787 (inventory flag F1; adaptation 20's README records the same at services.ts:1769).
- I12 is a property and gains ` | undefined`.
- JSDoc stays where it is, and P11 and P12 keep their source line breaks inside the braces. Only the member's own text changes.

| Id | Declaration | Rows | Before | After |
|---|---|---|---|---|
| P1 | types.ts:7999 `ModuleResolutionHost.trace` | D134 | `trace?(s: string): void;` | `trace?: { trace(s: string): void; }["trace"] \| undefined;` |
| P2 | types.ts:8000 `ModuleResolutionHost.directoryExists` | D049, D051, D134, D144 | `directoryExists?(directoryName: string): boolean;` | `directoryExists?: { directoryExists(directoryName: string): boolean; }["directoryExists"] \| undefined;` |
| P3 | types.ts:8005 `ModuleResolutionHost.realpath` | D043, D051, D134 | `realpath?(path: string): string;` | `realpath?: { realpath(path: string): string; }["realpath"] \| undefined;` |
| P4 | types.ts:8007 `ModuleResolutionHost.getDirectories` | D050, D051, D134 | `getDirectories?(path: string): string[];` | `getDirectories?: { getDirectories(path: string): string[]; }["getDirectories"] \| undefined;` |
| P5 | types.ts:4552 `ParseConfigHost.trace` | D051 | `trace?(s: string): void;` | `trace?: { trace(s: string): void; }["trace"] \| undefined;` |
| P6 | types.ts:8146 `CompilerHost.getDefaultLibLocation` | D134 | `getDefaultLibLocation?(): string;` | `getDefaultLibLocation?: { getDefaultLibLocation(): string; }["getDefaultLibLocation"] \| undefined;` |
| P7 | types.ts:8152 `CompilerHost.readDirectory` | D134 | `readDirectory?(rootDir: string, …, depth?: number): string[];` | `readDirectory?: { readDirectory(…same…): string[]; }["readDirectory"] \| undefined;` |
| P8 | types.ts:8162 `CompilerHost.resolveModuleNames` | D128, D139 | `resolveModuleNames?(moduleNames: string[], …): (ResolvedModule \| undefined)[];` | `resolveModuleNames?: { resolveModuleNames(…same…): (ResolvedModule \| undefined)[]; }["resolveModuleNames"] \| undefined;` |
| P9 | types.ts:8166 `CompilerHost.getModuleResolutionCache` | D130, D142 | `getModuleResolutionCache?(): ModuleResolutionCache \| undefined;` | `getModuleResolutionCache?: { getModuleResolutionCache(): ModuleResolutionCache \| undefined; }["getModuleResolutionCache"] \| undefined;` |
| P10 | types.ts:8172 `CompilerHost.resolveTypeReferenceDirectives` | D129, D141 | `resolveTypeReferenceDirectives?(…): (ResolvedTypeReferenceDirective \| undefined)[];` | `resolveTypeReferenceDirectives?: { resolveTypeReferenceDirectives(…same…): …; }["resolveTypeReferenceDirectives"] \| undefined;` |
| P11 | types.ts:8173-8180 `CompilerHost.resolveModuleNameLiterals` | D125, D138 | `resolveModuleNameLiterals?(` … `): readonly ResolvedModuleWithFailedLookupLocations[];` | `resolveModuleNameLiterals?: { resolveModuleNameLiterals(` … `): …; }["resolveModuleNameLiterals"] \| undefined;` |
| P12 | types.ts:8181-8188 `CompilerHost.resolveTypeReferenceDirectiveReferences` | D126, D140 | `resolveTypeReferenceDirectiveReferences?<T extends FileReference \| string>(` … `): …;` | `resolveTypeReferenceDirectiveReferences?: { resolveTypeReferenceDirectiveReferences<T extends FileReference \| string>(` … `): …; }["resolveTypeReferenceDirectiveReferences"] \| undefined;` |
| P13 | types.ts:8207 `CompilerHost.createHash` | D043, D134, D136 | `createHash?(data: string): string;` | `createHash?: { createHash(data: string): string; }["createHash"] \| undefined;` |
| P14 | watchPublic.ts:186 `ProgramHost.createHash` | D135 | `createHash?(data: string): string;` | `createHash?: { createHash(data: string): string; }["createHash"] \| undefined;` |
| P15 | watchPublic.ts:200 `ProgramHost.directoryExists` | D131 | `directoryExists?(path: string): boolean;` | `directoryExists?: { directoryExists(path: string): boolean; }["directoryExists"] \| undefined;` |
| P16 | watchPublic.ts:207 `ProgramHost.realpath` | D135 | `realpath?(path: string): string;` | `realpath?: { realpath(path: string): string; }["realpath"] \| undefined;` |
| P17 | watchPublic.ts:211 `ProgramHost.getEnvironmentVariable` | D135 | `getEnvironmentVariable?(name: string): string \| undefined;` | `getEnvironmentVariable?: { getEnvironmentVariable(name: string): string \| undefined; }["getEnvironmentVariable"] \| undefined;` |
| P18 | tsbuildPublic.ts:217 `SolutionBuilderHostBase.createDirectory` | D132 | `createDirectory?(path: string): void;` | `createDirectory?: { createDirectory(path: string): void; }["createDirectory"] \| undefined;` |
| P19 | tsbuildPublic.ts:222 `SolutionBuilderHostBase.writeFile` | D133 | `writeFile?(path: string, data: string, writeByteOrderMark?: boolean): void;` | `writeFile?: { writeFile(path: string, data: string, writeByteOrderMark?: boolean): void; }["writeFile"] \| undefined;` |
| I12 | program.ts:505 `CompilerHostLikeForCache.writeFile` (interface `@internal`) | cascade of P19 | `writeFile?: WriteFileCallback;` | `writeFile?: WriteFileCallback \| undefined;` |
| I13 | watchPublic.ts:267 `ProgramHost.createDirectory` (member `@internal`) | cascade of P18 | `createDirectory?(path: string): void;` | `createDirectory?: { createDirectory(path: string): void; }["createDirectory"] \| undefined;` |
| I14 | watchPublic.ts:269 `ProgramHost.writeFile` (member `@internal`) | cascade of P19 | `writeFile?(path: string, data: string, writeByteOrderMark?: boolean): void;` | `writeFile?: { writeFile(path: string, data: string, writeByteOrderMark?: boolean): void; }["writeFile"] \| undefined;` |

"…same…" is the exact original parameter list. adapt.cjs holds every text in full.

Why the cascade owners: `SolutionBuilderHostBase` extends `ProgramHost`, so once P18 and P19 admit undefined, `ProgramHost`'s internal `createDirectory` and `writeFile` must too (I13, I14). `CompilerHostLikeForCache` takes a `SolutionBuilderHost` at tsbuildPublic.ts:760, so its `writeFile` must too (I12). 97b moved them here because restoring them there brought nothing back.

`api-sanctions.json` holds the 19 public rows. Each row has its owner path in `generate-api.cjs`'s `owned()` convention (`ownerPath()` drops the `ts` namespace, as adaptations 20 and 40 record it), `kind: "PropertySignature"`, stock typescript.d.ts's exact member text as `before`, the member text after this head, and what tsc really stores there. For a method there is no separate type node, so `before` and `after` are whole member texts.

## Why it's sound

- **No reader gains an error, in the build:** upstream compiles this tree without `exactOptionalPropertyTypes`. Without it, `x?: T | undefined` reads exactly as `x?: T`, and the indexed method type is the method's own bivariant function type. To check this, I applied 97b and 97d to a copy of the pinned upstream source (050880ce, the same member texts) and type-checked all 601 files of `src` (compiler, services, server, jsTyping, deprecatedCompat, the typings installer, tsserver, tsc, typescript, watchGuard, harness, testRunner) with upstream's base options. Before and after give identical diagnostics.
- **No reader gains an error, under Adamic's options:** the inventory's checker over src/compiler (strict, exactOptionalPropertyTypes, noUncheckedIndexedAccess) finds nothing new. Under `exactOptionalPropertyTypes`, reading `x?: T` and `x?: T | undefined` both give `T | undefined`. Every read is an optional call, a guard, `maybeBind`, or an indexed read that already includes undefined (inventory flag F2).
- **Writers:** the only stores the widening newly admits are the undefined values tsc already stores. Defined values were admitted before and still are.
- **Runtime:** the edits are type-only. Each file's JavaScript is byte-identical.
- **What a consumer can see:** see below.

## What moves for a consumer

The declared API surface is what moves: 19 lines of typescript.d.ts, each an optional method that becomes an optional property of the same method type plus undefined (`proofs/typescript-d-ts-lines.md`). Every one of them remains optional.

- **Without `exactOptionalPropertyTypes`:** a consumer sees the same read type, the same assignability and the same bivariance. A class may still implement the member with a method.
- **With `exactOptionalPropertyTypes`:** a consumer that passes one of these hosts into its own narrower `{ x?(…): R }` shape gets a TS2379. That's what the inventory's cascade owners show inside the tree, and the cost Adamic accepted. Two such places are in upstream's own services, which nothing here compiles with that option: services.ts:1751:63 and services/utilities.ts:3518:67, `LanguageServiceHost` into `directoryProbablyExists`'s narrow parameter. These are the same shape adaptation 32's function-view casts handle in moduleNameResolver.ts. Also, a consumer's `Required<CompilerHost>` keeps the explicit undefined under that option.

## Which tsc baselines can move

Only the declared API surface. The JavaScript is byte-identical, so no emit, runtime or tsserver behavior changes.

- **`api/typescript.d.ts`**, the one baseline the lane already permits to differ (`unittests:: Public APIs for typescript.d.ts should be acknowledged when they change`). Its content now also differs in these 19 members. The built file is formatted by upstream's build, so its line breaks inside the new type literals may differ from the single-line spelling posted here. The lane's normalizer ignores that formatting.
- **The 10 upstream cases that compile against the built API through `/.ts/typescript.d.ts`:** `APILibCheck` and the nine `APISample_*` cases, all in tests/cases/compiler. `proofs/api-samples.cjs` compiles each one against stock typescript.d.ts and against the adapted one, and both give the same diagnostics: none for the nine samples, and for APILibCheck only the two internal-bundle imports this environment doesn't have, on both sides. APILibCheck checks the API file itself under `exactOptionalPropertyTypes`. A mutant that leaves `ModuleResolutionHost.trace` a method in the API makes it report `ParseConfigHost` incorrectly extends `ModuleResolutionHost` (TS2430), so the check does bite.
- **Whether the 13,453 baseline configurations' outputs include the API d.ts:** not established. No tracked file in the tree defines a 13,453 count. The pinned checkout has 12,445 compiler and conformance case files, and how option variants expand that count isn't recorded here. What is established is the list of readers: only the 10 cases above and the public API unittest read the built typescript.d.ts. Every other compiler and conformance case compiles its own inputs against the lib files.
- **APILibCheck's internal half:** it also imports `typescript.internal.d.ts` and `tsserverlibrary.internal.d.ts`, which only a full build produces, so they weren't checked here.

Outside this head: the lane's sanction (`stage3/lane/sanctioned-api.json`, generated by `generate-api.cjs`) does not yet include these 19. Until `generate-api.cjs` reads `api-sanctions.json` and the sanction is regenerated, the lane's `api_check` will report them as unsanctioned declarations.

## Rows and mutant pins

With 97b applied, this head clears 21 rows: D043, D049, D050, D051, D125, D126, D128, D129, D130, D131, D132, D133, D134, D135, D136, D138, D139, D140, D141, D142 and D144. With 97c's D143 that makes all 22 rows left after 97b, and with 97b's six, all 28 host-slot rows.

Each mutant restores one member of the adapted tree to its before text and reruns the check. Each of the 22 brings something back, and every returning TS2375 names the restored member:

| Id | Restored alone, it brings back |
|---|---|
| P1 | D134 (`trace`), and a TS2430 for `ParseConfigHost` extending `ModuleResolutionHost` at types.ts:4540:18, and a TS2379 at commandLineParser.ts:3672:106 |
| P2 | D049, D144, D051 and D134 (`directoryExists`), and TS2379s at tsbuildPublic.ts:454:17, 475:17 and 488:17 |
| P3 | D043, D051 and D134 (`realpath`), and the same three TS2379s |
| P4 | D050, D051 and D134 (`getDirectories`) |
| P5 | D051 (`trace`) |
| P6 | D134 (`getDefaultLibLocation`) |
| P7 | D134 (`readDirectory`) |
| P8 | D128 and D139 |
| P9 | D130 and D142 |
| P10 | D129 and D141 |
| P11 | D125 and D138 |
| P12 | D126 and D140 |
| P13 | D043, D134 (`createHash`) and D136, and a TS2379 at tsbuildPublic.ts:1642:90 |
| P14 | D135 (`createHash`) |
| P15 | D131 |
| P16 | D135 (`realpath`) |
| P17 | D135 (`getEnvironmentVariable`) |
| P18 | D132 |
| P19 | D133 |
| I12 | a TS2379, `SolutionBuilderHost` into `CompilerHostLikeForCache`, at tsbuildPublic.ts:760:9 |
| I13 | a TS2430, `SolutionBuilderHostBase` extending `ProgramHost`, at tsbuildPublic.ts:216:18, and TS2379s at 435:60 and 499:66 (`createDirectory`) |
| I14 | the same TS2430 and two TS2379s (`writeFile`) |

## Proof

- **The shape:** the adapter parses with stock TypeScript 6.0.3. It finds each owner as the one top-level exported interface of that name in its file and requires exactly one member of that name in it.
  - A public member's owner and the member itself must not be `@internal`. I12's interface and I13's and I14's members must still be `@internal`.
  - The member's text must be exactly the inventoried before or after, with a multi-line member spelled in the file's own line breaks.
  - 97b's 11 members must each read as 97b's after, or the adapter refuses, saying 97b is not applied.
- **Refusals:** each was probed on a scratch copy, and each exits 1 with no write:
  - the adapted tree without 97b;
  - 97b with only its I9 reverted;
  - types.ts with only P13 adapted (a partly adapted file);
  - types.ts adapted and the other three files not (a partly adapted tree);
  - `ProgramHost` renamed (a missing owner);
  - P19 in the arrow form (a shape not inventoried);
  - a public member marked `@internal`;
  - `CompilerHostLikeForCache` without its `@internal`;
  - a second `ProgramHost.realpath`;
  - watchPublic.ts with one bare LF (mixed line breaks).

  A copy of the adapter that appends a statement to watchPublic.ts, with its byte-count check removed, is refused by its JavaScript check.
- **Planning:** the adapter plans every file before writing any. For each file it re-reads its own result as fully adapted, checks that only the members' bytes moved, and checks the JavaScript. It re-reads each file just before writing, then writes each changed file once.
- **Source:** on a fresh copy of the adapted tree with 97b applied, the diff is exactly the 22 members, 24 lines in four files. That is types.ts 15 (P11 and P12 two each), watchPublic.ts 6, tsbuildPublic.ts 2 and program.ts 1. The written files are identical to the inventory's in-memory widening (scratchpad a97/inv/widen.cjs).
- **JavaScript:** `ts.transpileModule` before and after, with an ES2020 target, is byte-identical under both CommonJS and ESNext. Under CommonJS the sizes are types.ts 143,226 bytes, watchPublic.ts 41,577, tsbuildPublic.ts 87,806 and program.ts 240,346.
- **Idempotence:** a second run edits nothing and leaves the SHA-256s unchanged. An input that is already adapted is accepted with no write.
- **Declarations:** `proofs/check-api.cjs <stock typescript.d.ts> <tree before> <tree after> <out>` proves `api-sanctions.json`:
  - each row names one optional method in stock typescript.d.ts by `ownerPath()`, spelled exactly as `before`;
  - each `after` is that method in the idiom, and nothing else;
  - the lane's own `normalize-api.cjs` finds exactly the 19 `member:["ts", …]` keys changed in typescript.d.ts, each from `MethodSignature` to `PropertySignature`, with no comment change;
  - the adapted tree's declarations (`transpileDeclaration` with `stripInternal`, all four files) change in the same 19 members and nothing else. program.ts's are byte-identical, and each emitted member normalizes to its sanction's `after`.

  It writes `proofs/typescript.d.ts.diff`, `proofs/emitted-public.d.ts.diff` and `proofs/typescript.d.ts.lines.json`.
- **Diagnostics:** measured with the inventory's in-memory checker (stock TS 6.0.3, all of src/compiler, strict, exactOptionalPropertyTypes, noUncheckedIndexedAccess), on the adapted tree with 97b applied:
  - 135 before and 114 after;
  - the 21 removed are exactly the 21 rows above;
  - nothing new appears anywhere, and no remaining diagnostic's text changes;
  - the written tree's log is byte-identical to the in-memory log;
  - with 97c also applied, 134 before and 113 after, and no host-slot row is left. What remains of the inventory's 33 rows is the two TS18046 catch sites and 97a's three TS2345.
- **Upstream src:** with 97b and 97d applied to a copy of the pinned upstream source, all 601 files give the same diagnostics as before under upstream's base options.
- **Golem, `adamic types`** on the adapted tree at main 7b9c0fe7 with 97b applied first: whole program 138 to 117, the 21 rows removed and nothing new. Own-file: tsbuildPublic.ts 8 to 0, watch.ts 2 to 0, watchPublic.ts 8 to 1 (D143, 97c's), program.ts 9 to 5 (97a's three and the two catch sites). A second run edits nothing.

Not yet run: the sharded lane oracle. It runs in Loom as part of the wave, per Correctness' rule on #axochk.
