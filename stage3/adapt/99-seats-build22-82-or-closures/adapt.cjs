'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

const before = 'or(hasJSFileExtension, hasTSFileExtension)';
const after = 'text => hasJSFileExtension(text) || hasTSFileExtension(text)';

function inspect(tree) {
    const file = path.join(tree, 'src/compiler/utilities.ts');
    const text = fs.readFileSync(file, 'utf8');
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, 'utilities.ts: parse error');
    const owners = source.statements.filter(node =>
        ts.isFunctionDeclaration(node) && node.name?.text === 'usesExtensionsOnImports');
    assert.equal(owners.length, 1, 'expected one usesExtensionsOnImports declaration');
    const parameter = owners[0].parameters[1];
    assert(parameter && ts.isIdentifier(parameter.name) && parameter.name.text === 'hasExtension'
        && parameter.type?.getText(source) === '(text: string) => boolean'
        && parameter.initializer, 'usesExtensionsOnImports: default parameter drift');
    const initializer = parameter.initializer;
    const current = initializer.getText(source);
    const adapted = current === after;
    assert(adapted || current === before, 'usesExtensionsOnImports: callback snapshot drift');
    assert(adapted ? ts.isArrowFunction(initializer) : ts.isCallExpression(initializer),
        'usesExtensionsOnImports: callback shape drift');
    return { file, text, start: initializer.getStart(source), end: initializer.end, current, adapted };
}

const checkerBefore = 'or(isPropertyDeclaration, isAccessor, isVariableDeclaration, isPropertySignature, isBinaryExpression, isPropertyAccessExpression)';
const checkerAfter = '(d): boolean => isPropertyDeclaration(d) || isAccessor(d) || isVariableDeclaration(d) || isPropertySignature(d) || isBinaryExpression(d) || isPropertyAccessExpression(d)';

function inspectChecker(tree) {
    const file = path.join(tree, 'src/compiler/checker.ts');
    const text = fs.readFileSync(file, 'utf8');
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, 'checker.ts: parse error');
    const owners = [];
    function visit(node) {
        if (ts.isVariableDeclaration(node) && ts.isIdentifier(node.name) && node.name.text === 'firstPropertyLikeDecl') {
            owners.push(node);
        }
        ts.forEachChild(node, visit);
    }
    visit(source);
    assert.equal(owners.length, 1, 'expected one firstPropertyLikeDecl declaration');
    const find = owners[0].initializer;
    assert(find && ts.isCallExpression(find) && find.expression.getText(source) === 'p.declarations?.find'
        && find.arguments.length === 1, 'firstPropertyLikeDecl: find call drift');
    const callback = find.arguments[0];
    const current = callback.getText(source);
    const adapted = current === checkerAfter;
    assert(adapted || current === checkerBefore, 'firstPropertyLikeDecl: callback snapshot drift');
    assert(adapted ? ts.isArrowFunction(callback) : ts.isCallExpression(callback),
        'firstPropertyLikeDecl: callback shape drift');
    return { file, text, start: callback.getStart(source), end: callback.end, current, adapted };
}

function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const utilities = inspect(tree);
    const checker = inspectChecker(tree);
    assert.equal(checker.adapted, utilities.adapted, 'partly adapted tree');
    if (utilities.adapted) return 0;
    for (const [plan, replacement] of [[utilities, after], [checker, checkerAfter]]) {
        fs.writeFileSync(plan.file, plan.text.slice(0, plan.start) + replacement + plan.text.slice(plan.end));
    }
    return 2;
}

module.exports = { before, inspect, checkerBefore, inspectChecker };
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(`fixed-list or callbacks replaced: ${adapt(process.argv[2])}`);
}
