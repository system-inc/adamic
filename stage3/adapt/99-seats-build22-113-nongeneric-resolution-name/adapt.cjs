#!/usr/bin/env node
"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const ts = require("typescript");

const name = "getTypeReferenceResolutionName";
const before = "function getTypeReferenceResolutionName<T extends FileReference | string>(entry: T)";
const after = "function getTypeReferenceResolutionName(entry: FileReference | string): string";
const body = "{\r\n    return !isString(entry) ? entry.fileName : entry;\r\n}";

function adapt(tree) {
    const filename = path.join(tree, "src/compiler/program.ts");
    const text = fs.readFileSync(filename, "utf8");
    const file = ts.createSourceFile(filename, text, ts.ScriptTarget.Latest, true);
    assert.equal(file.parseDiagnostics.length, 0, "program.ts must parse");

    const declarations = file.statements.filter(node =>
        ts.isFunctionDeclaration(node) && node.name?.text === name);
    assert.equal(declarations.length, 1, `expected one declaration of ${name}`);
    const [declaration] = declarations;
    assert(declaration.body, `missing body for ${name}`);
    assert.equal(declaration.body.getText(file), body, `unexpected body for ${name}`);

    const start = declaration.getStart(file);
    const bodyStart = declaration.body.getStart(file);
    const signature = text.slice(start, bodyStart);
    if (signature === after + " ") return false;
    assert.equal(signature, before + " ", `unexpected signature for ${name}`);

    fs.writeFileSync(filename, text.slice(0, start) + after + text.slice(bodyStart - 1));
    return true;
}

assert.equal(process.argv.length, 3, "usage: node adapt.cjs <typescript tree>");
const changed = adapt(process.argv[2]);
console.log(JSON.stringify({ files: changed ? 1 : 0, edits: changed ? 1 : 0 }));
