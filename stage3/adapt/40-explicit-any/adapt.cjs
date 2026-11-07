// Reviewed owner types and redundant Error assertions are changed.
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
if (ts.version !== '6.0.3') throw Error('requires stock TypeScript 6.0.3');
if (process.argv.length !== 3) throw Error('usage: node adapt.cjs <tree>');
const file = path.join(path.resolve(process.argv[2]), 'src/compiler/core.ts');
const text = fs.readFileSync(file, 'utf8');
const helper = require('./hasproperty.cjs');
// Preserve downstream adapter inputs; the shared owner's final generic phase runs at 99.
const helperPlan = helper.plan(file, text, 'early');
const source = helper.parse(file, text);
const names = new Set(['length', 'toOffset', 'hasProperty', 'isArray']);
const bodies = new Map([
    ['isArray', '{\n    // See: https://github.com/microsoft/TypeScript/issues/17002\n    return Array.isArray(value);\n}'],
    ['length', '{\n    return array !== undefined ? array.length : 0;\n}'],
    ['toOffset', '{\n    return offset < 0 ? array.length + offset : offset;\n}'],
    ['hasProperty', '{\n    return hasOwnProperty.call(map, key);\n}'],
]);
const edits = [];
for (const node of source.statements) {
    if (!ts.isFunctionDeclaration(node) || !names.has(node.name?.text)) continue;
    names.delete(node.name.text);
    if (node.name.text === 'hasProperty') {
        if (helperPlan.edit) edits.push(helperPlan.edit);
        continue;
    }
    if (node.body?.getText(source).replace(/\r\n/g, '\n') !== bodies.get(node.name.text)) throw Error('unreviewed uses: ' + node.name.text);
    const parameter = node.parameters[0];
    const type = parameter.type;
    if (node.name.text === 'isArray') {
        if (parameter.name.text !== 'value' || node.typeParameters ||
            node.type?.getText(source) !== 'value is readonly unknown[]') throw Error('isArray predicate contract changed');
        if (type.kind === ts.SyntaxKind.UnknownKeyword) continue;
        if (type.kind !== ts.SyntaxKind.AnyKeyword) throw Error('unreviewed isArray input');
        edits.push({start: type.getStart(source), end: type.end, value: 'unknown'});
        continue;
    }
    const array = ts.isUnionTypeNode(type) ? type.types[0] : type;
    if (!ts.isTypeOperatorNode(array) || array.operator !== ts.SyntaxKind.ReadonlyKeyword || !ts.isArrayTypeNode(array.type)) throw Error('array contract changed');
    const element = array.type.elementType;
    if (node.typeParameters?.length === 1 && node.typeParameters[0].name.text === 'T' && element?.getText(source) === 'T') continue;
    if (node.typeParameters || element?.kind !== ts.SyntaxKind.AnyKeyword) throw Error('unreviewed signature: ' + node.name.text);
    edits.push({start: node.name.end, end: node.name.end, value: '<T>'});
    edits.push({start: element.getStart(source), end: element.end, value: 'T'});
}
if (names.size) throw Error('missing owners: ' + [...names].join(', '));
let result = text;
for (const edit of edits.sort((a, b) => b.start - a.start)) result = result.slice(0, edit.start) + edit.value + result.slice(edit.end);
for (const removeComments of [false, true]) if (helper.transpile(text, removeComments) !== helper.transpile(result, removeComments)) throw Error('core.ts JavaScript changed');
if (fs.readFileSync(file, 'utf8') !== text) throw Error('core changed while planning');
if (result !== text) fs.writeFileSync(file, result);
console.log(JSON.stringify({owners: edits.filter(e => e.value !== '<T>').length, removed: edits.filter(e => e.value !== '<T>').length}));

require('./classes.cjs').apply(path.resolve(process.argv[2]));
require('./classes.cjs').applyEnums(path.resolve(process.argv[2]));
require('./classes.cjs').applyEnums(path.resolve(process.argv[2]), 'diagnostic');
require('./classes.cjs').applyDiagnosticReference(path.resolve(process.argv[2]));
require('./classes.cjs').applyDiagnosticDeclarations(path.resolve(process.argv[2]));
require('./classes.cjs').applyEnums(path.resolve(process.argv[2]), 'filesystem');

require('./error-host.cjs').apply(path.resolve(process.argv[2]));

require('./classes.cjs').applyTimerSlots(path.resolve(process.argv[2]));
