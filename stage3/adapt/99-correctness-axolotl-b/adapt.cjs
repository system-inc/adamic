'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require('typescript');

// Both requested presence invariants are false. Validate the exact declined
// reads, including their unasserted state, and make no writes. See README.md.
const sites = [
    {function: 'parseOptionValue', expression: 'args[i]', kind: 'element', occurrence: 4, total: 8, parent: 'parseInt(args[i])'},
    {function: 'formatValueOrArray', expression: 'formatSingleValue(value, map)', kind: 'call', occurrence: 1, total: 1, parent: 'return formatSingleValue(value, map);'},
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
    const fileName = path.join(tree, 'src/compiler/commandLineParser.ts');
    const text = fs.readFileSync(fileName, 'utf8');
    const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, 'cannot parse commandLineParser.ts');
    for (const site of sites) {
        const functions = nodes(source, n => ts.isFunctionDeclaration(n) && n.name?.text === site.function);
        assert.equal(functions.length, 1, `${site.function} moved`);
        const reads = nodes(functions[0], n => (site.kind === 'element' ? ts.isElementAccessExpression(n) : ts.isCallExpression(n))
            && n.getText(source) === site.expression);
        assert.equal(reads.length, site.total, `${site.function}: ${site.expression} count drift`);
        const read = reads[site.occurrence - 1];
        assert(read && !isStore(read), `${site.function}: declined read became a store`);
        assert(!ts.isNonNullExpression(read.parent), `${site.function}: disproved read was asserted`);
        assert.equal(read.parent.getText(source), site.parent, `${site.function}: read context drift`);
    }
    // There are no valid edits to plan. Refuse any asserted or partly asserted
    // input above and validate the unchanged input before returning, without writes.
    assert.equal(fs.readFileSync(fileName, 'utf8'), text, 'source changed while validating');
    return {files: 0, assertions: 0, javascriptChanging: false};
}
module.exports = {adapt, sites};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
