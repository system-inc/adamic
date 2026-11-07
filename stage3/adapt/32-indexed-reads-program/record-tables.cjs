"use strict";
const pins = require("./record-tables-owners.json");
function plan(ts, text, check = false) {
    const source = ts.createSourceFile("symbolWalker.ts", text, ts.ScriptTarget.Latest, true);
    const owners = [];
    function visit(node) {
        if (ts.isFunctionDeclaration(node) && node.name?.text === "getSymbolWalker") owners.push(node);
        ts.forEachChild(node, visit);
    }
    visit(source);
    const imports = source.statements.filter(ts.isImportDeclaration);
    const owner = owners[0];
    if (source.parseDiagnostics.length || owners.length !== 1 || imports.length !== 1 ||
        !ts.isBlock(owner.parent) || !ts.isFunctionDeclaration(owner.parent.parent) ||
        owner.parent.parent.name?.text !== "createGetSymbolWalker") throw new Error("record table lexical owner drift");
    const actual = owner.getText(source), imported = imports[0].getText(source);
    for (const variant of pins.variants) {
        if (actual === variant.after && imported === pins.imports.after) return { text, edits: 0 };
        if (actual !== variant.before || imported !== pins.imports.before) continue;
        if (check) throw new Error("symbolWalker record table amendment missing");
        // Plan both complete owners together. No partial table/import/cleanup state is accepted.
        const edits = [{ start: owner.getStart(source), end: owner.end, text: variant.after },
            { start: imports[0].getStart(source), end: imports[0].end, text: pins.imports.after }];
        for (const edit of edits.sort((a, b) => b.start - a.start)) text = text.slice(0, edit.start) + edit.text + text.slice(edit.end);
        return { text, edits: 1 };
    }
    throw new Error("symbolWalker whole owner/import drift or half-adapted owner");
}
module.exports = { plan };
