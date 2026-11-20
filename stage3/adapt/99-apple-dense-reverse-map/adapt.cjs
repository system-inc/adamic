'use strict';
// #s59veah: scanner.ts's makeReverseMap builds its reverse map densely. Upstream writes result[value] = name into an
// empty array, so the array grows by writes past its end and holds holes where no entry names an index; rulings.md:318
// lets a write land only at an index the array already holds, and stops the sparse build. The dense build first sizes
// the array to the largest value plus one, filled with undefined by push, then writes each entry at an index it holds.
// Every index read gives what it gave before: a hole read undefined and now holds undefined, and each written index
// holds the same last-written name, since both builds write the map's entries in the same order. The length is the
// same, the largest value plus one. Its only readers are tokenToString's and regularExpressionFlagToCharacterCode's
// index reads. The owner is pinned whole, before and after, with its bytes and hashes; anything else is refused, and a
// second application writes nothing.
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const crypto = require('node:crypto');
const ts = require('typescript');
const pins = require('./pins.json');
const digest = text => crypto.createHash('sha256').update(text).digest('hex');

function owner(text) {
    const source = ts.createSourceFile('scanner.ts', text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, 'scanner.ts: parse error');
    const found = [];
    ts.forEachChild(source, function visit(node) {
        if (ts.isFunctionDeclaration(node) && node.name?.text === pins.owner && node.body) found.push(node);
        ts.forEachChild(node, visit);
    });
    assert.equal(found.length, 1, `${pins.owner}: missing or duplicate owner`);
    const start = found[0].getStart(source), end = found[0].end, spelling = text.slice(start, end);
    const state = spelling === pins.before ? 'before' : spelling === pins.after ? 'after' : undefined;
    assert(state, `${pins.owner}: unreviewed owner text`);
    assert.equal(digest(spelling), pins[`${state}Sha256`], `${pins.owner}: owner hash drift`);
    return {start, end, state};
}

function plan(text) {
    assert.equal(ts.version, '6.0.3');
    const found = owner(text);
    if (found.state === 'after') return text;
    const changed = text.slice(0, found.start) + pins.after + text.slice(found.end);
    assert.equal(owner(changed).state, 'after', `${pins.owner}: invalid planned owner`);
    // The JavaScript changes only inside the owner: the emit of the file with the before owner put back is the
    // before file's emit, for both comment settings.
    for (const removeComments of [true, false]) {
        const emit = source => ts.transpileModule(source, {compilerOptions: {target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments}}).outputText;
        const before = emit(text), after = emit(changed);
        // The owner alone is a script, so its emit opens with a "use strict" prologue the module's emit doesn't hold.
        const alone = source => emit(source).replace(/^"use strict";\r?\n/, '');
        const ownerBefore = alone(pins.before), ownerAfter = alone(pins.after);
        assert.equal(before.split(ownerBefore.trim()).length - 1, 1, `${pins.owner}: before owner's JavaScript not found once`);
        assert.equal(after, before.replace(ownerBefore.trim(), ownerAfter.trim()), `${pins.owner}: unreviewed JavaScript change`);
    }
    // The owner is internal, unexported, so no declaration changes.
    for (const stripInternal of [true, false]) {
        const emit = source => ts.transpileDeclaration(source, {fileName: 'scanner.ts', compilerOptions: {stripInternal}}).outputText;
        assert.equal(emit(changed), emit(text), 'declarations changed');
    }
    return changed;
}

function adapt(tree) {
    const file = path.join(tree, pins.file), before = fs.readFileSync(file, 'utf8'), after = plan(before);
    assert.equal(fs.readFileSync(file, 'utf8'), before, `${pins.file}: changed while planning`);
    if (after !== before) fs.writeFileSync(file, after);
    return {files: Number(after !== before), edits: Number(after !== before), javascriptChanging: true};
}

module.exports = {owner, plan, adapt, pins, digest};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs TREE');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2]))));
}
