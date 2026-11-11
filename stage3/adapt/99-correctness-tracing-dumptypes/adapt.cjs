'use strict';
const fs = require('node:fs'), path = require('node:path');
const assert = require('node:assert/strict'), crypto = require('node:crypto');
const ts = require('typescript'), pins = require('./pins.json');
const objectBefore = 'const objectFlags = (type as any).objectFlags;';
const objectAfter = 'const objectFlags = getObjectFlags(type);';
const intrinsicBefore = 'intrinsicName: (type as any).intrinsicName,';
const intrinsicAfter = 'intrinsicName: type.flags & TypeFlags.Intrinsic ? (type as IntrinsicType).intrinsicName : undefined,';
const sha = text => crypto.createHash('sha256').update(text).digest('hex');
function owner(text) {
    const source = ts.createSourceFile('tracing.ts', text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, 'tracing.ts parse drift');
    const found = [];
    function walk(node) {
        if (ts.isFunctionDeclaration(node) && node.name?.text === 'dumpTypes' && node.body) found.push(node);
        ts.forEachChild(node, walk);
    }
    walk(source);
    assert.equal(found.length, 1, 'dumpTypes owner drift');
    const node = found[0], parent = node.parent;
    assert(ts.isModuleBlock(parent) && ts.isModuleDeclaration(parent.parent) &&
        parent.parent.name.text === 'tracingEnabled', 'dumpTypes namespace drift');
    return {source, node, text: node.getText(source), start: node.getStart(source), end: node.end};
}
function plan(text) {
    assert.equal(ts.version, pins.version);
    for (const state of ['before', 'after']) assert.equal(sha(pins[state]), pins[state + 'Sha256']);
    assert.equal(pins.before.replace(objectBefore, objectAfter).replace(intrinsicBefore, intrinsicAfter), pins.after,
        'only the two reviewed subtype reads may change');
    const original = owner(text);
    let state, newline;
    for (const candidate of ['before', 'after']) {
        if (original.text === pins[candidate]) { state = candidate; newline = '\r\n'; }
        else if (original.text === pins[candidate].replace(/\r\n/g, '\n')) { state = candidate; newline = '\n'; }
    }
    assert(state, 'unreviewed or partially adapted dumpTypes');
    const imports = original.source.statements.filter(node => ts.isImportDeclaration(node) &&
        ts.isStringLiteral(node.moduleSpecifier) && node.moduleSpecifier.text === './_namespaces/ts.js');
    assert.equal(imports.length, 1, 'namespace import drift');
    const declaration = imports[0];
    assert(!declaration.importClause.isTypeOnly && ts.isNamedImports(declaration.importClause.namedBindings));
    const elements = declaration.importClause.namedBindings.elements;
    const edits = [];
    for (const [name, anchor, typeOnly] of [
        ['getObjectFlags', 'getLineAndCharacterOfPosition', false],
        ['IntrinsicType', 'IntersectionType', true],
    ]) {
        const current = elements.filter(e => e.name.text === name || e.propertyName?.text === name);
        assert.equal(current.length, state === 'after' ? 1 : 0, 'partial or duplicate import: ' + name);
        if (state === 'after') {
            assert.equal(current[0].name.text, name);
            assert(!current[0].propertyName, 'aliased import: ' + name);
            assert.equal(!!current[0].isTypeOnly, typeOnly, 'import kind drift: ' + name);
            continue;
        }
        const anchors = elements.filter(e => e.name.text === anchor);
        assert.equal(anchors.length, 1, 'import anchor drift: ' + anchor);
        const entry = anchors[0];
        assert.equal(entry.getText(original.source), (typeOnly ? 'type ' : '') + anchor);
        const importNewline = declaration.getText(original.source).includes('\r\n') ? '\r\n' : '\n';
        edits.push({start: entry.end, end: entry.end, text: ',' + importNewline + '    ' +
            (typeOnly ? 'type ' : '') + name});
    }
    if (state === 'after') return text;
    const replacement = pins.after.replace(/\r\n/g, newline);
    for (const removeComments of [false, true]) {
        const emit = value => ts.transpileModule(value, {compilerOptions: {
            target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments,
        }}).outputText;
        const before = emit(original.text);
        const expected = before.replace('const objectFlags = type.objectFlags;', objectAfter)
            .replace('intrinsicName: type.intrinsicName,',
                'intrinsicName: type.flags & TypeFlags.Intrinsic ? type.intrinsicName : undefined,');
        assert.notEqual(before, expected, 'emitted read drift');
        assert.equal(emit(replacement), expected, 'unreviewed dumpTypes JavaScript change');
    }
    edits.push({start: original.start, end: original.end, text: replacement});
    let changed = text;
    for (const edit of edits.sort((a, b) => b.start - a.start)) {
        changed = changed.slice(0, edit.start) + edit.text + changed.slice(edit.end);
    }
    assert.equal(owner(changed).text, replacement);
    assert.equal(plan(changed), changed, 'after-state validation');
    return changed;
}
function adapt(tree) {
    const file = path.join(tree, 'src/compiler/tracing.ts');
    const before = fs.readFileSync(file, 'utf8'), after = plan(before);
    assert.equal(fs.readFileSync(file, 'utf8'), before, 'concurrent source change');
    if (after !== before) fs.writeFileSync(file, after);
    return {files: Number(after !== before), edits: after === before ? 0 : 2, javascriptChanging: true};
}
module.exports = {owner, plan, adapt};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(process.argv[2])));
}
