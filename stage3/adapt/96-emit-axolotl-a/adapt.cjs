'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require('typescript');

// transformNodes' lexical environment, declared T[] but holding undefined placeholders between environments (#vy7ncrg).
// Each local and its stack tell the type the truth; every reader already tests absence (README.md).
const file = 'src/compiler/transformer.ts';
const locals = [
    {name: 'lexicalEnvironmentVariableDeclarations', before: 'VariableDeclaration[]', after: 'VariableDeclaration[] | undefined'},
    {name: 'lexicalEnvironmentFunctionDeclarations', before: 'FunctionDeclaration[]', after: 'FunctionDeclaration[] | undefined'},
    {name: 'lexicalEnvironmentStatements', before: 'Statement[]', after: 'Statement[] | undefined'},
    {name: 'blockScopedVariableDeclarations', before: 'Identifier[]', after: 'Identifier[] | undefined'},
];
const stacks = [
    {name: 'lexicalEnvironmentVariableDeclarationsStack', before: 'VariableDeclaration[][]', after: '(VariableDeclaration[] | undefined)[]'},
    {name: 'lexicalEnvironmentFunctionDeclarationsStack', before: 'FunctionDeclaration[][]', after: '(FunctionDeclaration[] | undefined)[]'},
    {name: 'lexicalEnvironmentStatementsStack', before: 'Statement[][]', after: '(Statement[] | undefined)[]'},
    {name: 'blockScopedVariableDeclarationsStack', before: 'Identifier[][]', after: '(Identifier[] | undefined)[]'},
];
// A local's placeholder becomes plain undefined; the stacks' disposal placeholders stay, as #9wc5q5j reads them.
const placeholders = {lexicalEnvironmentVariableDeclarations: 2, lexicalEnvironmentFunctionDeclarations: 2, lexicalEnvironmentStatements: 1, blockScopedVariableDeclarations: 1};

function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const fileName = path.join(tree, file);
    const text = fs.readFileSync(fileName, 'utf8');
    const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, 'cannot parse transformer.ts');
    const functions = source.statements.filter(n => ts.isFunctionDeclaration(n) && n.name?.text === 'transformNodes');
    assert.equal(functions.length, 1, 'transformNodes moved');
    const body = functions[0].body;
    const edits = [];
    let adapted = 0;
    for (const target of [...locals, ...stacks]) {
        const declarations = body.statements.filter(ts.isVariableStatement).flatMap(s => s.declarationList.declarations).filter(d => ts.isIdentifier(d.name) && d.name.text === target.name);
        assert.equal(declarations.length, 1, `${target.name} moved`);
        const type = declarations[0].type;
        assert(type, `${target.name} lost its annotation`);
        const written = type.getText(source);
        assert([target.before, target.after].includes(written), `${target.name} is ${written}, not an inventoried shape`);
        if (written === target.after) adapted++;
        else edits.push({start: type.getStart(source), end: type.end, text: target.after});
    }
    const found = {};
    function visit(node) {
        if (ts.isBinaryExpression(node) && node.operatorToken.kind === ts.SyntaxKind.EqualsToken && ts.isIdentifier(node.left) && node.left.text in placeholders) {
            const right = node.right;
            if (ts.isNonNullExpression(right) && right.expression.kind === ts.SyntaxKind.UndefinedKeyword || ts.isNonNullExpression(right) && ts.isIdentifier(right.expression) && right.expression.text === 'undefined') {
                found[node.left.text] = (found[node.left.text] ?? 0) + 1;
                edits.push({start: right.getStart(source), end: right.end, text: 'undefined'});
            } else if (ts.isIdentifier(right) && right.text === 'undefined') {
                found[node.left.text] = (found[node.left.text] ?? 0) + 1;
            }
        }
        ts.forEachChild(node, visit);
    }
    visit(body);
    for (const [name, count] of Object.entries(placeholders)) assert.equal(found[name] ?? 0, count, `${name}'s placeholders moved`);
    // Either untouched or already adapted, never half: a partial shape is drift.
    assert(adapted === 0 || adapted === locals.length + stacks.length, 'transformer.ts is partly adapted');
    let result = text;
    for (const edit of edits.sort((a, b) => b.start - a.start)) result = result.slice(0, edit.start) + edit.text + result.slice(edit.end);
    assert.equal(ts.createSourceFile(fileName, result, ts.ScriptTarget.Latest, true).parseDiagnostics.length, 0);
    assert.equal(fs.readFileSync(fileName, 'utf8'), text, 'source changed while planning');
    if (edits.length) fs.writeFileSync(fileName, result);
    return {files: edits.length ? 1 : 0, edits: edits.length, javascriptChanging: false};
}
module.exports = {adapt, locals, stacks, placeholders};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
