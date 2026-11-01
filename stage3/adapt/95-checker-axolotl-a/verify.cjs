#!/usr/bin/env node
'use strict';
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict');
const crypto = require('node:crypto'), ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
const [beforeTree, afterTree] = process.argv.slice(2);
assert(beforeTree && afterTree, 'usage: node verify.cjs BEFORE AFTER');
assert.equal(ts.version, '6.0.3');
const hash = text => crypto.createHash('sha256').update(text).digest('hex');
const walk = directory => fs.readdirSync(directory, {withFileTypes: true}).flatMap(entry => entry.isDirectory() ?
    walk(path.join(directory, entry.name)) : [path.join(directory, entry.name)]).sort();
const files = tree => walk(path.join(tree, 'src')).filter(file => file.endsWith('.ts')).map(file => path.relative(tree, file));
assert.deepEqual(files(beforeTree), files(afterTree));
let changed = 0, javascript = [], crlf;
for (const file of files(beforeTree)) {
    const before = fs.readFileSync(path.join(beforeTree, file), 'utf8'), after = fs.readFileSync(path.join(afterTree, file), 'utf8');
    if (file !== 'src/compiler/checker.ts') { assert.equal(after, before, 'other source changed: ' + file); continue; }
    const parseOwner = text => {
        const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true), found = [];
        assert.equal(source.parseDiagnostics.length, 0);
        function visit(node) {
            if (ts.isFunctionDeclaration(node) && node.name?.text === 'checkGrammarIndexSignatureParameters') found.push(node);
            ts.forEachChild(node, visit);
        }
        visit(source); assert.equal(found.length, 1);
        return {source, node: found[0]};
    };
    const old = parseOwner(before), actual = parseOwner(after);
    const oldText = old.node.getText(old.source), actualText = actual.node.getText(actual.source);
    for (const index of [0, 1]) assert.equal(actual.node.body.statements[index].getText(actual.source),
        old.node.body.statements[index].getText(old.source), 'initializer and invalid-arity handling must be unchanged');
    assert.equal((oldText.match(/parameter!/g) || []).length, 0);
    assert.equal((actualText.match(/parameter!/g) || []).length, 13, 'all 13 required read checks must be present');
    assert.equal(actualText.replaceAll('parameter!', 'parameter'), oldText, 'exact private function edit');
    assert.equal(after.slice(0, actual.node.getStart(actual.source)), before.slice(0, old.node.getStart(old.source)), 'prefix');
    assert.equal(after.slice(actual.node.end), before.slice(old.node.end), 'suffix');
    crlf = {before: (before.match(/\r\n/g) || []).length, after: (after.match(/\r\n/g) || []).length};
    assert.equal(crlf.after, crlf.before, 'CRLF count');
    for (const removeComments of [false, true]) {
        const options = {fileName: file, compilerOptions: {target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments}};
        const expected = ts.transpileModule(before, options).outputText, emitted = ts.transpileModule(after, options).outputText;
        assert.equal(emitted, expected, 'stock emitted JavaScript bytes');
        javascript.push({removeComments, bytes: Buffer.byteLength(emitted), sha256: hash(emitted)});
    }
    changed += Number(before !== after);
}
assert.equal(changed, 1);
console.log(JSON.stringify({sourceFiles: files(beforeTree).length, changedFiles: changed, assertions: 13, crlf, javascript, publicDeclarationsUnchanged: true}, null, 2));
