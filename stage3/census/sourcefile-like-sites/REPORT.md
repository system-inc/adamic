# SourceFileLike sites (baobab)

Measured main: `d690f276749f91e6b36fe3a9726a5a453fcd7d39`. Upstream TypeScript 6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`. Report only; no compiler or adaptation changes.

Attempted 79 of 79 compiler files; 78 completed full-mode streams. Physical sites: 34; distinct finding/phase/unit observations: 34.

**Counts are lower bounds.** Incomplete streams: `checker.ts`.

## By callee

| Callee | Sites | SourceFile sites | Other source types | Files | First-stop roots on c504dd8f | Reads optional method | Method-only candidate sites for road 2 |
| --- | ---: | ---: | ---: | ---: | ---: | --- | --- |
| `getEndLinePosition` | 1 | 1 | 0 | 1 | 0 | no | 0 |
| `getLineAndCharacterOfPosition` | 18 | 17 | 1 | 7 | 73 | no | 18 |
| `getLineStarts` | 4 | 4 | 0 | 4 | 0 | no | 4 |
| `getLinesBetweenPositions` | 5 | 5 | 0 | 1 | 0 | no | 0 |
| `getPositionOfLineAndCharacter` | 2 | 2 | 0 | 1 | 0 | yes | 0 |
| `transferSourceFileChildren` | 2 | 2 | 0 | 1 | 0 | no | 0 |
| `unsetNodeChildren` | 2 | 2 | 0 | 1 | 0 | no | 0 |

Reason census: 33 site(s): optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; 1 site(s): optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceMapSource, which can hide fields.

First-stop attribution comes from `typescript/next-stops-c504dd8f` at `fe08a0d47a3d28f646db181ac9e8456e5c60885f`, file `stage3/census/next-stops-c504dd8f/grouped-first-stops.json`, copied as prior-first-stops.json. It names tracing.ts:206:67 for 73 roots. Its other six roots have unrelated outcomes. First-stop roots count only sites explicitly named in that evidence; it does not predict their next blocker.

## Initial body-only road 2 hypothesis

Body inspection identifies 22 of 34 observed sites whose optional-method requirement would be removed by the two parameter changes. This was the original report hypothesis, not a measured clearance result. It does not establish that the replacement lineMap view is legal. The follow-up measurements below, when present, take precedence.

The exact proposal changes only getLineAndCharacterOfPosition and getLineStarts. Their bodies and the functions receiving their source-file value use only text and lineMap. getLineStarts reads lineMap, writes the cache when absent, and passes only text to computeLineStarts. getLineAndCharacterOfPosition passes the value only to getLineStarts, then passes a number array to computeLineAndCharacterOfPosition and computeLineOfPosition. These paths never inspect the optional method.

getPositionOfLineAndCharacter tests and conditionally calls sourceFile.getPositionOfLineAndCharacter; its fallback uses getLineStarts and text. Widening the two proposed signatures does not remove its optional-method contract. Other safe callees in the inventory would require their own parameter changes if findings are at those call boundaries. WeakMap users read no SourceFileLike members at all, and are called out separately.

SourceFileLike declares optional lineMap and getPositionOfLineAndCharacter. The proposed minimal type retains optional lineMap; SourceFile already supplies lineMap. Counts include every reason naming SourceFileLike. The by-callee JSON separately records how many sites name structural source SourceFile. Clearance counts refer only to the identified optional-method refusal at a changed signature; a reason concerning retained lineMap or another contract is not marked cleared. Each collected reason is preserved below and in findings.json.

## Every observed site

All observed views are call arguments; no assignment or return view was observed.

Locations are in the adapted compiler. Body-inventory line numbers in the next table refer to pinned upstream. A physical site is (location, reason), deduplicated across phases and units. sites.json preserves every distinct finding/phase/unit observation and the complete call expression.

| Location | Phase(s) | Enclosing function | Receiving callee / view | Value use | Method-only candidate | Reason |
| --- | --- | --- | --- | --- | --- | --- |
| emitter.ts:1444:66 | refusal_scan | createPrinter / getCurrentLineMap | getLineStarts (call argument) | reads lineMap, writes computed lineMap; text -&gt; computeLineStarts | yes | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| emitter.ts:3961:112 | refusal_scan | createPrinter / emitJsxExpression | getLineAndCharacterOfPosition (call argument) | sourceFile -&gt; getLineStarts; line-start array -&gt; computeLineAndCharacterOfPosition -&gt; computeLineOfPosition | yes | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| emitter.ts:3961:180 | refusal_scan | createPrinter / emitJsxExpression | getLineAndCharacterOfPosition (call argument) | sourceFile -&gt; getLineStarts; line-start array -&gt; computeLineAndCharacterOfPosition -&gt; computeLineOfPosition | yes | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| emitter.ts:6233:96 | refusal_scan | createPrinter / emitPos | getLineAndCharacterOfPosition (call argument) | sourceFile -&gt; getLineStarts; line-start array -&gt; computeLineAndCharacterOfPosition -&gt; computeLineOfPosition | yes | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceMapSource, which can hide fields |
| executeCommandLine.ts:109:41 | refusal_scan | countLines / &lt;anonymous@107&gt; | getLineStarts (call argument) | reads lineMap, writes computed lineMap; text -&gt; computeLineStarts | yes | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| parser.ts:10032:36 | refusal_scan | updateSourceFile | transferSourceFileChildren (call argument) | sourceFile and targetSourceFile are WeakMap keys; reads no SourceFileLike member | no | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| parser.ts:10032:48 | refusal_scan | updateSourceFile | transferSourceFileChildren (call argument) | sourceFile and targetSourceFile are WeakMap keys; reads no SourceFileLike member | no | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| parser.ts:10104:37 | refusal_scan | moveElementEntirelyPastChangeRange / visitNode | unsetNodeChildren (call argument) | origSourceFile is a WeakMap key; reads no SourceFileLike member | no | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| parser.ts:10261:42 | refusal_scan | updateTokenPositionsAndMarkElements / visitNode | unsetNodeChildren (call argument) | origSourceFile is a WeakMap key; reads no SourceFileLike member | no | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| program.ts:671:67 | refusal_scan | formatDiagnostic | getLineAndCharacterOfPosition (call argument) | sourceFile -&gt; getLineStarts; line-start array -&gt; computeLineAndCharacterOfPosition -&gt; computeLineOfPosition | yes | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| program.ts:713:89 | refusal_scan | formatCodeSpan | getLineAndCharacterOfPosition (call argument) | sourceFile -&gt; getLineStarts; line-start array -&gt; computeLineAndCharacterOfPosition -&gt; computeLineOfPosition | yes | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| program.ts:714:87 | refusal_scan | formatCodeSpan | getLineAndCharacterOfPosition (call argument) | sourceFile -&gt; getLineStarts; line-start array -&gt; computeLineAndCharacterOfPosition -&gt; computeLineOfPosition | yes | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| program.ts:715:58 | refusal_scan | formatCodeSpan | getLineAndCharacterOfPosition (call argument) | sourceFile -&gt; getLineStarts; line-start array -&gt; computeLineAndCharacterOfPosition -&gt; computeLineOfPosition | yes | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| program.ts:733:57 | refusal_scan | formatCodeSpan | getPositionOfLineAndCharacter (call argument) | sourceFile.getPositionOfLineAndCharacter; otherwise getLineStarts(sourceFile), sourceFile.text -&gt; computePositionOfLineAndCharacter | no | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| program.ts:734:76 | refusal_scan | formatCodeSpan | getPositionOfLineAndCharacter (call argument) | sourceFile.getPositionOfLineAndCharacter; otherwise getLineStarts(sourceFile), sourceFile.text -&gt; computePositionOfLineAndCharacter | no | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| program.ts:768:89 | refusal_scan | formatLocation | getLineAndCharacterOfPosition (call argument) | sourceFile -&gt; getLineStarts; line-start array -&gt; computeLineAndCharacterOfPosition -&gt; computeLineOfPosition | yes | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| program.ts:2973:42 | refusal_scan | createProgram / markPrecedingCommentDirectiveLine | getLineStarts (call argument) | reads lineMap, writes computed lineMap; text -&gt; computeLineStarts | yes | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| tracing.ts:206:67 | refusal_scan | getLocation | getLineAndCharacterOfPosition (call argument) | sourceFile -&gt; getLineStarts; line-start array -&gt; computeLineAndCharacterOfPosition -&gt; computeLineOfPosition | yes | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| tracing.ts:207:65 | refusal_scan | getLocation | getLineAndCharacterOfPosition (call argument) | sourceFile -&gt; getLineStarts; line-start array -&gt; computeLineAndCharacterOfPosition -&gt; computeLineOfPosition | yes | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| transformers/declarations.ts:1264:70 | refusal_scan | transformDeclarations / visitDeclarationSubtree | getLineAndCharacterOfPosition (call argument) | sourceFile -&gt; getLineStarts; line-start array -&gt; computeLineAndCharacterOfPosition -&gt; computeLineOfPosition | yes | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| transformers/declarations.ts:1264:139 | refusal_scan | transformDeclarations / visitDeclarationSubtree | getLineAndCharacterOfPosition (call argument) | sourceFile -&gt; getLineStarts; line-start array -&gt; computeLineAndCharacterOfPosition -&gt; computeLineOfPosition | yes | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| transformers/jsx.ts:359:63 | refusal_scan | transformJsx / visitJsxOpeningLikeElementOrFragmentJSX | getLineAndCharacterOfPosition (call argument) | sourceFile -&gt; getLineStarts; line-start array -&gt; computeLineAndCharacterOfPosition -&gt; computeLineOfPosition | yes | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| utilities.ts:1025:47 | refusal_scan | nodePosToString | getLineAndCharacterOfPosition (call argument) | sourceFile -&gt; getLineStarts; line-start array -&gt; computeLineAndCharacterOfPosition -&gt; computeLineOfPosition | yes | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| utilities.ts:1213:46 | refusal_scan | createCommentDirectivesMap / &lt;anonymous@1212&gt; | getLineAndCharacterOfPosition (call argument) | sourceFile -&gt; getLineStarts; line-start array -&gt; computeLineAndCharacterOfPosition -&gt; computeLineOfPosition | yes | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| utilities.ts:2526:67 | refusal_scan | getErrorSpanForArrowFunction | getLineAndCharacterOfPosition (call argument) | sourceFile -&gt; getLineStarts; line-start array -&gt; computeLineAndCharacterOfPosition -&gt; computeLineOfPosition | yes | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| utilities.ts:2527:65 | refusal_scan | getErrorSpanForArrowFunction | getLineAndCharacterOfPosition (call argument) | sourceFile -&gt; getLineStarts; line-start array -&gt; computeLineAndCharacterOfPosition -&gt; computeLineOfPosition | yes | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| utilities.ts:2531:70 | refusal_scan | getErrorSpanForArrowFunction | getEndLinePosition (call argument) | sourceFile -&gt; getLineStarts; sourceFile.text; only numbers and strings passed onward | no | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| utilities.ts:6748:38 | refusal_scan | getLineOfLocalPosition | getLineStarts (call argument) | reads lineMap, writes computed lineMap; text -&gt; computeLineStarts | yes | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| utilities.ts:7944:37 | refusal_scan | getLinesBetweenRangeEndAndRangeStart | getLinesBetweenPositions (call argument) | sourceFile -&gt; getLineStarts; line-start array -&gt; computeLineOfPosition | no | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| utilities.ts:7949:37 | refusal_scan | getLinesBetweenRangeEndPositions | getLinesBetweenPositions (call argument) | sourceFile -&gt; getLineStarts; line-start array -&gt; computeLineOfPosition | no | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| utilities.ts:7959:37 | refusal_scan | positionsAreOnSameLine | getLinesBetweenPositions (call argument) | sourceFile -&gt; getLineStarts; line-start array -&gt; computeLineOfPosition | no | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| utilities.ts:7971:37 | refusal_scan | getLinesBetweenPositionAndPrecedingNonWhitespaceCharacter | getLinesBetweenPositions (call argument) | sourceFile -&gt; getLineStarts; line-start array -&gt; computeLineOfPosition | no | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| utilities.ts:7977:37 | refusal_scan | getLinesBetweenPositionAndNextNonWhitespaceCharacter | getLinesBetweenPositions (call argument) | sourceFile -&gt; getLineStarts; line-start array -&gt; computeLineOfPosition | no | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| watch.ts:249:60 | refusal_scan | getFilesInErrorForSummary / &lt;anonymous@241&gt; | getLineAndCharacterOfPosition (call argument) | sourceFile -&gt; getLineStarts; line-start array -&gt; computeLineAndCharacterOfPosition -&gt; computeLineOfPosition | yes | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |

## All exported compiler functions with a SourceFileLike parameter

The AST inventory was checked against both pinned upstream and the fully adapted compiler: 12 implementations and 14 declarations including overloads. This function inventory and body audit is complete even though the site census has cut-off streams. This includes @internal exports as well as the published API. Only getPositionOfLineAndCharacter and getLineAndCharacterOfPosition lack @internal. Overload declarations of getPositionOfLineAndCharacter (scanner.ts:465 and :467) share the implementation at :468. SourceFileLike-returning host methods and WeakMap type arguments are not additional functions with such a parameter.

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

The first pass used one process for piece 1 and at most two for subsequent pieces. The follow-up six-file baseline runs sequentially with a 1500-second bound; completed follow-up streams supersede the corresponding five-minute streams in these counts. Original streams remain under raw/. Each process retains its own selection, stream, timing, and five-minute limit. All file invocations use LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, and LATENT_ONLY_FILE equal to one exact adapted file. The meter scratch overlay disables ordinary Lower and exposes no usable IR. All 79 compiler files remain roots of each loaded program; selection restricts attempted units, not program loading. The same binary and adapted tree serve every invocation. No backend is invoked.

The overlay retains its historical measurement label even when checker_rejected is false. Use the header flag and diagnostics, copied into coverage.json, for the actual checker status. A successful guarded Load test here proves the overlay disables output; it does not imply the source program has checker errors.

This is a latent observational census, not an exhaustive proof of all runtime paths or a compilation result. Checker-diagnosed bodies, failed signatures, boundaries, and relation-overflow effects are retained in the raw streams. Sites hidden behind these remain unknown. A zero exit plus a full header and relation trailer establishes stream completion, not body completeness. The initial table identifies removal of the optional-method requirement only; it does not claim those calls or roots will lower afterward.

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
| checker.ts | -15 | 1500.23 | no | yes |
| commandLineParser.ts | 0 | 131.84 | yes | no |
| core.ts | 0 | 39.20 | yes | no |
| corePublic.ts | 0 | 25.18 | yes | no |
| debug.ts | 0 | 50.04 | yes | no |
| diagnosticInformationMap.generated.ts | 0 | 28.94 | yes | no |
| emitter.ts | 0 | 479.19 | yes | yes |
| executeCommandLine.ts | 0 | 133.88 | yes | no |
| expressionToTypeNode.ts | 0 | 109.09 | yes | no |
| factory/baseNodeFactory.ts | 0 | 23.66 | yes | no |
| factory/emitHelpers.ts | 0 | 60.04 | yes | no |
| factory/emitNode.ts | 0 | 30.66 | yes | no |
| factory/nodeChildren.ts | 0 | 30.26 | yes | no |
| factory/nodeConverters.ts | 0 | 50.43 | yes | no |
| factory/nodeFactory.ts | 0 | 292.41 | yes | no |
| factory/nodeTests.ts | 0 | 110.43 | yes | no |
| factory/parenthesizerRules.ts | 0 | 57.96 | yes | no |
| factory/utilities.ts | 0 | 97.59 | yes | no |
| factory/utilitiesPublic.ts | 0 | 23.45 | yes | no |
| hostErrors.ts | 0 | 22.47 | yes | no |
| moduleNameResolver.ts | 0 | 198.43 | yes | no |
| moduleSpecifiers.ts | 0 | 65.49 | yes | no |
| parser.ts | 0 | 722.02 | yes | yes |
| path.ts | 0 | 54.93 | yes | no |
| performance.ts | 0 | 32.63 | yes | no |
| performanceCore.ts | 0 | 25.52 | yes | no |
| program.ts | 0 | 729.78 | yes | yes |
| programDiagnostics.ts | 0 | 47.93 | yes | no |
| resolutionCache.ts | 0 | 65.13 | yes | no |
| scanner.ts | 0 | 253.30 | yes | no |
| semver.ts | 0 | 60.31 | yes | no |
| sourcemap.ts | 0 | 67.29 | yes | no |
| symbolWalker.ts | 0 | 47.18 | yes | no |
| sys.ts | 0 | 122.74 | yes | no |
| tracing.ts | 0 | 37.34 | yes | no |
| transformer.ts | 0 | 33.97 | yes | no |
| transformers/classFields.ts | 0 | 178.12 | yes | no |
| transformers/classThis.ts | 0 | 23.18 | yes | no |
| transformers/declarations.ts | 0 | 130.53 | yes | no |
| transformers/declarations/diagnostics.ts | 0 | 71.27 | yes | no |
| transformers/destructuring.ts | 0 | 55.35 | yes | no |
| transformers/es2015.ts | 0 | 211.99 | yes | no |
| transformers/es2016.ts | 0 | 38.06 | yes | no |
| transformers/es2017.ts | 0 | 72.33 | yes | no |
| transformers/es2018.ts | 0 | 93.46 | yes | no |
| transformers/es2019.ts | 0 | 32.65 | yes | no |
| transformers/es2020.ts | 0 | 68.69 | yes | no |
| transformers/es2021.ts | 0 | 30.09 | yes | no |
| transformers/esDecorators.ts | 0 | 153.10 | yes | no |
| transformers/esnext.ts | 0 | 87.89 | yes | no |
| transformers/generators.ts | 0 | 169.97 | yes | no |
| transformers/jsx.ts | 0 | 107.95 | yes | no |
| transformers/legacyDecorators.ts | 0 | 117.01 | yes | no |
| transformers/module/esnextAnd2015.ts | 0 | 83.96 | yes | no |
| transformers/module/impliedNodeFormatDependent.ts | 0 | 31.52 | yes | no |
| transformers/module/module.ts | 0 | 131.97 | yes | no |
| transformers/module/system.ts | 0 | 168.17 | yes | no |
| transformers/namedEvaluation.ts | 0 | 30.90 | yes | no |
| transformers/taggedTemplate.ts | 0 | 25.05 | yes | no |
| transformers/ts.ts | 0 | 183.42 | yes | no |
| transformers/typeSerializer.ts | 0 | 65.98 | yes | no |
| transformers/utilities.ts | 0 | 103.87 | yes | no |
| tsbuild.ts | 0 | 32.97 | yes | no |
| tsbuildPublic.ts | 0 | 72.01 | yes | no |
| types.ts | 0 | 94.51 | yes | no |
| utilities.ts | 0 | 643.31 | yes | yes |
| utilitiesPublic.ts | 0 | 207.54 | yes | no |
| visitorPublic.ts | 0 | 37.85 | yes | no |
| watch.ts | 0 | 52.46 | yes | no |
| watchPublic.ts | 0 | 23.93 | yes | no |
| watchUtilities.ts | 0 | 67.72 | yes | no |

**Incomplete census.** Counts are observed lower bounds. Missing or timed-out file results must not be treated as zero sites. See coverage.json and raw per-file logs/status files.

Raw per-file findings, timing streams, logs, and child statuses are under raw/. Files over 1,000,000 bytes are gzip compressed. run-piece.py, collect.py, classify.cjs, and report.py reproduce the run and tables; provenance.json records pins. Setup and preparation logs are under preparation/. Checkpoints are snapshot commits on typescript/sourcefile-like-sites-checkpoint; the first report is one commit on the measured main tip; later requested measurements append report-only commits.

## Follow-up: measured road 2 variants

The follow-up uses the same main pin, upstream pin, frozen adapted input, and guarded binary as the initial report. The six old cut-offs were retried first, sequentially, with a 1500-second bound each. Only after all six attempts and their checkpoints completed were the variants applied or measured.

### Six longer baseline attempts

| File | Child exit | Seconds | Complete file result | Completed timing records | Last completed timed unit |
| --- | ---: | ---: | --- | ---: | --- |
| checker.ts | -15 | 1500.23 | no | 1293 | checker.ts:27590:5 |
| emitter.ts | 0 | 479.19 | yes | 423 | emitter.ts:6370:1 |
| parser.ts | 0 | 722.02 | yes | 545 | parser.ts:10800:1 |
| program.ts | 0 | 729.78 | yes | 235 | program.ts:5191:1 |
| scanner.ts | 0 | 253.30 | yes | 150 | scanner.ts:4103:1 |
| utilities.ts | 0 | 643.31 | yes | 867 | utilities.ts:12436:1 |

**The count remains a lower bound because these files still did not finish in 25 minutes: checker.ts.** Their timing progress is recorded above and their partial streams are retained. A partial stream without an emitted file result does not mean zero sites.

Successful longer streams supersede their first-pass streams in the primary count tables. Both sets of raw streams remain in the evidence. The complete function/body inventory is independent of whether a file census finishes.

### Exact variants and measured outcomes

(a) changes the first parameter of both helpers to `{ readonly text: string; lineMap?: readonly number[] }`.

(b) changes it to `{ readonly text: string; lineMap: readonly number[] | undefined }`.

Each scratch tree was produced with the pinned apply script/product before replacing those two annotations. Full src-file hashes prove that scanner.ts is the only changed source file; byte comparisons prove that its only edits are the two parameter types. No SourceFile or services interface was changed. SourceFile.lineMap in this adapted tree is required and has the exact type `readonly number[] | undefined` (compiler/types.ts:4436). SourceFileLike.lineMap is optional. The upstream SourceFile member is required readonly number[]; the unchanged applied baseline adds the undefined union. The applied SourceMapSource.lineMap is optional too. lineMap-declarations.txt quotes both source versions. These are baseline adaptations, not changes made by either variant.

Adamic enables exactOptionalPropertyTypes (internal/load/load.go:63; load-options.go.txt records the options). Its (a) core check reports TS2379 because the adapted required SourceFile.lineMap can be undefined while (a) does not explicitly permit undefined when the optional property is present. The stock whole-src checks below retain the repository project defaults, where exactOptionalPropertyTypes is false. Thus a zero stock (a) error count does not imply that (a) passes Adamic. The core records and guarded-census diagnostics use the actual Adamic settings.

The variants use separate single-file processes, at most one process for each variant concurrently. Each retains all 79 compiler roots and uses LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, and an exact LATENT_ONLY_FILE. Each variant file has the same 1500-second safety bound. Checkpoints follow every completed pair of file attempts.

| Variant | Completed file streams | Baseline view locations | Unchanged refusal at same view | Replacement finding at same view | Checker diagnostic at same call | No finding there, unit attempted | Other / incomplete | Whole-src stock errors | Services TS2386 | Core first stop |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| (a) | 9/9 | 34 | 9 | 0 | 22 | 0 | 3 | 0 | 0 | checker |
| (b) | 9/9 | 34 | 12 | 0 | 1 | 21 | 0 | 15 | 0 | checker |

These columns report observations at the original views, not successful lowering. A replacement finding is a different kind or reason at the same call/argument span. Unchanged findings include unmodified SourceFileLike callees. A call with two baseline SourceFileLike arguments is compared by argument span so findings are not attributed twice. A missing finding is not counted as a clearance if the file stream is incomplete or its enclosing unit is excluded. Other blockers elsewhere remain in the complete raw streams.

The original core.ts first stop and both variant first stops follow. The ordinary census driver emits a per-root record and then its one-root whole-program record; the first record is the requested per-root first stop. It never invokes a backend.

Core root original:

````json
{
  "roots": [
    "/workspace/scratch/baobab-adapted/src/compiler/core.ts"
  ],
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-adapted/src/compiler/tracing.ts:206:67",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "fix": "declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
  "seconds": 14.400678442
}
````


Core root a:

````json
{
  "roots": [
    "/workspace/scratch/baobab-road2-a/src/compiler/core.ts"
  ],
  "diagnostics": [
    "/workspace/scratch/baobab-road2-a/src/compiler/checker.ts:52424:57: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/checker.ts:52425:55: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts:1444:66: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts:3961:112: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts:3961:180: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts:6233:96: error TS2379: Argument of type 'SourceMapSource' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/executeCommandLine.ts:109:41: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:2973:42: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:671:67: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:713:89: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:714:87: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:715:58: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:768:89: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/scanner.ts:471:57: error TS2379: Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/scanner.ts:539:38: error TS2379: Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/tracing.ts:206:67: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/tracing.ts:207:65: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/transformers/declarations.ts:1264:139: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/transformers/declarations.ts:1264:70: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/transformers/jsx.ts:359:63: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:1018:26: error TS2379: Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:1025:47: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:1032:38: error TS2379: Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:1213:46: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:2526:67: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:2527:65: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:6748:38: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
    "/workspace/scratch/baobab-road2-a/src/compiler/watch.ts:249:60: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
  ],
  "kind": "checker",
  "seconds": 5.314485692
}
````


Core root b:

````json
{
  "roots": [
    "/workspace/scratch/baobab-road2-b/src/compiler/core.ts"
  ],
  "diagnostics": [
    "/workspace/scratch/baobab-road2-b/src/compiler/emitter.ts:6233:96: error TS2345: Argument of type 'SourceMapSource' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceMapSource' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
    "/workspace/scratch/baobab-road2-b/src/compiler/scanner.ts:471:57: error TS2345: Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceFileLike' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
    "/workspace/scratch/baobab-road2-b/src/compiler/scanner.ts:539:38: error TS2345: Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceFileLike' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
    "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:1018:26: error TS2345: Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceFileLike' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
    "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:1032:38: error TS2345: Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceFileLike' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'."
  ],
  "kind": "checker",
  "seconds": 3.924856896
}
````


### Every original view under each variant

The baseline table above names every observed original SourceFileLike view; the checker cutoff remains unknown. The blocks below quote every distinct finding matching each original call/argument span, including reasons that no longer name SourceFileLike. Checker diagnostics at those calls are quoted too. Unit status and stream completeness distinguish an unreported view from an excluded or unfinished measurement. Finding identities are deduplicated by kind, where, reason, text, phase, and unit; the raw streams retain repeated observations.

#### Variant (a)

**emitter.ts:1444:66**, `getLineStarts`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts",
  "start": 58857,
  "end": 58894,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts:1444:66: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "emitter.ts:1444:66"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts:1443:5",
    "name": "getCurrentLineMap",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts:1444:66: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**emitter.ts:3961:112**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts",
  "start": 171010,
  "end": 171027,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts:3961:112: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "emitter.ts:3961:112"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts:3959:5",
    "name": "emitJsxExpression",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts:3961:112: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts:3961:180: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**emitter.ts:3961:180**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts",
  "start": 171078,
  "end": 171095,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts:3961:180: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "emitter.ts:3961:180"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts:3959:5",
    "name": "emitJsxExpression",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts:3961:112: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts:3961:180: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**emitter.ts:6233:96**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts",
  "start": 268874,
  "end": 268889,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts:6233:96: error TS2379: Argument of type 'SourceMapSource' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "emitter.ts:6233:96"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts:6228:5",
    "name": "emitPos",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/emitter.ts:6233:96: error TS2379: Argument of type 'SourceMapSource' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**executeCommandLine.ts:109:41**, `getLineStarts`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/executeCommandLine.ts",
  "start": 2883,
  "end": 2887,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/executeCommandLine.ts:109:41: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "executeCommandLine.ts:109:41"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/executeCommandLine.ts:105:1",
    "name": "countLines",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/executeCommandLine.ts:109:41: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**parser.ts:10032:36**, `transferSourceFileChildren`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-a/src/compiler/parser.ts:9947:5",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-a/src/compiler/parser.ts:10032:36",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/parser.ts:10032:36: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/parser.ts:9947:5",
    "name": "updateSourceFile",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**parser.ts:10032:48**, `transferSourceFileChildren`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-a/src/compiler/parser.ts:9947:5",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-a/src/compiler/parser.ts:10032:48",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/parser.ts:10032:48: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/parser.ts:9947:5",
    "name": "updateSourceFile",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**parser.ts:10104:37**, `unsetNodeChildren`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-a/src/compiler/parser.ts:10096:9",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-a/src/compiler/parser.ts:10104:37",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/parser.ts:10104:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/parser.ts:10096:9",
    "name": "visitNode",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**parser.ts:10261:42**, `unsetNodeChildren`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-a/src/compiler/parser.ts:10246:9",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-a/src/compiler/parser.ts:10261:42",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/parser.ts:10261:42: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/parser.ts:10246:9",
    "name": "visitNode",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**program.ts:671:67**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts",
  "start": 24412,
  "end": 24427,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:671:67: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "program.ts:671:67"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:667:1",
    "name": "formatDiagnostic",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:671:67: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**program.ts:713:89**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts",
  "start": 26154,
  "end": 26158,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:713:89: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "program.ts:713:89"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:712:1",
    "name": "formatCodeSpan",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:713:89: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:714:87: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:715:58: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**program.ts:714:87**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts",
  "start": 26255,
  "end": 26259,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:714:87: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "program.ts:714:87"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:712:1",
    "name": "formatCodeSpan",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:713:89: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:714:87: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:715:58: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**program.ts:715:58**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts",
  "start": 26336,
  "end": 26340,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:715:58: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "program.ts:715:58"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:712:1",
    "name": "formatCodeSpan",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:713:89: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:714:87: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:715:58: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**program.ts:733:57**, `getPositionOfLineAndCharacter`: enclosing unit excluded or not attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:712:1",
    "name": "formatCodeSpan",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:713:89: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:714:87: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:715:58: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**program.ts:734:76**, `getPositionOfLineAndCharacter`: enclosing unit excluded or not attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:712:1",
    "name": "formatCodeSpan",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:713:89: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:714:87: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:715:58: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**program.ts:768:89**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts",
  "start": 29082,
  "end": 29086,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:768:89: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "program.ts:768:89"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:767:1",
    "name": "formatLocation",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:768:89: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**program.ts:2973:42**, `getLineStarts`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts",
  "start": 145144,
  "end": 145148,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:2973:42: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "program.ts:2973:42"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:2966:5",
    "name": "markPrecedingCommentDirectiveLine",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/program.ts:2973:42: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**tracing.ts:206:67**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/tracing.ts",
  "start": 8172,
  "end": 8176,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/tracing.ts:206:67: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "tracing.ts:206:67"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/tracing.ts:200:5",
    "name": "getLocation",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/tracing.ts:206:67: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/tracing.ts:207:65: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**tracing.ts:207:65**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/tracing.ts",
  "start": 8256,
  "end": 8260,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/tracing.ts:207:65: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "tracing.ts:207:65"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/tracing.ts:200:5",
    "name": "getLocation",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/tracing.ts:206:67: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/tracing.ts:207:65: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**transformers/declarations.ts:1264:70**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/transformers/declarations.ts",
  "start": 62455,
  "end": 62472,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/transformers/declarations.ts:1264:70: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "transformers/declarations.ts:1264:70"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/transformers/declarations.ts:1006:5",
    "name": "visitDeclarationSubtree",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/transformers/declarations.ts:1264:70: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/transformers/declarations.ts:1264:139: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**transformers/declarations.ts:1264:139**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/transformers/declarations.ts",
  "start": 62524,
  "end": 62541,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/transformers/declarations.ts:1264:139: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "transformers/declarations.ts:1264:139"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/transformers/declarations.ts:1006:5",
    "name": "visitDeclarationSubtree",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/transformers/declarations.ts:1264:70: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/transformers/declarations.ts:1264:139: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**transformers/jsx.ts:359:63**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/transformers/jsx.ts",
  "start": 16101,
  "end": 16113,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/transformers/jsx.ts:359:63: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "transformers/jsx.ts:359:63"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/transformers/jsx.ts:333:5",
    "name": "visitJsxOpeningLikeElementOrFragmentJSX",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/transformers/jsx.ts:359:63: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**utilities.ts:1025:47**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts",
  "start": 33578,
  "end": 33582,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:1025:47: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "utilities.ts:1025:47"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:1023:1",
    "name": "nodePosToString",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:1025:47: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**utilities.ts:1213:46**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts",
  "start": 42435,
  "end": 42445,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:1213:46: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "utilities.ts:1213:46"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:1210:1",
    "name": "createCommentDirectivesMap",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:1213:46: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**utilities.ts:2526:67**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts",
  "start": 88023,
  "end": 88033,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:2526:67: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "utilities.ts:2526:67"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:2523:1",
    "name": "getErrorSpanForArrowFunction",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:2526:67: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:2527:65: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**utilities.ts:2527:65**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts",
  "start": 88116,
  "end": 88126,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:2527:65: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "utilities.ts:2527:65"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:2523:1",
    "name": "getErrorSpanForArrowFunction",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:2526:67: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:2527:65: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**utilities.ts:2531:70**, `getEndLinePosition`: enclosing unit excluded or not attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:2523:1",
    "name": "getErrorSpanForArrowFunction",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:2526:67: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
      "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:2527:65: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**utilities.ts:6748:38**, `getLineStarts`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts",
  "start": 266275,
  "end": 266285,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:6748:38: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "utilities.ts:6748:38"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:6747:1",
    "name": "getLineOfLocalPosition",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:6748:38: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


**utilities.ts:7944:37**, `getLinesBetweenPositions`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7942:1",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7944:37",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7944:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7942:1",
    "name": "getLinesBetweenRangeEndAndRangeStart",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**utilities.ts:7949:37**, `getLinesBetweenPositions`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7948:1",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7949:37",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7949:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7948:1",
    "name": "getLinesBetweenRangeEndPositions",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**utilities.ts:7959:37**, `getLinesBetweenPositions`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7958:1",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7959:37",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7959:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7958:1",
    "name": "positionsAreOnSameLine",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**utilities.ts:7971:37**, `getLinesBetweenPositions`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7968:1",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7971:37",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7971:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7968:1",
    "name": "getLinesBetweenPositionAndPrecedingNonWhitespaceCharacter",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**utilities.ts:7977:37**, `getLinesBetweenPositions`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7975:1",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7977:37",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7977:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/utilities.ts:7975:1",
    "name": "getLinesBetweenPositionAndNextNonWhitespaceCharacter",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**watch.ts:249:60**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-a/src/compiler/watch.ts",
  "start": 8760,
  "end": 8787,
  "text": "/workspace/scratch/baobab-road2-a/src/compiler/watch.ts:249:60: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'.",
  "mapped_location": "watch.ts:249:60"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-a/src/compiler/watch.ts:233:1",
    "name": "getFilesInErrorForSummary",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-a/src/compiler/watch.ts:249:60: error TS2379: Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap?: readonly number[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.\n  Types of property 'lineMap' are incompatible.\n    Type 'readonly number[] | undefined' is not assignable to type 'readonly number[]'.\n      Type 'undefined' is not assignable to type 'readonly number[]'."
    ]
  }
]
````


#### Variant (b)

**emitter.ts:1444:66**, `getLineStarts`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/emitter.ts:1443:5",
    "name": "getCurrentLineMap",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**emitter.ts:3961:112**, `getLineAndCharacterOfPosition`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/emitter.ts:3959:5",
    "name": "emitJsxExpression",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**emitter.ts:3961:180**, `getLineAndCharacterOfPosition`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/emitter.ts:3959:5",
    "name": "emitJsxExpression",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**emitter.ts:6233:96**, `getLineAndCharacterOfPosition`: checker diagnostic at original call.

No census finding was emitted at this original view.

````json
{
  "file": "src/compiler/emitter.ts",
  "line": 6233,
  "column": 96,
  "start": 268874,
  "length": 15,
  "code": 2345,
  "category": "Error",
  "message": "Argument of type 'SourceMapSource' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceMapSource' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
  "related": []
}
````


The guarded census checker recorded:

````json
{
  "measurement": "measured on a checker-rejected program",
  "file": "/workspace/scratch/baobab-road2-b/src/compiler/emitter.ts",
  "start": 268874,
  "end": 268889,
  "text": "/workspace/scratch/baobab-road2-b/src/compiler/emitter.ts:6233:96: error TS2345: Argument of type 'SourceMapSource' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceMapSource' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
  "mapped_location": "emitter.ts:6233:96"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/emitter.ts:6228:5",
    "name": "emitPos",
    "status": "split_checker_body",
    "checker_diagnostics": [
      "/workspace/scratch/baobab-road2-b/src/compiler/emitter.ts:6233:96: error TS2345: Argument of type 'SourceMapSource' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceMapSource' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'."
    ]
  }
]
````


**executeCommandLine.ts:109:41**, `getLineStarts`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/executeCommandLine.ts:105:1",
    "name": "countLines",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**parser.ts:10032:36**, `transferSourceFileChildren`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-b/src/compiler/parser.ts:9947:5",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-b/src/compiler/parser.ts:10032:36",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-b/src/compiler/parser.ts:10032:36: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/parser.ts:9947:5",
    "name": "updateSourceFile",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**parser.ts:10032:48**, `transferSourceFileChildren`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-b/src/compiler/parser.ts:9947:5",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-b/src/compiler/parser.ts:10032:48",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-b/src/compiler/parser.ts:10032:48: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/parser.ts:9947:5",
    "name": "updateSourceFile",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**parser.ts:10104:37**, `unsetNodeChildren`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-b/src/compiler/parser.ts:10096:9",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-b/src/compiler/parser.ts:10104:37",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-b/src/compiler/parser.ts:10104:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/parser.ts:10096:9",
    "name": "visitNode",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**parser.ts:10261:42**, `unsetNodeChildren`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-b/src/compiler/parser.ts:10246:9",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-b/src/compiler/parser.ts:10261:42",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-b/src/compiler/parser.ts:10261:42: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/parser.ts:10246:9",
    "name": "visitNode",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**program.ts:671:67**, `getLineAndCharacterOfPosition`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/program.ts:667:1",
    "name": "formatDiagnostic",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**program.ts:713:89**, `getLineAndCharacterOfPosition`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/program.ts:712:1",
    "name": "formatCodeSpan",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**program.ts:714:87**, `getLineAndCharacterOfPosition`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/program.ts:712:1",
    "name": "formatCodeSpan",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**program.ts:715:58**, `getLineAndCharacterOfPosition`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/program.ts:712:1",
    "name": "formatCodeSpan",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**program.ts:733:57**, `getPositionOfLineAndCharacter`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-b/src/compiler/program.ts:712:1",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-b/src/compiler/program.ts:733:57",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-b/src/compiler/program.ts:733:57: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/program.ts:712:1",
    "name": "formatCodeSpan",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**program.ts:734:76**, `getPositionOfLineAndCharacter`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-b/src/compiler/program.ts:712:1",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-b/src/compiler/program.ts:734:76",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-b/src/compiler/program.ts:734:76: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/program.ts:712:1",
    "name": "formatCodeSpan",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**program.ts:768:89**, `getLineAndCharacterOfPosition`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/program.ts:767:1",
    "name": "formatLocation",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**program.ts:2973:42**, `getLineStarts`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/program.ts:2966:5",
    "name": "markPrecedingCommentDirectiveLine",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**tracing.ts:206:67**, `getLineAndCharacterOfPosition`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/tracing.ts:200:5",
    "name": "getLocation",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**tracing.ts:207:65**, `getLineAndCharacterOfPosition`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/tracing.ts:200:5",
    "name": "getLocation",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**transformers/declarations.ts:1264:70**, `getLineAndCharacterOfPosition`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/transformers/declarations.ts:1006:5",
    "name": "visitDeclarationSubtree",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**transformers/declarations.ts:1264:139**, `getLineAndCharacterOfPosition`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/transformers/declarations.ts:1006:5",
    "name": "visitDeclarationSubtree",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**transformers/jsx.ts:359:63**, `getLineAndCharacterOfPosition`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/transformers/jsx.ts:333:5",
    "name": "visitJsxOpeningLikeElementOrFragmentJSX",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**utilities.ts:1025:47**, `getLineAndCharacterOfPosition`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:1023:1",
    "name": "nodePosToString",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**utilities.ts:1213:46**, `getLineAndCharacterOfPosition`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:1210:1",
    "name": "createCommentDirectivesMap",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**utilities.ts:2526:67**, `getLineAndCharacterOfPosition`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:2523:1",
    "name": "getErrorSpanForArrowFunction",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**utilities.ts:2527:65**, `getLineAndCharacterOfPosition`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:2523:1",
    "name": "getErrorSpanForArrowFunction",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**utilities.ts:2531:70**, `getEndLinePosition`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:2523:1",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:2531:70",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:2531:70: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:2523:1",
    "name": "getErrorSpanForArrowFunction",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**utilities.ts:6748:38**, `getLineStarts`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:6747:1",
    "name": "getLineOfLocalPosition",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**utilities.ts:7944:37**, `getLinesBetweenPositions`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7942:1",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7944:37",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7944:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7942:1",
    "name": "getLinesBetweenRangeEndAndRangeStart",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**utilities.ts:7949:37**, `getLinesBetweenPositions`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7948:1",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7949:37",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7949:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7948:1",
    "name": "getLinesBetweenRangeEndPositions",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**utilities.ts:7959:37**, `getLinesBetweenPositions`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7958:1",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7959:37",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7959:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7958:1",
    "name": "positionsAreOnSameLine",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**utilities.ts:7971:37**, `getLinesBetweenPositions`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7968:1",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7971:37",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7971:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7968:1",
    "name": "getLinesBetweenPositionAndPrecedingNonWhitespaceCharacter",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**utilities.ts:7977:37**, `getLinesBetweenPositions`: finding at original call.

````json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7975:1",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7977:37",
  "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
  "text": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7977:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
}
````


Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/utilities.ts:7975:1",
    "name": "getLinesBetweenPositionAndNextNonWhitespaceCharacter",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


**watch.ts:249:60**, `getLineAndCharacterOfPosition`: no finding at original call; unit attempted.

No census finding was emitted at this original view.

Stream complete: true. Enclosing unit records:
````json
[
  {
    "where": "/workspace/scratch/baobab-road2-b/src/compiler/watch.ts:233:1",
    "name": "getFilesInErrorForSummary",
    "status": "attempted",
    "checker_diagnostics": []
  }
]
````


### Every src call and its argument type

The stock TypeScript 6.0.3 checker loaded all 604 source roots from the 13 src solution projects. A separate syntax scan covered 712 src source/declaration files: 604 were loaded, and the remaining library declarations have no references to either helper. There are 93 resolved calls to the two scanner functions. The 61 same-named method calls have different resolved signatures and are listed separately in excluded-or-unresolved-calls.json. No helper reference remains outside the checked program.

External dependencies came from a separate lockfile installation, with stock module resolution falling back to that installation. No node_modules directory was added to the frozen census tree. Whole-src checking uses strict source checks on the union of solution roots; composite, project-output, and isolated-declaration restrictions are disabled. Original whole-src diagnostics: 0. Full diagnostics and exact options are retained for each variant.

The latent census loads the compiler-only 79-root program. This stock call audit intentionally loads services too, including its SourceFile augmentation. Declared type means the variable/parameter/property declaration, declared return type for a call expression, or explicit/class/contextual this type. Flow type is recorded separately because a Node or optional SourceFile binding may be narrowed at the call. Exact declarations and variant types are in each calls.json.

| Call location | Callee | Argument | Original declared type | Original flow type | (b) call errors |
| --- | --- | --- | --- | --- | --- |
| src/compiler/checker.ts:52424:27 | `getLineAndCharacterOfPosition` | `file` | `SourceFile` | `SourceFile` | none |
| src/compiler/checker.ts:52425:25 | `getLineAndCharacterOfPosition` | `file` | `SourceFile` | `SourceFile` | none |
| src/compiler/emitter.ts:1444:52 | `getLineStarts` | `Debug.checkDefined(currentSourceFile)` | `ts.SourceFile` | `ts.SourceFile` | none |
| src/compiler/emitter.ts:3961:82 | `getLineAndCharacterOfPosition` | `currentSourceFile` | `ts.SourceFile \| undefined` | `ts.SourceFile` | none |
| src/compiler/emitter.ts:3961:150 | `getLineAndCharacterOfPosition` | `currentSourceFile` | `ts.SourceFile \| undefined` | `ts.SourceFile` | none |
| src/compiler/emitter.ts:6233:66 | `getLineAndCharacterOfPosition` | `sourceMapSource` | `ts.SourceMapSource` | `ts.SourceMapSource` | TS2345 |
| src/compiler/executeCommandLine.ts:109:27 | `getLineStarts` | `file` | `SourceFile` | `SourceFile` | none |
| src/compiler/program.ts:671:37 | `getLineAndCharacterOfPosition` | `diagnostic.file` | `SourceFile \| undefined` | `SourceFile` | none |
| src/compiler/program.ts:713:59 | `getLineAndCharacterOfPosition` | `file` | `SourceFile` | `SourceFile` | none |
| src/compiler/program.ts:714:57 | `getLineAndCharacterOfPosition` | `file` | `SourceFile` | `SourceFile` | none |
| src/compiler/program.ts:715:28 | `getLineAndCharacterOfPosition` | `file` | `SourceFile` | `SourceFile` | none |
| src/compiler/program.ts:768:59 | `getLineAndCharacterOfPosition` | `file` | `SourceFile` | `SourceFile` | none |
| src/compiler/program.ts:2973:28 | `getLineStarts` | `file` | `SourceFile \| undefined` | `SourceFile` | none |
| src/compiler/scanner.ts:471:43 | `getLineStarts` | `sourceFile` | `SourceFileLike` | `SourceFileLike` | TS2345 |
| src/compiler/scanner.ts:539:24 | `getLineStarts` | `sourceFile` | `SourceFileLike` | `SourceFileLike` | TS2345 |
| src/compiler/scanner.ts:549:46 | `getLineStarts` | `sourceFile` | `SourceFileLike` | `SourceFileLike` | none |
| src/compiler/tracing.ts:206:37 | `getLineAndCharacterOfPosition` | `file` | `SourceFile \| undefined` | `SourceFile` | none |
| src/compiler/tracing.ts:207:35 | `getLineAndCharacterOfPosition` | `file` | `SourceFile \| undefined` | `SourceFile` | none |
| src/compiler/transformers/declarations.ts:1264:40 | `getLineAndCharacterOfPosition` | `currentSourceFile` | `SourceFile` | `SourceFile` | none |
| src/compiler/transformers/declarations.ts:1264:109 | `getLineAndCharacterOfPosition` | `currentSourceFile` | `SourceFile` | `SourceFile` | none |
| src/compiler/transformers/jsx.ts:359:33 | `getLineAndCharacterOfPosition` | `originalFile` | `Node` | `SourceFile` | none |
| src/compiler/utilities.ts:1018:12 | `getLineStarts` | `sourceFile` | `SourceFileLike` | `SourceFileLike` | TS2345 |
| src/compiler/utilities.ts:1025:17 | `getLineAndCharacterOfPosition` | `file` | `SourceFile` | `SourceFile` | none |
| src/compiler/utilities.ts:1032:24 | `getLineStarts` | `sourceFile` | `SourceFileLike` | `SourceFileLike` | TS2345 |
| src/compiler/utilities.ts:1213:16 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/compiler/utilities.ts:2526:37 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/compiler/utilities.ts:2527:35 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/compiler/utilities.ts:6748:24 | `getLineStarts` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/compiler/watch.ts:249:30 | `getLineAndCharacterOfPosition` | `diagnosticForFileName.file!` | `SourceFile \| undefined` | `SourceFile` | none |
| src/harness/fourslashImpl.ts:726:41 | `getLineAndCharacterOfPosition` | `file` | `ts.SourceFile` | `ts.SourceFile` | none |
| src/server/session.ts:265:38 | `getLineAndCharacterOfPosition` | `info.file` | `SourceFile \| undefined` | `SourceFile` | none |
| src/server/session.ts:266:36 | `getLineAndCharacterOfPosition` | `info.file` | `SourceFile \| undefined` | `SourceFile` | none |
| src/server/session.ts:285:51 | `getLineAndCharacterOfPosition` | `diag.file` | `SourceFile \| undefined` | `SourceFile` | none |
| src/server/session.ts:286:49 | `getLineAndCharacterOfPosition` | `diag.file` | `SourceFile \| undefined` | `SourceFile` | none |
| src/server/session.ts:1557:57 | `getLineAndCharacterOfPosition` | `d.file` | `SourceFile \| undefined` | `SourceFile` | none |
| src/server/session.ts:1558:55 | `getLineAndCharacterOfPosition` | `d.file` | `SourceFile \| undefined` | `SourceFile` | none |
| src/services/codefixes/disableJsDiagnostics.ts:76:34 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/services/completions.ts:1801:67 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/services/completions.ts:1801:141 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/services/completions.ts:2171:9 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/services/completions.ts:2172:15 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/services/completions.ts:2462:24 | `getLineAndCharacterOfPosition` | `this` | `{ text: string; getLineAndCharacterOfPosition(pos: number): LineAndCharacter; }` | `{ text: string; getLineAndCharacterOfPosition(pos: number): LineAndCharacter; }` | TS2345 |
| src/services/completions.ts:2511:24 | `getLineAndCharacterOfPosition` | `this` | `{ text: string; getLineAndCharacterOfPosition(pos: number): LineAndCharacter; }` | `{ text: string; getLineAndCharacterOfPosition(pos: number): LineAndCharacter; }` | TS2345 |
| src/services/completions.ts:4393:33 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/services/completions.ts:4394:33 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/services/completions.ts:4415:33 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/services/completions.ts:4416:33 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/services/completions.ts:5344:34 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/services/completions.ts:5344:108 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/services/completions.ts:5738:21 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/services/completions.ts:5738:95 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/services/formatting/smartIndenter.ts:162:30 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/services/formatting/smartIndenter.ts:163:34 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/services/formatting/smartIndenter.ts:402:51 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFileLike` | `SourceFileLike` | TS2345 |
| src/services/formatting/smartIndenter.ts:441:38 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFileLike` | `SourceFileLike` | TS2345 |
| src/services/formatting/smartIndenter.ts:454:37 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFileLike` | `SourceFileLike` | TS2345 |
| src/services/formatting/smartIndenter.ts:471:40 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFileLike` | `SourceFileLike` | TS2345 |
| src/services/refactors/extractSymbol.ts:252:44 | `getLineAndCharacterOfPosition` | `context.file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/extractSymbol.ts:252:125 | `getLineAndCharacterOfPosition` | `context.file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/extractSymbol.ts:253:42 | `getLineAndCharacterOfPosition` | `context.file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/extractSymbol.ts:253:123 | `getLineAndCharacterOfPosition` | `context.file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/extractSymbol.ts:281:44 | `getLineAndCharacterOfPosition` | `context.file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/extractSymbol.ts:281:125 | `getLineAndCharacterOfPosition` | `context.file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/extractSymbol.ts:282:42 | `getLineAndCharacterOfPosition` | `context.file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/extractSymbol.ts:282:123 | `getLineAndCharacterOfPosition` | `context.file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/extractType.ts:117:40 | `getLineAndCharacterOfPosition` | `context.file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/extractType.ts:117:121 | `getLineAndCharacterOfPosition` | `context.file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/extractType.ts:118:38 | `getLineAndCharacterOfPosition` | `context.file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/extractType.ts:118:119 | `getLineAndCharacterOfPosition` | `context.file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/extractType.ts:330:47 | `getLineAndCharacterOfPosition` | `file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/extractType.ts:330:102 | `getLineAndCharacterOfPosition` | `file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/moveToFile.ts:193:32 | `getLineAndCharacterOfPosition` | `file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/moveToFile.ts:193:116 | `getLineAndCharacterOfPosition` | `file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/moveToFile.ts:194:30 | `getLineAndCharacterOfPosition` | `file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/moveToFile.ts:194:106 | `getLineAndCharacterOfPosition` | `file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/moveToNewFile.ts:60:32 | `getLineAndCharacterOfPosition` | `file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/moveToNewFile.ts:60:116 | `getLineAndCharacterOfPosition` | `file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/moveToNewFile.ts:61:30 | `getLineAndCharacterOfPosition` | `file` | `SourceFile` | `SourceFile` | none |
| src/services/refactors/moveToNewFile.ts:61:106 | `getLineAndCharacterOfPosition` | `file` | `SourceFile` | `SourceFile` | none |
| src/services/services.ts:1126:16 | `getLineAndCharacterOfPosition` | `this` | `SourceFileObject` | `this` | none |
| src/services/services.ts:1130:16 | `getLineStarts` | `this` | `SourceFileObject` | `this` | none |
| src/services/services.ts:1134:50 | `getLineStarts` | `this` | `SourceFileObject` | `this` | none |
| src/services/services.ts:1325:16 | `getLineAndCharacterOfPosition` | `this` | `SourceMapSourceObject` | `this` | none |
| src/services/sourcemaps.ts:85:40 | `getLineStarts` | `file` | `SourceFileLike \| undefined` | `SourceFileLike` | TS2345 |
| src/services/sourcemaps.ts:236:54 | `getLineStarts` | `this` | `SourceFileLike` | `SourceFileLike` | TS2345 |
| src/services/textChanges.ts:1343:24 | `getLineAndCharacterOfPosition` | `this` | `SourceFileLike` | `SourceFileLike` | TS2345 |
| src/services/utilities.ts:913:24 | `getLineStarts` | `sourceFile` | `SourceFileLike` | `SourceFileLike` | TS2345 |
| src/services/utilities.ts:3475:39 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/services/utilities.ts:3476:39 | `getLineAndCharacterOfPosition` | `sourceFile` | `SourceFile` | `SourceFile` | none |
| src/testRunner/unittests/tsserver/completionsIncomplete.ts:193:41 | `getLineAndCharacterOfPosition` | `file` | `ts.SourceFile` | `ts.SourceFile` | none |
| src/testRunner/unittests/tsserver/completionsIncomplete.ts:232:40 | `getLineAndCharacterOfPosition` | `file` | `ts.SourceFile` | `ts.SourceFile` | none |
| src/testRunner/unittests/tsserver/completionsIncomplete.ts:233:38 | `getLineAndCharacterOfPosition` | `file` | `ts.SourceFile` | `ts.SourceFile` | none |
| src/testRunner/unittests/tsserver/completionsIncomplete.ts:262:37 | `getLineAndCharacterOfPosition` | `file` | `ts.SourceFile` | `ts.SourceFile` | none |

Declared argument types changed by the edited helper parameter itself (all other entries retain the original declared type):

````json
[
  {
    "variant": "a",
    "file": "src/compiler/scanner.ts",
    "line": 549,
    "callee": "getLineStarts",
    "original": "SourceFileLike",
    "variant_declared": "{ readonly text: string; lineMap?: readonly number[]; }"
  },
  {
    "variant": "b",
    "file": "src/compiler/scanner.ts",
    "line": 549,
    "callee": "getLineStarts",
    "original": "SourceFileLike",
    "variant_declared": "{ readonly text: string; lineMap: readonly number[] | undefined; }"
  }
]
````


### Every failing call under (b)

These are actual stock TypeScript diagnostics from the whole-src program, not a classification based only on the spelling SourceFileLike. This preserves flow narrowing and class/contextual this types. The whole diagnostic stream, including any diagnostics outside calls, is retained in b/whole-src-diagnostics.json.

**src/compiler/emitter.ts:6233:66** `getLineAndCharacterOfPosition` argument `sourceMapSource`, declared `ts.SourceMapSource`, flow `ts.SourceMapSource`.

````json
{
  "file": "src/compiler/emitter.ts",
  "line": 6233,
  "column": 96,
  "start": 268874,
  "length": 15,
  "code": 2345,
  "category": "Error",
  "message": "Argument of type 'SourceMapSource' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceMapSource' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
  "related": []
}
````


**src/compiler/scanner.ts:471:43** `getLineStarts` argument `sourceFile`, declared `SourceFileLike`, flow `SourceFileLike`.

````json
{
  "file": "src/compiler/scanner.ts",
  "line": 471,
  "column": 57,
  "start": 47328,
  "length": 10,
  "code": 2345,
  "category": "Error",
  "message": "Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceFileLike' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
  "related": []
}
````


**src/compiler/scanner.ts:539:24** `getLineStarts` argument `sourceFile`, declared `SourceFileLike`, flow `SourceFileLike`.

````json
{
  "file": "src/compiler/scanner.ts",
  "line": 539,
  "column": 38,
  "start": 50647,
  "length": 10,
  "code": 2345,
  "category": "Error",
  "message": "Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceFileLike' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
  "related": []
}
````


**src/compiler/utilities.ts:1018:12** `getLineStarts` argument `sourceFile`, declared `SourceFileLike`, flow `SourceFileLike`.

````json
{
  "file": "src/compiler/utilities.ts",
  "line": 1018,
  "column": 26,
  "start": 33322,
  "length": 10,
  "code": 2345,
  "category": "Error",
  "message": "Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceFileLike' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
  "related": []
}
````


**src/compiler/utilities.ts:1032:24** `getLineStarts` argument `sourceFile`, declared `SourceFileLike`, flow `SourceFileLike`.

````json
{
  "file": "src/compiler/utilities.ts",
  "line": 1032,
  "column": 38,
  "start": 33844,
  "length": 10,
  "code": 2345,
  "category": "Error",
  "message": "Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceFileLike' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
  "related": []
}
````


**src/services/completions.ts:2462:24** `getLineAndCharacterOfPosition` argument `this`, declared `{ text: string; getLineAndCharacterOfPosition(pos: number): LineAndCharacter; }`, flow `{ text: string; getLineAndCharacterOfPosition(pos: number): LineAndCharacter; }`.

````json
{
  "file": "src/services/completions.ts",
  "line": 2462,
  "column": 54,
  "start": 97738,
  "length": 4,
  "code": 2345,
  "category": "Error",
  "message": "Argument of type '{ text: string; getLineAndCharacterOfPosition(pos: number): LineAndCharacter; }' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is missing in type '{ text: string; getLineAndCharacterOfPosition(pos: number): LineAndCharacter; }' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
  "related": [
    {
      "file": "src/compiler/scanner.ts",
      "line": 548,
      "column": 84,
      "start": 51088,
      "length": 7,
      "code": 2728,
      "category": "Message",
      "message": "'lineMap' is declared here.",
      "related": []
    }
  ]
}
````


**src/services/completions.ts:2511:24** `getLineAndCharacterOfPosition` argument `this`, declared `{ text: string; getLineAndCharacterOfPosition(pos: number): LineAndCharacter; }`, flow `{ text: string; getLineAndCharacterOfPosition(pos: number): LineAndCharacter; }`.

````json
{
  "file": "src/services/completions.ts",
  "line": 2511,
  "column": 54,
  "start": 99585,
  "length": 4,
  "code": 2345,
  "category": "Error",
  "message": "Argument of type '{ text: string; getLineAndCharacterOfPosition(pos: number): LineAndCharacter; }' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is missing in type '{ text: string; getLineAndCharacterOfPosition(pos: number): LineAndCharacter; }' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
  "related": [
    {
      "file": "src/compiler/scanner.ts",
      "line": 548,
      "column": 84,
      "start": 51088,
      "length": 7,
      "code": 2728,
      "category": "Message",
      "message": "'lineMap' is declared here.",
      "related": []
    }
  ]
}
````


**src/services/formatting/smartIndenter.ts:402:51** `getLineAndCharacterOfPosition` argument `sourceFile`, declared `SourceFileLike`, flow `SourceFileLike`.

````json
{
  "file": "src/services/formatting/smartIndenter.ts",
  "line": 402,
  "column": 81,
  "start": 19913,
  "length": 10,
  "code": 2345,
  "category": "Error",
  "message": "Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceFileLike' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
  "related": []
}
````


**src/services/formatting/smartIndenter.ts:441:38** `getLineAndCharacterOfPosition` argument `sourceFile`, declared `SourceFileLike`, flow `SourceFileLike`.

````json
{
  "file": "src/services/formatting/smartIndenter.ts",
  "line": 441,
  "column": 68,
  "start": 21974,
  "length": 10,
  "code": 2345,
  "category": "Error",
  "message": "Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceFileLike' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
  "related": []
}
````


**src/services/formatting/smartIndenter.ts:454:37** `getLineAndCharacterOfPosition` argument `sourceFile`, declared `SourceFileLike`, flow `SourceFileLike`.

````json
{
  "file": "src/services/formatting/smartIndenter.ts",
  "line": 454,
  "column": 67,
  "start": 22839,
  "length": 10,
  "code": 2345,
  "category": "Error",
  "message": "Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceFileLike' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
  "related": []
}
````


**src/services/formatting/smartIndenter.ts:471:40** `getLineAndCharacterOfPosition` argument `sourceFile`, declared `SourceFileLike`, flow `SourceFileLike`.

````json
{
  "file": "src/services/formatting/smartIndenter.ts",
  "line": 471,
  "column": 70,
  "start": 23802,
  "length": 10,
  "code": 2345,
  "category": "Error",
  "message": "Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceFileLike' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
  "related": []
}
````


**src/services/sourcemaps.ts:85:40** `getLineStarts` argument `file`, declared `SourceFileLike \| undefined`, flow `SourceFileLike`.

````json
{
  "file": "src/services/sourcemaps.ts",
  "line": 85,
  "column": 54,
  "start": 3217,
  "length": 4,
  "code": 2345,
  "category": "Error",
  "message": "Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceFileLike' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
  "related": []
}
````


**src/services/sourcemaps.ts:236:54** `getLineStarts` argument `this`, declared `SourceFileLike`, flow `SourceFileLike`.

````json
{
  "file": "src/services/sourcemaps.ts",
  "line": 236,
  "column": 68,
  "start": 9564,
  "length": 4,
  "code": 2345,
  "category": "Error",
  "message": "Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceFileLike' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
  "related": []
}
````


**src/services/textChanges.ts:1343:24** `getLineAndCharacterOfPosition` argument `this`, declared `SourceFileLike`, flow `SourceFileLike`.

````json
{
  "file": "src/services/textChanges.ts",
  "line": 1343,
  "column": 54,
  "start": 66040,
  "length": 4,
  "code": 2345,
  "category": "Error",
  "message": "Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceFileLike' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
  "related": []
}
````


**src/services/utilities.ts:913:24** `getLineStarts` argument `sourceFile`, declared `SourceFileLike`, flow `SourceFileLike`.

````json
{
  "file": "src/services/utilities.ts",
  "line": 913,
  "column": 38,
  "start": 31468,
  "length": 10,
  "code": 2345,
  "category": "Error",
  "message": "Argument of type 'SourceFileLike' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is optional in type 'SourceFileLike' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.",
  "related": []
}
````


15 calls have diagnostics under (b). The corresponding whole-source and services summaries are:

````json
{
  "label": "a",
  "whole_src_roots": 604,
  "whole_src_diagnostics": 0,
  "calls": 93,
  "services_files": 253,
  "services_diagnostics": 0,
  "services_TS2386": 0
}
````


````json
{
  "label": "b",
  "whole_src_roots": 604,
  "whole_src_diagnostics": 15,
  "calls": 93,
  "services_files": 253,
  "services_diagnostics": 15,
  "services_TS2386": 0
}
````


### Emitted public declarations with stripInternal

All three public bundles were built from actual stock TypeScript declaration emit with stripInternal=true and emitDeclarationOnly=true, then the upstream declaration bundler was run on the emitted .d.ts entry. No JavaScript or backend output was produced. Emission is allowed despite semantic errors so that (b) can be compared; its source diagnostics above are not suppressed. The exact options, emit diagnostics, emitted-file hashes, public bundles, and textual diffs are retained.

| Variant | Free getLineStarts in public bundle | Only getLineAndCharacterOfPosition parameter differs | Public SourceFile has lineMap | Public SourceFile caller errors |
| --- | ---: | --- | --- | ---: |
| original | 0 | true | false | 0 |
| a | 0 | true | false | 0 |
| b | 0 | true | false | 1 |

getLineStarts is @internal, so changing its parameter contributes no free-function declaration to public typescript.d.ts. The public API comparison normalizes only getLineAndCharacterOfPosition's first parameter back to SourceFileLike, then prints and compares the complete declaration AST. Equality confirms that this parameter is the sole public declaration difference, independent of formatting. The direct diffs and emitted signatures are retained.

Public signature original:

````json
[
  {
    "name": "getLineAndCharacterOfPosition",
    "text": "function getLineAndCharacterOfPosition(sourceFile: SourceFileLike, position: number): LineAndCharacter;"
  }
]
````


Public signature a:

````json
[
  {
    "name": "getLineAndCharacterOfPosition",
    "text": "function getLineAndCharacterOfPosition(sourceFile: {\n        readonly text: string;\n        lineMap?: readonly number[];\n    }, position: number): LineAndCharacter;"
  }
]
````


Public signature b:

````json
[
  {
    "name": "getLineAndCharacterOfPosition",
    "text": "function getLineAndCharacterOfPosition(sourceFile: {\n        readonly text: string;\n        lineMap: readonly number[] | undefined;\n    }, position: number): LineAndCharacter;"
  }
]
````


The emitted-public-API caller probe for b reports:

````json
[
  {
    "code": 2345,
    "file": "public-sourcefile-consumer.ts",
    "line": 3,
    "column": 34,
    "message": "Argument of type 'SourceFile' is not assignable to parameter of type '{ readonly text: string; lineMap: readonly number[] | undefined; }'.\n  Property 'lineMap' is missing in type 'SourceFile' but required in type '{ readonly text: string; lineMap: readonly number[] | undefined; }'."
  }
]
````


The public caller probe uses the emitted ts.SourceFile type, not the internal source interface. Its lineMap property can disappear under stripInternal even though the internal SourceFile member is required. This distinction is separate from the internal src caller audit.

Follow-up evidence is under continuation/. Original first-pass raw files remain under raw/; longer main retries are under continuation/baseline/raw/, and variant streams are under continuation/a/raw/ and continuation/b/raw/. All started measurement children are waited for, with process-group termination at their bound. The production compiler and main adaptations are unchanged; the only TypeScript source edits are in the two scratch variants.

## Third measurement: named LineMapSource, variant (c)

This measurement starts again from the same frozen adapted input, main d690f276749f91e6b36fe3a9726a5a453fcd7d39, TypeScript 6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8, and the same guarded meter binary. It does not apply the (a) or (b) scratch edits. The second report commit is 7944b20a23c7e2e40c00e925603c20bd7267a569. Nothing in compiler, services, cohere, or the adaptation rules lands.

**Overall outcome: no tested exact (c) form meets every requested gate. The internal marker breaks declaration bundling; removing it introduces a new public SDK optional-widening refusal. The internal compiler view counts below must not be read as end-to-end acceptance.**

### Exact source changes

Immediately before SourceFileLike in compiler/types.ts, (c) inserts:

```ts
/** @internal */
export interface LineMapSource {
    readonly text: string;
    lineMap?: readonly number[] | undefined;
}
```

SourceFileLike becomes `extends LineMapSource`. Its own text and lineMap member types are byte-for-byte identical to these declarations in the frozen adapted input, so both are removed. Its optional getPositionOfLineAndCharacter member is retained unchanged. The removed lineMap member had its own @internal comment. That comment is removed with the member; LineMapSource.lineMap has exactly the requested spelling, without a member-level internal marker. This matters when LineMapSource is made public.

Only getLineStarts and getLineAndCharacterOfPosition change their first parameter to LineMapSource. A type-only import of LineMapSource is necessary in scanner.ts. It is placed on the existing import line to retain all scanner line coordinates. The two function bodies, all other parameters, and all call expressions are unchanged. Full src hashes verify that scanner.ts and types.ts are the only changed source files. The patches and hashes are retained.

The public alternative removes only LineMapSource's @internal tag. A blank comment retains all line coordinates. Its runtime and relation-bearing syntax is otherwise identical, but it is independently rechecked and independently censused over the same ten files.

### Measured result for all 34 original views

| Variant | Complete single-file streams | Cleared at original view | Unchanged finding | Replacement finding | Checker diagnostic | Incomplete or excluded | Core checker diagnostics | Core first stop |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| (c), internal LineMapSource | 10/10 | 22 | 12 | 0 | 0 | 0 | 0 | `Refused tracing.ts:232:34`: a cast the runtime can't check |
| (c), public LineMapSource | 10/10 | 22 | 12 | 0 | 0 | 0 | 0 | `Refused tracing.ts:232:34`: a cast the runtime can't check |

A cleared view means that its enclosing unit was attempted and the guarded full census reports no finding or checker diagnostic at that original argument/call span. It does not mean that the whole function or root lowers. The 22 helper views consist of 21 SourceFile values and one SourceMapSource value. The other 12 views still call unmodified SourceFileLike callees. The original checker.ts census remains cut off after 25 minutes, so this is a complete comparison of the 34 observed views, not a new claim that the original 79-file site count is exhaustive.

| Callee | Original sites | (c) cleared | (c) unchanged | Public (c) cleared | Public (c) unchanged |
| --- | ---: | ---: | ---: | ---: | ---: |
| `getEndLinePosition` | 1 | 0 | 1 | 0 | 1 |
| `getLineAndCharacterOfPosition` | 18 | 18 | 0 | 18 | 0 |
| `getLineStarts` | 4 | 4 | 0 | 4 | 0 |
| `getLinesBetweenPositions` | 5 | 0 | 5 | 0 | 5 |
| `getPositionOfLineAndCharacter` | 2 | 0 | 2 | 0 | 2 |
| `transferSourceFileChildren` | 2 | 0 | 2 | 0 | 2 |
| `unsetNodeChildren` | 2 | 0 | 2 | 0 | 2 |

### Adamic relations, independent of stock tsc

The pinned cohere binary runs the full cohere:adamic rule set under strict, noUncheckedIndexedAccess, exactOptionalPropertyTypes, verbatimModuleSyntax, ES2024, ESNext/Bundler, and the remaining production loader options recorded in cohere-tsconfig.json and load-options.go.txt. The CLI actually walks 77 compiler source files out of its 79-file program; its complete scope, rule counts, unresolved-input warnings, and findings are retained. The first CLI attempt refused to lint import-only files; its zero-rule failure is preserved separately and is not used as a measurement. The successful rerun explicitly claims all 79 compiler files.

This runs the actual no-optional-widening and invariant-mutable implementations, not a stock tsc approximation. Required SourceFile.lineMap is readonly number[] | undefined in the adapted tree. The target optional slot explicitly holds undefined too. No rule reports this SourceFile-to-LineMapSource relation at any of the 22 helper views. Existing non-null assertion findings on some enclosing calls are quoted below; they are not new relation findings.

| Variant | Cohere rules | Compiler files checked | New cohere findings | Removed cohere findings | New measured lowering findings | Removed measured lowering findings |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| (c), internal LineMapSource | 40 | 77 | 0 | 24 | 0 | 22 |
| (c), public LineMapSource | 40 | 77 | 0 | 24 | 0 | 22 |

The cohere comparison includes every finding, not just the helper calls. Its removed findings include the two checker.ts calls that were not in the 34-view lowering list. The lowering comparison includes every finding in the ten completed selected-file streams, including boundaries and unrelated reasons. Path and unchanged types.ts line shifts are normalized; no reason is filtered out. Exact additions and removals are in c-analysis.json. A newly visible later boundary, if present in that comparison, is not silently counted as successful lowering.

Cohere additions and removals for (c):

````json
{
  "baseline_unique": 9728,
  "variant_unique": 9704,
  "added": [],
  "removed": [
    {
      "rule": "adamic/no-optional-widening",
      "path": "<tree>/src/compiler/emitter.ts",
      "line": 3961,
      "column": 112,
      "severity": "error",
      "message": "`currentSourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    },
    {
      "rule": "adamic/no-optional-widening",
      "path": "<tree>/src/compiler/transformers/declarations.ts",
      "line": 1264,
      "column": 139,
      "severity": "error",
      "message": "`currentSourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    },
    {
      "rule": "adamic/no-optional-widening",
      "path": "<tree>/src/compiler/emitter.ts",
      "line": 3961,
      "column": 180,
      "severity": "error",
      "message": "`currentSourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    },
    {
      "rule": "adamic/no-optional-widening",
      "path": "<tree>/src/compiler/utilities.ts",
      "line": 6748,
      "column": 38,
      "severity": "error",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    },
    {
      "rule": "adamic/no-optional-widening",
      "path": "<tree>/src/compiler/executeCommandLine.ts",
      "line": 109,
      "column": 41,
      "severity": "error",
      "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    },
    {
      "rule": "adamic/no-optional-widening",
      "path": "<tree>/src/compiler/program.ts",
      "line": 2973,
      "column": 42,
      "severity": "error",
      "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    },
    {
      "rule": "adamic/no-optional-widening",
      "path": "<tree>/src/compiler/utilities.ts",
      "line": 1213,
      "column": 46,
      "severity": "error",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    },
    {
      "rule": "adamic/no-optional-widening",
      "path": "<tree>/src/compiler/utilities.ts",
      "line": 1025,
      "column": 47,
      "severity": "error",
      "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    },
    {
      "rule": "adamic/no-optional-widening",
      "path": "<tree>/src/compiler/checker.ts",
      "line": 52425,
      "column": 55,
      "severity": "error",
      "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    },
    {
      "rule": "adamic/no-optional-widening",
      "path": "<tree>/src/compiler/checker.ts",
      "line": 52424,
      "column": 57,
      "severity": "error",
      "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    },
    {
      "rule": "adamic/no-optional-widening",
      "path": "<tree>/src/compiler/program.ts",
      "line": 715,
      "column": 58,
      "severity": "error",
      "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    },
    {
      "rule": "adamic/no-optional-widening",
      "path": "<tree>/src/compiler/watch.ts",
      "line": 249,
      "column": 60,
      "severity": "error",
      "message": "`diagnosticForFileName.file!` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    },
    {
      "rule": "adamic/no-optional-widening",
      "path": "<tree>/src/compiler/transformers/jsx.ts",
      "line": 359,
      "column": 63,
      "severity": "error",
      "message": "`originalFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    },
    {
      "rule": "adamic/no-optional-widening",
      "path": "<tree>/src/compiler/tracing.ts",
      "line": 207,
      "column": 65,
      "severity": "error",
      "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    },
    {
      "rule": "adamic/no-optional-widening",
      "path": "<tree>/src/compiler/utilities.ts",
      "line": 2527,
      "column": 65,
      "severity": "error",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    },
    {
      "rule": "adamic/no-optional-widening",
      "path": "<tree>/src/compiler/emitter.ts",
      "line": 1444,
      "column": 66,
      "severity": "error",
      "message": "`Debug.checkDefined(currentSourceFile)` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    },
    {
      "rule": "adamic/no-optional-widening",
      "path": "<tree>/src/compiler/tracing.ts",
      "line": 206,
      "column": 67,
      "severity": "error",
      "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    },
    {
      "rule": "adamic/no-optional-widening",
      "path": "<tree>/src/compiler/utilities.ts",
      "line": 2526,
      "column": 67,
      "severity": "error",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    },
    {
      "rule": "adamic/no-optional-widening",
      "path": "<tree>/src/compiler/program.ts",
      "line": 671,
      "column": 67,
      "severity": "error",
      "message": "`diagnostic.file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    },
    {
      "rule": "adamic/no-optional-widening",
      "path": "<tree>/src/compiler/transformers/declarations.ts",
      "line": 1264,
      "column": 70,
      "severity": "error",
      "message": "`currentSourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    },
    {
      "rule": "adamic/no-optional-widening",
      "path": "<tree>/src/compiler/program.ts",
      "line": 714,
      "column": 87,
      "severity": "error",
      "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    },
    {
      "rule": "adamic/no-optional-widening",
      "path": "<tree>/src/compiler/program.ts",
      "line": 713,
      "column": 89,
      "severity": "error",
      "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    },
    {
      "rule": "adamic/no-optional-widening",
      "path": "<tree>/src/compiler/program.ts",
      "line": 768,
      "column": 89,
      "severity": "error",
      "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    },
    {
      "rule": "adamic/no-optional-widening",
      "path": "<tree>/src/compiler/emitter.ts",
      "line": 6233,
      "column": 96,
      "severity": "error",
      "message": "`sourceMapSource` is a 'SourceMapSource', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceMapSource' does not declare. A 'SourceMapSource' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    }
  ]
}
````


### getLineStarts own unit and the cache write

The exact unchanged cache expression is:

```ts
return sourceFile.lineMap || (sourceFile.lineMap = computeLineStarts(sourceFile.text));
```

getLineStarts is scanner.ts:503:1; the assignment is scanner.ts:504:35. Every finding and the unit status from its own full-mode stream follows.

**(c), internal LineMapSource**

````json
{
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/scanner.ts:503:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 49075,
      "body_end": 49173,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLineStarts",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/scanner.ts:503:1"
    }
  ],
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-c/src/compiler/scanner.ts:503:1",
      "phase": "lowering",
      "kind": "NotYet",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/scanner.ts:504:35",
      "reason": "an assignment value to a member",
      "text": "/workspace/scratch/baobab-road2-c/src/compiler/scanner.ts:504:35: stage 0 can't lower an assignment value to a member yet"
    },
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-c/src/compiler/scanner.ts:503:1",
      "phase": "lowering",
      "kind": "Boundary",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/scanner.ts:504:5",
      "reason": "failed statement; state rolled back; continued at next statement or declaration",
      "text": "/workspace/scratch/baobab-road2-c/src/compiler/scanner.ts:504:35: stage 0 can't lower an assignment value to a member yet",
      "start": 49077,
      "end": 49170
    }
  ]
}
````


**(c), public LineMapSource**

````json
{
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/scanner.ts:503:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 49075,
      "body_end": 49173,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLineStarts",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/scanner.ts:503:1"
    }
  ],
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-c-public/src/compiler/scanner.ts:503:1",
      "phase": "lowering",
      "kind": "NotYet",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/scanner.ts:504:35",
      "reason": "an assignment value to a member",
      "text": "/workspace/scratch/baobab-road2-c-public/src/compiler/scanner.ts:504:35: stage 0 can't lower an assignment value to a member yet"
    },
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-c-public/src/compiler/scanner.ts:503:1",
      "phase": "lowering",
      "kind": "Boundary",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/scanner.ts:504:5",
      "reason": "failed statement; state rolled back; continued at next statement or declaration",
      "text": "/workspace/scratch/baobab-road2-c-public/src/compiler/scanner.ts:504:35: stage 0 can't lower an assignment value to a member yet",
      "start": 49077,
      "end": 49170
    }
  ]
}
````


The corresponding baseline findings are:

````json
[
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/baobab-adapted/src/compiler/scanner.ts:503:1",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/workspace/scratch/baobab-adapted/src/compiler/scanner.ts:504:35",
    "reason": "an assignment value to a member",
    "text": "/workspace/scratch/baobab-adapted/src/compiler/scanner.ts:504:35: stage 0 can't lower an assignment value to a member yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/workspace/scratch/baobab-adapted/src/compiler/scanner.ts:503:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/workspace/scratch/baobab-adapted/src/compiler/scanner.ts:504:5",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/workspace/scratch/baobab-adapted/src/compiler/scanner.ts:504:35: stage 0 can't lower an assignment value to a member yet",
    "start": 49058,
    "end": 49151
  }
]
````


Thus the measured writer stop is the pre-existing NotYet reason `an assignment value to a member`, not the #x0nd1mh optional-checked-view write reason. The normal core.ts first stop after each form of (c) is the unrelated cast in tracing.ts:232:34, not this write. The conditional request to count all distinct callers if the write becomes that first stop does not apply.

The implementation of checkViewOptionalWrites is preserved in view-optional-writes.go.txt. It refuses a write to an optional field when that receiver/field is checked; it does not refuse every optional field assignment merely because the declaration uses a question mark. The existing assignment-expression stop precedes that later check here, so it cannot be used as proof that the cache write is supported through every checked view. No actual (c) stream or cohere relation finding reports the required-to-optional writable-slot refusal posited in the conditional request. The requested condition for (d) therefore does not fire; (d) is not substituted for the measured (c) source.

Two separately labeled minimal controls are retained in c/probes/. The original expression reproduces the assignment-value NotYet. The control that separates the write into a statement is accepted for its required-to-optional record pair. Neither control changes (c), replaces its real-file findings, or establishes support for other shapes.

Core first-stop records (the driver also repeats its one-root input as a whole-program record; the first record is quoted):

**(c), internal LineMapSource**

````json
{
  "roots": [
    "/workspace/scratch/baobab-road2-c/src/compiler/core.ts"
  ],
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-c/src/compiler/tracing.ts:232:34",
  "reason": "a cast the runtime can't check",
  "fix": "use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)",
  "seconds": 21.274722228999998
}
````


**(c), public LineMapSource**

````json
{
  "roots": [
    "/workspace/scratch/baobab-road2-c-public/src/compiler/core.ts"
  ],
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/tracing.ts:232:34",
  "reason": "a cast the runtime can't check",
  "fix": "use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)",
  "seconds": 44.793428969
}
````


### Whole src, production checker, and services

All whole-src checks use the exact same union of 604 roots from all 13 src solution projects. The stock checks keep the separate lockfile dependency installation and the project source-check settings used for (a) and (b). They do not write node_modules into any frozen or census input tree. The extra exact-optional stock run changes only exactOptionalPropertyTypes. This is distinct from production load.Load, which uses every Adamic compiler option and its own environment.

For the actual production checker, adamic-check.go.txt calls unmodified load.Load over those 604 root paths and never calls lowering or an emitter. Every checker-rejected result is required to have no usable Program. This is the native Adamic checker, not stock tsc configured to approximate it. Its own full-src baseline is already rejected: many diagnostics concern type-only imports under verbatimModuleSyntax and test globals unavailable under types: []. Those errors are retained and compared, not attributed to (c).

| Tree | Stock whole-src errors | Stock with exactOptionalPropertyTypes | Production Adamic whole-src errors | New production checker errors | Services errors, stock | Services errors, exact optional | Services TS2386, stock / exact |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| Frozen baseline | 0 | 416 | 6851 | 0 | 0 | 142 | 0 / 0 |
| (c), internal LineMapSource | 0 | 416 | 6851 | 0 | 0 | 142 | 0 / 0 |
| (c), public LineMapSource | 0 | 416 | 6851 | 0 | 0 | 142 | 0 / 0 |

The production diagnostic sets are compared by their complete text after scratch-path and unchanged type-declaration line mapping. The stock exact-optional diagnostic sets are compared by file, location, code, category, and message. The same baseline errors remain, with additions and removals recorded in c-analysis.json. Core.ts is separately checked by the ordinary first-stop driver with actual Adamic options; it has no checker diagnostics under either form of (c).

Every one of the 93 resolved scanner-helper calls remains in the stock whole-src audit. SourceFileLike arguments and text-only object arguments are accepted by the optional LineMapSource contract; (b)'s 15 call errors do not recur. The only declared argument type changed by this edit is scanner.ts:549's sourceFile parameter, now LineMapSource instead of SourceFileLike. The declarations, flow types, and full diagnostic streams for all calls remain in each calls.json. The two diagnostics whose spans happen to enclose calls under exact optional checking are the same pre-existing enclosing object-literal TS2375 diagnostics in server/session.ts, not helper-parameter incompatibilities.

### Public declarations with stripInternal on and off

Actual stock declaration emission and the unchanged upstream declaration bundler are used, as before. All emitted-file hashes and options are retained. The bundler independently separates its public typescript.d.ts from typescript.internal.d.ts using @internal tags, even when the TypeScript compiler's stripInternal option is off. This distinction changes what the off result means.

| LineMapSource | Compiler stripInternal | Declaration emit | Upstream bundle | Public LineMapSource | Public getLineStarts | Public closure/consumer check | Lane API guard |
| --- | --- | --- | --- | --- | --- | --- | --- |
| @internal | on | 278 files, no emit diagnostics | Fails, child exit 1 | Stripped | Stripped | Dangling emitted references | Build fails before a bundle exists |
| @internal | off | 278 files, no emit diagnostics | Succeeds | absent | absent | 2 diagnostic(s) | 4 unsanctioned API changes |
| public | on | 278 files, no emit diagnostics | Succeeds | present | absent | 0 diagnostic(s) | 7 unsanctioned API changes |
| public | off | 278 files, no emit diagnostics | Succeeds | present | absent | 0 diagnostic(s) | 7 unsanctioned API changes |

**Internal (c), stripInternal on.** The raw emitted scanner.d.ts still imports LineMapSource and declares `getLineAndCharacterOfPosition(sourceFile: LineMapSource, position: number): LineAndCharacter`. Emitted types.d.ts declares `SourceFileLike extends LineMapSource` but no LineMapSource. The upstream bundler then aborts with `AssertionError [ERR_ASSERTION]: No symbol for node at /workspace/scratch/baobab-emit-c/compiler/types.d.ts:2000:40`. No public typescript.d.ts is produced in this configuration. It would be inaccurate to present a fabricated successful bundle or diff. The raw emitted selected declarations and all diagnostics are:

````json
{
  "stripInternal": true,
  "diagnostics": [
    {
      "code": 2305,
      "file": "compiler/scanner.d.ts",
      "line": 1,
      "column": 181,
      "message": "Module '\"./_namespaces/ts.js\"' has no exported member 'LineMapSource'."
    },
    {
      "code": 2304,
      "file": "compiler/types.d.ts",
      "line": 2000,
      "column": 41,
      "message": "Cannot find name 'LineMapSource'."
    },
    {
      "code": 2304,
      "file": "compiler/utilities.d.ts",
      "line": 23,
      "column": 29,
      "message": "Cannot find name 'Mutable'."
    },
    {
      "code": 2304,
      "file": "server/project.d.ts",
      "line": 114,
      "column": 118,
      "message": "Cannot find name 'EmitResult'."
    }
  ],
  "selected_declarations": [
    {
      "file": "compiler/types.d.ts",
      "name": "SourceFileLike",
      "text": "export interface SourceFileLike extends LineMapSource {\n}"
    },
    {
      "file": "compiler/scanner.d.ts",
      "name": "getLineAndCharacterOfPosition",
      "text": "export declare function getLineAndCharacterOfPosition(sourceFile: LineMapSource, position: number): LineAndCharacter;"
    }
  ],
  "bundled_typescript_dts_produced": false,
  "bundler_child_exit": 1
}
````


The raw stripped baseline has pre-existing dangling Mutable and EmitResult references too; its analysis is retained as original/raw-declaration-analysis.json. The additional c references are the missing scanner import TS2305 and missing LineMapSource base TS2304. The baseline upstream public bundle succeeds and has zero closure/consumer diagnostics.

**Internal (c), stripInternal off.** The raw module declarations retain LineMapSource, so bundling succeeds. The bundler still excludes its @internal interface from the public bundle. The public d.ts literally contains these declarations:

````json
[
  {
    "name": "SourceFileLike",
    "kind": "InterfaceDeclaration",
    "text": "interface SourceFileLike extends LineMapSource {\n    }"
  },
  {
    "name": "SourceFileLike",
    "kind": "InterfaceDeclaration",
    "text": "interface SourceFileLike {\n        getLineAndCharacterOfPosition(pos: number): LineAndCharacter;\n    }"
  },
  {
    "name": "getLineAndCharacterOfPosition",
    "kind": "FunctionDeclaration",
    "text": "function getLineAndCharacterOfPosition(sourceFile: LineMapSource, position: number): LineAndCharacter;"
  }
]
````


The public declaration checker reports:

````json
[
  {
    "code": 2304,
    "file": "typescript.d.ts",
    "line": 5906,
    "column": 38,
    "message": "Cannot find name 'LineMapSource'."
  },
  {
    "code": 2304,
    "file": "typescript.d.ts",
    "line": 8553,
    "column": 56,
    "message": "Cannot find name 'LineMapSource'."
  }
]
````


Its unchanged stage3/lane check.api_check rejects the API snapshot as follows:

````json
[
  "unsanctioned API declaration: FunctionDeclaration:[\"ts\",\"getLineAndCharacterOfPosition\"]#0",
  "unsanctioned API declaration: InterfaceDeclaration:[\"ts\",\"SourceFileLike\"]#0",
  "unsanctioned API declaration: comments#0",
  "unsanctioned API declaration: member:[\"ts\",\"SourceFileLike\",\"text\"]#0"
]
````


**Public (c), both settings.** Removing the interface-level marker produces a public LineMapSource and a valid public consumer. Its inherited lineMap member is now public too, because the requested interface did not put @internal on that member. The public declarations are:

````json
[
  {
    "name": "LineMapSource",
    "kind": "InterfaceDeclaration",
    "text": "interface LineMapSource {\n        readonly text: string;\n        lineMap?: readonly number[] | undefined;\n    }"
  },
  {
    "name": "SourceFileLike",
    "kind": "InterfaceDeclaration",
    "text": "interface SourceFileLike extends LineMapSource {\n    }"
  },
  {
    "name": "SourceFileLike",
    "kind": "InterfaceDeclaration",
    "text": "interface SourceFileLike {\n        getLineAndCharacterOfPosition(pos: number): LineAndCharacter;\n    }"
  },
  {
    "name": "getLineAndCharacterOfPosition",
    "kind": "FunctionDeclaration",
    "text": "function getLineAndCharacterOfPosition(sourceFile: LineMapSource, position: number): LineAndCharacter;"
  }
]
````


The complete textual diffs against the matching baseline settings are retained in c-public/typescript-public.diff and c-public-unstripped/typescript-public.diff. getLineStarts remains internal and absent from public typescript.d.ts in both cases. With stripInternal off it is retained in the raw scanner declarations and the internal bundle; it does not enter the public bundle. The lane guard still rejects seven unsanctioned public API changes:

````json
[
  "unsanctioned API declaration: FunctionDeclaration:[\"ts\",\"getLineAndCharacterOfPosition\"]#0",
  "unsanctioned API declaration: InterfaceDeclaration:[\"ts\",\"LineMapSource\"]#0",
  "unsanctioned API declaration: InterfaceDeclaration:[\"ts\",\"SourceFileLike\"]#0",
  "unsanctioned API declaration: comments#0",
  "unsanctioned API declaration: member:[\"ts\",\"LineMapSource\",\"lineMap\"]#0",
  "unsanctioned API declaration: member:[\"ts\",\"LineMapSource\",\"text\"]#0",
  "unsanctioned API declaration: member:[\"ts\",\"SourceFileLike\",\"text\"]#0"
]
````


The frozen baseline passes the actual lane API guard with zero errors. No API sanction is added for this report-only experiment. stage3/lane uses its normalized declaration/snapshot guard, not an installed Microsoft API Extractor invocation; the independent no-skipLibCheck public consumer check above supplies the missing-type closure diagnostics. Thus making LineMapSource public repairs the dangling type but does not make the unchanged landing lane accept a new public API.

The successful public-(c) typescript.d.ts bundles are byte-for-byte identical with compiler stripInternal on and off, as are the matching baseline public bundles. The raw declarations and internal bundles differ: getLineStarts is retained only when compiler stripInternal is off. c-declaration-matrix.json records the exact selected declarations for every mode.

### New public SDK refusal: the requested no-new-refusal outcome is not met

**Neither exact form of (c) meets all requested gates.** Keeping LineMapSource internal breaks declaration bundling. Making it public repairs the missing type but introduces a new Adamic refusal for an ordinary public SourceFile caller. This is outside the 34 internal compiler views, so it must not be hidden by their clearance count.

The emitted-public-API probe is:

```ts
import * as ts from "./typescript.js";
export function probe(file: ts.SourceFile): ts.LineAndCharacter {
    return ts.getLineAndCharacterOfPosition(file, 0);
}
```

| Emitted SDK | Production checker errors | Cohere findings | Guarded optional-widening refusals |
| --- | ---: | ---: | ---: |
| original | 0 | 0 | 0 |
| c-public | 0 | 1 | 1 |
| c-public-unstripped | 0 | 1 | 1 |
| c-public-readonly-sdk-control | 0 | 1 | 1 |

Both cohere and the same guarded full-mode lowering binary report the new refusal, including when only the exported slot is made readonly:

```text
optional property lineMap in LineMapSource absent from structural source SourceFile, which can hide fields
```

Public SourceFile has no lineMap because its original member remains @internal. Newly public LineMapSource declares an optional lineMap, so this public view reintroduces the structural hiding rule. Its readonly variant is still refused because optional widening is independent of writable-slot invariance. The production checker accepts all four SDK probes; this is an Adamic relation refusal, not a TypeScript checker error.

The readonly SDK control changes only the emitted interface member for this probe. It is not a full implementation or measurement of (d), and no narrow writer is claimed. The conditional compiler-side writable-slot refusal that requested (d) was not observed. Even a separate narrow writer would not change this SDK argument relation while its first parameter retains that public optional member. A further public API/cache-annotation change would be required; it is not silently folded into the requested experiment.

Every SDK finding, including the baseline external function-without-a-body NotYet, is retained below. The SDK census is isolated to the consumer and emitted declaration bundle, uses LATENT_FULL=1 and LATENT_ASSERT_NO_OUTPUT=1, and has no backend output. It is additional evidence, not a substitute for the real ten-file runs.

**original**

````json
{
  "production_checker": {
    "accepted": true,
    "diagnostics": [],
    "driver": "production load.Load, no lowering or output IR",
    "roots": 1,
    "seconds": 0.251802112
  },
  "cohere_findings": [],
  "guarded_findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-sdk-original/consumer.ts:2:1",
      "phase": "lowering",
      "kind": "NotYet",
      "where": "/workspace/scratch/baobab-sdk-original/typescript.d.ts:8557:5",
      "reason": "a function without a body",
      "text": "/workspace/scratch/baobab-sdk-original/typescript.d.ts:8557:5: stage 0 can't lower a function without a body yet"
    },
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-sdk-original/consumer.ts:2:1",
      "phase": "lowering",
      "kind": "Boundary",
      "where": "/workspace/scratch/baobab-sdk-original/consumer.ts:3:5",
      "reason": "failed statement; state rolled back; continued at next statement or declaration",
      "text": "/workspace/scratch/baobab-sdk-original/typescript.d.ts:8557:5: stage 0 can't lower a function without a body yet",
      "start": 104,
      "end": 158
    }
  ]
}
````


**c-public**

````json
{
  "production_checker": {
    "accepted": true,
    "diagnostics": [],
    "driver": "production load.Load, no lowering or output IR",
    "roots": 1,
    "seconds": 0.171637924
  },
  "cohere_findings": [
    {
      "kind": "finding",
      "path": "/workspace/adamic/stage3/census/sourcefile-like-sites/continuation/c-public/public-adamic-consumer.ts",
      "line": 3,
      "column": 45,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`file` is a 'SourceFile', seen here as 'LineMapSource', which adds the optional property `lineMap` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `lineMap` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a 'readonly number[] | undefined' and holds something else. Declare `lineMap` in the source's type, or build the value with the property named."
    }
  ],
  "guarded_findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-sdk-c-public/consumer.ts:2:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-sdk-c-public/consumer.ts:3:45",
      "reason": "optional property lineMap in LineMapSource absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-sdk-c-public/consumer.ts:3:45: Adamic 0.1 refuses optional property lineMap in LineMapSource absent from structural source SourceFile, which can hide fields; declare lineMap on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    },
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-sdk-c-public/consumer.ts:2:1",
      "phase": "lowering",
      "kind": "NotYet",
      "where": "/workspace/scratch/baobab-sdk-c-public/typescript.d.ts:8561:5",
      "reason": "a function without a body",
      "text": "/workspace/scratch/baobab-sdk-c-public/typescript.d.ts:8561:5: stage 0 can't lower a function without a body yet"
    },
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-sdk-c-public/consumer.ts:2:1",
      "phase": "lowering",
      "kind": "Boundary",
      "where": "/workspace/scratch/baobab-sdk-c-public/consumer.ts:3:5",
      "reason": "failed statement; state rolled back; continued at next statement or declaration",
      "text": "/workspace/scratch/baobab-sdk-c-public/typescript.d.ts:8561:5: stage 0 can't lower a function without a body yet",
      "start": 104,
      "end": 158
    }
  ]
}
````


**c-public-unstripped**

````json
{
  "production_checker": {
    "accepted": true,
    "diagnostics": [],
    "driver": "production load.Load, no lowering or output IR",
    "roots": 1,
    "seconds": 0.181255141
  },
  "cohere_findings": [
    {
      "kind": "finding",
      "path": "/workspace/adamic/stage3/census/sourcefile-like-sites/continuation/c-public-unstripped/public-adamic-consumer.ts",
      "line": 3,
      "column": 45,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`file` is a 'SourceFile', seen here as 'LineMapSource', which adds the optional property `lineMap` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `lineMap` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a 'readonly number[] | undefined' and holds something else. Declare `lineMap` in the source's type, or build the value with the property named."
    }
  ],
  "guarded_findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-sdk-c-public-unstripped/consumer.ts:2:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-sdk-c-public-unstripped/consumer.ts:3:45",
      "reason": "optional property lineMap in LineMapSource absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-sdk-c-public-unstripped/consumer.ts:3:45: Adamic 0.1 refuses optional property lineMap in LineMapSource absent from structural source SourceFile, which can hide fields; declare lineMap on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    },
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-sdk-c-public-unstripped/consumer.ts:2:1",
      "phase": "lowering",
      "kind": "NotYet",
      "where": "/workspace/scratch/baobab-sdk-c-public-unstripped/typescript.d.ts:8561:5",
      "reason": "a function without a body",
      "text": "/workspace/scratch/baobab-sdk-c-public-unstripped/typescript.d.ts:8561:5: stage 0 can't lower a function without a body yet"
    },
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-sdk-c-public-unstripped/consumer.ts:2:1",
      "phase": "lowering",
      "kind": "Boundary",
      "where": "/workspace/scratch/baobab-sdk-c-public-unstripped/consumer.ts:3:5",
      "reason": "failed statement; state rolled back; continued at next statement or declaration",
      "text": "/workspace/scratch/baobab-sdk-c-public-unstripped/typescript.d.ts:8561:5: stage 0 can't lower a function without a body yet",
      "start": 104,
      "end": 158
    }
  ]
}
````


**c-public-readonly-sdk-control**

````json
{
  "production_checker": {
    "accepted": true,
    "diagnostics": [],
    "driver": "production load.Load, no lowering or output IR",
    "roots": 1,
    "seconds": 0.161772996
  },
  "cohere_findings": [
    {
      "kind": "finding",
      "path": "/workspace/adamic/stage3/census/sourcefile-like-sites/continuation/c-public-readonly-sdk-control/public-adamic-consumer.ts",
      "line": 3,
      "column": 45,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`file` is a 'SourceFile', seen here as 'LineMapSource', which adds the optional property `lineMap` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `lineMap` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a 'readonly number[] | undefined' and holds something else. Declare `lineMap` in the source's type, or build the value with the property named."
    }
  ],
  "guarded_findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-sdk-c-public-readonly-sdk-control/consumer.ts:2:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-sdk-c-public-readonly-sdk-control/consumer.ts:3:45",
      "reason": "optional property lineMap in LineMapSource absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-sdk-c-public-readonly-sdk-control/consumer.ts:3:45: Adamic 0.1 refuses optional property lineMap in LineMapSource absent from structural source SourceFile, which can hide fields; declare lineMap on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    },
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-sdk-c-public-readonly-sdk-control/consumer.ts:2:1",
      "phase": "lowering",
      "kind": "NotYet",
      "where": "/workspace/scratch/baobab-sdk-c-public-readonly-sdk-control/typescript.d.ts:8561:5",
      "reason": "a function without a body",
      "text": "/workspace/scratch/baobab-sdk-c-public-readonly-sdk-control/typescript.d.ts:8561:5: stage 0 can't lower a function without a body yet"
    },
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-sdk-c-public-readonly-sdk-control/consumer.ts:2:1",
      "phase": "lowering",
      "kind": "Boundary",
      "where": "/workspace/scratch/baobab-sdk-c-public-readonly-sdk-control/consumer.ts:3:5",
      "reason": "failed statement; state rolled back; continued at next statement or declaration",
      "text": "/workspace/scratch/baobab-sdk-c-public-readonly-sdk-control/typescript.d.ts:8561:5: stage 0 can't lower a function without a body yet",
      "start": 104,
      "end": 158
    }
  ]
}
````



### Every original view under both forms of (c)

Each row keeps the original location, phase, enclosing function, callee, every current finding at that argument/call, and checker diagnostics. When a call has two original views its argument span is used separately, as in the earlier comparison. The cohere list includes existing non-null assertion findings on enclosing calls, so relation acceptance is not confused with an all-rules-clean source expression.

#### (c), internal LineMapSource

| Original location | Phase | Enclosing function | Callee | Measured classification |
| --- | --- | --- | --- | --- |
| emitter.ts:1444:66 | refusal_scan | `getCurrentLineMap` | `getLineStarts` | cleared at original view |
| emitter.ts:3961:112 | refusal_scan | `emitJsxExpression` | `getLineAndCharacterOfPosition` | cleared at original view |
| emitter.ts:3961:180 | refusal_scan | `emitJsxExpression` | `getLineAndCharacterOfPosition` | cleared at original view |
| emitter.ts:6233:96 | refusal_scan | `emitPos` | `getLineAndCharacterOfPosition` | cleared at original view |
| executeCommandLine.ts:109:41 | refusal_scan | `<anonymous@107>` | `getLineStarts` | cleared at original view |
| parser.ts:10032:36 | refusal_scan | `updateSourceFile` | `transferSourceFileChildren` | unchanged finding |
| parser.ts:10032:48 | refusal_scan | `updateSourceFile` | `transferSourceFileChildren` | unchanged finding |
| parser.ts:10104:37 | refusal_scan | `visitNode` | `unsetNodeChildren` | unchanged finding |
| parser.ts:10261:42 | refusal_scan | `visitNode` | `unsetNodeChildren` | unchanged finding |
| program.ts:2973:42 | refusal_scan | `markPrecedingCommentDirectiveLine` | `getLineStarts` | cleared at original view |
| program.ts:671:67 | refusal_scan | `formatDiagnostic` | `getLineAndCharacterOfPosition` | cleared at original view |
| program.ts:713:89 | refusal_scan | `formatCodeSpan` | `getLineAndCharacterOfPosition` | cleared at original view |
| program.ts:714:87 | refusal_scan | `formatCodeSpan` | `getLineAndCharacterOfPosition` | cleared at original view |
| program.ts:715:58 | refusal_scan | `formatCodeSpan` | `getLineAndCharacterOfPosition` | cleared at original view |
| program.ts:733:57 | refusal_scan | `formatCodeSpan` | `getPositionOfLineAndCharacter` | unchanged finding |
| program.ts:734:76 | refusal_scan | `formatCodeSpan` | `getPositionOfLineAndCharacter` | unchanged finding |
| program.ts:768:89 | refusal_scan | `formatLocation` | `getLineAndCharacterOfPosition` | cleared at original view |
| tracing.ts:206:67 | refusal_scan | `getLocation` | `getLineAndCharacterOfPosition` | cleared at original view |
| tracing.ts:207:65 | refusal_scan | `getLocation` | `getLineAndCharacterOfPosition` | cleared at original view |
| transformers/declarations.ts:1264:139 | refusal_scan | `visitDeclarationSubtree` | `getLineAndCharacterOfPosition` | cleared at original view |
| transformers/declarations.ts:1264:70 | refusal_scan | `visitDeclarationSubtree` | `getLineAndCharacterOfPosition` | cleared at original view |
| transformers/jsx.ts:359:63 | refusal_scan | `visitJsxOpeningLikeElementOrFragmentJSX` | `getLineAndCharacterOfPosition` | cleared at original view |
| utilities.ts:1025:47 | refusal_scan | `nodePosToString` | `getLineAndCharacterOfPosition` | cleared at original view |
| utilities.ts:1213:46 | refusal_scan | `<anonymous@1212>` | `getLineAndCharacterOfPosition` | cleared at original view |
| utilities.ts:2526:67 | refusal_scan | `getErrorSpanForArrowFunction` | `getLineAndCharacterOfPosition` | cleared at original view |
| utilities.ts:2527:65 | refusal_scan | `getErrorSpanForArrowFunction` | `getLineAndCharacterOfPosition` | cleared at original view |
| utilities.ts:2531:70 | refusal_scan | `getErrorSpanForArrowFunction` | `getEndLinePosition` | unchanged finding |
| utilities.ts:6748:38 | refusal_scan | `getLineOfLocalPosition` | `getLineStarts` | cleared at original view |
| utilities.ts:7944:37 | refusal_scan | `getLinesBetweenRangeEndAndRangeStart` | `getLinesBetweenPositions` | unchanged finding |
| utilities.ts:7949:37 | refusal_scan | `getLinesBetweenRangeEndPositions` | `getLinesBetweenPositions` | unchanged finding |
| utilities.ts:7959:37 | refusal_scan | `positionsAreOnSameLine` | `getLinesBetweenPositions` | unchanged finding |
| utilities.ts:7971:37 | refusal_scan | `getLinesBetweenPositionAndPrecedingNonWhitespaceCharacter` | `getLinesBetweenPositions` | unchanged finding |
| utilities.ts:7977:37 | refusal_scan | `getLinesBetweenPositionAndNextNonWhitespaceCharacter` | `getLinesBetweenPositions` | unchanged finding |
| watch.ts:249:60 | refusal_scan | `<anonymous@241>` | `getLineAndCharacterOfPosition` | cleared at original view |

**emitter.ts:1444:66** `getLineStarts`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/emitter.ts:1443:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 58788,
      "body_end": 58904,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getCurrentLineMap",
      "parent": "/workspace/scratch/baobab-road2-c/src/compiler/emitter.ts:1211:1",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/emitter.ts:1211:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": []
}
````


**emitter.ts:3961:112** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/emitter.ts:3959:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 170721,
      "body_end": 171620,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "emitJsxExpression",
      "parent": "/workspace/scratch/baobab-road2-c/src/compiler/emitter.ts:1211:1",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/emitter.ts:1211:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": []
}
````


**emitter.ts:3961:180** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/emitter.ts:3959:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 170721,
      "body_end": 171620,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "emitJsxExpression",
      "parent": "/workspace/scratch/baobab-road2-c/src/compiler/emitter.ts:1211:1",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/emitter.ts:1211:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": []
}
````


**emitter.ts:6233:96** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/emitter.ts:6228:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 268634,
      "body_end": 269150,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "emitPos",
      "parent": "/workspace/scratch/baobab-road2-c/src/compiler/emitter.ts:1211:1",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/emitter.ts:1211:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": []
}
````


**executeCommandLine.ts:109:41** `getLineStarts`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/executeCommandLine.ts:105:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 2705,
      "body_end": 2984,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "countLines",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/executeCommandLine.ts:105:1"
    }
  ],
  "cohere_findings_at_call": []
}
````


**parser.ts:10032:36** `transferSourceFileChildren`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:9947:5",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:10032:36",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:10032:36: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:9947:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 497989,
      "body_end": 503771,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "updateSourceFile",
      "parent": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:9946:1",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:9946:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts",
      "line": 10032,
      "column": 36,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    },
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts",
      "line": 10032,
      "column": 48,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`result` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    }
  ]
}
````


**parser.ts:10032:48** `transferSourceFileChildren`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:9947:5",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:10032:48",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:10032:48: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:9947:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 497989,
      "body_end": 503771,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "updateSourceFile",
      "parent": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:9946:1",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:9946:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts",
      "line": 10032,
      "column": 36,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    },
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts",
      "line": 10032,
      "column": 48,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`result` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    }
  ]
}
````


**parser.ts:10104:37** `unsetNodeChildren`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:10096:9",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:10104:37",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:10104:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:10096:9",
      "kind": "KindFunctionDeclaration",
      "body_start": 506691,
      "body_end": 507655,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "visitNode",
      "parent": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:10087:5",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:9946:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts",
      "line": 10104,
      "column": 37,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`origSourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    }
  ]
}
````


**parser.ts:10261:42** `unsetNodeChildren`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:10246:9",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:10261:42",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:10261:42: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:10246:9",
      "kind": "KindFunctionDeclaration",
      "body_start": 513350,
      "body_end": 515010,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "visitNode",
      "parent": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:10233:5",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts:9946:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/parser.ts",
      "line": 10261,
      "column": 42,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    }
  ]
}
````


**program.ts:2973:42** `getLineStarts`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:2966:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 144938,
      "body_end": 145801,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "markPrecedingCommentDirectiveLine",
      "parent": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:1516:1",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:1516:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": []
}
````


**program.ts:671:67** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:667:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 24128,
      "body_end": 24786,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatDiagnostic",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:667:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts",
      "line": 671,
      "column": 84,
      "severity": "error",
      "rule": "@typescript-eslint/no-non-null-assertion",
      "messageId": "noNonNull",
      "message": "A non-null assertion tells the compiler to stop checking, and it is erased before the code runs, so when the value is missing anyway the failure lands somewhere else, as a property access on `undefined` in code the types promised was safe. Fix it in this order. If the value can really be missing on a reachable path (input, disk, the network, an external API), guard it: return early, `continue`, or throw an error that says what was missing, because that is the case the type is warning about. If it cannot, fix the type that says it can (a type predicate in `.filter`, a tuple type, a field made required), so there is no claim left to check. If an index is what makes it optional, restructure so the index goes away (`for...of`, `.entries()`, a destructured head checked once), because a value you never index is never possibly missing. Where none of those fit, state the proof with `assert(value !== undefined, 'why it holds')` or, in an expression, `required(value, 'why it holds')` from `@nexus/source/errors/Assert`, which narrows the same way and throws that sentence if the proof is ever wrong. `?.`, `?? ''` and `as T` are not repairs on their own: each makes the missing value disappear instead of handling it."
    }
  ]
}
````


**program.ts:713:89** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:712:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 26062,
      "body_end": 28811,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatCodeSpan",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:712:1"
    }
  ],
  "cohere_findings_at_call": []
}
````


**program.ts:714:87** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:712:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 26062,
      "body_end": 28811,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatCodeSpan",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:712:1"
    }
  ],
  "cohere_findings_at_call": []
}
````


**program.ts:715:58** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:712:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 26062,
      "body_end": 28811,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatCodeSpan",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:712:1"
    }
  ],
  "cohere_findings_at_call": []
}
````


**program.ts:733:57** `getPositionOfLineAndCharacter`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:712:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:733:57",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:733:57: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:712:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 26062,
      "body_end": 28811,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatCodeSpan",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:712:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts",
      "line": 733,
      "column": 57,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    }
  ]
}
````


**program.ts:734:76** `getPositionOfLineAndCharacter`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:712:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:734:76",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:734:76: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:712:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 26062,
      "body_end": 28811,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatCodeSpan",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:712:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts",
      "line": 734,
      "column": 76,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    }
  ]
}
````


**program.ts:768:89** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:767:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 28990,
      "body_end": 29611,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatLocation",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/program.ts:767:1"
    }
  ],
  "cohere_findings_at_call": []
}
````


**tracing.ts:206:67** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/tracing.ts:200:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 7955,
      "body_end": 8503,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLocation",
      "parent": "/workspace/scratch/baobab-road2-c/src/compiler/tracing.ts:38:1",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/tracing.ts:38:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/tracing.ts",
      "line": 206,
      "column": 73,
      "severity": "error",
      "rule": "@typescript-eslint/no-non-null-assertion",
      "messageId": "noNonNull",
      "message": "A non-null assertion tells the compiler to stop checking, and it is erased before the code runs, so when the value is missing anyway the failure lands somewhere else, as a property access on `undefined` in code the types promised was safe. Fix it in this order. If the value can really be missing on a reachable path (input, disk, the network, an external API), guard it: return early, `continue`, or throw an error that says what was missing, because that is the case the type is warning about. If it cannot, fix the type that says it can (a type predicate in `.filter`, a tuple type, a field made required), so there is no claim left to check. If an index is what makes it optional, restructure so the index goes away (`for...of`, `.entries()`, a destructured head checked once), because a value you never index is never possibly missing. Where none of those fit, state the proof with `assert(value !== undefined, 'why it holds')` or, in an expression, `required(value, 'why it holds')` from `@nexus/source/errors/Assert`, which narrows the same way and throws that sentence if the proof is ever wrong. `?.`, `?? ''` and `as T` are not repairs on their own: each makes the missing value disappear instead of handling it."
    }
  ]
}
````


**tracing.ts:207:65** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/tracing.ts:200:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 7955,
      "body_end": 8503,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLocation",
      "parent": "/workspace/scratch/baobab-road2-c/src/compiler/tracing.ts:38:1",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/tracing.ts:38:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/tracing.ts",
      "line": 207,
      "column": 71,
      "severity": "error",
      "rule": "@typescript-eslint/no-non-null-assertion",
      "messageId": "noNonNull",
      "message": "A non-null assertion tells the compiler to stop checking, and it is erased before the code runs, so when the value is missing anyway the failure lands somewhere else, as a property access on `undefined` in code the types promised was safe. Fix it in this order. If the value can really be missing on a reachable path (input, disk, the network, an external API), guard it: return early, `continue`, or throw an error that says what was missing, because that is the case the type is warning about. If it cannot, fix the type that says it can (a type predicate in `.filter`, a tuple type, a field made required), so there is no claim left to check. If an index is what makes it optional, restructure so the index goes away (`for...of`, `.entries()`, a destructured head checked once), because a value you never index is never possibly missing. Where none of those fit, state the proof with `assert(value !== undefined, 'why it holds')` or, in an expression, `required(value, 'why it holds')` from `@nexus/source/errors/Assert`, which narrows the same way and throws that sentence if the proof is ever wrong. `?.`, `?? ''` and `as T` are not repairs on their own: each makes the missing value disappear instead of handling it."
    }
  ]
}
````


**transformers/declarations.ts:1264:139** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/transformers/declarations.ts:1006:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 47618,
      "body_end": 63600,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "visitDeclarationSubtree",
      "parent": "/workspace/scratch/baobab-road2-c/src/compiler/transformers/declarations.ts:259:1",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/transformers/declarations.ts:259:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": []
}
````


**transformers/declarations.ts:1264:70** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/transformers/declarations.ts:1006:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 47618,
      "body_end": 63600,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "visitDeclarationSubtree",
      "parent": "/workspace/scratch/baobab-road2-c/src/compiler/transformers/declarations.ts:259:1",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/transformers/declarations.ts:259:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": []
}
````


**transformers/jsx.ts:359:63** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/transformers/jsx.ts:333:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 14898,
      "body_end": 16980,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "visitJsxOpeningLikeElementOrFragmentJSX",
      "parent": "/workspace/scratch/baobab-road2-c/src/compiler/transformers/jsx.ts:92:1",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/transformers/jsx.ts:92:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": []
}
````


**utilities.ts:1025:47** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:1023:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 33483,
      "body_end": 33667,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "nodePosToString",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:1023:1"
    }
  ],
  "cohere_findings_at_call": []
}
````


**utilities.ts:1213:46** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:1210:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 42294,
      "body_end": 43143,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "createCommentDirectivesMap",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:1210:1"
    }
  ],
  "cohere_findings_at_call": []
}
````


**utilities.ts:2526:67** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:2523:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 87836,
      "body_end": 88470,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getErrorSpanForArrowFunction",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:2523:1"
    }
  ],
  "cohere_findings_at_call": []
}
````


**utilities.ts:2527:65** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:2523:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 87836,
      "body_end": 88470,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getErrorSpanForArrowFunction",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:2523:1"
    }
  ],
  "cohere_findings_at_call": []
}
````


**utilities.ts:2531:70** `getEndLinePosition`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:2523:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:2531:70",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:2531:70: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:2523:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 87836,
      "body_end": 88470,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getErrorSpanForArrowFunction",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:2523:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts",
      "line": 2531,
      "column": 70,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    }
  ]
}
````


**utilities.ts:6748:38** `getLineStarts`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:6747:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 266234,
      "body_end": 266342,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLineOfLocalPosition",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:6747:1"
    }
  ],
  "cohere_findings_at_call": []
}
````


**utilities.ts:7944:37** `getLinesBetweenPositions`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7942:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7944:37",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7944:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7942:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 313431,
      "body_end": 313609,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLinesBetweenRangeEndAndRangeStart",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7942:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts",
      "line": 7944,
      "column": 37,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    }
  ]
}
````


**utilities.ts:7949:37** `getLinesBetweenPositions`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7948:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7949:37",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7949:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7948:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 313781,
      "body_end": 313860,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLinesBetweenRangeEndPositions",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7948:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts",
      "line": 7949,
      "column": 37,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    }
  ]
}
````


**utilities.ts:7959:37** `getLinesBetweenPositions`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7958:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7959:37",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7959:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7958:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 314169,
      "body_end": 314242,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "positionsAreOnSameLine",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7958:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts",
      "line": 7959,
      "column": 37,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    }
  ]
}
````


**utilities.ts:7971:37** `getLinesBetweenPositions`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7968:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7971:37",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7971:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7968:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 314729,
      "body_end": 315003,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLinesBetweenPositionAndPrecedingNonWhitespaceCharacter",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7968:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts",
      "line": 7971,
      "column": 37,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    }
  ]
}
````


**utilities.ts:7977:37** `getLinesBetweenPositions`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7975:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7977:37",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7977:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7975:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 315182,
      "body_end": 315372,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLinesBetweenPositionAndNextNonWhitespaceCharacter",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts:7975:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/utilities.ts",
      "line": 7977,
      "column": 37,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    }
  ]
}
````


**watch.ts:249:60** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c/src/compiler/watch.ts:233:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 8061,
      "body_end": 8940,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getFilesInErrorForSummary",
      "top": "/workspace/scratch/baobab-road2-c/src/compiler/watch.ts:233:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/watch.ts",
      "line": 249,
      "column": 60,
      "severity": "error",
      "rule": "@typescript-eslint/no-non-null-assertion",
      "messageId": "noNonNull",
      "message": "A non-null assertion tells the compiler to stop checking, and it is erased before the code runs, so when the value is missing anyway the failure lands somewhere else, as a property access on `undefined` in code the types promised was safe. Fix it in this order. If the value can really be missing on a reachable path (input, disk, the network, an external API), guard it: return early, `continue`, or throw an error that says what was missing, because that is the case the type is warning about. If it cannot, fix the type that says it can (a type predicate in `.filter`, a tuple type, a field made required), so there is no claim left to check. If an index is what makes it optional, restructure so the index goes away (`for...of`, `.entries()`, a destructured head checked once), because a value you never index is never possibly missing. Where none of those fit, state the proof with `assert(value !== undefined, 'why it holds')` or, in an expression, `required(value, 'why it holds')` from `@nexus/source/errors/Assert`, which narrows the same way and throws that sentence if the proof is ever wrong. `?.`, `?? ''` and `as T` are not repairs on their own: each makes the missing value disappear instead of handling it."
    },
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c/src/compiler/watch.ts",
      "line": 249,
      "column": 89,
      "severity": "error",
      "rule": "@typescript-eslint/no-non-null-assertion",
      "messageId": "noNonNull",
      "message": "A non-null assertion tells the compiler to stop checking, and it is erased before the code runs, so when the value is missing anyway the failure lands somewhere else, as a property access on `undefined` in code the types promised was safe. Fix it in this order. If the value can really be missing on a reachable path (input, disk, the network, an external API), guard it: return early, `continue`, or throw an error that says what was missing, because that is the case the type is warning about. If it cannot, fix the type that says it can (a type predicate in `.filter`, a tuple type, a field made required), so there is no claim left to check. If an index is what makes it optional, restructure so the index goes away (`for...of`, `.entries()`, a destructured head checked once), because a value you never index is never possibly missing. Where none of those fit, state the proof with `assert(value !== undefined, 'why it holds')` or, in an expression, `required(value, 'why it holds')` from `@nexus/source/errors/Assert`, which narrows the same way and throws that sentence if the proof is ever wrong. `?.`, `?? ''` and `as T` are not repairs on their own: each makes the missing value disappear instead of handling it."
    }
  ]
}
````


#### (c), public LineMapSource

| Original location | Phase | Enclosing function | Callee | Measured classification |
| --- | --- | --- | --- | --- |
| emitter.ts:1444:66 | refusal_scan | `getCurrentLineMap` | `getLineStarts` | cleared at original view |
| emitter.ts:3961:112 | refusal_scan | `emitJsxExpression` | `getLineAndCharacterOfPosition` | cleared at original view |
| emitter.ts:3961:180 | refusal_scan | `emitJsxExpression` | `getLineAndCharacterOfPosition` | cleared at original view |
| emitter.ts:6233:96 | refusal_scan | `emitPos` | `getLineAndCharacterOfPosition` | cleared at original view |
| executeCommandLine.ts:109:41 | refusal_scan | `<anonymous@107>` | `getLineStarts` | cleared at original view |
| parser.ts:10032:36 | refusal_scan | `updateSourceFile` | `transferSourceFileChildren` | unchanged finding |
| parser.ts:10032:48 | refusal_scan | `updateSourceFile` | `transferSourceFileChildren` | unchanged finding |
| parser.ts:10104:37 | refusal_scan | `visitNode` | `unsetNodeChildren` | unchanged finding |
| parser.ts:10261:42 | refusal_scan | `visitNode` | `unsetNodeChildren` | unchanged finding |
| program.ts:2973:42 | refusal_scan | `markPrecedingCommentDirectiveLine` | `getLineStarts` | cleared at original view |
| program.ts:671:67 | refusal_scan | `formatDiagnostic` | `getLineAndCharacterOfPosition` | cleared at original view |
| program.ts:713:89 | refusal_scan | `formatCodeSpan` | `getLineAndCharacterOfPosition` | cleared at original view |
| program.ts:714:87 | refusal_scan | `formatCodeSpan` | `getLineAndCharacterOfPosition` | cleared at original view |
| program.ts:715:58 | refusal_scan | `formatCodeSpan` | `getLineAndCharacterOfPosition` | cleared at original view |
| program.ts:733:57 | refusal_scan | `formatCodeSpan` | `getPositionOfLineAndCharacter` | unchanged finding |
| program.ts:734:76 | refusal_scan | `formatCodeSpan` | `getPositionOfLineAndCharacter` | unchanged finding |
| program.ts:768:89 | refusal_scan | `formatLocation` | `getLineAndCharacterOfPosition` | cleared at original view |
| tracing.ts:206:67 | refusal_scan | `getLocation` | `getLineAndCharacterOfPosition` | cleared at original view |
| tracing.ts:207:65 | refusal_scan | `getLocation` | `getLineAndCharacterOfPosition` | cleared at original view |
| transformers/declarations.ts:1264:139 | refusal_scan | `visitDeclarationSubtree` | `getLineAndCharacterOfPosition` | cleared at original view |
| transformers/declarations.ts:1264:70 | refusal_scan | `visitDeclarationSubtree` | `getLineAndCharacterOfPosition` | cleared at original view |
| transformers/jsx.ts:359:63 | refusal_scan | `visitJsxOpeningLikeElementOrFragmentJSX` | `getLineAndCharacterOfPosition` | cleared at original view |
| utilities.ts:1025:47 | refusal_scan | `nodePosToString` | `getLineAndCharacterOfPosition` | cleared at original view |
| utilities.ts:1213:46 | refusal_scan | `<anonymous@1212>` | `getLineAndCharacterOfPosition` | cleared at original view |
| utilities.ts:2526:67 | refusal_scan | `getErrorSpanForArrowFunction` | `getLineAndCharacterOfPosition` | cleared at original view |
| utilities.ts:2527:65 | refusal_scan | `getErrorSpanForArrowFunction` | `getLineAndCharacterOfPosition` | cleared at original view |
| utilities.ts:2531:70 | refusal_scan | `getErrorSpanForArrowFunction` | `getEndLinePosition` | unchanged finding |
| utilities.ts:6748:38 | refusal_scan | `getLineOfLocalPosition` | `getLineStarts` | cleared at original view |
| utilities.ts:7944:37 | refusal_scan | `getLinesBetweenRangeEndAndRangeStart` | `getLinesBetweenPositions` | unchanged finding |
| utilities.ts:7949:37 | refusal_scan | `getLinesBetweenRangeEndPositions` | `getLinesBetweenPositions` | unchanged finding |
| utilities.ts:7959:37 | refusal_scan | `positionsAreOnSameLine` | `getLinesBetweenPositions` | unchanged finding |
| utilities.ts:7971:37 | refusal_scan | `getLinesBetweenPositionAndPrecedingNonWhitespaceCharacter` | `getLinesBetweenPositions` | unchanged finding |
| utilities.ts:7977:37 | refusal_scan | `getLinesBetweenPositionAndNextNonWhitespaceCharacter` | `getLinesBetweenPositions` | unchanged finding |
| watch.ts:249:60 | refusal_scan | `<anonymous@241>` | `getLineAndCharacterOfPosition` | cleared at original view |

**emitter.ts:1444:66** `getLineStarts`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/emitter.ts:1443:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 58788,
      "body_end": 58904,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getCurrentLineMap",
      "parent": "/workspace/scratch/baobab-road2-c-public/src/compiler/emitter.ts:1211:1",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/emitter.ts:1211:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": []
}
````


**emitter.ts:3961:112** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/emitter.ts:3959:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 170721,
      "body_end": 171620,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "emitJsxExpression",
      "parent": "/workspace/scratch/baobab-road2-c-public/src/compiler/emitter.ts:1211:1",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/emitter.ts:1211:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": []
}
````


**emitter.ts:3961:180** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/emitter.ts:3959:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 170721,
      "body_end": 171620,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "emitJsxExpression",
      "parent": "/workspace/scratch/baobab-road2-c-public/src/compiler/emitter.ts:1211:1",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/emitter.ts:1211:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": []
}
````


**emitter.ts:6233:96** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/emitter.ts:6228:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 268634,
      "body_end": 269150,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "emitPos",
      "parent": "/workspace/scratch/baobab-road2-c-public/src/compiler/emitter.ts:1211:1",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/emitter.ts:1211:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": []
}
````


**executeCommandLine.ts:109:41** `getLineStarts`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/executeCommandLine.ts:105:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 2705,
      "body_end": 2984,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "countLines",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/executeCommandLine.ts:105:1"
    }
  ],
  "cohere_findings_at_call": []
}
````


**parser.ts:10032:36** `transferSourceFileChildren`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:9947:5",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:10032:36",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:10032:36: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:9947:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 497989,
      "body_end": 503771,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "updateSourceFile",
      "parent": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:9946:1",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:9946:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts",
      "line": 10032,
      "column": 36,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    },
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts",
      "line": 10032,
      "column": 48,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`result` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    }
  ]
}
````


**parser.ts:10032:48** `transferSourceFileChildren`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:9947:5",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:10032:48",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:10032:48: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:9947:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 497989,
      "body_end": 503771,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "updateSourceFile",
      "parent": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:9946:1",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:9946:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts",
      "line": 10032,
      "column": 36,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    },
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts",
      "line": 10032,
      "column": 48,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`result` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    }
  ]
}
````


**parser.ts:10104:37** `unsetNodeChildren`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:10096:9",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:10104:37",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:10104:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:10096:9",
      "kind": "KindFunctionDeclaration",
      "body_start": 506691,
      "body_end": 507655,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "visitNode",
      "parent": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:10087:5",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:9946:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts",
      "line": 10104,
      "column": 37,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`origSourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    }
  ]
}
````


**parser.ts:10261:42** `unsetNodeChildren`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:10246:9",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:10261:42",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:10261:42: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:10246:9",
      "kind": "KindFunctionDeclaration",
      "body_start": 513350,
      "body_end": 515010,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "visitNode",
      "parent": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:10233:5",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts:9946:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/parser.ts",
      "line": 10261,
      "column": 42,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    }
  ]
}
````


**program.ts:2973:42** `getLineStarts`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:2966:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 144938,
      "body_end": 145801,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "markPrecedingCommentDirectiveLine",
      "parent": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:1516:1",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:1516:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": []
}
````


**program.ts:671:67** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:667:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 24128,
      "body_end": 24786,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatDiagnostic",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:667:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts",
      "line": 671,
      "column": 84,
      "severity": "error",
      "rule": "@typescript-eslint/no-non-null-assertion",
      "messageId": "noNonNull",
      "message": "A non-null assertion tells the compiler to stop checking, and it is erased before the code runs, so when the value is missing anyway the failure lands somewhere else, as a property access on `undefined` in code the types promised was safe. Fix it in this order. If the value can really be missing on a reachable path (input, disk, the network, an external API), guard it: return early, `continue`, or throw an error that says what was missing, because that is the case the type is warning about. If it cannot, fix the type that says it can (a type predicate in `.filter`, a tuple type, a field made required), so there is no claim left to check. If an index is what makes it optional, restructure so the index goes away (`for...of`, `.entries()`, a destructured head checked once), because a value you never index is never possibly missing. Where none of those fit, state the proof with `assert(value !== undefined, 'why it holds')` or, in an expression, `required(value, 'why it holds')` from `@nexus/source/errors/Assert`, which narrows the same way and throws that sentence if the proof is ever wrong. `?.`, `?? ''` and `as T` are not repairs on their own: each makes the missing value disappear instead of handling it."
    }
  ]
}
````


**program.ts:713:89** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:712:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 26062,
      "body_end": 28811,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatCodeSpan",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:712:1"
    }
  ],
  "cohere_findings_at_call": []
}
````


**program.ts:714:87** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:712:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 26062,
      "body_end": 28811,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatCodeSpan",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:712:1"
    }
  ],
  "cohere_findings_at_call": []
}
````


**program.ts:715:58** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:712:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 26062,
      "body_end": 28811,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatCodeSpan",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:712:1"
    }
  ],
  "cohere_findings_at_call": []
}
````


**program.ts:733:57** `getPositionOfLineAndCharacter`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:712:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:733:57",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:733:57: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:712:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 26062,
      "body_end": 28811,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatCodeSpan",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:712:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts",
      "line": 733,
      "column": 57,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    }
  ]
}
````


**program.ts:734:76** `getPositionOfLineAndCharacter`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:712:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:734:76",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:734:76: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:712:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 26062,
      "body_end": 28811,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatCodeSpan",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:712:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts",
      "line": 734,
      "column": 76,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    }
  ]
}
````


**program.ts:768:89** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:767:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 28990,
      "body_end": 29611,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatLocation",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/program.ts:767:1"
    }
  ],
  "cohere_findings_at_call": []
}
````


**tracing.ts:206:67** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/tracing.ts:200:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 7955,
      "body_end": 8503,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLocation",
      "parent": "/workspace/scratch/baobab-road2-c-public/src/compiler/tracing.ts:38:1",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/tracing.ts:38:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/tracing.ts",
      "line": 206,
      "column": 73,
      "severity": "error",
      "rule": "@typescript-eslint/no-non-null-assertion",
      "messageId": "noNonNull",
      "message": "A non-null assertion tells the compiler to stop checking, and it is erased before the code runs, so when the value is missing anyway the failure lands somewhere else, as a property access on `undefined` in code the types promised was safe. Fix it in this order. If the value can really be missing on a reachable path (input, disk, the network, an external API), guard it: return early, `continue`, or throw an error that says what was missing, because that is the case the type is warning about. If it cannot, fix the type that says it can (a type predicate in `.filter`, a tuple type, a field made required), so there is no claim left to check. If an index is what makes it optional, restructure so the index goes away (`for...of`, `.entries()`, a destructured head checked once), because a value you never index is never possibly missing. Where none of those fit, state the proof with `assert(value !== undefined, 'why it holds')` or, in an expression, `required(value, 'why it holds')` from `@nexus/source/errors/Assert`, which narrows the same way and throws that sentence if the proof is ever wrong. `?.`, `?? ''` and `as T` are not repairs on their own: each makes the missing value disappear instead of handling it."
    }
  ]
}
````


**tracing.ts:207:65** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/tracing.ts:200:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 7955,
      "body_end": 8503,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLocation",
      "parent": "/workspace/scratch/baobab-road2-c-public/src/compiler/tracing.ts:38:1",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/tracing.ts:38:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/tracing.ts",
      "line": 207,
      "column": 71,
      "severity": "error",
      "rule": "@typescript-eslint/no-non-null-assertion",
      "messageId": "noNonNull",
      "message": "A non-null assertion tells the compiler to stop checking, and it is erased before the code runs, so when the value is missing anyway the failure lands somewhere else, as a property access on `undefined` in code the types promised was safe. Fix it in this order. If the value can really be missing on a reachable path (input, disk, the network, an external API), guard it: return early, `continue`, or throw an error that says what was missing, because that is the case the type is warning about. If it cannot, fix the type that says it can (a type predicate in `.filter`, a tuple type, a field made required), so there is no claim left to check. If an index is what makes it optional, restructure so the index goes away (`for...of`, `.entries()`, a destructured head checked once), because a value you never index is never possibly missing. Where none of those fit, state the proof with `assert(value !== undefined, 'why it holds')` or, in an expression, `required(value, 'why it holds')` from `@nexus/source/errors/Assert`, which narrows the same way and throws that sentence if the proof is ever wrong. `?.`, `?? ''` and `as T` are not repairs on their own: each makes the missing value disappear instead of handling it."
    }
  ]
}
````


**transformers/declarations.ts:1264:139** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/transformers/declarations.ts:1006:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 47618,
      "body_end": 63600,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "visitDeclarationSubtree",
      "parent": "/workspace/scratch/baobab-road2-c-public/src/compiler/transformers/declarations.ts:259:1",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/transformers/declarations.ts:259:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": []
}
````


**transformers/declarations.ts:1264:70** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/transformers/declarations.ts:1006:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 47618,
      "body_end": 63600,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "visitDeclarationSubtree",
      "parent": "/workspace/scratch/baobab-road2-c-public/src/compiler/transformers/declarations.ts:259:1",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/transformers/declarations.ts:259:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": []
}
````


**transformers/jsx.ts:359:63** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/transformers/jsx.ts:333:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 14898,
      "body_end": 16980,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "visitJsxOpeningLikeElementOrFragmentJSX",
      "parent": "/workspace/scratch/baobab-road2-c-public/src/compiler/transformers/jsx.ts:92:1",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/transformers/jsx.ts:92:1",
      "depth": 1
    }
  ],
  "cohere_findings_at_call": []
}
````


**utilities.ts:1025:47** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:1023:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 33483,
      "body_end": 33667,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "nodePosToString",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:1023:1"
    }
  ],
  "cohere_findings_at_call": []
}
````


**utilities.ts:1213:46** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:1210:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 42294,
      "body_end": 43143,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "createCommentDirectivesMap",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:1210:1"
    }
  ],
  "cohere_findings_at_call": []
}
````


**utilities.ts:2526:67** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:2523:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 87836,
      "body_end": 88470,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getErrorSpanForArrowFunction",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:2523:1"
    }
  ],
  "cohere_findings_at_call": []
}
````


**utilities.ts:2527:65** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:2523:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 87836,
      "body_end": 88470,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getErrorSpanForArrowFunction",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:2523:1"
    }
  ],
  "cohere_findings_at_call": []
}
````


**utilities.ts:2531:70** `getEndLinePosition`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:2523:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:2531:70",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:2531:70: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:2523:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 87836,
      "body_end": 88470,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getErrorSpanForArrowFunction",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:2523:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts",
      "line": 2531,
      "column": 70,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    }
  ]
}
````


**utilities.ts:6748:38** `getLineStarts`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:6747:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 266234,
      "body_end": 266342,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLineOfLocalPosition",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:6747:1"
    }
  ],
  "cohere_findings_at_call": []
}
````


**utilities.ts:7944:37** `getLinesBetweenPositions`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7942:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7944:37",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7944:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7942:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 313431,
      "body_end": 313609,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLinesBetweenRangeEndAndRangeStart",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7942:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts",
      "line": 7944,
      "column": 37,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    }
  ]
}
````


**utilities.ts:7949:37** `getLinesBetweenPositions`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7948:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7949:37",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7949:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7948:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 313781,
      "body_end": 313860,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLinesBetweenRangeEndPositions",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7948:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts",
      "line": 7949,
      "column": 37,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    }
  ]
}
````


**utilities.ts:7959:37** `getLinesBetweenPositions`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7958:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7959:37",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7959:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7958:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 314169,
      "body_end": 314242,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "positionsAreOnSameLine",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7958:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts",
      "line": 7959,
      "column": 37,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    }
  ]
}
````


**utilities.ts:7971:37** `getLinesBetweenPositions`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7968:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7971:37",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7971:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7968:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 314729,
      "body_end": 315003,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLinesBetweenPositionAndPrecedingNonWhitespaceCharacter",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7968:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts",
      "line": 7971,
      "column": 37,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    }
  ]
}
````


**utilities.ts:7977:37** `getLinesBetweenPositions`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7975:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7977:37",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7977:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7975:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 315182,
      "body_end": 315372,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLinesBetweenPositionAndNextNonWhitespaceCharacter",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts:7975:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/utilities.ts",
      "line": 7977,
      "column": 37,
      "severity": "error",
      "rule": "adamic/no-optional-widening",
      "messageId": "optionalWidening",
      "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
    }
  ]
}
````


**watch.ts:249:60** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "adamic_checker_diagnostics": [],
  "enclosing_unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-c-public/src/compiler/watch.ts:233:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 8061,
      "body_end": 8940,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getFilesInErrorForSummary",
      "top": "/workspace/scratch/baobab-road2-c-public/src/compiler/watch.ts:233:1"
    }
  ],
  "cohere_findings_at_call": [
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/watch.ts",
      "line": 249,
      "column": 60,
      "severity": "error",
      "rule": "@typescript-eslint/no-non-null-assertion",
      "messageId": "noNonNull",
      "message": "A non-null assertion tells the compiler to stop checking, and it is erased before the code runs, so when the value is missing anyway the failure lands somewhere else, as a property access on `undefined` in code the types promised was safe. Fix it in this order. If the value can really be missing on a reachable path (input, disk, the network, an external API), guard it: return early, `continue`, or throw an error that says what was missing, because that is the case the type is warning about. If it cannot, fix the type that says it can (a type predicate in `.filter`, a tuple type, a field made required), so there is no claim left to check. If an index is what makes it optional, restructure so the index goes away (`for...of`, `.entries()`, a destructured head checked once), because a value you never index is never possibly missing. Where none of those fit, state the proof with `assert(value !== undefined, 'why it holds')` or, in an expression, `required(value, 'why it holds')` from `@nexus/source/errors/Assert`, which narrows the same way and throws that sentence if the proof is ever wrong. `?.`, `?? ''` and `as T` are not repairs on their own: each makes the missing value disappear instead of handling it."
    },
    {
      "kind": "finding",
      "path": "/workspace/scratch/baobab-road2-c-public/src/compiler/watch.ts",
      "line": 249,
      "column": 89,
      "severity": "error",
      "rule": "@typescript-eslint/no-non-null-assertion",
      "messageId": "noNonNull",
      "message": "A non-null assertion tells the compiler to stop checking, and it is erased before the code runs, so when the value is missing anyway the failure lands somewhere else, as a property access on `undefined` in code the types promised was safe. Fix it in this order. If the value can really be missing on a reachable path (input, disk, the network, an external API), guard it: return early, `continue`, or throw an error that says what was missing, because that is the case the type is warning about. If it cannot, fix the type that says it can (a type predicate in `.filter`, a tuple type, a field made required), so there is no claim left to check. If an index is what makes it optional, restructure so the index goes away (`for...of`, `.entries()`, a destructured head checked once), because a value you never index is never possibly missing. Where none of those fit, state the proof with `assert(value !== undefined, 'why it holds')` or, in an expression, `required(value, 'why it holds')` from `@nexus/source/errors/Assert`, which narrows the same way and throws that sentence if the proof is ever wrong. `?.`, `?? ''` and `as T` are not repairs on their own: each makes the missing value disappear instead of handling it."
    }
  ]
}
````


### Reproducibility, bounds, and limits

Each of the nine files with original findings, plus scanner.ts for the writer unit, is run in an independent process with LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, and LATENT_ONLY_FILE set to that exact file. All 79 compiler files remain roots. There are at most two concurrent file censuses; the public alternative starts only after all ten internal variant attempts finish. Every file has a 1500-second bound. Each child is waited for and its own exit code is recorded. Checkpoints are pushed after completed pieces.

| Variant | File | Child exit | Seconds | Complete full stream | Completed timing units |
| --- | --- | ---: | ---: | --- | ---: |
| (c), internal LineMapSource | scanner.ts | 0 | 613.42 | true | 150 |
| (c), internal LineMapSource | tracing.ts | 0 | 214.85 | true | 23 |
| (c), internal LineMapSource | watch.ts | 0 | 117.76 | true | 36 |
| (c), internal LineMapSource | transformers/jsx.ts | 0 | 151.39 | true | 40 |
| (c), internal LineMapSource | transformers/declarations.ts | 0 | 184.41 | true | 69 |
| (c), internal LineMapSource | executeCommandLine.ts | 0 | 191.80 | true | 68 |
| (c), internal LineMapSource | emitter.ts | 0 | 673.43 | true | 423 |
| (c), internal LineMapSource | utilities.ts | 0 | 746.02 | true | 867 |
| (c), internal LineMapSource | program.ts | 0 | 904.84 | true | 235 |
| (c), internal LineMapSource | parser.ts | 0 | 871.26 | true | 545 |
| (c), public LineMapSource | scanner.ts | 0 | 361.56 | true | 150 |
| (c), public LineMapSource | tracing.ts | 0 | 48.95 | true | 23 |
| (c), public LineMapSource | watch.ts | 0 | 78.66 | true | 36 |
| (c), public LineMapSource | transformers/jsx.ts | 0 | 121.01 | true | 40 |
| (c), public LineMapSource | transformers/declarations.ts | 0 | 153.92 | true | 69 |
| (c), public LineMapSource | executeCommandLine.ts | 0 | 155.27 | true | 68 |
| (c), public LineMapSource | emitter.ts | 0 | 612.61 | true | 423 |
| (c), public LineMapSource | utilities.ts | 0 | 712.32 | true | 867 |
| (c), public LineMapSource | program.ts | 0 | 839.37 | true | 235 |
| (c), public LineMapSource | parser.ts | 0 | 797.68 | true | 545 |

The overlay's historical measurement label remains constant even on a checker-accepted input. Every compiler census header for (c) records checker_rejected=false and diagnostics=[]. Use those actual flags rather than the legacy label. Completed streams still contain latent body boundaries and are not whole-program compilation proofs. No backend or JavaScript output is invoked. Full-src checker rejection, the unchanged writer NotYet, the unrelated core first stop, and the API guard failures above remain real limitations.

Evidence is under continuation/c/, c-public/, their declaration variants, and c-exact/, c-public-exact/, original-exact/, and original-rules/. c-site-comparison.json preserves all 68 view/variant rows. c-analysis.json preserves every comparison addition and removal. New Go evidence files end in .go.txt. Files over 1,000,000 bytes are gzip compressed. SHA256SUMS covers all retained evidence.

## Fourth measurement: internal LineMapSource overload, variant (e)

Measured from the same frozen adapted input on main d690f276749f91e6b36fe3a9726a5a453fcd7d39 and TypeScript 6.0.3 source 050880ce59e30b356b686bd3144efe24f875ebc8. The guarded binary and pinned cohere are unchanged. The preceding report commit is e87b5ec71092a42bf931b8f0f0a6b8840d3718b6. This is report only, with no compiler or rule changes landed.

### Exact source and lint

A fresh scratch copy inserts the requested @internal LineMapSource immediately before the whole documented SourceFileLike declaration. SourceFileLike, including its existing leading documentation, text member, lineMap member and their @internal annotations, remains byte-for-byte unchanged. It does not extend LineMapSource. The only other source file changed is scanner.ts: one required type-only import, the new overload set, and the first parameter type of the already-internal getLineStarts. Both function bodies are unchanged.

The complete overload set and unchanged body in the final scanner.ts are:

```ts
/** @internal */
export function getLineAndCharacterOfPosition(sourceFile: LineMapSource, position: number): LineAndCharacter;
export function getLineAndCharacterOfPosition(sourceFile: SourceFileLike, position: number): LineAndCharacter; // eslint-disable-line @typescript-eslint/unified-signatures
export function getLineAndCharacterOfPosition(sourceFile: LineMapSource, position: number): LineAndCharacter {
    return computeLineAndCharacterOfPosition(getLineStarts(sourceFile), position);
}
```

The internal overload is number 1 at scanner.ts:549:1. The public overload is number 2 at 550:1. The implementation is 551:1. getLineStarts remains at 503:1. The six new lines in types.ts shift its old declaration locations by six; the three new scanner lines shift old locations after the implementation header by three. Comparisons map those unchanged locations back to baseline, while raw streams retain actual coordinates.

The pinned full ESLint configuration requires the trailing unified-signatures suppression on the public overload. Its initial diagnostic was:

````json
[
  {
    "ruleId": "@typescript-eslint/unified-signatures",
    "severity": 2,
    "message": "These overloads can be combined into one signature taking `LineMapSource | SourceFileLike`.",
    "line": 550,
    "column": 47,
    "messageId": "singleParameterDifference",
    "endLine": 550,
    "endColumn": 73
  }
]
````


The exact // eslint-disable-line @typescript-eslint/unified-signatures comment is the one used by getPositionOfLineAndCharacter. No lint rule is globally disabled and no other diagnostic is repaired. The final scanner lint has 10 errors, matching the frozen baseline count of 10. These are existing no-unnecessary-type-assertion diagnostics; there is no remaining overload-order or unified-signatures diagnostic.

Two preliminary attempts are retained as clearly labeled controls, not counted as final (e) measurements. e-unsuppressed-control predates the required lint comment. e-displaced-doc-control exposed an insertion error: the new interface had captured SourceFileLike's existing documentation, causing a comment-only API diff. The final tree preserves that whole documented declaration. Native children in abandoned attempts were terminated and reaped by their own bounded wrappers before source changes. A transient HTTP 503 interrupted one checkpoint push; ordinary non-force push retries were added. All final results below come from the corrected, fixed-byte tree.

### All 34 original views

| Complete file streams | Cleared at original view | Unchanged finding | Replacement finding | Checker diagnostic | Incomplete or excluded |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 10/10 | 22 | 12 | 0 | 0 | 0 |

Cleared means that the original argument/call span has no guarded finding or checker diagnostic and its enclosing unit was attempted. It does not mean that the entire root lowers. The 22 helper views contain 21 SourceFile values and one SourceMapSource value. The 12 unchanged views still target the unmodified SourceFileLike callees. The earlier checker.ts 25-minute cutoff remains a limitation of the original site census; this comparison covers all 34 original observed views.

| Callee | Original sites | Cleared | Unchanged | Replacement / checker / unknown |
| --- | ---: | ---: | ---: | ---: |
| `getEndLinePosition` | 1 | 0 | 1 | 0 |
| `getLineAndCharacterOfPosition` | 18 | 18 | 0 | 0 |
| `getLineStarts` | 4 | 4 | 0 | 0 |
| `getLinesBetweenPositions` | 5 | 0 | 5 | 0 |
| `getPositionOfLineAndCharacter` | 2 | 0 | 2 | 0 |
| `transferSourceFileChildren` | 2 | 0 | 2 | 0 |
| `unsetNodeChildren` | 2 | 0 | 2 | 0 |

### Which overload the 18 original internal calls resolve to

This is measured twice. The stock TypeScript 6.0.3 checker resolves calls in the same whole-src 604-root program used for the previous caller audit. Independently, e-overloads.go.txt calls production load.Load on all 79 compiler roots and production Checker.GetResolvedSignature for each call. It does not infer a choice from overload order, and it does not lower or emit anything. The native checker sees 20 compiler calls, including the two checker.ts calls outside the original 34-view list. All returned choices are retained in e/native-overload-resolution.json. The requested 18-call subset follows.

| Original argument location | Enclosing function | Argument | Declared type | Native flow type | Stock overload | Native overload | Native chosen declaration |
| --- | --- | --- | --- | --- | ---: | ---: | --- |
| emitter.ts:3961:112 | `emitJsxExpression` | `currentSourceFile` | `ts.SourceFile \| undefined` | `SourceFile` | 1 | 1 | `scanner.ts:549:1` |
| emitter.ts:3961:180 | `emitJsxExpression` | `currentSourceFile` | `ts.SourceFile \| undefined` | `SourceFile` | 1 | 1 | `scanner.ts:549:1` |
| emitter.ts:6233:96 | `emitPos` | `sourceMapSource` | `ts.SourceMapSource` | `SourceMapSource` | 1 | 1 | `scanner.ts:549:1` |
| program.ts:671:67 | `formatDiagnostic` | `diagnostic.file` | `SourceFile \| undefined` | `SourceFile` | 1 | 1 | `scanner.ts:549:1` |
| program.ts:713:89 | `formatCodeSpan` | `file` | `SourceFile` | `SourceFile` | 1 | 1 | `scanner.ts:549:1` |
| program.ts:714:87 | `formatCodeSpan` | `file` | `SourceFile` | `SourceFile` | 1 | 1 | `scanner.ts:549:1` |
| program.ts:715:58 | `formatCodeSpan` | `file` | `SourceFile` | `SourceFile` | 1 | 1 | `scanner.ts:549:1` |
| program.ts:768:89 | `formatLocation` | `file` | `SourceFile` | `SourceFile` | 1 | 1 | `scanner.ts:549:1` |
| tracing.ts:206:67 | `getLocation` | `file` | `SourceFile \| undefined` | `SourceFile` | 1 | 1 | `scanner.ts:549:1` |
| tracing.ts:207:65 | `getLocation` | `file` | `SourceFile \| undefined` | `SourceFile` | 1 | 1 | `scanner.ts:549:1` |
| transformers/declarations.ts:1264:139 | `visitDeclarationSubtree` | `currentSourceFile` | `SourceFile` | `SourceFile` | 1 | 1 | `scanner.ts:549:1` |
| transformers/declarations.ts:1264:70 | `visitDeclarationSubtree` | `currentSourceFile` | `SourceFile` | `SourceFile` | 1 | 1 | `scanner.ts:549:1` |
| transformers/jsx.ts:359:63 | `visitJsxOpeningLikeElementOrFragmentJSX` | `originalFile` | `Node` | `SourceFile` | 1 | 1 | `scanner.ts:549:1` |
| utilities.ts:1025:47 | `nodePosToString` | `file` | `SourceFile` | `SourceFile` | 1 | 1 | `scanner.ts:549:1` |
| utilities.ts:1213:46 | `<anonymous@1212>` | `sourceFile` | `SourceFile` | `SourceFile` | 1 | 1 | `scanner.ts:549:1` |
| utilities.ts:2526:67 | `getErrorSpanForArrowFunction` | `sourceFile` | `SourceFile` | `SourceFile` | 1 | 1 | `scanner.ts:549:1` |
| utilities.ts:2527:65 | `getErrorSpanForArrowFunction` | `sourceFile` | `SourceFile` | `SourceFile` | 1 | 1 | `scanner.ts:549:1` |
| watch.ts:249:60 | `<anonymous@241>` | `diagnosticForFileName.file!` | `SourceFile \| undefined` | `SourceFile` | 1 | 1 | `scanner.ts:549:1` |

Both checkers resolve all 18 to overload 1, whose first parameter is LineMapSource. The public SourceFileLike overload is not selected at these sites. e/calls.json also retains all 93 source calls, both available signatures, chosen declaration text, the internal marker, and argument declared/flow types. All 79 getLineAndCharacterOfPosition calls in that full source program resolve to the internal first overload. The separate SDK probe instead reads only the emitted public declaration and sees the original SourceFileLike signature.

### Scanner own stream and complete overload handling

The following is the complete set of unit records and findings for getLineAndCharacterOfPosition in the final full-mode scanner stream, including both bodyless overload declarations and the implementation. No finding is summarized away:

````json
{
  "units": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/scanner.ts:549:1",
      "kind": "KindFunctionDeclaration",
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLineAndCharacterOfPosition",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/scanner.ts:549:1"
    },
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/scanner.ts:550:1",
      "kind": "KindFunctionDeclaration",
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLineAndCharacterOfPosition",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/scanner.ts:550:1"
    },
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/scanner.ts:551:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 51383,
      "body_end": 51472,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLineAndCharacterOfPosition",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/scanner.ts:551:1"
    }
  ],
  "findings": []
}
````


Every new overload-related finding anywhere across the ten selected files is:

````json
[]
````


In particular, new getLineAndCharacterOfPosition "overload N ... cannot be served by implementation" findings are:

````json
[]
````


The signature path checks censusOverloads before the body (functions.go:68), and module registration checks it too (modules.go:195). censusOverload compares each promised argument to its implementation parameter and the implementation result to each promised result. The full corresponding source is retained as e-census-small.go.txt, e-functions.go.txt and e-modules.go.txt. These are the unchanged pinned rules, not a new rule or exception. All three getLineAndCharacterOfPosition units are attempted, with no signature failure or checker rejection.

getLineStarts own complete unit records and findings are:

````json
{
  "units": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/scanner.ts:503:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 49075,
      "body_end": 49173,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLineStarts",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/scanner.ts:503:1"
    }
  ],
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-e/src/compiler/scanner.ts:503:1",
      "phase": "lowering",
      "kind": "NotYet",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/scanner.ts:504:35",
      "reason": "an assignment value to a member",
      "text": "/workspace/scratch/baobab-road2-e/src/compiler/scanner.ts:504:35: stage 0 can't lower an assignment value to a member yet"
    },
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-e/src/compiler/scanner.ts:503:1",
      "phase": "lowering",
      "kind": "Boundary",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/scanner.ts:504:5",
      "reason": "failed statement; state rolled back; continued at next statement or declaration",
      "text": "/workspace/scratch/baobab-road2-e/src/compiler/scanner.ts:504:35: stage 0 can't lower an assignment value to a member yet",
      "start": 49077,
      "end": 49170
    }
  ]
}
````


Its assignment-value NotYet at scanner.ts:504:35 is the same pre-existing baseline stop. The new overloads do not repair that expression, and it is not the core.ts first stop. The complete scanner stream is retained under e/raw/scanner.jsonl or scanner.jsonl.gz; all scanner overload findings, including unrelated existing ones, are also retained in e/scanner-overload-handling.json.

### All cohere and lowering additions/removals

The same pinned cohere binary runs the complete cohere:adamic set under the actual Adamic checker settings. Its successful run checks 77 compiler files with 40 rules, preserving all findings and unresolved-input warnings. The baseline comparison includes every finding in the same compiler scope, not just the two helpers. The guarded lowering comparison includes every finding, phase, reason and kind across all ten selected-file streams.

| Measurement | Added | Removed |
| --- | ---: | ---: |
| Cohere findings | 0 | 24 |
| Guarded full-census findings | 0 | 22 |

Per-phase guarded differences:

````json
{
  "deferred": {
    "added": 0,
    "removed": 0
  },
  "lowering": {
    "added": 0,
    "removed": 0
  },
  "refusal_scan": {
    "added": 0,
    "removed": 22
  }
}
````


Every added cohere finding:

````json
[]
````


Every removed cohere finding:

````json
[
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/emitter.ts",
    "line": 3961,
    "column": 112,
    "severity": "error",
    "message": "`currentSourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/transformers/declarations.ts",
    "line": 1264,
    "column": 139,
    "severity": "error",
    "message": "`currentSourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/emitter.ts",
    "line": 3961,
    "column": 180,
    "severity": "error",
    "message": "`currentSourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/utilities.ts",
    "line": 6748,
    "column": 38,
    "severity": "error",
    "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/executeCommandLine.ts",
    "line": 109,
    "column": 41,
    "severity": "error",
    "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/program.ts",
    "line": 2973,
    "column": 42,
    "severity": "error",
    "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/utilities.ts",
    "line": 1213,
    "column": 46,
    "severity": "error",
    "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/utilities.ts",
    "line": 1025,
    "column": 47,
    "severity": "error",
    "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/checker.ts",
    "line": 52425,
    "column": 55,
    "severity": "error",
    "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/checker.ts",
    "line": 52424,
    "column": 57,
    "severity": "error",
    "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/program.ts",
    "line": 715,
    "column": 58,
    "severity": "error",
    "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/watch.ts",
    "line": 249,
    "column": 60,
    "severity": "error",
    "message": "`diagnosticForFileName.file!` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/transformers/jsx.ts",
    "line": 359,
    "column": 63,
    "severity": "error",
    "message": "`originalFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/tracing.ts",
    "line": 207,
    "column": 65,
    "severity": "error",
    "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/utilities.ts",
    "line": 2527,
    "column": 65,
    "severity": "error",
    "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/emitter.ts",
    "line": 1444,
    "column": 66,
    "severity": "error",
    "message": "`Debug.checkDefined(currentSourceFile)` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/tracing.ts",
    "line": 206,
    "column": 67,
    "severity": "error",
    "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/utilities.ts",
    "line": 2526,
    "column": 67,
    "severity": "error",
    "message": "`sourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/program.ts",
    "line": 671,
    "column": 67,
    "severity": "error",
    "message": "`diagnostic.file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/transformers/declarations.ts",
    "line": 1264,
    "column": 70,
    "severity": "error",
    "message": "`currentSourceFile` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/program.ts",
    "line": 714,
    "column": 87,
    "severity": "error",
    "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/program.ts",
    "line": 713,
    "column": 89,
    "severity": "error",
    "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/program.ts",
    "line": 768,
    "column": 89,
    "severity": "error",
    "message": "`file` is a 'SourceFile', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceFile' does not declare. A 'SourceFile' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  },
  {
    "rule": "adamic/no-optional-widening",
    "path": "<tree>/src/compiler/emitter.ts",
    "line": 6233,
    "column": 96,
    "severity": "error",
    "message": "`sourceMapSource` is a 'SourceMapSource', seen here as 'SourceFileLike', which adds the optional property `getPositionOfLineAndCharacter` that 'SourceMapSource' does not declare. A 'SourceMapSource' can still carry a `getPositionOfLineAndCharacter` of any type, since an earlier, wider value may have been seen as it, and tsc now gives that property the optional one's type: the value reads as a '(line: number, character: number, allowEdits?: true | undefined) => number | undefined' and holds something else. Declare `getPositionOfLineAndCharacter` in the source's type, or build the value with the property named."
  }
]
````


Every added guarded finding:

````json
[]
````


Every removed guarded finding:

````json
[
  {
    "unit": "<tree>/src/compiler/emitter.ts:1443:5",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/emitter.ts:1444:66",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/emitter.ts:1444:66: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "emitter.ts"
  },
  {
    "unit": "<tree>/src/compiler/emitter.ts:3959:5",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/emitter.ts:3961:112",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/emitter.ts:3961:112: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "emitter.ts"
  },
  {
    "unit": "<tree>/src/compiler/emitter.ts:3959:5",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/emitter.ts:3961:180",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/emitter.ts:3961:180: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "emitter.ts"
  },
  {
    "unit": "<tree>/src/compiler/executeCommandLine.ts:105:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/executeCommandLine.ts:109:41",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/executeCommandLine.ts:109:41: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "executeCommandLine.ts"
  },
  {
    "unit": "<tree>/src/compiler/program.ts:2966:5",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/program.ts:2973:42",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/program.ts:2973:42: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "program.ts"
  },
  {
    "unit": "<tree>/src/compiler/program.ts:667:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/program.ts:671:67",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/program.ts:671:67: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "program.ts"
  },
  {
    "unit": "<tree>/src/compiler/program.ts:712:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/program.ts:713:89",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/program.ts:713:89: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "program.ts"
  },
  {
    "unit": "<tree>/src/compiler/program.ts:712:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/program.ts:714:87",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/program.ts:714:87: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "program.ts"
  },
  {
    "unit": "<tree>/src/compiler/program.ts:712:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/program.ts:715:58",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/program.ts:715:58: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "program.ts"
  },
  {
    "unit": "<tree>/src/compiler/program.ts:767:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/program.ts:768:89",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/program.ts:768:89: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "program.ts"
  },
  {
    "unit": "<tree>/src/compiler/tracing.ts:200:5",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/tracing.ts:206:67",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/tracing.ts:206:67: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "tracing.ts"
  },
  {
    "unit": "<tree>/src/compiler/tracing.ts:200:5",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/tracing.ts:207:65",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/tracing.ts:207:65: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "tracing.ts"
  },
  {
    "unit": "<tree>/src/compiler/transformers/declarations.ts:1006:5",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/transformers/declarations.ts:1264:139",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/transformers/declarations.ts:1264:139: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "transformers/declarations.ts"
  },
  {
    "unit": "<tree>/src/compiler/transformers/declarations.ts:1006:5",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/transformers/declarations.ts:1264:70",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/transformers/declarations.ts:1264:70: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "transformers/declarations.ts"
  },
  {
    "unit": "<tree>/src/compiler/transformers/jsx.ts:333:5",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/transformers/jsx.ts:359:63",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/transformers/jsx.ts:359:63: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "transformers/jsx.ts"
  },
  {
    "unit": "<tree>/src/compiler/utilities.ts:1023:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/utilities.ts:1025:47",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/utilities.ts:1025:47: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "utilities.ts"
  },
  {
    "unit": "<tree>/src/compiler/utilities.ts:1210:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/utilities.ts:1213:46",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/utilities.ts:1213:46: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "utilities.ts"
  },
  {
    "unit": "<tree>/src/compiler/utilities.ts:2523:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/utilities.ts:2526:67",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/utilities.ts:2526:67: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "utilities.ts"
  },
  {
    "unit": "<tree>/src/compiler/utilities.ts:2523:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/utilities.ts:2527:65",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/utilities.ts:2527:65: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "utilities.ts"
  },
  {
    "unit": "<tree>/src/compiler/utilities.ts:6747:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/utilities.ts:6748:38",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/utilities.ts:6748:38: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "utilities.ts"
  },
  {
    "unit": "<tree>/src/compiler/watch.ts:233:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/watch.ts:249:60",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
    "text": "<tree>/src/compiler/watch.ts:249:60: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "watch.ts"
  },
  {
    "unit": "<tree>/src/compiler/emitter.ts:6228:5",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "<tree>/src/compiler/emitter.ts:6233:96",
    "site_where": null,
    "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceMapSource, which can hide fields",
    "text": "<tree>/src/compiler/emitter.ts:6233:96: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceMapSource, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)",
    "root": "emitter.ts"
  }
]
````


Cohere removes the 22 helper-site optional-method findings plus the two checker.ts helper-call findings outside the original 34-site lowering list. The unchanged SourceFileLike contracts still account for the other 12 original views. Scratch roots and unchanged source-line shifts are normalized solely for comparison; raw streams preserve their original paths and locations.

### Whole src, services, and core.ts

| Input | Whole-src stock errors | Stock exact optional errors | Production Adamic whole-src errors | Services stock errors | Services exact optional errors | Services TS2386, stock / exact |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| Frozen input | 0 | 416 | 6851 | 0 | 142 | 0 / 0 |
| (e) | 0 | 416 | 6851 | 0 | 142 | 0 / 0 |

All whole-src runs use the same 604 roots from 13 solution projects. Stock defaults are preserved; the extra stock run enables exactOptionalPropertyTypes. The production run uses unmodified load.Load and every Adamic option, including exactOptionalPropertyTypes and noUncheckedIndexedAccess. Its baseline errors, including type-only imports and unavailable test globals, remain real errors; they are not attributed to (e). The production and exact-optional diagnostic additions/removals are:

````json
{
  "production": {
    "baseline_unique": 6840,
    "variant_unique": 6840,
    "added": [],
    "removed": []
  },
  "stock_exact_optional": {
    "baseline_unique": 416,
    "variant_unique": 416,
    "added": [],
    "removed": []
  }
}
````


The actual core.ts first-stop record is:

````json
{
  "roots": [
    "/workspace/scratch/baobab-road2-e/src/compiler/core.ts"
  ],
  "kind": "Refused",
  "where": "/workspace/scratch/baobab-road2-e/src/compiler/tracing.ts:232:34",
  "reason": "a cast the runtime can't check",
  "fix": "use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)",
  "seconds": 16.533875995
}
````


The compiler-only core root has no checker diagnostics. Its first stop is the unrelated tracing.ts:232:34 cast, not the cache write and not an overload-serving refusal. The driver repeats the same one-root input as a whole-program record; the first record is quoted.

### stripInternal-on declarations, bundle, and lane API guard

Stock TypeScript performs actual emitDeclarationOnly with stripInternal=true, then the unchanged upstream declaration bundler runs on the emitted module entry. The emit options, emitted-file hashes, bundler log and public bundle are retained. The internal LineMapSource and internal first overload are stripped, while the original public SourceFileLike overload remains. getLineStarts stays internal and absent.

| Public result | Measured value |
| --- | --- |
| Emitted declaration files | 278 |
| Emit diagnostics | 0 |
| Normalized public API changes | 0 |
| Public closure/consumer diagnostics | 0 |
| Lane API guard errors / unsanctioned changes | 0 |

The complete selected emitted public declarations are:

````json
[
  {
    "name": "SourceFileLike",
    "kind": "InterfaceDeclaration",
    "text": "interface SourceFileLike {\n        readonly text: string;\n    }"
  },
  {
    "name": "SourceFileLike",
    "kind": "InterfaceDeclaration",
    "text": "interface SourceFileLike {\n        getLineAndCharacterOfPosition(pos: number): LineAndCharacter;\n    }"
  },
  {
    "name": "getLineAndCharacterOfPosition",
    "kind": "FunctionDeclaration",
    "text": "function getLineAndCharacterOfPosition(sourceFile: SourceFileLike, position: number): LineAndCharacter;"
  }
]
````


The textual typescript.d.ts diff is retained in e/typescript-public.diff, including the diff child exit code. Its public bytes and normalized declaration map are compared against the frozen input, not a hand-edited expected snapshot. The actual stage3/lane check.api_check result is:

````json
{
  "errors": [],
  "facts": {
    "reference_declarations": 0,
    "composed_declarations": 267,
    "sanctioned_declarations": 267,
    "added_lines": 349,
    "removed_lines": 267
  }
}
````


No API sanction is added or changed. SourceFileLike retains its original documentation and public members. The internal implementation type is not a dangling name in this public d.ts.

### SDK probe

The same emitted-public-API consumer used for (c) is rechecked with the production checker, cohere adamic rules and the same guarded full-mode lowering binary. It imports the real emitted public bundle and accepts a ts.SourceFile parameter. All SDK findings, including the baseline external function-without-a-body NotYet, are quoted in full:

````json
{
  "checker": {
    "accepted": true,
    "diagnostics": [],
    "driver": "production load.Load, no lowering or output IR",
    "roots": 1,
    "seconds": 0.300982083
  },
  "cohere_findings": [],
  "guarded_findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-sdk-e/consumer.ts:2:1",
      "phase": "lowering",
      "kind": "NotYet",
      "where": "/workspace/scratch/baobab-sdk-e/typescript.d.ts:8557:5",
      "reason": "a function without a body",
      "text": "/workspace/scratch/baobab-sdk-e/typescript.d.ts:8557:5: stage 0 can't lower a function without a body yet"
    },
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-sdk-e/consumer.ts:2:1",
      "phase": "lowering",
      "kind": "Boundary",
      "where": "/workspace/scratch/baobab-sdk-e/consumer.ts:3:5",
      "reason": "failed statement; state rolled back; continued at next statement or declaration",
      "text": "/workspace/scratch/baobab-sdk-e/typescript.d.ts:8557:5: stage 0 can't lower a function without a body yet",
      "start": 104,
      "end": 158
    }
  ],
  "guarded_delta": {
    "baseline_unique": 2,
    "variant_unique": 2,
    "added": [],
    "removed": []
  }
}
````


A zero new-refusal count is relative to the original SDK probe. It is not a claim that an implementation-free declaration bundle can be compiled into native runtime code. The public signature stays SourceFileLike, so the LineMapSource optional-cache hiding refusal introduced by public (c) does not recur.

### Every original view, with literal findings

| Original location | Phase | Enclosing function | Callee | Classification |
| --- | --- | --- | --- | --- |
| emitter.ts:1444:66 | refusal_scan | `getCurrentLineMap` | `getLineStarts` | cleared at original view |
| emitter.ts:3961:112 | refusal_scan | `emitJsxExpression` | `getLineAndCharacterOfPosition` | cleared at original view |
| emitter.ts:3961:180 | refusal_scan | `emitJsxExpression` | `getLineAndCharacterOfPosition` | cleared at original view |
| emitter.ts:6233:96 | refusal_scan | `emitPos` | `getLineAndCharacterOfPosition` | cleared at original view |
| executeCommandLine.ts:109:41 | refusal_scan | `<anonymous@107>` | `getLineStarts` | cleared at original view |
| parser.ts:10032:36 | refusal_scan | `updateSourceFile` | `transferSourceFileChildren` | unchanged finding |
| parser.ts:10032:48 | refusal_scan | `updateSourceFile` | `transferSourceFileChildren` | unchanged finding |
| parser.ts:10104:37 | refusal_scan | `visitNode` | `unsetNodeChildren` | unchanged finding |
| parser.ts:10261:42 | refusal_scan | `visitNode` | `unsetNodeChildren` | unchanged finding |
| program.ts:2973:42 | refusal_scan | `markPrecedingCommentDirectiveLine` | `getLineStarts` | cleared at original view |
| program.ts:671:67 | refusal_scan | `formatDiagnostic` | `getLineAndCharacterOfPosition` | cleared at original view |
| program.ts:713:89 | refusal_scan | `formatCodeSpan` | `getLineAndCharacterOfPosition` | cleared at original view |
| program.ts:714:87 | refusal_scan | `formatCodeSpan` | `getLineAndCharacterOfPosition` | cleared at original view |
| program.ts:715:58 | refusal_scan | `formatCodeSpan` | `getLineAndCharacterOfPosition` | cleared at original view |
| program.ts:733:57 | refusal_scan | `formatCodeSpan` | `getPositionOfLineAndCharacter` | unchanged finding |
| program.ts:734:76 | refusal_scan | `formatCodeSpan` | `getPositionOfLineAndCharacter` | unchanged finding |
| program.ts:768:89 | refusal_scan | `formatLocation` | `getLineAndCharacterOfPosition` | cleared at original view |
| tracing.ts:206:67 | refusal_scan | `getLocation` | `getLineAndCharacterOfPosition` | cleared at original view |
| tracing.ts:207:65 | refusal_scan | `getLocation` | `getLineAndCharacterOfPosition` | cleared at original view |
| transformers/declarations.ts:1264:139 | refusal_scan | `visitDeclarationSubtree` | `getLineAndCharacterOfPosition` | cleared at original view |
| transformers/declarations.ts:1264:70 | refusal_scan | `visitDeclarationSubtree` | `getLineAndCharacterOfPosition` | cleared at original view |
| transformers/jsx.ts:359:63 | refusal_scan | `visitJsxOpeningLikeElementOrFragmentJSX` | `getLineAndCharacterOfPosition` | cleared at original view |
| utilities.ts:1025:47 | refusal_scan | `nodePosToString` | `getLineAndCharacterOfPosition` | cleared at original view |
| utilities.ts:1213:46 | refusal_scan | `<anonymous@1212>` | `getLineAndCharacterOfPosition` | cleared at original view |
| utilities.ts:2526:67 | refusal_scan | `getErrorSpanForArrowFunction` | `getLineAndCharacterOfPosition` | cleared at original view |
| utilities.ts:2527:65 | refusal_scan | `getErrorSpanForArrowFunction` | `getLineAndCharacterOfPosition` | cleared at original view |
| utilities.ts:2531:70 | refusal_scan | `getErrorSpanForArrowFunction` | `getEndLinePosition` | unchanged finding |
| utilities.ts:6748:38 | refusal_scan | `getLineOfLocalPosition` | `getLineStarts` | cleared at original view |
| utilities.ts:7944:37 | refusal_scan | `getLinesBetweenRangeEndAndRangeStart` | `getLinesBetweenPositions` | unchanged finding |
| utilities.ts:7949:37 | refusal_scan | `getLinesBetweenRangeEndPositions` | `getLinesBetweenPositions` | unchanged finding |
| utilities.ts:7959:37 | refusal_scan | `positionsAreOnSameLine` | `getLinesBetweenPositions` | unchanged finding |
| utilities.ts:7971:37 | refusal_scan | `getLinesBetweenPositionAndPrecedingNonWhitespaceCharacter` | `getLinesBetweenPositions` | unchanged finding |
| utilities.ts:7977:37 | refusal_scan | `getLinesBetweenPositionAndNextNonWhitespaceCharacter` | `getLinesBetweenPositions` | unchanged finding |
| watch.ts:249:60 | refusal_scan | `<anonymous@241>` | `getLineAndCharacterOfPosition` | cleared at original view |

**emitter.ts:1444:66** `getLineStarts`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/emitter.ts:1443:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 58788,
      "body_end": 58904,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getCurrentLineMap",
      "parent": "/workspace/scratch/baobab-road2-e/src/compiler/emitter.ts:1211:1",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/emitter.ts:1211:1",
      "depth": 1
    }
  ]
}
````


**emitter.ts:3961:112** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/emitter.ts:3959:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 170721,
      "body_end": 171620,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "emitJsxExpression",
      "parent": "/workspace/scratch/baobab-road2-e/src/compiler/emitter.ts:1211:1",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/emitter.ts:1211:1",
      "depth": 1
    }
  ]
}
````


**emitter.ts:3961:180** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/emitter.ts:3959:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 170721,
      "body_end": 171620,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "emitJsxExpression",
      "parent": "/workspace/scratch/baobab-road2-e/src/compiler/emitter.ts:1211:1",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/emitter.ts:1211:1",
      "depth": 1
    }
  ]
}
````


**emitter.ts:6233:96** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/emitter.ts:6228:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 268634,
      "body_end": 269150,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "emitPos",
      "parent": "/workspace/scratch/baobab-road2-e/src/compiler/emitter.ts:1211:1",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/emitter.ts:1211:1",
      "depth": 1
    }
  ]
}
````


**executeCommandLine.ts:109:41** `getLineStarts`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/executeCommandLine.ts:105:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 2705,
      "body_end": 2984,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "countLines",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/executeCommandLine.ts:105:1"
    }
  ]
}
````


**parser.ts:10032:36** `transferSourceFileChildren`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:9947:5",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:10032:36",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:10032:36: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:9947:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 497989,
      "body_end": 503771,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "updateSourceFile",
      "parent": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:9946:1",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:9946:1"
    }
  ]
}
````


**parser.ts:10032:48** `transferSourceFileChildren`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:9947:5",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:10032:48",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:10032:48: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:9947:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 497989,
      "body_end": 503771,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "updateSourceFile",
      "parent": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:9946:1",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:9946:1"
    }
  ]
}
````


**parser.ts:10104:37** `unsetNodeChildren`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:10096:9",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:10104:37",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:10104:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:10096:9",
      "kind": "KindFunctionDeclaration",
      "body_start": 506691,
      "body_end": 507655,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "visitNode",
      "parent": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:10087:5",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:9946:1",
      "depth": 1
    }
  ]
}
````


**parser.ts:10261:42** `unsetNodeChildren`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:10246:9",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:10261:42",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:10261:42: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:10246:9",
      "kind": "KindFunctionDeclaration",
      "body_start": 513350,
      "body_end": 515010,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "visitNode",
      "parent": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:10233:5",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/parser.ts:9946:1",
      "depth": 1
    }
  ]
}
````


**program.ts:2973:42** `getLineStarts`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:2966:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 144938,
      "body_end": 145801,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "markPrecedingCommentDirectiveLine",
      "parent": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:1516:1",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:1516:1",
      "depth": 1
    }
  ]
}
````


**program.ts:671:67** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:667:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 24128,
      "body_end": 24786,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatDiagnostic",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:667:1"
    }
  ]
}
````


**program.ts:713:89** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:712:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 26062,
      "body_end": 28811,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatCodeSpan",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:712:1"
    }
  ]
}
````


**program.ts:714:87** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:712:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 26062,
      "body_end": 28811,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatCodeSpan",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:712:1"
    }
  ]
}
````


**program.ts:715:58** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:712:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 26062,
      "body_end": 28811,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatCodeSpan",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:712:1"
    }
  ]
}
````


**program.ts:733:57** `getPositionOfLineAndCharacter`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:712:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:733:57",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:733:57: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:712:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 26062,
      "body_end": 28811,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatCodeSpan",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:712:1"
    }
  ]
}
````


**program.ts:734:76** `getPositionOfLineAndCharacter`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:712:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:734:76",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:734:76: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:712:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 26062,
      "body_end": 28811,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatCodeSpan",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:712:1"
    }
  ]
}
````


**program.ts:768:89** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:767:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 28990,
      "body_end": 29611,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "formatLocation",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/program.ts:767:1"
    }
  ]
}
````


**tracing.ts:206:67** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/tracing.ts:200:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 7955,
      "body_end": 8503,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLocation",
      "parent": "/workspace/scratch/baobab-road2-e/src/compiler/tracing.ts:38:1",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/tracing.ts:38:1"
    }
  ]
}
````


**tracing.ts:207:65** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/tracing.ts:200:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 7955,
      "body_end": 8503,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLocation",
      "parent": "/workspace/scratch/baobab-road2-e/src/compiler/tracing.ts:38:1",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/tracing.ts:38:1"
    }
  ]
}
````


**transformers/declarations.ts:1264:139** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/transformers/declarations.ts:1006:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 47618,
      "body_end": 63600,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "visitDeclarationSubtree",
      "parent": "/workspace/scratch/baobab-road2-e/src/compiler/transformers/declarations.ts:259:1",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/transformers/declarations.ts:259:1",
      "depth": 1
    }
  ]
}
````


**transformers/declarations.ts:1264:70** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/transformers/declarations.ts:1006:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 47618,
      "body_end": 63600,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "visitDeclarationSubtree",
      "parent": "/workspace/scratch/baobab-road2-e/src/compiler/transformers/declarations.ts:259:1",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/transformers/declarations.ts:259:1",
      "depth": 1
    }
  ]
}
````


**transformers/jsx.ts:359:63** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/transformers/jsx.ts:333:5",
      "kind": "KindFunctionDeclaration",
      "body_start": 14898,
      "body_end": 16980,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "visitJsxOpeningLikeElementOrFragmentJSX",
      "parent": "/workspace/scratch/baobab-road2-e/src/compiler/transformers/jsx.ts:92:1",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/transformers/jsx.ts:92:1",
      "depth": 1
    }
  ]
}
````


**utilities.ts:1025:47** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:1023:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 33483,
      "body_end": 33667,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "nodePosToString",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:1023:1"
    }
  ]
}
````


**utilities.ts:1213:46** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:1210:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 42294,
      "body_end": 43143,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "createCommentDirectivesMap",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:1210:1"
    }
  ]
}
````


**utilities.ts:2526:67** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:2523:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 87836,
      "body_end": 88470,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getErrorSpanForArrowFunction",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:2523:1"
    }
  ]
}
````


**utilities.ts:2527:65** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:2523:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 87836,
      "body_end": 88470,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getErrorSpanForArrowFunction",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:2523:1"
    }
  ]
}
````


**utilities.ts:2531:70** `getEndLinePosition`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:2523:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:2531:70",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:2531:70: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:2523:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 87836,
      "body_end": 88470,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getErrorSpanForArrowFunction",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:2523:1"
    }
  ]
}
````


**utilities.ts:6748:38** `getLineStarts`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:6747:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 266234,
      "body_end": 266342,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLineOfLocalPosition",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:6747:1"
    }
  ]
}
````


**utilities.ts:7944:37** `getLinesBetweenPositions`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7942:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7944:37",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7944:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7942:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 313431,
      "body_end": 313609,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLinesBetweenRangeEndAndRangeStart",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7942:1"
    }
  ]
}
````


**utilities.ts:7949:37** `getLinesBetweenPositions`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7948:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7949:37",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7949:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7948:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 313781,
      "body_end": 313860,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLinesBetweenRangeEndPositions",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7948:1"
    }
  ]
}
````


**utilities.ts:7959:37** `getLinesBetweenPositions`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7958:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7959:37",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7959:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7958:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 314169,
      "body_end": 314242,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "positionsAreOnSameLine",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7958:1"
    }
  ]
}
````


**utilities.ts:7971:37** `getLinesBetweenPositions`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7968:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7971:37",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7971:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7968:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 314729,
      "body_end": 315003,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLinesBetweenPositionAndPrecedingNonWhitespaceCharacter",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7968:1"
    }
  ]
}
````


**utilities.ts:7977:37** `getLinesBetweenPositions`, unchanged finding.

````json
{
  "findings": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7975:1",
      "phase": "refusal_scan",
      "kind": "Refused",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7977:37",
      "reason": "optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields",
      "text": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7977:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
    }
  ],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7975:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 315182,
      "body_end": 315372,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getLinesBetweenPositionAndNextNonWhitespaceCharacter",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/utilities.ts:7975:1"
    }
  ]
}
````


**watch.ts:249:60** `getLineAndCharacterOfPosition`, cleared at original view.

````json
{
  "findings": [],
  "checker_diagnostics": [],
  "stock_diagnostics": [],
  "unit_statuses": [
    {
      "measurement": "measured on a checker-rejected program",
      "where": "/workspace/scratch/baobab-road2-e/src/compiler/watch.ts:233:1",
      "kind": "KindFunctionDeclaration",
      "body_start": 8061,
      "body_end": 8940,
      "instantiated": false,
      "status": "attempted",
      "checker_diagnostics": [],
      "name": "getFilesInErrorForSummary",
      "top": "/workspace/scratch/baobab-road2-e/src/compiler/watch.ts:233:1"
    }
  ]
}
````


### Full-mode bounds and retained evidence

Each final invocation retains all 79 compiler roots and uses LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1 and LATENT_ONLY_FILE equal to the exact selected file. The nine files with original findings plus scanner.ts are measured independently, with at most two concurrent file censuses and a 1500-second bound per file. Every native child is waited for, terminated as a process group at its bound if needed, and reports its own exit. Checkpoints are pushed after pieces to the checkpoint branch.

| File | Child exit | Seconds | Full stream completed | Completed timing units |
| --- | ---: | ---: | --- | ---: |
| scanner.ts | 0 | 297.22 | true | 152 |
| tracing.ts | 0 | 40.73 | true | 23 |
| watch.ts | 0 | 59.77 | true | 36 |
| transformers/jsx.ts | 0 | 110.21 | true | 40 |
| transformers/declarations.ts | 0 | 117.43 | true | 69 |
| executeCommandLine.ts | 0 | 133.51 | true | 68 |
| emitter.ts | 0 | 488.36 | true | 423 |
| utilities.ts | 0 | 598.82 | true | 867 |
| program.ts | 0 | 794.22 | true | 235 |
| parser.ts | 0 | 748.93 | true | 545 |

The legacy measurement label is fixed text; actual header checker_rejected flags and diagnostics determine checker status. All final (e) compiler headers report checker_rejected=false. Stream completion does not erase latent body boundaries or the original checker.ts site-count limitation. No backend or JavaScript output is produced.

Final data is under continuation/e/, e-exact/ and sdk-e/. e-analysis.json and e-site-comparison.json retain the complete tables and additions/removals. Abandoned controls have ABANDONED.json markers and are excluded from every final count. New Go evidence uses .go.txt. Files over 1,000,000 bytes are gzip compressed and SHA256SUMS covers the retained files.
