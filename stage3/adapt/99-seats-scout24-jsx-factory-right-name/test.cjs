'use strict';

const assert = require('node:assert/strict');
const { execFileSync } = require('node:child_process');
const ts = require('typescript');
const { relative, findOwner, plan } = require('./adapt.cjs');

assert.equal(process.argv.length, 3, 'usage: node test.cjs <applied tree>');
const tree = process.argv[2];
const adapterRef = 'refs/stage3/99-seats-scout24-jsx-factory-right-name';
function snapshot(side) {
    return execFileSync('git', ['-C', tree, 'show', adapterRef + '-' + side + ':' + relative],
        { encoding: 'utf8', maxBuffer: 4 * 1024 * 1024 });
}
const before = snapshot('before');
const after = snapshot('after');

function compile(text) {
    const owners = [findOwner(text, 'createReactNamespace'), findOwner(text)];
    const js = ts.transpileModule(owners.map(owner => owner.getText()).join('\n'), {
        compilerOptions: { target: ts.ScriptTarget.ES2020 },
    }).outputText;
    return new Function('parseNodeFactory', 'setParent', 'getParseTreeNode', 'isQualifiedName', 'idText',
        js + '\nreturn createJsxFactoryExpressionFromEntityName;');
}

const names = [
    'React.createElement', 'a.b.c', 'h.__x', 'h.___x', '__a.__b', '_a._b',
    'h.default', 'h.await', 'h.constructor', 'h.__proto__', String.raw`h.\u0061b`, 'h.__',
];
const cases = names.map(name => ({
    name,
    entity: () => ts.parseIsolatedEntityName(name, ts.ScriptTarget.ES2020),
}));
for (const namespace of ['React', '__fallback']) {
    cases.push({
        name: 'synthesized ' + namespace + '.createElement',
        entity: () => ts.factory.createQualifiedName(ts.factory.createIdentifier(namespace), 'createElement'),
    });
}

function chain(node) {
    if (ts.isPropertyAccessExpression(node)) {
        return [...chain(node.expression), ...chain(node.name)];
    }
    assert(ts.isIdentifier(node));
    return [{ escapedText: node.escapedText, text: ts.idText(node) }];
}

function run(instantiate, scenario) {
    const calls = [];
    const recording = {
        createIdentifier(text) {
            const identifier = ts.factory.createIdentifier(text);
            calls.push(['createIdentifier', text, identifier.escapedText]);
            return identifier;
        },
        createPropertyAccessExpression(left, right) {
            calls.push(['createPropertyAccessExpression', chain(left), chain(right)]);
            return ts.factory.createPropertyAccessExpression(left, right);
        },
    };
    const source = ts.createSourceFile('case.tsx', '<div />;', ts.ScriptTarget.ES2020, true, ts.ScriptKind.TSX);
    const parent = source.statements[0].expression;
    const entity = scenario.entity();
    assert(entity && ts.isQualifiedName(entity), scenario.name);
    const factory = instantiate(recording, ts.setParent, ts.getParseTreeNode, ts.isQualifiedName, ts.idText);
    const result = factory(recording, entity, parent);
    return {
        chain: chain(result),
        printed: ts.createPrinter().printNode(ts.EmitHint.Expression, result, source),
        calls,
    };
}

const original = compile(before);
const adapted = compile(after);
for (const scenario of cases) {
    const expected = run(original, scenario);
    assert.deepEqual(run(adapted, scenario), expected, scenario.name);
    console.log('PASS ' + scenario.name + ' ' + JSON.stringify(expected));
}
assert.equal(plan(before), after);
assert.equal(plan(after), after);
