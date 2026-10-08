#!/usr/bin/env node
"use strict";
// Reconstruct adaptation 20's allowance from its parsed declaration-owner ledger.
const fs = require("node:fs");
const path = require("node:path");
const crypto = require("node:crypto");
const assert = require("node:assert/strict");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
const [pristine, before, after, ledgerFile, output, mode] = process.argv.slice(2);
assert(pristine && before && after && ledgerFile && output && (!mode || mode === "--prepare-oracle"));
assert.equal(ts.version, "6.0.3");
function parse(name, text) { const source = ts.createSourceFile(name, text, ts.ScriptTarget.Latest, true); assert.equal(source.parseDiagnostics.length, 0); return source; }
function properties(source) {
    const result = [];
    function visit(node) { if ((ts.isPropertySignature(node) || ts.isPropertyDeclaration(node)) && node.questionToken && node.type) result.push(node); ts.forEachChild(node, visit); }
    visit(source); return result;
}
function owner(node) {
    const names = [];
    for (let current = node; current; current = current.parent) if (current.name && !(ts.isModuleDeclaration(current) && current.name.getText(node.getSourceFile()) === "ts")) names.unshift(current.name.getText(node.getSourceFile()));
    return JSON.stringify(names);
}
function tokens(text) {
    const scanner = ts.createScanner(ts.ScriptTarget.Latest, true, ts.LanguageVariant.Standard, text);
    const result = [];
    for (let kind = scanner.scan(); kind !== ts.SyntaxKind.EndOfFileToken; kind = scanner.scan()) result.push([kind, scanner.getTokenText()]);
    return JSON.stringify(result);
}
const owners = new Set();
const sources = new Map();
for (const edit of JSON.parse(fs.readFileSync(ledgerFile, "utf8")).declarations) {
    if (!sources.has(edit.file)) sources.set(edit.file, parse(edit.file, fs.readFileSync(path.join(pristine, edit.file), "utf8")));
    const source = sources.get(edit.file);
    const selected = properties(source).filter(node => owner(node) === JSON.stringify(edit.owner_path));
    assert.equal(selected.length, 1);
    assert.equal(selected[0].name.getText(source), edit.name);
    assert.equal(tokens(selected[0].type.getText(source)), tokens(edit.before));
    owners.add(owner(selected[0]));
}
const relative = "tests/baselines/reference/api/typescript.d.ts";
const original = fs.readFileSync(path.join(pristine, relative), "utf8");
assert.equal(fs.readFileSync(path.join(pristine, "built/local/typescript.d.ts"), "utf8"), original);
const api = parse("api.d.ts", original);
const changes = [];
for (const node of properties(api)) {
    if (!owners.has(owner(node))) continue;
    const wrap = ts.isFunctionTypeNode(node.type) || ts.isConstructorTypeNode(node.type) || ts.isConditionalTypeNode(node.type);
    if (wrap) changes.push({ position: node.type.getStart(api), text: "(" });
    changes.push({ position: node.type.end, text: (wrap ? ")" : "") + " | undefined" });
}
let expected = original;
for (const change of changes.sort((a, b) => b.position - a.position)) expected = expected.slice(0, change.position) + change.text + expected.slice(change.position);
const previous = fs.readFileSync(path.join(before, "built/local/typescript.d.ts"), "utf8");
const actual = fs.readFileSync(path.join(after, "built/local/typescript.d.ts"), "utf8");
assert.equal(actual, previous, "adaptation 47 changed the public API");
const publicAPI = parse("after-api.d.ts", actual);
const helperDeclarations = [];
function findHelpers(node) {
    if (ts.isFunctionDeclaration(node) && ["errorCode", "errorMessage"].includes(node.name?.text)) helperDeclarations.push(node.name.text);
    ts.forEachChild(node, findHelpers);
}
findHelpers(publicAPI);
assert.deepEqual(helperDeclarations, [], "helper leaked into public API");
assert(!actual.includes("hostErrors"));
const oldLines = original.split("\n"), expectedLines = expected.split("\n"), actualLines = actual.split("\n");
assert.equal(oldLines.length, expectedLines.length);
assert.equal(oldLines.length, actualLines.length);
const sanctionedLines = oldLines.flatMap((line, i) => line === expectedLines[i] ? [] : [i + 1]);
assert.equal(sanctionedLines.length, 189);
const preexistingChanges = expectedLines.flatMap((line, i) => line === actualLines[i] ? [] : [{ line: i + 1, expected: line, actual: actualLines[i] }]);
if (mode === "--prepare-oracle") fs.writeFileSync(path.join(after, relative), expected);
const digest = text => crypto.createHash("sha256").update(text).digest("hex");
const report = { status: "pass", helperChangesPublicAPI: false, beforeSHA256: digest(previous), afterSHA256: digest(actual), adaptation20Lines: sanctionedLines, preexistingChanges, preparedOracle: mode === "--prepare-oracle" };
fs.writeFileSync(output, JSON.stringify(report, null, 2) + "\n");
console.log(JSON.stringify({ status: "pass", helperChangesPublicAPI: false, sanctionedLines: sanctionedLines.length, preexistingChanges: preexistingChanges.length }));
