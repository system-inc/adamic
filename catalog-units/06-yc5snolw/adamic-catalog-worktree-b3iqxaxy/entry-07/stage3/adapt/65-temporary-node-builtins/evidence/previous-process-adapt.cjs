"use strict";
// Only the existing Node host probe's type contract and erased cast change.
const fs = require("node:fs"), path = require("node:path"), assert = require("node:assert/strict");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
assert.equal(ts.version, "6.0.3");
assert.equal(process.argv.length, 3, "usage: adapt.cjs <adapted-tree-or-slice>");
const nodeTypes = process.env.CENSUS_NODE_TYPES || path.resolve(__dirname, "../../api/node_modules/@types/node");
const pin = JSON.parse(fs.readFileSync(path.join(nodeTypes, "package.json"), "utf8"));
assert.equal(pin.name, "@types/node");
assert.equal(pin.version, "25.3.3", "this temporary requires the library loader's exact Node bindings");
const file = "src/compiler/core.ts", target = path.join(process.argv[2], file);
const text = fs.readFileSync(target, "utf8"), source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
const owners = source.statements.filter(n => ts.isFunctionDeclaration(n) && n.name?.text === "isNodeLikeSystem" && n.body);
assert.equal(owners.length, 1, "isNodeLikeSystem owner drift");
const owner = owners[0], body = owner.getText(source);
const before = "(process as any).browser", after = "process.browser";
const oldDeclaration = "declare const process: { nextTick?: unknown; browser?: unknown; } | undefined;";
const newDeclaration = "declare const process: (NodeJS.Process & { browser?: unknown }) | undefined;";
const declarationCount = text.split(oldDeclaration).length - 1 + text.split(newDeclaration).length - 1;
assert.equal(declarationCount, 1, "local process declaration drift");
assert.equal(body.split(before).length - 1 + body.split(after).length - 1, 1, "browser probe drift");
if (body.includes(after)) {
    assert(text.includes(newDeclaration), "partially adapted process declaration");
    console.log(JSON.stringify({ file, already_adapted: true, node_types: pin.version }));
    process.exit(0);
}
assert(text.includes(oldDeclaration), "partially adapted browser probe");
const offset = owner.getStart(source) + body.indexOf(before);
let changed = text.slice(0, offset) + after + text.slice(offset + before.length);
changed = changed.replace(oldDeclaration, newDeclaration);
for (const removeComments of [false, true]) {
    const compilerOptions = { target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments };
    assert.equal(ts.transpileModule(changed, { compilerOptions }).outputText,
        ts.transpileModule(text, { compilerOptions }).outputText, "runtime JavaScript changed");
}
fs.writeFileSync(target, changed);
console.log(JSON.stringify({ file, owner: "isNodeLikeSystem", node_types: pin.version,
    edits: 2, javascript_identical: true }));
