#!/usr/bin/env node
"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const ts = require("typescript");
const { nodes } = require("../47-host-errors/adapt.cjs");

const ownerName = "formatDiagnosticsWithColorAndContext";
const array = "diagnostic.relatedInformation";
const fields = ["file", "start", "length", "messageText"];
const newline = "\r\n";
const body = [
    "                if (file) {",
    "                    output += host.getNewLine();",
    "                    output += halfIndent + formatLocation(file, start!, host); // TODO: GH#18217",
    "                    output += formatCodeSpan(file, start!, length!, indent, ForegroundColorEscapeSequences.Cyan, host); // TODO: GH#18217",
    "                }",
    "                output += host.getNewLine();",
    "                output += indent + flattenDiagnosticMessageText(messageText, host.getNewLine());",
    "            }",
];
const original = [`for (const { ${fields.join(", ")} } of ${array}) {`, ...body].join(newline);
const replacement = [
    `for (const info of ${array}) {`,
    ...fields.map(field => `                const ${field} = info.${field};`),
    ...body,
].join(newline);

function adapt(tree) {
    assert.equal(ts.version, "6.0.3");
    const filename = path.join(tree, "src/compiler/program.ts");
    const text = fs.readFileSync(filename, "utf8");
    const parse = source => ts.createSourceFile(filename, source, ts.ScriptTarget.Latest, true);
    const source = parse(text);
    assert.equal(source.parseDiagnostics.length, 0);
    const owners = source.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === ownerName);
    assert.equal(owners.length, 1, ownerName);
    const [owner] = owners;
    const loops = nodes(ts, owner, node => ts.isForOfStatement(node) && node.expression.getText(source) === array);
    assert.equal(loops.length, 1, `${ownerName}: ${array} loops`);
    const [loop] = loops;
    const loopText = loop.getText(source);
    if (loopText === replacement) return;
    assert.equal(loopText, original, `${ownerName}: loop snapshot moved`);
    assert.equal(nodes(ts, owner, node => ts.isIdentifier(node) && node.text === "info").length, 0,
        `${ownerName}: info already occurs in scope`);
    const updated = text.slice(0, loop.getStart(source)) + replacement + text.slice(loop.end);
    assert.equal(parse(updated).parseDiagnostics.length, 0);
    fs.writeFileSync(filename, updated);
}

if (require.main === module) adapt(process.argv[2]);
module.exports = { adapt };
