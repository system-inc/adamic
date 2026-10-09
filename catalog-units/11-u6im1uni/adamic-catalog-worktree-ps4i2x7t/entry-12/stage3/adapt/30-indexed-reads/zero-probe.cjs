#!/usr/bin/env node
"use strict";
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const assert = require("node:assert/strict");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
if (ts.version !== "6.0.3") throw new Error(`want TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 4) throw new Error("usage: node zero-probe.cjs <before-tree> <after-tree>");
function extract(tree) {
    const file = path.join(tree, "src/compiler/utilities.ts");
    const text = fs.readFileSync(file, "utf8");
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    const names = new Set(["getExpandedCharCodes", "convertToBase64", "getStringFromExpandedCharCodes", "base64decode"]);
    const spans = source.statements.filter(node => ts.isFunctionDeclaration(node) && names.has(node.name?.text) ||
        ts.isVariableStatement(node) && node.declarationList.declarations.some(declaration => declaration.name.getText(source) === "base64Digits"));
    assert.equal(spans.length, 5);
    const code = ts.transpileModule(spans.map(node => node.getText(source)).join("\n"), {
        compilerOptions: { target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.CommonJS },
    }).outputText;
    const context = vm.createContext({ exports: {}, Debug: { assert: condition => assert.ok(condition) } });
    vm.runInContext(code, context);
    return context;
}
const before = extract(process.argv[2]), after = extract(process.argv[3]);
const inputs = ["", "a", "ab", "abc", "abcd", "abcdef", "\0", "\0\0", "\0\0\0", "é", "中", "aé中"];
for (const input of inputs) {
    const expected = Buffer.from(input, "utf8").toString("base64");
    assert.equal(before.convertToBase64(input), expected);
    assert.equal(after.convertToBase64(input), expected);
    assert.equal(before.base64decode(undefined, expected), input);
    assert.equal(after.base64decode(undefined, expected), input);
}
console.log(`PASS ${inputs.length} source-extracted encode/decode cases against Node Buffer, including all tail lengths and decoder exhaustion`);
