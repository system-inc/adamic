#!/usr/bin/env node
"use strict";
const fs = require("node:fs");
const path = require("node:path");
const { spawnSync } = require("node:child_process");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
const sites = [...require("./sites.json"), ...require("./required-values.json")];
const { files } = require("./adapt.cjs");
if (ts.version !== "6.0.3") throw new Error(`want TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 5) throw new Error("usage: node mutant.cjs <before-tree> <after-tree> <new-mutant-tree>");
const [before, after, mutant] = process.argv.slice(2).map(name => path.resolve(name));
if (fs.existsSync(mutant)) throw new Error("mutant output already exists");
for (const file of files) {
    const target = path.join(mutant, "src/compiler", file);
    fs.mkdirSync(path.dirname(target), { recursive: true });
    fs.copyFileSync(path.join(after, "src/compiler", file), target);
}
const site = sites.find(site => site.action === "assert" &&
    (!process.env.MUTANT_FILE || site.file === process.env.MUTANT_FILE) &&
    (!process.env.MUTANT_EXPRESSION || site.expression === process.env.MUTANT_EXPRESSION));
if (!site) throw new Error("no required assertion to mutate");
const target = path.join(mutant, "src/compiler", site.file);
const text = fs.readFileSync(target, "utf8");
const source = ts.createSourceFile(site.file, text, ts.ScriptTarget.Latest, true);
const nodes = [];
function visit(node) {
    if (ts.SyntaxKind[node.kind] === (site.kind || "ElementAccessExpression") && node.getText(source).replaceAll("!", "") === site.expression) nodes.push(node);
    ts.forEachChild(node, visit);
}
visit(source);
if (nodes.length !== site.total) throw new Error("mutant site drift");
const node = nodes[site.occurrence - 1];
if (!ts.isNonNullExpression(node.parent)) throw new Error("required read is not asserted");
fs.writeFileSync(target, text.slice(0, node.getStart(source)) + "(" + node.getText(source) + " ?? 0)" + text.slice(node.parent.end));
const checks = [
    ["emitted-JavaScript", "verify.cjs", before, mutant],
    ["site-contract", "adapt.cjs", "--check", mutant],
].map(([name, script, ...args]) => {
    const result = spawnSync(process.execPath, [path.join(__dirname, script), ...args], { encoding: "utf8" });
    if (result.status !== 1) throw new Error(`${name} did not catch mutant: ${result.status}`);
    if (!result.stderr.includes(name === "emitted-JavaScript" ? "unexpected emitted JavaScript change" : "required read defaulted")) {
        throw new Error(`${name} failed for an unrelated reason: ${result.stderr}`);
    }
    return { name, exit: result.status, stdout: result.stdout, stderr: result.stderr };
});
console.log(JSON.stringify({ site, checks }, null, 2));
