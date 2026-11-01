'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require('typescript');

// Required reads only, at their original evaluation points. README.md records
// each invariant and the declaration read that cannot honestly be asserted.
const sites = [
    {file: 'esDecorators.ts', function: 'transformConstructorBodyWorker', expression: 'superPath[superPathDepth]', assert: [1], total: 1},
    {file: 'esDecorators.ts', function: 'transformConstructorBodyWorker', expression: 'statementsIn[superStatementIndex]', assert: [1], total: 1},
    {file: 'esnext.ts', function: 'visitSourceFile', expression: 'node.statements[pos]', assert: [1], total: 1},
    {file: 'esnext.ts', function: 'transformUsingDeclarations', expression: 'statementsIn[i]', assert: [1], total: 1},
    {file: 'esnext.ts', function: 'countPrologueStatements', expression: 'statements[i]', assert: [1, 2], total: 2},
    {file: 'jsx.ts', function: 'convertJsxChildrenToChildrenPropAssignment', expression: 'nonWhitespaceChildren[0]', assert: [2], total: 2},
    {file: 'jsx.ts', function: 'transformJsxAttributesToExpression', expression: 'expressions[0]', assert: [1], total: 1},
    {file: 'module/esnextAnd2015.ts', function: 'transformSourceFile', expression: 'node.arguments[0]', assert: [1], total: 2},
    {file: 'module/esnextAnd2015.ts', function: 'visitor', expression: 'importsAndRequiresToRewriteOrShim[0]', assert: [1], total: 1},
    {file: 'module/esnextAnd2015.ts', function: 'visitImportOrRequireCall', expression: 'node.arguments[0]', assert: [1, 3], total: 3},
    {file: 'module/esnextAnd2015.ts', function: 'createRequireCall', expression: 'importRequireStatements[1].declarationList.declarations[0]', assert: [1], total: 1},
];

function nodes(root, predicate) {
    const found = [];
    (function visit(node) { if (predicate(node)) found.push(node); ts.forEachChild(node, visit); })(root);
    return found;
}

function isStore(read) {
    let target = read;
    while (ts.isParenthesizedExpression(target.parent) || ts.isNonNullExpression(target.parent)) target = target.parent;
    const parent = target.parent;
    return (ts.isBinaryExpression(parent) && parent.left === target && ts.isAssignmentOperator(parent.operatorToken.kind))
        || ((ts.isPrefixUnaryExpression(parent) || ts.isPostfixUnaryExpression(parent))
            && (parent.operator === ts.SyntaxKind.PlusPlusToken || parent.operator === ts.SyntaxKind.MinusMinusToken))
        || ts.isDeleteExpression(parent)
        || ((ts.isForInStatement(parent) || ts.isForOfStatement(parent)) && parent.initializer === target);
}

function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const plans = [];
    for (const file of new Set(sites.map(site => site.file))) {
        const fileName = path.join(tree, 'src/compiler/transformers', file);
        const text = fs.readFileSync(fileName, 'utf8');
        const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
        assert.equal(source.parseDiagnostics.length, 0, `cannot parse ${file}`);
        const edits = [];
        let asserted = 0;
        let expected = 0;
        for (const site of sites.filter(site => site.file === file)) {
            const functions = nodes(source, n => ts.isFunctionDeclaration(n) && n.name?.text === site.function);
            assert.equal(functions.length, 1, `${file}: ${site.function} moved`);
            const reads = nodes(functions[0], n => ts.isElementAccessExpression(n) && n.getText(source) === site.expression);
            assert.equal(reads.length, site.total, `${file}: ${site.function}'s ${site.expression} count drift`);
            expected += site.assert.length;
            for (const occurrence of site.assert) {
                const read = reads[occurrence - 1];
                assert(read && !isStore(read), `${file}: ${site.function}'s ${site.expression} #${occurrence} is a store`);
                if (ts.isNonNullExpression(read.parent)) asserted++;
                else edits.push({at: read.end, text: '!'});
            }
        }
        assert(asserted === 0 || asserted === expected, `${file} is partly adapted (${asserted} of ${expected})`);
        let result = text;
        for (const edit of edits.sort((a, b) => b.at - a.at)) result = result.slice(0, edit.at) + edit.text + result.slice(edit.at);
        assert.equal(ts.createSourceFile(fileName, result, ts.ScriptTarget.Latest, true).parseDiagnostics.length, 0, `cannot parse adapted ${file}`);
        plans.push({fileName, text, result, assertions: edits.length});
    }
    // Validate every file and every unchanged input before the first write.
    for (const plan of plans) assert.equal(fs.readFileSync(plan.fileName, 'utf8'), plan.text, 'source changed while planning');
    for (const plan of plans) if (plan.assertions) fs.writeFileSync(plan.fileName, plan.result);
    return {files: plans.filter(plan => plan.assertions).length, assertions: plans.reduce((sum, plan) => sum + plan.assertions, 0), javascriptChanging: false};
}
module.exports = {adapt, sites};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
