# 99x: bind host methods at the read

Task kpp0f8s. javascriptChanging true.
Base main: 3e395f9e5a7038d859ec8930be72cde68d8be3ff.
Pinned upstream: TypeScript 6.0.3, 050880ce59e30b356b686bd3144efe24f875ebc8.

Adamic's November 6, 2026 ruling authorizes stage 3 to replace
`maybeBind(X, X.m)` with `X.m?.bind(X)` where X is one identifier.
Apple's apple/bind-1 head supplies the separate bound-method lowering work.
This branch changes no compiler implementation or policy.

The equivalence argument supplied in the ruling is that both forms read X,
then X.m, then X again with nothing between. If the slot holds a function,
both return it bound to X; undefined returns undefined; anything else truthy
throws TypeError at .bind in both. The core.ts maybeBind body is unchanged.

## Adapter

Run `NODE_PATH=stage3/api/node_modules node stage3/adapt/99-correctness-axolotl-x/adapt.cjs TREE`.
The API must be stock TypeScript 6.0.3. The scanner resolves the callee's symbol
to core.ts's exported declaration, requires exactly two arguments, requires
an Identifier receiver and a non-optional property access, and compares the
receiver identifiers' symbols. binder.ts's two calls to its own one-argument
helper are excluded by symbol identity and remain unchanged.

owners.json pins exact before/after text and SHA-256 for all 15 innermost named
owners. Nested getEmitHost and its createProgram parent are both pinned.
No whole-file hash gates adaptation. Missing, duplicate, moved, changed or
half-adapted owners and mixed file/tree/import states fail before any write.
Every source file is reread before writes. CRLF is preserved. An adapted replay
returns zero files and preserves bytes and mtimes. generate.cjs is an explicit
pin-generation tool, never called by apply.

The frozen inventory has 46 rewrites, zero refused core calls and zero sites
left as written. No cast is added. All six maybeBind import specifiers become
unused and are removed: builder.ts, checker.ts, program.ts, tsbuildPublic.ts,
watch.ts and watchPublic.ts. binder.ts's local helper is not a refused core call.

Python's actual directory sort places 99x after 99u and before
99-wasm-crypto-import. The fixed requested directory name cannot sort after
99-wasm-crypto-import; no other adaptation directory has been changed.

## Site inventory

Coordinates below are from the pinned pristine source. inventory.json also
records the fully adapted main coordinates used to generate the owner pins.

| File:line:column | Owner | Original call |
| --- | --- | --- |
| builder.ts:1727:30 | createBuilderProgram.emitBuildInfo | `maybeBind(host, host.writeFile)` |
| builder.ts:1815:34 | createBuilderProgram.emitNextAffectedFileOrDtsErrors | `maybeBind(host, host.writeFile)` |
| builder.ts:1914:78 | createBuilderProgram.getWriteFileCallback | `maybeBind(host, host.writeFile)` |
| checker.ts:54264:26 | createBasicNodeBuilderModuleSpecifierResolutionHost | `maybeBind(host, host.getSymlinkCache)` |
| checker.ts:54275:40 | createBasicNodeBuilderModuleSpecifierResolutionHost | `maybeBind(host, host.getGlobalTypingsCacheLocation)` |
| program.ts:494:21 | createCompilerHostWorker | `maybeBind(system, system.createHash)` |
| program.ts:1956:40 | createProgram | `maybeBind(host, host.getGlobalTypingsCacheLocation)` |
| program.ts:2626:23 | createProgram.getEmitHost | `maybeBind(host, host.realpath)` |
| program.ts:2632:25 | createProgram.getEmitHost | `maybeBind(host, host.createHash)` |
| program.ts:2634:20 | createProgram.getEmitHost | `maybeBind(host, host.trace)` |
| program.ts:5114:26 | parseConfigHostFromCompilerHostLike | `maybeBind(directoryStructureHost, directoryStructureHost.directoryExists)` |
| program.ts:5115:25 | parseConfigHostFromCompilerHostLike | `maybeBind(directoryStructureHost, directoryStructureHost.getDirectories)` |
| program.ts:5116:19 | parseConfigHostFromCompilerHostLike | `maybeBind(directoryStructureHost, directoryStructureHost.realpath)` |
| tsbuildPublic.ts:302:16 | createSolutionBuilderHostBase | `maybeBind(system, system.now)` |
| tsbuildPublic.ts:438:46 | createSolutionBuilderState | `maybeBind(host, host.resolveModuleNameLiterals)` |
| tsbuildPublic.ts:439:60 | createSolutionBuilderState | `maybeBind(host, host.resolveTypeReferenceDirectiveReferences)` |
| tsbuildPublic.ts:440:35 | createSolutionBuilderState | `maybeBind(host, host.resolveLibrary)` |
| tsbuildPublic.ts:441:39 | createSolutionBuilderState | `maybeBind(host, host.resolveModuleNames)` |
| tsbuildPublic.ts:442:51 | createSolutionBuilderState | `maybeBind(host, host.resolveTypeReferenceDirectives)` |
| tsbuildPublic.ts:443:45 | createSolutionBuilderState | `maybeBind(host, host.getModuleResolutionCache)` |
| tsbuildPublic.ts:500:16 | createSolutionBuilderState | `maybeBind(host, host.trace)` |
| watch.ts:677:20 | createWatchHost | `maybeBind(system, system.watchFile)` |
| watch.ts:678:25 | createWatchHost | `maybeBind(system, system.watchDirectory)` |
| watch.ts:679:21 | createWatchHost | `maybeBind(system, system.setTimeout)` |
| watch.ts:680:23 | createWatchHost | `maybeBind(system, system.clearTimeout)` |
| watch.ts:761:32 | createCompilerHostFromProgramHost | `maybeBind(host, host.getDefaultLibLocation)` |
| watch.ts:774:16 | createCompilerHostFromProgramHost | `maybeBind(host, host.trace)` |
| watch.ts:775:26 | createCompilerHostFromProgramHost | `maybeBind(directoryStructureHost, directoryStructureHost.directoryExists)` |
| watch.ts:776:25 | createCompilerHostFromProgramHost | `maybeBind(directoryStructureHost, directoryStructureHost.getDirectories)` |
| watch.ts:777:19 | createCompilerHostFromProgramHost | `maybeBind(host, host.realpath)` |
| watch.ts:778:33 | createCompilerHostFromProgramHost | `maybeBind(host, host.getEnvironmentVariable)` |
| watch.ts:779:21 | createCompilerHostFromProgramHost | `maybeBind(host, host.createHash)` |
| watch.ts:780:24 | createCompilerHostFromProgramHost | `maybeBind(host, host.readDirectory)` |
| watch.ts:856:19 | createProgramHost | `maybeBind(system, system.realpath)` |
| watch.ts:857:33 | createProgramHost | `maybeBind(system, system.getEnvironmentVariable)` |
| watch.ts:861:21 | createProgramHost | `maybeBind(system, system.createHash)` |
| watch.ts:864:14 | createProgramHost | `maybeBind(system, system.now)` |
| watchPublic.ts:125:23 | createIncrementalCompilerHost | `maybeBind(system, system.createHash)` |
| watchPublic.ts:506:56 | createWatchProgram | `maybeBind(host, host.useSourceOfProjectReferenceRedirect)` |
| watchPublic.ts:529:46 | createWatchProgram | `maybeBind(host, host.resolveModuleNameLiterals)` |
| watchPublic.ts:530:39 | createWatchProgram | `maybeBind(host, host.resolveModuleNames)` |
| watchPublic.ts:534:60 | createWatchProgram | `maybeBind(host, host.resolveTypeReferenceDirectiveReferences)` |
| watchPublic.ts:535:51 | createWatchProgram | `maybeBind(host, host.resolveTypeReferenceDirectives)` |
| watchPublic.ts:543:9 | createWatchProgram | `maybeBind(host, host.getModuleResolutionCache)` |
| watchPublic.ts:549:9 | createWatchProgram | `maybeBind(host, host.hasInvalidatedResolutions)` |
| watchPublic.ts:552:9 | createWatchProgram | `maybeBind(host, host.hasInvalidatedLibResolutions)` |

## Evidence

The earlier missing-source stop is retained in check1-before.log and
prerequisites.log as historical evidence. The user supplied the reference
location and setup authorization on continuation. 99w's adapt.cjs, README.md
and PROOF.md were read from 06dad75d; 99k's main adapter and documentation were
also read. Setup completed successfully, the pinned source was cloned, and
the full baseline apply passed. New logs supersede the historical stop.
PROOF.md records the current checks. Tests write logs outside this hashed
directory while running; completed evidence is copied here before publication.

## Current verification and stopping point

The final replay rule compares outcomes only. Matching refusals, including
40's builder.ts brand refusal, are baseline records. The first 44 adapter
pairs match. Pair 45, 95-checker-axolotl-h, differs: main reaches a createProgram
writer-pin refusal, while after refuses earlier on
createBasicNodeBuilderModuleSpecifierResolutionHost in checker.ts. That writer
is one of 99x's owners. Its 95h pin matches main and fails after the two checker
rewrites. Complete paired records, exact refusal text and their nonempty diff
are retained under paired-replay/. No other adaptation was modified.

Both full applies, owner contracts and zero-write direct replay pass. The
79-root type check has zero diagnostics. All nine JavaScript products pass
the emitted comparison, and the restored-site control detects exactly one
missed site. The ordinary 79-root census has zero changed first stops, and
all six CW051 units have before/after measurements. The whole upstream lane
passes with 106,366 passing, one sanctioned API failure and zero pending.
The first 40 of 398 standard shards pass; the remaining 358 were not run
because of the first actual paired-replay difference. See PROOF.md for the
full measured table and exact stopping point. This branch is not ready to land.
