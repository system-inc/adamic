#!/usr/bin/env node
"use strict";
const fs = require("node:fs");
const path = require("node:path");
const sites = require("./sites.json");
const { planWhole } = require("./whole-files.cjs");
const files = [
    "core.ts", "utilities.ts", "utilitiesPublic.ts", "parser.ts", "scanner.ts",
    "factory/baseNodeFactory.ts", "factory/emitHelpers.ts", "factory/emitNode.ts",
    "factory/nodeChildren.ts", "factory/nodeConverters.ts", "factory/nodeFactory.ts",
    "factory/nodeTests.ts", "factory/parenthesizerRules.ts", "factory/utilities.ts",
    "factory/utilitiesPublic.ts", "debug.ts", "path.ts",
];

// Addresses are parsed expressions plus their occurrence in the pinned file.
// Survey line/column fields are documentation, never edit offsets.
function plan(ts, file, text, check = false, zeroOnly = false) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    if (source.parseDiagnostics.length) throw new Error(`cannot parse ${file}`);
    const groups = new Map();
    function visit(node) {
        if (ts.isElementAccessExpression(node)) {
            const expression = node.getText(source).replaceAll("!", "");
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
        const group = groups.get(site.expression);
        if (!group || group.length !== site.total) throw new Error(`site drift: ${file}:${site.line} ${site.expression}`);
        const node = group[site.occurrence - 1];
        const parent = node.parent;
        const asserted = ts.isNonNullExpression(parent) && parent.expression === node;
        const coalesced = ts.isBinaryExpression(parent) && parent.left === node &&
            parent.operatorToken.kind === ts.SyntaxKind.QuestionQuestionToken;
        const zeroed = coalesced && ts.isNumericLiteral(parent.right) && parent.right.text === "0" &&
            ts.isParenthesizedExpression(parent.parent);
        const label = `${file}:${site.line}:${site.column}`;
        // Adaptation 46 owns this already-approved optional quote selection.
        // Keep the original declined capture reads on a fully adapted rerun.
        const quoteSelection = file === "parser.ts" && site.expression === "matchResult[2]" &&
            coalesced && parent.right.getText(source) === "matchResult[3]" &&
            ts.isParenthesizedExpression(parent.parent) && ts.isNonNullExpression(parent.parent.parent);
        if (site.action === "decline") {
            if (asserted || coalesced && !quoteSelection) throw new Error(`declined site changed: ${label}`);
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
    for (const edit of edits.sort((a, b) => b.at - a.at)) {
        text = text.slice(0, edit.at) + edit.text + text.slice(edit.at);
    }
    const closure = zeroOnly ? { text, edits: 0 } : planWhole(ts, file, text, check);
    return { text: closure.text, assertions, zeros, declined, closures: closure.edits };
}

function main() {
    const check = process.argv[2] === "--check";
    if (process.argv.length !== (check ? 4 : 3)) throw new Error("usage: node adapt.cjs [--check] <tree>");
    const tree = path.resolve(process.argv[check ? 3 : 2]);
    const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
    if (ts.version !== "6.0.3") throw new Error(`want TypeScript 6.0.3, got ${ts.version}`);
    // Validate every scoped file before writing any; retain original bytes and reads.
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
    console.log(JSON.stringify(plans.map(({ file, assertions, zeros, declined, closures }) => ({ file, assertions, zeros, declined, closures })), null, 2));
}
module.exports = { plan, files };
if (require.main === module) {
    try { main(); }
    catch (error) { console.error(error.message); process.exitCode = 1; }
}
