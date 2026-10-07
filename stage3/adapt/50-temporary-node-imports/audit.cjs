#!/usr/bin/env node
"use strict";
const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const os = require("node:os");
const crypto = require("node:crypto");
const { execFileSync } = require("node:child_process");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
const tree = path.resolve(process.argv[2]);
const adapter = path.join(__dirname, "adapt.cjs");
function run(root) { return JSON.parse(execFileSync(process.execPath, [adapter, root], { encoding: "utf8" })); }
function hashes(root) {
    return ts.sys.readDirectory(path.join(root, "src"), [".ts"]).sort().map(file =>
        [path.relative(root, file), crypto.createHash("sha256").update(fs.readFileSync(file)).digest("hex")]);
}
const expected = [
    ["src/compiler/performanceCore.ts", 35, "perf_hooks"],
    ["src/compiler/sys.ts", 1470, "fs"],
    ["src/compiler/sys.ts", 1471, "path"],
    ["src/compiler/sys.ts", 1472, "os"],
    ["src/compiler/sys.ts", 1476, "crypto"],
    ["src/compiler/sys.ts", 1650, "inspector"],
    ["src/compiler/tracing.ts", 63, "fs"],
];
function validate(report) {
    assert.equal(report.changed, 0);
    assert.deepEqual(report.declined.filter(site => site.file.startsWith("src/compiler/")).map(site => [site.file, site.line, site.specifier]), expected);
    for (const site of report.declined) assert.ok(site.reason.includes("load is lazy"));
    assert.ok(report.retained.some(site => site.specifier === "source-map-support"));
    assert.ok(report.retained.some(site => site.file === "src/compiler/sys.ts" && site.specifier === null));
}
const before = hashes(tree);
const first = run(tree);
validate(first);
assert.deepEqual(hashes(tree), before);
const second = run(tree);
assert.deepEqual(second, first);
assert.deepEqual(hashes(tree), before);
console.log("pass: seven builtin decline sites; source-map-support and dynamic plugin retained; source hashes unchanged; idempotent");
// A real source-input mutant changes the inspector load's specifier. The
// independently pinned site census must reject it, even though parsing succeeds.
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), "host-adapt-mutant-"));
try {
    fs.mkdirSync(path.join(scratch, "src/compiler"), { recursive: true });
    fs.copyFileSync(path.join(tree, "package-lock.json"), path.join(scratch, "package-lock.json"));
    for (const file of ["performanceCore.ts", "sys.ts", "tracing.ts"]) {
        let text = fs.readFileSync(path.join(tree, "src/compiler", file), "utf8");
        if (file === "sys.ts") {
            assert.ok(text.includes('require("inspector")'));
            text = text.replace('require("inspector")', 'require("http")');
        }
        fs.writeFileSync(path.join(scratch, "src/compiler", file), text);
    }
    assert.throws(() => validate(run(scratch)), { code: "ERR_ASSERTION" });
    console.log("caught mutant: inspector literal changed to http, rejected by exact builtin-site census");
    // Restore the load before a comment-only input mutation. Site inventory
    // and repeat reports still pass; only the byte-preservation check fails.
    const sysFile = path.join(scratch, "src/compiler/sys.ts");
    fs.copyFileSync(path.join(tree, "src/compiler/sys.ts"), sysFile);
    const control = run(scratch);
    validate(control);
    const controlHashes = hashes(scratch);
    fs.appendFileSync(sysFile, "\n// host source hash mutant\n");
    const commented = run(scratch);
    validate(commented);
    assert.deepEqual(commented, control);
    assert.throws(() => assert.deepEqual(hashes(scratch), controlHashes), { code: "ERR_ASSERTION" });
    console.log("caught mutant: trailing source comment changed bytes, rejected only by source-hash comparison");
    const lockFile = path.join(scratch, "package-lock.json");
    const lock = JSON.parse(fs.readFileSync(lockFile, "utf8"));
    lock.packages["node_modules/@types/node"].version = "25.3.2";
    fs.writeFileSync(lockFile, JSON.stringify(lock));
    assert.throws(() => run(scratch), error => error.stderr?.includes("want upstream @types/node@25.3.3"));
    console.log("caught mutant: @types/node lock pin changed to 25.3.2, rejected by declaration-version guard");
}
finally { fs.rmSync(scratch, { recursive: true, force: true }); }
