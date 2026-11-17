# The 17 mapped compiler/services calls

Counts are static call sites, never the number of repeated observations. A missing storage refusal is not proof of call success: a generic body may be deferred or another stop may precede it. Entry closure contains 11 of these calls; the other six are service calls measured under their own full-file entry roots.

Observed storage-refusal sites: **4 of 17**. Required coverage complete: False.

## arrayFrom — src/compiler/core.ts:1454:12

```ts
arrayFrom(arrayToMultiMap(values, getGroupId).values(), resultSelector)
```

No certified-storage refusal observed in the completed required streams. This does not establish that the call passed.

## arrayFrom — src/compiler/moduleSpecifiers.ts:868:32

```ts
arrayFrom<[string, { path: string; isRedirect: boolean; isInNodeModules: boolean; }], ModulePath>(
            allFileNames.entries(),
            ([fileName, { isRedirect, isInNodeModules }]): ModulePath => ({ path: fileName, isRedirect, isInNodeModules }),
        )
```

No certified-storage refusal observed in the completed required streams. This does not establish that the call passed.

## arrayFrom — src/compiler/sourcemap.ts:743:30

```ts
arrayFrom(decoder, processMapping)
```

No certified-storage refusal observed in the completed required streams. This does not establish that the call passed.

## arrayFrom — src/compiler/transformers/jsx.ts:186:68

```ts
arrayFrom(importSpecifiersMap.values(), s => factory.createBindingElement(/*dotDotDotToken*/ undefined, s.propertyName, s.name))
```

No certified-storage refusal observed in the completed required streams. This does not establish that the call passed.

## arrayToMap — src/compiler/builder.ts:2416:31

```ts
arrayToMap(diagnostics, value => toFilePath(value[0]), value => value[1])
```

```text
a certified overload whose implementation stores Map<Path, V2 | IncrementalBuildInfoDiagnosticOfFile> where the call reads Map<Path, readonly ReusableDiagnostic[]>
```

Observed in: entry src/tsc/tsc.ts, full src/compiler/builder.ts

## arrayToMap — src/compiler/builder.ts:2344:77

```ts
arrayToMap(buildInfo.affectedFilesPendingEmit, value => toFilePath(isNumber(value) ? value : value[0]), value => toBuilderFileEmit(value, fullEmitForOptions!))
```

No certified-storage refusal observed in the completed required streams. This does not establish that the call passed.

## arrayFrom — src/compiler/builder.ts:1421:25

```ts
arrayFrom(set.keys(), toFileId)
```

```text
a certified overload whose implementation stores (IncrementalBuildInfoFileId | Path)[] where the call reads IncrementalBuildInfoFileId[]
```

Observed in: entry src/tsc/tsc.ts, full src/compiler/builder.ts

## arrayFrom — src/compiler/builder.ts:1315:23

```ts
arrayFrom(state.fileInfos.entries(), ([key, value]): IncrementalMultiFileEmitBuildInfoFileInfo => {
        // Ensure fileId
        const fileId = toFileId(key);
        tryAddRoot(key, fileId);
        Debug.assert(fileNames[fileId - 1] === relativeToBuildInfo(key));
        const oldSignature = state.oldSignatures?.get(key);
        const actualSignature = oldSignature !== undefined ? oldSignature || undefined : value.signature;
        if (state.compilerOptions.composite) {
            const file = state.program.getSourceFileByPath(key)!;
            if (!isJsonSourceFile(file) && sourceFileMayBeEmitted(file, state.program)) {
                const emitSignature = state.emitSignatures?.get(key);
                if (emitSignature !== actualSignature) {
                    emitSignatures = append(
                        emitSignatures,
                        emitSignature === undefined ?
                            fileId : // There is no emit, encode as false
                            // fileId, signature: emptyArray if signature only differs in dtsMap option than our own compilerOptions otherwise EmitSignature
                            [fileId, !isString(emitSignature) && emitSignature[0] === actualSignature ? emptyArray as [] : emitSignature],
                    );
                }
            }
        }
        return value.version === actualSignature ?
            value.affectsGlobalScope || value.impliedFormat ?
                // If file version is same as signature, dont serialize signature
                { version: value.version, signature: undefined, affectsGlobalScope: value.affectsGlobalScope, impliedFormat: value.impliedFormat } :
                // If file info only contains version and signature and both are same we can just write string
                value.version :
            actualSignature !== undefined ? // If signature is not same as version, encode signature in the fileInfo
            oldSignature === undefined ?
                // If we havent computed signature, use fileInfo as is
                value :
                // Serialize fileInfo with new updated signature
                { version: value.version, signature: actualSignature, affectsGlobalScope: value.affectsGlobalScope, impliedFormat: value.impliedFormat } :
            // Signature of the FileInfo is undefined, serialize it as false
            { version: value.version, signature: false, affectsGlobalScope: value.affectsGlobalScope, impliedFormat: value.impliedFormat };
    })
```

```text
a certified overload whose implementation stores ([Path, FileInfo] | IncrementalMultiFileEmitBuildInfoFileInfo)[] where the call reads IncrementalMultiFileEmitBuildInfoFileInfo[]
```

Observed in: full src/compiler/builder.ts

## arrayFrom — src/compiler/builder.ts:1281:27

```ts
arrayFrom(state.fileInfos.entries(), ([key, value]): IncrementalBundleEmitBuildInfoFileInfo => {
            // Ensure fileId
            const fileId = toFileId(key);
            tryAddRoot(key, fileId);
            return value.impliedFormat ?
                { version: value.version, impliedFormat: value.impliedFormat, signature: undefined, affectsGlobalScope: undefined } :
                value.version;
        })
```

No certified-storage refusal observed in the completed required streams. This does not establish that the call passed.

## arrayFrom — src/compiler/builder.ts:1269:19

```ts
arrayFrom(rootFileNames, r => relativeToBuildInfo(r))
```

No certified-storage refusal observed in the completed required streams. This does not establish that the call passed.

## arrayFrom — src/compiler/tsbuildPublic.ts:1060:40

```ts
arrayFrom(
                    internalMap.values(),
                    data =>
                        state.host.realpath && (isPackageJsonInfo(data) || data.directoryExists) ?
                            state.host.realpath(combinePaths(data.packageDirectory, "package.json")) :
                            combinePaths(data.packageDirectory, "package.json"),
                )
```

No certified-storage refusal observed in the completed required streams. This does not establish that the call passed.

## arrayFrom — src/services/refactors/convertImport.ts:276:52

```ts
arrayFrom(neededNamedImports.values(), element => factory.createImportSpecifier(element.isTypeOnly, element.propertyName && factory.cloneNode(element.propertyName), factory.cloneNode(element.name)))
```

No certified-storage refusal observed in the completed required streams. This does not establish that the call passed.

## arrayFrom — src/services/refactors/extractSymbol.ts:1083:43

```ts
arrayFrom(typeParameterUsages.values(), type => ({ type, declaration: getFirstDeclarationBeforePosition(type, context.startPosition) }))
```

No certified-storage refusal observed in the completed required streams. This does not establish that the call passed.

## arrayFrom — src/services/codefixes/importFixes.ts:1582:16

```ts
arrayFrom(
            flatMapIterator(exportInfo.values(), exportInfos => getImportFixes(exportInfos, symbolToken.getStart(sourceFile), isValidTypeOnlyUseSite, useRequire, program, sourceFile, host, preferences).fixes),
            fix => ({ fix, symbolName, errorIdentifierText: symbolToken.text, isJsxNamespaceFix: symbolName !== symbolToken.text }),
        )
```

No certified-storage refusal observed in the completed required streams. This does not establish that the call passed.

## arrayFrom — src/services/codefixes/importFixes.ts:687:33

```ts
arrayFrom(namedImports.entries(), ([name, [addAsTypeOnly, propertyName]]) => ({ addAsTypeOnly, propertyName, name }))
```

No certified-storage refusal observed in the completed required streams. This does not establish that the call passed.

## arrayFrom — src/services/codefixes/importFixes.ts:673:17

```ts
arrayFrom(namedImports.entries(), ([name, { addAsTypeOnly, propertyName }]) => ({ addAsTypeOnly, propertyName, name }))
```

No certified-storage refusal observed in the completed required streams. This does not establish that the call passed.

## arrayFrom — src/services/codefixes/helpers.ts:586:12

```ts
arrayFrom(
        usedNames.values(),
        usedName => factory.createTypeParameterDeclaration(/*modifiers*/ undefined, usedName, constraintsByName.get(usedName)?.constraint),
    )
```

```text
a certified overload whose implementation stores (string | TypeParameterDeclaration)[] where the call reads TypeParameterDeclaration[]
```

Observed in: full src/services/codefixes/helpers.ts

