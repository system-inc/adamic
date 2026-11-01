'use strict';
// Node witnesses for 96b, before against after: the real enable and reader functions of transformer.ts, taken from each
// tree by their names, transpiled as stock TypeScript 6.0.3 emits them, and run on Node over a sparse feature array.
// Each scenario enables flags on a kind whose bit is unset, on one whose same bit is already set, and on one holding the
// other bit; it then reads every kind back through both readers and records the array's own keys and values.
// Usage: node witness.cjs <tree before 96b> <tree after 96b>
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const vm = require('node:vm');
const ts = require('typescript');

const file = 'src/compiler/transformer.ts';
const names = ['enableSubstitution', 'isSubstitutionEnabled', 'enableEmitNotification', 'isEmitNotificationEnabled'];
const scenarios = [
    {name: 'unset bit', calls: [['enableSubstitution', 1], ['enableEmitNotification', 2]]},
    {name: 'same bit already set', calls: [['enableSubstitution', 1], ['enableSubstitution', 1], ['enableEmitNotification', 2], ['enableEmitNotification', 2]]},
    {name: 'other bit already set', calls: [['enableEmitNotification', 1], ['enableSubstitution', 1], ['enableSubstitution', 2], ['enableEmitNotification', 2]]},
];

function program(tree) {
    const text = fs.readFileSync(path.join(tree, file), 'utf8');
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    const found = {};
    (function visit(node) {
        if ((ts.isFunctionDeclaration(node) && names.includes(node.name?.text)) || (ts.isEnumDeclaration(node) && node.name.text === 'SyntaxKindFeatureFlags')) {
            found[node.name.text] = node.getText(source).replace(/^const enum /, 'enum ');
        }
        ts.forEachChild(node, visit);
    })(source);
    for (const name of [...names, 'SyntaxKindFeatureFlags']) assert(found[name], `${name} moved`);
    // The functions read only the array, state, Debug.assert, getEmitFlags and EmitFlags; each is given the value
    // transformNodes would hold during transformation, so the scenarios reach the write.
    const prelude = `enum TransformationState { Uninitialized, Initialized, Completed, Disposed }
const state = TransformationState.Initialized;
const Debug = { assert(condition: boolean) { if (!condition) throw new Error("assert"); } };
const EmitFlags = { AdviseOnEmitNode: 1 << 19 };
const getEmitFlags = (_node: unknown) => 0;
const enabledSyntaxKindFeatures = new Array<number>(4);
`;
    const body = prelude + found.SyntaxKindFeatureFlags + '\n' + names.map(name => found[name]).join('\n') +
        `\nexports.run = (calls: [string, number][]) => {
    // A fresh sparse array per scenario: every slot a hole again, as transformNodes starts.
    enabledSyntaxKindFeatures.length = 0;
    enabledSyntaxKindFeatures.length = 4;
    const table: Record<string, (value: any) => any> = { ${names.join(', ')} };
    for (const [name, kind] of calls) table[name](kind);
    const read = [0, 1, 2, 3].map(kind => [isSubstitutionEnabled({ kind } as any), isEmitNotificationEnabled({ kind } as any)]);
    return { keys: Object.keys(enabledSyntaxKindFeatures), values: [...enabledSyntaxKindFeatures], read };
};`;
    const emitted = ts.transpileModule(body, {compilerOptions: {target: ts.ScriptTarget.ES2020, module: ts.ModuleKind.CommonJS}}).outputText;
    const context = {exports: {}};
    vm.runInNewContext(emitted, context);
    return {emitted, run: context.exports.run, enable: [found.enableSubstitution, found.enableEmitNotification]};
}

const [, , beforeTree, afterTree] = process.argv;
assert(beforeTree && afterTree, 'usage: node witness.cjs <tree before 96b> <tree after 96b>');
assert.equal(ts.version, '6.0.3');
const before = program(path.resolve(beforeTree)), after = program(path.resolve(afterTree));
assert.notEqual(before.emitted, after.emitted, 'the two trees emit the same JavaScript, so this witnesses nothing');
const results = scenarios.map(scenario => {
    const b = before.run(scenario.calls), a = after.run(scenario.calls);
    assert.deepEqual(JSON.parse(JSON.stringify(a)), JSON.parse(JSON.stringify(b)), `${scenario.name}: after differs from before`);
    return {scenario: scenario.name, calls: scenario.calls, result: b};
});
console.log(JSON.stringify({typescript: ts.version, node: process.version, before: before.enable, after: after.enable, scenarios: results, agree: true}, null, 2));
