"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const ts = require("typescript");

const relative = "src/compiler/performanceCore.ts";
const local = "const timeOrigin = performance.timeOrigin;";
const baseGuard = 'typeof performance.timeOrigin === "number" && typeof performance.now === "function"';
const adaptedGuard = 'typeof timeOrigin === "number" && typeof performance.now === "function"';
const baseStore = "hooks.performanceTime = performance;";
const adaptedStore = "hooks.performanceTime = { now: () => performance.now(), timeOrigin };";

function only(matches, label) {
    assert.equal(matches.length, 1, label);
    return matches[0];
}

function adapt(tree) {
    assert.equal(ts.version, "6.0.3");
    const file = path.join(tree, relative);
    const text = fs.readFileSync(file, "utf8");
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, relative);
    const owner = only(source.statements.filter(node =>
        ts.isFunctionDeclaration(node) && node.name?.text === "tryGetPerformanceHooks" && node.body),
    "expected one tryGetPerformanceHooks");
    const statements = owner.body.statements;
    const check = only(statements.filter(node =>
        ts.isIfStatement(node) && [baseGuard, adaptedGuard].includes(node.expression.getText(source))),
    "expected one pinned timing guard");
    assert(ts.isBlock(check.thenStatement) && !check.elseStatement, "timing guard shape changed");
    const store = only(check.thenStatement.statements, "timing guard must own one statement");
    assert(ts.isExpressionStatement(store), "timing store must be an expression statement");

    const origins = statements.filter(node => ts.isVariableStatement(node) &&
        node.declarationList.declarations.some(declaration => declaration.name.getText(source) === "timeOrigin"));
    const adapted = check.expression.getText(source) === adaptedGuard;
    if (adapted) {
        const capture = only(origins, "expected one timing origin local");
        assert.equal(capture, statements[statements.indexOf(check) - 1], "timing origin local must precede the guard");
        assert.equal(capture.getText(source), local, "unexpected timing origin local");
        assert.equal(store.getText(source), adaptedStore, "unexpected adapted timing store");
    } else {
        assert.equal(origins.length, 0, "unexpected timing origin local");
        assert.equal(store.getText(source), baseStore, "unexpected base timing store");
        const start = check.getStart(source);
        const lineStart = text.lastIndexOf("\n", start - 1) + 1;
        const indent = text.slice(lineStart, start);
        assert.match(indent, /^[ \t]*$/, "timing guard must start its own line");
        const newline = text[lineStart - 2] === "\r" ? "\r\n" : "\n";
        const guardStart = check.expression.getStart(source);
        const storeStart = store.getStart(source);
        fs.writeFileSync(file, text.slice(0, start) + local + newline + indent +
            text.slice(start, guardStart) + adaptedGuard + text.slice(check.expression.end, storeStart) +
            adaptedStore + text.slice(store.end));
    }
    return {
        file: relative,
        line: source.getLineAndCharacterOfPosition(store.getStart(source)).line + 1,
        edits: adapted ? 0 : 3,
        javascriptChanging: true,
    };
}

assert.equal(process.argv.length, 3, "usage: node adapt.cjs TREE");
console.log(JSON.stringify(adapt(path.resolve(process.argv[2]))));
