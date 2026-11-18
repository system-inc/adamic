#!/usr/bin/env node
'use strict';
// #38kdd4d, family 4's typeExpression (Adamic's variant B, Nov 18): nodeFactory.ts's JSDoc type-like tag arrows and
// workers took `typeExpression` as the constraint's `JSDocTypeExpression | undefined` under a writable constraint
// `{ typeExpression?: JSDocTypeExpression }`. Each tag type, whose typeExpression is readonly and for four of them
// required, was refused where it instantiated the constraint (10 sites), and the worker's write needed
// 61-temporary-factory-type-expression's cast. Now, as the JSDoc unary type workers take `type`:
// - the four type parameters are constrained `{ readonly typeExpression?: JSDocTypeExpression }`;
// - the two arrows and the two workers take `typeExpression: T["typeExpression"]`, and the arrows call their worker as
//   `<T>`, so each tag's write is the plain `node.typeExpression = typeExpression` of its own type. 61, which stood in
//   for that write, is taken out whole beside this adaptation.
// The one real lie stays refused, by its own name: NodeFactory's public updateJSDocThisTag declares
// `typeExpression: JSDocTypeExpression | undefined`, and its getter now hands it an arrow that takes a JSDocThisTag's
// required JSDocTypeExpression. This directory sorts after 96-emit-jsdoc-tag-comment-parameters, whose pin reads the two
// arrows' parameter lists as upstream wrote them. Type-only, internal: no line of typescript.d.ts moves.
// Through the stock AST each edit is found by its owner and its exact upstream text, refused before any write if
// anything has moved, and left as it is on a second run.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
const file = 'src/compiler/factory/nodeFactory.ts';
const fileName = path.join(path.resolve(process.argv[2]), file);
const text = fs.readFileSync(fileName, 'utf8');
const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
const constraint = { upstream: 'JSDocTag & { typeExpression?: JSDocTypeExpression; }', adapted: 'JSDocTag & { readonly typeExpression?: JSDocTypeExpression; }' };
const typed = 'typeExpression: T["typeExpression"]';
const changes = [];
const report = [];

function where(node) {
    const at = source.getLineAndCharacterOfPosition(node.getStart(source));
    return `${file}:${at.line + 1}:${at.character + 1}`;
}
// One edit: the node's text is upstream's or already adapted, and nothing else.
function edit(label, node, upstream, adapted) {
    const written = node.getText(source);
    if (written === adapted) {
        report.push({ edit: label, where: where(node), edited: false });
        return;
    }
    if (written !== upstream) throw new Error(`unexpected text at ${label}: ${written}`);
    changes.push({ at: node.getStart(source), end: node.end, adapted });
    report.push({ edit: label, where: where(node), edited: true });
}
function only(nodes, label) {
    if (nodes.length !== 1) throw new Error(`expected exactly one ${label}, found ${nodes.length}`);
    return nodes[0];
}
function parameterNamed(parameters, name, label) {
    return only(parameters.filter(parameter => ts.isIdentifier(parameter.name) && parameter.name.text === name), `${label} parameter ${name}`);
}

const arrows = {}, workers = {};
ts.forEachChild(source, function find(node) {
    if (ts.isVariableDeclaration(node) && ts.isIdentifier(node.name) && /^getJSDocTypeLikeTag(Create|Update)Function$/.test(node.name.text)) {
        (arrows[node.name.text] ||= []).push(node);
    }
    if (ts.isFunctionDeclaration(node) && node.name && /^(create|update)JSDocTypeLikeTagWorker$/.test(node.name.text) && node.body) {
        (workers[node.name.text] ||= []).push(node);
    }
    ts.forEachChild(node, find);
});

// The arrows: memoizeOne(<T extends ...>(kind) => (..., typeExpression, ...) => worker(...)).
for (const [name, worker] of [['getJSDocTypeLikeTagCreateFunction', 'createJSDocTypeLikeTagWorker'], ['getJSDocTypeLikeTagUpdateFunction', 'updateJSDocTypeLikeTagWorker']]) {
    const declaration = only(arrows[name] || [], name);
    const call = declaration.initializer;
    if (!call || !ts.isCallExpression(call) || call.expression.getText(source) !== 'memoizeOne' || call.arguments.length !== 1 || !ts.isArrowFunction(call.arguments[0])) {
        throw new Error(`${name} is not memoizeOne of an arrow`);
    }
    const outer = call.arguments[0];
    const typeParameter = only(outer.typeParameters || [], `${name}'s type parameter`);
    if (!typeParameter.constraint) throw new Error(`${name}'s T has no constraint`);
    edit(`${name}: T's constraint`, typeParameter.constraint, constraint.upstream, constraint.adapted);
    const inner = outer.body;
    if (!ts.isArrowFunction(inner) || !ts.isCallExpression(inner.body)) throw new Error(`${name} doesn't return an arrow calling its worker`);
    edit(`${name}: typeExpression`, parameterNamed(inner.parameters, 'typeExpression', name), 'typeExpression?: JSDocTypeExpression', typed);
    const callee = inner.body;
    if (callee.expression.getText(source) !== worker) throw new Error(`${name} calls ${callee.expression.getText(source)}, not ${worker}`);
    if (!callee.typeArguments) {
        changes.push({ at: callee.expression.end, end: callee.expression.end, adapted: '<T>' });
        report.push({ edit: `${name}: calls ${worker}<T>`, where: where(callee), edited: true });
    }
    else if (callee.typeArguments.length === 1 && callee.typeArguments[0].getText(source) === 'T') {
        report.push({ edit: `${name}: calls ${worker}<T>`, where: where(callee), edited: false });
    }
    else throw new Error(`${name} calls ${worker} with type arguments ${callee.typeArguments.map(argument => argument.getText(source)).join(', ')}`);
}

// The workers: their own constraint and parameter.
for (const [name, upstream] of [['createJSDocTypeLikeTagWorker', 'typeExpression?: JSDocTypeExpression'], ['updateJSDocTypeLikeTagWorker', 'typeExpression: JSDocTypeExpression | undefined']]) {
    const worker = only(workers[name] || [], name);
    const typeParameter = only(worker.typeParameters || [], `${name}'s type parameter`);
    if (!typeParameter.constraint) throw new Error(`${name}'s T has no constraint`);
    edit(`${name}: T's constraint`, typeParameter.constraint, constraint.upstream, constraint.adapted);
    edit(`${name}: typeExpression`, parameterNamed(worker.parameters, 'typeExpression', name), upstream, typed);
}
// The create worker writes the plain field, as upstream does. 61's cast form means 61 is still in the patch set.
const createBody = workers.createJSDocTypeLikeTagWorker[0].body.getText(source);
if (createBody.includes('(node as { typeExpression: JSDocTypeExpression | undefined }).typeExpression')) {
    throw new Error('createJSDocTypeLikeTagWorker still writes through 61-temporary-factory-type-expression\'s cast; 61 comes out with this adaptation');
}
if (createBody.split('node.typeExpression = typeExpression;').length !== 2) throw new Error('createJSDocTypeLikeTagWorker doesn\'t write node.typeExpression once');

let adapted = text;
for (const change of changes.sort((a, b) => b.at - a.at)) adapted = adapted.slice(0, change.at) + change.adapted + adapted.slice(change.end);
if (adapted !== text) {
    if (fs.readFileSync(fileName, 'utf8') !== text) throw new Error(`source changed during adaptation: ${fileName}`);
    fs.writeFileSync(fileName, adapted);
}
console.log(JSON.stringify({ typescript: ts.version, task: '#38kdd4d', ruling: 'Adamic, Nov 18, variant B', edits: report }, null, 2));
