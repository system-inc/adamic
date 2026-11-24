"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const ts = require("typescript");
const { nodes } = require("../47-host-errors/adapt.cjs");

const fileName = path.join(process.argv[2], "src/compiler/utilities.ts");
const before = fs.readFileSync(fileName, "utf8");
const sites = [
    ["createModeMismatchDetails", "() => true", '(): import("./types.js").RepopulateDiagnosticChainInfo => true'],
    ["getEnclosingContainer", "n => !!(getContainerFlags(n) & ContainerFlags.IsContainer)",
        '(n): boolean | "quit" => !!(getContainerFlags(n) & ContainerFlags.IsContainer)'],
    ["getEnclosingBlockScopeContainer", "current => isBlockScope(current, current.parent)",
        '(current): boolean | "quit" => isBlockScope(current, current.parent)'],
];

function locate(text, owner, original, adapted) {
    const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, fileName);
    const owners = source.statements.filter(node =>
        ts.isFunctionDeclaration(node) && node.name?.text === owner);
    assert.equal(owners.length, 1, owner);
    const arrows = nodes(ts, owners[0].body, node => {
        if (!ts.isArrowFunction(node)) return false;
        const parent = node.parent;
        if (owner === "createModeMismatchDetails") {
            return ts.isBinaryExpression(parent) && parent.right === node
                && parent.operatorToken.kind === ts.SyntaxKind.EqualsToken
                && parent.left.getText(source) === "result.repopulateInfo";
        }
        return ts.isCallExpression(parent) && parent.expression.getText(source) === "findAncestor"
            && parent.arguments[0]?.getText(source) === "node.parent" && parent.arguments[1] === node;
    });
    assert.equal(arrows.length, 1, owner);
    const arrow = arrows[0];
    const snapshot = arrow.getText(source);
    assert(snapshot === original || snapshot === adapted, "unreviewed callback " + snapshot);
    return { start: arrow.getStart(source), end: arrow.end, snapshot };
}

let after = before;
for (const [owner, original, adapted] of sites) {
    const { start, end, snapshot } = locate(after, owner, original, adapted);
    if (snapshot === original) after = after.slice(0, start) + adapted + after.slice(end);
    assert.equal(locate(after, owner, original, adapted).snapshot, adapted);
}
if (after !== before) {
    const options = { target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext };
    const js = text => ts.transpileModule(text, { compilerOptions: options, fileName }).outputText;
    let expected = js(before);
    for (const [, original] of sites.slice(1)) {
        const [parameter, body] = original.split(" => ");
        const oldLine = "findAncestor(node.parent, " + original;
        const newLine = "findAncestor(node.parent, (" + parameter + ") => " + body;
        if (expected.includes(oldLine)) {
            assert.equal(expected.split(oldLine).length, 2, oldLine);
            expected = expected.replace(oldLine, newLine);
        } else {
            assert.equal(expected.split(newLine).length, 2, newLine);
        }
    }
    assert.equal(js(after), expected, "utilities.ts JavaScript changed beyond the ruled parentheses");
    const declaration = text => ts.transpileDeclaration(text, {
        compilerOptions: { stripInternal: true }, fileName,
    }).outputText;
    assert.equal(declaration(after), declaration(before), "public declaration output changed");
    fs.writeFileSync(fileName, after);
}
