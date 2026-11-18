#!/usr/bin/env node
'use strict';
// Adaptation 80, #aryr7ea: Node and NodeArray in src/compiler/types.ts extend TextRange in place of ReadonlyTextRange,
// so their pos and end are writable, as tsc's own public setTextRange writes them on a node after the factory made it.
// Adamic ruled it (Nov 17) as adding truth under rulings.md:391, for these two interfaces and these two members by
// name: the readonly was a promise to consumers, never a fact about the value, and rulings.md:93 and :112 stand whole.
// It's type-only. Through the stock AST the edit replaces the one ReadonlyTextRange in each heritage clause, refuses
// before any write unless it finds exactly the two heritage clauses upstream 6.0.3 writes, and changes nothing on a
// second run. It sorts after 72-readonly-diagnostic-slots, so every later adaptation and every consumer reads apply's
// tree with Node's and NodeArray's pos and end writable. README.md names the premise, the cost and the measurements.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
const tree = path.resolve(process.argv[2]);
const fileName = path.join(tree, 'src/compiler/types.ts');
// Each interface's heritage as upstream 6.0.3 writes it, and as this adaptation leaves it: an interface whose heritage
// has moved is not the one Adamic ruled on.
const heritage = {
    Node: { upstream: 'extends ReadonlyTextRange', adapted: 'extends TextRange' },
    NodeArray: { upstream: 'extends ReadonlyArray<T>, ReadonlyTextRange', adapted: 'extends ReadonlyArray<T>, TextRange' },
};
const text = fs.readFileSync(fileName, 'utf8');
const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
const interfaces = [];
ts.forEachChild(source, function find(node) {
    if (ts.isInterfaceDeclaration(node) && Object.hasOwn(heritage, node.name.text)) interfaces.push(node);
    ts.forEachChild(node, find);
});
const names = interfaces.map(node => node.name.text).sort();
if (names.join(',') !== Object.keys(heritage).sort().join(','))
    throw new Error(`expected exactly one Node and one NodeArray interface in types.ts, found ${JSON.stringify(names)}`);
const edits = [];
for (const node of interfaces) {
    const clauses = node.heritageClauses || [];
    const where = source.getLineAndCharacterOfPosition(node.getStart(source));
    const record = { interface: node.name.text, where: `src/compiler/types.ts:${where.line + 1}:${where.character + 1}` };
    if (clauses.length !== 1 || clauses[0].token !== ts.SyntaxKind.ExtendsKeyword)
        throw new Error(`unexpected heritage on ${node.name.text}: ${clauses.map(clause => clause.getText(source)).join(' ')}`);
    const clause = clauses[0].getText(source);
    // Idempotent: a heritage already adapted keeps its bytes, so a second run changes nothing.
    if (clause === heritage[node.name.text].adapted) {
        edits.push({ ...record, edited: false });
        continue;
    }
    if (clause !== heritage[node.name.text].upstream)
        throw new Error(`unexpected heritage on ${node.name.text}: ${clause}`);
    const readonly = clauses[0].types.filter(type => type.expression.getText(source) === 'ReadonlyTextRange' && !type.typeArguments);
    if (readonly.length !== 1) throw new Error(`expected one ReadonlyTextRange in ${node.name.text}'s heritage`);
    edits.push({ ...record, edited: true, at: readonly[0].getStart(source), end: readonly[0].end });
}
let adapted = text;
for (const edit of edits.filter(edit => edit.edited).sort((a, b) => b.at - a.at))
    adapted = adapted.slice(0, edit.at) + 'TextRange' + adapted.slice(edit.end);
if (adapted !== text) {
    if (fs.readFileSync(fileName, 'utf8') !== text) throw new Error(`source changed during adaptation: ${fileName}`);
    fs.writeFileSync(fileName, adapted);
}
console.log(JSON.stringify({ typescript: ts.version, task: '#aryr7ea', ruling: 'rulings.md:391, Adamic, Nov 17',
    interfaces: edits.map(({ at, end, ...edit }) => edit) }, null, 2));
