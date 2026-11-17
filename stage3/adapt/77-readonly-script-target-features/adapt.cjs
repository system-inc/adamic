#!/usr/bin/env node
'use strict';
// Adaptation 77, #r11wppy item 1: the internal alias ScriptTargetFeatures in src/compiler/utilities.ts becomes
// ReadonlyMap<string, ReadonlyMap<string, readonly string[]>>, so the shared emptyArray, typed readonly never[] by the
// item 1 change, is a view the alias can no longer write a string into. Adamic ruled it as its own change; README.md
// says why and how it is proved. The alias is @internal, so the published declaration files do not change and no
// api-sanctions.json is needed. Through the stock AST the edit inserts `readonly ` before the innermost `string[]`,
// refuses unless it finds exactly one internal alias of upstream's shape, and changes nothing on a second run. It
// sorts after 76-truthful-casts, so every later adaptation and every consumer reads apply's own tree with the
// innermost array readonly.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
const tree = path.resolve(process.argv[2]);
const fileName = path.join(tree, 'src/compiler/utilities.ts');
// The alias's type as upstream 6.0.3 writes it, and as this adaptation leaves it: any other type is not the alias
// Adamic ruled on.
const upstream = 'ReadonlyMap<string, ReadonlyMap<string, string[]>>';
const adaptedType = 'ReadonlyMap<string, ReadonlyMap<string, readonly string[]>>';
const text = fs.readFileSync(fileName, 'utf8');
const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
const aliases = [];
ts.forEachChild(source, function find(node) {
    if (ts.isTypeAliasDeclaration(node) && node.name.text === 'ScriptTargetFeatures') aliases.push(node);
    ts.forEachChild(node, find);
});
if (aliases.length !== 1) throw new Error(`expected one ScriptTargetFeatures in utilities.ts, found ${aliases.length}`);
const alias = aliases[0];
const where = source.getLineAndCharacterOfPosition(alias.getStart(source));
const record = { alias: 'ScriptTargetFeatures', where: `src/compiler/utilities.ts:${where.line + 1}:${where.character + 1}` };
if (!ts.getJSDocTags(alias).some(tag => tag.tagName.text === 'internal'))
    throw new Error(`expected ScriptTargetFeatures to carry @internal: ${alias.getText(source)}`);
const type = alias.type.getText(source);
if (type !== upstream && type !== adaptedType)
    throw new Error(`unexpected alias ScriptTargetFeatures: ${alias.getText(source)}`);
// The innermost array, reached through the two ReadonlyMap references' second type arguments.
const inner = alias.type.typeArguments[1].typeArguments[1];
let edit;
// Idempotent: an innermost array already readonly keeps its bytes, so a second run changes nothing.
if (type === adaptedType) edit = { ...record, edited: false };
else if (!ts.isArrayTypeNode(inner) || inner.getText(source) !== 'string[]')
    throw new Error(`unexpected innermost array in ScriptTargetFeatures: ${inner.getText(source)}`);
else edit = { ...record, edited: true, at: inner.getStart(source) };
let adapted = text;
if (edit.edited) adapted = adapted.slice(0, edit.at) + 'readonly ' + adapted.slice(edit.at);
if (adapted !== text) {
    if (fs.readFileSync(fileName, 'utf8') !== text) throw new Error(`source changed during adaptation: ${fileName}`);
    fs.writeFileSync(fileName, adapted);
}
const { at, ...printed } = edit;
console.log(JSON.stringify({ typescript: ts.version, task: '#r11wppy', item: 1, alias: printed }, null, 2));
