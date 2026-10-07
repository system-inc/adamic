#!/usr/bin/env node
'use strict';
// Tooling only. Copy reviewed optional contracts at their declaration owners.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.TSC_ADAPT_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error(`expected TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
const tree = path.resolve(process.argv[2]);
const filename = path.join(tree, 'src/compiler/types.ts');
const program = ts.createProgram([filename], {
    strict: true, exactOptionalPropertyTypes: true, noUncheckedIndexedAccess: true,
    verbatimModuleSyntax: true, erasableSyntaxOnly: true, noEmit: true,
    target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext,
    moduleResolution: ts.ModuleResolutionKind.Bundler,
});
const checker = program.getTypeChecker();
const source = program.getSourceFile(filename);
const interfaces = new Map();
function visit(node) {
    if (ts.isInterfaceDeclaration(node)) {
        if (!interfaces.has(node.name.text)) interfaces.set(node.name.text, []);
        interfaces.get(node.name.text).push(node);
    }
    ts.forEachChild(node, visit);
}
visit(source);
function declaration(name) {
    const nodes = interfaces.get(name);
    if (nodes?.length !== 1) throw new Error(`ambiguous owner ${name}`);
    return nodes[0];
}
const specifications = [
    { owner: 'Type', target: 'InstantiableType', properties: ['resolvedBaseConstraint', 'resolvedIndexType', 'resolvedStringIndexType'] },
    { owner: 'SymbolVisibilityResult', target: 'SymbolAccessibilityResult', properties: ['errorModuleName'] },
];
const edits = [];
const records = [];
for (const spec of specifications) {
    const owner = declaration(spec.owner);
    const target = declaration(spec.target);
    const ownerType = checker.getTypeAtLocation(owner);
    const targetType = checker.getTypeAtLocation(target);
    for (const name of spec.properties) {
        const member = target.members.find(node => node.name && ts.isIdentifier(node.name) && node.name.text === name);
        if (!member || !ts.isPropertySignature(member) || !member.questionToken || !member.type) {
            throw new Error(`expected optional contract ${spec.target}.${name}`);
        }
        const actual = checker.getPropertyOfType(ownerType, name);
        const expected = checker.getPropertyOfType(targetType, name);
        if (actual) {
            const actualType = checker.getNonNullableType(checker.getTypeOfSymbolAtLocation(actual, owner));
            const expectedType = checker.getNonNullableType(checker.getTypeOfSymbolAtLocation(expected, target));
            if (!(actual.flags & ts.SymbolFlags.Optional) ||
                !checker.isTypeAssignableTo(actualType, expectedType) ||
                !checker.isTypeAssignableTo(expectedType, actualType)) {
                throw new Error(`incompatible existing contract ${spec.owner}.${name}`);
            }
            continue;
        }
        let typeText = member.type.getText(source);
        if (!checker.isTypeAssignableTo(checker.getUndefinedType(), checker.getTypeFromTypeNode(member.type))) {
            typeText += ' | undefined';
        }
        const newline = source.text.includes('\r\n') ? '\r\n' : '\n';
        edits.push({ at: owner.members.end, text: `${newline}    /** @internal */${newline}    ${name}?: ${typeText};` });
        records.push({ owner: spec.owner, property: name, target: spec.target, type: typeText });
    }
}
// These lazy caches are absent on newly constructed union/intersection types.
// Their parallel required declarations must match the truthful optional base
// slot, including at multiple-inheritance joins such as ResolvedType.
const unionOwner = declaration('UnionOrIntersectionType');
for (const name of specifications[0].properties) {
    const member = unionOwner.members.find(node => node.name?.getText(source) === name);
    if (!member || !ts.isPropertySignature(member) || !member.type) throw new Error(`missing union cache ${name}`);
    if (!member.questionToken) edits.push({ at: member.name.end, text: '?' });
    if (!checker.isTypeAssignableTo(checker.getUndefinedType(), checker.getTypeFromTypeNode(member.type))) {
        edits.push({ at: member.type.end, text: ' | undefined' });
    }
    if (!member.questionToken || !checker.isTypeAssignableTo(checker.getUndefinedType(), checker.getTypeFromTypeNode(member.type))) {
        records.push({ owner: 'UnionOrIntersectionType', property: name, target: 'InstantiableType', type: member.type.getText(source) + ' | undefined', correction: 'lazy cache can be absent' });
    }
}
let text = source.text;
for (const edit of edits.sort((a, b) => b.at - a.at)) text = text.slice(0, edit.at) + edit.text + text.slice(edit.at);
if (edits.length) fs.writeFileSync(filename, text);
console.log(JSON.stringify({ files: edits.length ? 1 : 0, additions: records }, null, 2));
