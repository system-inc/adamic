#!/usr/bin/env node
"use strict";
const fs = require("node:fs");
const path = require("node:path");
const cp = require("node:child_process");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
if (ts.version !== "6.0.3") throw new Error(`want TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 7) throw new Error("usage: node wave-mutant.cjs <before-tree> <after-tree> <new-mutant-tree> <file> <expression>");
const [before, after, mutant] = process.argv.slice(2, 5).map(value => path.resolve(value));
const [file, expression] = process.argv.slice(5);
const sites = require("./sites.json").filter(site => site.file === file && site.expression === expression && site.action === "assert");
if (sites.length !== 1) throw new Error("mutant requires one ledgered assertion occurrence");
if (fs.existsSync(mutant)) throw new Error(`refusing existing mutant tree: ${mutant}`);
fs.mkdirSync(path.join(mutant, "src"), { recursive: true });
fs.cpSync(path.join(after, "src/compiler"), path.join(mutant, "src/compiler"), { recursive: true });
const name = path.join(mutant, "src/compiler", file);
const text = fs.readFileSync(name, "utf8");
const source = ts.createSourceFile(name, text, ts.ScriptTarget.Latest, true);
if (source.parseDiagnostics.length) throw new Error("cannot parse mutant input");
let target;
function visit(node) {
    if (ts.isNonNullExpression(node) && node.expression.getText(source) === expression) {
        if (target) throw new Error("duplicate mutant target");
        target = node;
    }
    ts.forEachChild(node, visit);
}
visit(source);
if (!target) throw new Error("missing mutant target");
fs.writeFileSync(name, text.slice(0, target.getStart(source)) + `(${expression} ?? 0)` + text.slice(target.end));
const results = {};
for (const [label, args] of Object.entries({ emit: ["verify.cjs", before, mutant], contract: ["adapt.cjs", "--check", mutant] })) {
    const result = cp.spawnSync(process.execPath, [path.join(__dirname, args[0]), ...args.slice(1)], { encoding: "utf8" });
    const expected = label === "emit" ? `unexpected emitted JavaScript change in ${file}` : `required read defaulted: ${file}:${sites[0].line}:${sites[0].column}`;
    if (result.status !== 1 || !result.stderr.includes(expected)) throw new Error(`mutant not caught by ${label}: ${result.stderr}`);
    results[label] = { exit: result.status, stdout: result.stdout, stderr: result.stderr };
}
console.log(JSON.stringify({ file, expression, results }, null, 2));
