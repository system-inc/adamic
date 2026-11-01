'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const crypto = require('node:crypto');
const ts = require('typescript');
const {compiler, nodes} = require('../47-host-errors/adapt.cjs');
const manifest = require('./proofs/targets.json');
const targets = [
    {local: '_fs', module: 'fs', namespace: 'adamicNodeFs93', declaration: 'const _fs: typeof import("fs")'},
    {local: '_path', module: 'path', namespace: 'adamicNodePath93', declaration: 'const _path: typeof import("path")'},
    {local: '_os', module: 'os', namespace: 'adamicNodeOs93', declaration: 'const _os'},
];
function hash(text) { return crypto.createHash('sha256').update(text).digest('hex'); }
function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const program = compiler(tree, ts);
    const file = program.getSourceFile(path.join(tree, manifest[0].file));
    assert(file && !file.parseDiagnostics.length, 'cannot parse sys.ts');
    const digest = hash(file.text);
    assert([manifest[0].beforeSha256, manifest[0].afterSha256].includes(digest), 'target source moved: sys.ts');
    const done = digest === manifest[0].afterSha256;
    const functions = nodes(ts, file, n => ts.isFunctionDeclaration(n) && n.name?.text === 'getNodeSystem');
    assert.equal(functions.length, 1, 'getNodeSystem moved');
    const imports = file.statements.filter(ts.isImportDeclaration);
    const edits = [], evidence = [];
    for (const target of targets) {
        const statements = functions[0].body.statements.filter(ts.isVariableStatement).filter(s => s.declarationList.declarations.some(d => d.name.getText(file) === target.local));
        assert.equal(statements.length, 1, 'builtin local is not directly in getNodeSystem');
        const list = statements[0].declarationList;
        assert.equal(list.declarations.length, 1, 'builtin declaration list moved');
        assert.equal(list.flags & ts.NodeFlags.Const, ts.NodeFlags.Const);
        const declaration = list.declarations[0];
        assert.equal('const ' + declaration.getText(file).slice(0, declaration.initializer.getStart(file) - declaration.getStart(file)).trim(), target.declaration + ' =');
        const expression = declaration.initializer;
        assert.equal(expression.getText(file), done ? target.namespace : `require("${target.module}")`, 'builtin initializer moved');
        const existing = imports.filter(n => n.moduleSpecifier.text === `node:${target.module}`);
        if (done) {
            assert.equal(existing.length, 1, 'builtin import moved');
            assert.equal(existing[0].getText(file), `import * as ${target.namespace} from "node:${target.module}";`);
        } else {
            assert.equal(existing.length, 0, 'builtin already imported');
            assert.equal(nodes(ts, file, n => ts.isIdentifier(n) && n.text === target.namespace).length, 0, 'namespace name collides');
            edits.push({start: expression.getStart(file), end: expression.end, text: target.namespace});
        }
        evidence.push({local: target.local, module: `node:${target.module}`, namespace: target.namespace, line: file.getLineAndCharacterOfPosition(expression.getStart(file)).line + 1});
    }
    if (!done) {
        assert(imports.length, 'existing imports moved');
        const at = imports[imports.length - 1].end;
        const newline = file.text.includes('\r\n') ? '\r\n' : '\n';
        edits.push({start: at, end: at, text: targets.map(t => `${newline}import * as ${t.namespace} from "node:${t.module}";`).join('')});
    }
    let text = file.text;
    for (const edit of edits.sort((a,b) => b.start-a.start)) text = text.slice(0,edit.start)+edit.text+text.slice(edit.end);
    assert.equal(hash(text), manifest[0].afterSha256, 'unreviewed builtin import change');
    assert.equal(ts.createSourceFile(file.fileName,text,ts.ScriptTarget.Latest,true).parseDiagnostics.length,0);
    // Validate every site, collisions and the full planned bytes before the sole write.
    assert.equal(fs.readFileSync(file.fileName,'utf8'),file.text,'source changed while planning');
    if (edits.length) fs.writeFileSync(file.fileName,text);
    return {files: edits.length ? 1 : 0, javascriptChanging: true, sites: evidence};
}
module.exports = {adapt, targets};
if (require.main === module) {
    assert.equal(process.argv.length,3,'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])),null,2));
}
