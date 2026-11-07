'use strict';
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict');
const ts = require('typescript');
const pins = require('./owners.json');
function parse(file, text) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, file + ': parse error');
    return source;
}
function nodes(root, predicate) {
    const found = [];
    function visit(node) { if (predicate(node)) found.push(node); ts.forEachChild(node, visit); }
    visit(root); return found;
}
function one(found, label) { assert.equal(found.length, 1, label + ': missing or duplicate'); return found[0]; }
function ancestry(node) {
    const names = [];
    for (let p = node.parent; p; p = p.parent) {
        if (ts.isConstructorDeclaration(p)) names.unshift('constructor');
        else if ((ts.isFunctionLike(p) || ts.isClassDeclaration(p)) && p.name) names.unshift(p.name.getText());
    }
    return names;
}
function inspect(core, harness) {
    const c = parse(pins.core.file, core), h = parse(pins.harness.file, harness);
    const declarations = nodes(c, n => ts.isFunctionDeclaration(n) && n.name?.text === 'getAllKeys');
    assert(declarations.length <= 1, 'duplicate core getAllKeys');
    const declaration = declarations[0];
    // A removed declaration must not leave its internal JSDoc on the next owner.
    const next = one(c.statements.filter(n => ts.isFunctionDeclaration(n) && n.name?.text === 'getOwnValues'), 'getOwnValues boundary');
    assert.equal(next.jsDoc?.length, 1, 'orphaned getAllKeys comment');
    assert.equal(core.slice(next.jsDoc[0].pos, next.jsDoc[0].end), '/** @internal */', 'changed following comment boundary');
    if (declaration) {
        assert.equal(declaration.parent, c, 'core getAllKeys must be module level');
        assert.equal(declaration.jsDoc?.length, 1, 'getAllKeys internal comment moved');
        assert.equal(core.slice(declaration.jsDoc[0].pos, declaration.end), pins.core.before, 'changed core owner');
    }
    const statement = one(nodes(h, n => ts.isVariableStatement(n) && n.declarationList.declarations.some(d => ts.isIdentifier(d.name) && d.name.text === 'keys') && JSON.stringify(ancestry(n)) === JSON.stringify(pins.harness.path)), 'proxy keys statement');
    const spelling = statement.getText(h);
    assert([pins.harness.before, pins.harness.after].includes(spelling), 'changed harness statement owner');
    const locals = nodes(h, n => ts.isFunctionDeclaration(n) && n.name?.text === 'getAllKeys');
    assert(locals.length <= 1, 'duplicate harness getAllKeys');
    if (locals.length) {
        assert.equal(locals[0].parent, h, 'harness getAllKeys must be module local');
        assert.equal(locals[0].getText(h), pins.local.after, 'changed local function owner');
    }
    const imports = nodes(h, n => ts.isImportSpecifier(n) && n.name.text === 'pushIfUnique');
    assert(imports.length <= 1, 'duplicate pushIfUnique import');
    if (imports.length) assert.equal(imports[0].parent.parent.parent.getText(h), pins.import.after, 'changed helper import');
    const alias = one(h.statements.filter(n => ts.isImportDeclaration(n) && n.moduleSpecifier.text === './_namespaces/ts.js' && n.importClause?.namedBindings && ts.isNamespaceImport(n.importClause.namedBindings)), 'ts namespace import');
    assert.equal(alias.getText(h), pins.namespace, 'changed namespace import');
    const states = [!declaration, spelling === pins.harness.after, locals.length === 1, imports.length === 1];
    assert(states.every(s => s === states[0]), 'half-adapted getAllKeys tree');
    const adapted = states[0];
    assert.equal(nodes(h, n => ts.isIdentifier(n) && n.text === 'getAllKeys').length, adapted ? 2 : 1, 'unexpected harness getAllKeys reference');
    assert.equal(nodes(c, n => ts.isIdentifier(n) && n.text === 'getAllKeys').length, adapted ? 0 : 1, 'unexpected core getAllKeys reference');
    assert.equal(nodes(h, n => ts.isIdentifier(n) && n.text === 'pushIfUnique').length, adapted ? 2 : 0, 'pushIfUnique capture');
    return { adapted, declaration, statement, alias };
}
function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const files = [pins.core.file, pins.harness.file].map(file => path.join(tree, file));
    const raw = files.map(file => fs.readFileSync(file, 'utf8'));
    const newline = raw.map(text => text.includes('\r\n') ? '\r\n' : '\n');
    assert.equal(newline[0], newline[1], 'owner newline styles differ; body must move byte-identically');
    const [core, harness] = raw.map(text => text.replace(/\r\n/g, '\n'));
    const found = inspect(core, harness);
    if (found.adapted) return {files: 0, edits: 0, javascriptChanging: true};
    const afterCore = core.slice(0, found.declaration.jsDoc[0].pos) + core.slice(found.declaration.end);
    const edits = [
        {start: found.statement.getStart(), end: found.statement.end, text: pins.harness.after},
        {start: found.alias.end, end: found.alias.end, text: '\n' + pins.import.after},
        {start: harness.length, end: harness.length, text: '\n' + pins.local.after + '\n'},
    ];
    let afterHarness = harness;
    for (const edit of edits.sort((a,b) => b.start - a.start)) afterHarness = afterHarness.slice(0, edit.start) + edit.text + afterHarness.slice(edit.end);
    assert(inspect(afterCore, afterHarness).adapted, 'planned tree not adapted');
    const originalBody = found.declaration.body.getText();
    const local = one(parse(pins.harness.file, afterHarness).statements.filter(n => ts.isFunctionDeclaration(n) && n.name?.text === 'getAllKeys'), 'planned local function');
    assert.equal(local.body.getText(), originalBody, 'function body changed');
    const output = [afterCore, afterHarness].map(text => newline[0] === '\n' ? text : text.replace(/\n/g, '\r\n'));
    const beforeBody = originalBody.replace(/\n/g, newline[0]);
    assert(output[1].includes(beforeBody), 'body bytes did not move intact');
    files.forEach((file, i) => assert.equal(fs.readFileSync(file, 'utf8'), raw[i], 'source changed while planning'));
    files.forEach((file, i) => fs.writeFileSync(file, output[i]));
    return {files: 2, edits: 4, javascriptChanging: true};
}
module.exports = {adapt, inspect, pins};
if (require.main === module) {
    try { assert.equal(process.argv.length, 3, 'usage: node adapt.cjs TREE'); console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2)); }
    catch (error) { console.error(error.stack); process.exitCode = 1; }
}
