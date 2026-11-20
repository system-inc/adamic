'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const vm = require('node:vm');
const ts = require('typescript');
assert.equal(ts.version, '6.0.3');
const [before, after] = process.argv.slice(2);
const names = ['getInferTypeParameters', 'getDeclaredTypeOfSymbol', 'tryGetDeclaredTypeOfSymbol', 'getDeclaredTypeOfTypeParameter'];
function extract(root, file, names) {
    const text = fs.readFileSync(root + '/src/compiler/' + file, 'utf8');
    const ast = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    const found = new Map();
    function visit(node) {
        if (ts.isFunctionDeclaration(node) && node.body && names.includes(node.name?.text)) {
            assert(!found.has(node.name.text));
            found.set(node.name.text, node.getText(ast));
        }
        ts.forEachChild(node, visit);
    }
    visit(ast);
    return names.map(name => {
        assert(found.has(name), name);
        return found.get(name);
    }).join('\n');
}
const links = new Map();
let created = [];
const env = {
    SymbolFlags: ts.SymbolFlags,
    getSymbolLinks(s) {
        if (!links.has(s)) links.set(s, {});
        return links.get(s);
    },
    createTypeParameter(s) {
        created.push(s.name);
        return s.expected;
    },
};
const compile = root => ts.transpileModule(extract(root, 'checker.ts', names) + '\n' + extract(root, 'core.ts', ['append']).replace(/^export /, ''), { compilerOptions: { target: ts.ScriptTarget.ES2024 } }).outputText;
const functions = [before, after].map(root => vm.runInNewContext(compile(root) + '\ngetInferTypeParameters', env));
const first = { name: 'first', flags: ts.SymbolFlags.TypeParameter, expected: {} };
const second = { name: 'second', flags: ts.SymbolFlags.TypeParameter | ts.SymbolFlags.Transient, expected: {} };
const skipped = { name: 'skipped', flags: ts.SymbolFlags.TypeAlias, expected: {} };
const locals = new Map([[first.name, first], [skipped.name, skipped], [second.name, second]]);
for (const fn of functions) {
    links.clear();
    created = [];
    assert.equal(fn({}), undefined);
    assert.equal(fn({ locals: new Map([[skipped.name, skipped]]) }), undefined);
    const result = fn({ locals });
    assert.equal(result.length, 2);
    assert.equal(result[0], first.expected);
    assert.equal(result[1], second.expected);
    assert.deepEqual(created, ['first', 'second']);
    const cached = fn({ locals });
    assert.equal(cached[0], result[0]);
    assert.equal(cached[1], result[1]);
    assert.deepEqual(created, ['first', 'second']);
}
console.log(JSON.stringify({ noLocals: true, skippedNonParameter: true, mixedFlags: second.flags, identicalParameterValues: true, cachedValuesReused: true }));
