'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const crypto = require('node:crypto');
const ts = require('typescript');
const pins = require('./pins.json');
const hash = text => crypto.createHash('sha256').update(text).digest('hex');
function owners(text, file) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, `parse error: ${file}`);
    const found = new Map();
    function visit(node, ancestry) {
        const method = ts.isMethodDeclaration(node) && ts.isClassDeclaration(node.parent);
        const named = (ts.isModuleDeclaration(node) || ts.isClassDeclaration(node) || ts.isFunctionDeclaration(node) || method) && node.name;
        const chain = named ? [...ancestry, node.name.getText(source)] : ancestry;
        if ((ts.isFunctionDeclaration(node) || method) && node.body && node.name) {
            const key = chain.join('.');
            assert(!found.has(key), `duplicate owner: ${key}`);
            found.set(key, {start: node.getStart(source), end: node.end, text: node.getText(source)});
        }
        ts.forEachChild(node, child => visit(child, chain));
    }
    visit(source, []);
    return found;
}
function inspect(texts) {
    const parsed = new Map([...texts].map(([file, text]) => [file, owners(text, file)]));
    return pins.map(pin => {
        const owner = parsed.get(pin.file).get(pin.owner);
        assert(owner, `missing owner: ${pin.owner}`);
        const digest = hash(owner.text);
        const state = digest === pin.ownerBeforeSha256 ? 'before' : digest === pin.ownerAfterSha256 ? 'after' : undefined;
        assert(state, `changed owner: ${pin.owner}`);
        const expected = state === 'before' ? pin.before : pin.after;
        assert.equal(owner.text, expected, `owner text mismatch: ${pin.owner}`);
        return {pin, owner, state};
    });
}
function plan(texts) {
    assert.equal(ts.version, '6.0.3');
    const sites = inspect(texts);
    assert(sites.every(site => site.state === sites[0].state), 'half-adapted tree');
    if (sites[0].state === 'after') return {texts, edits: 0};
    const after = new Map(texts);
    for (const file of texts.keys()) {
        let text = texts.get(file);
        for (const {pin, owner} of sites.filter(site => site.pin.file === file).sort((a,b) => b.owner.start - a.owner.start)) {
            for (const exact of pin.exactSites) assert.equal(owner.text.split(exact).length - 1, 1, `exact site missing: ${pin.owner}`);
            text = text.slice(0, owner.start) + pin.after + text.slice(owner.end);
        }
        after.set(file, text);
    }
    assert(inspect(after).every(site => site.state === 'after'), 'incomplete plan');
    return {texts: after, edits: pins.reduce((n,p) => n + p.exactSites.length, 0)};
}
function adapt(tree) {
    const files = [...new Set(pins.map(pin => pin.file))];
    const before = new Map(files.map(file => [file, fs.readFileSync(path.join(tree, 'src/compiler', file), 'utf8')]));
    const result = plan(before);
    // Validate every owner and reread every input before the first write.
    for (const [file, text] of before) assert.equal(fs.readFileSync(path.join(tree, 'src/compiler', file), 'utf8'), text, `concurrent source change: ${file}`);
    let changed = 0;
    for (const [file, text] of result.texts) if (text !== before.get(file)) {
        fs.writeFileSync(path.join(tree, 'src/compiler', file), text);
        changed++;
    }
    return {files: changed, edits: result.edits, javascriptChanging: true};
}
module.exports = {adapt, plan, inspect, owners, hash, pins};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs TREE');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
