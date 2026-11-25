'use strict';

const assert = require('node:assert/strict');
const { execFileSync } = require('node:child_process');
const ts = require('typescript');
const { argumentType, plan, relative } = require('./adapt.cjs');

assert.equal(process.argv.length, 3, 'usage: node test.cjs <tree produced by stage3/apply.sh>');
const tree = process.argv[2];
const ref = 'refs/stage3/99-seats-build22-104-require-readonly-tuple';

function snapshot(side) {
    return execFileSync('git', ['-C', tree, 'show', `${ref}-${side}:${relative}`],
        { encoding: 'utf8', maxBuffer: 4 * 1024 * 1024 });
}

const before = snapshot('before');
const after = snapshot('after');
const args = argumentType(after);
assert(ts.isTypeOperatorNode(args) && args.operator === ts.SyntaxKind.ReadonlyKeyword,
    'RequireOrImportCall arguments must be readonly');
assert(ts.isTupleTypeNode(args.type) && args.type.elements.length === 1);
assert.equal(args.type.elements[0].getText(), 'StringLiteralLike');
assert.equal(plan(before), after);
assert.equal(plan(after), after);
console.log('PASS readonly single-string-literal tuple and exact snapshot replay');
