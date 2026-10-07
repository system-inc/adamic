#!/usr/bin/env node
"use strict";
const fs = require("node:fs");
const path = require("node:path");
const sites = [...require("./sites.json"), ...require("./required-values.json")];
const files = require("./files.json");
const typeEdits = require("./type-edits.json");

// Addresses are parsed expressions plus their occurrence in the pinned file.
// Survey line/column fields are documentation, never edit offsets.
function plan(ts, file, text, check = false, zeroOnly = false) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    if (source.parseDiagnostics.length) throw new Error(`cannot parse ${file}`);
    const groups = new Map();
    function visit(node) {
        if (ts.isElementAccessExpression(node) || ts.isCallExpression(node) || ts.isPropertyAccessExpression(node) || ts.isIdentifier(node)) {
            const expression = ts.SyntaxKind[node.kind] + ":" + node.getText(source).replaceAll("!", "");
            const group = groups.get(expression) || [];
            group.push(node);
            groups.set(expression, group);
        }
        ts.forEachChild(node, visit);
    }
    visit(source);
    const edits = [];
    let assertions = 0, zeros = 0, declined = 0;
    for (const site of sites.filter(site => site.file === file)) {
        const group = groups.get((site.kind || "ElementAccessExpression") + ":" + site.expression);
        if (!group || group.length !== site.total) throw new Error(`site drift: ${file}:${site.line} ${site.expression}`);
        const node = group[site.occurrence - 1];
        const parent = node.parent;
        const asserted = ts.isNonNullExpression(parent) && parent.expression === node;
        const coalesced = ts.isBinaryExpression(parent) && parent.left === node &&
            parent.operatorToken.kind === ts.SyntaxKind.QuestionQuestionToken;
        const zeroed = coalesced && ts.isNumericLiteral(parent.right) && parent.right.text === "0" &&
            ts.isParenthesizedExpression(parent.parent);
        const label = `${file}:${site.line}:${site.column}`;
        if (site.action === "decline") {
            if (asserted || coalesced) throw new Error(`declined site changed: ${label}`);
            declined++;
        }
        else if (site.action === "assert") {
            if (coalesced) throw new Error(`required read defaulted: ${label}`);
            if (!zeroOnly && !asserted) {
                if (check) throw new Error(`required assertion missing: ${label}`);
                edits.push({ at: node.end, text: "!" });
                assertions++;
            }
        }
        else if (site.action === "zero") {
            if (asserted || coalesced && !zeroed) throw new Error(`invalid U-zero adaptation: ${label}`);
            if (!zeroed) {
                if (check) throw new Error(`U-zero adaptation missing: ${label}`);
                edits.push({ at: node.getStart(source), text: "(" }, { at: node.end, text: " ?? 0)" });
                zeros++;
            }
        }
        else throw new Error(`unknown action: ${site.action}`);
    }
    for (const site of typeEdits.filter(site => site.file === file)) {
        const matches = [];
        function find(node) {
            if (site.kind === "VariableDeclaration" && ts.isVariableDeclaration(node) && node.name.getText(source) === site.name) matches.push(node);
            if (site.kind === "CallExpression" && ts.isCallExpression(node) && node.expression.getText(source) === site.name &&
                JSON.stringify(node.arguments.map(arg => arg.getText(source))) === JSON.stringify(site.arguments)) matches.push(node);
            ts.forEachChild(node, find);
        }
        find(source);
        if (matches.length !== site.total) throw new Error(`type site drift: ${file}:${site.name}`);
        const node = matches[site.occurrence - 1];
        if (site.kind === "VariableDeclaration") {
            if (!node.type) throw new Error(`missing owner type: ${file}:${site.name}`);
            const actual = node.type.getText(source);
            if (actual !== site.before && actual !== site.after) throw new Error(`unexpected owner type: ${file}:${site.name}`);
            if (!zeroOnly && actual !== site.after) {
                if (check) throw new Error(`optional owner missing: ${file}:${site.name}`);
                edits.push({at: node.type.end, text: " | undefined"});
            }
        }
        else {
            const args = node.typeArguments;
            if (args && (args.length !== 1 || args[0].getText(source) !== site.typeArgument)) throw new Error(`unexpected overload type: ${file}:${site.name}`);
            if (!zeroOnly && !args) {
                if (check) throw new Error(`overload type missing: ${file}:${site.name}`);
                edits.push({at: node.expression.end, text: `<${site.typeArgument}>`});
            }
        }
    }
    for (const edit of edits.sort((a, b) => b.at - a.at)) {
        text = text.slice(0, edit.at) + edit.text + text.slice(edit.at);
    }
    return { text, assertions, zeros, declined };
}

function main() {
    const check = process.argv[2] === "--check";
    if (process.argv.length !== (check ? 4 : 3)) throw new Error("usage: node adapt.cjs [--check] <tree>");
    const tree = path.resolve(process.argv[check ? 3 : 2]);
    const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
    if (ts.version !== "6.0.3") throw new Error(`want TypeScript 6.0.3, got ${ts.version}`);
    // Validate all owned files before writing any; retain original bytes and reads.
    const plans = files.map(file => {
        const name = path.join(tree, "src/compiler", file);
        const before = fs.readFileSync(name, "utf8");
        return { file, name, before, ...plan(ts, file, before, check) };
    });
    for (const item of plans) {
        if (fs.readFileSync(item.name, "utf8") !== item.before) throw new Error(`source changed: ${item.name}`);
    }
    for (const item of plans) {
        if (!check && item.text !== item.before) fs.writeFileSync(item.name, item.text);
    }
    console.log(JSON.stringify(plans.map(({ file, assertions, zeros, declined }) => ({ file, assertions, zeros, declined })), null, 2));
}
module.exports = { plan, files };
if (require.main === module) {
    try { main(); }
    catch (error) { console.error(error.message); process.exitCode = 1; }
}
