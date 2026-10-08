#!/usr/bin/env node
"use strict";
const fs = require("node:fs");
const path = require("node:path");
const assert = require("node:assert/strict");
const { createHash } = require("node:crypto");
const { plan, files } = require("./adapt.cjs");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
const { normalizeClampJS, compareClamps } = require("./scanner-proof.cjs");
const { normalizeMissingJS, compareMissingGuards } = require("./parser-proof.cjs");
const { normalizeDecoderJS, compareDecoders } = require("./utilities-proof.cjs");
if (ts.version !== "6.0.3") throw new Error(`want TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 4) throw new Error("usage: node verify.cjs <before-tree> <after-tree>");
const options = { target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext };
for (const file of require("../00-setup/test-shards.cjs").select(files)) {
    const before = fs.readFileSync(path.join(process.argv[2], "src/compiler", file), "utf8");
    const after = fs.readFileSync(path.join(process.argv[3], "src/compiler", file), "utf8");
    // Permit only the four U-zero operands and the separately proven scanner narrowing.
    // All required-read assertions must disappear under the independent stock emitter.
    const zeroOnly = plan(ts, file, before, false, true).text;
    const expected = ts.transpileModule(zeroOnly, { fileName: file, compilerOptions: options });
    const actual = ts.transpileModule(after, { fileName: file, compilerOptions: options });
    const normalize = file === "scanner.ts" ? normalizeClampJS : file === "parser.ts" ? normalizeMissingJS : file === "utilities.ts" ? normalizeDecoderJS : text => text;
    if (normalize(actual.outputText) !== normalize(expected.outputText)) {
        const hash = text => createHash("sha256").update(text).digest("hex");
        throw new Error(`unexpected emitted JavaScript change in ${file}: ${hash(expected.outputText)} != ${hash(actual.outputText)}`);
    }
    if (file === "scanner.ts") console.log(JSON.stringify(compareClamps(ts, before, after)));
    if (file === "parser.ts") console.log(JSON.stringify(compareMissingGuards(ts, before, after, fs.readFileSync(path.join(process.argv[3], "src/compiler/utilities.ts"), "utf8"))));
    if (file === "utilities.ts") console.log(JSON.stringify(compareDecoders(ts, before, after)));
    plan(ts, file, after, true);
    assert.equal(plan(ts, file, after).text, after, `not idempotent: ${file}`);
    console.log(`PASS ${file}: ${file === "scanner.ts" || file === "parser.ts" || file === "utilities.ts" ? "only independently proven narrowing restructuring" : "stock emitted JavaScript bytes"}, site contract, idempotence`);
}
