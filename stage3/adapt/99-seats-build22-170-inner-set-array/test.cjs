'use strict';

const assert = require('node:assert/strict');
const { execFileSync } = require('node:child_process');
const path = require('node:path');
const vm = require('node:vm');
const ts = require('typescript');
const { declaration, parse, plan, relativeFile } = require('./adapt.cjs');

assert.equal(process.argv.length, 3, 'usage: node test.cjs APPLIED_TREE');
const applied = path.resolve(process.argv[2]);
const ref = 'refs/stage3/99-seats-build22-170-inner-set-array';
function snapshot(side) {
    return execFileSync('git', ['-C', applied, 'show', `${ref}-${side}:${relativeFile}`], { encoding: 'utf8', maxBuffer: 8 * 1024 * 1024 });
}
const [base, head] = ['before', 'after'].map(snapshot);
assert.deepEqual(plan(base), { text: head, edits: 1 });
assert.deepEqual(plan(head), { text: head, edits: 0 });

const plain = 'unprefixedNodeCoreModulesList';
const exclusive = 'exclusivelyPrefixedNodeCoreModules';
const [stock, adapted] = [base, head].map(text => {
    const source = parse(text);
    return {
        lists: [plain, exclusive].map(name => `const ${declaration(source, name).getText(source)};`).join('\n'),
        initializer: `const ${declaration(source, 'nodeCoreModules').getText(source)};`,
    };
});
assert.equal(adapted.lists, stock.lists);

function order(lists, initializer) {
    const { outputText } = ts.transpileModule(`${lists}\n${initializer}\n[...nodeCoreModules];`, {
        compilerOptions: { target: ts.ScriptTarget.ES2024 },
    });
    return Array.from(vm.runInNewContext(outputText));
}

const realOrder = order(stock.lists, stock.initializer);
assert.deepEqual(order(adapted.lists, adapted.initializer), realOrder);
console.log(`Real adapted lists: ${realOrder.length} elements matched in insertion order`);

const witness = `
const ${plain}: string[] = ["fs"];
const ${exclusive}: Set<string> = new Set(["node:z", "node:fs", "node:a", "node:m"]);
`;
const expected = ['fs', 'node:fs', 'node:z', 'node:a', 'node:m'];
assert.deepEqual(order(witness, stock.initializer), expected);
assert.deepEqual(order(witness, adapted.initializer), expected);
console.log(`Order and de-duplication witness: ${expected.join(',')}`);
