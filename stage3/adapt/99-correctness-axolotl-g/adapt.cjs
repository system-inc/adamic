'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require('typescript');

const beforeValue = 'const value = (options[setting] ?? defaultValue) as PresetValue;';
const afterValue = 'const value = options[setting];';
const formatted = 'const formattedValue = hasProperty(options, setting) && value === undefined ? "null" : formatValueOrArray(setting, (value ?? defaultValue) as PresetValue);';
const templates = [
    '`${tab}${tab}// "${setting}": ${formatValueOrArray(setting, value)},`',
    '`${tab}${tab}"${setting}": ${formatValueOrArray(setting, value)},`',
];
function nodes(root, predicate) {
    const found = [];
    (function visit(node) { if (predicate(node)) found.push(node); ts.forEachChild(node, visit); })(root);
    return found;
}
function locate(fileName, text) {
    const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, 'cannot parse commandLineParser.ts');
    const generators = nodes(source, n => ts.isFunctionDeclaration(n) && n.name?.text === 'generateTSConfig');
    assert.equal(generators.length, 1, 'generateTSConfig moved');
    const emitters = nodes(source, n => ts.isFunctionDeclaration(n) && n.name?.text === 'emitOption');
    assert.equal(emitters.length, 1, 'emitOption count drift');
    const emitter = emitters[0];
    assert.equal(emitter.parent, generators[0].body, 'emitOption left generateTSConfig');
    const statements = emitter.body.statements;
    const values = statements.filter(n => ts.isVariableStatement(n) && n.declarationList.declarations.some(d => d.name.getText(source) === 'value'));
    assert.equal(values.length, 1, 'value declaration count drift');
    const value = values[0];
    const shape = value.getText(source) === beforeValue ? 'before' : value.getText(source) === afterValue ? 'after' : undefined;
    assert(shape, 'value declaration is not an inventoried spelling');
    const formattedNodes = nodes(emitter, n => ts.isIdentifier(n) && n.text === 'formattedValue');
    assert.equal(formattedNodes.length, shape === 'before' ? 0 : 3, 'file is partly adapted');
    assert.equal(statements.length, shape === 'before' ? 6 : 7, 'emitOption statement count drift');
    assert.equal(statements[4], value, 'value declaration moved');
    if (shape === 'after') assert.equal(statements[5].getText(source), formatted, 'formatted declaration moved');
    const branch = statements[shape === 'before' ? 5 : 6];
    assert(ts.isIfStatement(branch) && branch.expression.getText(source) === 'comment' && branch.elseStatement, 'comment branch moved');
    const outputs = nodes(branch, n => ts.isTemplateExpression(n));
    assert.equal(outputs.length, 2, 'output template count drift');
    for (let i = 0; i < outputs.length; i++) {
        const expected = shape === 'before' ? templates[i] : templates[i].replace('formatValueOrArray(setting, value)', 'formattedValue');
        assert.equal(outputs[i].getText(source), expected, 'output template spelling drift');
    }
    const calls = nodes(emitter, n => ts.isCallExpression(n) && n.expression.getText(source) === 'formatValueOrArray');
    assert.equal(calls.length, shape === 'before' ? 2 : 1, 'formatting call count drift');
    return {shape, source, value, outputs};
}
function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const fileName = path.join(tree, 'src/compiler/commandLineParser.ts');
    const text = fs.readFileSync(fileName, 'utf8');
    const site = locate(fileName, text);
    const edits = [];
    if (site.shape === 'before') {
        const gap = text.slice(site.value.end, site.outputs[0].getStart(site.source));
        const newline = gap.startsWith('\r\n') ? '\r\n' : '\n';
        edits.push({start: site.value.getStart(site.source), end: site.value.end, text: afterValue + newline + '        ' + formatted});
        for (const output of site.outputs) edits.push({start: output.getStart(site.source), end: output.end, text: output.getText(site.source).replace('formatValueOrArray(setting, value)', 'formattedValue')});
    }
    let result = text;
    for (const edit of edits.sort((a, b) => b.start - a.start)) result = result.slice(0, edit.start) + edit.text + result.slice(edit.end);
    assert.equal(locate(fileName, result).shape, 'after', 'planned result is not fully adapted');
    assert.equal(fs.readFileSync(fileName, 'utf8'), text, 'source changed while planning');
    if (edits.length) fs.writeFileSync(fileName, result);
    return {files: edits.length ? 1 : 0, edits: edits.length, javascriptChanging: true, behaviorChanging: true};
}
module.exports = {adapt, beforeValue, afterValue, formatted, templates};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
