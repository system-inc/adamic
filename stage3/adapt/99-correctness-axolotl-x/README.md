# 99x: explicit host method binding with earlier-owner handoffs

Task kpp0f8s. javascriptChanging true. Pinned TypeScript 6.0.3 source:
050880ce59e30b356b686bd3144efe24f875ebc8. Base main remains 3e395f9e5.

Adamic's November 6 ruling permits maybeBind(X, X.m) to become X.m?.bind(X)
when X is one identifier. The supplied equivalence argument is that both
forms read X, then X.m, then X again with nothing between. A function returns
bound to X; undefined returns undefined; anything else truthy throws TypeError
at .bind in both. The core.ts helper body remains unchanged.

## Owner handoffs and current count

The user requires earlier owner pins to own later edits to that owner.
The runtime-input audit finds five overlapping owner pins:

| Earlier adapter | Owner | Sites left |
| --- | --- | ---: |
| 95-checker-axolotl-h | checker.ts createBasicNodeBuilderModuleSpecifierResolutionHost | 2 |
| 95-checker-axolotl-h | program.ts createProgram, including its nested getEmitHost | 4 |
| 32-indexed-reads-program | watch.ts createCompilerHostFromProgramHost | 8 |
| 32-indexed-reads-program | watch.ts createProgramHost | 4 |

95h separately pins getEmitHost as well as its createProgram ancestor.
32's public-host-guards.cjs hashes each complete emitted function body.
The audit follows earlier adapters' local CJS/JSON runtime inputs, checks
literal JSON references too, and distinguishes full-owner pins from local
expression or ancestry guards. owner-pin-audit.json, owner-pin-audit-extra.json
and ownership.json retain the evidence. No earlier adaptation directory changes.

Current expectation: **28 rewritten, 18 left**. Six left sites belong to 95h;
twelve belong to 32. The requested two checker sites carry the exact reason
"owner pinned by 95-checker-axolotl-h; folded there" below. "Folded there"
records the ownership handoff, not an edit by this branch to another adapter.
There are zero unsupported-shape refusals and no type-driven fallback yet.

## Adapter contract

Run `NODE_PATH=stage3/api/node_modules node stage3/adapt/99-correctness-axolotl-x/adapt.cjs TREE`.
The parser must be stock TypeScript 6.0.3. Callee symbols must resolve to the
core export; both receiver identifiers must name the same symbol; the method
argument must be a non-optional property access. binder.ts's two calls to its
separate one-argument helper remain excluded by symbol identity.

owners.json pins the exact before/after text and SHA-256 of ten edited owners,
never whole-file hashes. Changed, missing, duplicate, moved and half-adapted
owners and mixed tree/import states fail before any write. Every source file
is reread before writing. CRLF is preserved; replay writes zero files and
preserves bytes and mtimes. Handed-off calls are preserved and checked against
their separate inventory. No cast is added.

Five source files change. The maybeBind imports are removed only in builder.ts,
tsbuildPublic.ts and watchPublic.ts. program.ts and watch.ts retain imports for
their handed-off calls; checker.ts is untouched. Python's actual sort places
99x after 99u and before 99-wasm-crypto-import. No other directory is reordered.

## All 46 original sites

Coordinates are from the pinned pristine source. The earlier-owner rule also
covers calls nested within a pinned ancestor function.

| File:line:column | Owner | Original call | 99x action |
| --- | --- | --- | --- |
| builder.ts:1727:30 | createBuilderProgram.emitBuildInfo | `maybeBind(host, host.writeFile)` | rewrite |
| builder.ts:1815:34 | createBuilderProgram.emitNextAffectedFileOrDtsErrors | `maybeBind(host, host.writeFile)` | rewrite |
| builder.ts:1914:78 | createBuilderProgram.getWriteFileCallback | `maybeBind(host, host.writeFile)` | rewrite |
| checker.ts:54264:26 | createBasicNodeBuilderModuleSpecifierResolutionHost | `maybeBind(host, host.getSymlinkCache)` | owner pinned by 95-checker-axolotl-h; folded there |
| checker.ts:54275:40 | createBasicNodeBuilderModuleSpecifierResolutionHost | `maybeBind(host, host.getGlobalTypingsCacheLocation)` | owner pinned by 95-checker-axolotl-h; folded there |
| program.ts:494:21 | createCompilerHostWorker | `maybeBind(system, system.createHash)` | rewrite |
| program.ts:1956:40 | createProgram | `maybeBind(host, host.getGlobalTypingsCacheLocation)` | owner pinned by 95-checker-axolotl-h; folded there |
| program.ts:2626:23 | createProgram.getEmitHost | `maybeBind(host, host.realpath)` | owner pinned by 95-checker-axolotl-h; folded there |
| program.ts:2632:25 | createProgram.getEmitHost | `maybeBind(host, host.createHash)` | owner pinned by 95-checker-axolotl-h; folded there |
| program.ts:2634:20 | createProgram.getEmitHost | `maybeBind(host, host.trace)` | owner pinned by 95-checker-axolotl-h; folded there |
| program.ts:5114:26 | parseConfigHostFromCompilerHostLike | `maybeBind(directoryStructureHost, directoryStructureHost.directoryExists)` | rewrite |
| program.ts:5115:25 | parseConfigHostFromCompilerHostLike | `maybeBind(directoryStructureHost, directoryStructureHost.getDirectories)` | rewrite |
| program.ts:5116:19 | parseConfigHostFromCompilerHostLike | `maybeBind(directoryStructureHost, directoryStructureHost.realpath)` | rewrite |
| tsbuildPublic.ts:302:16 | createSolutionBuilderHostBase | `maybeBind(system, system.now)` | rewrite |
| tsbuildPublic.ts:438:46 | createSolutionBuilderState | `maybeBind(host, host.resolveModuleNameLiterals)` | rewrite |
| tsbuildPublic.ts:439:60 | createSolutionBuilderState | `maybeBind(host, host.resolveTypeReferenceDirectiveReferences)` | rewrite |
| tsbuildPublic.ts:440:35 | createSolutionBuilderState | `maybeBind(host, host.resolveLibrary)` | rewrite |
| tsbuildPublic.ts:441:39 | createSolutionBuilderState | `maybeBind(host, host.resolveModuleNames)` | rewrite |
| tsbuildPublic.ts:442:51 | createSolutionBuilderState | `maybeBind(host, host.resolveTypeReferenceDirectives)` | rewrite |
| tsbuildPublic.ts:443:45 | createSolutionBuilderState | `maybeBind(host, host.getModuleResolutionCache)` | rewrite |
| tsbuildPublic.ts:500:16 | createSolutionBuilderState | `maybeBind(host, host.trace)` | rewrite |
| watch.ts:677:20 | createWatchHost | `maybeBind(system, system.watchFile)` | rewrite |
| watch.ts:678:25 | createWatchHost | `maybeBind(system, system.watchDirectory)` | rewrite |
| watch.ts:679:21 | createWatchHost | `maybeBind(system, system.setTimeout)` | rewrite |
| watch.ts:680:23 | createWatchHost | `maybeBind(system, system.clearTimeout)` | rewrite |
| watch.ts:761:32 | createCompilerHostFromProgramHost | `maybeBind(host, host.getDefaultLibLocation)` | owner pinned by 32-indexed-reads-program; folded there |
| watch.ts:774:16 | createCompilerHostFromProgramHost | `maybeBind(host, host.trace)` | owner pinned by 32-indexed-reads-program; folded there |
| watch.ts:775:26 | createCompilerHostFromProgramHost | `maybeBind(directoryStructureHost, directoryStructureHost.directoryExists)` | owner pinned by 32-indexed-reads-program; folded there |
| watch.ts:776:25 | createCompilerHostFromProgramHost | `maybeBind(directoryStructureHost, directoryStructureHost.getDirectories)` | owner pinned by 32-indexed-reads-program; folded there |
| watch.ts:777:19 | createCompilerHostFromProgramHost | `maybeBind(host, host.realpath)` | owner pinned by 32-indexed-reads-program; folded there |
| watch.ts:778:33 | createCompilerHostFromProgramHost | `maybeBind(host, host.getEnvironmentVariable)` | owner pinned by 32-indexed-reads-program; folded there |
| watch.ts:779:21 | createCompilerHostFromProgramHost | `maybeBind(host, host.createHash)` | owner pinned by 32-indexed-reads-program; folded there |
| watch.ts:780:24 | createCompilerHostFromProgramHost | `maybeBind(host, host.readDirectory)` | owner pinned by 32-indexed-reads-program; folded there |
| watch.ts:856:19 | createProgramHost | `maybeBind(system, system.realpath)` | owner pinned by 32-indexed-reads-program; folded there |
| watch.ts:857:33 | createProgramHost | `maybeBind(system, system.getEnvironmentVariable)` | owner pinned by 32-indexed-reads-program; folded there |
| watch.ts:861:21 | createProgramHost | `maybeBind(system, system.createHash)` | owner pinned by 32-indexed-reads-program; folded there |
| watch.ts:864:14 | createProgramHost | `maybeBind(system, system.now)` | owner pinned by 32-indexed-reads-program; folded there |
| watchPublic.ts:125:23 | createIncrementalCompilerHost | `maybeBind(system, system.createHash)` | rewrite |
| watchPublic.ts:506:56 | createWatchProgram | `maybeBind(host, host.useSourceOfProjectReferenceRedirect)` | rewrite |
| watchPublic.ts:529:46 | createWatchProgram | `maybeBind(host, host.resolveModuleNameLiterals)` | rewrite |
| watchPublic.ts:530:39 | createWatchProgram | `maybeBind(host, host.resolveModuleNames)` | rewrite |
| watchPublic.ts:534:60 | createWatchProgram | `maybeBind(host, host.resolveTypeReferenceDirectiveReferences)` | rewrite |
| watchPublic.ts:535:51 | createWatchProgram | `maybeBind(host, host.resolveTypeReferenceDirectives)` | rewrite |
| watchPublic.ts:543:9 | createWatchProgram | `maybeBind(host, host.getModuleResolutionCache)` | rewrite |
| watchPublic.ts:549:9 | createWatchProgram | `maybeBind(host, host.hasInvalidatedResolutions)` | rewrite |
| watchPublic.ts:552:9 | createWatchProgram | `maybeBind(host, host.hasInvalidatedLibResolutions)` | rewrite |

## Current verification

This ownership revision supersedes the earlier 46-site run. Its new full after
apply and all-adapter paired replay are in progress. The replay records every
outcome before deciding whether any differ; shared refusals are baselines.
It compares all 80 baseline adapters, with 99x's after replay recorded separately.
If clean, types, emitted JavaScript, the one-site negative control, the complete
lane and all 398 shards, and CW051 measurements will be rerun for this scope.
Prior 46-site results are historical and do not certify the new scope.
See PROOF.md and the new owned/ evidence as it is completed.
