#!/usr/bin/env node
'use strict';
// Adaptation 99-readonly-script-target-entries, #r11wppy class 2: the Object.entries calls in getScriptTargetFeatures'
// literal in src/compiler/utilities.ts that tsc would otherwise infer a union for get explicit type arguments, so each
// inner map is judged against ScriptTargetFeatures' read-only entry type. Adamic ruled the minimal form; README.md says
// why and how it is proved. getScriptTargetFeatures is @internal, so the published declaration files do not change and
// no api-sanctions.json is needed. Through the stock AST the edit inserts `<ReadonlyMap<string, readonly string[]>>`
// after the outer `Object.entries`, and `<readonly string[]>` after the inner one of each map whose literal mixes a bare
// `emptyArray` with an array literal, the three upstream holds. It refuses before any write unless the function has
// upstream's shape, adaptation 77's alias and exactly those three mixed maps, and changes nothing on a second run. It
// sorts after 77-readonly-script-target-features, whose alias the outer argument must match.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
const tree = path.resolve(process.argv[2]);
const fileName = path.join(tree, 'src/compiler/utilities.ts');
// The alias as adaptation 77 leaves it: the outer argument is its entry type, so any other alias is not the one
// Adamic ruled on.
const alias = 'ReadonlyMap<string, ReadonlyMap<string, readonly string[]>>';
const outerArgument = 'ReadonlyMap<string, readonly string[]>';
const innerArgument = 'readonly string[]';
// The maps upstream 6.0.3's literal mixes, the three the measurement found stops in (J-ezkkktv). A map that mixes and
// is not one of these is a literal Adamic did not rule on, so it refuses rather than be annotated by guess.
const mixedNames = ['Promise', 'BigInt64Array', 'BigUint64Array'];
const text = fs.readFileSync(fileName, 'utf8');
const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
const nodes = (root, predicate) => {
    const found = [];
    ts.forEachChild(root, function visit(node) {
        if (predicate(node)) found.push(node);
        ts.forEachChild(node, visit);
    });
    return found;
};
const where = node => {
    const at = source.getLineAndCharacterOfPosition(node.getStart(source));
    return `src/compiler/utilities.ts:${at.line + 1}:${at.character + 1}`;
};
const isEntries = node => ts.isCallExpression(node) && ts.isPropertyAccessExpression(node.expression)
    && ts.isIdentifier(node.expression.expression) && node.expression.expression.text === 'Object'
    && ts.isIdentifier(node.expression.name) && node.expression.name.text === 'entries';
const aliases = nodes(source, node => ts.isTypeAliasDeclaration(node) && node.name.text === 'ScriptTargetFeatures');
if (aliases.length !== 1 || aliases[0].type.getText(source) !== alias)
    throw new Error(`expected one ScriptTargetFeatures of adaptation 77's shape ${alias}, found ${aliases.map(found => found.getText(source)).join(' | ') || 'none'}`);
const declarations = nodes(source, node => ts.isVariableDeclaration(node) && ts.isIdentifier(node.name)
    && node.name.text === 'getScriptTargetFeatures');
if (declarations.length !== 1) throw new Error(`expected one getScriptTargetFeatures in utilities.ts, found ${declarations.length}`);
const declaration = declarations[0];
const statement = declaration.parent.parent;
if (!ts.isVariableStatement(statement) || declaration.parent.declarations.length !== 1
    || (declaration.parent.flags & ts.NodeFlags.BlockScoped) !== ts.NodeFlags.Const
    || !statement.modifiers?.some(modifier => modifier.kind === ts.SyntaxKind.ExportKeyword)
    || !ts.getJSDocTags(statement).some(tag => tag.tagName.text === 'internal'))
    throw new Error(`expected getScriptTargetFeatures to be the one declaration of an exported @internal const: ${statement.getText(source).split('\n')[0]}`);
// Upstream's `memoize((): ScriptTargetFeatures => new Map(Object.entries({...})))`, read node by node.
const initializer = declaration.initializer;
const arrow = initializer && ts.isCallExpression(initializer) && ts.isIdentifier(initializer.expression)
    && initializer.expression.text === 'memoize' && !initializer.typeArguments && initializer.arguments.length === 1
    ? initializer.arguments[0] : undefined;
if (declaration.type?.getText(source) !== '() => ScriptTargetFeatures' || !arrow || !ts.isArrowFunction(arrow)
    || arrow.parameters.length !== 0 || arrow.type?.getText(source) !== 'ScriptTargetFeatures')
    throw new Error(`unexpected getScriptTargetFeatures at ${where(declaration)}: ${statement.getText(source).split('\n')[0]}`);
// A `new Map(Object.entries({...}))`: its call and its literal, or a refusal naming what it found.
function entries(node, what) {
    const call = ts.isNewExpression(node) && ts.isIdentifier(node.expression) && node.expression.text === 'Map'
        && !node.typeArguments && node.arguments?.length === 1 ? node.arguments[0] : undefined;
    if (!call || !isEntries(call) || call.arguments.length !== 1 || !ts.isObjectLiteralExpression(call.arguments[0]))
        throw new Error(`expected ${what} at ${where(node)} to be new Map(Object.entries({...})): ${node.getText(source).split('\n')[0]}`);
    return { call, literal: call.arguments[0] };
}
// An inner literal mixes when one of its properties holds the bare identifier `emptyArray` and another holds an array
// literal: there Object.entries infers `string[] | never[]`, and the bare `emptyArray` is judged against that union.
// Every member is a plain property assignment: a shorthand or a spread could hide an `emptyArray` from this check, so
// one refuses instead of being read past.
function mixes(literal) {
    const other = literal.properties.find(property => !ts.isPropertyAssignment(property));
    if (other)
        throw new Error(`expected only property assignments in the inner literal at ${where(literal)}: ${other.getText(source).split('\n')[0]}`);
    const values = literal.properties.map(property => property.initializer);
    return values.some(value => ts.isIdentifier(value) && value.text === 'emptyArray')
        && values.some(ts.isArrayLiteralExpression);
}
// A call that already carries exactly its argument keeps its bytes, so a second run changes nothing; a call that
// carries any other argument refuses.
function slot(call, argument, record) {
    if (!call.typeArguments) return { ...record, edited: true, at: call.expression.end, argument };
    const carried = call.typeArguments.map(type => type.getText(source)).join(', ');
    if (carried !== argument) throw new Error(`unexpected type argument on ${record.slot} at ${record.where}: <${carried}>`);
    return { ...record, edited: false, argument };
}
const outer = entries(arrow.body, 'getScriptTargetFeatures\' body');
const edits = [slot(outer.call, outerArgument, { slot: 'outer', where: where(outer.call) })];
// Every Object.entries call the walk reads, the outer one and each inner map's, edited or not.
let read = 1;
for (const property of outer.literal.properties) {
    if (!ts.isPropertyAssignment(property))
        throw new Error(`unexpected member of getScriptTargetFeatures' literal at ${where(property)}: ${property.getText(source).split('\n')[0]}`);
    const name = property.name.getText(source);
    const map = entries(property.initializer, `property ${name}`);
    read++;
    if (mixes(map.literal)) {
        if (!mixedNames.includes(name))
            throw new Error(`property ${name} at ${where(map.call)} mixes emptyArray with an array literal, and Adamic ruled only on ${mixedNames.join(', ')}`);
        edits.push(slot(map.call, innerArgument, { slot: name, where: where(map.call) }));
    }
    else if (map.call.typeArguments)
        throw new Error(`unexpected type argument on ${name} at ${where(map.call)}, which does not mix: <${map.call.typeArguments.map(type => type.getText(source)).join(', ')}>`);
}
const mixed = edits.slice(1);
const mixedFound = mixed.map(edit => edit.slot);
if (mixedFound.length !== mixedNames.length || mixedNames.some(name => !mixedFound.includes(name)))
    throw new Error(`expected exactly the mixed maps ${mixedNames.join(', ')} in getScriptTargetFeatures, found ${mixedFound.join(', ') || 'none'}`);
// Every Object.entries in the function is one the walk read: none sits deeper or elsewhere in it.
const calls = nodes(initializer, isEntries).length;
if (calls !== read) throw new Error(`expected ${read} Object.entries calls in getScriptTargetFeatures, found ${calls}`);
// Every slot passed: only now is anything written.
let adapted = text;
for (const edit of edits.filter(edit => edit.edited).sort((a, b) => b.at - a.at))
    adapted = adapted.slice(0, edit.at) + `<${edit.argument}>` + adapted.slice(edit.at);
if (adapted !== text) {
    if (fs.readFileSync(fileName, 'utf8') !== text) throw new Error(`source changed during adaptation: ${fileName}`);
    fs.writeFileSync(fileName, adapted);
}
const printed = ({ at, argument, ...edit }) => edit;
console.log(JSON.stringify({ typescript: ts.version, task: '#r11wppy', class: 2, change: 'internal',
    function: 'getScriptTargetFeatures', where: where(declaration),
    outer: { argument: outerArgument, ...printed(edits[0]) },
    mixed: { argument: innerArgument, maps: mixedFound, edited: mixed.filter(edit => edit.edited).length,
        slots: mixed.map(printed) } }, null, 2));
