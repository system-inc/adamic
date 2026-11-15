'use strict';
const fs = require('node:fs'), path = require('node:path'), crypto = require('node:crypto'), assert = require('node:assert/strict'), ts = require('typescript');
const owners = require('./owners.json');
const digest = text => crypto.createHash('sha256').update(text).digest('hex');
function plan(tree, restore = false) {
    assert.equal(ts.version, '6.0.3');
    const files = [...new Set(owners.map(p => p.file))], plans = [], states = [];
    for (const file of files) {
        const name = path.join(tree, file), raw = fs.readFileSync(name, 'utf8');
        assert(!/\r(?!\n)/.test(raw), `${file}: bare CR`);
        assert(!(raw.includes('\r\n') && /(?<!\r)\n/.test(raw)), `${file}: mixed line endings`);
        const crlf = raw.includes('\r\n'), text = raw.replace(/\r\n/g, '\n');
        const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
        assert.equal(source.parseDiagnostics.length, 0, `${file}: parse diagnostics`);
        const edits = [];
        for (const pin of owners.filter(p => p.file === file)) {
            assert.equal(digest(pin.before), pin.ownerBeforeSha256, 'before pin corrupt');
            assert.equal(digest(pin.after), pin.ownerAfterSha256, 'after pin corrupt');
            const found = [];
            function visit(n) {
                if (ts.SyntaxKind[n.kind] === pin.kind && (ts.isImportDeclaration(n) ? n.moduleSpecifier.text : n.name?.text) === pin.owner) {
                    const ownerPath = []; for (let p = n; p; p = p.parent) if ((ts.isFunctionLike(p) || ts.isModuleDeclaration(p)) && p.name) ownerPath.unshift(p.name.getText(source));
                    if (!pin.ownerPath || JSON.stringify(ownerPath) === JSON.stringify(pin.ownerPath)) found.push(n);
                }
                ts.forEachChild(n, visit);
            }
            visit(source);
            assert.equal(found.length, pin.declarations || 1, `${file}:${pin.owner}: missing or duplicate owner`);
            found.sort((a, b) => a.pos - b.pos);
            const node = found[0], start = node.getStart(source), end = found.at(-1).end, current = text.slice(start, end), hash = digest(current);
            assert(hash === pin.ownerBeforeSha256 || hash === pin.ownerAfterSha256, `${file}:${pin.owner}: drift or half-adapted owner`);
            const adapted = hash === pin.ownerAfterSha256;
            assert.equal(current, adapted ? pin.after : pin.before);
            states.push(adapted);
            if (!pin.coveredBy && (restore ? adapted : !adapted)) edits.push({start, end, value: restore ? pin.before : pin.after});
        }
        let result = text;
        for (const edit of edits.sort((a, b) => b.start - a.start)) result = result.slice(0, edit.start) + edit.value + result.slice(edit.end);
        assert.equal(ts.createSourceFile(file, result, ts.ScriptTarget.Latest, true).parseDiagnostics.length, 0);
        if (crlf) result = result.replace(/\n/g, '\r\n');
        plans.push({name, raw, result, edits: edits.length});
    }
    assert(states.length && states.every(s => s === states[0]), 'half-adapted tree');
    return plans;
}
function adapt(tree, restore = false) {
    const plans = plan(tree, restore);
    for (const p of plans) assert.equal(fs.readFileSync(p.name, 'utf8'), p.raw, 'source changed while planning');
    for (const p of plans) if (p.result !== p.raw) fs.writeFileSync(p.name, p.result);
    return {files: plans.filter(p => p.result !== p.raw).length, writes: plans.filter(p => p.result !== p.raw).length, owners: plans.reduce((n, p) => n + p.edits, 0), javascriptChanging: true};
}
module.exports = {adapt, plan, owners};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs TREE');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
