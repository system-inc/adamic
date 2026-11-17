#!/usr/bin/env node
'use strict';
// Adaptation 95-readonly-empty-array-slots, #r11wppy item 1, the internal change: 20 local or @internal array slots in
// src/compiler/checker.ts and src/compiler/types.ts become readonly, so a slot that can hold tsc's shared emptyArray,
// typed never[], is a read-only view of it and Adamic's never[]-seen-as-T[] refusal clears at its site. Adamic ruled it
// as proposed; README.md says why, which slots were left out, and how it is proved. No slot is public, so the published
// declaration files do not change and no api-sanctions.json is needed. Through the stock AST the edit inserts
// `readonly ` before each slot's array type, refuses before any write unless every slot has upstream's type or is
// already adapted, and changes nothing on a second run. It sorts after 95-checker-axolotl-c, which pins the whole text
// of SymbolLinks, so that pin reads SymbolLinks.variances as upstream writes it.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
const tree = path.resolve(process.argv[2]);
// Each slot, addressed by name: a checker.ts function's return type or parameter, or a types.ts interface's method
// return type, method parameter or property. The type as upstream 6.0.3 writes it, and as this adaptation leaves it:
// any other type is not the slot Adamic ruled on.
const checker = 'src/compiler/checker.ts';
const types = 'src/compiler/types.ts';
const slots = [
    { file: checker, function: 'getNamedOrIndexSignatureMembers', upstream: 'Symbol[]' },
    { file: checker, function: 'getImplementsTypes', upstream: 'BaseType[]' },
    { file: checker, function: 'createInstantiatedSymbolTable', parameter: 'symbols', upstream: 'Symbol[]' },
    { file: checker, function: 'addInheritedMembers', parameter: 'baseSymbols', upstream: 'Symbol[]' },
    { file: checker, function: 'getUnionSignatures', upstream: 'Signature[]' },
    { file: checker, function: 'getUnionIndexInfos', upstream: 'IndexInfo[]' },
    { file: checker, function: 'relateVariances', parameter: 'variances', upstream: 'VarianceFlags[]' },
    { file: checker, function: 'excludeProperties', parameter: 'properties', upstream: 'Symbol[]' },
    { file: checker, function: 'getVariances', upstream: 'VarianceFlags[]' },
    { file: checker, function: 'getVariancesWorker', upstream: 'VarianceFlags[]' },
    { file: checker, function: 'hasCovariantVoidArgument', parameter: 'variances', upstream: 'VarianceFlags[]' },
    { file: checker, function: 'findDiscriminantProperties', parameter: 'sourceProperties', upstream: 'Symbol[]' },
    { file: checker, function: 'getSpellingSuggestionForName', parameter: 'symbols', upstream: 'Symbol[]' },
    { file: checker, function: 'getNonInheritedProperties', parameter: 'baseTypes', upstream: 'BaseType[]' },
    { file: checker, function: 'getNonInheritedProperties', parameter: 'properties', upstream: 'Symbol[]' },
    { file: checker, function: 'getPropertiesOfContainerFunction', upstream: 'Symbol[]' },
    { file: types, interface: 'TypeChecker', method: 'createAnonymousType', parameter: 'constructSignatures',
        upstream: 'Signature[]' },
    { file: types, interface: 'TypeChecker', method: 'createAnonymousType', parameter: 'indexInfos',
        upstream: 'IndexInfo[]' },
    { file: types, interface: 'EmitResolver', method: 'getPropertiesOfContainerFunction', upstream: 'Symbol[]' },
    { file: types, interface: 'SymbolLinks', property: 'variances', upstream: 'VarianceFlags[]' },
].map(slot => ({ ...slot, adapted: `readonly ${slot.upstream}` }));
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
const internal = node => ts.getJSDocTags(node).some(tag => tag.tagName.text === 'internal');
const name = slot => slot.function
    ? `${slot.function} ${slot.parameter ?? 'return'}`
    : `${slot.interface}.${slot.method ?? slot.property}${slot.method ? ` ${slot.parameter ?? 'return'}` : ''}`;
// The slot's declared type node, found by name: never a line or an offset.
function find(source, slot) {
    let owner;
    if (slot.function) {
        owner = one(nodes(source, node => ts.isFunctionDeclaration(node) && node.name?.text === slot.function),
            `function ${slot.function} in ${slot.file}`);
        // Local: a function of createTypeChecker's, exported by nothing.
        let outer = owner.parent;
        while (outer && !(ts.isFunctionDeclaration(outer) && outer.name?.text === 'createTypeChecker')) outer = outer.parent;
        if (!outer || owner.modifiers?.length) throw new Error(`expected ${slot.function} to be local to createTypeChecker`);
    } else {
        const container = one(nodes(source, node => ts.isInterfaceDeclaration(node) && node.name.text === slot.interface),
            `interface ${slot.interface} in ${slot.file}`);
        // A call or index signature has no name, so the name is checked before it is read.
        owner = one(container.members.filter(member => member.name && ts.isIdentifier(member.name)
            && member.name.text === (slot.method ?? slot.property)
            && (slot.method ? ts.isMethodSignature(member) : ts.isPropertySignature(member))),
            `${slot.method ? 'method' : 'property'} ${name(slot)} in ${slot.file}`);
        if (!internal(owner) && !internal(container))
            throw new Error(`expected ${name(slot)} or its interface to carry @internal`);
    }
    const declaration = slot.parameter
        ? one(owner.parameters.filter(parameter => ts.isIdentifier(parameter.name) && parameter.name.text === slot.parameter),
            `parameter ${name(slot)}`)
        : owner;
    if (!declaration.type) throw new Error(`expected ${name(slot)} to declare its type: ${owner.getText(source)}`);
    return declaration.type;
}
const files = [...new Set(slots.map(slot => slot.file))].map(file => {
    const fileName = path.join(tree, file);
    const text = fs.readFileSync(fileName, 'utf8');
    return { file, fileName, text, source: ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true) };
});
const edits = [];
for (const slot of slots) {
    const { source } = files.find(file => file.file === slot.file);
    const type = find(source, slot);
    const where = source.getLineAndCharacterOfPosition(type.getStart(source));
    const record = { slot: name(slot), where: `${slot.file}:${where.line + 1}:${where.character + 1}` };
    const typeText = type.getText(source);
    // Idempotent: a slot already readonly keeps its bytes, so a second run changes nothing.
    if (typeText === slot.adapted) {
        edits.push({ ...record, edited: false });
        continue;
    }
    if (typeText !== slot.upstream) throw new Error(`unexpected slot ${record.slot} at ${record.where}: ${typeText}`);
    // The array takes the readonly: the type itself, or a union's one array member (`X[] | undefined`).
    const arrays = ts.isUnionTypeNode(type) ? type.types.filter(ts.isArrayTypeNode) : [type].filter(ts.isArrayTypeNode);
    const array = one(arrays, `array type in slot ${record.slot}`);
    const offset = array.getStart(source) - type.getStart(source);
    if (typeText.slice(0, offset) + 'readonly ' + typeText.slice(offset) !== slot.adapted)
        throw new Error(`slot ${record.slot} would not read ${slot.adapted}`);
    edits.push({ ...record, edited: true, file: slot.file, at: array.getStart(source) });
}
// Every slot passed: only now is anything written.
for (const { file, fileName, text } of files) {
    let adapted = text;
    for (const edit of edits.filter(edit => edit.edited && edit.file === file).sort((a, b) => b.at - a.at))
        adapted = adapted.slice(0, edit.at) + 'readonly ' + adapted.slice(edit.at);
    if (adapted === text) continue;
    if (fs.readFileSync(fileName, 'utf8') !== text) throw new Error(`source changed during adaptation: ${fileName}`);
    fs.writeFileSync(fileName, adapted);
}
console.log(JSON.stringify({ typescript: ts.version, task: '#r11wppy', item: 1, change: 'internal',
    slots: edits.map(({ file, at, ...edit }) => edit) }, null, 2));
