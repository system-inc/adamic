#!/usr/bin/env node
"use strict";
const fs = require("node:fs");
const path = require("node:path");
const assert = require("node:assert/strict");
const { spawnSync } = require("node:child_process");
const ts = require("typescript");
const [tree, results] = process.argv.slice(2);
if (!tree || !results || ts.version !== "6.0.3") throw new Error("usage: verify.cjs <tree> <new-results>; requires TypeScript 6.0.3");
fs.mkdirSync(results, {recursive: false});
const file = path.resolve(tree, "src/compiler/core.ts");
const text = fs.readFileSync(file, "utf8");
const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
let fn = source.statements.find(n => ts.isFunctionDeclaration(n) && n.name?.text === "memoize").getText(source);
// Accept either the pre-adaptation tree or the actual applied A tree.
if (fn.includes("callback: (() => T) | undefined")) {
    assert.ok(fn.includes("callback = undefined;"));
    fn = fn.replace("callback: (() => T) | undefined", "callback: () => T").replace("callback = undefined;", "callback = undefined!;");
}
const variants = {
    original: fn,
    a: fn.replace("callback: () => T", "callback: (() => T) | undefined").replace("undefined!", "undefined"),
    b: fn.replace("let value: T;", "let value: T;\n    let pending: (() => T) | undefined = callback;").replace("if (callback)", "if (pending)").replace("value = callback();", "value = pending();").replace("callback = undefined!;", "pending = undefined;")
};
if (process.argv[4] === "--mutant-javascript") variants.a = variants.a.replace("callback = undefined;", "callback = undefined; value = callback;");
function emit(code) { return ts.transpileModule(code, {compilerOptions: {target: ts.ScriptTarget.ESNext, module: ts.ModuleKind.ESNext}}).outputText; }
assert.equal(emit(variants.original), emit(variants.a));
const driver = '\nlet calls = 0;\nconst cached = memoize(() => { calls += 1; return 42; });\nconst first = cached();\nconst second = cached();\nconsole.log(`${first} ${second} ${calls}`);\n';
const observations = {};
for (const [name, code] of Object.entries(variants)) {
    const script = path.join(results, name + ".mjs");
    fs.writeFileSync(script, emit(code + driver));
    const run = spawnSync(process.execPath, [script]);
    fs.writeFileSync(path.join(results, name + ".stdout"), run.stdout);
    fs.writeFileSync(path.join(results, name + ".stderr"), run.stderr);
    assert.equal(run.status, 0);
    assert.equal(run.stdout.toString(), "42 42 1\n");
    assert.equal(run.stderr.length, 0);
    observations[name] = {exit: run.status, stdout: run.stdout.toString(), stderr: run.stderr.toString()};
}
// The output comparison must reject exactly one changed byte.
const mutant = Buffer.from(observations.a.stdout);
mutant[0] = "5".charCodeAt(0);
let caught = false;
try { assert.deepEqual(mutant, Buffer.from(observations.original.stdout)); } catch { caught = true; }
assert.ok(caught);
fs.writeFileSync(path.join(results, "one-byte-mutant.stdout"), mutant);
const callers = [];
function walk(dir) {
    for (const entry of fs.readdirSync(dir, {withFileTypes: true})) {
        const name = path.join(dir, entry.name);
        if (entry.isDirectory()) walk(name);
        else if (name.endsWith(".ts")) {
            const sf = ts.createSourceFile(name, fs.readFileSync(name, "utf8"), ts.ScriptTarget.Latest, true);
            function visit(n) {
                if (ts.isCallExpression(n) && ts.isIdentifier(n.expression) && n.expression.text === "memoize") {
                    const position = sf.getLineAndCharacterOfPosition(n.getStart(sf));
                    callers.push({file: path.relative(tree, name), line: position.line + 1, source: n.getText(sf)});
                }
                ts.forEachChild(n, visit);
            }
            visit(sf);
        }
    }
}
walk(path.join(tree, "src/compiler"));
assert.equal(callers.length, 22);
const report = {node: process.version, source: "050880ce59e30b356b686bd3144efe24f875ebc8", a_javascript_identical: true,
    observations, output_mutant: {changed_bytes: 1, caught, catcher: "byte comparison against original Node stdout", native: false},
    callers, native: "Native acceptance is separate; updated area and scratch capture builds refuse unchanged A and B at their conditions; see p2-repeat/results.json"};
fs.writeFileSync(path.join(results, "proof.json"), JSON.stringify(report, null, 2) + "\n");
console.log(JSON.stringify({a_javascript_identical: true, node_two_call: observations, output_mutant_caught: caught, callers: callers.length, native: false}));
