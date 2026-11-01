#!/usr/bin/env node
"use strict";
// 98-axolotl-b: required indexed reads in transformers/es2015.ts.
// Each read in sites.json gains a `!` at its original evaluation point: Node erases it, Adamic checks it loudly.
// Sites are parsed expressions plus their occurrence; line and column fields are documentation, never offsets.
const fs = require("node:fs");
const path = require("node:path");
const sites = require("./sites.json");

const files = [...new Set(sites.map(site => site.file))];

// The nearest named function declaration. Every es2015.ts site sits directly in one (no arrow, method or function
// expression between), so 98-a's walk is kept unchanged.
function enclosingFunction(ts, node) {
    for (let current = node.parent; current; current = current.parent) {
        if (ts.isFunctionDeclaration(current) && current.name) return current.name.text;
    }
    return undefined;
}

// Plans one file's edits. Throws on any shape the ledger does not inventory: a missing or extra occurrence, a read
// moved to another function, a parent other than the recorded one, or a default where an assertion belongs.
function plan(ts, file, text) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    if (source.parseDiagnostics.length) throw new Error(`cannot parse ${file}`);
    const groups = new Map();
    (function visit(node) {
        if (ts.isElementAccessExpression(node)) {
            const expression = node.getText(source);
            if (!groups.has(expression)) groups.set(expression, []);
            groups.get(expression).push(node);
        }
        ts.forEachChild(node, visit);
    })(source);
    const edits = [];
    const states = [];
    for (const site of sites.filter(site => site.file === file)) {
        const label = `${file}:${site.line}:${site.column} ${site.expression}`;
        const group = groups.get(site.expression) || [];
        if (group.length !== site.total) throw new Error(`site drift (${group.length} occurrences, want ${site.total}): ${label}`);
        const node = group[site.occurrence - 1];
        if (enclosingFunction(ts, node) !== site.function) throw new Error(`site moved out of ${site.function}: ${label}`);
        if (site.action !== "assert") throw new Error(`unknown action ${site.action}: ${label}`);
        const asserted = ts.isNonNullExpression(node.parent) && node.parent.expression === node;
        const outer = asserted ? node.parent : node;
        if (ts.SyntaxKind[outer.parent.kind] !== site.parent) {
            throw new Error(`unexpected parent ${ts.SyntaxKind[outer.parent.kind]}, want ${site.parent}: ${label}`);
        }
        if (asserted && ts.isNonNullExpression(outer.parent)) throw new Error(`doubled assertion: ${label}`);
        states.push(asserted);
        if (!asserted) edits.push({ at: node.end, text: "!" });
    }
    let changed = text;
    for (const edit of edits.sort((a, b) => b.at - a.at)) {
        changed = changed.slice(0, edit.at) + edit.text + changed.slice(edit.at);
    }
    return { text: changed, states, assertions: edits.length };
}

function main() {
    if (process.argv.length !== 3) throw new Error("usage: node adapt.cjs <tree>");
    const tree = path.resolve(process.argv[2]);
    const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
    if (ts.version !== "6.0.3") throw new Error(`want TypeScript 6.0.3, got ${ts.version}`);
    // Plan and prove every file before writing any.
    const plans = files.map(file => {
        const name = path.join(tree, "src/compiler", file);
        const before = fs.readFileSync(name, "utf8");
        return { file, name, before, ...plan(ts, file, before) };
    });
    const states = plans.flatMap(item => item.states);
    if (new Set(states).size !== 1) throw new Error("partially applied: some sites asserted, others not");
    for (const item of plans) {
        // `!` is type syntax only: the JavaScript must not move by a byte, with comments kept or removed.
        for (const removeComments of [false, true]) {
            const options = { fileName: item.name, compilerOptions: { target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments } };
            if (ts.transpileModule(item.before, options).outputText !== ts.transpileModule(item.text, options).outputText) {
                throw new Error(`runtime JavaScript changed: ${item.file}`);
            }
        }
        if (fs.readFileSync(item.name, "utf8") !== item.before) throw new Error(`source changed while planning: ${item.name}`);
    }
    for (const item of plans) {
        if (item.text !== item.before) fs.writeFileSync(item.name, item.text);
    }
    console.log(JSON.stringify(plans.map(({ file, assertions }) => ({ file, assertions })), null, 2));
}

module.exports = { plan, files };
if (require.main === module) {
    try { main(); }
    catch (error) { console.error(error.message); process.exitCode = 1; }
}
