'use strict';

const assert = require('node:assert/strict');
const path = require('node:path');

function probe(ts) {
    const source = ts.createSourceFile('test.ts', '/** @deprecated hello */\nfunction f() {}',
        ts.ScriptTarget.Latest, true);
    const original = source.statements[0];
    assert(ts.isFunctionDeclaration(original));
    assert(original.jsDoc);
    assert.equal(Object.hasOwn(original.jsDoc, "jsDocCache"), false);
    const clone = ts.factory.cloneNode(original);
    assert.notEqual(clone, original);
    assert.equal(clone.jsDoc, original.jsDoc, 'clone must share the original array before either read');
    const first = ts.getJSDocTags(original);
    assert.equal(first.length, 1);
    assert.equal(first[0].kind, ts.SyntaxKind.JSDocDeprecatedTag);
    assert.equal(first[0].comment, 'hello');
    const second = ts.getJSDocTags(clone);
    assert.deepEqual(second, first);
    assert.equal(second, first, 'both nodes must read the same cached tag array');
    return [first[0].kind, first[0].comment];
}

const results = process.argv.slice(2).map(file => probe(require(path.resolve(file))));
assert(results.length > 0, 'pass one or more built typescript.js files');
for (const result of results) assert.deepEqual(result, results[0]);
console.log(JSON.stringify(results));
