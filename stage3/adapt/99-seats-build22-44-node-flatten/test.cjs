'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const { createRequire } = require('node:module');
const { test } = require('node:test');
const [api, root, mode] = process.argv.slice(2);
assert(api && root && ['base', 'head'].includes(mode), 'usage: node test.cjs <typescript package> <adapted tree> <base|head>');
const entry = path.resolve(api, 'lib/typescript.js');
const ts = require(entry);
const target = ts.ScriptTarget.ES2022;
const source = name => fs.readFileSync(path.join(root, 'src/compiler', name + '.ts'), 'utf8');
function find(text, file, name) {
    const tree = ts.createSourceFile(file, text, target, true);
    const matches = [];
    function visit(node) {
        if (ts.isFunctionDeclaration(node) && node.name?.text === name && node.body) matches.push(node.getText(tree));
        ts.forEachChild(node, visit);
    }
    visit(tree);
    assert.equal(matches.length, 1, name);
    return matches[0];
}
function emit(text) {
    return ts.transpileModule(text.replace(/^export /, ''),
        { compilerOptions: { target, module: ts.ModuleKind.ESNext } }).outputText;
}
let bundle = fs.readFileSync(entry, 'utf8');
const splice = [
    ['utilities', 'getJSDocTypeParameterDeclarations'],
    ['utilitiesPublic', 'getEffectiveTypeParameterDeclarations'],
];
for (const [file, name] of splice) {
    const original = find(bundle, entry, name);
    assert.equal(bundle.split(original).length, 2);
    bundle = bundle.replace(original, () => emit(find(source(file), file + '.ts', name)));
}
const close = '})({ get exports()';
assert.equal(bundle.split(close).length, 2);
const helper = mode === 'base' ? '' : emit(find(source('core'), 'core.ts', 'flatMapNodeLists'));
bundle = bundle.replace(close, () => helper + '\nmodule.exports.__auditFlatten = ' +
    (mode === 'base' ? 'flatMap' : 'flatMapNodeLists') + ';\n' + close);
const mod = { exports: {} };
vm.runInThisContext('(function(module,exports,require,__filename,__dirname){' + bundle + '\n})',
    { filename: entry })(mod, mod.exports, createRequire(entry), entry, path.dirname(entry));
const c = mod.exports;
function parse(code) {
    return c.createSourceFile('audit.js', code, c.ScriptTarget.Latest, true, c.ScriptKind.JS);
}
test('FourCallbacksPreserveIndexOrderAndNodeIdentity', () => {
    const fn = parse('function a() {}\nfunction b() {}\nfunction c() {}').statements;
    const values = [undefined, [fn[0], fn[1]], [], [fn[2]]];
    const reads = [];
    const out = c.__auditFlatten(values, (value, i) => { reads.push(i); return value; });
    console.log(JSON.stringify({ mode, reads, result: out.map(x => x.name.text) }));
    assert.deepEqual(reads, [0, 1, 2, 3]);
    assert.equal(out.length, 3);
    for (let i = 0; i < 3; i++) assert.equal(out[i], fn[i]);
    assert.notEqual(out, values[1]);
});
test('EmptyAndAbsentInputsShareEmptyAndThrowsStopIteration', () => {
    for (const values of [undefined, [], [undefined], [[]]]) {
        assert.equal(c.__auditFlatten(values, x => x), c.emptyArray);
    }
    const seen = [];
    const error = new Error('callback failure');
    assert.throws(() => c.__auditFlatten([1, 2, 3], (x, i) => {
        seen.push(i);
        if (x === 2) throw error;
        return undefined;
    }), e => e === error);
    assert.deepEqual(seen, [0, 1]);
});
test('DirectTemplateOverloadAndTypedefKeepTheirParameterNodes', () => {
    const [f, g, decl] = parse('/** @template T, U */\nfunction f() {}\n' +
        '/** @template V\n * @overload\n * @returns {V}\n */\nfunction g() {}\n' +
        '/** @template W\n * @typedef {W} Alias\n */\nconst marker=1;').statements;
    const [template] = f.jsDoc[0].tags;
    const [overloadTemplate, overload] = g.jsDoc[0].tags;
    const [aliasTemplate, alias] = decl.jsDoc[0].tags;
    for (const [value, expected] of [
        [c.getJSDocTypeParameterDeclarations(f), template.typeParameters],
        [c.getEffectiveTypeParameterDeclarations(overload.typeExpression), overloadTemplate.typeParameters],
        [c.getEffectiveTypeParameterDeclarations(alias), aliasTemplate.typeParameters],
    ]) {
        console.log(JSON.stringify({ mode, names: value.map(t => t.name.text) }));
        assert.equal(value.length, expected.length);
        value.forEach((x, i) => assert.equal(x, expected[i]));
    }
});
