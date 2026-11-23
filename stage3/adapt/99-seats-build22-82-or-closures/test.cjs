'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const { before, inspect } = require('./adapt.cjs');

assert.equal(process.argv.length, 3, 'usage: node test.cjs <adapted tree>');
assert.equal(ts.version, '6.0.3');
const tree = process.argv[2];
const file = path.join(tree, 'src/compiler/core.ts');
const core = ts.createSourceFile(file, fs.readFileSync(file, 'utf8'), ts.ScriptTarget.Latest, true);
const implementations = core.statements.filter(node =>
    ts.isFunctionDeclaration(node) && node.name?.text === 'or' && node.body);
assert.equal(implementations.length, 1);
const options = { compilerOptions: { target: ts.ScriptTarget.ES2020 } };
const implementation = implementations[0].getText(core).replace(/^export /, '');
const or = new Function(ts.transpileModule(implementation, options).outputText + '\nreturn or;')();

const { current, adapted } = inspect(tree);
assert(adapted, 'expected the adapted utilities closure');
const names = ['hasJSFileExtension', 'hasTSFileExtension'];
const original = new Function('or', ...names, `return ${before};`);
const closure = new Function(...names, ts.transpileModule(`const callback = ${current};`, options)
    .outputText + '\nreturn callback;');

function check(results, value, called) {
    const input = 'source.ts';
    const calls = [];
    const stubs = names.map((name, index) => argument => {
        calls.push([name, argument]);
        return results[index];
    });
    const want = called.map(name => [name, input]);
    const originalValue = original(or, ...stubs)(input);
    assert.equal(originalValue, value);
    assert.deepEqual(calls, want);
    calls.length = 0;
    assert.equal(closure(...stubs)(input), originalValue);
    assert.deepEqual(calls, want);
}

check([true, true], true, [names[0]]);
check([false, true], true, names);
check([false, 0], 0, names);
const truthy = { value: 'truthy' };
check([truthy, true], truthy, [names[0]]);
console.log('utilities.ts: 4 cases, same result and ordered calls');

const { checkerBefore, inspectChecker } = require('./adapt.cjs');
const checker = inspectChecker(tree);
assert(checker.adapted, 'expected the adapted checker closure');
const checkerNames = [
    'isPropertyDeclaration', 'isAccessor', 'isVariableDeclaration', 'isPropertySignature',
    'isBinaryExpression', 'isPropertyAccessExpression',
];
const checkerOriginal = new Function('or', ...checkerNames, `return ${checkerBefore};`);
const checkerClosure = new Function(...checkerNames, ts.transpileModule(`const callback = ${checker.current};`, options)
    .outputText + '\nreturn callback;');

function checkChecker(results, expected, count) {
    const input = { kind: 'declaration' };
    const calls = [];
    const stubs = checkerNames.map((name, index) => argument => {
        calls.push([name, argument]);
        return results[index];
    });
    const want = checkerNames.slice(0, count).map(name => [name, input]);
    const originalValue = checkerOriginal(or, ...stubs)(input);
    assert.equal(originalValue, expected);
    assert.deepEqual(calls, want);
    calls.length = 0;
    assert.equal(checkerClosure(...stubs)(input), originalValue);
    assert.deepEqual(calls, want);
}

for (let first = 0; first < checkerNames.length; first++) {
    checkChecker(checkerNames.map((_, index) => index >= first), true, first + 1);
}
checkChecker([false, false, false, false, false, 0], 0, 6);
checkChecker([false, truthy, true, true, true, true], truthy, 2);
console.log('checker.ts: 8 cases, same result and ordered calls');
