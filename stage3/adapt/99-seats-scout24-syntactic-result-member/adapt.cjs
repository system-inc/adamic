#!/usr/bin/env node
"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const ts = require("typescript");

assert.equal(ts.version, "6.0.3");
assert.equal(process.argv.length, 3, "usage: node adapt.cjs <tree>");
const file = path.join(process.argv[2], "src/compiler/expressionToTypeNode.ts");
const text = fs.readFileSync(file, "utf8");
const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
assert.equal(source.parseDiagnostics.length, 0);

function firstMember(parsed) {
    const declarations = parsed.statements.filter(node =>
        ts.isTypeAliasDeclaration(node) && node.name.text === "SyntacticResult");
    assert.equal(declarations.length, 1, "SyntacticResult declaration");
    const declaration = declarations[0];
    assert(ts.isUnionTypeNode(declaration.type));
    const first = declaration.type.types[0];
    assert(ts.isTypeLiteralNode(first));
    return { declaration, first };
}
const { declaration, first } = firstMember(source);
const [fallback] = first.members.filter(node => ts.isPropertySignature(node) && node.name.getText(source) === "reportFallback");
assert(fallback?.type, "first member's reportFallback type");

const reviewedConstructors = [
    "function syntacticResult(type: TypeNode | undefined): SyntacticResult;",
    "function syntacticResult(type: undefined, reportFallback: boolean): SyntacticResult;",
    "function syntacticResult(type: TypeNode | undefined, reportFallback: boolean = true) {\r\n    return { type, reportFallback } as SyntacticResult;\r\n}"
];
const constructors = source.statements.filter(node =>
    ts.isFunctionDeclaration(node) && node.name?.text === "syntacticResult");
assert.deepEqual(constructors.map(node => node.getText(source)), reviewedConstructors);

const before = "type SyntacticResult =\r\n    | { type: TypeNode; reportFallback: undefined; }\r\n    | { type: undefined; reportFallback: true; }\r\n    | { type: undefined; reportFallback: false; };";
const expectedFirstMember = "{ type: TypeNode; reportFallback: true; }";
const replacement = "true";
const after = before.replace("{ type: TypeNode; reportFallback: undefined; }", expectedFirstMember);

const pins = require("./pins.json");

function checkPaths() {
    const producers = [];
    const readers = [];
    const occurrences = new Map();
    const names = new Set(["syntacticResult", "SyntacticResult", "reportFallback"]);
    function line(node) {
        return source.getLineAndCharacterOfPosition(node.getStart(source)).line + 1;
    }
    function count(key, node) {
        const nodes = occurrences.get(key) ?? [];
        nodes.push(node);
        occurrences.set(key, nodes);
    }
    function enclosing(node, test, label) {
        for (let parent = node.parent; parent; parent = parent.parent) {
            if (test(parent)) return parent;
        }
        assert.fail(`${label} at line ${line(node)} has no reviewed enclosure`);
    }
    function visit(node) {
        if ((ts.isIdentifier(node) || ts.isStringLiteralLike(node)) && names.has(node.text)) {
            count(`${node.text}/${ts.SyntaxKind[node.kind]}/${ts.SyntaxKind[node.parent.kind]}`, node);
        }
        if (ts.isSpreadAssignment(node)) count("object spread", node);
        if (ts.isCallExpression(node)) {
            const callee = node.expression;
            const dotAssign = ts.isPropertyAccessExpression(callee) && callee.name.text === "assign";
            const bracketAssign = ts.isElementAccessExpression(callee) && ts.isStringLiteralLike(callee.argumentExpression) &&
                callee.argumentExpression.text === "assign";
            if ((dotAssign || bracketAssign) && callee.expression.getText(source) === "Object") {
                count("Object.assign", node);
            }
            if (ts.isIdentifier(callee) && callee.text === "syntacticResult") {
                producers.push({ node, enclosure: enclosing(node, ts.isStatement, "syntacticResult call") });
            }
        }
        if ((ts.isPropertyAccessExpression(node) && node.name.text === "reportFallback") ||
            (ts.isElementAccessExpression(node) && ts.isStringLiteralLike(node.argumentExpression) &&
                node.argumentExpression.text === "reportFallback")) {
            // The reviewed alias and constructor contain no property-access expressions.
            readers.push({ node });
        }
        ts.forEachChild(node, visit);
    }
    visit(source);

    function checkText(nodes, expected, label, getEnclosure) {
        for (const [index, { node, enclosure }] of nodes.entries()) {
            const actual = (enclosure ?? getEnclosure(node)).getText(source);
            assert(actual === expected[index], `${label} ${node.getText(source)} at line ${line(node)}`);
        }
        assert.equal(nodes.length, expected.length, `${label} inventory at ${file}:${source.getLineAndCharacterOfPosition(source.end).line + 1}`);
    }

    checkText(producers, pins.producers, "reviewed syntacticResult statement");
    checkText(readers, pins.readers, "reviewed reportFallback guard", node =>
        enclosing(node, parent => ts.isIfStatement(parent) || ts.isConditionalExpression(parent), "reportFallback access"));
    for (const [key, nodes] of occurrences) {
        const expected = pins.counts[key] ?? 0;
        assert.equal(nodes.length, expected, `reviewed ${key} count at line ${line(nodes[expected] ?? nodes[0])}`);
    }
    for (const [key, expected] of Object.entries(pins.counts)) {
        assert.equal(occurrences.get(key)?.length ?? 0, expected, `reviewed ${key} count at ${file}:${source.getLineAndCharacterOfPosition(source.end).line + 1}`);
    }
}

checkPaths();

function assertFirstMember(written) {
    const parsed = ts.createSourceFile(file, written, ts.ScriptTarget.Latest, true);
    assert.equal(firstMember(parsed).first.getText(parsed), expectedFirstMember, "SyntacticResult first-member postcondition");
}

const actual = declaration.getText(source);
if (actual === after) {
    assertFirstMember(text);
} else {
    assert.equal(actual, before, "reviewed SyntacticResult declaration");
    assert.equal(fallback.type.getText(source), "undefined");
    const next = text.slice(0, fallback.type.getStart(source)) + replacement + text.slice(fallback.type.end);
    assertFirstMember(next);
    assert.equal(fs.readFileSync(file, "utf8"), text, "concurrent source edit");
    fs.writeFileSync(file, next);
    assertFirstMember(fs.readFileSync(file, "utf8"));
}
