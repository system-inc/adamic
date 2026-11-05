'use strict';
const fs = require('node:fs'), path = require('node:path'), vm = require('node:vm'), assert = require('node:assert/strict');
const adapter = require('./adapt.cjs'), ts = require('typescript');
const args = process.argv.slice(2), mutant = args[0] === '--mutant';
if (mutant) args.shift();
assert.equal(args.length, 3, 'usage: witnesses.cjs [--mutant] BEFORE AFTER OUTPUT');
const [before, after, output] = args.map(value => path.resolve(value));
const stock = require(path.join(before, 'built/local/typescript.js'));
function extract(tree, file, name) {
    const text = fs.readFileSync(path.join(tree, 'src/compiler', file), 'utf8');
    const source = adapter.parse(file, text);
    return source.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === name).map(node => node.getText(source)).join('\n').replaceAll('export function', 'function');
}
function build(tree, mutate) {
    let builder = ['convertOrRepopulateDiagnosticMessageChain', 'convertOrRepopulateDiagnosticMessageChainArray'].map(name => extract(tree, 'builder.ts', name)).join('\n');
    if (mutate) {
        const old = 'Debug.assert(isFullChain(chain));';
        assert(builder.includes(old));
        builder = builder.replace(old, '').replace('return chain;', 'return chain as DiagnosticMessageChain;');
    }
    return vm.runInNewContext(adapter.transpile(extract(tree, 'core.ts', 'sameMap') + '\n' + builder)
        + '\n({convert:convertOrRepopulateDiagnosticMessageChain, array:convertOrRepopulateDiagnosticMessageChainArray})', {
        Debug: stock.Debug, createModeMismatchDetails: stock.createModeMismatchDetails, createModuleNotFoundChain: stock.createModuleNotFoundChain,
    });
}
const source = {fileName: '/project/input.ts'};
const program = {getResolvedModule: () => undefined, typesPackageExists: () => false, packageBundlesTypes: () => false};
function chain(text = 'ordinary', next) { return {messageText: text, category: 1, code: 1234, next}; }
function serialize(value) {
    return JSON.parse(JSON.stringify(value, (key, val) => typeof val === 'function' ? {returns: val()} : val));
}
function cases(api) {
    const rows = [];
    for (const name of ['unchanged', 'unchanged-nested', 'zero-fields', 'full-repopulate-mode', 'full-repopulate-module', 'repopulate-mode', 'repopulate-module', 'parent-of-repopulated', 'mixed-array', 'empty-array', 'undefined-array']) {
        const leaf = chain('leaf'); let input;
        if (name === 'unchanged') input = chain();
        if (name === 'unchanged-nested') input = chain('parent', [leaf]);
        if (name === 'zero-fields') input = {messageText: '', category: 0, code: 0};
        if (name === 'full-repopulate-mode') input = {...chain('mode'), repopulateInfo: () => true};
        if (name === 'full-repopulate-module') input = {...chain('module'), repopulateInfo: () => ({moduleReference: 'missing', mode: undefined, packageName: undefined})};
        if (name === 'repopulate-mode') input = {info: true, next: [leaf]};
        if (name === 'repopulate-module') input = {info: {moduleReference: 'missing', mode: undefined, packageName: undefined}, next: [leaf]};
        if (name === 'parent-of-repopulated') input = chain('parent', [{info: true}]);
        if (name === 'mixed-array') input = [leaf, {info: true}];
        if (name === 'empty-array') input = [];
        const calls = [];
        const info = value => { calls.push('info' in value ? serialize(value.info) : value.messageText); return value.info || value.repopulateInfo?.(); };
        const result = name.endsWith('array') ? api.array(input, source, program, info) : api.convert(input, source, program, info);
        rows.push({name, output: result === undefined ? 'undefined' : serialize(result), same: result === input, sameNext: input && !Array.isArray(input) ? result.next === input.next : undefined, calls});
    }
    return rows;
}
const oldApi = build(before, false), newApi = build(after, mutant), oldRows = cases(oldApi), newRows = cases(newApi);
assert.deepEqual(newRows, oldRows);
const invalid = [];
for (const missing of [['messageText'], ['category'], ['code'], ['messageText', 'category', 'code']]) {
    const value = chain(); for (const key of missing) delete value[key];
    function outcome(api) { try { return {returnedOriginal: api.convert(value, source, program, () => undefined) === value}; } catch (error) { return {error: error.message}; } }
    const original = outcome(oldApi), adapted = outcome(newApi);
    invalid.push({missing, before: original, after: adapted, guarded: !!adapted.error});
}
const coreBefore = fs.readFileSync(path.join(before, 'src/compiler/core.ts'), 'utf8'), coreAfter = fs.readFileSync(path.join(after, 'src/compiler/core.ts'), 'utf8');
const jsBefore = adapter.transpile(coreBefore), jsAfter = adapter.transpile(coreAfter);
assert.equal(jsBefore, jsAfter);
const prefix = output.replace(/\.json$/, '');
fs.writeFileSync(prefix + '-core-before.js', jsBefore); fs.writeFileSync(prefix + '-core-after.js', jsAfter);
const result = {mutant, validOutputsIdentical: true, beforeRows: oldRows, afterRows: newRows, invalid, coreJavascriptIdentical: true, coreJavascriptSha256: adapter.digest(jsBefore), pass: invalid.every(row => row.guarded)};
fs.writeFileSync(output, JSON.stringify(result, null, 2) + '\n');
console.log(JSON.stringify({mutant, validOutputsIdentical: true, invalidGuarded: invalid.filter(row => row.guarded).length, pass: result.pass}));
process.exitCode = result.pass ? 0 : 1;
