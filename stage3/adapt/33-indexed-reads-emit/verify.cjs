#!/usr/bin/env node
"use strict";
const fs = require("node:fs");
const path = require("node:path");
const assert = require("node:assert/strict");
const { createHash } = require("node:crypto");
const { plan, files } = require("./adapt.cjs");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
if (ts.version !== "6.0.3") throw new Error(`want TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 4) throw new Error("usage: node verify.cjs <before-tree> <after-tree>");
const options = { target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext };
for (const file of require("../00-setup/test-shards.cjs").select(files)) {
    const before = fs.readFileSync(path.join(process.argv[2], "src/compiler", file), "utf8");
    const after = fs.readFileSync(path.join(process.argv[3], "src/compiler", file), "utf8");
    // Reviewed numeric U-zero operands are the sole allowed emitted changes.
    // All required-read assertions must disappear under the independent stock emitter.
    const zeroOnly = plan(ts, file, before, false, true).text;
    const expected = ts.transpileModule(zeroOnly, { fileName: file, compilerOptions: options });
    const actual = ts.transpileModule(after, { fileName: file, compilerOptions: options });
    if (actual.outputText !== expected.outputText) {
        const hash = text => createHash("sha256").update(text).digest("hex");
        throw new Error(`unexpected emitted JavaScript change in ${file}: ${hash(expected.outputText)} != ${hash(actual.outputText)}`);
    }
    plan(ts, file, after, true);
    assert.equal(plan(ts, file, after).text, after, `not idempotent: ${file}`);
    console.log(`PASS ${file}: stock emitted JavaScript bytes, site contract, idempotence`);
}
