"use strict";
// Each mutant tree holds only the two adapted files; every refusal happens before the adapter builds a program.
// usage: node refusals.cjs <adapt.cjs> <tree before> <tree after> <scratch directory> <output.json>
const fs = require("node:fs");
const path = require("node:path");
const crypto = require("node:crypto");
const { execFileSync } = require("node:child_process");
const [adapter, before, after, scratch, output] = process.argv.slice(2);
const read = (tree, name) => fs.readFileSync(path.join(tree, "src/compiler", name), "utf8");
const original = { core: read(before, "core.ts"), performance: read(before, "performanceCore.ts") };
const adapted = { core: read(after, "core.ts"), performance: read(after, "performanceCore.ts") };
const once = (text, from, to) => {
    if (text.split(from).length !== 2) throw new Error(`mutant anchor drift: ${from}`);
    return text.replace(from, to);
};
const cases = [
    { name: "core.ts adapted, performanceCore.ts original", core: adapted.core, performance: original.performance },
    { name: "core.ts original, performanceCore.ts adapted", core: original.core, performance: adapted.performance },
    { name: "the require test removed, its declaration kept", core: once(original.core, "        && !process.browser\r\n        && typeof require !== \"undefined\";", "        && !process.browser;"), performance: original.performance },
    { name: "the import added, the rest original", core: original.core, performance: once(original.performance, "import { isNodeLikeSystem } from \"./_namespaces/ts.js\";", "import { isNodeLikeSystem } from \"./_namespaces/ts.js\";\r\nimport * as perfHooks from \"node:perf_hooks\";") },
    { name: "adapted, 65's declaration kept", core: adapted.core, performance: adapted.performance + "\ndeclare const require: (specifier: \"perf_hooks\") => any;\n" },
    { name: "99q's before text, `(process as any).browser`", core: once(original.core, "        && !process.browser\r\n", "        && !(process as any).browser\r\n"), performance: original.performance },
    { name: "original, with require read in another function", core: once(original.core, "export function isNodeLikeSystem(): boolean {", "export function hostRequire(): unknown { return require; }\r\n/** @internal */\r\nexport function isNodeLikeSystem(): boolean {"), performance: original.performance },
    { name: "original, perfHooks already named", core: original.core, performance: once(original.performance, "const nativePerformanceHooks = tryGetPerformanceHooks();", "const perfHooks = 0;\r\nconst nativePerformanceHooks = tryGetPerformanceHooks();") },
    { name: "adapted, the binding outside its try block", core: adapted.core, performance: once(once(adapted.performance, "            const { performance } = perfHooks;\r\n", ""), "    if (isNodeLikeSystem()) {\r\n", "    const { performance } = perfHooks;\r\n    if (isNodeLikeSystem()) {\r\n") },
];
const hash = text => crypto.createHash("sha256").update(text).digest("hex");
const results = cases.map((entry, index) => {
    const tree = path.join(scratch, `refusal-${index}`);
    fs.mkdirSync(path.join(tree, "src/compiler"), { recursive: true });
    fs.writeFileSync(path.join(tree, "src/compiler/core.ts"), entry.core);
    fs.writeFileSync(path.join(tree, "src/compiler/performanceCore.ts"), entry.performance);
    let exit = 0, stderr = "";
    try { execFileSync(process.execPath, [adapter, tree], { encoding: "utf8", stdio: ["ignore", "pipe", "pipe"] }); }
    catch (error) { exit = error.status; stderr = String(error.stderr).trim(); }
    const unchanged = hash(read(tree, "core.ts")) === hash(entry.core) && hash(read(tree, "performanceCore.ts")) === hash(entry.performance);
    return { name: entry.name, exit, refusal: stderr, unchanged };
});
fs.writeFileSync(output, JSON.stringify(results, null, 2) + "\n");
for (const result of results) console.log(`${result.exit === 1 && result.unchanged ? "refused" : "accepted"}  ${result.name}: ${result.refusal}`);
if (results.some(result => result.exit !== 1 || !result.unchanged)) process.exitCode = 1;
