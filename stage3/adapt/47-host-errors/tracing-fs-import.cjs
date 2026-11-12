'use strict';
const assert = require('node:assert/strict'), path = require('node:path');
function bindings(ts, program, source, nodes, done) {
    const checker = program.getTypeChecker();
    const declarations = nodes(ts, source, n => ts.isVariableDeclaration(n) && ts.isIdentifier(n.name) && n.name.text === 'fs');
    const imports = source.statements.filter(n => ts.isImportDeclaration(n) && n.moduleSpecifier.text === 'node:fs');
    assert.equal(declarations.length, done ? 0 : 1, 'tracing fs local drift');
    assert.equal(imports.length, done ? 1 : 0, 'tracing fs import drift');
    let binding;
    if (done) {
        assert.equal(imports[0].getText(source), 'import * as fs from "node:fs";');
        binding = imports[0].importClause.namedBindings.name;
    } else {
        assert.equal(declarations[0].parent.parent.getText(source), 'let fs: typeof import("fs");', 'tracing fs declaration drift');
        binding = declarations[0].name;
    }
    const symbol = checker.getSymbolAtLocation(binding);
    assert(symbol, 'tracing fs symbol missing');
    const uses = nodes(ts, source, n => ts.isIdentifier(n) && n.text === 'fs' && n !== binding);
    let qualified = 0, lazy = 0;
    for (const use of uses) {
        assert(checker.getSymbolAtLocation(use) === symbol, 'tracing fs captured or shadowed');
        if (ts.isPropertyAccessExpression(use.parent) && use.parent.expression === use) qualified++;
        else {
            assert(!done && ts.isBinaryExpression(use.parent) && use.parent.left === use &&
                ['fs === undefined', 'fs = require("fs")'].includes(use.parent.getText(source)), 'collision or bare tracing namespace use');
            lazy++;
        }
    }
    assert.equal(lazy, done ? 0 : 2, 'tracing lazy use count drift');
    assert.equal(qualified, 17, 'tracing qualified use count drift');
    return {declaration: declarations[0]?.parent.parent, qualified};
}
function plan({ts, program, tracing, start, startText, metadata, nodes, edit, planned}) {
    const done = startText === metadata.after;
    const checked = bindings(ts, program, tracing, nodes, done);
    const imports = tracing.statements.filter(ts.isImportDeclaration);
    const helpers = imports.filter(n => n.moduleSpecifier.text === './hostErrors.js');
    assert.equal(helpers.length, startText === metadata.lazyAfter ? 1 : 0, 'tracing helper anchor drift');
    if (helpers.length) assert.equal(helpers[0].getText(tracing), 'import { errorMessage } from "./hostErrors.js";');
    if (done) return;
    const loads = nodes(ts, start, ts.isTryStatement);
    assert.equal(loads.length, 1, 'tracing loader try drift');
    const load = loads[0];
    assert.equal(load.tryBlock.statements.length, 1);
    assert.equal(load.tryBlock.statements[0].getText(tracing), 'fs = require("fs");');
    assert(ts.isBlock(load.parent) && ts.isIfStatement(load.parent.parent));
    const guard = load.parent.parent;
    assert.equal(guard.expression.getText(tracing), 'fs === undefined');
    // The sole potentially throwing statement moves to module evaluation. With
    // it removed, this try block is empty and its original rethrow unreachable.
    // Pinning the whole owner above also pins the exact original catch text.
    const removeLine = (node, trailingBlank) => {
        const begin = tracing.text.lastIndexOf('\n', node.getStart(tracing) - 1) + 1;
        let end = tracing.text.indexOf('\n', node.end) + 1;
        assert(end > node.end, 'tracing statement newline drift');
        if (trailingBlank) {
            const next = tracing.text.indexOf('\n', end) + 1;
            assert.equal(tracing.text.slice(end, next).trim(), '', 'tracing loader spacing drift');
            end = next;
        }
        if (!planned.has(tracing)) planned.set(tracing, []);
        planned.get(tracing).push({start: begin, end, text: ''});
    };
    removeLine(guard, true);
    removeLine(checked.declaration, false);
    const newline = tracing.text.includes('\r\n') ? '\r\n' : '\n';
    if (helpers.length) {
        assert.equal(imports[imports.length - 1], helpers[0], 'tracing helper import position drift');
        removeLine(helpers[0], false);
    }
    const at = imports[imports.length - 1 - helpers.length].end;
    if (!planned.has(tracing)) planned.set(tracing, []);
    planned.get(tracing).push({start: at, end: at, text: `${newline}import * as fs from "node:fs";`});
}
function validatePlanned(ts, program, source, text, metadata, nodes) {
    const host = ts.createCompilerHost(program.getCompilerOptions(), true), original = host.getSourceFile.bind(host);
    host.getSourceFile = (name, language, onError, fresh) => path.resolve(name) === path.resolve(source.fileName)
        ? ts.createSourceFile(name, text, language, true) : original(name, language, onError, fresh);
    const planned = ts.createProgram(program.getRootFileNames(), program.getCompilerOptions(), host);
    const changed = planned.getSourceFile(source.fileName);
    const owners = nodes(ts, changed, n => ts.isFunctionDeclaration(n) && n.name?.text === 'startTracing');
    assert.equal(owners.length, 1);
    // Removing the complete guard retains its whitespace. Compare precisely
    // after the same deletion from the pinned lazy owner.
    assert.equal(owners[0].getText(changed).replace(/\r\n/g, '\n'), metadata.after, 'unreviewed eager startTracing');
    assert.equal(changed.statements.filter(n => ts.isImportDeclaration(n) && n.moduleSpecifier.text === './hostErrors.js').length, 0, 'unused tracing helper import remains');
    bindings(ts, planned, changed, nodes, true);
}
module.exports = {plan, bindings, validatePlanned};
