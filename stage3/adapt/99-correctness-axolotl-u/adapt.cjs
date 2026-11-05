'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const crypto = require('node:crypto');
const ts = require('typescript');
const pins = require('./pins.json');
const imports = require('./imports.json');
const normalize = text => text.replace(/\r\n/g, '\n');
const hash = text => crypto.createHash('sha256').update(normalize(text)).digest('hex');
function parse(text, file) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, `parse error: ${file}`);
    return source;
}
function owners(text, file) {
    const source = parse(text, file);
    const found = new Map();
    function visit(node, ancestry) {
        const scoped = (ts.isModuleDeclaration(node) || ts.isClassDeclaration(node) || ts.isFunctionDeclaration(node)) && node.name;
        const chain = scoped ? [...ancestry, node.name.getText(source)] : ancestry;
        if (ts.isFunctionDeclaration(node) && node.body && node.name) {
            const key = chain.join('.');
            assert(!found.has(key), `duplicate owner: ${key}`);
            found.set(key, {start: node.getStart(source), end: node.end, text: node.getText(source)});
        }
        ts.forEachChild(node, child => visit(child, chain));
    }
    visit(source, []);
    return found;
}
function importState(text) {
    const source = parse(text, imports.file);
    const matches = source.statements.filter(node => ts.isImportDeclaration(node) && node.moduleSpecifier.text === imports.module);
    assert.equal(matches.length, 1, 'missing or duplicate guard import');
    const clause = matches[0].importClause;
    assert(clause && !clause.isTypeOnly && clause.namedBindings && ts.isNamedImports(clause.namedBindings), 'changed guard import');
    const elements = clause.namedBindings.elements;
    const present = imports.names.map(name => {
        const rows = elements.filter(element => element.name.text === name || element.propertyName?.text === name);
        assert(rows.length <= 1, `duplicate guard import: ${name}`);
        if (rows.length) assert(!rows[0].isTypeOnly && !rows[0].propertyName && rows[0].name.text === name, `changed guard import: ${name}`);
        return rows.length === 1;
    });
    assert(present.every(Boolean) || present.every(value => !value), 'half-adapted guard imports');
    const anchor = elements.filter(element => element.name.text === imports.anchor);
    assert.equal(anchor.length, 1, 'missing guard import anchor');
    assert(!anchor[0].propertyName && !anchor[0].isTypeOnly, 'changed guard import anchor');
    const start = anchor[0].getStart(source);
    const lineStart = text.lastIndexOf('\n', start) + 1;
    assert.equal(text.slice(lineStart, start), '    ', 'changed guard import anchor indentation');
    // A second declaration must not shadow a newly introduced guard binding.
    for (const statement of source.statements) {
        if (statement === matches[0]) continue;
        function visit(node) {
            if ((ts.isImportSpecifier(node) || ts.isVariableDeclaration(node) || ts.isFunctionDeclaration(node) || ts.isClassDeclaration(node)) && node.name && ts.isIdentifier(node.name)) {
                assert(!imports.names.includes(node.name.text), `conflicting guard binding: ${node.name.text}`);
            }
            ts.forEachChild(node, visit);
        }
        visit(statement);
    }
    return {state: present[0] ? 'after' : 'before', lineStart};
}
function inspect(texts) {
    const parsed = new Map([...texts].map(([file, text]) => [file, owners(text, file)]));
    const sites = pins.map(pin => {
        const owner = parsed.get(pin.file).get(pin.owner);
        assert(owner, `missing owner: ${pin.owner}`);
        const digest = hash(owner.text);
        const state = digest === pin.ownerBeforeSha256 ? 'before' : digest === pin.ownerAfterSha256 ? 'after' : undefined;
        assert(state, `changed owner: ${pin.owner}`);
        assert.equal(normalize(owner.text), pin[state], `owner text mismatch: ${pin.owner}`);
        return {pin, owner, state};
    });
    assert(sites.every(site => site.state === sites[0].state), 'half-adapted tree');
    assert.equal(importState(texts.get(imports.file)).state, sites[0].state, 'owner/import state mismatch');
    return sites;
}
function plan(texts) {
    assert.equal(ts.version, '6.0.3');
    const sites = inspect(texts);
    if (sites[0].state === 'after') return {texts, edits: 0};
    const after = new Map(texts);
    for (const [file, before] of texts) {
        let text = before;
        for (const {pin, owner} of sites.filter(site => site.pin.file === file).sort((a,b) => b.owner.start - a.owner.start)) {
            for (const exact of pin.exactSites) assert.equal(normalize(owner.text).split(exact).length - 1, 1, `exact site missing: ${pin.owner}`);
            const eol = owner.text.includes('\r\n') ? '\r\n' : '\n';
            text = text.slice(0, owner.start) + pin.after.replace(/\n/g, eol) + text.slice(owner.end);
        }
        if (file === imports.file) {
            const {lineStart} = importState(text);
            const eol = text.includes('\r\n') ? '\r\n' : '\n';
            text = text.slice(0, lineStart) + imports.names.map(name => `    ${name},${eol}`).join('') + text.slice(lineStart);
        }
        after.set(file, text);
    }
    assert(inspect(after).every(site => site.state === 'after'), 'incomplete planned tree');
    return {texts: after, edits: pins.length + imports.names.length};
}
function adapt(tree) {
    const files = [...new Set(pins.map(pin => pin.file))];
    const before = new Map(files.map(file => [file, fs.readFileSync(path.join(tree, 'src/compiler', file), 'utf8')]));
    const result = plan(before);
    // All owners, imports and planned files are validated before any write.
    for (const [file, text] of before) assert.equal(fs.readFileSync(path.join(tree, 'src/compiler', file), 'utf8'), text, `concurrent source change: ${file}`);
    let filesChanged = 0;
    for (const [file, text] of result.texts) if (text !== before.get(file)) {
        fs.writeFileSync(path.join(tree, 'src/compiler', file), text);
        filesChanged++;
    }
    return {files: filesChanged, edits: result.edits, javascriptChanging: true};
}
module.exports = {adapt, plan, inspect, owners, hash, pins, imports};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs TREE');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
