#!/usr/bin/env node
'use strict';
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict'), crypto = require('node:crypto'), ts = require('typescript');
const site = require('./sites.json'), [control, actual] = process.argv.slice(2);
assert(control && actual); assert.equal(ts.version, '6.0.3');
const walk = directory => fs.readdirSync(directory, {withFileTypes: true}).flatMap(e => e.isDirectory() ? walk(path.join(directory, e.name)) : [path.join(directory, e.name)]).sort();
const files = tree => walk(path.join(tree, 'src')).filter(p => p.endsWith('.ts')).map(p => path.relative(tree, p));
assert.deepEqual(files(control), files(actual)); let changed = 0, proof;
for (const file of files(control)) {
    const before = fs.readFileSync(path.join(control, file), 'utf8'), after = fs.readFileSync(path.join(actual, file), 'utf8');
    if (file !== site.file) { assert.equal(after, before, file); continue; }
    const source = ts.createSourceFile(file, before, ts.ScriptTarget.Latest, true), found = [], statements = [];
    function visit(n) { if (ts.isFunctionDeclaration(n) && n.name?.text === 'getApplicableIndexSymbol') found.push(n); ts.forEachChild(n, visit); }
    visit(source); assert.equal(found.length, 1);
    assert.equal(crypto.createHash('sha256').update(found[0].getText(source)).digest('hex'), '35b9308f3edbb3ae332535bd198677d3cfba0044bd1588e327cb98d9b2a4eaeb');
    function locate(n) { if (ts.isExpressionStatement(n) && n.getText(source) === site.before) statements.push(n); ts.forEachChild(n, locate); }
    locate(found[0]); assert.equal(statements.length, 1); const old = statements[0];
    assert.equal(after === before.slice(0, old.getStart(source)) + site.after + before.slice(old.end), true, 'exact scoped fallback replacement required');
    assert.equal(ts.createSourceFile(file, after, ts.ScriptTarget.Latest, true).parseDiagnostics.length, 0);
    const crlf = text => (text.match(/\r\n/g) || []).length;
    assert.equal(crlf(after), crlf(before) + 10);
    const javascript = [false, true].map(removeComments => {
        const options = {fileName: file, compilerOptions: {target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments}};
        const old = ts.transpileModule(before, options).outputText, emitted = ts.transpileModule(after, options).outputText;
        assert.notEqual(emitted, old); return {removeComments, beforeBytes: Buffer.byteLength(old), afterBytes: Buffer.byteLength(emitted)};
    });
    proof = {javascriptChanging: true, crlfBefore: crlf(before), crlfAfter: crlf(after), javascript}; changed++;
}
assert.equal(changed, 1); console.log(JSON.stringify({sourceFiles: files(control).length, changedFiles: changed, ...proof}, null, 2));
