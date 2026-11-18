'use strict';
// #rv75m9k, road A: each @internal method signature tsc reads as a value becomes a readonly function-typed property of
// the same type, so the read is a field read and not "a method read as a value". Type-only: no emitted JavaScript
// changes, and every member is @internal, so typescript.d.ts doesn't move. owners.json pins each member whole, before
// and after, by its interface; the adapter refuses drift, a duplicate or missing member, a changed overload group, a
// half-adapted tree and mixed endings before writing anything, and a second run writes nothing. mutant.cjs restores every member.
const fs = require('node:fs'), path = require('node:path'), crypto = require('node:crypto'), assert = require('node:assert/strict'), ts = require('typescript');
const owners = require('./owners.json');
const digest = text => crypto.createHash('sha256').update(text).digest('hex');
function plan(tree, restore = false) {
    assert.equal(ts.version, '6.0.3');
    const files = [...new Set(owners.map(pin => pin.file))], plans = [], states = [];
    for (const file of files) {
        const name = path.join(tree, file), raw = fs.readFileSync(name, 'utf8');
        assert(!/\r(?!\n)/.test(raw), `${file}: bare CR`);
        assert(!(raw.includes('\r\n') && /(?<!\r)\n/.test(raw)), `${file}: mixed line endings`);
        const crlf = raw.includes('\r\n'), text = raw.replace(/\r\n/g, '\n');
        const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
        assert.equal(source.parseDiagnostics.length, 0, `${file}: parse diagnostics`);
        const edits = [];
        for (const pin of owners.filter(candidate => candidate.file === file)) {
            assert.equal(digest(pin.before), pin.beforeSha256, `${pin.interface}.${pin.member}: before pin corrupt`);
            assert.equal(digest(pin.after), pin.afterSha256, `${pin.interface}.${pin.member}: after pin corrupt`);
            // Interfaces merge: the member lives in exactly one declaration of its interface in this file.
            const named = member => member.name && ts.isIdentifier(member.name) && member.name.text === pin.member;
            const holders = [];
            ts.forEachChild(source, function find(node) {
                if (ts.isInterfaceDeclaration(node) && node.name.text === pin.interface && node.members.some(named)) holders.push(node);
                ts.forEachChild(node, find);
            });
            assert.equal(holders.length, 1, `${file}: expected one declaration of ${pin.interface} holding ${pin.member}, found ${holders.length}`);
            // Every member of that name, adjacent: before, an overload group of pin.overloads method signatures; after, one property.
            const members = holders[0].members.filter(named);
            const adapted = members.length === 1 && ts.isPropertySignature(members[0]);
            assert.equal(members.length, adapted ? 1 : pin.overloads, `${pin.interface}.${pin.member}: expected ${pin.overloads} member(s), found ${members.length} (missing, duplicate or a changed overload group)`);
            assert(adapted || members.every(member => ts.isMethodSignature(member)), `${pin.interface}.${pin.member}: unexpected member kind`);
            const start = members[0].getStart(source), end = members.at(-1).end, current = text.slice(start, end), hash = digest(current);
            assert(hash === (adapted ? pin.afterSha256 : pin.beforeSha256), `${pin.interface}.${pin.member}: drift or half-adapted member`);
            assert.equal(current, adapted ? pin.after : pin.before);
            states.push(adapted);
            if (restore ? adapted : !adapted) edits.push({ start, end, value: restore ? pin.before : pin.after });
        }
        let result = text;
        for (const edit of edits.sort((a, b) => b.start - a.start)) result = result.slice(0, edit.start) + edit.value + result.slice(edit.end);
        assert.equal(ts.createSourceFile(file, result, ts.ScriptTarget.Latest, true).parseDiagnostics.length, 0, `${file}: adapted text doesn't parse`);
        if (crlf) result = result.replace(/\n/g, '\r\n');
        plans.push({ name, raw, result, edits: edits.length });
    }
    assert(states.length && states.every(state => state === states[0]), 'half-adapted tree');
    return plans;
}
function adapt(tree, restore = false) {
    const plans = plan(tree, restore);
    for (const filePlan of plans) assert.equal(fs.readFileSync(filePlan.name, 'utf8'), filePlan.raw, 'source changed while planning');
    for (const filePlan of plans) if (filePlan.result !== filePlan.raw) fs.writeFileSync(filePlan.name, filePlan.result);
    const written = plans.filter(filePlan => filePlan.result !== filePlan.raw).length;
    return { task: '#rv75m9k', files: written, members: plans.reduce((sum, filePlan) => sum + filePlan.edits, 0), javascriptChanging: false };
}
module.exports = { adapt, plan, owners };
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs TREE');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
