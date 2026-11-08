# Graph adoption boundary census on bcbe5d12

Task jawqwv7. Report only. Counts cover lowered unit results only, including retained successful statements after full-mode rollback. They do not measure accepted whole programs.

Compiler/apply: `bcbe5d1244ee8a218d860415dea19cca114bae46`. Reference report tree: `30057eca`. TypeScript 6.0.3 source: `050880ce59e30b356b686bd3144efe24f875ebc8`.

Box: `755daee0c6ba`; cgroup CPU quota `400000 100000`; memory limit `17179869184` bytes; 5 visible CPUs. Environment: `ccarenv_b64_Y2NhcmVudl85M2MyY2I4NzA3ZTA4MTkxODMwYTI0NDZlY2ZhNjAyZg`.

## Four-way split, lowered IR only

| Kind | Observations | Distinct reported locations |
| --- | ---: | ---: |
| views | 59681 | 630 |
| record_literals | 0 | 0 |
| dynamic_reads | 68 | 40 |
| checked_object_calls | 0 | 0 |

Views combine 1981 ViewOrigins entries and 57700 CheckedFields keys, as requested; these are metadata observations, not that many distinct adapter expressions.

The 79 compiler-file subtotal is `{'views': 59681, 'record_literals': 0, 'dynamic_reads': 68, 'checked_object_calls': 0}`. The two entry-file subtotal is `{'views': 0, 'record_literals': 0, 'dynamic_reads': 0, 'checked_object_calls': 0}`. Both cover lowered unit results only.

Long pole by observed count: **views, 59681**. Coverage caveat: lowered IR only; failed statements, checker-skipped bodies, deferred generics and uncompleted units are absent. Inference: this ranking may differ after those gaps lower.

## Coverage and interpretation

Coverage continuation: 51/51 previously unstarted files now have a single-file full-mode attempt; 0 remain. executeCommandLine.ts is also retried because the earlier emitter piece ended before reaching it. Every continuation child is capped at 1490 seconds, leaving shutdown time below 25 minutes. A nonzero or timed-out single-file attempt is recorded and the runner moves on. 0 continuation files exited nonzero so far.

Bounded-pass stops: first: child exit -15, configured elapsed bound, 1790.787 seconds; full-01: child exit -15, configured elapsed bound, 1791.304 seconds; full-07: child exit -15, configured elapsed bound, 571.096 seconds. Interrupted active units are excluded from completed observation counts.

Entry checker diagnostics: 0. The latent observations still do not establish whole-program lowering or native behavior.

Observed 8690 unit results: {'lowered': 5024, 'stopped': 3666}. First-error pass observed 3447 results; full pieces observed 8679. The summed table uses one result per unit, preferring full mode over first-error mode and later full observations over earlier ones. Earlier completed unit results remain when a later attempt does not reach that unit. Repeated sites in different units are repeated observations, not independent source sites.

Stock unit inventory has 7566 units in 73 emitted files out of 81 resolved files. Units without an observer result, by stock status: `{'deferred_generic_instantiation': 751}`. Missing stock file records after an interrupted pass are not evidence of zero units.

In first-error mode a unit with a lowering finding contributes zero sites: its body is discarded. In full mode the stock hook rolls failed statements back and continues. A `stopped` full-mode observation can therefore contain retained successful statements; its failed statements contribute no sites. Panicked unit results contribute zero sites. `lowered` means no new lowering finding in that unit, not whole-program acceptance. File-level refusal-scan findings do not suppress independent-unit examination.

Every listed expression and metadata entry lacks a source-position field at this base. All positions are explicitly labeled unit fallbacks in the raw stream. Distinct reported locations deduplicate `(kind, unit file:line:column)`, and are not a bound on distinct actual source sites. Several expressions can share a fallback, while dependency bodies repeated in different units receive different owner fallbacks. Exact source-site deduplication cannot be recovered from this IR. The observer does not use the boundary's last-seen graphViewSite or graphDynamicSite.

GraphTypes non-empty in 0 observed results. The latent hook never runs `findCycles`; GraphTypes is not computed per unit. Empty flags here do not establish acyclicity. No graph membership is inferred.

## Per file, lowered IR only

Attribution is to the unit's file, not an independent root. Complete traversal means the census visited the file's candidates; findings, rollback and deferred generic gaps still apply. A zero with zero observed unit results is a coverage gap, not evidence of an empty file. Dependency function bodies or metadata can be included in that unit's result.

| Unit file | Views | Record literals | Dynamic reads | Checked object calls | Observed unit results | Full-mode traversal | Any observed GraphTypes |
| --- | ---: | ---: | ---: | ---: | ---: | --- | --- |
| _namespaces/ts.moduleSpecifiers.ts | 0 | 0 | 0 | 0 | 1 | complete traversal | no, not computed |
| _namespaces/ts.performance.ts | 0 | 0 | 0 | 0 | 1 | complete traversal | no, not computed |
| _namespaces/ts.ts | 0 | 0 | 0 | 0 | 78 | complete traversal | no, not computed |
| binder.ts | 3609 | 0 | 0 | 0 | 180 | complete traversal | no, not computed |
| builder.ts | 0 | 0 | 0 | 0 | 124 | complete traversal | no, not computed |
| builderPublic.ts | 0 | 0 | 0 | 0 | 16 | complete traversal | no, not computed |
| builderState.ts | 0 | 0 | 0 | 0 | 33 | complete traversal | no, not computed |
| builderStatePublic.ts | 0 | 0 | 0 | 0 | 3 | complete traversal | no, not computed |
| checker.ts | 10465 | 0 | 0 | 0 | 1402 | partial | no, not computed |
| commandLineParser.ts | 561 | 0 | 6 | 0 | 173 | complete traversal | no, not computed |
| core.ts | 0 | 0 | 4 | 0 | 59 | complete traversal | no, not computed |
| corePublic.ts | 0 | 0 | 0 | 0 | 9 | complete traversal | no, not computed |
| debug.ts | 0 | 0 | 0 | 0 | 85 | complete traversal | no, not computed |
| diagnosticInformationMap.generated.ts | 0 | 0 | 0 | 0 | 3 | complete traversal | no, not computed |
| emitter.ts | 3433 | 0 | 9 | 0 | 346 | partial | no, not computed |
| executeCommandLine.ts | 160 | 0 | 0 | 0 | 68 | complete traversal | no, not computed |
| expressionToTypeNode.ts | 0 | 0 | 0 | 0 | 0 | not yet attempted | no, not computed |
| factory/baseNodeFactory.ts | 0 | 0 | 0 | 0 | 8 | partial | no, not computed |
| factory/emitHelpers.ts | 0 | 0 | 0 | 0 | 83 | partial | no, not computed |
| factory/emitNode.ts | 0 | 0 | 0 | 0 | 18 | partial | no, not computed |
| factory/nodeChildren.ts | 12 | 0 | 0 | 0 | 6 | partial | no, not computed |
| factory/nodeConverters.ts | 204 | 0 | 0 | 0 | 12 | partial | no, not computed |
| factory/nodeFactory.ts | 621 | 0 | 1 | 0 | 502 | complete traversal | no, not computed |
| factory/nodeTests.ts | 0 | 0 | 0 | 0 | 228 | complete traversal | no, not computed |
| factory/parenthesizerRules.ts | 224 | 0 | 0 | 0 | 42 | complete traversal | no, not computed |
| factory/utilities.ts | 789 | 0 | 0 | 0 | 101 | complete traversal | no, not computed |
| factory/utilitiesPublic.ts | 0 | 0 | 0 | 0 | 3 | complete traversal | no, not computed |
| hostErrors.ts | 0 | 0 | 0 | 0 | 3 | complete traversal | no, not computed |
| moduleNameResolver.ts | 66 | 0 | 1 | 0 | 168 | complete traversal | no, not computed |
| moduleSpecifiers.ts | 0 | 0 | 0 | 0 | 52 | complete traversal | no, not computed |
| parser.ts | 2290 | 0 | 30 | 0 | 545 | complete traversal | no, not computed |
| path.ts | 0 | 0 | 0 | 0 | 77 | complete traversal | no, not computed |
| performance.ts | 0 | 0 | 0 | 0 | 26 | complete traversal | no, not computed |
| performanceCore.ts | 0 | 0 | 0 | 0 | 12 | complete traversal | no, not computed |
| program.ts | 4452 | 0 | 5 | 0 | 235 | complete traversal | no, not computed |
| programDiagnostics.ts | 0 | 0 | 0 | 0 | 12 | complete traversal | no, not computed |
| resolutionCache.ts | 29 | 0 | 0 | 0 | 75 | complete traversal | no, not computed |
| scanner.ts | 0 | 0 | 0 | 0 | 150 | complete traversal | no, not computed |
| semver.ts | 0 | 0 | 0 | 0 | 29 | complete traversal | no, not computed |
| sourcemap.ts | 0 | 0 | 0 | 0 | 58 | complete traversal | no, not computed |
| symbolWalker.ts | 0 | 0 | 0 | 0 | 14 | complete traversal | no, not computed |
| sys.ts | 6 | 0 | 0 | 0 | 137 | complete traversal | no, not computed |
| tracing.ts | 0 | 0 | 0 | 0 | 23 | complete traversal | no, not computed |
| transformer.ts | 0 | 0 | 0 | 0 | 15 | complete traversal | no, not computed |
| transformers/classFields.ts | 3176 | 0 | 0 | 0 | 127 | complete traversal | no, not computed |
| transformers/classThis.ts | 0 | 0 | 0 | 0 | 5 | complete traversal | no, not computed |
| transformers/declarations.ts | 688 | 0 | 1 | 0 | 69 | complete traversal | no, not computed |
| transformers/declarations/diagnostics.ts | 832 | 0 | 0 | 0 | 35 | complete traversal | no, not computed |
| transformers/destructuring.ts | 644 | 0 | 0 | 0 | 26 | complete traversal | no, not computed |
| transformers/es2015.ts | 1844 | 0 | 2 | 0 | 182 | complete traversal | no, not computed |
| transformers/es2016.ts | 78 | 0 | 0 | 0 | 7 | complete traversal | no, not computed |
| transformers/es2017.ts | 112 | 0 | 0 | 0 | 52 | complete traversal | no, not computed |
| transformers/es2018.ts | 815 | 0 | 0 | 0 | 62 | complete traversal | no, not computed |
| transformers/es2019.ts | 0 | 0 | 0 | 0 | 5 | complete traversal | no, not computed |
| transformers/es2020.ts | 853 | 0 | 0 | 0 | 13 | complete traversal | no, not computed |
| transformers/es2021.ts | 40 | 0 | 0 | 0 | 5 | complete traversal | no, not computed |
| transformers/esDecorators.ts | 1514 | 0 | 0 | 0 | 89 | complete traversal | no, not computed |
| transformers/esnext.ts | 589 | 0 | 0 | 0 | 29 | complete traversal | no, not computed |
| transformers/generators.ts | 500 | 0 | 0 | 0 | 145 | complete traversal | no, not computed |
| transformers/jsx.ts | 684 | 0 | 0 | 0 | 40 | complete traversal | no, not computed |
| transformers/legacyDecorators.ts | 867 | 0 | 0 | 0 | 40 | complete traversal | no, not computed |
| transformers/module/esnextAnd2015.ts | 799 | 0 | 0 | 0 | 15 | complete traversal | no, not computed |
| transformers/module/impliedNodeFormatDependent.ts | 0 | 0 | 0 | 0 | 8 | complete traversal | no, not computed |
| transformers/module/module.ts | 1075 | 0 | 0 | 0 | 77 | complete traversal | no, not computed |
| transformers/module/system.ts | 1770 | 0 | 0 | 0 | 71 | complete traversal | no, not computed |
| transformers/namedEvaluation.ts | 0 | 0 | 0 | 0 | 19 | complete traversal | no, not computed |
| transformers/taggedTemplate.ts | 0 | 0 | 0 | 0 | 5 | complete traversal | no, not computed |
| transformers/ts.ts | 1566 | 0 | 1 | 0 | 122 | complete traversal | no, not computed |
| transformers/typeSerializer.ts | 258 | 0 | 0 | 0 | 22 | complete traversal | no, not computed |
| transformers/utilities.ts | 1130 | 0 | 0 | 0 | 48 | complete traversal | no, not computed |
| tsbuild.ts | 0 | 0 | 0 | 0 | 5 | complete traversal | no, not computed |
| tsbuildPublic.ts | 0 | 0 | 0 | 0 | 34 | complete traversal | no, not computed |
| tsc/_namespaces/ts.ts | 0 | 0 | 0 | 0 | 1 | complete traversal | no, not computed |
| tsc/tsc.ts | 0 | 0 | 0 | 0 | 6 | complete traversal | no, not computed |
| types.ts | 0 | 0 | 0 | 0 | 875 | complete traversal | no, not computed |
| utilities.ts | 10398 | 0 | 8 | 0 | 867 | complete traversal | no, not computed |
| utilitiesPublic.ts | 2528 | 0 | 0 | 0 | 243 | complete traversal | no, not computed |
| visitorPublic.ts | 0 | 0 | 0 | 0 | 28 | complete traversal | no, not computed |
| watch.ts | 0 | 0 | 0 | 0 | 36 | complete traversal | no, not computed |
| watchPublic.ts | 0 | 0 | 0 | 0 | 18 | complete traversal | no, not computed |
| watchUtilities.ts | 0 | 0 | 0 | 0 | 44 | complete traversal | no, not computed |

## Top 10 files by record_literals, lowered IR only

| Unit file | Count |
| --- | ---: |
| no sites observed in covered unit results | 0 |

Examples, when available, use unit fallback positions and adapted unit source text; exact expression locations are unavailable.

## Top 10 files by views, lowered IR only

| Unit file | Count |
| --- | ---: |
| checker.ts | 10465 |
| utilities.ts | 10398 |
| program.ts | 4452 |
| binder.ts | 3609 |
| emitter.ts | 3433 |
| transformers/classFields.ts | 3176 |
| utilitiesPublic.ts | 2528 |
| parser.ts | 2290 |
| transformers/es2015.ts | 1844 |
| transformers/module/system.ts | 1770 |

Examples, when available, use unit fallback positions and adapted unit source text; exact expression locations are unavailable.

`checker.ts:2437:5` (unit fallback):
```ts
    function getJsxNamespace(location: Node | undefined): __String {
        if (location) {
            const file = getSourceFileOfNode(location);
```

`checker.ts:3112:9` (unit fallback):
```ts
        function isUsedInFunctionOrInstancePropertyWorker(usage: Node, declaration: Node): boolean {
            return !!findAncestor(usage, current => {
                if (current === declContainer) {
```

`checker.ts:3299:5` (unit fallback):
```ts
    function onSuccessfullyResolvedSymbol(
        errorLocation: Node | undefined,
        result: Symbol,
```

`utilities.ts:1097:1` (unit fallback):
```ts
export function isGrammarError(parent: Node, child: Node | NodeArray<Node>): boolean {
    if (isTypeParameterDeclaration(parent)) return child === parent.expression;
    if (isClassStaticBlockDeclaration(parent)) return child === parent.modifiers;
```

`utilities.ts:1143:1` (unit fallback):
```ts
function isAnyPrologueDirective(node: Node) {
    return isPrologueDirective(node) || !!(getEmitFlags(node) & EmitFlags.CustomPrologue);
}
```

`utilities.ts:1296:1` (unit fallback):
```ts
export function isExportNamespaceAsDefaultDeclaration(node: Node): boolean {
    return !!(isExportDeclaration(node) && node.exportClause && isNamespaceExport(node.exportClause) && moduleExportNameIsDefault(node.exportClause.name));
}
```

`program.ts:937:1` (unit fallback):
```ts
function getModeForUsageLocationWorker(file: Pick<SourceFile, "fileName" | "impliedNodeFormat" | "packageJsonScope">, usage: StringLiteralLike, compilerOptions?: CompilerOptions) {
    if (isImportDeclaration(usage.parent) || isExportDeclaration(usage.parent) || isJSDocImportTag(usage.parent)) {
        const isTypeOnly = isExclusivelyTypeOnlyImportOrExport(usage.parent);
```

`program.ts:959:1` (unit fallback):
```ts
function getEmitSyntaxForUsageLocationWorker(file: Pick<SourceFile, "fileName" | "impliedNodeFormat" | "packageJsonScope">, usage: StringLiteralLike, compilerOptions?: CompilerOptions): ResolutionMode {
    if (!compilerOptions) {
        // This should always be provided, but we try to fail somewhat
```

`program.ts:986:1` (unit fallback):
```ts
export function getResolutionModeOverride(node: ImportAttributes | undefined, grammarErrorOnNode?: (node: Node, diagnostic: DiagnosticMessage) => void): ResolutionMode | undefined {
    if (!node) return undefined;
    if (length(node.elements) !== 1) {
```

`binder.ts:350:1` (unit fallback):
```ts
function getModuleInstanceStateWorker(node: Node, visited: Map<number, ModuleInstanceState | undefined>): ModuleInstanceState {
    // A module is uninstantiated if it contains only
    switch (node.kind) {
```

`binder.ts:3815:1` (unit fallback):
```ts
export function getContainerFlags(node: Node): ContainerFlags {
    switch (node.kind) {
        case SyntaxKind.ClassExpression:
```

`binder.ts:661:5` (unit fallback):
```ts
    function getDeclarationName(node: Declaration): __String | undefined {
        if (node.kind === SyntaxKind.ExportAssignment) {
            return (node as ExportAssignment).isExportEquals ? InternalSymbolName.ExportEquals : InternalSymbolName.Default;
```

`emitter.ts:752:1` (unit fallback):
```ts
export function emitFiles(
    resolver: EmitResolver,
    host: EmitHost,
```

`emitter.ts:970:5` (unit fallback):
```ts
    function collectLinkedAliases(node: Node): undefined {
        if (isExportAssignment(node)) {
            if (node.expression.kind === SyntaxKind.Identifier) {
```

`emitter.ts:984:5` (unit fallback):
```ts
    function markLinkedReferences(file: SourceFile) {
        if (ts.isSourceFileJS(file)) return; // JS files don't use reference calculations as they don't do import ellision, no need to calculate it
        ts.forEachChildRecursively(file, n => {
```

`transformers/classFields.ts:3343:1` (unit fallback):
```ts
function isReservedPrivateName(node: PrivateIdentifier) {
    return !isGeneratedPrivateIdentifier(node) && node.escapedText === "#constructor";
}
```

`transformers/classFields.ts:3349:1` (unit fallback):
```ts
function isPrivateIdentifierInExpression(node: BinaryExpression): node is PrivateIdentifierInExpression {
    return isPrivateIdentifier(node.left)
        && node.operatorToken.kind === SyntaxKind.InKeyword;
```

`transformers/classFields.ts:3354:1` (unit fallback):
```ts
function isStaticPropertyDeclaration(node: Node): node is PropertyDeclaration {
    return isPropertyDeclaration(node) && hasStaticModifier(node);
}
```

`utilitiesPublic.ts:615:1` (unit fallback):
```ts
export function isParameterPropertyDeclaration(node: Node, parent: Node): node is ParameterPropertyDeclaration {
    return isParameter(node) && hasSyntacticModifier(node, ModifierFlags.ParameterPropertyModifier) && parent.kind === SyntaxKind.Constructor;
}
```

`utilitiesPublic.ts:629:1` (unit fallback):
```ts
export function isEmptyBindingElement(node: BindingElement | ArrayBindingElement): boolean {
    if (isOmittedExpression(node)) {
        return true;
```

`utilitiesPublic.ts:870:1` (unit fallback):
```ts
export function symbolName(symbol: Symbol): string {
    if (symbol.valueDeclaration && isPrivateIdentifierClassElementDeclaration(symbol.valueDeclaration)) {
        return idText(symbol.valueDeclaration.name);
```

`parser.ts:10800:1` (unit fallback):
```ts
export function tagNamesAreEquivalent(lhs: JsxTagNameExpression, rhs: JsxTagNameExpression): boolean {
    if (lhs.kind !== rhs.kind) {
        return false;
```

`parser.ts:3399:5` (unit fallback):
```ts
    function isReusableParameter(node: Node) {
        if (node.kind !== SyntaxKind.Parameter) {
            return false;
```

`parser.ts:3735:5` (unit fallback):
```ts
    function parseLiteralNode(): LiteralExpression {
        return parseLiteralLikeNode(token()) as LiteralExpression;
    }
```

`transformers/es2015.ts:593:5` (unit fallback):
```ts
    function shouldVisitNode(node: Node): boolean {
        return (node.transformFlags & TransformFlags.ContainsES2015) !== 0
            || convertedLoopState !== undefined
```

`transformers/es2015.ts:609:5` (unit fallback):
```ts
    function classWrapperStatementVisitor(node: Node): VisitResult<Node | undefined> {
        if (shouldVisitNode(node)) {
            const original = getOriginalNode(node);
```

`transformers/es2015.ts:633:5` (unit fallback):
```ts
    function visitorWorker(node: Node, expressionResultIsUnused: boolean): VisitResult<Node | undefined> {
        switch (node.kind) {
            case SyntaxKind.StaticKeyword:
```

`transformers/module/system.ts:137:1` (unit fallback):
```ts
export function transformSystemModule(context: TransformationContext): (x: SourceFile | Bundle) => SourceFile | Bundle {
    interface DependencyGroup {
        name: StringLiteral;
```

`transformers/module/system.ts:310:5` (unit fallback):
```ts
    function createSystemModuleBody(node: SourceFile, dependencyGroups: DependencyGroup[]) {
        // Shape of the body in system modules:
        //
```

`transformers/module/system.ts:729:5` (unit fallback):
```ts
    function topLevelVisitor(node: Node): VisitResult<Node | undefined> {
        switch (node.kind) {
            case SyntaxKind.ImportDeclaration:
```

## Pass exits and limits

| Pass | Child exit | Seconds | Stopping reason |
| --- | ---: | ---: | --- |
| apply-add | 0 | 6.017 | completed |
| apply-commit | 0 | 1.006 | completed |
| apply-push | 0 | 3.002 | completed |
| apply-report | 0 | 1.004 | completed |
| apply-snapshot-final | 0 | 1.002 | completed |
| apply-snapshot | 0 | 1.001 | completed |
| apply | 0 | 917.51 | completed |
| census-build | 0 | 473.132 | completed |
| census-final-build | 0 | 1.006 | completed |
| coverage-final-add | 0 | 1.001 | completed |
| coverage-final-audit | 0 | 3.002 | completed |
| coverage-final-checksums | 0 | 1.003 | completed |
| coverage-final-commit | 0 | 1.009 | completed |
| coverage-final-push | 0 | 2.003 | completed |
| coverage-final-report | 0 | 2.001 | completed |
| coverage-final-snapshot-pre | 0 | 2.001 | completed |
| coverage-final-snapshot | 0 | 2.004 | completed |
| coverage-preflight-audit | 0 | 2.002 | completed |
| coverage-preflight-report | 0 | 1.001 | completed |
| coverage-preflight-snapshot | 0 | 2.006 | completed |
| entry-build-fixed | 0 | 33.1 | completed |
| entry-build | 1 | 10.017 | nonzero exit; cause not inferred |
| entry-fixture | 0 | 1.007 | completed |
| export-audit | 0 | 2.003 | completed |
| export-report | 0 | 1.003 | completed |
| final-audit | 0 | 2.01 | completed |
| final-checksums-corrected | 0 | 1.002 | completed |
| final-checksums | 1 | 1.004 | nonzero exit; cause not inferred |
| final-report | 0 | 2.003 | completed |
| final-snapshot-pre | 0 | 1.001 | completed |
| final-snapshot | 0 | 2.007 | completed |
| first-add | 0 | 1.008 | completed |
| first-commit | 0 | 1.007 | completed |
| first-push | 0 | 3.006 | completed |
| first-report | 0 | 1.009 | completed |
| first-snapshot-final | 0 | 1.004 | completed |
| first-snapshot | 0 | 1.005 | completed |
| first | -15 | 1790.787 | configured elapsed bound |
| full-00-add | 0 | 1.003 | completed |
| full-00-commit | 0 | 1.002 | completed |
| full-00-push | 0 | 3.006 | completed |
| full-00-report | 0 | 1.002 | completed |
| full-00-snapshot-final | 0 | 1.001 | completed |
| full-00-snapshot | 0 | 1.001 | completed |
| full-00 | 0 | 24.088 | completed |
| full-01-add | 0 | 1.021 | completed |
| full-01-commit | 0 | 1.009 | completed |
| full-01-push | 0 | 3.015 | completed |
| full-01-report | 0 | 1.01 | completed |
| full-01-snapshot-final | 0 | 1.005 | completed |
| full-01-snapshot | 0 | 1.007 | completed |
| full-01 | -15 | 1791.304 | configured elapsed bound |
| full-02-add | 0 | 1.003 | completed |
| full-02-commit | 0 | 1.002 | completed |
| full-02-push | 0 | 3.006 | completed |
| full-02-report | 0 | 1.005 | completed |
| full-02-snapshot-final | 0 | 1.002 | completed |
| full-02-snapshot | 0 | 1.005 | completed |
| full-02 | 0 | 79.359 | completed |
| full-03-add | 0 | 1.004 | completed |
| full-03-commit | 0 | 1.002 | completed |
| full-03-push | 0 | 2.002 | completed |
| full-03-report | 0 | 1.001 | completed |
| full-03-snapshot-final | 0 | 1.002 | completed |
| full-03-snapshot | 0 | 1.003 | completed |
| full-03 | 0 | 119.289 | completed |
| full-04-add | 0 | 1.003 | completed |
| full-04-commit | 0 | 1.003 | completed |
| full-04-push | 0 | 3.011 | completed |
| full-04-report | 0 | 1.007 | completed |
| full-04-snapshot-final | 0 | 1.002 | completed |
| full-04-snapshot | 0 | 1.002 | completed |
| full-04 | 0 | 177.329 | completed |
| full-05-add | 0 | 1.007 | completed |
| full-05-commit | 0 | 1.005 | completed |
| full-05-push | 0 | 3.004 | completed |
| full-05-report | 0 | 1.006 | completed |
| full-05-snapshot-final | 0 | 1.003 | completed |
| full-05-snapshot | 0 | 1.003 | completed |
| full-05 | 0 | 328.32 | completed |
| full-06-add | 0 | 1.009 | completed |
| full-06-commit | 0 | 1.002 | completed |
| full-06-push | 0 | 3.005 | completed |
| full-06-report | 0 | 1.002 | completed |
| full-06-snapshot-final | 0 | 1.001 | completed |
| full-06-snapshot | 0 | 1.003 | completed |
| full-06 | 0 | 454.859 | completed |
| full-07-add | 0 | 1.003 | completed |
| full-07-commit | 0 | 1.002 | completed |
| full-07-push | 0 | 3.004 | completed |
| full-07-report | 0 | 1.005 | completed |
| full-07-snapshot-final | 0 | 1.002 | completed |
| full-07-snapshot | 0 | 1.003 | completed |
| full-07 | -15 | 571.096 | configured elapsed bound |
| full-20-000-add | 0 | 1.006 | completed |
| full-20-000-audit | 0 | 2.007 | completed |
| full-20-000-commit | 0 | 1.004 | completed |
| full-20-000-push | 0 | 3.007 | completed |
| full-20-000-report | 0 | 1.002 | completed |
| full-20-000-snapshot-final | 0 | 1.002 | completed |
| full-20-000-snapshot | 0 | 1.003 | completed |
| full-20-000 | 0 | 580.765 | completed |
| full-20-001-add | 0 | 1.005 | completed |
| full-20-001-audit | 0 | 1.003 | completed |
| full-20-001-commit | 0 | 1.003 | completed |
| full-20-001-push | 0 | 3.006 | completed |
| full-20-001-report | 0 | 1.021 | completed |
| full-20-001-snapshot-final | 0 | 1.001 | completed |
| full-20-001-snapshot | 0 | 1.002 | completed |
| full-20-001 | 0 | 75.047 | completed |
| full-20-002-add | 0 | 1.002 | completed |
| full-20-002-audit | 0 | 1.002 | completed |
| full-20-002-commit | 0 | 1.005 | completed |
| full-20-002-push | 0 | 3.006 | completed |
| full-20-002-report | 0 | 1.003 | completed |
| full-20-002-snapshot-final | 0 | 1.004 | completed |
| full-20-002-snapshot | 0 | 1.001 | completed |
| full-20-002 | 0 | 18.023 | completed |
| full-20-003-add | 0 | 1.003 | completed |
| full-20-003-audit | 0 | 1.001 | completed |
| full-20-003-commit | 0 | 1.002 | completed |
| full-20-003-push | 0 | 3.005 | completed |
| full-20-003-report | 0 | 1.001 | completed |
| full-20-003-snapshot-final | 0 | 1.023 | completed |
| full-20-003-snapshot | 0 | 1.002 | completed |
| full-20-003 | 0 | 42.029 | completed |
| full-20-004-add | 0 | 1.003 | completed |
| full-20-004-audit | 0 | 1.001 | completed |
| full-20-004-commit | 0 | 1.024 | completed |
| full-20-004-push | 0 | 3.006 | completed |
| full-20-004-report | 0 | 1.001 | completed |
| full-20-004-snapshot-final | 0 | 1.004 | completed |
| full-20-004-snapshot | 0 | 1.001 | completed |
| full-20-004 | 0 | 17.014 | completed |
| full-20-005-add | 0 | 1.004 | completed |
| full-20-005-audit | 0 | 2.004 | completed |
| full-20-005-commit | 0 | 1.004 | completed |
| full-20-005-push | 0 | 2.008 | completed |
| full-20-005-report | 0 | 1.004 | completed |
| full-20-005-snapshot-final | 0 | 1.002 | completed |
| full-20-005-snapshot | 0 | 1.001 | completed |
| full-20-005 | 0 | 41.043 | completed |
| full-20-006-add | 0 | 1.003 | completed |
| full-20-006-audit | 0 | 1.003 | completed |
| full-20-006-commit | 0 | 1.003 | completed |
| full-20-006-push | 0 | 3.004 | completed |
| full-20-006-report | 0 | 1.005 | completed |
| full-20-006-snapshot-final | 0 | 1.004 | completed |
| full-20-006-snapshot | 0 | 1.002 | completed |
| full-20-006 | 0 | 71.166 | completed |
| full-20-007-add | 0 | 1.005 | completed |
| full-20-007-audit | 0 | 2.029 | completed |
| full-20-007-commit | 0 | 1.005 | completed |
| full-20-007-push | 0 | 3.009 | completed |
| full-20-007-report | 0 | 1.002 | completed |
| full-20-007-snapshot-final | 0 | 1.002 | completed |
| full-20-007-snapshot | 0 | 1.002 | completed |
| full-20-007 | 0 | 86.22 | completed |
| full-20-008-add | 0 | 1.005 | completed |
| full-20-008-audit | 0 | 2.011 | completed |
| full-20-008-commit | 0 | 1.003 | completed |
| full-20-008-push | 0 | 3.011 | completed |
| full-20-008-report | 0 | 2.01 | completed |
| full-20-008-snapshot-final | 0 | 1.001 | completed |
| full-20-008-snapshot | 0 | 2.006 | completed |
| full-20-008 | 0 | 524.788 | completed |
| full-20-009-add | 0 | 1.004 | completed |
| full-20-009-audit | 0 | 2.005 | completed |
| full-20-009-commit | 0 | 1.004 | completed |
| full-20-009-push | 0 | 3.004 | completed |
| full-20-009-report | 0 | 2.004 | completed |
| full-20-009-snapshot-final | 0 | 2.008 | completed |
| full-20-009-snapshot | 0 | 2.002 | completed |
| full-20-009 | 0 | 68.568 | completed |
| full-20-010-add | 0 | 1.004 | completed |
| full-20-010-audit | 0 | 2.003 | completed |
| full-20-010-commit | 0 | 1.006 | completed |
| full-20-010-push | 0 | 3.012 | completed |
| full-20-010-report | 0 | 2.007 | completed |
| full-20-010-snapshot-final | 0 | 2.004 | completed |
| full-20-010-snapshot | 0 | 2.003 | completed |
| full-20-010 | 0 | 70.571 | completed |
| full-20-011-add | 0 | 1.006 | completed |
| full-20-011-audit | 0 | 3.004 | completed |
| full-20-011-commit | 0 | 1.001 | completed |
| full-20-011-push | 0 | 2.001 | completed |
| full-20-011-report | 0 | 2.005 | completed |
| full-20-011-snapshot-final | 0 | 1.003 | completed |
| full-20-011-snapshot | 0 | 2.006 | completed |
| full-20-011 | 0 | 54.11 | completed |
| full-20-012-add | 0 | 1.002 | completed |
| full-20-012-audit | 0 | 2.003 | completed |
| full-20-012-commit | 0 | 1.001 | completed |
| full-20-012-push | 0 | 3.007 | completed |
| full-20-012-report | 0 | 2.002 | completed |
| full-20-012-snapshot-final | 0 | 1.004 | completed |
| full-20-012-snapshot | 0 | 1.003 | completed |
| full-20-012 | 0 | 117.241 | completed |
| full-20-013-add | 0 | 1.002 | completed |
| full-20-013-audit | 0 | 2.009 | completed |
| full-20-013-commit | 0 | 1.006 | completed |
| full-20-013-push | 0 | 3.005 | completed |
| full-20-013-report | 0 | 2.002 | completed |
| full-20-013-snapshot-final | 0 | 1.003 | completed |
| full-20-013-snapshot | 0 | 2.003 | completed |
| full-20-013 | 0 | 42.068 | completed |
| full-20-014-add | 0 | 1.006 | completed |
| full-20-014-audit | 0 | 2.002 | completed |
| full-20-014-commit | 0 | 1.005 | completed |
| full-20-014-push | 0 | 3.006 | completed |
| full-20-014-report | 0 | 2.007 | completed |
| full-20-014-snapshot-final | 0 | 2.004 | completed |
| full-20-014-snapshot | 0 | 1.001 | completed |
| full-20-014 | 0 | 38.079 | completed |
| full-20-015-add | 0 | 1.005 | completed |
| full-20-015-audit | 0 | 3.035 | completed |
| full-20-015-commit | 0 | 1.005 | completed |
| full-20-015-push | 0 | 3.017 | completed |
| full-20-015-report | 0 | 2.003 | completed |
| full-20-015-snapshot-final | 0 | 2.002 | completed |
| full-20-015-snapshot | 0 | 2.002 | completed |
| full-20-015 | 0 | 197.235 | completed |
| full-20-016-add | 0 | 1.003 | completed |
| full-20-016-audit | 0 | 3.024 | completed |
| full-20-016-commit | 0 | 1.002 | completed |
| full-20-016-push | 0 | 3.006 | completed |
| full-20-016-report | 0 | 2.006 | completed |
| full-20-016-snapshot-final | 0 | 1.001 | completed |
| full-20-016-snapshot | 0 | 2.007 | completed |
| full-20-016 | 0 | 31.135 | completed |
| full-20-017-add | 0 | 1.007 | completed |
| full-20-017-audit | 0 | 2.006 | completed |
| full-20-017-commit | 0 | 1.003 | completed |
| full-20-017-push | 0 | 3.004 | completed |
| full-20-017-report | 0 | 2.01 | completed |
| full-20-017-snapshot-final | 0 | 1.002 | completed |
| full-20-017-snapshot | 0 | 2.006 | completed |
| full-20-017 | 0 | 133.257 | completed |
| full-20-018-add | 0 | 1.004 | completed |
| full-20-018-audit | 0 | 2.003 | completed |
| full-20-018-commit | 0 | 1.001 | completed |
| full-20-018-push | 0 | 3.002 | completed |
| full-20-018-report | 0 | 2.004 | completed |
| full-20-018-snapshot-final | 0 | 2.002 | completed |
| full-20-018-snapshot | 0 | 3.003 | completed |
| full-20-018 | 0 | 80.1 | completed |
| full-20-019-add | 0 | 1.002 | completed |
| full-20-019-audit | 0 | 2.004 | completed |
| full-20-019-commit | 0 | 1.001 | completed |
| full-20-019-push | 0 | 3.005 | completed |
| full-20-019-report | 0 | 2.001 | completed |
| full-20-019-snapshot-final | 0 | 2.001 | completed |
| full-20-019-snapshot | 0 | 2.005 | completed |
| full-20-019 | 0 | 70.1 | completed |
| full-20-020-add | 0 | 1.009 | completed |
| full-20-020-audit | 0 | 2.011 | completed |
| full-20-020-commit | 0 | 1.004 | completed |
| full-20-020-push | 0 | 3.004 | completed |
| full-20-020-report | 0 | 2.008 | completed |
| full-20-020-snapshot-final | 0 | 1.003 | completed |
| full-20-020-snapshot | 0 | 2.006 | completed |
| full-20-020 | 0 | 253.583 | completed |
| full-20-021-add | 0 | 1.002 | completed |
| full-20-021-audit | 0 | 2.002 | completed |
| full-20-021-commit | 0 | 1.008 | completed |
| full-20-021-push | 0 | 3.006 | completed |
| full-20-021-report | 0 | 2.001 | completed |
| full-20-021-snapshot-final | 0 | 1.005 | completed |
| full-20-021-snapshot | 0 | 1.003 | completed |
| full-20-021 | 0 | 48.111 | completed |
| full-20-022-add | 0 | 1.004 | completed |
| full-20-022-audit | 0 | 2.003 | completed |
| full-20-022-commit | 0 | 1.004 | completed |
| full-20-022-push | 0 | 3.007 | completed |
| full-20-022-report | 0 | 2.002 | completed |
| full-20-022-snapshot-final | 0 | 2.001 | completed |
| full-20-022-snapshot | 0 | 1.001 | completed |
| full-20-022 | 0 | 80.193 | completed |
| full-20-023-add | 0 | 1.007 | completed |
| full-20-023-audit | 0 | 2.004 | completed |
| full-20-023-commit | 0 | 1.003 | completed |
| full-20-023-push | 0 | 3.007 | completed |
| full-20-023-report | 0 | 2.01 | completed |
| full-20-023-snapshot-final | 0 | 2.006 | completed |
| full-20-023-snapshot | 0 | 2.005 | completed |
| full-20-023 | 0 | 98.189 | completed |
| full-20-024-add | 0 | 1.001 | completed |
| full-20-024-audit | 0 | 3.008 | completed |
| full-20-024-commit | 0 | 1.002 | completed |
| full-20-024-push | 0 | 2.003 | completed |
| full-20-024-report | 0 | 2.002 | completed |
| full-20-024-snapshot-final | 0 | 2.001 | completed |
| full-20-024-snapshot | 0 | 2.002 | completed |
| full-20-024 | 0 | 37.102 | completed |
| full-20-025-add | 0 | 1.006 | completed |
| full-20-025-audit | 0 | 3.01 | completed |
| full-20-025-commit | 0 | 1.004 | completed |
| full-20-025-push | 0 | 3.005 | completed |
| full-20-025-report | 0 | 2.006 | completed |
| full-20-025-snapshot-final | 0 | 2.009 | completed |
| full-20-025-snapshot | 0 | 1.002 | completed |
| full-20-025 | 0 | 79.186 | completed |
| full-20-026-add | 0 | 1.011 | completed |
| full-20-026-audit | 0 | 3.006 | completed |
| full-20-026-commit | 0 | 1.009 | completed |
| full-20-026-push | 0 | 3.008 | completed |
| full-20-026-report | 0 | 2.023 | completed |
| full-20-026-snapshot-final | 0 | 2.003 | completed |
| full-20-026-snapshot | 0 | 2.006 | completed |
| full-20-026 | 0 | 42.112 | completed |
| full-20-027-add | 0 | 1.005 | completed |
| full-20-027-audit | 0 | 2.004 | completed |
| full-20-027-commit | 0 | 1.001 | completed |
| full-20-027-push | 0 | 3.005 | completed |
| full-20-027-report | 0 | 2.001 | completed |
| full-20-027-snapshot-final | 0 | 1.006 | completed |
| full-20-027-snapshot | 0 | 1.001 | completed |
| full-20-027 | 0 | 148.3 | completed |
| full-20-028-add | 0 | 1.003 | completed |
| full-20-028-audit | 0 | 4.034 | completed |
| full-20-028-commit | 0 | 1.004 | completed |
| full-20-028-push | 0 | 3.003 | completed |
| full-20-028-report | 0 | 3.005 | completed |
| full-20-028-snapshot-final | 0 | 2.005 | completed |
| full-20-028-snapshot | 0 | 2.01 | completed |
| full-20-028 | 0 | 97.184 | completed |
| full-20-029-add | 0 | 1.002 | completed |
| full-20-029-audit | 0 | 2.055 | completed |
| full-20-029-commit | 0 | 1.002 | completed |
| full-20-029-push | 0 | 2.01 | completed |
| full-20-029-report | 0 | 2.006 | completed |
| full-20-029-snapshot-final | 0 | 1.002 | completed |
| full-20-029-snapshot | 0 | 2.006 | completed |
| full-20-029 | 0 | 29.053 | completed |
| full-20-030-add | 0 | 1.006 | completed |
| full-20-030-audit | 0 | 2.004 | completed |
| full-20-030-commit | 0 | 1.003 | completed |
| full-20-030-push | 0 | 3.004 | completed |
| full-20-030-report | 0 | 2.011 | completed |
| full-20-030-snapshot-final | 0 | 2.003 | completed |
| full-20-030-snapshot | 0 | 2.005 | completed |
| full-20-030 | 0 | 28.044 | completed |
| full-20-031-add | 0 | 1.002 | completed |
| full-20-031-audit | 0 | 3.006 | completed |
| full-20-031-commit | 0 | 1.002 | completed |
| full-20-031-push | 0 | 3.002 | completed |
| full-20-031-report | 0 | 2.002 | completed |
| full-20-031-snapshot-final | 0 | 2.005 | completed |
| full-20-031-snapshot | 0 | 2.003 | completed |
| full-20-031 | 0 | 39.091 | completed |
| full-20-032-add | 0 | 1.003 | completed |
| full-20-032-audit | 0 | 2.002 | completed |
| full-20-032-commit | 0 | 1.003 | completed |
| full-20-032-push | 0 | 3.006 | completed |
| full-20-032-report | 0 | 2.002 | completed |
| full-20-032-snapshot-final | 0 | 2.004 | completed |
| full-20-032-snapshot | 0 | 2.004 | completed |
| full-20-032 | 0 | 239.423 | completed |
| full-20-033-add | 0 | 1.004 | completed |
| full-20-033-audit | 0 | 2.003 | completed |
| full-20-033-commit | 0 | 1.003 | completed |
| full-20-033-push | 0 | 3.005 | completed |
| full-20-033-report | 0 | 2.005 | completed |
| full-20-033-snapshot-final | 0 | 1.001 | completed |
| full-20-033-snapshot | 0 | 2.005 | completed |
| full-20-033 | 0 | 139.22 | completed |
| full-20-034-add | 0 | 1.006 | completed |
| full-20-034-audit | 0 | 2.004 | completed |
| full-20-034-commit | 0 | 1.004 | completed |
| full-20-034-push | 0 | 3.006 | completed |
| full-20-034-report | 0 | 2.007 | completed |
| full-20-034-snapshot-final | 0 | 1.003 | completed |
| full-20-034-snapshot | 0 | 2.002 | completed |
| full-20-034 | 0 | 24.023 | completed |
| full-20-035-add | 0 | 1.005 | completed |
| full-20-035-audit | 0 | 3.005 | completed |
| full-20-035-commit | 0 | 1.013 | completed |
| full-20-035-push | 0 | 3.004 | completed |
| full-20-035-report | 0 | 2.008 | completed |
| full-20-035-snapshot-final | 0 | 2.008 | completed |
| full-20-035-snapshot | 0 | 2.004 | completed |
| full-20-035 | 0 | 175.345 | completed |
| full-20-036-add | 0 | 1.005 | completed |
| full-20-036-audit | 0 | 3.004 | completed |
| full-20-036-commit | 0 | 1.002 | completed |
| full-20-036-push | 0 | 3.006 | completed |
| full-20-036-report | 0 | 2.004 | completed |
| full-20-036-snapshot-final | 0 | 2.003 | completed |
| full-20-036-snapshot | 0 | 2.036 | completed |
| full-20-036 | 0 | 120.255 | completed |
| full-20-037-add | 0 | 1.005 | completed |
| full-20-037-audit | 0 | 2.008 | completed |
| full-20-037-commit | 0 | 1.004 | completed |
| full-20-037-push | 0 | 3.004 | completed |
| full-20-037-report | 0 | 2.007 | completed |
| full-20-037-snapshot-final | 0 | 2.001 | completed |
| full-20-037-snapshot | 0 | 2.009 | completed |
| full-20-037 | 0 | 131.314 | completed |
| full-20-038-add | 0 | 1.002 | completed |
| full-20-038-audit | 0 | 3.009 | completed |
| full-20-038-commit | 0 | 1.013 | completed |
| full-20-038-push | 0 | 3.009 | completed |
| full-20-038-report | 0 | 3.012 | completed |
| full-20-038-snapshot-final | 0 | 2.001 | completed |
| full-20-038-snapshot | 0 | 2.002 | completed |
| full-20-038 | 0 | 107.377 | completed |
| full-20-039-add | 0 | 1.007 | completed |
| full-20-039-audit | 0 | 3.004 | completed |
| full-20-039-commit | 0 | 1.002 | completed |
| full-20-039-push | 0 | 3.01 | completed |
| full-20-039-report | 0 | 2.002 | completed |
| full-20-039-snapshot-final | 0 | 2.003 | completed |
| full-20-039-snapshot | 0 | 2.004 | completed |
| full-20-039 | 0 | 40.167 | completed |
| full-20-040-add | 0 | 1.003 | completed |
| full-20-040-audit | 0 | 3.004 | completed |
| full-20-040-commit | 0 | 1.004 | completed |
| full-20-040-push | 0 | 3.006 | completed |
| full-20-040-report | 0 | 2.006 | completed |
| full-20-040-snapshot-final | 0 | 2.004 | completed |
| full-20-040-snapshot | 0 | 2.006 | completed |
| full-20-040 | 0 | 127.238 | completed |
| full-20-041-add | 0 | 1.003 | completed |
| full-20-041-audit | 0 | 3.007 | completed |
| full-20-041-commit | 0 | 1.004 | completed |
| full-20-041-push | 0 | 3.008 | completed |
| full-20-041-report | 0 | 2.002 | completed |
| full-20-041-snapshot-final | 0 | 2.001 | completed |
| full-20-041-snapshot | 0 | 2.005 | completed |
| full-20-041 | 0 | 172.409 | completed |
| full-20-042-add | 0 | 1.003 | completed |
| full-20-042-audit | 0 | 3.003 | completed |
| full-20-042-commit | 0 | 1.002 | completed |
| full-20-042-push | 0 | 3.002 | completed |
| full-20-042-report | 0 | 2.002 | completed |
| full-20-042-snapshot-final | 0 | 2.003 | completed |
| full-20-042-snapshot | 0 | 2.003 | completed |
| full-20-042 | 0 | 43.175 | completed |
| full-20-043-add | 0 | 1.002 | completed |
| full-20-043-audit | 0 | 3.003 | completed |
| full-20-043-commit | 0 | 1.005 | completed |
| full-20-043-push | 0 | 2.003 | completed |
| full-20-043-report | 0 | 2.001 | completed |
| full-20-043-snapshot-final | 0 | 2.004 | completed |
| full-20-043-snapshot | 0 | 2.006 | completed |
| full-20-043 | 0 | 33.135 | completed |
| full-20-044-add | 0 | 1.003 | completed |
| full-20-044-audit | 0 | 3.014 | completed |
| full-20-044-commit | 0 | 1.004 | completed |
| full-20-044-push | 0 | 3.007 | completed |
| full-20-044-report | 0 | 2.011 | completed |
| full-20-044-snapshot-final | 0 | 2.006 | completed |
| full-20-044-snapshot | 0 | 2.01 | completed |
| full-20-044 | 0 | 190.794 | completed |
| full-20-045-add | 0 | 1.004 | completed |
| full-20-045-audit | 0 | 2.002 | completed |
| full-20-045-commit | 0 | 1.003 | completed |
| full-20-045-push | 0 | 3.006 | completed |
| full-20-045-report | 0 | 2.002 | completed |
| full-20-045-snapshot-final | 0 | 2.001 | completed |
| full-20-045-snapshot | 0 | 2.005 | completed |
| full-20-045 | 0 | 78.131 | completed |
| full-20-046-add | 0 | 1.001 | completed |
| full-20-046-audit | 0 | 3.008 | completed |
| full-20-046-commit | 0 | 1.005 | completed |
| full-20-046-push | 0 | 3.038 | completed |
| full-20-046-report | 0 | 2.015 | completed |
| full-20-046-snapshot-final | 0 | 2.01 | completed |
| full-20-046-snapshot | 0 | 2.004 | completed |
| full-20-046 | 0 | 62.113 | completed |
| full-20-047-add | 0 | 1.008 | completed |
| full-20-047-audit | 0 | 3.006 | completed |
| full-20-047-commit | 0 | 1.003 | completed |
| full-20-047-push | 0 | 3.008 | completed |
| full-20-047-report | 0 | 2.004 | completed |
| full-20-047-snapshot-final | 0 | 2.004 | completed |
| full-20-047-snapshot | 0 | 2.002 | completed |
| full-20-047 | 0 | 252.37 | completed |
| full-20-048-add | 0 | 1.004 | completed |
| full-20-048-audit | 0 | 3.004 | completed |
| full-20-048-commit | 0 | 1.005 | completed |
| full-20-048-push | 0 | 3.006 | completed |
| full-20-048-report | 0 | 2.006 | completed |
| full-20-048-snapshot-final | 0 | 2.004 | completed |
| full-20-048-snapshot | 0 | 3.004 | completed |
| full-20-048 | 0 | 656.156 | completed |
| full-20-049-add | 0 | 1.004 | completed |
| full-20-049-audit | 0 | 3.01 | completed |
| full-20-049-commit | 0 | 1.004 | completed |
| full-20-049-push | 0 | 2.003 | completed |
| full-20-049-report | 0 | 3.005 | completed |
| full-20-049-snapshot-final | 0 | 2.003 | completed |
| full-20-049-snapshot | 0 | 2.008 | completed |
| full-20-049 | 0 | 58.124 | completed |
| full-20-050-add | 0 | 1.001 | completed |
| full-20-050-audit | 0 | 3.002 | completed |
| full-20-050-commit | 0 | 1.004 | completed |
| full-20-050-push | 0 | 3.003 | completed |
| full-20-050-report | 0 | 4.002 | completed |
| full-20-050-snapshot-final | 0 | 2.019 | completed |
| full-20-050-snapshot | 0 | 2.009 | completed |
| full-20-050 | 0 | 36.063 | completed |
| full-20-051-add | 0 | 1.001 | completed |
| full-20-051-audit | 0 | 3.01 | completed |
| full-20-051-commit | 0 | 1.002 | completed |
| full-20-051-push | 0 | 3.008 | completed |
| full-20-051-report | 0 | 2.01 | completed |
| full-20-051-snapshot-final | 0 | 2.004 | completed |
| full-20-051-snapshot | 0 | 2.009 | completed |
| full-20-051 | 0 | 25.078 | completed |
| instrument-entry-fixed | 0 | 1.001 | completed |
| instrument-entry | 0 | 1.002 | completed |
| interim-audit | 0 | 3.002 | completed |
| interim-report | 0 | 1.002 | completed |
| npm-ci | 0 | not timed | completed |
| overlay-entry-fixed | 0 | 2.011 | completed |
| overlay-entry | 0 | 3.006 | completed |
| overlay | 0 | 6.029 | completed |
| publish-add | 0 | 1.001 | completed |
| publish-branch | 0 | 1.001 | completed |
| publish-checkpoint-commit | 0 | 1.001 | completed |
| publish-checkpoint-push | 0 | 3.007 | completed |
| publish-commit | 0 | 1.002 | completed |
| publish-count | 0 | 1.002 | completed |
| publish-final-scope | 0 | 1.002 | completed |
| publish-parent | 0 | 1.003 | completed |
| publish-push | 0 | 4.005 | completed |
| publish-remote | 0 | 1.002 | completed |
| publish-scope | 0 | 1.001 | completed |
| publish-staged-scope | 0 | 1.001 | completed |
| publish-switch | 0 | 1.002 | completed |
| publish-tree | 0 | 1.003 | completed |
| record-final-audit | 0 | 3.002 | completed |
| record-final-grouping | 0 | 2.004 | completed |
| record-final-report | 0 | 4.004 | completed |
| record-final-snapshot-pre | 0 | 3.002 | completed |
| record-recovery-00 | 0 | 26.079 | completed |
| record-recovery-01 | 0 | 51.767 | completed |
| record-recovery-02 | 0 | 21.033 | completed |
| record-recovery-03 | 0 | 32.029 | completed |
| record-recovery-04 | 0 | 24.015 | completed |
| record-recovery-05 | 0 | 1336.728 | completed |
| record-recovery-06 | -15 | 535.69 | explicit stop: first-error path ignores LATENT_ONLY_FILE and scans unrelated files; retry missing payloads only with scratch file-selection filter |
| record-recovery-07 | 0 | 284.321 | completed |
| record-recovery-08 | 0 | 143.6 | completed |
| record-recovery-build | 0 | 26.039 | completed |
| record-recovery-file-filter-build | 0 | 35.106 | completed |
| record-source-inspection-checker-complete | 1 | 7.01 | nonzero exit; cause not inferred |
| record-source-inspection-checker-corrected | 0 | 17.654 | completed |
| record-source-inspection-corrected | 0 | 7.013 | completed |
| record-source-inspection-final | 0 | 9.014 | completed |
| record-source-inspection-literal-context | 0 | 18.089 | completed |
| record-source-inspection-recovery | 0 | 14.07 | completed |
| record-source-inspection | 0 | 9.01 | completed |
| record-source-inspection-string-keys | 0 | 13.043 | completed |
| selftest | 0 | 1.028 | completed |
| setup | 0 | 758.331 | completed |
| source | 0 | 12.004 | completed |
| verify-apply | 0 | 1.003 | completed |

## Method and reproduction

Read `graph_types.go:460` and its callers at `graph_types.go:355` and `program_region.go:140`. Count ViewOrigins entries plus CheckedFields keys without deduplicating their overlap; walk Main and every function body with the boundary's own walk for record ObjectLiteral, both dynamic-read nodes, and checked ObjectCall. These are hypothetical boundary candidates before the rest of ordinary lowering, not demonstrated graph-program refusals.

The entry resolves 81 implementation files: the 79 compiler files from the independent-root report plus `src/tsc/tsc.ts` and `src/tsc/_namespaces/ts.ts`. The sums cover this resolved entry program. `instrument.py` uses the meter entry overlay to keep the entry first and expand Files to its implementation dependencies. It removes only the meter driver's full-mode-only checker-rejection gate so the requested first-error latent pass can retain diagnostics and proceed. `instrument.py` extends the existing guarded latent overlay in scratch only. It retains successful top-level statement returns in the measurement result's Main, which the stock hook discards. Ordinary Load and Lower remain disabled and every real pass asserts those guards. `overlay/` contains all generated Go inputs as `.go.txt`; no production lower source is changed. The deterministic self-test exercises Main, a function body, both dynamic reads, both view sources, GraphTypes presence, and failed-unit exclusion.

An initial entry build omitted the observer mapping and failed; the corrected build and entry fixture passed before the real census. Both attempts are retained. The flattened selected-site stream is [`sites.jsonl.gz`](sites.jsonl.gz), beside this report; each line includes kind, reported position, unit, owner file, GraphTypes flag and original stream. It is derived without changing the preserved raw pass streams. Raw boundary streams, original latent streams, timing streams, logs and each child's own exit status are compressed separately beside this report in `raw/`. `source-manifest.json`, `provenance.json` and `SHA256SUMS` retain source and binary identities. Piece selections and measurement environment variables are retained in raw JSON. The initial full pieces used Go's `GOMEMLIMIT=10GiB` soft memory limit and a 60-minute aggregate budget, with at most 1790 seconds per child. Continuation pieces select one file each, use `GOMEMLIMIT=4GiB` and at most 1490 seconds per child. Priority and files larger than 300000 bytes run alone; at most two smaller files run concurrently. Only completed child streams enter a checkpoint. The continuation has no aggregate 60-minute cutoff. Each process group is killed and reaped by the supervisor on completion or stopping.

```sh
source /workspace/adamic-tools/env.sh
bash cloud/setup.sh > setup.log 2>&1
npm ci --prefix stage3/api > npm-ci.log 2>&1
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source
bash stage3/apply.sh /workspace/graph-boundary-adapted > apply.log 2>&1
python3 stage3/census/latent/make_overlay.py /workspace/graph-boundary /workspace/scratch/jawqwv7/overlay
python3 stage3/census/graph-boundary-bcbe5d12/instrument.py /workspace/graph-boundary /workspace/scratch/jawqwv7/overlay
go build -overlay /workspace/scratch/jawqwv7/overlay/overlay.json -o /workspace/scratch/jawqwv7/census ./stage3/census/latent/tool
# run-pieces.py records the exact entry invocation and file selections
```

## Record-shaped findings among stopped units

All 3,666 stopped unit results are covered: 2,762 already had finding payloads; 904 payloads missing from interrupted passes were recovered. The inherited latent driver buffers file records until pass completion, while the boundary observer streams each unit. Available finding payloads were not rerun; missing-payload recovery does not replace any four-way IR census row. Each recovery uses only the originally observed units, the same mode, unchanged registration and lowering, and an observer that flushes findings per unit. Every recovered boundary status, findings-added count, GraphTypes flag and retained IR site list matches its original row. Recovery children are capped at 1490 seconds and their own exits are preserved.

**Record-shaped totals: 85 finding observations, 58 distinct finding locations, in 18 exact-reason groups and 49 stopped units. Of these, 78 findings at 53 locations are causal failures; the rest are bookkeeping in record-shaped context. 31 units have a record-shaped first stop; 18 have one only behind another stop.** Excluding bookkeeping, the causal split is 31 first-stop units and 15 only-behind units. 3 additional units have record-shaped context only in bookkeeping, with no verified record-shaped cause. 3,620 units have no verified record-shaped causal finding.

Grouped 24,806 emitted findings into 1235 exact reason texts, including 12,941 Boundary bookkeeping entries. These wrappers describe rollback or continuation and do not add an independent cause; 7 have source record context. They are included in overall record-shaped finding totals when the source context qualifies, but excluded from causal totals and first-stop ordering. Refusal-scan findings are excluded because they are file-level diagnostics, not the findings that stopped these unit results.

Findings without a source location: 1. They remain in reason counts and contribute no distinct site; classification there can use explicit reason evidence only.

Selection and ordering follow the existing census: one observation per (owner file, unit), latest full mode preferred, otherwise first error. First means first causal finding in emission order, not the smallest source line. The overall only-behind count includes marked bookkeeping that follows a non-record first stop; the separate causal only-behind count excludes it. Repeated dependency failures in different units count separately. Distinct sites use the finding’s actual file:line:column, unlike the four-way IR table’s unit fallbacks. Top files below use the finding location’s file, not its owner unit; both are retained in the observation stream.

Classification uses exact reason text and source inspection with pinned TypeScript 6.0.3. Record/dictionary/index-signature reasons are marked. Generic reasons are marked only at sites where the source checker verifies an index-signature object literal (including a member initializer, without crossing a function boundary) or a string-keyed operation on such an object. CompilerOptions, WatchOptions and TypeAcquisition declare string index signatures in types.ts. A mixed reason does not make all its occurrences record-shaped. Source context establishes a record-shaped operation, not that record support alone would fix it. Arrays, tuples, primitives, Maps and Sets do not qualify by indexing alone. Uncertain cases remain unmarked; each group's decision is in the table and JSON.

Coverage: all stopped-unit payloads, not all source programs or failed expressions. Failed statements roll back; later statements may be visited but expressions behind an internal first failure can remain hidden. Deferred generics, checker skips and interrupted unit traversal remain the same gaps described above. These findings do not imply GraphTypes membership or native lowering success. These totals describe failed-lowering evidence; the four-way table continues to count retained IR only.

Machine-readable exact groups, decisions, locations, top files and examples: [record-findings.json](record-findings.json). Per-observation classification and provenance: [record-findings-observations.jsonl.gz](record-findings-observations.jsonl.gz). No reason text is normalized; Markdown escapes only table syntax.

### Every exact finding reason

| Exact reason | Findings | Distinct locations | Record-shaped findings | Record sites | Decision |
| --- | ---: | ---: | ---: | ---: | --- |
| failed statement; state rolled back; continued at next statement or declaration | 12622 | 8151 | 7 | 6 | some: Continuation bookkeeping, not an independent failure cause; marked only when the source checker verifies record-shaped context. |
| a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1063 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| reading node | 1042 | 1028 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| writing a checked object field without its source-slot type certificate | 990 | 44 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a cast the runtime can't check | 570 | 426 | 2 | 2 | some: Source checker verifies string-keyed read/write on an index-signature object. Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type T seen as Mutable&lt;T&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of T["pos"] would replace | 433 | 4 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate membership over an open tag domain | 321 | 279 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| function signature or prologue failed; continued at next declaration | 319 | 319 | 0 | 0 | unmarked: Continuation bookkeeping, not an independent failure cause; marked only when the source checker verifies record-shaped context. |
| a call through ?. (an optional call) | 236 | 149 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| reading result | 182 | 109 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| overload 1 of visitNode parameter node cannot be served by implementation parameter node | 181 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type T["kind"] | 180 | 5 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 2 of visitEachChild parameter nodesVisitor cannot be served by implementation parameter nodesVisitor | 175 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a Map or Set interface implemented by a value that isn't a Map or Set | 141 | 120 | 0 | 0 | unmarked: Unmarked: Map/Set representation alone is not a dictionary or record object; no verified record source context. Uncertain cases remain unmarked. |
| a method sharing its name with a getter; rename the getter or use a callable arrow field in the interface and implementation | 118 | 100 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a narrowed union member whose object tag cannot be checked with typeof; keep differently held object kinds in separately typed variables | 110 | 30 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| assigning to an Identifier | 107 | 81 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an optional Map get through a structural receiver | 106 | 44 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of visitNodes parameter visitor cannot be served by implementation parameter visitor | 89 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| reading type | 87 | 87 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| a call spreading elements that cannot be packed | 85 | 37 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| reading expression | 83 | 75 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| a value of type any | 82 | 54 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| assigning a field of a value | 79 | 14 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| var | 73 | 63 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions | 66 | 16 | 0 | 0 | unmarked: Unmarked: Map/Set representation alone is not a dictionary or record object; no verified record source context. Uncertain cases remain unmarked. |
| overload 1 of filter result U[] cannot be served by implementation result readonly T[] \| undefined | 61 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| reading updated | 61 | 47 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading modifiers | 58 | 43 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| an assignment value to a member | 57 | 52 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type HasJSDoc | 55 | 4 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type JSDocArray \| undefined | 54 | 4 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| reading jsDoc | 52 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| a boolean \| undefined variable a function value captures | 51 | 35 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a union of differently held members variable a function value captures | 51 | 20 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type never | 49 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a function returning U \| undefined | 46 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a function returning T \| undefined | 42 | 9 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a generic instantiated at void whose body reads the value: emitNodeList's Child = void (read from the signature), read at /workspace/graph-boundary-adapted/src/compiler/emitter.ts:4679:270 as a condition | 42 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| reading name | 40 | 40 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading sourceFilePath | 40 | 10 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading visited | 38 | 24 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| checked predicate overload target Expression | 37 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 3 of getGlobalType result GenericType cannot be served by implementation result ObjectType \| undefined | 37 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of arrayFrom result U[] cannot be served by implementation result (T \| U)[] | 35 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of concatenate result T[] cannot be served by implementation result readonly T[] \| undefined | 35 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| reading array | 35 | 7 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| a case whose type differs from the switch's | 34 | 16 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an array going where a plain object type is held | 33 | 28 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a SpreadElement | 31 | 10 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an array of any | 31 | 15 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a dictionary member narrowed to an array without an element-checked view | 30 | 19 | 30 | 19 | all: Reason explicitly describes a record, dictionary or index-signature object. |
| a checked predicate target whose check would read past the depth bound (TupleTypeReference) | 29 | 3 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| reading statement | 28 | 27 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading target | 28 | 22 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| checked predicate overload target Node | 27 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| reading file | 27 | 10 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| an unproven relation from Node \| undefined to T \| undefined: the source is not assignable to the target | 25 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| reading host | 25 | 25 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading res | 25 | 5 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| a function value that captures the variable its own initializer declares | 24 | 6 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| new an Identifier | 24 | 20 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of sameMap result U[] cannot be served by implementation result readonly U[] \| undefined | 24 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| reading refPath | 24 | 7 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| overload 1 of reduceLeft parameter f cannot be served by implementation parameter f | 22 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of skipParentheses result Expression cannot be served by implementation result Node | 22 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| reading left | 22 | 12 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading statements | 22 | 15 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| a comparator that isn't a function | 21 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| reading diagnostics | 21 | 21 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| replacing a represented method at runtime | 21 | 18 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a branded supertype inside a checked predicate's array target | 20 | 17 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| in that narrows a union of object types (any object with the key passes) | 20 | 7 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| optional chaining to .size on a value | 20 | 16 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| reading resolutions | 20 | 5 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| returning a property or element assignment | 20 | 20 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| writing a possibly absent optional own field | 20 | 17 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type AnonymousFunctionDefinition | 19 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | 19 | 12 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of parseOptionalToken result Token&lt;TKind&gt; cannot be served by implementation result Node \| undefined | 19 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of skipPartiallyEmittedExpressions result Expression cannot be served by implementation result Node | 19 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| reading clone | 19 | 16 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading fn | 19 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| a checked predicate result read by an ArrowFunction inside a condition, which its direction walk doesn't read | 18 | 15 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a destructured name that isn't plain | 18 | 4 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a literal method through a view that erases its receiver | 18 | 13 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of visitParameterList parameter nodesVisitor cannot be served by implementation parameter nodesVisitor | 18 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| reading objectFlags | 18 | 10 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading resolution | 18 | 6 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading resolved | 18 | 18 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading value | 18 | 8 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading varStatement | 18 | 14 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| a method in object destructuring | 17 | 15 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an unproven relation from string to T: the source is not assignable to the target | 17 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate target whose check would read past the depth bound (TypeReference) | 16 | 3 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an optional chain longer than one step | 16 | 9 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of singleOrMany result T \| T[] cannot be served by implementation result T \| readonly T[] \| undefined | 16 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of skipOuterExpressions result T cannot be served by implementation result Node | 16 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| reading dupFile | 16 | 4 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading existing | 16 | 7 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading includes | 16 | 16 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading parentPath | 16 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading types | 16 | 11 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| a checked predicate target whose check would read past the depth bound (MappedType) | 15 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a destructured name held otherwise than its field | 15 | 7 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a template interpolating an object, an array, a map, a function or undefined | 15 | 3 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an unproven relation from string to P: the source is not assignable to the target | 15 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of parseModifiers result NodeArray&lt;Modifier&gt; \| undefined cannot be served by implementation result NodeArray&lt;ModifierLike&gt; \| undefined | 15 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; } | 14 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| function identity comparison in a program with specialized generic function values | 14 | 9 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| inherited library member push read as an own field | 14 | 8 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of createToken result SuperExpression cannot be served by implementation result Mutable&lt;Token&lt;TKind&gt;&gt; | 14 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of findAncestor result T \| undefined cannot be served by implementation result Node \| undefined (return at /workspace/graph-boundary-adapted/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) | 14 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of getPathComponents result PathPathComponents cannot be served by implementation result string[] | 14 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| reading major | 14 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| a call spreading something other than an array | 13 | 11 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a generic instantiated at void whose body reads the value: mutateMapSkippingNewValues's U = void (read from the signature), read at /workspace/graph-boundary-adapted/src/compiler/utilities.ts:8213:48 as a destructured value | 13 | 13 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| assigning to a NonNullExpression | 13 | 12 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| reading block | 13 | 13 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading sig | 13 | 13 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading specifier | 13 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading varDecl | 13 | 11 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| sort without a comparator | 13 | 6 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| Object enumeration with getter or accessor descriptors | 12 | 12 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a dynamic getter or method read, or a nullable field (runtime dispatch, slot tags and narrowed rereads) | 12 | 7 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a function returning this | 12 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type ResolvedConfigFileName | 12 | 7 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| assigning an element of a value | 12 | 7 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target ReferencedFile | 12 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| new a ParenthesizedExpression | 12 | 7 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of getThisContainer result ThisContainer cannot be served by implementation result ArrowFunction \| ComputedPropertyName \| ThisContainer | 12 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| reading compilerHost | 12 | 12 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading exception | 12 | 12 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading isFromNodeModulesSearch | 12 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading parameters | 12 | 12 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| a [specifier: string, mode: ResolutionMode] \| undefined where [any, any] goes, with no representation Adamic can name for the slot's elements (string stored, read as any) | 11 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type (ExportDeclaration &amp; { readonly isTypeOnly: true; readonly moduleSpecifier: Expression; }) \| undefined | 11 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type Readonly&lt;PathPathComponents&gt; | 11 | 7 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target LiteralImportTypeNode | 11 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of assertNode parameter test cannot be served by implementation parameter test | 11 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| reading symbol | 11 | 11 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| an array of boolean \| undefined | 10 | 6 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| iterating a value | 10 | 6 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of writeTokenText result void cannot be served by implementation result number | 10 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| reading chain | 10 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading constraint | 10 | 6 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading members | 10 | 7 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading propertyName | 10 | 9 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading sourceFile | 10 | 10 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| a presence test of a method whose presence the receiver's construction doesn't prove; a view or host value has to read it | 9 | 8 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a spread after the first field | 9 | 3 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type InitializedVariableDeclaration | 9 | 4 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target ImportEqualsDeclaration &amp; { moduleReference: ExternalModuleReference; } | 9 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| optional record property access | 9 | 7 | 9 | 7 | all: Reason explicitly describes a record, dictionary or index-signature object. |
| overload 1 of createBuilderProgram result SemanticDiagnosticsBuilderProgram cannot be served by implementation result BuilderProgram \| undefined | 9 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of getSupportedExtensions result readonly Extension[][] cannot be served by implementation result readonly string[][] | 9 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of nodeCanBeDecorated result true cannot be served by implementation result boolean | 9 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| reading commonDir | 9 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading declaration | 9 | 9 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading forStatement | 9 | 9 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading origin | 9 | 7 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading text | 9 | 9 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| a Map or Set going where a plain object type is held | 8 | 6 | 0 | 0 | unmarked: Unmarked: Map/Set representation alone is not a dictionary or record object; no verified record source context. Uncertain cases remain unmarked. |
| a checked predicate target whose check would read past the depth bound (BuilderProgramStateWithDefinedProgram) | 8 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a union or optional field read in a program with record storage | 8 | 8 | 8 | 8 | all: Reason explicitly describes a record, dictionary or index-signature object. |
| a value of type (ModuleSpecifierResolutionHost &amp; { getCommonSourceDirectory(): string; }) \| undefined | 8 | 3 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an overloaded function as a value | 8 | 6 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target Identifier | 8 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target NamedDeclaration &amp; { name: DeclarationName; } | 8 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| inherited library member shift read as an own field | 8 | 6 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of getSupportedExtensionsWithJsonIfResolveJsonModule parameter supportedExtensions cannot be served by implementation parameter supportedExtensions | 8 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of visitFunctionBody result Block cannot be served by implementation result ConciseBody \| undefined | 8 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| reading assignment | 8 | 6 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading currentModifiedTime | 8 | 4 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading diag | 8 | 8 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading initializer | 8 | 7 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading outputDts | 8 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading resolvedTypeReferenceDirective | 8 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading scope | 8 | 8 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading sources | 8 | 4 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading suggestion | 8 | 4 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| union-typed array built from a narrower literal | 8 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| RegExp with a nonconstant pattern | 7 | 7 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a Map of CompilerOptionsValue | 7 | 7 | 0 | 0 | unmarked: Unmarked: CompilerOptionsValue includes MapLike&lt;string[]&gt; in types.ts:7412, but this Map representation reason does not identify the failing union member; record involvement is uncertain. Uncertain cases remain unmarked. |
| a checked predicate target whose base has no required field to check (CompilerOptions &amp; { types: string[]; }) | 7 | 1 | 7 | 1 | all: Predicate target is CompilerOptions, whose types.ts declaration has a string index signature. |
| a function or method with a this parameter followed by another parameter (dynamic receivers are not implemented) | 7 | 7 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type BindableObjectDefinePropertyCall | 7 | 3 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type BindableStaticNameExpression | 7 | 5 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type T \| undefined | 7 | 4 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of parseExpectedToken result Token&lt;TKind&gt; cannot be served by implementation result Node | 7 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of visitArray parameter visitor cannot be served by implementation parameter visitor | 7 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| reading container | 7 | 7 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading expr | 7 | 7 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading globalCache | 7 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading t | 7 | 5 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| a PostfixUnaryExpression | 6 | 4 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a comparator that doesn't take two elements and return a number | 6 | 4 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a generic instantiated at void whose body reads the value: getPropFromRaw's T = void (read from the signature), read at /workspace/graph-boundary-adapted/src/compiler/commandLineParser.ts:3253:24 as a returned value | 6 | 3 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a method of a class stage 0 doesn't have | 6 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an array of unknown with erased element storage (retain its declared element type before reading elements) | 6 | 5 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an unproven relation from "" to T: the source is not assignable to the target | 6 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| for...in without a proven fixed plain-object origin (arrays, prototypes and absent synthetic fields cannot be enumerated soundly) | 6 | 4 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| in when a possibly absent spread has no property presence descriptors | 6 | 6 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| inherited library member call read as an own field | 6 | 6 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of assertEachNode parameter test cannot be served by implementation parameter test | 6 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of serializePropertySymbolsForClassOrInterface result TypeElement[] cannot be served by implementation result (ClassElement \| TypeElement)[] | 6 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| reading capturedLeft | 6 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading data | 6 | 4 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading declarationTransform | 6 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading diagnostic | 6 | 6 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading encodeURI | 6 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading enumStatement | 6 | 6 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading exportSpecifiers | 6 | 4 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading extendsRaw | 6 | 2 | 6 | 2 | all: Source checker verifies string-keyed read/write on an index-signature object. |
| reading flags | 6 | 6 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading innerExpression | 6 | 5 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading iterator | 6 | 6 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading map | 6 | 4 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading moduleStatement | 6 | 6 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading needCompilerDiagnostic | 6 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading newParsedCommandLine | 6 | 3 | 2 | 1 | some: Source checker verifies string-keyed read/write on an index-signature object. Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading ownKey | 6 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading propType | 6 | 5 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading props | 6 | 4 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading questionToken | 6 | 6 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading reference | 6 | 6 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading tag | 6 | 6 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading typeParameters | 6 | 6 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| ?. to a number, which would be number \| undefined | 5 | 5 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) | 5 | 5 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a Map of ExportDeclaration &amp; { readonly isTypeOnly: true; readonly moduleSpecifier: Expression; } | 5 | 3 | 0 | 0 | unmarked: Unmarked: Map/Set representation alone is not a dictionary or record object; no verified record source context. Uncertain cases remain unmarked. |
| a checked predicate target whose check would read past the depth bound (NumberLiteralType) | 5 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a conditional whose branches have different types | 5 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a for...of destructuring an object | 5 | 5 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a function returning (ConstructorDeclaration &amp; { body: Block; }) \| undefined | 5 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a function returning CapturedThis | 5 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a function returning ImmediatelyInvokedArrowFunction | 5 | 4 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a generic instantiated at void whose body reads the value: forEachEmittedFile's T = void, read at /workspace/graph-boundary-adapted/src/compiler/emitter.ts:461:21 as a value | 5 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a rest parameter that isn't an array | 5 | 3 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type EndOfFileToken | 5 | 3 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type ExpressionWithTypeArguments &amp; { readonly expression: Identifier \| PropertyAccessEntityNameExpression; } | 5 | 5 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type Node seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 5 | 5 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type T | 5 | 4 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an ElementAccessExpression | 5 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an array of T | 5 | 3 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| detached hasOwnProperty through an object view (an Error may be hidden by the view) | 5 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of getGlobalBuiltinTypes result ObjectType[] cannot be served by implementation result Type[] | 5 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of getProperties result readonly InitializedPropertyDeclaration[] cannot be served by implementation result readonly PropertyDeclaration[] | 5 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of parseNamedImportsOrExports result NamedImports cannot be served by implementation result NamedImportsOrExports | 5 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of symbolToName result Identifier cannot be served by implementation result EntityName | 5 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 2 of getParseTreeNode result T \| undefined cannot be served by implementation result Node \| undefined | 5 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| reading __filename | 5 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading asteriskToken | 5 | 5 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading buildOptions | 5 | 5 | 4 | 4 | some: Source checker verifies string-keyed read/write on an index-signature object. Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading cache | 5 | 5 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading candidate | 5 | 4 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading classType | 5 | 5 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading elements | 5 | 5 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading emittedExpression | 5 | 5 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading fakespace | 5 | 5 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading func | 5 | 5 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading iife | 5 | 5 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading indexType | 5 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading isAsync | 5 | 5 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading isGenerator | 5 | 5 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading key | 5 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading kind | 5 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading objectType | 5 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading param | 5 | 5 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading parent | 5 | 5 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading seen | 5 | 5 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| Object.defineProperties | 4 | 4 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a (n: Node) =&gt; boolean seen as a (element: Node) =&gt; "quit" \| boolean (one keeps something weakly that the other keeps strongly) | 4 | 4 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a BinaryExpression with a number \| undefined and a number | 4 | 4 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a computed key without an own data field | 4 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a filter callback that doesn't return a boolean | 4 | 4 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a function returning Declaration &amp; HasModifiers | 4 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a function returning PollingIntervalQueue | 4 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a function without a body | 4 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a generic function value without a concrete contextual signature | 4 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a logical string operand converted to number \| undefined, whose falsy "" the result can't hold | 4 | 4 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a undefined[] where (boolean \| undefined)[] goes: its undefined elements are stored as a reference and read there as boolean \| undefined, held as boolean \| undefined (build it as the slot's type where it's made, or copy its elements into one) | 4 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type (ConstructorDeclaration &amp; { body: Block; }) \| undefined | 4 | 4 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type (VariableDeclaration &amp; { name: Identifier; }) \| undefined | 4 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type HasJSDoc \| undefined | 4 | 4 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type ImmediatelyInvokedArrowFunction | 4 | 4 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type ImportDeclaration seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 4 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type ParameterPropertyDeclaration | 4 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type StringLiteral seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 4 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an unproven relation from Type to TypeVariable: optional field constraint has no proven compatible presence/type | 4 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target BindableStaticElementAccessExpression | 4 | 3 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target Block | 4 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target Comparator[][] | 4 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target NodeArray&lt;Expression&gt; | 4 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target ParameterPropertyDeclaration | 4 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target SourceFile | 4 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target Statement | 4 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target TsConfigSourceFile | 4 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of captureMapping result Required&lt;Mapping&gt; cannot be served by implementation result Mapping | 4 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of evaluate result EvaluatorResult&lt;string \| undefined&gt; cannot be served by implementation result EvaluatorResult&lt;string \| number \| undefined&gt; | 4 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of getDiagnostics result DiagnosticWithLocation[] cannot be served by implementation result Diagnostic[] | 4 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of parenthesizeConciseBodyOfArrowFunction result Expression cannot be served by implementation result ConciseBody | 4 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of transformAsyncFunctionBody result Block cannot be served by implementation result ConciseBody | 4 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of transformFunctionBody result Block cannot be served by implementation result ConciseBody | 4 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| reading actualFileName | 4 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading affectedSourceFile | 4 | 4 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading args | 4 | 4 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading assignClassAliasInStaticBlock | 4 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading baseObjectType | 4 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading basePath | 4 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading baseUrl | 4 | 4 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading classExpression | 4 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading compilerOptionsProperty | 4 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading constructor | 4 | 4 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading d | 4 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading decl | 4 | 4 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading declBlocked | 4 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading declarationList | 4 | 4 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading element | 4 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading elideImport | 4 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading emitNode | 4 | 4 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading emitResult | 4 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading emittedFiles | 4 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading exportStatement | 4 | 4 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading extensions | 4 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading fileIdListId | 4 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading filesForEmit | 4 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading fromCache | 4 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading funcStatements | 4 | 4 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading getCanonicalFileName | 4 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading hasExistingReasonToReportErrorOn | 4 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading i | 4 | 4 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading inference | 4 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading inferredProp | 4 | 4 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading isDeferredMappedIndex | 4 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading isJsFileFromNodeModules | 4 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading links | 4 | 4 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading moduleResolutionName | 4 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading moduleSymbol | 4 | 4 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading newValue | 4 | 4 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading normalized | 4 | 4 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading oldResolution | 4 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading operand | 4 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading parseNode | 4 | 4 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading parts | 4 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading pragma | 4 | 4 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading presentWatcher | 4 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading prop | 4 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading questionDotToken | 4 | 4 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading rawText | 4 | 4 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading referencesOfRaw | 4 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading regularType | 4 | 4 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading related | 4 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading relatedInfo | 4 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading resolvedFileName | 4 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading resultFromHost | 4 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading reversed | 4 | 4 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading rightExpression | 4 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading signatureNode | 4 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading sourceDiscriminantTypes | 4 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading sourceUnionOrIntersection | 4 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading spanResult | 4 | 4 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading targetIsOptional | 4 | 4 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading targetType | 4 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading targets | 4 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading typeSet | 4 | 4 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| spreading an array of other elements | 4 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| Object.setPrototypeOf | 3 | 3 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a Map whose key and value types aren't known | 3 | 3 | 0 | 0 | unmarked: Unmarked: Map/Set representation alone is not a dictionary or record object; no verified record source context. Uncertain cases remain unmarked. |
| a checked predicate target Expression that narrows undefined, which isn't its supertype | 3 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate target whose check would read past the depth bound (IncrementalBundleEmitBuildInfo) | 3 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate target whose check would read past the depth bound (IndexedAccessType) | 3 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate target whose check would read past the depth bound (ResolvedProjectReference) | 3 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate target whose check would read past the depth bound (TransientSymbol) | 3 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a computed field name | 3 | 3 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a function returning MissingList&lt;ParameterDeclaration&gt; | 3 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a function returning MissingList&lt;TypeNode&gt; | 3 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a function returning PropertyAccessExpression \| SyntheticSuper | 3 | 3 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a generic instantiated at void whose body reads the value: parseConfigHostFromCompilerHostLike's T = void (read from the signature), read at /workspace/graph-boundary-adapted/src/compiler/program.ts:5106:54 as a binding's initializer | 3 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a logical string operand converted to value, whose falsy "" the result can't hold | 3 | 3 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a non-accessor method in an accessor literal | 3 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a number[] where string \| number \| boolean \| Args \| readonly (string \| number \| boolean \| Args \| null \| undefined)[] \| null \| undefined goes: its number elements are stored as number and read there as string \| number \| boolean \| Args \| null \| undefined, held as a boxed union (build it as the slot's type where it's made, or copy its elements into one) | 3 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a parameter that isn't a plain name | 3 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a placeholder reset of a parameter | 3 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a readonly T[] where readonly (string \| FileReference)[] goes, with no representation Adamic can name for the stored elements (T stored, read as string \| FileReference) | 3 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a tuple literal leaving out an element of type ModuleSpecifierEnding | 3 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type (EmitNode &amp; { autoGenerate: AutoGenerateInfo; }) \| (EmitNode &amp; { autoGenerate: AutoGenerateInfo; }) | 3 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type CompilerHost &amp; ReadBuildProgramHost | 3 | 3 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type NodeArray&lt;BindingElement&gt; \| NodeArray&lt;ArrayBindingElement&gt; where an array goes | 3 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type PollingIntervalQueue | 3 | 3 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type ResolvedConfigFilePath | 3 | 3 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type WrappedExpression&lt;AnonymousFunctionDefinition&gt; | 3 | 3 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type false \| TypeOnlyAliasDeclaration \| undefined | 3 | 3 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an array of InitializedVariableDeclaration | 3 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an array of ResolvedConfigFileName | 3 | 3 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an unproven relation from Declaration to T: the source is not assignable to the target | 3 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target ArrayBindingPattern | 3 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target AssignmentExpression&lt;EqualsToken&gt; &amp; { readonly left: Identifier; readonly right: WrappedExpression&lt;AnonymousFunctionDefinition&gt;; } | 3 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target BinaryExpression &amp; { operatorToken: Token&lt;SyntaxKind.CommaToken&gt;; } | 3 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target ElementAccessExpression &amp; Declaration &amp; { readonly argumentExpression: NumericLiteral \| StringLiteralLike; } &amp; ... | 3 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target LeftHandSideExpression | 3 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target SourceMappedPosition | 3 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target Symbol | 3 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target TransformedSuperCall | 3 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target VariableDeclarationList &amp; { _usingBrand: undefined; } | 3 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked view field modifiers of type NodeArray&lt;Modifier&gt; \| NodeArray&lt;ModifierLike&gt; \| undefined | 3 | 3 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| destructuring a value made in the assignment itself | 3 | 3 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| for...of over an object | 3 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| inherited library member get read as an own field | 3 | 3 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| inherited library member has read as an own field | 3 | 3 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| method bind with an assigned receiver identifier | 3 | 3 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| new a PropertyAccessExpression | 3 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| node:fs.StatsBase.isDirectory through an optional call | 3 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| node:fs.writeFileSync: JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) | 3 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of assertOptionalNode parameter test cannot be served by implementation parameter test | 3 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of createImmediatelyInvokedArrowFunction result ImmediatelyInvokedArrowFunction cannot be served by implementation result Mutable&lt;CallExpression&gt; | 3 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of createImmediatelyInvokedFunctionExpression result ImmediatelyInvokedFunctionExpression cannot be served by implementation result Mutable&lt;CallExpression&gt; | 3 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of getSuperContainer result SuperContainer \| undefined cannot be served by implementation result SuperContainerOrFunctions \| undefined | 3 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of mergeLexicalEnvironment result NodeArray&lt;Statement&gt; cannot be served by implementation result Statement[] \| NodeArray&lt;Statement&gt; | 3 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of visitBindingElement parameter elem cannot be served by implementation parameter elem | 3 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 2 of arrayToMap result Map&lt;K, V2&gt; cannot be served by implementation result Map&lt;K, V1 \| V2&gt; | 3 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| reading aliasAssignment | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading argument | 3 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading autoGenerate | 3 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading backingField | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading boundTag | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading buildOrder | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading cachedPackageJson | 3 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading call | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading captureThisStatement | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading childComponents | 3 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading classDeclaration | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading constructSignatures | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading constructorFunction | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading diff | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading dotDotDotToken | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading expressionText | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading extendedSpecs | 3 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading filtered | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading firstArgument | 3 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading fragment | 3 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading getter | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading graphNode | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading heritageClause | 3 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading inner | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading inspector | 3 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading invocation | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading isImmediatelyInvoked | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading isStaticMethodSymbol | 3 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading jsdocAliasDecl | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading lastDecorator | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading match | 3 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading memberDecoratorsAssignment | 3 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading mode | 3 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading modifier | 3 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading named | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading newBaseType | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading newVarStatement | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading nodes | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading oldEmitKind | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading outer | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading outerTypeParameters | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading parametersWithPropertyAssignments | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading params | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading parentComponents | 3 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading pathComponents | 3 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading pos | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading qualifiedName | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading regularNew | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading resolvedProject | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading root | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading savedPreserveSourceNewlines | 3 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading session | 3 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading setter | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading start | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading statementExpression | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading superProperty | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading timerToUpdateChildWatches | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading typeCopy | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading typeOnlyDeclaration | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading visitedAccessorName | 3 | 3 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| Array.isArray on a tuple or an erased object/any/unknown view | 2 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| JSON.parse: its result's type can't be proven from the text | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| JSON.stringify a union containing containers without runtime element metadata | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a (d: Node) =&gt; boolean seen as a (element: Node) =&gt; "quit" \| boolean (one keeps something weakly that the other keeps strongly) | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a BinaryExpression with a string and a union of differently held members | 2 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a DeleteExpression as a statement | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a Set of ResolvedConfigFilePath (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far) | 2 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a SpreadElement in a tuple literal | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a [message: DiagnosticMessage, ...args: (string \| number \| boolean \| readonly string[] \| SourceFile \| undefined)[]] \| undefined where [any, ...any[]] goes, with no representation Adamic can name for the slot's elements (DiagnosticMessage stored, read as any) | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a [string, string] \| undefined where [any, any] goes, with no representation Adamic can name for the slot's elements (string stored, read as any) | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate target Expression \| VariableDeclarationList over a generic argument T \| null \| undefined, read on the true side as Expression \| VariableDeclarationList | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate target LateBoundDeclaration that narrows ClassElement &amp; { name: ComputedPropertyName; }, which isn't its supertype | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate target Statement that narrows ParameterDeclaration, which isn't its supertype | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate target whose check would read past the depth bound (BuilderProgram) | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate target whose check would read past the depth bound (CircularBuildOrder) | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate target whose check would read past the depth bound (ReferenceFileLocation) | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate target whose check would read past the depth bound (Type) | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a class method through a view that erases its prototype origin | 2 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a function returning CanonicalKey | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a function returning ClassExpression \| ImmediatelyInvokedArrowFunction | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a function returning ExpressionWithTypeArguments &amp; { expression: Identifier \| PropertyAccessEntityNameExpression; } | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a function returning HasJSDoc \| undefined | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a function returning NodeArray&lt;never&gt; | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a function returning ResolvedConfigFileName | 2 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a function returning SyntheticSuper | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a function returning T | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a function taking string seen as one taking string \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a generic instantiated at void whose body reads the value: createWatchProgram's T = void (read from the signature), read at /workspace/graph-boundary-adapted/src/compiler/watchPublic.ts:455:142 as a destructured value | 2 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a tagged template other than the intrinsic String.raw | 2 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a type argument makes a value of type (left: MappedPosition, right: MappedPosition) =&gt; boolean seen as EqualityComparer&lt;SourceMappedPosition&gt; \| undefined, which can write string \| undefined where string is read | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type AccessorDeclaration &amp; { readonly name: BigIntLiteral \| ComputedPropertyName \| Identifier \| NoSubstitutionTemplateLiteral \| NumericLiteral \| StringLiteral; } | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type ArrowFunction \| BinaryExpression \| BindingElement \| Block \| BreakStatement \| CallSignatureDeclaration \| ... 64 more ... \| EndOfFileToken | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type BindablePropertyAssignmentExpression \| PropertyAccessExpression \| LiteralLikeElementAccessExpression | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type BindableStaticAccessExpression | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type ConstructorDeclaration seen as Mutable&lt;ConstructorDeclaration&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type EmitNode &amp; { autoGenerate: AutoGenerateInfo; } | 2 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type ExpressionWithTypeArguments &amp; { expression: Identifier \| PropertyAccessEntityNameExpression; } | 2 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type MissingDeclaration seen as Mutable&lt;MissingDeclaration&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type NonNullable&lt;T&gt; | 2 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type PrivateIdentifierInExpression | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type PropertyDeclaration \| ParameterPropertyDeclaration | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type SourceFile seen as Mutable&lt;SourceFile&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type VariableDeclaration &amp; { name: Identifier; } | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an array of ParameterPropertyDeclaration | 2 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an assignment result with different storage | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an object with a nullable field viewed as unknown or object (null and undefined slot tags) | 2 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an unproven relation from ObjectType to AnonymousType: optional field target has no proven compatible presence/type | 2 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an unproven relation from PackageJsonPathFields to PackageJson: optional field version has no proven compatible presence/type | 2 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an unproven relation from UnionOrIntersectionType to UnionType: optional field resolvedReducedType has no proven compatible presence/type | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| apply without a dense argument literal (length, presence and argument representations must be proven) | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target BigIntLiteral | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target BindableObjectDefinePropertyCall | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target ClassNamedEvaluationHelperBlock | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target ClassThisAssignmentBlock | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target Expression &amp; { parent: OptionalChainRoot; } | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target JSDocTag | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target NodeArray&lt;Statement&gt; | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target NodeArray&lt;TypeNode&gt; | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target ParameterDeclaration &amp; { initializer: Expression; } | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target PrefixUnaryExpression &amp; { operand: NumericLiteral; } | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target Program | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target SourceMapping | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target ThisCapturingAssignmentExpression | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target ThisCapturingVariableStatement | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked view field id of type number \| undefined | 2 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| destructuring a value | 2 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| hasOwnProperty through an object view (an Error may be hidden by the view) | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| indexOf on a tuple (a tuple is held as an object, not an array, so far; write it as an array where it's made) | 2 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| inherited library member add read as an own field | 2 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| inherited library member compare read as an own field | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| inherited library member set read as an own field | 2 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| instanceof against a value that isn't a declared class | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| new Map from something that isn't [key, value] pairs | 2 | 2 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| node:fs.StatsBase.isFile through an optional call | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of convertOptionsFromJson result WatchOptions \| undefined cannot be served by implementation result CompilerOptions \| TypeAcquisition \| WatchOptions \| undefined | 2 | 1 | 2 | 1 | all: Result types CompilerOptions, TypeAcquisition and WatchOptions all declare string index signatures in types.ts. |
| overload 1 of createBinaryExpressionTrampoline result (node: BinaryExpression) =&gt; TResult cannot be served by implementation result (node: BinaryExpression, outerState: TOuterState) =&gt; TResult | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of findAncestor result T \| undefined cannot be served by implementation result Node \| undefined (return at /workspace/graph-boundary-adapted/src/compiler/utilitiesPublic.ts:792:13 isn't the value its predicate checked) | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of groupBy type parameter U cannot satisfy implementation parameter K | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of injectClassNamedEvaluationHelperBlockIfMissing result Extract&lt;ClassDeclaration, Pick&lt;...&gt;&gt; \| Extract&lt;...&gt; cannot be served by implementation result ClassLikeDeclaration | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of or type parameter P cannot satisfy implementation parameter T | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of parseExpectedTokenJSDoc result Token&lt;TKind&gt; cannot be served by implementation result Node | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 2 of arrayToMultiMap result MultiMap&lt;K, U&gt; cannot be served by implementation result MultiMap&lt;K, U \| V&gt; | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| process.recordreplay as a value (use its supported process operation directly) | 2 | 1 | 0 | 0 | unmarked: Unmarked: process.recordreplay is a process operation name, not a record object. Uncertain cases remain unmarked. |
| push on a tuple (a tuple is held as an object, not an array, so far; write it as an array where it's made) | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| reading aComponents | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading aParts | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading abstractSignatures | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading allTypeFlags | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading antecedents | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading areAllModuleExports | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading argumentExpression | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading attrs | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading baseIndexedAccess | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading baseTypes | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading body | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading buildInfo | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading byteOrderMarkIndicator | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading callArgument | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading callResult | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading candidates | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading checkFlags | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading child | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading chosenPragma | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading classFunction | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading className | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading classStatements | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading cleanup | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading columns | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading combinedFlags | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading commonResolved | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading commonSourceDirectory | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading comparer | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading condition | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading config | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading constraints | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading constructorDeclaration | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading containingCall | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading declarations | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading declaredType | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading decorators | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading defaultConstraint | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading directoryPath | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading discriminantCombinations | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading distributiveConstraint | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading elementAccess | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading elementTypes | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading emitComments | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading emitSignature | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading emitSourceMaps | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading emittedOperand | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading endLabel | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading equalsGreaterThanToken | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading equalsToken | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading errorNode | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading exclamationToken | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading executeStatements | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading exportName | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading exports | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading fileIds | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading first | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading formattedValue | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading freshType | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading fromComponents | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading getModifiedTime | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading hasAntecedentList | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading hasPrivateIdentifier | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading hasZeroOrNoReferences | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading id | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading implementsTypeNodes | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading impliedNodeFormat | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading importOrExport | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading index | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading indexConstraint | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading init | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading innerModuleSymbol | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading inputListOrBundle | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading instantiation | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading instantiationExpressionType | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading isAmbient | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading isInExternalModule | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading isJsonFile | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading isStaticChildren | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading iteratorValueStatement | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading jsFilePath | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading keyAttr | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading lastId | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading lastModifier | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading lastParamVariadicType | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading leadingNewlines | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading leftExpression | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading leftThisArg | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading length | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading libName | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading libReference | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading locationMap | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading mapByKeyProperty | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading mappings | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading maybeParameters | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading metadata | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading moduleArg | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading moduleAugmentation | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading moduleObject | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading moduleReference | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading msg | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading nameText | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading nameType | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading namespaceDeclaration | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading needsParens | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading newReturnType | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading newStatements | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading nodeFlags | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading numericNamesOnly | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading objectProperties | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading operation | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading option | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading origTypeParameter | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading originalClassDecl | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading originalCreateDirectory | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading originalDecl | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading paramSymbol | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading parameterPropertyAssignments | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading partsAfterFirst | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading pendingKind | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading primaryDeclaration | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading primaryTypes | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading properties | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading property | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading rawtext | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading real | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading reduced | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading reducedTypes | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading ref | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading remainingPaths | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading requireOptionalProperties | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading restElement | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading restType | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading reuseIdentifierExpressions | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading right | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading savedInStrictMode | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading scanner | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading seenKind | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading signature | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading skipCaching | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading sourceExtends | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading sourceFiles | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading sourceIsJSConstructor | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading sourceMapUrlPos | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading sourceOrigin | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading sourceParams | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading sourceSignatures | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading sourceSize | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading sourceTypes | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading specifiers | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading spread | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading state | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading subpathParts | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading suggestedType | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading tags | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading targetFile | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading targetFlags | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading targetIsJSConstructor | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading targetOrigin | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading targetProp | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading targetReturnType | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading targetTypes | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading targetUnionOrIntersection | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading tempSources | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading toComponents | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading toWatch | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading typeAlias | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading typeNode | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading typeOnlyDeclarationIsExportStar | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading typeVariable | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading undefinedStrippedTarget | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading unmatched | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading unmatchedProperty | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading unqualifiedLibName | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading validatedFilesSpec | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading validatedFilesSpecBeforeSubstitution | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading valueDeclaration | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading values | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading varDeclList | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading varInitializer | 2 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading variable | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading version | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading visibleDefaultBinding | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading widened | 2 | 2 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| statement panic: Unhandled case in Node.Text: *ast.BindingPattern | 2 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| .types on a value | 1 | 1 | 1 | 1 | all: Source checker verifies string-keyed read/write on an index-signature object. |
| Array as a value outside equality or typeof (overloaded calls and static properties need their own representation) | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| Object.assign on a shape not proven by a plain literal or its const binding | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| Object.create | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| Object.defineProperty | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| Object.fromEntries | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| Reflect.get by a run-time key in a program with a class getter that does more than return its fields and literals joined by operators (the getter runs inside the read, unseen by the passes that follow calls) | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a () =&gt; never seen as a (node: ElementAccessExpression \| EnumMember \| PropertyAccessExpression) =&gt; string \| number \| undefined (one keeps something weakly that the other keeps strongly) | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a () =&gt; true seen as a () =&gt; RepopulateDiagnosticChainInfo (one keeps something weakly that the other keeps strongly) | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a (_?: unknown) =&gt; void seen as a (flags: LexicalEnvironmentFlags, value: boolean) =&gt; void (one keeps something weakly that the other keeps strongly) | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a (current: Node) =&gt; boolean seen as a (element: Node) =&gt; "quit" \| boolean (one keeps something weakly that the other keeps strongly) | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a (declaration: Node) =&gt; boolean seen as a (element: Node) =&gt; "quit" \| boolean (one keeps something weakly that the other keeps strongly) | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a (file: SourceFile, options: CompilerOptions) =&gt; true \| undefined seen as a (file: SourceFile, options: CompilerOptions) =&gt; true \| Node \| undefined (one keeps something weakly that the other keeps strongly) | 1 | 1 | 0 | 0 | unmarked: Unmarked: CompilerOptions is mentioned in another feature signature, but record-shaped causality is not established at this site. Uncertain cases remain unmarked. |
| a (location: Node) =&gt; boolean seen as a (element: Node) =&gt; "quit" \| boolean (one keeps something weakly that the other keeps strongly) | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a (node: Node) =&gt; boolean seen as a (element: Node) =&gt; "quit" \| boolean (one keeps something weakly that the other keeps strongly) | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a ClassExpression | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a Map of ResolvedConfigFilePath | 1 | 1 | 0 | 0 | unmarked: Unmarked: Map/Set representation alone is not a dictionary or record object; no verified record source context. Uncertain cases remain unmarked. |
| a Map&lt;__String, ExportDeclaration &amp; { readonly isTypeOnly: true; readonly moduleSpecifier: Expression; }&gt; \| undefined where Map&lt;__String, ExportDeclaration &amp; { readonly isTypeOnly: true; readonly moduleSpecifier: Expression; }&gt; \| undefined goes, with no representation Adamic can name for the stored elements (ExportDeclaration &amp; { readonly isTypeOnly: true; readonly moduleSpecifier: Expression; } stored, read as ExportDeclaration &amp; { readonly isTypeOnly: true; readonly moduleSpecifier: Expression; }) | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a Map&lt;string, JsxEmit&gt; where Map&lt;string, string \| number&gt; goes: its JsxEmit elements are stored as number and read there as string \| number, held as a boxed union (build it as the slot's type where it's made, or copy its elements into one) | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a block-scoped nested function declaration | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a call returning any | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a call to a PropertyAccessExpression | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked cast from never[] to [] whose value isn't a plain object | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked generic predicate used as a value | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate result read by a ConditionalExpression inside a condition, which its direction walk doesn't read | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate result returned from a function the checker infers a type predicate for, an ArrowFunction, whose uses the walk doesn't follow, which its direction walk doesn't read | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate result returned from a function the checker infers a type predicate for, isEnclosingDeclaration, whose false side the call at /workspace/graph-boundary-adapted/src/compiler/transformers/declarations.ts:1408:13 reads, which its direction walk doesn't read | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate target Declaration that narrows undefined, which isn't its supertype | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate target DynamicNamedDeclaration that narrows BindablePropertyAssignmentExpression, which isn't its supertype | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate target DynamicNamedDeclaration that narrows GetAccessorDeclaration, which isn't its supertype | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate target Expression that narrows ArrayBindingPattern, which isn't its supertype | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate target Expression that narrows BigIntLiteral, which isn't its supertype | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate target Expression that narrows Declaration, which isn't its supertype | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate target LateBoundDeclaration that narrows ClassElement, which isn't its supertype | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate target LateBoundDeclaration that narrows DynamicNamedBinaryExpression, which isn't its supertype | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate target LeftHandSideExpression that narrows ArrowFunction, which isn't its supertype | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate target NamedDeclaration that narrows ArrowFunction, which isn't its supertype | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate target NamedDeclaration that narrows BindingElement, which isn't its supertype | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate target NamedDeclaration that narrows ClassDeclaration, which isn't its supertype | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate target TypeNode that narrows undefined, which isn't its supertype | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate target whose base has no required field to check ((externalModuleName: string) =&gt; string) | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate target whose base has no required field to check ((name: GeneratedIdentifier \| GeneratedPrivateIdentifier) =&gt; string) | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate target whose base has no required field to check (void &amp; { __escapedIdentifier: void; }) | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate target whose base has no required field to check ({ [key: string]: unknown; }) | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate target whose check would read past the depth bound (CommandLineOptionOfCustomType) | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate target whose check would read past the depth bound (DiagnosticWithDetachedLocation) | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate target whose check would read past the depth bound (FlowSwitchClause) | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate target whose check would read past the depth bound (HostDirectoryWatcher) | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate target whose check would read past the depth bound (PackageJsonInfo) | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a checked predicate target whose check would read past the depth bound (ReusableRepopulateInfoChain &amp; DiagnosticMessageChain) | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a constructor object escaping before static fields are initialized | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a function returning (AssignmentExpression&lt;EqualsToken&gt; &amp; { readonly left: GeneratedIdentifier; }) \| undefined | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a function returning AnyValidImportOrReExport | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a function returning AnyValidImportOrReExport \| undefined | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a function returning BindableStaticAccessExpression \| undefined | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a function returning ClassNamedEvaluationHelperBlock | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a function returning ClassThisAssignmentBlock | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a function returning ExpressionWithTypeArguments &amp; { readonly expression: Identifier \| PropertyAccessEntityNameExpression; } | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a function returning MemberName \| (Expression &amp; (NumericLiteral \| StringLiteralLike)) | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a function returning MissingList&lt;TypeElement&gt; | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a function returning MissingList&lt;TypeParameterDeclaration&gt; | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a function returning MissingList&lt;VariableDeclaration&gt; | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a function returning TypeOnlyAliasDeclaration \| undefined | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a function returning any | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a generator function whose generator is used other than by a for-of over its call or a call as a statement (#qk8rztp) | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a generator written in place and called, used other than by a for-of over its call, a call as a statement or a one-read Iterable parameter (#qk8rztp) | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a generic instantiated at void whose body reads the value: getEmitListItem's T = void (read from the signature), read at /workspace/graph-boundary-adapted/src/compiler/emitter.ts:6376:16 as typeof's operand | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a generic instantiated at void whose body reads the value: speculationHelper's T = void, read at /workspace/graph-boundary-adapted/src/compiler/parser.ts:2280:14 as a value | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a placeholder reset without a represented variable or field slot | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a readonly ["cs", "de", "es", "fr", "it", "ja", "ko", "pl", "pt-br", "ru", "tr", "zh-cn", "zh-tw"] seen as a readonly string[] (a tuple is held as an object, not an array, so far; write it as an array where it's made, or copy it into one: [pair[0], pair[1]]) | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a readonly view read by a run-time key in a program with a class getter that does more than return its fields and literals joined by operators (the getter runs inside the read, unseen by the passes that follow calls) | 1 | 1 | 1 | 1 | all: Source checker verifies string-keyed read/write on an index-signature object. |
| a spread into record storage from a fixed object | 1 | 1 | 1 | 1 | all: Reason explicitly describes a record, dictionary or index-signature object. |
| a template interpolating a union with an object, an array, a map or a function in it | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a tuple literal with more values than its tuple has elements | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a union or optional tuple length; declare arrays where the values are constructed, or carry an explicit length alongside each tuple | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type "" \| ResolvedConfigFileName \| undefined | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type (AmbientModuleDeclaration &amp; { name: StringLiteral; }) \| undefined | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type (AssignmentExpression&lt;EqualsToken&gt; &amp; { readonly left: GeneratedIdentifier; }) \| undefined | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type (AssignmentExpression&lt;EqualsToken&gt; &amp; { readonly left: Identifier; readonly right: WrappedExpression&lt;AnonymousFunctionDefinition&gt;; } &amp; BinaryExpression) \| (... &amp; ... 1 more ... &amp; BinaryExpression) | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type (ModuleDeclaration &amp; { name: StringLiteral; }) \| undefined | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type BindableAccessExpression | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type BindingElement &amp; { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression&lt;AnonymousFunctionDefinition&gt;; } | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type CallExpression \| BindableStaticAccessExpression | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type CanonicalKey | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type ClassStaticBlockDeclaration seen as Mutable&lt;ClassStaticBlockDeclaration&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type ExportAssignment &amp; { readonly expression: WrappedExpression&lt;AnonymousFunctionDefinition&gt;; } | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type GetAccessorDeclaration \| SetAccessorDeclaration seen as Mutable&lt;AccessorDeclaration&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type Identifier seen as Mutable&lt;Identifier&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type Identifier seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type IncrementalMultiFileEmitBuildInfoFileInfo | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type JSDocArray | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type JSDocImportTag \| CanHaveModuleSpecifier | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type MethodDeclaration seen as Mutable&lt;MethodDeclaration&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type MissingList&lt;Statement&gt; | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type ModuleResolutionState &amp; { host: GetPackageJsonEntrypointsHost; } | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type NamespaceExportDeclaration seen as Mutable&lt;NamespaceExportDeclaration&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type NodeArray&lt;never&gt; | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type NonNullExpression seen as Mutable&lt;NonNullExpression&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type OptionalTypeNode seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type ParameterDeclaration &amp; { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression&lt;AnonymousFunctionDefinition&gt;; } | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type PathPathComponents | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type PragmaPseudoMap[TKey][] \| PragmaPseudoMap[TKey] | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type PropertyAccessExpression \| LiteralLikeElementAccessExpression | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type PropertyAccessExpression \| SyntheticSuper | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type PropertyAssignment &amp; { readonly name: Identifier; readonly initializer: WrappedExpression&lt;AnonymousFunctionDefinition&gt;; } | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type PropertyDeclaration &amp; { readonly initializer: WrappedExpression&lt;AnonymousFunctionDefinition&gt;; } | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type PropertySignature seen as Mutable&lt;PropertySignature&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type Record&lt;number, FlowGraphNode&gt; | 1 | 1 | 1 | 1 | all: Reason explicitly describes a record, dictionary or index-signature object. |
| a value of type ReplaceableIndexedAccessType | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type RequireOrImportCall | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type ShorthandPropertyAssignment &amp; { readonly objectAssignmentInitializer: WrappedExpression&lt;AnonymousFunctionDefinition&gt;; } | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type TSConfig &amp; { watchOptions?: object \| undefined; } | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type TaggedTemplateExpression seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type TemplateStringsArray | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type ThisCapturingVariableDeclaration | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type TupleType &amp; InterfaceTypeWithDeclaredMembers | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type TypeNode &amp; LiteralTypeNode &amp; { readonly literal: StringLiteral; } | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type TypeOnlyAliasDeclaration \| undefined | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type TypeParameterDeclaration &amp; { parent: JSDocTemplateTag; } | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type VariableDeclaration &amp; { readonly name: Identifier; readonly initializer: WrappedExpression&lt;AnonymousFunctionDefinition&gt;; } | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type VariableDeclarationList &amp; { readonly declarations: { 0: ThisCapturingVariableDeclaration; }; } | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type WatchFactoryHost &amp; { trace?(s: string): void; } | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type readonly IncrementalBundleEmitBuildInfoFileInfo[] \| readonly IncrementalMultiFileEmitBuildInfoFileInfo[] where an array goes | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| a value of type string[] \| PluginImport[] \| ProjectReference[] \| (string \| number)[] seen as unknown[], which can write unknown where string is read | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an array of CanonicalKey | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an array of InitializedPropertyDeclaration | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an array of NonNullable&lt;T&gt; | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an array union read with non-array arm NodeArray&lt;ParameterDeclaration&gt; | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an uninitialized object field without a supported declared slot type | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an unproven relation from Declaration to NamedDeclaration: optional field name has no proven compatible presence/type | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an unproven relation from Identifier \| MissingDeclaration \| NumericLiteral \| StringLiteral \| TemplateLiteralLikeNode \| Token&lt;...&gt; to T: the source is not assignable to the target | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an unproven relation from ImportEqualsDeclaration &amp; { moduleReference: ExternalModuleReference; } to ImportEqualsDeclaration: optional field moduleReference.id has no proven compatible presence/type | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an unproven relation from InstantiableType \| UnionOrIntersectionType to TypeParameter: optional field constraint has no proven compatible presence/type | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an unproven relation from LiteralLikeNode to TemplateLiteralLikeNode: optional field rawText has no proven compatible presence/type | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an unproven relation from NodeArray&lt;BindingElement&gt; \| NodeArray&lt;Expression&gt; \| NodeArray&lt;ArrayBindingElement&gt; to NodeArray&lt;Node&gt;: optional field concat.parameter.element.slice.element.propertyName has no proven compatible presence/type | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an unproven relation from NodeArray&lt;ModifierLike&gt; &amp; readonly Decorator[] to NodeArray&lt;Decorator&gt;: optional field concat.parameter.element.slice.element.parent.name has no proven compatible presence/type | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an unproven relation from SolutionBuilderHost&lt;T&gt; \| SolutionBuilderWithWatchHost&lt;T&gt; to SolutionBuilderHost&lt;T&gt;: optional field reportErrorSummary has no proven compatible presence/type | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an unproven relation from SolutionBuilderHostBase&lt;T&gt; to SolutionBuilderHost&lt;T&gt;: optional field reportErrorSummary has no proven compatible presence/type | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| an unrestricted dictionary write that can violate named property all; use a literal key or a value accepted by every named property | 1 | 1 | 1 | 1 | all: Reason explicitly describes a record, dictionary or index-signature object. |
| an unrestricted dictionary write that can violate named property configFile; use a literal key or a value accepted by every named property | 1 | 1 | 1 | 1 | all: Reason explicitly describes a record, dictionary or index-signature object. |
| assigning to an AsExpression | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| bind of a rest-parameter function | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target AssignmentExpression&lt;AmpersandAmpersandEqualsToken \| BarBarEqualsToken \| EqualsToken \| QuestionQuestionEqualsToken&gt; &amp; ... | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target BindingElementOfBareOrAccessedRequire | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target CapturedThis | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target ConstructorDeclaration &amp; { body: Block; } | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target FileInfo | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target FunctionDeclaration &amp; { body: Block; } | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target FunctionExpression &amp; { body: Block; } | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target ImplicitSuperCall | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target ImportClause &amp; { readonly isTypeOnly: true; readonly name: Identifier; } | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target JSDocCallbackTag | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target LiteralLikeElementAccessExpression | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target Map&lt;Path, string \| false&gt; | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target Map&lt;number, Symbol&gt; | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target MethodDeclaration &amp; { body: Block; } | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target NonIncrementalBuildInfo | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target PackageDirWatcher | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target PrivateIdentifierInExpression | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target Symbol[] | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target SyntheticSuper | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target TypeNode | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target TypeReferenceNode &amp; { typeName: Identifier; } | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target readonly Type[] | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked predicate overload target readonly string[] | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked view field parent of type CatchClause \| VariableDeclarationList \| BindingPattern | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| checked view field parent of type JSDoc \| ModuleBlock \| SourceFile | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| destructuring anything but a tuple into [names] | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| findIndex on a tuple (a tuple is held as an object, not an array, so far; write it as an array where it's made) | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| for...in over an array (holes and own enumerable properties are not represented; use for...of for elements) | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| forEach with a callback returning boolean \| undefined | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| hasOwnProperty on a function (its prototype property depends on whether it was declared or made as an arrow) | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| inherited library member name read as an own field | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| join on an array of objects, arrays, maps or functions | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| null comparison with a scalar | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of cloneNode type parameter T cannot satisfy implementation parameter T | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of compact parameter array cannot be served by implementation parameter array | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of convertOrRepopulateDiagnosticMessageChainArray parameter array cannot be served by implementation parameter array | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of convertOrRepopulateDiagnosticMessageChainArray parameter repopulateInfo cannot be served by implementation parameter repopulateInfo | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of parseExpectedToken parameter t cannot be served by implementation parameter t | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of parseOptionalToken parameter t cannot be served by implementation parameter t | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of parseOptionalTokenJSDoc result Token&lt;TKind&gt; cannot be served by implementation result Node \| undefined | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of replaceDecoratorsAndModifiers result T cannot be served by implementation result ClassDeclaration \| ClassExpression \| GetAccessorDeclaration \| MethodDeclaration \| ParameterDeclaration \| PropertyDeclaration \| SetAccessorDeclaration | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of replaceModifiers result T cannot be served by implementation result ArrowFunction \| ClassDeclaration \| ClassExpression \| ConstructorDeclaration \| ConstructorTypeNode \| ... 19 more ... \| VariableStatement | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of replacePropertyName result T cannot be served by implementation result GetAccessorDeclaration \| MethodDeclaration \| MethodSignature \| PropertyAssignment \| PropertyDeclaration \| PropertySignature \| SetAccessorDeclaration | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of transformNamedEvaluation result Extract&lt;AssignmentExpression&lt;EqualsToken&gt; &amp; { readonly left: Identifier; readonly right: WrappedExpression&lt;AnonymousFunctionDefinition&gt;; }, Pick&lt;...&gt;&gt; \| ... 7 more ... \| Extract&lt;...&gt; cannot be served by implementation result BinaryExpression \| BindingElement \| ExportAssignment \| ParameterDeclaration \| PropertyAssignment \| PropertyDeclaration \| ShorthandPropertyAssignment \| VariableDeclaration | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 1 of visitArrayWorker parameter visitor cannot be served by implementation parameter visitor | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| overload 2 of setSerializerContextAnd parameter cb cannot be served by implementation parameter cb | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| reading accessModifier | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading add | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading allowTrailingComma | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading alwaysStrict | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading ambientModuleDeclare | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading ancestor | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading antecedent | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading arg | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading arrayArg | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading asPos | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading assertsModifier | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading assignedName | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading awaitToken | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading baseCandidates | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading baseType | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading bases | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading best | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading bestMatch | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading bindingList | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading blockedByExports | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading cacheAssignment | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading cacheKey | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading cached | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading callExpr | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading callTarget | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading callee | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading canSuggestTypeof | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading catchClause | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading check | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading classDeclarations | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading classThis | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading commentRange | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading commentRanges | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading commentText | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading commonJSPropertyAccess | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading components | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading conditionPrecedence | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading conditionalType | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading constantValue | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading containingQualifiedName | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading contextualType | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading ctor | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading currentFlags | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading currentNamespace | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading declName | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading declarationFilePath | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading declarationName | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading deduplicated | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading defaultDeclaration | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading defaultFallbackPolling | 1 | 1 | 1 | 1 | all: Source checker verifies object literal with an index-signature type. |
| reading defaultOnlyType | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading defaultReplaced | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading defaultType | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading deprecatedNode | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading deprecatedTag | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading dirPath | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading e | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading early | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading elementFlags | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading emitListItem | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading emitSuperHelpers | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading emittedCondition | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading enclosingBlockScopeContainer | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading enclosingContainer | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading end | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading entityName | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading errorInfo | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading excludeRegex | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading existingFlags | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading existingKind | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading existingPath | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading existingPending | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading existingSpecifier | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading exitNonUserCodeExpression | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading exitNonUserCodeStatement | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading exitStatus | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading expected | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading exportClause | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading expressionPrecedence | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading expressionResult | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading ext | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading extendedConfigCache | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading extendedConstraint | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading extendsType | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading extensionlessPriority | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading fallbackPolling | 1 | 1 | 1 | 1 | all: Source checker verifies object literal with an index-signature type. |
| reading fileInfos | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading fileName | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading fileSystemEntryExists | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading files | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading filesInError | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading following | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading freshTypeParameter | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading functionDeclaration | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading functionName | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading functionType | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading getCurrentDirectory | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading grandParent | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading hasEmptyObject | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading hasJSDocFunctionType | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading hasSyntheticDefault | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading helpOptions | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading helpers | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading heritageClauses | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading hostDeclaration | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading identifier | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading ifStatement | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading immediate | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading importClause | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading imports | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading indexedAccessType | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading inferences | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading initialElements | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading initializeExpression | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading initializerWithoutParens | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading innerAssignment | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading innerParameters | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading instanceProperties | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading instantiated | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading instantiatedTemplateType | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading instantiations | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading internalFlags | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading invokedExpression | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading isAnonymous | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading isCallToReadHelper | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading isData | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading isDerivedClass | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading isDirectExport | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading isEsmCjsRef | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading isExternalImportAlias | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading isFinite | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading isJsonEmittedToSameLocation | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading isLengthPushOrUnshift | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading isOptionalParameter | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading isProperty | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading isPrototypeAssignment | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading isSyncImport | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading isUntypedSignatureInJSFile | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading iteratorValueExpression | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading jsdocPropertyTag | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading jsonSourceFile | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading jsxImportSourcePragma | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading jsxImportSourcePragmas | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading jsxRuntimePragmas | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading last | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading lastElement | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading lastLeft | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading lastParamTags | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading lastPart | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading lastStatement | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading leadingComments | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading leftmostExpressionKind | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading libraryResolutionCache | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading limitedConstraint | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading lines | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading literalFiles | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading literals | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading local | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading localProps | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading localSymbol | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading localVsRemoteMap | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading location | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading mainExport | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading mapped | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading mappedSource | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading memberProps | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading membersSet | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading memoryUsed | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading mergedMembers | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading methodDeclaration | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading moduleBlock | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading moduleFileToTry | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading moduleName | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading moduleResolutionCache | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading moduleSpecifier | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading moduleTag | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading nameParts | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading namesOfType | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading needsConstructorBody | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading needsUniqueClassThis | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading neverProp | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading newConstraint | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading newDeclList | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading newFlags | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading newRoot | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading newTypes | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading next | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading nonParameterProperties | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading normalizedElements | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading ns | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading nsIndex | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading objectLiterals | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading oldEmitSignature | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading oldProgram | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading oldUncommittedSignature | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading operandPrecedence | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading operator | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading original | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading originalImplements | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading originalMethod | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading originalModuleSpecifier | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading originalNode | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading originalReadFile | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading otherOption | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading outerAssignment | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading output | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading overloadSignatures | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading packageJsonContent | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading packageRootPath | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading paramIdx | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading parameterNode | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading parameterProperties | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading parentType | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading path | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading paths | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading pattern | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading pollScheduled | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading possibleOption | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading postSuper | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading previousSibling | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading privateProp | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading propName | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading propNode | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading propertyAccess | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading providedKeys | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading reachable | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading readFileWithCache | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading realMembers | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading recursionIdentityMap | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading recursiveInnerModule | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading reexports | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading referenced | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading remainingStatements | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading remove | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading resolutionMode | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading resolvedModuleType | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading rest | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading returnStatement | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading rhsReference | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading seenSymbols | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading setReadFileCache | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading setterModifiers | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading setterName | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading sharedLength | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading shouldBundle | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading shouldSkip | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading skipped | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading snippetElement | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading sourceMapRange | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading sourceReturnType | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading span | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading startsOnNewLine | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading stateVariable | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading staticBlocks | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading str | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading stringNames | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading subcontext | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading substitutionType | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading supportedExtensionsWithJsonIfResolveJsonModule | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading symbolProps | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading symbols | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading targetNode | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading targetPropertyName | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading targetStr | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading targetSymbol | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading templateType | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading templates | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading terminalWidth | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading thisContainer | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading thisType | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading token | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading tokenSourceMapRanges | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading tracingEnabled | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading trailingComments | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading trailingNewlines | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading trailingParts | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading transformedStatements | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading tryStatement | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading tsPriority | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading typeAnnotation | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading typeArguments | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading typeKey | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading typeParameter | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading typeReferenceDirectiveResolutionCache | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading typeRoots | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading umdHeader | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading uniqueFilled | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading unwidenedType | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading updatedName | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading varDecls | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading variableDeclaration | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading variableDeclarator | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading verbatimFromExports | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| reading widenedTypes | 1 | 1 | 0 | 0 | unmarked: Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked. |
| the host binding performance, pending Runtime's clearance of its native column, Runtime and Platforms' clearance of its wasm32 column and #3sh9398's host model, | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column, | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| this outside a method | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |
| toLocaleTimeString: Date methods other than getTime/valueOf in the fs file host | 1 | 1 | 0 | 0 | unmarked: Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked. |

### Missing-stream recovery

The initial first-error recovery ignored LATENT_ONLY_FILE in its inherited driver and scanned unrelated files. It was explicitly stopped, then retried with a scratch file filter; unit registration and lowering were unchanged. All previously completed payloads were reused.

| Recovery piece | Mode | Target payloads | Child exit | Seconds | Stopping reason |
| --- | --- | ---: | ---: | ---: | --- |
| record-recovery-00 | full | 6 | 0 | 26.079 | completed |
| record-recovery-01 | full | 9 | 0 | 51.767 | completed |
| record-recovery-02 | full | 3 | 0 | 21.033 | completed |
| record-recovery-03 | full | 13 | 0 | 32.029 | completed |
| record-recovery-04 | full | 5 | 0 | 24.015 | completed |
| record-recovery-05 | full | 736 | 0 | 1336.728 | completed |
| record-recovery-06 | first-error | 4 | -15 | 535.69 | explicit stop: first-error path ignores LATENT_ONLY_FILE and scans unrelated files; retry missing payloads only with scratch file-selection filter |
| record-recovery-07 | full | 128 | 0 | 284.321 | completed |
| record-recovery-08 | first-error | 4 | 0 | 143.6 | completed |

### Record-shaped groups, top files and examples

#### failed statement; state rolled back; continued at next statement or declaration

7 record-shaped findings at 6 distinct sites, out of 12622 findings in this reason group.

Decision: Continuation bookkeeping, not an independent failure cause; marked only when the source checker verifies record-shaped context.

Top finding files: commandLineParser.ts: 4 findings, 3 sites; tsbuildPublic.ts: 1 findings, 1 sites; watchUtilities.ts: 1 findings, 1 sites; sys.ts: 1 findings, 1 sites.

`commandLineParser.ts:2530:21` (source starts at line 2529):
```ts
                if (returnValue) {
                    result![keyText] = value;
                }

                // Notify key value set, if user asked for it
```

`commandLineParser.ts:2994:13` (source starts at line 2993):
```ts
        if (hasProperty(options, name)) {
            result[name] = convertToOptionValueWithAbsolutePaths(
                optionsNameMap.get(name.toLowerCase()),
                options[name] as CompilerOptionsValue,
                toAbsolutePath,
```

#### a cast the runtime can't check

2 record-shaped findings at 2 distinct sites, out of 570 findings in this reason group.

Decision: Source checker verifies string-keyed read/write on an index-signature object. Unmarked: Reason and source inspection do not establish a record-shaped failure. Uncertain cases remain unmarked.

Top finding files: commandLineParser.ts: 2 findings, 2 sites.

`commandLineParser.ts:2810:27` (source starts at line 2809):
```ts
            }
            const value = options[name] as CompilerOptionsValue;
            const optionDefinition = optionsNameMap.get(name.toLowerCase());
            if (optionDefinition) {
                Debug.assert(optionDefinition.type !== "listOrElement");
```

`commandLineParser.ts:2996:17` (source starts at line 2995):
```ts
                optionsNameMap.get(name.toLowerCase()),
                options[name] as CompilerOptionsValue,
                toAbsolutePath,
            );
        }
```

#### a dictionary member narrowed to an array without an element-checked view

30 record-shaped findings at 19 distinct sites, out of 30 findings in this reason group.

Decision: Reason explicitly describes a record, dictionary or index-signature object.

Top finding files: moduleNameResolver.ts: 12 findings, 10 sites; program.ts: 11 findings, 3 sites; commandLineParser.ts: 2 findings, 2 sites; programDiagnostics.ts: 2 findings, 1 sites; utilities.ts: 1 findings, 1 sites.

`commandLineParser.ts:2871:9` (source starts at line 2870):
```ts
    emitOption("types", []);
    if (options.lib) {
        emitOption("lib", options.lib);
    }
    emitHeader(Diagnostics.For_nodejs_Colon);
```

`commandLineParser.ts:2949:23` (source starts at line 2948):
```ts

        const value = options[setting];
        const formattedValue = hasProperty(options, setting) && value === undefined ? "null" : formatValueOrArray(setting, (value ?? defaultValue) as PresetValue);
        if (comment) {
            result.push(`${tab}${tab}// "${setting}": ${String(formattedValue)},`);
```

#### optional record property access

9 record-shaped findings at 7 distinct sites, out of 9 findings in this reason group.

Decision: Reason explicitly describes a record, dictionary or index-signature object.

Top finding files: sys.ts: 4 findings, 4 sites; commandLineParser.ts: 3 findings, 1 sites; watchUtilities.ts: 1 findings, 1 sites; builder.ts: 1 findings, 1 sites.

`builder.ts:1200:14` (source starts at line 1199):
```ts
export function isIncrementalBundleEmitBuildInfo(info: IncrementalBuildInfo): info is IncrementalBundleEmitBuildInfo {
    return !!info.options?.outFile;
}

/** @internal */
```

`commandLineParser.ts:3447:9` (source starts at line 3446):
```ts

    if (ownConfig.options?.paths) {
        // If we end up needing to resolve relative paths from 'paths' relative to
        // the config file location, we'll need to know where that config file was.
        // Since 'paths' can be inherited from an extended config in another directory,
```

#### a union or optional field read in a program with record storage

8 record-shaped findings at 8 distinct sites, out of 8 findings in this reason group.

Decision: Reason explicitly describes a record, dictionary or index-signature object.

Top finding files: moduleNameResolver.ts: 5 findings, 5 sites; factory/nodeFactory.ts: 3 findings, 3 sites.

`factory/nodeFactory.ts:6551:73` (source starts at line 6550):
```ts
        const properties: PropertyAssignment[] = [];
        tryAddPropertyAssignment(properties, "enumerable", asExpression(attributes.enumerable));
        tryAddPropertyAssignment(properties, "configurable", asExpression(attributes.configurable));

        let isData = tryAddPropertyAssignment(properties, "writable", asExpression(attributes.writable));
```

`factory/nodeFactory.ts:6552:75` (source starts at line 6551):
```ts
        tryAddPropertyAssignment(properties, "enumerable", asExpression(attributes.enumerable));
        tryAddPropertyAssignment(properties, "configurable", asExpression(attributes.configurable));

        let isData = tryAddPropertyAssignment(properties, "writable", asExpression(attributes.writable));
        isData = tryAddPropertyAssignment(properties, "value", attributes.value) || isData;
```

#### a checked predicate target whose base has no required field to check (CompilerOptions &amp; { types: string[]; })

7 record-shaped findings at 1 distinct sites, out of 7 findings in this reason group.

Decision: Predicate target is CompilerOptions, whose types.ts declaration has a string index signature.

Top finding files: utilities.ts: 7 findings, 1 sites.

`utilities.ts:9026:1` (source starts at line 9025):
```ts
 */
export function usesWildcardTypes(options: CompilerOptions): options is CompilerOptions & { types: string[]; } {
    return some(options.types, t => t === "*");
}

```

Only one distinct site exists in this group; a second distinct example is unavailable.

#### reading extendsRaw

6 record-shaped findings at 2 distinct sites, out of 6 findings in this reason group.

Decision: Source checker verifies string-keyed read/write on an index-signature object.

Top finding files: commandLineParser.ts: 6 findings, 2 sites.

`commandLineParser.ts:3485:39` (source starts at line 3484):
```ts
                if (ownRaw[propertyName]) return; // No need to calculate if already set in own config
                const extendedSpecs = extendsRaw[propertyName];
                if (extendedSpecs) {
                    if (!isArray(extendedSpecs) || !extendedSpecs.every(isString)) throw new TypeError("Expected inherited string specs");
                    result[propertyName] = map(extendedSpecs, (path: string) =>
```

`commandLineParser.ts:3500:17` (source starts at line 3499):
```ts
            setPropertyInResultIfNotUndefined("files");
            if (extendsRaw.compileOnSave !== undefined) {
                result.compileOnSave = extendsRaw.compileOnSave;
            }
            assign(result.options, extendedConfig.options);
```

#### reading newParsedCommandLine

2 record-shaped findings at 1 distinct sites, out of 6 findings in this reason group.

Decision: Source checker verifies string-keyed read/write on an index-signature object. Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked.

Top finding files: program.ts: 2 findings, 1 sites.

`program.ts:1298:67` (source starts at line 1297):
```ts
            // If change in source file
            if (oldResolvedRef.commandLine.options.configFile !== newParsedCommandLine.options.configFile) return false;

            // check file names
            if (!arrayIsEqualTo(oldResolvedRef.commandLine.fileNames, newParsedCommandLine.fileNames)) return false;
```

Only one distinct site exists in this group; a second distinct example is unavailable.

#### reading buildOptions

4 record-shaped findings at 4 distinct sites, out of 5 findings in this reason group.

Decision: Source checker verifies string-keyed read/write on an index-signature object. Unmarked: Generic read reason; no verified index-signature object operation or typed record literal at this site. Uncertain cases remain unmarked.

Top finding files: commandLineParser.ts: 4 findings, 4 sites.

`commandLineParser.ts:2190:9` (source starts at line 2189):
```ts
    // Nonsensical combinations
    if (buildOptions.clean && buildOptions.force) {
        errors.push(createCompilerDiagnostic(Diagnostics.Options_0_and_1_cannot_be_combined, "clean", "force"));
    }
    if (buildOptions.clean && buildOptions.verbose) {
```

`commandLineParser.ts:2193:9` (source starts at line 2192):
```ts
    }
    if (buildOptions.clean && buildOptions.verbose) {
        errors.push(createCompilerDiagnostic(Diagnostics.Options_0_and_1_cannot_be_combined, "clean", "verbose"));
    }
    if (buildOptions.clean && buildOptions.watch) {
```

#### overload 1 of convertOptionsFromJson result WatchOptions \| undefined cannot be served by implementation result CompilerOptions \| TypeAcquisition \| WatchOptions \| undefined

2 record-shaped findings at 1 distinct sites, out of 2 findings in this reason group.

Decision: Result types CompilerOptions, TypeAcquisition and WatchOptions all declare string index signatures in types.ts.

Top finding files: commandLineParser.ts: 2 findings, 1 sites.

`commandLineParser.ts:3797:1` (source starts at line 3796):
```ts

function convertOptionsFromJson<TRaw>(optionsNameMap: Map<string, CommandLineOption>, jsonOptions: TRaw, basePath: string, defaultOptions: undefined, diagnostics: DidYouMeanOptionsDiagnostics, errors: Diagnostic[]): WatchOptions | undefined;
function convertOptionsFromJson<TRaw>(optionsNameMap: Map<string, CommandLineOption>, jsonOptions: TRaw, basePath: string, defaultOptions: CompilerOptions | TypeAcquisition, diagnostics: DidYouMeanOptionsDiagnostics, errors: Diagnostic[]): CompilerOptions | TypeAcquisition | undefined;
function convertOptionsFromJson<TRaw>(optionsNameMap: Map<string, CommandLineOption>, jsonOptions: TRaw, basePath: string, defaultOptions: CompilerOptions | TypeAcquisition | WatchOptions | undefined, diagnostics: DidYouMeanOptionsDiagnostics, errors: Diagnostic[]) {
    if (!jsonOptions) {
```

Only one distinct site exists in this group; a second distinct example is unavailable.

#### .types on a value

1 record-shaped findings at 1 distinct sites, out of 1 findings in this reason group.

Decision: Source checker verifies string-keyed read/write on an index-signature object.

Top finding files: moduleNameResolver.ts: 1 findings, 1 sites.

`moduleNameResolver.ts:849:32` (source starts at line 848):
```ts
    // in the wildcard in the position it was specified in the types array
    return deduplicate(flatten(options.types.map(t => t === "*" ? wildcardMatches : t)), equateValues);
}

export interface TypeReferenceDirectiveResolutionCache extends PerDirectoryResolutionCache<ResolvedTypeReferenceDirectiveWithFailedLookupLocations>, NonRelativeNameResolutionCache<ResolvedTypeReferenceDirectiveWithFailedLookupLocations>, PackageJsonInfoCache {
```

Only one distinct site exists in this group; a second distinct example is unavailable.

#### a readonly view read by a run-time key in a program with a class getter that does more than return its fields and literals joined by operators (the getter runs inside the read, unseen by the passes that follow calls)

1 record-shaped findings at 1 distinct sites, out of 1 findings in this reason group.

Decision: Source checker verifies string-keyed read/write on an index-signature object.

Top finding files: debug.ts: 1 findings, 1 sites.

`debug.ts:412:27` (source starts at line 411):
```ts
        for (const name in enumObject) {
            const value = enumObject[name];
            if (typeof value === "number") {
                result.push([value, name]);
            }
```

Only one distinct site exists in this group; a second distinct example is unavailable.

#### a spread into record storage from a fixed object

1 record-shaped findings at 1 distinct sites, out of 1 findings in this reason group.

Decision: Reason explicitly describes a record, dictionary or index-signature object.

Top finding files: tracing.ts: 1 findings, 1 sites.

`tracing.ts:183:44` (source starts at line 182):
```ts
        else if (sampleInterval - (time % sampleInterval) <= endTime - time) {
            writeEvent("X", phase, name, { ...args, results }, `"dur":${endTime - time}`, time);
        }
    }

```

Only one distinct site exists in this group; a second distinct example is unavailable.

#### a value of type Record&lt;number, FlowGraphNode&gt;

1 record-shaped findings at 1 distinct sites, out of 1 findings in this reason group.

Decision: Reason explicitly describes a record, dictionary or index-signature object.

Top finding files: debug.ts: 1 findings, 1 sites.

`debug.ts:959:15` (source starts at line 958):
```ts

        const links: Record<number, FlowGraphNode> = Object.create(/*o*/ null); // eslint-disable-line no-restricted-syntax
        const nodes: FlowGraphNode[] = [];
        const edges: FlowGraphEdge[] = [];
        const root = buildGraphNode(flowNode, new Set());
```

Only one distinct site exists in this group; a second distinct example is unavailable.

#### an unrestricted dictionary write that can violate named property all; use a literal key or a value accepted by every named property

1 record-shaped findings at 1 distinct sites, out of 1 findings in this reason group.

Decision: Reason explicitly describes a record, dictionary or index-signature object.

Top finding files: tsbuildPublic.ts: 1 findings, 1 sites.

`tsbuildPublic.ts:333:53` (source starts at line 332):
```ts
    commonOptionsWithBuild.forEach(option => {
        if (hasProperty(buildOptions, option.name)) result[option.name] = buildOptions[option.name];
    });
    result.tscBuild = true;
    return result;
```

Only one distinct site exists in this group; a second distinct example is unavailable.

#### an unrestricted dictionary write that can violate named property configFile; use a literal key or a value accepted by every named property

1 record-shaped findings at 1 distinct sites, out of 1 findings in this reason group.

Decision: Reason explicitly describes a record, dictionary or index-signature object.

Top finding files: commandLineParser.ts: 1 findings, 1 sites.

`commandLineParser.ts:3052:99` (source starts at line 3051):
```ts
    if (configFile) {
        Object.defineProperty(options, "configFile", { enumerable: false, writable: false, value: configFile });
    }
}

```

Only one distinct site exists in this group; a second distinct example is unavailable.

#### reading defaultFallbackPolling

1 record-shaped findings at 1 distinct sites, out of 1 findings in this reason group.

Decision: Source checker verifies object literal with an index-signature type.

Top finding files: sys.ts: 1 findings, 1 sites.

`sys.ts:1103:30` (source starts at line 1102):
```ts
            watchFile,
            fallbackPolling: defaultFallbackPolling === undefined ?
                fallbackPolling :
                defaultFallbackPolling,
        };
```

Only one distinct site exists in this group; a second distinct example is unavailable.

#### reading fallbackPolling

1 record-shaped findings at 1 distinct sites, out of 1 findings in this reason group.

Decision: Source checker verifies object literal with an index-signature type.

Top finding files: watchUtilities.ts: 1 findings, 1 sites.

`watchUtilities.ts:838:20` (source starts at line 837):
```ts
    return {
        watchFile: fallbackPolling !== undefined ?
            fallbackPolling as unknown as WatchFileKind :
            WatchFileKind.PriorityPollingInterval,
    };
```

Only one distinct site exists in this group; a second distinct example is unavailable.

Reproduce source classification and this section without lowering:
```sh
python3 stage3/census/graph-boundary-bcbe5d12/record-findings.py --prepare
node stage3/census/graph-boundary-bcbe5d12/classify-record-sites.cjs /workspace/scratch/jawqwv7/record-all-query-sites.json /workspace/scratch/jawqwv7/record-all-source-classification.json
python3 stage3/census/graph-boundary-bcbe5d12/record-findings.py --source-inspection /workspace/scratch/jawqwv7/record-all-source-classification.json
```
