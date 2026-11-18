# Whole construction/helper evidence

Original source, with CRLF normalized to LF only in these Markdown excerpts. Original exact strings and locations are retained in the compressed JSON evidence.

## 1. Node

`utilities.ts:8507:1`

```ts
function Node(this: Mutable<Node>, kind: SyntaxKind, pos: number, end: number): void {
    // Note: if modifying this, be sure to update NodeObject in src/services/services.ts
    this.pos = pos;
    this.end = end;
    this.kind = kind;
    this.id = 0;
    this.flags = NodeFlags.None;
    this.modifierFlagsCache = ModifierFlags.None;
    this.transformFlags = TransformFlags.None;
    this.parent = undefined!;
    this.original = undefined;
    this.emitNode = undefined;
}
```

## 2. Token

`utilities.ts:8521:1`

```ts
function Token(this: Mutable<Node>, kind: SyntaxKind, pos: number, end: number): void {
    // Note: if modifying this, be sure to update TokenOrIdentifierObject in src/services/services.ts
    this.pos = pos;
    this.end = end;
    this.kind = kind;
    this.id = 0;
    this.flags = NodeFlags.None;
    this.transformFlags = TransformFlags.None;
    this.parent = undefined!;
    this.emitNode = undefined;
}
```

## 3. Identifier

`utilities.ts:8533:1`

```ts
function Identifier(this: Mutable<Node>, kind: SyntaxKind, pos: number, end: number): void {
    // Note: if modifying this, be sure to update TokenOrIdentifierObject in src/services/services.ts
    this.pos = pos;
    this.end = end;
    this.kind = kind;
    this.id = 0;
    this.flags = NodeFlags.None;
    this.transformFlags = TransformFlags.None;
    this.parent = undefined!;
    this.original = undefined;
    this.emitNode = undefined;
}
```

## 4. Type

`utilities.ts:8491:1`

```ts
function Type(this: Type, checker: TypeChecker, flags: TypeFlags): void {
    // Note: if modifying this, be sure to update TypeObject in src/services/services.ts
    this.flags = flags;
    if (Debug.isDebugging || tracing) {
        this.checker = checker;
    }
}
```

## 5. SourceMapSource

`utilities.ts:8546:1`

```ts
function SourceMapSource(this: Partial<SourceMapSource>, fileName: string, text: string, skipTrivia?: (pos: number) => number): void {
    // Note: if modifying this, be sure to update SourceMapSourceObject in src/services/services.ts
    this.fileName = fileName;
    this.text = text;
    this.skipTrivia = skipTrivia || (pos => pos);
}
```

## 6. moveRangePos

`utilities.ts:7869:1`

```ts
export function moveRangePos(range: Readonly<TextRange>, pos: number): TextRange {
    return createRange(pos, range.end);
}
```

## 7. moveRangePastModifiers

`utilities.ts:7890:1`

```ts
export function moveRangePastModifiers(node: Node): TextRange {
    if (isPropertyDeclaration(node) || isMethodDeclaration(node)) {
        return moveRangePos(node, node.name.pos);
    }

    const lastModifier = canHaveModifiers(node) ? lastOrUndefined(node.modifiers) : undefined;
    return lastModifier && !positionIsSynthesized(lastModifier.end)
        ? moveRangePos(node, lastModifier.end)
        : moveRangePastDecorators(node);
}
```

## 8. createBaseNodeFactory

`factory/baseNodeFactory.ts:26:1`

```ts
export function createBaseNodeFactory(): BaseNodeFactory {
    let NodeConstructor: new (kind: SyntaxKind, pos: number, end: number) => Node;
    let TokenConstructor: new (kind: SyntaxKind, pos: number, end: number) => Node;
    let IdentifierConstructor: new (kind: SyntaxKind.Identifier, pos: number, end: number) => Node;
    let PrivateIdentifierConstructor: new (kind: SyntaxKind.PrivateIdentifier, pos: number, end: number) => Node;
    let SourceFileConstructor: new (kind: SyntaxKind.SourceFile, pos: number, end: number) => Node;

    return {
        createBaseSourceFileNode,
        createBaseIdentifierNode,
        createBasePrivateIdentifierNode,
        createBaseTokenNode,
        createBaseNode,
    };

    function createBaseSourceFileNode(kind: SyntaxKind.SourceFile): Node {
        return new (SourceFileConstructor || (SourceFileConstructor = objectAllocator.getSourceFileConstructor()))(kind, /*pos*/ -1, /*end*/ -1);
    }

    function createBaseIdentifierNode(kind: SyntaxKind.Identifier): Node {
        return new (IdentifierConstructor || (IdentifierConstructor = objectAllocator.getIdentifierConstructor()))(kind, /*pos*/ -1, /*end*/ -1);
    }

    function createBasePrivateIdentifierNode(kind: SyntaxKind.PrivateIdentifier): Node {
        return new (PrivateIdentifierConstructor || (PrivateIdentifierConstructor = objectAllocator.getPrivateIdentifierConstructor()))(kind, /*pos*/ -1, /*end*/ -1);
    }

    function createBaseTokenNode(kind: SyntaxKind): Node {
        return new (TokenConstructor || (TokenConstructor = objectAllocator.getTokenConstructor()))(kind, /*pos*/ -1, /*end*/ -1);
    }

    function createBaseNode(kind: SyntaxKind): Node {
        return new (NodeConstructor || (NodeConstructor = objectAllocator.getNodeConstructor()))(kind, /*pos*/ -1, /*end*/ -1);
    }
}
```

## 9. createIdentifier

`factory/nodeFactory.ts:1326:5`

```ts
function createIdentifier(text: string, originalKeywordKind?: SyntaxKind, hasExtendedUnicodeEscape?: boolean): Identifier {
        if (originalKeywordKind === undefined && text) {
            originalKeywordKind = stringToToken(text);
        }
        if (originalKeywordKind === SyntaxKind.Identifier) {
            originalKeywordKind = undefined;
        }

        const node = createBaseIdentifier(escapeLeadingUnderscores(text));
        if (hasExtendedUnicodeEscape) node.flags |= NodeFlags.IdentifierHasExtendedUnicodeEscape;

        // NOTE: we do not include transform flags of typeArguments in an identifier as they do not contribute to transformations
        if (node.escapedText === "await") {
            node.transformFlags |= TransformFlags.ContainsPossibleTopLevelAwait;
        }
        if (node.flags & NodeFlags.IdentifierHasExtendedUnicodeEscape) {
            node.transformFlags |= TransformFlags.ContainsES2015;
        }

        return node;
    }
```

## 10. createBigIntLiteral

`factory/nodeFactory.ts:1244:5`

```ts
function createBigIntLiteral(value: string | PseudoBigInt): BigIntLiteral {
        const node = createBaseToken<BigIntLiteral>(SyntaxKind.BigIntLiteral);
        node.text = typeof value === "string" ? value : pseudoBigIntToString(value) + "n";
        node.transformFlags |= TransformFlags.ContainsES2020;
        return node;
    }
```

## 11. createVariableDeclaration

`factory/nodeFactory.ts:4251:5`

```ts
function createVariableDeclaration(name: string | BindingName, exclamationToken: ExclamationToken | undefined, type: TypeNode | undefined, initializer: Expression | undefined) {
        const node = createBaseDeclaration<VariableDeclaration>(SyntaxKind.VariableDeclaration);
        node.name = asName(name);
        node.exclamationToken = exclamationToken;
        node.type = type;
        node.initializer = asInitializer(initializer);
        node.transformFlags |= propagateNameFlags(node.name) |
            propagateChildFlags(node.initializer) |
            (node.exclamationToken ?? node.type ? TransformFlags.ContainsTypeScript : TransformFlags.None);

        node.jsDoc = undefined; // initialized by parser (JsDocContainer)
        return node;
    }
```

## 12. createVariableStatement

`factory/nodeFactory.ts:3876:5`

```ts
function createVariableStatement(modifiers: readonly ModifierLike[] | undefined, declarationList: VariableDeclarationList | readonly VariableDeclaration[]) {
        const node = createBaseNode<VariableStatement>(SyntaxKind.VariableStatement);
        node.modifiers = asNodeArray(modifiers);
        node.declarationList = isArray(declarationList) ? createVariableDeclarationList(declarationList) : declarationList;
        node.transformFlags |= propagateChildrenFlags(node.modifiers) |
            propagateChildFlags(node.declarationList);
        if (modifiersToFlags(node.modifiers) & ModifierFlags.Ambient) {
            node.transformFlags = TransformFlags.ContainsTypeScript;
        }

        node.jsDoc = undefined; // initialized by parser (JsDocContainer)
        node.flowNode = undefined; // initialized by binder (FlowContainer)
        return node;
    }
```

## 13. createArrayBindingPattern

`factory/nodeFactory.ts:2799:5`

```ts
function createArrayBindingPattern(elements: readonly ArrayBindingElement[]) {
        const node = createBaseNode<ArrayBindingPattern>(SyntaxKind.ArrayBindingPattern);
        node.elements = createNodeArray(elements);
        node.transformFlags |= propagateChildrenFlags(node.elements) |
            TransformFlags.ContainsES2015 |
            TransformFlags.ContainsBindingPattern;
        return node;
    }
```

## 14. createClassDeclaration

`factory/nodeFactory.ts:4383:5`

```ts
function createClassDeclaration(
        modifiers: readonly ModifierLike[] | undefined,
        name: string | Identifier | undefined,
        typeParameters: readonly TypeParameterDeclaration[] | undefined,
        heritageClauses: readonly HeritageClause[] | undefined,
        members: readonly ClassElement[],
    ) {
        const node = createBaseDeclaration<ClassDeclaration>(SyntaxKind.ClassDeclaration);
        node.modifiers = asNodeArray(modifiers);
        node.name = asName(name);
        node.typeParameters = asNodeArray(typeParameters);
        node.heritageClauses = asNodeArray(heritageClauses);
        node.members = createNodeArray(members);

        if (modifiersToFlags(node.modifiers) & ModifierFlags.Ambient) {
            node.transformFlags = TransformFlags.ContainsTypeScript;
        }
        else {
            node.transformFlags |= propagateChildrenFlags(node.modifiers) |
                propagateNameFlags(node.name) |
                propagateChildrenFlags(node.typeParameters) |
                propagateChildrenFlags(node.heritageClauses) |
                propagateChildrenFlags(node.members) |
                (node.typeParameters ? TransformFlags.ContainsTypeScript : TransformFlags.None) |
                TransformFlags.ContainsES2015;
            if (node.transformFlags & TransformFlags.ContainsTypeScriptClassSyntax) {
                node.transformFlags |= TransformFlags.ContainsTypeScript;
            }
        }

        node.jsDoc = undefined; // initialized by parser (JsDocContainer)
        return node;
    }
```

## 15. createEnumDeclaration

`factory/nodeFactory.ts:4510:5`

```ts
function createEnumDeclaration(
        modifiers: readonly ModifierLike[] | undefined,
        name: string | Identifier,
        members: readonly EnumMember[],
    ) {
        const node = createBaseDeclaration<EnumDeclaration>(SyntaxKind.EnumDeclaration);
        node.modifiers = asNodeArray(modifiers);
        node.name = asName(name);
        node.members = createNodeArray(members);
        node.transformFlags |= propagateChildrenFlags(node.modifiers) |
            propagateChildFlags(node.name) |
            propagateChildrenFlags(node.members) |
            TransformFlags.ContainsTypeScript;
        node.transformFlags &= ~TransformFlags.ContainsPossibleTopLevelAwait; // Enum declarations cannot contain `await`

        node.jsDoc = undefined; // initialized by parser (JsDocContainer)
        return node;
    }
```

## 16. createModuleDeclaration

`factory/nodeFactory.ts:4544:5`

```ts
function createModuleDeclaration(
        modifiers: readonly ModifierLike[] | undefined,
        name: ModuleName,
        body: ModuleBody | undefined,
        flags = NodeFlags.None,
    ) {
        const node = createBaseDeclaration<ModuleDeclaration>(SyntaxKind.ModuleDeclaration);
        node.modifiers = asNodeArray(modifiers);
        node.flags |= flags & (NodeFlags.Namespace | NodeFlags.NestedNamespace | NodeFlags.GlobalAugmentation);
        node.name = name;
        node.body = body;
        if (modifiersToFlags(node.modifiers) & ModifierFlags.Ambient) {
            node.transformFlags = TransformFlags.ContainsTypeScript;
        }
        else {
            node.transformFlags |= propagateChildrenFlags(node.modifiers) |
                propagateChildFlags(node.name) |
                propagateChildFlags(node.body) |
                TransformFlags.ContainsTypeScript;
        }
        node.transformFlags &= ~TransformFlags.ContainsPossibleTopLevelAwait; // Module declarations cannot contain `await`.

        node.jsDoc = undefined; // initialized by parser (JsDocContainer)
        node.locals = undefined; // initialized by binder (LocalsContainer)
        node.nextContainer = undefined; // initialized by binder (LocalsContainer)
        return node;
    }
```

## 17. createParameterDeclaration

`factory/nodeFactory.ts:1645:5`

```ts
function createParameterDeclaration(
        modifiers: readonly ModifierLike[] | undefined,
        dotDotDotToken: DotDotDotToken | undefined,
        name: string | BindingName,
        questionToken?: QuestionToken,
        type?: TypeNode,
        initializer?: Expression,
    ) {
        const node = createBaseDeclaration<ParameterDeclaration>(SyntaxKind.Parameter);
        node.modifiers = asNodeArray(modifiers);
        node.dotDotDotToken = dotDotDotToken;
        node.name = asName(name);
        node.questionToken = questionToken;
        node.type = type;
        node.initializer = asInitializer(initializer);

        if (isThisIdentifier(node.name)) {
            node.transformFlags = TransformFlags.ContainsTypeScript;
        }
        else {
            node.transformFlags = propagateChildrenFlags(node.modifiers) |
                propagateChildFlags(node.dotDotDotToken) |
                propagateNameFlags(node.name) |
                propagateChildFlags(node.questionToken) |
                propagateChildFlags(node.initializer) |
                (node.questionToken ?? node.type ? TransformFlags.ContainsTypeScript : TransformFlags.None) |
                (node.dotDotDotToken ?? node.initializer ? TransformFlags.ContainsES2015 : TransformFlags.None) |
                (modifiersToFlags(node.modifiers) & ModifierFlags.ParameterPropertyModifier ? TransformFlags.ContainsTypeScriptClassSyntax : TransformFlags.None);
        }

        node.jsDoc = undefined; // initialized by parser (JsDocContainer)
        return node;
    }
```

## 18. createSuper

`factory/nodeFactory.ts:1521:5`

```ts
function createSuper() {
        return createToken(SyntaxKind.SuperKeyword);
    }
```

## 19. createType

`checker.ts:5521:5`

```ts
function createType(flags: TypeFlags): Type {
        const result = new Type(checker, flags);
        typeCount++;
        result.id = typeCount;
        tracing?.recordType(result);
        return result;
    }
```

## 20. createTypeParameter

`checker.ts:5575:5`

```ts
function createTypeParameter(symbol?: Symbol) {
        return createTypeWithSymbol(TypeFlags.TypeParameter, symbol!) as TypeParameter;
    }
```

## 21. createProgramHost

`watch.ts:843:1`

```ts
export function createProgramHost<T extends BuilderProgram = EmitAndSemanticDiagnosticsBuilderProgram>(system: System, createProgram: CreateProgram<T> | undefined): ProgramHost<T> {
    const getDefaultLibLocation = memoize(() => getDirectoryPath(normalizePath(system.getExecutingFilePath())));
    return {
        useCaseSensitiveFileNames: () => system.useCaseSensitiveFileNames,
        getNewLine: () => system.newLine,
        getCurrentDirectory: memoize(() => system.getCurrentDirectory()),
        getDefaultLibLocation,
        getDefaultLibFileName: options => combinePaths(getDefaultLibLocation(), getDefaultLibFileName(options)),
        fileExists: path => system.fileExists(path),
        readFile: (path, encoding) => system.readFile(path, encoding),
        directoryExists: path => system.directoryExists(path),
        getDirectories: path => system.getDirectories(path),
        readDirectory: (path, extensions, exclude, include, depth) => system.readDirectory(path, extensions, exclude, include, depth),
        realpath: system.realpath?.bind(system),
        getEnvironmentVariable: system.getEnvironmentVariable?.bind(system),
        trace: s => system.write(s + system.newLine),
        createDirectory: path => system.createDirectory(path),
        writeFile: (path, data, writeByteOrderMark) => system.writeFile(path, data, writeByteOrderMark),
        createHash: system.createHash?.bind(system),
        createProgram: createProgram || createEmitAndSemanticDiagnosticsBuilderProgram as any as CreateProgram<T>,
        storeSignatureInfo: system.storeSignatureInfo,
        now: system.now?.bind(system),
    };
}
```

## 22. createSolutionBuilderHostBase

`tsbuildPublic.ts:294:1`

```ts
function createSolutionBuilderHostBase<T extends BuilderProgram>(system: System, createProgram: CreateProgram<T> | undefined, reportDiagnostic?: DiagnosticReporter, reportSolutionBuilderStatus?: DiagnosticReporter) {
    const host = createProgramHost(system, createProgram) as SolutionBuilderHostBase<T>;
    host.getModifiedTime = system.getModifiedTime ? path => system.getModifiedTime!(path) : returnUndefined;
    host.setModifiedTime = system.setModifiedTime ? (path, date) => system.setModifiedTime!(path, date) : noop;
    host.deleteFile = system.deleteFile ? path => system.deleteFile!(path) : noop;
    host.reportDiagnostic = reportDiagnostic || createDiagnosticReporter(system);
    host.reportSolutionBuilderStatus = reportSolutionBuilderStatus || createBuilderStatusReporter(system);
    host.now = system.now?.bind(system); // For testing
    return host;
}
```

## 23. createCompilerHostWorker

`program.ts:450:1`

```ts
export function createCompilerHostWorker(
    options: CompilerOptions,
    setParentNodes?: boolean,
    system: System = sys,
): CompilerHost {
    const systemCreateDirectory = (path: string): void => system.createDirectory(path);
    const systemDirectoryExists = (path: string): boolean => system.directoryExists(path);
    const existingDirectories = new Map<string, boolean>();
    const getCanonicalFileName = createGetCanonicalFileName(system.useCaseSensitiveFileNames);
    function directoryExists(directoryPath: string): boolean {
        if (existingDirectories.has(directoryPath)) {
            return true;
        }
        if ((compilerHost.directoryExists || systemDirectoryExists)(directoryPath)) {
            existingDirectories.set(directoryPath, true);
            return true;
        }
        return false;
    }

    function getDefaultLibLocation(): string {
        return getDirectoryPath(normalizePath(system.getExecutingFilePath()));
    }

    const newLine = getNewLineCharacter(options);
    const realpath = system.realpath && ((path: string) => system.realpath!(path));
    const compilerHost: CompilerHost = {
        getSourceFile: createGetSourceFile(fileName => compilerHost.readFile(fileName), setParentNodes),
        getDefaultLibLocation,
        getDefaultLibFileName: options => combinePaths(getDefaultLibLocation(), getDefaultLibFileName(options)),
        writeFile: createWriteFileMeasuringIO(
            (path, data, writeByteOrderMark) => system.writeFile(path, data, writeByteOrderMark),
            path => (compilerHost.createDirectory || systemCreateDirectory)(path),
            path => directoryExists(path),
        ),
        getCurrentDirectory: memoize(() => system.getCurrentDirectory()),
        useCaseSensitiveFileNames: () => system.useCaseSensitiveFileNames,
        getCanonicalFileName,
        getNewLine: () => newLine,
        fileExists: fileName => system.fileExists(fileName),
        readFile: fileName => system.readFile(fileName),
        trace: (s: string) => system.write(s + newLine),
        directoryExists: directoryName => system.directoryExists(directoryName),
        getEnvironmentVariable: name => system.getEnvironmentVariable ? system.getEnvironmentVariable(name) : "",
        getDirectories: (path: string) => system.getDirectories(path),
        realpath,
        readDirectory: (path, extensions, include, exclude, depth) => system.readDirectory(path, extensions, include, exclude, depth),
        createDirectory: d => system.createDirectory(d),
        createHash: system.createHash?.bind(system),
    };
    return compilerHost;
}
```

## 24. mapDefined

`core.ts:494:1`

```ts
export function mapDefined<T, U>(array: readonly T[] | undefined, mapFn: (x: T, i: number) => U | undefined): U[] {
    const result: U[] = [];
    if (array !== undefined) {
        for (let i = 0; i < array.length; i++) {
            const mapped = mapFn(array[i]!, i);
            if (mapped !== undefined) {
                result.push(mapped);
            }
        }
    }
    return result;
}
```

## 25. serializeOptionBaseObject

`commandLineParser.ts:2795:1`

```ts
function serializeOptionBaseObject(
    options: OptionsBase,
    { optionsNameMap }: OptionsNameMap,
    pathOptions?: { configFilePath: string; useCaseSensitiveFileNames: boolean; },
): Map<string, CompilerOptionsValue> {
    const result = new Map<string, CompilerOptionsValue>();
    const getCanonicalFileName = pathOptions && createGetCanonicalFileName(pathOptions.useCaseSensitiveFileNames);

    for (const name in options) {
        if (hasProperty(options, name)) {
            // tsconfig only options cannot be specified via command line,
            // so we can assume that only types that can appear here string | number | boolean
            if (optionsNameMap.has(name) && (optionsNameMap.get(name)!.category === Diagnostics.Command_line_Options || optionsNameMap.get(name)!.category === Diagnostics.Output_Formatting)) {
                continue;
            }
            const value = options[name] as CompilerOptionsValue;
            const optionDefinition = optionsNameMap.get(name.toLowerCase());
            if (optionDefinition) {
                Debug.assert(optionDefinition.type !== "listOrElement");
                const customTypeMap = getCustomTypeMapOfCommandLineOption(optionDefinition);
                if (!customTypeMap) {
                    // There is no map associated with this compiler option then use the value as-is
                    // This is the case if the value is expect to be string, number, boolean or list of string
                    if (pathOptions && optionDefinition.isFilePath) {
                        Debug.assert(typeof value === "string");
                        result.set(name, getRelativePathFromFile(pathOptions.configFilePath, getNormalizedAbsolutePath(value, getDirectoryPath(pathOptions.configFilePath)), getCanonicalFileName!));
                    }
                    else if (pathOptions && optionDefinition.type === "list" && optionDefinition.element.isFilePath) {
                        result.set(name, (value as readonly string[]).map(v => getRelativePathFromFile(pathOptions.configFilePath, getNormalizedAbsolutePath(v, getDirectoryPath(pathOptions.configFilePath)), getCanonicalFileName!)));
                    }
                    else {
                        result.set(name, value);
                    }
                }
                else {
                    if (optionDefinition.type === "list") {
                        result.set(name, (value as readonly (string | number)[]).map(element => getNameOfCompilerOptionValue(element, customTypeMap)!)); // TODO: GH#18217
                    }
                    else {
                        // There is a typeMap associated with this command-line option so use it to map value back to its name
                        result.set(name, getNameOfCompilerOptionValue(value, customTypeMap));
                    }
                }
            }
        }
    }
    return result;
}
```

## 26. optionMapToObject

`commandLineParser.ts:2729:1`

```ts
export function optionMapToObject(optionMap: Map<string, CompilerOptionsValue>): object {
    return Object.fromEntries(optionMap);
}
```

## 27. getNamedPragmaArguments

`parser.ts:10764:1`

```ts
function getNamedPragmaArguments(pragma: PragmaDefinition, text: string | undefined): { [index: string]: string | undefined; } | "fail" {
    if (!text) return {};
    if (!pragma.args) return {};
    const args = text.trim().split(/\s+/) as string[];
    const argMap: { [index: string]: string | undefined; } = {};
    for (let i = 0; i < pragma.args.length; i++) {
        const argument = pragma.args[i]!;
        if (!args[i] && !argument.optional) {
            return "fail";
        }
        if (argument.captureSpan) {
            return Debug.fail("Capture spans not yet implemented for non-xml pragmas");
        }
        argMap[argument.name] = args[i];
    }
    return argMap;
}
```

## 28. tryGetModuleNameFromExportsOrImports

`moduleSpecifiers.ts:1020:1`

```ts
function tryGetModuleNameFromExportsOrImports(
    options: CompilerOptions,
    host: ModuleSpecifierResolutionHost,
    targetFilePath: string,
    packageDirectory: string,
    packageName: string,
    exports: unknown,
    conditions: string[],
    mode: MatchingMode,
    isImports: boolean,
    preferTsExtension: boolean,
): { moduleFileToTry: string; } | undefined {
    if (typeof exports === "string") {
        const ignoreCase = !hostUsesCaseSensitiveFileNames(host);
        const getCommonSourceDirectory = () => host.getCommonSourceDirectory();
        const outputFile = isImports && getOutputJSFileNameWorker(targetFilePath, options, ignoreCase, getCommonSourceDirectory);
        const declarationFile = isImports && getOutputDeclarationFileNameWorker(targetFilePath, options, ignoreCase, getCommonSourceDirectory);

        const pathOrPattern = getNormalizedAbsolutePath(combinePaths(packageDirectory, exports), /*currentDirectory*/ undefined);
        const extensionSwappedTarget = hasTSFileExtension(targetFilePath) ? removeFileExtension(targetFilePath) + tryGetJSExtensionForFile(targetFilePath, options) : undefined;
        const canTryTsExtension = preferTsExtension && hasImplementationTSFileExtension(targetFilePath);

        switch (mode) {
            case MatchingMode.Exact:
                if (
                    extensionSwappedTarget && comparePaths(extensionSwappedTarget, pathOrPattern, ignoreCase) === Comparison.EqualTo ||
                    comparePaths(targetFilePath, pathOrPattern, ignoreCase) === Comparison.EqualTo ||
                    outputFile && comparePaths(outputFile, pathOrPattern, ignoreCase) === Comparison.EqualTo ||
                    declarationFile && comparePaths(declarationFile, pathOrPattern, ignoreCase) === Comparison.EqualTo
                ) {
                    return { moduleFileToTry: packageName };
                }
                break;
            case MatchingMode.Directory:
                if (canTryTsExtension && containsPath(targetFilePath, pathOrPattern, ignoreCase)) {
                    const fragment = getRelativePathFromDirectory(pathOrPattern, targetFilePath, /*ignoreCase*/ false);
                    return { moduleFileToTry: getNormalizedAbsolutePath(combinePaths(combinePaths(packageName, exports), fragment), /*currentDirectory*/ undefined) };
                }
                if (extensionSwappedTarget && containsPath(pathOrPattern, extensionSwappedTarget, ignoreCase)) {
                    const fragment = getRelativePathFromDirectory(pathOrPattern, extensionSwappedTarget, /*ignoreCase*/ false);
                    return { moduleFileToTry: getNormalizedAbsolutePath(combinePaths(combinePaths(packageName, exports), fragment), /*currentDirectory*/ undefined) };
                }
                if (!canTryTsExtension && containsPath(pathOrPattern, targetFilePath, ignoreCase)) {
                    const fragment = getRelativePathFromDirectory(pathOrPattern, targetFilePath, /*ignoreCase*/ false);
                    return { moduleFileToTry: getNormalizedAbsolutePath(combinePaths(combinePaths(packageName, exports), fragment), /*currentDirectory*/ undefined) };
                }
                if (outputFile && containsPath(pathOrPattern, outputFile, ignoreCase)) {
                    const fragment = getRelativePathFromDirectory(pathOrPattern, outputFile, /*ignoreCase*/ false);
                    return { moduleFileToTry: combinePaths(packageName, fragment) };
                }
                if (declarationFile && containsPath(pathOrPattern, declarationFile, ignoreCase)) {
                    const fragment = changeFullExtension(getRelativePathFromDirectory(pathOrPattern, declarationFile, /*ignoreCase*/ false), getJSExtensionForFile(declarationFile, options));
                    return { moduleFileToTry: combinePaths(packageName, fragment) };
                }
                break;
            case MatchingMode.Pattern:
                const starPos = pathOrPattern.indexOf("*");
                const leadingSlice = pathOrPattern.slice(0, starPos);
                const trailingSlice = pathOrPattern.slice(starPos + 1);
                if (canTryTsExtension && startsWith(targetFilePath, leadingSlice, ignoreCase) && endsWith(targetFilePath, trailingSlice, ignoreCase)) {
                    const starReplacement = targetFilePath.slice(leadingSlice.length, targetFilePath.length - trailingSlice.length);
                    return { moduleFileToTry: replaceFirstStar(packageName, starReplacement) };
                }
                if (extensionSwappedTarget && startsWith(extensionSwappedTarget, leadingSlice, ignoreCase) && endsWith(extensionSwappedTarget, trailingSlice, ignoreCase)) {
                    const starReplacement = extensionSwappedTarget.slice(leadingSlice.length, extensionSwappedTarget.length - trailingSlice.length);
                    return { moduleFileToTry: replaceFirstStar(packageName, starReplacement) };
                }
                if (!canTryTsExtension && startsWith(targetFilePath, leadingSlice, ignoreCase) && endsWith(targetFilePath, trailingSlice, ignoreCase)) {
                    const starReplacement = targetFilePath.slice(leadingSlice.length, targetFilePath.length - trailingSlice.length);
                    return { moduleFileToTry: replaceFirstStar(packageName, starReplacement) };
                }
                if (outputFile && startsWith(outputFile, leadingSlice, ignoreCase) && endsWith(outputFile, trailingSlice, ignoreCase)) {
                    const starReplacement = outputFile.slice(leadingSlice.length, outputFile.length - trailingSlice.length);
                    return { moduleFileToTry: replaceFirstStar(packageName, starReplacement) };
                }
                if (declarationFile && startsWith(declarationFile, leadingSlice, ignoreCase) && endsWith(declarationFile, trailingSlice, ignoreCase)) {
                    const starReplacement = declarationFile.slice(leadingSlice.length, declarationFile.length - trailingSlice.length);
                    const substituted = replaceFirstStar(packageName, starReplacement);
                    const jsExtension = tryGetJSExtensionForFile(declarationFile, options);
                    return jsExtension ? { moduleFileToTry: changeFullExtension(substituted, jsExtension) } : undefined;
                }
                break;
        }
    }
    else if (Array.isArray(exports)) {
        return forEach(exports, e => tryGetModuleNameFromExportsOrImports(options, host, targetFilePath, packageDirectory, packageName, e, conditions, mode, isImports, preferTsExtension));
    }
    else if (typeof exports === "object" && exports !== null) { // eslint-disable-line no-restricted-syntax
        // conditional mapping
        for (const key of getOwnKeys(exports as MapLike<unknown>)) {
            if (key === "default" || conditions.indexOf(key) >= 0 || isApplicableVersionedTypesKey(conditions, key)) {
                const subTarget = (exports as MapLike<unknown>)[key];
                const result = tryGetModuleNameFromExportsOrImports(options, host, targetFilePath, packageDirectory, packageName, subTarget, conditions, mode, isImports, preferTsExtension);
                if (result) {
                    return result;
                }
            }
        }
    }
    return undefined;
}
```

## 29. tryGetModuleNameFromExports

`moduleSpecifiers.ts:1122:1`

```ts
function tryGetModuleNameFromExports(options: CompilerOptions, host: ModuleSpecifierResolutionHost, targetFilePath: string, packageDirectory: string, packageName: string, exports: unknown, conditions: string[]): { moduleFileToTry: string; } | undefined {
    if (typeof exports === "object" && exports !== null && !Array.isArray(exports) && allKeysStartWithDot(exports as MapLike<unknown>)) { // eslint-disable-line no-restricted-syntax
        // sub-mappings
        // 3 cases:
        // * directory mappings (legacyish, key ends with / (technically allows index/extension resolution under cjs mode))
        // * pattern mappings (contains a *)
        // * exact mappings (no *, does not end with /)
        return forEach(getOwnKeys(exports as MapLike<unknown>), k => {
            const subPackageName = getNormalizedAbsolutePath(combinePaths(packageName, k), /*currentDirectory*/ undefined);
            const mode = endsWith(k, "/") ? MatchingMode.Directory
                : k.includes("*") ? MatchingMode.Pattern
                : MatchingMode.Exact;
            return tryGetModuleNameFromExportsOrImports(options, host, targetFilePath, packageDirectory, subPackageName, (exports as MapLike<unknown>)[k], conditions, mode, /*isImports*/ false, /*preferTsExtension*/ false);
        });
    }
    return tryGetModuleNameFromExportsOrImports(options, host, targetFilePath, packageDirectory, packageName, exports, conditions, MatchingMode.Exact, /*isImports*/ false, /*preferTsExtension*/ false);
}
```
