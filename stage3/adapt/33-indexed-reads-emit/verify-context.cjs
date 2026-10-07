#!/usr/bin/env node
"use strict";
const fs = require("node:fs");
const path = require("node:path");
const assert = require("node:assert/strict");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
assert.equal(ts.version, "6.0.3");
const file = "transformers/module/module.ts";
const input = path.resolve(process.argv[2], "src/compiler", file);
const program = ts.createProgram([input], {noResolve:true, noLib:true, target:ts.ScriptTarget.Latest});
const source = program.getSourceFile(input);
const checker = program.getTypeChecker();
assert.equal(source.parseDiagnostics.length, 0);
const functions = new Map();
function collect(node) {
    if (ts.isFunctionDeclaration(node) && node.name) {
        assert(!functions.has(node.name.text), "duplicate local function name");
        functions.set(node.name.text, node);
    }
    ts.forEachChild(node, collect);
}
collect(source);
// Conservatively include references passed as callbacks and returned delegates,
// not merely direct calls. Nested declarations have their own graph vertices.
const graph = new Map();
for (const [name, fn] of functions) {
    const edges = new Set();
    function visit(node) {
        if (ts.isFunctionDeclaration(node)) return;
        if (ts.isIdentifier(node) && functions.has(node.text)) {
            const symbol = checker.getSymbolAtLocation(node);
            if (symbol?.declarations?.includes(functions.get(node.text))) edges.add(node.text);
        }
        ts.forEachChild(node, visit);
    }
    visit(fn.body);
    graph.set(name, edges);
}
function reachable(root) {
    const seen = new Set();
    function visit(name) {
        if (seen.has(name)) return;
        seen.add(name);
        for (const next of graph.get(name) || []) visit(next);
    }
    visit(root);
    return seen;
}
const transformation = reachable("transformSourceFile");
const emission = new Set([...reachable("onEmitNode"), ...reachable("onSubstituteNode")]);
const sites = require("./required-values.json").filter(site => site.file === file && site.kind === "Identifier");
assert.equal(sites.length, 22);
const owners = [...new Set(sites.map(site => site.function))];
for (const name of owners) {
    assert(transformation.has(name), "required reader outside transformation: " + name);
    assert(!emission.has(name), "required reader reachable in emission: " + name);
}
for (const name of transformation) assert(!graph.get(name).has("transformSourceFile"), "recursive source-file transformation");
const writes = [];
function writesVisit(node) {
    if (ts.isBinaryExpression(node) && node.operatorToken.kind === ts.SyntaxKind.EqualsToken &&
        ts.isIdentifier(node.left) && node.left.text === "currentModuleInfo") {
        let owner = node.parent;
        while (owner && !ts.isFunctionDeclaration(owner)) owner = owner.parent;
        writes.push({node, owner: owner.name.text, right: node.right.getText(source)});
    }
    ts.forEachChild(node, writesVisit);
}
writesVisit(source);
assert.deepEqual(writes.map(write => [write.owner, write.right]), [
    ["transformSourceFile", "collectExternalModuleInfo(context, node)"],
    ["transformSourceFile", "undefined!"],
    ["onEmitNode", "moduleInfoMap[getOriginalNodeId(currentSourceFile)]"],
    ["onEmitNode", "undefined!"],
]);
const delegateCalls = [];
function callsVisit(node) {
    if (ts.isCallExpression(node) && node.expression.getText(source) === "transformModule") delegateCalls.push(node);
    ts.forEachChild(node, callsVisit);
}
callsVisit(functions.get("transformSourceFile").body);
assert.equal(delegateCalls.length, 1);
assert(writes[0].node.end < delegateCalls[0].pos && delegateCalls[0].end < writes[1].node.pos);
console.log(JSON.stringify({status:"pass", required_reads:sites.length, required_reader_functions:owners,
    all_required_readers_reachable_in_transformation:true, all_required_readers_unreachable_in_emission:true,
    writes_confined_to_transform_and_emit_roots:true, initialization_precedes_delegate:true,
    reset_follows_delegate:true, no_recursive_source_file_transform:true}, null, 2));
