#!/usr/bin/env node
"use strict";

const fs = require("node:fs");
const path = require("node:path");

function adapt(tree, ts) {
    if (ts.version !== "6.0.3") throw new Error(`want stock typescript@6.0.3, got ${ts.version}`);
    const file = path.join(tree, "src/compiler/parser.ts");
    const text = fs.readFileSync(file, "utf8");
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    if (source.parseDiagnostics.length) throw new Error("cannot parse parser.ts");
    function one(root, predicate, label) {
        const found = [];
        function visit(node) {
            if (predicate(node)) found.push(node);
            ts.forEachChild(node, visit);
        }
        visit(root);
        if (found.length !== 1) throw new Error(`${label}: want one node, got ${found.length}`);
        return found[0];
    }
    const constructorFunction = one(source, node => ts.isFunctionDeclaration(node) && node.name?.text === "getNamedArgRegEx", "named argument matcher");
    const constructor = one(constructorFunction, node => ts.isVariableDeclaration(node) && node.name.getText() === "result", "matcher constructor").initializer;
    const template = constructor?.arguments?.[0];
    if (!ts.isNewExpression(constructor) || constructor.expression.getText() !== "RegExp" || !template ||
        !ts.isTemplateExpression(template) || template.templateSpans.length !== 1 ||
        template.head.text !== "(\\s" || template.templateSpans[0].expression.getText() !== "name" ||
        template.templateSpans[0].literal.text !== "\\s*=\\s*)(?:(?:'([^']*)')|(?:\"([^\"]*)\"))" ||
        constructor.arguments[1]?.text !== "im") throw new Error("unreviewed quote alternatives");
    const definitionsFile = path.join(tree, "src/compiler/types.ts");
    const definitions = ts.createSourceFile(definitionsFile, fs.readFileSync(definitionsFile, "utf8"), ts.ScriptTarget.Latest, true);
    if (definitions.parseDiagnostics.length) throw new Error("cannot parse pragma definitions");
    const pragmas = one(definitions, node => ts.isVariableDeclaration(node) && node.name.getText() === "commentPragmas", "pragma definitions");
    let names = 0;
    function checkName(node) {
        if (ts.isPropertyAssignment(node) && node.name.getText() === "name") {
            names++;
            if (!ts.isStringLiteral(node.initializer) || !/^[a-z-]+$/.test(node.initializer.text)) throw new Error("pragma argument names may inject regex syntax");
        }
        ts.forEachChild(node, checkName);
    }
    checkName(pragmas);
    if (!names) throw new Error("no fixed pragma argument names");
    const extract = one(source, node => ts.isFunctionDeclaration(node) && node.name?.text === "extractPragmas", "extractPragmas");
    const value = one(extract, node => ts.isVariableDeclaration(node) && node.name.getText() === "value", "argument value").initializer;
    if (!ts.isNonNullExpression(value) || !ts.isParenthesizedExpression(value.expression)) throw new Error("requires adaptation 45's whole-expression boundary");
    const expression = value.expression.expression;
    function capture(node, index) {
        return ts.isElementAccessExpression(node) && node.expression.getText() === "matchResult" &&
            ts.isNumericLiteral(node.argumentExpression) && Number(node.argumentExpression.text) === index;
    }
    if (!ts.isBinaryExpression(expression) || !capture(expression.left, 2) || !capture(expression.right, 3) ||
        ![ts.SyntaxKind.BarBarToken, ts.SyntaxKind.QuestionQuestionToken].includes(expression.operatorToken.kind)) {
        throw new Error("unreviewed argument selection");
    }
    if (expression.operatorToken.kind === ts.SyntaxKind.QuestionQuestionToken) return { files: 0, edits: 0 };
    const start = expression.operatorToken.getStart(source);
    const result = text.slice(0, start) + "??" + text.slice(expression.operatorToken.end);
    if (ts.createSourceFile(file, result, ts.ScriptTarget.Latest, true).parseDiagnostics.length) throw new Error("adapted syntax error");
    if (fs.readFileSync(file, "utf8") !== text) throw new Error("parser changed while planning");
    fs.writeFileSync(file, result);
    const where = source.getLineAndCharacterOfPosition(start);
    return { files: 1, edits: 1, file: "parser.ts", line: where.line + 1,
        before: "(matchResult[2] || matchResult[3])!", after: "(matchResult[2] ?? matchResult[3])!" };
}

module.exports = { adapt };
if (require.main === module) {
    try {
        if (process.argv.length !== 3) throw new Error("usage: node adapt.cjs <tree>");
        console.log(JSON.stringify(adapt(path.resolve(process.argv[2]), require(process.env.CENSUS_TYPESCRIPT || "typescript")), null, 2));
    }
    catch (error) { console.error(error.message); process.exitCode = 1; }
}
