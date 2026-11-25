#!/usr/bin/env node
"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const ts = require("typescript");

const unitName = "loadModuleFromSpecificNodeModulesDirectory";
const originalCall = 'hasProperty((rootPackageInfo = getPackageJsonInfo(packageDirectory, !nodeModulesDirectoryExists, state))?.contents.packageJsonContent ?? emptyArray, "exports")';
const adaptedCall = 'hasProperty(rootPackageJsonContent, "exports")';
const adaptedOperand = '!((rootPackageJsonContent = (rootPackageInfo = getPackageJsonInfo(packageDirectory, !nodeModulesDirectoryExists, state))?.contents.packageJsonContent) !== undefined && ' + adaptedCall + ')';

function findUnit(source, name) {
    const units = source.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === name);
    assert.equal(units.length, 1, name);
    return units[0];
}

function plan(original, file) {
    assert.equal(ts.version, "6.0.3");
    const source = ts.createSourceFile(file, original, ts.ScriptTarget.Latest, true);
    assert.deepEqual(source.parseDiagnostics, []);
    const unit = findUnit(source, unitName);
    const declarations = [];
    const calls = [];
    function visit(node) {
        if (ts.isVariableDeclaration(node) && ts.isIdentifier(node.name) &&
            ["rootPackageInfo", "rootPackageJsonContent"].includes(node.name.text)) {
            declarations.push(node);
        }
        if (ts.isCallExpression(node) && ts.isIdentifier(node.expression) &&
            node.expression.text === "hasProperty" && node.arguments.length === 2 &&
            ts.isStringLiteral(node.arguments[1]) && node.arguments[1].text === "exports") {
            calls.push(node);
        }
        ts.forEachChild(node, visit);
    }
    visit(unit);
    assert.equal(calls.length, 1, "expected one exports ownership call");
    const [call] = calls;
    const [info, content] = declarations;
    assert(info);
    assert.equal(info.getText(source), "rootPackageInfo: PackageJsonInfo | undefined");
    const statement = info.parent.parent;
    assert(ts.isVariableStatement(statement));
    assert.equal(statement.getText(source), "let rootPackageInfo: PackageJsonInfo | undefined;");
    if (content) {
        assert.equal(declarations.length, 2);
        assert.equal(content.parent.parent.getText(source), "let rootPackageJsonContent: PackageJsonPathFields | undefined;");
        assert.equal(call.getText(source), adaptedCall);
        assert.equal(call.parent.parent.parent.getText(source), adaptedOperand);
        return original;
    }

    assert.equal(declarations.length, 1);
    assert.equal(call.getText(source), originalCall);
    const argument = call.arguments[0];
    assert(ts.isBinaryExpression(argument));
    assert.equal(argument.operatorToken.kind, ts.SyntaxKind.QuestionQuestionToken);
    assert(ts.isIdentifier(argument.right) && argument.right.text === "emptyArray");
    assert(ts.isPrefixUnaryExpression(call.parent));
    assert.equal(call.parent.operator, ts.SyntaxKind.ExclamationToken);

    const newline = original.includes("\r\n") ? "\r\n" : "\n";
    return original.slice(0, statement.end) +
        newline + "    let rootPackageJsonContent: PackageJsonPathFields | undefined;" +
        original.slice(statement.end, call.parent.getStart(source)) +
        adaptedOperand + original.slice(call.parent.end);
}

if (require.main === module) {
    assert.equal(process.argv.length, 3, "usage: node adapt.cjs <tree>");
    const file = path.join(process.argv[2], "src/compiler/moduleNameResolver.ts");
    const original = fs.readFileSync(file, "utf8");
    const adapted = plan(original, file);
    if (adapted !== original) fs.writeFileSync(file, adapted);
}

module.exports = { findUnit, plan };
