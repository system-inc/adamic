'use strict';
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const assert = require('node:assert/strict');
const ts = require('typescript');
const owners = require('./owners.json');

const digest = text => crypto.createHash('sha256').update(text).digest('hex');

function locate(file, text, pins) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, `${file}: parse error`);
    return pins.map(pin => {
        const matches = source.statements.filter(node => {
            if (node.kind !== ts.SyntaxKind[pin.kind]) return false;
            if (ts.isFunctionDeclaration(node)) return node.name?.text === pin.owner;
            if (ts.isImportDeclaration(node)) return node.moduleSpecifier.text === pin.owner;
            return ts.isVariableStatement(node) && node.declarationList.declarations.some(
                declaration => ts.isIdentifier(declaration.name) && declaration.name.text === pin.owner);
        });
        assert.equal(matches.length, 1, `${file}: ${pin.owner}: owner count drift`);
        const start = matches[0].getStart(source, true);
        const end = matches[0].end;
        const spelling = text.slice(start, end);
        const hash = digest(spelling);
        const adapted = hash === pin.ownerAfterSha256;
        assert(adapted || hash === pin.ownerBeforeSha256, `${file}: ${pin.owner}: changed owner`);
        assert.equal(spelling, adapted ? pin.after : pin.before, `${file}: ${pin.owner}: exact text drift`);
        return {pin, start, end, adapted};
    });
}

function transpile(text, removeComments) {
    return ts.transpileModule(text, {compilerOptions: {
        target: ts.ScriptTarget.ES2024,
        module: ts.ModuleKind.ESNext,
        removeComments,
    }}).outputText;
}

function rewrite(tree, restoring = false) {
    assert.equal(ts.version, '6.0.3');
    const wanted = !restoring;
    const plans = [];
    let inputState;
    for (const [file, pins] of Object.entries(Object.groupBy(owners, pin => pin.file))) {
        const name = path.join(tree, 'src/compiler', file);
        const raw = fs.readFileSync(name, 'utf8');
        const crlf = raw.includes('\r\n');
        const text = raw.replace(/\r\n/g, '\n');
        assert(!text.includes('\r') && (!crlf || !raw.replace(/\r\n/g, '').includes('\n')), `${file}: mixed line endings`);
        const states = locate(file, text, pins).sort((a, b) => b.start - a.start);
        inputState ??= states[0].adapted;
        assert(states.every(state => state.adapted === inputState), `${file}: half-adapted owners`);
        if (inputState === wanted) continue;
        let edited = text;
        for (const {pin, start, end} of states) {
            edited = edited.slice(0, start) + (wanted ? pin.after : pin.before) + edited.slice(end);
        }
        assert(locate(file, edited, pins).every(state => state.adapted === wanted), `${file}: invalid planned result`);
        for (const removeComments of [false, true]) {
            assert.equal(transpile(edited, removeComments), transpile(text, removeComments), `${file}: JavaScript changed`);
        }
        plans.push({name, raw, after: crlf ? edited.replace(/\n/g, '\r\n') : edited});
    }
    for (const {name, raw} of plans) {
        assert.equal(fs.readFileSync(name, 'utf8'), raw, `${name}: source changed while planning`);
    }
    for (const {name, after} of plans) fs.writeFileSync(name, after);
    return {files: plans.length, edits: plans.length ? owners.length : 0, javascriptChanging: false};
}
module.exports = {rewrite};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs TREE');
    console.log(JSON.stringify(rewrite(path.resolve(process.argv[2])), null, 2));
}
