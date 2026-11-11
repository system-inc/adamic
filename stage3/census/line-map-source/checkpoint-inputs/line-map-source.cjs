'use strict';
// Adamic's baobab ruling, Nov 11: the internal overload follows the existing
// getPositionOfLineAndCharacter precedent. The public signature stays intact.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const pins = require('./line-map-source.json');

function plan(tree, ts, pending = new Map()) {
    assert.equal(ts.version, '6.0.3');
    const read = name => pending.get(path.join(tree, name)) ?? fs.readFileSync(path.join(tree, name), 'utf8');
    const typesName = 'src/compiler/types.ts', scannerName = 'src/compiler/scanner.ts';
    const texts = new Map([[typesName, read(typesName)], [scannerName, read(scannerName)]]);
    const parsed = new Map([...texts].map(([name, text]) => [name, ts.createSourceFile(name, text, ts.ScriptTarget.Latest, true)]));
    for (const file of parsed.values()) assert.equal(file.parseDiagnostics.length, 0, `cannot parse ${file.fileName}`);
    const types = parsed.get(typesName), scanner = parsed.get(scannerName);
    const one = (items, label) => { assert.equal(items.length, 1, `${label} count`); return items[0]; };
    const documented = node => {
        const start = node.jsDoc?.[0]?.pos ?? node.getStart(node.getSourceFile());
        return node.getSourceFile().text.slice(start, node.end);
    };
    const source = one(types.statements.filter(n => ts.isInterfaceDeclaration(n) && n.name.text === 'SourceFileLike'), 'SourceFileLike');
    assert.equal(documented(source), pins.sourceFileLike, 'documented SourceFileLike drift');
    const views = types.statements.filter(n => ts.isInterfaceDeclaration(n) && n.name.text === 'LineMapSource');
    const calls = scanner.statements.filter(n => ts.isFunctionDeclaration(n) && n.name?.text === 'getLineAndCharacterOfPosition');
    const starts = one(scanner.statements.filter(n => ts.isFunctionDeclaration(n) && n.name?.text === 'getLineStarts'), 'getLineStarts');
    const imports = scanner.statements.filter(ts.isImportDeclaration).flatMap(n => n.importClause?.namedBindings && ts.isNamedImports(n.importClause.namedBindings) ? [...n.importClause.namedBindings.elements] : []);
    const imported = imports.filter(n => n.name.text === 'LineMapSource');
    const states = [];
    states.push(views.length === 0 ? 'before' : views.length === 1 && documented(views[0]) === pins.lineMapSource ? 'after' : 'drift');
    states.push(calls.length === 1 && documented(calls[0]) === pins.position.before ? 'before' : calls.length === 3 &&
        scanner.text.slice(calls[0].jsDoc[0].pos, calls[2].end) === pins.position.after ? 'after' : 'drift');
    states.push(documented(starts) === pins.starts.before ? 'before' : documented(starts) === pins.starts.after ? 'after' : 'drift');
    const importBefore = '    type LineAndCharacter,', importAfter = importBefore + ' type LineMapSource,';
    states.push(imported.length === 0 && scanner.text.split(importBefore).length === 2 ? 'before' : imported.length === 1 && imported[0].isTypeOnly && scanner.text.split(importAfter).length === 2 ? 'after' : 'drift');
    assert(states.every(s => s === states[0]) && states[0] !== 'drift', `line map drift or half-adapted files: ${states.join(', ')}`);
    if (states[0] === 'after') {
        assert.equal(types.statements.indexOf(views[0]) + 1, types.statements.indexOf(source), 'LineMapSource must immediately precede SourceFileLike');
        assert(types.text.includes(pins.lineMapSource + '\r\n\r\n' + pins.sourceFileLike), 'LineMapSource documentation placement drift');
        return new Map();
    }
    const replace = (text, before, after) => { assert.equal(text.split(before).length, 2, 'pin must occur exactly once'); return text.replace(before, after); };
    const output = new Map();
    output.set(path.join(tree, typesName), replace(types.text, pins.sourceFileLike, pins.lineMapSource + '\r\n\r\n' + pins.sourceFileLike));
    let changed = replace(scanner.text, pins.position.before, pins.position.after);
    changed = replace(changed, pins.starts.before, pins.starts.after);
    changed = replace(changed, importBefore, importAfter);
    output.set(path.join(tree, scannerName), changed);
    // Actual stock JavaScript bytes, including comments, must stay unchanged.
    for (const [name, after] of output) {
        const before = read(path.relative(tree, name));
        const emit = text => ts.transpileModule(text, { fileName: name, compilerOptions: { target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, verbatimModuleSyntax: true } }).outputText;
        assert.equal(emit(after), emit(before), `JavaScript changed: ${name}`);
    }
    plan(tree, ts, new Map([...pending, ...output]));
    return output;
}
module.exports = { plan };
