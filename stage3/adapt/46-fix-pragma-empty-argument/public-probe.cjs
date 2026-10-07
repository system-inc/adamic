"use strict";

const assert = require("node:assert/strict");
const before = require(process.argv[2]);
const after = require(process.argv[3]);
assert.equal(before.version, "6.0.3");
assert.equal(after.version, "6.0.3");
function observe(compiler, input) {
    try {
        const source = compiler.createSourceFile("empty.ts", input, compiler.ScriptTarget.Latest);
        return { referencedFiles: source.referencedFiles, typeReferenceDirectives: source.typeReferenceDirectives,
            libReferenceDirectives: source.libReferenceDirectives, diagnostics: source.parseDiagnostics.map(diagnostic => diagnostic.code) };
    }
    catch (error) { return { error: error.name + ": " + error.message }; }
}
const empty = "/// <reference path='' />";
const old = observe(before, empty);
const fixed = observe(after, empty);
assert.match(old.error, /^TypeError: Cannot read properties of undefined \(reading 'length'\)$/);
assert.deepEqual(fixed.referencedFiles.map(file => file.fileName), [""]);
assert.equal(fixed.referencedFiles[0].end - fixed.referencedFiles[0].pos, 0);
assert.deepEqual(fixed.diagnostics, []);
console.log("input: " + JSON.stringify(empty));
console.log("before: " + JSON.stringify(old));
console.log("after: " + JSON.stringify(fixed));
for (const input of [
    '/// <reference path="" />', "/// <reference path='x.ts' />", '/// <reference path="x.ts" />',
    "/// <reference path='a b.ts' />", "/// <reference types='node' />", "/// <reference lib='es2020' />",
    "/// <reference />", "// not a reference", "/* @jsx h */", "/* @jsxruntime automatic */",
]) {
    assert.deepEqual(observe(after, input), observe(before, input), input);
    console.log("unchanged: " + JSON.stringify(input));
}
// This comparison would fail with the original compiler put back as the fixed one.
assert.throws(() => assert.deepEqual(observe(before, empty), fixed), assert.AssertionError);
console.log("PASS: real compiler public API accepts the empty argument; controls unchanged; original-parser mutant caught");
