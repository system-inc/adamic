#!/usr/bin/env node
"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const ts = require("typescript");

assert.equal(ts.version, "6.0.3");
assert.equal(process.argv.length, 3, "usage: node adapt.cjs <tree>");

const filename = path.join(process.argv[2], "src/compiler/sourcemap.ts");
const text = fs.readFileSync(filename, "utf8");
const source = ts.createSourceFile(filename, text, ts.ScriptTarget.Latest, true);
const newline = text.includes("\r\n") ? "\r\n" : "\n";

function namedFunction(statements, name) {
    const matches = statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === name);
    assert.equal(matches.length, 1, `expected one ${name} function`);
    assert.ok(matches[0].body, `expected a body for ${name}`);
    return matches[0];
}

const decoder = namedFunction(source.statements, "decodeMappings");
const completion = namedFunction(decoder.body.statements, "stopIterating");
const snapshot = (type, value) => [
    `function stopIterating(): { value: ${type}; done: true; } {`,
    "        done = true;",
    `        return { value: ${value}, done: true };`,
    "    }",
].join(newline);
const original = snapshot("never", "undefined!");
const adapted = snapshot("undefined", "undefined");
const current = completion.getText(source);

if (current === original) {
    const start = completion.getStart(source);
    fs.writeFileSync(filename, text.slice(0, start) + adapted + text.slice(completion.end));
} else {
    assert.equal(current, adapted, "unexpected decodeMappings.stopIterating snapshot");
}

console.log(JSON.stringify({ family: 156, file: "src/compiler/sourcemap.ts", edits: current === original ? 2 : 0 }));
