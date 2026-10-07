#!/usr/bin/env node
"use strict";
const fs = require("node:fs");
const path = require("node:path");
const cp = require("node:child_process");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
if (ts.version !== "6.0.3") throw new Error(`want TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 7 && process.argv.length !== 8) throw new Error("usage: node wave-mutant.cjs <before-tree> <after-tree> <new-mutant-tree> <file> <expression> [occurrence]");
const [before, after, mutant] = process.argv.slice(2, 5).map(value => path.resolve(value));
const [file, expression] = process.argv.slice(5);
const occurrence = process.argv[7] === undefined ? undefined : Number(process.argv[7]);
if (occurrence !== undefined && (!Number.isInteger(occurrence) || occurrence < 1)) throw new Error("invalid mutant occurrence");
const sites = require("./sites.json").filter(site => site.file === file && site.expression === expression && site.action === "assert" && (occurrence === undefined || site.occurrence === occurrence));
if (sites.length !== 1) throw new Error("mutant requires one ledgered assertion occurrence");
if (fs.existsSync(mutant)) throw new Error(`refusing existing mutant tree: ${mutant}`);
fs.mkdirSync(path.join(mutant, "src"), { recursive: true });
fs.cpSync(path.join(after, "src/compiler"), path.join(mutant, "src/compiler"), { recursive: true });
const name = path.join(mutant, "src/compiler", file);
const text = fs.readFileSync(name, "utf8");
const source = ts.createSourceFile(name, text, ts.ScriptTarget.Latest, true);
if (source.parseDiagnostics.length) throw new Error("cannot parse mutant input");
const matches = [];
function visit(node) {
    if (ts.isElementAccessExpression(node) && node.getText(source).replaceAll("!", "") === expression) matches.push(node);
    ts.forEachChild(node, visit);
}
visit(source);
if (matches.length !== sites[0].total) throw new Error("mutant occurrence count drift");
const read = matches[sites[0].occurrence - 1];
if (!read || !ts.isNonNullExpression(read.parent)) throw new Error("missing mutant assertion");
const target = read.parent;
fs.writeFileSync(name, text.slice(0, target.getStart(source)) + `(${expression} ?? 0)` + text.slice(target.end));
const results = {};
for (const [label, args] of Object.entries({ emit: ["verify.cjs", before, mutant], contract: ["adapt.cjs", "--check", mutant] })) {
    const result = cp.spawnSync(process.execPath, [path.join(__dirname, args[0]), ...args.slice(1)], { encoding: "utf8" });
    const expected = label === "emit" ? `unexpected emitted JavaScript change in ${file}` : `required read defaulted: ${file}:${sites[0].line}:${sites[0].column}`;
    if (result.status !== 1 || !result.stderr.includes(expected)) throw new Error(`mutant not caught by ${label}: ${result.stderr}`);
    results[label] = { exit: result.status, stdout: result.stdout, stderr: result.stderr };
}
console.log(JSON.stringify({ file, expression, results }, null, 2));
