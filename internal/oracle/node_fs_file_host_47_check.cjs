#!/usr/bin/env node
"use strict";
// Extract actual adapted tsc functions, never rewrite the recorded Node truth.
const fs = require("node:fs");
const path = require("node:path");
const assert = require("node:assert/strict");
const crypto = require("node:crypto");
const { spawnSync } = require("node:child_process");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
const { nodes } = require("../../stage3/adapt/47-host-errors/adapt.cjs");
const [tree, compiler, logs, reportPath] = process.argv.slice(2);
assert(tree && compiler && logs && reportPath, "usage: check.cjs <adapted-tsc-tree> <adamic> <logs> <report>");
const repo = path.resolve(__dirname, "../..");
const bucket = path.join(repo, "stage3/fixtures/host");
fs.mkdirSync(logs, { recursive: true });
const generated = path.join(logs, "fixtures");
fs.mkdirSync(generated, { recursive: true });
const sysText = fs.readFileSync(path.join(tree, "src/compiler/sys.ts"), "utf8");
const helperText = fs.readFileSync(path.join(tree, "src/compiler/hostErrors.ts"), "utf8");
const sys = parse("sys.ts", sysText);
const helper = parse("hostErrors.ts", helperText);
function parse(name, text) { return ts.createSourceFile(name, text, ts.ScriptTarget.Latest, true); }
function one(root, predicate) { const found = nodes(ts, root, predicate); assert.equal(found.length, 1); return found[0]; }
const readFile = one(sys, node => ts.isFunctionDeclaration(node) && node.name?.text === "readFile").getText(sys);
const createDirectory = one(sys, node => ts.isMethodDeclaration(node) && node.name.getText(sys) === "createDirectory").getText(sys);
const errorCode = one(helper, node => ts.isFunctionDeclaration(node) && node.name?.text === "errorCode").getText(helper).replace("export ", "");
function observe(command, label) {
    const result = spawnSync(command[0], command.slice(1), { cwd: repo, encoding: "utf8" });
    if (result.error || result.signal) throw result.error || new Error(result.signal);
    for (const [name, value] of Object.entries({ stdout: result.stdout, stderr: result.stderr, exit: String(result.status) + "\n" })) fs.writeFileSync(path.join(logs, label + "." + name), value);
    return { stdout: result.stdout, stderr: result.stderr, exit: result.status };
}
function onNode(file, label) { return observe([process.execPath, "--disable-warning=ExperimentalWarning", "oracle/node.mjs", file], label); }
function stage(built) {
    const text = built.stdout + built.stderr;
    if (built.exit === 0) return "Compiles";
    if (text.includes(": error TS")) return "Checker";
    if (text.includes("Adamic 0.1 refuses")) return "Refused";
    if (text.includes("stage 0 can't lower")) return "NotYet";
    throw new Error("unexpected compiler/tool failure: " + text);
}
const results = [];
for (const row of JSON.parse(fs.readFileSync(path.join(bucket, "status.json"), "utf8"))) {
    if (!/^(0[1-4]|13)_/.test(row.file)) continue;
    let source = fs.readFileSync(path.join(bucket, row.file), "utf8");
    const original = parse(row.file, source);
    const isDirectory = row.file.startsWith("13_");
    const target = one(original, node => isDirectory ? ts.isMethodDeclaration(node) && node.name.getText(original) === "createDirectory" : ts.isFunctionDeclaration(node) && node.name?.text === "readFile");
    source = source.slice(0, target.getStart(original)) + (isDirectory ? createDirectory : readFile) + source.slice(target.end);
    if (isDirectory) source = errorCode + "\n" + source;
    const fixture = path.join(generated, row.file);
    fs.writeFileSync(fixture, source);
    const truth = onNode(fixture, row.file + ".node");
    assert.deepEqual(truth, row.node, row.file + ": adapted source differs from recorded Node truth");
    const result = { file: row.file, nodeMatches: true, adaptedSourceSha256: crypto.createHash("sha256").update(source).digest("hex") };
    for (const backend of ["native", "javascript"]) {
        const binary = path.join(logs, row.file + ".binary");
        const built = observe([compiler, backend === "native" ? "build" : "js", fixture, ...(backend === "native" ? ["-o", binary] : [])], row.file + "." + backend + "-build");
        result[backend] = { stage: stage(built), diagnostic: built.exit === 0 ? "" : built.stdout + built.stderr };
        if (built.exit === 0) {
            const module = path.join(generated, row.file + ".mjs");
            if (backend === "javascript") fs.writeFileSync(module, built.stdout);
            const got = backend === "native" ? observe([binary], row.file + ".native") : onNode(module, row.file + ".javascript");
            assert.deepEqual(got, truth, "SILENT MISCOMPILE: " + row.file + " " + backend);
            result[backend].agrees = true;
        }
    }
    let mutated;
    if (isDirectory) mutated = source.replaceAll("error.code", "error.message");
    else {
        const changes = {
            "01": ['return buffer.toString("utf8", 3);', 'return buffer.toString("utf8", 0);'],
            "02": ['len >= 2 && buffer[0]! === 0xFF && buffer[1]! === 0xFE', 'false'],
            "03": ['buffer[i] = buffer[i + 1]!;', 'buffer[i] = temp;'],
            "04": ['return undefined;', 'return "";'],
        };
        const [before, after] = changes[row.file.slice(0, 2)];
        assert(source.includes(before), row.file + ": mutation site drift");
        mutated = source.replace(before, after);
    }
    assert.notEqual(mutated, source);
    const mutant = path.join(generated, "mutant_" + row.file);
    fs.writeFileSync(mutant, mutated);
    assert.notDeepEqual(onNode(mutant, row.file + ".mutant"), truth, "SURVIVING MUTANT: " + row.file);
    result.nodeMutantCaught = true;
    results.push(result);
    console.log(`${row.file}: Node agrees; native=${result.native.stage}; javascript=${result.javascript.stage}; mutant caught`);
}
fs.writeFileSync(reportPath, JSON.stringify(results, null, 2) + "\n");
process.exitCode = results.every(row => ["native", "javascript"].every(backend => row[backend].agrees)) ? 0 : 1;
