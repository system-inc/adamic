#!/usr/bin/env node
'use strict';
// Adaptation 99-readonly-inferred-bindings, #r11wppy classes 1 and 4: 12 unannotated local bindings and 4 unannotated
// function returns in src/compiler get an explicit readonly array annotation, so a value that can be tsc's shared
// emptyArray, typed never[], lands in a read-only view of it rather than in the mutable type tsc infers. Adamic ruled
// it; README.md says why, which binding was left out, and how it is proved. Every binding is local and no annotated
// function is exported, so the published declaration files do not change and no api-sanctions.json is needed. Through
// the stock AST the edit inserts `: <annotation>` after each binding's name, or after a function's parameter list,
// refuses before any write unless every slot is unannotated with upstream's text or already carries exactly its
// annotation, and changes nothing on a second run. It sorts after 99-correctness-binding-views, which pins the whole
// text of program.ts's createProgram (through 95-checker-axolotl-h's diagnostic-owners.json), and after
// 99-correctness-axolotl-r, which pins the whole text of checker.ts's getTypeArguments, so both pins read those
// functions as they expect.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
const tree = path.resolve(process.argv[2]);
// Each slot, addressed by name: a binding by its keyword and name in the function that encloses it, or a function's
// return by the function's name and the function or module it is local to. `upstream` is the slot's first line as
// upstream 6.0.3 writes it, trimmed: the statement's or the function's text must start with it, or with it carrying
// the annotation. Any other text is not the slot Adamic ruled on.
const checker = 'src/compiler/checker.ts';
const utilities = 'src/compiler/utilities.ts';
const transformer = 'src/compiler/transformer.ts';
const program = 'src/compiler/program.ts';
const builder = 'src/compiler/builder.ts';
const slots = [
    { file: utilities, function: 'inferPreference', binding: 'specifiers', annotation: 'readonly StringLiteralLike[]',
        upstream: 'const specifiers = sourceFile?.imports.length ? sourceFile.imports :' },
    { file: checker, function: 'serializeModule', binding: 'realMembers', annotation: 'readonly Symbol[]',
        upstream: 'const realMembers = locationMap.get("real") || emptyArray;' },
    { file: checker, function: 'serializeModule', binding: 'mergedMembers', annotation: 'readonly Symbol[]',
        upstream: 'const mergedMembers = locationMap.get("merged") || emptyArray;' },
    { file: checker, function: 'serializeAsNamespaceDeclaration', binding: 'localProps', annotation: 'readonly Symbol[]',
        upstream: 'const localProps = localVsRemoteMap.get("local") || emptyArray;' },
    { file: checker, function: 'serializeAsClass', binding: 'privateProperties', annotation: 'readonly ClassElement[]',
        upstream: 'const privateProperties = hasPrivateIdentifier ?' },
    { file: checker, function: 'resolveBaseTypesOfClass', local: 'createTypeChecker', annotation: 'readonly BaseType[]',
        upstream: 'function resolveBaseTypesOfClass(type: InterfaceType) {' },
    { file: checker, function: 'resolveAnonymousTypeMembers', binding: 'constructSignatures', keyword: 'let',
        annotation: 'readonly Signature[]',
        upstream: 'let constructSignatures = symbol.members ? getSignaturesOfSymbol(symbol.members.get(InternalSymbolName.Constructor)) : emptyArray;' },
    { file: checker, function: 'resolveReverseMappedTypeMembers', binding: 'indexInfos', annotation: 'readonly IndexInfo[]',
        upstream: 'const indexInfos = indexInfo ? [createIndexInfo(stringType, inferReverseMappedType(indexInfo.type, type.mappedType, type.constraintType) || unknownType, readonlyMask && indexInfo.isReadonly)] : emptyArray;' },
    { file: checker, function: 'getTypeArguments', binding: 'typeArguments', annotation: 'readonly Type[] | undefined',
        upstream: 'const typeArguments = !node ? emptyArray :' },
    { file: checker, function: 'getIntendedTypeFromJSDocTypeReference', binding: 'indexInfo', annotation: 'readonly IndexInfo[]',
        upstream: 'const indexInfo = indexed === stringType || indexed === numberType ? [createIndexInfo(indexed, target, /*isReadonly*/ false)] : emptyArray;' },
    { file: checker, function: 'getExactOptionalUnassignableProperties', local: 'createTypeChecker',
        annotation: 'readonly Symbol[]',
        upstream: 'function getExactOptionalUnassignableProperties(source: Type, target: Type) {' },
    { file: checker, function: 'createEmptyObjectTypeFromStringLiteral', binding: 'indexInfos', annotation: 'readonly IndexInfo[]',
        upstream: 'const indexInfos = type.flags & TypeFlags.String ? [createIndexInfo(stringType, emptyObjectType, /*isReadonly*/ false)] : emptyArray;' },
    { file: checker, function: 'addImplementationSuccessElaboration', binding: 'failedSignatureDeclarations',
        annotation: 'readonly Declaration[]',
        upstream: 'const failedSignatureDeclarations = failed.declaration?.symbol?.declarations || emptyArray;' },
    { file: transformer, function: 'getScriptTransformers', local: null,
        annotation: 'readonly TransformerFactory<SourceFile | Bundle>[]',
        upstream: 'function getScriptTransformers(compilerOptions: CompilerOptions, customTransformers?: CustomTransformers, emitOnly?: boolean | EmitOnly) {' },
    { file: program, function: 'getOptionsDiagnosticsOfConfigFile', local: 'createProgram',
        annotation: 'readonly DiagnosticWithLocation[]',
        upstream: 'function getOptionsDiagnosticsOfConfigFile() {' },
    { file: builder, function: 'getSemanticDiagnosticsOfNextAffectedFile', binding: 'result', keyword: 'let',
        annotation: 'readonly Diagnostic[] | undefined', upstream: 'let result;' },
].map(slot => ({ keyword: 'const', ...slot }));
const nodes = (root, predicate) => {
    const found = [];
    ts.forEachChild(root, function visit(node) {
        if (predicate(node)) found.push(node);
        ts.forEachChild(node, visit);
    });
    return found;
};
const one = (found, what) => {
    if (found.length !== 1) throw new Error(`expected one ${what}, found ${found.length}`);
    return found[0];
};
const name = slot => slot.binding ? `${slot.function} ${slot.keyword} ${slot.binding}` : `${slot.function} return`;
// The slot's node, its first line's text, its declared type, and where the annotation goes, found by name: never a
// line or an offset.
function find(source, slot) {
    const owner = one(nodes(source, node => ts.isFunctionDeclaration(node) && node.name?.text === slot.function),
        `function ${slot.function} in ${slot.file}`);
    if (slot.binding) {
        // The binding's enclosing function is the named one, not a callback or a function nested in it.
        const declaration = one(nodes(owner, node => ts.isVariableDeclaration(node) && ts.isIdentifier(node.name)
            && node.name.text === slot.binding && ts.findAncestor(node.parent, ts.isFunctionLike) === owner
            && ts.isVariableDeclarationList(node.parent)
            && (node.parent.flags & ts.NodeFlags.BlockScoped) === (slot.keyword === 'let' ? ts.NodeFlags.Let : ts.NodeFlags.Const)),
            `${slot.keyword} ${slot.binding} in ${slot.function}`);
        const statement = declaration.parent.parent;
        if (!ts.isVariableStatement(statement) || declaration.parent.declarations.length !== 1)
            throw new Error(`expected ${name(slot)} to be the one declaration of its statement`);
        return { node: statement, type: declaration.type, at: declaration.name.end };
    }
    // Local: a function of the named function's, or a module function nothing exports.
    const outer = ts.findAncestor(owner.parent, ts.isFunctionLike);
    if (owner.modifiers?.length || (slot.local ? outer?.name?.text !== slot.local : !ts.isSourceFile(owner.parent)))
        throw new Error(`expected ${slot.function} to be local to ${slot.local ?? `${slot.file}, unexported`}`);
    const close = one(owner.getChildren(source).filter(child => child.kind === ts.SyntaxKind.CloseParenToken),
        `parameter list in ${slot.function}`);
    return { node: owner, type: owner.type, at: close.end };
}
const files = [...new Set(slots.map(slot => slot.file))].map(file => {
    const fileName = path.join(tree, file);
    const text = fs.readFileSync(fileName, 'utf8');
    return { file, fileName, text, source: ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true) };
});
const edits = [];
for (const slot of slots) {
    const { source } = files.find(file => file.file === slot.file);
    const { node, type, at } = find(source, slot);
    const start = node.getStart(source);
    const where = source.getLineAndCharacterOfPosition(start);
    const record = { slot: name(slot), annotation: slot.annotation, where: `${slot.file}:${where.line + 1}:${where.character + 1}` };
    const nodeText = node.getText(source);
    const offset = at - start;
    if (offset > slot.upstream.length) throw new Error(`slot ${record.slot} at ${record.where}: the annotation would fall past its first line`);
    const adapted = `${slot.upstream.slice(0, offset)}: ${slot.annotation}${slot.upstream.slice(offset)}`;
    // Idempotent: a slot that already carries its annotation keeps its bytes, so a second run changes nothing.
    if (type && type.getText(source) === slot.annotation && nodeText.startsWith(adapted)) {
        edits.push({ ...record, edited: false });
        continue;
    }
    if (type || !nodeText.startsWith(slot.upstream))
        throw new Error(`unexpected slot ${record.slot} at ${record.where}: ${nodeText.split('\n')[0]}`);
    edits.push({ ...record, edited: true, file: slot.file, at });
}
// Every slot passed: only now is anything written.
for (const { file, fileName, text } of files) {
    let adapted = text;
    for (const edit of edits.filter(edit => edit.edited && edit.file === file).sort((a, b) => b.at - a.at))
        adapted = adapted.slice(0, edit.at) + `: ${edit.annotation}` + adapted.slice(edit.at);
    if (adapted === text) continue;
    if (fs.readFileSync(fileName, 'utf8') !== text) throw new Error(`source changed during adaptation: ${fileName}`);
    fs.writeFileSync(fileName, adapted);
}
console.log(JSON.stringify({ typescript: ts.version, task: '#r11wppy', classes: [1, 4], change: 'internal',
    slots: edits.map(({ file, at, ...edit }) => edit) }, null, 2));
