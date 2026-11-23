"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const ts = require("typescript");

const helperName = "scanConflictMarkerTrivia";
const originalType = "(diag: DiagnosticMessage, pos?: number, len?: number) => void";
const requiredType = "(diag: DiagnosticMessage, pos: number, len: number) => void";
const originalCall = "scanConflictMarkerTrivia(text!, pos, error)";
const wrapper = "(diag, at, len) => error(diag, at, len)";
const wrappedCall = "scanConflictMarkerTrivia(text!, pos, " + wrapper + ")";
const markerCases = ["lessThan", "equals", "greaterThan", "bar"];

function reviewedHelper(callbackType) {
    return `function scanConflictMarkerTrivia(text: string, pos: number, error?: ${callbackType}) {
    if (error) {
        error(Diagnostics.Merge_conflict_marker_encountered, pos, mergeConflictMarkerLength);
    }

    const ch = text.charCodeAt(pos);
    const len = text.length;

    if (ch === CharacterCodes.lessThan || ch === CharacterCodes.greaterThan) {
        while (pos < len && !isLineBreak(text.charCodeAt(pos))) {
            pos++;
        }
    }
    else {
        Debug.assert(ch === CharacterCodes.bar || ch === CharacterCodes.equals);
        // Consume everything from the start of a ||||||| or ======= marker to the start
        // of the next ======= or >>>>>>> marker.
        while (pos < len) {
            const currentChar = text.charCodeAt(pos);
            if ((currentChar === CharacterCodes.equals || currentChar === CharacterCodes.greaterThan) && currentChar !== ch && isConflictMarkerTrivia(text, pos)) {
                break;
            }

            pos++;
        }
    }

    return pos;
}`.replaceAll("\n", "\r\n");
}

function one(items, label) {
    assert.equal(items.length, 1, label);
    return items[0];
}

function descendants(root, predicate) {
    const matches = [];
    function visit(node) {
        if (predicate(node)) matches.push(node);
        ts.forEachChild(node, visit);
    }
    visit(root);
    return matches;
}

function declaration(owner, name) {
    return one(owner.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === name), name);
}

function apply(tree) {
    assert.equal(ts.version, "6.0.3");
    const file = path.join(tree, "src/compiler/scanner.ts");
    const before = fs.readFileSync(file, "utf8");
    const source = ts.createSourceFile(file, before, ts.ScriptTarget.ES2020, true);
    assert.equal(source.parseDiagnostics.length, 0);

    const helper = declaration(source, helperName);
    assert.equal(helper.modifiers, undefined, "helper must be private");
    const callback = one(helper.parameters.filter(p => p.name.getText(source) === "error"), "callback parameter");
    const type = callback.type;
    assert(callback.questionToken && type && ts.isFunctionTypeNode(type), "optional callback with a function type");
    const typeText = type.getText(source);
    assert(typeText === originalType || typeText === requiredType, "reviewed callback type");
    assert(helper.getText(source) === reviewedHelper(typeText), "reviewed helper declaration");
    const reads = descendants(helper.body, node => ts.isIdentifier(node) && node.text === "error");
    assert.equal(reads.length, 2, "callback is only tested for presence and called");
    const [guard, callTarget] = reads;
    assert(ts.isIfStatement(guard.parent) && guard.parent.expression === guard, "callback presence guard");
    assert(ts.isCallExpression(callTarget.parent) && callTarget.parent.expression === callTarget, "direct callback call");
    assert.equal(callTarget.parent.getText(source),
        "error(Diagnostics.Merge_conflict_marker_encountered, pos, mergeConflictMarkerLength)");

    const inventory = descendants(source, node => ts.isCallExpression(node) &&
        ts.isIdentifier(node.expression) && node.expression.text === helperName);
    assert.equal(inventory.length, 6, "helper call inventory");
    // A second binding or an alias would require another identifier with this name.
    const reviewedNames = new Set([helper.name, ...inventory.map(call => call.expression)]);
    const names = descendants(source, node => ts.isIdentifier(node) && node.text === helperName);
    assert.equal(names.length, reviewedNames.size, "reviewed helper binding inventory");
    assert(names.every(name => reviewedNames.has(name)), "reviewed helper binding inventory");

    const scanner = declaration(source, "createScanner");
    const scan = declaration(scanner.body, "scan");
    const jsx = declaration(scanner.body, "scanJsxToken");
    const trivia = declaration(source, "skipTrivia");
    function owner(node) {
        for (let parent = node.parent; parent; parent = parent.parent) {
            if (ts.isFunctionLike(parent)) return parent;
        }
    }
    const omitted = one(inventory.filter(call => owner(call) === trivia), "skipTrivia helper call");
    assert.equal(omitted.getText(source), "scanConflictMarkerTrivia(text, pos)");
    const retained = one(inventory.filter(call => owner(call) === jsx), "scanJsxToken helper call");
    assert.equal(retained.getText(source), originalCall);

    const selected = inventory.filter(call => owner(call) === scan);
    assert.equal(selected.length, markerCases.length, "scan helper call inventory");
    const calls = markerCases.map(name => {
        const branch = one(descendants(scan.body, node => ts.isCaseClause(node) &&
            node.expression.getText(source) === "CharacterCodes." + name), "scanner " + name + " case");
        const guard = one(branch.statements.filter(node => ts.isIfStatement(node) &&
            node.expression.getText(source) === "isConflictMarkerTrivia(text!, pos)"), "conflict-marker guard in " + name);
        assert(ts.isBlock(guard.thenStatement), "conflict-marker block in " + name);
        const statement = guard.thenStatement.statements[0];
        assert(statement && ts.isExpressionStatement(statement) && ts.isBinaryExpression(statement.expression),
            "first conflict-marker statement in " + name);
        const call = one(selected.filter(node => statement.expression.right === node),
            "first conflict-marker statement in " + name);
        assert.equal(statement.getText(source), "pos = " + call.getText(source) + ";");
        return call;
    });
    if (typeText === requiredType && calls.every(call => call.getText(source) === wrappedCall)) return;
    assert.equal(typeText, originalType);
    for (const call of calls) assert.equal(call.getText(source), originalCall);

    let after = before;
    const argumentsToWrap = calls.map(call => call.arguments[2])
        .sort((a, b) => b.getStart(source) - a.getStart(source));
    for (const argument of argumentsToWrap) {
        after = after.slice(0, argument.getStart(source)) + wrapper + after.slice(argument.end);
    }
    after = after.slice(0, type.getStart(source)) + requiredType + after.slice(type.end);
    assert.equal(fs.readFileSync(file, "utf8"), before, "concurrent scanner edit");
    fs.writeFileSync(file, after);
}

assert.equal(process.argv.length, 3, "usage: node adapt.cjs <tree>");
apply(process.argv[2]);
