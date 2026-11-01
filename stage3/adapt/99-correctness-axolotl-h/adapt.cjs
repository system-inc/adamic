'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require('typescript');

const prerequisite = 'const formattedValue = hasProperty(options, setting) && value === undefined ? "null" : formatValueOrArray(setting, (value ?? defaultValue) as PresetValue);';
const headers = [
    ['formatValueOrArray', 'function formatValueOrArray(settingName: string, value: PresetValue): string ', 'function formatValueOrArray(settingName: string, value: PresetValue): string | undefined '],
    ['formatSingleValue', 'function formatSingleValue(value: PresetValue, map: Map<string, string | number> | undefined) ', 'function formatSingleValue(value: PresetValue, map: Map<string, string | number> | undefined): string | undefined '],
];
const arrayBefore = 'return `[${value.map(v => formatSingleValue(v, map)).join(", ")}]`;';
const arrayAfter = 'return `[${String(value.map(v => formatSingleValue(v, map)).join(", "))}]`;';
const templates = ['`${tab}${tab}// "${setting}": ${formattedValue},`', '`${tab}${tab}"${setting}": ${formattedValue},`'];
function nodes(root, predicate) {
    const found = [];
    (function visit(node) { if (predicate(node)) found.push(node); ts.forEachChild(node, visit); })(root);
    return found;
}
function locate(fileName, text) {
    const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, 'cannot parse commandLineParser.ts');
    function named(name) {
        const found = nodes(source, n => ts.isFunctionDeclaration(n) && n.name?.text === name);
        assert.equal(found.length, 1, `${name} count drift`);
        return found[0];
    }
    const generator = named('generateTSConfig'), emitter = named('emitOption');
    assert.equal(emitter.parent, generator.body, 'emitOption left generateTSConfig');
    const boundary = nodes(emitter, n => ts.isVariableStatement(n) && n.getText(source) === prerequisite);
    assert.equal(boundary.length, 1, '99h requires the exact 99g formatter declaration');
    assert.equal(emitter.body.statements.length, 7, '99g emitOption statement count drift');
    const regions = [];
    for (const [name, before, after] of headers) {
        const fn = named(name);
        assert.equal(fn.parent, generator.body, `${name} left generateTSConfig`);
        regions.push({start: fn.getStart(source), end: fn.body.getStart(source), before, after});
    }
    const formatter = named('formatValueOrArray');
    const calls = nodes(formatter, n => ts.isCallExpression(n) && n.expression.getText(source) === 'formatSingleValue');
    assert.equal(calls.length, 2, 'formatSingleValue call count drift');
    assert.equal(calls[0].getText(source), 'formatSingleValue(v, map)', 'array callback moved');
    assert.equal(calls[1].getText(source), 'formatSingleValue(value, map)', 'scalar return moved');
    assert.equal(calls[1].parent.getText(source), 'return formatSingleValue(value, map);', 'scalar return was asserted or changed');
    const arrays = nodes(formatter, n => ts.isReturnStatement(n) && ts.isTemplateExpression(n.expression));
    assert.equal(arrays.length, 1, 'array wrapper count drift');
    regions.push({start: arrays[0].getStart(source), end: arrays[0].end, before: arrayBefore, after: arrayAfter});
    const outputs = nodes(emitter, n => ts.isTemplateExpression(n));
    assert.equal(outputs.length, 2, 'config template count drift');
    for (let i = 0; i < outputs.length; i++) regions.push({start: outputs[i].getStart(source), end: outputs[i].end, before: templates[i], after: templates[i].replace('${formattedValue}', '${String(formattedValue)}')});
    const states = regions.map(region => {
        const spelling = text.slice(region.start, region.end);
        assert(spelling === region.before || spelling === region.after, 'site is not an inventoried spelling');
        return spelling === region.after;
    });
    assert(states.every(state => state === states[0]), 'file is partly adapted');
    return {adapted: states[0], regions};
}
function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const fileName = path.join(tree, 'src/compiler/commandLineParser.ts');
    const text = fs.readFileSync(fileName, 'utf8');
    const site = locate(fileName, text);
    const edits = site.adapted ? [] : site.regions;
    let result = text;
    for (const edit of edits.sort((a, b) => b.start - a.start)) result = result.slice(0, edit.start) + edit.after + result.slice(edit.end);
    assert(locate(fileName, result).adapted, 'planned result is not fully adapted');
    assert.equal(fs.readFileSync(fileName, 'utf8'), text, 'source changed while planning');
    if (edits.length) fs.writeFileSync(fileName, result);
    return {files: edits.length ? 1 : 0, edits: edits.length, javascriptChanging: true, behaviorChanging: false};
}
module.exports = {adapt, prerequisite, headers, arrayBefore, arrayAfter, templates};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
