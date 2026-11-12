# Final main overload-result census

Measured **adamic main `219dbb6e7e20ea99d3e5d4b61c0cb14102dc9205`**, fetched and frozen at the start of this run. No pair or scratch composition is used. This commit holds only this census directory.

The landed step 2, step 2b, step 3 and view-entry-fix commits are all verified ancestors of the measured main tip:

| Change | Landed commit |
| --- | --- |
| abebb00b | abebb00bf6312fb92798095515fbee86b8ed78e3 |
| a57ffec0 | a57ffec0d6037086adc4847ee9410d74b8b50522 |
| dd4c82e1 | dd4c82e159072c2b3ead05396ac10ad5ffa895c2 |
| 9ae9f988 | 9ae9f9888b29ef3108952246ce755a707c722c2f |


## Counting definitions

Every observation whose exact `reason` starts with `overload N of <function> result` or `overload N of <function> parameter` is included, in every phase and from every attempting unit. `N` is numeric and the function is the diagnostic’s function-name token; `type parameter` reasons are not silently counted as `parameter` reasons.

The four columns distinguish different questions: **observations** retain repeated diagnostics from different attempting units; **raw unique** uses the prior census meter identity `(kind, source identity, reason, text)`; **distinct** uses source path, source SHA-256, outermost contiguous property/element access start and exact reason; **first-stop units** counts units whose first lowering/deferred finding matches the prefix. Raw and distinct normalization use main’s unmodified `stage3/census/pair/normalize.cjs`; no comparison or paired delta is performed.

All matching observations, full reason/text, phase, unit names/locations, and every raw/distinct representative and its observation references are in [overload-findings.json](overload-findings.json). The stock normalization request and identities are retained separately.

## 1. Overload-result findings

| Function | Observations | Raw unique | Distinct | First-stop units | Handoff | Check |
| --- | --- | --- | --- | --- | --- | --- |
| arrayFrom | 4 | 1 | 1 | 4 | 4 | matches first-stop units |
| arrayToMap | 2 | 1 | 1 | 2 | 2 | matches first-stop units |
| convertOptionsFromJson | 2 | 1 | 1 | 2 | 2 | matches first-stop units |
| createBuilderProgram | 9 | 1 | 1 | 9 | 9 | matches first-stop units |
| filter | 19 | 1 | 1 | 19 | 19 | matches first-stop units |
| getDiagnostics | 1 | 1 | 1 | 1 | 1 | matches first-stop units |
| getParseTreeNode | 2 | 1 | 1 | 2 | 2 | matches first-stop units |
| getPathComponents | 11 | 1 | 1 | 11 | — | not specified |
| getProperties | 3 | 1 | 1 | 3 | 3 | matches first-stop units |
| injectClassNamedEvaluationHelperBlockIfMissing | 1 | 1 | 1 | 1 | 1 | matches first-stop units |
| sameMap | 1 | 1 | 1 | 1 | 1 | matches first-stop units |
| transformNamedEvaluation | 1 | 1 | 1 | 1 | 1 | matches first-stop units |
| TOTAL | 56 | 12 | 12 | 56 | 45 | first-stop total corrected |

The handoff's 45 becomes **56** per-unit result first stops. All 11 listed per-function counts are confirmed; the additional **getPathComponents=11** accounts for the difference. Its exact reason is `overload 1 of getPathComponents result PathPathComponents cannot be served by implementation result string[]` at `src/compiler/path.ts:492:1`.

Observed result-prefix functions: **12**. Listed but absent: **none**. Present but not in the handoff: **getPathComponents**.

### Complete result inventories

Each raw row is one complete raw identity; each distinct row is one outermost-access/reason identity. Locations are relative to the frozen adapted tree. Full diagnostic text, source hashes and observation-to-identity references are in `overload-findings.json`. Distinct offsets below are zero-based UTF-8 byte offsets.

#### arrayFrom

Raw identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/core.ts:1334:1 | overload 1 of arrayFrom result U[] cannot be served by implementation result (T \| U)[] | 4 | createCommentDirectivesMap @ src/compiler/utilities.ts:1210:1; createDiagnosticForInvalidCustomType @ src/compiler/commandLineParser.ts:1873:1; generateOptionOutput @ src/compiler/executeCommandLine.ts:244:1; getBuildInfo @ src/compiler/builder.ts:1258:1 |


Distinct identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/core.ts:byte 46498 | overload 1 of arrayFrom result U[] cannot be served by implementation result (T \| U)[] | 4 | createCommentDirectivesMap @ src/compiler/utilities.ts:1210:1; createDiagnosticForInvalidCustomType @ src/compiler/commandLineParser.ts:1873:1; generateOptionOutput @ src/compiler/executeCommandLine.ts:244:1; getBuildInfo @ src/compiler/builder.ts:1258:1 |


#### arrayToMap

Raw identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/core.ts:1400:1 | overload 2 of arrayToMap result Map&lt;K, V2&gt; cannot be served by implementation result Map&lt;K, V1 \| V2&gt; | 2 | commandLineOptionsToMap @ src/compiler/commandLineParser.ts:2307:1; createBuilderProgramUsingIncrementalBuildInfo @ src/compiler/builder.ts:2277:1 |


Distinct identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/core.ts:byte 48714 | overload 2 of arrayToMap result Map&lt;K, V2&gt; cannot be served by implementation result Map&lt;K, V1 \| V2&gt; | 2 | commandLineOptionsToMap @ src/compiler/commandLineParser.ts:2307:1; createBuilderProgramUsingIncrementalBuildInfo @ src/compiler/builder.ts:2277:1 |


#### convertOptionsFromJson

Raw identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/commandLineParser.ts:3797:1 | overload 1 of convertOptionsFromJson result WatchOptions \| undefined cannot be served by implementation result CompilerOptions \| TypeAcquisition \| WatchOptions \| undefined | 2 | convertCompilerOptionsFromJson @ src/compiler/commandLineParser.ts:3755:1; convertTypeAcquisitionFromJson @ src/compiler/commandLineParser.ts:3761:1 |


Distinct identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/commandLineParser.ts:byte 159676 | overload 1 of convertOptionsFromJson result WatchOptions \| undefined cannot be served by implementation result CompilerOptions \| TypeAcquisition \| WatchOptions \| undefined | 2 | convertCompilerOptionsFromJson @ src/compiler/commandLineParser.ts:3755:1; convertTypeAcquisitionFromJson @ src/compiler/commandLineParser.ts:3761:1 |


#### createBuilderProgram

Raw identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/builder.ts:1686:1 | overload 1 of createBuilderProgram result SemanticDiagnosticsBuilderProgram cannot be served by implementation result BuilderProgram \| undefined | 9 | createBuilderProgram @ src/compiler/builder.ts:1686:1; createBuilderProgram @ src/compiler/builder.ts:1691:1; createBuilderProgram @ src/compiler/builder.ts:1696:1; createEmitAndSemanticDiagnosticsBuilderProgram @ src/compiler/builderPublic.ts:201:1; createEmitAndSemanticDiagnosticsBuilderProgram @ src/compiler/builderPublic.ts:207:1; createEmitAndSemanticDiagnosticsBuilderProgram @ src/compiler/builderPublic.ts:215:1; createSemanticDiagnosticsBuilderProgram @ src/compiler/builderPublic.ts:162:1; createSemanticDiagnosticsBuilderProgram @ src/compiler/builderPublic.ts:168:1; createSemanticDiagnosticsBuilderProgram @ src/compiler/builderPublic.ts:176:1 |


Distinct identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/builder.ts:byte 76960 | overload 1 of createBuilderProgram result SemanticDiagnosticsBuilderProgram cannot be served by implementation result BuilderProgram \| undefined | 9 | createBuilderProgram @ src/compiler/builder.ts:1686:1; createBuilderProgram @ src/compiler/builder.ts:1691:1; createBuilderProgram @ src/compiler/builder.ts:1696:1; createEmitAndSemanticDiagnosticsBuilderProgram @ src/compiler/builderPublic.ts:201:1; createEmitAndSemanticDiagnosticsBuilderProgram @ src/compiler/builderPublic.ts:207:1; createEmitAndSemanticDiagnosticsBuilderProgram @ src/compiler/builderPublic.ts:215:1; createSemanticDiagnosticsBuilderProgram @ src/compiler/builderPublic.ts:162:1; createSemanticDiagnosticsBuilderProgram @ src/compiler/builderPublic.ts:168:1; createSemanticDiagnosticsBuilderProgram @ src/compiler/builderPublic.ts:176:1 |


#### filter

Raw identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/core.ts:261:1 | overload 1 of filter result U[] cannot be served by implementation result readonly T[] \| undefined | 19 | convertToTSConfig @ src/compiler/commandLineParser.ts:2660:1; emitDetachedComments @ src/compiler/utilities.ts:7010:1; filterSemanticDiagnostics @ src/compiler/program.ts:5099:1; getCommonSourceDirectoryOfConfig @ src/compiler/emitter.ts:685:1; getDeclarationDiagnostics @ src/compiler/transformers/declarations.ts:221:1; getDecorators @ src/compiler/utilitiesPublic.ts:1024:1; getFilesInErrorForSummary @ src/compiler/watch.ts:233:1; getImportedHelpers @ src/compiler/factory/utilities.ts:760:1; getInitializedVariables @ src/compiler/utilities.ts:8017:1; getJSDocCommentRanges @ src/compiler/utilities.ts:2761:1; getOptionsForHelp @ src/compiler/executeCommandLine.ts:176:1; getSemanticJsxChildren @ src/compiler/utilities.ts:6044:1; getSourceFilesToEmit @ src/compiler/utilities.ts:6632:1; getStaticPropertiesAndClassStaticBlock @ src/compiler/transformers/utilities.ts:598:1; getStaticPropertiesAndClassStaticBlock @ src/compiler/transformers/utilities.ts:600:1; getStaticPropertiesAndClassStaticBlock @ src/compiler/transformers/utilities.ts:602:1; matchesExclude @ src/compiler/commandLineParser.ts:4077:1; printAllHelp @ src/compiler/executeCommandLine.ts:513:1; printBuildHelp @ src/compiler/executeCommandLine.ts:523:1 |


Distinct identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/core.ts:byte 8777 | overload 1 of filter result U[] cannot be served by implementation result readonly T[] \| undefined | 19 | convertToTSConfig @ src/compiler/commandLineParser.ts:2660:1; emitDetachedComments @ src/compiler/utilities.ts:7010:1; filterSemanticDiagnostics @ src/compiler/program.ts:5099:1; getCommonSourceDirectoryOfConfig @ src/compiler/emitter.ts:685:1; getDeclarationDiagnostics @ src/compiler/transformers/declarations.ts:221:1; getDecorators @ src/compiler/utilitiesPublic.ts:1024:1; getFilesInErrorForSummary @ src/compiler/watch.ts:233:1; getImportedHelpers @ src/compiler/factory/utilities.ts:760:1; getInitializedVariables @ src/compiler/utilities.ts:8017:1; getJSDocCommentRanges @ src/compiler/utilities.ts:2761:1; getOptionsForHelp @ src/compiler/executeCommandLine.ts:176:1; getSemanticJsxChildren @ src/compiler/utilities.ts:6044:1; getSourceFilesToEmit @ src/compiler/utilities.ts:6632:1; getStaticPropertiesAndClassStaticBlock @ src/compiler/transformers/utilities.ts:598:1; getStaticPropertiesAndClassStaticBlock @ src/compiler/transformers/utilities.ts:600:1; getStaticPropertiesAndClassStaticBlock @ src/compiler/transformers/utilities.ts:602:1; matchesExclude @ src/compiler/commandLineParser.ts:4077:1; printAllHelp @ src/compiler/executeCommandLine.ts:513:1; printBuildHelp @ src/compiler/executeCommandLine.ts:523:1 |


#### getDiagnostics

Raw identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/utilities.ts:6120:5 | overload 1 of getDiagnostics result DiagnosticWithLocation[] cannot be served by implementation result Diagnostic[] | 1 | createDiagnosticCollection @ src/compiler/utilities.ts:6058:1 |


Distinct identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/utilities.ts:byte 240569 | overload 1 of getDiagnostics result DiagnosticWithLocation[] cannot be served by implementation result Diagnostic[] | 1 | createDiagnosticCollection @ src/compiler/utilities.ts:6058:1 |


#### getParseTreeNode

Raw identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/utilitiesPublic.ts:826:1 | overload 2 of getParseTreeNode result T \| undefined cannot be served by implementation result Node \| undefined | 2 | getParseTreeNode @ src/compiler/utilitiesPublic.ts:817:1; getParseTreeNode @ src/compiler/utilitiesPublic.ts:827:1 |


Distinct identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/utilitiesPublic.ts:byte 31802 | overload 2 of getParseTreeNode result T \| undefined cannot be served by implementation result Node \| undefined | 2 | getParseTreeNode @ src/compiler/utilitiesPublic.ts:817:1; getParseTreeNode @ src/compiler/utilitiesPublic.ts:827:1 |


#### getPathComponents

Raw identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/path.ts:492:1 | overload 1 of getPathComponents result PathPathComponents cannot be served by implementation result string[] | 11 | comparePathsWorker @ src/compiler/path.ts:886:1; containsPath @ src/compiler/path.ts:965:1; containsPath @ src/compiler/path.ts:967:1; containsPath @ src/compiler/path.ts:969:1; getNormalizedPathComponents @ src/compiler/path.ts:624:1; getPathComponents @ src/compiler/path.ts:492:1; getPathComponents @ src/compiler/path.ts:494:1; getPathComponents @ src/compiler/path.ts:495:1; getPathComponentsRelativeTo @ src/compiler/path.ts:1012:1; guessDirectorySymlink @ src/compiler/utilities.ts:9559:1; normalizePathForCJSResolution @ src/compiler/moduleNameResolver.ts:1966:1 |


Distinct identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/path.ts:byte 19623 | overload 1 of getPathComponents result PathPathComponents cannot be served by implementation result string[] | 11 | comparePathsWorker @ src/compiler/path.ts:886:1; containsPath @ src/compiler/path.ts:965:1; containsPath @ src/compiler/path.ts:967:1; containsPath @ src/compiler/path.ts:969:1; getNormalizedPathComponents @ src/compiler/path.ts:624:1; getPathComponents @ src/compiler/path.ts:492:1; getPathComponents @ src/compiler/path.ts:494:1; getPathComponents @ src/compiler/path.ts:495:1; getPathComponentsRelativeTo @ src/compiler/path.ts:1012:1; guessDirectorySymlink @ src/compiler/utilities.ts:9559:1; normalizePathForCJSResolution @ src/compiler/moduleNameResolver.ts:1966:1 |


#### getProperties

Raw identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/transformers/utilities.ts:585:1 | overload 1 of getProperties result readonly InitializedPropertyDeclaration[] cannot be served by implementation result readonly PropertyDeclaration[] | 3 | getProperties @ src/compiler/transformers/utilities.ts:585:1; getProperties @ src/compiler/transformers/utilities.ts:587:1; getProperties @ src/compiler/transformers/utilities.ts:589:1 |


Distinct identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/transformers/utilities.ts:byte 23717 | overload 1 of getProperties result readonly InitializedPropertyDeclaration[] cannot be served by implementation result readonly PropertyDeclaration[] | 3 | getProperties @ src/compiler/transformers/utilities.ts:585:1; getProperties @ src/compiler/transformers/utilities.ts:587:1; getProperties @ src/compiler/transformers/utilities.ts:589:1 |


#### injectClassNamedEvaluationHelperBlockIfMissing

Raw identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/transformers/namedEvaluation.ts:173:1 | overload 1 of injectClassNamedEvaluationHelperBlockIfMissing result Extract&lt;ClassDeclaration, Pick&lt;...&gt;&gt; \| Extract&lt;...&gt; cannot be served by implementation result ClassLikeDeclaration | 1 | injectClassNamedEvaluationHelperBlockIfMissing @ src/compiler/transformers/namedEvaluation.ts:179:1 |


Distinct identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/transformers/namedEvaluation.ts:byte 6604 | overload 1 of injectClassNamedEvaluationHelperBlockIfMissing result Extract&lt;ClassDeclaration, Pick&lt;...&gt;&gt; \| Extract&lt;...&gt; cannot be served by implementation result ClassLikeDeclaration | 1 | injectClassNamedEvaluationHelperBlockIfMissing @ src/compiler/transformers/namedEvaluation.ts:179:1 |


#### sameMap

Raw identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/core.ts:342:1 | overload 1 of sameMap result U[] cannot be served by implementation result readonly U[] \| undefined | 1 | repopulateDiagnostics @ src/compiler/builder.ts:520:1 |


Distinct identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/core.ts:byte 11572 | overload 1 of sameMap result U[] cannot be served by implementation result readonly U[] \| undefined | 1 | repopulateDiagnostics @ src/compiler/builder.ts:520:1 |


#### transformNamedEvaluation

Raw identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/transformers/namedEvaluation.ts:469:1 | overload 1 of transformNamedEvaluation result Extract&lt;AssignmentExpression&lt;EqualsToken&gt; &amp; { readonly left: Identifier; readonly right: WrappedExpression&lt;AnonymousFunctionDefinition&gt;; }, Pick&lt;...&gt;&gt; \| ... 7 more ... \| Extract&lt;...&gt; cannot be served by implementation result BinaryExpression \| BindingElement \| ExportAssignment \| ParameterDeclaration \| PropertyAssignment \| PropertyDeclaration \| ShorthandPropertyAssignment \| VariableDeclaration | 1 | transformNamedEvaluation @ src/compiler/transformers/namedEvaluation.ts:470:1 |


Distinct identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/transformers/namedEvaluation.ts:byte 20533 | overload 1 of transformNamedEvaluation result Extract&lt;AssignmentExpression&lt;EqualsToken&gt; &amp; { readonly left: Identifier; readonly right: WrappedExpression&lt;AnonymousFunctionDefinition&gt;; }, Pick&lt;...&gt;&gt; \| ... 7 more ... \| Extract&lt;...&gt; cannot be served by implementation result BinaryExpression \| BindingElement \| ExportAssignment \| ParameterDeclaration \| PropertyAssignment \| PropertyDeclaration \| ShorthandPropertyAssignment \| VariableDeclaration | 1 | transformNamedEvaluation @ src/compiler/transformers/namedEvaluation.ts:470:1 |


Handoff first-stop corrections: No count correction among listed functions; presence differences are named above.

## 2. Overload-parameter findings

| Function | Observations | Raw unique | Distinct | First-stop units | Handoff | Check |
| --- | --- | --- | --- | --- | --- | --- |
| assertEachNode | 4 | 1 | 1 | 4 | — | not specified |
| assertNode | 1 | 1 | 1 | 1 | — | not specified |
| convertOrRepopulateDiagnosticMessageChainArray | 1 | 1 | 1 | 1 | — | not specified |
| getSupportedExtensionsWithJsonIfResolveJsonModule | 5 | 1 | 1 | 5 | — | not specified |
| reduceLeft | 1 | 1 | 1 | 1 | — | not specified |
| visitArray | 1 | 1 | 1 | 1 | — | not specified |
| visitArrayWorker | 1 | 1 | 1 | 1 | — | not specified |
| visitEachChild | 6 | 1 | 1 | 6 | 6 | matches first-stop units |
| visitNode | 4 | 1 | 1 | 4 | — | not specified |
| visitNodes | 2 | 1 | 1 | 2 | — | not specified |
| visitParameterList | 3 | 1 | 1 | 3 | 3 | matches first-stop units |
| TOTAL | 29 | 11 | 11 | 29 | 27 | first-stop total corrected |

Observed parameter-prefix functions: **11**. Handoff total 27 versus measured **29 first-stop units**, **29 observations**, **11 raw unique findings**, and **11 distinct findings**.

### Complete parameter inventories

Each raw row is one complete raw identity; each distinct row is one outermost-access/reason identity. Locations are relative to the frozen adapted tree. Full diagnostic text, source hashes and observation-to-identity references are in `overload-findings.json`. Distinct offsets below are zero-based UTF-8 byte offsets.

#### assertEachNode

Raw identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/debug.ts:259:98 | overload 4 of assertEachNode parameter test cannot be served by implementation parameter test | 4 | makeArrayAssignmentPattern @ src/compiler/transformers/destructuring.ts:605:1; makeArrayBindingPattern @ src/compiler/transformers/destructuring.ts:600:1; makeObjectAssignmentPattern @ src/compiler/transformers/destructuring.ts:615:1; makeObjectBindingPattern @ src/compiler/transformers/destructuring.ts:610:1 |


Distinct identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/debug.ts:byte 9284 | overload 4 of assertEachNode parameter test cannot be served by implementation parameter test | 4 | makeArrayAssignmentPattern @ src/compiler/transformers/destructuring.ts:605:1; makeArrayBindingPattern @ src/compiler/transformers/destructuring.ts:600:1; makeObjectAssignmentPattern @ src/compiler/transformers/destructuring.ts:615:1; makeObjectBindingPattern @ src/compiler/transformers/destructuring.ts:610:1 |


#### assertNode

Raw identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/debug.ts:281:82 | overload 1 of assertNode parameter test cannot be served by implementation parameter test | 1 | flattenDestructuringBinding @ src/compiler/transformers/destructuring.ts:244:1 |


Distinct identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/debug.ts:byte 10433 | overload 1 of assertNode parameter test cannot be served by implementation parameter test | 1 | flattenDestructuringBinding @ src/compiler/transformers/destructuring.ts:244:1 |


#### convertOrRepopulateDiagnosticMessageChainArray

Raw identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/builder.ts:567:5 | overload 1 of convertOrRepopulateDiagnosticMessageChainArray parameter array cannot be served by implementation parameter array | 1 | convertOrRepopulateDiagnosticMessageChainArray @ src/compiler/builder.ts:572:1 |


Distinct identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/builder.ts:byte 24439 | overload 1 of convertOrRepopulateDiagnosticMessageChainArray parameter array cannot be served by implementation parameter array | 1 | convertOrRepopulateDiagnosticMessageChainArray @ src/compiler/builder.ts:572:1 |


#### getSupportedExtensionsWithJsonIfResolveJsonModule

Raw identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/utilities.ts:10012:105 | overload 1 of getSupportedExtensionsWithJsonIfResolveJsonModule parameter supportedExtensions cannot be served by implementation parameter supportedExtensions | 5 | getFileNamesFromConfigSpecs @ src/compiler/commandLineParser.ts:3950:1; getSupportedExtensionsWithJsonIfResolveJsonModule @ src/compiler/utilities.ts:10012:1; getSupportedExtensionsWithJsonIfResolveJsonModule @ src/compiler/utilities.ts:10014:1; getSupportedExtensionsWithJsonIfResolveJsonModule @ src/compiler/utilities.ts:10016:1; isSupportedSourceFileName @ src/compiler/utilities.ts:10146:1 |


Distinct identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/utilities.ts:byte 401660 | overload 1 of getSupportedExtensionsWithJsonIfResolveJsonModule parameter supportedExtensions cannot be served by implementation parameter supportedExtensions | 5 | getFileNamesFromConfigSpecs @ src/compiler/commandLineParser.ts:3950:1; getSupportedExtensionsWithJsonIfResolveJsonModule @ src/compiler/utilities.ts:10012:1; getSupportedExtensionsWithJsonIfResolveJsonModule @ src/compiler/utilities.ts:10014:1; getSupportedExtensionsWithJsonIfResolveJsonModule @ src/compiler/utilities.ts:10016:1; isSupportedSourceFileName @ src/compiler/utilities.ts:10146:1 |


#### reduceLeft

Raw identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/core.ts:1236:67 | overload 1 of reduceLeft parameter f cannot be served by implementation parameter f | 1 | tryGetModuleNameFromRootDirs @ src/compiler/moduleSpecifiers.ts:1169:1 |


Distinct identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/core.ts:byte 43844 | overload 1 of reduceLeft parameter f cannot be served by implementation parameter f | 1 | tryGetModuleNameFromRootDirs @ src/compiler/moduleSpecifiers.ts:1169:1 |


#### visitArray

Raw identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/visitorPublic.ts:276:5 | overload 1 of visitArray parameter visitor cannot be served by implementation parameter visitor | 1 | visitArray @ src/compiler/visitorPublic.ts:289:1 |


Distinct identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/visitorPublic.ts:byte 9732 | overload 1 of visitArray parameter visitor cannot be served by implementation parameter visitor | 1 | visitArray @ src/compiler/visitorPublic.ts:289:1 |


#### visitArrayWorker

Raw identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/visitorPublic.ts:316:5 | overload 1 of visitArrayWorker parameter visitor cannot be served by implementation parameter visitor | 1 | visitArrayWorker @ src/compiler/visitorPublic.ts:328:1 |


Distinct identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/visitorPublic.ts:byte 11011 | overload 1 of visitArrayWorker parameter visitor cannot be served by implementation parameter visitor | 1 | visitArrayWorker @ src/compiler/visitorPublic.ts:328:1 |


#### visitEachChild

Raw identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/visitorPublic.ts:587:119 | overload 2 of visitEachChild parameter nodesVisitor cannot be served by implementation parameter nodesVisitor | 6 | transformES2016 @ src/compiler/transformers/es2016.ts:21:1; transformES2017 @ src/compiler/transformers/es2017.ts:119:1; transformES2019 @ src/compiler/transformers/es2019.ts:17:1; transformES2020 @ src/compiler/transformers/es2020.ts:42:1; transformES2021 @ src/compiler/transformers/es2021.ts:25:1; transformLegacyDecorators @ src/compiler/transformers/legacyDecorators.ts:88:1 |


Distinct identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/visitorPublic.ts:byte 24998 | overload 2 of visitEachChild parameter nodesVisitor cannot be served by implementation parameter nodesVisitor | 6 | transformES2016 @ src/compiler/transformers/es2016.ts:21:1; transformES2017 @ src/compiler/transformers/es2017.ts:119:1; transformES2019 @ src/compiler/transformers/es2019.ts:17:1; transformES2020 @ src/compiler/transformers/es2020.ts:42:1; transformES2021 @ src/compiler/transformers/es2021.ts:25:1; transformLegacyDecorators @ src/compiler/transformers/legacyDecorators.ts:88:1 |


#### visitNode

Raw identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/visitorPublic.ts:123:5 | overload 1 of visitNode parameter node cannot be served by implementation parameter node | 4 | flattenBindingOrAssignmentElement @ src/compiler/transformers/destructuring.ts:338:1; processTaggedTemplateExpression @ src/compiler/transformers/taggedTemplate.ts:35:1; transformESNext @ src/compiler/transformers/esnext.ts:71:1; visitNode @ src/compiler/visitorPublic.ts:148:1 |


Distinct identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/visitorPublic.ts:byte 3355 | overload 1 of visitNode parameter node cannot be served by implementation parameter node | 4 | flattenBindingOrAssignmentElement @ src/compiler/transformers/destructuring.ts:338:1; processTaggedTemplateExpression @ src/compiler/transformers/taggedTemplate.ts:35:1; transformESNext @ src/compiler/transformers/esnext.ts:71:1; visitNode @ src/compiler/visitorPublic.ts:148:1 |


#### visitNodes

Raw identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/visitorPublic.ts:196:5 | overload 1 of visitNodes parameter visitor cannot be served by implementation parameter visitor | 2 | visitLexicalEnvironment @ src/compiler/visitorPublic.ts:382:1; visitNodes @ src/compiler/visitorPublic.ts:223:1 |


Distinct identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/visitorPublic.ts:byte 6508 | overload 1 of visitNodes parameter visitor cannot be served by implementation parameter visitor | 2 | visitLexicalEnvironment @ src/compiler/visitorPublic.ts:382:1; visitNodes @ src/compiler/visitorPublic.ts:223:1 |


#### visitParameterList

Raw identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/visitorPublic.ts:393:126 | overload 1 of visitParameterList parameter nodesVisitor cannot be served by implementation parameter nodesVisitor | 3 | visitParameterList @ src/compiler/visitorPublic.ts:393:1; visitParameterList @ src/compiler/visitorPublic.ts:394:1; visitParameterList @ src/compiler/visitorPublic.ts:395:1 |


Distinct identities:

| Location | Exact reason | Observations | Attempting units |
| --- | --- | --- | --- |
| src/compiler/visitorPublic.ts:byte 14064 | overload 1 of visitParameterList parameter nodesVisitor cannot be served by implementation parameter nodesVisitor | 3 | visitParameterList @ src/compiler/visitorPublic.ts:393:1; visitParameterList @ src/compiler/visitorPublic.ts:394:1; visitParameterList @ src/compiler/visitorPublic.ts:395:1 |


The table includes every observed function, not only the five visitor functions mentioned in the handoff. The explicit visitEachChild=6 and visitParameterList=3 expectations are checked individually; other per-function expectations were not specified.

## 3. First-stop comparison

| Measure | 78933e52 composition head | Current main | Change |
| --- | --- | --- | --- |
| Overload-result prefix first-stop units | 45 | 56 | 11 |
| Units with no first lowering/deferred stop | 3187 | 3186 | -1 |

The prior canonical `cannot be served by implementation result` first-stop definition counts **56** on current main. A unit with no first stop is not proof of complete compilation.

| Current-main coverage | Count |
| --- | --- |
| Resolved source files | 81 |
| Checker diagnostics | 0 |
| Units | 4566 |
| Attempted units | 4063 |
| Units with no first stop | 3186 |

The comparison is numerical, against the previous report’s 45 and 3,187. Adapted input manifests match: **False**; changed input files: **2**. Differences are retained in [input-difference-from-78933e52.json](input-difference-from-78933e52.json). No old stream is reused and no paired regression claim is made.

## Method and validation

This is the same stock **first-error latent entry census** as `stage3/census/overload-phantom-78933e52`: `LATENT_FULL=0`, `LATENT_ASSERT_NO_OUTPUT=1`, rooted at the adapted `src/tsc/tsc.ts`. The compiler overlay is built from the measured main tree itself with its own `make_overlay.py` and `entry_overlay.py`, exact recursive submodule pins, and source-generated adaptation from its own `apply.sh`. The TypeScript source pin, complete adapted-source hashes, overlay maps and executable hashes are retained.

Ordinary Load and Lower remain disabled. No backend or usable IR is produced. Each of eight disjoint attempted-unit shards keeps the full project and registration graph. Shards 0 and 1 run complete refusal scans, which must agree exactly; shards 2–7 omit only repeated scans. Two shards run concurrently with `GOMEMLIMIT=3GiB`. Eligibility and generic deferral remain unchanged. Deferred/skipped records deduplicate, and generic call witnesses union under the original Go order and five-witness cap. Each attempted unit appears once. Process-scoped relation/cache reports stay separate, never summed; any overflow or global finding stops composition.

Both binary variants pass the unsplit-versus-sharded fixture, including a real overloaded callback-value stop and cross-shard generic witnesses. The existing latent audit and all 20 normalization/pair-tool tests pass; the tests do not constitute a paired census. A stock-parser inventory independently checks every top-level unit and source order. Validation verifies timings, shard membership, exact archived recomposition, normalization hashes, both deduplication counts, complete prefix selection, landed ancestry pins, and handoff comparisons.

Every timing wrapper reports its child’s own exit code and propagates it; shell wrappers also report and propagate their child status. Each child is bounded at 1,740 seconds with at most ten seconds to kill, below 30 minutes. No measured child timed out. Go evidence is named `.go.txt`; logs and JSONL are gzip compressed without changing their bytes. Commands and all analysis/build/shard scripts are retained.

Each census child’s own measurements (overlapping runs):

| Shard | Wall s | User s | System s | Peak RSS MiB | Child exit |
| --- | --- | --- | --- | --- | --- |
| 0 | 1096.903 | 1248.806 | 20.451 | 2782.08 | 0 |
| 1 | 1069.95 | 1221.49 | 20.297 | 2745.02 | 0 |
| 2 | 132.603 | 135.696 | 3.074 | 979.38 | 0 |
| 3 | 121.094 | 124.406 | 3.344 | 969.0 | 0 |
| 4 | 133.959 | 136.95 | 3.467 | 1119.85 | 0 |
| 5 | 129.797 | 131.966 | 4.47 | 1326.11 | 0 |
| 6 | 128.64 | 132.968 | 2.844 | 1140.61 | 0 |
| 7 | 124.924 | 128.256 | 2.664 | 1008.67 | 0 |
