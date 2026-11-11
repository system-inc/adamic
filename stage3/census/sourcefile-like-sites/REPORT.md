# SourceFileLike sites (baobab)

Measured main: `d690f276749f91e6b36fe3a9726a5a453fcd7d39`. Upstream TypeScript 6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`. Report only; no compiler or adaptation changes.

Attempted 10 of 79 compiler files; 9 completed full-mode streams. Physical sites: 0; distinct finding/phase/unit observations: 0.

## By callee

| Callee | Sites | Files | First-stop roots on c504dd8f | Reads optional method | Cleared by the two proposed signatures |
| --- | ---: | ---: | ---: | --- | --- |

First-stop attribution comes from `typescript/next-stops-c504dd8f` at `fe08a0d47a3d28f646db181ac9e8456e5c60885f`, file `stage3/census/next-stops-c504dd8f/grouped-first-stops.json`, copied as prior-first-stops.json. It names tracing.ts:206:67 for 73 roots. Its other six roots have unrelated outcomes. First-stop roots count only sites explicitly named in that evidence; it does not predict their next blocker.

## What road 2 clears

The exact proposal changes only getLineAndCharacterOfPosition and getLineStarts. Their bodies and the functions receiving their source-file value use only text and lineMap. getLineStarts reads lineMap, writes the cache when absent, and passes only text to computeLineStarts. getLineAndCharacterOfPosition passes the value only to getLineStarts, then passes a number array to computeLineAndCharacterOfPosition and computeLineOfPosition. These paths never inspect the optional method.

getPositionOfLineAndCharacter tests and conditionally calls sourceFile.getPositionOfLineAndCharacter; its fallback uses getLineStarts and text. Widening the two proposed signatures does not remove its optional-method contract. Other safe callees in the inventory would require their own parameter changes if findings are at those call boundaries. WeakMap users read no SourceFileLike members at all, and are called out separately.

SourceFileLike declares optional lineMap and getPositionOfLineAndCharacter. The proposed minimal type retains optional lineMap; SourceFile already supplies lineMap. Each collected reason is preserved below and in findings.json.

## Every observed site

Locations are in the adapted compiler. Body-inventory line numbers in the next table refer to pinned upstream. A physical site is (location, reason), deduplicated across phases and units. sites.json preserves every distinct finding/phase/unit observation and the complete call expression.

| Location | Phase(s) | Enclosing function | Receiving callee / view | Value use | Reason |
| --- | --- | --- | --- | --- | --- |

## All exported compiler functions with a SourceFileLike parameter

This includes @internal exports as well as the published API. Only getPositionOfLineAndCharacter and getLineAndCharacterOfPosition lack @internal. Overload declarations of getPositionOfLineAndCharacter (scanner.ts:465 and :467) share the implementation at :468. SourceFileLike-returning host methods and WeakMap type arguments are not additional functions with such a parameter.

| Upstream location | Function | Internal | Reads optional method, directly or transitively | Body / onward use |
| --- | --- | --- | --- | --- |
| scanner.ts:468 | `getPositionOfLineAndCharacter` | no | reads and conditionally calls getPositionOfLineAndCharacter | sourceFile.getPositionOfLineAndCharacter; otherwise getLineStarts(sourceFile), sourceFile.text -> computePositionOfLineAndCharacter |
| scanner.ts:503 | `getLineStarts` | yes | no | reads lineMap, writes computed lineMap; text -> computeLineStarts |
| scanner.ts:537 | `getLinesBetweenPositions` | yes | no | sourceFile -> getLineStarts; line-start array -> computeLineOfPosition |
| scanner.ts:548 | `getLineAndCharacterOfPosition` | no | no | sourceFile -> getLineStarts; line-start array -> computeLineAndCharacterOfPosition -> computeLineOfPosition |
| utilities.ts:1015 | `getStartPositionOfLine` | yes | no | sourceFile -> getLineStarts |
| utilities.ts:1029 | `getEndLinePosition` | yes | no | sourceFile -> getLineStarts; sourceFile.text; only numbers and strings passed onward |
| utilities.ts:1238 | `getTokenPosOfNode` | yes | no | reads text -> skipTrivia; recursive getTokenPosOfNode; sourceFile -> getNodeChildren (WeakMap key only) |
| utilities.ts:1276 | `getNonDecoratorTokenPosOfNode` | yes | no | sourceFile -> getTokenPosOfNode; otherwise sourceFile.text -> skipTrivia |
| factory/nodeChildren.ts:14 | `getNodeChildren` | yes | no | sourceFile is a WeakMap key; reads no SourceFileLike member |
| factory/nodeChildren.ts:27 | `setNodeChildren` | yes | no | sourceFile is a WeakMap key; reads no SourceFileLike member |
| factory/nodeChildren.ts:44 | `unsetNodeChildren` | yes | no | origSourceFile is a WeakMap key; reads no SourceFileLike member |
| factory/nodeChildren.ts:55 | `transferSourceFileChildren` | yes | no | sourceFile and targetSourceFile are WeakMap keys; reads no SourceFileLike member |

The transitive read audit includes getNodeChildren for the token-position utilities, getLineStarts for the line helpers, and recursive getTokenPosOfNode calls. Remaining downstream functions receive only strings, number arrays, nodes, or scalar values, so cannot read the source-file optional method. See body-evidence.txt.

## Coverage and reproducibility

Piece 1 runs sequentially; subsequent pieces use at most two concurrent independent single-file processes. Each process retains its own selection, stream, timing, and five-minute limit. All file invocations use LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, and LATENT_ONLY_FILE equal to one exact adapted file. The meter scratch overlay disables ordinary Lower and exposes no usable IR. All 79 compiler files remain roots of each loaded program; selection restricts attempted units, not program loading. The same binary and adapted tree serve every invocation. No backend is invoked.

The overlay retains its historical measurement label even when checker_rejected is false. Use the header flag and diagnostics, copied into coverage.json, for the actual checker status. A successful guarded Load test here proves the overlay disables output; it does not imply the source program has checker errors.

This is a latent observational census, not an exhaustive proof of all runtime paths or a compilation result. Checker-diagnosed bodies, failed signatures, boundaries, and relation-overflow effects are retained in the raw streams. Sites hidden behind these remain unknown. A zero exit plus a full header and relation trailer establishes stream completion, not body completeness. The report predicts removal of the identified parameter-view contract only; it does not claim those calls or roots will lower afterward.

| File | Child exit | Seconds | Complete stream | Over five minutes |
| --- | ---: | ---: | --- | --- |
| _namespaces/ts.moduleSpecifiers.ts | 0 | 20.82 | yes | no |
| _namespaces/ts.performance.ts | 0 | 19.31 | yes | no |
| _namespaces/ts.ts | 0 | 38.79 | yes | no |
| binder.ts | 0 | 205.84 | yes | no |
| builder.ts | 0 | 96.46 | yes | no |
| builderPublic.ts | 0 | 25.11 | yes | no |
| builderState.ts | 0 | 63.30 | yes | no |
| builderStatePublic.ts | 0 | 22.21 | yes | no |
| checker.ts | -15 | 300.12 | no | yes |
| commandLineParser.ts | 0 | 131.84 | yes | no |

**Incomplete census.** Counts are observed lower bounds. Missing or timed-out file results must not be treated as zero sites. See coverage.json and raw per-file logs/status files.

Raw per-file findings, timing streams, logs, and child statuses are under raw/. Files over 1,000,000 bytes are gzip compressed. run-piece.py, collect.py, classify.cjs, and report.py reproduce the run and tables; provenance.json records pins. Setup and preparation logs are under preparation/. Checkpoints are snapshot commits on typescript/sourcefile-like-sites-checkpoint; the final report branch contains one commit on the measured main tip.
