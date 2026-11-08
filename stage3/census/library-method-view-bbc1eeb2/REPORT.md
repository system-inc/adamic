# Library method view: paired latent first-error census

Task sn788gx. Report only. No compiler or adaptation change is committed.

| Compiler | Pin |
| --- | --- |
| Parent and apply | `f7390e5b75975418eb6cbd1a8ee24e1f792fd4ca` |
| Head | `bbc1eeb2900538192d08ddb45b493d703e2b79ad` |

The head adds lib-declared method signatures to `libraryUnstoredMember`. The adapted input comes from the unmodified parent apply, using TypeScript 6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`. Both compiler binaries and their guarded overlays were built from their own pins; checker dependency `06af2e399c6c1fb93fd570b774027942747400b4` is shared unchanged.

The entry is `src/tsc/tsc.ts`, with 81 reached source files and 0 checker diagnostics on each compiler. Only latent **first-error** mode ran. `LATENT_ASSERT_NO_OUTPUT=1` checked both no-output guards. No full-core mode, replay, second pairing, native build or full gate ran.

## Counts

| View | Cleared findings | New findings | Changed findings | New sites | First-stop unit changes | Regression entries |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Raw | 0 | 74 | 0 | 74 | 0 | 74 |
| Distinct outermost access chains | 0 | 74 | 0 | 74 | 0 | 74 |

Raw means pair.py exact raw-position keys, not event multiplicity. Distinct uses the stock TypeScript 6.0.3 outermost-access-chain normalization. New findings include replacement reasons, so new and changed are not disjoint. A newly stopped unit was finding-free during latent lowering before and now has a lowering stop. A different first stop in an already-blocked unit is listed separately. This is a lowering observation, not proof of previous native correctness.

## Every newly stopped unit (0)

| Unit | Unit location | Head stop site | Exact head reason |
| --- | --- | --- | --- |
| None | | | |

## Every changed first stop in an already-blocked unit (0)

| Unit | Unit location | Parent site and exact reason | Head site and exact reason |
| --- | --- | --- | --- |
| None | | | |

## Cleared unit first stops (0)

None.

## Head method-view NotYet sites and source text

Each unique location and exact reason matching the guard template is quoted below. The JSON retains the recording unit, phase, exact site line and source context. Retained parent findings are labeled rather than attributed to the new method arm.

### compiler/builder.ts:1011:41

**New exact finding**. Unit: `handleDtsMayChangeOfReferencingExportOfAffectedFile` at `compiler/builder.ts:956:1`; phase `refusal_scan`.

a view naming has, which lib declares on ReadonlySet but the runtime's ReadonlySet doesn't hold (leave has out of the view's type)

Source lines 1009 through 1013; diagnostic column 41:

```typescript
        if (handleDtsMayChangeOfGlobalScope(state, exportedFromPath, invalidateJsFiles, cancellationToken, host)) return true;
        const references = state.referencedMap!.getKeys(exportedFromPath);
        return references && forEachKey(references, filePath =>
            handleDtsMayChangeOfFileAndExportsOfFile(
                state,
```

### compiler/builder.ts:394:41

**New exact finding**. Unit: `createBuilderProgramState` at `compiler/builder.ts:320:1`; phase `refusal_scan`.

a view naming has, which lib declares on ReadonlySet but the runtime's ReadonlySet doesn't hold (leave has out of the view's type)

Source lines 392 through 396; diagnostic column 41:

```typescript
            !hasSameKeys(newReferences = referencedMap && referencedMap.getValues(sourceFilePath), oldReferencedMap && oldReferencedMap.getValues(sourceFilePath)) ||
            // Referenced file was deleted in the new program
            newReferences && forEachKey(newReferences, path => !state.fileInfos.has(path) && oldState!.fileInfos.has(path))
        ) {
            // Register file as changed file and do not copy semantic diagnostics, since all changed files need to be re-evaluated
```

### compiler/checker.ts:11134:157

**New exact finding**. Unit: `createTypeChecker` at `compiler/checker.ts:1487:1`; phase `refusal_scan`.

a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type)

Source lines 11132 through 11136; diagnostic column 157:

```typescript
                    localName = "_exports";
                }
                localName = isIdentifierText(localName, languageVersion) && !isStringANonContextualKeyword(localName) ? localName : "_" + localName.replace(/[^a-z0-9]/gi, "_");
                return localName;
            }
```

### compiler/checker.ts:46735:36

**New exact finding**. Unit: `createTypeChecker` at `compiler/checker.ts:1487:1`; phase `refusal_scan`.

a view naming has, which lib declares on Map but the runtime's Map doesn't hold (leave has out of the view's type)

Source lines 46733 through 46737; diagnostic column 36:

```typescript
                    const blockLocals = catchClause.block.locals;
                    if (blockLocals) {
                        forEachKey(catchClause.locals!, caughtName => {
                            const blockLocal = blockLocals.get(caughtName);
                            if (blockLocal?.valueDeclaration && (blockLocal.flags & SymbolFlags.BlockScopedVariable) !== 0) {
```

### compiler/checker.ts:7878:140

**New exact finding**. Unit: `createTypeChecker` at `compiler/checker.ts:1487:1`; phase `refusal_scan`.

a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type)

Source lines 7876 through 7880; diagnostic column 140:

```typescript
                    const commentText = getTextOfJSDocComment(jsdocPropertyTag.comment);
                    if (commentText) {
                        setSyntheticLeadingComments(node, [{ kind: SyntaxKind.MultiLineCommentTrivia, text: "*\n * " + commentText.replace(/\n/g, "\n * ") + "\n ", pos: -1, end: -1, hasTrailingNewLine: true }]);
                    }
                }
```

### compiler/checker.ts:8063:127

**New exact finding**. Unit: `createTypeChecker` at `compiler/checker.ts:1487:1`; phase `refusal_scan`.

a view naming �@split@1091, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@split@1091 out of the view's type)

Source lines 8061 through 8065; diagnostic column 127:

```typescript
            }
            if (signature.declaration?.kind === SyntaxKind.JSDocSignature && signature.declaration.parent.kind === SyntaxKind.JSDocOverloadTag) {
                const comment = getTextOfNode(signature.declaration.parent.parent, /*includeTrivia*/ true).slice(2, -2).split(/\r\n|\n|\r/).map(line => line!.replace(/^\s+/, " ")).join("\n");
                addSyntheticLeadingComment(node, SyntaxKind.MultiLineCommentTrivia, comment, /*hasTrailingNewLine*/ true);
            }
```

### compiler/checker.ts:8063:167

**New exact finding**. Unit: `createTypeChecker` at `compiler/checker.ts:1487:1`; phase `refusal_scan`.

a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type)

Source lines 8061 through 8065; diagnostic column 167:

```typescript
            }
            if (signature.declaration?.kind === SyntaxKind.JSDocSignature && signature.declaration.parent.kind === SyntaxKind.JSDocOverloadTag) {
                const comment = getTextOfNode(signature.declaration.parent.parent, /*includeTrivia*/ true).slice(2, -2).split(/\r\n|\n|\r/).map(line => line!.replace(/^\s+/, " ")).join("\n");
                addSyntheticLeadingComment(node, SyntaxKind.MultiLineCommentTrivia, comment, /*hasTrailingNewLine*/ true);
            }
```

### compiler/checker.ts:8944:77

**New exact finding**. Unit: `createTypeChecker` at `compiler/checker.ts:1487:1`; phase `refusal_scan`.

a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type)

Source lines 8942 through 8946; diagnostic column 77:

```typescript
                    let expression: Expression | undefined;
                    if (isSingleOrDoubleQuote(firstChar) && !(symbol.flags & SymbolFlags.EnumMember)) {
                        const literalText = stripQuotes(symbolName).replace(/\\./g, s => s.substring(1));
                        context.approximateLength += literalText.length + 2; // "literalText"
                        expression = factory.createStringLiteral(literalText, firstChar === CharacterCodes.singleQuote);
```

### compiler/checker.ts:9920:126

**New exact finding**. Unit: `createTypeChecker` at `compiler/checker.ts:1487:1`; phase `refusal_scan`.

a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type)

Source lines 9918 through 9922; diagnostic column 126:

```typescript
                    setSyntheticLeadingComments(
                        factory.createTypeAliasDeclaration(/*modifiers*/ undefined, internalSymbolName, typeParamDecls, typeNode),
                        !commentText ? [] : [{ kind: SyntaxKind.MultiLineCommentTrivia, text: "*\n * " + commentText.replace(/\n/g, "\n * ") + "\n ", pos: -1, end: -1, hasTrailingNewLine: true }],
                    ),
                    modifierFlags,
```

### compiler/core.ts:1882:19

**New exact finding**. Unit: `toFileNameLowerCase` at `compiler/core.ts:1880:1`; phase `refusal_scan`.

a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type)

Source lines 1880 through 1884; diagnostic column 19:

```typescript
export function toFileNameLowerCase(x: string): string {
    return fileNameLowerCaseRegExp.test(x) ?
        x.replace(fileNameLowerCaseRegExp, toLowerCase) :
        x;
}
```

### compiler/debug.ts:563:69

**New exact finding**. Unit: `Debug` at `compiler/debug.ts:111:1`; phase `refusal_scan`.

a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type)

Source lines 561 through 565; diagnostic column 69:

```typescript
                        // We don't care, this is debug code that's only enabled with a debugger attached -
                        // we're just taking note of it for anyone checking regex performance in the future.
                        defaultValue = String(defaultValue).replace(/(?:,[\s\w]+:[^,]+)+\]$/, "]");
                        return `NodeArray ${defaultValue}`;
                    },
```

### compiler/emitter.ts:4105:42

**New exact finding**. Unit: `createPrinter` at `compiler/emitter.ts:1211:1`; phase `refusal_scan`.

a view naming �@split@1091, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@split@1091 out of the view's type)

Source lines 4103 through 4107; diagnostic column 42:

```typescript
            const text = getTextOfJSDocComment(node.comment);
            if (text) {
                const lines = text.split(/\r\n?|\n/) as string[];
                for (const line of lines) {
                    writeLine();
```

### compiler/emitter.ts:4976:34

**New exact finding**. Unit: `createPrinter` at `compiler/emitter.ts:1211:1`; phase `refusal_scan`.

a view naming �@split@1091, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@split@1091 out of the view's type)

Source lines 4974 through 4978; diagnostic column 34:

```typescript

    function writeLines(text: string): void {
        const lines = text.split(/\r\n?|\n/) as string[];
        const indentation = guessIndentation(lines);
        for (const lineText of lines) {
```

### compiler/moduleNameResolver.ts:2763:69

**New exact finding**. Unit: `getLoadModuleFromTargetExportOrImport` at `compiler/moduleNameResolver.ts:2751:1`; phase `refusal_scan`.

a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type)

Source lines 2761 through 2765; diagnostic column 69:

```typescript
            if (!startsWith(target, "./")) {
                if (isImports && !startsWith(target, "../") && !startsWith(target, "/") && !isRootedDiskPath(target)) {
                    const combinedLookup = pattern ? target.replace(/\*/g, subpath) : target + subpath;
                    traceIfEnabled(state, Diagnostics.Using_0_subpath_1_with_target_2, "imports", key, combinedLookup);
                    traceIfEnabled(state, Diagnostics.Resolving_module_0_from_1, combinedLookup, scope.packageDirectory + "/");
```

### compiler/moduleNameResolver.ts:2807:145

**New exact finding**. Unit: `getLoadModuleFromTargetExportOrImport` at `compiler/moduleNameResolver.ts:2751:1`; phase `refusal_scan`.

a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type)

Source lines 2805 through 2809; diagnostic column 145:

```typescript

            if (state.traceEnabled) {
                trace(state.host, Diagnostics.Using_0_subpath_1_with_target_2, isImports ? "imports" : "exports", key, pattern ? target.replace(/\*/g, subpath) : target + subpath);
            }
            const finalPath = toAbsolutePath(pattern ? resolvedTarget.replace(/\*/g, subpath) : resolvedTarget + subpath);
```

### compiler/moduleNameResolver.ts:2809:79

**New exact finding**. Unit: `getLoadModuleFromTargetExportOrImport` at `compiler/moduleNameResolver.ts:2751:1`; phase `refusal_scan`.

a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type)

Source lines 2807 through 2811; diagnostic column 79:

```typescript
                trace(state.host, Diagnostics.Using_0_subpath_1_with_target_2, isImports ? "imports" : "exports", key, pattern ? target.replace(/\*/g, subpath) : target + subpath);
            }
            const finalPath = toAbsolutePath(pattern ? resolvedTarget.replace(/\*/g, subpath) : resolvedTarget + subpath);
            const inputLink = tryLoadInputFileForPath(finalPath, subpath, combinePaths(scope.packageDirectory, "package.json"), isImports);
            if (inputLink) return inputLink;
```

### compiler/moduleSpecifiers.ts:152:64

**New exact finding**. Unit: `KindVariableStatement` at `compiler/moduleSpecifiers.ts:133:1`; phase `refusal_scan`.

a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type)

Source lines 150 through 154; diagnostic column 64:

```typescript
        }
        // Only case-insensitive and unicode flags make sense
        const flags = pattern.substring(lastSlash + 1).replace(/[^iu]/g, "");
        pattern = pattern.substring(1, lastSlash);
        return new RegExp(pattern, flags);
```

### compiler/parser.ts:10784:36

**New exact finding**. Unit: `getNamedPragmaArguments` at `compiler/parser.ts:10781:1`; phase `refusal_scan`.

a view naming �@split@1091, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@split@1091 out of the view's type)

Source lines 10782 through 10786; diagnostic column 36:

```typescript
    if (!text) return {};
    if (!pragma.args) return {};
    const args = text.trim().split(/\s+/) as string[];
    const argMap: { [index: string]: string | undefined; } = {};
    for (let i = 0; i < pragma.args.length; i++) {
```

### compiler/path.ts:528:24

**New exact finding**. Unit: `normalizeSlashes` at `compiler/path.ts:526:1`; phase `refusal_scan`.

a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type)

Source lines 526 through 530; diagnostic column 24:

```typescript
export function normalizeSlashes(path: string): string {
    return path.includes("\\")
        ? path.replace(backslashRegExp, directorySeparator)
        : path;
}
```

### compiler/path.ts:738:35

**New exact finding**. Unit: `simpleNormalizePath` at `compiler/path.ts:732:1`; phase `refusal_scan`.

a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type)

Source lines 736 through 740; diagnostic column 35:

```typescript
    }
    // Some paths only require cleanup of `/./` or leading `./`
    let simplified = path.replace(/\/\.\//g, "/");
    if (simplified.startsWith("./")) {
        simplified = simplified.slice(2);
```

### compiler/program.ts:4985:13

**New exact finding**. Unit: `updateHostForUseSourceOfProjectReferenceRedirect` at `compiler/program.ts:4892:1`; phase `refusal_scan`.

a view naming has, which lib declares on Set but the runtime's Set doesn't hold (leave has out of the view's type)

Source lines 4983 through 4987; diagnostic column 13:

```typescript
        const dirPathWithTrailingDirectorySeparator = `${dirPath}${directorySeparator}`;
        return forEachKey(
            setOfDeclarationDirectories!,
            declDirPath =>
                dirPath === declDirPath ||
```

### compiler/program.ts:5047:75

**New exact finding**. Unit: `updateHostForUseSourceOfProjectReferenceRedirect` at `compiler/program.ts:4892:1`; phase `refusal_scan`.

a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type)

Source lines 5045 through 5049; diagnostic column 75:

```typescript
                    symlinkCache.setSymlinkedFile(
                        fileOrDirectoryPath,
                        `${symlinkedDirectory.real}${absolutePath.replace(new RegExp(directoryPath, "i"), "")}`,
                    );
                }
```

### compiler/program.ts:737:43

**New exact finding**. Unit: `formatCodeSpan` at `compiler/program.ts:712:1`; phase `refusal_scan`.

a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type)

Source lines 735 through 739; diagnostic column 43:

```typescript
        let lineContent = file.text.slice(lineStart, lineEnd);
        lineContent = lineContent.trimEnd(); // trim from end
        lineContent = lineContent.replace(/\t/g, " "); // convert tabs to single spaces

        // Output the gutter and the actual contents of the line.
```

### compiler/program.ts:751:68

**New exact finding**. Unit: `formatCodeSpan` at `compiler/program.ts:712:1`; phase `refusal_scan`.

a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type)

Source lines 749 through 753; diagnostic column 68:

```typescript
            const lastCharForLine = i === lastLine ? lastLineChar : undefined;

            context += lineContent.slice(0, firstLineChar).replace(/\S/g, " ");
            context += lineContent.slice(firstLineChar, lastCharForLine).replace(/./g, "~");
        }
```

### compiler/program.ts:752:82

**New exact finding**. Unit: `formatCodeSpan` at `compiler/program.ts:712:1`; phase `refusal_scan`.

a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type)

Source lines 750 through 754; diagnostic column 82:

```typescript

            context += lineContent.slice(0, firstLineChar).replace(/\S/g, " ");
            context += lineContent.slice(firstLineChar, lastCharForLine).replace(/./g, "~");
        }
        else if (i === lastLine) {
```

### compiler/program.ts:755:67

**New exact finding**. Unit: `formatCodeSpan` at `compiler/program.ts:712:1`; phase `refusal_scan`.

a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type)

Source lines 753 through 757; diagnostic column 67:

```typescript
        }
        else if (i === lastLine) {
            context += lineContent.slice(0, lastLineChar).replace(/./g, "~");
        }
        else {
```

### compiler/program.ts:759:44

**New exact finding**. Unit: `formatCodeSpan` at `compiler/program.ts:712:1`; phase `refusal_scan`.

a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type)

Source lines 757 through 761; diagnostic column 44:

```typescript
        else {
            // Squiggle the entire line.
            context += lineContent.replace(/./g, "~");
        }
        context += resetEscapeSequence;
```

### compiler/resolutionCache.ts:1627:18

**New exact finding**. Unit: `createResolutionCache` at `compiler/resolutionCache.ts:579:1`; phase `refusal_scan`.

a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type)

Source lines 1625 through 1629; diagnostic column 18:

```typescript

    function closeTypeRootsWatch() {
        clearMap(typeRootsWatches, closeFileWatcher);
    }

```

### compiler/resolutionCache.ts:287:50

**New exact finding**. Unit: `perceivedOsRootLengthForWatching` at `compiler/resolutionCache.ts:269:1`; phase `refusal_scan`.

a view naming �@match@1085, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@match@1085 out of the view's type)

Source lines 285 through 289; diagnostic column 50:

```typescript
    if (
        isDosStyle &&
        !pathComponents[indexAfterOsRoot]!.match(/^users$/i)
    ) {
        // Paths like c:/notUsers
```

### compiler/resolutionCache.ts:293:49

**New exact finding**. Unit: `perceivedOsRootLengthForWatching` at `compiler/resolutionCache.ts:269:1`; phase `refusal_scan`.

a view naming �@match@1085, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@match@1085 out of the view's type)

Source lines 291 through 295; diagnostic column 49:

```typescript
    }

    if (pathComponents[indexAfterOsRoot]!.match(/^workspaces$/i)) {
        // Paths like: /workspaces as codespaces hoist the repos in /workspaces so we have to exempt these from "2" level from root rule
        return indexAfterOsRoot + 1;
```

### compiler/resolutionCache.ts:681:18

**New exact finding**. Unit: `createResolutionCache` at `compiler/resolutionCache.ts:579:1`; phase `refusal_scan`.

a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type)

Source lines 679 through 683; diagnostic column 18:

```typescript

    function clear() {
        clearMap(directoryWatchesOfFailedLookups, closeFileWatcherOf);
        clearMap(fileWatchesOfAffectingLocations, closeFileWatcherOf);
        isSymlinkCache.clear();
```

### compiler/resolutionCache.ts:682:18

**New exact finding**. Unit: `createResolutionCache` at `compiler/resolutionCache.ts:579:1`; phase `refusal_scan`.

a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type)

Source lines 680 through 684; diagnostic column 18:

```typescript
    function clear() {
        clearMap(directoryWatchesOfFailedLookups, closeFileWatcherOf);
        clearMap(fileWatchesOfAffectingLocations, closeFileWatcherOf);
        isSymlinkCache.clear();
        packageDirWatchers.clear();
```

### compiler/semver.ts:278:41

**New exact finding**. Unit: `parseRange` at `compiler/semver.ts:276:1`; phase `refusal_scan`.

a view naming �@split@1091, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@split@1091 out of the view's type)

Source lines 276 through 280; diagnostic column 41:

```typescript
function parseRange(text: string) {
    const alternatives: Comparator[][] = [];
    for (let range of text.trim().split(logicalOrRegExp)) {
        if (!range) continue;
        const comparators: Comparator[] = [];
```

### compiler/semver.ts:287:46

**New exact finding**. Unit: `parseRange` at `compiler/semver.ts:276:1`; phase `refusal_scan`.

a view naming �@split@1091, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@split@1091 out of the view's type)

Source lines 285 through 289; diagnostic column 46:

```typescript
        }
        else {
            for (const simple of range.split(whitespaceRegExp) as string[]) {
                const match = rangeRegExp.exec(simple.trim());
                if (!match || !parseComparator(match[1], match[2]!, comparators)) return undefined;
```

### compiler/sourcemap.ts:400:30

**New exact finding**. Unit: `tryGetSourceMappingURL` at `compiler/sourcemap.ts:392:1`; phase `refusal_scan`.

a view naming �@match@1085, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@match@1085 out of the view's type)

Source lines 398 through 402; diagnostic column 30:

```typescript
        }
        // If we see a non-whitespace/map comment-like line, break, to avoid scanning up the entire file
        else if (!line.match(whitespaceOrMapCommentRegExp)) {
            break;
        }
```

### compiler/sys.ts:1695:105

**New exact finding**. Unit: `KindVariableStatement` at `compiler/sys.ts:1469:1`; phase `refusal_scan`.

a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type)

Source lines 1693 through 1697; diagnostic column 105:

```typescript
                    if (!err) {
                        if (statSync(profilePath)?.isDirectory()) {
                            profilePath = _path.join(profilePath, `${(new Date()).toISOString().replace(/:/g, "-")}+P${process.pid}.cpuprofile`);
                        }
                        try {
```

### compiler/sys.ts:1729:30

**New exact finding**. Unit: `KindVariableStatement` at `compiler/sys.ts:1469:1`; phase `refusal_scan`.

a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type)

Source lines 1727 through 1731; diagnostic column 30:

```typescript
        /** Convert all lowercase chars to uppercase, and vice-versa */
        function swapCase(s: string): string {
            return s.replace(/\w/g, ch => {
                const up = ch.toUpperCase();
                return ch === up ? ch.toLowerCase() : up;
```

### compiler/transformers/jsx.ts:624:29

**New exact finding**. Unit: `transformJsx` at `compiler/transformers/jsx.ts:92:1`; phase `refusal_scan`.

a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type)

Source lines 622 through 626; diagnostic column 29:

```typescript
     */
    function decodeEntities(text: string): string {
        return text.replace(/&((#((\d+)|x([\da-fA-F]+)))|(\w+));/g, (match, _all, _number, _digits, decimal, hex, word) => {
            if (decimal) {
                return utf16EncodeAsString(parseInt(decimal, 10));
```

### compiler/transformers/taggedTemplate.ts:130:25

**New exact finding**. Unit: `getRawLiteral` at `compiler/transformers/taggedTemplate.ts:110:1`; phase `refusal_scan`.

a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type)

Source lines 128 through 132; diagnostic column 25:

```typescript
    // ES6 Spec 11.8.6.1 - Static Semantics of TV's and TRV's
    // <CR><LF> and <CR> LineTerminatorSequences are normalized to <LF> for both TV and TRV.
    text = text.replace(/\r\n?/g, "\n");
    return setTextRange(factory.createStringLiteral(text), node);
}
```

### compiler/transformers/ts.ts:2718:30

**New exact finding**. Unit: `transformTypeScript` at `compiler/transformers/ts.ts:235:1`; phase `refusal_scan`.

a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type)

Source lines 2716 through 2720; diagnostic column 30:

```typescript

    function safeMultiLineComment(value: string): string {
        return value.replace(/\*\//g, "*_/");
    }

```

### compiler/tsbuildPublic.ts:1774:9

**New exact finding**. Unit: `getUpToDateStatusWorker` at `compiler/tsbuildPublic.ts:1468:1`; phase `refusal_scan`.

a view naming has, which lib declares on Set but the runtime's Set doesn't hold (leave has out of the view's type)

Source lines 1772 through 1776; diagnostic column 9:

```typescript
    const packageJsonLookups = state.lastCachedPackageJsonLookups.get(resolvedPath);
    const dependentPackageFileStatus = packageJsonLookups && forEachKey(
        packageJsonLookups,
        path => checkConfigFileUpToDateStatus(state, path, oldestOutputFileTime, oldestOutputFileName),
    );
```

### compiler/tsbuildPublic.ts:2247:14

**New exact finding**. Unit: `stopWatching` at `compiler/tsbuildPublic.ts:2246:1`; phase `refusal_scan`.

a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type)

Source lines 2245 through 2249; diagnostic column 14:

```typescript

function stopWatching<T extends BuilderProgram>(state: SolutionBuilderState<T>) {
    clearMap(state.allWatchedConfigFiles, closeFileWatcher);
    clearMap(state.allWatchedExtendedConfigFiles, closeFileWatcherOf);
    clearMap(state.allWatchedWildcardDirectories, watchedWildcardDirectories => clearMap(watchedWildcardDirectories, closeFileWatcherOf));
```

### compiler/tsbuildPublic.ts:2248:14

**New exact finding**. Unit: `stopWatching` at `compiler/tsbuildPublic.ts:2246:1`; phase `refusal_scan`.

a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type)

Source lines 2246 through 2250; diagnostic column 14:

```typescript
function stopWatching<T extends BuilderProgram>(state: SolutionBuilderState<T>) {
    clearMap(state.allWatchedConfigFiles, closeFileWatcher);
    clearMap(state.allWatchedExtendedConfigFiles, closeFileWatcherOf);
    clearMap(state.allWatchedWildcardDirectories, watchedWildcardDirectories => clearMap(watchedWildcardDirectories, closeFileWatcherOf));
    clearMap(state.allWatchedInputFiles, watchedWildcardDirectories => clearMap(watchedWildcardDirectories, closeFileWatcher));
```

### compiler/tsbuildPublic.ts:2249:14

**New exact finding**. Unit: `stopWatching` at `compiler/tsbuildPublic.ts:2246:1`; phase `refusal_scan`.

a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type)

Source lines 2247 through 2251; diagnostic column 14:

```typescript
    clearMap(state.allWatchedConfigFiles, closeFileWatcher);
    clearMap(state.allWatchedExtendedConfigFiles, closeFileWatcherOf);
    clearMap(state.allWatchedWildcardDirectories, watchedWildcardDirectories => clearMap(watchedWildcardDirectories, closeFileWatcherOf));
    clearMap(state.allWatchedInputFiles, watchedWildcardDirectories => clearMap(watchedWildcardDirectories, closeFileWatcher));
    clearMap(state.allWatchedPackageJsonFiles, watchedPacageJsonFiles => clearMap(watchedPacageJsonFiles, closeFileWatcher));
```

### compiler/tsbuildPublic.ts:2249:90

**New exact finding**. Unit: `stopWatching` at `compiler/tsbuildPublic.ts:2246:1`; phase `refusal_scan`.

a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type)

Source lines 2247 through 2251; diagnostic column 90:

```typescript
    clearMap(state.allWatchedConfigFiles, closeFileWatcher);
    clearMap(state.allWatchedExtendedConfigFiles, closeFileWatcherOf);
    clearMap(state.allWatchedWildcardDirectories, watchedWildcardDirectories => clearMap(watchedWildcardDirectories, closeFileWatcherOf));
    clearMap(state.allWatchedInputFiles, watchedWildcardDirectories => clearMap(watchedWildcardDirectories, closeFileWatcher));
    clearMap(state.allWatchedPackageJsonFiles, watchedPacageJsonFiles => clearMap(watchedPacageJsonFiles, closeFileWatcher));
```

### compiler/tsbuildPublic.ts:2250:14

**New exact finding**. Unit: `stopWatching` at `compiler/tsbuildPublic.ts:2246:1`; phase `refusal_scan`.

a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type)

Source lines 2248 through 2252; diagnostic column 14:

```typescript
    clearMap(state.allWatchedExtendedConfigFiles, closeFileWatcherOf);
    clearMap(state.allWatchedWildcardDirectories, watchedWildcardDirectories => clearMap(watchedWildcardDirectories, closeFileWatcherOf));
    clearMap(state.allWatchedInputFiles, watchedWildcardDirectories => clearMap(watchedWildcardDirectories, closeFileWatcher));
    clearMap(state.allWatchedPackageJsonFiles, watchedPacageJsonFiles => clearMap(watchedPacageJsonFiles, closeFileWatcher));
}
```

### compiler/tsbuildPublic.ts:2250:81

**New exact finding**. Unit: `stopWatching` at `compiler/tsbuildPublic.ts:2246:1`; phase `refusal_scan`.

a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type)

Source lines 2248 through 2252; diagnostic column 81:

```typescript
    clearMap(state.allWatchedExtendedConfigFiles, closeFileWatcherOf);
    clearMap(state.allWatchedWildcardDirectories, watchedWildcardDirectories => clearMap(watchedWildcardDirectories, closeFileWatcherOf));
    clearMap(state.allWatchedInputFiles, watchedWildcardDirectories => clearMap(watchedWildcardDirectories, closeFileWatcher));
    clearMap(state.allWatchedPackageJsonFiles, watchedPacageJsonFiles => clearMap(watchedPacageJsonFiles, closeFileWatcher));
}
```

### compiler/tsbuildPublic.ts:2251:14

**New exact finding**. Unit: `stopWatching` at `compiler/tsbuildPublic.ts:2246:1`; phase `refusal_scan`.

a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type)

Source lines 2249 through 2253; diagnostic column 14:

```typescript
    clearMap(state.allWatchedWildcardDirectories, watchedWildcardDirectories => clearMap(watchedWildcardDirectories, closeFileWatcherOf));
    clearMap(state.allWatchedInputFiles, watchedWildcardDirectories => clearMap(watchedWildcardDirectories, closeFileWatcher));
    clearMap(state.allWatchedPackageJsonFiles, watchedPacageJsonFiles => clearMap(watchedPacageJsonFiles, closeFileWatcher));
}

```

### compiler/tsbuildPublic.ts:2251:83

**New exact finding**. Unit: `stopWatching` at `compiler/tsbuildPublic.ts:2246:1`; phase `refusal_scan`.

a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type)

Source lines 2249 through 2253; diagnostic column 83:

```typescript
    clearMap(state.allWatchedWildcardDirectories, watchedWildcardDirectories => clearMap(watchedWildcardDirectories, closeFileWatcherOf));
    clearMap(state.allWatchedInputFiles, watchedWildcardDirectories => clearMap(watchedWildcardDirectories, closeFileWatcher));
    clearMap(state.allWatchedPackageJsonFiles, watchedPacageJsonFiles => clearMap(watchedPacageJsonFiles, closeFileWatcher));
}

```

### compiler/utilities.ts:10159:29

**New exact finding**. Unit: `numberOfDirectorySeparators` at `compiler/utilities.ts:10158:1`; phase `refusal_scan`.

a view naming �@match@1085, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@match@1085 out of the view's type)

Source lines 10157 through 10161; diagnostic column 29:

```typescript

function numberOfDirectorySeparators(str: string) {
    const match = str.match(/\//g);
    return match ? match.length : 0;
}
```

### compiler/utilities.ts:10917:25

**New exact finding**. Unit: `escapeSnippetText` at `compiler/utilities.ts:10916:1`; phase `refusal_scan`.

a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type)

Source lines 10915 through 10919; diagnostic column 25:

```typescript
/** @internal */
export function escapeSnippetText(text: string): string {
    return text.replace(/\$/g, () => "\\$");
}

```

### compiler/utilities.ts:1331:27

**New exact finding**. Unit: `getTextOfNodeFromSourceText` at `compiler/utilities.ts:1322:1`; phase `refusal_scan`.

a view naming �@split@1091, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@split@1091 out of the view's type)

Source lines 1329 through 1333; diagnostic column 27:

```typescript
    if (isJSDocTypeExpressionOrChild(node)) {
        // strip space + asterisk at line start
        text = text.split(/\r\n|\n|\r/).map(line => line!.replace(/^\s*\*/, "").trimStart()).join("\n");
    }

```

### compiler/utilities.ts:1331:67

**New exact finding**. Unit: `getTextOfNodeFromSourceText` at `compiler/utilities.ts:1322:1`; phase `refusal_scan`.

a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type)

Source lines 1329 through 1333; diagnostic column 67:

```typescript
    if (isJSDocTypeExpressionOrChild(node)) {
        // strip space + asterisk at line start
        text = text.split(/\r\n|\n|\r/).map(line => line!.replace(/^\s*\*/, "").trimStart()).join("\n");
    }

```

### compiler/utilities.ts:2063:48

**New exact finding**. Unit: `makeIdentifierFromModuleName` at `compiler/utilities.ts:2062:1`; phase `refusal_scan`.

a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type)

Source lines 2061 through 2065; diagnostic column 48:

```typescript
/** @internal */
export function makeIdentifierFromModuleName(moduleName: string): string {
    return getBaseFileName(moduleName).replace(/^(\d)/, "_$1").replace(/\W/g, "_");
}

```

### compiler/utilities.ts:2063:72

**New exact finding**. Unit: `makeIdentifierFromModuleName` at `compiler/utilities.ts:2062:1`; phase `refusal_scan`.

a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type)

Source lines 2061 through 2065; diagnostic column 72:

```typescript
/** @internal */
export function makeIdentifierFromModuleName(moduleName: string): string {
    return getBaseFileName(moduleName).replace(/^(\d)/, "_$1").replace(/\W/g, "_");
}

```

### compiler/utilities.ts:6139:24

**New exact finding**. Unit: `escapeTemplateSubstitution` at `compiler/utilities.ts:6138:1`; phase `refusal_scan`.

a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type)

Source lines 6137 through 6141; diagnostic column 24:

```typescript
/** @internal */
export function escapeTemplateSubstitution(str: string): string {
    return str.replace(templateSubstitutionRegExp, "\\${");
}

```

### compiler/utilities.ts:6209:22

**New exact finding**. Unit: `escapeString` at `compiler/utilities.ts:6205:1`; phase `refusal_scan`.

a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type)

Source lines 6207 through 6211; diagnostic column 22:

```typescript
        quoteChar === CharacterCodes.singleQuote ? singleQuoteEscapedCharsRegExp :
        doubleQuoteEscapedCharsRegExp;
    return s.replace(escapedCharsRegExp, getReplacement);
}

```

### compiler/utilities.ts:6219:19

**New exact finding**. Unit: `escapeNonAsciiString` at `compiler/utilities.ts:6214:1`; phase `refusal_scan`.

a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type)

Source lines 6217 through 6221; diagnostic column 19:

```typescript
    // Otherwise just return the original string.
    return nonAsciiCharacters.test(s) ?
        s.replace(nonAsciiCharacters, c => encodeUtf16EscapeSequence(c.charCodeAt(0))) :
        s;
}
```

### compiler/utilities.ts:6250:22

**New exact finding**. Unit: `escapeJsxAttributeString` at `compiler/utilities.ts:6247:1`; phase `refusal_scan`.

a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type)

Source lines 6248 through 6252; diagnostic column 22:

```typescript
    const escapedCharsRegExp = quoteChar === CharacterCodes.singleQuote ? jsxSingleQuoteEscapedCharsRegExp :
        jsxDoubleQuoteEscapedCharsRegExp;
    return s.replace(escapedCharsRegExp, getJsxAttributeStringReplacement);
}

```

### compiler/utilities.ts:8585:25

**New exact finding**. Unit: `formatStringFromArgs` at `compiler/utilities.ts:8584:1`; phase `refusal_scan`.

a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type)

Source lines 8583 through 8587; diagnostic column 25:

```typescript
/** @internal */
export function formatStringFromArgs(text: string, args: DiagnosticArguments): string {
    return text.replace(/\{(\d+)\}/g, (_match, index: string) => "" + Debug.checkDefined(args[+index]));
}

```

### compiler/utilities.ts:9599:25

**New exact finding**. Unit: `regExpEscape` at `compiler/utilities.ts:9598:1`; phase `refusal_scan`.

a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type)

Source lines 9597 through 9601; diagnostic column 25:

```typescript
/** @internal */
export function regExpEscape(text: string): string {
    return text.replace(reservedCharacterPattern, escapeRegExpCharacter);
}

```

### compiler/utilities.ts:9747:55

**New exact finding**. Unit: `getSubPatternFromSpec` at `compiler/utilities.ts:9696:1`; phase `refusal_scan`.

a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type)

Source lines 9745 through 9749; diagnostic column 55:

```typescript
                }

                componentPattern += component.replace(reservedCharacterPattern, replaceWildcardCharacter);

                // Patterns should not include subfolders like node_modules unless they are
```

### compiler/utilities.ts:9762:49

**New exact finding**. Unit: `getSubPatternFromSpec` at `compiler/utilities.ts:9696:1`; phase `refusal_scan`.

a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type)

Source lines 9760 through 9764; diagnostic column 49:

```typescript
            }
            else {
                subpattern += component.replace(reservedCharacterPattern, replaceWildcardCharacter);
            }
        }
```

### compiler/watch.ts:789:20

**New exact finding**. Unit: `getSourceFileVersionAsHashFromText` at `compiler/watch.ts:787:1`; phase `refusal_scan`.

a view naming �@match@1085, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@match@1085 out of the view's type)

Source lines 787 through 791; diagnostic column 20:

```typescript
export function getSourceFileVersionAsHashFromText(host: Pick<CompilerHost, "createHash">, text: string): string {
    // If text can contain the sourceMapUrl ignore sourceMapUrl for calcualting hash
    if (text.match(sourceMapCommentRegExpDontCareLineStart)) {
        let lineEnd = text.length;
        let lineStart = lineEnd;
```

### compiler/watch.ts:811:28

**New exact finding**. Unit: `getSourceFileVersionAsHashFromText` at `compiler/watch.ts:787:1`; phase `refusal_scan`.

a view naming �@match@1085, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@match@1085 out of the view's type)

Source lines 809 through 813; diagnostic column 28:

```typescript
            // This is start of the line
            const line = text.substring(lineStart, lineEnd);
            if (line.match(sourceMapCommentRegExp)) {
                text = text.substring(0, lineStart);
                break;
```

### compiler/watch.ts:816:34

**New exact finding**. Unit: `getSourceFileVersionAsHashFromText` at `compiler/watch.ts:787:1`; phase `refusal_scan`.

a view naming �@match@1085, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@match@1085 out of the view's type)

Source lines 814 through 818; diagnostic column 34:

```typescript
            }
            // If we see a non-whitespace/map comment-like line, break, to avoid scanning up the entire file
            else if (!line.match(whitespaceOrMapCommentRegExp)) {
                break;
            }
```

### compiler/watchPublic.ts:1039:49

**New exact finding**. Unit: `createWatchProgram` at `compiler/watchPublic.ts:419:1`; phase `refusal_scan`.

a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type)

Source lines 1037 through 1041; diagnostic column 49:

```typescript

        parsedConfigs!.delete(path);
        if (config.watchedDirectories) clearMap(config.watchedDirectories, closeFileWatcherOf);
        config.watcher?.close();
        clearSharedExtendedConfigFileWatcher(path, sharedExtendedConfigFileWatchers);
```

### compiler/watchPublic.ts:564:18

**New exact finding**. Unit: `createWatchProgram` at `compiler/watchPublic.ts:419:1`; phase `refusal_scan`.

a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type)

Source lines 562 through 566; diagnostic column 18:

```typescript
        clearInvalidateResolutionsOfFailedLookupLocations();
        resolutionCache.clear();
        clearMap(sourceFilesCache, value => {
            if (value && value.fileWatcher) {
                value.fileWatcher.close();
```

### compiler/watchPublic.ts:577:22

**New exact finding**. Unit: `createWatchProgram` at `compiler/watchPublic.ts:419:1`; phase `refusal_scan`.

a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type)

Source lines 575 through 579; diagnostic column 22:

```typescript
        extendedConfigCache = undefined;
        if (sharedExtendedConfigFileWatchers) {
            clearMap(sharedExtendedConfigFileWatchers, closeFileWatcherOf);
            sharedExtendedConfigFileWatchers = undefined!;
        }
```

### compiler/watchPublic.ts:581:22

**New exact finding**. Unit: `createWatchProgram` at `compiler/watchPublic.ts:419:1`; phase `refusal_scan`.

a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type)

Source lines 579 through 583; diagnostic column 22:

```typescript
        }
        if (watchedWildcardDirectories) {
            clearMap(watchedWildcardDirectories, closeFileWatcherOf);
            watchedWildcardDirectories = undefined!;
        }
```

### compiler/watchPublic.ts:585:22

**New exact finding**. Unit: `createWatchProgram` at `compiler/watchPublic.ts:419:1`; phase `refusal_scan`.

a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type)

Source lines 583 through 587; diagnostic column 22:

```typescript
        }
        if (missingFilesMap) {
            clearMap(missingFilesMap, closeFileWatcher);
            missingFilesMap = undefined!;
        }
```

### compiler/watchPublic.ts:589:22

**New exact finding**. Unit: `createWatchProgram` at `compiler/watchPublic.ts:419:1`; phase `refusal_scan`.

a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type)

Source lines 587 through 591; diagnostic column 22:

```typescript
        }
        if (parsedConfigs) {
            clearMap(parsedConfigs, config => {
                config.watcher?.close();
                config.watcher = undefined;
```

### compiler/watchPublic.ts:592:57

**New exact finding**. Unit: `createWatchProgram` at `compiler/watchPublic.ts:419:1`; phase `refusal_scan`.

a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type)

Source lines 590 through 594; diagnostic column 57:

```typescript
                config.watcher?.close();
                config.watcher = undefined;
                if (config.watchedDirectories) clearMap(config.watchedDirectories, closeFileWatcherOf);
                config.watchedDirectories = undefined;
            });
```

### compiler/watchUtilities.ts:534:18

**New exact finding**. Unit: `updateWatchingWildcardDirectories` at `compiler/watchUtilities.ts:514:1`; phase `refusal_scan`.

a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type)

Source lines 532 through 536; diagnostic column 18:

```typescript
    }
    else {
        clearMap(existingWatchedForWildcards, closeFileWatcherOf);
    }

```

## Panics and regression details

- before: 0 raw-position panic findings.
- after: 0 raw-position panic findings.
- previously_finding_free_scan_site: NotYet: a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type) at compiler/core.ts:1882:19
- previously_finding_free_scan_site: NotYet: a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type) at compiler/debug.ts:563:69
- previously_finding_free_scan_site: NotYet: a view naming �@split@1091, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@split@1091 out of the view's type) at compiler/semver.ts:278:41
- previously_finding_free_scan_site: NotYet: a view naming �@split@1091, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@split@1091 out of the view's type) at compiler/semver.ts:287:46
- previously_finding_free_scan_site: NotYet: a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type) at compiler/sys.ts:1695:105
- previously_finding_free_scan_site: NotYet: a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type) at compiler/sys.ts:1729:30
- previously_finding_free_scan_site: NotYet: a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type) at compiler/path.ts:528:24
- previously_finding_free_scan_site: NotYet: a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type) at compiler/path.ts:738:35
- previously_finding_free_scan_site: NotYet: a view naming �@split@1091, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@split@1091 out of the view's type) at compiler/utilities.ts:1331:27
- previously_finding_free_scan_site: NotYet: a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type) at compiler/utilities.ts:1331:67
- previously_finding_free_scan_site: NotYet: a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type) at compiler/utilities.ts:2063:48
- previously_finding_free_scan_site: NotYet: a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type) at compiler/utilities.ts:2063:72
- previously_finding_free_scan_site: NotYet: a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type) at compiler/utilities.ts:6139:24
- previously_finding_free_scan_site: NotYet: a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type) at compiler/utilities.ts:6209:22
- previously_finding_free_scan_site: NotYet: a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type) at compiler/utilities.ts:6219:19
- previously_finding_free_scan_site: NotYet: a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type) at compiler/utilities.ts:6250:22
- previously_finding_free_scan_site: NotYet: a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type) at compiler/utilities.ts:8585:25
- previously_finding_free_scan_site: NotYet: a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type) at compiler/utilities.ts:9599:25
- previously_finding_free_scan_site: NotYet: a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type) at compiler/utilities.ts:9747:55
- previously_finding_free_scan_site: NotYet: a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type) at compiler/utilities.ts:9762:49
- previously_finding_free_scan_site: NotYet: a view naming �@match@1085, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@match@1085 out of the view's type) at compiler/utilities.ts:10159:29
- previously_finding_free_scan_site: NotYet: a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type) at compiler/utilities.ts:10917:25
- previously_finding_free_scan_site: NotYet: a view naming �@split@1091, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@split@1091 out of the view's type) at compiler/parser.ts:10784:36
- previously_finding_free_scan_site: NotYet: a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type) at compiler/moduleNameResolver.ts:2763:69
- previously_finding_free_scan_site: NotYet: a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type) at compiler/moduleNameResolver.ts:2807:145
- previously_finding_free_scan_site: NotYet: a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type) at compiler/moduleNameResolver.ts:2809:79
- previously_finding_free_scan_site: NotYet: a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type) at compiler/moduleSpecifiers.ts:152:64
- previously_finding_free_scan_site: NotYet: a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type) at compiler/checker.ts:7878:140
- previously_finding_free_scan_site: NotYet: a view naming �@split@1091, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@split@1091 out of the view's type) at compiler/checker.ts:8063:127
- previously_finding_free_scan_site: NotYet: a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type) at compiler/checker.ts:8063:167
- previously_finding_free_scan_site: NotYet: a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type) at compiler/checker.ts:8944:77
- previously_finding_free_scan_site: NotYet: a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type) at compiler/checker.ts:9920:126
- previously_finding_free_scan_site: NotYet: a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type) at compiler/checker.ts:11134:157
- previously_finding_free_scan_site: NotYet: a view naming has, which lib declares on Map but the runtime's Map doesn't hold (leave has out of the view's type) at compiler/checker.ts:46735:36
- previously_finding_free_scan_site: NotYet: a view naming �@match@1085, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@match@1085 out of the view's type) at compiler/sourcemap.ts:400:30
- previously_finding_free_scan_site: NotYet: a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type) at compiler/transformers/taggedTemplate.ts:130:25
- previously_finding_free_scan_site: NotYet: a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type) at compiler/transformers/ts.ts:2718:30
- previously_finding_free_scan_site: NotYet: a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type) at compiler/transformers/jsx.ts:624:29
- previously_finding_free_scan_site: NotYet: a view naming �@split@1091, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@split@1091 out of the view's type) at compiler/emitter.ts:4105:42
- previously_finding_free_scan_site: NotYet: a view naming �@split@1091, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@split@1091 out of the view's type) at compiler/emitter.ts:4976:34
- previously_finding_free_scan_site: NotYet: a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type) at compiler/watchUtilities.ts:534:18
- previously_finding_free_scan_site: NotYet: a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type) at compiler/program.ts:737:43
- previously_finding_free_scan_site: NotYet: a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type) at compiler/program.ts:751:68
- previously_finding_free_scan_site: NotYet: a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type) at compiler/program.ts:752:82
- previously_finding_free_scan_site: NotYet: a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type) at compiler/program.ts:755:67
- previously_finding_free_scan_site: NotYet: a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type) at compiler/program.ts:759:44
- previously_finding_free_scan_site: NotYet: a view naming has, which lib declares on Set but the runtime's Set doesn't hold (leave has out of the view's type) at compiler/program.ts:4985:13
- previously_finding_free_scan_site: NotYet: a view naming �@replace@1087, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@replace@1087 out of the view's type) at compiler/program.ts:5047:75
- previously_finding_free_scan_site: NotYet: a view naming has, which lib declares on ReadonlySet but the runtime's ReadonlySet doesn't hold (leave has out of the view's type) at compiler/builder.ts:394:41
- previously_finding_free_scan_site: NotYet: a view naming has, which lib declares on ReadonlySet but the runtime's ReadonlySet doesn't hold (leave has out of the view's type) at compiler/builder.ts:1011:41
- previously_finding_free_scan_site: NotYet: a view naming �@match@1085, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@match@1085 out of the view's type) at compiler/resolutionCache.ts:287:50
- previously_finding_free_scan_site: NotYet: a view naming �@match@1085, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@match@1085 out of the view's type) at compiler/resolutionCache.ts:293:49
- previously_finding_free_scan_site: NotYet: a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type) at compiler/resolutionCache.ts:681:18
- previously_finding_free_scan_site: NotYet: a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type) at compiler/resolutionCache.ts:682:18
- previously_finding_free_scan_site: NotYet: a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type) at compiler/resolutionCache.ts:1627:18
- previously_finding_free_scan_site: NotYet: a view naming �@match@1085, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@match@1085 out of the view's type) at compiler/watch.ts:789:20
- previously_finding_free_scan_site: NotYet: a view naming �@match@1085, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@match@1085 out of the view's type) at compiler/watch.ts:811:28
- previously_finding_free_scan_site: NotYet: a view naming �@match@1085, which lib declares on RegExp but the runtime's RegExp doesn't hold (leave �@match@1085 out of the view's type) at compiler/watch.ts:816:34
- previously_finding_free_scan_site: NotYet: a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type) at compiler/watchPublic.ts:564:18
- previously_finding_free_scan_site: NotYet: a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type) at compiler/watchPublic.ts:577:22
- previously_finding_free_scan_site: NotYet: a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type) at compiler/watchPublic.ts:581:22
- previously_finding_free_scan_site: NotYet: a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type) at compiler/watchPublic.ts:585:22
- previously_finding_free_scan_site: NotYet: a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type) at compiler/watchPublic.ts:589:22
- previously_finding_free_scan_site: NotYet: a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type) at compiler/watchPublic.ts:592:57
- previously_finding_free_scan_site: NotYet: a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type) at compiler/watchPublic.ts:1039:49
- previously_finding_free_scan_site: NotYet: a view naming has, which lib declares on Set but the runtime's Set doesn't hold (leave has out of the view's type) at compiler/tsbuildPublic.ts:1774:9
- previously_finding_free_scan_site: NotYet: a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type) at compiler/tsbuildPublic.ts:2247:14
- previously_finding_free_scan_site: NotYet: a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type) at compiler/tsbuildPublic.ts:2248:14
- previously_finding_free_scan_site: NotYet: a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type) at compiler/tsbuildPublic.ts:2249:14
- previously_finding_free_scan_site: NotYet: a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type) at compiler/tsbuildPublic.ts:2249:90
- previously_finding_free_scan_site: NotYet: a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type) at compiler/tsbuildPublic.ts:2250:14
- previously_finding_free_scan_site: NotYet: a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type) at compiler/tsbuildPublic.ts:2250:81
- previously_finding_free_scan_site: NotYet: a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type) at compiler/tsbuildPublic.ts:2251:14
- previously_finding_free_scan_site: NotYet: a view naming forEach, which lib declares on Map but the runtime's Map doesn't hold (leave forEach out of the view's type) at compiler/tsbuildPublic.ts:2251:83

## Time and peak memory

| Run | Child exit | Wall seconds | Peak RSS MiB |
| --- | ---: | ---: | ---: |
| parent | 0 | 2571.931 | 8152.4 |
| head | 0 | 2527.855 | 8027.7 |

Peak RSS is Linux child `getrusage` in KiB converted to MiB. Wall observations include up to one second of polling resolution. Parent and head overlap after a memory-aware stagger ([launch record](head-start.json)); these are shared-host observations, not benchmarks. Every wrapper records its child exit, wall time and peak RSS, and stops a child after 30 minutes without output/progress. Unit timing streams provide progress even while the census buffers file records.

## Evidence and validation

- [Parent stream](parent-run.jsonl), [head stream](head-run.jsonl), [full paired JSON](paired.json), and [paired ledger](paired.md).
- [Newly stopped units](newly-stopped.json), [changed first stops](changed-first-stops.json), and [method-view source quotations](method-view-sites.json).
- [Applied patch set](patch-set.md), [all measured source hashes](source-hashes.json), [compiler/source/binary provenance](provenance.json), and [guarded overlays](overlays/). Go evidence uses `.go.txt` filenames.
- [Parent timings](parent-run-units.jsonl), [head timings](head-run-units.jsonl), run logs and `*.status.json` / `*.time.json` files retain commands, own child exits and resources.
- The independent audit below recounts raw keys, unit first stops and newly stopped units, verifies every input and compiler source hash, and checks the 81-file scope against the measurement header. No measured source was removed or changed. Only unused upstream tests were removed after apply to reserve RAM.

The environment and pinned checker were reused from the preceding task after verifying the dependency pin. The API packages were installed with `npm ci`. The parent apply and both own-pin builds exited 0. A sparse-checkout exclusion initially prevented staging the checkpoint directory; it was corrected before any census run. Logs are retained and never piped into head or tail.

The apply checkpoint uses a separate suffixed ref ([record](checkpoint-apply.json)), allowing this final branch to contain one commit directly on f7390e5b without any force push. No main branch is pushed.
