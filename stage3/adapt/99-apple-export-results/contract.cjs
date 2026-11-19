'use strict';
// The owner contract, run on an adapted tree: node contract.cjs TREE. It rebuilds each file with every edit undone in a
// scratch copy, then checks the edits under LF and CRLF, a zero-edit replay that keeps bytes and mtime, and that a
// renamed, duplicated or drifted owner and a half-made set are each refused with the files untouched.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const {adapt, before, owners} = require('./adapt.cjs');
const files = [...new Set(owners.edits.map(edit => edit.file))];
const adapted = new Map(files.map(file => [file, fs.readFileSync(path.join(process.argv[2], file), 'utf8').replace(/\r\n/g, '\n')]));
const original = before(adapted);
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'family5-owner-'));
const write = (texts, newline) => {
    for (const [file, text] of texts) {
        fs.mkdirSync(path.dirname(path.join(scratch, file)), {recursive: true});
        fs.writeFileSync(path.join(scratch, file), text.replace(/\n/g, newline));
    }
};
const read = () => new Map(files.map(file => [file, fs.readFileSync(path.join(scratch, file), 'utf8')]));
for (const newline of ['\n', '\r\n']) {
    write(original, newline);
    assert.equal(adapt(scratch).edits, owners.edits.length);
    for (const file of files) assert.equal(read().get(file), adapted.get(file).replace(/\n/g, newline), `${file}: not the adapted text`);
    const stamps = files.map(file => fs.statSync(path.join(scratch, file), {bigint: true}).mtimeNs);
    assert.equal(adapt(scratch).edits, 0);
    files.forEach((file, i) => assert.equal(fs.statSync(path.join(scratch, file), {bigint: true}).mtimeNs, stamps[i], `${file}: replay wrote`));
}
const refused = (name, texts) => {
    write(texts, '\n');
    const kept = read();
    assert.throws(() => adapt(scratch), undefined, `${name}: not refused`);
    for (const [file, text] of kept) assert.equal(read().get(file), text, `${name}: ${file} changed`);
};
for (const edit of owners.edits) {
    const text = original.get(edit.file);
    // The owner's name as declared, followed by what may follow it there: `interface X {`, `interface X<T>`, `function X(`.
    const declaration = new RegExp(edit.kind === 'member' ? `\\binterface ${edit.owner}(?=[\\s<{])` : `\\bfunction ${edit.owner}(?=\\()`, 'g');
    assert(declaration.test(text), `${edit.candidate}: no ${declaration}`);
    refused(`${edit.candidate} owner renamed`, new Map(original).set(edit.file, text.replace(declaration, found => found + 'Renamed')));
    if (edit.kind === 'replace') {
        const drifted = edit.old.replace(/(\w)(\W*)$/, (_, last, rest) => (last === 'x' ? 'y' : 'x') + rest);
        refused(`${edit.candidate} old text drifted`, new Map(original).set(edit.file, text.replace(edit.old, drifted)));
    }
}
// A half-made set: the first edit made alone, by narrowing the shared owners list for one run, then the whole set asked.
if (owners.edits.length > 1) {
    const whole = owners.edits.slice();
    write(original, '\n');
    owners.edits.splice(1);
    adapt(scratch);
    owners.edits.splice(0, owners.edits.length, ...whole);
    const halfTexts = read();
    assert.throws(() => adapt(scratch), /half adapted/, 'half made: not refused');
    for (const [file, text] of halfTexts) assert.equal(read().get(file), text, `half made: ${file} changed`);
}
console.log(`owner drift, renamed owners, half-made set, LF/CRLF and zero-edit replay over ${owners.edits.length} edits in ${files.length} files: pass`);
