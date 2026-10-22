'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require('typescript');
const crypto = require('node:crypto');
const manifest = require('./proofs/targets.json');
const {compiler, nodes} = require('../47-host-errors/adapt.cjs');
const targets = [
    {file: 'src/compiler/factory/nodeFactory.ts', owner: 'createNodeFactory', registry: 'nodeFactoryPatchers', argument: 'factory', type: 'NodeFactory'},
    {file: 'src/compiler/utilities.ts', owner: 'setObjectAllocator', registry: 'objectAllocatorPatchers', argument: 'objectAllocator', type: 'ObjectAllocator'},
];
function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const program = compiler(tree, ts), checker = program.getTypeChecker();
    const outputs = [], evidence = [];
    for (const target of targets) {
        const file = program.getSourceFile(path.join(tree, target.file));
        assert(file && !file.parseDiagnostics.length, target.file);
        const pin = manifest.find(row => row.file === target.file);
        const digest = crypto.createHash('sha256').update(file.text).digest('hex');
        assert([pin.beforeSha256, pin.afterSha256].includes(digest), 'target source moved: ' + target.file);
        const owners = nodes(ts, file, n => ts.isFunctionDeclaration(n) && n.name?.text === target.owner);
        assert.equal(owners.length, 1, 'target owner moved');
        const calls = nodes(ts, owners[0], n => ts.isCallExpression(n) && n.expression.getText(file) === 'forEach' && n.arguments[0]?.getText(file) === target.registry);
        assert.equal(calls.length, 1, 'patcher call moved');
        const call = calls[0], arrow = call.arguments[1];
        assert.equal(call.arguments.length, 2);
        assert(ts.isExpressionStatement(call.parent), 'patcher call context moved');
        assert(ts.isArrowFunction(arrow) && arrow.parameters.length === 1);
        assert.equal(arrow.parameters[0].getText(file), 'fn');
        const before = `fn(${target.argument})`, after = `{ ${before}; }`;
        assert([before, after].includes(arrow.body.getText(file)), 'patcher arrow body moved');
        const inner = ts.isBlock(arrow.body) ? arrow.body.statements[0]?.expression : arrow.body;
        assert(inner && ts.isCallExpression(inner));
        assert.equal(checker.getReturnTypeOfSignature(checker.getResolvedSignature(inner)).flags, ts.TypeFlags.Void, 'patcher return is no longer exactly void');
        const signature = checker.getResolvedSignature(inner);
        assert.equal(signature.parameters.length, 1);
        assert.equal(checker.typeToString(checker.getTypeOfSymbolAtLocation(signature.parameters[0], inner)), target.type);
        const registry = checker.getSymbolAtLocation(call.arguments[0]);
        assert.equal(registry?.declarations?.length, 1);
        assert.equal(registry.declarations[0].getSourceFile(), file);
        const forEach = checker.getResolvedSignature(call)?.declaration;
        assert(forEach && forEach.getSourceFile().fileName.endsWith('/src/compiler/core.ts'), 'forEach resolution moved');
        const location = file.getLineAndCharacterOfPosition(call.getStart(file));
        evidence.push({...target, line: location.line + 1, before: `forEach(${target.registry}, fn => ${before});`, after: `forEach(${target.registry}, fn => ${after});`});
        if (arrow.body.getText(file) === after) continue;
        const text = file.text.slice(0, arrow.body.getStart(file)) + after + file.text.slice(arrow.body.end);
        assert.equal(ts.createSourceFile(file.fileName, text, ts.ScriptTarget.Latest, true).parseDiagnostics.length, 0);
        outputs.push([file, text]);
    }
    // Validate both targets before writing either file.
    for (const [file] of outputs) assert.equal(fs.readFileSync(file.fileName, 'utf8'), file.text, 'source changed while planning');
    for (const [file, text] of outputs) fs.writeFileSync(file.fileName, text);
    return {files: outputs.length, javascriptChanging: true, sites: evidence};
}
module.exports = {adapt, targets};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
