#!/usr/bin/env node
'use strict';
// #38kdd4d: a generated name's emitNode stops claiming more than tsc keeps (Adamic's road 3, Nov 18). This directory
// sorts after every adaptation whose pin holds text it edits: 91-patcher-discard (createNodeFactory's whole text, where
// the clones' two reads are), 99-correctness-binding-views (utilitiesPublic.ts's namespace import) and
// 99-method-value-views (classFields.ts's visitClassExpressionInNewClassLexicalEnvironment, where one read is).
// GeneratedIdentifier and GeneratedPrivateIdentifier redeclared `readonly emitNode: EmitNode & { autoGenerate:
// AutoGenerateInfo; }`, a slot narrower than Node's `EmitNode | undefined`. tsc's own constructors break that claim for a
// statement or two: getOrCreateEmitNode writes `{}` before setIdentifierAutoGenerate sets autoGenerate, and the two clones
// call setOriginal first, which writes a merged emitNode that never carries autoGenerate. Every value of the two types
// seen as a plain Identifier, PrivateIdentifier or Node was refused as a view that could write the narrower slot (family
// 4, 58 sites). So:
// - types.ts: the two interfaces drop the redeclaration and keep Node's emitNode. Each gains a phantom brand in its place,
//   `readonly __generatedIdentifier: undefined` or `readonly __generatedPrivateIdentifier: undefined` (rulings.md:32),
//   because the narrower slot was also what told the two types apart from Identifier and PrivateIdentifier: without a
//   brand, isGeneratedIdentifier's false branch narrows to `never` (Adamic's fourth road, Nov 18). Both are @internal, so
//   no line of the published typescript.d.ts moves.
// - the nine places that read `x.emitNode.autoGenerate` as present read it through tsc's own getter beside
//   setIdentifierAutoGenerate, as `Debug.checkDefined(getIdentifierAutoGenerate(x))`: a check that stops loudly if a
//   generated name ever lacks it, and never fires on tsc's own paths, where a constructor sets it before it returns.
//   Each file imports Debug and getIdentifierAutoGenerate where it doesn't yet.
// Through the stock AST each edit is found by its owner and its exact upstream text, refused before any write if anything
// has moved, and left as it is on a second run.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
const tree = path.resolve(process.argv[2]);
const redeclaration = 'readonly emitNode: EmitNode & { autoGenerate: AutoGenerateInfo; };';
// Each file that reads a generated name's autoGenerate as present, and how many times it does.
const readers = {
    'src/compiler/emitter.ts': 2,
    'src/compiler/utilitiesPublic.ts': 1,
    'src/compiler/transformers/classFields.ts': 1,
    'src/compiler/transformers/utilities.ts': 1,
    'src/compiler/transformers/module/module.ts': 1,
    'src/compiler/factory/utilities.ts': 1,
    'src/compiler/factory/nodeFactory.ts': 2,
};
const checked = name => `Debug.checkDefined(getIdentifierAutoGenerate(${name}))`;
const report = [];
const writes = [];

function parse(file) {
    const fileName = path.join(tree, file);
    const text = fs.readFileSync(fileName, 'utf8');
    return { file, fileName, text, source: ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true), changes: [] };
}
function where(parsed, node) {
    const at = parsed.source.getLineAndCharacterOfPosition(node.getStart(parsed.source));
    return `${parsed.file}:${at.line + 1}:${at.character + 1}`;
}
function walk(node, visit) {
    visit(node);
    ts.forEachChild(node, child => walk(child, visit));
}
function finish(parsed) {
    let adapted = parsed.text;
    for (const change of parsed.changes.sort((a, b) => b.at - a.at)) adapted = adapted.slice(0, change.at) + change.adapted + adapted.slice(change.end);
    if (adapted !== parsed.text) writes.push({ fileName: parsed.fileName, text: parsed.text, adapted });
}

// types.ts: each interface's one redeclaration becomes its phantom brand, in the same line (rulings.md:32's object
// brand: a required member typed undefined, under a plain string name). The brand keeps each type a strict subtype of
// Identifier or PrivateIdentifier, so isGeneratedIdentifier's false branch stays Identifier, never `never`.
const types = parse('src/compiler/types.ts');
for (const [name, brandName] of [['GeneratedIdentifier', '__generatedIdentifier'], ['GeneratedPrivateIdentifier', '__generatedPrivateIdentifier']]) {
    const brand = `readonly ${brandName}: undefined;`;
    const owners = [];
    walk(types.source, node => {
        if (ts.isInterfaceDeclaration(node) && node.name.text === name) owners.push(node);
    });
    if (owners.length !== 1) throw new Error(`expected one interface ${name} in types.ts, found ${owners.length}`);
    const named = text => owners[0].members.filter(member => member.name && ts.isIdentifier(member.name) && member.name.text === text);
    const members = named('emitNode'), brands = named(brandName);
    if (members.length === 0 && brands.length === 1 && brands[0].getText(types.source) === brand) {
        report.push({ edit: `${name}: emitNode becomes the phantom brand ${brandName}`, where: where(types, brands[0]), edited: false });
        continue;
    }
    if (members.length !== 1 || brands.length !== 0 || members[0].getText(types.source) !== redeclaration) {
        throw new Error(`unexpected emitNode in ${name}: ${members.map(member => member.getText(types.source)).join(' | ')}, ${brands.length} brands`);
    }
    types.changes.push({ at: members[0].getStart(types.source), end: members[0].end, adapted: brand });
    report.push({ edit: `${name}: emitNode becomes the phantom brand ${brandName}`, where: where(types, members[0]), edited: true });
}
finish(types);

// The readers: `x.emitNode.autoGenerate`, no optional chain, x a plain name, so x is read once either way.
for (const [file, expected] of Object.entries(readers)) {
    const parsed = parse(file);
    const upstream = [];
    let adapted = 0;
    walk(parsed.source, node => {
        if (ts.isPropertyAccessExpression(node) && node.name.text === 'autoGenerate' && !node.questionDotToken &&
            ts.isPropertyAccessExpression(node.expression) && node.expression.name.text === 'emitNode' && !node.expression.questionDotToken) {
            if (!ts.isIdentifier(node.expression.expression)) throw new Error(`${where(parsed, node)}: emitNode.autoGenerate read off ${node.expression.expression.getText(parsed.source)}, not a name`);
            upstream.push(node);
        }
        if (ts.isCallExpression(node) && node.getText(parsed.source).startsWith('Debug.checkDefined(getIdentifierAutoGenerate(')) adapted++;
    });
    if (!(upstream.length === expected && adapted === 0) && !(upstream.length === 0 && adapted === expected)) {
        throw new Error(`${file}: expected ${expected} reads of emitNode.autoGenerate, found ${upstream.length} as upstream and ${adapted} checked`);
    }
    for (const node of upstream) {
        parsed.changes.push({ at: node.getStart(parsed.source), end: node.end, adapted: checked(node.expression.expression.text) });
        report.push({ edit: `${file}: autoGenerate read through its getter, checked`, where: where(parsed, node), edited: true });
    }
    if (upstream.length === 0) report.push({ edit: `${file}: ${expected} autoGenerate reads through the getter, checked`, where: file, edited: false });
    // The namespace import gains Debug and getIdentifierAutoGenerate where it lacks them, each in the list's own order.
    const imports = parsed.source.statements.filter(statement =>
        ts.isImportDeclaration(statement) && ts.isStringLiteral(statement.moduleSpecifier) && /(^|\/)_namespaces\/ts\.js$/.test(statement.moduleSpecifier.text) &&
        statement.importClause && statement.importClause.namedBindings && ts.isNamedImports(statement.importClause.namedBindings));
    if (imports.length !== 1) throw new Error(`${file}: expected one named import from _namespaces/ts.js, found ${imports.length}`);
    const specifiers = imports[0].importClause.namedBindings.elements;
    for (const name of ['Debug', 'getIdentifierAutoGenerate']) {
        if (specifiers.some(specifier => specifier.name.text === name)) continue;
        const after = specifiers.find(specifier => specifier.name.text.toLowerCase() > name.toLowerCase());
        if (!after) throw new Error(`${file}: no specifier sorts after ${name}`);
        const at = after.getStart(parsed.source);
        const lineStart = parsed.text.lastIndexOf('\n', at) + 1;
        const indent = parsed.text.slice(lineStart, at);
        parsed.changes.push({ at, end: at, adapted: /^\s*$/.test(indent) ? `${name},\n${indent}` : `${name}, ` });
        report.push({ edit: `${file}: imports ${name}`, where: where(parsed, after), edited: true });
    }
    finish(parsed);
}

for (const write of writes) {
    if (fs.readFileSync(write.fileName, 'utf8') !== write.text) throw new Error(`source changed during adaptation: ${write.fileName}`);
    fs.writeFileSync(write.fileName, write.adapted);
}
console.log(JSON.stringify({ typescript: ts.version, task: '#38kdd4d', ruling: 'Adamic, Nov 18, road 3', edits: report }, null, 2));
