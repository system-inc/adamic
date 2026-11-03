# 99i: module cache uses its declared value type

lane: run as part of Correctness's wave

Base: origin/main fec34067c11f6ad76c3a8a0811546daad1f18109. Upstream:
TypeScript 6.0.3, 050880ce59e30b356b686bd3144efe24f875ebc8.
This anticipates Runtime Performance C2, 024c5b2189abea29441452b883c74894b92a8a73.

## Claim checked before the edit

The fully composed source's createProgram declares resolvedModules as
Map<Path, ModeAwareCache<ResolvedModuleWithFailedLookupLocations>> | undefined.
In processImportedModules, the local resolutionsInFile has one population site:
`resolutionsInFile.set(moduleName, mode, resolutions[index]!);`.
There are only three references to this local: its declaration, publication in
resolvedModules, and this set. There is no other local alias or callback escape.

The resolutions array comes from either resolvedModulesProcessing.get(file.path)
or resolveModuleNamesReusingOldState(moduleNames, file). The processing map's
value is readonly ResolvedModuleWithFailedLookupLocations[]; its sole writer in
tryReuseStructureFromOldProgram stores resolveModuleNamesReusingOldState's result.
That function explicitly returns readonly ResolvedModuleWithFailedLookupLocations[].
Its generic helper returns the module worker's results, oldProgram.getResolvedModule
results, or emptyResolution. The worker explicitly returns the same module array;
old-program access returns the module type; emptyResolution is explicitly the
intersection of module and type-directive results, with resolvedModule: undefined.
All values written therefore have the required module type, including failed and
ambient resolutions.

Publication is the only insertion into resolvedModules. Its other assignment
reuses oldProgram.resolvedModules, which has the identical declared map type.
The returned program exposes that same typed map. Searching all src finds no
other property access beyond that old-program reuse. The factory's private alias
cache and underlying Map<ModeAwareCacheKey, T> only store the set method's value;
get and forEach return it unchanged, delete removes it, and has/size read metadata.
No hidden factory write introduces a different resolution type.

Reads in createProgram are getResolvedModule and forEachResolvedModule, through
the generic forEachResolution's resolutionCache/resolutions aliases. Both expose
the module type. Consumers are getResolvedModuleFromModuleSpecifier, old-state
reuse and equality comparison, getPackagesMap, and symlink processing. External
consumers in checker.ts and utilities.ts read resolvedModule or alternateResult;
server/project.ts reads resolvedModule for unresolved imports and forwards the
iterator to the symlink cache. The symlink cache reads resolvedModule's paths.
The test and harness consumers in reuseProgramStructure.ts, incrementalUtils.ts,
and tsserver/typingsInstaller.ts serialize results or read resolvedModule and its
file name. The generic harness constraint ResolutionWithFailedLookupLocations
uses shared fields and remains structurally satisfied by the module type; it
does not read watcher-only isInvalidated/files fields.

The getResolvedModuleFromModuleSpecifier route additionally serves
services/goToDefinition.ts, suggestionDiagnostics.ts, findAllReferences.ts,
getEditsForFileRename.ts, and codefixes/fixImportNonExportedMember.ts,
fixSpelling.ts, convertToEsModule.ts, addMissingResolutionModeImportAttribute.ts,
and importFixes.ts. These read resolvedModule (including resolvedFileName and
resolvedUsingTsExtension); the rename helper also reads failedLookupLocations,
which the module result declares. program.ts's import-reason helper reads the
resolved module's packageId. None reads or writes watcher-only fields.

No reader needs the wider
watcher-cache type. The interface's required resolvedModule field is precisely
what the wider type fails to promise.

## Exact change and contract

In createProgram's processImportedModules, before (after preceding adapters):

```ts
const resolutionsInFile = createModeAwareCache<ResolutionWithFailedLookupLocations>();
```

After:

```ts
const resolutionsInFile = createModeAwareCache<ResolvedModuleWithFailedLookupLocations>();
```

Stock upstream has an argumentless generic call; the preceding adaptations make
its wider inferred type explicit. 99i replaces that explicit type argument.
This is type-only: no cast, runtime expression, public declaration or handling of
failed resolution changes. Type arguments erase from emitted JavaScript.

The adapter uses stock TypeScript 6.0.3 to parse, locates unique named enclosing
functions, and pins the exact declaration, guard, sibling statements, publication,
all local references, population write and loop. Any unlisted spelling or partial
edit is refused before a write. It validates the planned result, rereads the
original, then writes once. An already adapted input returns zero edits without
writing. Only adapt.cjs and README.md are adapter inputs.

## Proofs

Raw commands and logs are retained under /workspace/scratch with the 99i prefix
and the named apply, setup, build and checker logs. The reverse-edit mutant is
named undo-99i-on-C2: replace the after declaration with the before declaration
on the C2 checker tree. It must restore program.ts TS2345.

Measured results:

- Setup passed with the requested GOPROXY and timeout command.
- Main apply passed all 60 adapters; main plus 99i passed all 61. Every
  adapter has both refs/stage3 before/after refs and its patch-table row.
- The adapter's first run makes one edit; its second run makes zero edits
  and preserves mtime. Eight refusal fixtures preserve both bytes and mtime:
  untyped call, wrong type, moved owner, changed publication, half population,
  extra alias, parse error and duplicate owner.
- Whole-program `adamic types` uses the 79 recursively sorted compiler .ts/.a
  roots, unchanged checker options and prelude, as axolotl-wall-2. Main has
  **9 diagnostics before and 9 after**, with identical full diagnostics after
  normalizing only the absolute tree prefix. No diagnostic is added or removed.
  These are fresh apply trees without upstream node_modules; the inherited
  source-map-support TS2307 is retained in every count.
- C2 was merged independently into disposable main and main-plus-99i checkouts,
  and Adamic was rebuilt in each. Neither scratch merge is pushed. C2 without
  99i has **10 diagnostics**, adding only program.ts:3921:60 TS2345: a
  ModeAwareCache<ResolutionWithFailedLookupLocations> lacks the required
  resolvedModule field. C2 with 99i has **9**, exactly the original main
  diagnostic text and chains, with nothing else added. undo-99i-on-C2 has
  **10**, exactly matching the C2-without-99i observation. The mutant changes
  just the type argument in the after tree, uses its C2 binary, and restores
  the source afterward.
- Independent adapted-tree copies use the upstream lockfile's npm install
  and TypeScript's own LKG tsc: `node lib/tsc.js -b src/compiler
  --emitDeclarationOnly false --pretty false`. Both builds exit 0. The
  emitted built/local/compiler/program.js is byte-identical, **240875 bytes**,
  SHA-256 `84416ea952adee2078357a097bddcdfd5ec553b9a526d7000574b3044f957fa3`.
  The original checker trees never receive those build dependencies.
- `python3 -B stage3/lane/generate_shards.py --check` passes.
- With ADAMIC_TYPESCRIPT_SOURCE set to the pinned upstream, `GOMAXPROCS=1
  go test -count=1 -skip 'TestShard_|TestProduct_' ./stage3/lane/` passes.
  The worker has a four-core quota. Initial default-concurrency attempts
  exposed missing input configuration and then the existing 55-second Python
  group timeouts under load. The passing run changes concurrency only, with
  no test, assertion, timeout, baseline or expected-count edits.
- There is no public API line. sanctioned-api.json, the shard plan and all
  generated lane files remain unchanged.

Reproduction inputs and logs: /workspace/scratch/99i-main is the adaptation
checkout; 99i-before and 99i-after are full apply trees; 99i-c2-before and
99i-c2-after are disposable compiler checkouts. 99i-results/commands.json
records all checker argument lists and exit statuses; its stderr files retain
all diagnostic chains, and exact-diagnostic-delta.json records the sole C2
addition. 99i-measure.py, 99i-contract.cjs, 99i-apply-check.py and 99i-emit.py
retain the proof commands. apply-before.log, apply-after.log, 99i-contract.log,
99i-emit-results.json, 99i-shards.log and 99i-lane-serial.log retain the raw
results. No generated evidence is an adapter input.

Full-apply replay uses ordinary apply's product-cache interface and a fresh
output destination, as apply refuses to replace an existing output. The second
ordinary apply restores the same product without running any adapter or making
source edits; compare its complete non-.git payload byte for byte with the first
output. Separately, 99i's own second invocation verifies zero edits and no write.

An additional attempted in-place replay with --adapt-checkout reaches the
pre-existing 30-indexed-reads guard and stops at `declined site changed:
core.ts:992:21`. Later adapters have changed that earlier adapter's inventoried
site. Therefore this README does not claim all existing adapters support
in-place replay of the final composition. No existing adapter is changed to
suppress the refusal. This limitation is separate from ordinary apply's cached
second invocation and 99i's own zero-edit contract.
