#!/usr/bin/env node
"use strict";
const fs = require("node:fs");
const path = require("node:path");
const cp = require("node:child_process");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
if (ts.version !== "6.0.3") throw new Error(`want TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 5) throw new Error("usage: node mutant.cjs <before-tree> <after-tree> <new-mutant-tree>");
const [before, after, mutant] = process.argv.slice(2).map(value => path.resolve(value));
if (fs.existsSync(mutant)) throw new Error(`refusing existing mutant tree: ${mutant}`);
fs.mkdirSync(path.join(mutant, "src"), { recursive: true });
fs.cpSync(path.join(after, "src/compiler"), path.join(mutant, "src/compiler"), { recursive: true });
const name = path.join(mutant, "src/compiler/semver.ts");
const text = fs.readFileSync(name, "utf8");
const source = ts.createSourceFile(name, text, ts.ScriptTarget.Latest, true);
if (source.parseDiagnostics.length) throw new Error("cannot parse mutant input");
let target;
function visit(node) {
    if (ts.isNonNullExpression(node) && node.expression.getText(source) === "left[i]") {
        if (target) throw new Error("duplicate mutant target");
        target = node;
    }
    ts.forEachChild(node, visit);
}
visit(source);
if (!target) throw new Error("missing mutant target");
fs.writeFileSync(name, text.slice(0, target.getStart(source)) + "(left[i] ?? 0)" + text.slice(target.end));
const results = {};
for (const [label, args] of Object.entries({
    emit: ["verify.cjs", before, mutant],
    contract: ["adapt.cjs", "--check", mutant],
})) {
    const result = cp.spawnSync(process.execPath, [path.join(__dirname, args[0]), ...args.slice(1)], { encoding: "utf8" });
    const expected = label === "emit" ? "unexpected emitted JavaScript change in semver.ts" : "required read defaulted: semver.ts:171:32";
    if (result.status !== 1 || !result.stderr.includes(expected)) throw new Error(`mutant not caught by ${label}: ${result.stderr}`);
    results[label] = { exit: result.status, stdout: result.stdout, stderr: result.stderr };
}
console.log(JSON.stringify(results, null, 2));
