"use strict";
const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const ts = require("typescript");
const pins = require("./pins.json");
const relative = "src/compiler/utilities.ts";

function locate(raw) {
    const source = ts.createSourceFile(relative, raw, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, relative + ": parse error");
    const matches = [];
    function scan(node) {
        if (ts.isFunctionDeclaration(node) && node.name?.text === "processResolution") matches.push(node);
        ts.forEachChild(node, scan);
    }
    scan(source);
    assert.equal(matches.length, 1, "Expected exactly one processResolution");
    const node = matches[0];
    assert(ts.isBlock(node.parent) && ts.isFunctionDeclaration(node.parent.parent) &&
        node.parent.parent.name?.text === "createSymlinkCache", "processResolution moved");
    const spelling = node.getText(source).replace(/\r\n/g, "\n");
    assert(spelling === pins.before || spelling === pins.after, "processResolution text drift");
    return { start: node.getStart(source), end: node.end, adapted: spelling === pins.after };
}

function adapt(tree) {
    assert.equal(ts.version, "6.0.3");
    const file = path.join(tree, relative);
    const raw = fs.readFileSync(file, "utf8");
    const state = locate(raw);
    if (state.adapted) return { files: 0, edits: 0 };
    const replacement = raw.includes("\r\n") ? pins.after.replace(/\n/g, "\r\n") : pins.after;
    const after = raw.slice(0, state.start) + replacement + raw.slice(state.end);
    assert(locate(after).adapted);
    assert.equal(fs.readFileSync(file, "utf8"), raw, "Source changed while planning");
    fs.writeFileSync(file, after);
    return { files: 1, edits: 1 };
}
module.exports = { adapt, locate, relative };
if (require.main === module) {
    assert.equal(process.argv.length, 3, "usage: node adapt.cjs TREE");
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2]))));
}
