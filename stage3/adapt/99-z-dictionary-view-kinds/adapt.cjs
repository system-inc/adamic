'use strict';
// #2v3xtgm, Compiler's ruling (b): the two owners where Node hands tsc's `object as MapLike<unknown>` something other
// than a record get value-identical guards, so Adamic's checked dictionary view, which stops the program on anything
// but a record of references, only ever sees records, and Node's every answer stays the same.
//   - utilities.ts's isJsonEqual compares arrays itself, as equalOwnProperties answers for an array holding every
//     index below its length and no other key, and casts only when both sides are records. Its one caller,
//     optionsHaveChanges, hands it every option value, so paths' arrays, lib, types, rootDirs and plugins all reached
//     the cast.
//   - moduleNameResolver.ts's loadModuleFromExports reads exports through the view only once it's an object, and at
//     allKeysStartWithDot only once it's an object that isn't an array. A package.json's exports can be true, a number,
//     a string or an array, and every such kind gets the same trace and the same undefined it got before.
// README.md has the kind tables. Each owner is pinned whole, before and after, with its CRLF bytes and hashes
// (pins.json), and found by its name, never a coordinate; anything else is refused, and a second application writes
// nothing. Directory 99-z sorts after every adaptation that edits either file (99-getter-callable-views and
// 99-readonly-inferred-bindings edit utilities.ts), so the files it plans over are the ones they leave.
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const crypto = require('node:crypto');
const ts = require('typescript');
const pins = require('./pins.json');
const digest = text => crypto.createHash('sha256').update(text).digest('hex');

// owner finds pin's owner in text, the one function declaration with a body that the pin names, and says whether it
// holds the pinned before or after text.
function owner(text, pin) {
    const source = ts.createSourceFile(path.basename(pin.file), text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, `${pin.file}: parse error`);
    const found = [];
    ts.forEachChild(source, function visit(node) {
        if (ts.isFunctionDeclaration(node) && node.name?.text === pin.owner && node.body) found.push(node);
        ts.forEachChild(node, visit);
    });
    assert.equal(found.length, 1, `${pin.owner}: missing or duplicate owner`);
    const start = found[0].getStart(source), end = found[0].end, spelling = text.slice(start, end);
    const state = spelling === pin.before ? 'before' : spelling === pin.after ? 'after' : undefined;
    assert(state, `${pin.owner}: unreviewed owner text`);
    assert.equal(digest(spelling), pin[`${state}Sha256`], `${pin.owner}: owner hash drift`);
    return {start, end, state};
}

function plan(text, pin) {
    assert.equal(ts.version, '6.0.3');
    const found = owner(text, pin);
    if (found.state === 'after') return text;
    const changed = text.slice(0, found.start) + pin.after + text.slice(found.end);
    assert.equal(owner(changed, pin).state, 'after', `${pin.owner}: invalid planned owner`);
    // The JavaScript changes only inside the owner: the emit of the file with the after owner is the before file's
    // emit with the owner's own emit replaced, for both comment settings. The imports the after owner reads (isArray,
    // hasProperty, every) are each read elsewhere in its file already, so no import's elision moves.
    for (const removeComments of [true, false]) {
        const emit = source => ts.transpileModule(source, {compilerOptions: {target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments}}).outputText;
        const before = emit(text), after = emit(changed);
        // An owner alone without export is a script, whose emit opens with a "use strict" prologue the module's
        // emit doesn't hold.
        const alone = source => emit(source).replace(/^"use strict";\r?\n/, '');
        const ownerBefore = alone(pin.before), ownerAfter = alone(pin.after);
        assert.equal(before.split(ownerBefore.trim()).length - 1, 1, `${pin.owner}: before owner's JavaScript not found once`);
        assert.equal(after, before.replace(ownerBefore.trim(), ownerAfter.trim()), `${pin.owner}: unreviewed JavaScript change`);
    }
    // isJsonEqual is internal and keeps its signature, and loadModuleFromExports is unexported, so no declaration
    // changes under either value of stripInternal.
    for (const stripInternal of [true, false]) {
        const emit = source => ts.transpileDeclaration(source, {fileName: path.basename(pin.file), compilerOptions: {stripInternal}}).outputText;
        assert.equal(emit(changed), emit(text), `${pin.file}: declarations changed`);
    }
    return changed;
}

function adapt(tree) {
    let edits = 0;
    for (const pin of pins.owners) {
        const file = path.join(tree, pin.file), before = fs.readFileSync(file, 'utf8'), after = plan(before, pin);
        assert.equal(fs.readFileSync(file, 'utf8'), before, `${pin.file}: changed while planning`);
        if (after !== before) {
            fs.writeFileSync(file, after);
            edits++;
        }
    }
    return {files: edits, edits, javascriptChanging: true, task: '#2v3xtgm'};
}

module.exports = {owner, plan, adapt, pins, digest};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs TREE');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2]))));
}
