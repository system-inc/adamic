# SourceFileLike sites (baobab)

Measured main: `d690f276749f91e6b36fe3a9726a5a453fcd7d39`. Upstream TypeScript 6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`. Report only; no compiler or adaptation changes.

Attempted 79 of 79 compiler files; 75 completed full-mode streams. Physical sites: 15; distinct finding/phase/unit observations: 15.

**Counts are lower bounds.** Incomplete streams: `checker.ts`, `program.ts`, `scanner.ts`, `utilities.ts`.

## By callee

| Callee | Sites | SourceFile sites | Other source types | Files | First-stop roots on c504dd8f | Reads optional method | Method-only candidate sites for road 2 |
| --- | ---: | ---: | ---: | ---: | ---: | --- | --- |
| `getLineAndCharacterOfPosition` | 9 | 8 | 1 | 5 | 73 | no | 9 |
| `getLineStarts` | 2 | 2 | 0 | 2 | 0 | no | 2 |
| `getPositionOfLineAndCharacter` | 0 | 0 | 0 | 0 | 0 | yes | 0 |
| `transferSourceFileChildren` | 2 | 2 | 0 | 1 | 0 | no | 0 |
| `unsetNodeChildren` | 2 | 2 | 0 | 1 | 0 | no | 0 |

Reason census: 14 site(s): optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; 1 site(s): optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceMapSource, which can hide fields.

First-stop attribution comes from `typescript/next-stops-c504dd8f` at `fe08a0d47a3d28f646db181ac9e8456e5c60885f`, file `stage3/census/next-stops-c504dd8f/grouped-first-stops.json`, copied as prior-first-stops.json. It names tracing.ts:206:67 for 73 roots. Its other six roots have unrelated outcomes. First-stop roots count only sites explicitly named in that evidence; it does not predict their next blocker.

## Initial body-only road 2 hypothesis

Body inspection identifies 11 of 15 observed sites whose optional-method requirement would be removed by the two parameter changes. This was the original report hypothesis, not a measured clearance result. It does not establish that the replacement lineMap view is legal. The follow-up measurements below, when present, take precedence.

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
| tracing.ts:206:67 | refusal_scan | getLocation | getLineAndCharacterOfPosition (call argument) | sourceFile -&gt; getLineStarts; line-start array -&gt; computeLineAndCharacterOfPosition -&gt; computeLineOfPosition | yes | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| tracing.ts:207:65 | refusal_scan | getLocation | getLineAndCharacterOfPosition (call argument) | sourceFile -&gt; getLineStarts; line-start array -&gt; computeLineAndCharacterOfPosition -&gt; computeLineOfPosition | yes | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| transformers/declarations.ts:1264:70 | refusal_scan | transformDeclarations / visitDeclarationSubtree | getLineAndCharacterOfPosition (call argument) | sourceFile -&gt; getLineStarts; line-start array -&gt; computeLineAndCharacterOfPosition -&gt; computeLineOfPosition | yes | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| transformers/declarations.ts:1264:139 | refusal_scan | transformDeclarations / visitDeclarationSubtree | getLineAndCharacterOfPosition (call argument) | sourceFile -&gt; getLineStarts; line-start array -&gt; computeLineAndCharacterOfPosition -&gt; computeLineOfPosition | yes | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
| transformers/jsx.ts:359:63 | refusal_scan | transformJsx / visitJsxOpeningLikeElementOrFragmentJSX | getLineAndCharacterOfPosition (call argument) | sourceFile -&gt; getLineStarts; line-start array -&gt; computeLineAndCharacterOfPosition -&gt; computeLineOfPosition | yes | optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields |
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
| program.ts | -15 | 300.42 | no | yes |
| programDiagnostics.ts | 0 | 47.93 | yes | no |
| resolutionCache.ts | 0 | 65.13 | yes | no |
| scanner.ts | -15 | 300.12 | no | yes |
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
| utilities.ts | -15 | 300.12 | no | yes |
| utilitiesPublic.ts | 0 | 207.54 | yes | no |
| visitorPublic.ts | 0 | 37.85 | yes | no |
| watch.ts | 0 | 52.46 | yes | no |
| watchPublic.ts | 0 | 23.93 | yes | no |
| watchUtilities.ts | 0 | 67.72 | yes | no |

**Incomplete census.** Counts are observed lower bounds. Missing or timed-out file results must not be treated as zero sites. See coverage.json and raw per-file logs/status files.

Raw per-file findings, timing streams, logs, and child statuses are under raw/. Files over 1,000,000 bytes are gzip compressed. run-piece.py, collect.py, classify.cjs, and report.py reproduce the run and tables; provenance.json records pins. Setup and preparation logs are under preparation/. Checkpoints are snapshot commits on typescript/sourcefile-like-sites-checkpoint; the final report branch contains one commit on the measured main tip.
