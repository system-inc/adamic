'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require('typescript');

// Make parseInt's existing ToString conversion explicit. The other requested
// read remains declined and must validate before any write. See README.md.
const sites = [
    {function: 'parseOptionValue', expression: 'args[i]', kind: 'element', occurrence: 4, total: 8, parent: 'parseInt(args[i])', adaptedParent: 'String(args[i])'},
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

function locate(fileName, text) {
    const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, 'cannot parse commandLineParser.ts');
    let edit;
    for (const site of sites) {
        const functions = nodes(source, n => ts.isFunctionDeclaration(n) && n.name?.text === site.function);
        assert.equal(functions.length, 1, `${site.function} moved`);
        const reads = nodes(functions[0], n => (site.kind === 'element' ? ts.isElementAccessExpression(n) : ts.isCallExpression(n))
            && n.getText(source) === site.expression);
        assert.equal(reads.length, site.total, `${site.function}: ${site.expression} count drift`);
        const read = reads[site.occurrence - 1];
        assert(read && !isStore(read), `${site.function}: declined read became a store`);
        assert(!ts.isNonNullExpression(read.parent), `${site.function}: disproved read was asserted`);
        const context = read.parent.getText(source);
        if (site.adaptedParent) {
            const adapted = context === site.adaptedParent;
            assert(context === site.parent || adapted, `${site.function}: read context drift`);
            const call = adapted ? read.parent.parent : read.parent;
            assert(ts.isCallExpression(call), 'numeric read left its call');
            assert.equal(call.getText(source), adapted ? 'parseInt(String(args[i]))' : 'parseInt(args[i])');
            const statement = call.parent.parent.parent;
            assert.equal(statement.getText(source), adapted
                ? 'options[opt.name] = validateJsonOptionValue(opt, parseInt(String(args[i])), errors);'
                : 'options[opt.name] = validateJsonOptionValue(opt, parseInt(args[i]), errors);');
            const calls = nodes(functions[0], n => ts.isCallExpression(n) && ts.isIdentifier(n.expression) && n.expression.text === 'parseInt');
            assert.equal(calls.length, 1, 'numeric parseInt count drift');
            if (!adapted) edit = {start: read.getStart(source), end: read.end, text: 'String(args[i])'};
        }
        else assert.equal(context, site.parent, `${site.function}: read context drift`);
    }
    return edit;
}

function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const fileName = path.join(tree, 'src/compiler/commandLineParser.ts');
    const text = fs.readFileSync(fileName, 'utf8');
    const edit = locate(fileName, text);
    const result = edit ? text.slice(0, edit.start) + edit.text + text.slice(edit.end) : text;
    assert.equal(locate(fileName, result), undefined, 'planned result is not fully adapted');
    assert.equal(result.length - text.length, edit ? 'String()'.length : 0, 'unplanned source changes');
    assert.equal(fs.readFileSync(fileName, 'utf8'), text, 'source changed while planning');
    if (edit) fs.writeFileSync(fileName, result);
    return {files: edit ? 1 : 0, edits: edit ? 1 : 0, javascriptChanging: true, behaviorChanging: false};
}
module.exports = {adapt, sites};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
