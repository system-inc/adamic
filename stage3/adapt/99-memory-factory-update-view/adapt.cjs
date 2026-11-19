#!/usr/bin/env node
'use strict';
// Adaptation 99-memory-factory-update-view, #38kdd4d: nodeFactory.ts's `update`, the helper inside createNodeFactory
// that every updateX hands its fresh node to, takes that node as `updated: T`, not `updated: Mutable<T>`. Its body writes
// nothing itself: it calls setOriginal(updated, original), which writes the node's `original` and `emitNode`, and
// setTextRange(updated, original), which writes its `pos` and `end`. Node declares `original` and `emitNode` writable,
// and `pos` and `end` are writable since 80-writable-node-range. So the Mutable view, which lifts readonly from every
// field, was wider than anything the function writes, and Adamic refused each `update(createX(...), node)` call as a
// value of type X seen as Mutable<X>, whose readonly field flags or kind becomes writable: 44 sites. The ruling is the
// one Adamic delegated on #k8ed1f2 and 99-readonly-family3-views applies: a view narrowed where tsc never writes.
// Type-only: the parameter's type erases, so the emitted JavaScript does not change. `update` is a function declared in
// createNodeFactory's body, itself `@internal`, so typescript.d.ts does not change and no api-sanctions.json is needed.
// The finishUpdate helpers keep their Mutable parameters, since each writes a field through it; README.md says why.
// Through the stock AST the parameter is found by its function's name and its own, never by line, offset or whole-line
// text, and the edit replaces only that parameter's type node. It refuses before any write unless there is exactly one
// function named update in nodeFactory.ts, a statement of createNodeFactory's own body, with upstream's signature and
// the body the premise reads, setOriginal bound as upstream binds it, setOriginalNode writing only `original` and
// `emitNode`, and every use of the name a call. A parameter already `T` keeps its bytes, so a second run changes
// nothing. No adaptation before it edits createNodeFactory's `update`, setOriginal or setOriginalNode, so it expects
// upstream's text there; 91-patcher-discard pins createNodeFactory's whole text and runs first, so its pin reads
// upstream's `update`. None of the eight that sort after it reads or pins `update`'s declaration.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
const tree = path.resolve(process.argv[2]);
const relative = 'src/compiler/factory/nodeFactory.ts';
const fileName = path.join(tree, relative);
const text = fs.readFileSync(fileName, 'utf8');
const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);

// The parameter as upstream 6.0.3 writes it, and as this adaptation leaves it: any other text is not the slot the
// ruling covers.
const parameter = { name: 'updated', upstream: 'Mutable<T>', adapted: 'T' };
// update's signature apart from that parameter, and its body, statement by statement, whitespace folded: the premise is
// that the body writes nothing itself and hands the node to setOriginal and setTextRange alone.
const typeParameters = 'T extends Node';
const original = { name: 'original', type: 'T' };
const returnType = 'T';
const body = [
    'if (updated !== original) { setOriginal(updated, original); setTextRange(updated, original); }',
    'return updated;',
];
// setOriginal, createNodeFactory's const, as upstream binds it: identity, which writes nothing, or setOriginalNode.
const setOriginalBinding = 'flags & NodeFactoryFlags.NoOriginalNode ? identity : setOriginalNode';
// The node fields setOriginalNode assigns, each declared writable on Node.
const setOriginalNodeWrites = ['node.emitNode', 'node.original'];

const fold = node => node.getText(source).replace(/\s+/g, ' ');
function where(node) {
    const at = source.getLineAndCharacterOfPosition(node.getStart(source));
    return `${relative}:${at.line + 1}:${at.character + 1}`;
}
function nodes(root, predicate) {
    const found = [];
    ts.forEachChild(root, function visit(node) {
        if (predicate(node)) found.push(node);
        ts.forEachChild(node, visit);
    });
    return found;
}
function one(found, what) {
    if (found.length !== 1) throw new Error(`expected one ${what} in ${relative}, found ${found.length}`);
    return found[0];
}
const functionNamed = name => one(nodes(source, node => ts.isFunctionDeclaration(node) && node.name?.text === name),
    `function ${name}`);

// createNodeFactory, `@internal`, and update, a statement of its own body: local to it, so never in typescript.d.ts.
const factory = functionNamed('createNodeFactory');
if (!ts.getJSDocTags(factory).some(tag => tag.tagName.text === 'internal'))
    throw new Error(`expected createNodeFactory at ${where(factory)} to be @internal`);
const update = functionNamed('update');
if (update.parent !== factory.body)
    throw new Error(`expected update at ${where(update)} to be a statement of createNodeFactory's own body`);
if (update.modifiers?.length) throw new Error(`expected update at ${where(update)} to carry no modifiers`);

// The signature.
const typeParameterText = (update.typeParameters ?? []).map(node => node.getText(source)).join(', ');
if (typeParameterText !== typeParameters)
    throw new Error(`expected update<${typeParameters}> at ${where(update)}, found <${typeParameterText}>`);
if (update.type?.getText(source) !== returnType)
    throw new Error(`expected update at ${where(update)} to return ${returnType}, found ${update.type?.getText(source)}`);
const [updated, originalParameter, ...rest] = update.parameters;
if (rest.length || !updated || !originalParameter)
    throw new Error(`expected update at ${where(update)} to take two parameters, found ${update.parameters.length}`);
for (const [node, name] of [[updated, parameter.name], [originalParameter, original.name]]) {
    if (!ts.isIdentifier(node.name) || node.name.text !== name || node.dotDotDotToken || node.questionToken ||
        node.initializer || node.modifiers?.length || !node.type)
        throw new Error(`expected update's parameter ${name} at ${where(node)}, plain and typed, found ${node.getText(source)}`);
}
if (originalParameter.type.getText(source) !== original.type)
    throw new Error(`expected update's ${original.name}: ${original.type} at ${where(originalParameter)}, ` +
        `found ${originalParameter.type.getText(source)}`);

// The premise, read from the source: the body writes nothing itself, setOriginal is identity or setOriginalNode, and
// setOriginalNode assigns only `original` and `emitNode` on the node. setTextRange is an import, utilitiesPublic.ts's,
// whose writes go through setTextRangePos and setTextRangeEnd to `pos` and `end`; README.md names those lines.
const statements = update.body?.statements.map(fold) ?? [];
if (statements.join('\n') !== body.join('\n'))
    throw new Error(`expected update's body at ${where(update)} to be upstream's, found ${JSON.stringify(statements)}`);
const setOriginal = one(nodes(factory.body, node => ts.isVariableDeclaration(node) && ts.isIdentifier(node.name)
    && node.name.text === 'setOriginal'), "createNodeFactory's const setOriginal");
if (setOriginal.parent.parent.parent !== factory.body
    || (setOriginal.parent.flags & ts.NodeFlags.BlockScoped) !== ts.NodeFlags.Const
    || setOriginal.initializer?.getText(source) !== setOriginalBinding)
    throw new Error(`expected setOriginal at ${where(setOriginal)} to be createNodeFactory's \`const setOriginal = ${setOriginalBinding}\``);
const setOriginalNode = functionNamed('setOriginalNode');
const assigned = nodes(setOriginalNode.body, node => ts.isBinaryExpression(node)
    && node.operatorToken.kind >= ts.SyntaxKind.FirstAssignment && node.operatorToken.kind <= ts.SyntaxKind.LastAssignment)
    .map(node => node.left.getText(source)).sort();
const mutated = nodes(setOriginalNode.body, node => (ts.isPrefixUnaryExpression(node) || ts.isPostfixUnaryExpression(node))
    && [ts.SyntaxKind.PlusPlusToken, ts.SyntaxKind.MinusMinusToken].includes(node.operator)
    || ts.isDeleteExpression(node));
if (assigned.join(', ') !== setOriginalNodeWrites.join(', ') || mutated.length)
    throw new Error(`expected setOriginalNode at ${where(setOriginalNode)} to assign ${setOriginalNodeWrites.join(' and ')} alone, ` +
        `found ${JSON.stringify(assigned)} and ${mutated.length} increments or deletes`);

// Every use of the name in createNodeFactory is a call of it, so its parameter's type is read nowhere else. A property
// name (`.update`, `update:`) is another symbol and is passed over.
let calls = 0;
for (const use of nodes(factory.body, node => ts.isIdentifier(node) && node.text === 'update')) {
    if (use === update.name) continue;
    const parent = use.parent;
    if ((ts.isPropertyAccessExpression(parent) && parent.name === use) || (ts.isPropertyAssignment(parent) && parent.name === use))
        continue;
    if (!ts.isCallExpression(parent) || parent.expression !== use)
        throw new Error(`expected every use of update in createNodeFactory to be a call of it, found ${fold(parent)} at ${where(use)}`);
    calls++;
}
const outside = nodes(source, node => ts.isIdentifier(node) && node.text === 'update'
    && !(node.pos >= factory.pos && node.end <= factory.end)
    && !(ts.isPropertyAccessExpression(node.parent) && node.parent.name === node));
if (outside.length) throw new Error(`expected no use of update outside createNodeFactory, found one at ${where(outside[0])}`);

// The one edit.
const typeText = updated.type.getText(source);
let adapted = text;
if (typeText === parameter.upstream)
    adapted = text.slice(0, updated.type.getStart(source)) + parameter.adapted + text.slice(updated.type.end);
else if (typeText !== parameter.adapted)
    throw new Error(`unexpected update parameter ${parameter.name}: ${typeText} at ${where(updated.type)}`);
if (adapted !== text) {
    if (fs.readFileSync(fileName, 'utf8') !== text) throw new Error(`source changed during adaptation: ${fileName}`);
    fs.writeFileSync(fileName, adapted);
}
console.log(`update ${parameter.name}: ${adapted !== text ? `edited, ${parameter.upstream} to ${parameter.adapted}` : 'already adapted'} ` +
    `(${where(updated.type)}, local to createNodeFactory, ${calls} calls)`);
