#!/usr/bin/env node
'use strict';
// Adaptation 87, #38kdd4d: the three writes of a node's emitNode that tsc makes by hand, outside emitNode.ts's setters,
// go away (Adamic's ruling of item 3, Nov 18, kept under road 3):
// - factory/emitHelpers.ts: createAsyncGeneratorHelper and createAwaiterHelper write
//   `(generatorFunc.emitNode || (generatorFunc.emitNode = {} as EmitNode)).flags |= ...`. Each now calls
//   getOrCreateEmitNode(generatorFunc), which differs from the inline `||` only for a parse-tree node (it records the node
//   in its source file's annotatedNodes) or an Immutable emitNode (it asserts). generatorFunc is neither: createAwaiterHelper
//   makes it six lines above, and createAsyncGeneratorHelper's one caller, es2018.ts, hands it a createFunctionExpression it
//   has just made. The lane witnessed it with Debug.assert(!isParseTreeNode(generatorFunc)) before both writes.
//   EmitNode leaves the import list, since nothing else in the file names it, and getOrCreateEmitNode joins it.
// - expressionToTypeNode.ts: typeFromArrayLiteral makes a tuple, writes its emitNode, and returns notImplemented on the
//   next line, so nothing ever reads the tuple. The write is dropped and the create stays, as a statement.
// Through the stock AST each edit is found by its owner and its exact upstream text, refused before any write if anything
// has moved, and left as it is on a second run.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
const tree = path.resolve(process.argv[2]);
const inlineWrite = '(generatorFunc.emitNode || (generatorFunc.emitNode = {} as EmitNode)).flags |= EmitFlags.AsyncFunctionBody | EmitFlags.ReuseTempVariableScope;';
const helperWrite = 'getOrCreateEmitNode(generatorFunc).flags |= EmitFlags.AsyncFunctionBody | EmitFlags.ReuseTempVariableScope;';
const tupleCreate = 'const tupleType = factory.createTupleTypeNode(elementTypesInfo);';
const tupleWrite = 'tupleType.emitNode = { flags: 1, autoGenerate: undefined, internalFlags: 0 };';
const tupleStatement = 'factory.createTupleTypeNode(elementTypesInfo);';
const report = [];
const writes = [];

function parse(file) {
    const fileName = path.join(tree, file);
    const text = fs.readFileSync(fileName, 'utf8');
    return { file, fileName, text, source: ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true), changes: [] };
}
function functionsNamed(parsed, name) {
    const found = [];
    ts.forEachChild(parsed.source, function find(node) {
        if (ts.isFunctionDeclaration(node) && node.name && node.name.text === name) found.push(node);
        ts.forEachChild(node, find);
    });
    if (found.length !== 1 || !found[0].body) throw new Error(`expected exactly one ${name} with a body in ${parsed.file}, found ${found.length}`);
    return found[0];
}
function where(parsed, node) {
    const at = parsed.source.getLineAndCharacterOfPosition(node.getStart(parsed.source));
    return `${parsed.file}:${at.line + 1}:${at.character + 1}`;
}
function finish(parsed) {
    let adapted = parsed.text;
    for (const change of parsed.changes.sort((a, b) => b.at - a.at)) adapted = adapted.slice(0, change.at) + change.adapted + adapted.slice(change.end);
    if (adapted !== parsed.text) writes.push({ fileName: parsed.fileName, text: parsed.text, adapted });
}

// emitHelpers.ts: each helper's own statement, by its exact text among the function body's statements.
const helpers = parse('src/compiler/factory/emitHelpers.ts');
for (const owner of ['createAsyncGeneratorHelper', 'createAwaiterHelper']) {
    const body = functionsNamed(helpers, owner).body.statements;
    const upstream = body.filter(statement => statement.getText(helpers.source) === inlineWrite);
    const adapted = body.filter(statement => statement.getText(helpers.source) === helperWrite);
    if (upstream.length + adapted.length !== 1) throw new Error(`expected one emitNode write in ${owner}, found ${upstream.length} inline and ${adapted.length} through getOrCreateEmitNode`);
    const statement = upstream[0] || adapted[0];
    report.push({ edit: `${owner}: generatorFunc's emitNode through getOrCreateEmitNode`, where: where(helpers, statement), edited: upstream.length === 1 });
    if (upstream.length === 1) helpers.changes.push({ at: statement.getStart(helpers.source), end: statement.end, adapted: helperWrite });
}
// The import list from the namespace: EmitNode goes, since the two writes were its only readers, and getOrCreateEmitNode
// comes in after getEmitScriptTarget, in the list's own order.
const imports = helpers.source.statements.filter(statement =>
    ts.isImportDeclaration(statement) && ts.isStringLiteral(statement.moduleSpecifier) && statement.moduleSpecifier.text === '../_namespaces/ts.js');
if (imports.length !== 1 || !imports[0].importClause || !imports[0].importClause.namedBindings || !ts.isNamedImports(imports[0].importClause.namedBindings)) {
    throw new Error(`expected one named import from ../_namespaces/ts.js in emitHelpers.ts, found ${imports.length}`);
}
const specifiers = imports[0].importClause.namedBindings.elements;
const named = name => specifiers.filter(specifier => specifier.name.text === name && !specifier.propertyName);
const emitNode = named('EmitNode'), helper = named('getOrCreateEmitNode');
const stillNamed = [];
ts.forEachChild(helpers.source, function find(node) {
    if (ts.isIdentifier(node) && node.text === 'EmitNode' && !ts.isImportSpecifier(node.parent)) stillNamed.push(where(helpers, node));
    ts.forEachChild(node, find);
});
const writesLeft = helpers.changes.length;
// Once both writes are adapted, EmitNode is named nowhere but its import.
if (stillNamed.length !== writesLeft) throw new Error(`EmitNode is named outside the two writes in emitHelpers.ts: ${stillNamed.join(', ')}`);
if (emitNode.length > 1 || helper.length > 1 || emitNode.length + helper.length === 0) {
    throw new Error(`expected EmitNode or getOrCreateEmitNode once in emitHelpers.ts's import list, found ${emitNode.length} and ${helper.length}`);
}
report.push({ edit: 'emitHelpers.ts imports getOrCreateEmitNode for EmitNode', where: where(helpers, (emitNode[0] || helper[0])), edited: emitNode.length === 1 });
if (emitNode.length === 1) {
    const specifier = emitNode[0];
    if (helper.length === 0) {
        // The specifier's name is rewritten where it stands, whatever the list's layout.
        helpers.changes.push({ at: specifier.getStart(helpers.source), end: specifier.end, adapted: 'getOrCreateEmitNode' });
    }
    else {
        // Another adaptation already imports getOrCreateEmitNode, so EmitNode's specifier goes with the separator
        // after it, up to the next specifier.
        const index = specifiers.indexOf(specifier);
        const next = specifiers[index + 1];
        if (!next) throw new Error('EmitNode is the last specifier in emitHelpers.ts\'s import list');
        helpers.changes.push({ at: specifier.getStart(helpers.source), end: next.getStart(helpers.source), adapted: '' });
    }
}
finish(helpers);

// expressionToTypeNode.ts: the dead tuple's create and write, as two neighboring statements of typeFromArrayLiteral,
// become the create alone.
const builder = parse('src/compiler/expressionToTypeNode.ts');
const owner = functionsNamed(builder, 'typeFromArrayLiteral');
const statements = owner.body.statements;
const createAt = statements.findIndex(statement => statement.getText(builder.source) === tupleCreate);
const adaptedAt = statements.findIndex(statement => statement.getText(builder.source) === tupleStatement);
const tupleNames = [];
ts.forEachChild(owner, function find(node) {
    if (ts.isIdentifier(node) && node.text === 'tupleType') tupleNames.push(node);
    ts.forEachChild(node, find);
});
if (createAt >= 0) {
    if (adaptedAt >= 0 || statements[createAt + 1] === undefined || statements[createAt + 1].getText(builder.source) !== tupleWrite) {
        throw new Error('typeFromArrayLiteral: the tuple\'s create is not followed by its emitNode write');
    }
    // The tuple is named twice, by its declaration and by the write, and read nowhere.
    if (tupleNames.length !== 2) throw new Error(`typeFromArrayLiteral names tupleType ${tupleNames.length} times, not 2`);
    // The function ends `return notImplemented;`, after the write and one restore of the context.
    const last = statements[statements.length - 1];
    if (statements.length !== createAt + 4 || !ts.isReturnStatement(last) || !last.expression || last.expression.getText(builder.source) !== 'notImplemented') {
        throw new Error('typeFromArrayLiteral: the tuple\'s write is not two statements before `return notImplemented`');
    }
    builder.changes.push({ at: statements[createAt].getStart(builder.source), end: statements[createAt + 1].end, adapted: tupleStatement });
    report.push({ edit: 'typeFromArrayLiteral: the dead tuple\'s emitNode write dropped', where: where(builder, statements[createAt]), edited: true });
}
else {
    if (adaptedAt < 0 || tupleNames.length !== 0) throw new Error('typeFromArrayLiteral holds neither the tuple\'s create and write nor the adapted create');
    report.push({ edit: 'typeFromArrayLiteral: the dead tuple\'s emitNode write dropped', where: where(builder, statements[adaptedAt]), edited: false });
}
finish(builder);

for (const write of writes) {
    if (fs.readFileSync(write.fileName, 'utf8') !== write.text) throw new Error(`source changed during adaptation: ${write.fileName}`);
    fs.writeFileSync(write.fileName, write.adapted);
}
console.log(JSON.stringify({ typescript: ts.version, task: '#38kdd4d', ruling: 'Adamic, Nov 18, item 3', edits: report }, null, 2));
