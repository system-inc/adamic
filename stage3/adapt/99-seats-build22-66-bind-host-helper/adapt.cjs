"use strict";
const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const ts = require("typescript");
const { nodes } = require("../47-host-errors/adapt.cjs");

const name = "bindHostWriteFile";
const before = "writeFile || host.writeFile?.bind(host)";
const after = "writeFile || bindHostWriteFile(host)";
const owners = ["emitBuildInfo", "emitNextAffectedFileOrDtsErrors", "getWriteFileCallback"];

assert.equal(process.argv.length, 3, "usage: node adapt.cjs <tree>");
assert.equal(ts.version, "6.0.3");
const file = path.join(process.argv[2], "src/compiler/builder.ts");
const text = fs.readFileSync(file, "utf8");
const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
assert.equal(source.parseDiagnostics.length, 0);
const newline = text.includes("\r\n") ? "\r\n" : "\n";
const helper = [
    "function bindHostWriteFile(current: BuilderProgramHost): WriteFileCallback | undefined {",
    "    return current.writeFile?.bind(current);",
    "}",
].join(newline);

const implementations = source.statements.filter(node =>
    ts.isFunctionDeclaration(node) && node.name?.text === "createBuilderProgram" && node.body);
assert.equal(implementations.length, 1, "createBuilderProgram implementation");
const fallbacks = owners.map(owner => {
    const functions = nodes(ts, implementations[0].body, node =>
        ts.isFunctionDeclaration(node) && node.name?.text === owner && node.body);
    assert.equal(functions.length, 1, owner);
    const matches = nodes(ts, functions[0].body, node => {
        if (!ts.isBinaryExpression(node) || node.operatorToken.kind !== ts.SyntaxKind.BarBarToken ||
            !ts.isIdentifier(node.left) || node.left.text !== "writeFile" || !ts.isCallExpression(node.right)) {
            return false;
        }
        const callee = node.right.expression;
        return (ts.isPropertyAccessExpression(callee) && callee.name.text === "bind") ||
            (ts.isIdentifier(callee) && callee.text === name);
    });
    assert.equal(matches.length, 1, `${owner} writeFile fallback`);
    const fallback = matches[0];
    assert([before, after].includes(fallback.getText(source)), `${owner} fallback drift: ${fallback.getText(source)}`);
    return fallback;
});

const helpers = source.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === name);
const last = source.statements.at(-1);
if (helpers.length > 0) {
    assert.equal(helpers.length, 1, name);
    assert.equal(helpers[0], last, "helper moved");
    assert.equal(last.getText(source), helper, "helper text drift");
    assert(fallbacks.every(node => node.getText(source) === after), "partly adapted builder fallbacks");
} else {
    assert(fallbacks.every(node => node.getText(source) === before), "partly adapted builder fallbacks");
    assert(ts.isFunctionDeclaration(last) && last.name?.text === "createRedirectedBuilderProgram",
        "builder insertion anchor moved");
    assert.equal(text.slice(last.end - 1), "}" + newline, "builder insertion text drift");
    let result = text.slice(0, last.end) + newline + newline + helper + text.slice(last.end);
    for (const fallback of fallbacks.sort((left, right) => right.getStart(source) - left.getStart(source))) {
        result = result.slice(0, fallback.getStart(source)) + after + result.slice(fallback.end);
    }
    fs.writeFileSync(file, result);
}
