"use strict";
// Inventory every retained reference, not a file-graph or runtime-coverage proof.
const fs = require("node:fs"), path = require("node:path"), assert = require("node:assert/strict"), crypto = require("node:crypto");
const ts = require(process.env.CENSUS_TYPESCRIPT || "typescript");
assert.equal(ts.version, "6.0.3");
const [tree, output] = process.argv.slice(2);
assert(tree && output, "usage: audit-range-callers.cjs <slice> <output.json>");
const calls = [], references = [], callbacks = [], hashes = {};
function owner(node) {
    const names = [];
    for (let p = node.parent; p; p = p.parent) {
        if (ts.isFunctionDeclaration(p) && p.name) names.unshift(p.name.text);
        if (ts.isModuleDeclaration(p)) names.unshift(p.name.text);
    }
    return names.join(".");
}
function walkFiles(dir) {
    for (const entry of fs.readdirSync(dir, { withFileTypes: true }).sort((a,b) => a.name.localeCompare(b.name))) {
        const file = path.join(dir, entry.name);
        if (entry.isDirectory()) walkFiles(file);
        else if (entry.name.endsWith(".ts")) {
            const relative = path.relative(tree, file), text = fs.readFileSync(file, "utf8");
            hashes[relative] = crypto.createHash("sha256").update(text).digest("hex");
            const sf = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
            function visit(n) {
                if (ts.isIdentifier(n) && n.text === "addRange") {
                    const p = n.parent;
                    if (!ts.isImportSpecifier(p) && !ts.isExportSpecifier(p) && !(ts.isFunctionDeclaration(p) && p.name === n)) {
                        const site = { file: relative, line: sf.getLineAndCharacterOfPosition(n.getStart(sf)).line + 1, owner: owner(n), text: p.getText(sf) };
                        references.push(site);
                        assert(ts.isCallExpression(p) && p.expression === n, "unreviewed addRange alias/reference: " + JSON.stringify(site));
                        calls.push({ ...site, arguments: p.arguments.map(a => a.getText(sf)) });
                    }
                }
                if (ts.isCallExpression(n) && ts.isIdentifier(n.expression) && ["flatMap", "sameFlatMap"].includes(n.expression.text)) {
                    callbacks.push({ file: relative, owner: owner(n), callee: n.expression.text, text: n.getText(sf), callback: n.arguments[1].getText(sf) });
                }
                ts.forEachChild(n, visit);
            }
            visit(sf);
        }
    }
}
walkFiles(path.join(tree, "src/compiler"));
const expected = new Map([
    ["src/compiler/core.ts:flatMap", 1], ["src/compiler/core.ts:sameFlatMap", 1],
    ["src/compiler/factory/nodeFactory.ts:mergeEmitNode", 2],
    ["src/compiler/utilities.ts:getJSDocCommentsAndTags", 4],
    ["src/compiler/parser.ts:Parser.reparseTopLevelAwait", 4],
    ["src/compiler/parser.ts:Parser.JSDocParser.parseJSDocComment", 1],
]);
for (const site of calls) {
    const key = site.file + ":" + site.owner;
    assert(expected.has(key) && expected.get(key) > 0, "unreviewed caller: " + key);
    expected.set(key, expected.get(key) - 1);
}
assert([...expected.values()].every(n => n === 0));
assert.equal(calls.length, 13);
assert.equal(callbacks.length, 3, "unreviewed flattening callback");
assert(callbacks.some(c => c.owner === "filterOwnedJSDocTags" && c.callback.startsWith("jsDoc =>")));
assert(callbacks.some(c => c.owner === "getJSDocTagsWorker" && c.callback === "j => isJSDoc(j) ? j.tags : j"));
assert(callbacks.some(c => c.owner === "createNodeFactory.createCommaListExpression" && c.callback === "flattenCommaElements"));
fs.writeFileSync(output, JSON.stringify({ calls, callbacks, references: references.length, sourceHashes: hashes,
    scope: "Retained declaration inventory; producer evidence in adaptation 64 README. Corpus observations are not universal reachability proofs." }, null, 2) + "\n");
console.log("13 reviewed direct calls, no addRange aliases, three internal flattening callbacks.");
