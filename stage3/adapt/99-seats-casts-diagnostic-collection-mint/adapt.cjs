#!/usr/bin/env node
"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const ts = require("typescript");

assert.equal(ts.version, "6.0.3");
assert.equal(process.argv.length, 3, "usage: node adapt.cjs <tree>");

const relative = "src/compiler/utilities.ts";
const file = path.join(process.argv[2], relative);
const text = fs.readFileSync(file, "utf8");
const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
assert.equal(source.parseDiagnostics.length, 0, `cannot parse ${relative}`);

function functions(nodes, name) {
    return nodes.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === name);
}

function body(nodes, name, header) {
    const declarations = functions(nodes, name);
    assert.equal(declarations.length, 1, `expected one ${name} function declaration in ${relative}`);
    const declaration = declarations[0];
    assert(declaration.body, `missing ${name} implementation in ${relative}`);
    const actual = text.slice(declaration.getStart(source), declaration.body.getStart(source)).trimEnd();
    assert.equal(actual, header, `unreviewed ${name} header in ${relative}`);
    return declaration.body;
}

const collection = body(source.statements, "createDiagnosticCollection",
    "export function createDiagnosticCollection(): DiagnosticCollection");
const add = body(collection.statements, "add", "function add(diagnostic: Diagnostic): void");

const before = "[] as Diagnostic[] as SortedArray<DiagnosticWithLocation>";
const after = "[] as unknown as SortedArray<DiagnosticWithLocation>";
const literals = [];
const bindings = [];

function visit(node) {
    if (ts.isArrayLiteralExpression(node)) literals.push(node);
    if (ts.isDeclaration(node) && node.name && ts.isIdentifier(node.name) && node.name.text === "diagnostics") {
        bindings.push(node);
    }
    ts.forEachChild(node, visit);
}

visit(add);
assert.equal(literals.length, 1, "expected one array literal in createDiagnosticCollection.add");
let cast = literals[0];
function wrapsExpression(node, expression) {
    return (ts.isParenthesizedExpression(node) || ts.isAsExpression(node)
        || ts.isTypeAssertionExpression(node) || ts.isSatisfiesExpression(node))
        && node.expression === expression;
}
while (wrapsExpression(cast.parent, cast)) cast = cast.parent;
assert(ts.isAsExpression(cast), "expected an outer as cast in createDiagnosticCollection.add");
const actual = cast.getText(source);
assert(actual === before || actual === after, `unreviewed createDiagnosticCollection.add cast: ${actual}`);
let owner = cast.parent;
while (owner && !ts.isFunctionLike(owner) && !ts.isClassLike(owner)) owner = owner.parent;
assert(owner === add.parent, "unreviewed createDiagnosticCollection.add cast owner");

const assignment = cast.parent;
assert(ts.isBinaryExpression(assignment) && assignment.operatorToken.kind === ts.SyntaxKind.EqualsToken
    && ts.isIdentifier(assignment.left) && assignment.left.text === "diagnostics" && assignment.right === cast,
    "unreviewed createDiagnosticCollection.add cast assignment");

assert.equal(bindings.length, 1, "expected one diagnostics declaration in createDiagnosticCollection.add");
const binding = bindings[0];
const statement = binding.parent.parent;
assert(ts.isVariableDeclaration(binding) && ts.isVariableStatement(statement) && statement.parent === add
    && statement.getText(source) === "let diagnostics: SortedArray<Diagnostic> | undefined;",
    "unreviewed createDiagnosticCollection.add diagnostics binding");

const edited = actual === before;
if (edited) {
    assert.equal(fs.readFileSync(file, "utf8"), text, `source changed while planning ${relative}`);
    fs.writeFileSync(file, text.slice(0, cast.getStart(source)) + after + text.slice(cast.end));
}
console.log(JSON.stringify({ files: Number(edited), edits: Number(edited) }));
