#!/usr/bin/env node
'use strict';
// Adaptation 85, #aryr7ea: tsc's four range setters in src/compiler/utilities.ts take a TextRange in place of a
// ReadonlyTextRange, since each writes the range's pos or end: setTextRangePos and setTextRangeEnd through a cast, and
// setTextRangePosEnd and setTextRangePosWidth through them. Adamic ruled it (Nov 18) under rulings.md:391, extended by
// name for these four internal functions: the signature says what the body does, and after adaptation 80 every Node
// and NodeArray already is a TextRange. It's type-only. Through the stock AST the edit replaces the one
// ReadonlyTextRange constraint on each setter's type parameter, refuses before any write unless it finds exactly the
// four declarations upstream 6.0.3 writes, and changes nothing on a second run. README.md names the premise and the
// measurements.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
const tree = path.resolve(process.argv[2]);
const fileName = path.join(tree, 'src/compiler/utilities.ts');
// Each setter's parameters as upstream 6.0.3 writes them: a setter whose parameters have moved is not the one Adamic
// ruled on.
const setters = {
    setTextRangePos: 'range: T, pos: number',
    setTextRangeEnd: 'range: T, end: number',
    setTextRangePosEnd: 'range: T, pos: number, end: number',
    setTextRangePosWidth: 'range: T, pos: number, width: number',
};
const text = fs.readFileSync(fileName, 'utf8');
const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
const functions = [];
ts.forEachChild(source, node => {
    if (ts.isFunctionDeclaration(node) && node.name && Object.hasOwn(setters, node.name.text)) functions.push(node);
});
const names = functions.map(node => node.name.text).sort();
if (names.join(',') !== Object.keys(setters).sort().join(','))
    throw new Error(`expected exactly one declaration of each range setter in utilities.ts, found ${JSON.stringify(names)}`);
const edits = [];
for (const node of functions) {
    const where = source.getLineAndCharacterOfPosition(node.getStart(source));
    const record = { function: node.name.text, where: `src/compiler/utilities.ts:${where.line + 1}:${where.character + 1}` };
    const parameters = node.parameters.map(parameter => parameter.getText(source)).join(', ');
    const typeParameters = node.typeParameters || [];
    if (parameters !== setters[node.name.text] || typeParameters.length !== 1 || typeParameters[0].name.text !== 'T' || !typeParameters[0].constraint)
        throw new Error(`unexpected signature on ${node.name.text}: <${typeParameters.map(type => type.getText(source)).join(', ')}>(${parameters})`);
    const constraint = typeParameters[0].constraint;
    const written = constraint.getText(source);
    // Idempotent: a constraint already adapted keeps its bytes, so a second run changes nothing.
    if (written === 'TextRange') {
        edits.push({ ...record, edited: false });
        continue;
    }
    if (written !== 'ReadonlyTextRange') throw new Error(`unexpected constraint on ${node.name.text}: ${written}`);
    edits.push({ ...record, edited: true, at: constraint.getStart(source), end: constraint.end });
}
let adapted = text;
for (const edit of edits.filter(edit => edit.edited).sort((a, b) => b.at - a.at))
    adapted = adapted.slice(0, edit.at) + 'TextRange' + adapted.slice(edit.end);
if (adapted !== text) {
    if (fs.readFileSync(fileName, 'utf8') !== text) throw new Error(`source changed during adaptation: ${fileName}`);
    fs.writeFileSync(fileName, adapted);
}
console.log(JSON.stringify({ typescript: ts.version, task: '#aryr7ea', ruling: 'rulings.md:391, Adamic, Nov 18',
    setters: edits.map(({ at, end, ...edit }) => edit) }, null, 2));
