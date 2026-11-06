# 99x: explicit host method binding with earlier-owner handoffs

Task kpp0f8s. javascriptChanging true. Pinned TypeScript 6.0.3 source:
050880ce59e30b356b686bd3144efe24f875ebc8. This continuation starts from
adapt-1 a6d5f494 on main 322f8056. The original delivery measured main 3e395f9e5.

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

Baseline without earlier folds: **28 rewritten, 18 left**. Complete earlier
folds reduce the remaining-original count; the 28 own rewrites stay unchanged. Six left sites belong to 95h;
twelve belong to 32. The requested two checker sites carry the exact reason
"owner pinned by 95-checker-axolotl-h; folded there" below. "Folded there"
records the ownership handoff, not an edit by this branch to another adapter.
There are zero unsupported-shape refusals and zero type-driven fallbacks.

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
preserves bytes and mtimes. For every ownership.json row, all original handed-off calls must remain or
all must be absent. A mixed group refuses and names its owner. 99x does not
validate or edit the replacement code in a wholly absent group. No cast is added.

Five source files are eligible to change. Their maybeBind import specifier is
removed exactly when no core maybeBind call remains after 99x. inspect derives
its import requirement from all symbol-resolved core calls in each file,
including parenthesized callees, rather than
the historical inventory.imports list. On main the removed imports are in
builder.ts, tsbuildPublic.ts and watchPublic.ts; a complete 32 fold also lets
99x remove watch.ts's import. A complete program owner fold lets 99x remove
program.ts's import. checker.ts has no 99x site, so 99x never changes or checks
its import. Removing that import after its final use is 95h's fold's job. Python's actual sort places
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

## Historical verification

Checks 1 through 6 pass for 28 rewrites and 18 earlier-owner handoffs.
All 80 baseline replay outcomes match. Types pass over all 79 roots, nine
emitted JavaScript products match the expected changes, and the one-site
negative control is detected. The unchanged whole lane and all 398 standard
shards pass with exactly matching observations and sanctioned API bytes.
All 79 ordinary first stops are unchanged; all six CW051 units were measured
before and after. See PROOF.md and owned/ for complete current evidence.
Original receipts outside owned/ and composition/ describe superseded
46-site experiments. owned/ certifies the earlier 28-site delivery;
composition/ contains the current continuation.

## Fold composition continuation

This continuation starts at typescript/maybebind-adapt-1, a6d5f494, on main
322f8056. The 28 own sites and ten exact owner pins are unchanged. The old
18-left total is now conditional: complete earlier-owner folds reduce it,
without granting 99x ownership of their replacement expressions. Strict own
owner, half-adapted tree and import checks remain. composition/ contains the
new receipts; the prior owned/ receipts certify the historical run only.

The no-write controls cover each partial owner group and restored unused
imports in builder.ts, watch.ts and program.ts. A complete group may contain
other core helper call shapes; 99x counts only inventoried originals when
checking the all-present/all-absent boundary and leaves replacement code opaque. A complete synthetic owner-fold control exercises zero remaining
program/watch calls, import cleanup, checker byte preservation and a replay
that preserves source bytes and nanosecond mtimes. Full apply, type, emitted
JavaScript, oracle and CW051 composition measurements are in progress.
