"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const ts = require("typescript");

const name = "tagNamesAreEquivalent";
const snapshotHead = [
    "export function tagNamesAreEquivalent(lhs: JsxTagNameExpression, rhs: JsxTagNameExpression): boolean {",
    "    if (lhs.kind !== rhs.kind) {",
    "        return false;",
    "    }",
    "",
    "    if (lhs.kind === SyntaxKind.Identifier) {",
    "        return lhs.escapedText === (rhs as Identifier).escapedText;",
    "    }",
    "",
    "    if (lhs.kind === SyntaxKind.ThisKeyword) {",
    "        return true;",
    "    }",
    "",
    "    if (lhs.kind === SyntaxKind.JsxNamespacedName) {",
    "        return lhs.namespace.escapedText === (rhs as JsxNamespacedName).namespace.escapedText &&",
    "            lhs.name.escapedText === (rhs as JsxNamespacedName).name.escapedText;",
    "    }",
    "",
    "    // If we are at this statement then we must have PropertyAccessExpression and because tag name in Jsx element can only",
    "    // take forms of JsxTagNameExpression which includes an identifier, \"this\" expression, or another propertyAccessExpression",
    "    // it is safe to case the expression property as such. See parseJsxElementName for how we parse tag name in Jsx element",
    "    ",
].join("\r\n");
const original = [
    "return (lhs as PropertyAccessExpression).name.escapedText === (rhs as PropertyAccessExpression).name.escapedText &&",
    "        tagNamesAreEquivalent((lhs as PropertyAccessExpression).expression as JsxTagNameExpression, (rhs as PropertyAccessExpression).expression as JsxTagNameExpression);",
].join("\r\n");
const replacement = [
    "return lhs.kind === SyntaxKind.PropertyAccessExpression && rhs.kind === SyntaxKind.PropertyAccessExpression &&",
    "        lhs.name.escapedText === rhs.name.escapedText && tagNamesAreEquivalent(lhs.expression, rhs.expression);",
].join("\r\n");

function findSite(fileName, text) {
    const file = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
    assert.equal(file.parseDiagnostics.length, 0, "parser.ts must parse");
    const functions = file.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === name);
    assert.equal(functions.length, 1, `expected one ${name} declaration`);
    const declaration = functions[0];
    const statement = declaration.body?.statements.at(-1);
    assert(statement && ts.isReturnStatement(statement), "expected the closing return");
    return { start: statement.getStart(file), end: statement.end, declaration: declaration.getText(file) };
}

function adapt(tree) {
    assert.equal(ts.version, "6.0.3");
    const fileName = path.join(tree, "src/compiler/parser.ts");
    const before = fs.readFileSync(fileName, "utf8");
    const site = findSite(fileName, before);
    const current = before.slice(site.start, site.end);
    // Pin the guards and the final return together, since the proof depends on both.
    assert.equal(site.declaration, snapshotHead + current + "\r\n}", `unreviewed ${name} guards`);
    if (current === replacement) return 0;
    assert.equal(current, original, `unreviewed ${name} return`);
    const after = before.slice(0, site.start) + replacement + before.slice(site.end);
    assert.equal(findSite(fileName, after).declaration, snapshotHead + replacement + "\r\n}");
    fs.writeFileSync(fileName, after);
    return 1;
}

if (require.main === module) {
    assert.equal(process.argv.length, 3, "usage: node adapt.cjs <tree>");
    console.log(JSON.stringify({ files: adapt(path.resolve(process.argv[2])), javascriptChanging: true }));
}

module.exports = { adapt };
