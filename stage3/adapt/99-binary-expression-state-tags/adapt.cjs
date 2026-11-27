#!/usr/bin/env node
"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const ts = require("typescript");
const snapshot = require("./snapshot.json");

const relative = "src/compiler/factory/utilities.ts";
const snapshotText = lines => lines.join("\r\n");

function exactlyOne(nodes, label) {
    assert.equal(nodes.length, 1, `${label}: expected one declaration`);
    return nodes[0];
}

function owners(text) {
    const source = ts.createSourceFile(relative, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, "factory utilities must parse");
    function find(predicate, name) {
        return exactlyOne(source.statements.filter(node => predicate(node) && node.name?.text === name), name);
    }
    return {
        state: find(ts.isTypeAliasDeclaration, "BinaryExpressionState"),
        handlers: find(ts.isModuleDeclaration, "BinaryExpressionState"),
        machine: find(ts.isClassDeclaration, "BinaryExpressionStateMachine"),
        walker: find(node => ts.isFunctionDeclaration(node) && node.body, "createBinaryExpressionTrampoline"),
    };
}

function adapt(tree) {
    assert.equal(ts.version, "6.0.3");
    const file = path.join(tree, relative);
    const source = fs.readFileSync(file, "utf8");
    const declarations = owners(source);
    assert.equal(declarations.machine.getText(), snapshotText(snapshot.machine), "state machine callbacks changed");

    const edits = [];
    let originals = 0;
    for (const name of ["state", "handlers", "walker"]) {
        const node = declarations[name];
        const { before, after } = snapshot[name];
        const text = node.getText();
        if (text === snapshotText(before)) {
            originals++;
            edits.push({ start: node.getStart(), end: node.end, after: snapshotText(after) });
        } else {
            assert.equal(text, snapshotText(after), `${name}: unrecognized snapshot`);
        }
    }
    assert.ok(originals === 0 || originals === 3, "partially adapted walker");
    if (originals === 0) return;

    let result = source;
    for (const { start, end, after } of edits.sort((a, b) => b.start - a.start)) {
        result = result.slice(0, start) + after + result.slice(end);
    }
    const adapted = owners(result);
    for (const name of ["state", "handlers", "walker"]) {
        assert.equal(adapted[name].getText(), snapshotText(snapshot[name].after));
    }
    fs.writeFileSync(file, result);
}

assert.equal(process.argv.length, 3, "usage: node adapt.cjs <tree>");
adapt(path.resolve(process.argv[2]));
