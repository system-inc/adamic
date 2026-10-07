#!/usr/bin/env node
"use strict";
const fs = require("node:fs");
const path = require("node:path");
const sites = require("./sites.json");
const files = ["corePublic.ts", "performanceCore.ts", "types.ts", "expressionToTypeNode.ts", "executeCommandLine.ts", "builder.ts", "builderPublic.ts", "builderState.ts", "builderStatePublic.ts", "watch.ts", "watchPublic.ts", "watchUtilities.ts", "resolutionCache.ts", "tsbuild.ts", "tsbuildPublic.ts", "moduleNameResolver.ts", "moduleSpecifiers.ts", "sys.ts", "program.ts", "commandLineParser.ts", "binder.ts", "semver.ts", "programDiagnostics.ts", "tracing.ts", "symbolWalker.ts", "performance.ts"];

// Addresses are parsed expressions plus their occurrence in the pinned file.
// Survey line/column fields are documentation, never edit offsets.
function plan(ts, file, text, check = false, zeroOnly = false) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    if (source.parseDiagnostics.length) throw new Error(`cannot parse ${file}`);
    const groups = new Map();
    const laterExpressions = new Map(sites.filter(s => s.file === file && s.laterExpression).map(s => [s.laterExpression, s.expression]));
    const laterCaptures = new Set();
    function visit(node) {
        if (ts.isElementAccessExpression(node)) {
            let expression = node.getText(source).replaceAll("!", "");
            if (file === "semver.ts" && expression === "match[1]") {
                let owner = node.parent;
                while (owner && !(ts.isFunctionDeclaration(owner) && owner.name)) owner = owner.parent;
                if (owner && ["tryParseComponents", "parsePartial"].includes(owner.name.text)) {
                    if (!ts.isNonNullExpression(node.parent) || !ts.isVariableDeclaration(node.parent.parent) || node.parent.parent.name.getText(source) !== "major" || laterCaptures.has(owner.name.text)) throw new Error("adaptation 45 capture drift");
                    laterCaptures.add(owner.name.text);
                    return;
                }
            }
            expression = laterExpressions.get(expression) || expression;
            const group = groups.get(expression) || [];
            group.push(node);
            groups.set(expression, group);
        }
        ts.forEachChild(node, visit);
    }
    visit(source);
    if (laterCaptures.size !== 0 && laterCaptures.size !== 2) throw new Error("incomplete adaptation 45 captures");
    const fileSites = sites.filter(site => site.file === file);
    const indexedCount = [...groups.values()].reduce((count, group) => count + group.length, 0);
    if (indexedCount !== fileSites.length) throw new Error(`indexed inventory drift: ${file}: ${indexedCount} != ${fileSites.length}`);
    const edits = [];
    let assertions = 0, zeros = 0, declined = 0;
    for (const site of fileSites) {
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
        if (site.action === "decline") {
            if (site.existingDefault !== undefined) {
                if (asserted || !coalesced || parent.right.getText(source) !== site.existingDefault) throw new Error(`existing default changed: ${label}`);
            }
            else if ((asserted && site.laterAssertion !== "47-host-errors") || coalesced) throw new Error(`declined site changed: ${label}`);
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
    if (file === "watchUtilities.ts") require("./watch-protocol.cjs").validate(ts, source);
    if (file === "moduleSpecifiers.ts") require("./nonempty-endings.cjs").validate(ts, source);
    for (const edit of edits.sort((a, b) => b.at - a.at)) {
        text = text.slice(0, edit.at) + edit.text + text.slice(edit.at);
    }
    const closure = require("./closure.cjs").plan(ts, file, text, check);
    const handoff = require("./handoffs.cjs").plan(ts, file, closure.text, check);
    return { ...handoff, contracts: closure.contracts + handoff.contracts, assertions, zeros, declined };
}

function main() {
    const check = process.argv[2] === "--check";
    if (process.argv.length !== (check ? 4 : 3)) throw new Error("usage: node adapt.cjs [--check] <tree>");
    const tree = path.resolve(process.argv[check ? 3 : 2]);
    const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
    if (ts.version !== "6.0.3") throw new Error(`want TypeScript 6.0.3, got ${ts.version}`);
    require("./handoffs.cjs").validate(ts, tree);
    require("./public-host-guards.cjs").validate(ts, tree);
    // Validate all selected files before writing any; retain original bytes and reads.
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
    console.log(JSON.stringify(plans.map(({ file, assertions, zeros, declined, contracts }) => ({ file, assertions, zeros, declined, contracts })), null, 2));
}
module.exports = { plan, files };
if (require.main === module) {
    try { main(); }
    catch (error) { console.error(error.message); process.exitCode = 1; }
}
