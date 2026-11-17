#!/usr/bin/env node
'use strict';
// Adaptation 96-emit, #v879py5: in src/compiler/factory/nodeFactory.ts, the four memoized arrows that build the JSDoc simple
// and type-like tag factories type their comment parameter narrower than both sides of them. Each is handed out by a
// lazy getter as a NodeFactory member, which types.ts declares as taking `comment?: string | NodeArray<JSDocComment>`,
// and each only forwards comment to a worker typed `string | NodeArray<JSDocComment>` too. tsc relates a method's
// parameters both ways, so the narrow annotation lets a public caller's string through to the arrow; Adamic judges the
// getter's function value seen as the member's type and refuses all 26 getters (13 tags, a create and an update each).
// Widening the four parameters to the workers' type makes the annotation say what the code does: the JavaScript is
// byte-identical, the arrows are file-local consts inside createNodeFactory, so typescript.d.ts doesn't move, and it's
// certified by the compiled program's gate as adaptation 72 is. README.md names the premise and the measurement.
//
// Through the stock AST the edit inserts `string | ` before each comment parameter's type, refuses unless it finds
// exactly the four arrows with upstream's own parameter lists, and changes nothing on a second run.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
const tree = path.resolve(process.argv[2]);
const fileName = path.join(tree, 'src/compiler/factory/nodeFactory.ts');
// Each arrow's returned function's parameter list, as upstream 6.0.3 writes it: an arrow whose parameters have moved
// is not the arrow this adaptation was measured on.
const narrow = 'NodeArray<JSDocComment>';
const wide = `string | ${narrow}`;
const parameters = {
    getJSDocSimpleTagCreateFunction: 'tagName: Identifier | undefined, comment?: COMMENT',
    getJSDocSimpleTagUpdateFunction: 'node: T, tagName: Identifier | undefined, comment?: COMMENT',
    getJSDocTypeLikeTagCreateFunction: 'tagName: Identifier | undefined, typeExpression?: JSDocTypeExpression, comment?: COMMENT',
    getJSDocTypeLikeTagUpdateFunction: 'node: T, tagName: Identifier | undefined, typeExpression?: JSDocTypeExpression, comment?: COMMENT',
};
const text = fs.readFileSync(fileName, 'utf8');
const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
const declarations = [];
ts.forEachChild(source, function find(node) {
    if (ts.isVariableDeclaration(node) && ts.isIdentifier(node.name) && Object.hasOwn(parameters, node.name.text))
        declarations.push(node);
    ts.forEachChild(node, find);
});
const names = declarations.map(declaration => declaration.name.text).sort();
if (names.join(',') !== Object.keys(parameters).sort().join(','))
    throw new Error(`expected exactly one of each of ${Object.keys(parameters).sort().join(', ')} in nodeFactory.ts, found ${JSON.stringify(names)}`);
const edits = [];
for (const declaration of declarations) {
    const name = declaration.name.text;
    const where = source.getLineAndCharacterOfPosition(declaration.getStart(source));
    const record = { arrow: name, where: `src/compiler/factory/nodeFactory.ts:${where.line + 1}:${where.character + 1}` };
    // memoizeOne(<T ...>(kind: T["kind"]) => (<parameters>) => <worker>(...)): the returned function is the inner arrow.
    const call = declaration.initializer;
    const outer = call && ts.isCallExpression(call) && ts.isIdentifier(call.expression) && call.expression.text === 'memoizeOne' &&
        call.arguments.length === 1 && ts.isArrowFunction(call.arguments[0]) ? call.arguments[0] : undefined;
    const inner = outer && ts.isArrowFunction(outer.body) ? outer.body : undefined;
    if (!inner) throw new Error(`unexpected ${name}: not memoizeOne of an arrow returning an arrow: ${declaration.getText(source)}`);
    const list = inner.parameters.map(parameter => parameter.getText(source)).join(', ');
    const comment = inner.parameters[inner.parameters.length - 1];
    const typeText = comment.type && comment.type.getText(source);
    // Idempotent: a parameter already widened keeps its bytes, so a second run changes nothing.
    if (list === parameters[name].replace('COMMENT', wide)) {
        edits.push({ ...record, edited: false });
        continue;
    }
    if (list !== parameters[name].replace('COMMENT', narrow) || typeText !== narrow)
        throw new Error(`unexpected ${name}: its function's parameters are (${list}), where upstream's are (${parameters[name].replace('COMMENT', narrow)})`);
    edits.push({ ...record, edited: true, at: comment.type.getStart(source) });
}
let adapted = text;
for (const edit of edits.filter(edit => edit.edited).sort((a, b) => b.at - a.at))
    adapted = adapted.slice(0, edit.at) + 'string | ' + adapted.slice(edit.at);
if (adapted !== text) {
    if (fs.readFileSync(fileName, 'utf8') !== text) throw new Error(`source changed during adaptation: ${fileName}`);
    fs.writeFileSync(fileName, adapted);
}
console.log(JSON.stringify({ typescript: ts.version, task: '#v879py5', certificate: "the compiled program's gate",
    arrows: edits.sort((a, b) => a.arrow.localeCompare(b.arrow)).map(({ at, ...edit }) => edit) }, null, 2));
