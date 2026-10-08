"use strict";
// Checked second-read invariant; stock TypeScript erases only this assertion.
const fs = require("node:fs"), path = require("node:path"), assert = require("node:assert/strict");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
assert.equal(ts.version, "6.0.3");
assert.equal(process.argv.length, 3, "usage: adapt.cjs <adapted-tree-or-slice>");
const file = "src/compiler/core.ts";
const owner = "addRange";
const before = "to.push(from[i])";
const after = "to.push(from[i]!)";
const target = path.join(process.argv[2], file), text = fs.readFileSync(target, "utf8");
const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
const owners = [];
function walk(node) {
    if (ts.isFunctionDeclaration(node) && node.name?.text === owner && node.body) owners.push(node);
    ts.forEachChild(node, walk);
}
walk(source);
assert.equal(owners.length, 1, "owner drift: " + owner);
const node = owners[0], body = node.getText(source);
assert.equal(body.split(before).length - 1 + body.split(after).length - 1, 1, "site drift/duplicate");
if (body.includes(after)) { console.log("already adapted " + file + ":" + owner); process.exit(0); }
const offset = node.getStart(source) + body.indexOf(before);
const changed = text.slice(0, offset) + after + text.slice(offset + before.length);
const options = { target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments: true };
assert.equal(ts.transpileModule(changed, { compilerOptions: options }).outputText,
    ts.transpileModule(text, { compilerOptions: options }).outputText, "runtime JavaScript changed");
fs.writeFileSync(target, changed);
console.log("adapted " + file + ":" + owner + " (runtime JavaScript identical)");
