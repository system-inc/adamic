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

function resultReferences(source, producers, readers) {
    const names = new Set();
    for (const { node, enclosure } of producers) {
        const parent = node.parent;
        if (ts.isVariableDeclaration(parent) && parent.initializer === node && ts.isIdentifier(parent.name)) {
            names.add(parent.name.text);
        }
        if (ts.isExpressionStatement(enclosure) && enclosure.expression === parent &&
            ts.isBinaryExpression(parent) && parent.operatorToken.kind === ts.SyntaxKind.EqualsToken &&
            parent.right === node && ts.isIdentifier(parent.left)) {
            names.add(parent.left.text);
        }
    }
    for (const { node } of readers) {
        if (ts.isIdentifier(node.expression)) {
            names.add(node.expression.text);
        }
    }

    const references = [];
    function visit(node) {
        if (ts.isIdentifier(node) && names.has(node.text)) {
            references.push({
                node,
                pin: [node.text, ts.SyntaxKind[node.parent.kind], node.parent.getText(source)],
            });
        }
        ts.forEachChild(node, visit);
    }
    visit(source);
    return { names: [...names].sort(), references };
}

function checkResultReferences(source, references) {
    for (const [index, { node, pin }] of references.entries()) {
        const expected = pins.resultReferences[index];
        const { line } = source.getLineAndCharacterOfPosition(node.getStart(source));
        assert(expected && pin.every((value, field) => value === expected[field]),
            `reviewed result reference ${node.text} at line ${line + 1}: ${pin[2]}`);
    }
    const missing = pins.resultReferences[references.length];
    if (missing) {
        const { line } = source.getLineAndCharacterOfPosition(source.end);
        assert.fail(`reviewed result reference ${missing[0]} missing at line ${line + 1}: ${missing[2]}`);
    }
}

function singleFileProgram(file, text) {
    const options = { noResolve: true, noLib: true, types: [] };
    const host = ts.createCompilerHost(options);
    host.getSourceFile = (name, languageVersion) =>
        name === file ? ts.createSourceFile(name, text, languageVersion, true) : undefined;
    host.fileExists = name => name === file;
    host.readFile = name => name === file ? text : undefined;
    const program = ts.createProgram([file], options, host);
    assert.equal(program.getSourceFiles().length, 1, "reviewed single-file checker");
    assert.equal(program.getSourceFiles()[0].text, text, "reviewed checker input bytes");
    return program;
}

function resultTypes(checker, source) {
    const alias = firstMember(source).declaration;
    const type = checker.getTypeAtLocation(alias.name);
    assert(type.isUnion() && type.types.length === 3, "reviewed SyntacticResult union");
    const members = new Set(type.types);
    return {
        alias,
        matches: candidate => candidate === type || members.has(candidate) ||
            (candidate.isUnion() && candidate.types.some(member => members.has(member))),
    };
}

function typedResults(checker, source, types) {
    const bindings = [];
    const references = [];
    function visit(node) {
        if (ts.isTypeNode(node)) return;
        if (ts.isIdentifier(node) && node !== types.alias.name && types.matches(checker.getTypeAtLocation(node))) {
            const parent = node.parent;
            const reference = { node, pin: [node.text, ts.SyntaxKind[parent.kind], parent.getText(source)] };
            references.push(reference);
            if (parent.name === node && (ts.isVariableDeclaration(parent) || ts.isParameter(parent) ||
                ts.isBindingElement(parent) || ts.isPropertyDeclaration(parent) || ts.isPropertyAssignment(parent) ||
                ts.isShorthandPropertyAssignment(parent))) {
                bindings.push(reference);
            }
        }
        ts.forEachChild(node, visit);
    }
    visit(source);
    return { bindings, references };
}

function resultCalls(checker, source, types) {
    const functions = new Set();
    const calls = [];
    function visit(node) {
        if (ts.isFunctionDeclaration(node) && node.name) {
            const signature = checker.getSignatureFromDeclaration(node);
            if (signature && ((node.type && ts.isTypeReferenceNode(node.type) &&
                ts.isIdentifier(node.type.typeName) && node.type.typeName.text === "SyntacticResult") ||
                types.matches(checker.getReturnTypeOfSignature(signature)))) {
                functions.add(node);
            }
        }
        if (ts.isCallExpression(node)) calls.push(node);
        ts.forEachChild(node, visit);
    }
    visit(source);
    return {
        functions,
        calls: calls.filter(call => functions.has(checker.getResolvedSignature(call)?.declaration)),
    };
}

function checkTypedPins(source, values, expected, label) {
    for (const [index, { node, pin }] of values.entries()) {
        const match = expected[index];
        const { line } = source.getLineAndCharacterOfPosition(node.getStart(source));
        assert(match && pin.every((value, field) => value === match[field]),
            `${label} ${pin[0]} at line ${line + 1}: ${pin[2]}`);
    }
    const missing = expected[values.length];
    if (missing) {
        const { line } = source.getLineAndCharacterOfPosition(source.end);
        assert.fail(`${label} ${missing[0]} missing at line ${line + 1}: ${missing[2]}`);
    }
}

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
    const results = resultReferences(source, producers, readers);
    assert.deepEqual(results.names, pins.resultNames, "reviewed result binding names");
    checkResultReferences(source, results.references);
    const program = singleFileProgram(file, text);
    const checkedSource = program.getSourceFiles()[0];
    const checker = program.getTypeChecker();
    const types = resultTypes(checker, checkedSource);
    const typed = typedResults(checker, checkedSource, types);
    checkTypedPins(checkedSource, typed.bindings, pins.typedBindings, "reviewed typed result binding");
    checkTypedPins(checkedSource, typed.references, pins.typedReferences, "reviewed typed result reference");

    const { calls } = resultCalls(checker, checkedSource, types);
    const statements = calls.filter(call => !ts.isIdentifier(call.expression) || call.expression.text !== "syntacticResult")
        .map(call => {
            const statement = enclosing(call, ts.isStatement, "result call");
            return {
                node: call,
                pin: [call.expression.getText(checkedSource), ts.SyntaxKind[statement.kind], statement.getText(checkedSource)],
            };
        });
    checkTypedPins(checkedSource, statements, pins.resultCallStatements, "reviewed result call statement");
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
