#!/usr/bin/env node
'use strict';
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict');
const crypto = require('node:crypto'), ts = require('typescript');
const [beforeTree, afterTree] = process.argv.slice(2);
assert(beforeTree && afterTree, 'usage: node verify.cjs BEFORE AFTER');
assert.equal(ts.version, '6.0.3');
const hash = text => crypto.createHash('sha256').update(text).digest('hex');
const walk = dir => fs.readdirSync(dir, {withFileTypes: true}).flatMap(e => e.isDirectory() ? walk(path.join(dir, e.name)) : [path.join(dir, e.name)]).sort();
const files = tree => walk(path.join(tree, 'src')).filter(p => p.endsWith('.ts')).map(p => path.relative(tree, p));
assert.deepEqual(files(beforeTree), files(afterTree));
let javascript = [], crlf, changed = 0;
for (const file of files(beforeTree)) {
    const before = fs.readFileSync(path.join(beforeTree, file), 'utf8'), after = fs.readFileSync(path.join(afterTree, file), 'utf8');
    if (file !== 'src/compiler/checker.ts') { assert.equal(after, before, 'other source changed: ' + file); continue; }
    const source = ts.createSourceFile(file, before, ts.ScriptTarget.Latest, true), owners = [];
    function visit(n) {
        if (ts.isFunctionDeclaration(n) && n.name?.text === 'getFlowTypeFromCommonJSExport') owners.push(n);
        ts.forEachChild(n, visit);
    }
    visit(source); assert.equal(owners.length, 1);
    const owner = owners[0];
    assert.equal(hash(owner.getText(source)), '57faba456412535bb02760aad9cdbcf7569cc83624729da3f6ddf40b6ac5a80c');
    const initializer = owner.body.statements[0].declarationList.declarations[0].initializer;
    assert(ts.isCallExpression(initializer));
    assert.equal(initializer.arguments[0].getText(source), 'symbol.declarations![0]');
    const at = initializer.arguments[0].end;
    assert.equal(after, before.slice(0, at) + '!' + before.slice(at), 'exact required declaration read missing or uninventoried edit');
    assert.equal(ts.createSourceFile(file, after, ts.ScriptTarget.Latest, true).parseDiagnostics.length, 0);
    crlf = {before: (before.match(/\r\n/g) || []).length, after: (after.match(/\r\n/g) || []).length};
    assert.equal(crlf.before, crlf.after);
    for (const removeComments of [false, true]) {
        const options = {fileName: file, compilerOptions: {target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments}};
        const expected = ts.transpileModule(before, options).outputText, emitted = ts.transpileModule(after, options).outputText;
        assert.equal(emitted, expected, 'stock emitted JavaScript bytes');
        javascript.push({removeComments, bytes: Buffer.byteLength(emitted), sha256: hash(emitted)});
    }
    changed += Number(before !== after);
}
assert.equal(changed, 1);
console.log(JSON.stringify({sourceFiles: files(beforeTree).length, changedFiles: changed, assertions: 1, crlf, javascript, publicDeclarationsUnchanged: true}, null, 2));
