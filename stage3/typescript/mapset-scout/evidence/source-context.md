# Whole functions containing the 35 first-error sites

Adapted-tree coordinates at 1d3efe7b; function bodies retain their source text with line endings and trailing whitespace normalized (exact extracted text is in sites.json).

## 1. utilitiesPublic.ts:625:22
Enclosing function: `utilitiesPublic.ts:623`.

```typescript
export function isEmptyBindingPattern(node: BindingName): node is BindingPattern {
    if (isBindingPattern(node)) {
        return every(node.elements, isEmptyBindingElement);
    }
    return false;
}
```

## 2. utilitiesPublic.ts:1035:30
Enclosing function: `utilitiesPublic.ts:1034`.

```typescript
export function getModifiers(node: HasModifiers): readonly Modifier[] | undefined {
    if (hasSyntacticModifier(node, ModifierFlags.Modifier)) {
        return filter(node.modifiers, isModifier);
    }
}
```

## 3. utilitiesPublic.ts:1082:31
Enclosing function: `utilitiesPublic.ts:1080`.

```typescript
function getJSDocTypeParameterTagsWorker(param: TypeParameterDeclaration, noCache?: boolean): readonly JSDocTemplateTag[] {
    const name = param.name.escapedText;
    return getJSDocTagsWorker(param.parent, noCache).filter((tag): tag is JSDocTemplateTag => isJSDocTemplateTag(tag) && tag.typeParameters.some(tp => tp.name.escapedText === name));
}
```

## 4. utilitiesPublic.ts:1111:37
Enclosing function: `utilitiesPublic.ts:1110`.

```typescript
export function hasJSDocParameterTags(node: FunctionLikeDeclaration | SignatureDeclaration): boolean {
    return !!getFirstJSDocTag(node, isJSDocParameterTag);
}
```

Value declaration: `src/compiler/factory/nodeTests.ts:1148`.

```typescript
export function isJSDocParameterTag(node: Node): node is JSDocParameterTag {
    return node.kind === SyntaxKind.JSDocParameterTag;
}
```

## 5. utilitiesPublic.ts:1116:35
Enclosing function: `utilitiesPublic.ts:1115`.

```typescript
export function getJSDocAugmentsTag(node: Node): JSDocAugmentsTag | undefined {
    return getFirstJSDocTag(node, isJSDocAugmentsTag);
}
```

Value declaration: `src/compiler/factory/nodeTests.ts:1096`.

```typescript
export function isJSDocAugmentsTag(node: Node): node is JSDocAugmentsTag {
    return node.kind === SyntaxKind.JSDocAugmentsTag;
}
```

## 6. utilitiesPublic.ts:1121:34
Enclosing function: `utilitiesPublic.ts:1120`.

```typescript
export function getJSDocImplementsTags(node: Node): readonly JSDocImplementsTag[] {
    return getAllJSDocTags(node, isJSDocImplementsTag);
}
```

Value declaration: `src/compiler/factory/nodeTests.ts:1180`.

```typescript
export function isJSDocImplementsTag(node: Node): node is JSDocImplementsTag {
    return node.kind === SyntaxKind.JSDocImplementsTag;
}
```

## 7. utilitiesPublic.ts:1126:35
Enclosing function: `utilitiesPublic.ts:1125`.

```typescript
export function getJSDocClassTag(node: Node): JSDocClassTag | undefined {
    return getFirstJSDocTag(node, isJSDocClassTag);
}
```

Value declaration: `src/compiler/factory/nodeTests.ts:1104`.

```typescript
export function isJSDocClassTag(node: Node): node is JSDocClassTag {
    return node.kind === SyntaxKind.JSDocClassTag;
}
```

## 8. utilitiesPublic.ts:1131:35
Enclosing function: `utilitiesPublic.ts:1130`.

```typescript
export function getJSDocPublicTag(node: Node): JSDocPublicTag | undefined {
    return getFirstJSDocTag(node, isJSDocPublicTag);
}
```

Value declaration: `src/compiler/factory/nodeTests.ts:1112`.

```typescript
export function isJSDocPublicTag(node: Node): node is JSDocPublicTag {
    return node.kind === SyntaxKind.JSDocPublicTag;
}
```

## 9. utilitiesPublic.ts:1136:35
Enclosing function: `utilitiesPublic.ts:1135`.

```typescript
export function getJSDocPublicTagNoCache(node: Node): JSDocPublicTag | undefined {
    return getFirstJSDocTag(node, isJSDocPublicTag, /*noCache*/ true);
}
```

Value declaration: `src/compiler/factory/nodeTests.ts:1112`.

```typescript
export function isJSDocPublicTag(node: Node): node is JSDocPublicTag {
    return node.kind === SyntaxKind.JSDocPublicTag;
}
```

## 10. utilitiesPublic.ts:1141:35
Enclosing function: `utilitiesPublic.ts:1140`.

```typescript
export function getJSDocPrivateTag(node: Node): JSDocPrivateTag | undefined {
    return getFirstJSDocTag(node, isJSDocPrivateTag);
}
```

Value declaration: `src/compiler/factory/nodeTests.ts:1116`.

```typescript
export function isJSDocPrivateTag(node: Node): node is JSDocPrivateTag {
    return node.kind === SyntaxKind.JSDocPrivateTag;
}
```

## 11. utilitiesPublic.ts:1146:35
Enclosing function: `utilitiesPublic.ts:1145`.

```typescript
export function getJSDocPrivateTagNoCache(node: Node): JSDocPrivateTag | undefined {
    return getFirstJSDocTag(node, isJSDocPrivateTag, /*noCache*/ true);
}
```

Value declaration: `src/compiler/factory/nodeTests.ts:1116`.

```typescript
export function isJSDocPrivateTag(node: Node): node is JSDocPrivateTag {
    return node.kind === SyntaxKind.JSDocPrivateTag;
}
```

## 12. utilitiesPublic.ts:1151:35
Enclosing function: `utilitiesPublic.ts:1150`.

```typescript
export function getJSDocProtectedTag(node: Node): JSDocProtectedTag | undefined {
    return getFirstJSDocTag(node, isJSDocProtectedTag);
}
```

Value declaration: `src/compiler/factory/nodeTests.ts:1120`.

```typescript
export function isJSDocProtectedTag(node: Node): node is JSDocProtectedTag {
    return node.kind === SyntaxKind.JSDocProtectedTag;
}
```

## 13. utilitiesPublic.ts:1156:35
Enclosing function: `utilitiesPublic.ts:1155`.

```typescript
export function getJSDocProtectedTagNoCache(node: Node): JSDocProtectedTag | undefined {
    return getFirstJSDocTag(node, isJSDocProtectedTag, /*noCache*/ true);
}
```

Value declaration: `src/compiler/factory/nodeTests.ts:1120`.

```typescript
export function isJSDocProtectedTag(node: Node): node is JSDocProtectedTag {
    return node.kind === SyntaxKind.JSDocProtectedTag;
}
```

## 14. utilitiesPublic.ts:1161:35
Enclosing function: `utilitiesPublic.ts:1160`.

```typescript
export function getJSDocReadonlyTag(node: Node): JSDocReadonlyTag | undefined {
    return getFirstJSDocTag(node, isJSDocReadonlyTag);
}
```

Value declaration: `src/compiler/factory/nodeTests.ts:1124`.

```typescript
export function isJSDocReadonlyTag(node: Node): node is JSDocReadonlyTag {
    return node.kind === SyntaxKind.JSDocReadonlyTag;
}
```

## 15. utilitiesPublic.ts:1166:35
Enclosing function: `utilitiesPublic.ts:1165`.

```typescript
export function getJSDocReadonlyTagNoCache(node: Node): JSDocReadonlyTag | undefined {
    return getFirstJSDocTag(node, isJSDocReadonlyTag, /*noCache*/ true);
}
```

Value declaration: `src/compiler/factory/nodeTests.ts:1124`.

```typescript
export function isJSDocReadonlyTag(node: Node): node is JSDocReadonlyTag {
    return node.kind === SyntaxKind.JSDocReadonlyTag;
}
```

## 16. utilitiesPublic.ts:1170:35
Enclosing function: `utilitiesPublic.ts:1169`.

```typescript
export function getJSDocOverrideTagNoCache(node: Node): JSDocOverrideTag | undefined {
    return getFirstJSDocTag(node, isJSDocOverrideTag, /*noCache*/ true);
}
```

Value declaration: `src/compiler/factory/nodeTests.ts:1128`.

```typescript
export function isJSDocOverrideTag(node: Node): node is JSDocOverrideTag {
    return node.kind === SyntaxKind.JSDocOverrideTag;
}
```

## 17. utilitiesPublic.ts:1175:35
Enclosing function: `utilitiesPublic.ts:1174`.

```typescript
export function getJSDocDeprecatedTag(node: Node): JSDocDeprecatedTag | undefined {
    return getFirstJSDocTag(node, isJSDocDeprecatedTag);
}
```

Value declaration: `src/compiler/factory/nodeTests.ts:1136`.

```typescript
export function isJSDocDeprecatedTag(node: Node): node is JSDocDeprecatedTag {
    return node.kind === SyntaxKind.JSDocDeprecatedTag;
}
```

## 18. utilitiesPublic.ts:1180:35
Enclosing function: `utilitiesPublic.ts:1179`.

```typescript
export function getJSDocDeprecatedTagNoCache(node: Node): JSDocDeprecatedTag | undefined {
    return getFirstJSDocTag(node, isJSDocDeprecatedTag, /*noCache*/ true);
}
```

Value declaration: `src/compiler/factory/nodeTests.ts:1136`.

```typescript
export function isJSDocDeprecatedTag(node: Node): node is JSDocDeprecatedTag {
    return node.kind === SyntaxKind.JSDocDeprecatedTag;
}
```

## 19. utilitiesPublic.ts:1190:35
Enclosing function: `utilitiesPublic.ts:1189`.

```typescript
export function getJSDocThisTag(node: Node): JSDocThisTag | undefined {
    return getFirstJSDocTag(node, isJSDocThisTag);
}
```

Value declaration: `src/compiler/factory/nodeTests.ts:1156`.

```typescript
export function isJSDocThisTag(node: Node): node is JSDocThisTag {
    return node.kind === SyntaxKind.JSDocThisTag;
}
```

## 20. utilitiesPublic.ts:1195:35
Enclosing function: `utilitiesPublic.ts:1194`.

```typescript
export function getJSDocReturnTag(node: Node): JSDocReturnTag | undefined {
    return getFirstJSDocTag(node, isJSDocReturnTag);
}
```

Value declaration: `src/compiler/factory/nodeTests.ts:1152`.

```typescript
export function isJSDocReturnTag(node: Node): node is JSDocReturnTag {
    return node.kind === SyntaxKind.JSDocReturnTag;
}
```

## 21. utilitiesPublic.ts:1200:35
Enclosing function: `utilitiesPublic.ts:1199`.

```typescript
export function getJSDocTemplateTag(node: Node): JSDocTemplateTag | undefined {
    return getFirstJSDocTag(node, isJSDocTemplateTag);
}
```

Value declaration: `src/compiler/factory/nodeTests.ts:1164`.

```typescript
export function isJSDocTemplateTag(node: Node): node is JSDocTemplateTag {
    return node.kind === SyntaxKind.JSDocTemplateTag;
}
```

## 22. utilitiesPublic.ts:1204:35
Enclosing function: `utilitiesPublic.ts:1203`.

```typescript
export function getJSDocSatisfiesTag(node: Node): JSDocSatisfiesTag | undefined {
    return getFirstJSDocTag(node, isJSDocSatisfiesTag);
}
```

Value declaration: `src/compiler/factory/nodeTests.ts:1184`.

```typescript
export function isJSDocSatisfiesTag(node: Node): node is JSDocSatisfiesTag {
    return node.kind === SyntaxKind.JSDocSatisfiesTag;
}
```

## 23. utilitiesPublic.ts:1210:40
Enclosing function: `utilitiesPublic.ts:1208`.

```typescript
export function getJSDocTypeTag(node: Node): JSDocTypeTag | undefined {
    // We should have already issued an error if there were multiple type jsdocs, so just use the first one.
    const tag = getFirstJSDocTag(node, isJSDocTypeTag);
    if (tag && tag.typeExpression && tag.typeExpression.type) {
        return tag;
    }
    return undefined;
}
```

Value declaration: `src/compiler/factory/nodeTests.ts:1160`.

```typescript
export function isJSDocTypeTag(node: Node): node is JSDocTypeTag {
    return node.kind === SyntaxKind.JSDocTypeTag;
}
```

## 24. utilitiesPublic.ts:1229:84
Enclosing function: `utilitiesPublic.ts:1228`.

```typescript
export function getJSDocType(node: Node): TypeNode | undefined {
    let tag: JSDocTypeTag | JSDocParameterTag | undefined = getFirstJSDocTag(node, isJSDocTypeTag);
    if (!tag && isParameter(node)) {
        tag = find(getJSDocParameterTags(node), tag => !!tag.typeExpression);
    }

    return tag && tag.typeExpression && tag.typeExpression.type;
}
```

Value declaration: `src/compiler/factory/nodeTests.ts:1160`.

```typescript
export function isJSDocTypeTag(node: Node): node is JSDocTypeTag {
    return node.kind === SyntaxKind.JSDocTypeTag;
}
```

## 25. utilitiesPublic.ts:1357:28
Enclosing function: `utilitiesPublic.ts:1355`.

```typescript
export function getEffectiveConstraintOfTypeParameter(node: TypeParameterDeclaration): TypeNode | undefined {
    return node.constraint ? node.constraint :
        isJSDocTemplateTag(node.parent) && node === node.parent.typeParameters[0] ? node.parent.constraint :
        undefined;
}
```

## 26. utilitiesPublic.ts:2638:76
Enclosing function: `utilitiesPublic.ts:2637`.

```typescript
export function hasRestParameter(s: SignatureDeclaration | JSDocSignature): boolean {
    const last = lastOrUndefined<ParameterDeclaration | JSDocParameterTag>(s.parameters);
    return !!last && isRestParameter(last);
}
```

## 27. utilities.ts:3095:20
Enclosing function: `utilities.ts:3094`.

```typescript
export function forEachPropertyAssignment<T>(objectLiteral: ObjectLiteralExpression | undefined, key: string, callback: (property: PropertyAssignment) => T | undefined, key2?: string): T | undefined {
    return forEach(objectLiteral?.properties, property => {
        if (!isPropertyAssignment(property)) return undefined;
        const propName = tryGetTextOfPropertyName(property.name);
        return key === propName || (key2 && key2 === propName) ?
            callback(property) :
            undefined;
    });
}
```

## 28. utilities.ts:4429:82
Enclosing function: `utilities.ts:4426`.

```typescript
export function getNamespaceDeclarationNode(node: ImportDeclaration | ImportEqualsDeclaration | ExportDeclaration): ImportEqualsDeclaration | NamespaceImport | NamespaceExport | undefined {
    switch (node.kind) {
        case SyntaxKind.ImportDeclaration:
            return node.importClause && tryCast(node.importClause.namedBindings, isNamespaceImport);
        case SyntaxKind.ImportEqualsDeclaration:
            return node;
        case SyntaxKind.ExportDeclaration:
            return node.exportClause && tryCast(node.exportClause, isNamespaceExport);
        default:
            return Debug.assertNever(node);
    }
}
```

Value declaration: `src/compiler/factory/nodeTests.ts:870`.

```typescript
export function isNamespaceImport(node: Node): node is NamespaceImport {
    return node.kind === SyntaxKind.NamespaceImport;
}
```

## 29. utilities.ts:4780:34
Enclosing function: `utilities.ts:4779`.

```typescript
export function getJSDocOverloadTags(node: Node): readonly JSDocOverloadTag[] {
    return getAllJSDocTags(node, isJSDocOverloadTag);
}
```

Value declaration: `src/compiler/factory/nodeTests.ts:1132`.

```typescript
export function isJSDocOverloadTag(node: Node): node is JSDocOverloadTag {
    return node.kind === SyntaxKind.JSDocOverloadTag;
}
```

## 30. factory/utilities.ts:1614:9
Enclosing function: `factory/utilities.ts:1611`.

```typescript
export function createAccessorPropertyBackingField(factory: NodeFactory, node: PropertyDeclaration, modifiers: ModifiersArray | undefined, initializer: Expression | undefined): PropertyDeclaration {
    return factory.updatePropertyDeclaration(
        node,
        modifiers,
        factory.getGeneratedPrivateNameForNode(node.name, /*prefix*/ undefined, "_accessor_storage"),
        /*questionOrExclamationToken*/ undefined,
        /*type*/ undefined,
        initializer,
    );
}
```

## 31. parser.ts:472:12
Enclosing function: `parser.ts:469`.

```typescript
export function isFileProbablyExternalModule(sourceFile: SourceFile): Node | undefined {
    // Try to use the first top-level import/export when available, then
    // fall back to looking for an 'import.meta' somewhere in the tree if necessary.
    return forEach(sourceFile.statements, isAnExternalModuleIndicatorNode) ||
        getImportMetaIfNecessary(sourceFile);
}
```

## 32. parser.ts:496:17
Enclosing function: `parser.ts:495`.

```typescript
function hasModifierOfKind(node: HasModifiers, kind: SyntaxKind) {
    return some(node.modifiers, m => m.kind === kind);
}
```

## 33. transformers/utilities.ts:126:25
Enclosing function: `transformers/utilities.ts:124`.

```typescript
function containsDefaultReference(node: NamedImportBindings | NamedExportBindings | undefined) {
    if (!node) return false;
    if (!isNamedImports(node) && !isNamedExports(node)) return false;
    return some(node.elements, isNamedDefaultReference);
}
```

## 34. transformers/utilities.ts:161:25
Enclosing function: `transformers/utilities.ts:153`.

```typescript
export function getImportNeedsImportStarHelper(node: ImportDeclaration): boolean {
    if (!!getNamespaceDeclarationNode(node)) {
        return true;
    }
    const bindings = node.importClause && node.importClause.namedBindings;
    if (!bindings) {
        return false;
    }
    if (!isNamedImports(bindings)) return false;
    let defaultRefCount = 0;
    for (const binding of bindings.elements) {
        if (isNamedDefaultReference(binding)) {
            defaultRefCount++;
        }
    }
    // Import star is required if there's default named refs mixed with non-default refs, or if theres non-default refs and it has a default import
    return (defaultRefCount > 0 && defaultRefCount !== bindings.elements.length) || (!!(bindings.elements.length - defaultRefCount) && isDefaultImport(node));
}
```

## 35. transformers/utilities.ts:175:118
Enclosing function: `transformers/utilities.ts:173`.

```typescript
export function getImportNeedsImportDefaultHelper(node: ImportDeclaration): boolean {
    // Import default is needed if there's a default import or a default ref and no other refs (meaning an import star helper wasn't requested)
    return !getImportNeedsImportStarHelper(node) && (isDefaultImport(node) || (!!node.importClause && isNamedImports(node.importClause.namedBindings!) && containsDefaultReference(node.importClause.namedBindings))); // TODO: GH#18217
}
```
